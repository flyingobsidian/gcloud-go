package cli

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcGoogleApiSources() {
	for _, c := range []*cobra.Command{grp.evGASCreateCmd, grp.evGASDeleteCmd, grp.evGASDescribeCmd, grp.evGASListCmd, grp.evGASUpdateCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvGASLocation, true)
	}
	for _, c := range []*cobra.Command{grp.evGASCreateCmd, grp.evGASUpdateCmd} {
		c.Flags().StringVar(&grp.flagEvGASConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the GoogleApiSource message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	grp.evGASUpdateCmd.Flags().StringVar(&grp.flagEvGASUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{grp.evGASCreateCmd, grp.evGASDeleteCmd, grp.evGASUpdateCmd} {
		c.Flags().BoolVar(&grp.flagEvGASAsync, "async", false, "Return the long-running operation without waiting")
	}
	grp.evGASDescribeCmd.Flags().StringVar(&grp.flagEvGASFormat, "format", "", "Output format")
	grp.evGASListCmd.Flags().StringVar(&grp.flagEvGASFormat, "format", "", "Output format")
	grp.evGASListCmd.Flags().Int64Var(&grp.flagEvGASListPage, "page-size", 0, "Page size")
	grp.evGASListCmd.Flags().Int64Var(&grp.flagEvGASListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evGASListCmd.Flags().StringVar(&grp.flagEvGASListFilter, "filter", "", "Server-side filter expression")
	grp.evGASListCmd.Flags().BoolVar(&grp.flagEvGASListURI, "uri", false, "Print resource names only")

	grp.eventarcGASCmd.AddCommand(grp.evGASCreateCmd, grp.evGASDeleteCmd, grp.evGASDescribeCmd, grp.evGASListCmd, grp.evGASUpdateCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcGASCmd)
}

func (grp *eventarcCmds) runEvGASCreate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	g := &eventarc.GoogleApiSource{}
	if err := loadYAMLOrJSONInto(grp.flagEvGASConfigFile, g); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoogleApiSources.Create(eventarcLocationParent(project, grp.flagEvGASLocation), g).
		GoogleApiSourceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating Google API source: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Create Google API source", args[0], grp.flagEvGASAsync)
}

func (grp *eventarcCmds) runEvGASDelete(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoogleApiSources.Delete(eventarcResourceName("googleApiSources", args[0], project, grp.flagEvGASLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting Google API source: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Delete Google API source", args[0], grp.flagEvGASAsync)
}

func (grp *eventarcCmds) runEvGASDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	g, err := svc.Projects.Locations.GoogleApiSources.Get(eventarcResourceName("googleApiSources", args[0], project, grp.flagEvGASLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Google API source: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, g, grp.flagEvGASFormat)
}

func (grp *eventarcCmds) runEvGASList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, grp.flagEvGASLocation)
	var all []*eventarc.GoogleApiSource
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoogleApiSources.List(parent).Context(ctx)
		if grp.flagEvGASListFilter != "" {
			call = call.Filter(grp.flagEvGASListFilter)
		}
		if grp.flagEvGASListPage > 0 {
			call = call.PageSize(grp.flagEvGASListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing Google API sources: %w", err)
		}
		all = append(all, resp.GoogleApiSources...)
		if grp.flagEvGASListLimit > 0 && int64(len(all)) >= grp.flagEvGASListLimit {
			all = all[:grp.flagEvGASListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvGASListURI {
		for _, g := range all {
			fmt.Fprintln(grp.inv.stdout, g.Name)
		}
		return nil
	}
	if grp.flagEvGASFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvGASFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-40s %s\n", "NAME", "DESTINATION")
	for _, g := range all {
		fmt.Fprintf(grp.inv.stdout, "%-40s %s\n", path.Base(g.Name), g.Destination)
	}
	return nil
}

func (grp *eventarcCmds) runEvGASUpdate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	g := &eventarc.GoogleApiSource{}
	if err := loadYAMLOrJSONInto(grp.flagEvGASConfigFile, g); err != nil {
		return err
	}
	mask := grp.flagEvGASUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(g))
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoogleApiSources.Patch(eventarcResourceName("googleApiSources", args[0], project, grp.flagEvGASLocation), g).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating Google API source: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Update Google API source", args[0], grp.flagEvGASAsync)
}
