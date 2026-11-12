package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging buckets (#912) ---

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

func (g *loggingCmds) bucketFromFlags(body *logging.LogBucket) error {
	if g.flagLogBucketDescription != "" {
		body.Description = g.flagLogBucketDescription
	}
	if g.flagLogBucketRetentionDays > 0 {
		body.RetentionDays = g.flagLogBucketRetentionDays
	}
	if g.flagLogBucketRestrictedFields != "" {
		body.RestrictedFields = splitCSV(g.flagLogBucketRestrictedFields)
	}
	if g.flagLogBucketCMEKKey != "" {
		body.CmekSettings = &logging.CmekSettings{KmsKeyName: g.flagLogBucketCMEKKey}
	}
	if g.flagLogBucketEnableAnalytics {
		body.AnalyticsEnabled = true
	}
	idx, err := parseBucketIndexes(g.flagLogBucketIndexes)
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

func (g *loggingCmds) runLogBucketCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	body := &logging.LogBucket{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	if err := g.bucketFromFlags(body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, g.loggingLocation())
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
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogBucketDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "buckets", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Deleted log bucket [%s].\n", args[0])
	return nil
}

func (g *loggingCmds) runLogBucketDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "buckets", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogBucketList(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	loc := loggingLocationParent(parent, g.loggingLocation())
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %-10s %s\n", "NAME", "RETENTION", "STATE")
	for _, b := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %-10s %s\n",
			loggingBasename(b.Name),
			strconv.FormatInt(b.RetentionDays, 10),
			b.LifecycleState)
	}
	return nil
}

func (g *loggingCmds) runLogBucketUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "buckets", args[0])
	body := &logging.LogBucket{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	if err := g.bucketFromFlags(body); err != nil {
		return err
	}
	mask := g.loggingResolveMask(body)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogBucketUndelete(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := loggingLocationChildName(parent, g.loggingLocation(), "buckets", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Undeleted log bucket [%s].\n", args[0])
	return nil
}

func (g *loggingCmds) initLoggingBuckets() {
	all := []*cobra.Command{g.loggingBucketsCreateCmd, g.loggingBucketsDeleteCmd, g.loggingBucketsDescribeCmd,
		g.loggingBucketsListCmd, g.loggingBucketsUpdateCmd, g.loggingBucketsUndeleteCmd}
	g.addLogScopeFlags(all...)
	g.addLogLocationFlag(all...)
	g.addLogFormatFlag(g.loggingBucketsCreateCmd, g.loggingBucketsDescribeCmd, g.loggingBucketsListCmd, g.loggingBucketsUpdateCmd)
	g.addLogPageSizeFlag(g.loggingBucketsListCmd)
	for _, c := range []*cobra.Command{g.loggingBucketsCreateCmd, g.loggingBucketsUpdateCmd} {
		c.Flags().StringVar(&g.flagLogBucketDescription, "description", "", "A textual description for the bucket")
		c.Flags().Int64Var(&g.flagLogBucketRetentionDays, "retention-days", 0, "Retention period in days")
		c.Flags().StringVar(&g.flagLogBucketRestrictedFields, "restricted-fields", "", "Comma-separated list of restricted field paths")
		c.Flags().StringVar(&g.flagLogBucketCMEKKey, "cmek-kms-key-name", "", "CMEK KMS key name for this bucket")
		c.Flags().BoolVar(&g.flagLogBucketEnableAnalytics, "enable-analytics", false, "Enable Log Analytics for this bucket")
		c.Flags().StringSliceVar(&g.flagLogBucketIndexes, "index", nil, "Bucket index (fieldPath=X,type=Y); may repeat")
		c.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogBucket body")
	}
	g.loggingBucketsUpdateCmd.Flags().StringVar(&g.flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	g.loggingBucketsCmd.AddCommand(all...)
	g.loggingCmd.AddCommand(g.loggingBucketsCmd)
}
