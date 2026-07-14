package cmd

import "github.com/spf13/cobra"

// --- gcloud firestore (#341) ---

var firestoreCmd = &cobra.Command{Use: "firestore", Short: "Manage Cloud Firestore"}

func init() {
	// Subgroups (databases, backups, fields, indexes, locations, operations,
	// user-creds) are implemented in dedicated firestore_*.go files. The three
	// top-level bulk-delete / export / import commands remain stubs pending
	// their own implementations.
	for _, name := range []string{"bulk-delete", "export", "import"} {
		registerStubCommand(firestoreCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(firestoreCmd)
}
