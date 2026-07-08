package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nattergabriel/reseed/internal/reseed"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

func runTUI(cmd *cobra.Command, args []string) error {
	lib, err := reseed.OpenLibrary()
	if err != nil {
		return err
	}

	if len(lib.Skills) == 0 {
		fmt.Println("No skills in library.")
		return nil
	}

	proj, err := reseed.OpenProject(flagDir)
	if err != nil {
		return err
	}

	names, err := proj.Installed()
	if err != nil {
		return err
	}
	installed := make(map[string]bool, len(names))
	for _, n := range names {
		installed[n] = true
	}

	m := tuiModel{
		items:     buildItems(lib.Skills),
		proj:      proj,
		installed: installed,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// tuiItem is one top-level row: a skill, or a folder with its member skills.
type tuiItem struct {
	name     string
	skill    reseed.Skill   // valid when this row is a plain skill
	skills   []reseed.Skill // folder members; nil for a plain skill
	expanded bool
}

func (it tuiItem) isFolder() bool {
	return it.skills != nil
}

// buildItems converts skills into a single alphabetical list of top-level
// skills and folders. Skills arrive sorted by group, so folder members can be
// collected by adjacency.
func buildItems(skills []reseed.Skill) []tuiItem {
	var items []tuiItem
	for _, s := range skills {
		if s.Group == "" {
			items = append(items, tuiItem{name: s.Name, skill: s})
			continue
		}
		if len(items) == 0 || items[len(items)-1].name != s.Group {
			items = append(items, tuiItem{name: s.Group, skills: []reseed.Skill{}})
		}
		last := &items[len(items)-1]
		last.skills = append(last.skills, s)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].name < items[j].name })
	return items
}

// visibleItem is a flattened row the cursor can land on.
type visibleItem struct {
	itemIdx  int
	childIdx int // index into the folder's skills; -1 for the item row itself
}

type tuiModel struct {
	items []tuiItem

	cursor int
	offset int

	height    int
	proj      reseed.Project
	installed map[string]bool
	status    string
	statusErr bool
}

func (m tuiModel) visibleItems() []visibleItem {
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

// skillAt returns the skill a row refers to; ok is false for folder rows.
func (m tuiModel) skillAt(row visibleItem) (reseed.Skill, bool) {
	it := m.items[row.itemIdx]
	if row.childIdx >= 0 {
		return it.skills[row.childIdx], true
	}
	if it.isFolder() {
		return reseed.Skill{}, false
	}
	return it.skill, true
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *tuiModel) toggleFold() {
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

func (m *tuiModel) toggleCurrent() {
	rows := m.visibleItems()
	if len(rows) == 0 {
		return
	}
	row := rows[m.cursor]
	if s, ok := m.skillAt(row); ok {
		m.toggleSkill(s)
	} else {
		m.toggleFolder(row.itemIdx)
	}
}

func (m *tuiModel) toggleFolder(idx int) {
	folder := m.items[idx]
	if m.isFolderFullyInstalled(folder) {
		for _, s := range folder.skills {
			if err := m.proj.Remove(s.Name); err != nil {
				m.status = fmt.Sprintf("Error removing %s: %s", s.Name, err)
				m.statusErr = true
				return
			}
			delete(m.installed, s.Name)
		}
		m.status = fmt.Sprintf("Removed %d %s from %s", len(folder.skills), skillNoun(len(folder.skills)), folder.name)
		return
	}

	var added int
	for _, s := range folder.skills {
		if m.installed[s.Name] {
			continue
		}
		if err := m.proj.Add(s); err != nil {
			m.status = fmt.Sprintf("Error adding %s: %s", s.Name, err)
			m.statusErr = true
			return
		}
		m.installed[s.Name] = true
		added++
	}
	m.status = fmt.Sprintf("Added %d %s from %s", added, skillNoun(added), folder.name)
}

func (m *tuiModel) toggleSkill(s reseed.Skill) {
	if m.installed[s.Name] {
		if err := m.proj.Remove(s.Name); err != nil {
			m.status = err.Error()
			m.statusErr = true
			return
		}
		delete(m.installed, s.Name)
		m.status = fmt.Sprintf("Removed %s", s.Name)
	} else {
		if err := m.proj.Add(s); err != nil {
			m.status = err.Error()
			m.statusErr = true
			return
		}
		m.installed[s.Name] = true
		m.status = fmt.Sprintf("Added %s", s.Name)
	}
}

func (m tuiModel) isFolderFullyInstalled(folder tuiItem) bool {
	for _, s := range folder.skills {
		if !m.installed[s.Name] {
			return false
		}
	}
	return true
}

func (m *tuiModel) clampOffset() {
	available := m.viewHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+available {
		m.offset = m.cursor - available + 1
	}
}

// viewHeight is the number of content lines that fit between the fixed-size
// header and footer.
func (m tuiModel) viewHeight() int {
	chrome := 2 + 1 + 2 // header, footer, blank separators
	if m.status != "" {
		chrome++
	}
	if available := m.height - chrome; available > 1 {
		return available
	}
	return 1
}

func (m tuiModel) contextualAction() string {
	rows := m.visibleItems()
	if len(rows) == 0 {
		return "add"
	}
	row := rows[m.cursor]
	if s, ok := m.skillAt(row); ok {
		if m.installed[s.Name] {
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

func (m tuiModel) folderCountInfo(folder tuiItem) string {
	total := len(folder.skills)
	var count int
	for _, s := range folder.skills {
		if m.installed[s.Name] {
			count++
		}
	}
	if count == 0 {
		return fmt.Sprintf("(%d %s)", total, skillNoun(total))
	}
	return fmt.Sprintf("(%d/%d added)", count, total)
}

func (m tuiModel) renderHeader() string {
	return fmt.Sprintf("  %s\n%s", styleTitle.Render("Library"), styleSeparator.Render("  ────────────────"))
}

func (m tuiModel) View() string {
	available := m.viewHeight()

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
	view := lipgloss.JoinVertical(lipgloss.Left, m.renderHeader(), "", content, "", m.renderFooter())
	return lipgloss.NewStyle().Height(m.height).Render(view)
}

func (m tuiModel) renderItems() []string {
	rows := m.visibleItems()
	lines := make([]string, len(rows))
	for i, row := range rows {
		lines[i] = m.renderRow(row, i == m.cursor)
	}
	return lines
}

func (m tuiModel) renderRow(row visibleItem, selected bool) string {
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

	s, _ := m.skillAt(row)
	indent := ""
	if row.childIdx >= 0 {
		indent = "  " // skills inside an expanded folder
	}
	cursor := "  "
	if selected {
		cursor = styleCursor.Render("> ")
	}
	check := "  "
	name := s.Name
	if m.installed[s.Name] {
		check = styleCheck.Render("✓ ")
		name = styleInstalled.Render(name)
	}
	return fmt.Sprintf("%s%s%s%s", indent, cursor, check, name)
}

func footerItem(key, desc string) string {
	return styleFooterKey.Render(key) + styleFooter.Render(" "+desc)
}

func (m tuiModel) renderFooter() string {
	sep := styleFooter.Render("  ")
	parts := []string{footerItem("esc", "quit")}

	if rows := m.visibleItems(); len(rows) > 0 {
		row := rows[m.cursor]
		if it := m.items[row.itemIdx]; it.isFolder() {
			hint := "expand"
			if it.expanded {
				hint = "collapse"
			}
			parts = append(parts, footerItem("enter", hint))
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
