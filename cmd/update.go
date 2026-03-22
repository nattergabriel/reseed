package cmd

import (
	"fmt"
	"strings"

	"github.com/nattergabriel/reseed/internal/github"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Fetch latest versions of external skills",
	GroupID: groupLibrary,
	Long:  "Re-fetches external skills from GitHub. Pinned versions are skipped; 'latest' skills get the newest tag.",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		if len(lib.Config.Sources) == 0 {
			fmt.Println("No external skills to update.")
			return nil
		}

		client := github.NewClient()
		var errors []string

		for name, src := range lib.Config.Sources {
			if src.Version != "latest" {
				fmt.Printf("  - %s (pinned at %s, skipped)\n", name, src.Version)
				continue
			}

			// Parse the source string to get owner/repo/skill
			ref, err := parseSourceString(src.Source, name)
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

// parseSourceString parses "user/repo/skill" back into a SkillRef.
func parseSourceString(source, skillName string) (*github.SkillRef, error) {
	parts := strings.Split(source, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid source: %s", source)
	}

	ref := &github.SkillRef{
		Owner:   parts[0],
		Repo:    parts[1],
		Version: "latest",
	}

	if len(parts) >= 3 {
		ref.Skill = parts[2]
	} else {
		ref.Skill = skillName
	}

	return ref, nil
}
