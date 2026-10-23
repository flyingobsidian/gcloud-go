package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	observability "google.golang.org/api/observability/v1"
)

func observabilitySubgroup(name string) *cobra.Command {
	for _, c := range observabilityCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestObservabilityScopesSubcommands(t *testing.T) {
	g := observabilitySubgroup("scopes")
	if g == nil {
		t.Fatal("observability scopes missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestObservabilityTraceScopesSubcommands(t *testing.T) {
	g := observabilitySubgroup("trace-scopes")
	if g == nil {
		t.Fatal("observability trace-scopes missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestObservabilityBucketsCommands(t *testing.T) {
	g := findSub(observabilityCmd, "buckets")
	if g == nil {
		t.Fatal("observability buckets missing")
	}
	assertSubcommands(t, g, []string{"create", "describe", "list", "update"})
	for _, f := range []string{"async", "display-name", "description", "cmek-settings-kms-key", "clear-cmek-settings"} {
		if obBucketUpdateCmd.Flags().Lookup(f) == nil {
			t.Errorf("--%s missing on buckets update", f)
		}
	}
}

func TestObBucketUpdateMask(t *testing.T) {
	defer func() {
		flagObBucketDisplayName, flagObBucketKmsKey, flagObBucketClearCmek = "", "", false
		for _, f := range []string{"display-name", "cmek-settings-kms-key"} {
			obBucketUpdateCmd.Flags().Lookup(f).Changed = false
		}
	}()
	if err := obBucketUpdateCmd.Flags().Set("display-name", "My bucket"); err != nil {
		t.Fatal(err)
	}
	if err := obBucketUpdateCmd.Flags().Set("cmek-settings-kms-key", "projects/p/locations/l/keyRings/r/cryptoKeys/k"); err != nil {
		t.Fatal(err)
	}
	b := &observability.Bucket{Description: "keep"}
	if mask := obBucketUpdateMask(obBucketUpdateCmd, b); mask != "displayName,cmekSettings.kmsKey" {
		t.Errorf("mask = %q", mask)
	}
	if b.DisplayName != "My bucket" || b.Description != "keep" || b.CmekSettings.KmsKey == "" {
		t.Errorf("bucket = %+v", b)
	}
}
