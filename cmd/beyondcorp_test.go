package cmd

import (
	"testing"

	"github.com/spf13/cobra"
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
	assertSubcommands(t, apps, []string{"create", "delete", "describe", "list", "update"})
}
