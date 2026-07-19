package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// --- gcloud pubsub lite-topics (#1174) ---

var pubsubLiteTopicsCmd = &cobra.Command{
	Use:   "lite-topics",
	Short: "Manage Pub/Sub Lite topics",
}

var (
	flagPSLTopicLocation   string
	flagPSLTopicFormat     string
	flagPSLTopicConfigFile string
	flagPSLTopicUpdateMask string
	flagPSLTopicPageSize   int64
)

var (
	pubsubLiteTopicCreateCmd = &cobra.Command{
		Use: "create TOPIC", Short: "Create a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: runPSLTopicCreate,
	}
	pubsubLiteTopicDeleteCmd = &cobra.Command{
		Use: "delete TOPIC", Short: "Delete a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: runPSLTopicDelete,
	}
	pubsubLiteTopicDescribeCmd = &cobra.Command{
		Use: "describe TOPIC", Short: "Describe a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: runPSLTopicDescribe,
	}
	pubsubLiteTopicListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub Lite topics in a location",
		Args: cobra.NoArgs, RunE: runPSLTopicList,
	}
	pubsubLiteTopicListSubsCmd = &cobra.Command{
		Use: "list-subscriptions TOPIC", Short: "List subscriptions attached to a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: runPSLTopicListSubs,
	}
	pubsubLiteTopicPublishCmd = &cobra.Command{
		Use: "publish TOPIC", Short: "Publish to a Pub/Sub Lite topic (unsupported)",
		Args: cobra.ExactArgs(1), RunE: runPSLTopicPublish,
	}
	pubsubLiteTopicUpdateCmd = &cobra.Command{
		Use: "update TOPIC", Short: "Update a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: runPSLTopicUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		pubsubLiteTopicCreateCmd, pubsubLiteTopicDeleteCmd, pubsubLiteTopicDescribeCmd,
		pubsubLiteTopicListCmd, pubsubLiteTopicListSubsCmd, pubsubLiteTopicPublishCmd,
		pubsubLiteTopicUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagPSLTopicLocation, "location", "",
			"Regional or zonal location, e.g. us-central1 or us-central1-a (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagPSLTopicFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{pubsubLiteTopicCreateCmd, pubsubLiteTopicUpdateCmd} {
		c.Flags().StringVar(&flagPSLTopicConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Topic body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	pubsubLiteTopicUpdateCmd.Flags().StringVar(&flagPSLTopicUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (default: top-level fields in --config-file)")
	for _, c := range []*cobra.Command{pubsubLiteTopicListCmd, pubsubLiteTopicListSubsCmd} {
		c.Flags().Int64Var(&flagPSLTopicPageSize, "page-size", 0, "Maximum number of results per page")
	}

	pubsubLiteTopicsCmd.AddCommand(all...)
	pubsubCmd.AddCommand(pubsubLiteTopicsCmd)
}

func pslTopicName(id, project, location string) string {
	return pubsubLiteChild("topics", id, pubsubLiteLocationParent(project, location))
}

func pslTopicService(ctx context.Context) (*pubsublite.Service, string, error) {
	project, err := resolveProject()
	if err != nil {
		return nil, "", err
	}
	region, err := pubsubLiteRegion(flagPSLTopicLocation)
	if err != nil {
		return nil, "", err
	}
	svc, err := gcp.PubSubLiteService(ctx, flagAccount, region)
	if err != nil {
		return nil, "", err
	}
	return svc, project, nil
}

func runPSLTopicCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslTopicService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Topic{}
	if err := loadYAMLOrJSONInto(flagPSLTopicConfigFile, body); err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Topics.
		Create(pubsubLiteLocationParent(project, flagPSLTopicLocation), body).
		TopicId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating topic: %w", err)
	}
	return emitFormatted(got, flagPSLTopicFormat)
}

func runPSLTopicDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslTopicService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Admin.Projects.Locations.Topics.
		Delete(pslTopicName(args[0], project, flagPSLTopicLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting topic: %w", err)
	}
	fmt.Printf("Deleted topic [%s].\n", args[0])
	return nil
}

func runPSLTopicDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslTopicService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Topics.
		Get(pslTopicName(args[0], project, flagPSLTopicLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing topic: %w", err)
	}
	return emitFormatted(got, flagPSLTopicFormat)
}

func runPSLTopicList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslTopicService(ctx)
	if err != nil {
		return err
	}
	parent := pubsubLiteLocationParent(project, flagPSLTopicLocation)
	var all []*pubsublite.Topic
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Topics.List(parent).Context(ctx)
		if flagPSLTopicPageSize > 0 {
			call = call.PageSize(flagPSLTopicPageSize)
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
	return emitFormatted(all, flagPSLTopicFormat)
}

func runPSLTopicListSubs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslTopicService(ctx)
	if err != nil {
		return err
	}
	name := pslTopicName(args[0], project, flagPSLTopicLocation)
	var all []string
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Topics.Subscriptions.List(name).Context(ctx)
		if flagPSLTopicPageSize > 0 {
			call = call.PageSize(flagPSLTopicPageSize)
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
	return emitFormatted(all, flagPSLTopicFormat)
}

func runPSLTopicPublish(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("publish is a streaming gRPC operation not implemented over REST; use a native pub/sub client library for publishing")
}

func runPSLTopicUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslTopicService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Topic{}
	if err := loadYAMLOrJSONInto(flagPSLTopicConfigFile, body); err != nil {
		return err
	}
	mask := flagPSLTopicUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	got, err := svc.Admin.Projects.Locations.Topics.
		Patch(pslTopicName(args[0], project, flagPSLTopicLocation), body).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating topic: %w", err)
	}
	return emitFormatted(got, flagPSLTopicFormat)
}
