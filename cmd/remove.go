package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/reseed"
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
		proj, err := reseed.OpenProject(flagDir)
		if err != nil {
			return err
		}

		for _, name := range args {
			if err := proj.Remove(name); err != nil {
				return err
			}
			fmt.Printf("  - %s\n", name)
		}

		printSummary("Removed", len(args))
		return nil
	},
}
