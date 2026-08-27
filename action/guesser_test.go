package action

import "testing"

func TestGuessWKT(t *testing.T) {
	tests := []struct {
		name string
		v    string
		want bool
	}{
		{"point", "POINT(-0.4539761 48.0930043)", true},
		{"point with space", "POINT (-0.4539761 48.0930043)", true},
		{"lowercase point", "point(-0.4539761 48.0930043)", true},
		{"leading whitespace", "\n  POINT(-0.4539761 48.0930043)", true},
		{"point z", "POINT Z (1 2 3)", true},
		{"point zm", "POINT ZM (1 2 3 4)", true},
		{"point empty", "POINT EMPTY", true},
		{"linestring", "LINESTRING(0 0,1 1)", true},
		{"polygon", "POLYGON((0 0,0 1,1 1,1 0,0 0))", true},
		{"multipolygon", "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)))", true},
		{"geometrycollection", "GEOMETRYCOLLECTION(POINT(0 0),POINT(1 1))", true},
		{"keyword only", "POINT", false},
		{"text starting with point", "Point break", false},
		{"text with paren", "Multiple things (foo)", false},
		{"json", `{"type":"Point","coordinates":[2.2,48.8]}`, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GuessWKT([]byte(tt.v)); got != tt.want {
				t.Errorf("GuessWKT(%q) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

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
