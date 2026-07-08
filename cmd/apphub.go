package cmd

import "github.com/spf13/cobra"

// --- gcloud apphub (#300) ---

var apphubCmd = &cobra.Command{Use: "apphub", Short: "Manage App Hub (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(apphubCmd, "applications", "Manage applications", append(crud, "services", "workloads")...)
	registerStubGroup(apphubCmd, "boundary", "Manage boundaries", "describe", "update")
	registerStubGroup(apphubCmd, "discovered-services", "Manage discovered services", "describe", "list", "lookup", "find-unregistered")
	registerStubGroup(apphubCmd, "discovered-workloads", "Manage discovered workloads", "describe", "list", "lookup", "find-unregistered")
	registerStubGroup(apphubCmd, "locations", "Manage locations", "describe", "list")
	registerStubGroup(apphubCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(apphubCmd, "service-projects", "Manage service projects", "add", "remove", "describe", "list", "lookup")
	rootCmd.AddCommand(apphubCmd)
}
