package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
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
	removeAction     key.Binding
	showDetails      key.Binding
	openEditor       key.Binding
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
		showDetails: key.NewBinding(
			key.WithKeys("v", "V"),
			key.WithHelp("v", "show details view"),
		),
		removeAction: key.NewBinding(
			key.WithKeys("backspace", "d"),
			key.WithHelp("backspace", "undo last action"),
		),
		openEditor: key.NewBinding(
			key.WithKeys("e", "E"),
			key.WithHelp("e", "open editor"),
		),
	}
}

func main() {
	readStdin := flag.Bool("s", false, "Use Stdin as input, default to clipboard")
	debug := flag.Bool("debug", false, "Debug in debug.log file")

	flag.Parse()

	if *debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	var input []byte

	if *readStdin {
		stdin, _ := io.ReadAll(os.Stdin)
		input = stdin
	} else {
		input = clipboard.Read(clipboard.FmtText)
	}

	p := tea.NewProgram(
		newModel(input),
		// tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	if m, ok := m.(model); ok {
		fmt.Printf("%s\n---\n%s\n", m.out.StackString(), m.out.String())

		// putting output in clipboard
		if !*readStdin {
			clipboard.Write(clipboard.FmtText, []byte(m.out.String()))
		}
	}
}
