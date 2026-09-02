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

// TestCMTagsFlagOnCreates locks in that --tags is exposed on the five create
// commands listed in the gcloud-python 580.0.0 release notes but not on
// update, delete, describe, list, or `maps entries create` (which uses a
// separate registration path).
func TestCMTagsFlagOnCreates(t *testing.T) {
	cases := []struct {
		group string
	}{{"certificates"}, {"dns-authorizations"}, {"issuance-configs"}, {"maps"}, {"trust-configs"}}
	for _, tc := range cases {
		g := cmSubgroup(tc.group)
		if g == nil {
			t.Fatalf("%s missing", tc.group)
		}
		create := findSub(g, "create")
		if create == nil {
			t.Fatalf("%s create missing", tc.group)
		}
		if create.Flags().Lookup("tags") == nil {
			t.Errorf("%s create missing --tags", tc.group)
		}
		for _, verb := range []string{"update", "delete", "describe", "list"} {
			sub := findSub(g, verb)
			if sub != nil && sub.Flags().Lookup("tags") != nil {
				t.Errorf("%s %s should not expose --tags", tc.group, verb)
			}
		}
	}
	// maps entries create uses its own flag registration and should stay
	// clear of --tags.
	maps := cmSubgroup("maps")
	entries := findSub(maps, "entries")
	if entries != nil {
		if entry := findSub(entries, "create"); entry != nil && entry.Flags().Lookup("tags") != nil {
			t.Error("maps entries create should not expose --tags")
		}
	}
}

func TestCMMergeTags(t *testing.T) {
	orig := flagCMTags
	t.Cleanup(func() { flagCMTags = orig })

	// Empty flag → keep existing map (including nil) untouched.
	flagCMTags = nil
	if got := cmMergeTags(nil); got != nil {
		t.Errorf("empty flag, nil input: got=%v want=nil", got)
	}
	existing := map[string]string{"k": "v"}
	got := cmMergeTags(existing)
	if len(got) != 1 || got["k"] != "v" {
		t.Errorf("empty flag, existing input mutated: got=%v", got)
	}

	// Flag overlays and wins on collision.
	flagCMTags = map[string]string{"k": "override", "b": "2"}
	got = cmMergeTags(map[string]string{"k": "v", "a": "1"})
	if got["k"] != "override" || got["a"] != "1" || got["b"] != "2" {
		t.Errorf("overlay: got=%v", got)
	}

	// Flag with nil input allocates a fresh map.
	got = cmMergeTags(nil)
	if got["k"] != "override" || got["b"] != "2" {
		t.Errorf("allocate: got=%v", got)
	}
}
