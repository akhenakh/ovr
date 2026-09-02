package main

import (
	"fmt"
	"io"
	"log"
	"os"

	. "go.hasen.dev/shirei"
	app "go.hasen.dev/shirei/app"

	"github.com/akhenakh/ovr/internal/clipboard"
)

func main() {
	// shirei's LargeText logs every scan to the std logger; keep stdout clean
	log.SetOutput(io.Discard)

	if len(os.Args) >= 3 && os.Args[1] == "--png" {
		input := []byte("POINT(-0.4539761 48.0930043)")
		if len(os.Args) >= 4 {
			b, err := os.ReadFile(os.Args[3])
			if err != nil {
				fmt.Println("read input file failed:", err)
				os.Exit(1)
			}
			input = b
		}
		setInput(input)
		selectFirst()
		if err := RenderToPNG(os.Args[2], 1100, 760, RootView); err != nil {
			fmt.Println("render failed:", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) >= 3 && os.Args[1] == "--input" {
		b, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Println("read input file failed:", err)
			os.Exit(1)
		}
		setInput(b)
	} else {
		initClipboard()
		if clipboardReady {
			reloadInput(clipboard.Read(clipboard.FmtText))
		}
	}

	app.SetupWindow("ovr", 1100, 760)
	app.Run(RootView)
}

var clipboardReady bool

func initClipboard() {
	if err := clipboard.Init(); err != nil {
		setInput([]byte("could not read the clipboard: " + err.Error()))
		return
	}
	clipboardReady = true
}

func reloadClipboard() {
	if !clipboardReady {
		setStatus("clipboard not available", true)
		return
	}
	reloadInput(clipboard.Read(clipboard.FmtText))
}
