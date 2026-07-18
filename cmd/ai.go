package cmd

import "github.com/spf13/cobra"

// --- gcloud ai (#291) ---
//
// Vertex AI is served by the aiplatform API surface. Every subgroup uses a
// regional aiplatform client (endpoint of the form
// https://<region>-aiplatform.googleapis.com/) and takes --region on every
// leaf command.

var aiCmd = &cobra.Command{Use: "ai", Short: "Manage Vertex AI"}

func init() {
	rootCmd.AddCommand(aiCmd)
}
