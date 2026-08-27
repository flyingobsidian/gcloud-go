package cmd

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

// sampleProject mirrors the shape of a Cloud Resource Manager Project response
// so the format tests can exercise the same rendering paths that projects
// describe uses in production.
type sampleProject struct {
	CreateTime     string `json:"createTime"`
	LifecycleState string `json:"lifecycleState"`
	Name           string `json:"name"`
	ProjectId      string `json:"projectId"`
	ProjectNumber  string `json:"projectNumber"`
}

func fixtureProject() sampleProject {
	return sampleProject{
		CreateTime:     "2026-07-11T11:40:24.977Z",
		LifecycleState: "ACTIVE",
		Name:           "MY_PROJECT",
		ProjectId:      "MY_PROJECT",
		ProjectNumber:  "123456789012",
	}
}

func TestParseFormat(t *testing.T) {
	cases := []struct {
		in         string
		wantName   string
		wantFields []string
		wantErr    bool
	}{
		{in: "", wantName: ""},
		{in: "yaml", wantName: "yaml"},
		{in: "json", wantName: "json"},
		{in: "table(a,b,c)", wantName: "table", wantFields: []string{"a", "b", "c"}},
		{in: "csv(a, b )", wantName: "csv", wantFields: []string{"a", "b"}},
		{in: "value(x)", wantName: "value", wantFields: []string{"x"}},
		{in: "text(a.b,c)", wantName: "text", wantFields: []string{"a.b", "c"}},
		{in: "config(a)", wantName: "config", wantFields: []string{"a"}},
		{in: "get(a)", wantName: "get", wantFields: []string{"a"}},
		{in: "csv", wantErr: true},                                     // columnar formats require fields
		{in: "table(", wantErr: true},                                  // missing close paren
		{in: "table()", wantErr: true},                                 // empty field list
		{in: "bogus", wantErr: true},                                   // bare unknown format
		{in: "bogus(a)", wantName: "bogus", wantFields: []string{"a"}}, // parens variant is caught downstream by emitFormattedTo
	}
	for _, c := range cases {
		name, fields, err := parseFormat(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseFormat(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseFormat(%q) err = %v", c.in, err)
			continue
		}
		if name != c.wantName || !reflect.DeepEqual(fields, c.wantFields) {
			t.Errorf("parseFormat(%q) = (%q, %v), want (%q, %v)",
				c.in, name, fields, c.wantName, c.wantFields)
		}
	}
}

func TestCamelToSnake(t *testing.T) {
	cases := []struct{ in, want string }{
		{"createTime", "create_time"},
		{"projectId", "project_id"},
		{"projectNumber", "project_number"},
		{"lifecycleState", "lifecycle_state"},
		{"name", "name"},
	}
	for _, c := range cases {
		if got := camelToSnake(c.in); got != c.want {
			t.Errorf("camelToSnake(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBaseField(t *testing.T) {
	cases := []struct{ in, want string }{
		{"name", "name"},
		{"a.b.c", "c"},
		{"networkInterfaces[0].networkIP", "networkIP"},
		{"labels[0]", "labels"},
	}
	for _, c := range cases {
		if got := baseField(c.in); got != c.want {
			t.Errorf("baseField(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFieldValueDottedAndIndexed(t *testing.T) {
	m := map[string]any{
		"a": map[string]any{
			"b": "hello",
		},
		"arr": []any{
			map[string]any{"x": "first"},
			map[string]any{"x": "second"},
		},
		"missing": nil,
	}
	cases := []struct{ field, want string }{
		{"a.b", "hello"},
		{"arr[0].x", "first"},
		{"arr[1].x", "second"},
		{"arr[7].x", ""},
		{"missing", ""},
		{"nope", ""},
	}
	for _, c := range cases {
		if got := fieldValue(m, c.field); got != c.want {
			t.Errorf("fieldValue(%q) = %q, want %q", c.field, got, c.want)
		}
	}
}

func runFormat(t *testing.T, v any, format string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := emitFormattedTo(&buf, v, format); err != nil {
		t.Fatalf("emitFormattedTo(%q) err = %v", format, err)
	}
	return buf.String()
}

func TestEmitYAML(t *testing.T) {
	got := runFormat(t, fixtureProject(), "yaml")
	// yaml.v3 double-quotes strings starting with a digit, which is fine.
	for _, need := range []string{
		"createTime:",
		"lifecycleState: ACTIVE",
		"name: MY_PROJECT",
		"projectId: MY_PROJECT",
		"projectNumber: ",
	} {
		if !strings.Contains(got, need) {
			t.Errorf("yaml output missing %q; got:\n%s", need, got)
		}
	}
}

func TestEmitJSON(t *testing.T) {
	got := runFormat(t, fixtureProject(), "json")
	if !strings.Contains(got, `"projectId": "MY_PROJECT"`) {
		t.Errorf("json output missing projectId; got:\n%s", got)
	}
	if !strings.Contains(got, `"lifecycleState": "ACTIVE"`) {
		t.Errorf("json output missing lifecycleState; got:\n%s", got)
	}
}

func TestEmitCSV(t *testing.T) {
	got := runFormat(t, fixtureProject(), "csv(createTime,lifecycleState,name,projectId,projectNumber)")
	want := "create_time,lifecycle_state,name,project_id,project_number\n" +
		"2026-07-11T11:40:24.977Z,ACTIVE,MY_PROJECT,MY_PROJECT,123456789012\n"
	if got != want {
		t.Errorf("csv output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestEmitCSVEscapesCommas(t *testing.T) {
	v := map[string]any{"a": "one,two", "b": `he said "hi"`}
	got := runFormat(t, v, "csv(a,b)")
	want := "a,b\n" +
		"\"one,two\",\"he said \"\"hi\"\"\"\n"
	if got != want {
		t.Errorf("csv escape mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestEmitTable(t *testing.T) {
	got := runFormat(t, fixtureProject(), "table(createTime,lifecycleState,name,projectId,projectNumber)")
	want := "CREATE_TIME               LIFECYCLE_STATE  NAME        PROJECT_ID  PROJECT_NUMBER\n" +
		"2026-07-11T11:40:24.977Z  ACTIVE           MY_PROJECT  MY_PROJECT  123456789012\n"
	if got != want {
		t.Errorf("table output mismatch:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestEmitText(t *testing.T) {
	got := runFormat(t, fixtureProject(), "text(createTime,lifecycleState,name,projectId,projectNumber)")
	want := "createTime:     2026-07-11T11:40:24.977Z\n" +
		"lifecycleState: ACTIVE\n" +
		"name:           MY_PROJECT\n" +
		"projectId:      MY_PROJECT\n" +
		"projectNumber:  123456789012\n"
	if got != want {
		t.Errorf("text output mismatch:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestEmitValue(t *testing.T) {
	got := runFormat(t, fixtureProject(), "value(createTime,lifecycleState,name,projectId,projectNumber)")
	want := "2026-07-11T11:40:24.977Z\tACTIVE\tMY_PROJECT\tMY_PROJECT\t123456789012\n"
	if got != want {
		t.Errorf("value output mismatch:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestEmitConfig(t *testing.T) {
	got := runFormat(t, fixtureProject(), "config(createTime,lifecycleState,name,projectId,projectNumber)")
	want := "createTime = 2026-07-11T11:40:24.977Z\n" +
		"lifecycleState = ACTIVE\n" +
		"name = MY_PROJECT\n" +
		"projectId = MY_PROJECT\n" +
		"projectNumber = 123456789012\n"
	if got != want {
		t.Errorf("config output mismatch:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestEmitGet(t *testing.T) {
	got := runFormat(t, fixtureProject(), "get(projectId)")
	if got != "MY_PROJECT\n" {
		t.Errorf("get output = %q, want %q", got, "MY_PROJECT\n")
	}
}

func TestEmitFormattedOnList(t *testing.T) {
	list := []sampleProject{fixtureProject(), {
		CreateTime:     "2026-08-01T00:00:00Z",
		LifecycleState: "ACTIVE",
		Name:           "OTHER",
		ProjectId:      "OTHER",
		ProjectNumber:  "1",
	}}
	got := runFormat(t, list, "value(projectId,projectNumber)")
	want := "MY_PROJECT\t123456789012\nOTHER\t1\n"
	if got != want {
		t.Errorf("list value output mismatch:\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestEmitFormattedUnknown(t *testing.T) {
	var buf bytes.Buffer
	if err := emitFormattedTo(&buf, fixtureProject(), "quux(a)"); err == nil {
		t.Fatal("expected error for unknown format")
	}
}

// TestEmitFlattened covers the `flattened` format added for #1742.
func TestEmitFlattened(t *testing.T) {
	got := runFormat(t, fixtureProject(), "flattened")
	// Deterministic key order (sorted). Compare exact bytes.
	want := "createTime:     2026-07-11T11:40:24.977Z\n" +
		"lifecycleState: ACTIVE\n" +
		"name:           MY_PROJECT\n" +
		"projectId:      MY_PROJECT\n" +
		"projectNumber:  123456789012\n"
	if got != want {
		t.Errorf("flattened output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestEmitFlattenedWithFields exercises the parenthesised form
// `flattened(field1,field2)` -- restricts output to those subtrees.
func TestEmitFlattenedWithFields(t *testing.T) {
	got := runFormat(t, fixtureProject(), "flattened(projectId,name)")
	want := "projectId: MY_PROJECT\n" +
		"name:      MY_PROJECT\n"
	if got != want {
		t.Errorf("flattened(fields) output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestEmitFlattenedQuotesMultilineValue covers #1745: a value containing
// newlines must be emitted as one physical line with `\n` escape sequences
// inside double quotes -- matching gcloud-python's FlattenedPrinter.
func TestEmitFlattenedQuotesMultilineValue(t *testing.T) {
	obj := map[string]any{
		"commonInstanceMetadata": map[string]any{
			"items": []any{
				map[string]any{
					"key":   "ssh-keys",
					"value": "user:ssh-rsa KEY1 user@host\nuser:ssh-rsa KEY2 user@host\nuser:ssh-rsa KEY3 user@host",
				},
			},
		},
	}
	got := runFormat(t, obj, "flattened(commonInstanceMetadata.items)")
	want := "commonInstanceMetadata.items[0].key:   ssh-keys\n" +
		"commonInstanceMetadata.items[0].value: \"user:ssh-rsa KEY1 user@host\\nuser:ssh-rsa KEY2 user@host\\nuser:ssh-rsa KEY3 user@host\"\n"
	if got != want {
		t.Errorf("multiline flattened mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
	// The value line must render on exactly one physical line: the raw
	// newlines have to become the two-character escape sequence \n.
	if lines := strings.Count(got, "\n"); lines != 2 {
		t.Errorf("expected 2 lines of output, got %d:\n%s", lines, got)
	}
}

// TestFlattenedValueDisplay exercises the per-value quoting rules directly,
// so regressions in the escape table (\, ", \f, \n, \r, \t) are caught even
// when they don't happen to appear in a printer fixture.
func TestFlattenedValueDisplay(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"plain", "plain"},
		{"one line", "one line"},
		// Newline forces quoting; embedded \n becomes the two-char escape.
		{"a\nb", `"a\nb"`},
		// Leading/trailing whitespace forces quoting.
		{" leading", `" leading"`},
		{"trailing ", `"trailing "`},
		// Once we quote, backslashes and quotes must be escaped too.
		{"has \"quote\" and \\slash\n", `"has \"quote\" and \\slash\n"`},
		// Tabs / carriage returns / form feeds get their escapes only when
		// quoting is already in effect (here forced by the newline).
		{"tab\there\n", `"tab\there\n"`},
		{"cr\rhere\n", `"cr\rhere\n"`},
		{"ff\fhere\n", `"ff\fhere\n"`},
		// A mid-string tab alone doesn't trigger quoting in gcloud-python;
		// stay bug-compatible so operators diffing output see identical text.
		{"mid\ttab", "mid\ttab"},
	}
	for _, c := range cases {
		if got := flattenedValueDisplay(c.in); got != c.want {
			t.Errorf("flattenedValueDisplay(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestEmitFlattenedNested asserts that nested objects and arrays produce
// dotted-path leaves (metadata.items[0].key etc.).
func TestEmitFlattenedNested(t *testing.T) {
	// Fake Project with a common-metadata style shape.
	obj := map[string]any{
		"name": "p",
		"metadata": map[string]any{
			"items": []any{
				map[string]any{"key": "K1", "value": "V1"},
				map[string]any{"key": "K2", "value": "V2"},
			},
		},
	}
	got := runFormat(t, obj, "flattened")
	// Keys sort alphabetically at each level: metadata < name; within
	// each item, key < value.
	want := "metadata.items[0].key:   K1\n" +
		"metadata.items[0].value: V1\n" +
		"metadata.items[1].key:   K2\n" +
		"metadata.items[1].value: V2\n" +
		"name:                    p\n"
	if got != want {
		t.Errorf("nested flattened mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
