package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// --- gcloud workspace-add-ons (#399) ---
//
// The Google Workspace Add-ons API (`gsuiteaddons.googleapis.com`, v1) is not
// exposed by google.golang.org/api; subgroups use the shared restClient from
// rest_helpers.go with a per-service endpoint.

var workspaceAddOnsCmd = &cobra.Command{Use: "workspace-add-ons", Short: "Manage Google Workspace Add-ons"}

var workspaceAddOnsRest = newRESTClient("https://gsuiteaddons.googleapis.com/v1")

var (
	flagWAOGetAuthFormat string
)

var workspaceAddOnsGetAuthCmd = &cobra.Command{
	Use: "get-authorization", Short: "Get the authorization information for deployments in a project",
	Args: cobra.NoArgs, RunE: runWAOGetAuthorization,
}

func init() {
	workspaceAddOnsGetAuthCmd.Flags().StringVar(&flagWAOGetAuthFormat, "format", "", "Output format")
	workspaceAddOnsCmd.AddCommand(workspaceAddOnsGetAuthCmd)
	rootCmd.AddCommand(workspaceAddOnsCmd)
}

func runWAOGetAuthorization(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := workspaceAddOnsRest.do(ctx, http.MethodGet, fmt.Sprintf("/projects/%s/authorization", project), nil, nil, &got); err != nil {
		return fmt.Errorf("getting workspace add-ons authorization: %w", err)
	}
	return emitFormatted(got, flagWAOGetAuthFormat)
}
