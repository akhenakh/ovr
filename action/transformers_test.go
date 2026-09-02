package action

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func (r *ActionRegistry) TextAction(name string, in []byte) ([]byte, error) {
	a, ok := r.ActionByName(TextFormat, name)
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for text input", name)
	}

	out, err := a.Transform(NewDataText(in))
	if err != nil {
		return nil, err
	}
	return out.RawValue, nil
}

func (r *ActionRegistry) TextTimeAction(name string, in []byte) (time.Time, error) {
	a, ok := r.ActionByName(TextFormat, name)
	if !ok {
		return time.Time{}, fmt.Errorf("action %s does not exist for text input", name)
	}

	out, err := a.Transform(NewDataText(in))
	if err != nil {
		return time.Time{}, err
	}
	t, ok := out.Value.(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("output is not a time.Time")
	}
	return t, nil
}

func (r *ActionRegistry) TimeTextAction(name string, in time.Time) ([]byte, error) {
	a, ok := r.ActionByName(TimeFormat, name)
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for time input", name)
	}

	out, err := a.Transform(NewDataTime(in))
	if err != nil {
		return nil, err
	}
	return out.RawValue, nil
}

func (r *ActionRegistry) TimeAction(name string, in time.Time) (time.Time, error) {
	a, ok := r.ActionByName(TimeFormat, name)
	if !ok {
		return time.Time{}, fmt.Errorf("action %s does not exist for time input", name)
	}

	out, err := a.Transform(NewDataTime(in))
	if err != nil {
		return time.Time{}, err
	}
	t, ok := out.Value.(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("output is not a time.Time")
	}
	return t, nil
}

func (r *ActionRegistry) TextTextListAction(name string, in []byte) ([]string, error) {
	a, ok := r.ActionByName(TextFormat, name)
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for text input", name)
	}

	out, err := a.Transform(NewDataText(in))
	if err != nil {
		return nil, err
	}
	l, ok := out.Value.([]string)
	if !ok {
		return nil, fmt.Errorf("output is not a []string")
	}
	return l, nil
}

func (r *ActionRegistry) TextListTextListAction(name string, in []string) ([]string, error) {
	a, ok := r.ActionByName(TextListFormat, name)
	if !ok {
		a, ok = r.ActionByName(TextFormat, name)
		if !ok {
			return nil, fmt.Errorf("action %s does not exist for list of string input", name)
		}
	}

	out, err := a.Transform(NewDataTextList(in))
	if err != nil {
		return nil, err
	}
	l, ok := out.Value.([]string)
	if !ok {
		return nil, fmt.Errorf("output is not a []string")
	}
	return l, nil
}

func (r *ActionRegistry) TextListTextAction(name string, in []string) ([]byte, error) {
	a, ok := r.ActionByName(TextListFormat, name)
	if !ok {
		return nil, fmt.Errorf("action %s does not exist for list of string input", name)
	}

	out, err := a.Transform(NewDataTextList(in))
	if err != nil {
		return nil, err
	}
	return out.RawValue, nil
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
		{"space", "a b c", []string{"a", "b", "c"}, false},
		{"space", "a", nil, true},
		{"pipe", "a|b|c", []string{"a", "b", "c"}, false},
		{
			"jwt",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlJvYmVydCIsImlhdCI6MTUxNjIzOTAyMn0.fiHN5qbwhxBjwxLKSXfDV4wkVeuNeV8URADmuiYYYQo",
			[]string{`{"alg":"HS256","typ":"JWT"}`, `{"sub":"1234567890","name":"Robert","iat":1516239022}`},
			false,
		},
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
		params  []any
		in      []string
		want    string
		wantErr bool
	}{
		{action: "comma", in: []string{"a", "b"}, want: "a,b", wantErr: false},
		{action: "line", in: []string{"a", "b"}, want: "a\nb", wantErr: false},
		{action: "first", in: []string{"A", "B", "C"}, want: "A", wantErr: false},
		{action: "last", in: []string{"A", "B", "C"}, want: "C", wantErr: false},
		{action: "join", params: []any{"-"}, in: []string{"a", "b"}, want: "a-b", wantErr: false},
		{action: "index", params: []any{2}, in: []string{"A", "B", "C"}, want: "C", wantErr: false},
		{action: "index", params: []any{5}, in: []string{"A", "B", "C"}, want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			if tt.params != nil {
				a := r.MustActionByName(TextListFormat, tt.action)
				require.NoError(t, a.SetInputParameters(tt.params...))
			}
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

func TestAction_Parameters(t *testing.T) {
	r := NewRegistry()

	a := r.MustActionByName(TextListFormat, "join")
	_, err := a.Transform(NewDataTextList([]string{"a", "b"}))
	require.Error(t, err)
	require.Error(t, a.SetInputParameters(42))
	require.Error(t, a.SetInputParameters("a", "b"))
	require.NoError(t, a.SetInputParameters("-"))

	out, err := a.Transform(NewDataTextList([]string{"a", "b"}))
	require.NoError(t, err)
	require.Equal(t, "a-b", string(out.RawValue))
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
		{action: "jsonpretty", in: `{"engine_version":"1.7"}`, want: "{\n  \"engine_version\": \"1.7\"\n}", wantErr: false},
		{action: "unescape", in: `hello\nworld`, want: "hello\nworld", wantErr: false},
		{action: "unescape", in: `a\tb`, want: "a\tb", wantErr: false},
		{action: "trimspace", in: " hello ", want: "hello", wantErr: false},
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

	got, err := r.TextTimeAction("jsondate", []byte("2012-04-23T18:25:43Z"))
	require.NoError(t, err)
	require.Equal(t, "2012-04-23 18:25:43 +0000 UTC", got.String())

	// epoch is parsed in the process local timezone, compare the instant instead
	got, err = r.TextTimeAction("epoch", []byte("1257894000"))
	require.NoError(t, err)
	require.Equal(t, int64(1257894000), got.Unix())
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
		{
			// DST aware, summer is EDT
			"et",
			time.Date(2009, time.July, 10, 23, 0, 0, 0, time.UTC),
			"2009-07-10 19:00:00 -0400 EDT",
			false,
		},
		{
			"pst",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			"2009-11-10 15:00:00 -0800 PST",
			false,
		},
		{
			"jst",
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			"2009-11-11 08:00:00 +0900 JST",
			false,
		},
		{
			"cet",
			time.Date(2009, time.July, 10, 23, 0, 0, 0, time.UTC),
			"2009-07-11 01:00:00 +0200 CEST",
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

func TestAction_Tz(t *testing.T) {
	r := NewRegistry()

	a := r.MustActionByName(TimeFormat, "tz")
	require.NoError(t, a.SetInputParameters("Asia/Tokyo"))
	out, err := a.Transform(NewDataTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)))
	require.NoError(t, err)
	tm, ok := out.Value.(time.Time)
	require.True(t, ok)
	require.Equal(t, "2009-11-11 08:00:00 +0900 JST", tm.String())

	bad := r.MustActionByName(TimeFormat, "tz")
	require.NoError(t, bad.SetInputParameters("Mars/Olympus"))
	_, err = bad.Transform(NewDataTime(time.Now()))
	require.Error(t, err)
}

func TestAction_TzCaseInsensitive(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		param string
		want  string
	}{
		{"america/montreal", "America/Montreal"},
		{"AMERICA/NEW_YORK", "America/New_York"},
		{"asia/kolkata", "Asia/Kolkata"},
		{"australia/sydney", "Australia/Sydney"},
		{"utc", "UTC"},
		{"etc/gmt-14", "Etc/GMT-14"},
		{"Etc/GMT-14", "Etc/GMT-14"},
		{"America/Argentina/Buenos_Aires", "America/Argentina/Buenos_Aires"},
	}
	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			a := r.MustActionByName(TimeFormat, "tz")
			require.NoError(t, a.SetInputParameters(tt.param))
			out, err := a.Transform(NewDataTime(time.Now()))
			require.NoError(t, err)
			tm, ok := out.Value.(time.Time)
			require.True(t, ok)
			require.Equal(t, tt.want, tm.Location().String())
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
