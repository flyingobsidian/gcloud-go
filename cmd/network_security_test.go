package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func nsSubgroup(name string) *cobra.Command {
	for _, c := range networkSecurityCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

var nsCRUDSubcommands = []string{"create", "delete", "describe", "list", "update"}

type nsGroupCase struct {
	name string
	subs []string
}

func TestNetworkSecuritySubgroups(t *testing.T) {
	cases := []nsGroupCase{
		{"address-groups", append([]string{"add-items", "clone-items", "list-references", "remove-items"}, nsCRUDSubcommands...)},
		{"authorization-policies", nsCRUDSubcommands},
		{"authz-policies", nsCRUDSubcommands},
		{"backend-authentication-configs", nsCRUDSubcommands},
		{"client-tls-policies", nsCRUDSubcommands},
		{"dns-threat-detectors", nsCRUDSubcommands},
		{"firewall-endpoint-associations", nsCRUDSubcommands},
		{"firewall-endpoints", nsCRUDSubcommands},
		{"gateway-security-policies", append([]string{"rules"}, nsCRUDSubcommands...)},
		{"intercept-deployment-groups", nsCRUDSubcommands},
		{"intercept-deployments", nsCRUDSubcommands},
		{"intercept-endpoint-group-associations", nsCRUDSubcommands},
		{"intercept-endpoint-groups", nsCRUDSubcommands},
		{"mirroring-deployment-groups", nsCRUDSubcommands},
		{"mirroring-deployments", nsCRUDSubcommands},
		{"mirroring-endpoint-group-associations", nsCRUDSubcommands},
		{"mirroring-endpoint-groups", nsCRUDSubcommands},
		{"operations", []string{"cancel", "delete", "describe", "list"}},
		{"org-address-groups", append([]string{"add-items", "clone-items", "list-references", "remove-items"}, nsCRUDSubcommands...)},
		{"secure-access-connect", []string{"attachments", "realms"}},
		{"security-profile-groups", nsCRUDSubcommands},
		{"security-profiles", nsCRUDSubcommands},
		{"server-tls-policies", nsCRUDSubcommands},
		{"tls-inspection-policies", nsCRUDSubcommands},
		{"url-lists", nsCRUDSubcommands},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			g := nsSubgroup(tc.name)
			if g == nil {
				t.Fatalf("network-security %s missing", tc.name)
			}
			assertSubcommands(t, g, tc.subs)
		})
	}
}

func TestNetworkSecurityGatewayRules(t *testing.T) {
	gsp := nsSubgroup("gateway-security-policies")
	if gsp == nil {
		t.Fatal("gateway-security-policies missing")
	}
	var rules *cobra.Command
	for _, c := range gsp.Commands() {
		if c.Name() == "rules" {
			rules = c
			break
		}
	}
	if rules == nil {
		t.Fatal("gateway-security-policies rules subgroup missing")
	}
	assertSubcommands(t, rules, nsCRUDSubcommands)
}

func TestNetworkSecuritySecureAccessConnectSubgroups(t *testing.T) {
	sac := nsSubgroup("secure-access-connect")
	if sac == nil {
		t.Fatal("secure-access-connect missing")
	}
	var attachments, realms *cobra.Command
	for _, c := range sac.Commands() {
		switch c.Name() {
		case "attachments":
			attachments = c
		case "realms":
			realms = c
		}
	}
	if attachments == nil {
		t.Fatal("secure-access-connect attachments subgroup missing")
	}
	if realms == nil {
		t.Fatal("secure-access-connect realms subgroup missing")
	}
	sacLeaves := []string{"create", "delete", "describe", "list"}
	assertSubcommands(t, attachments, sacLeaves)
	assertSubcommands(t, realms, sacLeaves)
}
