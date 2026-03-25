package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/charmbracelet/huh/spinner"
	"github.com/nattergabriel/reseed/internal/config"
	"github.com/nattergabriel/reseed/internal/github"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().StringP("pack", "p", "", "create a pack with all installed skills")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:     "install <user/repo[/path][@version]>",
	Short:   "Fetch skills from a GitHub repo into your library",
	GroupID: groupLibrary,
	Long: `Downloads skills from GitHub repositories and adds them to your library.

Examples:
  reseed install user/repo                    # all skills from the repo
  reseed install user/repo/src/skills/commit  # one specific skill
  reseed install user/repo/src/skills         # all skills under a directory
  reseed install user/repo@v2.0               # pin to a tag
  reseed install user/repo user2/repo2        # multiple sources at once
  reseed install user/repo/src/skills -p kit  # install and create a pack`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		packName, _ := cmd.Flags().GetString("pack")
		client := github.NewClient()
		var packSkills []string // only used when packName != ""

		for _, arg := range args {
			ref, err := github.ParseRef(arg)
			if err != nil {
				return err
			}

			versionStr := ref.Version
			if versionStr == "" {
				versionStr = github.VersionLatest
			}

			source := fmt.Sprintf("%s/%s", ref.Owner, ref.Repo)
			if ref.Path != "" {
				source += "/" + ref.Path
			}

			var skills []github.ExtractedSkill
			err = spinner.New().
				Title(fmt.Sprintf("  Fetching %s (%s)...", source, versionStr)).
				ActionWithErr(func(ctx context.Context) error {
					var ferr error
					skills, ferr = client.FetchSkills(ctx, ref, lib.SkillsDir())
					return ferr
				}).
				Run()
			if err != nil {
				return err
			}

			for _, skill := range skills {
				lib.Config.Sources[skill.Name] = config.Source{
					Source:  ref.SourceString(skill.Path),
					Version: versionStr,
				}
				if packName != "" {
					packSkills = append(packSkills, skill.Name)
				}
				fmt.Printf("  + %s\n", skill.Name)
			}
		}

		if packName != "" && len(packSkills) > 0 {
			sort.Strings(packSkills)
			lib.Config.Packs[packName] = packSkills
			fmt.Printf("  Pack %q created with %d skills\n", packName, len(packSkills))
		}

		if err := lib.SaveConfig(); err != nil {
			return err
		}

		return nil
	},
}
