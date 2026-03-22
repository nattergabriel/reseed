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
	Use:     "install <user/repo[/skill][@version]>",
	Short:   "Fetch skills from a GitHub repo into your library",
	GroupID: groupLibrary,
	Long: `Downloads skills from GitHub repositories and adds them to your library.

Examples:
  reseed install user/repo              # all skills from the repo
  reseed install user/repo/my-skill     # one specific skill
  reseed install user/repo@v2.0         # pin to a tag
  reseed install user/repo user2/repo2  # multiple sources at once`,
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
				versionStr = "latest"
			}

			fmt.Printf("Fetching from %s/%s", ref.Owner, ref.Repo)
			if ref.Skill != "" {
				fmt.Printf("/%s", ref.Skill)
			}
			fmt.Printf(" (%s)...\n", versionStr)

			skills, err := client.FetchSkills(ref, lib.SkillsDir())
			if err != nil {
				return err
			}

			for _, name := range skills {
				lib.Config.Sources[name] = config.Source{
					Source:  ref.SourceString(name),
					Version: versionStr,
				}
				fmt.Printf("  + %s\n", name)
			}
		}

		if err := lib.SaveConfig(); err != nil {
			return err
		}

		return nil
	},
}
