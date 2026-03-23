package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/config"
	"github.com/nattergabriel/reseed/internal/github"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
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
  reseed install user/repo user2/repo2        # multiple sources at once`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		client := github.NewClient()

		for _, arg := range args {
			ref, err := github.ParseRef(arg)
			if err != nil {
				return err
			}

			versionStr := ref.Version
			if versionStr == "" {
				versionStr = github.VersionLatest
			}

			fmt.Printf("Fetching from %s/%s", ref.Owner, ref.Repo)
			if ref.Path != "" {
				fmt.Printf("/%s", ref.Path)
			}
			fmt.Printf(" (%s)...\n", versionStr)

			skills, err := client.FetchSkills(ref, lib.SkillsDir())
			if err != nil {
				return err
			}

			for _, skill := range skills {
				lib.Config.Sources[skill.Name] = config.Source{
					Source:  ref.SourceString(skill.Path),
					Version: versionStr,
				}
				fmt.Printf("  + %s\n", skill.Name)
			}
		}

		if err := lib.SaveConfig(); err != nil {
			return err
		}

		return nil
	},
}
