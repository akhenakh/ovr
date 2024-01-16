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

var smallFont *g.FontInfo

type App struct {
	in                    []byte
	out                   *action.Data
	r                     *action.ActionRegistry
	editor                *g.CodeEditorWidget
	statusMsg             string
	dataMsg               string
	listBox               *g.ListBoxWidget
	disableGlobalShortcut bool // disable global shortcut while using editor or /
	visibleWidgets        []g.Widget
	searchInput           string
}

func newApp(in []byte) *App {
	out := action.NewDataText(in)

	smallFont = g.Context.FontAtlas.AddFont("iosevskanerdfont.ttf", 15)

	g.Context.FontAtlas.SetDefaultFont("iosevskanerdfont.ttf", 24)

	editor := g.CodeEditor().
		ShowWhitespaces(false).
		Text("").
		Border(true).LanguageDefinition(giu.LanguageDefinitionC)

	statusMsg := "q quit, v view, / search"

	r := action.DefaultRegistry()

	// Make initial list of items
	actions := r.ActionsForData(out)
	items := make([]string, len(actions))
	for i := 0; i < len(actions); i++ {
		items[i] = actions[i].Title()
	}
	listBox := g.ListBox("actionList", items).Size(g.Auto, -20)

	visibleWidgets := []g.Widget{
		g.Row(g.Style().
			SetColor(g.StyleColorText, color.RGBA{0x11, 0xDD, 0x11, 255}).
			To(
				g.Label("TEXT"),
			),
			g.Label(string(in)).Font(smallFont)),
		listBox,
		g.Label(statusMsg).Font(smallFont),
		// g.Column(
		// 	// g.Row(
		// 	// 	g.Label(statusMsg).Font(smallFont),
		// 	// ),

		// ),

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

	}

	a := &App{
		in:             in,
		out:            out,
		r:              r,
		editor:         editor,
		listBox:        listBox,
		statusMsg:      statusMsg,
		visibleWidgets: visibleWidgets,
	}

	// when an action is selected in the list
	listBox.OnDClick(func(idx int) {
		// defer giu.Update()
		act := actions[idx]

		out, err := act.Transform(a.out)
		if err != nil {
			statusMsg = "Error " + err.Error()
			a.visibleWidgets = []g.Widget{
				g.Row(g.Style().
					SetColor(g.StyleColorText, color.RGBA{0x11, 0xDD, 0x11, 255}).
					To(
						g.Label(strings.ToUpper(a.out.Format.Name)),
					),
					g.Label(string(a.out.String())).Font(smallFont)),
				listBox,
				g.Label(statusMsg).Font(smallFont),
			}
			return
		}
		a.out = out
		fmt.Println("action selected", act.Title(), "updated", out.String())

		a.editor.Text(string(a.out.String()))

		actions := r.ActionsForData(out)
		items := make([]string, len(actions))
		for i := 0; i < len(actions); i++ {
			items[i] = actions[i].Title()
		}
		a.listBox = g.ListBox("actionList", items).Size(g.Auto, -20)

		a.visibleWidgets = []g.Widget{
			g.Row(g.Style().
				SetColor(g.StyleColorText, color.RGBA{0x11, 0xDD, 0x11, 255}).
				To(
					g.Label(strings.ToUpper(a.out.Format.Name)),
				),
				g.Label(string(a.out.String())).Font(smallFont)),
			listBox,
			g.Label(statusMsg).Font(smallFont),
		}
	})

	return a
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
				a.statusMsg = "Search: Type to find an action"
				a.visibleWidgets = []g.Widget{
					g.InputText(&a.searchInput).Hint("Type to search for an action, ESC to quit"),
					a.listBox,
					g.Label(a.statusMsg).Font(smallFont),
				}
				// giu.SetKeyboardFocusHere()
			}
		}},

		// view command
		giu.WindowShortcut{Key: giu.KeyV, Callback: func() {
			if !a.disableGlobalShortcut {
				a.statusMsg = "ESC to quit the editor"

				a.editor.Text(string(a.in))
				a.visibleWidgets = []g.Widget{
					a.editor.Size(g.Auto, -20),
					g.Label(a.statusMsg).Font(smallFont),
				}
			}
		}},

		// delete from stack command
		giu.WindowShortcut{Key: giu.KeyBackspace, Callback: func() {
			if !a.disableGlobalShortcut {
				d, oa, err := a.out.Undo(a.in)
				if err != nil { // we should not have errors in the stack
					a.statusMsg = "Error " + err.Error()
					return
				}
				a.out = d
				a.statusMsg = "Removed action: " + oa.Title()
				a.editor.Text(string(a.in))

				actions := a.r.ActionsForData(a.out)
				items := make([]string, len(actions))
				for i := 0; i < len(actions); i++ {
					items[i] = actions[i].Title()
				}
				a.listBox = g.ListBox("actionList", items).Size(g.Auto, -20)
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
