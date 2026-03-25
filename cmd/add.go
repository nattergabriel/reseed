package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
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
	count  int
}

func (i addItem) FilterValue() string {
	if i.isPack {
		return fmt.Sprintf("%s (%d skills)", i.name, i.count)
	}
	return i.name
}

func addInteractive(lib *library.Library) ([]string, error) {
	entries, err := lib.ListSkillEntries()
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		fmt.Println("No skills or packs in library.")
		return nil, nil
	}

	// Derive packs and standalone skills from entries
	var packs []string
	packCounts := make(map[string]int)
	selected := make(map[int]bool)
	var items []list.Item
	for _, e := range entries {
		if e.Pack != "" {
			if packCounts[e.Pack] == 0 {
				packs = append(packs, e.Pack)
			}
			packCounts[e.Pack]++
		}
	}
	for _, p := range packs {
		items = append(items, addItem{name: p, isPack: true, count: packCounts[p]})
	}
	for _, e := range entries {
		if e.Pack == "" {
			items = append(items, addItem{name: e.Name})
		}
	}

	delegate := checkboxDelegate{selected: selected}
	l := list.New(items, delegate, 0, 0)
	l.Title = "Select skills or packs to add (space: toggle, enter: confirm)"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	m := checkboxModel{
		list:     l,
		selected: selected,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := result.(checkboxModel)
	if final.cancelled {
		fmt.Println("Cancelled.")
		return nil, nil
	}

	// Expand selected packs into skill names
	seen := make(map[string]bool)
	var chosen []string
	for i, listItem := range final.list.Items() {
		if !final.selected[i] {
			continue
		}
		item := listItem.(addItem)
		if item.isPack {
			for _, e := range entries {
				if e.Pack == item.name && !seen[e.Name] {
					seen[e.Name] = true
					chosen = append(chosen, e.Name)
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
