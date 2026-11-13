package main

import (
	"testing"

	"github.com/akhenakh/ovr/action"

	. "go.hasen.dev/shirei"
)

// resetEditState clears any editor session left by a previous test.
func resetEditState() {
	appData.editOpen = false
	appData.editAction = nil
}

func TestEditActionOpensInAppEditor(t *testing.T) {
	resetEditState()
	t.Cleanup(resetEditState)

	setInput([]byte("hello"))
	edit := findAction(t, action.TextFormat, "edit")

	if applyAction(edit) {
		t.Fatal("edit action reported an immediate apply")
	}
	if !appData.editOpen {
		t.Fatal("edit action did not open the document editor")
	}
	if got := appData.editDoc.Text(); got != "hello" {
		t.Fatalf("editor content is %q, want hello", got)
	}
	if got := appData.out.String(); got != "hello" {
		t.Fatalf("output changed before saving: %q", got)
	}
	// the action needs its content before it can be replayed
	if appData.out.StackString() != "" {
		t.Fatalf("stack changed before saving: %q", appData.out.StackString())
	}
}

func TestEditSaveAppliesAndUndoRestores(t *testing.T) {
	resetEditState()
	t.Cleanup(resetEditState)

	setInput([]byte("hello"))
	edit := findAction(t, action.TextFormat, "edit")
	applyAction(edit)

	appData.editDoc.SetText("edited text")
	commitEdit()

	if appData.editOpen {
		t.Fatal("editor stayed open after save")
	}
	if got := appData.out.String(); got != "edited text" {
		t.Fatalf("output is %q, want edited text", got)
	}
	if got := appData.out.StackString(); got == "" {
		t.Fatal("applied edit did not reach the stack")
	}

	// undo replays the stack from the raw input; the edit content is stored
	// as the action input parameter, so no editor is needed again
	undoAction()
	if got := appData.out.String(); got != "hello" {
		t.Fatalf("output is %q, want hello after undo", got)
	}
	if appData.editOpen {
		t.Fatal("undo reopened the editor")
	}
}

func TestEditCancelKeepsOutput(t *testing.T) {
	resetEditState()
	t.Cleanup(resetEditState)

	setInput([]byte("hello"))
	edit := findAction(t, action.TextFormat, "edit")
	applyAction(edit)

	appData.editDoc.SetText("discarded")
	cancelEdit()

	if appData.editOpen {
		t.Fatal("editor stayed open after cancel")
	}
	if appData.editAction != nil {
		t.Fatal("cancelled action was not cleared")
	}
	if got := appData.out.String(); got != "hello" {
		t.Fatalf("output is %q, want hello after cancel", got)
	}
	if got := appData.out.StackString(); got != "" {
		t.Fatalf("cancelled edit reached the stack: %q", got)
	}
}

func TestEditSaveChord(t *testing.T) {
	resetEditState()
	t.Cleanup(resetEditState)

	setInput([]byte("hello"))
	applyAction(findAction(t, action.TextFormat, "edit"))
	appData.editDoc.SetText("chord saved")

	GetInputState().Modifiers = PrimaryMod()
	uiFrame(Vec2{-1000, -1000}, 0, KeyS)
	GetInputState().Modifiers = 0

	if appData.editOpen {
		t.Fatal("primary modifier + S did not save the editor")
	}
	if got := appData.out.String(); got != "chord saved" {
		t.Fatalf("output is %q, want chord saved", got)
	}
}

func TestEditEscapeCancels(t *testing.T) {
	resetEditState()
	t.Cleanup(resetEditState)

	setInput([]byte("hello"))
	applyAction(findAction(t, action.TextFormat, "edit"))
	appData.editDoc.SetText("discarded")

	// the first frame mounts the modal, the second delivers Escape
	uiFrame(Vec2{-1000, -1000}, 0, 0)
	uiFrame(Vec2{-1000, -1000}, 0, KeyEscape)

	if appData.editOpen {
		t.Fatal("escape did not dismiss the editor")
	}
	if got := appData.out.String(); got != "hello" {
		t.Fatalf("output is %q, want hello after cancel", got)
	}
}

func TestEditModalOwnsTheKeyboard(t *testing.T) {
	resetEditState()
	t.Cleanup(resetEditState)

	setInput([]byte("hello"))
	applyAction(findAction(t, action.TextFormat, "edit"))
	appData.search = "tobase64"

	// Enter must reach the document editor, not activate the actions list
	GetHost().WindowSize = Vec2{1100, 760}
	GetInputState().MousePoint = Vec2{-1000, -1000}
	GetFrameInput().Mouse = 0
	GetFrameInput().Text = ""
	GetFrameInput().Key = KeyEnter
	RunFrameFn(RootView)

	if !appData.editOpen {
		t.Fatal("enter closed the editor")
	}
	if appData.selected != nil {
		t.Fatal("enter applied an action behind the editor")
	}
	if appData.listFocus {
		t.Fatal("enter moved focus to the actions list behind the editor")
	}
}
