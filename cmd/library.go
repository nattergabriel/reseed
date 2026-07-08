package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nattergabriel/reseed/internal/library"
	"github.com/nattergabriel/reseed/internal/project"
	"github.com/nattergabriel/reseed/internal/skill"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

func runLibrary(cmd *cobra.Command, args []string) error {
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

	installed, err := project.InstalledSet()
	if err != nil {
		return err
	}

	m := libraryModel{
		items:     buildItems(entries),
		lib:       lib,
		installed: installed,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// libraryItem is one top-level row: a skill, or a folder with its member skills.
type libraryItem struct {
	name     string
	skills   []string // nil for a plain skill
	expanded bool
}

func (it libraryItem) isFolder() bool {
	return it.skills != nil
}

// buildItems converts entries into a single alphabetical list of top-level
// skills and folders. Entries arrive sorted by group, so folder members can
// be collected by adjacency.
func buildItems(entries []skill.SkillEntry) []libraryItem {
	var items []libraryItem
	for _, e := range entries {
		if e.Group == "" {
			items = append(items, libraryItem{name: e.Name})
			continue
		}
		if len(items) == 0 || items[len(items)-1].name != e.Group {
			items = append(items, libraryItem{name: e.Group, skills: []string{}})
		}
		last := &items[len(items)-1]
		last.skills = append(last.skills, e.Name)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].name < items[j].name })
	return items
}

// visibleItem is a flattened row the cursor can land on.
type visibleItem struct {
	itemIdx  int
	childIdx int // index into the folder's skills; -1 for the item row itself
}

type libraryModel struct {
	items []libraryItem

	cursor int
	offset int

	height    int
	lib       *library.Library
	installed map[string]bool
	status    string
	statusErr bool
}

func (m libraryModel) visibleItems() []visibleItem {
	var rows []visibleItem
	for i, it := range m.items {
		rows = append(rows, visibleItem{itemIdx: i, childIdx: -1})
		if it.isFolder() && it.expanded {
			for c := range it.skills {
				rows = append(rows, visibleItem{itemIdx: i, childIdx: c})
			}
		}
	}
	return rows
}

// skillName returns the skill a row refers to, or "" if the row is a folder.
func (m libraryModel) skillName(row visibleItem) string {
	it := m.items[row.itemIdx]
	if row.childIdx >= 0 {
		return it.skills[row.childIdx]
	}
	if it.isFolder() {
		return ""
	}
	return it.name
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

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.clampOffset()
			}
		case "down", "j":
			if m.cursor < len(m.visibleItems())-1 {
				m.cursor++
				m.clampOffset()
			}
		case "enter":
			m.toggleFold()
			m.clampOffset()
		case " ":
			m.toggleCurrent()
			m.clampOffset()
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// toggleFold expands or collapses the folder under the cursor. On a skill
// inside an expanded folder it collapses that folder and moves the cursor to
// the folder row. On a plain skill it does nothing.
func (m *libraryModel) toggleFold() {
	rows := m.visibleItems()
	if len(rows) == 0 {
		return
	}
	row := rows[m.cursor]
	it := &m.items[row.itemIdx]
	if !it.isFolder() {
		return
	}
	it.expanded = !it.expanded

	// Land on the folder row (a no-op when already there).
	for i, r := range m.visibleItems() {
		if r.itemIdx == row.itemIdx && r.childIdx == -1 {
			m.cursor = i
			return
		}
	}
}

func (m *libraryModel) toggleCurrent() {
	rows := m.visibleItems()
	if len(rows) == 0 {
		return
	}
	row := rows[m.cursor]
	if name := m.skillName(row); name != "" {
		m.toggleSkill(name)
	} else {
		m.toggleFolder(row.itemIdx)
	}
}

func (m *libraryModel) toggleFolder(idx int) {
	folder := m.items[idx]
	if m.isFolderFullyInstalled(folder) {
		for _, name := range folder.skills {
			if err := project.RemoveSkill(name); err != nil {
				m.status = fmt.Sprintf("Error removing %s: %s", name, err)
				m.statusErr = true
				return
			}
			delete(m.installed, name)
		}
		m.status = fmt.Sprintf("Removed %d %s from %s", len(folder.skills), skillNoun(len(folder.skills)), folder.name)
		return
	}

	var added int
	for _, name := range folder.skills {
		if m.installed[name] {
			continue
		}
		if err := project.AddSkill(m.lib, name); err != nil {
			m.status = fmt.Sprintf("Error adding %s: %s", name, err)
			m.statusErr = true
			return
		}
		m.installed[name] = true
		added++
	}
	m.status = fmt.Sprintf("Added %d %s from %s", added, skillNoun(added), folder.name)
}

func (m *libraryModel) toggleSkill(name string) {
	if m.installed[name] {
		if err := project.RemoveSkill(name); err != nil {
			m.status = err.Error()
			m.statusErr = true
			return
		}
		delete(m.installed, name)
		m.status = fmt.Sprintf("Removed %s", name)
	} else {
		if err := project.AddSkill(m.lib, name); err != nil {
			m.status = err.Error()
			m.statusErr = true
			return
		}
		m.installed[name] = true
		m.status = fmt.Sprintf("Added %s", name)
	}
}

func (m libraryModel) isFolderFullyInstalled(folder libraryItem) bool {
	for _, s := range folder.skills {
		if !m.installed[s] {
			return false
		}
	}
	return true
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

func (m libraryModel) availableHeight(header, footer string) int {
	chrome := lipgloss.Height(header) + lipgloss.Height(footer) + 2
	if available := m.height - chrome; available > 1 {
		return available
	}
	return 1
}

func (m libraryModel) viewHeight() int {
	return m.availableHeight(m.renderHeader(), m.renderFooter())
}

func (m libraryModel) contextualAction() string {
	rows := m.visibleItems()
	if len(rows) == 0 {
		return "add"
	}
	row := rows[m.cursor]
	if name := m.skillName(row); name != "" {
		if m.installed[name] {
			return "remove"
		}
		return "add"
	}
	if m.isFolderFullyInstalled(m.items[row.itemIdx]) {
		return "remove"
	}
	return "add"
}

var (
	styleFolder      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	styleFolderCount = lipgloss.NewStyle().Faint(true)
	styleSkill       = lipgloss.NewStyle()
	styleCursor      = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleInstalled   = lipgloss.NewStyle().Faint(true)
	styleCheck       = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleStatus      = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleStatusErr   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleFooterKey   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleFooter      = lipgloss.NewStyle().Faint(true)
	styleTitle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleSeparator   = lipgloss.NewStyle().Faint(true)
)

func (m libraryModel) folderCountInfo(folder libraryItem) string {
	total := len(folder.skills)
	var count int
	for _, s := range folder.skills {
		if m.installed[s] {
			count++
		}
	}
	if count == 0 {
		return fmt.Sprintf("(%d %s)", total, skillNoun(total))
	}
	return fmt.Sprintf("(%d/%d added)", count, total)
}

func (m libraryModel) renderHeader() string {
	return fmt.Sprintf("  %s\n%s", styleTitle.Render("Library"), styleSeparator.Render("  ────────────────"))
}

func (m libraryModel) View() string {
	header := m.renderHeader()
	footer := m.renderFooter()
	available := m.availableHeight(header, footer)

	lines := m.renderItems()
	start := m.offset
	if start > len(lines) {
		start = len(lines)
	}
	end := start + available
	if end > len(lines) {
		end = len(lines)
	}

	content := strings.Join(lines[start:end], "\n")
	view := lipgloss.JoinVertical(lipgloss.Left, header, "", content, "", footer)
	return lipgloss.NewStyle().Height(m.height).Render(view)
}

func (m libraryModel) renderItems() []string {
	rows := m.visibleItems()
	lines := make([]string, len(rows))
	for i, row := range rows {
		lines[i] = m.renderRow(row, i == m.cursor)
	}
	return lines
}

func (m libraryModel) renderRow(row visibleItem, selected bool) string {
	it := m.items[row.itemIdx]

	if it.isFolder() && row.childIdx == -1 {
		cursor := "  "
		if selected {
			cursor = styleCursor.Render("> ")
		}
		arrow := "▶"
		if it.expanded {
			arrow = "▼"
		}
		return fmt.Sprintf("%s%s %s %s",
			cursor,
			styleFolderCount.Render(arrow),
			styleFolder.Render(it.name),
			styleFolderCount.Render(m.folderCountInfo(it)),
		)
	}

	name := m.skillName(row)
	indent := ""
	if row.childIdx >= 0 {
		indent = "  " // skills inside an expanded folder
	}
	cursor := "  "
	if selected {
		cursor = styleCursor.Render("> ")
	}
	check := "  "
	nameStyle := styleSkill
	if m.installed[name] {
		check = styleCheck.Render("✓ ")
		nameStyle = styleInstalled
	}
	return fmt.Sprintf("%s%s%s%s", indent, cursor, check, nameStyle.Render(name))
}

func footerItem(key, desc string) string {
	return styleFooterKey.Render(key) + styleFooter.Render(" "+desc)
}

func (m libraryModel) renderFooter() string {
	sep := styleFooter.Render("  ")
	parts := []string{footerItem("esc", "quit")}

	if rows := m.visibleItems(); len(rows) > 0 {
		row := rows[m.cursor]
		it := m.items[row.itemIdx]
		switch {
		case it.isFolder() && row.childIdx == -1 && it.expanded:
			parts = append(parts, footerItem("enter", "collapse"))
		case it.isFolder() && row.childIdx == -1:
			parts = append(parts, footerItem("enter", "expand"))
		case row.childIdx >= 0:
			parts = append(parts, footerItem("enter", "collapse"))
		}
	}

	parts = append(parts, footerItem("space", m.contextualAction()))
	footer := strings.Join(parts, sep)

	if m.status != "" {
		st := styleStatus
		if m.statusErr {
			st = styleStatusErr
		}
		return st.Render(m.status) + "\n" + footer
	}
	return footer
}
