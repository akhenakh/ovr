package main

import (
	"os"
	"path/filepath"

	. "go.hasen.dev/shirei"
	. "go.hasen.dev/shirei/widgets"

	"github.com/akhenakh/ovr/action"
	"gopkg.in/yaml.v3"
)

// uiTheme carries every color role used by the ovrui views.
type uiTheme struct {
	name       string // config key
	label      string // action title
	actionName string // registry name

	rootBg      Vec4
	headerBg    Vec4
	headerTitle Vec4
	headerSub   Vec4
	chipBg      Vec4
	chipText    Vec4

	fieldBg     Vec4
	fieldBorder Vec4
	fieldFocus  Vec4
	fieldText   Vec4
	fieldCaret  Vec4
	fieldHint   Vec4

	paneBg          Vec4
	paneBorder      Vec4
	paneFocusBg     Vec4
	paneFocusBorder Vec4
	paneTitle       Vec4
	emptyText       Vec4

	rowTitle Vec4
	rowDoc   Vec4
	selRowBg Vec4
	selTitle Vec4
	selDoc   Vec4
	hoverRow Vec4

	tabActiveBg   Vec4
	tabActiveText Vec4
	tabIdleText   Vec4
	tabHoverBg    Vec4
	outputText    Vec4

	statusOK   Vec4
	statusErr  Vec4
	statusHint Vec4
	statusInfo Vec4

	paramsLabel Vec4
	paramsDoc   Vec4

	buttonAccent Vec4
	floatBg      Vec4
}

func mkTheme(name, label, actionName string, t uiTheme) *uiTheme {
	t.name = name
	t.label = label
	t.actionName = actionName
	return &t
}

// themes holds every available theme; the first is the pre-solarized default.
var themes = []*uiTheme{
	mkTheme("default", "Default theme", "default theme", uiTheme{
		rootBg:      Vec4{220, 10, 97, 1},
		headerBg:    Vec4{220, 25, 18, 1},
		headerTitle: Vec4{0, 0, 100, 1},
		headerSub:   Vec4{0, 0, 78, 1},
		chipBg:      Vec4{220, 30, 32, 1},
		chipText:    Vec4{0, 0, 92, 1},

		fieldBg:     Vec4{0, 0, 100, 1},
		fieldBorder: Vec4{0, 0, 0, 0.16},
		fieldFocus:  Vec4{0, 0, 0, 0.30},
		fieldText:   Vec4{0, 0, 0, 1},
		fieldCaret:  Vec4{0, 0, 30, 1},
		fieldHint:   Vec4{0, 0, 0, 0.4},

		paneBg:          Vec4{220, 12, 94, 1},
		paneBorder:      Vec4{220, 12, 85, 1},
		paneFocusBg:     Vec4{210, 35, 95, 1},
		paneFocusBorder: Vec4{210, 60, 55, 1},
		paneTitle:       Vec4{0, 0, 30, 1},
		emptyText:       Vec4{0, 0, 55, 1},

		rowTitle: Vec4{0, 0, 18, 1},
		rowDoc:   Vec4{0, 0, 42, 1},
		selRowBg: Vec4{210, 65, 86, 1},
		selTitle: Vec4{0, 0, 18, 1},
		selDoc:   Vec4{0, 0, 42, 1},
		hoverRow: Vec4{220, 18, 90, 1},

		tabActiveBg:   Vec4{210, 60, 45, 1},
		tabActiveText: Vec4{0, 0, 100, 1},
		tabIdleText:   Vec4{0, 0, 30, 1},
		tabHoverBg:    Vec4{220, 18, 88, 1},
		outputText:    Vec4{0, 0, 10, 1},

		statusOK:   Vec4{150, 60, 32, 1},
		statusErr:  Vec4{0, 70, 45, 1},
		statusHint: Vec4{0, 0, 55, 1},
		statusInfo: Vec4{0, 0, 45, 1},

		paramsLabel: Vec4{0, 0, 30, 1},
		paramsDoc:   Vec4{0, 0, 45, 1},

		buttonAccent: Vec4{214, 20, 90, 1},
		floatBg:      Vec4{220, 16, 98, 1},
	}),
	mkTheme("solarized-dark", "Solarized dark theme", "solarized dark theme", uiTheme{
		// solarized dark palette
		rootBg:      Vec4{192, 100, 11, 1}, // base03
		headerBg:    Vec4{192, 81, 14, 1},  // base02
		headerTitle: Vec4{68, 100, 30, 1},  // green
		headerSub:   Vec4{186, 13, 55, 1},  // base0
		chipBg:      Vec4{192, 100, 11, 1}, // base03
		chipText:    Vec4{175, 61, 40, 1},  // cyan

		fieldBg:     Vec4{192, 81, 14, 1},   // base02
		fieldBorder: Vec4{187, 14, 40, 0.6}, // base01
		fieldFocus:  Vec4{205, 69, 49, 1},   // blue
		fieldText:   Vec4{180, 8, 61, 1},    // base1
		fieldCaret:  Vec4{180, 8, 61, 1},    // base1
		fieldHint:   Vec4{187, 14, 55, 0.7}, // base01

		paneBg:          Vec4{192, 81, 14, 1}, // base02
		paneBorder:      Vec4{187, 14, 40, 1}, // base01
		paneFocusBg:     Vec4{192, 50, 17, 1}, // base02 tinted
		paneFocusBorder: Vec4{205, 69, 49, 1}, // blue
		paneTitle:       Vec4{180, 8, 61, 1},  // base1
		emptyText:       Vec4{194, 14, 45, 1}, // base00

		rowTitle: Vec4{180, 8, 61, 1},  // base1
		rowDoc:   Vec4{186, 13, 55, 1}, // base0
		selRowBg: Vec4{205, 55, 32, 1}, // dim blue
		selTitle: Vec4{44, 45, 88, 1},  // base2
		selDoc:   Vec4{44, 45, 80, 1},  // base2 dimmed
		hoverRow: Vec4{192, 50, 17, 1}, // base02 tinted

		tabActiveBg:   Vec4{205, 55, 32, 1}, // dim blue
		tabActiveText: Vec4{44, 45, 88, 1},  // base2
		tabIdleText:   Vec4{186, 13, 55, 1}, // base0
		tabHoverBg:    Vec4{192, 50, 17, 1}, // base02 tinted
		outputText:    Vec4{180, 8, 61, 1},  // base1

		statusOK:   Vec4{68, 100, 30, 1}, // green
		statusErr:  Vec4{1, 67, 52, 1},   // red
		statusHint: Vec4{194, 14, 45, 1}, // base00
		statusInfo: Vec4{186, 13, 55, 1}, // base0

		paramsLabel: Vec4{45, 100, 36, 1}, // yellow
		paramsDoc:   Vec4{186, 13, 55, 1}, // base0

		buttonAccent: Vec4{192, 81, 14, 1}, // base02
		floatBg:      Vec4{192, 81, 14, 1}, // base02
	}),
	mkTheme("solarized-light", "Solarized light theme", "solarized light theme", uiTheme{
		// solarized light palette
		rootBg:      Vec4{44, 87, 94, 1},  // base3
		headerBg:    Vec4{44, 46, 88, 1},  // base2
		headerTitle: Vec4{68, 100, 25, 1}, // green, darkened
		headerSub:   Vec4{194, 14, 45, 1}, // base00
		chipBg:      Vec4{44, 87, 94, 1},  // base3
		chipText:    Vec4{175, 61, 32, 1}, // cyan, darkened

		fieldBg:     Vec4{44, 87, 94, 1},    // base3
		fieldBorder: Vec4{187, 14, 40, 0.4}, // base01
		fieldFocus:  Vec4{205, 69, 42, 1},   // blue, darkened
		fieldText:   Vec4{187, 14, 40, 1},   // base01
		fieldCaret:  Vec4{187, 14, 40, 1},   // base01
		fieldHint:   Vec4{194, 14, 45, 0.6}, // base00

		paneBg:          Vec4{44, 46, 88, 1},    // base2
		paneBorder:      Vec4{187, 14, 40, 0.4}, // base01
		paneFocusBg:     Vec4{44, 60, 84, 1},    // base2 tinted
		paneFocusBorder: Vec4{205, 69, 42, 1},   // blue, darkened
		paneTitle:       Vec4{187, 14, 40, 1},   // base01
		emptyText:       Vec4{194, 14, 45, 1},   // base00

		rowTitle: Vec4{187, 14, 40, 1}, // base01
		rowDoc:   Vec4{194, 14, 45, 1}, // base00
		selRowBg: Vec4{205, 60, 78, 1}, // light blue
		selTitle: Vec4{187, 14, 30, 1}, // base01, darkened
		selDoc:   Vec4{194, 14, 35, 1}, // base00, darkened
		hoverRow: Vec4{44, 40, 82, 1},  // base2 tinted

		tabActiveBg:   Vec4{205, 60, 78, 1}, // light blue
		tabActiveText: Vec4{187, 14, 30, 1}, // base01, darkened
		tabIdleText:   Vec4{194, 14, 45, 1}, // base00
		tabHoverBg:    Vec4{44, 40, 82, 1},  // base2 tinted
		outputText:    Vec4{187, 14, 40, 1}, // base01

		statusOK:   Vec4{68, 100, 25, 1}, // green, darkened
		statusErr:  Vec4{1, 70, 40, 1},   // red, darkened
		statusHint: Vec4{194, 14, 40, 1}, // base00
		statusInfo: Vec4{194, 14, 45, 1}, // base00

		paramsLabel: Vec4{45, 100, 22, 1}, // yellow, darkened
		paramsDoc:   Vec4{194, 14, 45, 1}, // base00

		buttonAccent: Vec4{44, 46, 85, 1}, // base2, slightly darker
		floatBg:      Vec4{44, 87, 94, 1}, // base3
	}),
}

// theme is the active theme
var theme = themes[1]

func init() {
	// apply the stock-widget chrome for the default theme; loadConfig may
	// override it later
	ButtonAccent = theme.buttonAccent
	DefaultBackground = theme.floatBg
}

// registerThemeActions adds one action per theme so switching is available
// from the actions list for any data format.
func registerThemeActions(r *action.ActionRegistry) {
	for _, t := range themes {
		theme := t
		r.RegisterActionForAllFormats(action.NewCallbackAction(
			[]string{theme.actionName},
			"switch the UI theme",
			func() error {
				setTheme(theme)
				return nil
			},
		))
	}
}

func themeByName(name string) *uiTheme {
	for _, t := range themes {
		if t.name == name {
			return t
		}
	}
	return nil
}

// setTheme activates a theme and persists it in the config file.
func setTheme(t *uiTheme) {
	theme = t
	// retheme stock widgets (buttons, floating panels)
	ButtonAccent = t.buttonAccent
	DefaultBackground = t.floatBg
	saveConfig()
}

// appConfig is the persisted user configuration.
type appConfig struct {
	Theme string `yaml:"theme"`
}

// configDirOverride redirects the config file (used by tests)
var configDirOverride string

func configPath() (string, error) {
	dir := configDirOverride
	if dir == "" {
		u, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		dir = u
	}
	return filepath.Join(dir, "ovrui", "config.yaml"), nil
}

func loadConfig() {
	path, err := configPath()
	if err != nil {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return // no config yet: keep the default theme
	}
	var cfg appConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return
	}
	if t := themeByName(cfg.Theme); t != nil {
		theme = t
		ButtonAccent = t.buttonAccent
		DefaultBackground = t.floatBg
	}
}

func saveConfig() {
	path, err := configPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	b, err := yaml.Marshal(appConfig{Theme: theme.name})
	if err != nil {
		return
	}
	_ = os.WriteFile(path, b, 0o644)
}
