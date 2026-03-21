package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect <pack>",
	Short: "Show the skills in a pack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		skills, err := lib.ResolvePack(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Pack %q:\n", args[0])
		for _, name := range skills {
			status := "ok"
			if !lib.HasSkill(name) {
				status = "missing"
			}
			fmt.Printf("  %s (%s)\n", name, status)
		}

		return nil
	},
}
