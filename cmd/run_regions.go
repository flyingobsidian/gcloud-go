package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	runv1 "google.golang.org/api/run/v1"
)

// --- gcloud run regions (#1052) ---
//
// Cloud Run's v2 surface does not expose a Locations.List RPC in v0.279.0
// (only Export* variants). Fall back to the v1 API, which does support
// Projects.Locations.List and enumerates the same set of regions.

var runRegionsCmd = &cobra.Command{Use: "regions", Short: "View Cloud Run regions"}

var (
	flagRunRegionsFormat   string
	flagRunRegionsPageSize int64
)

var runRegionsListCmd = &cobra.Command{
	Use: "list", Short: "List the regions where Cloud Run is available",
	Args: cobra.NoArgs, RunE: runRegionsList,
}

func init() {
	runRegionsListCmd.Flags().StringVar(&flagRunRegionsFormat, "format", "", "Output format")
	runRegionsListCmd.Flags().Int64Var(&flagRunRegionsPageSize, "page-size", 0, "Maximum results per page")
	runRegionsCmd.AddCommand(runRegionsListCmd)
	runCmd.AddCommand(runRegionsCmd)
}

func runRegionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV1Service(ctx, flagAccount, "")
	if err != nil {
		return err
	}
	var all []*runv1.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagRunRegionsPageSize > 0 {
			call = call.PageSize(flagRunRegionsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing regions: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagRunRegionsFormat)
}
