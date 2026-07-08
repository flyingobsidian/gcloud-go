package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	asset "google.golang.org/api/cloudasset/v1"
)

var assetFeedsCmd = &cobra.Command{
	Use:   "feeds",
	Short: "Manage Cloud Asset Inventory feeds",
}

var assetFeedCreateCmd = &cobra.Command{
	Use:   "create FEED_ID",
	Short: "Create an asset feed",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetFeedCreate,
}

var assetFeedDeleteCmd = &cobra.Command{
	Use:   "delete FEED_ID",
	Short: "Delete an asset feed",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetFeedDelete,
}

var assetFeedDescribeCmd = &cobra.Command{
	Use:   "describe FEED_ID",
	Short: "Describe an asset feed",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetFeedDescribe,
}

var assetFeedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List asset feeds under a parent",
	Args:  cobra.NoArgs,
	RunE:  runAssetFeedList,
}

var assetFeedUpdateCmd = &cobra.Command{
	Use:   "update FEED_ID",
	Short: "Update an asset feed",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetFeedUpdate,
}

var (
	flagAssetFeedProject         string
	flagAssetFeedFolder          string
	flagAssetFeedOrg             string
	flagAssetFeedAssetNames      []string
	flagAssetFeedAssetTypes      []string
	flagAssetFeedContentType     string
	flagAssetFeedPubsubTopic     string
	flagAssetFeedConditionExpr   string
	flagAssetFeedConditionTitle  string
	flagAssetFeedConditionDesc   string
	flagAssetFeedRelationships   []string
	flagAssetFeedListFormat      string
)

func init() {
	for _, c := range []*cobra.Command{
		assetFeedCreateCmd, assetFeedDeleteCmd, assetFeedDescribeCmd, assetFeedListCmd, assetFeedUpdateCmd,
	} {
		c.Flags().StringVar(&flagAssetFeedProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&flagAssetFeedFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&flagAssetFeedOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}

	for _, c := range []*cobra.Command{assetFeedCreateCmd, assetFeedUpdateCmd} {
		c.Flags().StringSliceVar(&flagAssetFeedAssetNames, "asset-names", nil, "Full asset resource names to receive updates for")
		c.Flags().StringSliceVar(&flagAssetFeedAssetTypes, "asset-types", nil, "Asset types to receive updates for (e.g. compute.googleapis.com/Disk)")
		c.Flags().StringVar(&flagAssetFeedContentType, "content-type", "", "Asset content type (resource, iam-policy, org-policy, access-policy, os-inventory, relationship)")
		c.Flags().StringVar(&flagAssetFeedPubsubTopic, "pubsub-topic", "", "Full Pub/Sub topic resource name (required)")
		c.Flags().StringVar(&flagAssetFeedConditionExpr, "condition-expression", "", "CEL expression that filters published updates")
		c.Flags().StringVar(&flagAssetFeedConditionTitle, "condition-title", "", "Title for the condition")
		c.Flags().StringVar(&flagAssetFeedConditionDesc, "condition-description", "", "Description for the condition")
		c.Flags().StringSliceVar(&flagAssetFeedRelationships, "relationship-types", nil, "Relationship types to receive updates for")
	}
	assetFeedCreateCmd.MarkFlagRequired("pubsub-topic")

	assetFeedListCmd.Flags().StringVar(&flagAssetFeedListFormat, "format", "", "Output format (json, yaml, or table)")

	assetFeedsCmd.AddCommand(
		assetFeedCreateCmd, assetFeedDeleteCmd, assetFeedDescribeCmd, assetFeedListCmd, assetFeedUpdateCmd,
	)
	assetCmd.AddCommand(assetFeedsCmd)
}

func feedParent() (string, error) {
	return resolveAssetScope(flagAssetFeedProject, flagAssetFeedFolder, flagAssetFeedOrg)
}

func feedName(parent, feedID string) string {
	if strings.Contains(feedID, "/feeds/") {
		return feedID
	}
	return parent + "/feeds/" + strings.TrimPrefix(feedID, "feeds/")
}

// normalizeContentType maps user-friendly content type flags to the API's
// upper-snake enum values.
func normalizeContentType(v string) string {
	if v == "" {
		return ""
	}
	up := strings.ToUpper(strings.ReplaceAll(v, "-", "_"))
	return up
}

func buildFeed(parent string, args []string) *asset.Feed {
	f := &asset.Feed{
		AssetNames:        flagAssetFeedAssetNames,
		AssetTypes:        flagAssetFeedAssetTypes,
		ContentType:       normalizeContentType(flagAssetFeedContentType),
		RelationshipTypes: flagAssetFeedRelationships,
	}
	if flagAssetFeedPubsubTopic != "" {
		f.FeedOutputConfig = &asset.FeedOutputConfig{
			PubsubDestination: &asset.PubsubDestination{Topic: flagAssetFeedPubsubTopic},
		}
	}
	if flagAssetFeedConditionExpr != "" || flagAssetFeedConditionTitle != "" || flagAssetFeedConditionDesc != "" {
		f.Condition = &asset.Expr{
			Expression:  flagAssetFeedConditionExpr,
			Title:       flagAssetFeedConditionTitle,
			Description: flagAssetFeedConditionDesc,
		}
	}
	return f
}

func runAssetFeedCreate(cmd *cobra.Command, args []string) error {
	parent, err := feedParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	feed, err := svc.Feeds.Create(parent, &asset.CreateFeedRequest{
		FeedId: args[0],
		Feed:   buildFeed(parent, args),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating feed: %w", err)
	}
	return yamlEncode(feed)
}

func runAssetFeedDelete(cmd *cobra.Command, args []string) error {
	parent, err := feedParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Feeds.Delete(feedName(parent, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting feed: %w", err)
	}
	fmt.Printf("Deleted feed [%s].\n", args[0])
	return nil
}

func runAssetFeedDescribe(cmd *cobra.Command, args []string) error {
	parent, err := feedParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	feed, err := svc.Feeds.Get(feedName(parent, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing feed: %w", err)
	}
	return yamlEncode(feed)
}

func runAssetFeedList(cmd *cobra.Command, args []string) error {
	parent, err := feedParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Feeds.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	return printListResults(resp.Feeds, flagAssetFeedListFormat, func() {
		fmt.Printf("%-60s %s\n", "NAME", "TOPIC")
		for _, f := range resp.Feeds {
			topic := ""
			if f.FeedOutputConfig != nil && f.FeedOutputConfig.PubsubDestination != nil {
				topic = f.FeedOutputConfig.PubsubDestination.Topic
			}
			fmt.Printf("%-60s %s\n", f.Name, topic)
		}
	})
}

func runAssetFeedUpdate(cmd *cobra.Command, args []string) error {
	parent, err := feedParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}

	feed := buildFeed(parent, args)
	name := feedName(parent, args[0])
	feed.Name = name

	// Build the update mask from flags the user explicitly set.
	var masks []string
	if cmd.Flags().Changed("asset-names") {
		masks = append(masks, "assetNames")
	}
	if cmd.Flags().Changed("asset-types") {
		masks = append(masks, "assetTypes")
	}
	if cmd.Flags().Changed("content-type") {
		masks = append(masks, "contentType")
	}
	if cmd.Flags().Changed("pubsub-topic") {
		masks = append(masks, "feedOutputConfig")
	}
	if cmd.Flags().Changed("condition-expression") || cmd.Flags().Changed("condition-title") || cmd.Flags().Changed("condition-description") {
		masks = append(masks, "condition")
	}
	if cmd.Flags().Changed("relationship-types") {
		masks = append(masks, "relationshipTypes")
	}
	if len(masks) == 0 {
		return fmt.Errorf("at least one field must be provided for update")
	}

	updated, err := svc.Feeds.Patch(name, &asset.UpdateFeedRequest{
		Feed:       feed,
		UpdateMask: strings.Join(masks, ","),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating feed: %w", err)
	}
	return yamlEncode(updated)
}
