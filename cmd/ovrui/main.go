package main

import (
	"fmt"
	"os"

	"github.com/AllenDang/giu"
	g "github.com/AllenDang/giu"
)

var (
	editor                *g.CodeEditorWidget
	statusMsg             string
	disableGlobalShortcut bool // disable global shortcut while using editor or /
	visibleWidgets        []g.Widget
	smallFont             *g.FontInfo
)

func onClickMe() {
	fmt.Println("\ueb28 Hello world!")
}

func onImSoCute() {
	fmt.Println("Im sooooooo cute!!")
}

func loop() {
	g.SingleWindow().RegisterKeyboardShortcuts(
		// quit command
		giu.WindowShortcut{Key: giu.KeyQ, Callback: func() {
			if !disableGlobalShortcut {
				os.Exit(0)
			}
		}},
		// search command
		giu.WindowShortcut{Key: giu.KeySlash, Callback: func() {
			if !disableGlobalShortcut {
				os.Exit(0)
			}
		}},
		// view command
		giu.WindowShortcut{Key: giu.KeyV, Callback: func() {
			if !disableGlobalShortcut {
				statusMsg = "ESC quit the editor"

				visibleWidgets = []g.Widget{
					g.Column(
						g.Row(
							g.Label(statusMsg).Font(smallFont),
						),
						editor,
					),
				}
			}
		}},
		// close editor
		giu.WindowShortcut{Key: giu.KeyEscape, Callback: func() {
			if !disableGlobalShortcut {
				os.Exit(0)
			}
		}},
	).Layout(visibleWidgets...)
}

func main() {
	wnd := g.NewMasterWindow("Hello world", 640, 480, g.MasterWindowFlagsFrameless)

	smallFont = g.Context.FontAtlas.AddFont("iosevskanerdfont.ttf", 15)

	g.Context.FontAtlas.SetDefaultFont("iosevskanerdfont.ttf", 24)

	editor = g.CodeEditor().
		ShowWhitespaces(false).
		TabSize(2).
		Text("select * from greeting\nwhere date > current_timestamp\norder by date").
		Border(true)

	statusMsg = "q quit, v view"
	visibleWidgets = []g.Widget{
		g.Column(
			g.Row(
				g.Label(statusMsg).Font(smallFont),
			),
			g.ListBox("actionList", []string{"Jwt Parse", "Base64 Decode", "aaaaa", "bbbbb", "cccccc", "dddddd", "eeeeee", "ffffff", "aaaaa", "bbbbb", "cccccc", "dddddd", "eeeeee"}),
			// g.Style().
			// 	SetStyle(g.StyleVarFramePadding, 10, -30).
			// 	SetFontSize(90).To(
			// 	g.Button(string(0xe342)),
			// ),
			//
			// 	g.Button("Click Me").OnClick(onClickMe),
			// 	g.Button("I'm so cute").OnClick(onImSoCute),
			// 	g.Button("salut").OnClick(func() { fmt.Println("yo") }),
			// ),

		),
	}

	wnd.Run(loop)
}
