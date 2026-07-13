package cmd

import "github.com/spf13/cobra"

// --- gcloud certificate-manager (#313) ---

var certificateManagerCmd = &cobra.Command{Use: "certificate-manager", Short: "Manage Certificate Manager"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(certificateManagerCmd, "certificates", "Manage certificates", crud...)
	registerStubGroup(certificateManagerCmd, "dns-authorizations", "Manage DNS authorizations", crud...)
	registerStubGroup(certificateManagerCmd, "issuance-configs", "Manage issuance configs", crud...)
	registerStubGroup(certificateManagerCmd, "maps", "Manage certificate maps", crud...)
	registerStubGroup(certificateManagerCmd, "operations", "Manage operations", "describe", "list")
	registerStubGroup(certificateManagerCmd, "trust-configs", "Manage trust configs", crud...)
	rootCmd.AddCommand(certificateManagerCmd)
}
