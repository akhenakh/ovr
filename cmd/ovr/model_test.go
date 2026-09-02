package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/akhenakh/ovr/action"
)

func TestRegistryItems(t *testing.T) {
	r := action.DefaultRegistry()
	actions := r.ActionsForText("")
	if len(actions) == 0 {
		t.Fatal("ActionsForText returned no actions")
	}

	m := newModel([]byte("hello"))
	if got := len(m.list.Items()); got != len(actions) {
		t.Fatalf("list has %d items, want %d", got, len(actions))
	}
}

// Regression test: actions must satisfy list.DefaultItem (Title, Description
// and FilterValue) or the list delegate silently renders nothing.
func TestActionsRenderInList(t *testing.T) {
	m := newModel([]byte("hello"))

	newM, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mm := newM.(model)

	v := mm.View()
	if !strings.Contains(v.Content, "Base64") {
		t.Fatal("no actions rendered in the view")
	}
}

// The edit action must be offered for the initial data and applying the
// editor result must store it as the new output.
func TestEditActionApplied(t *testing.T) {
	m := newModel([]byte("hello"))

	var found bool
	for _, it := range m.list.Items() {
		if a, ok := it.(action.Action); ok && a.Names()[0] == "edit" {
			found = true
		}
	}
	if !found {
		t.Fatal("edit action is not in the list")
	}

	ea, ok := m.editActionForData()
	if !ok {
		t.Fatal("no edit action for text data")
	}

	newM, _ := m.Update(editorFinishedMsg{a: ea, edited: []byte("edited text")})
	mm := newM.(model)
	if string(mm.out.RawValue) != "edited text" {
		t.Fatalf("out = %q, want %q", mm.out.RawValue, "edited text")
	}
	if !strings.Contains(mm.list.Title, "edited text") {
		t.Fatalf("title = %q, want it to contain the edited content", mm.list.Title)
	}
}
