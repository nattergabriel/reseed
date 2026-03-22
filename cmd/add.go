package cmd

import (
	"fmt"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add [skills or packs...]",
	Short:   "Add skills or packs to the current project",
	GroupID: groupProject,
	Long:    "Copies skills or packs from your library into the project's .agents/skills/ directory. Use --all to add every skill in your library. Run without arguments for interactive selection.",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		lib, err := library.Open()
		if err != nil {
			return err
		}

		var skills []string
		if all {
			skills, err = lib.ListSkills()
			if err != nil {
				return err
			}
			if len(skills) == 0 {
				fmt.Println("No skills in library.")
				return nil
			}
		} else if len(args) == 0 {
			skills, err = addInteractive(lib)
			if err != nil {
				return err
			}
			if len(skills) == 0 {
				return nil
			}
		} else {
			for _, arg := range args {
				resolved, err := lib.ResolveSkillOrPack(arg)
				if err != nil {
					return err
				}
				skills = append(skills, resolved...)
			}
		}

		for _, name := range skills {
			if err := project.AddSkill(lib, name); err != nil {
				return fmt.Errorf("adding %s: %w", name, err)
			}
			fmt.Printf("  + %s\n", name)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().Bool("all", false, "Add all skills from the library")
}

type addItem struct {
	name   string
	isPack bool
	count  int // number of skills in pack
}

func addInteractive(lib *library.Library) ([]string, error) {
	available, err := lib.ListSkills()
	if err != nil {
		return nil, err
	}

	if len(available) == 0 && len(lib.Config.Packs) == 0 {
		fmt.Println("No skills or packs in library.")
		return nil, nil
	}

	var packNames []string
	for name := range lib.Config.Packs {
		packNames = append(packNames, name)
	}
	sort.Strings(packNames)

	var items []addItem
	for _, name := range packNames {
		items = append(items, addItem{name: name, isPack: true, count: len(lib.Config.Packs[name])})
	}
	for _, name := range available {
		items = append(items, addItem{name: name})
	}

	m := addModel{
		items:    items,
		packs:    lib.Config.Packs,
		selected: make(map[int]bool),
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := result.(addModel)
	if final.cancelled {
		fmt.Println("Cancelled.")
		return nil, nil
	}

	seen := make(map[string]bool)
	var chosen []string
	for i, item := range final.items {
		if !final.selected[i] {
			continue
		}
		if item.isPack {
			for _, s := range final.packs[item.name] {
				if !seen[s] {
					seen[s] = true
					chosen = append(chosen, s)
				}
			}
		} else {
			if !seen[item.name] {
				seen[item.name] = true
				chosen = append(chosen, item.name)
			}
		}
	}

	return chosen, nil
}

type addModel struct {
	items     []addItem
	packs     map[string][]string
	selected  map[int]bool
	cursor    int
	cancelled bool
}

func (m addModel) Init() tea.Cmd {
	return nil
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
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

func (m addModel) View() string {
	s := "Select skills to add (space to toggle, enter to confirm):\n\n"

	hasPacks := false
	hasSkills := false
	for _, item := range m.items {
		if item.isPack {
			hasPacks = true
		} else {
			hasSkills = true
		}
	}

	if hasPacks {
		s += "Packs:\n"
		for i, item := range m.items {
			if !item.isPack {
				continue
			}
			s += formatItem(i, item, m.cursor, m.selected)
		}
		if hasSkills {
			s += "\nSkills:\n"
		}
	}

	for i, item := range m.items {
		if item.isPack {
			continue
		}
		s += formatItem(i, item, m.cursor, m.selected)
	}

	s += "\nq/esc to cancel"
	return s
}

func formatItem(index int, item addItem, cursor int, selected map[int]bool) string {
	c := "  "
	if cursor == index {
		c = "> "
	}

	check := "[ ]"
	if selected[index] {
		check = "[x]"
	}

	label := item.name
	if item.isPack {
		label = fmt.Sprintf("%s (%d skills)", item.name, item.count)
	}

	return fmt.Sprintf("%s%s %s\n", c, check, label)
}
