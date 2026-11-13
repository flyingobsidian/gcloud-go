package cli

import "testing"

func TestAuthHasPrintIdentityToken(t *testing.T) {
	grp := newAuthCmds(legacyTestInvocation())
	found := false
	for _, c := range grp.authCmd.Commands() {
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
