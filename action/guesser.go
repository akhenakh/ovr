package action

import "net/http"

func GuessFormat(v []byte) Format {
	if len(v) >= 3 {
		if GuessFormatIsBinary(v) {
			return binFormat
		}
	}

	return textFormat
}

func GuessContentType(v []byte) string {
	return http.DetectContentType(v)
}

func GuessFormatIsBinary(v []byte) bool {
	// look at the first 512 bytes
	l := 512
	if len(v) < l {
		l = len(v)
	}

	b := v[:l]

	// https://datatracker.ietf.org/doc/html/draft-ietf-websec-mime-sniff#rfc.section.5
	// 0x00 -- 0x08  0x0B   0x0E -- 0x1A  0x1C -- 0x1F
	for _, c := range b {
		switch c {
		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08:
			fallthrough
		case 0x0B:
			fallthrough
		case 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a:
			fallthrough
		case 0x1C, 0x1D, 0x1E, 0x1F:
			return true
		}
	}

	return false
}
