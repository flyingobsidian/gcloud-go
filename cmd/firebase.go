package cmd

import "github.com/spf13/cobra"

// --- gcloud firebase (#340) ---

var firebaseCmd = &cobra.Command{Use: "firebase", Short: "Work with Google Firebase"}

func init() {
	test := &cobra.Command{Use: "test", Short: "Interact with Firebase Test Lab"}
	registerStubGroup(test, "android", "Android testing", "run", "models", "versions", "locales", "orientations")
	registerStubGroup(test, "ios", "iOS testing", "run", "models", "versions", "locales", "orientations")
	registerStubGroup(test, "network-profiles", "Manage network profiles", "list", "describe")
	firebaseCmd.AddCommand(test)
	rootCmd.AddCommand(firebaseCmd)
}
