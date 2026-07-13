package cmd

import "github.com/spf13/cobra"

// --- gcloud essential-contacts (#337) ---

var essentialContactsCmd = &cobra.Command{Use: "essential-contacts", Short: "Manage essential contacts"}

func init() {
	for _, name := range []string{"compute", "create", "delete", "describe", "list", "update"} {
		registerStubCommand(essentialContactsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(essentialContactsCmd)
}
