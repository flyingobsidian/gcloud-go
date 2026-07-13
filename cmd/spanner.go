package cmd

import "github.com/spf13/cobra"

// --- gcloud spanner (#387) ---

var spannerCmd = &cobra.Command{Use: "spanner", Short: "Manage Cloud Spanner"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(spannerCmd, "backups", "Manage backups", crud...)
	registerStubGroup(spannerCmd, "databases", "Manage databases", append(crud, "execute-sql", "ddl", "sessions", "restore", "add-split-points", "change-quorum")...)
	registerStubGroup(spannerCmd, "instance-configs", "Manage instance configs", crud...)
	registerStubGroup(spannerCmd, "instance-partitions", "Manage instance partitions", crud...)
	registerStubGroup(spannerCmd, "instances", "Manage instances", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding", "move")...)
	registerStubGroup(spannerCmd, "operations", "Manage operations", "cancel", "describe", "list")
	registerStubGroup(spannerCmd, "rows", "Manage rows", "delete", "insert", "read", "update")
	registerStubGroup(spannerCmd, "samples", "Sample apps", "list", "run")
	registerStubCommand(spannerCmd, "cli", "Interactive Spanner shell")
		registerStubGroup(spannerCmd, "backup-schedules", "Manage backup-schedules", "list", "describe")
	rootCmd.AddCommand(spannerCmd)
}
