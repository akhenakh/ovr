package main

import (
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
