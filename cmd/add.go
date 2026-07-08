package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <skills or folders...>",
	Short:   "Add skills or folders to the current project",
	GroupID: groupProject,
	Long:    "Copies skills from your library into the project's .agents/skills/ directory. Naming a library folder adds every skill under it.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		var skills []string
		for _, arg := range args {
			resolved, err := lib.Resolve(arg)
			if err != nil {
				return err
			}
			skills = append(skills, resolved...)
		}

		for _, name := range skills {
			if err := project.AddSkill(lib, name); err != nil {
				return fmt.Errorf("adding %s: %w", name, err)
			}
			fmt.Printf("  + %s\n", name)
		}

		printSummary("Added", len(skills))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
