package cmd

import "github.com/spf13/cobra"

// --- gcloud telco-automation (#390) ---
//
// The Telco Automation API (`telcoautomation.googleapis.com`, v1) is not
// exposed by google.golang.org/api; subgroups use the shared restClient from
// rest_helpers.go with a per-service endpoint.

var telcoAutomationCmd = &cobra.Command{Use: "telco-automation", Short: "Manage Telco Automation"}

var telcoAutomationRest = newRESTClient("https://telcoautomation.googleapis.com/v1")

func init() {
	rootCmd.AddCommand(telcoAutomationCmd)
}
