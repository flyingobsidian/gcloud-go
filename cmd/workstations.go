package cmd

import "github.com/spf13/cobra"

// --- gcloud workstations (#400) ---

var workstationsCmd = &cobra.Command{Use: "workstations", Short: "Manage Cloud Workstations (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(workstationsCmd, "clusters", "Manage workstation clusters", crud...)
	registerStubGroup(workstationsCmd, "configs", "Manage workstation configurations", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	for _, name := range []string{
		"create", "delete", "describe", "get-iam-policy", "list", "list-usable",
		"set-iam-policy", "ssh", "start", "start-tcp-tunnel", "stop", "update",
	} {
		registerStubCommand(workstationsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(workstationsCmd)
}
