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

		if listLong {
			printListLong(entries)
		} else {
			printListShort(entries)
		}

		return nil
	},
}

func printListShort(entries []skill.SkillEntry) {
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
}

func printListLong(entries []skill.SkillEntry) {
	descByKey := make(map[string]string, len(entries))
	for _, e := range entries {
		key := e.Name
		if e.Pack != "" {
			key = e.Pack + "/" + e.Name
		}
		descByKey[key] = skill.ReadDescription(e.Path)
	}

	skills, packs := buildSkillsAndPacks(entries)

	for _, name := range skills {
		printSkillLong(name, descByKey[name])
	}

	for i, p := range packs {
		if len(skills) > 0 || i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s:\n", p.name)
		for _, s := range p.skills {
			printSkillLong("  "+s, descByKey[p.name+"/"+s])
		}
	}
}

func printSkillLong(name, desc string) {
	if desc != "" {
		fmt.Printf("%s - %s\n", name, desc)
	} else {
		fmt.Println(name)
	}
}
