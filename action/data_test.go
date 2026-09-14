package action

import (
	"strings"
	"testing"
)

func TestAction_Preview(t *testing.T) {
	tests := []struct {
		name     string
		data     *Data
		maxRunes int
		want     string
	}{
		{"short text", NewDataText([]byte("hello")), 10, "hello"},
		{"text multiline collapsed", NewDataText([]byte("hello\nworld\r\nagain")), 20, "hello world again"},
		{"long text truncated", NewDataText([]byte(strings.Repeat("a", 500))), 10, "aaaaaaaaaa…"},
		{"long single line utf8", NewDataText([]byte(strings.Repeat("é", 500))), 5, "ééééé…"},
		{
			"dict pretty json collapsed on one line",
			NewDataDict(map[string]any{"type": "FeatureCollection", "features": []any{"a", "b"}}),
			60,
			`{ "features": [ "a", "b" ], "type": "FeatureCollection" }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.data.Preview(tt.maxRunes)
			if got != tt.want {
				t.Errorf("Preview = %q, want %q", got, tt.want)
			}
		})
	}

	// a preview must stay short whatever the entry size
	big := NewDataText([]byte(strings.Repeat("word ", 100000)))
	if got := big.Preview(120); len(got) > 200 {
		t.Errorf("Preview length = %d, want <= 200", len(got))
	}
}
