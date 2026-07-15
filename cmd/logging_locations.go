package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging locations (#914) ---
// --- gcloud logging logs (#915) ---
// --- gcloud logging resource-descriptors (#919) ---

// locations subgroup

var loggingLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Cloud Logging locations"}

var (
	loggingLocationsDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Cloud Logging location",
		Args: cobra.ExactArgs(1), RunE: runLogLocationDescribe,
	}
	loggingLocationsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Logging locations",
		Args: cobra.NoArgs, RunE: runLogLocationList,
	}
)

func runLogLocationDescribe(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "locations", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(got, flagLogFormat)
}

func runLogLocationList(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
			if flagLogFilter != "" {
				call = call.Filter(flagLogFilter)
			}
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogFilter != "" {
				call = call.Filter(flagLogFilter)
			}
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogFilter != "" {
				call = call.Filter(flagLogFilter)
			}
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogFilter != "" {
				call = call.Filter(flagLogFilter)
			}
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION_ID", "DISPLAY_NAME")
	for _, l := range all {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// logs subgroup

var loggingLogsCmd = &cobra.Command{Use: "logs", Short: "Manage logs"}

var (
	loggingLogsDeleteCmd = &cobra.Command{
		Use: "delete LOG", Short: "Delete a log and all its entries",
		Args: cobra.ExactArgs(1), RunE: runLogLogDelete,
	}
	loggingLogsListCmd = &cobra.Command{
		Use: "list", Short: "List logs in a resource",
		Args: cobra.NoArgs, RunE: runLogLogList,
	}
)

func runLogLogDelete(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "logs", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	fmt.Printf("Deleted log [%s].\n", args[0])
	return nil
}

func runLogLogList(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	for _, n := range all {
		fmt.Println(n)
	}
	return nil
}

// resource-descriptors subgroup

var loggingResourceDescriptorsCmd = &cobra.Command{Use: "resource-descriptors", Short: "Manage monitored resource descriptors"}

var (
	loggingResourceDescriptorsListCmd = &cobra.Command{
		Use: "list", Short: "List monitored resource descriptors",
		Args: cobra.NoArgs, RunE: runLogRDList,
	}
	loggingResourceDescriptorsDescribeCmd = &cobra.Command{
		Use: "describe TYPE", Short: "Describe a monitored resource descriptor",
		Args: cobra.ExactArgs(1), RunE: runLogRDDescribe,
	}
)

func runLogRDList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.MonitoredResourceDescriptor
	pageToken := ""
	for {
		call := svc.MonitoredResourceDescriptors.List().Context(ctx)
		if flagLogPageSize > 0 {
			call = call.PageSize(flagLogPageSize)
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
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	fmt.Printf("%-30s %s\n", "TYPE", "DISPLAY_NAME")
	for _, r := range all {
		fmt.Printf("%-30s %s\n", r.Type, r.DisplayName)
	}
	return nil
}

func runLogRDDescribe(cmd *cobra.Command, args []string) error {
	// The MRD API only exposes list, so we scan for a matching type.
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
				return emitFormatted(r, flagLogFormat)
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return fmt.Errorf("no resource descriptor with type %q", args[0])
}

func init() {
	// locations
	addLogScopeFlags(loggingLocationsDescribeCmd, loggingLocationsListCmd)
	addLogFormatFlag(loggingLocationsDescribeCmd, loggingLocationsListCmd)
	addLogFilterFlag(loggingLocationsListCmd)
	addLogPageSizeFlag(loggingLocationsListCmd)
	loggingLocationsCmd.AddCommand(loggingLocationsDescribeCmd, loggingLocationsListCmd)
	loggingCmd.AddCommand(loggingLocationsCmd)

	// logs
	addLogScopeFlags(loggingLogsDeleteCmd, loggingLogsListCmd)
	addLogFormatFlag(loggingLogsListCmd)
	addLogPageSizeFlag(loggingLogsListCmd)
	loggingLogsCmd.AddCommand(loggingLogsDeleteCmd, loggingLogsListCmd)
	loggingCmd.AddCommand(loggingLogsCmd)

	// resource-descriptors
	addLogFormatFlag(loggingResourceDescriptorsListCmd, loggingResourceDescriptorsDescribeCmd)
	addLogPageSizeFlag(loggingResourceDescriptorsListCmd)
	loggingResourceDescriptorsCmd.AddCommand(loggingResourceDescriptorsListCmd, loggingResourceDescriptorsDescribeCmd)
	loggingCmd.AddCommand(loggingResourceDescriptorsCmd)
}
