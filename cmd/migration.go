package cmd

import "github.com/spf13/cobra"

// --- gcloud migration (#357, #895-#896) ---
//
// Only `vms target-projects` and `vms image-imports` are implemented against
// the real VM Migration API (see migration_vms.go). The rest of the Migrate
// to Virtual Machines surface remains stubbed here so the command paths still
// exist in --help output.

var migrationCmd = &cobra.Command{Use: "migration", Short: "Migrate to Virtual Machines"}
var migrationVMsCmd = &cobra.Command{Use: "vms", Short: "Migrate to VMs"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(migrationVMsCmd, "sources", "Manage migration sources", crud...)
	registerStubGroup(migrationVMsCmd, "groups", "Manage migration groups",
		append(crud, "add-migration", "remove-migration")...)
	registerStubGroup(migrationVMsCmd, "machine-images", "Manage machine images", crud...)
	registerStubGroup(migrationVMsCmd, "migrations", "Manage VM migrations",
		append(crud, "start-replication", "pause-replication", "resume-replication",
			"cancel-cutover", "finalize-migration", "cutover")...)
	registerStubGroup(migrationVMsCmd, "cutover-jobs", "Manage cutover jobs",
		"cancel", "describe", "list")
	registerStubGroup(migrationVMsCmd, "replication-cycles", "Manage replication cycles",
		"describe", "list")
	registerStubGroup(migrationVMsCmd, "utilization-reports", "Manage utilization reports",
		"create", "delete", "describe", "list")
	registerStubGroup(migrationVMsCmd, "operations", "Manage operations",
		"cancel", "delete", "describe", "list")
	migrationCmd.AddCommand(migrationVMsCmd)
	rootCmd.AddCommand(migrationCmd)
}
