package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/reseed"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <skills or folders...>",
	Short:   "Add skills or folders to the current project",
	GroupID: groupProject,
	Long:    "Copies skills from your library into the project's .agents/skills/ directory. Naming a library folder adds every skill under it.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := reseed.OpenLibrary()
		if err != nil {
			return err
		}
		proj, err := reseed.OpenProject(flagDir)
		if err != nil {
			return err
		}

		var skills []reseed.Skill
		for _, arg := range args {
			resolved, err := lib.Resolve(arg)
			if err != nil {
				return err
			}
			skills = append(skills, resolved...)
		}

		for _, s := range skills {
			if err := proj.Add(s); err != nil {
				return fmt.Errorf("adding %s: %w", s.Name, err)
			}
			fmt.Printf("  + %s\n", s.Name)
		}

		printSummary("Added", len(skills))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
