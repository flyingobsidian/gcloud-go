package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func oracleSubgroup(name string) *cobra.Command {
	for _, c := range oracleDatabaseCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestOracleAutonomousDatabaseCharacterSetsSubcommands(t *testing.T) {
	g := oracleSubgroup("autonomous-database-character-sets")
	if g == nil {
		t.Fatal("oracle-database autonomous-database-character-sets missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleDbSystemShapesSubcommands(t *testing.T) {
	g := oracleSubgroup("db-system-shapes")
	if g == nil {
		t.Fatal("oracle-database db-system-shapes missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleEntitlementsSubcommands(t *testing.T) {
	g := oracleSubgroup("entitlements")
	if g == nil {
		t.Fatal("oracle-database entitlements missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleGiVersionsSubcommands(t *testing.T) {
	g := oracleSubgroup("gi-versions")
	if g == nil {
		t.Fatal("oracle-database gi-versions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleOperationsSubcommands(t *testing.T) {
	g := oracleSubgroup("operations")
	if g == nil {
		t.Fatal("oracle-database operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}

func TestOracleAutonomousDbVersionsSubcommands(t *testing.T) {
	g := oracleSubgroup("autonomous-db-versions")
	if g == nil {
		t.Fatal("oracle-database autonomous-db-versions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleDatabaseCharacterSetsSubcommands(t *testing.T) {
	g := oracleSubgroup("database-character-sets")
	if g == nil {
		t.Fatal("oracle-database database-character-sets missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleDbSystemInitialStorageSizesSubcommands(t *testing.T) {
	g := oracleSubgroup("db-system-initial-storage-sizes")
	if g == nil {
		t.Fatal("oracle-database db-system-initial-storage-sizes missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleDbVersionsSubcommands(t *testing.T) {
	g := oracleSubgroup("db-versions")
	if g == nil {
		t.Fatal("oracle-database db-versions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}
