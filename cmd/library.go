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
	Use:     "library",
	Short:   "List all skills in your library",
	GroupID: groupLibrary,
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

		var local, external []string
		for _, name := range skills {
			if lib.IsExternal(name) {
				external = append(external, name)
			} else {
				local = append(local, name)
			}
		}

		if len(local) > 0 {
			fmt.Println("Local:")
			for _, name := range local {
				fmt.Printf("  %s\n", name)
			}
		}

		if len(external) > 0 {
			if len(local) > 0 {
				fmt.Println()
			}
			fmt.Println("External:")
			for _, name := range external {
				src := lib.Config.Sources[name]
				fmt.Printf("  %s  (%s, %s)\n", name, src.Source, src.Version)
			}
		}

		return nil
	},
}
