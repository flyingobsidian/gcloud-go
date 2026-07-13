package cmd

import "github.com/spf13/cobra"

// --- gcloud dataproc (#324) ---

var dataprocCmd = &cobra.Command{Use: "dataproc", Short: "Manage Dataproc"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(dataprocCmd, "autoscaling-policies", "Manage autoscaling policies", append(crud, "import", "export", "get-iam-policy", "set-iam-policy")...)
	registerStubGroup(dataprocCmd, "batches", "Submit batch jobs", append(crud, "cancel", "submit", "wait")...)
	registerStubGroup(dataprocCmd, "clusters", "Manage clusters", append(crud, "diagnose", "start", "stop", "import", "export", "get-iam-policy", "set-iam-policy")...)
	registerStubGroup(dataprocCmd, "jobs", "Manage jobs", append(crud, "submit", "kill", "wait")...)
	registerStubGroup(dataprocCmd, "node-groups", "Manage node groups", append(crud, "resize", "repair")...)
	registerStubGroup(dataprocCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list", "wait")
	registerStubGroup(dataprocCmd, "workflow-templates", "Manage workflow templates", append(crud, "instantiate", "instantiate-from-file", "import", "export", "add-job", "remove-job", "set-managed-cluster", "set-cluster-selector", "run")...)
	rootCmd.AddCommand(dataprocCmd)
}
