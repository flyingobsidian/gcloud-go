package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	agentidentityalpha "google.golang.org/api/agentidentity/v1alpha"
)

func TestAgentIdentityAuthorizationsSubcommands(t *testing.T) {
	assertSubcommands(t, agentIdentityAuthorizationsCmd, []string{
		"add-iam-policy-binding", "get-iam-policy", "remove-iam-policy-binding",
		"set-iam-policy", "test-iam-permissions",
	})
	if agentIdentityAuthorizationsCmd.Parent() != agentIdentityAuthProvidersCmd {
		t.Fatal("authorizations should be registered under agent-identity auth-providers")
	}
}

func TestAgentIdentityAuthorizationsFlags(t *testing.T) {
	cases := map[string][]string{
		"get-iam-policy":            {"auth-provider", "endpoint-mode", "location", "limit", "page-size", "sort-by"},
		"set-iam-policy":            {"auth-provider", "endpoint-mode", "location"},
		"add-iam-policy-binding":    {"auth-provider", "endpoint-mode", "location", "member", "role"},
		"remove-iam-policy-binding": {"auth-provider", "endpoint-mode", "location", "member", "role"},
		"test-iam-permissions":      {"auth-provider", "endpoint-mode", "location", "permissions"},
	}
	for _, c := range agentIdentityAuthorizationsCmd.Commands() {
		want, ok := cases[c.Name()]
		if !ok {
			continue
		}
		for _, f := range want {
			if c.Flags().Lookup(f) == nil {
				t.Errorf("%s: missing --%s", c.Name(), f)
			}
		}
		// enable_condition: false in the reference spec.
		if c.Flags().Lookup("condition") != nil || c.Flags().Lookup("condition-expression") != nil {
			t.Errorf("%s: unexpected condition flag", c.Name())
		}
	}
}

func TestAIAuthorizationName(t *testing.T) {
	prevLoc, prevAP, prevProj := flagAIAuthzLocation, flagAIAuthzAuthProvider, flagProject
	defer func() { flagAIAuthzLocation, flagAIAuthzAuthProvider, flagProject = prevLoc, prevAP, prevProj }()

	full := "projects/p/locations/us-central1/authProviders/ap/authorizations/a"
	flagAIAuthzLocation, flagAIAuthzAuthProvider = "", ""
	if got, err := aiAuthorizationName(full); err != nil || got != full {
		t.Fatalf("full name: got %q, %v", got, err)
	}
	if _, err := aiAuthorizationName("projects/p/locations/l/authorizations/a"); err == nil {
		t.Fatal("expected error for malformed resource name")
	}
	if _, err := aiAuthorizationName("a"); err == nil {
		t.Fatal("expected error without --location")
	}
	flagAIAuthzLocation = "us-central1"
	if _, err := aiAuthorizationName("a"); err == nil {
		t.Fatal("expected error without --auth-provider")
	}
	flagAIAuthzAuthProvider, flagProject = "ap", "p"
	if got, err := aiAuthorizationName("a"); err != nil || got != full {
		t.Fatalf("short name: got %q, %v", got, err)
	}
}

func TestAIAuthzEndpoint(t *testing.T) {
	cases := map[string]string{
		"":                   "",
		"global":             "",
		"regional-preferred": "",
		"regional":           "https://agentidentity.us-central1.rep.googleapis.com/",
	}
	for mode, want := range cases {
		got, err := aiAuthzEndpoint(mode, "us-central1")
		if err != nil || got != want {
			t.Errorf("mode %q: got %q, %v; want %q", mode, got, err, want)
		}
	}
	if _, err := aiAuthzEndpoint("bogus", "us-central1"); err == nil {
		t.Fatal("expected error for invalid endpoint mode")
	}
}

func TestAIAuthzAddBinding(t *testing.T) {
	pol := &agentidentityalpha.Policy{Bindings: []*agentidentityalpha.Binding{
		{Role: "roles/a", Members: []string{"user:x"}},
	}}
	aiAuthzAddBinding(pol, "roles/a", "user:x")
	aiAuthzAddBinding(pol, "roles/a", "user:y")
	aiAuthzAddBinding(pol, "roles/b", "user:z")
	if len(pol.Bindings) != 2 {
		t.Fatalf("want 2 bindings, got %d", len(pol.Bindings))
	}
	if !reflect.DeepEqual(pol.Bindings[0].Members, []string{"user:x", "user:y"}) {
		t.Errorf("roles/a members: %v", pol.Bindings[0].Members)
	}
	if pol.Bindings[1].Role != "roles/b" || !reflect.DeepEqual(pol.Bindings[1].Members, []string{"user:z"}) {
		t.Errorf("roles/b binding: %+v", pol.Bindings[1])
	}
}

func TestAIAuthzRemoveBinding(t *testing.T) {
	pol := &agentidentityalpha.Policy{Bindings: []*agentidentityalpha.Binding{
		{Role: "roles/a", Members: []string{"user:x", "user:y"}},
		{Role: "roles/b", Members: []string{"user:x"}},
	}}
	if err := aiAuthzRemoveBinding(pol, "roles/b", "user:x"); err != nil {
		t.Fatal(err)
	}
	if len(pol.Bindings) != 1 || pol.Bindings[0].Role != "roles/a" {
		t.Fatalf("empty binding not dropped: %+v", pol.Bindings)
	}
	if err := aiAuthzRemoveBinding(pol, "roles/a", "user:x"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pol.Bindings[0].Members, []string{"user:y"}) {
		t.Errorf("roles/a members: %v", pol.Bindings[0].Members)
	}
	err := aiAuthzRemoveBinding(pol, "roles/a", "user:x")
	if err == nil || err.Error() != "Policy binding with the specified principal and role not found!" {
		t.Fatalf("want not-found error, got %v", err)
	}
}

func TestAIAuthzPolicyUpdateMask(t *testing.T) {
	p := filepath.Join(t.TempDir(), "policy.yaml")
	if err := os.WriteFile(p, []byte("version: 1\nbindings:\n- role: roles/a\n  members: [user:x]\netag: BwX=\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := aiAuthzPolicyUpdateMask(p)
	if err != nil {
		t.Fatal(err)
	}
	if got != "bindings,etag,version" {
		t.Fatalf("got %q", got)
	}
}
