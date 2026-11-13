package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction_IntBase(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		in       string
		from, to int
		want     string
		wantErr  bool
	}{
		{"ff", 16, 10, "255", false},
		{"255", 10, 16, "ff", false},
		{"0xff", 0, 10, "255", false},
		{"0b1010", 0, 10, "10", false},
		{"0o17", 0, 16, "f", false},
		{"-ff", 16, 2, "-11111111", false},
		{"1_000", 10, 10, "1000", false},
		{"18446744073709551616", 10, 16, "10000000000000000", false}, // above uint64
		{"zz", 10, 16, "", true},
		{"10", 10, 1, "", true},
	}
	for _, tc := range tests {
		got, err := textWithParams(t, r, "intbase", []byte(tc.in), tc.from, tc.to)
		if tc.wantErr {
			require.Error(t, err, "intbase(%q, %d, %d)", tc.in, tc.from, tc.to)
			continue
		}
		require.NoError(t, err, "intbase(%q, %d, %d)", tc.in, tc.from, tc.to)
		require.Equal(t, tc.want, string(got), "intbase(%q, %d, %d)", tc.in, tc.from, tc.to)
	}
}

func TestAction_Convert(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		in, param, want string
		wantErr         bool
	}{
		{"12 km", "mi", "7.45645 mi", false},
		{"12 km", "km>mi", "7.45645 mi", false},
		{"12", "km>mi", "7.45645 mi", false},
		{"100", "c>f", "212 f", false},
		{"32 f", "c", "0 c", false},
		{"273.15 k", "c", "0 c", false},
		{"1", "kib>kb", "1.024 kb", false},
		{"2.5", "h>min", "150 min", false},
		{"1", "km>kg", "", true},
		{"1", "parsec>m", "", true},
		{"abc", "km>mi", "", true},
		{"1", "mi", "", true},
	}
	for _, tc := range tests {
		got, err := textWithParams(t, r, "convert", []byte(tc.in), tc.param)
		if tc.wantErr {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
		require.Equal(t, tc.want, string(got))
	}
}
