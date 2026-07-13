package cmd

import "github.com/spf13/cobra"

// --- gcloud migration (#357) ---

var migrationCmd = &cobra.Command{Use: "migration", Short: "Migrate to Virtual Machines"}

func init() {
	vms := &cobra.Command{Use: "vms", Short: "Migrate to VMs"}
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(vms, "sources", "Manage migration sources", crud...)
	registerStubGroup(vms, "target-projects", "Manage target projects", crud...)
	registerStubGroup(vms, "groups", "Manage migration groups", append(crud, "add-migration", "remove-migration")...)
	registerStubGroup(vms, "machine-images", "Manage machine images", crud...)
	registerStubGroup(vms, "migrations", "Manage VM migrations", append(crud, "start-replication", "pause-replication", "resume-replication", "cancel-cutover", "finalize-migration", "cutover")...)
	registerStubGroup(vms, "cutover-jobs", "Manage cutover jobs", "cancel", "describe", "list")
	registerStubGroup(vms, "replication-cycles", "Manage replication cycles", "describe", "list")
	registerStubGroup(vms, "utilization-reports", "Manage utilization reports", "create", "delete", "describe", "list")
	registerStubGroup(vms, "image-imports", "Manage image imports", "create", "cancel", "describe", "list")
	registerStubGroup(vms, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	migrationCmd.AddCommand(vms)
	rootCmd.AddCommand(migrationCmd)
}
