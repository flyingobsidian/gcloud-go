package cmd

import "github.com/spf13/cobra"

// --- gcloud recaptcha (#378) ---

var recaptchaCmd = &cobra.Command{Use: "recaptcha", Short: "Manage reCAPTCHA (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(recaptchaCmd, "firewall-policies", "Manage firewall policies", crud...)
	registerStubGroup(recaptchaCmd, "keys", "Manage reCAPTCHA keys", append(crud, "migrate", "get-metrics")...)
	rootCmd.AddCommand(recaptchaCmd)
}
