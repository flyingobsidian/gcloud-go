package cmd

import "github.com/spf13/cobra"

// --- gcloud firestore (#341) ---

var firestoreCmd = &cobra.Command{Use: "firestore", Short: "Manage Cloud Firestore (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(firestoreCmd, "databases", "Manage databases", append(crud, "restore", "clone")...)
	registerStubGroup(firestoreCmd, "fields", "Manage field metadata", "describe", "list", "update")
	registerStubGroup(firestoreCmd, "indexes", "Manage indexes",
		"composite", "fields")
	registerStubGroup(firestoreCmd, "locations", "Manage locations", "list", "describe")
	registerStubGroup(firestoreCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(firestoreCmd, "user-creds", "Manage user credentials", "create", "delete", "describe", "list", "enable", "disable", "reset-password")
	for _, name := range []string{"bulk-delete", "export", "import"} {
		registerStubCommand(firestoreCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(firestoreCmd)
}
