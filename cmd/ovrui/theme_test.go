package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akhenakh/ovr/action"
	"go.hasen.dev/shirei/widgets"
)

func TestThemeByName(t *testing.T) {
	if themeByName("solarized-dark") == nil {
		t.Fatal("solarized-dark theme not found")
	}
	if themeByName("default") == nil {
		t.Fatal("default theme not found")
	}
	if themeByName("nope") != nil {
		t.Fatal("unknown theme returned")
	}
	if len(themes) != 2 {
		t.Fatalf("got %d themes, want 2", len(themes))
	}
}

func findThemeAction(t *testing.T, name string) action.Action {
	t.Helper()
	a, ok := appData.registry.ActionByName(action.TextFormat, name)
	if !ok {
		t.Fatalf("theme action %q not registered", name)
	}
	return a
}

func TestThemeSwitchAction(t *testing.T) {
	// sandbox the config file
	dir := t.TempDir()
	configDirOverride = dir
	defer func() { configDirOverride = "" }()

	setInput([]byte("hello"))
	if appData.registry == nil {
		t.Fatal("registry was not created")
	}

	// switch to the default theme through its action
	def := findThemeAction(t, "default theme")
	if !applyAction(def) {
		t.Fatal("applying the default theme action failed")
	}
	if theme != themes[0] {
		t.Fatalf("active theme is %q, want default", theme.name)
	}
	if widgets.ButtonAccent != themes[0].buttonAccent {
		t.Fatal("stock widget accent was not updated")
	}

	// the choice is persisted
	if _, err := os.ReadFile(filepath.Join(dir, "ovrui", "config.yaml")); err != nil {
		t.Fatalf("config file was not written: %v", err)
	}
	loadConfig()
	if theme.name != "default" {
		t.Fatalf("loaded theme is %q, want default", theme.name)
	}

	// switch back to solarized dark
	sol := findThemeAction(t, "solarized dark theme")
	if !applyAction(sol) {
		t.Fatal("applying the solarized dark theme action failed")
	}
	if theme.name != "solarized-dark" {
		t.Fatalf("active theme is %q, want solarized-dark", theme.name)
	}

	// the data is untouched by theme actions
	if got := appData.out.String(); got != "hello" {
		t.Fatalf("output changed to %q, want hello", got)
	}
}
