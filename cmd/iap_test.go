package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// findRootSub returns a top-level cobra subcommand by name.
func findRootSub(name string) *cobra.Command {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestIapWebIamStubsIncludeAgentRegistryFlags(t *testing.T) {
	iap := findRootSub("iap")
	if iap == nil {
		t.Fatal("iap missing")
	}
	web := findSub(iap, "web")
	if web == nil {
		t.Fatal("iap web missing")
	}
	iamCmds := []string{"get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding"}
	for _, name := range iamCmds {
		sub := findSub(web, name)
		if sub == nil {
			t.Fatalf("iap web %s missing", name)
		}
		// agent-registry selectors promoted to GA in 576.0.0.
		for _, flag := range []string{"resource-type", "agent", "mcp-server", "endpoint"} {
			if sub.Flags().Lookup(flag) == nil {
				t.Errorf("iap web %s missing --%s flag", name, flag)
			}
		}
		usage := sub.Flag("resource-type").Usage
		if !strings.Contains(usage, "agent-registry") {
			t.Errorf("iap web %s --resource-type usage missing agent-registry: %q", name, usage)
		}
	}
}
