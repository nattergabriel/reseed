package cmd

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type checkboxDelegate struct {
	selected map[int]bool
}

func (d checkboxDelegate) Height() int                             { return 1 }
func (d checkboxDelegate) Spacing() int                            { return 0 }
func (d checkboxDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d checkboxDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}
	check := "[ ]"
	if d.selected[index] {
		check = "[x]"
	}
	_, _ = fmt.Fprintf(w, "%s%s %s", cursor, check, listItem.FilterValue())
}

type checkboxModel struct {
	list      list.Model
	selected  map[int]bool
	cancelled bool
}

func (m checkboxModel) Init() tea.Cmd {
	return nil
}

func (m checkboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			m.selected[m.list.Index()] = !m.selected[m.list.Index()]
			return m, nil
		case "enter":
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m checkboxModel) View() string {
	return m.list.View()
}
