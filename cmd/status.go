package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show skills installed in the current project",
	GroupID: groupProject,
	RunE: func(cmd *cobra.Command, args []string) error {
		installed, err := project.ListInstalled()
		if err != nil {
			return err
		}

		if len(installed) == 0 {
			fmt.Println("No skills installed.")
			return nil
		}

		for _, name := range installed {
			fmt.Println(name)
		}

		return nil
	},
}
