package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcChannelConnections() {
	for _, c := range []*cobra.Command{grp.evCCCreateCmd, grp.evCCDeleteCmd, grp.evCCDescribeCmd, grp.evCCListCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvCCLocation, true)
	}
	grp.evCCCreateCmd.Flags().StringVar(&grp.flagEvCCConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ChannelConnection message body (required)")
	_ = grp.evCCCreateCmd.MarkFlagRequired("config-file")
	for _, c := range []*cobra.Command{grp.evCCCreateCmd, grp.evCCDeleteCmd} {
		c.Flags().BoolVar(&grp.flagEvCCAsync, "async", false, "Return the long-running operation without waiting")
	}
	grp.evCCDescribeCmd.Flags().StringVar(&grp.flagEvCCFormat, "format", "", "Output format")
	grp.evCCListCmd.Flags().StringVar(&grp.flagEvCCFormat, "format", "", "Output format")
	grp.evCCListCmd.Flags().Int64Var(&grp.flagEvCCListPage, "page-size", 0, "Page size")
	grp.evCCListCmd.Flags().Int64Var(&grp.flagEvCCListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evCCListCmd.Flags().BoolVar(&grp.flagEvCCListURI, "uri", false, "Print resource names only")

	grp.eventarcChannelConnCmd.AddCommand(grp.evCCCreateCmd, grp.evCCDeleteCmd, grp.evCCDescribeCmd, grp.evCCListCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcChannelConnCmd)
}

func (grp *eventarcCmds) runEvCCCreate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	cc := &eventarc.ChannelConnection{}
	if err := loadYAMLOrJSONInto(grp.flagEvCCConfigFile, cc); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ChannelConnections.Create(eventarcLocationParent(project, grp.flagEvCCLocation), cc).
		ChannelConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating channel connection: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Create channel connection", args[0], grp.flagEvCCAsync)
}

func (grp *eventarcCmds) runEvCCDelete(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ChannelConnections.Delete(eventarcResourceName("channelConnections", args[0], project, grp.flagEvCCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting channel connection: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Delete channel connection", args[0], grp.flagEvCCAsync)
}

func (grp *eventarcCmds) runEvCCDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	cc, err := svc.Projects.Locations.ChannelConnections.Get(eventarcResourceName("channelConnections", args[0], project, grp.flagEvCCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing channel connection: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, cc, grp.flagEvCCFormat)
}

func (grp *eventarcCmds) runEvCCList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, grp.flagEvCCLocation)
	var all []*eventarc.ChannelConnection
	pageToken := ""
	for {
		call := svc.Projects.Locations.ChannelConnections.List(parent).Context(ctx)
		if grp.flagEvCCListPage > 0 {
			call = call.PageSize(grp.flagEvCCListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing channel connections: %w", err)
		}
		all = append(all, resp.ChannelConnections...)
		if grp.flagEvCCListLimit > 0 && int64(len(all)) >= grp.flagEvCCListLimit {
			all = all[:grp.flagEvCCListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvCCListURI {
		for _, c := range all {
			fmt.Fprintln(grp.inv.stdout, c.Name)
		}
		return nil
	}
	if grp.flagEvCCFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvCCFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-40s %s\n", "NAME", "CHANNEL")
	for _, c := range all {
		fmt.Fprintf(grp.inv.stdout, "%-40s %s\n", path.Base(c.Name), c.Channel)
	}
	return nil
}
