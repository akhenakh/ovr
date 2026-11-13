package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction_XMLPretty(t *testing.T) {
	r := NewRegistry()

	got, err := r.TextAction("xmlpretty", []byte(`<a><b x="1">text</b><c/></a>`))
	require.NoError(t, err)
	require.Equal(t, `<a>
  <b x="1">text</b>
  <c></c>
</a>`, string(got))

	_, err = r.TextAction("xmlpretty", []byte(`<a><b></a>`))
	require.Error(t, err)
}

func TestAction_Markdown(t *testing.T) {
	r := NewRegistry()

	got, err := r.TextAction("markdown", []byte("# Title\n\n- one\n- two"))
	require.NoError(t, err)
	s := string(got)
	require.Contains(t, s, "<h1>Title</h1>")
	require.Contains(t, s, "<li>one</li>")
}
