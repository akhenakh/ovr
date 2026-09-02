package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/akhenakh/ovr/internal/clipboard"
)

func main() {
	readStdin := flag.Bool("s", false, "Use Stdin as input data (conflicts with TUI interaction if piped directly)")
	inputFile := flag.String("f", "", "Input filename (read data from file)")
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
	// Only fail on clipboard init error if we absolutely need it (no input file/stdin provided)
	if err != nil && *inputFile == "" && !*readStdin {
		fmt.Fprintf(os.Stderr, "Clipboard unavailable: %v\nUse -s to read from stdin or -f <filename> to read from a file.\n", err)
		os.Exit(1)
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
	} else if clipboard.Available() {
		input = clipboard.Read(clipboard.FmtText)
	}

	p := tea.NewProgram(newModel(input))

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
