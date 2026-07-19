package cmd

import "testing"

func TestHealthcareConsentStoresSubcommands(t *testing.T) {
	g := healthcareSubgroup("consent-stores")
	if g == nil {
		t.Fatal("healthcare consent-stores missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy",
	})
}

func TestHealthcareDicomStoresSubcommands(t *testing.T) {
	g := healthcareSubgroup("dicom-stores")
	if g == nil {
		t.Fatal("healthcare dicom-stores missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy",
	})
}

func TestHealthcareFhirStoresSubcommands(t *testing.T) {
	g := healthcareSubgroup("fhir-stores")
	if g == nil {
		t.Fatal("healthcare fhir-stores missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy",
	})
}

func TestHealthcareHl7v2StoresSubcommands(t *testing.T) {
	g := healthcareSubgroup("hl7v2-stores")
	if g == nil {
		t.Fatal("healthcare hl7v2-stores missing")
	}
	assertSubcommands(t, g, []string{
		"create", "delete", "describe", "list", "update",
		"get-iam-policy", "set-iam-policy",
	})
}
