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

	skills, packs := buildSkillsAndPacks(entries)

	installed, _ := project.InstalledSet()
	if installed == nil {
		installed = make(map[string]bool)
	}

	// Default to packs tab if there are no standalone skills
	startTab := tabSkills
	if len(skills) == 0 {
		startTab = tabPacks
	}

	m := libraryModel{
		skills:    skills,
		packs:     packs,
		tab:       startTab,
		lib:       lib,
		installed: installed,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func buildSkillsAndPacks(entries []skill.SkillEntry) ([]string, []libraryPack) {
	var skills []string
	var packs []libraryPack
	var currentPack string
	var packSkills []string

	flushPack := func() {
		if currentPack != "" {
			packs = append(packs, libraryPack{name: currentPack, skills: packSkills})
			currentPack = ""
			packSkills = nil
		}
	}

	for _, e := range entries {
		if e.Pack == "" {
			flushPack()
			skills = append(skills, e.Name)
		} else {
			if e.Pack != currentPack {
				flushPack()
				currentPack = e.Pack
			}
			packSkills = append(packSkills, e.Name)
		}
	}
	flushPack()
	return skills, packs
}

type tab int

const (
	tabSkills tab = iota
	tabPacks
)

type libraryPack struct {
	name     string
	skills   []string
	expanded bool
}

// visibleItem is a flattened entry the cursor can land on in the packs tab.
type visibleItem struct {
	name    string
	isPack  bool
	packIdx int // index into m.packs
}

type libraryModel struct {
	skills []string
	packs  []libraryPack

	tab tab

	skillsCursor int
	skillsOffset int
	packsCursor  int
	packsOffset  int

	height    int
	lib       *library.Library
	installed map[string]bool
	status    string
	statusErr bool
}

func (m libraryModel) packVisibleItems() []visibleItem {
	var items []visibleItem
	for i, p := range m.packs {
		items = append(items, visibleItem{name: p.name, isPack: true, packIdx: i})
		if p.expanded {
			for _, s := range p.skills {
				items = append(items, visibleItem{name: s, packIdx: i})
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

		switch msg.String() {
		case "left", "right", "tab":
			if m.tab == tabSkills && len(m.packs) > 0 {
				m.tab = tabPacks
			} else if m.tab == tabPacks && len(m.skills) > 0 {
				m.tab = tabSkills
			}
			m.clampOffset()
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "enter":
			if m.tab == tabPacks {
				items := m.packVisibleItems()
				if len(items) > 0 {
					item := items[m.packsCursor]
					if item.isPack {
						m.packs[item.packIdx].expanded = !m.packs[item.packIdx].expanded
						if !m.packs[item.packIdx].expanded {
							m.packsCursor = m.packIndexOf(item.packIdx)
						}
						m.clampOffset()
					} else {
						m.toggleCurrent()
						m.clampOffset()
					}
				}
			} else {
				m.toggleCurrent()
				m.clampOffset()
			}
		case " ":
			m.toggleCurrent()
			m.clampOffset()
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *libraryModel) moveCursor(dir int) {
	if m.tab == tabSkills {
		next := m.skillsCursor + dir
		if next >= 0 && next < len(m.skills) {
			m.skillsCursor = next
			m.clampOffset()
		}
	} else {
		items := m.packVisibleItems()
		next := m.packsCursor + dir
		if next >= 0 && next < len(items) {
			m.packsCursor = next
			m.clampOffset()
		}
	}
}

// packIndexOf returns the visible item index for a given pack index.
func (m libraryModel) packIndexOf(packIdx int) int {
	idx := 0
	for i, p := range m.packs {
		if i == packIdx {
			return idx
		}
		idx++
		if p.expanded {
			idx += len(p.skills)
		}
	}
	return idx
}

func (m *libraryModel) toggleCurrent() {
	if m.tab == tabSkills {
		if len(m.skills) == 0 {
			return
		}
		m.toggleSkill(m.skills[m.skillsCursor])
	} else {
		items := m.packVisibleItems()
		if len(items) == 0 {
			return
		}
		item := items[m.packsCursor]

		if item.isPack {
			pack := m.packs[item.packIdx]
			if m.isPackFullyInstalled(pack) {
				for _, skillName := range pack.skills {
					if err := project.RemoveSkill(skillName); err != nil {
						m.status = fmt.Sprintf("Error removing %s: %s", skillName, err)
						m.statusErr = true
						return
					}
					delete(m.installed, skillName)
				}
				m.status = fmt.Sprintf("Removed %d %s from %s", len(pack.skills), skillNoun(len(pack.skills)), item.name)
			} else {
				var added int
				for _, skillName := range pack.skills {
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
			m.toggleSkill(item.name)
		}
	}
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

func (m libraryModel) isPackFullyInstalled(pack libraryPack) bool {
	for _, s := range pack.skills {
		if !m.installed[s] {
			return false
		}
	}
	return true
}

func (m *libraryModel) clampOffset() {
	available := m.viewHeight()
	if m.tab == tabSkills {
		if m.skillsCursor < m.skillsOffset {
			m.skillsOffset = m.skillsCursor
		}
		if m.skillsCursor >= m.skillsOffset+available {
			m.skillsOffset = m.skillsCursor - available + 1
		}
	} else {
		if m.packsCursor < m.packsOffset {
			m.packsOffset = m.packsCursor
		}
		if m.packsCursor >= m.packsOffset+available {
			m.packsOffset = m.packsCursor - available + 1
		}
	}
}

func (m libraryModel) viewHeight() int {
	// tab header + separator + blank + blank + status/blank + footer
	available := m.height - 6
	if available < 1 {
		available = 1
	}
	return available
}

// contextualAction returns the action label based on cursor state.
func (m libraryModel) contextualAction() string {
	if m.tab == tabSkills {
		if len(m.skills) == 0 {
			return "add"
		}
		if m.installed[m.skills[m.skillsCursor]] {
			return "remove"
		}
		return "add"
	}

	items := m.packVisibleItems()
	if len(items) == 0 {
		return "add"
	}
	item := items[m.packsCursor]
	if item.isPack {
		if m.isPackFullyInstalled(m.packs[item.packIdx]) {
			return "remove"
		}
		return "add"
	}
	if m.installed[item.name] {
		return "remove"
	}
	return "add"
}

var (
	stylePack      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	stylePackCount = lipgloss.NewStyle().Faint(true)
	styleSkill     = lipgloss.NewStyle()
	styleCursor    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleInstalled = lipgloss.NewStyle().Faint(true)
	styleCheck     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleStatus    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleStatusErr = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleFooterKey = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleFooter    = lipgloss.NewStyle().Faint(true)
	styleActiveTab = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleTab       = lipgloss.NewStyle().Faint(true)
	styleSeparator = lipgloss.NewStyle().Faint(true)
)

func (m libraryModel) packCountInfo(pack libraryPack) string {
	total := len(pack.skills)
	var count int
	for _, s := range pack.skills {
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

	var s strings.Builder

	// Tab header
	skillsLabel := "Skills"
	packsLabel := "Packs"
	if m.tab == tabSkills {
		skillsLabel = styleActiveTab.Render(skillsLabel)
		packsLabel = styleTab.Render(packsLabel)
	} else {
		skillsLabel = styleTab.Render(skillsLabel)
		packsLabel = styleActiveTab.Render(packsLabel)
	}
	fmt.Fprintf(&s, "  %s    %s\n", skillsLabel, packsLabel)
	s.WriteString(styleSeparator.Render("  ────────────────"))
	s.WriteString("\n\n")

	// Content
	var lines []string
	var offset int
	if m.tab == tabSkills {
		lines = m.renderSkills()
		offset = m.skillsOffset
	} else {
		lines = m.renderPacks()
		offset = m.packsOffset
	}

	end := offset + available
	if end > len(lines) {
		end = len(lines)
	}
	start := offset
	if start > len(lines) {
		start = len(lines)
	}
	visibleLines := lines[start:end]
	s.WriteString(strings.Join(visibleLines, "\n"))

	// Pad remaining space so footer stays at the bottom
	padding := available - len(visibleLines)
	for range padding {
		s.WriteString("\n")
	}

	// Status + footer
	s.WriteString("\n")
	if m.status != "" {
		st := styleStatus
		if m.statusErr {
			st = styleStatusErr
		}
		s.WriteString(st.Render(m.status))
	}
	s.WriteString("\n")

	s.WriteString(m.renderFooter())

	return s.String()
}

func (m libraryModel) renderSkills() []string {
	var lines []string
	for i, name := range m.skills {
		cursor := "  "
		if i == m.skillsCursor {
			cursor = styleCursor.Render("> ")
		}
		check := "  "
		nameStyle := styleSkill
		if m.installed[name] {
			check = styleCheck.Render("✓ ")
			nameStyle = styleInstalled
		}
		lines = append(lines, fmt.Sprintf("%s%s%s", cursor, check, nameStyle.Render(name)))
	}
	return lines
}

func (m libraryModel) renderPacks() []string {
	items := m.packVisibleItems()
	var lines []string
	for i, item := range items {
		if item.isPack {
			cursor := "  "
			if i == m.packsCursor {
				cursor = styleCursor.Render("> ")
			}
			pack := m.packs[item.packIdx]
			arrow := "▶"
			if pack.expanded {
				arrow = "▼"
			}
			line := fmt.Sprintf("%s%s %s %s",
				cursor,
				stylePackCount.Render(arrow),
				stylePack.Render(item.name),
				stylePackCount.Render(m.packCountInfo(pack)),
			)
			lines = append(lines, line)
		} else {
			cursor := "    "
			if i == m.packsCursor {
				cursor = styleCursor.Render("  > ")
			}
			check := "  "
			nameStyle := styleSkill
			if m.installed[item.name] {
				check = styleCheck.Render("✓ ")
				nameStyle = styleInstalled
			}
			lines = append(lines, fmt.Sprintf("%s%s%s", cursor, check, nameStyle.Render(item.name)))
		}
	}
	return lines
}

func footerItem(key, desc string) string {
	return styleFooterKey.Render(key) + styleFooter.Render(" "+desc)
}

func (m libraryModel) renderFooter() string {
	sep := styleFooter.Render("  ")
	parts := []string{
		footerItem("q", "quit"),
		footerItem("tab", "switch"),
	}
	if m.tab == tabPacks {
		parts = append(parts, footerItem("enter", "expand"))
	}
	parts = append(parts, footerItem("space", m.contextualAction()))
	return strings.Join(parts, sep)
}
