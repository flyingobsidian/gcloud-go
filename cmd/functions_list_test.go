package cmd

import (
	"bytes"
	"strings"
	"testing"

	cloudfunctionsv1 "google.golang.org/api/cloudfunctions/v1"
	cloudfunctions "google.golang.org/api/cloudfunctions/v2"
)

func TestFuncV1Regions(t *testing.T) {
	got := funcV1Regions([]string{"-", "europe-west2", "v2-only"}, []string{"europe-west2", "us-central1"})
	if strings.Join(got, ",") != "-,europe-west2" {
		t.Errorf("funcV1Regions = %v", got)
	}
	if got := funcV1Regions([]string{"v2-only"}, []string{"europe-west2"}); got == nil || len(got) != 0 {
		t.Errorf("funcV1Regions(v2-only) = %v, want empty", got)
	}
}

func TestEmitFuncListDefaultTable(t *testing.T) {
	functions := []any{
		&cloudfunctions.Function{
			Name: "projects/p/locations/europe-west2/functions/zeta", State: "ACTIVE", Environment: "GEN_2",
			ServiceConfig: &cloudfunctions.ServiceConfig{},
			EventTrigger: &cloudfunctions.EventTrigger{EventType: "google.cloud.storage.object.v1.finalized",
				EventFilters: []*cloudfunctions.EventFilter{{Attribute: "bucket", Value: "uploads"}}},
		},
		&cloudfunctions.Function{
			Name: "projects/p/locations/us-central1/functions/web", State: "ACTIVE", Environment: "GEN_2",
			ServiceConfig: &cloudfunctions.ServiceConfig{},
		},
		&cloudfunctionsv1.CloudFunction{
			Name: "projects/p/locations/us-central1/functions/alpha", Status: "ACTIVE", Runtime: "python312",
			EventTrigger: &cloudfunctionsv1.EventTrigger{EventType: "providers/cloud.pubsub/eventTypes/topic.publish"},
		},
	}
	var out bytes.Buffer
	if err := emitFuncList(&out, &bytes.Buffer{}, functions, ""); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"NAME STATE TRIGGER REGION ENVIRONMENT",
		"alpha ACTIVE Event Trigger us-central1 1st gen",
		"web ACTIVE HTTP Trigger us-central1 2nd gen",
		"zeta ACTIVE bucket: uploads europe-west2 2nd gen",
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != len(want) {
		t.Fatalf("got:\n%s", out.String())
	}
	for i, w := range want {
		if got := strings.Join(strings.Fields(lines[i]), " "); got != w {
			t.Errorf("line %d = %q, want %q", i, got, w)
		}
	}
}

func TestEmitFuncListEmpty(t *testing.T) {
	var out, errOut bytes.Buffer
	if err := emitFuncList(&out, &errOut, []any{}, "json"); err != nil || strings.TrimSpace(out.String()) != "[]" {
		t.Errorf("json: %q, %v", out.String(), err)
	}
	out.Reset()
	if err := emitFuncList(&out, &errOut, nil, ""); err != nil || out.Len() != 0 || errOut.String() != "Listed 0 items.\n" {
		t.Errorf("table: stdout %q stderr %q err %v", out.String(), errOut.String(), err)
	}
}

func TestFuncTrigger(t *testing.T) {
	cases := []struct {
		name string
		rec  map[string]any
		want string
	}{
		{"v1 http", map[string]any{"runtime": "go122", "httpsTrigger": map[string]any{}}, "HTTP Trigger"},
		{"v1 bucket", map[string]any{"entryPoint": "F", "gcsTrigger": "b"}, "bucket: b"},
		{"v1 event", map[string]any{"runtime": "go122", "eventTrigger": map[string]any{}}, "Event Trigger"},
		{"v2 audit", map[string]any{"serviceConfig": map[string]any{}, "eventTrigger": map[string]any{"eventType": "google.cloud.audit.log.v1.written"}}, "Cloud Audit Log"},
		{"v2 pubsub", map[string]any{"buildConfig": map[string]any{}, "eventTrigger": map[string]any{"pubsubTopic": "projects/p/topics/t"}}, "topic: t"},
		{"v2 http", map[string]any{"serviceConfig": map[string]any{}}, "HTTP Trigger"},
		{"unknown", map[string]any{}, ""},
	}
	for _, tc := range cases {
		if got := funcTrigger(tc.rec); got != tc.want {
			t.Errorf("%s: funcTrigger = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestFunctionsListRegistered(t *testing.T) {
	list, _, err := functionsCmd.Find([]string{"list"})
	if err != nil || list.RunE == nil || list.Flags().Lookup("regions") == nil || list.Flags().Lookup("v2") == nil {
		t.Fatalf("functions list not registered with --regions/--v2: %v", err)
	}
	if strings.Contains(list.Short, "Not yet implemented") {
		t.Error("functions list is still the stub")
	}
}
