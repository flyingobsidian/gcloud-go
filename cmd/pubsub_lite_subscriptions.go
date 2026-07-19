package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// --- gcloud pubsub lite-subscriptions (#1173) ---

var pubsubLiteSubsCmd = &cobra.Command{
	Use:   "lite-subscriptions",
	Short: "Manage Pub/Sub Lite subscriptions",
}

var (
	flagPSLSubLocation   string
	flagPSLSubFormat     string
	flagPSLSubConfigFile string
	flagPSLSubUpdateMask string
	flagPSLSubPageSize   int64
	flagPSLSubPartition  int64
	flagPSLSubOffset     int64
)

var (
	pubsubLiteSubAckUpToCmd = &cobra.Command{
		Use: "ack-up-to SUBSCRIPTION", Short: "Commit the cursor for a subscription partition",
		Args: cobra.ExactArgs(1), RunE: runPSLSubAckUpTo,
	}
	pubsubLiteSubCreateCmd = &cobra.Command{
		Use: "create SUBSCRIPTION", Short: "Create a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: runPSLSubCreate,
	}
	pubsubLiteSubDeleteCmd = &cobra.Command{
		Use: "delete SUBSCRIPTION", Short: "Delete a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: runPSLSubDelete,
	}
	pubsubLiteSubDescribeCmd = &cobra.Command{
		Use: "describe SUBSCRIPTION", Short: "Describe a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: runPSLSubDescribe,
	}
	pubsubLiteSubListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub Lite subscriptions in a location",
		Args: cobra.NoArgs, RunE: runPSLSubList,
	}
	pubsubLiteSubSeekCmd = &cobra.Command{
		Use: "seek SUBSCRIPTION", Short: "Seek a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: runPSLSubSeek,
	}
	pubsubLiteSubSubscribeCmd = &cobra.Command{
		Use: "subscribe SUBSCRIPTION", Short: "Subscribe to a Pub/Sub Lite subscription (unsupported)",
		Args: cobra.ExactArgs(1), RunE: runPSLSubSubscribe,
	}
	pubsubLiteSubUpdateCmd = &cobra.Command{
		Use: "update SUBSCRIPTION", Short: "Update a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: runPSLSubUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		pubsubLiteSubAckUpToCmd, pubsubLiteSubCreateCmd, pubsubLiteSubDeleteCmd,
		pubsubLiteSubDescribeCmd, pubsubLiteSubListCmd, pubsubLiteSubSeekCmd,
		pubsubLiteSubSubscribeCmd, pubsubLiteSubUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagPSLSubLocation, "location", "",
			"Regional or zonal location, e.g. us-central1 or us-central1-a (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagPSLSubFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{pubsubLiteSubCreateCmd, pubsubLiteSubSeekCmd, pubsubLiteSubUpdateCmd} {
		c.Flags().StringVar(&flagPSLSubConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	pubsubLiteSubUpdateCmd.Flags().StringVar(&flagPSLSubUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (default: top-level fields in --config-file)")
	pubsubLiteSubListCmd.Flags().Int64Var(&flagPSLSubPageSize, "page-size", 0, "Maximum number of results per page")
	pubsubLiteSubAckUpToCmd.Flags().Int64Var(&flagPSLSubPartition, "partition", 0,
		"Partition index for the cursor (required)")
	pubsubLiteSubAckUpToCmd.Flags().Int64Var(&flagPSLSubOffset, "offset", 0,
		"Offset to commit as the cursor for the partition (required)")
	_ = pubsubLiteSubAckUpToCmd.MarkFlagRequired("partition")
	_ = pubsubLiteSubAckUpToCmd.MarkFlagRequired("offset")

	pubsubLiteSubsCmd.AddCommand(all...)
	pubsubCmd.AddCommand(pubsubLiteSubsCmd)
}

func pslSubName(id, project, location string) string {
	return pubsubLiteChild("subscriptions", id, pubsubLiteLocationParent(project, location))
}

func pslSubService(ctx context.Context) (*pubsublite.Service, string, error) {
	project, err := resolveProject()
	if err != nil {
		return nil, "", err
	}
	region, err := pubsubLiteRegion(flagPSLSubLocation)
	if err != nil {
		return nil, "", err
	}
	svc, err := gcp.PubSubLiteService(ctx, flagAccount, region)
	if err != nil {
		return nil, "", err
	}
	return svc, project, nil
}

func runPSLSubAckUpTo(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslSubService(ctx)
	if err != nil {
		return err
	}
	name := pslSubName(args[0], project, flagPSLSubLocation)
	req := &pubsublite.CommitCursorRequest{
		Partition: flagPSLSubPartition,
		Cursor:    &pubsublite.Cursor{Offset: flagPSLSubOffset},
	}
	got, err := svc.Cursor.Projects.Locations.Subscriptions.CommitCursor(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing cursor: %w", err)
	}
	return emitFormatted(got, flagPSLSubFormat)
}

func runPSLSubCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslSubService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Subscription{}
	if err := loadYAMLOrJSONInto(flagPSLSubConfigFile, body); err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Subscriptions.
		Create(pubsubLiteLocationParent(project, flagPSLSubLocation), body).
		SubscriptionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}
	return emitFormatted(got, flagPSLSubFormat)
}

func runPSLSubDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslSubService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Admin.Projects.Locations.Subscriptions.
		Delete(pslSubName(args[0], project, flagPSLSubLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting subscription: %w", err)
	}
	fmt.Printf("Deleted subscription [%s].\n", args[0])
	return nil
}

func runPSLSubDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslSubService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Subscriptions.
		Get(pslSubName(args[0], project, flagPSLSubLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing subscription: %w", err)
	}
	return emitFormatted(got, flagPSLSubFormat)
}

func runPSLSubList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslSubService(ctx)
	if err != nil {
		return err
	}
	parent := pubsubLiteLocationParent(project, flagPSLSubLocation)
	var all []*pubsublite.Subscription
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Subscriptions.List(parent).Context(ctx)
		if flagPSLSubPageSize > 0 {
			call = call.PageSize(flagPSLSubPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing subscriptions: %w", err)
		}
		all = append(all, resp.Subscriptions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagPSLSubFormat)
}

func runPSLSubSeek(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslSubService(ctx)
	if err != nil {
		return err
	}
	req := &pubsublite.SeekSubscriptionRequest{}
	if err := loadYAMLOrJSONInto(flagPSLSubConfigFile, req); err != nil {
		return err
	}
	op, err := svc.Admin.Projects.Locations.Subscriptions.
		Seek(pslSubName(args[0], project, flagPSLSubLocation), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("seeking subscription: %w", err)
	}
	return emitFormatted(op, flagPSLSubFormat)
}

func runPSLSubSubscribe(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("subscribe is a streaming gRPC operation not implemented over REST; use a native pub/sub client library for streaming reads")
}

func runPSLSubUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslSubService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Subscription{}
	if err := loadYAMLOrJSONInto(flagPSLSubConfigFile, body); err != nil {
		return err
	}
	mask := flagPSLSubUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	got, err := svc.Admin.Projects.Locations.Subscriptions.
		Patch(pslSubName(args[0], project, flagPSLSubLocation), body).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating subscription: %w", err)
	}
	return emitFormatted(got, flagPSLSubFormat)
}
