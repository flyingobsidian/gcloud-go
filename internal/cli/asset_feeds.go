package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	asset "google.golang.org/api/cloudasset/v1"
)

func (s *assetCmds) initAssetFeeds() {
	for _, c := range []*cobra.Command{
		s.assetFeedCreateCmd, s.assetFeedDeleteCmd, s.assetFeedDescribeCmd, s.assetFeedListCmd, s.assetFeedUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagAssetFeedProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&s.flagAssetFeedFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&s.flagAssetFeedOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}

	for _, c := range []*cobra.Command{s.assetFeedCreateCmd, s.assetFeedUpdateCmd} {
		c.Flags().StringSliceVar(&s.flagAssetFeedAssetNames, "asset-names", nil, "Full asset resource names to receive updates for")
		c.Flags().StringSliceVar(&s.flagAssetFeedAssetTypes, "asset-types", nil, "Asset types to receive updates for (e.g. compute.googleapis.com/Disk)")
		c.Flags().StringVar(&s.flagAssetFeedContentType, "content-type", "", "Asset content type (resource, iam-policy, org-policy, access-policy, os-inventory, relationship)")
		c.Flags().StringVar(&s.flagAssetFeedPubsubTopic, "pubsub-topic", "", "Full Pub/Sub topic resource name (required)")
		c.Flags().StringVar(&s.flagAssetFeedConditionExpr, "condition-expression", "", "CEL expression that filters published updates")
		c.Flags().StringVar(&s.flagAssetFeedConditionTitle, "condition-title", "", "Title for the condition")
		c.Flags().StringVar(&s.flagAssetFeedConditionDesc, "condition-description", "", "Description for the condition")
		c.Flags().StringSliceVar(&s.flagAssetFeedRelationships, "relationship-types", nil, "Relationship types to receive updates for")
	}
	s.assetFeedCreateCmd.MarkFlagRequired("pubsub-topic")

	s.assetFeedListCmd.Flags().StringVar(&s.flagAssetFeedListFormat, "format", "", "Output format (json, yaml, or table)")

	s.assetFeedsCmd.AddCommand(
		s.assetFeedCreateCmd, s.assetFeedDeleteCmd, s.assetFeedDescribeCmd, s.assetFeedListCmd, s.assetFeedUpdateCmd,
	)
	s.assetCmd.AddCommand(s.assetFeedsCmd)
}

func (s *assetCmds) feedParent() (string, error) {
	return resolveAssetScope(s.flagAssetFeedProject, s.flagAssetFeedFolder, s.flagAssetFeedOrg)
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

func (s *assetCmds) buildFeed(parent string, args []string) *asset.Feed {
	f := &asset.Feed{
		AssetNames:        s.flagAssetFeedAssetNames,
		AssetTypes:        s.flagAssetFeedAssetTypes,
		ContentType:       normalizeContentType(s.flagAssetFeedContentType),
		RelationshipTypes: s.flagAssetFeedRelationships,
	}
	if s.flagAssetFeedPubsubTopic != "" {
		f.FeedOutputConfig = &asset.FeedOutputConfig{
			PubsubDestination: &asset.PubsubDestination{Topic: s.flagAssetFeedPubsubTopic},
		}
	}
	if s.flagAssetFeedConditionExpr != "" || s.flagAssetFeedConditionTitle != "" || s.flagAssetFeedConditionDesc != "" {
		f.Condition = &asset.Expr{
			Expression:  s.flagAssetFeedConditionExpr,
			Title:       s.flagAssetFeedConditionTitle,
			Description: s.flagAssetFeedConditionDesc,
		}
	}
	return f
}

func (s *assetCmds) runAssetFeedCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.feedParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	feed, err := svc.Feeds.Create(parent, &asset.CreateFeedRequest{
		FeedId: args[0],
		Feed:   s.buildFeed(parent, args),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating feed: %w", err)
	}
	return s.inv.yaml(feed)
}

func (s *assetCmds) runAssetFeedDelete(cmd *cobra.Command, args []string) error {
	parent, err := s.feedParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Feeds.Delete(feedName(parent, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting feed: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted feed [%s].\n", args[0])
	return nil
}

func (s *assetCmds) runAssetFeedDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.feedParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	feed, err := svc.Feeds.Get(feedName(parent, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing feed: %w", err)
	}
	return s.inv.yaml(feed)
}

func (s *assetCmds) runAssetFeedList(cmd *cobra.Command, args []string) error {
	parent, err := s.feedParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Feeds.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing feeds: %w", err)
	}

	return s.inv.printListResults(resp.Feeds, s.flagAssetFeedListFormat, func() {
		printTableHeadingTo(s.inv.stdout, len(resp.Feeds), "%-60s %s\n", "NAME", "TOPIC")
		for _, f := range resp.Feeds {
			topic := ""
			if f.FeedOutputConfig != nil && f.FeedOutputConfig.PubsubDestination != nil {
				topic = f.FeedOutputConfig.PubsubDestination.Topic
			}
			fmt.Fprintf(s.inv.stdout, "%-60s %s\n", f.Name, topic)
		}
	})
}

func (s *assetCmds) runAssetFeedUpdate(cmd *cobra.Command, args []string) error {
	parent, err := s.feedParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}

	feed := s.buildFeed(parent, args)
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
	return s.inv.yaml(updated)
}
