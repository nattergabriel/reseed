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

		entries, err := lib.ListSkillEntries()
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Println("No skills in library.")
			return nil
		}

		// Group by pack
		var standalone []string
		packs := make(map[string][]string)
		var packOrder []string
		for _, e := range entries {
			if e.Pack == "" {
				standalone = append(standalone, e.Name)
			} else {
				if _, exists := packs[e.Pack]; !exists {
					packOrder = append(packOrder, e.Pack)
				}
				packs[e.Pack] = append(packs[e.Pack], e.Name)
			}
		}

		if len(standalone) > 0 {
			for _, name := range standalone {
				fmt.Printf("  %s\n", name)
			}
		}

		for _, pack := range packOrder {
			if len(standalone) > 0 || pack != packOrder[0] {
				fmt.Println()
			}
			fmt.Printf("%s:\n", pack)
			for _, name := range packs[pack] {
				fmt.Printf("  %s\n", name)
			}
		}

		return nil
	},
}
