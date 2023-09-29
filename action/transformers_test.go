package action

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func (r *ActionRegistry) TextAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[textFormat.Prefix+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for text input", action)
	}

	ab, err := a.Func(in)
	return ab.([]byte), err
}

func (r *ActionRegistry) BinAction(action string, in []byte) ([]byte, error) {
	a, ok := r.m[binFormat.Prefix+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for bin input", action)
	}
	ab, err := a.Func(in)
	return ab.([]byte), err
}

func (r *ActionRegistry) TextTimeAction(action string, in []byte) (time.Time, error) {
	a, ok := r.m[textFormat.Prefix+","+action]
	if !ok {
		return time.Time{}, fmt.Errorf("action %s does not exist for text input", action)
	}
	ab, err := a.Func(in)
	return ab.(time.Time), err
}

func (r *ActionRegistry) TimeTextAction(action string, in time.Time) ([]byte, error) {
	a, ok := r.m[timeFormat.Prefix+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for time input", action)
	}
	ab, err := a.Func(in)
	return ab.([]byte), err
}

func (r *ActionRegistry) TimeAction(action string, in time.Time) (time.Time, error) {
	a, ok := r.m[timeFormat.Prefix+","+action]
	if !ok {
		return time.Time{}, fmt.Errorf("action %s does not exist for time input", action)
	}
	ab, err := a.Func(in)
	return ab.(time.Time), err
}

func (r *ActionRegistry) TextTextListAction(action string, in []byte) ([]string, error) {
	a, ok := r.m[textFormat.Prefix+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for list of string input", action)
	}
	ab, err := a.Func(in)
	return ab.([]string), err
}

func (r *ActionRegistry) TextListTextListAction(action string, in []string) ([]string, error) {
	a, ok := r.m[textListFormat.Prefix+","+action]
	if !ok {
		// special case to apply text to list of text
		a, ok = r.m[textFormat.Prefix+","+action]
		if !ok {
			return nil, fmt.Errorf("action %s does not exist for list of string input", action)
		}

		resp := make([]string, len(in))
		for i, s := range in {
			v, err := a.Func([]byte(s))
			if err != nil {
				return nil, err
			}
			resp[i] = string(v.([]byte))
		}
		return resp, nil
	}
	ab, err := a.Func(in)
	return ab.([]string), err
}

func (r *ActionRegistry) TextListTextAction(action string, in []string) ([]byte, error) {
	a, ok := r.m[textListFormat.Prefix+","+action]
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for list of string input", action)
	}
	ab, err := a.Func(in)
	return ab.([]byte), err
}

func TestAction_TextTextListTransform(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		action  string
		in      string
		want    []string
		wantErr bool
	}{
		{action: "comma", in: "hello", want: nil, wantErr: true},
		{"comma", "a,b", []string{"a", "b"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TextTextListAction(tt.action, []byte(tt.in))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAction_TextListTextListTransform(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		action  string
		in      []string
		want    []string
		wantErr bool
	}{
		{"upper", []string{"a", "b"}, []string{"A", "B"}, false},
		{"lower", []string{"A", "B"}, []string{"a", "b"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TextListTextListAction(tt.action, tt.in)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAction_TextListTextTransform(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		action  string
		in      []string
		want    string
		wantErr bool
	}{
		{action: "comma", in: []string{"a", "b"}, want: "a,b", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TextListTextAction(tt.action, tt.in)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, string(got))
			}
		})
	}
}

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
		{"hex", "48454 c4c4f", "HELLO", false},
		{"tohex", "HELLO", "48454c4c4f", false},
		{action: "base64", in: "aGVsbG8=", want: "hello", wantErr: false},
		{action: "tobase64", in: "hello", want: "aGVsbG8=", wantErr: false},
		{action: "minify", in: `{
			"engine_version":   "1.7"
		  }`, want: `{"engine_version":"1.7"}`, wantErr: false},
		{action: "jsoncompact", in: `{
			"engine_version":   "1.7"
		  }`, want: `{"engine_version":"1.7"}`, wantErr: false},
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
		{"epoch", "1257894000", "2009-11-10 18:00:00 -0500 EST", false},
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

func TestAction_TimeTransform(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		action  string
		in      time.Time
		want    string
		wantErr bool
	}{
		{
			"utc",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			"2009-11-10 23:00:00 +0000 UTC",
			false,
		},
		{
			"est",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			"2009-11-10 18:00:00 -0500 EST",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TimeAction(tt.action, tt.in)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got.String())
			}
		})
	}
}

func TestAction_TimeTextTransform(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		action  string
		in      time.Time
		want    string
		wantErr bool
	}{
		{
			"iso",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			"2009-11-10T23:00:00Z",
			false,
		},
		{
			"epoch",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			"1257894000",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			got, err := r.TimeTextAction(tt.action, tt.in)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, string(got))
			}
		})
	}
}
