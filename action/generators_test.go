package action

import (
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestAction_NewUUID(t *testing.T) {
	r := NewRegistry()

	for _, tt := range []struct {
		name    string
		version byte
	}{
		{"uuid4", 4},
		{"uuid7", 7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := r.MustActionByName(TextFormat, tt.name)
			out, err := a.Transform(NewDataText([]byte("input is ignored")))
			require.NoError(t, err)
			require.Equal(t, TextFormat, out.Format)

			u, err := uuid.Parse(string(out.RawValue))
			require.NoError(t, err)
			require.Equal(t, tt.version, u[6]>>4)

			// generators ignore the input, even empty
			out2, err := a.Transform(NewDataText(nil))
			require.NoError(t, err)
			require.NotEqual(t, string(out.RawValue), string(out2.RawValue))
		})
	}
}

func TestAction_NewUUIDFromAnyFormat(t *testing.T) {
	r := NewRegistry()

	a := r.MustActionByName(TimeFormat, "uuid4")
	out, err := a.Transform(NewDataTime(time.Now()))
	require.NoError(t, err)
	require.Equal(t, TextFormat, out.Format)
	_, err = uuid.Parse(string(out.RawValue))
	require.NoError(t, err)
}

func TestAction_Now(t *testing.T) {
	r := NewRegistry()

	a := r.MustActionByName(TextFormat, "now")
	before := time.Now().Add(-time.Second)
	out, err := a.Transform(NewDataText([]byte("x")))
	require.NoError(t, err)
	require.Equal(t, TimeFormat, out.Format)
	tm, ok := out.Value.(time.Time)
	require.True(t, ok)
	require.False(t, tm.Before(before))

	_, ok = r.ActionByName(TextFormat, "newtime")
	require.True(t, ok)
}

func TestAction_Repeat(t *testing.T) {
	r := NewRegistry()

	uuid4 := r.MustActionByName(TextFormat, "uuid4")
	out, err := uuid4.Transform(NewDataText([]byte("x")))
	require.NoError(t, err)

	rep := r.MustActionByName(TextFormat, "repeat")
	require.NoError(t, rep.SetInputParameters(3))

	lout, err := rep.Transform(out)
	require.NoError(t, err)
	require.Equal(t, TextListFormat, lout.Format)

	l, ok := lout.Value.([]string)
	require.True(t, ok)
	require.Len(t, l, 3)
	seen := make(map[string]bool, len(l))
	for _, s := range l {
		u, err := uuid.Parse(s)
		require.NoError(t, err)
		require.Equal(t, byte(4), u[6]>>4)
		seen[s] = true
	}
	require.Len(t, seen, len(l))
}

func TestAction_RepeatFromTimeData(t *testing.T) {
	r := NewRegistry()

	now := r.MustActionByName(TextFormat, "now")
	tout, err := now.Transform(NewDataText([]byte("x")))
	require.NoError(t, err)
	require.Equal(t, TimeFormat, tout.Format)

	trep := r.MustActionByName(TimeFormat, "repeat")
	require.NoError(t, trep.SetInputParameters(2))
	tlout, err := trep.Transform(tout)
	require.NoError(t, err)
	require.Equal(t, TextListFormat, tlout.Format)

	tl, ok := tlout.Value.([]string)
	require.True(t, ok)
	require.Len(t, tl, 2)
	_, err = time.Parse("2006-01-02 15:04:05 -0700 MST", strings.Split(tl[0], " m=+")[0])
	require.NoError(t, err)
}

func TestAction_StackString(t *testing.T) {
	r := NewRegistry()

	out, err := r.MustActionByName(TextFormat, "uuid4").Transform(NewDataText([]byte("x")))
	require.NoError(t, err)

	rep := r.MustActionByName(TextFormat, "repeat")
	require.NoError(t, rep.SetInputParameters(3))
	out, err = rep.Transform(out)
	require.NoError(t, err)

	out, err = r.MustActionByName(TextListFormat, "comma").Transform(out)
	require.NoError(t, err)

	require.Equal(t, "Uuid4,Repeat(3),Comma", out.StackString())
}

func TestAction_RepeatErrors(t *testing.T) {
	r := NewRegistry()

	uuid4 := r.MustActionByName(TextFormat, "uuid4")
	out, err := uuid4.Transform(NewDataText([]byte("x")))
	require.NoError(t, err)

	// missing parameter
	bad := r.MustActionByName(TextFormat, "repeat")
	_, err = bad.Transform(out)
	require.Error(t, err)

	// invalid count
	bad = r.MustActionByName(TextFormat, "repeat")
	require.NoError(t, bad.SetInputParameters(0))
	_, err = bad.Transform(out)
	require.Error(t, err)

	bad = r.MustActionByName(TextFormat, "repeat")
	require.NoError(t, bad.SetInputParameters(-1))
	_, err = bad.Transform(out)
	require.Error(t, err)

	// nothing to repeat
	bad = r.MustActionByName(TextFormat, "repeat")
	require.NoError(t, bad.SetInputParameters(1))
	_, err = bad.Transform(NewDataText([]byte("x")))
	require.Error(t, err)

	// can't repeat repeat
	rep := r.MustActionByName(TextFormat, "repeat")
	require.NoError(t, rep.SetInputParameters(2))
	lout, err := rep.Transform(out)
	require.NoError(t, err)
	require.NoError(t, rep.SetInputParameters(2))
	_, err = rep.Transform(lout)
	require.Error(t, err)
}
