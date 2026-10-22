package cmd

import (
	"bytes"
	"testing"

	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

func TestValidateACMLookupResource(t *testing.T) {
	for _, ok := range []string{"projects/123", "folders/456"} {
		if err := validateACMLookupResource(ok); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	want := `Invalid value for [--resource]: Resource must start with "projects/" or "folders/". Got: organizations/1`
	if err := validateACMLookupResource("organizations/1"); err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
}

func TestPrintACMConfiguredPerimeter(t *testing.T) {
	var buf bytes.Buffer
	printACMConfiguredPerimeter(&buf, &accesscontextmanager.LookupConfiguredServicePerimeterResponse{
		ServicePerimeter:   "accessPolicies/1/servicePerimeters/p",
		RestrictedResource: "projects/123",
	})
	want := "SERVICE_PERIMETER:           accessPolicies/1/servicePerimeters/p\n" +
		"RESTRICTED_RESOURCE:         projects/123\n" +
		"SERVICE_PERIMETER_DRY_RUN:   \n" +
		"RESTRICTED_RESOURCE_DRY_RUN: \n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestACMLookupConfiguredPerimeterRegistered(t *testing.T) {
	c, _, err := accessContextManagerCmd.Find([]string{"lookup-configured-perimeter"})
	if err != nil || c != acmLookupConfiguredPerimeterCmd {
		t.Fatalf("lookup-configured-perimeter not registered: %v", err)
	}
	if c.Flags().Lookup("resource") == nil {
		t.Error("missing --resource")
	}
}
