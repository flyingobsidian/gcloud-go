package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run revisions (#1053) ---
//
// Revisions live under a Cloud Run service; every subcommand requires
// --service and --region.

func (s *runCmds) initRunRevisions() {
	all := []*cobra.Command{s.runRevisionsDeleteCmd, s.runRevisionsDescribeCmd, s.runRevisionsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagRunRevisionsRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&s.flagRunRevisionsService, "service", "", "Parent Cloud Run service (required)")
		c.Flags().StringVar(&s.flagRunRevisionsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("service")
	}
	s.runRevisionsListCmd.Flags().Int64Var(&s.flagRunRevisionsPageSize, "page-size", 0, "Maximum results per page")
	s.runRevisionsListCmd.Flags().BoolVar(&s.flagRunRevisionsShowDel, "show-deleted", false, "Include deleted revisions")

	s.runRevisionsCmd.AddCommand(all...)
	s.runCmd.AddCommand(s.runRevisionsCmd)
}

// runRevisionsParent returns "projects/PROJ/locations/REGION/services/SERVICE".
func (s *runCmds) runRevisionsParent(project string) string {
	return runResourceName(project, s.flagRunRevisionsRegion, "services", s.flagRunRevisionsService)
}

// runRevisionsName joins the parent service with the revision id, passing
// through fully-qualified names unchanged.
func (s *runCmds) runRevisionsName(project, rev string) string {
	if hasProjectsPrefix(rev) {
		return rev
	}
	return fmt.Sprintf("%s/revisions/%s", s.runRevisionsParent(project), rev)
}

func (s *runCmds) runRevisionsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunRevisionsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Revisions.Delete(s.runRevisionsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting revision: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for revision [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunRevisionsFormat)
}

func (s *runCmds) runRevisionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunRevisionsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Revisions.Get(s.runRevisionsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing revision: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRunRevisionsFormat)
}

func (s *runCmds) runRevisionsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunRevisionsRegion)
	if err != nil {
		return err
	}
	var all []*runv2.GoogleCloudRunV2Revision
	pageToken := ""
	for {
		call := svc.Projects.Locations.Services.Revisions.List(s.runRevisionsParent(project)).Context(ctx)
		if s.flagRunRevisionsPageSize > 0 {
			call = call.PageSize(s.flagRunRevisionsPageSize)
		}
		if s.flagRunRevisionsShowDel {
			call = call.ShowDeleted(true)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing revisions: %w", err)
		}
		all = append(all, resp.Revisions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagRunRevisionsFormat)
}

// hasProjectsPrefix reports whether name is already a fully-qualified Cloud
// Run resource path ("projects/..."). It is used by helpers that accept both
// bare ids and full names.
func hasProjectsPrefix(name string) bool {
	const p = "projects/"
	return len(name) >= len(p) && name[:len(p)] == p
}
