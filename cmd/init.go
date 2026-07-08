package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/reseed"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:     "init [path]",
	Short:   "Initialize a skill library",
	GroupID: groupLibrary,
	Long:    "Creates a skill library at the given path (or current directory), or recognizes an existing one.",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		lib, err := reseed.InitLibrary(path)
		if err != nil {
			return err
		}

		fmt.Printf("Library initialized at %s\n", lib.Path)
		if len(lib.Skills) > 0 {
			fmt.Printf("Found %d existing %s:\n", len(lib.Skills), skillNoun(len(lib.Skills)))
			for _, s := range lib.Skills {
				fmt.Printf("  %s\n", s.Name)
			}
		}

		return nil
	},
}
