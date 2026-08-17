// Package gfilter implements a subset of gcloud's client-side --filter
// expression language ("gcloud topic filters"). It is used by list commands
// that want to filter results client-side rather than pushing the expression
// through the underlying API's server-side filter (which uses a different
// syntax, e.g. Compute Engine's `field eq value`).
//
// Grammar supported:
//
//	expr    = orExpr
//	orExpr  = andExpr ("OR" andExpr)*
//	andExpr = notExpr ( ("AND"|)  notExpr)*        // implicit AND between terms
//	notExpr = "NOT" notExpr | atom
//	atom    = "(" expr ")" | term
//	term    = KEY OP VALUE
//	OP      = ":" | "=" | "!=" | "~" | "!~" | "<" | "<=" | ">" | ">="
//	KEY     = dotted identifier (may include [] indexing; array traversal is
//	          implicit — see match())
//	VALUE   = bareword | quoted string
//
// Semantics for terms:
//   - Evaluate KEY against the JSON representation of the record. If any leg
//     of the path is an array, each element is tried and OR-combined.
//   - ":" is a substring/contains match on strings (case-insensitive). Against
//     numbers it is equality. Against arrays it is membership.
//   - "=" / "!=" are string equality after Sprintf-ing the value.
//   - "~" / "!~" compile the RHS as a Go regexp and match against the string
//     form of the LHS.
//   - "<", "<=", ">", ">=" compare numerically if both sides parse as float,
//     otherwise lexicographically.
package gfilter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Filter is a compiled --filter expression.
type Filter struct {
	root node
}

// Compile parses expr into a Filter. An empty expr compiles to a no-op filter
// that matches everything.
func Compile(expr string) (*Filter, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return &Filter{root: alwaysTrue{}}, nil
	}
	p := newParser(expr)
	n, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if !p.eof() {
		return nil, fmt.Errorf("unexpected trailing input at position %d: %q", p.pos, p.rest())
	}
	return &Filter{root: n}, nil
}

// Match reports whether v (which will be JSON-marshaled and unmarshaled into a
// map[string]any) satisfies the filter. Any marshal error causes a false
// return.
func (f *Filter) Match(v any) bool {
	if f == nil || f.root == nil {
		return true
	}
	m, err := toMap(v)
	if err != nil {
		return false
	}
	return f.root.eval(m)
}

// UsesContainsOperator reports whether the expression contains a `:` operator.
// gcloud-python emits a deprecation-style warning in that case ("--filter :
// operator evaluation is changing..."); callers can use this to mirror that
// behaviour.
func (f *Filter) UsesContainsOperator() bool {
	if f == nil || f.root == nil {
		return false
	}
	return usesContains(f.root)
}

// toMap converts v to a map[string]any (or an []any for slices at the top
// level) via JSON round-trip.
func toMap(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- AST ---

type node interface {
	eval(v any) bool
}

type alwaysTrue struct{}

func (alwaysTrue) eval(v any) bool { return true }

type orNode struct{ l, r node }

func (n orNode) eval(v any) bool { return n.l.eval(v) || n.r.eval(v) }

type andNode struct{ l, r node }

func (n andNode) eval(v any) bool { return n.l.eval(v) && n.r.eval(v) }

type notNode struct{ x node }

func (n notNode) eval(v any) bool { return !n.x.eval(v) }

type term struct {
	key string
	op  string
	val string

	// re is a lazily-compiled regexp for ~ / !~.
	re *regexp.Regexp
}

func (t *term) eval(v any) bool {
	values := lookup(v, t.key)
	for _, x := range values {
		if compare(x, t.op, t.val, t) {
			return true
		}
	}
	return false
}

func usesContains(n node) bool {
	switch nn := n.(type) {
	case *term:
		return nn.op == ":"
	case orNode:
		return usesContains(nn.l) || usesContains(nn.r)
	case andNode:
		return usesContains(nn.l) || usesContains(nn.r)
	case notNode:
		return usesContains(nn.x)
	}
	return false
}

// --- lookup ---

// lookup returns every value reachable from root by following the dotted path
// in key, expanding arrays along the way. If key is empty it returns [root].
func lookup(root any, key string) []any {
	if key == "" {
		return []any{root}
	}
	parts := strings.Split(key, ".")
	acc := []any{root}
	for _, p := range parts {
		var next []any
		for _, cur := range acc {
			for _, v := range descend(cur, p) {
				next = append(next, v)
			}
		}
		if len(next) == 0 {
			return nil
		}
		acc = next
	}
	return acc
}

// descend applies one path component to node, expanding arrays.
func descend(node any, key string) []any {
	switch t := node.(type) {
	case map[string]any:
		if v, ok := t[key]; ok {
			return []any{v}
		}
		return nil
	case []any:
		var out []any
		for _, elem := range t {
			out = append(out, descend(elem, key)...)
		}
		return out
	}
	return nil
}

// compare applies op to a single lookup result and a filter RHS value.
func compare(lhs any, op, rhs string, t *term) bool {
	switch op {
	case ":":
		return colonMatch(lhs, rhs)
	case "=":
		return equal(lhs, rhs)
	case "!=":
		return !equal(lhs, rhs)
	case "~", "!~":
		if t.re == nil {
			re, err := regexp.Compile(rhs)
			if err != nil {
				return false
			}
			t.re = re
		}
		ok := t.re.MatchString(stringify(lhs))
		if op == "!~" {
			return !ok
		}
		return ok
	case "<", "<=", ">", ">=":
		return numCompare(lhs, rhs, op)
	}
	return false
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func colonMatch(lhs any, rhs string) bool {
	rhsLower := strings.ToLower(rhs)
	switch t := lhs.(type) {
	case string:
		return strings.Contains(strings.ToLower(t), rhsLower)
	case []any:
		for _, elem := range t {
			if colonMatch(elem, rhs) {
				return true
			}
		}
		return false
	case map[string]any:
		// contains-a-key semantics
		_, ok := t[rhs]
		return ok
	}
	return strings.Contains(strings.ToLower(stringify(lhs)), rhsLower)
}

func equal(lhs any, rhs string) bool {
	return stringify(lhs) == rhs
}

func numCompare(lhs any, rhs, op string) bool {
	lf, lok := toFloat(lhs)
	rf, rok := toFloat(rhs)
	if lok && rok {
		switch op {
		case "<":
			return lf < rf
		case "<=":
			return lf <= rf
		case ">":
			return lf > rf
		case ">=":
			return lf >= rf
		}
	}
	// fall back to lexicographic
	ls := stringify(lhs)
	switch op {
	case "<":
		return ls < rhs
	case "<=":
		return ls <= rhs
	case ">":
		return ls > rhs
	case ">=":
		return ls >= rhs
	}
	return false
}

func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case string:
		f, err := strconv.ParseFloat(t, 64)
		return f, err == nil
	}
	return 0, false
}

// --- parser ---

type parser struct {
	src string
	pos int
}

func newParser(src string) *parser { return &parser{src: src} }

func (p *parser) eof() bool {
	p.skipSpace()
	return p.pos >= len(p.src)
}

func (p *parser) rest() string { return p.src[p.pos:] }

func (p *parser) skipSpace() {
	for p.pos < len(p.src) && unicode.IsSpace(rune(p.src[p.pos])) {
		p.pos++
	}
}

// parseExpr — top level: orExpr.
func (p *parser) parseExpr() (node, error) { return p.parseOr() }

func (p *parser) parseOr() (node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		if !p.matchKeyword("OR") {
			return left, nil
		}
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = orNode{l: left, r: right}
	}
}

func (p *parser) parseAnd() (node, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for {
		p.skipSpace()
		if p.eof() || p.peek() == ')' {
			return left, nil
		}
		// stop at OR — that's a lower-precedence boundary
		if p.peekKeyword("OR") {
			return left, nil
		}
		// optional explicit AND
		p.matchKeyword("AND")
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = andNode{l: left, r: right}
	}
}

func (p *parser) parseNot() (node, error) {
	if p.matchKeyword("NOT") {
		inner, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return notNode{x: inner}, nil
	}
	return p.parseAtom()
}

func (p *parser) parseAtom() (node, error) {
	p.skipSpace()
	if p.peek() == '(' {
		p.pos++
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if p.peek() != ')' {
			return nil, fmt.Errorf("expected ')' at position %d", p.pos)
		}
		p.pos++
		return inner, nil
	}
	return p.parseTerm()
}

func (p *parser) parseTerm() (node, error) {
	p.skipSpace()
	key, err := p.readKey()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	op, err := p.readOp()
	if err != nil {
		return nil, fmt.Errorf("in term for key %q: %w", key, err)
	}
	p.skipSpace()
	val, err := p.readValue()
	if err != nil {
		return nil, err
	}
	return &term{key: key, op: op, val: val}, nil
}

func (p *parser) peek() byte {
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

// isKeyByte reports whether c can appear in a key path (letters, digits,
// underscore, dot; `-` also allowed for hyphenated field names).
func isKeyByte(c byte) bool {
	if c == '_' || c == '.' || c == '-' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= 'A' && c <= 'Z' {
		return true
	}
	return false
}

func (p *parser) readKey() (string, error) {
	start := p.pos
	for p.pos < len(p.src) && isKeyByte(p.src[p.pos]) {
		p.pos++
	}
	if p.pos == start {
		return "", fmt.Errorf("expected key at position %d, got %q", p.pos, p.rest())
	}
	return p.src[start:p.pos], nil
}

func (p *parser) readOp() (string, error) {
	if p.pos >= len(p.src) {
		return "", fmt.Errorf("expected operator at end of input")
	}
	// two-char ops first
	if p.pos+1 < len(p.src) {
		two := p.src[p.pos : p.pos+2]
		switch two {
		case "!=", "<=", ">=", "!~":
			p.pos += 2
			return two, nil
		}
	}
	one := p.src[p.pos]
	switch one {
	case ':', '=', '<', '>', '~':
		p.pos++
		return string(one), nil
	}
	return "", fmt.Errorf("expected operator at position %d, got %q", p.pos, p.rest())
}

func (p *parser) readValue() (string, error) {
	if p.pos >= len(p.src) {
		return "", fmt.Errorf("expected value at end of input")
	}
	c := p.src[p.pos]
	if c == '"' || c == '\'' {
		return p.readQuoted(c)
	}
	start := p.pos
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if unicode.IsSpace(rune(c)) || c == ')' {
			break
		}
		p.pos++
	}
	return p.src[start:p.pos], nil
}

func (p *parser) readQuoted(quote byte) (string, error) {
	p.pos++ // consume opening quote
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != quote {
		if p.src[p.pos] == '\\' && p.pos+1 < len(p.src) {
			p.pos += 2
			continue
		}
		p.pos++
	}
	if p.pos >= len(p.src) {
		return "", fmt.Errorf("unterminated quoted string")
	}
	s := p.src[start:p.pos]
	p.pos++ // consume closing quote
	// unescape simple backslash sequences
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\'`, `'`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s, nil
}

// matchKeyword consumes kw (case-insensitive) if present at the cursor. Only
// matches when kw is followed by whitespace, '(' or EOF (so it doesn't eat the
// prefix of a key like "OR" inside "OR_ID").
func (p *parser) matchKeyword(kw string) bool {
	p.skipSpace()
	if !p.peekKeyword(kw) {
		return false
	}
	p.pos += len(kw)
	return true
}

func (p *parser) peekKeyword(kw string) bool {
	if p.pos+len(kw) > len(p.src) {
		return false
	}
	if !strings.EqualFold(p.src[p.pos:p.pos+len(kw)], kw) {
		return false
	}
	if p.pos+len(kw) == len(p.src) {
		return true
	}
	next := p.src[p.pos+len(kw)]
	if unicode.IsSpace(rune(next)) || next == '(' {
		return true
	}
	return false
}
