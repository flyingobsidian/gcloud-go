package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func privatecaSubgroup(name string) *cobra.Command {
	grp := newPrivatecaCmds(legacyTestInvocation())
	for _, c := range grp.privatecaCmd.Commands() {
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

func TestPrivatecaSubordinateActivateFirstPartyFlags(t *testing.T) {
	g := privatecaSubgroup("subordinates")
	if g == nil {
		t.Fatal("privateca subordinates missing")
	}
	activate := findSub(g, "activate")
	if activate == nil {
		t.Fatal("subordinates activate missing")
	}
	for _, name := range []string{"issuer-ca", "issuer-pool", "issuer-location"} {
		if activate.Flags().Lookup(name) == nil {
			t.Errorf("subordinates activate missing --%s", name)
		}
	}
}

func TestPCAResolveIssuerCA(t *testing.T) {
	grp := newPrivatecaCmds(legacyTestInvocation())
	saved := struct {
		ca, pool, loc, sPool, sLoc string
	}{
		ca:    grp.flagPCACAIssuerCA,
		pool:  grp.flagPCACAIssuerPool,
		loc:   grp.flagPCACAIssuerLocation,
		sPool: grp.flagPCACAPool,
		sLoc:  grp.flagPCACALocation,
	}
	t.Cleanup(func() {
		grp.flagPCACAIssuerCA = saved.ca
		grp.flagPCACAIssuerPool = saved.pool
		grp.flagPCACAIssuerLocation = saved.loc
		grp.flagPCACAPool = saved.sPool
		grp.flagPCACALocation = saved.sLoc
	})

	// Full name → passthrough.
	grp.flagPCACAIssuerCA = "projects/p/locations/us/caPools/pool/certificateAuthorities/ca"
	got, err := grp.pcaResolveIssuerCA("ignored")
	if err != nil || got != grp.flagPCACAIssuerCA {
		t.Fatalf("passthrough: got=%q err=%v", got, err)
	}

	// Short id inherits --location/--pool when issuer-* omitted.
	grp.flagPCACAIssuerCA = "ca1"
	grp.flagPCACAIssuerPool = ""
	grp.flagPCACAIssuerLocation = ""
	grp.flagPCACAPool = "pool"
	grp.flagPCACALocation = "us"
	got, err = grp.pcaResolveIssuerCA("proj")
	want := "projects/proj/locations/us/caPools/pool/certificateAuthorities/ca1"
	if err != nil || got != want {
		t.Fatalf("inherit: got=%q err=%v (want %q)", got, err, want)
	}

	// Explicit --issuer-pool/--issuer-location override.
	grp.flagPCACAIssuerPool = "otherPool"
	grp.flagPCACAIssuerLocation = "eu"
	got, _ = grp.pcaResolveIssuerCA("proj")
	want = "projects/proj/locations/eu/caPools/otherPool/certificateAuthorities/ca1"
	if got != want {
		t.Errorf("override: got=%q want=%q", got, want)
	}

	// Missing pool/location → error.
	grp.flagPCACAIssuerPool = ""
	grp.flagPCACAIssuerLocation = ""
	grp.flagPCACAPool = ""
	grp.flagPCACALocation = ""
	if _, err := grp.pcaResolveIssuerCA("proj"); err == nil {
		t.Error("expected error when both --issuer-pool/--issuer-location and --pool/--location are unset")
	}
}

func TestPrivatecaCertificatesCreateHasNotBeforeFlag(t *testing.T) {
	g := privatecaSubgroup("certificates")
	if g == nil {
		t.Fatal("privateca certificates missing")
	}
	create := findSub(g, "create")
	if create == nil {
		t.Fatal("certificates create missing")
	}
	if create.Flags().Lookup("requested-not-before-time") == nil {
		t.Error("certificates create missing --requested-not-before-time")
	}
}

func TestPrivatecaTemplatesSubcommands(t *testing.T) {
	g := privatecaSubgroup("templates")
	if g == nil {
		t.Fatal("privateca templates missing")
	}
	assertSubcommands(t, g, []string{"add-iam-policy-binding", "create", "delete", "describe", "get-iam-policy", "list", "remove-iam-policy-binding", "replicate", "set-iam-policy", "update"})
}
