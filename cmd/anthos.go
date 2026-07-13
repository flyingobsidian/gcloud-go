package cmd

import "github.com/spf13/cobra"

// --- gcloud anthos (#295) ---

var anthosCmd = &cobra.Command{Use: "anthos", Short: "Anthos commands"}

func init() {
	registerStubGroup(anthosCmd, "auth", "Authenticate clusters", "login", "token")
	registerStubGroup(anthosCmd, "config", "Anthos config", "controller")
	registerStubCommand(anthosCmd, "create-login-config", "Generate a login configuration file")
	rootCmd.AddCommand(anthosCmd)
}
