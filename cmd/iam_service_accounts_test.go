package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	iam "google.golang.org/api/iam/v1"
)

func TestBuildCreateServiceAccountRequest(t *testing.T) {
	t.Run("display name and description", func(t *testing.T) {
		flagSADisplayName = "My SA"
		flagSADescription = "does things"
		t.Cleanup(func() { flagSADisplayName = ""; flagSADescription = "" })

		req := buildCreateServiceAccountRequest("my-sa")
		if req.AccountId != "my-sa" {
			t.Errorf("AccountId = %q, want my-sa", req.AccountId)
		}
		if req.ServiceAccount == nil {
			t.Fatal("ServiceAccount is nil, want populated")
		}
		if req.ServiceAccount.DisplayName != "My SA" {
			t.Errorf("DisplayName = %q, want My SA", req.ServiceAccount.DisplayName)
		}
		if req.ServiceAccount.Description != "does things" {
			t.Errorf("Description = %q, want does things", req.ServiceAccount.Description)
		}
	})

	t.Run("display name only", func(t *testing.T) {
		flagSADisplayName = "Only Name"
		flagSADescription = ""
		t.Cleanup(func() { flagSADisplayName = "" })

		req := buildCreateServiceAccountRequest("sa2")
		if req.ServiceAccount == nil {
			t.Fatal("ServiceAccount is nil, want populated")
		}
		if req.ServiceAccount.DisplayName != "Only Name" {
			t.Errorf("DisplayName = %q, want Only Name", req.ServiceAccount.DisplayName)
		}
		if req.ServiceAccount.Description != "" {
			t.Errorf("Description = %q, want empty", req.ServiceAccount.Description)
		}
	})

	t.Run("no optional fields omits ServiceAccount body", func(t *testing.T) {
		flagSADisplayName = ""
		flagSADescription = ""

		req := buildCreateServiceAccountRequest("bare-sa")
		if req.AccountId != "bare-sa" {
			t.Errorf("AccountId = %q, want bare-sa", req.AccountId)
		}
		if req.ServiceAccount != nil {
			t.Errorf("ServiceAccount = %+v, want nil", req.ServiceAccount)
		}
	})
}

func TestSAResourceName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"sa@my-project.iam.gserviceaccount.com", "projects/-/serviceAccounts/sa@my-project.iam.gserviceaccount.com"},
		{"103271949540120710052", "projects/-/serviceAccounts/103271949540120710052"},
		{"projects/-/serviceAccounts/sa@my-project.iam.gserviceaccount.com", "projects/-/serviceAccounts/sa@my-project.iam.gserviceaccount.com"},
	}
	for _, c := range cases {
		if got := saResourceName(c.in); got != c.want {
			t.Errorf("saResourceName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsDigitString(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"103271949540120710052", true},
		{"123abc", false},
		{"sa@example.com", false},
	}
	for _, c := range cases {
		if got := isDigitString(c.in); got != c.want {
			t.Errorf("isDigitString(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIAMConditionsEqual(t *testing.T) {
	a := &iam.Expr{Expression: "x", Title: "t"}
	b := &iam.Expr{Expression: "x", Title: "t"}
	c := &iam.Expr{Expression: "y", Title: "t"}
	if !iamConditionsEqual(nil, nil) {
		t.Error("nil,nil should be equal")
	}
	if iamConditionsEqual(a, nil) || iamConditionsEqual(nil, a) {
		t.Error("nil vs non-nil should be unequal")
	}
	if !iamConditionsEqual(a, b) {
		t.Error("identical conditions should be equal")
	}
	if iamConditionsEqual(a, c) {
		t.Error("differing expressions should be unequal")
	}
}

func TestIAMAddBindingToPolicy(t *testing.T) {
	t.Run("new binding", func(t *testing.T) {
		p := &iam.Policy{}
		if !iamAddBindingToPolicy(p, "roles/iam.serviceAccountUser", "user:a@example.com", nil) {
			t.Fatal("expected change=true when adding to empty policy")
		}
		if len(p.Bindings) != 1 || p.Bindings[0].Role != "roles/iam.serviceAccountUser" {
			t.Fatalf("unexpected bindings: %+v", p.Bindings)
		}
		if !reflect.DeepEqual(p.Bindings[0].Members, []string{"user:a@example.com"}) {
			t.Errorf("members = %v", p.Bindings[0].Members)
		}
	})

	t.Run("existing role adds member", func(t *testing.T) {
		p := &iam.Policy{Bindings: []*iam.Binding{{
			Role:    "roles/iam.serviceAccountUser",
			Members: []string{"user:a@example.com"},
		}}}
		if !iamAddBindingToPolicy(p, "roles/iam.serviceAccountUser", "user:b@example.com", nil) {
			t.Fatal("expected change=true when adding new member")
		}
		want := []string{"user:a@example.com", "user:b@example.com"}
		if !reflect.DeepEqual(p.Bindings[0].Members, want) {
			t.Errorf("members = %v, want %v", p.Bindings[0].Members, want)
		}
	})

	t.Run("duplicate member", func(t *testing.T) {
		p := &iam.Policy{Bindings: []*iam.Binding{{
			Role:    "roles/iam.serviceAccountUser",
			Members: []string{"user:a@example.com"},
		}}}
		if iamAddBindingToPolicy(p, "roles/iam.serviceAccountUser", "user:a@example.com", nil) {
			t.Fatal("expected change=false for duplicate member")
		}
	})
}

func TestIAMRemoveBindingFromPolicy(t *testing.T) {
	t.Run("removes single member", func(t *testing.T) {
		p := &iam.Policy{Bindings: []*iam.Binding{{
			Role:    "roles/iam.serviceAccountUser",
			Members: []string{"user:a@example.com", "user:b@example.com"},
		}}}
		if !iamRemoveBindingFromPolicy(p, "roles/iam.serviceAccountUser", "user:a@example.com", nil, false) {
			t.Fatal("expected change=true")
		}
		if !reflect.DeepEqual(p.Bindings[0].Members, []string{"user:b@example.com"}) {
			t.Errorf("members = %v", p.Bindings[0].Members)
		}
	})

	t.Run("drops empty binding", func(t *testing.T) {
		p := &iam.Policy{Bindings: []*iam.Binding{{
			Role:    "roles/iam.serviceAccountUser",
			Members: []string{"user:a@example.com"},
		}}}
		if !iamRemoveBindingFromPolicy(p, "roles/iam.serviceAccountUser", "user:a@example.com", nil, false) {
			t.Fatal("expected change=true")
		}
		if len(p.Bindings) != 0 {
			t.Errorf("expected 0 bindings, got %+v", p.Bindings)
		}
	})

	t.Run("no match", func(t *testing.T) {
		p := &iam.Policy{Bindings: []*iam.Binding{{
			Role:    "roles/iam.serviceAccountUser",
			Members: []string{"user:a@example.com"},
		}}}
		if iamRemoveBindingFromPolicy(p, "roles/editor", "user:a@example.com", nil, false) {
			t.Error("expected change=false when role does not match")
		}
	})

	t.Run("all conditions", func(t *testing.T) {
		cond := &iam.Expr{Expression: "e", Title: "t"}
		p := &iam.Policy{Bindings: []*iam.Binding{
			{Role: "roles/iam.serviceAccountUser", Members: []string{"user:a@example.com"}},
			{Role: "roles/iam.serviceAccountUser", Members: []string{"user:a@example.com"}, Condition: cond},
		}}
		if !iamRemoveBindingFromPolicy(p, "roles/iam.serviceAccountUser", "user:a@example.com", nil, true) {
			t.Fatal("expected change=true")
		}
		if len(p.Bindings) != 0 {
			t.Errorf("expected all bindings removed, got %+v", p.Bindings)
		}
	})
}

func TestParseIAMPolicyFile(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "policy.json")
		content := `{"bindings":[{"role":"roles/iam.serviceAccountUser","members":["user:a@example.com"]}],"etag":"BwXYZ="}`
		if err := os.WriteFile(f, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		p, err := parseIAMPolicyFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Bindings) != 1 || p.Bindings[0].Role != "roles/iam.serviceAccountUser" {
			t.Errorf("unexpected bindings: %+v", p.Bindings)
		}
		if p.Etag != "BwXYZ=" {
			t.Errorf("etag = %q", p.Etag)
		}
	})

	t.Run("YAML", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "policy.yaml")
		content := "bindings:\n- role: roles/iam.serviceAccountUser\n  members:\n  - user:a@example.com\netag: BwXYZ=\n"
		if err := os.WriteFile(f, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		p, err := parseIAMPolicyFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Bindings) != 1 || p.Bindings[0].Role != "roles/iam.serviceAccountUser" {
			t.Errorf("unexpected bindings: %+v", p.Bindings)
		}
	})
}

func TestServiceAccountsCommandsRegistered(t *testing.T) {
	got := map[string]bool{}
	for _, c := range serviceAccountsCmd.Commands() {
		got[c.Name()] = true
	}
	want := []string{
		"add-iam-policy-binding",
		"create",
		"delete",
		"describe",
		"disable",
		"enable",
		"get-iam-policy",
		"list",
		"remove-iam-policy-binding",
		"set-iam-policy",
		"sign-blob",
		"sign-jwt",
		"undelete",
		"update",
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("service-accounts subcommand %q not registered", name)
		}
	}
}
