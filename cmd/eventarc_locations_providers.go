package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcLocationsProviders() {
	grp.eventarcLocationsListCmd.Flags().StringVar(&grp.flagEvLocFormat, "format", "", "Output format")
	grp.eventarcLocationsListCmd.Flags().Int64Var(&grp.flagEvLocListPage, "page-size", 0, "Page size")
	grp.eventarcLocationsListCmd.Flags().Int64Var(&grp.flagEvLocListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.eventarcLocationsCmd.AddCommand(grp.eventarcLocationsListCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcLocationsCmd)

	for _, c := range []*cobra.Command{grp.evProvDescribeCmd, grp.evProvListCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvProvLocation, true)
	}
	grp.evProvDescribeCmd.Flags().StringVar(&grp.flagEvProvFormat, "format", "", "Output format")
	grp.evProvListCmd.Flags().StringVar(&grp.flagEvProvFormat, "format", "", "Output format")
	grp.evProvListCmd.Flags().Int64Var(&grp.flagEvProvListPage, "page-size", 0, "Page size")
	grp.evProvListCmd.Flags().Int64Var(&grp.flagEvProvListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evProvListCmd.Flags().StringVar(&grp.flagEvProvListFilter, "filter", "", "Server-side filter expression")
	grp.evProvListCmd.Flags().BoolVar(&grp.flagEvProvListURI, "uri", false, "Print resource names only")
	grp.eventarcProvidersCmd.AddCommand(grp.evProvDescribeCmd, grp.evProvListCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcProvidersCmd)
}

func (grp *eventarcCmds) runEventarcLocationsList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	var all []*eventarc.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if grp.flagEvLocListPage > 0 {
			call = call.PageSize(grp.flagEvLocListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if grp.flagEvLocListLimit > 0 && int64(len(all)) >= grp.flagEvLocListLimit {
			all = all[:grp.flagEvLocListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvLocFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvLocFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range all {
		fmt.Fprintf(grp.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func (grp *eventarcCmds) runEvProvDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Locations.Providers.Get(eventarcResourceName("providers", args[0], project, grp.flagEvProvLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing provider: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, p, grp.flagEvProvFormat)
}

func (grp *eventarcCmds) runEvProvList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, grp.flagEvProvLocation)
	var all []*eventarc.Provider
	pageToken := ""
	for {
		call := svc.Projects.Locations.Providers.List(parent).Context(ctx)
		if grp.flagEvProvListFilter != "" {
			call = call.Filter(grp.flagEvProvListFilter)
		}
		if grp.flagEvProvListPage > 0 {
			call = call.PageSize(grp.flagEvProvListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing providers: %w", err)
		}
		all = append(all, resp.Providers...)
		if grp.flagEvProvListLimit > 0 && int64(len(all)) >= grp.flagEvProvListLimit {
			all = all[:grp.flagEvProvListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvProvListURI {
		for _, p := range all {
			fmt.Fprintln(grp.inv.stdout, p.Name)
		}
		return nil
	}
	if grp.flagEvProvFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvProvFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, p := range all {
		fmt.Fprintf(grp.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.DisplayName)
	}
	return nil
}
