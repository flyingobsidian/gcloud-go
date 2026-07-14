package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func policyIntSubgroup(name string) *cobra.Command {
	for _, c := range policyIntelligenceCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestPolicyIntSimulateSubcommands(t *testing.T) {
	g := policyIntSubgroup("simulate")
	if g == nil {
		t.Fatal("policy-intelligence simulate missing")
	}
	assertSubcommands(t, g, []string{"orgpolicy"})
	op := findSub(g, "orgpolicy")
	if op == nil {
		t.Fatal("simulate orgpolicy missing")
	}
	assertSubcommands(t, op, []string{"create", "describe", "list"})
}

func TestPolicyIntTroubleshootSubcommands(t *testing.T) {
	g := policyIntSubgroup("troubleshoot-policy")
	if g == nil {
		t.Fatal("policy-intelligence troubleshoot-policy missing")
	}
	assertSubcommands(t, g, []string{"iam"})
}

func TestPolicyIntQueryActivityRegistered(t *testing.T) {
	if policyIntSubgroup("query-activity") == nil {
		t.Fatal("policy-intelligence query-activity missing")
	}
}
