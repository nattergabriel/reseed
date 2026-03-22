package cmd

import (
	"fmt"

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

		m := packModel{
			packName: packName,
			skills:   skills,
			selected: selected,
		}

		p := tea.NewProgram(m)
		result, err := p.Run()
		if err != nil {
			return err
		}

		final := result.(packModel)
		if final.cancelled {
			fmt.Println("Cancelled.")
			return nil
		}

		var chosen []string
		for i, name := range final.skills {
			if final.selected[i] {
				chosen = append(chosen, name)
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

type packModel struct {
	packName  string
	skills    []string
	selected  map[int]bool
	cursor    int
	cancelled bool
}

func (m packModel) Init() tea.Cmd {
	return nil
}

func (m packModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.skills)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m packModel) View() string {
	s := fmt.Sprintf("Pack %q - select skills (space to toggle, enter to confirm):\n\n", m.packName)

	for i, name := range m.skills {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}

		s += fmt.Sprintf("%s%s %s\n", cursor, check, name)
	}

	s += "\nq/esc to cancel"
	return s
}
