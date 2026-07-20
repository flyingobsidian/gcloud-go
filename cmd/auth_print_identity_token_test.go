package cmd

import "testing"

func TestAuthHasPrintIdentityToken(t *testing.T) {
	found := false
	for _, c := range authCmd.Commands() {
		if c.Name() == "print-identity-token" {
			found = true
			if c.RunE == nil {
				t.Fatal("auth print-identity-token should have a RunE (not just a stub)")
			}
		}
	}
	if !found {
		t.Fatal("auth print-identity-token missing")
	}
}
