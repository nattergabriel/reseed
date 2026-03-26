package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh/spinner"
	"github.com/nattergabriel/reseed/internal/github"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().StringP("pack", "p", "", "install skills into a pack directory")
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
  reseed install user/repo/src/skills -p kit  # install into a pack`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		packName, _ := cmd.Flags().GetString("pack")
		client := github.NewClient()

		destDir := lib.SkillsDir()
		if packName != "" {
			destDir = filepath.Join(lib.SkillsDir(), packName)
		}

		var total int
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
					skills, ferr = client.FetchSkills(ctx, ref, destDir)
					return ferr
				}).
				Run()
			if err != nil {
				return err
			}

			for _, s := range skills {
				fmt.Printf("  + %s\n", s.Name)
			}
			total += len(skills)
		}

		printSummary("Installed", total)
		return nil
	},
}
