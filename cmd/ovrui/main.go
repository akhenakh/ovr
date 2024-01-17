package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"github.com/AllenDang/giu"
	g "github.com/AllenDang/giu"
	"github.com/sahilm/fuzzy"
	"golang.design/x/clipboard"

	"github.com/akhenakh/ovr/action"
)

var bigFont *g.FontInfo

type App struct {
	in             []byte
	out            *action.Data
	r              *action.ActionRegistry
	statusMsg      string
	dataMsg        string
	visibleWidgets []g.Widget
	searchInput    string
	state          UIState
}

type UIState uint8

const (
	HomeState UIState = iota
	SearchState
	ViewState
)

const defaultStatusMsg = "q quit, v view, / search"

func newApp(in []byte) *App {
	out := action.NewDataText(in)

	bigFont = g.Context.FontAtlas.AddFont("iosevskanerdfont.ttf", 20)

	g.Context.FontAtlas.SetDefaultFont("iosevskanerdfont.ttf", 15)

	statusMsg := defaultStatusMsg

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
	a.state = HomeState

	txtDisplay := a.out.String()
	if len(txtDisplay) > 400 {
		txtDisplay = txtDisplay[0:400]
	}
	lines := strings.Split(txtDisplay, "\n")
	if len(lines) > 4 {
		txtDisplay = strings.Join(lines[0:4], "\n")
	}

	var info string

	switch a.out.Format.Name {
	case action.TextFormat.Name:
		info = fmt.Sprintf("%d", len(a.out.String()))
	case action.TextListFormat.Name:
		l := a.out.Value.([]string)
		info = fmt.Sprintf("%d", len(l))
	case action.TimeFormat.Name:
		l := a.out.Value.(time.Time)
		_, offset := l.Zone()
		info = fmt.Sprintf("%+d", offset/60/60)
	}

	a.visibleWidgets = []g.Widget{
		g.Row(g.Style().
			SetColor(g.StyleColorText, color.RGBA{0x22, 0xDD, 0x22, 255}).
			To(
				g.Label(strings.ToUpper(fmt.Sprintf("%s\n%s", a.out.Format.Name, info))), // TODO: number of elements or lines ...
			),
			g.Label(string(txtDisplay)).Wrapped(true)),

		a.listBox(""),
		g.Label(statusMsg).Wrapped(true),
	}
}

// searchView display the filtered list of actions
func (a *App) searchView(search string) {
	a.state = SearchState

	a.statusMsg = "Search: Type to find an action, enter or double click to execute, ESC to close"
	a.visibleWidgets = []g.Widget{
		g.InputText(&a.searchInput).Hint("Type to fuzzy search for an action, ESC to close").
			OnChange(func() {
				a.searchView(a.searchInput)
			}),
		a.listBox(a.searchInput),
		g.Label(a.statusMsg),
	}
}

// editorView displays the full window editor
func (a *App) editorView(statusMsg string) {
	a.state = ViewState

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

func (a *App) listBox(filter string) g.Widget {
	actions := a.r.ActionsForData(a.out)

	items := make([]string, len(actions))
	for i := 0; i < len(actions); i++ {
		items[i] = strings.Title(actions[i].Title())
	}

	if filter != "" {
		matches := fuzzy.Find(a.searchInput, items)
		res := make([]string, matches.Len())
		for i, m := range matches {
			res[i] = m.Str
		}
		items = res
	}

	listBox := g.ListBox("actionList", items).Size(g.Auto, -20)

	// when an action is selected in the list
	listBox.OnDClick(func(idx int) {
		act := actions[idx]

		out, err := act.Transform(a.out)
		if err != nil {
			a.defaultView("Error " + err.Error())

			return
		}
		a.out = out

		a.defaultView("Applied " + act.Title())
	})

	return g.Style().SetFont(bigFont).To(listBox)
}

func (a *App) loop() {
	g.SingleWindow().RegisterKeyboardShortcuts(
		// up arrow command
		giu.WindowShortcut{Key: giu.KeyUp, Callback: func() {
			if a.state == HomeState || a.state == SearchState {
				fmt.Println("UP")
				// TODO
			}
		}},

		// down arrow command
		giu.WindowShortcut{Key: giu.KeyDown, Callback: func() {
			if a.state == HomeState || a.state == SearchState {
				fmt.Println("DOWN")
				// TODO
			}
		}},

		// enter command
		giu.WindowShortcut{Key: giu.KeyEnter, Callback: func() {
			if a.state == HomeState || a.state == SearchState {
				// TODO
			}
		}},

		// quit command
		giu.WindowShortcut{Key: giu.KeyQ, Callback: func() {
			if a.state == HomeState {
				fmt.Printf("%s\n---\n%s\n", a.out.StackString(), a.out.String())

				clipboard.Write(clipboard.FmtText, []byte(a.out.String()))

				os.Exit(0)
			}
		}},

		// search command
		giu.WindowShortcut{Key: giu.KeySlash, Callback: func() {
			if a.state == HomeState {
				a.statusMsg = "Search: Type to find an action, enter or double click to execute, ESC to close"
				a.searchView("")
			}
		}},

		// view command
		giu.WindowShortcut{Key: giu.KeyV, Callback: func() {
			if a.state == HomeState {
				a.editorView("ESC to quit the editor")
			}
		}},

		// delete from stack command
		giu.WindowShortcut{Key: giu.KeyBackspace, Callback: func() {
			if a.state == HomeState {
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
			if a.state == ViewState || a.state == SearchState {
				a.defaultView(defaultStatusMsg)
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
