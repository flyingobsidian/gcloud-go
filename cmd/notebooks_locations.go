package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks locations (#1063) ---

var notebooksLocCmd = &cobra.Command{Use: "locations", Short: "View notebook locations"}

var (
	flagNotebooksLocFormat   string
	flagNotebooksLocPageSize int64
)

var notebooksLocListCmd = &cobra.Command{
	Use: "list", Short: "List notebook locations for the current project",
	Args: cobra.NoArgs, RunE: runNotebooksLocList,
}

func init() {
	notebooksLocListCmd.Flags().StringVar(&flagNotebooksLocFormat, "format", "", "Output format")
	notebooksLocListCmd.Flags().Int64Var(&flagNotebooksLocPageSize, "page-size", 0, "Maximum results per page")
	notebooksLocCmd.AddCommand(notebooksLocListCmd)
	notebooksCmd.AddCommand(notebooksLocCmd)
}

func runNotebooksLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s", project)
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*notebooksv1.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(parent).Context(ctx)
		if flagNotebooksLocPageSize > 0 {
			call = call.PageSize(flagNotebooksLocPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNotebooksLocFormat)
}
