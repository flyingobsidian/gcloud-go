package cmd

import "github.com/spf13/cobra"

// --- gcloud biglake (#307) ---

var biglakeCmd = &cobra.Command{
	Use:   "biglake",
	Short: "Manage BigLake resources (stubbed)",
}

func init() {
	iceberg := &cobra.Command{Use: "iceberg", Short: "BigLake Iceberg REST catalogs"}
	registerStubGroup(iceberg, "catalogs", "Manage Iceberg REST catalogs", "create", "delete", "describe", "list", "update")
	biglakeCmd.AddCommand(iceberg)
	rootCmd.AddCommand(biglakeCmd)
}
