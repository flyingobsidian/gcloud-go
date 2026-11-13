package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// --- gcloud pubsub lite-operations (#1171) ---

func (g *pubsubCmds) initPubsubLiteOperations() {
	for _, c := range []*cobra.Command{g.pubsubLiteOpDescribeCmd, g.pubsubLiteOpListCmd} {
		c.Flags().StringVar(&g.flagPSLOpLocation, "location", "",
			"Zonal location containing the operation, e.g. us-central1-a (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagPSLOpFormat, "format", "", "Output format")
	}
	g.pubsubLiteOpListCmd.Flags().StringVar(&g.flagPSLOpFilter, "filter", "",
		"Server-side filter expression (overrides --done/--subscription)")
	g.pubsubLiteOpListCmd.Flags().StringVar(&g.flagPSLOpSubscription, "subscription", "",
		"Filter operations to those targeting this Lite subscription")
	g.pubsubLiteOpListCmd.Flags().BoolVar(&g.flagPSLOpDone, "done", false,
		"Filter operations by whether they are done")
	g.pubsubLiteOpListCmd.Flags().Int64Var(&g.flagPSLOpPageSize, "page-size", 0,
		"Maximum number of results per page")
	g.pubsubLiteOpListCmd.PreRun = func(cmd *cobra.Command, args []string) {
		g.flagPSLOpDoneSet = cmd.Flags().Changed("done")
	}

	g.pubsubLiteOperationsCmd.AddCommand(g.pubsubLiteOpDescribeCmd, g.pubsubLiteOpListCmd)
	g.pubsubCmd.AddCommand(g.pubsubLiteOperationsCmd)
}

func pslOpName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/operations/%s", pubsubLiteLocationParent(project, location), id)
}

// pslOpBuildFilter derives the server-side filter from --done and
// --subscription. This mirrors gcloud-python's UpdateListOperationsFilter hook,
// which composes an AND-joined filter unless the user supplied an explicit
// --filter.
func (g *pubsubCmds) pslOpBuildFilter(project string) string {
	if g.flagPSLOpFilter != "" {
		return g.flagPSLOpFilter
	}
	var parts []string
	if g.flagPSLOpDoneSet {
		parts = append(parts, fmt.Sprintf("done=%t", g.flagPSLOpDone))
	}
	if g.flagPSLOpSubscription != "" {
		sub := pubsubLiteChild("subscriptions", g.flagPSLOpSubscription,
			pubsubLiteLocationParent(project, g.flagPSLOpLocation))
		parts = append(parts, fmt.Sprintf("metadata.target=%s", sub))
	}
	return strings.Join(parts, " AND ")
}

func (g *pubsubCmds) runPSLOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	region, err := pubsubLiteRegion(g.flagPSLOpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSubLite(ctx, region)
	if err != nil {
		return err
	}
	op, err := svc.Admin.Projects.Locations.Operations.Get(pslOpName(args[0], project, g.flagPSLOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagPSLOpFormat)
}

func (g *pubsubCmds) runPSLOpList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	region, err := pubsubLiteRegion(g.flagPSLOpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSubLite(ctx, region)
	if err != nil {
		return err
	}
	filter := g.pslOpBuildFilter(project)
	parent := pubsubLiteLocationParent(project, g.flagPSLOpLocation)
	var all []*pubsublite.Operation
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Operations.List(parent).Context(ctx)
		if filter != "" {
			call = call.Filter(filter)
		}
		if g.flagPSLOpPageSize > 0 {
			call = call.PageSize(g.flagPSLOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagPSLOpFormat)
}
