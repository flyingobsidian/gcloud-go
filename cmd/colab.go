package cmd

import "github.com/spf13/cobra"

// --- gcloud colab (#316) ---

var colabCmd = &cobra.Command{Use: "colab", Short: "Manage Colab Enterprise (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list"}
	registerStubGroup(colabCmd, "executions", "Manage notebook executions", crud...)
	registerStubGroup(colabCmd, "runtime-templates", "Manage runtime templates", append(crud, "update")...)
	registerStubGroup(colabCmd, "runtimes", "Manage runtimes", append(crud, "start", "stop")...)
	registerStubGroup(colabCmd, "schedules", "Manage execution schedules", append(crud, "pause", "resume")...)
	rootCmd.AddCommand(colabCmd)
}
