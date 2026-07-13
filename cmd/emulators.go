package cmd

import "github.com/spf13/cobra"

// --- gcloud emulators (#335) ---

var emulatorsCmd = &cobra.Command{Use: "emulators", Short: "Local emulators"}

func init() {
	registerStubGroup(emulatorsCmd, "firestore", "Manage local Firestore emulator", "start", "env-init")
	registerStubGroup(emulatorsCmd, "spanner", "Manage local Spanner emulator", "start", "env-init")
	registerStubGroup(emulatorsCmd, "bigtable", "Manage local Bigtable emulator", "start", "env-init")
	registerStubGroup(emulatorsCmd, "datastore", "Manage local Datastore emulator", "start", "env-init", "env-unset")
	registerStubGroup(emulatorsCmd, "pubsub", "Manage local Pub/Sub emulator", "start", "env-init")
	rootCmd.AddCommand(emulatorsCmd)
}
