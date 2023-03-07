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
		{action: "quote", in: "hello", want: "\"hello\"", wantErr: false},
		{action: "quote", in: "hello\n", want: "\"hello\\n\"", wantErr: false},
		{action: "unquote", in: "\"hello\"", want: "hello", wantErr: false},
		{action: "sha1", in: "hello", want: "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d", wantErr: false},
		{action: "sha256", in: "hello", want: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", wantErr: false},
		{action: "sha512", in: "hello", want: "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043", wantErr: false},
		{action: "md5", in: "hello", want: "5d41402abc4b2a76b9719d911017c592", wantErr: false},
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
