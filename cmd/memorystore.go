package cmd

import "github.com/spf13/cobra"

// --- gcloud memorystore (#355) ---

var memorystoreCmd = &cobra.Command{Use: "memorystore", Short: "Manage Memorystore for Valkey"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(memorystoreCmd, "acl-policies", "Manage ACL policies", crud...)
	registerStubGroup(memorystoreCmd, "backup-collections", "Manage backup collections", crud...)
	registerStubGroup(memorystoreCmd, "instances", "Manage instances", append(crud, "backup", "reschedule-maintenance", "failover", "certificate-authorities")...)
	registerStubGroup(memorystoreCmd, "locations", "Manage locations", "list", "describe")
	registerStubGroup(memorystoreCmd, "operations", "Manage operations", "cancel", "describe", "list")
	rootCmd.AddCommand(memorystoreCmd)
}
