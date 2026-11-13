package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// --- gcloud pubsub lite-topics (#1174) ---

func (g *pubsubCmds) initPubsubLiteTopics() {
	all := []*cobra.Command{
		g.pubsubLiteTopicCreateCmd, g.pubsubLiteTopicDeleteCmd, g.pubsubLiteTopicDescribeCmd,
		g.pubsubLiteTopicListCmd, g.pubsubLiteTopicListSubsCmd, g.pubsubLiteTopicPublishCmd,
		g.pubsubLiteTopicUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagPSLTopicLocation, "location", "",
			"Regional or zonal location, e.g. us-central1 or us-central1-a (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagPSLTopicFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.pubsubLiteTopicCreateCmd, g.pubsubLiteTopicUpdateCmd} {
		c.Flags().StringVar(&g.flagPSLTopicConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Topic body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.pubsubLiteTopicUpdateCmd.Flags().StringVar(&g.flagPSLTopicUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (default: top-level fields in --config-file)")
	for _, c := range []*cobra.Command{g.pubsubLiteTopicListCmd, g.pubsubLiteTopicListSubsCmd} {
		c.Flags().Int64Var(&g.flagPSLTopicPageSize, "page-size", 0, "Maximum number of results per page")
	}

	g.pubsubLiteTopicsCmd.AddCommand(all...)
	g.pubsubCmd.AddCommand(g.pubsubLiteTopicsCmd)
}

func pslTopicName(id, project, location string) string {
	return pubsubLiteChild("topics", id, pubsubLiteLocationParent(project, location))
}

func (g *pubsubCmds) pslTopicService(ctx context.Context) (*pubsublite.Service, string, error) {
	project, err := g.inv.project()
	if err != nil {
		return nil, "", err
	}
	region, err := pubsubLiteRegion(g.flagPSLTopicLocation)
	if err != nil {
		return nil, "", err
	}
	svc, err := g.inv.clients().PubSubLite(ctx, region)
	if err != nil {
		return nil, "", err
	}
	return svc, project, nil
}

func (g *pubsubCmds) runPSLTopicCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslTopicService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Topic{}
	if err := loadYAMLOrJSONInto(g.flagPSLTopicConfigFile, body); err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Topics.
		Create(pubsubLiteLocationParent(project, g.flagPSLTopicLocation), body).
		TopicId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating topic: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLTopicFormat)
}

func (g *pubsubCmds) runPSLTopicDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslTopicService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Admin.Projects.Locations.Topics.
		Delete(pslTopicName(args[0], project, g.flagPSLTopicLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting topic: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted topic [%s].\n", args[0])
	return nil
}

func (g *pubsubCmds) runPSLTopicDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslTopicService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Topics.
		Get(pslTopicName(args[0], project, g.flagPSLTopicLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing topic: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLTopicFormat)
}

func (g *pubsubCmds) runPSLTopicList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslTopicService(ctx)
	if err != nil {
		return err
	}
	parent := pubsubLiteLocationParent(project, g.flagPSLTopicLocation)
	var all []*pubsublite.Topic
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Topics.List(parent).Context(ctx)
		if g.flagPSLTopicPageSize > 0 {
			call = call.PageSize(g.flagPSLTopicPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing topics: %w", err)
		}
		all = append(all, resp.Topics...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagPSLTopicFormat)
}

func (g *pubsubCmds) runPSLTopicListSubs(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslTopicService(ctx)
	if err != nil {
		return err
	}
	name := pslTopicName(args[0], project, g.flagPSLTopicLocation)
	var all []string
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Topics.Subscriptions.List(name).Context(ctx)
		if g.flagPSLTopicPageSize > 0 {
			call = call.PageSize(g.flagPSLTopicPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing topic subscriptions: %w", err)
		}
		all = append(all, resp.Subscriptions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagPSLTopicFormat)
}

func runPSLTopicPublish(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("publish is a streaming gRPC operation not implemented over REST; use a native pub/sub client library for publishing")
}

func (g *pubsubCmds) runPSLTopicUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslTopicService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Topic{}
	if err := loadYAMLOrJSONInto(g.flagPSLTopicConfigFile, body); err != nil {
		return err
	}
	mask := g.flagPSLTopicUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	got, err := svc.Admin.Projects.Locations.Topics.
		Patch(pslTopicName(args[0], project, g.flagPSLTopicLocation), body).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating topic: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLTopicFormat)
}
