package main

import (
	"regexp"
	"strings"
	"testing"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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

// The short help at the bottom must show the v key with the other keys.
func TestShortHelpShowsViewerKey(t *testing.T) {
	m := newModel([]byte("hello"))

	newM, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	content := newM.(model).View().Content

	// the help is colorized, strip the ANSI escape sequences
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	plain := ansi.ReplaceAllString(content, "")

	if !strings.Contains(plain, "v details") {
		t.Fatal("short help does not show the v key")
	}
	if !strings.Contains(plain, "backspace undo") {
		t.Fatal("short help does not show the backspace key")
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

// Regression test: pressing v must open the detail viewer with the entry
// content, sized to the full terminal.
func TestDetailViewer(t *testing.T) {
	m := newModel([]byte(strings.Repeat("hello line\n", 100)))

	newM, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mm := newM.(model)

	newM, _ = mm.Update(tea.KeyPressMsg{Code: 'v'})
	mm = newM.(model)

	if mm.state != detailState {
		t.Fatal("v did not switch to the detail state")
	}
	if mm.viewport.Width() != 80 {
		t.Fatalf("viewport width = %d, want 80", mm.viewport.Width())
	}
	if h := mm.viewport.Height(); h != 24-lipgloss.Height(mm.headerView())-lipgloss.Height(mm.footerView()) {
		t.Fatalf("viewport height = %d, want %d", h, 24-lipgloss.Height(mm.headerView())-lipgloss.Height(mm.footerView()))
	}
	if !strings.Contains(mm.viewport.View(), "hello line") {
		t.Fatal("the viewer is empty")
	}
	if !strings.Contains(mm.View().Content, "text:") {
		t.Fatal("the viewer header does not show the format name")
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
