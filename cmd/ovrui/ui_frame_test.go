package main

import (
	"os"
	"testing"

	. "go.hasen.dev/shirei"
)

func uiFrame(mouse Vec2, mouseAction MouseAction, key KeyCode) {
	GetHost().WindowSize = Vec2{1100, 760}
	GetInputState().MousePoint = mouse
	GetFrameInput().Mouse = mouseAction
	GetFrameInput().Key = key
	GetFrameInput().Text = ""
	RunFrameFn(RootView)
}

func TestCtrlQQuits(t *testing.T) {
	quitCalled := false
	quitApp = func(int) { quitCalled = true }
	defer func() { quitApp = os.Exit }()

	setInput([]byte("hello"))

	// plain q without the primary modifier must not quit
	uiFrame(Vec2{-1000, -1000}, 0, KeyQ)
	if quitCalled {
		t.Fatal("plain Q quit the app")
	}

	// primary modifier + Q quits
	GetInputState().Modifiers = PrimaryMod()
	uiFrame(Vec2{-1000, -1000}, 0, KeyQ)
	GetInputState().Modifiers = 0
	if !quitCalled {
		t.Fatal("primary modifier + Q did not quit")
	}
}

func TestClickSelectsAndEnterApplies(t *testing.T) {
	setInput([]byte("hello"))

	uiFrame(Vec2{-1000, -1000}, 0, 0)
	uiFrame(Vec2{-1000, -1000}, 0, 0)

	// filter down to a single action, then click on the first row
	appData.search = "tobase64"
	uiFrame(Vec2{-1000, -1000}, 0, 0)
	uiFrame(Vec2{200, 140}, MouseClick, 0)
	if appData.selected == nil {
		t.Fatal("click did not select an action")
	}
	if got := appData.selected.Title(); got != "Tobase64" {
		t.Fatalf("selected %q, want Tobase64", got)
	}

	// release, then Enter applies the selected action
	uiFrame(Vec2{200, 140}, MouseRelease, 0)
	uiFrame(Vec2{-1000, -1000}, 0, KeyEnter)

	if got := appData.out.String(); got != "aGVsbG8=" {
		t.Fatalf("output is %q, want aGVsbG8=", got)
	}
}

func TestFilterEnterActivatesList(t *testing.T) {
	setInput([]byte("hello"))

	uiFrame(Vec2{-1000, -1000}, 0, 0)
	uiFrame(Vec2{-1000, -1000}, 0, 0)

	// type a filter, then Enter: the list takes focus and the first match
	// is selected
	appData.search = "tobase64"
	uiFrame(Vec2{-1000, -1000}, 0, KeyEnter)
	if !appData.listFocus {
		t.Fatal("Enter did not move focus to the actions list")
	}
	if appData.selected == nil {
		t.Fatal("Enter did not select the first matching action")
	}
	if got := appData.selected.Title(); got != "Tobase64" {
		t.Fatalf("selected %q, want Tobase64", got)
	}

	// Enter again applies the selected action, resets the filter, and
	// moves focus back to the filter field
	uiFrame(Vec2{-1000, -1000}, 0, KeyEnter)
	if got := appData.out.String(); got != "aGVsbG8=" {
		t.Fatalf("output is %q, want aGVsbG8=", got)
	}
	if appData.search != "" {
		t.Fatalf("filter is %q, want empty", appData.search)
	}
	if appData.listFocus {
		t.Fatal("focus did not return to the filter field")
	}
}

func TestTabTogglesFilterAndList(t *testing.T) {
	setInput([]byte("hello"))

	uiFrame(Vec2{-1000, -1000}, 0, 0)
	uiFrame(Vec2{-1000, -1000}, 0, 0)

	// Tab from the filter moves focus to the list without selecting anything
	uiFrame(Vec2{-1000, -1000}, 0, KeyTab)
	if !appData.listFocus {
		t.Fatal("Tab did not move focus to the actions list")
	}
	if appData.selected != nil {
		t.Fatal("Tab must not select an action")
	}

	// Tab again returns focus to the filter field
	uiFrame(Vec2{-1000, -1000}, 0, KeyTab)
	if appData.listFocus {
		t.Fatal("Tab did not move focus back to the filter field")
	}
	if !IdHasFocus(appData.filterId) {
		t.Fatal("filter field does not hold keyboard focus")
	}
}

func TestArrowsNavigateList(t *testing.T) {
	setInput([]byte("hello"))

	uiFrame(Vec2{-1000, -1000}, 0, 0)
	uiFrame(Vec2{-1000, -1000}, 0, 0)

	// Tab focuses the list, then Down selects the first action
	uiFrame(Vec2{-1000, -1000}, 0, KeyTab)
	if !appData.listFocus {
		t.Fatal("Tab did not move focus to the actions list")
	}
	if appData.selected != nil {
		t.Fatal("Tab must not select an action")
	}

	uiFrame(Vec2{-1000, -1000}, 0, KeyDown)
	list := filteredActions()
	if len(list) == 0 {
		t.Fatal("no actions to navigate")
	}
	if appData.selected != list[0] {
		t.Fatal("Down did not select the first action")
	}

	// Down again moves to the second action, Up goes back
	uiFrame(Vec2{-1000, -1000}, 0, KeyDown)
	if appData.selected != list[1] {
		t.Fatal("Down did not advance the selection")
	}
	uiFrame(Vec2{-1000, -1000}, 0, KeyUp)
	if appData.selected != list[0] {
		t.Fatal("Up did not move the selection back")
	}

	// Up from the first row wraps to the last
	uiFrame(Vec2{-1000, -1000}, 0, KeyUp)
	if appData.selected != list[len(list)-1] {
		t.Fatal("Up did not wrap to the last action")
	}
}
