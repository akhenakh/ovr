package main

import (
	_ "embed"
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

//go:embed iosevskanerdfont.ttf
var bigFontBytes []byte

var bigFont *g.FontInfo

type App struct {
	in  []byte
	out *action.Data
	r   *action.ActionRegistry

	// UI current state
	state uiState

	statusMsg string

	visibleWidgets []g.Widget

	// list of actions
	selectedIndex int32
	actionsList   action.Actions

	// search
	searchInput string

	// params
	actionParams []any
}

type uiState uint8

const (
	homeState uiState = iota
	searchState
	viewState
	paramsState
)

const defaultStatusMsg = "ESC quit, v view, / search"

func newApp(in []byte) *App {
	out := action.NewDataText(in)

	bigFont = g.Context.FontAtlas.AddFontFromBytes("iosevskanerdfont.ttf", bigFontBytes, 20)

	g.Context.FontAtlas.SetDefaultFontFromBytes(bigFontBytes, 15)

	statusMsg := defaultStatusMsg

	r := action.DefaultRegistry()

	a := &App{
		in:            in,
		out:           out,
		r:             r,
		statusMsg:     statusMsg,
		selectedIndex: 0,
	}

	a.defaultView(statusMsg)

	return a
}

// defaultView display the "home" with the list of actions
func (a *App) defaultView(statusMsg string) {
	a.state = homeState

	txtDisplay := a.out.String()
	if len(txtDisplay) > 400 {
		txtDisplay = txtDisplay[0:400] + "..."
	}
	lines := strings.Split(txtDisplay, "\n")
	if len(lines) > 4 {
		txtDisplay = strings.Join(lines[0:4], "\n") + "..."
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
func (a *App) searchView(focus bool) {
	a.state = searchState

	inputText := g.InputText(&a.searchInput).Hint("Type to fuzzy search for an action, ESC to close").
		OnChange(func() {
			a.searchView(false)
			a.selectedIndex = 0
		})
	if !focus {
		inputText = a.visibleWidgets[1].(*g.InputTextWidget)
	}
	a.statusMsg = "Search: Type to find an action, enter or double click to execute, ESC to close"
	a.visibleWidgets = []g.Widget{
		g.Custom(func() {
			g.SetKeyboardFocusHere()
			g.SetItemDefaultFocus()
		}),
		inputText,
		a.listBox(a.searchInput),
		g.Label(a.statusMsg),
	}
}

// editorView displays the full window editor
func (a *App) editorView(statusMsg string) {
	a.state = viewState

	editor := g.CodeEditor().
		ShowWhitespaces(false).
		Text("").
		Border(true).LanguageDefinition(giu.LanguageDefinitionC)

	editor.Text(a.out.String())
	a.visibleWidgets = []g.Widget{
		giu.Custom(func() {
			giu.SetKeyboardFocusHere()
		}),
		editor.Size(g.Auto, -20),
		g.Label(statusMsg),
	}
}

// editorView displays the full window editor
func (a *App) paramsView(act *action.Action) {
	a.state = paramsState

	widgets := []g.Widget{g.Label(fmt.Sprintf("Enter parameters for %s action", act.Title()))}

	a.actionParams = make([]any, len(act.Parameters))

	for i, p := range act.Parameters {
		switch p {
		case action.IntActionParameter:
			var val int32

			if i == 0 {
				widgets = append(widgets, g.Custom(func() {
					giu.SetKeyboardFocusHere()
				}))
			}

			widgets = append(widgets, g.Row(
				g.Style().SetFont(bigFont).To(
					g.InputInt(&val).
						Label(fmt.Sprintf("%s %d param int", act.Title(), i)).
						OnChange(func() {
							a.actionParams[i] = int(val)
						}),
				),
			),
				g.Label("ESC to quit, enter to validate"))
		}
	}

	a.visibleWidgets = widgets
}

func (a *App) listBox(filter string) g.Widget {
	actions := action.Actions(a.r.ActionsForData(a.out))

	a.actionsList = actions

	if filter != "" {
		matches := fuzzy.FindFrom(a.searchInput, actions)
		res := make(action.Actions, matches.Len())
		for i, m := range matches {
			res[i] = actions[m.Index]
		}
		actions = res
		a.actionsList = res
	}

	items := make([]string, len(actions))
	for i := 0; i < len(actions); i++ {
		items[i] = actions[i].FullDescription()
	}

	listBox := g.ListBox(items).Size(g.Auto, -20).SelectedIndex(&a.selectedIndex)

	// when an action is selected in the list
	listBox.OnDClick(func(idx int) {
		act := actions[idx]
		if act == nil {
			a.defaultView("Error can't find this action")

			return
		}

		// This action has parameters, we need the ui to ask for those
		if len(act.Parameters) > 0 {
			a.paramsView(act)

			return
		}

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
			if a.state == homeState || a.state == searchState {
				if a.selectedIndex == 0 {
					return
				}
				a.selectedIndex--
			}
		}},

		// down arrow command
		giu.WindowShortcut{Key: giu.KeyDown, Callback: func() {
			if a.state == homeState || a.state == searchState {
				if a.selectedIndex+1 >= int32(a.actionsList.Len()) {
					return
				}
				a.selectedIndex++
			}
		}},

		// enter command
		giu.WindowShortcut{Key: giu.KeyEnter, Callback: func() {
			act := a.actionsList[a.selectedIndex]
			if act == nil {
				a.defaultView("Error can't find this action")

				return
			}

			if a.state == homeState || a.state == searchState {
				a.searchInput = ""

				// This action has parameters, we need the ui to ask for those
				if len(act.Parameters) > 0 {
					a.paramsView(act)

					return
				}

				out, err := act.Transform(a.out)
				if err != nil {
					a.defaultView("Error " + err.Error())

					return
				}
				a.out = out

				a.defaultView("Applied " + act.Title())
			} else if a.state == paramsState {
				if len(a.actionParams) != len(act.Parameters) {
					a.defaultView("Error incorrect number of parameters")

					return
				}

				act.InputParameters = a.actionParams

				out, err := act.Transform(a.out)
				if err != nil {
					a.defaultView("Error " + err.Error())

					return
				}
				a.out = out

				a.defaultView("Applied " + act.Title())

			}
		}},

		// search command
		giu.WindowShortcut{Key: giu.KeySlash, Callback: func() {
			if a.state == homeState {
				a.statusMsg = "Search: Type to find an action, enter or double click to execute, ESC to close"
				a.searchView(true)
			}
		}},

		// view command
		giu.WindowShortcut{Key: giu.KeyV, Callback: func() {
			if a.state == homeState {
				a.editorView("ESC to quit the editor")
			}
		}},

		// delete from stack command
		giu.WindowShortcut{Key: giu.KeyBackspace, Callback: func() {
			if a.state == homeState {
				d, oa, err := a.out.Undo(a.in)
				if err != nil { // we should not have errors in the stack
					a.statusMsg = "Error " + err.Error()
					return
				}
				a.out = d

				a.defaultView("Removed action: " + oa.Title())
			}
		}},

		// Back for Editor & View, Quit for Home
		giu.WindowShortcut{Key: giu.KeyEscape, Callback: func() {
			if a.state == viewState || a.state == searchState || a.state == paramsState {
				a.searchInput = ""
				a.defaultView(defaultStatusMsg)
			} else if a.state == homeState {
				fmt.Printf("%s\n---\n%s\n", a.out.StackString(), a.out.String())

				clipboard.Write(clipboard.FmtText, []byte(a.out.String()))

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

	wnd := g.NewMasterWindow("OVR", 640, 480, g.MasterWindowFlagsFrameless)
	app := newApp(input)
	wnd.Run(app.loop)
}
