package cmd

import "github.com/spf13/cobra"

// --- gcloud privateca (#374) ---

var privatecaCmd = &cobra.Command{Use: "privateca", Short: "Manage Private CA"}

func init() {
	// All subgroups (certificates, locations, operations, pools, roots,
	// subordinates, templates) are implemented in dedicated privateca_*.go
	// files.
	rootCmd.AddCommand(privatecaCmd)
}
