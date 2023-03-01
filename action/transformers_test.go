package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction_TextTransform(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		action  string
		in      string
		want    string
		wantErr bool
	}{
		{"upper", "hello", "HELLO", false},
		{"lower", "HELLO", "hello", false},
		{"hex", "48454c4c4f", "HELLO", false},
		{"tohex", "HELLO", "48454c4c4f", false},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TextAction(tt.action, []byte(tt.in))
			if !tt.wantErr {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, string(got))
		})
	}
}
