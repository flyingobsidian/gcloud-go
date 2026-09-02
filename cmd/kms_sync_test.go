package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func kmsSub(names ...string) *cobra.Command {
	var cur *cobra.Command = kmsCmd
	for _, n := range names {
		cur = findSub(cur, n)
		if cur == nil {
			return nil
		}
	}
	return cur
}

func TestKmsVersionsUpdateHasNewFlags(t *testing.T) {
	c := kmsSub("keys", "versions", "update")
	if c == nil {
		t.Fatal("kms keys versions update missing")
	}
	for _, name := range []string{"protection-level", "crypto-key-backend", "state", "update-mask", "config-file"} {
		if c.Flags().Lookup(name) == nil {
			t.Errorf("--%s flag missing on versions update", name)
		}
	}
}

func TestKmsKeysCreateHasHsmTrustedWrapping(t *testing.T) {
	c := kmsSub("keys", "create")
	if c == nil {
		t.Fatal("kms keys create missing")
	}
	for _, name := range []string{"hsm-trusted-wrapping", "crypto-key-backend"} {
		if c.Flags().Lookup(name) == nil {
			t.Errorf("--%s flag missing on keys create", name)
		}
	}
}

func TestKmsAutokeyCommandsHaveProjectAndFolder(t *testing.T) {
	for _, sub := range []string{"describe", "show-effective-config", "update"} {
		c := kmsSub("autokey-config", sub)
		if c == nil {
			t.Fatalf("kms autokey-config %s missing", sub)
		}
		for _, name := range []string{"project", "folder"} {
			if c.Flags().Lookup(name) == nil {
				t.Errorf("autokey-config %s missing --%s flag", sub, name)
			}
		}
		// --folder must be optional (not required) after 577.0.0.
		fold := c.Flags().Lookup("folder")
		if req, ok := fold.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(req) > 0 && req[0] == "true" {
			t.Errorf("autokey-config %s: --folder should be optional", sub)
		}
	}
}

func TestKmsHsmProposalCreateHasUpgradeKeyTrustFlags(t *testing.T) {
	c := kmsSub("single-tenant-hsm", "proposal", "create")
	if c == nil {
		t.Fatal("kms single-tenant-hsm proposal create missing")
	}
	for _, name := range []string{"operation-type", "crypto-key-version-name", "two-factor-public-key-pem"} {
		if c.Flags().Lookup(name) == nil {
			t.Errorf("proposal create missing --%s flag", name)
		}
	}
}

func TestKmsInventoryFlagsPresent(t *testing.T) {
	// 569.0.0: --fallback-scope on get-protected-resources-summary already present.
	c := kmsSub("inventory", "get-protected-resources-summary")
	if c == nil {
		t.Fatal("kms inventory get-protected-resources-summary missing")
	}
	if c.Flags().Lookup("fallback-scope") == nil {
		t.Error("get-protected-resources-summary missing --fallback-scope")
	}
	// 569.0.0: search-protected-resources supports projects/ scope (verified by
	// runtime dispatch); at minimum the --scope flag is present.
	c = kmsSub("inventory", "search-protected-resources")
	if c == nil || c.Flags().Lookup("scope") == nil {
		t.Error("search-protected-resources missing --scope flag")
	}
}
