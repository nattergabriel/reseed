package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(libraryCmd)
}

var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "List all skills in your library",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		skills, err := lib.ListSkills()
		if err != nil {
			return err
		}

		if len(skills) == 0 {
			fmt.Println("No skills in library.")
			return nil
		}

		for _, name := range skills {
			suffix := ""
			if lib.IsExternal(name) {
				suffix = " (external)"
			}
			fmt.Printf("  %s%s\n", name, suffix)
		}

		return nil
	},
}
