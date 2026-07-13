package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// gcloud-go does not run a satisfaction survey of its own; the command is
// registered for CLI parity with gcloud-python (#535) and points at the
// feedback tracker where users can share input.
var surveyCmd = &cobra.Command{
	Use:   "survey",
	Short: "Invoke the customer satisfaction survey",
	Args:  cobra.NoArgs,
	RunE:  runSurvey,
}

func init() {
	rootCmd.AddCommand(surveyCmd)
}

func runSurvey(cmd *cobra.Command, args []string) error {
	fmt.Println("gcloud-go does not currently run a survey.")
	fmt.Println("Please share feedback via:")
	fmt.Println("  " + feedbackURL)
	return nil
}
