package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:     "remove <skills...>",
	Short:   "Remove skills from the current project",
	GroupID: groupProject,
	Long:    "Removes skills from the project's skills directory.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		installedSet, err := project.InstalledSet()
		if err != nil {
			return err
		}

		var removed int
		for _, name := range args {
			if !installedSet[name] {
				return fmt.Errorf("skill %q not installed", name)
			}
			if err := project.RemoveSkill(name); err != nil {
				return fmt.Errorf("removing %s: %w", name, err)
			}
			fmt.Printf("  - %s\n", name)
			removed++
		}

		printSummary("Removed", removed)
		return nil
	},
}
