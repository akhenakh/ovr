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
