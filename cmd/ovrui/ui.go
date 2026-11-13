package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/akhenakh/ovr/action"

	. "go.hasen.dev/shirei" //nolint:staticcheck // shirei's DSL is dot-imported by design
	. "go.hasen.dev/shirei/widgets" //nolint:staticcheck // shirei's DSL is dot-imported by design
)

type f32 = float32

// keySlash is the physical "/" key: KeyCode carries US-QWERTY legends, so the
// ASCII code is the identity even though shirei only names letters.
const keySlash KeyCode = '/'

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
	// handle of the right-pane document editor, captured each frame
	paneEditorId ContainerId
	// set to re-assert focus on the filter field during the next build pass
	refocusFilter bool
	// paste popup: shown at startup when no input came from a file, stdin,
	// or the clipboard, and reopenable from the toolbar
	pasteOpen bool
	pasteBuf  string
	// read-only documents backing the right pane (output and input tabs);
	// documents are app-owned retained state, never rebuilt per frame
	outDoc TextDocument
	inDoc  TextDocument
	// edit action session: the in-app document editor stands in for $EDITOR
	editOpen   bool
	editAction action.Action
	editDoc    TextDocument
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

// editorStyle paints a document editor from the theme while keeping the
// stock selection chrome.
func editorStyle(bg, text, caret Vec4) TextInputStyle {
	s := CurrentColorScheme.TextInput
	s.Background = bg
	s.Text = text
	s.Caret = caret
	s.Selection = theme.selRowBg
	s.Border = theme.fieldBorder
	s.HoveredBorder = theme.fieldFocus
	return s
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

	// if the input looks like a JSON object, start with dict data so the
	// gjson actions are offered right away
	if action.GuessDict(in) {
		if a, ok := r.ActionByName(action.TextFormat, "dict"); ok {
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
	// the actions list is the home position when input loads
	appData.listFocus = true
	appData.refocusFilter = false
	appData.editOpen = false
	appData.editAction = nil
	syncDocs()
	refreshActions()
}

// syncDocs refreshes both read-only pane documents from the current data.
func syncDocs() {
	appData.outDoc.SetText(appData.out.String())
	appData.inDoc.SetText(string(appData.in))
}

// syncOutputDoc refreshes the output document after a transformation.
func syncOutputDoc() {
	appData.outDoc.SetText(appData.out.String())
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

// focusList gives the keyboard to the actions list. The list is the home
// position: startup, after applying, cancelling or clearing the filter.
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

// focusFilter gives the keyboard to the filter field, the "/" and Tab target.
func focusFilter() {
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
	// applying from the list resets the filter and keeps the keyboard there
	appData.search = ""
	focusList()
	return true
}

func applyAction(a action.Action) bool {
	// editor actions need the in-app document editor instead of a parameter
	// field; the action is applied once the user saves the edited content
	if ed, ok := a.(action.Editor); ok {
		startEditAction(a, ed)
		return false
	}
	if params := a.Parameters(); len(params) > 0 {
		if len(appData.paramBuf) != len(params) {
			setStatus(fmt.Sprintf("%s: missing parameters", a.Title()), true)
			return false
		}
		values := make([]any, len(params))
		for i, p := range params {
			// string parameters are literal (a space is a valid split
			// separator); only numbers ignore surrounding blanks
			v := appData.paramBuf[i]
			switch p.ActionParameterType {
			case action.IntParameter:
				n, err := strconv.Atoi(strings.TrimSpace(v))
				if err != nil {
					setStatus(fmt.Sprintf("%s: %s needs an integer", a.Title(), p.Doc), true)
					return false
				}
				values[i] = n
			case action.FloatParameter:
				f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
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

	return applyTransform(a)
}

// applyTransform runs the action on the current data and updates the app state.
func applyTransform(a action.Action) bool {
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
	syncOutputDoc()
	setStatus("applied "+a.Title(), false)
	return true
}

// startEditAction opens the in-app document editor for an edit action,
// replacing the external $EDITOR the TUI launches.
func startEditAction(a action.Action, ed action.Editor) {
	appData.editAction = a
	appData.editDoc.SetText(ed.EditContent(appData.out))
	appData.editOpen = true
	appData.listFocus = false
	setStatus("editing output, save to apply "+a.Title(), false)
}

// cancelEdit closes the editor and drops the edits.
func cancelEdit() {
	appData.editOpen = false
	appData.editAction = nil
	// the modal took the keyboard; hand it back to the actions list
	focusList()
	setStatus("edit cancelled", false)
}

// commitEdit feeds the edited content back to the action as its input
// parameter, then applies it so the action stays undoable by replay.
func commitEdit() {
	a := appData.editAction
	if a == nil {
		cancelEdit()
		return
	}
	content := appData.editDoc.Text()
	appData.editOpen = false
	appData.editAction = nil
	// the modal took the keyboard; hand it back to the actions list
	focusList()
	if err := a.SetInputParameters(content); err != nil {
		setStatus(err.Error(), true)
		return
	}
	applyTransform(a)
}

// undoLast removes the last applied action when there is one.
func undoLast() {
	if len(appData.out.Stack) == 0 {
		return
	}
	undoAction()
}

func undoAction() {
	nd, oa, err := appData.out.Undo(appData.in)
	if err != nil {
		setStatus(err.Error(), true)
		return
	}
	appData.out = nd
	refreshActions()
	syncOutputDoc()
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

// openPaste shows the paste popup with a fresh buffer.
func openPaste() {
	appData.pasteBuf = ""
	appData.pasteOpen = true
}

// closePaste dismisses the paste popup without loading anything.
func closePaste() {
	appData.pasteBuf = ""
	appData.pasteOpen = false
}

// loadPaste turns the popup buffer into the app input.
func loadPaste() {
	if strings.TrimSpace(appData.pasteBuf) == "" {
		setStatus("nothing to load, paste some text first", true)
		return
	}
	in := []byte(appData.pasteBuf)
	closePaste()
	setInput(in)
	setStatus(fmt.Sprintf("input loaded (%d bytes)", len(in)), false)
}

// themedModal renders a modal card painted from the active theme instead of
// shirei's default light card.
func themedModal(width f32, dismiss func(), fn func()) {
	st := DefaultModalStyle()
	st.Background = theme.paneBg
	st.Text = theme.outputText
	ModalStyled(width, dismiss, st, fn)
}

// pasteModal renders the paste-input dialog over the app.
func pasteModal() {
	themedModal(640, closePaste, func() {
		Label("Paste input", FontSize(16), FontWeight(WeightBold), TextColorVec(theme.paneTitle))
		themedTextInput(&appData.pasteBuf, TextInputAttrs{
			Placeholder: "paste text here…",
			Rows:        12,
			Wrap:        true,
		})
		Container(Attrs(Row, Expand, CrossMid, Gap(8)), func() {
			Filler(1)
			if Button(NoIcon, "Cancel") {
				closePaste()
			}
			if Button(SymDownload, "Load") {
				loadPaste()
			}
		})
	})
}

// editModal renders the document editor standing in for $EDITOR over the app.
func editModal() {
	themedModal(920, cancelEdit, func() {
		Container(Attrs(Row, Expand, CrossMid, Gap(10)), func() {
			Label("Edit output", FontSize(16), FontWeight(WeightBold), TextColorVec(theme.paneTitle))
			Filler(1)
			Label(saveHint(), FontSize(11), TextColorVec(theme.rowDoc))
		})
		Container(Attrs(Expand, FixHeight(470), Clip, Corners(4), BorderWidth(1), BorderColorVec(theme.fieldBorder)), func() {
			TextEditorStyled(&appData.editDoc, TextEditorAttrs{
				TextStyle:   TextStyle(Fonts(Monospace...), FontSize(13)),
				LineNumbers: true,
			}, editorStyle(theme.fieldBg, theme.fieldText, theme.fieldCaret), CurrentColorScheme.ScrollBar)
		})
		Container(Attrs(Row, Expand, CrossMid, Gap(8)), func() {
			Filler(1)
			if Button(SymCancel, "Cancel") {
				cancelEdit()
			}
			if Button(SymBoxTick, "Save & Apply") {
				commitEdit()
			}
		})
	})
}

func saveHint() string {
	if PrimaryMod() == ModCmd {
		return "Escape cancels · Cmd+S saves"
	}
	return "Escape cancels · Ctrl+S saves"
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
	// popups own the keyboard while open: the paste field and the document
	// editor handle their own keys, so the list shortcuts must not run
	if appData.editOpen {
		if GetFrameInput().Key == KeyS && ActiveCombo() == Combo(KeyS, PrimaryMod()) {
			commitEdit()
		}
	} else if !appData.pasteOpen {
		switch GetFrameInput().Key {
		case KeyEnter:
			if appData.listFocus {
				// on the list, Enter picks the first match or applies the
				// current selection
				if appData.selected == nil {
					selectFirst()
				} else {
					applySelected()
				}
			} else {
				// Enter in the filter field activates the actions list
				activateList()
			}
		case KeyTab:
			// Tab toggles between the filter field and the actions list
			if appData.listFocus {
				focusFilter()
			} else {
				focusList()
			}
		case KeyDown, KeyUp:
			if appData.listFocus {
				moveSelection(GetFrameInput().Key == KeyDown)
			}
		case KeyDeleteBackward:
			// Backspace on the list removes the last applied action, like
			// the Undo button; in the filter it keeps editing the text
			if appData.listFocus {
				undoLast()
			}
		case keySlash:
			// "/" jumps from the list to the filter, like a search prompt
			if appData.listFocus {
				focusFilter()
			}
		case KeyEscape:
			appData.search = ""
			focusList()
		case KeyQ:
			if ActiveCombo() == Combo(KeyQ, PrimaryMod()) {
				quitApp(0)
			}
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
		ChainBar()
		StatusBar()
	})

	if appData.editOpen {
		editModal()
	}
	if appData.pasteOpen {
		pasteModal()
	}
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
		themedTextInput(&appData.search, TextInputAttrs{
			Placeholder: "filter actions…",
			MaxWidth:    340,
			MaxLines:    1,
			// the actions list owns the keyboard at startup; focus reaches
			// the field through "/", Tab or a click
			NoAutoFocus: true,
		})
		appData.filterId = GetLastId()
		// clicking into the field returns keyboard focus mode to the filter
		if IdReceivedFocusNow(appData.filterId) {
			appData.listFocus = false
		}
		if appData.refocusFilter {
			appData.refocusFilter = false
			FocusImmediateOn(appData.filterId)
			// focus moved programmatically: show the keyboard focus indicator
			ShowFocusIndicator()
		}
		if Button(SymRefresh, "Reload") {
			reloadClipboard()
		}
		if Button(SymDownload, "Paste") {
			openPaste()
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
	// editor actions collect their content in the document editor modal
	if _, ok := a.(action.Editor); ok {
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
			// clicking a row takes the keyboard back from the filter field
			focusList()
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
			TextEditorStyled(&appData.outDoc, readOnlyEditorAttrs(),
				editorStyle(theme.paneBg, theme.outputText, theme.fieldCaret), CurrentColorScheme.ScrollBar)
		} else {
			TextEditorStyled(&appData.inDoc, readOnlyEditorAttrs(),
				editorStyle(theme.paneBg, theme.outputText, theme.fieldCaret), CurrentColorScheme.ScrollBar)
		}
		// selecting text in the pane moves the keyboard there: list shortcuts
		// must not also react to the same arrow keys
		appData.paneEditorId = GetLastId()
		if IdReceivedFocusNow(appData.paneEditorId) {
			appData.listFocus = false
		}
	})
}

// readOnlyEditorAttrs configures the right-pane document view: selectable and
// copyable, but never editable, and it must not steal focus from the filter.
func readOnlyEditorAttrs() TextEditorAttrs {
	return TextEditorAttrs{
		TextStyle:   TextStyle(Fonts(Monospace...), FontSize(12)),
		LineNumbers: true,
		ReadOnly:    true,
		NoAutoFocus: true,
	}
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

// ChainBar shows the actions applied so far as blue bubbles, oldest first.
// It keeps a fixed height so applying or undoing does not resize the panes.
func ChainBar() {
	Container(Attrs(Row, Expand, CrossMid, Gap(6), Pad2(4, 12), FixHeight(32), Clip), func() {
		stack := appData.out.Stack
		if len(stack) == 0 {
			Label("no transformations applied", FontSize(11), TextColorVec(theme.emptyText))
			return
		}
		for _, a := range stack {
			chainBubble(a)
		}
	})
}

func chainBubble(a action.Action) {
	Container(Attrs(Row, CrossMid, Pad2(3, 10), Corners(10), Clip, BackgroundVec(theme.tabActiveBg)), func() {
		Label(chainLabel(a), FontSize(11), TextColorVec(theme.tabActiveText))
	})
}

// chainParamMax caps how much of a parameter a bubble shows: edit stores the
// whole edited text as a parameter.
const chainParamMax = 32

// chainLabel renders one stack entry as "Title(param, param)".
func chainLabel(a action.Action) string {
	label := a.Title()
	params := a.InputParameters()
	if len(params) == 0 {
		return label
	}
	parts := make([]string, len(params))
	for i, p := range params {
		if s, ok := p.(string); ok {
			parts[i] = strconv.Quote(truncateBytes(s, chainParamMax))
		} else {
			parts[i] = truncateBytes(fmt.Sprint(p), chainParamMax)
		}
	}
	return label + "(" + strings.Join(parts, ", ") + ")"
}

// truncateBytes shortens s to at most n bytes on a rune boundary.
func truncateBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…"
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
			Label("↑↓ navigate · Enter selects/applies · / or Tab filters · Backspace undoes · Esc clears the filter · "+quitHint(), FontSize(11), TextColorVec(theme.statusHint))
		}
	})
}
