package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/nattergabriel/reseed/internal/project"
	"github.com/nattergabriel/reseed/internal/skill"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	rootCmd.AddCommand(libraryCmd)
}

var libraryCmd = &cobra.Command{
	Use:     "library",
	Short:   "Browse and manage your skill library",
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

		rows := buildRows(entries)

		installed := make(map[string]bool)
		if names, err := project.ListInstalled(); err == nil {
			for _, n := range names {
				installed[n] = true
			}
		}

		m := libraryModel{
			rows:      rows,
			lib:       lib,
			installed: installed,
		}
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func buildRows(entries []skill.SkillEntry) []libraryRow {
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
	return rows
}

// libraryRow is either a standalone skill or a pack header with nested skills.
type libraryRow struct {
	name     string
	isPack   bool
	skills   []string
	expanded bool
}

// visibleItem is a flattened entry the cursor can land on.
type visibleItem struct {
	name     string
	isPack   bool
	isNested bool
	packName string
	rowIdx   int
}

type libraryModel struct {
	rows      []libraryRow
	cursor    int
	height    int
	offset    int
	lib       *library.Library
	installed map[string]bool
	status    string
	statusErr bool
}

func (m libraryModel) visibleItems() []visibleItem {
	var items []visibleItem
	for i, r := range m.rows {
		items = append(items, visibleItem{
			name:   r.name,
			isPack: r.isPack,
			rowIdx: i,
		})
		if r.isPack && r.expanded {
			for _, s := range r.skills {
				items = append(items, visibleItem{
					name:     s,
					isNested: true,
					packName: r.name,
					rowIdx:   i,
				})
			}
		}
	}
	return items
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
		m.status = ""
		m.statusErr = false

		visible := m.visibleItems()
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.clampOffset()
			}
		case "down", "j":
			if m.cursor < len(visible)-1 {
				m.cursor++
				m.clampOffset()
			}
		case "enter", " ":
			item := visible[m.cursor]
			if item.isPack {
				m.rows[item.rowIdx].expanded = !m.rows[item.rowIdx].expanded
				// If collapsing, clamp cursor to pack header
				if !m.rows[item.rowIdx].expanded {
					m.cursor = m.indexOfRow(item.rowIdx)
				}
				m.clampOffset()
			}
		case "a":
			m.toggleCurrent()
			m.clampOffset()
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// indexOfRow returns the visible item index for a given row index.
func (m libraryModel) indexOfRow(rowIdx int) int {
	idx := 0
	for i, r := range m.rows {
		if i == rowIdx {
			return idx
		}
		idx++
		if r.isPack && r.expanded {
			idx += len(r.skills)
		}
	}
	return idx
}

func (m *libraryModel) toggleCurrent() {
	visible := m.visibleItems()
	item := visible[m.cursor]

	if item.isPack {
		row := m.rows[item.rowIdx]
		// If all installed, remove all. Otherwise, add all missing.
		allInstalled := true
		for _, s := range row.skills {
			if !m.installed[s] {
				allInstalled = false
				break
			}
		}
		if allInstalled {
			for _, skillName := range row.skills {
				if err := project.RemoveSkill(skillName); err != nil {
					m.status = fmt.Sprintf("Error removing %s: %s", skillName, err)
					m.statusErr = true
					return
				}
				delete(m.installed, skillName)
			}
			m.status = fmt.Sprintf("Removed %d %s from %s", len(row.skills), skillNoun(len(row.skills)), item.name)
		} else {
			var added int
			for _, skillName := range row.skills {
				if m.installed[skillName] {
					continue
				}
				if err := project.AddSkill(m.lib, skillName); err != nil {
					m.status = fmt.Sprintf("Error adding %s: %s", skillName, err)
					m.statusErr = true
					return
				}
				m.installed[skillName] = true
				added++
			}
			m.status = fmt.Sprintf("Added %d %s from %s", added, skillNoun(added), item.name)
		}
	} else {
		skillName := item.name
		if m.installed[skillName] {
			if err := project.RemoveSkill(skillName); err != nil {
				m.status = err.Error()
				m.statusErr = true
				return
			}
			delete(m.installed, skillName)
			m.status = fmt.Sprintf("Removed %s", skillName)
		} else {
			if err := project.AddSkill(m.lib, skillName); err != nil {
				m.status = err.Error()
				m.statusErr = true
				return
			}
			m.installed[skillName] = true
			m.status = fmt.Sprintf("Added %s", skillName)
		}
	}
}

func (m *libraryModel) clampOffset() {
	available := m.viewHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+available {
		m.offset = m.cursor - available + 1
	}
}

func (m libraryModel) viewHeight() int {
	// header + blank line + blank line + status/blank + footer
	available := m.height - 5
	if available < 1 {
		available = 1
	}
	return available
}

var (
	stylePack      = lipgloss.NewStyle().Bold(true)
	stylePackCount = lipgloss.NewStyle().Faint(true)
	styleSkill     = lipgloss.NewStyle()
	styleCursor    = lipgloss.NewStyle().Bold(true)
	styleNested    = lipgloss.NewStyle().Faint(true)
	styleInstalled = lipgloss.NewStyle().Faint(true)
	styleCheck     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleStatus    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleStatusErr = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleFooter    = lipgloss.NewStyle().Faint(true)
)

func (m libraryModel) packCountInfo(row libraryRow) string {
	total := len(row.skills)
	var count int
	for _, s := range row.skills {
		if m.installed[s] {
			count++
		}
	}
	if count == 0 {
		return fmt.Sprintf("(%d %s)", total, skillNoun(total))
	}
	return fmt.Sprintf("(%d/%d added)", count, total)
}

func (m libraryModel) View() string {
	available := m.viewHeight()
	visible := m.visibleItems()

	var lines []string
	for i, item := range visible {
		cursor := "  "
		if i == m.cursor {
			cursor = styleCursor.Render("> ")
		}

		if item.isPack {
			row := m.rows[item.rowIdx]
			arrow := ">"
			if row.expanded {
				arrow = "v"
			}
			line := fmt.Sprintf("%s%s %s %s",
				cursor,
				stylePackCount.Render(arrow),
				stylePack.Render(item.name),
				stylePackCount.Render(m.packCountInfo(row)),
			)
			lines = append(lines, line)
		} else if item.isNested {
			cursor := "    "
			if i == m.cursor {
				cursor = styleCursor.Render("  > ")
			}
			check := "  "
			nameStyle := styleNested
			if m.installed[item.name] {
				check = styleCheck.Render("* ")
				nameStyle = styleInstalled
			}
			lines = append(lines, fmt.Sprintf("%s%s%s", cursor, check, nameStyle.Render(item.name)))
		} else {
			check := "  "
			nameStyle := styleSkill
			if m.installed[item.name] {
				check = styleCheck.Render("* ")
				nameStyle = styleInstalled
			}
			lines = append(lines, fmt.Sprintf("%s%s%s", cursor, check, nameStyle.Render(item.name)))
		}
	}

	// Apply viewport
	end := m.offset + available
	if end > len(lines) {
		end = len(lines)
	}
	start := m.offset
	if start > len(lines) {
		start = len(lines)
	}
	viewLines := lines[start:end]

	var s strings.Builder
	s.WriteString(stylePack.Render("Library"))
	s.WriteString("\n\n")
	s.WriteString(strings.Join(viewLines, "\n"))
	s.WriteString("\n\n")
	if m.status != "" {
		st := styleStatus
		if m.statusErr {
			st = styleStatusErr
		}
		s.WriteString(st.Render(m.status))
	}
	s.WriteString("\n")
	s.WriteString(styleFooter.Render("q: quit  enter: expand/collapse  a: toggle"))

	return s.String()
}
