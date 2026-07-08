package cmd

import "github.com/spf13/cobra"

// --- gcloud domains (#332) ---

var domainsCmd = &cobra.Command{Use: "domains", Short: "Manage Cloud Domains (stubbed)"}

func init() {
	registerStubGroup(domainsCmd, "registrations", "Manage domain registrations",
		"register", "search-domains", "get-register-parameters",
		"describe", "list", "delete", "update", "renew-domain",
		"retrieve-authorization-code", "reset-authorization-code",
		"transfer", "get-transfer-parameters",
		"configure", "export", "cancel-export", "initiate-push-transfer")
	for _, name := range []string{"list-user-verified", "verify"} {
		registerStubCommand(domainsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(domainsCmd)
}
