package action

import (
	"fmt"
	"math"
	"strconv"
)

var calcAction = New(Definition[[]byte, []byte]{
	Doc:          "Evaluate the input as an arithmetic expression, e.g. 12 + 23, supports + - * / % ^ and parentheses",
	Names:        []string{"calc"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		v, err := evalCalc(string(in))
		if err != nil {
			return nil, err
		}
		return []byte(v), nil
	},
})

type calcParser struct {
	s string
	i int
}

func evalCalc(s string) (string, error) {
	p := &calcParser{s: s}
	v, err := p.additive()
	if err != nil {
		return "", err
	}
	p.skipSpaces()
	if p.i < len(p.s) {
		return "", fmt.Errorf("unexpected %q in expression %q", p.s[p.i:], s)
	}
	return strconv.FormatFloat(v, 'f', -1, 64), nil
}

func (p *calcParser) skipSpaces() {
	for p.i < len(p.s) {
		switch p.s[p.i] {
		case ' ', '\t', '\n', '\r':
			p.i++
		default:
			return
		}
	}
}

func (p *calcParser) additive() (float64, error) {
	v, err := p.multiplicative()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		if p.i >= len(p.s) || p.s[p.i] != '+' && p.s[p.i] != '-' {
			return v, nil
		}
		op := p.s[p.i]
		p.i++
		rhs, err := p.multiplicative()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			v += rhs
		} else {
			v -= rhs
		}
	}
}

func (p *calcParser) multiplicative() (float64, error) {
	v, err := p.unary()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		if p.i >= len(p.s) {
			return v, nil
		}
		switch p.s[p.i] {
		case '*':
			p.i++
			rhs, err := p.unary()
			if err != nil {
				return 0, err
			}
			v *= rhs
		case '/':
			p.i++
			rhs, err := p.unary()
			if err != nil {
				return 0, err
			}
			if rhs == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			v /= rhs
		case '%':
			p.i++
			rhs, err := p.unary()
			if err != nil {
				return 0, err
			}
			if rhs == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			v = math.Mod(v, rhs)
		default:
			return v, nil
		}
	}
}

func (p *calcParser) unary() (float64, error) {
	p.skipSpaces()
	if p.i < len(p.s) {
		switch p.s[p.i] {
		case '-':
			p.i++
			v, err := p.unary()
			if err != nil {
				return 0, err
			}
			return -v, nil
		case '+':
			p.i++
			return p.unary()
		}
	}
	return p.power()
}

func (p *calcParser) power() (float64, error) {
	v, err := p.primary()
	if err != nil {
		return 0, err
	}
	p.skipSpaces()
	if p.i < len(p.s) && p.s[p.i] == '^' {
		p.i++
		rhs, err := p.unary()
		if err != nil {
			return 0, err
		}
		v = math.Pow(v, rhs)
	}
	return v, nil
}

func (p *calcParser) primary() (float64, error) {
	p.skipSpaces()
	if p.i >= len(p.s) {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	if p.s[p.i] == '(' {
		p.i++
		v, err := p.additive()
		if err != nil {
			return 0, err
		}
		p.skipSpaces()
		if p.i >= len(p.s) || p.s[p.i] != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.i++
		return v, nil
	}

	start := p.i
	for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' || p.i < len(p.s) && p.s[p.i] == '.' {
		p.i++
	}
	if p.i < len(p.s) && (p.s[p.i] == 'e' || p.s[p.i] == 'E') {
		j := p.i + 1
		if j < len(p.s) && (p.s[j] == '+' || p.s[j] == '-') {
			j++
		}
		if j < len(p.s) && p.s[j] >= '0' && p.s[j] <= '9' {
			p.i = j
			for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
				p.i++
			}
		}
	}
	if start == p.i {
		return 0, fmt.Errorf("unexpected %q in expression", p.s[p.i])
	}
	v, err := strconv.ParseFloat(p.s[start:p.i], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", p.s[start:p.i], err)
	}
	return v, nil
}
