package cmd

import "testing"

func TestContainerHasGetServerConfig(t *testing.T) {
	found := false
	for _, c := range containerCmd.Commands() {
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
