package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging locations (#914) ---
// --- gcloud logging logs (#915) ---
// --- gcloud logging resource-descriptors (#919) ---

// locations subgroup

func (g *loggingCmds) runLogLocationDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "locations", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.Location
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Get(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Get(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogLocationList(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.Location
	pageToken := ""
	for {
		var (
			page []*logging.Location
			next string
		)
		switch loggingScope(parent) {
		case "projects":
			call := svc.Projects.Locations.List(parent).Context(ctx)
			if g.flagLogFilter != "" {
				call = call.Filter(g.flagLogFilter)
			}
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing locations: %w", err)
			}
			page, next = resp.Locations, resp.NextPageToken
		case "folders":
			call := svc.Folders.Locations.List(parent).Context(ctx)
			if g.flagLogFilter != "" {
				call = call.Filter(g.flagLogFilter)
			}
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing locations: %w", err)
			}
			page, next = resp.Locations, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Locations.List(parent).Context(ctx)
			if g.flagLogFilter != "" {
				call = call.Filter(g.flagLogFilter)
			}
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing locations: %w", err)
			}
			page, next = resp.Locations, resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Locations.List(parent).Context(ctx)
			if g.flagLogFilter != "" {
				call = call.Filter(g.flagLogFilter)
			}
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing locations: %w", err)
			}
			page, next = resp.Locations, resp.NextPageToken
		default:
			return fmt.Errorf("invalid parent %q", parent)
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-20s %s\n", "LOCATION_ID", "DISPLAY_NAME")
	for _, l := range all {
		fmt.Fprintf(g.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// logs subgroup

func (g *loggingCmds) runLogLogDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "logs", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	switch loggingScope(parent) {
	case "projects":
		_, err = svc.Projects.Logs.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.Logs.Delete(name).Context(ctx).Do()
	case "organizations":
		_, err = svc.Organizations.Logs.Delete(name).Context(ctx).Do()
	case "billingAccounts":
		_, err = svc.BillingAccounts.Logs.Delete(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("deleting log: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted log [%s].\n", args[0])
	return nil
}

func (g *loggingCmds) runLogLogList(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []string
	pageToken := ""
	for {
		var (
			page []string
			next string
		)
		switch loggingScope(parent) {
		case "projects":
			call := svc.Projects.Logs.List(parent).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing logs: %w", err)
			}
			page, next = resp.LogNames, resp.NextPageToken
		case "folders":
			call := svc.Folders.Logs.List(parent).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing logs: %w", err)
			}
			page, next = resp.LogNames, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Logs.List(parent).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing logs: %w", err)
			}
			page, next = resp.LogNames, resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Logs.List(parent).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing logs: %w", err)
			}
			page, next = resp.LogNames, resp.NextPageToken
		default:
			return fmt.Errorf("invalid parent %q", parent)
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	for _, n := range all {
		fmt.Fprintln(g.inv.stdout, n)
	}
	return nil
}

// resource-descriptors subgroup

func (g *loggingCmds) runLogRDList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.MonitoredResourceDescriptor
	pageToken := ""
	for {
		call := svc.MonitoredResourceDescriptors.List().Context(ctx)
		if g.flagLogPageSize > 0 {
			call = call.PageSize(g.flagLogPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing resource descriptors: %w", err)
		}
		all = append(all, resp.ResourceDescriptors...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-30s %s\n", "TYPE", "DISPLAY_NAME")
	for _, r := range all {
		fmt.Fprintf(g.inv.stdout, "%-30s %s\n", r.Type, r.DisplayName)
	}
	return nil
}

func (g *loggingCmds) runLogRDDescribe(cmd *cobra.Command, args []string) error {
	// The MRD API only exposes list, so we scan for a matching type.
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	pageToken := ""
	for {
		call := svc.MonitoredResourceDescriptors.List().Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing resource descriptors: %w", err)
		}
		for _, r := range resp.ResourceDescriptors {
			if r.Type == args[0] {
				return emitFormattedTo(g.inv.stdout, r, g.flagLogFormat)
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return fmt.Errorf("no resource descriptor with type %q", args[0])
}

func (g *loggingCmds) initLoggingLocations() {
	// locations
	g.addLogScopeFlags(g.loggingLocationsDescribeCmd, g.loggingLocationsListCmd)
	g.addLogFormatFlag(g.loggingLocationsDescribeCmd, g.loggingLocationsListCmd)
	g.addLogFilterFlag(g.loggingLocationsListCmd)
	g.addLogPageSizeFlag(g.loggingLocationsListCmd)
	g.loggingLocationsCmd.AddCommand(g.loggingLocationsDescribeCmd, g.loggingLocationsListCmd)
	g.loggingCmd.AddCommand(g.loggingLocationsCmd)

	// logs
	g.addLogScopeFlags(g.loggingLogsDeleteCmd, g.loggingLogsListCmd)
	g.addLogFormatFlag(g.loggingLogsListCmd)
	g.addLogPageSizeFlag(g.loggingLogsListCmd)
	g.loggingLogsCmd.AddCommand(g.loggingLogsDeleteCmd, g.loggingLogsListCmd)
	g.loggingCmd.AddCommand(g.loggingLogsCmd)

	// resource-descriptors
	g.addLogFormatFlag(g.loggingResourceDescriptorsListCmd, g.loggingResourceDescriptorsDescribeCmd)
	g.addLogPageSizeFlag(g.loggingResourceDescriptorsListCmd)
	g.loggingResourceDescriptorsCmd.AddCommand(g.loggingResourceDescriptorsListCmd, g.loggingResourceDescriptorsDescribeCmd)
	g.loggingCmd.AddCommand(g.loggingResourceDescriptorsCmd)
}
