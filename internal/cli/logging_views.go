package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging views (#924) ---

func viewBucketParent(parent, location, bucket string) string {
	return loggingLocationChildName(parent, location, "buckets", bucket)
}

func (g *loggingCmds) viewFromFlags(body *logging.LogView) {
	if g.flagLogViewFilter != "" {
		body.Filter = g.flagLogViewFilter
	}
	if g.flagLogViewDescription != "" {
		body.Description = g.flagLogViewDescription
	}
}

func (g *loggingCmds) runLogViewCreate(cmd *cobra.Command, args []string) error {
	if g.flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	body := &logging.LogView{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	g.viewFromFlags(body)
	bp := viewBucketParent(parent, g.loggingLocation(), g.flagLogViewBucket)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogViewDelete(cmd *cobra.Command, args []string) error {
	if g.flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	bp := viewBucketParent(parent, g.loggingLocation(), g.flagLogViewBucket)
	name := loggingChildName(bp, "views", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Deleted log view [%s].\n", args[0])
	return nil
}

func (g *loggingCmds) runLogViewDescribe(cmd *cobra.Command, args []string) error {
	if g.flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	bp := viewBucketParent(parent, g.loggingLocation(), g.flagLogViewBucket)
	name := loggingChildName(bp, "views", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogViewList(cmd *cobra.Command, args []string) error {
	if g.flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	bp := viewBucketParent(parent, g.loggingLocation(), g.flagLogViewBucket)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "FILTER")
	for _, v := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", loggingBasename(v.Name), v.Filter)
	}
	return nil
}

func (g *loggingCmds) runLogViewUpdate(cmd *cobra.Command, args []string) error {
	if g.flagLogViewBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	bp := viewBucketParent(parent, g.loggingLocation(), g.flagLogViewBucket)
	name := loggingChildName(bp, "views", args[0])
	body := &logging.LogView{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	g.viewFromFlags(body)
	mask := g.loggingResolveMask(body)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) initLoggingViews() {
	all := []*cobra.Command{g.loggingViewsCreateCmd, g.loggingViewsDeleteCmd, g.loggingViewsDescribeCmd,
		g.loggingViewsListCmd, g.loggingViewsUpdateCmd}
	g.addLogScopeFlags(all...)
	g.addLogLocationFlag(all...)
	g.addLogFormatFlag(all...)
	g.addLogPageSizeFlag(g.loggingViewsListCmd)
	for _, c := range all {
		c.Flags().StringVar(&g.flagLogViewBucket, "bucket", "", "Bucket ID that owns the view (required)")
	}
	for _, c := range []*cobra.Command{g.loggingViewsCreateCmd, g.loggingViewsUpdateCmd} {
		c.Flags().StringVar(&g.flagLogViewFilter, "log-filter", "", "Filter that restricts entries visible in this view")
		c.Flags().StringVar(&g.flagLogViewDescription, "description", "", "A textual description for the view")
		c.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogView body")
	}
	g.loggingViewsUpdateCmd.Flags().StringVar(&g.flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	g.loggingViewsCmd.AddCommand(all...)
	g.loggingCmd.AddCommand(g.loggingViewsCmd)
}
