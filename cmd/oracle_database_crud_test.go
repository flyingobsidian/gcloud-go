package cmd

import "testing"

func TestOracleCloudExadataInfrastructuresSubcommands(t *testing.T) {
	g := oracleSubgroup("cloud-exadata-infrastructures")
	if g == nil {
		t.Fatal("oracle-database cloud-exadata-infrastructures missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestOracleCloudVmClustersSubcommands(t *testing.T) {
	g := oracleSubgroup("cloud-vm-clusters")
	if g == nil {
		t.Fatal("oracle-database cloud-vm-clusters missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestOracleDatabasesSubcommands(t *testing.T) {
	g := oracleSubgroup("databases")
	if g == nil {
		t.Fatal("oracle-database databases missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestOracleDbSystemsSubcommands(t *testing.T) {
	g := oracleSubgroup("db-systems")
	if g == nil {
		t.Fatal("oracle-database db-systems missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestOracleExadbVmClustersSubcommands(t *testing.T) {
	g := oracleSubgroup("exadb-vm-clusters")
	if g == nil {
		t.Fatal("oracle-database exadb-vm-clusters missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "update"})
}

func TestOracleExascaleDbStorageVaultsSubcommands(t *testing.T) {
	g := oracleSubgroup("exascale-db-storage-vaults")
	if g == nil {
		t.Fatal("oracle-database exascale-db-storage-vaults missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestOracleOdbNetworksSubcommands(t *testing.T) {
	g := oracleSubgroup("odb-networks")
	if g == nil {
		t.Fatal("oracle-database odb-networks missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestOraclePluggableDatabasesSubcommands(t *testing.T) {
	g := oracleSubgroup("pluggable-databases")
	if g == nil {
		t.Fatal("oracle-database pluggable-databases missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}
