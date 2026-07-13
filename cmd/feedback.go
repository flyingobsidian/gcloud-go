package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// gcloud-go's issue tracker. gcloud-python points at its own; gcloud-go
// forwards feedback to the fork's repository.
const feedbackURL = "https://github.com/flyingobsidian/gcloud-go/issues/new"

var feedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Provide feedback to the gcloud-go team",
	Args:  cobra.NoArgs,
	RunE:  runFeedback,
}

var flagFeedbackLogFile string

func init() {
	feedbackCmd.Flags().StringVar(&flagFeedbackLogFile, "log-file", "",
		"Log file to attach in the browser feedback form (accepted for gcloud-python parity; not currently transmitted)")
	rootCmd.AddCommand(feedbackCmd)
}

func runFeedback(cmd *cobra.Command, args []string) error {
	fmt.Println("File a gcloud-go issue at:")
	fmt.Println("  " + feedbackURL)
	if flagFeedbackLogFile != "" {
		fmt.Printf("Log file to attach: %s\n", flagFeedbackLogFile)
	}
	// Best-effort browser open; ignore failure since not all environments
	// have a UI and gcloud-python also degrades gracefully.
	openBrowser(feedbackURL)
	return nil
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	_ = exec.Command(cmd, args...).Start()
}
