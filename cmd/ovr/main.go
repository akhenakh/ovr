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
	readStdin := flag.Bool("s", false, "Use Stdin as input data (conflicts with TUI interaction if piped directly)")
	inputFile := flag.String("i", "", "Input file path (read data from file)")
	rawOutput := flag.Bool("r", false, "Raw output, only the modified string")
	outputFile := flag.String("o", "", "Output file path (writes raw output to file)")
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
	// Only panic on clipboard init error if we absolutely need it (no input file/stdin provided)
	if err != nil && *inputFile == "" && !*readStdin {
		panic(err)
	}

	var input []byte

	// Priority: 1. Input File, 2. Stdin, 3. Clipboard
	if *inputFile != "" {
		f, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Printf("Error reading input file: %v\n", err)
			os.Exit(1)
		}
		input = f
	} else if *readStdin {
		stdin, _ := io.ReadAll(os.Stdin)
		input = stdin
	} else {
		input = clipboard.Read(clipboard.FmtText)
	}

	// Important: We strictly use NewProgram. If 'ovr' was invoked with a file input (-i),
	// stdin remains attached to the terminal, allowing the TUI to work.
	p := tea.NewProgram(
		newModel(input),
		tea.WithMouseCellMotion(),
	)

	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	if m, ok := m.(model); ok {
		finalOutput := m.out.String()

		if *outputFile != "" {
			err := os.WriteFile(*outputFile, []byte(finalOutput), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if *rawOutput {
			fmt.Print(finalOutput)
		} else {
			fmt.Printf("%s\n---\n%s\n", m.out.StackString(), finalOutput)
		}

		if *inputFile == "" && !*readStdin {
			clipboard.Write(clipboard.FmtText, []byte(finalOutput))
		}
	}
}
