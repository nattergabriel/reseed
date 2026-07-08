package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/nattergabriel/reseed/internal/skill"
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

		printList(entries, listLong)
		return nil
	},
}

// printList prints entries grouped by folder. Entries are sorted by group,
// so top-level skills come first, followed by one section per folder.
func printList(entries []skill.SkillEntry, long bool) {
	prevGroup := ""
	for i, e := range entries {
		if e.Group != prevGroup {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", e.Group)
			prevGroup = e.Group
		}

		name := e.Name
		if e.Group != "" {
			name = "  " + name
		}
		if long {
			if desc := skill.ReadDescription(e.Path); desc != "" {
				fmt.Printf("%s - %s\n", name, desc)
				continue
			}
		}
		fmt.Println(name)
	}
}
