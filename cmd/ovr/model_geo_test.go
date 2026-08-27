//go:build geo

package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/akhenakh/ovr/action"
)

// Regression test: WKT input should be detected and parsed into geometry
// data so geo actions are offered right away.
func TestWKTInputStartsAsGeometry(t *testing.T) {
	m := newModel([]byte("POINT(-0.4539761 48.0930043)"))

	if m.out.Format != action.GeoFormat {
		t.Fatalf("out format is %s, want geometry", m.out.Format.Name)
	}

	newM, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	mm := newM.(model)

	v := mm.View()
	for _, want := range []string{"Geometry: POINT(-0.4539761 48.0930043)", "Centroid", "Map"} {
		if !strings.Contains(v.Content, want) {
			t.Errorf("view does not contain %q", want)
		}
	}
}

func TestTextInputStillStartsAsText(t *testing.T) {
	m := newModel([]byte("hello"))

	if m.out.Format != action.TextFormat {
		t.Fatalf("out format is %s, want text", m.out.Format.Name)
	}
}
