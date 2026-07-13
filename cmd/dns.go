package cmd

import "github.com/spf13/cobra"

// --- gcloud dns (#331) ---

var dnsCmd = &cobra.Command{Use: "dns", Short: "Manage Cloud DNS"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(dnsCmd, "dns-keys", "Manage DNSKEY records", "describe", "list")
	registerStubGroup(dnsCmd, "managed-zones", "Manage managed zones", crud...)
	registerStubGroup(dnsCmd, "operations", "Manage operations", "describe", "list")
	registerStubGroup(dnsCmd, "policies", "Manage DNS policies", crud...)
	registerStubGroup(dnsCmd, "project-info", "View DNS project info", "describe")
	registerStubGroup(dnsCmd, "record-sets", "Manage record sets", append(crud, "import", "export", "transaction", "changes")...)
	registerStubGroup(dnsCmd, "response-policies", "Manage response policies", append(crud, "rules")...)
	rootCmd.AddCommand(dnsCmd)
}
