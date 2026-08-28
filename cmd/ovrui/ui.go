package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/akhenakh/ovr/action"

	. "go.hasen.dev/shirei"
	. "go.hasen.dev/shirei/widgets"
)

type f32 = float32

type AppState struct {
	registry *action.ActionRegistry
	in       []byte
	out      *action.Data
	actions  []action.Action
	search   string
	selected action.Action
	paramBuf []string
	tab      int
	status   string
	isErr    bool
}

var appData = &AppState{}

func setInput(in []byte) {
	if appData.registry == nil {
		appData.registry = action.DefaultRegistry()
	}
	r := appData.registry
	out := action.NewDataText(in)

	// if the input looks like WKT and the wkt action is available,
	// start with geometry data so geo actions are offered right away
	if action.GuessWKT(in) {
		if a, ok := r.ActionByName(action.TextFormat, "wkt"); ok {
			if d, err := a.Transform(action.NewDataText(in)); err == nil {
				out = d
			}
		}
	}

	appData.in = in
	appData.out = out
	appData.selected = nil
	appData.paramBuf = nil
	appData.tab = 0
	appData.status = ""
	appData.isErr = false
	refreshActions()
}

func refreshActions() {
	appData.actions = appData.registry.ActionsForData(appData.out)
}

func filteredActions() []action.Action {
	q := strings.ToLower(strings.TrimSpace(appData.search))
	if q == "" {
		return appData.actions
	}
	var out []action.Action
	for _, a := range appData.actions {
		if strings.Contains(strings.ToLower(a.Title()+" "+a.Doc()), q) {
			out = append(out, a)
		}
	}
	return out
}

func selectFirst() {
	list := filteredActions()
	if len(list) > 0 {
		selectAction(list[0])
	}
}

func selectAction(a action.Action) {
	if appData.selected == a {
		return
	}
	appData.selected = a
	appData.paramBuf = make([]string, len(a.Parameters()))
}

func setStatus(msg string, isErr bool) {
	appData.status = msg
	appData.isErr = isErr
}

func applySelected() {
	if appData.selected == nil {
		setStatus("select an action first", true)
		return
	}
	applyAction(appData.selected)
}

func applyAction(a action.Action) {
	if params := a.Parameters(); len(params) > 0 {
		values := make([]any, len(params))
		for i, p := range params {
			v := strings.TrimSpace(appData.paramBuf[i])
			switch p.ActionParameterType {
			case action.IntParameter:
				n, err := strconv.Atoi(v)
				if err != nil {
					setStatus(fmt.Sprintf("%s: %s needs an integer", a.Title(), p.Doc), true)
					return
				}
				values[i] = n
			case action.FloatParameter:
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					setStatus(fmt.Sprintf("%s: %s needs a number", a.Title(), p.Doc), true)
					return
				}
				values[i] = f
			case action.StringParameter:
				if v == "" {
					setStatus(fmt.Sprintf("%s: %s is required", a.Title(), p.Doc), true)
					return
				}
				values[i] = v
			}
		}
		if err := a.SetInputParameters(values...); err != nil {
			setStatus(err.Error(), true)
			return
		}
	}

	out, err := a.Transform(appData.out)
	if err != nil {
		setStatus(err.Error(), true)
		return
	}
	appData.out = out
	appData.tab = 0
	refreshActions()
	setStatus("applied "+a.Title(), false)
}

func undoAction() {
	nd, oa, err := appData.out.Undo(appData.in)
	if err != nil {
		setStatus(err.Error(), true)
		return
	}
	appData.out = nd
	refreshActions()
	setStatus("removed "+oa.Title(), false)
}

func copyOutput() {
	RequestTextCopy(appData.out.String())
	setStatus("output copied to clipboard", false)
}

func reloadInput(in []byte) {
	setInput(in)
	setStatus(fmt.Sprintf("input reloaded (%d bytes)", len(in)), false)
}

func dataSummary() string {
	s := appData.out.String()
	n := len(s)
	size := fmt.Sprintf("%d B", n)
	if n >= 1024 {
		size = fmt.Sprintf("%.1f KB", f32(n)/1024)
	}
	return fmt.Sprintf("%s · %s", appData.out.Format.Name, size)
}

func RootView() {
	switch GetFrameInput().Key {
	case KeyEnter:
		applySelected()
	case KeyEscape:
		appData.search = ""
	}

	Container(Attrs(Viewport, AmendTextStyle(Fonts(fontFamily)), Background(220, 10, 97, 1)), func() {
		Header()
		Toolbar()
		ParamsPanel()
		Container(Attrs(Row, Grow(1), Expand, Extrinsic, Clip, Gap(8), Pad2(4, 12)), func() {
			ActionsPane()
			OutputPane()
		})
		StatusBar()
	})
}

func Header() {
	Container(Attrs(Row, Expand, CrossMid, Gap(10), Pad2(10, 14), Background(220, 25, 18, 1)), func() {
		Label("ovr", FontSize(17), FontWeight(WeightBold), TextColor(0, 0, 100, 1))
		Label("companion", FontSize(11), TextColor(0, 0, 78, 1))
		Filler(1)
		Container(Attrs(Pad2(3, 10), Corners(10), Background(220, 30, 32, 1)), func() {
			Label(dataSummary(), FontSize(11), TextColor(0, 0, 92, 1))
		})
	})
}

func Toolbar() {
	Container(Attrs(Row, Expand, CrossMid, Gap(8), Pad2(8, 14)), func() {
		TextInputExt(&appData.search, TextInputAttrs{Placeholder: "filter actions…", MaxWidth: 340, MaxLines: 1})
		if Button(SymRefresh, "Reload") {
			reloadClipboard()
		}
		if Button(NoIcon, "Apply") {
			applySelected()
		}
		if Button(SymUndo, "Undo") {
			undoAction()
		}
		if Button(SymCopy, "Copy Output") {
			copyOutput()
		}
	})
}

func ParamsPanel() {
	a := appData.selected
	if a == nil || len(a.Parameters()) == 0 {
		return
	}
	Container(Attrs(Row, Expand, CrossMid, Gap(10), Pad2(2, 14), Wrap), func() {
		Label("Parameters", FontSize(12), FontWeight(WeightBold), TextColor(0, 0, 30, 1))
		for i, p := range a.Parameters() {
			Container(Attrs(Row, CrossMid, Gap(6)), func() {
				Label(p.Doc, FontSize(11), TextColor(0, 0, 45, 1))
				TextInputExt(&appData.paramBuf[i], TextInputAttrs{Placeholder: p.Doc, MaxWidth: 160, NoAutoFocus: true})
			})
		}
	})
}

func ActionsPane() {
	list := filteredActions()
	Container(Attrs(FixWidth(420), Expand, Clip, Corners(6), Background(220, 12, 94, 1), BorderWidth(1), BorderColor(220, 12, 85, 1)), func() {
		Container(Attrs(Row, CrossMid, Pad2(6, 10)), func() {
			Label(fmt.Sprintf("Actions (%d)", len(list)), FontSize(12), FontWeight(WeightBold), TextColor(0, 0, 30, 1))
		})
		if len(list) == 0 {
			Container(Attrs(Expand, Center), func() {
				Label("no matching actions", FontSize(12), TextColor(0, 0, 55, 1))
			})
			return
		}
		Container(Attrs(Grow(1), Expand, Clip), func() {
			VirtualListView(
				nil,
				len(list),
				func(i int) any { return list[i] },
				func(i int, w f32) f32 { return 52 },
				func(i int, w f32) { actionRow(list[i]) },
			)
		})
	})
}

func actionRow(a action.Action) {
	Container(Attrs(Expand, FixHeight(52), Pad2(4, 10), Clip), func() {
		selected := appData.selected == a
		if selected {
			ModAttrs(Background(210, 65, 86, 1), Corners(4))
		} else if IsHovered() {
			ModAttrs(Background(220, 18, 90, 1), Corners(4))
		}
		if IsClicked() {
			selectAction(a)
		}
		if IsDoubleClicked() {
			applySelected()
		}
		Container(Attrs(Expand, Clip), func() {
			Label(a.Title(), FontSize(13), FontWeight(WeightBold), TextColor(0, 0, 18, 1))
			Label(a.Doc(), FontSize(11), TextColor(0, 0, 42, 1))
		})
	})
}

func OutputPane() {
	Container(Attrs(Grow(1), Expand, Extrinsic, Clip, Corners(6), Background(220, 12, 95, 1), BorderWidth(1), BorderColor(220, 12, 85, 1)), func() {
		Container(Attrs(Row, CrossMid, Pad2(6, 10), Gap(6)), func() {
			tabButton("Output", 0)
			tabButton("Input", 1)
			Filler(1)
			Label(dataSummary(), FontSize(11), TextColor(0, 0, 45, 1))
		})
		if appData.tab == 0 {
			LargeText(appData.out.String())
		} else {
			LargeText(string(appData.in))
		}
	})
}

func tabButton(label string, tab int) {
	active := appData.tab == tab
	Container(Attrs(Pad2(4, 12), Corners(4)), func() {
		if active {
			ModAttrs(Background(210, 60, 45, 1))
		} else if IsHovered() {
			ModAttrs(Background(220, 18, 88, 1))
		}
		if IsClicked() {
			appData.tab = tab
		}
		if active {
			Label(label, FontSize(12), TextColor(0, 0, 100, 1))
		} else {
			Label(label, FontSize(12), TextColor(0, 0, 30, 1))
		}
	})
}

func StatusBar() {
	Container(Attrs(Row, Expand, CrossMid, Gap(8), Pad2(6, 14)), func() {
		if appData.status != "" {
			if appData.isErr {
				Label(appData.status, FontSize(12), TextColor(0, 70, 45, 1))
			} else {
				Label(appData.status, FontSize(12), TextColor(150, 60, 32, 1))
			}
		} else {
			Label("click to select · double-click or Enter to apply · Esc clears the filter", FontSize(11), TextColor(0, 0, 55, 1))
		}
		Filler(1)
		Label(appData.out.StackString(), FontSize(11), TextColor(0, 0, 45, 1))
	})
}
