package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// dmSubgroup returns the named subcommand of `database-migration`, or nil.
func dmSubgroup(name string) *cobra.Command {
	for _, c := range databaseMigrationCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestDMConversionWorkspacesSubcommands(t *testing.T) {
	g := dmSubgroup("conversion-workspaces")
	if g == nil {
		t.Fatal("database-migration conversion-workspaces missing")
	}
	assertSubcommands(t, g, []string{
		"apply", "commit", "convert", "create", "delete", "describe", "describe-ddls",
		"describe-entities", "describe-issues", "import-rules", "list",
		"list-background-jobs", "mapping-rules", "rollback", "seed", "update",
	})
	mr := findSub(g, "mapping-rules")
	if mr == nil {
		t.Fatal("mapping-rules subgroup missing")
	}
	assertSubcommands(t, mr, []string{"create", "delete", "describe", "list"})
}

func TestDMMigrationJobsSubcommands(t *testing.T) {
	g := dmSubgroup("migration-jobs")
	if g == nil {
		t.Fatal("database-migration migration-jobs missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "demote-destination", "describe", "fetch-source-objects",
		"generate-ssh-script", "list", "promote", "restart", "resume", "start", "stop",
		"update", "verify",
	})
}

func TestDMObjectsSubcommands(t *testing.T) {
	g := dmSubgroup("objects")
	if g == nil {
		t.Fatal("database-migration objects missing")
	}
	assertSubcommands(t, g, []string{"list", "lookup"})
}

func TestDMOperationsSubcommands(t *testing.T) {
	g := dmSubgroup("operations")
	if g == nil {
		t.Fatal("database-migration operations missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "list"})
}

func TestDMPrivateConnectionsSubcommands(t *testing.T) {
	g := dmSubgroup("private-connections")
	if g == nil {
		t.Fatal("database-migration private-connections missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestJoinMask(t *testing.T) {
	if got := joinMask([]string{"a", "b", "c"}); got != "a,b,c" {
		t.Errorf("joinMask = %q, want a,b,c", got)
	}
	if got := joinMask(nil); got != "" {
		t.Errorf("joinMask(nil) = %q, want empty", got)
	}
}

func TestNonEmptyJSONFieldsOnMigrationJob(t *testing.T) {
	type inline struct {
		DisplayName string `json:"displayName,omitempty"`
		Source      string `json:"source,omitempty"`
		Skipped     string `json:"-"`
	}
	got := nonEmptyJSONFields(inline{DisplayName: "x", Source: "src"})
	set := map[string]bool{}
	for _, f := range got {
		set[f] = true
	}
	if !set["displayName"] || !set["source"] {
		t.Errorf("expected displayName+source in %v", got)
	}
	if set["Skipped"] || set["-"] {
		t.Errorf("did not expect skipped fields: %v", got)
	}
}

func findSub(c *cobra.Command, name string) *cobra.Command {
	for _, sc := range c.Commands() {
		if sc.Name() == name {
			return sc
		}
	}
	return nil
}

func assertSubcommands(t *testing.T, c *cobra.Command, want []string) {
	t.Helper()
	got := map[string]bool{}
	for _, sc := range c.Commands() {
		got[sc.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("%s missing subcommand %q", c.Name(), name)
		}
	}
}
