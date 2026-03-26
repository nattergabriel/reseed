package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List skills and packs in your library",
	GroupID: groupLibrary,
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open()
		if err != nil {
			return err
		}

		entries, err := lib.ListSkillEntries()
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Println("No skills in library.")
			return nil
		}

		skills, packs := buildSkillsAndPacks(entries)

		for _, name := range skills {
			fmt.Println(name)
		}

		for i, p := range packs {
			if len(skills) > 0 || i > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", p.name)
			for _, s := range p.skills {
				fmt.Printf("  %s\n", s)
			}
		}

		return nil
	},
}
