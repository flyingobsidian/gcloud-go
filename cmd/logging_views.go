package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging views (#924) ---

var loggingViewsCmd = &cobra.Command{Use: "views", Short: "Manage log views"}

var (
	flagLogViewBucket      string
	flagLogViewFilter      string
	flagLogViewDescription string
)

var (
	loggingViewsCreateCmd = &cobra.Command{
		Use: "create VIEW_ID", Short: "Create a log view",
		Args: cobra.ExactArgs(1), RunE: runLogViewCreate,
	}
	loggingViewsDeleteCmd = &cobra.Command{
		Use: "delete VIEW_ID", Short: "Delete a log view",
		Args: cobra.ExactArgs(1), RunE: runLogViewDelete,
	}
	loggingViewsDescribeCmd = &cobra.Command{
		Use: "describe VIEW_ID", Short: "Describe a log view",
		Args: cobra.ExactArgs(1), RunE: runLogViewDescribe,
	}
	loggingViewsListCmd = &cobra.Command{
		Use: "list", Short: "List log views",
		Args: cobra.NoArgs, RunE: runLogViewList,
	}
	loggingViewsUpdateCmd = &cobra.Command{
		Use: "update VIEW_ID", Short: "Update a log view",
		Args: cobra.ExactArgs(1), RunE: runLogViewUpdate,
	}
)

func viewBucketParent(parent, location, bucket string) string {
	return loggingLocationChildName(parent, location, "buckets", bucket)
}

func viewFromFlags(body *logging.LogView) {
	if flagLogViewFilter != "" {
		body.Filter = flagLogViewFilter
	}
	if flagLogViewDescription != "" {
		body.Description = flagLogViewDescription
	}
}

func runLogViewCreate(cmd *cobra.Command, args []string) error {
	if flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	body := &logging.LogView{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	viewFromFlags(body)
	bp := viewBucketParent(parent, loggingLocation(), flagLogViewBucket)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogView
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.Buckets.Views.Create(bp, body).ViewId(args[0]).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Buckets.Views.Create(bp, body).ViewId(args[0]).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Buckets.Views.Create(bp, body).ViewId(args[0]).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Buckets.Views.Create(bp, body).ViewId(args[0]).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("creating log view: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogViewDelete(cmd *cobra.Command, args []string) error {
	if flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	bp := viewBucketParent(parent, loggingLocation(), flagLogViewBucket)
	name := loggingChildName(bp, "views", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	switch loggingScope(parent) {
	case "projects":
		_, err = svc.Projects.Locations.Buckets.Views.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.Locations.Buckets.Views.Delete(name).Context(ctx).Do()
	case "organizations":
		_, err = svc.Organizations.Locations.Buckets.Views.Delete(name).Context(ctx).Do()
	case "billingAccounts":
		_, err = svc.BillingAccounts.Locations.Buckets.Views.Delete(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("deleting log view: %w", err)
	}
	fmt.Printf("Deleted log view [%s].\n", args[0])
	return nil
}

func runLogViewDescribe(cmd *cobra.Command, args []string) error {
	if flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	bp := viewBucketParent(parent, loggingLocation(), flagLogViewBucket)
	name := loggingChildName(bp, "views", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogView
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.Buckets.Views.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Buckets.Views.Get(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Buckets.Views.Get(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Buckets.Views.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing log view: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogViewList(cmd *cobra.Command, args []string) error {
	if flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	bp := viewBucketParent(parent, loggingLocation(), flagLogViewBucket)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.LogView
	pageToken := ""
	for {
		var (
			page []*logging.LogView
			next string
		)
		switch loggingScope(parent) {
		case "projects":
			call := svc.Projects.Locations.Buckets.Views.List(bp).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log views: %w", err)
			}
			page, next = resp.Views, resp.NextPageToken
		case "folders":
			call := svc.Folders.Locations.Buckets.Views.List(bp).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log views: %w", err)
			}
			page, next = resp.Views, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Locations.Buckets.Views.List(bp).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log views: %w", err)
			}
			page, next = resp.Views, resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Locations.Buckets.Views.List(bp).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log views: %w", err)
			}
			page, next = resp.Views, resp.NextPageToken
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
	fmt.Printf("%-40s %s\n", "NAME", "FILTER")
	for _, v := range all {
		fmt.Printf("%-40s %s\n", loggingBasename(v.Name), v.Filter)
	}
	return nil
}

func runLogViewUpdate(cmd *cobra.Command, args []string) error {
	if flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	bp := viewBucketParent(parent, loggingLocation(), flagLogViewBucket)
	name := loggingChildName(bp, "views", args[0])
	body := &logging.LogView{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	viewFromFlags(body)
	mask := loggingResolveMask(body)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogView
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.Buckets.Views.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Buckets.Views.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Buckets.Views.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Buckets.Views.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("updating log view: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func init() {
	all := []*cobra.Command{loggingViewsCreateCmd, loggingViewsDeleteCmd, loggingViewsDescribeCmd,
		loggingViewsListCmd, loggingViewsUpdateCmd}
	addLogScopeFlags(all...)
	addLogLocationFlag(all...)
	addLogFormatFlag(all...)
	addLogPageSizeFlag(loggingViewsListCmd)
	for _, c := range all {
		c.Flags().StringVar(&flagLogViewBucket, "bucket", "", "Bucket ID that owns the view (required)")
	}
	for _, c := range []*cobra.Command{loggingViewsCreateCmd, loggingViewsUpdateCmd} {
		c.Flags().StringVar(&flagLogViewFilter, "log-filter", "", "Filter that restricts entries visible in this view")
		c.Flags().StringVar(&flagLogViewDescription, "description", "", "A textual description for the view")
		c.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogView body")
	}
	loggingViewsUpdateCmd.Flags().StringVar(&flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	loggingViewsCmd.AddCommand(all...)
	loggingCmd.AddCommand(loggingViewsCmd)
}
