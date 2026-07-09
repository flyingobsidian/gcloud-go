package cmd

import "github.com/spf13/cobra"

// --- gcloud transfer (#393) ---

var transferCmd = &cobra.Command{Use: "transfer", Short: "Manage Storage Transfer Service (stubbed)"}

func init() {
	registerStubGroup(transferCmd, "agent-pools", "Manage on-premise transfer agent pools", "create", "delete", "describe", "list", "update")
	registerStubGroup(transferCmd, "agents", "Manage transfer agents", "install", "delete", "list")
	registerStubGroup(transferCmd, "jobs", "Manage transfer jobs", "create", "delete", "describe", "list", "update", "run", "monitor")
	registerStubGroup(transferCmd, "operations", "Manage transfer operations", "cancel", "describe", "list", "monitor", "pause", "resume")
	registerStubCommand(transferCmd, "authorize", "Authorize an account for Transfer Service features")
	rootCmd.AddCommand(transferCmd)
}
