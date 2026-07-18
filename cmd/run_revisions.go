package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run revisions (#1053) ---
//
// Revisions live under a Cloud Run service; every subcommand requires
// --service and --region.

var runRevisionsCmd = &cobra.Command{Use: "revisions", Short: "Manage Cloud Run revisions"}

var (
	flagRunRevisionsRegion   string
	flagRunRevisionsService  string
	flagRunRevisionsFormat   string
	flagRunRevisionsPageSize int64
	flagRunRevisionsShowDel  bool
)

var (
	runRevisionsDeleteCmd = &cobra.Command{
		Use: "delete REVISION", Short: "Delete a Cloud Run revision",
		Args: cobra.ExactArgs(1), RunE: runRevisionsDelete,
	}
	runRevisionsDescribeCmd = &cobra.Command{
		Use: "describe REVISION", Short: "Describe a Cloud Run revision",
		Args: cobra.ExactArgs(1), RunE: runRevisionsDescribe,
	}
	runRevisionsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run revisions for a service",
		Args: cobra.NoArgs, RunE: runRevisionsList,
	}
)

func init() {
	all := []*cobra.Command{runRevisionsDeleteCmd, runRevisionsDescribeCmd, runRevisionsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagRunRevisionsRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&flagRunRevisionsService, "service", "", "Parent Cloud Run service (required)")
		c.Flags().StringVar(&flagRunRevisionsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("service")
	}
	runRevisionsListCmd.Flags().Int64Var(&flagRunRevisionsPageSize, "page-size", 0, "Maximum results per page")
	runRevisionsListCmd.Flags().BoolVar(&flagRunRevisionsShowDel, "show-deleted", false, "Include deleted revisions")

	runRevisionsCmd.AddCommand(all...)
	runCmd.AddCommand(runRevisionsCmd)
}

// runRevisionsParent returns "projects/PROJ/locations/REGION/services/SERVICE".
func runRevisionsParent(project string) string {
	return runResourceName(project, flagRunRevisionsRegion, "services", flagRunRevisionsService)
}

// runRevisionsName joins the parent service with the revision id, passing
// through fully-qualified names unchanged.
func runRevisionsName(project, rev string) string {
	if hasProjectsPrefix(rev) {
		return rev
	}
	return fmt.Sprintf("%s/revisions/%s", runRevisionsParent(project), rev)
}

func runRevisionsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunRevisionsRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Revisions.Delete(runRevisionsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting revision: %w", err)
	}
	fmt.Printf("Delete request issued for revision [%s].\n", args[0])
	return emitFormatted(op, flagRunRevisionsFormat)
}

func runRevisionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunRevisionsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Revisions.Get(runRevisionsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing revision: %w", err)
	}
	return emitFormatted(got, flagRunRevisionsFormat)
}

func runRevisionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunRevisionsRegion)
	if err != nil {
		return err
	}
	var all []*runv2.GoogleCloudRunV2Revision
	pageToken := ""
	for {
		call := svc.Projects.Locations.Services.Revisions.List(runRevisionsParent(project)).Context(ctx)
		if flagRunRevisionsPageSize > 0 {
			call = call.PageSize(flagRunRevisionsPageSize)
		}
		if flagRunRevisionsShowDel {
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
	return emitFormatted(all, flagRunRevisionsFormat)
}

// hasProjectsPrefix reports whether name is already a fully-qualified Cloud
// Run resource path ("projects/..."). It is used by helpers that accept both
// bare ids and full names.
func hasProjectsPrefix(name string) bool {
	const p = "projects/"
	return len(name) >= len(p) && name[:len(p)] == p
}
