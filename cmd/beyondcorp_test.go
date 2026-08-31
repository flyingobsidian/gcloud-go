package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	beyondcorp "google.golang.org/api/beyondcorp/v1"
)

func beyondcorpSubgroup(name string) *cobra.Command {
	for _, c := range beyondcorpCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestBeyondcorpOperationsSubcommands(t *testing.T) {
	g := beyondcorpSubgroup("operations")
	if g == nil {
		t.Fatal("beyondcorp operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestBeyondcorpSecurityGatewaysSubcommands(t *testing.T) {
	g := beyondcorpSubgroup("security-gateways")
	if g == nil {
		t.Fatal("beyondcorp security-gateways missing")
	}
	assertSubcommands(t, g, []string{
		"add-iam-policy-binding", "applications", "create", "delete", "describe",
		"get-iam-policy", "list", "remove-iam-policy-binding", "set-iam-policy", "update",
	})
}

func TestBeyondcorpSecurityGatewaysApplicationsSubcommands(t *testing.T) {
	sg := beyondcorpSubgroup("security-gateways")
	if sg == nil {
		t.Fatal("beyondcorp security-gateways missing")
	}
	apps := findSub(sg, "applications")
	if apps == nil {
		t.Fatal("beyondcorp security-gateways applications missing")
	}
	assertSubcommands(t, apps, []string{
		"add-iam-policy-binding", "create", "delete", "describe", "list",
		"remove-iam-policy-binding", "update",
	})
}

func TestBeyondcorpSecurityGatewaysAddIamBindingHasConditionFlags(t *testing.T) {
	sg := beyondcorpSubgroup("security-gateways")
	if sg == nil {
		t.Fatal("beyondcorp security-gateways missing")
	}
	for _, name := range []string{"add-iam-policy-binding", "remove-iam-policy-binding"} {
		cmd := findSub(sg, name)
		if cmd == nil {
			t.Fatalf("beyondcorp security-gateways %s missing", name)
		}
		for _, flag := range []string{"condition-expression", "condition-title", "condition-description"} {
			if cmd.Flag(flag) == nil {
				t.Errorf("beyondcorp security-gateways %s: missing --%s flag", name, flag)
			}
		}
	}
	if findSub(sg, "remove-iam-policy-binding").Flag("all") == nil {
		t.Error("beyondcorp security-gateways remove-iam-policy-binding: missing --all flag")
	}
}

func TestBeyondcorpApplicationsIamBindingCondFlags(t *testing.T) {
	sg := beyondcorpSubgroup("security-gateways")
	if sg == nil {
		t.Fatal("beyondcorp security-gateways missing")
	}
	apps := findSub(sg, "applications")
	if apps == nil {
		t.Fatal("beyondcorp security-gateways applications missing")
	}
	for _, name := range []string{"add-iam-policy-binding", "remove-iam-policy-binding"} {
		cmd := findSub(apps, name)
		if cmd == nil {
			t.Fatalf("beyondcorp security-gateways applications %s missing", name)
		}
		for _, flag := range []string{"member", "role", "condition-expression", "condition-title", "condition-description"} {
			if cmd.Flag(flag) == nil {
				t.Errorf("beyondcorp security-gateways applications %s: missing --%s flag", name, flag)
			}
		}
	}
	if findSub(apps, "remove-iam-policy-binding").Flag("all") == nil {
		t.Error("beyondcorp security-gateways applications remove-iam-policy-binding: missing --all flag")
	}
}

func TestBcIamAddRemoveBinding(t *testing.T) {
	cond := &beyondcorp.GoogleTypeExpr{Expression: "e", Title: "t"}
	pol := &beyondcorp.GoogleIamV1Policy{}
	// add first user without condition
	bcIamAddBinding(pol, "r1", "user:a@x", nil)
	// dedupes the same member
	bcIamAddBinding(pol, "r1", "user:a@x", nil)
	// add same role but with condition creates a separate binding
	bcIamAddBinding(pol, "r1", "user:b@x", cond)
	if len(pol.Bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(pol.Bindings))
	}
	if len(pol.Bindings[0].Members) != 1 || pol.Bindings[0].Members[0] != "user:a@x" {
		t.Errorf("unconditional binding has unexpected members: %v", pol.Bindings[0].Members)
	}
	if pol.Bindings[1].Condition == nil || pol.Bindings[1].Condition.Title != "t" {
		t.Errorf("conditional binding lost condition: %+v", pol.Bindings[1].Condition)
	}

	// remove missing member returns false
	if bcIamRemoveBinding(pol, "r1", "user:missing@x", nil, false) {
		t.Error("expected no change removing missing member")
	}
	// remove user:a from unconditional only
	if !bcIamRemoveBinding(pol, "r1", "user:a@x", nil, false) {
		t.Error("expected change removing user:a")
	}
	if len(pol.Bindings) != 1 || pol.Bindings[0].Condition == nil {
		t.Errorf("unconditional binding should be gone; got %+v", pol.Bindings)
	}
	// --all removes across conditions
	bcIamAddBinding(pol, "r1", "user:b@x", nil)
	if !bcIamRemoveBinding(pol, "r1", "user:b@x", nil, true) {
		t.Error("expected change removing user:b with --all")
	}
	if len(pol.Bindings) != 0 {
		t.Errorf("expected all r1 bindings removed; got %+v", pol.Bindings)
	}
}

func TestBcIamBuildAndCondsEqual(t *testing.T) {
	// no flags → nil
	flagBCIamCondExpr, flagBCIamCondTitle, flagBCIamCondDesc = "", "", ""
	if bcIamBuildCondition() != nil {
		t.Error("expected nil condition when no flags set")
	}
	// flags set → non-nil with correct values
	flagBCIamCondExpr = "request.time < timestamp('2026-01-01T00:00:00Z')"
	flagBCIamCondTitle = "expires_2026"
	flagBCIamCondDesc = "expires new year 2026"
	defer func() { flagBCIamCondExpr, flagBCIamCondTitle, flagBCIamCondDesc = "", "", "" }()
	c := bcIamBuildCondition()
	if c == nil || c.Expression != flagBCIamCondExpr || c.Title != flagBCIamCondTitle || c.Description != flagBCIamCondDesc {
		t.Errorf("bcIamBuildCondition returned %+v", c)
	}
	// conds equal
	if !bcIamCondsEqual(nil, nil) {
		t.Error("nil,nil should be equal")
	}
	if bcIamCondsEqual(c, nil) || bcIamCondsEqual(nil, c) {
		t.Error("nil and non-nil should not be equal")
	}
	if !bcIamCondsEqual(c, bcIamBuildCondition()) {
		t.Error("identical conditions should be equal")
	}
}
