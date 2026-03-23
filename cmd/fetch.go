package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nattergabriel/reseed/internal/github"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fetchCmd)
}

var fetchCmd = &cobra.Command{
	Use:     "fetch",
	Short:   "Fetch latest versions of external skills from GitHub",
	GroupID: groupLibrary,
	Long:  "Fetches external skills from GitHub into your library. Pinned versions are skipped; 'latest' skills get the newest tag.",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		if len(lib.Config.Sources) == 0 {
			fmt.Println("No external skills to fetch.")
			return nil
		}

		client := github.NewClient()
		var errors []string

		names := make([]string, 0, len(lib.Config.Sources))
		for name := range lib.Config.Sources {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			src := lib.Config.Sources[name]
			if src.Version != github.VersionLatest {
				fmt.Printf("  - %s (pinned at %s, skipped)\n", name, src.Version)
				continue
			}

			ref, err := github.ParseRef(src.Source)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", name, err))
				continue
			}

			fmt.Printf("  ~ %s...\n", name)

			_, err = client.FetchSkills(ref, lib.SkillsDir())
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", name, err))
				continue
			}
		}

		if len(errors) > 0 {
			return fmt.Errorf("some skills failed to update:\n  %s", strings.Join(errors, "\n  "))
		}

		return nil
	},
}
