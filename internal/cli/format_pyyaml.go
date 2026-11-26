package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// --- gcloud-python style YAML (#2016) ---
//
// emitPyYAML writes JSON-like values (maps, slices, strings, bools, numbers,
// json.Number, nil) the way gcloud-python's yaml printer (ruamel, block
// style, indent 2) does: sorted keys, sequences at their parent key's indent,
// strings plain unless they would read back as another type or are not
// valid plain scalars, in which case they are single-quoted (double-quoted
// when they contain a single quote). Long plain
// strings are not folded at 80 columns as ruamel would.

// pyYAMLImplicit matches plain scalars that gcloud-python's ruamel quotes
// because they would resolve to another type: null, true/false (yes/no/on/off
// stay plain), YAML 1.1 ints and floats, and timestamps.
func pyYAMLImplicit() *regexp.Regexp {
	return regexp.MustCompile(`^(?:~|null|Null|NULL` +
		`|true|True|TRUE|false|False|FALSE` +
		`|[-+]?(?:0|[1-9][0-9_]*)` +
		`|[-+]?0b[0-1_]+|[-+]?0o?[0-7_]+|[-+]?0x[0-9a-fA-F_]+` +
		`|[-+]?(?:[0-9][0-9_]*)?\.[0-9_]*(?:[eE][-+]?[0-9]+)?|[-+]?[0-9][0-9_]*(?:[eE][-+]?[0-9]+)` +
		`|[-+]?\.(?:inf|Inf|INF)|\.(?:nan|NaN|NAN)` +
		`|[0-9]{4}-[0-9]{1,2}-[0-9]{1,2}(?:(?:[Tt]|[ \t]+)[0-9]{1,2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]*)?` +
		`(?:[ \t]*(?:Z|[-+][0-9]{1,2}(?::[0-9]{2})?))?)?` +
		`)$`)
}

// pyYAMLNeedsQuotes reports whether s cannot be written as a plain scalar.
func pyYAMLNeedsQuotes(s string, implicit *regexp.Regexp) bool {
	if s == "" || implicit.MatchString(s) {
		return true
	}
	if strings.TrimSpace(s) != s || strings.ContainsAny(s, "\n\t\r") {
		return true
	}
	if strings.ContainsRune("-?:,[]{}#&*!|>'\"%@`", rune(s[0])) {
		// "-", "?" and ":" start a plain scalar only when followed by a
		// non-space character.
		if !(strings.ContainsRune("-?:", rune(s[0])) && len(s) > 1 && s[1] != ' ') {
			return true
		}
	}
	return strings.Contains(s, ": ") || strings.Contains(s, " #") || strings.HasSuffix(s, ":")
}

type pyYAMLEncoder struct {
	w        io.Writer
	implicit *regexp.Regexp
	err      error
}

func (e *pyYAMLEncoder) printf(format string, a ...any) {
	if e.err == nil {
		_, e.err = fmt.Fprintf(e.w, format, a...)
	}
}

func (e *pyYAMLEncoder) scalar(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		if pyYAMLNeedsQuotes(t, e.implicit) {
			if strings.Contains(t, "'") {
				return strconv.Quote(t)
			}
			return "'" + t + "'"
		}
		return t
	case bool:
		return strconv.FormatBool(t)
	case json.Number:
		return t.String()
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	default:
		return fmt.Sprint(t)
	}
}

// pyYAMLAsMap and pyYAMLAsList normalize the container types emitPyYAML
// accepts.
func pyYAMLAsMap(v any) (map[string]any, bool) {
	switch t := v.(type) {
	case map[string]any:
		return t, true
	case map[string]string:
		m := make(map[string]any, len(t))
		for k, s := range t {
			m[k] = s
		}
		return m, true
	}
	return nil, false
}

func pyYAMLAsList(v any) ([]any, bool) {
	switch t := v.(type) {
	case []any:
		return t, true
	case []string:
		l := make([]any, len(t))
		for i, s := range t {
			l[i] = s
		}
		return l, true
	}
	return nil, false
}

// value writes v, which follows "key:" or "- " on the current line; indent
// is the indent of that key or dash.
func (e *pyYAMLEncoder) value(v any, indent int, afterDash bool) {
	if m, ok := pyYAMLAsMap(v); ok {
		if len(m) == 0 {
			e.printf(" {}\n")
			return
		}
		if afterDash {
			e.printf(" ")
			e.mapping(m, indent+2, true)
			return
		}
		e.printf("\n")
		e.mapping(m, indent+2, false)
		return
	}
	if l, ok := pyYAMLAsList(v); ok {
		if len(l) == 0 {
			e.printf(" []\n")
			return
		}
		if afterDash {
			e.printf(" ")
			e.sequence(l, indent+2, true)
			return
		}
		e.printf("\n")
		e.sequence(l, indent, false)
		return
	}
	e.printf(" %s\n", e.scalar(v))
}

// mapping writes m's entries at indent; inline means the first entry
// continues the current line (after "- ").
func (e *pyYAMLEncoder) mapping(m map[string]any, indent int, inline bool) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		if !(inline && i == 0) {
			e.printf("%s", strings.Repeat(" ", indent))
		}
		e.printf("%s:", e.scalar(k))
		e.value(m[k], indent, false)
	}
}

// sequence writes l's items as "- " lines at indent.
func (e *pyYAMLEncoder) sequence(l []any, indent int, inline bool) {
	for i, item := range l {
		if !(inline && i == 0) {
			e.printf("%s", strings.Repeat(" ", indent))
		}
		e.printf("-")
		e.value(item, indent, true)
	}
}

// emitPyYAMLDocuments prints each record as a YAML document preceded by
// "---", as gcloud-python's default yaml format does for a list command.
func emitPyYAMLDocuments(w io.Writer, records []map[string]any) error {
	e := &pyYAMLEncoder{w: w, implicit: pyYAMLImplicit()}
	for _, r := range records {
		e.printf("---\n")
		if len(r) == 0 {
			e.printf("{}\n")
			continue
		}
		e.mapping(r, 0, false)
	}
	return e.err
}
