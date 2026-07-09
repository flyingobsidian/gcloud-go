package cmd

import "github.com/spf13/cobra"

// --- gcloud workbench (#396) ---

var workbenchCmd = &cobra.Command{Use: "workbench", Short: "Manage Vertex AI Workbench (stubbed)"}

func init() {
	registerStubGroup(workbenchCmd, "executions", "Manage executions", "create", "delete", "describe", "list")
	registerStubGroup(workbenchCmd, "instances", "Manage instances",
		"create", "delete", "describe", "list", "update", "start", "stop", "reset",
		"diagnose", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding",
		"remove-iam-policy-binding", "upgrade", "check-upgradability", "resize-disk",
		"migrate", "rollback", "resolve-notebooks-instance-id",
		"restore", "reset-jupyter-password", "report-event", "reject", "resize-config",
		"reset-instance-owner")
	registerStubGroup(workbenchCmd, "schedules", "Manage schedules", "create", "delete", "describe", "list", "update", "trigger", "pause", "resume")
	rootCmd.AddCommand(workbenchCmd)
}
