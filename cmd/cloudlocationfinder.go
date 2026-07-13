package cmd

import "github.com/spf13/cobra"

// --- gcloud cloudlocationfinder (#315) ---

var cloudLocationFinderCmd = &cobra.Command{Use: "cloudlocationfinder", Short: "Manage Cloud Location Finder"}

func init() {
	registerStubGroup(cloudLocationFinderCmd, "cloud-locations", "Manage cloud locations", "describe", "list")
	rootCmd.AddCommand(cloudLocationFinderCmd)
}
