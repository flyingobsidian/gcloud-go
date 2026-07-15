package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func recaptchaSubgroup(name string) *cobra.Command {
	for _, c := range recaptchaCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestRecaptchaHasFirewallPoliciesSubgroup(t *testing.T) {
	if recaptchaSubgroup("firewall-policies") == nil {
		t.Fatal("recaptcha missing firewall-policies subgroup")
	}
}

func TestRecaptchaFirewallPoliciesSubcommands(t *testing.T) {
	g := recaptchaSubgroup("firewall-policies")
	if g == nil {
		t.Fatal("firewall-policies missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "reorder", "update"})
}

func TestRecaptchaHasKeysSubgroup(t *testing.T) {
	if recaptchaSubgroup("keys") == nil {
		t.Fatal("recaptcha missing keys subgroup")
	}
}

func TestRecaptchaKeysSubcommands(t *testing.T) {
	g := recaptchaSubgroup("keys")
	if g == nil {
		t.Fatal("keys missing")
	}
	assertSubcommands(t, g, []string{
		"add-ip-override", "create", "delete", "describe", "list",
		"list-ip-overrides", "migrate", "remove-ip-override", "update",
	})
}

func TestParseRcpFirewallActions(t *testing.T) {
	got, err := parseRcpFirewallActions("allow,block,redirect,substitute=/foo,set_header=X-Custom=hi")
	if err != nil {
		t.Fatalf("parseRcpFirewallActions error: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5 actions, got %d", len(got))
	}
	if got[0].Allow == nil {
		t.Errorf("action 0: expected Allow")
	}
	if got[1].Block == nil {
		t.Errorf("action 1: expected Block")
	}
	if got[2].Redirect == nil {
		t.Errorf("action 2: expected Redirect")
	}
	if got[3].Substitute == nil || got[3].Substitute.Path != "/foo" {
		t.Errorf("action 3: expected Substitute(/foo)")
	}
	if got[4].SetHeader == nil || got[4].SetHeader.Key != "X-Custom" || got[4].SetHeader.Value != "hi" {
		t.Errorf("action 4: expected SetHeader(X-Custom=hi)")
	}
	for _, bad := range []string{"allow=x", "substitute", "set_header=k", "bogus"} {
		if _, err := parseRcpFirewallActions(bad); err == nil {
			t.Errorf("parseRcpFirewallActions(%q) expected error", bad)
		}
	}
}

func TestRcpKeyIntegrationEnum(t *testing.T) {
	cases := map[string]string{
		"":          "",
		"score":     "SCORE",
		"checkbox":  "CHECKBOX",
		"INVISIBLE": "INVISIBLE",
	}
	for in, want := range cases {
		got, err := rcpKeyIntegrationEnum(in)
		if err != nil {
			t.Errorf("rcpKeyIntegrationEnum(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("rcpKeyIntegrationEnum(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := rcpKeyIntegrationEnum("bogus"); err == nil {
		t.Errorf("rcpKeyIntegrationEnum(\"bogus\") expected error")
	}
}
