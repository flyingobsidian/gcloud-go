package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	apihub "google.golang.org/api/apihub/v1"
)

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

// TestApihubApisImportServiceType checks that the 588.0.0 serviceType field
// loads from an `apis import` source file and lands in the update mask.
func TestApihubApisImportServiceType(t *testing.T) {
	f := filepath.Join(t.TempDir(), "api.yaml")
	src := "displayName: x\nserviceType:\n  enumValues:\n    values:\n    - id: agent\n"
	if err := os.WriteFile(f, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	body := &apihub.GoogleCloudApihubV1Api{}
	if err := loadYAMLOrJSONInto(f, body); err != nil {
		t.Fatal(err)
	}
	if body.ServiceType == nil || body.ServiceType.EnumValues == nil ||
		len(body.ServiceType.EnumValues.Values) != 1 || body.ServiceType.EnumValues.Values[0].Id != "agent" {
		t.Fatalf("serviceType not loaded: %+v", body.ServiceType)
	}
	if mask := joinMask(nonEmptyJSONFields(body)); !strings.Contains(mask, "serviceType") {
		t.Errorf("update mask %q missing serviceType", mask)
	}
}
