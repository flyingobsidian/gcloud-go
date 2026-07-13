package cmd

import "github.com/spf13/cobra"

// --- gcloud edge-cache (#333) ---

var edgeCacheCmd = &cobra.Command{Use: "edge-cache", Short: "Manage Media CDN Edge Cache"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(edgeCacheCmd, "keysets", "Manage EdgeCacheKeyset resources", crud...)
	registerStubGroup(edgeCacheCmd, "operations", "Manage operations", "describe", "list", "cancel", "delete")
	registerStubGroup(edgeCacheCmd, "origins", "Manage EdgeCacheOrigin resources", crud...)
	registerStubGroup(edgeCacheCmd, "services", "Manage EdgeCacheService resources", crud...)
	rootCmd.AddCommand(edgeCacheCmd)
}
