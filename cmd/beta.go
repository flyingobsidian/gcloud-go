package cmd

import (
	"github.com/spf13/cobra"
)

// --- gcloud beta (#305) ---
//
// In gcloud-python, `gcloud beta X` mirrors `gcloud X` but exposes the BETA
// tracks of each surface. gcloud-go does not track release stages
// separately: every implemented command is exposed at its GA path only.
// This stub registers the `beta` group so `gcloud beta ...` invocations are
// recognized and redirected to the GA tree.

var betaCmd = &cobra.Command{
	Use:   "beta",
	Short: "BETA aliases for the top-level command tree",
	Long: `gcloud-go does not currently split commands by release stage. Use the
corresponding top-level command instead (for example, ` + "`gcloud-go compute ...`" +
		` in place of ` + "`gcloud-go beta compute ...`" + `).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(betaCmd)
}
