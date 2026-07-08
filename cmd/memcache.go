package cmd

import "github.com/spf13/cobra"

// --- gcloud memcache (#354) ---

var memcacheCmd = &cobra.Command{Use: "memcache", Short: "Manage Memorystore for Memcached (stubbed)"}

func init() {
	registerStubGroup(memcacheCmd, "instances", "Manage instances", "create", "delete", "describe", "list", "update", "upgrade", "apply-parameters")
	registerStubGroup(memcacheCmd, "operations", "Manage operations", "cancel", "describe", "list")
	registerStubGroup(memcacheCmd, "regions", "Manage regions", "list", "describe")
	rootCmd.AddCommand(memcacheCmd)
}
