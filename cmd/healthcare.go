package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud healthcare (#344) ---

var healthcareCmd = &cobra.Command{Use: "healthcare", Short: "Manage Cloud Healthcare API"}

func init() {
	// consent-stores (#1220), dicom-stores (#1222), fhir-stores (#1223),
	// hl7v2-stores (#1224), datasets (#1221) and operations (#1225) are
	// implemented in dedicated files. Specialty verbs (deidentify/export/
	// import/rollback/search/etc.) will land under separate issues.
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
