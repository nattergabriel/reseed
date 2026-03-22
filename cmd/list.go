package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List skills installed in the current project",
	GroupID: groupProject,
	RunE: func(cmd *cobra.Command, args []string) error {
		skills, err := project.ListInstalled()
		if err != nil {
			return err
		}

		if len(skills) == 0 {
			fmt.Println("No skills installed in this project.")
			return nil
		}

		for _, name := range skills {
			fmt.Printf("  %s\n", name)
		}

		return nil
	},
}
