package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh/spinner"
	"github.com/nattergabriel/reseed/internal/github"
	"github.com/nattergabriel/reseed/internal/reseed"
	"github.com/spf13/cobra"
)

var installInto string

func init() {
	installCmd.Flags().StringVar(&installInto, "into", "", "install skills into a library subfolder")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:     "install <user/repo[/path]>",
	Short:   "Fetch skills from a GitHub repo into your library",
	GroupID: groupLibrary,
	Long: `Downloads skills from GitHub repositories and adds them to your library.

Examples:
  reseed install user/repo                            # all skills from the repo
  reseed install user/repo/src/skills/commit          # one specific skill
  reseed install user/repo/src/skills                 # all skills under a directory
  reseed install user/repo user2/repo2                # multiple sources at once
  reseed install user/repo/src/skills --into kit      # install into a subfolder`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := reseed.OpenLibrary()
		if err != nil {
			return err
		}

		destDir := lib.SkillsDir()
		if installInto != "" {
			destDir = filepath.Join(destDir, installInto)
		}

		var total int
		for _, arg := range args {
			ref, err := github.ParseRef(arg)
			if err != nil {
				return err
			}

			var skills []string
			err = spinner.New().
				Title(fmt.Sprintf("  Fetching %s...", ref)).
				ActionWithErr(func(ctx context.Context) error {
					var ferr error
					skills, ferr = github.FetchSkills(ctx, ref, destDir)
					return ferr
				}).
				Run()
			if err != nil {
				return err
			}

			for _, s := range skills {
				fmt.Printf("  + %s\n", s)
			}
			total += len(skills)
		}

		printSummary("Installed", total)
		return nil
	},
}
