package main

import (
	"strings"
	"testing"

	"github.com/akhenakh/ovr/action"
)

func TestChainLabel(t *testing.T) {
	setInput([]byte("a b c"))

	if got := chainLabel(findAction(t, action.TextFormat, "upper")); got != "Upper" {
		t.Fatalf("label is %q, want Upper", got)
	}

	split := findAction(t, action.TextFormat, "split")
	if err := split.SetInputParameters(" "); err != nil {
		t.Fatalf("set split parameter: %v", err)
	}
	if got := chainLabel(split); got != `Split(" ")` {
		t.Fatalf("label is %q, want Split(\" \")", got)
	}
}

func TestChainLabelTruncatesLongParameters(t *testing.T) {
	setInput([]byte("hello"))
	edit := findAction(t, action.TextFormat, "edit")
	if err := edit.SetInputParameters(strings.Repeat("x", 400)); err != nil {
		t.Fatalf("set edit parameter: %v", err)
	}

	got := chainLabel(edit)
	if !strings.HasPrefix(got, `Edit("`) {
		t.Fatalf("label is %q, want an Edit bubble", got)
	}
	if !strings.HasSuffix(got, `…")`) {
		t.Fatalf("label is %q, want a truncated parameter", got)
	}
	if len(got) > chainParamMax+20 {
		t.Fatalf("label is %d bytes, want it capped", len(got))
	}
}

func TestChainTracksApplyAndUndo(t *testing.T) {
	setInput([]byte("a b c"))

	split := findAction(t, action.TextFormat, "split")
	selectAction(split)
	appData.paramBuf[0] = " "
	if !applyAction(split) {
		t.Fatalf("split failed: %s", appData.status)
	}

	count := findAction(t, action.TextListFormat, "count")
	if !applyAction(count) {
		t.Fatalf("count failed: %s", appData.status)
	}
	if got := len(appData.out.Stack); got != 2 {
		t.Fatalf("stack has %d entries, want 2", got)
	}

	undoAction()
	if got := len(appData.out.Stack); got != 1 {
		t.Fatalf("stack has %d entries after undo, want 1", got)
	}
	if got := chainLabel(appData.out.Stack[0]); got != `Split(" ")` {
		t.Fatalf("remaining bubble is %q, want Split(\" \")", got)
	}
}
