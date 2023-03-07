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
		{action: "lower", in: "HELLO", want: "hello", wantErr: false},
		{action: "title", in: "HELLO", want: "Hello", wantErr: false},
		{"hex", "48454c4c4f", "HELLO", false},
		{action: "hex", in: "gg", want: "HELLO", wantErr: true},
		{"tohex", "HELLO", "48454c4c4f", false},
		{action: "base64", in: "aGVsbG8=", want: "hello", wantErr: false},
		{action: "tobase64", in: "hello", want: "aGVsbG8=", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TextAction(tt.action, []byte(tt.in))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, string(got))
			}
		})
	}
}

func TestAction_TextTimeTransform(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		action  string
		in      string
		want    string
		wantErr bool
	}{
		{"jsondate", "2012-04-23T18:25:43Z", "2012-04-23 18:25:43 +0000 UTC", false},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TextTimeAction(tt.action, []byte(tt.in))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got.String())
			}
		})
	}
}
