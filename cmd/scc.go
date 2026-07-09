package cmd

import "github.com/spf13/cobra"

// --- gcloud scc (#381) ---

var sccCmd = &cobra.Command{Use: "scc", Short: "Manage Security Command Center (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(sccCmd, "assets", "Manage assets", "describe", "group", "list", "list-assets", "update-security-marks", "run-discovery")
	registerStubGroup(sccCmd, "bqexports", "Manage BigQuery exports", crud...)
	registerStubGroup(sccCmd, "custom-modules", "Manage custom modules", "etd", "sha")
	registerStubGroup(sccCmd, "findings", "Manage findings", append(crud, "group", "set-state", "set-mute", "update-security-marks", "bulk-mute", "list")...)
	registerStubGroup(sccCmd, "iac-validation-reports", "Manage IaC validation reports", "describe", "list", "create")
	registerStubGroup(sccCmd, "manage", "Manage SCC settings", "services", "settings")
	registerStubGroup(sccCmd, "muteconfigs", "Manage mute configs", crud...)
	registerStubGroup(sccCmd, "notifications", "Manage notifications", crud...)
	registerStubGroup(sccCmd, "operations", "Manage SCC operations", "cancel", "delete", "describe", "list", "wait")
	registerStubGroup(sccCmd, "posture-deployments", "Manage posture deployments", crud...)
	registerStubGroup(sccCmd, "posture-operations", "Manage posture operations", "describe", "list", "cancel")
	registerStubGroup(sccCmd, "posture-templates", "Manage posture templates", "describe", "list")
	registerStubGroup(sccCmd, "postures", "Manage postures", append(crud, "extract")...)
	registerStubGroup(sccCmd, "sources", "Manage finding sources", crud...)
	rootCmd.AddCommand(sccCmd)
}
