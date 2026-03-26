package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Sync project skills from your library",
	GroupID: groupProject,
	Long:  "Re-copies skills from the library into the project. Matches by name — skills not in the library are left untouched.",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		updated, err := project.SyncSkills(lib)
		if err != nil {
			return err
		}

		if len(updated) == 0 {
			fmt.Println("Nothing to sync.")
			return nil
		}

		for _, name := range updated {
			fmt.Printf("  ~ %s\n", name)
		}

		printSummary("Synced", len(updated))
		return nil
	},
}
