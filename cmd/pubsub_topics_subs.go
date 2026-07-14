package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	pubsub "google.golang.org/api/pubsub/v1"
)

// --- gcloud pubsub topics + subscriptions (#1178, #1179) ---

func psTopicName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/topics/%s", project, id)
}

func psSubName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/subscriptions/%s", project, id)
}

var (
	flagPSConfigFile string
	flagPSUpdateMask string
	flagPSFormat     string
	flagPSIamMember  string
	flagPSIamRole    string

	// topics publish
	flagPSPublishMessage    string
	flagPSPublishAttributes map[string]string
	flagPSPublishOrderKey   string

	// subscriptions ack / seek / modify
	flagPSAckDeadline  int64
	flagPSMaxMessages  int64
	flagPSSeekTime     string
	flagPSSeekSnapshot string
	flagPSPushEndpoint string
	flagPSPushAudience string

	// subscriptions create hints
	flagPSSubTopic string
)

// --- topics ---

var pubsubTopicsCmd = &cobra.Command{Use: "topics", Short: "Manage Pub/Sub topics"}

var (
	psTopCreateCmd = &cobra.Command{
		Use: "create TOPIC", Short: "Create a Pub/Sub topic (optionally from a --config-file)",
		Args: cobra.ExactArgs(1), RunE: runPSTopCreate,
	}
	psTopDeleteCmd = &cobra.Command{
		Use: "delete TOPIC", Short: "Delete a Pub/Sub topic",
		Args: cobra.ExactArgs(1), RunE: runPSTopDelete,
	}
	psTopDescribeCmd = &cobra.Command{
		Use: "describe TOPIC", Short: "Describe a Pub/Sub topic",
		Args: cobra.ExactArgs(1), RunE: runPSTopDescribe,
	}
	psTopListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub topics in the project",
		Args: cobra.NoArgs, RunE: runPSTopList,
	}
	psTopUpdateCmd = &cobra.Command{
		Use: "update TOPIC", Short: "Update a Pub/Sub topic from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPSTopUpdate,
	}
	psTopPublishCmd = &cobra.Command{
		Use: "publish TOPIC", Short: "Publish a message to a Pub/Sub topic",
		Args: cobra.ExactArgs(1), RunE: runPSTopPublish,
	}
	psTopListSubsCmd = &cobra.Command{
		Use: "list-subscriptions TOPIC", Short: "List subscriptions attached to a topic",
		Args: cobra.ExactArgs(1), RunE: runPSTopListSubs,
	}
	psTopDetachSubCmd = &cobra.Command{
		Use: "detach-subscription SUBSCRIPTION", Short: "Detach a subscription from its topic",
		Args: cobra.ExactArgs(1), RunE: runPSTopDetachSub,
	}
	psTopGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TOPIC", Short: "Get the IAM policy for a topic",
		Args: cobra.ExactArgs(1), RunE: runPSTopGetIam,
	}
	psTopSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TOPIC POLICY_FILE", Short: "Replace the IAM policy for a topic",
		Args: cobra.ExactArgs(2), RunE: runPSTopSetIam,
	}
	psTopAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TOPIC", Short: "Add an IAM binding to a topic",
		Args: cobra.ExactArgs(1), RunE: runPSTopAddIam,
	}
	psTopRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TOPIC", Short: "Remove an IAM binding from a topic",
		Args: cobra.ExactArgs(1), RunE: runPSTopRemoveIam,
	}
)

// --- subscriptions ---

var pubsubSubscriptionsCmd = &cobra.Command{Use: "subscriptions", Short: "Manage Pub/Sub subscriptions"}

var (
	psSubCreateCmd = &cobra.Command{
		Use: "create SUBSCRIPTION", Short: "Create a Pub/Sub subscription (optionally from a --config-file)",
		Args: cobra.ExactArgs(1), RunE: runPSSubCreate,
	}
	psSubDeleteCmd = &cobra.Command{
		Use: "delete SUBSCRIPTION", Short: "Delete a Pub/Sub subscription",
		Args: cobra.ExactArgs(1), RunE: runPSSubDelete,
	}
	psSubDescribeCmd = &cobra.Command{
		Use: "describe SUBSCRIPTION", Short: "Describe a Pub/Sub subscription",
		Args: cobra.ExactArgs(1), RunE: runPSSubDescribe,
	}
	psSubListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub subscriptions in the project",
		Args: cobra.NoArgs, RunE: runPSSubList,
	}
	psSubUpdateCmd = &cobra.Command{
		Use: "update SUBSCRIPTION", Short: "Update a subscription from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPSSubUpdate,
	}
	psSubAckCmd = &cobra.Command{
		Use: "ack SUBSCRIPTION ACK_ID [ACK_ID ...]", Short: "Acknowledge messages on a subscription",
		Args: cobra.MinimumNArgs(2), RunE: runPSSubAck,
	}
	psSubPullCmd = &cobra.Command{
		Use: "pull SUBSCRIPTION", Short: "Pull messages from a subscription",
		Args: cobra.ExactArgs(1), RunE: runPSSubPull,
	}
	psSubSeekCmd = &cobra.Command{
		Use: "seek SUBSCRIPTION", Short: "Seek a subscription to a time or snapshot",
		Args: cobra.ExactArgs(1), RunE: runPSSubSeek,
	}
	psSubModifyAckCmd = &cobra.Command{
		Use: "modify-message-ack-deadline SUBSCRIPTION ACK_ID [ACK_ID ...]",
		Short: "Modify the ack deadline for outstanding messages",
		Args: cobra.MinimumNArgs(2), RunE: runPSSubModifyAck,
	}
	psSubModifyPushCmd = &cobra.Command{
		Use: "modify-push-config SUBSCRIPTION", Short: "Modify the push configuration for a subscription",
		Args: cobra.ExactArgs(1), RunE: runPSSubModifyPush,
	}
	psSubGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SUBSCRIPTION", Short: "Get the IAM policy for a subscription",
		Args: cobra.ExactArgs(1), RunE: runPSSubGetIam,
	}
	psSubSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SUBSCRIPTION POLICY_FILE", Short: "Replace the IAM policy for a subscription",
		Args: cobra.ExactArgs(2), RunE: runPSSubSetIam,
	}
	psSubAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SUBSCRIPTION", Short: "Add an IAM binding to a subscription",
		Args: cobra.ExactArgs(1), RunE: runPSSubAddIam,
	}
	psSubRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SUBSCRIPTION", Short: "Remove an IAM binding from a subscription",
		Args: cobra.ExactArgs(1), RunE: runPSSubRemoveIam,
	}
)

func init() {
	// topics
	topAll := []*cobra.Command{psTopCreateCmd, psTopDeleteCmd, psTopDescribeCmd, psTopListCmd, psTopUpdateCmd,
		psTopPublishCmd, psTopListSubsCmd, psTopDetachSubCmd,
		psTopGetIamCmd, psTopSetIamCmd, psTopAddIamCmd, psTopRemoveIamCmd}
	for _, c := range []*cobra.Command{psTopCreateCmd, psTopUpdateCmd} {
		c.Flags().StringVar(&flagPSConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Topic body (required for update)")
	}
	_ = psTopUpdateCmd.MarkFlagRequired("config-file")
	psTopUpdateCmd.Flags().StringVar(&flagPSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{psTopDescribeCmd, psTopListCmd, psTopListSubsCmd, psTopGetIamCmd} {
		c.Flags().StringVar(&flagPSFormat, "format", "", "Output format")
	}
	psTopPublishCmd.Flags().StringVar(&flagPSPublishMessage, "message", "", "Message body (required unless --attribute is set)")
	psTopPublishCmd.Flags().StringToStringVar(&flagPSPublishAttributes, "attribute", nil, "Message attribute (key=value); may repeat")
	psTopPublishCmd.Flags().StringVar(&flagPSPublishOrderKey, "ordering-key", "", "Message ordering key")
	for _, c := range []*cobra.Command{psTopAddIamCmd, psTopRemoveIamCmd} {
		c.Flags().StringVar(&flagPSIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagPSIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	pubsubTopicsCmd.AddCommand(topAll...)
	pubsubCmd.AddCommand(pubsubTopicsCmd)

	// subscriptions
	subAll := []*cobra.Command{psSubCreateCmd, psSubDeleteCmd, psSubDescribeCmd, psSubListCmd, psSubUpdateCmd,
		psSubAckCmd, psSubPullCmd, psSubSeekCmd, psSubModifyAckCmd, psSubModifyPushCmd,
		psSubGetIamCmd, psSubSetIamCmd, psSubAddIamCmd, psSubRemoveIamCmd}
	psSubCreateCmd.Flags().StringVar(&flagPSConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Subscription body (recommended for non-default fields)")
	psSubCreateCmd.Flags().StringVar(&flagPSSubTopic, "topic", "",
		"Topic that the subscription is attached to (required unless supplied in --config-file)")
	psSubUpdateCmd.Flags().StringVar(&flagPSConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Subscription body (required)")
	_ = psSubUpdateCmd.MarkFlagRequired("config-file")
	psSubUpdateCmd.Flags().StringVar(&flagPSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{psSubDescribeCmd, psSubListCmd, psSubGetIamCmd, psSubPullCmd} {
		c.Flags().StringVar(&flagPSFormat, "format", "", "Output format")
	}
	psSubPullCmd.Flags().Int64Var(&flagPSMaxMessages, "limit", 1, "Maximum messages to pull")
	psSubSeekCmd.Flags().StringVar(&flagPSSeekTime, "time", "", "RFC 3339 timestamp to seek to")
	psSubSeekCmd.Flags().StringVar(&flagPSSeekSnapshot, "snapshot", "", "Snapshot to seek to")
	psSubModifyAckCmd.Flags().Int64Var(&flagPSAckDeadline, "ack-deadline", 10, "New ack deadline in seconds")
	psSubModifyPushCmd.Flags().StringVar(&flagPSPushEndpoint, "push-endpoint", "", "Push endpoint URL ('' clears push)")
	psSubModifyPushCmd.Flags().StringVar(&flagPSPushAudience, "push-auth-audience", "", "OIDC audience for push authentication")
	for _, c := range []*cobra.Command{psSubAddIamCmd, psSubRemoveIamCmd} {
		c.Flags().StringVar(&flagPSIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagPSIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	pubsubSubscriptionsCmd.AddCommand(subAll...)
	pubsubCmd.AddCommand(pubsubSubscriptionsCmd)
}

// --- topics impl ---

func runPSTopCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	t := &pubsub.Topic{}
	if flagPSConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagPSConfigFile, t); err != nil {
			return err
		}
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Topics.Create(psTopicName(project, args[0]), t).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating topic: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSTopDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Topics.Delete(psTopicName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting topic: %w", err)
	}
	fmt.Printf("Deleted topic [%s].\n", args[0])
	return nil
}

func runPSTopDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Topics.Get(psTopicName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing topic: %w", err)
	}
	return emitFormatted(got, flagPSFormat)
}

func runPSTopList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*pubsub.Topic
	pageToken := ""
	for {
		call := svc.Projects.Topics.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
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
	if flagPSFormat != "" {
		return emitFormatted(all, flagPSFormat)
	}
	fmt.Printf("%s\n", "NAME")
	for _, t := range all {
		fmt.Println(path.Base(t.Name))
	}
	return nil
}

func runPSTopUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	t := &pubsub.Topic{}
	if err := loadYAMLOrJSONInto(flagPSConfigFile, t); err != nil {
		return err
	}
	mask := flagPSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(t))
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Topics.Patch(psTopicName(project, args[0]), &pubsub.UpdateTopicRequest{
		Topic: t, UpdateMask: mask,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating topic: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSTopPublish(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagPSPublishMessage == "" && len(flagPSPublishAttributes) == 0 {
		return fmt.Errorf("--message or at least one --attribute is required")
	}
	msg := &pubsub.PubsubMessage{
		Data:        base64.StdEncoding.EncodeToString([]byte(flagPSPublishMessage)),
		Attributes:  flagPSPublishAttributes,
		OrderingKey: flagPSPublishOrderKey,
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Topics.Publish(psTopicName(project, args[0]),
		&pubsub.PublishRequest{Messages: []*pubsub.PubsubMessage{msg}}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("publishing message: %w", err)
	}
	return emitFormatted(resp, "")
}

func runPSTopListSubs(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Topics.Subscriptions.List(psTopicName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing subscriptions: %w", err)
	}
	if flagPSFormat != "" {
		return emitFormatted(resp.Subscriptions, flagPSFormat)
	}
	for _, s := range resp.Subscriptions {
		fmt.Println(s)
	}
	return nil
}

func runPSTopDetachSub(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.Detach(psSubName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("detaching subscription: %w", err)
	}
	fmt.Printf("Detached subscription [%s].\n", args[0])
	return nil
}

// topic IAM
func runPSTopGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Topics.GetIamPolicy(psTopicName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(p, flagPSFormat)
}

func runPSTopSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &pubsub.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Topics.SetIamPolicy(psTopicName(project, args[0]),
		&pubsub.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSTopAddIam(cmd *cobra.Command, args []string) error {
	return psModifyIamTop(args[0], func(p *pubsub.Policy) {
		psAddBinding(p, flagPSIamRole, flagPSIamMember)
	})
}

func runPSTopRemoveIam(cmd *cobra.Command, args []string) error {
	return psModifyIamTop(args[0], func(p *pubsub.Policy) {
		psRemoveBinding(p, flagPSIamRole, flagPSIamMember)
	})
}

func psModifyIamTop(topic string, mutate func(*pubsub.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := psTopicName(project, topic)
	p, err := svc.Projects.Topics.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(p)
	got, err := svc.Projects.Topics.SetIamPolicy(resource, &pubsub.SetIamPolicyRequest{Policy: p}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// --- subscriptions impl ---

func runPSSubCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &pubsub.Subscription{}
	if flagPSConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagPSConfigFile, s); err != nil {
			return err
		}
	}
	if s.Topic == "" {
		if flagPSSubTopic == "" {
			return fmt.Errorf("--topic is required")
		}
		s.Topic = psTopicName(project, flagPSSubTopic)
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Subscriptions.Create(psSubName(project, args[0]), s).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSubDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.Delete(psSubName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting subscription: %w", err)
	}
	fmt.Printf("Deleted subscription [%s].\n", args[0])
	return nil
}

func runPSSubDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Subscriptions.Get(psSubName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing subscription: %w", err)
	}
	return emitFormatted(got, flagPSFormat)
}

func runPSSubList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*pubsub.Subscription
	pageToken := ""
	for {
		call := svc.Projects.Subscriptions.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
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
	if flagPSFormat != "" {
		return emitFormatted(all, flagPSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "TOPIC")
	for _, s := range all {
		fmt.Printf("%-40s %s\n", path.Base(s.Name), s.Topic)
	}
	return nil
}

func runPSSubUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &pubsub.Subscription{}
	if err := loadYAMLOrJSONInto(flagPSConfigFile, s); err != nil {
		return err
	}
	mask := flagPSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Subscriptions.Patch(psSubName(project, args[0]), &pubsub.UpdateSubscriptionRequest{
		Subscription: s, UpdateMask: mask,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating subscription: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSubAck(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.Acknowledge(psSubName(project, args[0]),
		&pubsub.AcknowledgeRequest{AckIds: args[1:]}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("acknowledging messages: %w", err)
	}
	fmt.Printf("Acknowledged %d message(s) on subscription [%s].\n", len(args)-1, args[0])
	return nil
}

func runPSSubPull(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Subscriptions.Pull(psSubName(project, args[0]),
		&pubsub.PullRequest{MaxMessages: flagPSMaxMessages}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("pulling messages: %w", err)
	}
	return emitFormatted(resp, flagPSFormat)
}

func runPSSubSeek(cmd *cobra.Command, args []string) error {
	if flagPSSeekTime == "" && flagPSSeekSnapshot == "" {
		return fmt.Errorf("either --time or --snapshot is required")
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &pubsub.SeekRequest{}
	if flagPSSeekTime != "" {
		req.Time = flagPSSeekTime
	}
	if flagPSSeekSnapshot != "" {
		req.Snapshot = flagPSSeekSnapshot
		if !strings.HasPrefix(req.Snapshot, "projects/") {
			req.Snapshot = fmt.Sprintf("projects/%s/snapshots/%s", project, flagPSSeekSnapshot)
		}
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Subscriptions.Seek(psSubName(project, args[0]), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("seeking subscription: %w", err)
	}
	return emitFormatted(resp, "")
}

func runPSSubModifyAck(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.ModifyAckDeadline(psSubName(project, args[0]),
		&pubsub.ModifyAckDeadlineRequest{
			AckIds:             args[1:],
			AckDeadlineSeconds: flagPSAckDeadline,
		}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("modifying ack deadline: %w", err)
	}
	fmt.Printf("Modified ack deadline for %d message(s) on subscription [%s].\n", len(args)-1, args[0])
	return nil
}

func runPSSubModifyPush(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &pubsub.ModifyPushConfigRequest{PushConfig: &pubsub.PushConfig{PushEndpoint: flagPSPushEndpoint}}
	if flagPSPushAudience != "" {
		req.PushConfig.OidcToken = &pubsub.OidcToken{Audience: flagPSPushAudience}
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.ModifyPushConfig(psSubName(project, args[0]), req).Context(ctx).Do(); err != nil {
		return fmt.Errorf("modifying push config: %w", err)
	}
	fmt.Printf("Modified push configuration for subscription [%s].\n", args[0])
	return nil
}

// subscription IAM
func runPSSubGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Subscriptions.GetIamPolicy(psSubName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(p, flagPSFormat)
}

func runPSSubSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &pubsub.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Subscriptions.SetIamPolicy(psSubName(project, args[0]),
		&pubsub.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runPSSubAddIam(cmd *cobra.Command, args []string) error {
	return psModifyIamSub(args[0], func(p *pubsub.Policy) {
		psAddBinding(p, flagPSIamRole, flagPSIamMember)
	})
}

func runPSSubRemoveIam(cmd *cobra.Command, args []string) error {
	return psModifyIamSub(args[0], func(p *pubsub.Policy) {
		psRemoveBinding(p, flagPSIamRole, flagPSIamMember)
	})
}

func psModifyIamSub(sub string, mutate func(*pubsub.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PubSubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := psSubName(project, sub)
	p, err := svc.Projects.Subscriptions.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(p)
	got, err := svc.Projects.Subscriptions.SetIamPolicy(resource, &pubsub.SetIamPolicyRequest{Policy: p}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// --- IAM helpers ---

func psAddBinding(p *pubsub.Policy, role, member string) {
	for _, b := range p.Bindings {
		if b.Role == role {
			for _, m := range b.Members {
				if m == member {
					return
				}
			}
			b.Members = append(b.Members, member)
			return
		}
	}
	p.Bindings = append(p.Bindings, &pubsub.Binding{Role: role, Members: []string{member}})
}

func psRemoveBinding(p *pubsub.Policy, role, member string) {
	for _, b := range p.Bindings {
		if b.Role != role {
			continue
		}
		out := b.Members[:0]
		for _, m := range b.Members {
			if m != member {
				out = append(out, m)
			}
		}
		b.Members = out
	}
}
