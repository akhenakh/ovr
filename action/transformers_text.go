package action

import (
	"fmt"
	"html"
	"math"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// escapeMode reads the query/path mode parameter shared by the URL actions.
func escapeMode(a Action) (string, error) {
	p, ok := a.InputParameters()[0].(string)
	if !ok {
		return "", fmt.Errorf("mode parameter is not a string")
	}
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "query", "q":
		return "query", nil
	case "path", "component", "c":
		return "path", nil
	default:
		return "", fmt.Errorf("unknown mode %q, use query or path", p)
	}
}

var urlEncodeAction = New(Definition[[]byte, []byte]{
	Doc:          "Percent-encode the input for a URL, mode 'query' (space becomes +) or 'path' (space becomes %20)",
	Names:        []string{"urlencode"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "mode: query or path"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		mode, err := escapeMode(a)
		if err != nil {
			return nil, err
		}
		if mode == "path" {
			return []byte(url.PathEscape(string(in))), nil
		}
		return []byte(url.QueryEscape(string(in))), nil
	},
})

var urlDecodeAction = New(Definition[[]byte, []byte]{
	Doc:          "Percent-decode a URL component, mode 'query' (also turns + into a space) or 'path'",
	Names:        []string{"urldecode"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "mode: query or path"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		mode, err := escapeMode(a)
		if err != nil {
			return nil, err
		}
		if mode == "path" {
			out, err := url.PathUnescape(string(in))
			if err != nil {
				return nil, fmt.Errorf("invalid percent-encoding: %w", err)
			}
			return []byte(out), nil
		}
		out, err := url.QueryUnescape(string(in))
		if err != nil {
			return nil, fmt.Errorf("invalid percent-encoding: %w", err)
		}
		return []byte(out), nil
	},
})

var htmlEscapeAction = New(Definition[[]byte, []byte]{
	Doc:          "Escape HTML special characters, e.g. < to &lt;",
	Names:        []string{"htmlescape"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(html.EscapeString(string(in))), nil
	},
})

var htmlUnescapeAction = New(Definition[[]byte, []byte]{
	Doc:          "Unescape HTML entities, e.g. &lt; to <",
	Names:        []string{"htmlunescape"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(html.UnescapeString(string(in))), nil
	},
})

var toBinaryAction = New(Definition[[]byte, []byte]{
	Doc:          "Encode the input bytes as 8-bit binary groups, e.g. 'hi' to 01101000 01101001",
	Names:        []string{"tobinary"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		parts := make([]string, len(in))
		for i, b := range in {
			parts[i] = fmt.Sprintf("%08b", b)
		}
		return []byte(strings.Join(parts, " ")), nil
	},
})

var fromBinaryAction = New(Definition[[]byte, []byte]{
	Doc:          "Decode 8-bit binary groups back to text, separators and spacing are ignored",
	Names:        []string{"frombinary"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		bits := strings.Map(func(r rune) rune {
			if r == '0' || r == '1' {
				return r
			}
			return -1
		}, string(in))
		if len(bits) == 0 {
			return nil, fmt.Errorf("no binary digits in the input")
		}
		if len(bits)%8 != 0 {
			return nil, fmt.Errorf("binary input is %d bits, it must be a multiple of 8", len(bits))
		}
		out := make([]byte, len(bits)/8)
		for i := range out {
			var v byte
			for _, c := range bits[i*8 : i*8+8] {
				v <<= 1
				if c == '1' {
					v++
				}
			}
			out[i] = v
		}
		return out, nil
	},
})

var slugAction = New(Definition[[]byte, []byte]{
	Doc:          "Slugify the input: lowercase, accents folded, words separated by hyphens",
	Names:        []string{"slug"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(slugify(string(in))), nil
	},
})

// slugReplacements folds letters that Unicode decomposition leaves alone.
var slugReplacements = strings.NewReplacer(
	"ß", "ss", "æ", "ae", "Æ", "ae", "œ", "oe", "Œ", "oe",
	"ø", "o", "Ø", "o", "ł", "l", "Ł", "l", "đ", "d", "Đ", "d",
	"ð", "d", "Ð", "d", "þ", "th", "Þ", "th",
)

// slugify folds accents, lowercases and joins ASCII words with hyphens.
func slugify(s string) string {
	folded := norm.NFD.String(slugReplacements.Replace(s))
	var b strings.Builder
	lastSep := true
	for _, r := range folded {
		switch {
		case unicode.Is(unicode.Mn, r):
			// combining mark from the decomposition, drop it
		case r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			b.WriteRune(unicode.ToLower(r))
			lastSep = false
		case !lastSep:
			b.WriteByte('-')
			lastSep = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

var replaceTextAction = New(Definition[[]byte, []byte]{
	Doc:          "Replace every regexp match, groups are available as $1, e.g. pattern (\\w+)@ and replacement $1 at ",
	Names:        []string{"replace"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters: []ActionParameter{
		{StringParameter, "the regexp pattern to match"},
		{StringParameter, "the replacement, $1 refers to the first group"},
	},
	Func: func(a Action, in []byte) ([]byte, error) {
		params := a.InputParameters()
		pattern, ok := params[0].(string)
		if !ok {
			return nil, fmt.Errorf("pattern parameter is not a string")
		}
		replacement, ok := params[1].(string)
		if !ok {
			return nil, fmt.Errorf("replacement parameter is not a string")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		return []byte(re.ReplaceAllString(string(in), replacement)), nil
	},
})

var textStatsAction = New(Definition[[]byte, []byte]{
	Doc:          "Text statistics: characters, words, lines, sentences, paragraphs, unique words, reading time and entropy",
	Names:        []string{"stats"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		return []byte(textStats(string(in))), nil
	},
})

// wordsPerMinute is the reading speed used by the stats action.
const wordsPerMinute = 200

var paragraphRe = regexp.MustCompile(`\n\s*\n`)

// textStats renders a one-stat-per-line summary of the input.
func textStats(s string) string {
	words := strings.Fields(s)
	unique := make(map[string]struct{}, len(words))
	for _, w := range words {
		unique[strings.ToLower(w)] = struct{}{}
	}

	noSpaces := strings.Join(words, "")
	lines := 0
	if s != "" {
		lines = strings.Count(strings.TrimRight(s, "\n"), "\n") + 1
	}
	sentences := 0
	for _, r := range s {
		if r == '.' || r == '!' || r == '?' {
			sentences++
		}
	}
	paragraphs := 0
	for _, p := range paragraphRe.Split(strings.TrimSpace(s), -1) {
		if strings.TrimSpace(p) != "" {
			paragraphs++
		}
	}

	minutes := float64(len(words)) / wordsPerMinute
	reading := fmt.Sprintf("%.1fs", minutes*60)
	if minutes >= 1 {
		reading = fmt.Sprintf("%dm%02ds", int(minutes), int((minutes-float64(int(minutes)))*60))
	}

	return fmt.Sprintf(
		"characters: %d\ncharacters (no spaces): %d\nwords: %d\nunique words: %d\nlines: %d\nsentences: %d\nparagraphs: %d\nreading time: %s\nentropy: %.2f bits/char",
		len([]rune(s)), len([]rune(noSpaces)), len(words), len(unique), lines, sentences, paragraphs, reading, entropy(s),
	)
}

// entropy returns the Shannon entropy of the input characters in bits.
func entropy(s string) float64 {
	if s == "" {
		return 0
	}
	counts := make(map[rune]int)
	n := 0
	for _, r := range s {
		counts[r]++
		n++
	}
	var h float64
	for _, c := range counts {
		p := float64(c) / float64(n)
		h -= p * math.Log2(p)
	}
	return h
}
