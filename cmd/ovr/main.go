package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/akhenakh/ovr/action"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
	errorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#FF1111", Dark: "#FF1111"}).
				Render
)

type listKeyMap struct {
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type model struct {
	r            *action.ActionRegistry // items on the to-do list
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
	in           []byte
	out          []byte
}

func newModel(in []byte) model {
	var (
		r            = action.NewRegistry()
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	// Make initial list of items
	actions := r.ActionsForText("")
	items := make([]list.Item, len(actions))
	for i := 0; i < len(actions); i++ {
		items[i] = actions[i]
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	actionList := list.New(items, delegate, 0, 0)
	actionList.Title = fmt.Sprintf("Text Input: %s", strings.TrimRight(string(in), "\r\n"))
	actionList.Styles.Title = titleStyle
	actionList.SetShowStatusBar(false)
	actionList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
		}
	}

	return model{
		r:            r,
		list:         actionList,
		keys:         listKeys,
		delegateKeys: delegateKeys,
		in:           in,
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {

		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil
		case msg.String() == "enter":
			a, ok := m.list.SelectedItem().(*action.Action)
			if ok {
				out, err := a.TextTransform(m.in)
				if err != nil {
					m.list.NewStatusMessage(errorMessageStyle("Error " + err.Error()))
					return m, nil
				}
				m.list.Title = fmt.Sprintf("Text Input: %s", strings.TrimRight(string(out), "\r\n"))

				m.in = out
			}
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}

func main() {
	stdin, _ := io.ReadAll(os.Stdin)

	p := tea.NewProgram(newModel(stdin))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	if m, ok := m.(model); ok {
		fmt.Printf("\n---\n%s\n", string(m.in))
	}
}
