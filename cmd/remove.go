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
	Use:   "remove <skill>",
	Short: "Remove a skill from the current project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.RemoveSkill(args[0]); err != nil {
			return err
		}
		fmt.Printf("  - %s\n", args[0])
		return nil
	},
}
