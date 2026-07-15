package cmd

import "github.com/spf13/cobra"

// --- gcloud recaptcha (#378) ---

var recaptchaCmd = &cobra.Command{Use: "recaptcha", Short: "Manage reCAPTCHA"}

func init() {
	rootCmd.AddCommand(recaptchaCmd)
}
