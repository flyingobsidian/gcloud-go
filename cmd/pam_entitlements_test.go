package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func pamSubgroup(name string) *cobra.Command {
	for _, c := range pamCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestPamHasEntitlementsSubgroup(t *testing.T) {
	if pamSubgroup("entitlements") == nil {
		t.Fatal("pam missing entitlements subgroup")
	}
}

func TestPamEntitlementsSubcommands(t *testing.T) {
	g := pamSubgroup("entitlements")
	if g == nil {
		t.Fatal("entitlements missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "export", "list", "search", "update"})
}

func TestPamCallerAccessValue(t *testing.T) {
	cases := map[string]string{
		"grant-requester": "GRANT_REQUESTER",
		"GRANT-REQUESTER": "GRANT_REQUESTER",
		"grant-approver":  "GRANT_APPROVER",
	}
	for in, want := range cases {
		got, err := pamCallerAccessValue(in)
		if err != nil {
			t.Errorf("pamCallerAccessValue(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("pamCallerAccessValue(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := pamCallerAccessValue("bogus"); err == nil {
		t.Errorf("pamCallerAccessValue(\"bogus\") expected an error")
	}
}

func TestPamEntNamePassesThroughFullyQualified(t *testing.T) {
	full := "projects/x/locations/global/entitlements/y"
	got, err := pamEntName(full)
	if err != nil {
		t.Fatalf("pamEntName: %v", err)
	}
	if got != full {
		t.Errorf("pamEntName should pass through fully-qualified names, got %q", got)
	}
}
