package cli

import (
	"encoding/base64"
	"fmt"
	"path"
	"strings"

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

// --- topics ---

// --- subscriptions ---

func (g *pubsubCmds) initPubsubTopicsSubs() {
	// topics
	topAll := []*cobra.Command{g.psTopCreateCmd, g.psTopDeleteCmd, g.psTopDescribeCmd, g.psTopListCmd, g.psTopUpdateCmd,
		g.psTopPublishCmd, g.psTopListSubsCmd, g.psTopDetachSubCmd,
		g.psTopGetIamCmd, g.psTopSetIamCmd, g.psTopAddIamCmd, g.psTopRemoveIamCmd}
	for _, c := range []*cobra.Command{g.psTopCreateCmd, g.psTopUpdateCmd} {
		c.Flags().StringVar(&g.flagPSConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Topic body (required for update)")
	}
	_ = g.psTopUpdateCmd.MarkFlagRequired("config-file")
	g.psTopUpdateCmd.Flags().StringVar(&g.flagPSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.psTopDescribeCmd, g.psTopListCmd, g.psTopListSubsCmd, g.psTopGetIamCmd} {
		c.Flags().StringVar(&g.flagPSFormat, "format", "", "Output format")
	}
	g.psTopPublishCmd.Flags().StringVar(&g.flagPSPublishMessage, "message", "", "Message body (required unless --attribute is set)")
	g.psTopPublishCmd.Flags().StringToStringVar(&g.flagPSPublishAttributes, "attribute", nil, "Message attribute (key=value); may repeat")
	g.psTopPublishCmd.Flags().StringVar(&g.flagPSPublishOrderKey, "ordering-key", "", "Message ordering key")
	for _, c := range []*cobra.Command{g.psTopAddIamCmd, g.psTopRemoveIamCmd} {
		c.Flags().StringVar(&g.flagPSIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&g.flagPSIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	g.pubsubTopicsCmd.AddCommand(topAll...)
	g.pubsubCmd.AddCommand(g.pubsubTopicsCmd)

	// subscriptions
	subAll := []*cobra.Command{g.psSubCreateCmd, g.psSubDeleteCmd, g.psSubDescribeCmd, g.psSubListCmd, g.psSubUpdateCmd,
		g.psSubAckCmd, g.psSubPullCmd, g.psSubSeekCmd, g.psSubModifyAckCmd, g.psSubModifyPushCmd,
		g.psSubGetIamCmd, g.psSubSetIamCmd, g.psSubAddIamCmd, g.psSubRemoveIamCmd}
	g.psSubCreateCmd.Flags().StringVar(&g.flagPSConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Subscription body (recommended for non-default fields)")
	g.psSubCreateCmd.Flags().StringVar(&g.flagPSSubTopic, "topic", "",
		"Topic that the subscription is attached to (required unless supplied in --config-file)")
	g.psSubUpdateCmd.Flags().StringVar(&g.flagPSConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Subscription body (required)")
	_ = g.psSubUpdateCmd.MarkFlagRequired("config-file")
	g.psSubUpdateCmd.Flags().StringVar(&g.flagPSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.psSubDescribeCmd, g.psSubListCmd, g.psSubGetIamCmd, g.psSubPullCmd} {
		c.Flags().StringVar(&g.flagPSFormat, "format", "", "Output format")
	}
	g.psSubPullCmd.Flags().Int64Var(&g.flagPSMaxMessages, "limit", 1, "Maximum messages to pull")
	g.psSubSeekCmd.Flags().StringVar(&g.flagPSSeekTime, "time", "", "RFC 3339 timestamp to seek to")
	g.psSubSeekCmd.Flags().StringVar(&g.flagPSSeekSnapshot, "snapshot", "", "Snapshot to seek to")
	g.psSubModifyAckCmd.Flags().Int64Var(&g.flagPSAckDeadline, "ack-deadline", 10, "New ack deadline in seconds")
	g.psSubModifyPushCmd.Flags().StringVar(&g.flagPSPushEndpoint, "push-endpoint", "", "Push endpoint URL ('' clears push)")
	g.psSubModifyPushCmd.Flags().StringVar(&g.flagPSPushAudience, "push-auth-audience", "", "OIDC audience for push authentication")
	for _, c := range []*cobra.Command{g.psSubAddIamCmd, g.psSubRemoveIamCmd} {
		c.Flags().StringVar(&g.flagPSIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&g.flagPSIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	g.pubsubSubscriptionsCmd.AddCommand(subAll...)
	g.pubsubCmd.AddCommand(g.pubsubSubscriptionsCmd)
}

// --- topics impl ---

func (g *pubsubCmds) runPSTopCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	t := &pubsub.Topic{}
	if g.flagPSConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagPSConfigFile, t); err != nil {
			return err
		}
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Topics.Create(psTopicName(project, args[0]), t).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating topic: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSTopDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Topics.Delete(psTopicName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting topic: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted topic [%s].\n", args[0])
	return nil
}

func (g *pubsubCmds) runPSTopDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Topics.Get(psTopicName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing topic: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSFormat)
}

func (g *pubsubCmds) runPSTopList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
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
	if g.flagPSFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagPSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%s\n", "NAME")
	for _, t := range all {
		fmt.Fprintln(g.inv.stdout, path.Base(t.Name))
	}
	return nil
}

func (g *pubsubCmds) runPSTopUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	t := &pubsub.Topic{}
	if err := loadYAMLOrJSONInto(g.flagPSConfigFile, t); err != nil {
		return err
	}
	mask := g.flagPSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(t))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Topics.Patch(psTopicName(project, args[0]), &pubsub.UpdateTopicRequest{
		Topic: t, UpdateMask: mask,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating topic: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSTopPublish(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	if g.flagPSPublishMessage == "" && len(g.flagPSPublishAttributes) == 0 {
		return fmt.Errorf("--message or at least one --attribute is required")
	}
	msg := &pubsub.PubsubMessage{
		Data:        base64.StdEncoding.EncodeToString([]byte(g.flagPSPublishMessage)),
		Attributes:  g.flagPSPublishAttributes,
		OrderingKey: g.flagPSPublishOrderKey,
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Topics.Publish(psTopicName(project, args[0]),
		&pubsub.PublishRequest{Messages: []*pubsub.PubsubMessage{msg}}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("publishing message: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, "")
}

func (g *pubsubCmds) runPSTopListSubs(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Topics.Subscriptions.List(psTopicName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing subscriptions: %w", err)
	}
	if g.flagPSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Subscriptions, g.flagPSFormat)
	}
	for _, s := range resp.Subscriptions {
		fmt.Fprintln(g.inv.stdout, s)
	}
	return nil
}

func (g *pubsubCmds) runPSTopDetachSub(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.Detach(psSubName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("detaching subscription: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Detached subscription [%s].\n", args[0])
	return nil
}

// topic IAM
func (g *pubsubCmds) runPSTopGetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Topics.GetIamPolicy(psTopicName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, p, g.flagPSFormat)
}

func (g *pubsubCmds) runPSTopSetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	policy := &pubsub.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Topics.SetIamPolicy(psTopicName(project, args[0]),
		&pubsub.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSTopAddIam(cmd *cobra.Command, args []string) error {
	return g.psModifyIamTop(args[0], func(p *pubsub.Policy) {
		psAddBinding(p, g.flagPSIamRole, g.flagPSIamMember)
	})
}

func (g *pubsubCmds) runPSTopRemoveIam(cmd *cobra.Command, args []string) error {
	return g.psModifyIamTop(args[0], func(p *pubsub.Policy) {
		psRemoveBinding(p, g.flagPSIamRole, g.flagPSIamMember)
	})
}

func (g *pubsubCmds) psModifyIamTop(topic string, mutate func(*pubsub.Policy)) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().PubSub(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, "")
}

// --- subscriptions impl ---

func (g *pubsubCmds) runPSSubCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &pubsub.Subscription{}
	if g.flagPSConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagPSConfigFile, s); err != nil {
			return err
		}
	}
	if s.Topic == "" {
		if g.flagPSSubTopic == "" {
			return fmt.Errorf("--topic is required")
		}
		s.Topic = psTopicName(project, g.flagPSSubTopic)
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Subscriptions.Create(psSubName(project, args[0]), s).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSubDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.Delete(psSubName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting subscription: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted subscription [%s].\n", args[0])
	return nil
}

func (g *pubsubCmds) runPSSubDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Subscriptions.Get(psSubName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing subscription: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSFormat)
}

func (g *pubsubCmds) runPSSubList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
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
	if g.flagPSFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagPSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "TOPIC")
	for _, s := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(s.Name), s.Topic)
	}
	return nil
}

func (g *pubsubCmds) runPSSubUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &pubsub.Subscription{}
	if err := loadYAMLOrJSONInto(g.flagPSConfigFile, s); err != nil {
		return err
	}
	mask := g.flagPSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Subscriptions.Patch(psSubName(project, args[0]), &pubsub.UpdateSubscriptionRequest{
		Subscription: s, UpdateMask: mask,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating subscription: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSubAck(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.Acknowledge(psSubName(project, args[0]),
		&pubsub.AcknowledgeRequest{AckIds: args[1:]}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("acknowledging messages: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Acknowledged %d message(s) on subscription [%s].\n", len(args)-1, args[0])
	return nil
}

func (g *pubsubCmds) runPSSubPull(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Subscriptions.Pull(psSubName(project, args[0]),
		&pubsub.PullRequest{MaxMessages: g.flagPSMaxMessages}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("pulling messages: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagPSFormat)
}

func (g *pubsubCmds) runPSSubSeek(cmd *cobra.Command, args []string) error {
	if g.flagPSSeekTime == "" && g.flagPSSeekSnapshot == "" {
		return fmt.Errorf("either --time or --snapshot is required")
	}
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	req := &pubsub.SeekRequest{}
	if g.flagPSSeekTime != "" {
		req.Time = g.flagPSSeekTime
	}
	if g.flagPSSeekSnapshot != "" {
		req.Snapshot = g.flagPSSeekSnapshot
		if !strings.HasPrefix(req.Snapshot, "projects/") {
			req.Snapshot = fmt.Sprintf("projects/%s/snapshots/%s", project, g.flagPSSeekSnapshot)
		}
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Subscriptions.Seek(psSubName(project, args[0]), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("seeking subscription: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, "")
}

func (g *pubsubCmds) runPSSubModifyAck(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.ModifyAckDeadline(psSubName(project, args[0]),
		&pubsub.ModifyAckDeadlineRequest{
			AckIds:             args[1:],
			AckDeadlineSeconds: g.flagPSAckDeadline,
		}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("modifying ack deadline: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Modified ack deadline for %d message(s) on subscription [%s].\n", len(args)-1, args[0])
	return nil
}

func (g *pubsubCmds) runPSSubModifyPush(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	req := &pubsub.ModifyPushConfigRequest{PushConfig: &pubsub.PushConfig{PushEndpoint: g.flagPSPushEndpoint}}
	if g.flagPSPushAudience != "" {
		req.PushConfig.OidcToken = &pubsub.OidcToken{Audience: g.flagPSPushAudience}
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Subscriptions.ModifyPushConfig(psSubName(project, args[0]), req).Context(ctx).Do(); err != nil {
		return fmt.Errorf("modifying push config: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Modified push configuration for subscription [%s].\n", args[0])
	return nil
}

// subscription IAM
func (g *pubsubCmds) runPSSubGetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Subscriptions.GetIamPolicy(psSubName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, p, g.flagPSFormat)
}

func (g *pubsubCmds) runPSSubSetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	policy := &pubsub.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().PubSub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Subscriptions.SetIamPolicy(psSubName(project, args[0]),
		&pubsub.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *pubsubCmds) runPSSubAddIam(cmd *cobra.Command, args []string) error {
	return g.psModifyIamSub(args[0], func(p *pubsub.Policy) {
		psAddBinding(p, g.flagPSIamRole, g.flagPSIamMember)
	})
}

func (g *pubsubCmds) runPSSubRemoveIam(cmd *cobra.Command, args []string) error {
	return g.psModifyIamSub(args[0], func(p *pubsub.Policy) {
		psRemoveBinding(p, g.flagPSIamRole, g.flagPSIamMember)
	})
}

func (g *pubsubCmds) psModifyIamSub(sub string, mutate func(*pubsub.Policy)) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().PubSub(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, "")
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
