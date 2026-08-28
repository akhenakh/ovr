package main

import (
	"strings"
	"testing"

	"github.com/akhenakh/ovr/action"
)

func findAction(t *testing.T, format action.Format, name string) action.Action {
	t.Helper()
	a, ok := appData.registry.ActionByName(format, name)
	if !ok {
		t.Fatalf("action %s not found", name)
	}
	return a
}

func TestSetInputText(t *testing.T) {
	setInput([]byte("hello"))
	if appData.out.Format != action.TextFormat {
		t.Fatalf("format is %s, want text", appData.out.Format.Name)
	}
	if len(appData.actions) == 0 {
		t.Fatal("no actions for text input")
	}
}

func TestApplyUpperAndUndo(t *testing.T) {
	setInput([]byte("hello"))

	upper := findAction(t, action.TextFormat, "upper")
	applyAction(upper)
	if got := appData.out.String(); got != "HELLO" {
		t.Fatalf("output is %q, want HELLO", got)
	}

	undoAction()
	if got := appData.out.String(); got != "hello" {
		t.Fatalf("output is %q, want hello after undo", got)
	}
}

func TestApplyParams(t *testing.T) {
	setInput([]byte("a,b,c"))

	join := findAction(t, action.TextListFormat, "join")
	selectAction(join)
	if len(appData.paramBuf) != 1 {
		t.Fatalf("paramBuf len is %d, want 1", len(appData.paramBuf))
	}
	appData.paramBuf[0] = "-"

	// applying without a list first must fail cleanly
	applyAction(join)
	if !appData.isErr {
		t.Fatal("expected an error applying join to text")
	}

	comma := findAction(t, action.TextFormat, "comma")
	applyAction(comma)
	selectAction(join) // applying cleared the selection; select again like the UI flow
	appData.paramBuf[0] = "-"
	applyAction(join)
	if got := appData.out.String(); got != "a-b-c" {
		t.Fatalf("output is %q, want a-b-c", got)
	}
	if appData.selected != nil {
		t.Fatal("selection should be cleared after a successful apply")
	}
}

func TestFilterActions(t *testing.T) {
	setInput([]byte("hello"))

	appData.search = "jso"
	list := filteredActions()
	if len(list) == 0 {
		t.Fatal("filter jso returned no actions")
	}
	for _, a := range list {
		if !strings.Contains(strings.ToLower(a.Title()+a.Doc()), "jso") {
			t.Fatalf("action %s does not match filter jso", a.Title())
		}
	}

	appData.search = "zzzz"
	if list := filteredActions(); len(list) != 0 {
		t.Fatalf("filter zzzz returned %d actions", len(list))
	}
	appData.search = ""
}

func TestCopyOutputRequestsClipboard(t *testing.T) {
	setInput([]byte("hello"))
	copyOutput()
	if appData.isErr {
		t.Fatal("copyOutput set an error")
	}
}
