package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Manage Eventarc channels",
}

var (
	evChanCreateCmd = &cobra.Command{
		Use: "create CHANNEL", Short: "Create an Eventarc channel from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvChanCreate,
	}
	evChanDeleteCmd = &cobra.Command{
		Use: "delete CHANNEL", Short: "Delete an Eventarc channel",
		Args: cobra.ExactArgs(1), RunE: runEvChanDelete,
	}
	evChanDescribeCmd = &cobra.Command{
		Use: "describe CHANNEL", Short: "Describe an Eventarc channel",
		Args: cobra.ExactArgs(1), RunE: runEvChanDescribe,
	}
	evChanListCmd = &cobra.Command{
		Use: "list", Short: "List Eventarc channels in a location",
		Args: cobra.NoArgs, RunE: runEvChanList,
	}
	evChanUpdateCmd = &cobra.Command{
		Use: "update CHANNEL", Short: "Update an Eventarc channel from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvChanUpdate,
	}
)

var (
	flagEvChanLocation   string
	flagEvChanConfigFile string
	flagEvChanUpdateMask string
	flagEvChanFormat     string
	flagEvChanAsync      bool
	flagEvChanListLimit  int64
	flagEvChanListPage int64
	flagEvChanListURI  bool
)

func init() {
	for _, c := range []*cobra.Command{evChanCreateCmd, evChanDeleteCmd, evChanDescribeCmd, evChanListCmd, evChanUpdateCmd} {
		eventarcAddRegionFlag(c, &flagEvChanLocation, true)
	}
	for _, c := range []*cobra.Command{evChanCreateCmd, evChanUpdateCmd} {
		c.Flags().StringVar(&flagEvChanConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Channel message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	evChanUpdateCmd.Flags().StringVar(&flagEvChanUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{evChanCreateCmd, evChanDeleteCmd, evChanUpdateCmd} {
		c.Flags().BoolVar(&flagEvChanAsync, "async", false, "Return the long-running operation without waiting")
	}
	evChanDescribeCmd.Flags().StringVar(&flagEvChanFormat, "format", "", "Output format")
	evChanListCmd.Flags().StringVar(&flagEvChanFormat, "format", "", "Output format")
	evChanListCmd.Flags().Int64Var(&flagEvChanListPage, "page-size", 0, "Page size for API pagination")
	evChanListCmd.Flags().Int64Var(&flagEvChanListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evChanListCmd.Flags().BoolVar(&flagEvChanListURI, "uri", false, "Print resource names only")

	eventarcChannelsCmd.AddCommand(evChanCreateCmd, evChanDeleteCmd, evChanDescribeCmd, evChanListCmd, evChanUpdateCmd)
	eventarcCmd.AddCommand(eventarcChannelsCmd)
}

func runEvChanCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ch := &eventarc.Channel{}
	if err := loadYAMLOrJSONInto(flagEvChanConfigFile, ch); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Channels.Create(eventarcLocationParent(project, flagEvChanLocation), ch).
		ChannelId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating channel: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Create channel", args[0], flagEvChanAsync)
}

func runEvChanDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Channels.Delete(eventarcResourceName("channels", args[0], project, flagEvChanLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting channel: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Delete channel", args[0], flagEvChanAsync)
}

func runEvChanDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	ch, err := svc.Projects.Locations.Channels.Get(eventarcResourceName("channels", args[0], project, flagEvChanLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing channel: %w", err)
	}
	return emitFormatted(ch, flagEvChanFormat)
}

func runEvChanList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, flagEvChanLocation)
	var all []*eventarc.Channel
	pageToken := ""
	for {
		call := svc.Projects.Locations.Channels.List(parent).Context(ctx)
		if flagEvChanListPage > 0 {
			call = call.PageSize(flagEvChanListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing channels: %w", err)
		}
		all = append(all, resp.Channels...)
		if flagEvChanListLimit > 0 && int64(len(all)) >= flagEvChanListLimit {
			all = all[:flagEvChanListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvChanListURI {
		for _, c := range all {
			fmt.Println(c.Name)
		}
		return nil
	}
	if flagEvChanFormat != "" {
		return emitFormatted(all, flagEvChanFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, c := range all {
		fmt.Printf("%-40s %s\n", path.Base(c.Name), c.State)
	}
	return nil
}

func runEvChanUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ch := &eventarc.Channel{}
	if err := loadYAMLOrJSONInto(flagEvChanConfigFile, ch); err != nil {
		return err
	}
	mask := flagEvChanUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ch))
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Channels.Patch(eventarcResourceName("channels", args[0], project, flagEvChanLocation), ch).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating channel: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Update channel", args[0], flagEvChanAsync)
}
