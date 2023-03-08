package action

import "testing"

func TestGuessIsBinary(t *testing.T) {
	tests := []struct {
		name string
		v    []byte
		want bool
	}{
		{"not matching bin", []byte("NOOOP"), false},
		{"matching bin", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, true},
		{"matching bin 0x1C, 0x1D, 0x1E, 0x1F:", []byte{0x1C, 0x1D, 0x1E, 0x1F}, true},
		{"not machint long string", []byte("CompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressinCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressingCompressin"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GuessFormatIsBinary(tt.v); got != tt.want {
				t.Errorf("GuessIsBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}
