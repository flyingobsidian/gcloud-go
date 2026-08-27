package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// emitFormatted renders v to stdout in one of the supported gcloud output
// formats: yaml (default), json, csv, table, text, value, config, get. The
// columnar formats (csv/table/text/value/config/get) take a parenthesised
// field list, e.g. table(projectId,state).
//
// This is the single implementation of gcloud-go's --format flag; all callers
// route through here so behavior stays consistent across commands.
func emitFormatted(v any, format string) error {
	return emitFormattedTo(os.Stdout, v, format)
}

// emitFormattedTo is the io.Writer-taking form of emitFormatted, used by tests.
func emitFormattedTo(w io.Writer, v any, format string) error {
	name, fields, err := parseFormat(format)
	if err != nil {
		return err
	}

	rows, err := jsonRows(v)
	if err != nil {
		return err
	}

	switch name {
	case "", "yaml":
		return emitYAML(w, v)
	case "json":
		return emitJSON(w, v)
	case "csv":
		return emitCSV(w, rows, fields)
	case "table":
		return emitTable(w, rows, fields)
	case "text":
		return emitText(w, rows, fields)
	case "value":
		return emitValue(w, rows, fields)
	case "config":
		return emitConfig(w, rows, fields)
	case "get":
		return emitGet(w, rows, fields)
	case "flattened":
		return emitFlattened(w, rows, fields)
	default:
		return fmt.Errorf("unknown format %q", format)
	}
}

// parseFormat splits a format string into its name and optional field list.
// e.g. "table(projectId,state)" -> ("table", ["projectId", "state"], nil);
// "yaml" -> ("yaml", nil, nil); "csv" -> error (columnar formats require fields).
func parseFormat(format string) (string, []string, error) {
	format = strings.TrimSpace(format)
	if format == "" {
		return "", nil, nil
	}
	open := strings.IndexByte(format, '(')
	if open < 0 {
		switch format {
		case "yaml", "json":
			return format, nil, nil
		case "flattened":
			// `flattened` may be used without a field list -- it prints
			// every leaf of the record. Return empty fields to signal
			// "all fields" to emitFlattened.
			return format, nil, nil
		case "csv", "table", "text", "value", "config", "get":
			return "", nil, fmt.Errorf("format %q requires a field list, e.g. %s(field1,field2)", format, format)
		default:
			return "", nil, fmt.Errorf("unknown format %q", format)
		}
	}
	if !strings.HasSuffix(format, ")") {
		return "", nil, fmt.Errorf("malformed format %q (missing closing parenthesis)", format)
	}
	name := format[:open]
	inner := format[open+1 : len(format)-1]
	var fields []string
	for _, f := range strings.Split(inner, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		fields = append(fields, f)
	}
	if len(fields) == 0 {
		return "", nil, fmt.Errorf("format %q requires at least one field", format)
	}
	return name, fields, nil
}

// jsonRows normalizes any Go value into a slice of maps by round-tripping
// through JSON. A struct/object becomes a single-row slice; an array becomes
// the corresponding rows.
func jsonRows(v any) ([]map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var single map[string]any
	if err := json.Unmarshal(data, &single); err == nil {
		return []map[string]any{single}, nil
	}
	var many []map[string]any
	if err := json.Unmarshal(data, &many); err == nil {
		return many, nil
	}
	return nil, fmt.Errorf("value must be an object or an array of objects")
}

func emitYAML(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var m any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	out, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, string(out))
	return err
}

func emitJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func emitCSV(w io.Writer, rows []map[string]any, fields []string) error {
	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = camelToSnake(baseField(f))
	}
	if _, err := fmt.Fprintln(w, strings.Join(headers, ",")); err != nil {
		return err
	}
	for _, row := range rows {
		vals := make([]string, len(fields))
		for i, f := range fields {
			vals[i] = csvEscape(fieldValue(row, f))
		}
		if _, err := fmt.Fprintln(w, strings.Join(vals, ",")); err != nil {
			return err
		}
	}
	return nil
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func emitTable(w io.Writer, rows []map[string]any, fields []string) error {
	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = strings.ToUpper(camelToSnake(baseField(f)))
	}
	widths := make([]int, len(fields))
	for i, h := range headers {
		widths[i] = len(h)
	}
	cells := make([][]string, len(rows))
	for r, row := range rows {
		cells[r] = make([]string, len(fields))
		for i, f := range fields {
			v := fieldValue(row, f)
			cells[r][i] = v
			if len(v) > widths[i] {
				widths[i] = len(v)
			}
		}
	}

	if err := writePadded(w, headers, widths); err != nil {
		return err
	}
	for _, row := range cells {
		if err := writePadded(w, row, widths); err != nil {
			return err
		}
	}
	return nil
}

// writePadded prints cells space-padded to widths, with two spaces between
// columns and no trailing whitespace on the final column.
func writePadded(w io.Writer, cells []string, widths []int) error {
	for i, c := range cells {
		if i > 0 {
			if _, err := fmt.Fprint(w, "  "); err != nil {
				return err
			}
		}
		if i == len(cells)-1 {
			if _, err := fmt.Fprint(w, c); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "%-*s", widths[i], c); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func emitText(w io.Writer, rows []map[string]any, fields []string) error {
	labels := make([]string, len(fields))
	maxLen := 0
	for i, f := range fields {
		labels[i] = baseField(f)
		if len(labels[i]) > maxLen {
			maxLen = len(labels[i])
		}
	}
	for r, row := range rows {
		if r > 0 {
			if _, err := fmt.Fprintln(w, "---"); err != nil {
				return err
			}
		}
		for i, f := range fields {
			pad := strings.Repeat(" ", maxLen-len(labels[i]))
			if _, err := fmt.Fprintf(w, "%s:%s %s\n", labels[i], pad, fieldValue(row, f)); err != nil {
				return err
			}
		}
	}
	return nil
}

func emitValue(w io.Writer, rows []map[string]any, fields []string) error {
	for _, row := range rows {
		vals := make([]string, len(fields))
		for i, f := range fields {
			vals[i] = fieldValue(row, f)
		}
		if _, err := fmt.Fprintln(w, strings.Join(vals, "\t")); err != nil {
			return err
		}
	}
	return nil
}

func emitConfig(w io.Writer, rows []map[string]any, fields []string) error {
	for r, row := range rows {
		if r > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		for _, f := range fields {
			if _, err := fmt.Fprintf(w, "%s = %s\n", baseField(f), fieldValue(row, f)); err != nil {
				return err
			}
		}
	}
	return nil
}

func emitGet(w io.Writer, rows []map[string]any, fields []string) error {
	for _, row := range rows {
		vals := make([]string, len(fields))
		for i, f := range fields {
			vals[i] = fieldValue(row, f)
		}
		if _, err := fmt.Fprintln(w, strings.Join(vals, "\t")); err != nil {
			return err
		}
	}
	return nil
}

// flattenedEntry is a single path/value pair produced by flattenLeaves.
type flattenedEntry struct{ key, val string }

// emitFlattened writes v as gcloud's "flattened" format: one dotted-path
// KEY: VALUE line per leaf. When fields is empty, every leaf is emitted;
// when fields is provided, the output is restricted to those subtrees.
func emitFlattened(w io.Writer, rows []map[string]any, fields []string) error {
	for r, row := range rows {
		if r > 0 {
			// Records after the first are separated by ---, matching
			// gcloud-python's flattened multi-record output.
			if _, err := fmt.Fprintln(w, "---"); err != nil {
				return err
			}
		}
		var entries []flattenedEntry
		if len(fields) == 0 {
			entries = flattenLeaves("", row)
		} else {
			for _, f := range fields {
				val, ok := lookupPath(row, f)
				if !ok {
					continue
				}
				entries = append(entries, flattenLeaves(f, val)...)
			}
		}
		width := 0
		for _, e := range entries {
			if len(e.key) > width {
				width = len(e.key)
			}
		}
		for _, e := range entries {
			pad := strings.Repeat(" ", width-len(e.key))
			if _, err := fmt.Fprintf(w, "%s:%s %s\n", e.key, pad, flattenedValueDisplay(e.val)); err != nil {
				return err
			}
		}
	}
	return nil
}

// flattenLeaves walks v and returns dotted-path leaf entries, using prefix
// as the accumulated path.
func flattenLeaves(prefix string, v any) []flattenedEntry {
	switch t := v.(type) {
	case map[string]any:
		// Deterministic key order so callers see a stable table.
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var out []flattenedEntry
		for _, k := range keys {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			out = append(out, flattenLeaves(p, t[k])...)
		}
		return out
	case []any:
		var out []flattenedEntry
		for i, e := range t {
			p := fmt.Sprintf("%s[%d]", prefix, i)
			out = append(out, flattenLeaves(p, e)...)
		}
		return out
	default:
		return []flattenedEntry{{key: prefix, val: stringifyLeaf(v)}}
	}
}

func stringifyLeaf(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// flattenedValueDisplay renders a leaf value for the "flattened" printer.
// Each leaf must occupy a single line; when the raw value contains a newline
// or has leading/trailing whitespace, we surround it with double quotes and
// escape the special characters (\, ", \f, \n, \r, \t) so the output stays
// on one line -- mirroring gcloud-python's FlattenedPrinter._StringQuote.
func flattenedValueDisplay(s string) string {
	if !flattenedNeedsQuoting(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// flattenedNeedsQuoting reports whether a flattened value must be quoted to
// stay on a single line. Matches gcloud-python: embedded newlines or leading/
// trailing whitespace force quoting.
func flattenedNeedsQuoting(s string) bool {
	if strings.Contains(s, "\n") {
		return true
	}
	if s == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(s)
	last, _ := utf8.DecodeLastRuneInString(s)
	return unicode.IsSpace(first) || unicode.IsSpace(last)
}

// lookupPath walks m via a dotted path and returns the value at that
// location (or nil, false). Bracket indexing e.g. "items[0]" is honoured.
func lookupPath(m map[string]any, path string) (any, bool) {
	var cur any = m
	for _, part := range strings.Split(path, ".") {
		key, index, hasIndex := parseIndex(part)
		mp, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		val, ok := mp[key]
		if !ok {
			return nil, false
		}
		if hasIndex {
			arr, ok := val.([]any)
			if !ok || index < 0 || index >= len(arr) {
				return nil, false
			}
			val = arr[index]
		}
		cur = val
	}
	return cur, true
}

// fieldValue looks up a dotted field path (e.g. "networkInterfaces[0].networkIP")
// on the given JSON-shaped map and returns its stringified value, or the empty
// string if the path is missing. Values are rendered as their native fmt.Sprintf
// %v except nil/missing entries.
func fieldValue(m map[string]any, field string) string {
	parts := strings.Split(field, ".")
	var current any = m
	for _, part := range parts {
		key, index, hasIndex := parseIndex(part)
		switch c := current.(type) {
		case map[string]any:
			current = c[key]
		default:
			return ""
		}
		if hasIndex {
			arr, ok := current.([]any)
			if !ok || index < 0 || index >= len(arr) {
				return ""
			}
			current = arr[index]
		}
	}
	if current == nil {
		return ""
	}
	return fmt.Sprintf("%v", current)
}

// parseIndex extracts "name[N]" into ("name", N, true) or returns
// ("name", 0, false) when there is no bracketed index.
func parseIndex(s string) (string, int, bool) {
	open := strings.IndexByte(s, '[')
	if open < 0 || !strings.HasSuffix(s, "]") {
		return s, 0, false
	}
	name := s[:open]
	numStr := s[open+1 : len(s)-1]
	var n int
	if _, err := fmt.Sscanf(numStr, "%d", &n); err != nil {
		return s, 0, false
	}
	return name, n, true
}

// baseField returns the last dotted segment of a field path, stripped of any
// [N] index, e.g. "networkInterfaces[0].networkIP" -> "networkIP".
func baseField(field string) string {
	if i := strings.LastIndexByte(field, '.'); i >= 0 {
		field = field[i+1:]
	}
	if i := strings.IndexByte(field, '['); i >= 0 {
		field = field[:i]
	}
	return field
}

// camelToSnake converts an identifier from camelCase to snake_case.
func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}
