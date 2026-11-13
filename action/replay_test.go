package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The stack replays actions on undo: non-deterministic ones must replay the
// value they produced instead of inventing a new one.

func TestUndo_ReplaysGeneratedUUID(t *testing.T) {
	r := NewRegistry()

	uuid4, ok := r.ActionByName(TextFormat, "uuid4")
	require.True(t, ok)
	upper, ok := r.ActionByName(TextFormat, "upper")
	require.True(t, ok)

	first, err := uuid4.Transform(NewDataText([]byte("ignored")))
	require.NoError(t, err)
	second, err := upper.Transform(first)
	require.NoError(t, err)
	require.Len(t, second.Stack, 2)

	back, removed, err := second.Undo([]byte("ignored"))
	require.NoError(t, err)
	require.Equal(t, "Upper", removed.Title())
	require.Equal(t, first.String(), back.String(), "undo must replay the same uuid")
}

func TestUndo_ReplaysShuffledList(t *testing.T) {
	r := NewRegistry()

	shuffle, ok := r.ActionByName(TextListFormat, "shuffle")
	require.True(t, ok)
	sortAction, ok := r.ActionByName(TextListFormat, "sort")
	require.True(t, ok)

	in := []string{"a", "b", "c", "d", "e", "f", "g"}
	shuffled, err := shuffle.Transform(NewDataTextList(in))
	require.NoError(t, err)
	sorted, err := sortAction.Transform(shuffled)
	require.NoError(t, err)
	require.NotNil(t, sorted)

	back, _, err := sorted.Undo(nil)
	require.NoError(t, err)
	require.Equal(t, shuffled.Value, back.Value, "undo must replay the same shuffle")
}

func TestRepeat_GeneratesFreshValuesAndReplays(t *testing.T) {
	r := NewRegistry()

	uuid4, ok := r.ActionByName(TextFormat, "uuid4")
	require.True(t, ok)
	repeat, ok := r.ActionByName(TextFormat, "repeat")
	require.True(t, ok)
	require.NoError(t, repeat.SetInputParameters(3))

	generated, err := uuid4.Transform(NewDataText([]byte("seed")))
	require.NoError(t, err)
	repeated, err := repeat.Transform(generated)
	require.NoError(t, err)

	values, ok := repeated.Value.([]string)
	require.True(t, ok)
	require.Len(t, values, 3)
	require.NotEqual(t, values[0], values[1], "repeat must generate fresh values")
	require.NotEqual(t, values[1], values[2], "repeat must generate fresh values")

	// a following action makes the repeat undoable: the replay must not
	// generate three new uuids
	first, ok := r.ActionByName(TextListFormat, "first")
	require.True(t, ok)
	afterFirst, err := first.Transform(repeated)
	require.NoError(t, err)

	back, _, err := afterFirst.Undo([]byte("seed"))
	require.NoError(t, err)
	require.Equal(t, values, back.Value)
}
