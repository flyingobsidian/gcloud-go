package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func privatecaSubgroup(name string) *cobra.Command {
	for _, c := range privatecaCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestPrivatecaCertificatesSubcommands(t *testing.T) {
	g := privatecaSubgroup("certificates")
	if g == nil {
		t.Fatal("privateca certificates missing")
	}
	assertSubcommands(t, g, []string{"create", "describe", "export", "list", "revoke", "update"})
}

func TestPrivatecaLocationsSubcommands(t *testing.T) {
	g := privatecaSubgroup("locations")
	if g == nil {
		t.Fatal("privateca locations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestPrivatecaOperationsSubcommands(t *testing.T) {
	g := privatecaSubgroup("operations")
	if g == nil {
		t.Fatal("privateca operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}

func TestPrivatecaPoolsSubcommands(t *testing.T) {
	g := privatecaSubgroup("pools")
	if g == nil {
		t.Fatal("privateca pools missing")
	}
	assertSubcommands(t, g, []string{"add-iam-policy-binding", "create", "delete", "describe", "get-ca-certs", "get-iam-policy", "list", "remove-iam-policy-binding", "set-iam-policy", "update"})
}

func TestPrivatecaRootsSubcommands(t *testing.T) {
	g := privatecaSubgroup("roots")
	if g == nil {
		t.Fatal("privateca roots missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "disable", "enable", "list", "undelete", "update"})
}

func TestPrivatecaSubordinatesSubcommands(t *testing.T) {
	g := privatecaSubgroup("subordinates")
	if g == nil {
		t.Fatal("privateca subordinates missing")
	}
	assertSubcommands(t, g, []string{"activate", "create", "delete", "describe", "disable", "enable", "get-csr", "list", "undelete", "update"})
}

func TestPrivatecaTemplatesSubcommands(t *testing.T) {
	g := privatecaSubgroup("templates")
	if g == nil {
		t.Fatal("privateca templates missing")
	}
	assertSubcommands(t, g, []string{"add-iam-policy-binding", "create", "delete", "describe", "get-iam-policy", "list", "remove-iam-policy-binding", "replicate", "set-iam-policy", "update"})
}
