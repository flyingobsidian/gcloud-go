package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcChannelConnCmd = &cobra.Command{
	Use:   "channel-connections",
	Short: "Manage Eventarc channel connections",
}

var (
	evCCCreateCmd = &cobra.Command{
		Use: "create CHANNEL_CONNECTION", Short: "Create a channel connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvCCCreate,
	}
	evCCDeleteCmd = &cobra.Command{
		Use: "delete CHANNEL_CONNECTION", Short: "Delete a channel connection",
		Args: cobra.ExactArgs(1), RunE: runEvCCDelete,
	}
	evCCDescribeCmd = &cobra.Command{
		Use: "describe CHANNEL_CONNECTION", Short: "Describe a channel connection",
		Args: cobra.ExactArgs(1), RunE: runEvCCDescribe,
	}
	evCCListCmd = &cobra.Command{
		Use: "list", Short: "List channel connections in a location",
		Args: cobra.NoArgs, RunE: runEvCCList,
	}
)

var (
	flagEvCCLocation   string
	flagEvCCConfigFile string
	flagEvCCFormat     string
	flagEvCCAsync      bool
	flagEvCCListLimit  int64
	flagEvCCListPage   int64
	flagEvCCListURI    bool
)

func init() {
	for _, c := range []*cobra.Command{evCCCreateCmd, evCCDeleteCmd, evCCDescribeCmd, evCCListCmd} {
		eventarcAddRegionFlag(c, &flagEvCCLocation, true)
	}
	evCCCreateCmd.Flags().StringVar(&flagEvCCConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ChannelConnection message body (required)")
	_ = evCCCreateCmd.MarkFlagRequired("config-file")
	for _, c := range []*cobra.Command{evCCCreateCmd, evCCDeleteCmd} {
		c.Flags().BoolVar(&flagEvCCAsync, "async", false, "Return the long-running operation without waiting")
	}
	evCCDescribeCmd.Flags().StringVar(&flagEvCCFormat, "format", "", "Output format")
	evCCListCmd.Flags().StringVar(&flagEvCCFormat, "format", "", "Output format")
	evCCListCmd.Flags().Int64Var(&flagEvCCListPage, "page-size", 0, "Page size")
	evCCListCmd.Flags().Int64Var(&flagEvCCListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evCCListCmd.Flags().BoolVar(&flagEvCCListURI, "uri", false, "Print resource names only")

	eventarcChannelConnCmd.AddCommand(evCCCreateCmd, evCCDeleteCmd, evCCDescribeCmd, evCCListCmd)
	eventarcCmd.AddCommand(eventarcChannelConnCmd)
}

func runEvCCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cc := &eventarc.ChannelConnection{}
	if err := loadYAMLOrJSONInto(flagEvCCConfigFile, cc); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ChannelConnections.Create(eventarcLocationParent(project, flagEvCCLocation), cc).
		ChannelConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating channel connection: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Create channel connection", args[0], flagEvCCAsync)
}

func runEvCCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ChannelConnections.Delete(eventarcResourceName("channelConnections", args[0], project, flagEvCCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting channel connection: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Delete channel connection", args[0], flagEvCCAsync)
}

func runEvCCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cc, err := svc.Projects.Locations.ChannelConnections.Get(eventarcResourceName("channelConnections", args[0], project, flagEvCCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing channel connection: %w", err)
	}
	return emitFormatted(cc, flagEvCCFormat)
}

func runEvCCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, flagEvCCLocation)
	var all []*eventarc.ChannelConnection
	pageToken := ""
	for {
		call := svc.Projects.Locations.ChannelConnections.List(parent).Context(ctx)
		if flagEvCCListPage > 0 {
			call = call.PageSize(flagEvCCListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing channel connections: %w", err)
		}
		all = append(all, resp.ChannelConnections...)
		if flagEvCCListLimit > 0 && int64(len(all)) >= flagEvCCListLimit {
			all = all[:flagEvCCListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvCCListURI {
		for _, c := range all {
			fmt.Println(c.Name)
		}
		return nil
	}
	if flagEvCCFormat != "" {
		return emitFormatted(all, flagEvCCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "CHANNEL")
	for _, c := range all {
		fmt.Printf("%-40s %s\n", path.Base(c.Name), c.Channel)
	}
	return nil
}
