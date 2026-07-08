package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	crm "google.golang.org/api/cloudresourcemanager/v3"
)

func TestOrgResourceName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1234567890", "organizations/1234567890"},
		{"organizations/1234567890", "organizations/1234567890"},
	}
	for _, c := range cases {
		if got := orgResourceName(c.in); got != c.want {
			t.Errorf("orgResourceName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestConditionsEqual(t *testing.T) {
	a := &crm.Expr{Expression: "x", Title: "t"}
	b := &crm.Expr{Expression: "x", Title: "t"}
	c := &crm.Expr{Expression: "y", Title: "t"}
	if !conditionsEqual(nil, nil) {
		t.Error("nil,nil should be equal")
	}
	if conditionsEqual(a, nil) || conditionsEqual(nil, a) {
		t.Error("nil vs non-nil should be unequal")
	}
	if !conditionsEqual(a, b) {
		t.Error("identical conditions should be equal")
	}
	if conditionsEqual(a, c) {
		t.Error("differing expressions should be unequal")
	}
}

func TestAddBindingToPolicyNewBinding(t *testing.T) {
	p := &crm.Policy{}
	if !addBindingToPolicy(p, "roles/browser", "user:a@example.com", nil) {
		t.Fatal("expected change=true when adding to empty policy")
	}
	if len(p.Bindings) != 1 || p.Bindings[0].Role != "roles/browser" {
		t.Fatalf("unexpected bindings: %+v", p.Bindings)
	}
	if !reflect.DeepEqual(p.Bindings[0].Members, []string{"user:a@example.com"}) {
		t.Errorf("members = %v", p.Bindings[0].Members)
	}
}

func TestAddBindingToPolicyExistingRole(t *testing.T) {
	p := &crm.Policy{Bindings: []*crm.Binding{{
		Role:    "roles/browser",
		Members: []string{"user:a@example.com"},
	}}}
	if !addBindingToPolicy(p, "roles/browser", "user:b@example.com", nil) {
		t.Fatal("expected change=true when adding new member")
	}
	if len(p.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(p.Bindings))
	}
	want := []string{"user:a@example.com", "user:b@example.com"}
	if !reflect.DeepEqual(p.Bindings[0].Members, want) {
		t.Errorf("members = %v, want %v", p.Bindings[0].Members, want)
	}
}

func TestAddBindingToPolicyDuplicateMember(t *testing.T) {
	p := &crm.Policy{Bindings: []*crm.Binding{{
		Role:    "roles/browser",
		Members: []string{"user:a@example.com"},
	}}}
	if addBindingToPolicy(p, "roles/browser", "user:a@example.com", nil) {
		t.Fatal("expected change=false for duplicate member")
	}
	if len(p.Bindings[0].Members) != 1 {
		t.Errorf("members should not have changed: %v", p.Bindings[0].Members)
	}
}

func TestAddBindingToPolicyDifferentCondition(t *testing.T) {
	cond := &crm.Expr{Expression: "request.time < timestamp(\"2030-01-01T00:00:00Z\")", Title: "t"}
	p := &crm.Policy{Bindings: []*crm.Binding{{
		Role:    "roles/browser",
		Members: []string{"user:a@example.com"},
	}}}
	if !addBindingToPolicy(p, "roles/browser", "user:a@example.com", cond) {
		t.Fatal("expected change=true when condition differs")
	}
	if len(p.Bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(p.Bindings))
	}
	if p.Bindings[1].Condition == nil || p.Bindings[1].Condition.Title != "t" {
		t.Errorf("unexpected new binding condition: %+v", p.Bindings[1].Condition)
	}
}

func TestRemoveBindingFromPolicy(t *testing.T) {
	p := &crm.Policy{Bindings: []*crm.Binding{{
		Role:    "roles/browser",
		Members: []string{"user:a@example.com", "user:b@example.com"},
	}}}
	if !removeBindingFromPolicy(p, "roles/browser", "user:a@example.com", nil, false) {
		t.Fatal("expected change=true")
	}
	if !reflect.DeepEqual(p.Bindings[0].Members, []string{"user:b@example.com"}) {
		t.Errorf("members = %v", p.Bindings[0].Members)
	}
}

func TestRemoveBindingFromPolicyDropsEmptyBinding(t *testing.T) {
	p := &crm.Policy{Bindings: []*crm.Binding{{
		Role:    "roles/browser",
		Members: []string{"user:a@example.com"},
	}}}
	if !removeBindingFromPolicy(p, "roles/browser", "user:a@example.com", nil, false) {
		t.Fatal("expected change=true")
	}
	if len(p.Bindings) != 0 {
		t.Errorf("expected 0 bindings, got %+v", p.Bindings)
	}
}

func TestRemoveBindingFromPolicyNoMatch(t *testing.T) {
	p := &crm.Policy{Bindings: []*crm.Binding{{
		Role:    "roles/browser",
		Members: []string{"user:a@example.com"},
	}}}
	if removeBindingFromPolicy(p, "roles/editor", "user:a@example.com", nil, false) {
		t.Error("expected change=false when role does not match")
	}
	if removeBindingFromPolicy(p, "roles/browser", "user:z@example.com", nil, false) {
		t.Error("expected change=false when member is absent")
	}
}

func TestRemoveBindingFromPolicyAllConditions(t *testing.T) {
	cond := &crm.Expr{Expression: "e", Title: "t"}
	p := &crm.Policy{Bindings: []*crm.Binding{
		{Role: "roles/browser", Members: []string{"user:a@example.com"}},
		{Role: "roles/browser", Members: []string{"user:a@example.com"}, Condition: cond},
	}}
	if !removeBindingFromPolicy(p, "roles/browser", "user:a@example.com", nil, true) {
		t.Fatal("expected change=true")
	}
	if len(p.Bindings) != 0 {
		t.Errorf("expected all bindings removed, got %+v", p.Bindings)
	}
}

func TestParsePolicyFileJSON(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "policy.json")
	content := `{"bindings":[{"role":"roles/browser","members":["user:a@example.com"]}],"etag":"BwXYZ="}`
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := parsePolicyFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Bindings) != 1 || p.Bindings[0].Role != "roles/browser" {
		t.Errorf("unexpected bindings: %+v", p.Bindings)
	}
	if p.Etag != "BwXYZ=" {
		t.Errorf("etag = %q", p.Etag)
	}
}

func TestParsePolicyFileYAML(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "policy.yaml")
	content := "bindings:\n- role: roles/browser\n  members:\n  - user:a@example.com\netag: BwXYZ=\n"
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := parsePolicyFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Bindings) != 1 || p.Bindings[0].Role != "roles/browser" {
		t.Errorf("unexpected bindings: %+v", p.Bindings)
	}
}

func TestBuildConditionEmpty(t *testing.T) {
	flagOrgIamCondExpr, flagOrgIamCondTitle, flagOrgIamCondDesc = "", "", ""
	if buildCondition() != nil {
		t.Error("expected nil for empty flags")
	}
}

func TestBuildConditionPopulated(t *testing.T) {
	flagOrgIamCondExpr = "request.time > timestamp(\"2020-01-01T00:00:00Z\")"
	flagOrgIamCondTitle = "after-2020"
	flagOrgIamCondDesc = "Only after 2020"
	t.Cleanup(func() { flagOrgIamCondExpr, flagOrgIamCondTitle, flagOrgIamCondDesc = "", "", "" })

	c := buildCondition()
	if c == nil {
		t.Fatal("expected non-nil condition")
	}
	if c.Title != "after-2020" || c.Expression == "" || c.Description == "" {
		t.Errorf("unexpected condition: %+v", c)
	}
}
