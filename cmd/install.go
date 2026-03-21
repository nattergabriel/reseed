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
	Use:   "install <user/repo[/skill][@version]>",
	Short: "Fetch skills from a GitHub repo into your library",
	Long: `Downloads skills from a GitHub repository and adds them to your library.

Examples:
  reseed install user/repo              # all skills from the repo
  reseed install user/repo/my-skill     # one specific skill
  reseed install user/repo@v2.0         # pin to a tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := github.ParseRef(args[0])
		if err != nil {
			return err
		}

		lib, err := library.Open()
		if err != nil {
			return err
		}

		client := github.NewClient()

		// Resolve version for display
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

		// Track sources in config
		for _, name := range skills {
			lib.Config.Sources[name] = config.Source{
				Source:  ref.SourceString(name),
				Version: versionStr,
			}
		}

		if err := lib.SaveConfig(); err != nil {
			return err
		}

		for _, name := range skills {
			fmt.Printf("  + %s\n", name)
		}

		return nil
	},
}
