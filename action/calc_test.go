package action

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAction_Calc(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"12 + 23", "35", false},
		{"12+23", "35", false},
		{"2 + 3 * 4", "14", false},
		{"(2 + 3) * 4", "20", false},
		{"10 / 4", "2.5", false},
		{"10 % 3", "1", false},
		{"2 ^ 10", "1024", false},
		{"2 ^ 3 ^ 2", "512", false},
		{"-5 + 3", "-2", false},
		{"2 * -3", "-6", false},
		{"2^-2", "0.25", false},
		{"1.5e2 + 1", "151", false},
		{"  12 + 23\n", "35", false},
		{"0.1 + 0.2", "0.30000000000000004", false},
		{"", "", true},
		{"12 +", "", true},
		{"hello", "", true},
		{"12 + 23)", "", true},
		{"(12 + 23", "", true},
		{"1/0", "", true},
		{"10 % 0", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := r.TextAction("calc", []byte(tt.in))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, string(got))
			}
		})
	}
}

func TestAction_NewCalc(t *testing.T) {
	r := NewRegistry()

	a := r.MustActionByName(TextFormat, "newcalc")

	// missing parameter
	_, err := a.Transform(NewDataText(nil))
	require.Error(t, err)

	require.NoError(t, a.SetInputParameters("12 + 23"))

	// ignores the buffer, even empty
	out, err := a.Transform(NewDataText(nil))
	require.NoError(t, err)
	require.Equal(t, "35", string(out.RawValue))

	// offered from any data format
	a2 := r.MustActionByName(TimeFormat, "newcalc")
	require.NoError(t, a2.SetInputParameters("2 ^ 10"))
	out, err = a2.Transform(NewDataTime(time.Now()))
	require.NoError(t, err)
	require.Equal(t, "1024", string(out.RawValue))
}
