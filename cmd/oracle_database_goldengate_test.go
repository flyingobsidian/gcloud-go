package cmd

import "testing"

func TestOracleGoldengateConnectionAssignmentsSubcommands(t *testing.T) {
	g := oracleSubgroup("goldengate-connection-assignments")
	if g == nil {
		t.Fatal("oracle-database goldengate-connection-assignments missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "test"})
}

func TestOracleGoldengateConnectionTypesSubcommands(t *testing.T) {
	g := oracleSubgroup("goldengate-connection-types")
	if g == nil {
		t.Fatal("oracle-database goldengate-connection-types missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleGoldengateConnectionsSubcommands(t *testing.T) {
	g := oracleSubgroup("goldengate-connections")
	if g == nil {
		t.Fatal("oracle-database goldengate-connections missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestOracleGoldengateDeploymentEnvironmentsSubcommands(t *testing.T) {
	g := oracleSubgroup("goldengate-deployment-environments")
	if g == nil {
		t.Fatal("oracle-database goldengate-deployment-environments missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleGoldengateDeploymentTypesSubcommands(t *testing.T) {
	g := oracleSubgroup("goldengate-deployment-types")
	if g == nil {
		t.Fatal("oracle-database goldengate-deployment-types missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleGoldengateDeploymentVersionsSubcommands(t *testing.T) {
	g := oracleSubgroup("goldengate-deployment-versions")
	if g == nil {
		t.Fatal("oracle-database goldengate-deployment-versions missing")
	}
	assertSubcommands(t, g, []string{"list"})
}

func TestOracleGoldengateDeploymentsSubcommands(t *testing.T) {
	g := oracleSubgroup("goldengate-deployments")
	if g == nil {
		t.Fatal("oracle-database goldengate-deployments missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "start", "stop"})
}
