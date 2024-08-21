package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/akhenakh/ovr/action"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	state         sessionState
	width, height int
	r             *action.ActionRegistry // items for the search list
	list          list.Model
	viewport      viewport.Model // details view
	paramsInput   []textinput.Model
	keys          *listKeyMap
	delegateKeys  *delegateKeyMap
	in            []byte
	out           *action.Data

	// for the text input view
	cursorMode cursor.Mode
	focusIndex int
}

type sessionState int

const (
	mainListState sessionState = iota // the default view with list of applicable actions
	detailState                       // the detail view, that could change based on the type of current data
	paramState                        // displayed when a parameter is needed to execute the action
)

var infoStyle = func() lipgloss.Style {
	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	return titleStyle.Copy().BorderStyle(b)
}()

func newModel(in []byte) model {
	var (
		r            = action.DefaultRegistry()
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	// Make initial list of actions
	actions := r.ActionsForText("")
	items := make([]list.Item, len(actions))
	for i := 0; i < len(actions); i++ {
		items[i] = actions[i]
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	actionList := list.New(items, delegate, 0, 0)
	actionList.Title = fmt.Sprintf("Text: %s", strings.TrimRight(string(in), "\r\n"))
	actionList.Styles.Title = titleStyle
	actionList.SetShowStatusBar(false)
	actionList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
			listKeys.removeAction,
			listKeys.showDetails,
			listKeys.openEditor,
		}
	}

	return model{
		state:        mainListState,
		r:            r,
		list:         actionList,
		keys:         listKeys,
		delegateKeys: delegateKeys,
		in:           in,
		out:          action.NewDataText(in),
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.state == detailState {
		return updateDetail(msg, m)
	}

	if m.state == paramState {
		return updateParams(msg, m)
	}

	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.width = msg.Width - h
		m.height = msg.Height - v

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

		case key.Matches(msg, m.keys.showDetails):
			m.state = detailState
			return m, nil
		case key.Matches(msg, m.keys.openEditor):
			return m, m.openEditor()

		case key.Matches(msg, m.keys.removeAction):
			d, oa, err := m.out.Undo([]byte(m.out.String()))
			if err != nil { // we should not have errors in the stack
				m.list.NewStatusMessage(errorMessageStyle("Error " + err.Error()))
				return m, nil
			}
			m.out = d
			m.list.NewStatusMessage(statusMessageStyle("Removed action: " + oa.Title()))

			m.list.Title = fmt.Sprintf("%s: %s", m.out.Format.Name, strings.TrimRight(m.out.String(), "\r\n"))

			m.list.ResetFilter()

			actions := m.r.ActionsForData(m.out)
			items := make([]list.Item, len(actions))
			for i := 0; i < len(actions); i++ {
				items[i] = actions[i]
			}
			m.list.SetItems(items)

			return m, nil

		case msg.String() == "enter":
			a, ok := m.list.SelectedItem().(*action.Action)
			m.paramsInput = make([]textinput.Model, len(a.Parameters))

			if ok {
				if len(a.Parameters) > 0 {
					for i, ap := range a.Parameters {
						ti := textinput.New()
						switch ap.ActionParameterType {
						case action.IntParameter:
							ti.Placeholder = ap.Doc
							ti.CharLimit = 8
							ti.Width = 20
						case action.StringParameter:
							ti.Placeholder = ap.Doc
							ti.CharLimit = 256
							ti.Width = 80
						case action.FloatParameter:
							ti.Placeholder = ap.Doc
							ti.CharLimit = 16
							ti.Width = 20
						}

						if i == 0 {
							ti.Focus()
						}
						m.paramsInput[i] = ti
					}
					m.state = paramState

					return m, nil
				}

				out, err := a.Transform(m.out)
				if err != nil {
					m.list.NewStatusMessage(errorMessageStyle("Error " + err.Error()))
					return m, nil
				}
				m.list.Title = fmt.Sprintf("%s: %s", out.Format.Name, strings.TrimRight(out.String(), "\r\n"))
				m.out = out

				m.list.ResetFilter()

				actions := m.r.ActionsForData(m.out)
				items := make([]list.Item, len(actions))
				for i := 0; i < len(actions); i++ {
					items[i] = actions[i]
				}
				m.list.SetItems(items)
			}
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func updateParams(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var (
		focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		noStyle      = lipgloss.NewStyle()
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.paramsInput))
			for i := range m.paramsInput {
				cmds[i] = m.paramsInput[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit the view and try apply the values.
			if s == "enter" && m.focusIndex == len(m.paramsInput) {
				a, _ := m.list.SelectedItem().(*action.Action)
				m.state = mainListState

				for i, ap := range a.Parameters {
					switch ap.ActionParameterType {
					case action.IntParameter:
						v, err := strconv.Atoi(m.paramsInput[i].Value())
						if err != nil {
							m.list.NewStatusMessage(errorMessageStyle("Error " + err.Error()))
							return m, nil
						}
						a.InputParameters = append(a.InputParameters, v)
					case action.FloatParameter:
						v, err := strconv.ParseFloat(m.paramsInput[i].Value(), 64)
						if err != nil {
							m.list.NewStatusMessage(errorMessageStyle("Error " + err.Error()))
							return m, nil
						}
						a.InputParameters = append(a.InputParameters, v)
					case action.StringParameter:
						if m.paramsInput[i].Value() == "" {
							m.list.NewStatusMessage(errorMessageStyle("Error string parameter is empty"))
							return m, nil
						}
						a.InputParameters = append(a.InputParameters, m.paramsInput[i].Value())
					}
				}
				return m, nil
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.paramsInput) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.paramsInput)
			}

			cmds := make([]tea.Cmd, len(m.paramsInput))
			for i := 0; i <= len(m.paramsInput)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.paramsInput[i].Focus()
					m.paramsInput[i].PromptStyle = focusedStyle
					m.paramsInput[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.paramsInput[i].Blur()
				m.paramsInput[i].PromptStyle = noStyle
				m.paramsInput[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmds := make([]tea.Cmd, len(m.paramsInput))

	// Only text paramsInput with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.paramsInput {
		m.paramsInput[i], cmds[i] = m.paramsInput[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func updateDetail(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.Width = m.width
		m.viewport.Height = m.height - verticalMarginHeight

	case tea.KeyMsg:
		// come back from detail view
		if k := msg.String(); (k == "ctrl+c" || k == "q" || k == "esc") && m.state == detailState {
			m.state = mainListState
			return m, nil
		}
	}

	_, cmd := m.viewport.Update(msg)
	return m, cmd
}

// headerView for detailView
func (m model) headerView() string {
	title := titleStyle.Render("Text:")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

// footerView for detailView
func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m model) View() string {
	switch m.state {
	case detailState:
		return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	case paramState:
		var (
			focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

			helpStyle           = blurredStyle.Copy()
			cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

			focusedButton = focusedStyle.Copy().Render("[ Submit ]")
			blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
		)

		var b strings.Builder

		a, _ := m.list.SelectedItem().(*action.Action)

		fmt.Fprintf(&b, "Parameters for %s\n\n", a.Names[0])

		for i := range m.paramsInput {
			b.WriteString(m.paramsInput[i].View())
			if i < len(m.paramsInput)-1 {
				b.WriteRune('\n')
			}
		}

		button := &blurredButton
		if m.focusIndex == len(m.paramsInput) {
			button = &focusedButton
		}

		fmt.Fprintf(&b, "\n\n%s\n\n", *button)

		b.WriteString(helpStyle.Render("cursor mode is "))
		b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
		b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

		return b.String()

	default:
		return appStyle.Render(m.list.View())
	}
}

type editorFinishedMsg struct{ err error }

func (m model) openEditor() tea.Cmd {
	f, err := os.CreateTemp(os.TempDir(), "ovr")
	if err != nil {
		log.Println(err)
		// TODO: not showing up
		return m.list.NewStatusMessage(errorMessageStyle("Error " + err.Error()))
	}
	defer f.Close()
	defer os.Remove(f.Name())
	if _, err := f.Write(m.out.RawValue); err != nil {
		log.Println(err)
		// TODO: not showing up
		return m.list.NewStatusMessage(errorMessageStyle("Error " + err.Error()))
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	c := exec.Command(editor, f.Name()) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			log.Println(err)
			// TODO: not showing up
			return m.list.NewStatusMessage(errorMessageStyle("Error " + err.Error()))
		}
		return nil
	})
}
