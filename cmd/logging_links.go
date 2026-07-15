package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging links (#913) ---

var loggingLinksCmd = &cobra.Command{Use: "links", Short: "Manage log bucket links"}

var (
	flagLogLinkBucket      string
	flagLogLinkDescription string
)

var (
	loggingLinksCreateCmd = &cobra.Command{
		Use: "create LINK_ID", Short: "Create a link on a log bucket",
		Args: cobra.ExactArgs(1), RunE: runLogLinkCreate,
	}
	loggingLinksDeleteCmd = &cobra.Command{
		Use: "delete LINK_ID", Short: "Delete a bucket link",
		Args: cobra.ExactArgs(1), RunE: runLogLinkDelete,
	}
	loggingLinksDescribeCmd = &cobra.Command{
		Use: "describe LINK_ID", Short: "Describe a bucket link",
		Args: cobra.ExactArgs(1), RunE: runLogLinkDescribe,
	}
	loggingLinksListCmd = &cobra.Command{
		Use: "list", Short: "List bucket links",
		Args: cobra.NoArgs, RunE: runLogLinkList,
	}
)

func logLinkBucketParent(parent, location, bucket string) string {
	return loggingLocationChildName(parent, location, "buckets", bucket)
}

func runLogLinkCreate(cmd *cobra.Command, args []string) error {
	if flagLogLinkBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	body := &logging.Link{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	if flagLogLinkDescription != "" {
		body.Description = flagLogLinkDescription
	}
	bp := logLinkBucketParent(parent, loggingLocation(), flagLogLinkBucket)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(op, flagLogFormat)
}

func runLogLinkDelete(cmd *cobra.Command, args []string) error {
	if flagLogLinkBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	bp := logLinkBucketParent(parent, loggingLocation(), flagLogLinkBucket)
	name := loggingChildName(bp, "links", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(op, flagLogFormat)
}

func runLogLinkDescribe(cmd *cobra.Command, args []string) error {
	if flagLogLinkBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	bp := logLinkBucketParent(parent, loggingLocation(), flagLogLinkBucket)
	name := loggingChildName(bp, "links", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
	return emitFormatted(got, flagLogFormat)
}

func runLogLinkList(cmd *cobra.Command, args []string) error {
	if flagLogLinkBucket == "" {
		return fmt.Errorf("--bucket is required")
	}
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	bp := logLinkBucketParent(parent, loggingLocation(), flagLogLinkBucket)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
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
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, l := range all {
		fmt.Printf("%-40s %s\n", loggingBasename(l.Name), l.LifecycleState)
	}
	return nil
}

func init() {
	all := []*cobra.Command{loggingLinksCreateCmd, loggingLinksDeleteCmd, loggingLinksDescribeCmd, loggingLinksListCmd}
	addLogScopeFlags(all...)
	addLogLocationFlag(all...)
	addLogFormatFlag(all...)
	addLogPageSizeFlag(loggingLinksListCmd)
	for _, c := range all {
		c.Flags().StringVar(&flagLogLinkBucket, "bucket", "", "Bucket ID that owns the link (required)")
	}
	loggingLinksCreateCmd.Flags().StringVar(&flagLogLinkDescription, "description", "", "A textual description for the link")
	loggingLinksCreateCmd.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the Link body")
	loggingLinksCmd.AddCommand(all...)
	loggingCmd.AddCommand(loggingLinksCmd)
}
