package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
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

		// Group entries into rows: standalone skills and packs
		var rows []libraryRow
		var currentPack string
		var packSkills []string

		flushPack := func() {
			if currentPack != "" {
				rows = append(rows, libraryRow{name: currentPack, isPack: true, skills: packSkills})
				currentPack = ""
				packSkills = nil
			}
		}

		for _, e := range entries {
			if e.Pack == "" {
				flushPack()
				rows = append(rows, libraryRow{name: e.Name})
			} else {
				if e.Pack != currentPack {
					flushPack()
					currentPack = e.Pack
				}
				packSkills = append(packSkills, e.Name)
			}
		}
		flushPack()

		m := libraryModel{rows: rows}
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

type libraryRow struct {
	name     string
	isPack   bool
	skills   []string // only for packs
	expanded bool
}

type libraryModel struct {
	rows   []libraryRow
	cursor int
	height int
	offset int
}

func (m libraryModel) Init() tea.Cmd {
	return nil
}

func (m libraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.clampOffset()
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.clampOffset()
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
				m.clampOffset()
			}
		case "enter", " ":
			if m.rows[m.cursor].isPack {
				m.rows[m.cursor].expanded = !m.rows[m.cursor].expanded
				m.clampOffset()
			}
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// cursorLine returns the line index of the current cursor row.
func (m libraryModel) cursorLine() int {
	line := 0
	for i, r := range m.rows {
		if i == m.cursor {
			return line
		}
		line++
		if r.isPack && r.expanded {
			line += len(r.skills)
		}
	}
	return line
}

func (m *libraryModel) clampOffset() {
	available := m.height - 3 // header + blank + footer
	if available < 1 {
		available = 1
	}
	cl := m.cursorLine()
	if cl < m.offset {
		m.offset = cl
	}
	if cl >= m.offset+available {
		m.offset = cl - available + 1
	}
}

var (
	stylePack      = lipgloss.NewStyle().Bold(true)
	stylePackCount = lipgloss.NewStyle().Faint(true)
	styleSkill     = lipgloss.NewStyle()
	styleCursor    = lipgloss.NewStyle().Bold(true)
	styleNested    = lipgloss.NewStyle().Faint(true)
)

func (m libraryModel) View() string {
	available := m.height - 3
	if available < 1 {
		available = 1
	}

	var lines []string
	for i, r := range m.rows {
		cursor := "  "
		if i == m.cursor {
			cursor = styleCursor.Render("> ")
		}

		if r.isPack {
			arrow := ">"
			if r.expanded {
				arrow = "v"
			}
			line := fmt.Sprintf("%s%s %s %s",
				cursor,
				stylePackCount.Render(arrow),
				stylePack.Render(r.name),
				stylePackCount.Render(fmt.Sprintf("(%d skills)", len(r.skills))),
			)
			lines = append(lines, line)

			if r.expanded {
				for _, s := range r.skills {
					lines = append(lines, fmt.Sprintf("      %s", styleNested.Render(s)))
				}
			}
		} else {
			lines = append(lines, fmt.Sprintf("%s%s", cursor, styleSkill.Render(r.name)))
		}
	}

	// Apply viewport
	end := m.offset + available
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[m.offset:end]

	var s strings.Builder
	s.WriteString("Library\n\n")
	s.WriteString(strings.Join(visible, "\n"))
	s.WriteString("\n\nq: quit  enter: expand/collapse")
	return s.String()
}
