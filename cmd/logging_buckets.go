package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging buckets (#912) ---

var loggingBucketsCmd = &cobra.Command{Use: "buckets", Short: "Manage log buckets"}

var (
	flagLogBucketDescription      string
	flagLogBucketRetentionDays    int64
	flagLogBucketRestrictedFields string
	flagLogBucketCMEKKey          string
	flagLogBucketEnableAnalytics  bool
	flagLogBucketIndexes          []string
)

var (
	loggingBucketsCreateCmd = &cobra.Command{
		Use: "create BUCKET_ID", Short: "Create a log bucket",
		Args: cobra.ExactArgs(1), RunE: runLogBucketCreate,
	}
	loggingBucketsDeleteCmd = &cobra.Command{
		Use: "delete BUCKET_ID", Short: "Delete a log bucket",
		Args: cobra.ExactArgs(1), RunE: runLogBucketDelete,
	}
	loggingBucketsDescribeCmd = &cobra.Command{
		Use: "describe BUCKET_ID", Short: "Describe a log bucket",
		Args: cobra.ExactArgs(1), RunE: runLogBucketDescribe,
	}
	loggingBucketsListCmd = &cobra.Command{
		Use: "list", Short: "List log buckets",
		Args: cobra.NoArgs, RunE: runLogBucketList,
	}
	loggingBucketsUpdateCmd = &cobra.Command{
		Use: "update BUCKET_ID", Short: "Update a log bucket",
		Args: cobra.ExactArgs(1), RunE: runLogBucketUpdate,
	}
	loggingBucketsUndeleteCmd = &cobra.Command{
		Use: "undelete BUCKET_ID", Short: "Undelete a log bucket",
		Args: cobra.ExactArgs(1), RunE: runLogBucketUndelete,
	}
)

func parseBucketIndexes(vals []string) ([]*logging.IndexConfig, error) {
	if len(vals) == 0 {
		return nil, nil
	}
	out := make([]*logging.IndexConfig, 0, len(vals))
	for _, v := range vals {
		spec := &logging.IndexConfig{}
		for _, kv := range strings.Split(v, ",") {
			k, val, ok := strings.Cut(strings.TrimSpace(kv), "=")
			if !ok {
				return nil, fmt.Errorf("invalid --index entry %q; expected key=value pairs", v)
			}
			switch strings.ToLower(k) {
			case "fieldpath":
				spec.FieldPath = val
			case "type":
				spec.Type = val
			default:
				return nil, fmt.Errorf("unknown --index key %q", k)
			}
		}
		if spec.FieldPath == "" || spec.Type == "" {
			return nil, fmt.Errorf("--index requires fieldPath and type: got %q", v)
		}
		out = append(out, spec)
	}
	return out, nil
}

func bucketFromFlags(body *logging.LogBucket) error {
	if flagLogBucketDescription != "" {
		body.Description = flagLogBucketDescription
	}
	if flagLogBucketRetentionDays > 0 {
		body.RetentionDays = flagLogBucketRetentionDays
	}
	if flagLogBucketRestrictedFields != "" {
		body.RestrictedFields = splitCSV(flagLogBucketRestrictedFields)
	}
	if flagLogBucketCMEKKey != "" {
		body.CmekSettings = &logging.CmekSettings{KmsKeyName: flagLogBucketCMEKKey}
	}
	if flagLogBucketEnableAnalytics {
		body.AnalyticsEnabled = true
	}
	idx, err := parseBucketIndexes(flagLogBucketIndexes)
	if err != nil {
		return err
	}
	if idx != nil {
		body.IndexConfigs = idx
	}
	return nil
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func runLogBucketCreate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	body := &logging.LogBucket{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	if err := bucketFromFlags(body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, loggingLocation())
	var got *logging.LogBucket
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.Buckets.Create(loc, body).BucketId(args[0]).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Buckets.Create(loc, body).BucketId(args[0]).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Buckets.Create(loc, body).BucketId(args[0]).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Buckets.Create(loc, body).BucketId(args[0]).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("creating log bucket: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogBucketDelete(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "buckets", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	switch loggingScope(parent) {
	case "projects":
		_, err = svc.Projects.Locations.Buckets.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.Locations.Buckets.Delete(name).Context(ctx).Do()
	case "organizations":
		_, err = svc.Organizations.Locations.Buckets.Delete(name).Context(ctx).Do()
	case "billingAccounts":
		_, err = svc.BillingAccounts.Locations.Buckets.Delete(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("deleting log bucket: %w", err)
	}
	fmt.Printf("Deleted log bucket [%s].\n", args[0])
	return nil
}

func runLogBucketDescribe(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "buckets", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogBucket
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Buckets.Get(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Buckets.Get(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Buckets.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing log bucket: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogBucketList(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, loggingLocation())
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.LogBucket
	pageToken := ""
	for {
		var (
			page []*logging.LogBucket
			next string
		)
		switch loggingScope(parent) {
		case "projects":
			call := svc.Projects.Locations.Buckets.List(loc).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log buckets: %w", err)
			}
			page = resp.Buckets
			next = resp.NextPageToken
		case "folders":
			call := svc.Folders.Locations.Buckets.List(loc).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log buckets: %w", err)
			}
			page = resp.Buckets
			next = resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Locations.Buckets.List(loc).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log buckets: %w", err)
			}
			page = resp.Buckets
			next = resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Locations.Buckets.List(loc).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log buckets: %w", err)
			}
			page = resp.Buckets
			next = resp.NextPageToken
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
	fmt.Printf("%-40s %-10s %s\n", "NAME", "RETENTION", "STATE")
	for _, b := range all {
		fmt.Printf("%-40s %-10s %s\n",
			loggingBasename(b.Name),
			strconv.FormatInt(b.RetentionDays, 10),
			b.LifecycleState)
	}
	return nil
}

func runLogBucketUpdate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "buckets", args[0])
	body := &logging.LogBucket{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	if err := bucketFromFlags(body); err != nil {
		return err
	}
	mask := loggingResolveMask(body)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogBucket
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Locations.Buckets.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Buckets.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Buckets.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Buckets.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("updating log bucket: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogBucketUndelete(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, loggingLocation(), "buckets", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	req := &logging.UndeleteBucketRequest{}
	switch loggingScope(parent) {
	case "projects":
		_, err = svc.Projects.Locations.Buckets.Undelete(name, req).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.Locations.Buckets.Undelete(name, req).Context(ctx).Do()
	case "organizations":
		_, err = svc.Organizations.Locations.Buckets.Undelete(name, req).Context(ctx).Do()
	case "billingAccounts":
		_, err = svc.BillingAccounts.Locations.Buckets.Undelete(name, req).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("undeleting log bucket: %w", err)
	}
	fmt.Printf("Undeleted log bucket [%s].\n", args[0])
	return nil
}

func init() {
	all := []*cobra.Command{loggingBucketsCreateCmd, loggingBucketsDeleteCmd, loggingBucketsDescribeCmd,
		loggingBucketsListCmd, loggingBucketsUpdateCmd, loggingBucketsUndeleteCmd}
	addLogScopeFlags(all...)
	addLogLocationFlag(all...)
	addLogFormatFlag(loggingBucketsCreateCmd, loggingBucketsDescribeCmd, loggingBucketsListCmd, loggingBucketsUpdateCmd)
	addLogPageSizeFlag(loggingBucketsListCmd)
	for _, c := range []*cobra.Command{loggingBucketsCreateCmd, loggingBucketsUpdateCmd} {
		c.Flags().StringVar(&flagLogBucketDescription, "description", "", "A textual description for the bucket")
		c.Flags().Int64Var(&flagLogBucketRetentionDays, "retention-days", 0, "Retention period in days")
		c.Flags().StringVar(&flagLogBucketRestrictedFields, "restricted-fields", "", "Comma-separated list of restricted field paths")
		c.Flags().StringVar(&flagLogBucketCMEKKey, "cmek-kms-key-name", "", "CMEK KMS key name for this bucket")
		c.Flags().BoolVar(&flagLogBucketEnableAnalytics, "enable-analytics", false, "Enable Log Analytics for this bucket")
		c.Flags().StringSliceVar(&flagLogBucketIndexes, "index", nil, "Bucket index (fieldPath=X,type=Y); may repeat")
		c.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogBucket body")
	}
	loggingBucketsUpdateCmd.Flags().StringVar(&flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	loggingBucketsCmd.AddCommand(all...)
	loggingCmd.AddCommand(loggingBucketsCmd)
}
