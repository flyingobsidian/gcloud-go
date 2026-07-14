package cmd

import "github.com/spf13/cobra"

// --- gcloud certificate-manager (#313) ---

var certificateManagerCmd = &cobra.Command{Use: "certificate-manager", Short: "Manage Certificate Manager"}

func init() {
	// All subgroups (certificates, dns-authorizations, issuance-configs, maps
	// (with entries), operations, trust-configs) are implemented in
	// certificate_manager_all.go.
	rootCmd.AddCommand(certificateManagerCmd)
}
