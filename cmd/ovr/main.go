package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/akhenakh/ovr/action"
)

type model struct {
	r      *action.ActionRegistry // items on the to-do list
	cursor int                    // which to-do list item our cursor is pointing at
	choice *action.Action         // which to-do items are selected
	out    string
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			// Send the choice on the channel and exit.
			stdin, _ := io.ReadAll(os.Stdin)
			out, _ := m.r.ActionsForText("")[m.cursor].TextTransform(stdin)
			m.out = string(out)
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.r.ActionsForText("")) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.r.ActionsForText("")) - 1
			}
		}
	}

	return m, nil
}

func initialModel() model {
	return model{
		r: action.NewRegistry(),
	}
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString("Text filter to apply?\n\n")

	for i := 0; i < len(m.r.ActionsForText("")); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(m.r.ActionsForText("")[i].Names[0])
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n")

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	if m, ok := m.(model); ok && m.out != "" {
		fmt.Printf("\n---\n%s\n", m.out)
	}
}
