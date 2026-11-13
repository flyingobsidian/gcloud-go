package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging links (#913) ---

func logLinkBucketParent(parent, location, bucket string) string {
	return loggingLocationChildName(parent, location, "buckets", bucket)
}

func (g *loggingCmds) runLogLinkCreate(cmd *cobra.Command, args []string) error {
	if g.flagLogLinkBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	body := &logging.Link{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	if g.flagLogLinkDescription != "" {
		body.Description = g.flagLogLinkDescription
	}
	bp := logLinkBucketParent(parent, g.loggingLocation(), g.flagLogLinkBucket)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var op *logging.Operation
	switch loggingScope(parent) {
	case "projects":
		op, err = svc.Projects.Locations.Buckets.Links.Create(bp, body).LinkId(args[0]).Context(ctx).Do()
	case "folders":
		op, err = svc.Folders.Locations.Buckets.Links.Create(bp, body).LinkId(args[0]).Context(ctx).Do()
	case "organizations":
		op, err = svc.Organizations.Locations.Buckets.Links.Create(bp, body).LinkId(args[0]).Context(ctx).Do()
	case "billingAccounts":
		op, err = svc.BillingAccounts.Locations.Buckets.Links.Create(bp, body).LinkId(args[0]).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("creating bucket link: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagLogFormat)
}

func (g *loggingCmds) runLogLinkDelete(cmd *cobra.Command, args []string) error {
	if g.flagLogLinkBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	bp := logLinkBucketParent(parent, g.loggingLocation(), g.flagLogLinkBucket)
	name := loggingChildName(bp, "links", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var op *logging.Operation
	switch loggingScope(parent) {
	case "projects":
		op, err = svc.Projects.Locations.Buckets.Links.Delete(name).Context(ctx).Do()
	case "folders":
		op, err = svc.Folders.Locations.Buckets.Links.Delete(name).Context(ctx).Do()
	case "organizations":
		op, err = svc.Organizations.Locations.Buckets.Links.Delete(name).Context(ctx).Do()
	case "billingAccounts":
		op, err = svc.BillingAccounts.Locations.Buckets.Links.Delete(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("deleting bucket link: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagLogFormat)
}

func (g *loggingCmds) runLogLinkDescribe(cmd *cobra.Command, args []string) error {
	if g.flagLogLinkBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	bp := logLinkBucketParent(parent, g.loggingLocation(), g.flagLogLinkBucket)
	name := loggingChildName(bp, "links", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.Link
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.Buckets.Links.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Buckets.Links.Get(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Buckets.Links.Get(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Buckets.Links.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing bucket link: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogLinkList(cmd *cobra.Command, args []string) error {
	if g.flagLogLinkBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	bp := logLinkBucketParent(parent, g.loggingLocation(), g.flagLogLinkBucket)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.Link
	pageToken := ""
	for {
		var (
			page []*logging.Link
			next string
		)
		switch loggingScope(parent) {
		case "projects":
			call := svc.Projects.Locations.Buckets.Links.List(bp).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing bucket links: %w", err)
			}
			page, next = resp.Links, resp.NextPageToken
		case "folders":
			call := svc.Folders.Locations.Buckets.Links.List(bp).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing bucket links: %w", err)
			}
			page, next = resp.Links, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Locations.Buckets.Links.List(bp).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing bucket links: %w", err)
			}
			page, next = resp.Links, resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Locations.Buckets.Links.List(bp).Context(ctx)
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing bucket links: %w", err)
			}
			page, next = resp.Links, resp.NextPageToken
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
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "STATE")
	for _, l := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", loggingBasename(l.Name), l.LifecycleState)
	}
	return nil
}

func (g *loggingCmds) initLoggingLinks() {
	all := []*cobra.Command{g.loggingLinksCreateCmd, g.loggingLinksDeleteCmd, g.loggingLinksDescribeCmd, g.loggingLinksListCmd}
	g.addLogScopeFlags(all...)
	g.addLogLocationFlag(all...)
	g.addLogFormatFlag(all...)
	g.addLogPageSizeFlag(g.loggingLinksListCmd)
	for _, c := range all {
		c.Flags().StringVar(&g.flagLogLinkBucket, "bucket", "", "Bucket ID that owns the link (required)")
	}
	g.loggingLinksCreateCmd.Flags().StringVar(&g.flagLogLinkDescription, "description", "", "A textual description for the link")
	g.loggingLinksCreateCmd.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the Link body")
	g.loggingLinksCmd.AddCommand(all...)
	g.loggingCmd.AddCommand(g.loggingLinksCmd)
}
