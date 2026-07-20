package cmd

import "github.com/spf13/cobra"

// --- gcloud firebase (#340) ---

var firebaseCmd = &cobra.Command{Use: "firebase", Short: "Work with Google Firebase"}
var firebaseTestCmd = &cobra.Command{Use: "test", Short: "Interact with Firebase Test Lab"}

func init() {
	firebaseCmd.AddCommand(firebaseTestCmd)
	rootCmd.AddCommand(firebaseCmd)
}
