package cmd

import "github.com/spf13/cobra"

// --- gcloud vector-search (#394) ---

var vectorSearchCmd = &cobra.Command{Use: "vector-search", Short: "Manage Vector Search"}

func init() {
	registerStubGroup(vectorSearchCmd, "collections", "Manage collections",
		"create", "delete", "describe", "list", "update", "upload-records", "delete-records", "query")
	registerStubGroup(vectorSearchCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	rootCmd.AddCommand(vectorSearchCmd)
}
