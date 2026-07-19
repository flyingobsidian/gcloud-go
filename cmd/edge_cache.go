package cmd

import "github.com/spf13/cobra"

// --- gcloud edge-cache (#333) ---
//
// Media CDN Edge Cache is served by the networkservices.googleapis.com API.
// Only the IAM-only surfaces are generated in google.golang.org/api; the CRUD
// endpoints (keysets, origins, services) are reached via the shared restClient
// against edgeCacheRest so subcommand files can call them uniformly.

var edgeCacheCmd = &cobra.Command{Use: "edge-cache", Short: "Manage Google Cloud Media CDN Edge Cache"}

var edgeCacheRest = newRESTClient("https://networkservices.googleapis.com/v1")

func init() {
	rootCmd.AddCommand(edgeCacheCmd)
}
