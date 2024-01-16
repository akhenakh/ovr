package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/AllenDang/giu"
	g "github.com/AllenDang/giu"
	"github.com/akhenakh/ovr/action"
	"golang.design/x/clipboard"
)

var bigFont *g.FontInfo

type App struct {
	in                    []byte
	out                   *action.Data
	r                     *action.ActionRegistry
	statusMsg             string
	dataMsg               string
	disableGlobalShortcut bool // disable global shortcut while using editor or /
	visibleWidgets        []g.Widget
	searchInput           string
}

func newApp(in []byte) *App {
	out := action.NewDataText(in)

	bigFont = g.Context.FontAtlas.AddFont("iosevskanerdfont.ttf", 20)

	g.Context.FontAtlas.SetDefaultFont("iosevskanerdfont.ttf", 15)

	statusMsg := "q quit, v view, / search"

	r := action.DefaultRegistry()

	a := &App{
		in:        in,
		out:       out,
		r:         r,
		statusMsg: statusMsg,
	}

	a.defaultView(statusMsg)

	return a
}

// defaultView display the "home" with the list of actions
func (a *App) defaultView(statusMsg string) {
	a.visibleWidgets = []g.Widget{
		g.Row(g.Style().
			SetColor(g.StyleColorText, color.RGBA{0x11, 0xDD, 0x11, 255}).
			To(
				g.Label(strings.ToUpper(a.out.Format.Name)),
			),
			g.Label(string(a.out.String())).Wrapped(true)),

		a.listBox(),
		g.Label(statusMsg).Wrapped(true),
	}
}

// searchView display the filtered list of actions
func (a *App) searchView(search string) {
	a.statusMsg = "Search: Type to find an action, enter or double click to execute, ESC to close"
	a.visibleWidgets = []g.Widget{
		g.InputText(&a.searchInput).Hint("Type to search for an action, ESC to quit"),
		a.listBox(),
		g.Label(a.statusMsg),
	}
}

// editorView displays the full window editor
func (a *App) editorView(statusMsg string) {
	editor := g.CodeEditor().
		ShowWhitespaces(false).
		Text("").
		Border(true).LanguageDefinition(giu.LanguageDefinitionC)

	editor.Text(a.out.String())
	a.visibleWidgets = []g.Widget{
		editor.Size(g.Auto, -20),
		g.Label(statusMsg),
	}
}

func (a *App) listBox() g.Widget {
	actions := a.r.ActionsForData(a.out)
	items := make([]string, len(actions))
	for i := 0; i < len(actions); i++ {
		items[i] = strings.Title(actions[i].Title())
	}

	listBox := g.ListBox("actionList", items).Size(g.Auto, -20)

	// when an action is selected in the list
	listBox.OnDClick(func(idx int) {
		// defer giu.Update()
		act := actions[idx]

		out, err := act.Transform(a.out)
		if err != nil {
			a.defaultView("Error " + err.Error())

			return
		}
		a.out = out
		fmt.Println("action selected", act.Title(), "updated", out.String())

		a.defaultView("Applied " + act.Title())
	})

	return g.Style().SetFont(bigFont).To(listBox)
}

func (a *App) loop() {
	g.SingleWindow().RegisterKeyboardShortcuts(
		// quit command
		giu.WindowShortcut{Key: giu.KeyQ, Callback: func() {
			if !a.disableGlobalShortcut {
				os.Exit(0)
			}
		}},

		// search command
		giu.WindowShortcut{Key: giu.KeySlash, Callback: func() {
			fmt.Println("Slash")
			if !a.disableGlobalShortcut {
				a.statusMsg = "Search: Type to find an action, enter or double click to execute, ESC to close"
				a.searchView("")
			}
		}},

		// view command
		giu.WindowShortcut{Key: giu.KeyV, Callback: func() {
			if !a.disableGlobalShortcut {
				a.editorView("ESC to quit the editor")
			}
		}},

		// delete from stack command
		giu.WindowShortcut{Key: giu.KeyBackspace, Callback: func() {
			fmt.Println("action remove requested")

			if !a.disableGlobalShortcut {
				d, oa, err := a.out.Undo(a.in)
				if err != nil { // we should not have errors in the stack
					a.statusMsg = "Error " + err.Error()
					return
				}
				a.out = d

				a.defaultView("Removed action: " + oa.Title())
			}
		}},

		// close editor
		giu.WindowShortcut{Key: giu.KeyEscape, Callback: func() {
			if !a.disableGlobalShortcut {
				os.Exit(0)
			}
		}},
	).Layout(a.visibleWidgets...)
}

func main() {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	input := []byte(clipboard.Read(clipboard.FmtText))

	wnd := g.NewMasterWindow("OVR", 640, 480, 0)
	app := newApp(input)
	wnd.Run(app.loop)
}
