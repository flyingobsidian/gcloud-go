package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// --- gcloud pubsub lite-operations (#1171) ---

var pubsubLiteOperationsCmd = &cobra.Command{
	Use:   "lite-operations",
	Short: "Manage Pub/Sub Lite operations",
}

var (
	flagPSLOpLocation     string
	flagPSLOpFormat       string
	flagPSLOpFilter       string
	flagPSLOpSubscription string
	flagPSLOpPageSize     int64
)

var (
	flagPSLOpDone         bool
	flagPSLOpDoneSet      bool
)

var (
	pubsubLiteOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Pub/Sub Lite operation",
		Args: cobra.ExactArgs(1), RunE: runPSLOpDescribe,
	}
	pubsubLiteOpListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub Lite operations in a location",
		Args: cobra.NoArgs, RunE: runPSLOpList,
	}
)

func init() {
	for _, c := range []*cobra.Command{pubsubLiteOpDescribeCmd, pubsubLiteOpListCmd} {
		c.Flags().StringVar(&flagPSLOpLocation, "location", "",
			"Zonal location containing the operation, e.g. us-central1-a (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagPSLOpFormat, "format", "", "Output format")
	}
	pubsubLiteOpListCmd.Flags().StringVar(&flagPSLOpFilter, "filter", "",
		"Server-side filter expression (overrides --done/--subscription)")
	pubsubLiteOpListCmd.Flags().StringVar(&flagPSLOpSubscription, "subscription", "",
		"Filter operations to those targeting this Lite subscription")
	pubsubLiteOpListCmd.Flags().BoolVar(&flagPSLOpDone, "done", false,
		"Filter operations by whether they are done")
	pubsubLiteOpListCmd.Flags().Int64Var(&flagPSLOpPageSize, "page-size", 0,
		"Maximum number of results per page")
	pubsubLiteOpListCmd.PreRun = func(cmd *cobra.Command, args []string) {
		flagPSLOpDoneSet = cmd.Flags().Changed("done")
	}

	pubsubLiteOperationsCmd.AddCommand(pubsubLiteOpDescribeCmd, pubsubLiteOpListCmd)
	pubsubCmd.AddCommand(pubsubLiteOperationsCmd)
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
func pslOpBuildFilter(project string) string {
	if flagPSLOpFilter != "" {
		return flagPSLOpFilter
	}
	var parts []string
	if flagPSLOpDoneSet {
		parts = append(parts, fmt.Sprintf("done=%t", flagPSLOpDone))
	}
	if flagPSLOpSubscription != "" {
		sub := pubsubLiteChild("subscriptions", flagPSLOpSubscription,
			pubsubLiteLocationParent(project, flagPSLOpLocation))
		parts = append(parts, fmt.Sprintf("metadata.target=%s", sub))
	}
	return strings.Join(parts, " AND ")
}

func runPSLOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	region, err := pubsubLiteRegion(flagPSLOpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubLiteService(ctx, flagAccount, region)
	if err != nil {
		return err
	}
	op, err := svc.Admin.Projects.Locations.Operations.Get(pslOpName(args[0], project, flagPSLOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagPSLOpFormat)
}

func runPSLOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	region, err := pubsubLiteRegion(flagPSLOpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubLiteService(ctx, flagAccount, region)
	if err != nil {
		return err
	}
	filter := pslOpBuildFilter(project)
	parent := pubsubLiteLocationParent(project, flagPSLOpLocation)
	var all []*pubsublite.Operation
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Operations.List(parent).Context(ctx)
		if filter != "" {
			call = call.Filter(filter)
		}
		if flagPSLOpPageSize > 0 {
			call = call.PageSize(flagPSLOpPageSize)
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
	return emitFormatted(all, flagPSLOpFormat)
}
