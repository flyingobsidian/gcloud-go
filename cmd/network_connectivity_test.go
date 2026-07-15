package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func ncSubgroup(name string) *cobra.Command {
	for _, c := range networkConnectivityCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

var ncCRUDSubcommands = []string{"create", "delete", "describe", "list", "update"}

type ncGroupCase struct {
	name string
	subs []string
}

func TestNetworkConnectivitySubgroups(t *testing.T) {
	cases := []ncGroupCase{
		{"hubs", append([]string{"accept-spoke", "list-spokes", "reject-spoke"}, ncCRUDSubcommands...)},
		{"internal-ranges", ncCRUDSubcommands},
		{"locations", []string{"describe", "list"}},
		{"multicloud-data-transfer-configs", ncCRUDSubcommands},
		{"multicloud-data-transfer-supported-services", []string{"describe", "list"}},
		{"operations", []string{"cancel", "delete", "describe", "list"}},
		{"policy-based-routes", []string{"create", "delete", "describe", "list"}},
		{"regional-endpoints", []string{"create", "delete", "describe", "list"}},
		{"service-connection-policies", ncCRUDSubcommands},
		{"spokes", append([]string{"accept", "reject"}, ncCRUDSubcommands...)},
		{"transports", ncCRUDSubcommands},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			g := ncSubgroup(tc.name)
			if g == nil {
				t.Fatalf("network-connectivity %s missing", tc.name)
			}
			assertSubcommands(t, g, tc.subs)
		})
	}
}
