package cmd

import "testing"

func TestPolicyTroubleshootHasIam(t *testing.T) {
	found := false
	for _, c := range policyTroubleshootCmd.Commands() {
		if c.Name() == "iam" {
			found = true
			if !c.HasFlags() {
				t.Fatal("policy-troubleshoot iam should have flags")
			}
			if c.RunE == nil {
				t.Fatal("policy-troubleshoot iam should have a RunE (not just a stub)")
			}
		}
	}
	if !found {
		t.Fatal("policy-troubleshoot iam missing")
	}
}
