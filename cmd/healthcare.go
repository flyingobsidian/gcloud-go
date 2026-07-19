package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud healthcare (#344) ---

var healthcareCmd = &cobra.Command{Use: "healthcare", Short: "Manage Cloud Healthcare API"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	// consent-stores, dicom-stores, fhir-stores, hl7v2-stores are still
	// stubs pending their own issues; datasets (#1221) and operations
	// (#1225) are implemented in dedicated files.
	registerStubGroup(healthcareCmd, "consent-stores", "Manage consent stores", append(crud, "get-iam-policy", "set-iam-policy", "evaluate-user-consents", "query-accessible-data", "check-data-access", "attributes")...)
	registerStubGroup(healthcareCmd, "dicom-stores", "Manage DICOM stores", append(crud, "get-iam-policy", "set-iam-policy", "export", "import", "deidentify", "search")...)
	registerStubGroup(healthcareCmd, "fhir-stores", "Manage FHIR stores", append(crud, "get-iam-policy", "set-iam-policy", "export", "import", "deidentify", "rollback")...)
	registerStubGroup(healthcareCmd, "hl7v2-stores", "Manage HL7v2 stores", append(crud, "get-iam-policy", "set-iam-policy", "export", "import")...)
	rootCmd.AddCommand(healthcareCmd)
}

// hcLocationParent returns projects/PROJECT/locations/LOCATION.
func hcLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// hcDatasetName returns projects/.../datasets/DATASET.
func hcDatasetName(location, dataset string) (string, error) {
	if strings.HasPrefix(dataset, "projects/") {
		return dataset, nil
	}
	parent, err := hcLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/datasets/%s", parent, dataset), nil
}
