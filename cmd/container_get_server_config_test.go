package cmd

import "testing"

func TestContainerHasGetServerConfig(t *testing.T) {
	grp := newContainerCmds(legacyTestInvocation())
	found := false
	for _, c := range grp.containerCmd.Commands() {
		if c.Name() == "get-server-config" {
			found = true
			if c.RunE == nil {
				t.Fatal("container get-server-config should have a RunE (not just a stub)")
			}
		}
	}
	if !found {
		t.Fatal("container get-server-config missing")
	}
}
