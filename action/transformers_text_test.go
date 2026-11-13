package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// textWithParams runs a text action after setting its parameters.
func textWithParams(t *testing.T, r *ActionRegistry, name string, in []byte, params ...any) ([]byte, error) {
	t.Helper()
	a, ok := r.ActionByName(TextFormat, name)
	require.True(t, ok, "action %s does not exist for text input", name)
	require.NoError(t, a.SetInputParameters(params...))
	out, err := a.Transform(NewDataText(in))
	if err != nil {
		return nil, err
	}
	return out.RawValue, nil
}

func TestAction_URLEncodeDecode(t *testing.T) {
	r := NewRegistry()

	got, err := textWithParams(t, r, "urlencode", []byte("a b&c=d"), "query")
	require.NoError(t, err)
	require.Equal(t, "a+b%26c%3Dd", string(got))

	back, err := textWithParams(t, r, "urldecode", got, "query")
	require.NoError(t, err)
	require.Equal(t, "a b&c=d", string(back))

	path, err := textWithParams(t, r, "urlencode", []byte("a b/c"), "path")
	require.NoError(t, err)
	back, err = textWithParams(t, r, "urldecode", path, "path")
	require.NoError(t, err)
	require.Equal(t, "a b/c", string(back))

	_, err = textWithParams(t, r, "urlencode", []byte("x"), "nonsense")
	require.Error(t, err)

	_, err = textWithParams(t, r, "urldecode", []byte("%zz"), "query")
	require.Error(t, err)
}

func TestAction_HTMLEscape(t *testing.T) {
	r := NewRegistry()

	got, err := r.TextAction("htmlescape", []byte(`<b class="x">&</b>`))
	require.NoError(t, err)
	require.Equal(t, "&lt;b class=&#34;x&#34;&gt;&amp;&lt;/b&gt;", string(got))

	back, err := r.TextAction("htmlunescape", got)
	require.NoError(t, err)
	require.Equal(t, `<b class="x">&</b>`, string(back))
}

func TestAction_Binary(t *testing.T) {
	r := NewRegistry()

	got, err := r.TextAction("tobinary", []byte("hi"))
	require.NoError(t, err)
	require.Equal(t, "01101000 01101001", string(got))

	back, err := r.TextAction("frombinary", got)
	require.NoError(t, err)
	require.Equal(t, "hi", string(back))

	// separators and spacing are ignored
	back, err = r.TextAction("frombinary", []byte("01101000\n01101001"))
	require.NoError(t, err)
	require.Equal(t, "hi", string(back))

	_, err = r.TextAction("frombinary", []byte("0110100"))
	require.Error(t, err)

	_, err = r.TextAction("frombinary", []byte("  "))
	require.Error(t, err)
}

func TestAction_Slug(t *testing.T) {
	r := NewRegistry()

	tests := []struct{ in, want string }{
		{"Héllo, World!", "hello-world"},
		{"  --foo__bar  ", "foo-bar"},
		{"Already-slugged", "already-slugged"},
		{"Ünïcödé ß", "unicode-ss"},
		{"Æther Øresund", "aether-oresund"},
		{"!!!", ""},
	}
	for _, tc := range tests {
		got, err := r.TextAction("slug", []byte(tc.in))
		require.NoError(t, err)
		require.Equal(t, tc.want, string(got), "slug(%q)", tc.in)
	}
}

func TestAction_Replace(t *testing.T) {
	r := NewRegistry()

	got, err := textWithParams(t, r, "replace", []byte("bob@example.com"), `(\w+)@`, "$1 at ")
	require.NoError(t, err)
	require.Equal(t, "bob at example.com", string(got))

	_, err = textWithParams(t, r, "replace", []byte("x"), "(", "y")
	require.Error(t, err)
}

func TestAction_TextStats(t *testing.T) {
	r := NewRegistry()

	got, err := r.TextAction("stats", []byte("hello world. Bye!\n\nSecond paragraph."))
	require.NoError(t, err)
	s := string(got)
	require.Contains(t, s, "words: 5")
	require.Contains(t, s, "sentences: 3")
	require.Contains(t, s, "paragraphs: 2")
	require.Contains(t, s, "lines: 3")
	require.Contains(t, s, "unique words: 5")
}
