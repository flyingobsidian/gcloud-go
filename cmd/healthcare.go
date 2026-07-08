package cmd

import "github.com/spf13/cobra"

// --- gcloud healthcare (#344) ---

var healthcareCmd = &cobra.Command{Use: "healthcare", Short: "Manage Cloud Healthcare API (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(healthcareCmd, "consent-stores", "Manage consent stores", append(crud, "get-iam-policy", "set-iam-policy", "evaluate-user-consents", "query-accessible-data", "check-data-access", "attributes")...)
	registerStubGroup(healthcareCmd, "datasets", "Manage datasets", append(crud, "get-iam-policy", "set-iam-policy", "deidentify", "operations")...)
	registerStubGroup(healthcareCmd, "dicom-stores", "Manage DICOM stores", append(crud, "get-iam-policy", "set-iam-policy", "export", "import", "deidentify", "search")...)
	registerStubGroup(healthcareCmd, "fhir-stores", "Manage FHIR stores", append(crud, "get-iam-policy", "set-iam-policy", "export", "import", "deidentify", "rollback")...)
	registerStubGroup(healthcareCmd, "hl7v2-stores", "Manage HL7v2 stores", append(crud, "get-iam-policy", "set-iam-policy", "export", "import")...)
	registerStubGroup(healthcareCmd, "operations", "Manage operations", "cancel", "describe", "list")
	rootCmd.AddCommand(healthcareCmd)
}
