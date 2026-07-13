package cmd

import "github.com/spf13/cobra"

// --- gcloud ai-platform (#292) ---

var aiPlatformCmd = &cobra.Command{Use: "ai-platform", Short: "Manage AI Platform"}

func init() {
	registerStubGroup(aiPlatformCmd, "jobs", "AI Platform jobs", "cancel", "describe", "list", "stream-logs", "submit", "update")
	registerStubGroup(aiPlatformCmd, "local", "AI Platform local commands", "predict", "train")
	registerStubGroup(aiPlatformCmd, "models", "AI Platform models", "create", "delete", "describe", "list", "update", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")
	registerStubGroup(aiPlatformCmd, "operations", "AI Platform operations", "cancel", "delete", "describe", "list", "wait")
	registerStubGroup(aiPlatformCmd, "versions", "AI Platform versions", "create", "delete", "describe", "list", "set-default", "update")
	registerStubCommand(aiPlatformCmd, "predict", "Run AI Platform online prediction")
	rootCmd.AddCommand(aiPlatformCmd)
}
