package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	runv1 "google.golang.org/api/run/v1"
)

// --- gcloud run regions (#1052) ---
//
// Cloud Run's v2 surface does not expose a Locations.List RPC in v0.279.0
// (only Export* variants). Fall back to the v1 API, which does support
// Projects.Locations.List and enumerates the same set of regions.

func (s *runCmds) initRunRegions() {
	s.runRegionsListCmd.Flags().StringVar(&s.flagRunRegionsFormat, "format", "", "Output format")
	s.runRegionsListCmd.Flags().Int64Var(&s.flagRunRegionsPageSize, "page-size", 0, "Maximum results per page")
	s.runRegionsCmd.AddCommand(s.runRegionsListCmd)
	s.runCmd.AddCommand(s.runRegionsCmd)
}

func (s *runCmds) runRegionsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV1(ctx, "")
	if err != nil {
		return err
	}
	var all []*runv1.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if s.flagRunRegionsPageSize > 0 {
			call = call.PageSize(s.flagRunRegionsPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagRunRegionsFormat)
}
