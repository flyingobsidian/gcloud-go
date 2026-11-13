package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcChannels() {
	for _, c := range []*cobra.Command{grp.evChanCreateCmd, grp.evChanDeleteCmd, grp.evChanDescribeCmd, grp.evChanListCmd, grp.evChanUpdateCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvChanLocation, true)
	}
	for _, c := range []*cobra.Command{grp.evChanCreateCmd, grp.evChanUpdateCmd} {
		c.Flags().StringVar(&grp.flagEvChanConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Channel message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	grp.evChanUpdateCmd.Flags().StringVar(&grp.flagEvChanUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{grp.evChanCreateCmd, grp.evChanDeleteCmd, grp.evChanUpdateCmd} {
		c.Flags().BoolVar(&grp.flagEvChanAsync, "async", false, "Return the long-running operation without waiting")
	}
	grp.evChanDescribeCmd.Flags().StringVar(&grp.flagEvChanFormat, "format", "", "Output format")
	grp.evChanListCmd.Flags().StringVar(&grp.flagEvChanFormat, "format", "", "Output format")
	grp.evChanListCmd.Flags().Int64Var(&grp.flagEvChanListPage, "page-size", 0, "Page size for API pagination")
	grp.evChanListCmd.Flags().Int64Var(&grp.flagEvChanListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evChanListCmd.Flags().BoolVar(&grp.flagEvChanListURI, "uri", false, "Print resource names only")

	grp.eventarcChannelsCmd.AddCommand(grp.evChanCreateCmd, grp.evChanDeleteCmd, grp.evChanDescribeCmd, grp.evChanListCmd, grp.evChanUpdateCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcChannelsCmd)
}

func (grp *eventarcCmds) runEvChanCreate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ch := &eventarc.Channel{}
	if err := loadYAMLOrJSONInto(grp.flagEvChanConfigFile, ch); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Channels.Create(eventarcLocationParent(project, grp.flagEvChanLocation), ch).
		ChannelId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating channel: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Create channel", args[0], grp.flagEvChanAsync)
}

func (grp *eventarcCmds) runEvChanDelete(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Channels.Delete(eventarcResourceName("channels", args[0], project, grp.flagEvChanLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting channel: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Delete channel", args[0], grp.flagEvChanAsync)
}

func (grp *eventarcCmds) runEvChanDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	ch, err := svc.Projects.Locations.Channels.Get(eventarcResourceName("channels", args[0], project, grp.flagEvChanLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing channel: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, ch, grp.flagEvChanFormat)
}

func (grp *eventarcCmds) runEvChanList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, grp.flagEvChanLocation)
	var all []*eventarc.Channel
	pageToken := ""
	for {
		call := svc.Projects.Locations.Channels.List(parent).Context(ctx)
		if grp.flagEvChanListPage > 0 {
			call = call.PageSize(grp.flagEvChanListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing channels: %w", err)
		}
		all = append(all, resp.Channels...)
		if grp.flagEvChanListLimit > 0 && int64(len(all)) >= grp.flagEvChanListLimit {
			all = all[:grp.flagEvChanListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvChanListURI {
		for _, c := range all {
			fmt.Fprintln(grp.inv.stdout, c.Name)
		}
		return nil
	}
	if grp.flagEvChanFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvChanFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-40s %s\n", "NAME", "STATE")
	for _, c := range all {
		fmt.Fprintf(grp.inv.stdout, "%-40s %s\n", path.Base(c.Name), c.State)
	}
	return nil
}

func (grp *eventarcCmds) runEvChanUpdate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ch := &eventarc.Channel{}
	if err := loadYAMLOrJSONInto(grp.flagEvChanConfigFile, ch); err != nil {
		return err
	}
	mask := grp.flagEvChanUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ch))
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Channels.Patch(eventarcResourceName("channels", args[0], project, grp.flagEvChanLocation), ch).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating channel: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Update channel", args[0], grp.flagEvChanAsync)
}
