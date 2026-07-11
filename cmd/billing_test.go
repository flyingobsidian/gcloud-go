package cmd

import (
	"testing"

	billingbudgets "google.golang.org/api/billingbudgets/v1"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
)

func TestBillingAccountResourceName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"01ABCD-234567-8901EF", "billingAccounts/01ABCD-234567-8901EF"},
		{"billingAccounts/01ABCD-234567-8901EF", "billingAccounts/01ABCD-234567-8901EF"},
	}
	for _, c := range cases {
		if got := billingAccountResourceName(c.in); got != c.want {
			t.Errorf("billingAccountResourceName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestProjectBillingInfoName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"my-project", "projects/my-project/billingInfo"},
		{"projects/my-project", "projects/my-project/billingInfo"},
		{"projects/my-project/billingInfo", "projects/my-project/billingInfo"},
	}
	for _, c := range cases {
		if got := projectBillingInfoName(c.in); got != c.want {
			t.Errorf("projectBillingInfoName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBudgetResourceName(t *testing.T) {
	t.Run("qualified passes through", func(t *testing.T) {
		got, err := budgetResourceName("billingAccounts/123/budgets/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "billingAccounts/123/budgets/abc" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("bare id needs account", func(t *testing.T) {
		if _, err := budgetResourceName("abc", ""); err == nil {
			t.Fatal("expected error when billing-account is empty")
		}
	})
	t.Run("bare id with account", func(t *testing.T) {
		got, err := budgetResourceName("abc", "123")
		if err != nil {
			t.Fatal(err)
		}
		if got != "billingAccounts/123/budgets/abc" {
			t.Errorf("got %q", got)
		}
	})
}

func TestParseMoney(t *testing.T) {
	cases := []struct {
		in       string
		wantUnit int64
		wantNano int64
		wantCode string
		wantErr  bool
	}{
		{"100", 100, 0, "", false},
		{"100USD", 100, 0, "USD", false},
		{"100.75USD", 100, 75, "USD", false},
		{"987.65", 987, 65, "", false},
		{"abc", 0, 0, "", true},
		{"", 0, 0, "", true},
	}
	for _, c := range cases {
		got, err := parseMoney(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseMoney(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseMoney(%q) err = %v", c.in, err)
			continue
		}
		if got.Units != c.wantUnit || got.Nanos != c.wantNano || got.CurrencyCode != c.wantCode {
			t.Errorf("parseMoney(%q) = %+v, want units=%d nanos=%d code=%q",
				c.in, got, c.wantUnit, c.wantNano, c.wantCode)
		}
	}
}

func TestParseThresholdRule(t *testing.T) {
	t.Run("percent only", func(t *testing.T) {
		r, err := parseThresholdRule("percent=0.5")
		if err != nil {
			t.Fatal(err)
		}
		if r.ThresholdPercent != 0.5 || r.SpendBasis != "" {
			t.Errorf("got %+v", r)
		}
	})
	t.Run("with forecasted basis", func(t *testing.T) {
		r, err := parseThresholdRule("percent=0.75,basis=forecasted-spend")
		if err != nil {
			t.Fatal(err)
		}
		if r.ThresholdPercent != 0.75 || r.SpendBasis != "FORECASTED_SPEND" {
			t.Errorf("got %+v", r)
		}
	})
	t.Run("bad key", func(t *testing.T) {
		if _, err := parseThresholdRule("wat=1"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("bad basis", func(t *testing.T) {
		if _, err := parseThresholdRule("percent=0.5,basis=nope"); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBillingConditionsEqual(t *testing.T) {
	a := &cloudbilling.Expr{Expression: "x", Title: "t"}
	b := &cloudbilling.Expr{Expression: "x", Title: "t"}
	c := &cloudbilling.Expr{Expression: "y", Title: "t"}
	if !billingConditionsEqual(nil, nil) {
		t.Error("nil,nil should be equal")
	}
	if billingConditionsEqual(a, nil) || billingConditionsEqual(nil, a) {
		t.Error("nil vs non-nil should be unequal")
	}
	if !billingConditionsEqual(a, b) {
		t.Error("identical conditions should be equal")
	}
	if billingConditionsEqual(a, c) {
		t.Error("differing expressions should be unequal")
	}
}

func TestBillingAddBindingToPolicy(t *testing.T) {
	p := &cloudbilling.Policy{}
	if !billingAddBindingToPolicy(p, "roles/billing.viewer", "user:a@example.com", nil) {
		t.Fatal("expected change=true when adding to empty policy")
	}
	if billingAddBindingToPolicy(p, "roles/billing.viewer", "user:a@example.com", nil) {
		t.Fatal("expected change=false when adding duplicate member")
	}
	if !billingAddBindingToPolicy(p, "roles/billing.viewer", "user:b@example.com", nil) {
		t.Fatal("expected change=true when adding new member to existing binding")
	}
	if len(p.Bindings) != 1 || len(p.Bindings[0].Members) != 2 {
		t.Errorf("unexpected bindings: %+v", p.Bindings)
	}
}

func TestBillingRemoveBindingFromPolicy(t *testing.T) {
	t.Run("removes single member", func(t *testing.T) {
		p := &cloudbilling.Policy{Bindings: []*cloudbilling.Binding{{
			Role:    "roles/billing.viewer",
			Members: []string{"user:a@example.com", "user:b@example.com"},
		}}}
		if !billingRemoveBindingFromPolicy(p, "roles/billing.viewer", "user:a@example.com", nil, false) {
			t.Fatal("expected change=true")
		}
		if len(p.Bindings) != 1 || len(p.Bindings[0].Members) != 1 || p.Bindings[0].Members[0] != "user:b@example.com" {
			t.Errorf("unexpected bindings: %+v", p.Bindings)
		}
	})
	t.Run("no match", func(t *testing.T) {
		p := &cloudbilling.Policy{Bindings: []*cloudbilling.Binding{{
			Role:    "roles/billing.viewer",
			Members: []string{"user:a@example.com"},
		}}}
		if billingRemoveBindingFromPolicy(p, "roles/editor", "user:a@example.com", nil, false) {
			t.Error("expected change=false when role does not match")
		}
	})
}

func TestBillingCommandsRegistered(t *testing.T) {
	groups := map[string][]string{
		"accounts": {
			"add-iam-policy-binding", "describe", "get-iam-policy",
			"list", "remove-iam-policy-binding", "set-iam-policy",
		},
		"budgets":  {"create", "delete", "describe", "list", "update"},
		"projects": {"describe", "link", "list", "unlink"},
	}
	for _, sub := range billingCmd.Commands() {
		got := map[string]bool{}
		for _, c := range sub.Commands() {
			got[c.Name()] = true
		}
		want, ok := groups[sub.Name()]
		if !ok {
			t.Errorf("unexpected billing subgroup %q", sub.Name())
			continue
		}
		for _, name := range want {
			if !got[name] {
				t.Errorf("billing %s subcommand %q not registered", sub.Name(), name)
			}
		}
		delete(groups, sub.Name())
	}
	for name := range groups {
		t.Errorf("missing billing subgroup: %s", name)
	}
}

// The build depends on billingbudgets being referenced elsewhere; ensure the
// symbol is visible from the test package too.
var _ = (*billingbudgets.GoogleCloudBillingBudgetsV1Budget)(nil)
