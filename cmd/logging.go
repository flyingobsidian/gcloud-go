package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging (#912-#924) ---
//
// Real API-backed implementations for Cloud Logging v2
// (logging.googleapis.com). All resource groups accept the standard set of
// scope flags --project (default), --organization, --folder, and
// --billing-account, and route to the matching sub-service:
//   svc.Projects        -> projects/{id}
//   svc.Organizations   -> organizations/{id}
//   svc.Folders         -> folders/{id}
//   svc.BillingAccounts -> billingAccounts/{id}
//
// Groups that live under a location (buckets, links, views, saved-queries,
// scopes, recent-queries, operations) additionally accept --location, which
// defaults to "global" unless noted.
//
// The complex "read/write/copy/tail" data-plane commands remain best-effort
// wrappers of Entries.* Write/List/Copy/Tail so users can drive them with a
// --config-file that contains the full request body.

// --- shared flags ---

// loggingParent returns the resource-container parent for the currently
// selected scope: "projects/{id}", "organizations/{id}", "folders/{id}", or
// "billingAccounts/{id}". If no scope flag is set, falls back to the active
// gcloud project.
func (g *loggingCmds) loggingParent() (string, error) {
	set := 0
	for _, v := range []string{g.flagLogProject, g.flagLogOrg, g.flagLogFolder, g.flagLogBillingAccount} {
		if v != "" {
			set++
		}
	}
	if set > 1 {
		return "", fmt.Errorf("--project, --organization, --folder, and --billing-account are mutually exclusive")
	}
	switch {
	case g.flagLogOrg != "":
		return "organizations/" + strings.TrimPrefix(g.flagLogOrg, "organizations/"), nil
	case g.flagLogFolder != "":
		return "folders/" + strings.TrimPrefix(g.flagLogFolder, "folders/"), nil
	case g.flagLogBillingAccount != "":
		return "billingAccounts/" + strings.TrimPrefix(g.flagLogBillingAccount, "billingAccounts/"), nil
	case g.flagLogProject != "":
		return "projects/" + strings.TrimPrefix(g.flagLogProject, "projects/"), nil
	}
	proj, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + proj, nil
}

// loggingScope returns "projects" | "organizations" | "folders" |
// "billingAccounts" for a parent like "projects/foo".
func loggingScope(parent string) string {
	if i := strings.Index(parent, "/"); i > 0 {
		return parent[:i]
	}
	return ""
}

// loggingLocationParent appends "/locations/{location}" to parent.
func loggingLocationParent(parent, location string) string {
	if location == "" {
		location = "global"
	}
	return parent + "/locations/" + location
}

// loggingLocation returns --location or "global" when unset.
func (g *loggingCmds) loggingLocation() string {
	if g.flagLogLocation != "" {
		return g.flagLogLocation
	}
	return "global"
}

// loggingChildName appends "/{collection}/{id}" to parent, unless id already
// looks like a fully qualified resource name.
func loggingChildName(parent, collection, id string) string {
	if strings.Contains(id, "/") {
		return id
	}
	return parent + "/" + collection + "/" + id
}

// loggingLocationChildName resolves either a fully qualified {parent}/locations/{loc}/{coll}/{id}
// or a bare id (in which case it wraps id with parent + location + collection).
func loggingLocationChildName(parent, location, collection, id string) string {
	if strings.Contains(id, "/") {
		return id
	}
	return loggingLocationParent(parent, location) + "/" + collection + "/" + id
}

// loggingClient returns an initialised logging.Service.
func (g *loggingCmds) loggingClient(ctx context.Context) (*logging.Service, error) {
	return g.inv.clients().Logging(ctx)
}

// loggingResolveMask returns --update-mask or an auto-derived mask from body.
func (g *loggingCmds) loggingResolveMask(body any) string {
	if g.flagLogUpdateMask != "" {
		return g.flagLogUpdateMask
	}
	return joinMask(nonEmptyJSONFields(body))
}

// --- flag helpers ---

func (g *loggingCmds) addLogScopeFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagLogProject, "project", "", "Project ID or projects/{id}")
		c.Flags().StringVar(&g.flagLogOrg, "organization", "", "Organization ID or organizations/{id}")
		c.Flags().StringVar(&g.flagLogFolder, "folder", "", "Folder ID or folders/{id}")
		c.Flags().StringVar(&g.flagLogBillingAccount, "billing-account", "", "Billing Account ID or billingAccounts/{id}")
	}
}

func (g *loggingCmds) addLogFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagLogFormat, "format", "", "Output format")
	}
}

func (g *loggingCmds) addLogFilterFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagLogFilter, "filter", "", "Server-side list filter")
	}
}

func (g *loggingCmds) addLogPageSizeFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().Int64Var(&g.flagLogPageSize, "page-size", 0, "Page size for list requests")
	}
}

func (g *loggingCmds) addLogLocationFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagLogLocation, "location", "global", "Location (defaults to global)")
	}
}

func (g *loggingCmds) addLogOrderByFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagLogOrderBy, "order-by", "", "Server-side ordering expression")
	}
}

// basename returns the trailing element of name.
func loggingBasename(name string) string { return path.Base(name) }

// --- entries commands (data plane) ---
//
// The four data-plane commands (copy/read/write/tail) accept a --config-file
// with the full request body and forward it verbatim to the API. This mirrors
// how our other config-heavy verbs work (SCC, database-migration). For read
// the config file is optional: a positional LOG_FILTER plus --freshness,
// --order and --limit cover the common case, as in Python gcloud.

func (g *loggingCmds) runLoggingCopy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	body := &logging.CopyLogEntriesRequest{}
	if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
		return err
	}
	op, err := svc.Entries.Copy(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("copying log entries: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagLogFormat)
}

// loggingTimestampLayout matches Python gcloud's logging util.FormatTimestamp.
const loggingTimestampLayout = "2006-01-02T15:04:05.000000Z"

// loggingReadFilter mirrors Python gcloud's MakeTimestampFilters and
// JoinFilters: with descending order and a filter that does not mention
// "timestamp", it prepends a lower bound of now minus freshness.
func loggingReadFilter(logFilter, order string, freshness time.Duration, now time.Time) string {
	var clauses []string
	if order == "desc" && !strings.Contains(logFilter, "timestamp") {
		since := now.Add(-freshness).UTC().Format(loggingTimestampLayout)
		clauses = append(clauses, `timestamp>="`+since+`"`)
	}
	if logFilter != "" {
		clauses = append(clauses, logFilter)
	}
	return strings.Join(clauses, " AND ")
}

// logEntriesLister performs a single Entries.List call.
type logEntriesLister func(ctx context.Context, req *logging.ListLogEntriesRequest) (*logging.ListLogEntriesResponse, error)

// loggingListEntries pages through Entries.List until the results run out or
// limit entries have been collected (limit <= 0 means no limit). It never
// returns a nil slice, so an empty result renders as [] in JSON.
func loggingListEntries(ctx context.Context, list logEntriesLister, req *logging.ListLogEntriesRequest, limit int64) ([]*logging.LogEntry, error) {
	if req.PageSize == 0 {
		// The backend caps page size at 1000.
		req.PageSize = 1000
		if limit > 0 && limit < req.PageSize {
			req.PageSize = limit
		}
	}
	entries := []*logging.LogEntry{}
	for {
		resp, err := list(ctx, req)
		if err != nil {
			return nil, err
		}
		entries = append(entries, resp.Entries...)
		if limit > 0 && int64(len(entries)) >= limit {
			return entries[:limit], nil
		}
		if resp.NextPageToken == "" {
			return entries, nil
		}
		req.PageToken = resp.NextPageToken
	}
}

func (g *loggingCmds) runLoggingRead(cmd *cobra.Command, args []string) error {
	order := strings.ToLower(g.flagLogOrder)
	if order != "desc" && order != "asc" {
		return fmt.Errorf("invalid --order %q: must be one of desc, asc", g.flagLogOrder)
	}
	freshness, err := parseGcloudDuration(g.flagLogFreshness)
	if err != nil {
		return fmt.Errorf("invalid --freshness: %w", err)
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	body := &logging.ListLogEntriesRequest{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	// The filter comes from LOG_FILTER, else --filter, else --config-file.
	logFilter := body.Filter
	if g.flagLogFilter != "" {
		logFilter = g.flagLogFilter
	}
	if len(args) > 0 {
		logFilter = args[0]
	}
	body.Filter = loggingReadFilter(logFilter, order, freshness, time.Now())
	if g.flagLogOrderBy != "" {
		body.OrderBy = g.flagLogOrderBy
	}
	if body.OrderBy == "" {
		body.OrderBy = "timestamp " + order
	}
	if g.flagLogPageSize > 0 && body.PageSize == 0 {
		body.PageSize = g.flagLogPageSize
	}
	if len(body.ResourceNames) == 0 {
		p, err := g.loggingParent()
		if err != nil {
			return err
		}
		body.ResourceNames = []string{p}
	}
	list := func(ctx context.Context, req *logging.ListLogEntriesRequest) (*logging.ListLogEntriesResponse, error) {
		return svc.Entries.List(req).Context(ctx).Do()
	}
	entries, err := loggingListEntries(ctx, list, body, g.flagLogLimit)
	if err != nil {
		return fmt.Errorf("reading log entries: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, entries, g.flagLogFormat)
}

func (g *loggingCmds) runLoggingWrite(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	body := &logging.WriteLogEntriesRequest{}
	if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
		return err
	}
	resp, err := svc.Entries.Write(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("writing log entries: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagLogFormat)
}

func (g *loggingCmds) runLoggingTail(cmd *cobra.Command, args []string) error {
	// The Tail RPC is a bidi stream; the generated google-api client models it
	// as a single unary call that returns the initial response window. We
	// forward the request body and print whatever the server sends before
	// closing the stream, which matches how our other snapshot verbs render.
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	body := &logging.TailLogEntriesRequest{}
	if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
		return err
	}
	if len(body.ResourceNames) == 0 {
		p, err := g.loggingParent()
		if err != nil {
			return err
		}
		body.ResourceNames = []string{p}
	}
	resp, err := svc.Entries.Tail(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("tailing log entries: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagLogFormat)
}

func (g *loggingCmds) initLogging() {
	// entries data-plane commands
	for _, c := range []*cobra.Command{g.loggingCopyCmd, g.loggingWriteCmd, g.loggingTailCmd} {
		c.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
		c.Flags().StringVar(&g.flagLogFormat, "format", "", "Output format")
	}
	g.loggingReadCmd.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the request body")
	g.loggingReadCmd.Flags().StringVar(&g.flagLogFormat, "format", "", "Output format")
	g.loggingReadCmd.Flags().StringVar(&g.flagLogFilter, "filter", "", "Server-side log filter (overrides --config-file; LOG_FILTER overrides this)")
	g.loggingReadCmd.Flags().StringVar(&g.flagLogOrderBy, "order-by", "", "Server-side ordering expression (overrides --order)")
	g.loggingReadCmd.Flags().Int64Var(&g.flagLogPageSize, "page-size", 0, "Page size")
	g.loggingReadCmd.Flags().StringVar(&g.flagLogFreshness, "freshness", "1d",
		"Return entries that are not older than this value, e.g. 7d or 1h30m. Works only with desc ordering and filters without a timestamp")
	g.loggingReadCmd.Flags().StringVar(&g.flagLogOrder, "order", "desc", "Ordering of returned log entries based on timestamp: desc or asc")
	g.loggingReadCmd.Flags().Int64Var(&g.flagLogLimit, "limit", 0, "Maximum number of log entries to return (0 = no limit)")
	g.addLogScopeFlags(g.loggingReadCmd, g.loggingTailCmd)
	g.loggingCmd.AddCommand(g.loggingCopyCmd, g.loggingReadCmd, g.loggingWriteCmd, g.loggingTailCmd)

	// Subgroup registration happens in the per-group files.

}

// loggingCmds holds the commands and flag values of one invocation's "logging"
// command group, so invocations don't share state (#2070).
type loggingCmds struct {
	inv *invocation

	loggingCmd                            *cobra.Command
	flagLogProject                        string
	flagLogOrg                            string
	flagLogFolder                         string
	flagLogBillingAccount                 string
	flagLogLocation                       string
	flagLogFormat                         string
	flagLogFilter                         string
	flagLogOrderBy                        string
	flagLogPageSize                       int64
	flagLogConfigFile                     string
	flagLogUpdateMask                     string
	flagLogFreshness                      string
	flagLogOrder                          string
	flagLogLimit                          int64
	loggingCopyCmd                        *cobra.Command
	loggingReadCmd                        *cobra.Command
	loggingWriteCmd                       *cobra.Command
	loggingTailCmd                        *cobra.Command
	loggingBucketsCmd                     *cobra.Command
	flagLogBucketDescription              string
	flagLogBucketRetentionDays            int64
	flagLogBucketRestrictedFields         string
	flagLogBucketCMEKKey                  string
	flagLogBucketEnableAnalytics          bool
	flagLogBucketIndexes                  []string
	loggingBucketsCreateCmd               *cobra.Command
	loggingBucketsDeleteCmd               *cobra.Command
	loggingBucketsDescribeCmd             *cobra.Command
	loggingBucketsListCmd                 *cobra.Command
	loggingBucketsUpdateCmd               *cobra.Command
	loggingBucketsUndeleteCmd             *cobra.Command
	loggingLinksCmd                       *cobra.Command
	flagLogLinkBucket                     string
	flagLogLinkDescription                string
	loggingLinksCreateCmd                 *cobra.Command
	loggingLinksDeleteCmd                 *cobra.Command
	loggingLinksDescribeCmd               *cobra.Command
	loggingLinksListCmd                   *cobra.Command
	loggingLocationsCmd                   *cobra.Command
	loggingLocationsDescribeCmd           *cobra.Command
	loggingLocationsListCmd               *cobra.Command
	loggingLogsCmd                        *cobra.Command
	loggingLogsDeleteCmd                  *cobra.Command
	loggingLogsListCmd                    *cobra.Command
	loggingResourceDescriptorsCmd         *cobra.Command
	loggingResourceDescriptorsListCmd     *cobra.Command
	loggingResourceDescriptorsDescribeCmd *cobra.Command
	loggingMetricsCmd                     *cobra.Command
	flagLogMetricDescription              string
	flagLogMetricFilter                   string
	flagLogMetricBucketName               string
	loggingMetricsCreateCmd               *cobra.Command
	loggingMetricsDeleteCmd               *cobra.Command
	loggingMetricsDescribeCmd             *cobra.Command
	loggingMetricsListCmd                 *cobra.Command
	loggingMetricsUpdateCmd               *cobra.Command
	loggingOperationsCmd                  *cobra.Command
	loggingOperationsCancelCmd            *cobra.Command
	loggingOperationsDescribeCmd          *cobra.Command
	loggingOperationsListCmd              *cobra.Command
	loggingSavedQueriesCmd                *cobra.Command
	flagLogSavedQueryDescription          string
	flagLogSavedQueryDisplayName          string
	flagLogSavedQueryVisibility           string
	flagLogSavedQueryQuery                string
	loggingSavedQueriesCreateCmd          *cobra.Command
	loggingSavedQueriesDeleteCmd          *cobra.Command
	loggingSavedQueriesDescribeCmd        *cobra.Command
	loggingSavedQueriesListCmd            *cobra.Command
	loggingSavedQueriesUpdateCmd          *cobra.Command
	loggingRecentQueriesCmd               *cobra.Command
	loggingRecentQueriesListCmd           *cobra.Command
	loggingScopesCmd                      *cobra.Command
	flagLogScopeDescription               string
	flagLogScopeResourceNames             []string
	loggingScopesCreateCmd                *cobra.Command
	loggingScopesDeleteCmd                *cobra.Command
	loggingScopesDescribeCmd              *cobra.Command
	loggingScopesListCmd                  *cobra.Command
	loggingScopesUpdateCmd                *cobra.Command
	loggingSettingsCmd                    *cobra.Command
	flagLogSettingsKmsKey                 string
	flagLogSettingsStorageLocation        string
	flagLogSettingsDisableDefaultSink     bool
	loggingSettingsDescribeCmd            *cobra.Command
	loggingSettingsUpdateCmd              *cobra.Command
	loggingSinksCmd                       *cobra.Command
	flagLogSinkDestination                string
	flagLogSinkFilter                     string
	flagLogSinkDescription                string
	flagLogSinkDisabled                   bool
	flagLogSinkIncludeChildren            bool
	flagLogSinkUniqueWriter               bool
	flagLogSinkCustomWriter               string
	loggingSinksCreateCmd                 *cobra.Command
	loggingSinksDeleteCmd                 *cobra.Command
	loggingSinksDescribeCmd               *cobra.Command
	loggingSinksListCmd                   *cobra.Command
	loggingSinksUpdateCmd                 *cobra.Command
	loggingViewsCmd                       *cobra.Command
	flagLogViewBucket                     string
	flagLogViewFilter                     string
	flagLogViewDescription                string
	loggingViewsCreateCmd                 *cobra.Command
	loggingViewsDeleteCmd                 *cobra.Command
	loggingViewsDescribeCmd               *cobra.Command
	loggingViewsListCmd                   *cobra.Command
	loggingViewsUpdateCmd                 *cobra.Command
}

// newLoggingCmd returns the "logging" command group for inv.
func newLoggingCmd(inv *invocation) *cobra.Command {
	g := &loggingCmds{inv: inv}
	g.loggingCmd = &cobra.Command{Use: "logging", Short: "Manage Cloud Logging"}
	g.loggingCopyCmd = &cobra.Command{
		Use:   "copy",
		Short: "Copy log entries from a bucket into a destination",
		RunE:  g.runLoggingCopy,
	}
	g.loggingReadCmd = &cobra.Command{
		Use:   "read [LOG_FILTER]",
		Short: "Read log entries",
		Args:  cobra.MaximumNArgs(1),
		RunE:  g.runLoggingRead,
	}
	g.loggingWriteCmd = &cobra.Command{
		Use:   "write",
		Short: "Write a log entry",
		RunE:  g.runLoggingWrite,
	}
	g.loggingTailCmd = &cobra.Command{
		Use:   "tail",
		Short: "Tail log entries in real time",
		RunE:  g.runLoggingTail,
	}
	g.loggingBucketsCmd = &cobra.Command{Use: "buckets", Short: "Manage log buckets"}
	g.loggingBucketsCreateCmd = &cobra.Command{
		Use: "create BUCKET_ID", Short: "Create a log bucket",
		Args: cobra.ExactArgs(1), RunE: g.runLogBucketCreate,
	}
	g.loggingBucketsDeleteCmd = &cobra.Command{
		Use: "delete BUCKET_ID", Short: "Delete a log bucket",
		Args: cobra.ExactArgs(1), RunE: g.runLogBucketDelete,
	}
	g.loggingBucketsDescribeCmd = &cobra.Command{
		Use: "describe BUCKET_ID", Short: "Describe a log bucket",
		Args: cobra.ExactArgs(1), RunE: g.runLogBucketDescribe,
	}
	g.loggingBucketsListCmd = &cobra.Command{
		Use: "list", Short: "List log buckets",
		Args: cobra.NoArgs, RunE: g.runLogBucketList,
	}
	g.loggingBucketsUpdateCmd = &cobra.Command{
		Use: "update BUCKET_ID", Short: "Update a log bucket",
		Args: cobra.ExactArgs(1), RunE: g.runLogBucketUpdate,
	}
	g.loggingBucketsUndeleteCmd = &cobra.Command{
		Use: "undelete BUCKET_ID", Short: "Undelete a log bucket",
		Args: cobra.ExactArgs(1), RunE: g.runLogBucketUndelete,
	}
	g.loggingLinksCmd = &cobra.Command{Use: "links", Short: "Manage log bucket links"}
	g.loggingLinksCreateCmd = &cobra.Command{
		Use: "create LINK_ID", Short: "Create a link on a log bucket",
		Args: cobra.ExactArgs(1), RunE: g.runLogLinkCreate,
	}
	g.loggingLinksDeleteCmd = &cobra.Command{
		Use: "delete LINK_ID", Short: "Delete a bucket link",
		Args: cobra.ExactArgs(1), RunE: g.runLogLinkDelete,
	}
	g.loggingLinksDescribeCmd = &cobra.Command{
		Use: "describe LINK_ID", Short: "Describe a bucket link",
		Args: cobra.ExactArgs(1), RunE: g.runLogLinkDescribe,
	}
	g.loggingLinksListCmd = &cobra.Command{
		Use: "list", Short: "List bucket links",
		Args: cobra.NoArgs, RunE: g.runLogLinkList,
	}
	g.loggingLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Cloud Logging locations"}
	g.loggingLocationsDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Cloud Logging location",
		Args: cobra.ExactArgs(1), RunE: g.runLogLocationDescribe,
	}
	g.loggingLocationsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Logging locations",
		Args: cobra.NoArgs, RunE: g.runLogLocationList,
	}
	g.loggingLogsCmd = &cobra.Command{Use: "logs", Short: "Manage logs"}
	g.loggingLogsDeleteCmd = &cobra.Command{
		Use: "delete LOG", Short: "Delete a log and all its entries",
		Args: cobra.ExactArgs(1), RunE: g.runLogLogDelete,
	}
	g.loggingLogsListCmd = &cobra.Command{
		Use: "list", Short: "List logs in a resource",
		Args: cobra.NoArgs, RunE: g.runLogLogList,
	}
	g.loggingResourceDescriptorsCmd = &cobra.Command{Use: "resource-descriptors", Short: "Manage monitored resource descriptors"}
	g.loggingResourceDescriptorsListCmd = &cobra.Command{
		Use: "list", Short: "List monitored resource descriptors",
		Args: cobra.NoArgs, RunE: g.runLogRDList,
	}
	g.loggingResourceDescriptorsDescribeCmd = &cobra.Command{
		Use: "describe TYPE", Short: "Describe a monitored resource descriptor",
		Args: cobra.ExactArgs(1), RunE: g.runLogRDDescribe,
	}
	g.loggingMetricsCmd = &cobra.Command{Use: "metrics", Short: "Manage logs-based metrics"}
	g.loggingMetricsCreateCmd = &cobra.Command{
		Use: "create METRIC", Short: "Create a logs-based metric",
		Args: cobra.ExactArgs(1), RunE: g.runLogMetricCreate,
	}
	g.loggingMetricsDeleteCmd = &cobra.Command{
		Use: "delete METRIC", Short: "Delete a logs-based metric",
		Args: cobra.ExactArgs(1), RunE: g.runLogMetricDelete,
	}
	g.loggingMetricsDescribeCmd = &cobra.Command{
		Use: "describe METRIC", Short: "Describe a logs-based metric",
		Args: cobra.ExactArgs(1), RunE: g.runLogMetricDescribe,
	}
	g.loggingMetricsListCmd = &cobra.Command{
		Use: "list", Short: "List logs-based metrics",
		Args: cobra.NoArgs, RunE: g.runLogMetricList,
	}
	g.loggingMetricsUpdateCmd = &cobra.Command{
		Use: "update METRIC", Short: "Update a logs-based metric",
		Args: cobra.ExactArgs(1), RunE: g.runLogMetricUpdate,
	}
	g.loggingOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Logging long-running operations"}
	g.loggingOperationsCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a long-running operation",
		Args: cobra.ExactArgs(1), RunE: g.runLogOperationCancel,
	}
	g.loggingOperationsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a long-running operation",
		Args: cobra.ExactArgs(1), RunE: g.runLogOperationDescribe,
	}
	g.loggingOperationsListCmd = &cobra.Command{
		Use: "list", Short: "List long-running operations",
		Args: cobra.NoArgs, RunE: g.runLogOperationList,
	}
	g.loggingSavedQueriesCmd = &cobra.Command{Use: "saved-queries", Short: "Manage saved queries"}
	g.loggingSavedQueriesCreateCmd = &cobra.Command{
		Use: "create SAVED_QUERY_ID", Short: "Create a saved query",
		Args: cobra.ExactArgs(1), RunE: g.runLogSavedQueryCreate,
	}
	g.loggingSavedQueriesDeleteCmd = &cobra.Command{
		Use: "delete SAVED_QUERY_ID", Short: "Delete a saved query",
		Args: cobra.ExactArgs(1), RunE: g.runLogSavedQueryDelete,
	}
	g.loggingSavedQueriesDescribeCmd = &cobra.Command{
		Use: "describe SAVED_QUERY_ID", Short: "Describe a saved query",
		Args: cobra.ExactArgs(1), RunE: g.runLogSavedQueryDescribe,
	}
	g.loggingSavedQueriesListCmd = &cobra.Command{
		Use: "list", Short: "List saved queries",
		Args: cobra.NoArgs, RunE: g.runLogSavedQueryList,
	}
	g.loggingSavedQueriesUpdateCmd = &cobra.Command{
		Use: "update SAVED_QUERY_ID", Short: "Update a saved query",
		Args: cobra.ExactArgs(1), RunE: g.runLogSavedQueryUpdate,
	}
	g.loggingRecentQueriesCmd = &cobra.Command{Use: "recent-queries", Short: "Manage recent queries"}
	g.loggingRecentQueriesListCmd = &cobra.Command{
		Use: "list", Short: "List recent queries",
		Args: cobra.NoArgs, RunE: g.runLogRecentQueryList,
	}
	g.loggingScopesCmd = &cobra.Command{Use: "scopes", Short: "Manage log scopes"}
	g.loggingScopesCreateCmd = &cobra.Command{
		Use: "create LOG_SCOPE_ID", Short: "Create a log scope",
		Args: cobra.ExactArgs(1), RunE: g.runLogScopeCreate,
	}
	g.loggingScopesDeleteCmd = &cobra.Command{
		Use: "delete LOG_SCOPE_ID", Short: "Delete a log scope",
		Args: cobra.ExactArgs(1), RunE: g.runLogScopeDelete,
	}
	g.loggingScopesDescribeCmd = &cobra.Command{
		Use: "describe LOG_SCOPE_ID", Short: "Describe a log scope",
		Args: cobra.ExactArgs(1), RunE: g.runLogScopeDescribe,
	}
	g.loggingScopesListCmd = &cobra.Command{
		Use: "list", Short: "List log scopes",
		Args: cobra.NoArgs, RunE: g.runLogScopeList,
	}
	g.loggingScopesUpdateCmd = &cobra.Command{
		Use: "update LOG_SCOPE_ID", Short: "Update a log scope",
		Args: cobra.ExactArgs(1), RunE: g.runLogScopeUpdate,
	}
	g.loggingSettingsCmd = &cobra.Command{Use: "settings", Short: "Manage Logs Router settings"}
	g.loggingSettingsDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe Logs Router settings",
		Args: cobra.NoArgs, RunE: g.runLogSettingsDescribe,
	}
	g.loggingSettingsUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update Logs Router settings",
		Args: cobra.NoArgs, RunE: g.runLogSettingsUpdate,
	}
	g.loggingSinksCmd = &cobra.Command{Use: "sinks", Short: "Manage log sinks"}
	g.loggingSinksCreateCmd = &cobra.Command{
		Use: "create SINK", Short: "Create a log sink",
		Args: cobra.ExactArgs(1), RunE: g.runLogSinkCreate,
	}
	g.loggingSinksDeleteCmd = &cobra.Command{
		Use: "delete SINK", Short: "Delete a log sink",
		Args: cobra.ExactArgs(1), RunE: g.runLogSinkDelete,
	}
	g.loggingSinksDescribeCmd = &cobra.Command{
		Use: "describe SINK", Short: "Describe a log sink",
		Args: cobra.ExactArgs(1), RunE: g.runLogSinkDescribe,
	}
	g.loggingSinksListCmd = &cobra.Command{
		Use: "list", Short: "List log sinks",
		Args: cobra.NoArgs, RunE: g.runLogSinkList,
	}
	g.loggingSinksUpdateCmd = &cobra.Command{
		Use: "update SINK", Short: "Update a log sink",
		Args: cobra.ExactArgs(1), RunE: g.runLogSinkUpdate,
	}
	g.loggingViewsCmd = &cobra.Command{Use: "views", Short: "Manage log views"}
	g.loggingViewsCreateCmd = &cobra.Command{
		Use: "create VIEW_ID", Short: "Create a log view",
		Args: cobra.ExactArgs(1), RunE: g.runLogViewCreate,
	}
	g.loggingViewsDeleteCmd = &cobra.Command{
		Use: "delete VIEW_ID", Short: "Delete a log view",
		Args: cobra.ExactArgs(1), RunE: g.runLogViewDelete,
	}
	g.loggingViewsDescribeCmd = &cobra.Command{
		Use: "describe VIEW_ID", Short: "Describe a log view",
		Args: cobra.ExactArgs(1), RunE: g.runLogViewDescribe,
	}
	g.loggingViewsListCmd = &cobra.Command{
		Use: "list", Short: "List log views",
		Args: cobra.NoArgs, RunE: g.runLogViewList,
	}
	g.loggingViewsUpdateCmd = &cobra.Command{
		Use: "update VIEW_ID", Short: "Update a log view",
		Args: cobra.ExactArgs(1), RunE: g.runLogViewUpdate,
	}
	g.initLogging()
	g.initLoggingBuckets()
	g.initLoggingLinks()
	g.initLoggingLocations()
	g.initLoggingMetrics()
	g.initLoggingOperations()
	g.initLoggingQueries()
	g.initLoggingSettings()
	g.initLoggingSinks()
	g.initLoggingViews()
	return g.loggingCmd
}
