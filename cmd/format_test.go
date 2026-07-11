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
		{in: "csv", wantErr: true},          // columnar formats require fields
		{in: "table(", wantErr: true},       // missing close paren
		{in: "table()", wantErr: true},      // empty field list
		{in: "bogus", wantErr: true},                             // bare unknown format
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
