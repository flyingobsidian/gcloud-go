package cmd

import "github.com/spf13/cobra"

// --- gcloud migration (#357) ---
//
// vms target-projects, image-imports, machine-image-imports, and disk-migrations
// are implemented against the real VM Migration API in migration_vms.go and
// migration_vms_extra.go. gcloud-python's `migration vms` surface is limited to
// exactly these four subgroups; earlier fake stubs (sources, groups, migrations,
// cutover-jobs, replication-cycles, utilization-reports, operations, machine-images)
// have been removed as they don't exist in gcloud-python.

var migrationCmd = &cobra.Command{Use: "migration", Short: "Migrate to Virtual Machines"}
var migrationVMsCmd = &cobra.Command{Use: "vms", Short: "Migrate to VMs"}

func init() {
	migrationCmd.AddCommand(migrationVMsCmd)
	rootCmd.AddCommand(migrationCmd)
}
