package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a skill library",
	Long:  "Creates a skill library at the given path (or current directory), or recognizes an existing one.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		lib, err := library.Init(path)
		if err != nil {
			return err
		}

		skills, err := lib.ListSkills()
		if err != nil {
			return err
		}

		fmt.Printf("Library initialized at %s\n", lib.Path)
		if len(skills) > 0 {
			fmt.Printf("Found %d existing skill(s):\n", len(skills))
			for _, s := range skills {
				fmt.Printf("  %s\n", s)
			}
		}

		return nil
	},
}
