package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func cmSubgroup(name string) *cobra.Command {
	for _, c := range certificateManagerCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestCMCertificatesSubcommands(t *testing.T) {
	g := cmSubgroup("certificates")
	if g == nil {
		t.Fatal("certificate-manager certificates missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestCMDnsAuthorizationsSubcommands(t *testing.T) {
	g := cmSubgroup("dns-authorizations")
	if g == nil {
		t.Fatal("certificate-manager dns-authorizations missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestCMIssuanceConfigsSubcommands(t *testing.T) {
	g := cmSubgroup("issuance-configs")
	if g == nil {
		t.Fatal("certificate-manager issuance-configs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestCMMapsSubcommands(t *testing.T) {
	g := cmSubgroup("maps")
	if g == nil {
		t.Fatal("certificate-manager maps missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "entries", "list", "update"})
	entries := findSub(g, "entries")
	if entries == nil {
		t.Fatal("maps entries missing")
	}
	assertSubcommands(t, entries, []string{"create", "delete", "describe", "list", "update"})
}

func TestCMOperationsSubcommands(t *testing.T) {
	g := cmSubgroup("operations")
	if g == nil {
		t.Fatal("certificate-manager operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestCMTrustConfigsSubcommands(t *testing.T) {
	g := cmSubgroup("trust-configs")
	if g == nil {
		t.Fatal("certificate-manager trust-configs missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}
