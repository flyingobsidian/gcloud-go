package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func dataplexSubgroup(name string) *cobra.Command {
	for _, c := range dataplexCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// #1774: --enable-catalog-publishing and --mode on `dataplex datascans create`
// (covers the create + update data-documentation/data-profile bullet), plus
// `dataplex context lookup --all-schema-fields`.
func TestDataplex_1774_DatascansCreateFlags(t *testing.T) {
	if dataplexDatascansCreateCmd.Flags().Lookup("enable-catalog-publishing") == nil {
		t.Error("dataplex datascans create: --enable-catalog-publishing flag missing")
	}
	if dataplexDatascansCreateCmd.Flags().Lookup("mode") == nil {
		t.Error("dataplex datascans create: --mode flag missing")
	}
}

func TestDataplex_1774_ContextLookup(t *testing.T) {
	ctx := dataplexSubgroup("context")
	if ctx == nil {
		t.Fatal("dataplex context missing")
	}
	var lookup *cobra.Command
	for _, c := range ctx.Commands() {
		if c.Name() == "lookup" {
			lookup = c
			break
		}
	}
	if lookup == nil {
		t.Fatal("dataplex context lookup missing")
	}
	for _, name := range []string{"resources", "all-schema-fields", "context-format", "options"} {
		if lookup.Flags().Lookup(name) == nil {
			t.Errorf("dataplex context lookup: --%s flag missing", name)
		}
	}
}
