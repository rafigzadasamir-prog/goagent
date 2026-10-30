package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// EvalExpr evaluates a simple arithmetic expression with +, -, *, /,
// parentheses and float literals, using standard operator precedence.
// It is intentionally minimal: no variables, no functions, just the
// four basic operators, so the agent can do arithmetic without shelling
// out to anything.
func EvalExpr(expr string) (float64, error) {
	p := &exprParser{input: []rune(strings.TrimSpace(expr))}
	val, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	p.skipSpace()
	if p.pos != len(p.input) {
		return 0, fmt.Errorf("unexpected character %q at position %d", p.input[p.pos], p.pos)
	}
	return val, nil
}

type exprParser struct {
	input []rune
	pos   int
}

func (p *exprParser) skipSpace() {
	for p.pos < len(p.input) && unicode.IsSpace(p.input[p.pos]) {
		p.pos++
	}
}

func (p *exprParser) peek() (rune, bool) {
	p.skipSpace()
	if p.pos >= len(p.input) {
		return 0, false
	}
	return p.input[p.pos], true
}

// parseExpr handles + and - (lowest precedence).
func (p *exprParser) parseExpr() (float64, error) {
	val, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		c, ok := p.peek()
		if !ok || (c != '+' && c != '-') {
			return val, nil
		}
		p.pos++
		rhs, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if c == '+' {
			val += rhs
		} else {
			val -= rhs
		}
	}
}

// parseTerm handles * and / (higher precedence than + -).
func (p *exprParser) parseTerm() (float64, error) {
	val, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		c, ok := p.peek()
		if !ok || (c != '*' && c != '/') {
			return val, nil
		}
		p.pos++
		rhs, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if c == '*' {
			val *= rhs
		} else {
			if rhs == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			val /= rhs
		}
	}
}

// parseFactor handles unary minus, parentheses and numeric literals.
func (p *exprParser) parseFactor() (float64, error) {
	c, ok := p.peek()
	if !ok {
		return 0, fmt.Errorf("unexpected end of expression")
	}

	if c == '-' {
		p.pos++
		val, err := p.parseFactor()
		return -val, err
	}
	if c == '+' {
		p.pos++
		return p.parseFactor()
	}
	if c == '(' {
		p.pos++
		val, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		closer, ok := p.peek()
		if !ok || closer != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return val, nil
	}

	start := p.pos
	for p.pos < len(p.input) && (unicode.IsDigit(p.input[p.pos]) || p.input[p.pos] == '.') {
		p.pos++
	}
	if p.pos == start {
		return 0, fmt.Errorf("expected number at position %d, got %q", start, c)
	}
	return strconv.ParseFloat(string(p.input[start:p.pos]), 64)
}
