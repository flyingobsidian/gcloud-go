package cmd

import "github.com/spf13/cobra"

// --- gcloud logging (#350) ---

var loggingCmd = &cobra.Command{Use: "logging", Short: "Manage Cloud Logging"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(loggingCmd, "buckets", "Manage log buckets", append(crud, "undelete")...)
	registerStubGroup(loggingCmd, "links", "Manage log bucket links", "create", "delete", "describe", "list")
	registerStubGroup(loggingCmd, "locations", "Manage locations", "list", "describe")
	registerStubGroup(loggingCmd, "logs", "Manage logs", "delete", "list")
	registerStubGroup(loggingCmd, "metrics", "Manage logs-based metrics", crud...)
	registerStubGroup(loggingCmd, "operations", "Manage operations", "cancel", "describe", "list", "list-recent")
	registerStubGroup(loggingCmd, "recent-queries", "Manage recent queries", "list")
	registerStubGroup(loggingCmd, "resource-descriptors", "Resource descriptor info", "list", "describe")
	registerStubGroup(loggingCmd, "saved-queries", "Manage saved queries", crud...)
	registerStubGroup(loggingCmd, "scopes", "Manage log scopes", crud...)
	registerStubGroup(loggingCmd, "settings", "Manage router settings", "describe", "update")
	registerStubGroup(loggingCmd, "sinks", "Manage log sinks", crud...)
	registerStubGroup(loggingCmd, "views", "Manage log views", crud...)
	for _, name := range []string{"copy", "read", "write", "tail"} {
		registerStubCommand(loggingCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(loggingCmd)
}
