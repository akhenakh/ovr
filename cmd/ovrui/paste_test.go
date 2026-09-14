package main

import (
	"testing"

	. "go.hasen.dev/shirei"
)

// pasteFrame drives one frame of RootView with the paste popup state active
// and the mouse parked offscreen.
func pasteFrame(key KeyCode) {
	GetHost().WindowSize = Vec2{1100, 760}
	GetInputState().MousePoint = Vec2{-1000, -1000}
	GetFrameInput().Mouse = 0
	GetFrameInput().Key = key
	GetFrameInput().Text = ""
	RunFrameFn(RootView)
}

func resetPasteState() {
	appData.pasteOpen = false
	appData.pasteBuf = ""
	appData.search = ""
	appData.listFocus = false
	appData.selected = nil
}

func TestPastePopupEscapeDismisses(t *testing.T) {
	resetPasteState()
	t.Cleanup(resetPasteState)

	setInput([]byte("hello"))
	openPaste()

	pasteFrame(0)
	if !appData.pasteOpen {
		t.Fatal("paste popup closed on its own")
	}

	pasteFrame(KeyEscape)
	if appData.pasteOpen {
		t.Fatal("escape did not dismiss the paste popup")
	}
}

func TestPastePopupKeysDoNotTouchList(t *testing.T) {
	resetPasteState()
	t.Cleanup(resetPasteState)

	setInput([]byte("hello"))
	openPaste()
	pasteFrame(0)
	pasteFrame(KeyEnter)

	if !appData.pasteOpen {
		t.Fatal("enter closed the paste popup")
	}
	if appData.selected != nil {
		t.Fatal("enter activated the actions list behind the popup")
	}
	if appData.listFocus {
		t.Fatal("enter moved focus to the actions list")
	}
}

func TestLoadPaste(t *testing.T) {
	resetPasteState()
	t.Cleanup(resetPasteState)

	appData.pasteOpen = true
	appData.pasteBuf = "a,b,c"
	loadPaste()

	if appData.pasteOpen {
		t.Fatal("popup stayed open after load")
	}
	if got := string(appData.in); got != "a,b,c" {
		t.Fatalf("input is %q, want the pasted text", got)
	}
	if appData.isErr {
		t.Fatalf("loadPaste set an error status: %s", appData.status)
	}
	if len(appData.actions) == 0 {
		t.Fatal("no actions offered after loading pasted text")
	}
}

func TestLoadPasteEmptyKeepsPopupOpen(t *testing.T) {
	resetPasteState()
	t.Cleanup(resetPasteState)

	appData.pasteOpen = true
	appData.pasteBuf = "   "
	loadPaste()

	if !appData.pasteOpen {
		t.Fatal("empty paste should keep the popup open")
	}
	if !appData.isErr {
		t.Fatal("empty paste should set an error status")
	}
}
