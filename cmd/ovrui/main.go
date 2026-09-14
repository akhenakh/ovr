package main

import (
	"bytes"
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

	loadConfig()

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

	if input, ok := initialInput(); ok {
		setInput(input)
	} else {
		// no usable input: start with an empty document and let the user
		// paste input in the popup
		setInput(nil)
		appData.pasteOpen = true
	}

	app.SetupWindow("ovr", 1100, 760)
	app.Run(RootView)
}

// initialInput resolves the startup input. It reports ok=false when no
// usable source exists so the caller can open the paste popup.
func initialInput() ([]byte, bool) {
	// 1. file (-f or the legacy --input form)
	if len(os.Args) >= 3 && (os.Args[1] == "-f" || os.Args[1] == "--input") {
		b, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Println("read input file failed:", err)
			os.Exit(1)
		}
		return b, true
	}

	if onWeb {
		// the browser has no file paths or stdin; the paste popup is the
		// only input source
		return nil, false
	}

	// 2. piped stdin
	if stdinPiped() {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("read stdin failed:", err)
			os.Exit(1)
		}
		return b, true
	}

	// 3. clipboard
	initClipboard()
	if clipboardReady {
		if b := clipboard.Read(clipboard.FmtText); len(bytes.TrimSpace(b)) > 0 {
			return b, true
		}
	}

	return nil, false
}

// stdinPiped reports whether stdin carries piped data (not a terminal).
func stdinPiped() bool {
	stat, err := os.Stdin.Stat()
	return err == nil && stat.Mode()&os.ModeCharDevice == 0
}

var clipboardReady bool

func initClipboard() {
	clipboardReady = clipboard.Init() == nil
}

func reloadClipboard() {
	if !clipboardReady {
		setStatus("clipboard not available", true)
		return
	}
	reloadInput(clipboard.Read(clipboard.FmtText))
}
