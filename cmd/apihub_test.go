package cmd

import "testing"

func TestApihubApisSubcommands(t *testing.T) {
	g := apihubSubgroup("apis")
	if g == nil {
		t.Fatal("apihub apis missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestApihubAttributesSubcommands(t *testing.T) {
	g := apihubSubgroup("attributes")
	if g == nil {
		t.Fatal("apihub attributes missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestApihubCurationsSubcommands(t *testing.T) {
	g := apihubSubgroup("curations")
	if g == nil {
		t.Fatal("apihub curations missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestApihubDependenciesSubcommands(t *testing.T) {
	g := apihubSubgroup("dependencies")
	if g == nil {
		t.Fatal("apihub dependencies missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestApihubDeploymentsSubcommands(t *testing.T) {
	g := apihubSubgroup("deployments")
	if g == nil {
		t.Fatal("apihub deployments missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestApihubExternalApisSubcommands(t *testing.T) {
	g := apihubSubgroup("external-apis")
	if g == nil {
		t.Fatal("apihub external-apis missing")
	}
	assertSubcommands(t, g, []string{"delete", "describe", "export", "import", "list"})
}

func TestApihubDiscoveredApiObservationsSubcommands(t *testing.T) {
	g := apihubSubgroup("discovered-api-observations")
	if g == nil {
		t.Fatal("apihub discovered-api-observations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestApihubHostProjectRegistrationsSubcommands(t *testing.T) {
	g := apihubSubgroup("host-project-registrations")
	if g == nil {
		t.Fatal("apihub host-project-registrations missing")
	}
	assertSubcommands(t, g, []string{"create", "describe", "list"})
}

func TestApihubOperationsSubcommands(t *testing.T) {
	g := apihubSubgroup("operations")
	if g == nil {
		t.Fatal("apihub operations missing")
	}
	assertSubcommands(t, g, []string{"cancel", "delete", "describe", "list"})
}
