package cmd

import "github.com/spf13/cobra"

// --- gcloud alpha (#294) ---
//
// Like `beta` and `preview`, `alpha` in gcloud-python is a release-stage alias
// tree. gcloud-go does not track release stages separately.

var alphaCmd = &cobra.Command{
	Use:   "alpha",
	Short: "ALPHA aliases for the top-level command tree",
	Long:  "gcloud-go does not track release stages separately. Use the corresponding top-level command instead.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(alphaCmd)
}
