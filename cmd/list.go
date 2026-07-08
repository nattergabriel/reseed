package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/reseed"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.Flags().BoolVarP(&listLong, "long", "l", false, "Show skill descriptions")
	rootCmd.AddCommand(listCmd)
}

var listLong bool

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List skills in your library",
	GroupID: groupLibrary,
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := reseed.OpenLibrary()
		if err != nil {
			return err
		}

		if len(lib.Skills) == 0 {
			fmt.Println("No skills in library.")
			return nil
		}

		printList(lib.Skills, listLong)
		return nil
	},
}

// printList prints skills grouped by folder. Skills are sorted by group, so
// top-level skills come first, followed by one section per folder.
func printList(skills []reseed.Skill, long bool) {
	prevGroup := ""
	for i, s := range skills {
		if s.Group != prevGroup {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", s.Group)
			prevGroup = s.Group
		}

		name := s.Name
		if s.Group != "" {
			name = "  " + name
		}
		if long {
			if desc := reseed.ReadDescription(s.Path); desc != "" {
				fmt.Printf("%s - %s\n", name, desc)
				continue
			}
		}
		fmt.Println(name)
	}
}
