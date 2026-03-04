package worker

import (
	"math"
	"strconv"
	"strings"
)

func evalFormula(formula string, t float64) float64 {
	expr := strings.ReplaceAll(strings.TrimSpace(formula), " ", "")
	expr = strings.ReplaceAll(expr, "t", strconv.FormatFloat(t, 'f', -1, 64))
	expr = evalFn(expr, "sin", math.Sin)
	expr = evalFn(expr, "cos", math.Cos)
	p := &parser{in: expr}
	return p.parseExpr()
}

func evalFn(expr, fn string, f func(float64) float64) string {
	for {
		start := strings.Index(expr, fn+"(")
		if start < 0 {
			return expr
		}
		end := strings.Index(expr[start:], ")")
		if end < 0 {
			return expr
		}
		end += start
		inner := expr[start+len(fn)+1 : end]
		p := &parser{in: inner}
		v := f(p.parseExpr())
		expr = expr[:start] + strconv.FormatFloat(v, 'f', -1, 64) + expr[end+1:]
	}
}

type parser struct {
	in  string
	pos int
}

func (p *parser) parseExpr() float64 {
	v := p.parseTerm()
	for p.pos < len(p.in) {
		op := p.in[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++
		rhs := p.parseTerm()
		if op == '+' {
			v += rhs
		} else {
			v -= rhs
		}
	}
	return v
}

func (p *parser) parseTerm() float64 {
	v := p.parseNumber()
	for p.pos < len(p.in) {
		op := p.in[p.pos]
		if op != '*' && op != '/' {
			break
		}
		p.pos++
		rhs := p.parseNumber()
		if op == '*' {
			v *= rhs
		} else {
			v /= rhs
		}
	}
	return v
}

func (p *parser) parseNumber() float64 {
	start := p.pos
	if p.pos < len(p.in) && (p.in[p.pos] == '+' || p.in[p.pos] == '-') {
		p.pos++
	}
	for p.pos < len(p.in) && (p.in[p.pos] == '.' || (p.in[p.pos] >= '0' && p.in[p.pos] <= '9')) {
		p.pos++
	}
	if start == p.pos {
		return 0
	}
	v, _ := strconv.ParseFloat(p.in[start:p.pos], 64)
	return v
}
