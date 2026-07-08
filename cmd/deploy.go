package cmd

import "github.com/spf13/cobra"

// --- gcloud deploy (#327) ---

var deployCmd = &cobra.Command{Use: "deploy", Short: "Manage Cloud Deploy (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(deployCmd, "automation-runs", "Manage automation runs", "cancel", "describe", "list")
	registerStubGroup(deployCmd, "automations", "Manage automations", crud...)
	registerStubGroup(deployCmd, "custom-target-types", "Manage custom target types", crud...)
	registerStubGroup(deployCmd, "delivery-pipelines", "Manage delivery pipelines", append(crud, "rollback")...)
	registerStubGroup(deployCmd, "deploy-policies", "Manage deploy policies", crud...)
	registerStubGroup(deployCmd, "job-runs", "Manage job runs", "describe", "list", "retry", "terminate")
	registerStubGroup(deployCmd, "releases", "Manage releases", append(crud, "promote", "abandon")...)
	registerStubGroup(deployCmd, "rollouts", "Manage rollouts", append(crud, "approve", "advance", "cancel", "ignore-job", "retry-job", "terminate-job")...)
	registerStubGroup(deployCmd, "targets", "Manage targets", crud...)
	for _, name := range []string{"apply", "delete", "get-config"} {
		registerStubCommand(deployCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(deployCmd)
}
