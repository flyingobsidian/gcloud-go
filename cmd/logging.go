package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
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

var loggingCmd = &cobra.Command{Use: "logging", Short: "Manage Cloud Logging"}

// --- shared flags ---

var (
	flagLogProject        string
	flagLogOrg            string
	flagLogFolder         string
	flagLogBillingAccount string
	flagLogLocation       string
	flagLogFormat         string
	flagLogFilter         string
	flagLogOrderBy        string
	flagLogPageSize       int64
	flagLogConfigFile     string
	flagLogUpdateMask     string
)

// loggingParent returns the resource-container parent for the currently
// selected scope: "projects/{id}", "organizations/{id}", "folders/{id}", or
// "billingAccounts/{id}". If no scope flag is set, falls back to the active
// gcloud project.
func loggingParent() (string, error) {
	set := 0
	for _, v := range []string{flagLogProject, flagLogOrg, flagLogFolder, flagLogBillingAccount} {
		if v != "" {
			set++
		}
	}
	if set > 1 {
		return "", fmt.Errorf("--project, --organization, --folder, and --billing-account are mutually exclusive")
	}
	switch {
	case flagLogOrg != "":
		return "organizations/" + strings.TrimPrefix(flagLogOrg, "organizations/"), nil
	case flagLogFolder != "":
		return "folders/" + strings.TrimPrefix(flagLogFolder, "folders/"), nil
	case flagLogBillingAccount != "":
		return "billingAccounts/" + strings.TrimPrefix(flagLogBillingAccount, "billingAccounts/"), nil
	case flagLogProject != "":
		return "projects/" + strings.TrimPrefix(flagLogProject, "projects/"), nil
	}
	proj, err := resolveProject()
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
func loggingLocation() string {
	if flagLogLocation != "" {
		return flagLogLocation
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
func loggingClient(ctx context.Context) (*logging.Service, error) {
	return gcp.LoggingService(ctx, flagAccount)
}

// loggingResolveMask returns --update-mask or an auto-derived mask from body.
func loggingResolveMask(body any) string {
	if flagLogUpdateMask != "" {
		return flagLogUpdateMask
	}
	return joinMask(nonEmptyJSONFields(body))
}

// --- flag helpers ---

func addLogScopeFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagLogProject, "project", "", "Project ID or projects/{id}")
		c.Flags().StringVar(&flagLogOrg, "organization", "", "Organization ID or organizations/{id}")
		c.Flags().StringVar(&flagLogFolder, "folder", "", "Folder ID or folders/{id}")
		c.Flags().StringVar(&flagLogBillingAccount, "billing-account", "", "Billing Account ID or billingAccounts/{id}")
	}
}

func addLogFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagLogFormat, "format", "", "Output format")
	}
}

func addLogFilterFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagLogFilter, "filter", "", "Server-side list filter")
	}
}

func addLogPageSizeFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().Int64Var(&flagLogPageSize, "page-size", 0, "Page size for list requests")
	}
}

func addLogLocationFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagLogLocation, "location", "global", "Location (defaults to global)")
	}
}

func addLogOrderByFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagLogOrderBy, "order-by", "", "Server-side ordering expression")
	}
}

// basename returns the trailing element of name.
func loggingBasename(name string) string { return path.Base(name) }

// --- entries commands (data plane) ---
//
// The four data-plane commands (copy/read/write/tail) accept a --config-file
// with the full request body and forward it verbatim to the API. This mirrors
// how our other config-heavy verbs work (SCC, database-migration).

var (
	loggingCopyCmd = &cobra.Command{
		Use:   "copy",
		Short: "Copy log entries from a bucket into a destination",
		RunE:  runLoggingCopy,
	}
	loggingReadCmd = &cobra.Command{
		Use:   "read",
		Short: "Read log entries",
		RunE:  runLoggingRead,
	}
	loggingWriteCmd = &cobra.Command{
		Use:   "write",
		Short: "Write a log entry",
		RunE:  runLoggingWrite,
	}
	loggingTailCmd = &cobra.Command{
		Use:   "tail",
		Short: "Tail log entries in real time",
		RunE:  runLoggingTail,
	}
)

func runLoggingCopy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	body := &logging.CopyLogEntriesRequest{}
	if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
		return err
	}
	op, err := svc.Entries.Copy(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("copying log entries: %w", err)
	}
	return emitFormatted(op, flagLogFormat)
}

func runLoggingRead(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	body := &logging.ListLogEntriesRequest{}
	if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
		return err
	}
	if flagLogFilter != "" && body.Filter == "" {
		body.Filter = flagLogFilter
	}
	if flagLogOrderBy != "" && body.OrderBy == "" {
		body.OrderBy = flagLogOrderBy
	}
	if flagLogPageSize > 0 && body.PageSize == 0 {
		body.PageSize = flagLogPageSize
	}
	if len(body.ResourceNames) == 0 {
		p, err := loggingParent()
		if err != nil {
			return err
		}
		body.ResourceNames = []string{p}
	}
	resp, err := svc.Entries.List(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reading log entries: %w", err)
	}
	return emitFormatted(resp.Entries, flagLogFormat)
}

func runLoggingWrite(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	body := &logging.WriteLogEntriesRequest{}
	if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
		return err
	}
	resp, err := svc.Entries.Write(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("writing log entries: %w", err)
	}
	return emitFormatted(resp, flagLogFormat)
}

func runLoggingTail(cmd *cobra.Command, args []string) error {
	// The Tail RPC is a bidi stream; the generated google-api client models it
	// as a single unary call that returns the initial response window. We
	// forward the request body and print whatever the server sends before
	// closing the stream, which matches how our other snapshot verbs render.
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	body := &logging.TailLogEntriesRequest{}
	if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
		return err
	}
	if len(body.ResourceNames) == 0 {
		p, err := loggingParent()
		if err != nil {
			return err
		}
		body.ResourceNames = []string{p}
	}
	resp, err := svc.Entries.Tail(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("tailing log entries: %w", err)
	}
	return emitFormatted(resp, flagLogFormat)
}

func init() {
	// entries data-plane commands
	for _, c := range []*cobra.Command{loggingCopyCmd, loggingReadCmd, loggingWriteCmd, loggingTailCmd} {
		c.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
		c.Flags().StringVar(&flagLogFormat, "format", "", "Output format")
	}
	loggingReadCmd.Flags().StringVar(&flagLogFilter, "filter", "", "Server-side list filter (overrides --config-file)")
	loggingReadCmd.Flags().StringVar(&flagLogOrderBy, "order-by", "", "Server-side ordering expression")
	loggingReadCmd.Flags().Int64Var(&flagLogPageSize, "page-size", 0, "Page size")
	addLogScopeFlags(loggingReadCmd, loggingTailCmd)
	loggingCmd.AddCommand(loggingCopyCmd, loggingReadCmd, loggingWriteCmd, loggingTailCmd)

	// Subgroup registration happens in the per-group files.
	rootCmd.AddCommand(loggingCmd)
}
