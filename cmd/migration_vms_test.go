package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func migrationVMsSubgroup(name string) *cobra.Command {
	for _, c := range migrationVMsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestMigrationVMsTargetProjectsSubcommands(t *testing.T) {
	g := migrationVMsSubgroup("target-projects")
	if g == nil {
		t.Fatal("migration vms target-projects missing")
	}
	assertSubcommands(t, g, []string{"add", "delete", "describe", "list", "update"})
}

func TestMigrationVMsImageImportsSubcommands(t *testing.T) {
	g := migrationVMsSubgroup("image-imports")
	if g == nil {
		t.Fatal("migration vms image-imports missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestMigrationVMsResourceName(t *testing.T) {
	parent := "projects/foo/locations/us-central1"
	if got := mvmResourceName(parent, "imageImports", "bar"); got != parent+"/imageImports/bar" {
		t.Errorf("mvmResourceName rel: got %s", got)
	}
	full := "projects/x/locations/y/imageImports/z"
	if got := mvmResourceName(parent, "imageImports", full); got != full {
		t.Errorf("mvmResourceName full: got %s", got)
	}
}
