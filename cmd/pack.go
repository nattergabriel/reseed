package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(packCmd)
}

var packCmd = &cobra.Command{
	Use:     "pack <name>",
	Short:   "Create or edit a pack interactively",
	GroupID: groupLibrary,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packName := args[0]

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

		selected := make(map[int]bool)
		if current, ok := lib.Config.Packs[packName]; ok {
			inPack := make(map[string]bool, len(current))
			for _, s := range current {
				inPack[s] = true
			}
			for i, name := range skills {
				if inPack[name] {
					selected[i] = true
				}
			}
		}

		var items []list.Item
		for _, name := range skills {
			items = append(items, addItem{name: name})
		}

		delegate := checkboxDelegate{selected: selected}
		l := list.New(items, delegate, 0, 0)
		l.Title = fmt.Sprintf("Pack %q (space: toggle, enter: confirm)", packName)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)

		m := checkboxModel{
			list:     l,
			selected: selected,
		}

		p := tea.NewProgram(m, tea.WithAltScreen())
		result, err := p.Run()
		if err != nil {
			return err
		}

		final := result.(checkboxModel)
		if final.cancelled {
			fmt.Println("Cancelled.")
			return nil
		}

		var chosen []string
		for i, listItem := range final.list.Items() {
			if final.selected[i] {
				chosen = append(chosen, listItem.(addItem).name)
			}
		}

		if len(chosen) == 0 {
			delete(lib.Config.Packs, packName)
			if err := lib.SaveConfig(); err != nil {
				return err
			}
			fmt.Printf("Pack %q removed (no skills selected).\n", packName)
			return nil
		}

		lib.Config.Packs[packName] = chosen
		if err := lib.SaveConfig(); err != nil {
			return err
		}

		fmt.Printf("Pack %q saved with %d skill(s):\n", packName, len(chosen))
		for _, name := range chosen {
			fmt.Printf("  + %s\n", name)
		}

		return nil
	},
}
