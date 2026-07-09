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

// currentRow returns the row under the cursor; ok is false when the list is
// empty.
func (m tuiModel) currentRow() (visibleItem, bool) {
	rows := m.visibleItems()
	if len(rows) == 0 {
		return visibleItem{}, false
	}
	return rows[m.cursor], true
}

// rowSkills returns the skills a row acts on: the skill itself, or every
// member of a folder. folder is the folder name, "" for a skill row.
func (m tuiModel) rowSkills(row visibleItem) (skills []reseed.Skill, folder string) {
	if s, ok := m.skillAt(row); ok {
		return []reseed.Skill{s}, ""
	}
	it := m.items[row.itemIdx]
	return it.skills, it.name
}

func (m tuiModel) installedCount(skills []reseed.Skill) int {
	var count int
	for _, s := range skills {
		if m.installed[s.Name] {
			count++
		}
	}
	return count
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
	row, ok := m.currentRow()
	if !ok {
		return
	}
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

// toggleCurrent adds the skills the cursor's row acts on, or removes them
// when every one of them is already installed.
func (m *tuiModel) toggleCurrent() {
	row, ok := m.currentRow()
	if !ok {
		return
	}
	skills, folder := m.rowSkills(row)
	remove := m.installedCount(skills) == len(skills)

	var affected int
	for _, s := range skills {
		if !remove && m.installed[s.Name] {
			continue
		}
		var err error
		verb := "adding"
		if remove {
			verb = "removing"
			err = m.proj.Remove(s.Name)
		} else {
			err = m.proj.Add(s)
		}
		if err != nil {
			m.status = fmt.Sprintf("Error %s %s: %s", verb, s.Name, err)
			m.statusErr = true
			return
		}
		if remove {
			delete(m.installed, s.Name)
		} else {
			m.installed[s.Name] = true
		}
		affected++
	}

	verb := "Added"
	if remove {
		verb = "Removed"
	}
	if folder == "" {
		m.status = fmt.Sprintf("%s %s", verb, skills[0].Name)
		return
	}
	m.status = fmt.Sprintf("%s %d %s from %s", verb, affected, skillNoun(affected), folder)
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
	row, ok := m.currentRow()
	if !ok {
		return "add"
	}
	skills, _ := m.rowSkills(row)
	if m.installedCount(skills) == len(skills) {
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
	count := m.installedCount(folder.skills)
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
	start := min(m.offset, len(lines))
	end := min(start+available, len(lines))

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

	if row, ok := m.currentRow(); ok {
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
