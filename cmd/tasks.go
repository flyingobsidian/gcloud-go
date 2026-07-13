package cmd

import "github.com/spf13/cobra"

// --- gcloud tasks (#389) ---

var tasksCmd = &cobra.Command{Use: "tasks", Short: "Manage Cloud Tasks"}

func init() {
	registerStubGroup(tasksCmd, "locations", "Manage locations", "describe", "list")
	registerStubGroup(tasksCmd, "queues", "Manage queues",
		"create", "delete", "describe", "list", "update", "pause", "purge", "resume",
		"get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")
	for _, name := range []string{"buffer", "create-app-engine-task", "create-http-task", "delete", "describe", "list", "run"} {
		registerStubCommand(tasksCmd, name, "Not yet implemented")
	}
		registerStubGroup(tasksCmd, "cmek-config", "Manage cmek-config", "list", "describe")
	rootCmd.AddCommand(tasksCmd)
}
