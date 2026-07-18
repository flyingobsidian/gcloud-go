package cmd

import "github.com/spf13/cobra"

// --- gcloud memorystore (#355) ---
//
// The memorystore v1beta API is not exposed through google.golang.org/api, so
// this file wires the top-level command up and hands its subcommands a shared
// REST client that talks directly to the memorystore endpoint.

var memorystoreCmd = &cobra.Command{Use: "memorystore", Short: "Manage Memorystore"}

var memorystoreRest = newRESTClient("https://memorystore.googleapis.com/v1beta")

func init() {
	rootCmd.AddCommand(memorystoreCmd)
}
