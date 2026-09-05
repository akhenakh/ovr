package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/akhenakh/ovr/action"

	. "go.hasen.dev/shirei"
	. "go.hasen.dev/shirei/widgets"
)

type f32 = float32

// quitApp is indirected so tests can intercept the exit
var quitApp = os.Exit

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
	// keyboard focus mode: true when the actions list holds the keyboard,
	// false when the filter field does
	listFocus bool
	// handle of the filter text field, captured each frame
	filterId ContainerId
	// set to re-assert focus on the filter field during the next build pass
	refocusFilter bool
}

var appData = &AppState{}

// actionsListKey identifies the actions VirtualListView so scroll-into-view
// commands can target it
var actionsListKey = new(int)

// themedTextInput renders a single-line text field with the active theme's
// chrome (the stock TextInputExt hardcodes a light face).
func themedTextInput(buf *string, attrs TextInputAttrs) {
	raw := TextInputConfigFromAttrs(attrs)
	raw.TextColor = theme.fieldText
	raw.CaretColor = theme.fieldCaret
	raw.PlaceholderColor = theme.fieldHint

	// replicate the stock chrome sizing: comfort-scale font & padding
	cfg := raw
	if cfg.FontSize == 0 {
		cfg.FontSize = DefaultTextSize
	}
	if cfg.Padding == (Vec4{}) {
		cfg.Padding = N4(cfg.FontSize / 2)
	}
	s := ComfortScale()
	cfg.FontSize *= s
	for i := range cfg.Padding {
		cfg.Padding[i] *= s
	}
	rows := cfg.Rows
	if rows == 0 {
		rows = 1
	}
	padSize := PadSize(cfg.Padding)
	boxH := f32(rows)*cfg.FontSize + padSize[1]
	minW := padSize[0] + cfg.FontSize*10
	if cfg.MinWidth > 0 {
		minW = cfg.MinWidth
	}

	Container(Attrs(
		Focusable,
		Corners(4),
		BackgroundVec(theme.fieldBg),
		PadVec(cfg.Padding),
		MinSizeVec(Vec2{minW, boxH}),
		MaxSizeVec(Vec2{cfg.MaxWidth, boxH}),
		Clip,
		BorderWidth(1),
		BorderColorVec(theme.fieldBorder),
	), func() {
		if !cfg.FixedWidth && cfg.MaxWidth == 0 {
			ModAttrs(Expand)
			if GetAttrs().Row {
				ModAttrs(Grow(1))
			}
		}
		st := ProcessTextInput(buf, raw)
		if st.HasFocus {
			ModAttrs(BorderColorVec(theme.fieldFocus))
		} else {
			ModAttrs(BorderColorVec(theme.fieldBorder))
		}
		DrawTextInputPlain(st, cfg)
	})
}

func setInput(in []byte) {
	if appData.registry == nil {
		appData.registry = action.DefaultRegistry()
		registerThemeActions(appData.registry)
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
	appData.listFocus = false
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

// moveSelection moves the list selection one row down or up (wrapping at the
// ends) and scrolls the list so the newly selected row is visible.
func moveSelection(down bool) {
	list := filteredActions()
	if len(list) == 0 {
		return
	}
	idx := -1
	for i, a := range list {
		if appData.selected == a {
			idx = i
			break
		}
	}
	switch {
	case idx == -1 && down:
		idx = 0
	case idx == -1:
		idx = len(list) - 1
	case down:
		idx = (idx + 1) % len(list)
	default:
		idx = (idx - 1 + len(list)) % len(list)
	}
	selectAction(list[idx])
	VirtualListScrollIntoView(actionsListKey, list[idx])
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

// focusList moves keyboard focus from the filter field to the actions list
// without changing the selection.
func focusList() {
	appData.listFocus = true
	ClearFocus()
}

// activateList moves keyboard focus from the filter field to the actions
// list: the first matching action becomes the selection.
func activateList() {
	focusList()
	selectFirst()
}

// exitListFocus moves keyboard focus back to the filter field.
func exitListFocus() {
	appData.listFocus = false
	appData.refocusFilter = true
}

func applySelected() bool {
	if appData.selected == nil {
		setStatus("select an action first", true)
		return false
	}
	if !applyAction(appData.selected) {
		return false
	}
	// applying from the list resets the filter and returns focus to it
	appData.search = ""
	exitListFocus()
	return true
}

func applyAction(a action.Action) bool {
	if params := a.Parameters(); len(params) > 0 {
		if len(appData.paramBuf) != len(params) {
			setStatus(fmt.Sprintf("%s: missing parameters", a.Title()), true)
			return false
		}
		values := make([]any, len(params))
		for i, p := range params {
			v := strings.TrimSpace(appData.paramBuf[i])
			switch p.ActionParameterType {
			case action.IntParameter:
				n, err := strconv.Atoi(v)
				if err != nil {
					setStatus(fmt.Sprintf("%s: %s needs an integer", a.Title(), p.Doc), true)
					return false
				}
				values[i] = n
			case action.FloatParameter:
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					setStatus(fmt.Sprintf("%s: %s needs a number", a.Title(), p.Doc), true)
					return false
				}
				values[i] = f
			case action.StringParameter:
				if v == "" {
					setStatus(fmt.Sprintf("%s: %s is required", a.Title(), p.Doc), true)
					return false
				}
				values[i] = v
			}
		}
		if err := a.SetInputParameters(values...); err != nil {
			setStatus(err.Error(), true)
			return false
		}
	}

	out, err := a.Transform(appData.out)
	if err != nil {
		setStatus(err.Error(), true)
		return false
	}
	appData.out = out
	appData.tab = 0
	appData.selected = nil
	appData.paramBuf = nil
	refreshActions()
	setStatus("applied "+a.Title(), false)
	return true
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

func quitHint() string {
	if PrimaryMod() == ModCmd {
		return "Cmd+Q quits"
	}
	return "Ctrl+Q quits"
}

func dataSummary() string {
	if appData.out.Format == action.TextListFormat {
		if l, ok := appData.out.Value.([]string); ok {
			return fmt.Sprintf("%s · %d items", appData.out.Format.Name, len(l))
		}
	}
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
		if appData.listFocus {
			applySelected()
		} else {
			// Enter in the filter field activates the actions list
			activateList()
		}
	case KeyTab:
		// Tab toggles between the filter field and the actions list
		if appData.listFocus {
			exitListFocus()
		} else {
			focusList()
		}
	case KeyDown, KeyUp:
		if appData.listFocus {
			moveSelection(GetFrameInput().Key == KeyDown)
		}
	case KeyEscape:
		appData.search = ""
		exitListFocus()
	case KeyQ:
		if ActiveCombo() == Combo(KeyQ, PrimaryMod()) {
			quitApp(0)
		}
	}

	Container(Attrs(Viewport, AmendTextStyle(Fonts(fontFamily)), BackgroundVec(theme.rootBg)), func() {
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
	Container(Attrs(Row, Expand, CrossMid, Gap(10), Pad2(10, 14), BackgroundVec(theme.headerBg)), func() {
		Label("ovr", FontSize(17), FontWeight(WeightBold), TextColorVec(theme.headerTitle))
		Label("companion", FontSize(11), TextColorVec(theme.headerSub))
		Filler(1)
		Container(Attrs(Pad2(3, 10), Corners(10), BackgroundVec(theme.chipBg)), func() {
			Label(dataSummary(), FontSize(11), TextColorVec(theme.chipText))
		})
	})
}

func Toolbar() {
	Container(Attrs(Row, Expand, CrossMid, Gap(8), Pad2(8, 14)), func() {
		themedTextInput(&appData.search, TextInputAttrs{Placeholder: "filter actions…", MaxWidth: 340, MaxLines: 1})
		appData.filterId = GetLastId()
		// clicking into the field returns keyboard focus mode to the filter
		if IdReceivedFocusNow(appData.filterId) {
			appData.listFocus = false
		}
		if appData.refocusFilter {
			appData.refocusFilter = false
			FocusImmediateOn(appData.filterId)
		}
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
		Label("Parameters", FontSize(12), FontWeight(WeightBold), TextColorVec(theme.paramsLabel))
		for i, p := range a.Parameters() {
			Container(Attrs(Row, CrossMid, Gap(6)), func() {
				Label(p.Doc, FontSize(11), TextColorVec(theme.paramsDoc))
				themedTextInput(&appData.paramBuf[i], TextInputAttrs{
					Placeholder: p.Doc,
					MaxWidth:    160,
					MaxLines:    1,
					NoAutoFocus: i != 0,
				})
			})
		}
		if Button(NoIcon, "Apply") {
			applySelected()
		}
	})
}

func ActionsPane() {
	list := filteredActions()
	Container(Attrs(FixWidth(420), Expand, Clip, Corners(6), BackgroundVec(theme.paneBg), BorderWidth(1), BorderColorVec(theme.paneBorder)), func() {
		// tint the pane while the keyboard focus is on the list
		if appData.listFocus {
			ModAttrs(BackgroundVec(theme.paneFocusBg), BorderColorVec(theme.paneFocusBorder))
		}
		Container(Attrs(Row, CrossMid, Pad2(6, 10)), func() {
			Label(fmt.Sprintf("Actions (%d)", len(list)), FontSize(12), FontWeight(WeightBold), TextColorVec(theme.paneTitle))
		})
		if len(list) == 0 {
			Container(Attrs(Expand, Center), func() {
				Label("no matching actions", FontSize(12), TextColorVec(theme.emptyText))
			})
			return
		}
		Container(Attrs(Grow(1), Expand, Clip), func() {
			VirtualListView(
				actionsListKey,
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
			ModAttrs(BackgroundVec(theme.selRowBg), Corners(4))
		} else if IsHovered() {
			ModAttrs(BackgroundVec(theme.hoverRow), Corners(4))
		}
		if IsClicked() {
			selectAction(a)
			appData.listFocus = true
		}
		if IsDoubleClicked() {
			applySelected()
		}
		Container(Attrs(Expand, Clip), func() {
			if selected {
				Label(a.Title(), FontSize(13), FontWeight(WeightBold), TextColorVec(theme.selTitle))
				Label(a.Doc(), FontSize(11), TextColorVec(theme.selDoc))
			} else {
				Label(a.Title(), FontSize(13), FontWeight(WeightBold), TextColorVec(theme.rowTitle))
				Label(a.Doc(), FontSize(11), TextColorVec(theme.rowDoc))
			}
		})
	})
}

func OutputPane() {
	Container(Attrs(Grow(1), Expand, Extrinsic, Clip, Corners(6), BackgroundVec(theme.paneBg), BorderWidth(1), BorderColorVec(theme.paneBorder)), func() {
		Container(Attrs(Row, CrossMid, Pad2(6, 10), Gap(6)), func() {
			tabButton("Output", 0)
			tabButton("Input", 1)
			Filler(1)
			Label(dataSummary(), FontSize(11), TextColorVec(theme.rowDoc))
		})
		if appData.tab == 0 {
			LargeText(appData.out.String(), TextColorVec(theme.outputText))
		} else {
			LargeText(string(appData.in), TextColorVec(theme.outputText))
		}
	})
}

func tabButton(label string, tab int) {
	active := appData.tab == tab
	Container(Attrs(Pad2(4, 12), Corners(4)), func() {
		if active {
			ModAttrs(BackgroundVec(theme.tabActiveBg))
		} else if IsHovered() {
			ModAttrs(BackgroundVec(theme.tabHoverBg))
		}
		if IsClicked() {
			appData.tab = tab
		}
		if active {
			Label(label, FontSize(12), TextColorVec(theme.tabActiveText))
		} else {
			Label(label, FontSize(12), TextColorVec(theme.tabIdleText))
		}
	})
}

func StatusBar() {
	Container(Attrs(Row, Expand, CrossMid, Gap(8), Pad2(6, 14)), func() {
		if appData.status != "" {
			if appData.isErr {
				Label(appData.status, FontSize(12), TextColorVec(theme.statusErr))
			} else {
				Label(appData.status, FontSize(12), TextColorVec(theme.statusOK))
			}
		} else {
			Label("Enter selects · ↑↓ navigate · Enter applies · Tab toggles filter/list · Esc clears the filter · "+quitHint(), FontSize(11), TextColorVec(theme.statusHint))
		}
		Filler(1)
		Label(appData.out.StackString(), FontSize(11), TextColorVec(theme.statusInfo))
	})
}
