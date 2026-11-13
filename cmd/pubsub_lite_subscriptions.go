package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// --- gcloud pubsub lite-subscriptions (#1173) ---

func (g *pubsubCmds) initPubsubLiteSubscriptions() {
	all := []*cobra.Command{
		g.pubsubLiteSubAckUpToCmd, g.pubsubLiteSubCreateCmd, g.pubsubLiteSubDeleteCmd,
		g.pubsubLiteSubDescribeCmd, g.pubsubLiteSubListCmd, g.pubsubLiteSubSeekCmd,
		g.pubsubLiteSubSubscribeCmd, g.pubsubLiteSubUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagPSLSubLocation, "location", "",
			"Regional or zonal location, e.g. us-central1 or us-central1-a (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagPSLSubFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.pubsubLiteSubCreateCmd, g.pubsubLiteSubSeekCmd, g.pubsubLiteSubUpdateCmd} {
		c.Flags().StringVar(&g.flagPSLSubConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.pubsubLiteSubUpdateCmd.Flags().StringVar(&g.flagPSLSubUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (default: top-level fields in --config-file)")
	g.pubsubLiteSubListCmd.Flags().Int64Var(&g.flagPSLSubPageSize, "page-size", 0, "Maximum number of results per page")
	g.pubsubLiteSubAckUpToCmd.Flags().Int64Var(&g.flagPSLSubPartition, "partition", 0,
		"Partition index for the cursor (required)")
	g.pubsubLiteSubAckUpToCmd.Flags().Int64Var(&g.flagPSLSubOffset, "offset", 0,
		"Offset to commit as the cursor for the partition (required)")
	_ = g.pubsubLiteSubAckUpToCmd.MarkFlagRequired("partition")
	_ = g.pubsubLiteSubAckUpToCmd.MarkFlagRequired("offset")

	g.pubsubLiteSubsCmd.AddCommand(all...)
	g.pubsubCmd.AddCommand(g.pubsubLiteSubsCmd)
}

func pslSubName(id, project, location string) string {
	return pubsubLiteChild("subscriptions", id, pubsubLiteLocationParent(project, location))
}

func (g *pubsubCmds) pslSubService(ctx context.Context) (*pubsublite.Service, string, error) {
	project, err := g.inv.project()
	if err != nil {
		return nil, "", err
	}
	region, err := pubsubLiteRegion(g.flagPSLSubLocation)
	if err != nil {
		return nil, "", err
	}
	svc, err := g.inv.clients().PubSubLite(ctx, region)
	if err != nil {
		return nil, "", err
	}
	return svc, project, nil
}

func (g *pubsubCmds) runPSLSubAckUpTo(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslSubService(ctx)
	if err != nil {
		return err
	}
	name := pslSubName(args[0], project, g.flagPSLSubLocation)
	req := &pubsublite.CommitCursorRequest{
		Partition: g.flagPSLSubPartition,
		Cursor:    &pubsublite.Cursor{Offset: g.flagPSLSubOffset},
	}
	got, err := svc.Cursor.Projects.Locations.Subscriptions.CommitCursor(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing cursor: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLSubFormat)
}

func (g *pubsubCmds) runPSLSubCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslSubService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Subscription{}
	if err := loadYAMLOrJSONInto(g.flagPSLSubConfigFile, body); err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Subscriptions.
		Create(pubsubLiteLocationParent(project, g.flagPSLSubLocation), body).
		SubscriptionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLSubFormat)
}

func (g *pubsubCmds) runPSLSubDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslSubService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Admin.Projects.Locations.Subscriptions.
		Delete(pslSubName(args[0], project, g.flagPSLSubLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting subscription: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted subscription [%s].\n", args[0])
	return nil
}

func (g *pubsubCmds) runPSLSubDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslSubService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Subscriptions.
		Get(pslSubName(args[0], project, g.flagPSLSubLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing subscription: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLSubFormat)
}

func (g *pubsubCmds) runPSLSubList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslSubService(ctx)
	if err != nil {
		return err
	}
	parent := pubsubLiteLocationParent(project, g.flagPSLSubLocation)
	var all []*pubsublite.Subscription
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Subscriptions.List(parent).Context(ctx)
		if g.flagPSLSubPageSize > 0 {
			call = call.PageSize(g.flagPSLSubPageSize)
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
	return emitFormattedTo(g.inv.stdout, all, g.flagPSLSubFormat)
}

func (g *pubsubCmds) runPSLSubSeek(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslSubService(ctx)
	if err != nil {
		return err
	}
	req := &pubsublite.SeekSubscriptionRequest{}
	if err := loadYAMLOrJSONInto(g.flagPSLSubConfigFile, req); err != nil {
		return err
	}
	op, err := svc.Admin.Projects.Locations.Subscriptions.
		Seek(pslSubName(args[0], project, g.flagPSLSubLocation), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("seeking subscription: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagPSLSubFormat)
}

func runPSLSubSubscribe(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("subscribe is a streaming gRPC operation not implemented over REST; use a native pub/sub client library for streaming reads")
}

func (g *pubsubCmds) runPSLSubUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslSubService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Subscription{}
	if err := loadYAMLOrJSONInto(g.flagPSLSubConfigFile, body); err != nil {
		return err
	}
	mask := g.flagPSLSubUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	got, err := svc.Admin.Projects.Locations.Subscriptions.
		Patch(pslSubName(args[0], project, g.flagPSLSubLocation), body).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating subscription: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLSubFormat)
}
