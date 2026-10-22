package cmd

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestAppTopologySubcommands(t *testing.T) {
	assertSubcommands(t, appTopologyCmd, []string{"domains", "locations", "resources-graph"})
	assertSubcommands(t, appTopologyDomainsCmd, []string{"describe", "list", "schema"})
	assertSubcommands(t, appTopologySchemaCmd, []string{"describe"})
	assertSubcommands(t, appTopologyLocCmd, []string{"describe", "list"})
	assertSubcommands(t, appTopologyGraphCmd, []string{"generate"})
}

func TestAppTopologyNormalizeDomains(t *testing.T) {
	loc := "projects/p/locations/global"
	got := appTopologyNormalizeDomains([]string{
		" /SRE/ ",
		"domains/DEVOPS",
		"locations/us/domains/X",
		"projects/q/locations/global/domains/Y",
		"https://apptopology.googleapis.com/v1/projects/r/locations/global/domains/Z",
		"https://gw.internal/proxy/domains/W",
		"SRE",
		"",
	}, loc)
	want := []string{
		loc + "/domains/SRE",
		loc + "/domains/DEVOPS",
		"projects/p/locations/us/domains/X",
		"projects/q/locations/global/domains/Y",
		"projects/r/locations/global/domains/Z",
		loc + "/domains/W",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestAppTopologySnakeToCamel(t *testing.T) {
	cases := map[string]string{
		"starting_node":            "startingNode",
		"label-matcher-expr":       "labelMatcherExpr",
		"STARTING_NODE":            "startingNode",
		"EDGE":                     "edge",
		"StartingNode":             "startingNode",
		"startingNode":             "startingNode",
		"label_properties_pattern": "labelPropertiesPattern",
	}
	for in, want := range cases {
		if got := appTopologySnakeToCamel(in); got != want {
			t.Errorf("appTopologySnakeToCamel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAppTopologyParsePattern(t *testing.T) {
	got, err := appTopologyParsePattern(`{"starting_node": {"label_matcher_expr": "AppHub"}, "edges": [{"direction": " to "}]}`, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"startingNode": map[string]any{"labelMatcherExpr": "AppHub"},
		"edges":        []any{map[string]any{"direction": "TO"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, err := appTopologyParsePattern("# only a comment\n", "", nil); err != nil || got != nil {
		t.Errorf("comment-only pattern = %v, %v; want nil, nil", got, err)
	}
	if got, err := appTopologyParsePattern("", "-", strings.NewReader("a: 1\n")); err != nil || got["a"] != 1 {
		t.Errorf("stdin pattern = %v, %v", got, err)
	}
	if _, err := appTopologyParsePattern("[1, 2]", "", nil); err == nil {
		t.Error("expected error for non-object pattern")
	}
}

func TestAppTopologySchemaName(t *testing.T) {
	d := "projects/p/locations/global/domains/schema"
	if got := appTopologySchemaName(d); got != d+"/schema" {
		t.Errorf("got %q", got)
	}
	if got := appTopologySchemaName(d + "/schema"); got != d+"/schema" {
		t.Errorf("got %q", got)
	}
}

func TestEmitAppTopologyDomainsList(t *testing.T) {
	defer func(l int64, s []string) { flagATLimit, flagATSortBy = l, s }(flagATLimit, flagATSortBy)
	flagATLimit, flagATSortBy = 2, []string{"~name"}
	records := []map[string]any{
		{"name": "projects/p/locations/global/domains/A"},
		{"name": "projects/p/locations/global/domains/C"},
		{"name": "projects/p/locations/global/domains/B"},
	}
	var out, errOut bytes.Buffer
	scope := func(r map[string]any) map[string]any {
		n := r["name"].(string)
		return map[string]any{"name": n[strings.LastIndex(n, "/")+1:]}
	}
	if err := emitAppTopologyList(&out, &errOut, records, "table(name)", scope); err != nil {
		t.Fatal(err)
	}
	if want := "NAME\nC\nB\n"; out.String() != want {
		t.Errorf("out = %q, want %q", out.String(), want)
	}

	out.Reset()
	if err := emitAppTopologyList(&out, &errOut, nil, "table(name)", nil); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 || errOut.String() != "Listed 0 items.\n" {
		t.Errorf("empty list: out=%q err=%q", out.String(), errOut.String())
	}
}
