package cmd

import "github.com/spf13/cobra"

// --- gcloud run (#380) ---

var runCmd = &cobra.Command{Use: "run", Short: "Manage Cloud Run (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(runCmd, "compose", "Docker Compose workflows", "run", "list")
	registerStubGroup(runCmd, "domain-mappings", "Manage domain mappings", "create", "delete", "describe", "list")
	registerStubGroup(runCmd, "jobs", "Manage Cloud Run jobs", append(crud, "execute", "executions")...)
	registerStubGroup(runCmd, "multi-region-services", "Manage multi-region services", crud...)
	registerStubGroup(runCmd, "regions", "View regions", "list")
	registerStubGroup(runCmd, "revisions", "Manage revisions", "delete", "describe", "list", "adjust-traffic")
	registerStubGroup(runCmd, "services", "Manage services", append(crud, "add-iam-policy-binding", "remove-iam-policy-binding", "get-iam-policy", "set-iam-policy", "logs", "proxy", "replace", "update-traffic")...)
	registerStubGroup(runCmd, "worker-pools", "Manage worker pools", crud...)
	registerStubCommand(runCmd, "deploy", "Deploy a Cloud Run service")
	rootCmd.AddCommand(runCmd)
}
