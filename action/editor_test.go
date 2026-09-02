package action

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEditRegisteredForAllFormats(t *testing.T) {
	r := NewRegistry()
	for _, f := range inputFormats {
		a, ok := r.ActionByName(f, "edit")
		require.True(t, ok, "no edit action for format %s", f.Name)
		_, ok = a.(Editor)
		require.True(t, ok, "edit action for format %s is not an Editor", f.Name)
	}
}

func TestEditContent(t *testing.T) {
	r := NewRegistry()

	a, ok := r.ActionByName(TextFormat, "edit")
	require.True(t, ok)
	ea := a.(Editor)

	// text is kept as is
	require.Equal(t, "hello", ea.EditContent(NewDataText([]byte("hello"))))

	// non text entries are converted to their string representation
	tm := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	at, ok := r.ActionByName(TimeFormat, "edit")
	require.True(t, ok)
	require.Equal(t, tm.String(), at.(Editor).EditContent(NewDataTime(tm)))

	require.Equal(t, "a\nb", ea.EditContent(NewDataTextList([]string{"a", "b"})))

	require.Equal(t, "a,b\nc,d", ea.EditContent(NewDataTable([][]string{{"a", "b"}, {"c", "d"}})))
}

func TestEditTransform(t *testing.T) {
	r := NewRegistry()
	a, ok := r.ActionByName(TextFormat, "edit")
	require.True(t, ok)

	// without the edited content the action can't be applied
	_, err := a.Transform(NewDataText([]byte("hello")))
	require.Error(t, err)

	require.NoError(t, a.SetInputParameters("edited"))
	out, err := a.Transform(NewDataText([]byte("hello")))
	require.NoError(t, err)
	require.Equal(t, TextFormat, out.Format)
	require.Equal(t, []byte("edited"), out.RawValue)
}

func TestEditUndoReplay(t *testing.T) {
	r := NewRegistry()
	ea, ok := r.ActionByName(TextFormat, "edit")
	require.True(t, ok)
	ua, ok := r.ActionByName(TextFormat, "upper")
	require.True(t, ok)

	require.NoError(t, ea.SetInputParameters("edited"))
	d, err := ea.Transform(NewDataText([]byte("hello")))
	require.NoError(t, err)
	d, err = ua.Transform(d)
	require.NoError(t, err)
	require.Equal(t, []byte("EDITED"), d.RawValue)

	// undoing re-applies the edit action from the stack without
	// launching the editor
	nd, popped, err := d.Undo([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, "Upper", popped.Title())
	require.Equal(t, []byte("edited"), nd.RawValue)
}
