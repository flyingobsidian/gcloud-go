package cmd

import "github.com/spf13/cobra"

// --- gcloud cloud-shell (#314) ---

var cloudShellCmd = &cobra.Command{Use: "cloud-shell", Short: "Manage Cloud Shell"}

func init() {
	for _, name := range []string{"get-mount-command", "scp", "ssh"} {
		registerStubCommand(cloudShellCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(cloudShellCmd)
}
