package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	monitoring "google.golang.org/api/monitoring/v3"
)

var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Manage Cloud Monitoring",
}

// --- policies list ---

var monitoringPoliciesCmd = &cobra.Command{
	Use:   "policies",
	Short: "Manage alerting policies",
}

var monitoringPoliciesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List alerting policies",
	Args:  cobra.NoArgs,
	RunE:  runMonitoringPoliciesList,
}

var (
	flagMonPoliciesFormat string
	flagMonPoliciesFilter string
	flagMonPoliciesURI    bool
)

// --- snoozes create ---

var monitoringSnoozesCmd = &cobra.Command{
	Use:   "snoozes",
	Short: "Manage alert snoozes",
}

var monitoringSnoozesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an alert snooze",
	Args:  cobra.NoArgs,
	RunE:  runMonitoringSnoozesCreate,
}

var (
	flagSnoozeDisplayName  string
	flagSnoozeStartTime    string
	flagSnoozeEndTime      string
	flagSnoozePolicies     []string
	flagSnoozeFilter       string
	flagSnoozeFromFile     string
	flagSnoozeCreateFormat string
)

// --- snoozes describe ---

var monitoringSnoozesDescribeCmd = &cobra.Command{
	Use:   "describe SNOOZE_ID",
	Short: "Describe an alert snooze",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringSnoozesDescribe,
}

var flagSnoozeDescribeFormat string

// --- snoozes list ---

var monitoringSnoozesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List alert snoozes",
	Args:  cobra.NoArgs,
	RunE:  runMonitoringSnoozesList,
}

var (
	flagSnoozesListFormat string
	flagSnoozesListURI    bool
)

// --- snoozes update ---

var monitoringSnoozesUpdateCmd = &cobra.Command{
	Use:   "update SNOOZE_ID",
	Short: "Update an alert snooze",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringSnoozesUpdate,
}

var (
	flagSnoozeUpdateDisplayName string
	flagSnoozeUpdateStartTime   string
	flagSnoozeUpdateEndTime     string
	flagSnoozeUpdateFromFile    string
)

// --- snoozes cancel ---

var monitoringSnoozesCancelCmd = &cobra.Command{
	Use:   "cancel SNOOZE_ID",
	Short: "Cancel an alert snooze",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringSnoozesCancel,
}

// --- policies describe ---

var monitoringPoliciesDescribeCmd = &cobra.Command{
	Use:   "describe POLICY_ID",
	Short: "Describe an alerting policy",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringPoliciesDescribe,
}

// --- policies create ---

var monitoringPoliciesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an alerting policy",
	Args:  cobra.NoArgs,
	RunE:  runMonitoringPoliciesCreate,
}

var (
	flagPolCreateFromFile      string
	flagPolCreateDisplayName   string
	flagPolCreateEnabled       bool
	flagPolCreateNoEnabled     bool
	flagPolCreateChannels      []string
	flagPolCreateDocumentation string
)

// --- policies update ---

var monitoringPoliciesUpdateCmd = &cobra.Command{
	Use:   "update POLICY_ID",
	Short: "Update an alerting policy",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringPoliciesUpdate,
}

var (
	flagPolUpdateFromFile  string
	flagPolUpdateEnabled   bool
	flagPolUpdateNoEnabled bool
	flagPolUpdateFields    string
	flagPolAddChannels     []string
	flagPolRemoveChannels  []string
)

// --- policies delete ---

var monitoringPoliciesDeleteCmd = &cobra.Command{
	Use:   "delete POLICY_ID",
	Short: "Delete an alerting policy",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringPoliciesDelete,
}

// --- channels ---

var monitoringChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Manage notification channels",
}

var monitoringChannelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification channels",
	Args:  cobra.NoArgs,
	RunE:  runMonitoringChannelsList,
}

var (
	flagChanListFormat string
	flagChanListFilter string
	flagChanListURI    bool
)

var monitoringChannelsDescribeCmd = &cobra.Command{
	Use:   "describe CHANNEL_ID",
	Short: "Describe a notification channel",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringChannelsDescribe,
}

var monitoringChannelsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a notification channel",
	Args:  cobra.NoArgs,
	RunE:  runMonitoringChannelsCreate,
}

var (
	flagChanType        string
	flagChanDisplayName string
	flagChanLabels      map[string]string
)

var monitoringChannelsUpdateCmd = &cobra.Command{
	Use:   "update CHANNEL_ID",
	Short: "Update a notification channel",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringChannelsUpdate,
}

var (
	flagChanUpdateDisplayName string
	flagChanUpdateLabels      map[string]string
	flagChanUpdateFields      string
)

var monitoringChannelsDeleteCmd = &cobra.Command{
	Use:   "delete CHANNEL_ID",
	Short: "Delete a notification channel",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringChannelsDelete,
}

// --- snoozes list filter ---
var flagSnoozesListFilter string

func init() {
	monitoringPoliciesListCmd.Flags().StringVar(&flagMonPoliciesFormat, "format", "", "Output format (e.g. json)")
	monitoringPoliciesListCmd.Flags().StringVar(&flagMonPoliciesFilter, "filter", "", "Filter expression")
	monitoringPoliciesListCmd.Flags().BoolVar(&flagMonPoliciesURI, "uri", false, "Print resource names")

	monitoringPoliciesCreateCmd.Flags().StringVar(&flagPolCreateFromFile, "policy-from-file", "", "JSON file with policy definition")
	monitoringPoliciesCreateCmd.Flags().StringVar(&flagPolCreateDisplayName, "display-name", "", "Display name")
	monitoringPoliciesCreateCmd.Flags().BoolVar(&flagPolCreateEnabled, "enabled", true, "Enable the policy")
	monitoringPoliciesCreateCmd.Flags().BoolVar(&flagPolCreateNoEnabled, "no-enabled", false, "Disable the policy")
	monitoringPoliciesCreateCmd.Flags().StringSliceVar(&flagPolCreateChannels, "notification-channels", nil, "Notification channel names")
	monitoringPoliciesCreateCmd.Flags().StringVar(&flagPolCreateDocumentation, "documentation", "", "Documentation content")

	monitoringPoliciesUpdateCmd.Flags().StringVar(&flagPolUpdateFromFile, "policy-from-file", "", "JSON file with policy definition")
	monitoringPoliciesUpdateCmd.Flags().BoolVar(&flagPolUpdateEnabled, "enabled", false, "Enable the policy")
	monitoringPoliciesUpdateCmd.Flags().BoolVar(&flagPolUpdateNoEnabled, "no-enabled", false, "Disable the policy")
	monitoringPoliciesUpdateCmd.Flags().StringVar(&flagPolUpdateFields, "fields", "", "Comma-separated fields to update")
	monitoringPoliciesUpdateCmd.Flags().StringSliceVar(&flagPolAddChannels, "add-notification-channels", nil, "Channels to add")
	monitoringPoliciesUpdateCmd.Flags().StringSliceVar(&flagPolRemoveChannels, "remove-notification-channels", nil, "Channels to remove")

	monitoringPoliciesCmd.AddCommand(monitoringPoliciesListCmd)
	monitoringPoliciesCmd.AddCommand(monitoringPoliciesDescribeCmd)
	monitoringPoliciesCmd.AddCommand(monitoringPoliciesCreateCmd)
	monitoringPoliciesCmd.AddCommand(monitoringPoliciesUpdateCmd)
	monitoringPoliciesCmd.AddCommand(monitoringPoliciesDeleteCmd)

	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeDisplayName, "display-name", "", "Display name for the snooze")
	monitoringSnoozesCreateCmd.MarkFlagRequired("display-name")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeStartTime, "start-time", "", "Start time (RFC3339)")
	monitoringSnoozesCreateCmd.MarkFlagRequired("start-time")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeEndTime, "end-time", "", "End time (RFC3339)")
	monitoringSnoozesCreateCmd.MarkFlagRequired("end-time")
	monitoringSnoozesCreateCmd.Flags().StringSliceVar(&flagSnoozePolicies, "criteria-policies", nil, "Alert policy resource names to snooze")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeFilter, "criteria-filter", "", "Filter for snooze criteria")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeFromFile, "snooze-from-file", "", "JSON file containing snooze definition")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeCreateFormat, "format", "", "Output format (e.g. json, yaml, 'value(name)')")
	monitoringSnoozesCmd.AddCommand(monitoringSnoozesCreateCmd)

	monitoringSnoozesListCmd.Flags().StringVar(&flagSnoozesListFormat, "format", "", "Output format: yaml (default), json, 'csv(NAME,DISPLAY_NAME)'")
	monitoringSnoozesListCmd.Flags().StringVar(&flagSnoozesListFilter, "filter", "", "Filter expression")
	monitoringSnoozesListCmd.Flags().BoolVar(&flagSnoozesListURI, "uri", false, "Print resource names")
	monitoringSnoozesCmd.AddCommand(monitoringSnoozesListCmd)

	monitoringSnoozesDescribeCmd.Flags().StringVar(&flagSnoozeDescribeFormat, "format", "", "Output format: yaml (default), json, 'table(name,display_name)'")
	monitoringSnoozesCmd.AddCommand(monitoringSnoozesDescribeCmd)

	monitoringSnoozesUpdateCmd.Flags().StringVar(&flagSnoozeUpdateDisplayName, "display-name", "", "New display name")
	monitoringSnoozesUpdateCmd.Flags().StringVar(&flagSnoozeUpdateStartTime, "start-time", "", "New start time (RFC3339)")
	monitoringSnoozesUpdateCmd.Flags().StringVar(&flagSnoozeUpdateEndTime, "end-time", "", "New end time (RFC3339)")
	monitoringSnoozesUpdateCmd.Flags().StringVar(&flagSnoozeUpdateFromFile, "snooze-from-file", "", "JSON file containing snooze definition")
	monitoringSnoozesCmd.AddCommand(monitoringSnoozesUpdateCmd)

	monitoringSnoozesCmd.AddCommand(monitoringSnoozesCancelCmd)

	// channels
	monitoringChannelsListCmd.Flags().StringVar(&flagChanListFormat, "format", "", "Output format (e.g. json)")
	monitoringChannelsListCmd.Flags().StringVar(&flagChanListFilter, "filter", "", "Filter expression")
	monitoringChannelsListCmd.Flags().BoolVar(&flagChanListURI, "uri", false, "Print resource names")
	monitoringChannelsCreateCmd.Flags().StringVar(&flagChanType, "type", "", "Channel type (required)")
	monitoringChannelsCreateCmd.MarkFlagRequired("type")
	monitoringChannelsCreateCmd.Flags().StringVar(&flagChanDisplayName, "display-name", "", "Display name")
	monitoringChannelsCreateCmd.Flags().StringToStringVar(&flagChanLabels, "channel-labels", nil, "Channel labels (key=value)")
	monitoringChannelsUpdateCmd.Flags().StringVar(&flagChanUpdateDisplayName, "display-name", "", "New display name")
	monitoringChannelsUpdateCmd.Flags().StringToStringVar(&flagChanUpdateLabels, "channel-labels", nil, "Channel labels")
	monitoringChannelsUpdateCmd.Flags().StringVar(&flagChanUpdateFields, "fields", "", "Comma-separated fields to update")
	monitoringChannelsCmd.AddCommand(monitoringChannelsListCmd)
	monitoringChannelsCmd.AddCommand(monitoringChannelsDescribeCmd)
	monitoringChannelsCmd.AddCommand(monitoringChannelsCreateCmd)
	monitoringChannelsCmd.AddCommand(monitoringChannelsUpdateCmd)
	monitoringChannelsCmd.AddCommand(monitoringChannelsDeleteCmd)

	monitoringCmd.AddCommand(monitoringPoliciesCmd)
	monitoringCmd.AddCommand(monitoringSnoozesCmd)
	monitoringCmd.AddCommand(monitoringChannelsCmd)
	rootCmd.AddCommand(monitoringCmd)
}

func runMonitoringPoliciesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var allPolicies []*monitoring.AlertPolicy
	pageToken := ""
	for {
		call := svc.Projects.AlertPolicies.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagMonPoliciesFilter != "" {
			call = call.Filter(flagMonPoliciesFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing alert policies: %w", err)
		}
		allPolicies = append(allPolicies, resp.AlertPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagMonPoliciesURI {
		for _, policy := range allPolicies {
			fmt.Println(policy.Name)
		}
		return nil
	}

	if flagMonPoliciesFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allPolicies)
	}

	fmt.Printf("%-60s %-10s %s\n", "NAME", "ENABLED", "DISPLAY_NAME")
	for _, p := range allPolicies {
		fmt.Printf("%-60s %-10t %s\n", p.Name, p.Enabled, p.DisplayName)
	}
	return nil
}

func runMonitoringSnoozesCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var snooze *monitoring.Snooze
	if flagSnoozeFromFile != "" {
		data, err := os.ReadFile(flagSnoozeFromFile)
		if err != nil {
			return fmt.Errorf("reading snooze file: %w", err)
		}
		snooze = &monitoring.Snooze{}
		if err := json.Unmarshal(data, snooze); err != nil {
			return fmt.Errorf("parsing snooze file: %w", err)
		}
	} else {
		snooze = &monitoring.Snooze{}
	}
	// Apply flag overrides (flags take precedence over file values).
	if flagSnoozeDisplayName != "" {
		snooze.DisplayName = flagSnoozeDisplayName
	}
	if flagSnoozeStartTime != "" || flagSnoozeEndTime != "" {
		if snooze.Interval == nil {
			snooze.Interval = &monitoring.TimeInterval{}
		}
		if flagSnoozeStartTime != "" {
			snooze.Interval.StartTime = flagSnoozeStartTime
		}
		if flagSnoozeEndTime != "" {
			snooze.Interval.EndTime = flagSnoozeEndTime
		}
	}
	if len(flagSnoozePolicies) > 0 || flagSnoozeFilter != "" {
		if snooze.Criteria == nil {
			snooze.Criteria = &monitoring.Criteria{}
		}
		if len(flagSnoozePolicies) > 0 {
			snooze.Criteria.Policies = flagSnoozePolicies
		}
		if flagSnoozeFilter != "" {
			snooze.Criteria.Filter = flagSnoozeFilter
		}
	}

	result, err := svc.Projects.Snoozes.Create(fmt.Sprintf("projects/%s", project), snooze).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating snooze: %w", err)
	}

	if flagSnoozeCreateFormat != "" {
		return formatOutput(result, flagSnoozeCreateFormat)
	}

	fmt.Printf("Created snooze [%s].\n", result.Name)
	return nil
}

func snoozeName(project, snoozeID string) string {
	// Accept either a bare snooze ID or a full resource name.
	if strings.HasPrefix(snoozeID, "projects/") {
		return snoozeID
	}
	return fmt.Sprintf("projects/%s/snoozes/%s", project, snoozeID)
}

func runMonitoringSnoozesDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	snooze, err := svc.Projects.Snoozes.Get(snoozeName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing snooze: %w", err)
	}

	return formatSnooze(snooze, flagSnoozeDescribeFormat)
}

// formatSnooze renders a single snooze per --format. The default (empty) format
// is yaml, matching gcloud. It also supports json and csv(COLS)/table(COLS).
func formatSnooze(s *monitoring.Snooze, format string) error {
	switch {
	case format == "" || format == "yaml":
		return yamlEncode(s)
	case format == "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(s)
	case isCsvFormat(format):
		return printSnoozeColumns([]*monitoring.Snooze{s}, extractCsvFields(format), ",")
	case isTableFormat(format):
		return printSnoozeTable([]*monitoring.Snooze{s}, extractTableFields(format))
	}
	return yamlEncode(s)
}

// printSnoozeTable prints an aligned table with uppercased column headings,
// auto-sizing each column to the widest value (matching gcloud table output).
func printSnoozeTable(snoozes []*monitoring.Snooze, fields []string) error {
	headers := make([]string, len(fields))
	widths := make([]int, len(fields))
	for i, f := range fields {
		headers[i] = strings.ToUpper(f)
		widths[i] = len(headers[i])
	}
	rows := make([][]string, 0, len(snoozes))
	for _, s := range snoozes {
		row := make([]string, len(fields))
		for i, f := range fields {
			row[i] = snoozeField(s, f)
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
		rows = append(rows, row)
	}

	printRow := func(cols []string) {
		var b strings.Builder
		for i, c := range cols {
			if i > 0 {
				b.WriteString("  ")
			}
			if i < len(cols)-1 {
				fmt.Fprintf(&b, "%-*s", widths[i], c)
			} else {
				b.WriteString(c)
			}
		}
		fmt.Println(b.String())
	}

	printRow(headers)
	for _, row := range rows {
		printRow(row)
	}
	return nil
}

func runMonitoringSnoozesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var allSnoozes []*monitoring.Snooze
	pageToken := ""
	for {
		call := svc.Projects.Snoozes.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagSnoozesListFilter != "" {
			call = call.Filter(flagSnoozesListFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing snoozes: %w", err)
		}
		allSnoozes = append(allSnoozes, resp.Snoozes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagSnoozesListURI {
		for _, snooze := range allSnoozes {
			fmt.Println(snooze.Name)
		}
		return nil
	}

	return formatSnoozesList(allSnoozes, flagSnoozesListFormat)
}

// formatSnoozesList renders snoozes per --format. The default (empty) format is
// yaml, matching gcloud. It also supports json and csv(COLS)/table(COLS); any
// other value falls back to the default NAME/DISPLAY_NAME table.
func formatSnoozesList(snoozes []*monitoring.Snooze, format string) error {
	switch {
	case format == "" || format == "yaml":
		for _, s := range snoozes {
			fmt.Println("---")
			if err := yamlEncode(s); err != nil {
				return err
			}
		}
		return nil
	case format == "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(snoozes)
	case isCsvFormat(format):
		return printSnoozeColumns(snoozes, extractCsvFields(format), ",")
	case isTableFormat(format):
		return printSnoozeColumns(snoozes, extractTableFields(format), "\t")
	}

	// Default columnar table.
	fmt.Printf("%-60s %s\n", "NAME", "DISPLAY_NAME")
	for _, s := range snoozes {
		fmt.Printf("%-60s %s\n", s.Name, s.DisplayName)
	}
	return nil
}

// printSnoozeColumns prints a heading row of the requested column names (as
// given) followed by one row per snooze, joined by sep.
func printSnoozeColumns(snoozes []*monitoring.Snooze, fields []string, sep string) error {
	fmt.Println(strings.Join(fields, sep))
	for _, s := range snoozes {
		vals := make([]string, len(fields))
		for i, f := range fields {
			vals[i] = snoozeField(s, f)
		}
		fmt.Println(strings.Join(vals, sep))
	}
	return nil
}

// snoozeField extracts a display column from a snooze. Column names are
// case-insensitive.
func snoozeField(s *monitoring.Snooze, field string) string {
	switch strings.ToUpper(field) {
	case "NAME":
		return s.Name
	case "DISPLAY_NAME", "DISPLAYNAME":
		return s.DisplayName
	}
	return ""
}

func runMonitoringSnoozesUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := snoozeName(project, args[0])

	if flagSnoozeUpdateFromFile != "" {
		data, err := os.ReadFile(flagSnoozeUpdateFromFile)
		if err != nil {
			return fmt.Errorf("reading snooze file: %w", err)
		}
		snooze := &monitoring.Snooze{}
		if err := json.Unmarshal(data, snooze); err != nil {
			return fmt.Errorf("parsing snooze file: %w", err)
		}
		result, err := svc.Projects.Snoozes.Patch(name, snooze).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating snooze: %w", err)
		}
		fmt.Printf("Updated snooze [%s].\n", result.Name)
		return nil
	}

	// Read-modify-write: get current snooze then apply updates.
	snooze, err := svc.Projects.Snoozes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting snooze: %w", err)
	}

	var updateFields []string
	if flagSnoozeUpdateDisplayName != "" {
		snooze.DisplayName = flagSnoozeUpdateDisplayName
		updateFields = append(updateFields, "display_name")
	}
	if flagSnoozeUpdateStartTime != "" {
		if snooze.Interval == nil {
			snooze.Interval = &monitoring.TimeInterval{}
		}
		snooze.Interval.StartTime = flagSnoozeUpdateStartTime
		updateFields = append(updateFields, "interval.start_time")
	}
	if flagSnoozeUpdateEndTime != "" {
		if snooze.Interval == nil {
			snooze.Interval = &monitoring.TimeInterval{}
		}
		snooze.Interval.EndTime = flagSnoozeUpdateEndTime
		updateFields = append(updateFields, "interval.end_time")
	}

	if len(updateFields) == 0 {
		return fmt.Errorf("at least one of --display-name, --start-time, or --end-time is required")
	}

	result, err := svc.Projects.Snoozes.Patch(name, snooze).
		UpdateMask(joinFields(updateFields)).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating snooze: %w", err)
	}

	fmt.Printf("Updated snooze [%s].\n", result.Name)
	return nil
}

func joinFields(fields []string) string {
	result := ""
	for i, f := range fields {
		if i > 0 {
			result += ","
		}
		result += f
	}
	return result
}

func runMonitoringSnoozesCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := snoozeName(project, args[0])

	// Get current snooze to check timing.
	snooze, err := svc.Projects.Snoozes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting snooze: %w", err)
	}

	// Cancel = set end time to now (+ 1 second).
	// If start time is in the future, also set start to now.
	now := time.Now().UTC().Add(time.Second).Format(time.RFC3339)
	updateMask := "interval.end_time"

	if snooze.Interval != nil && snooze.Interval.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, snooze.Interval.StartTime)
		if err == nil && startTime.After(time.Now()) {
			snooze.Interval.StartTime = now
			updateMask = "interval.start_time,interval.end_time"
		}
	}

	if snooze.Interval == nil {
		snooze.Interval = &monitoring.TimeInterval{}
	}
	snooze.Interval.EndTime = now

	result, err := svc.Projects.Snoozes.Patch(name, snooze).
		UpdateMask(updateMask).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling snooze: %w", err)
	}

	fmt.Printf("Cancelled snooze [%s].\n", result.Name)
	return nil
}

// --- policies describe (#179) ---

func runMonitoringPoliciesDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/alertPolicies/%s", project, args[0])
	policy, err := svc.Projects.AlertPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing alert policy: %w", err)
	}

	return formatOutput(policy, "")
}

// --- policies create (#180) ---

func runMonitoringPoliciesCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var policy *monitoring.AlertPolicy
	if flagPolCreateFromFile != "" {
		data, err := os.ReadFile(flagPolCreateFromFile)
		if err != nil {
			return fmt.Errorf("reading policy file: %w", err)
		}
		policy = &monitoring.AlertPolicy{}
		if err := json.Unmarshal(data, policy); err != nil {
			return fmt.Errorf("parsing policy file: %w", err)
		}
	} else {
		policy = &monitoring.AlertPolicy{}
	}

	if flagPolCreateDisplayName != "" {
		policy.DisplayName = flagPolCreateDisplayName
	}
	if flagPolCreateNoEnabled {
		policy.Enabled = false
	} else {
		policy.Enabled = flagPolCreateEnabled
	}
	if len(flagPolCreateChannels) > 0 {
		policy.NotificationChannels = flagPolCreateChannels
	}
	if flagPolCreateDocumentation != "" {
		policy.Documentation = &monitoring.Documentation{
			Content: flagPolCreateDocumentation,
		}
	}

	result, err := svc.Projects.AlertPolicies.Create(fmt.Sprintf("projects/%s", project), policy).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating alert policy: %w", err)
	}

	fmt.Printf("Created alert policy [%s].\n", result.Name)
	return nil
}

// --- policies update (#181) ---

func runMonitoringPoliciesUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/alertPolicies/%s", project, args[0])

	if flagPolUpdateFromFile != "" {
		data, err := os.ReadFile(flagPolUpdateFromFile)
		if err != nil {
			return fmt.Errorf("reading policy file: %w", err)
		}
		policy := &monitoring.AlertPolicy{}
		if err := json.Unmarshal(data, policy); err != nil {
			return fmt.Errorf("parsing policy file: %w", err)
		}
		call := svc.Projects.AlertPolicies.Patch(name, policy).Context(ctx)
		if flagPolUpdateFields != "" {
			call = call.UpdateMask(flagPolUpdateFields)
		}
		result, err := call.Do()
		if err != nil {
			return fmt.Errorf("updating alert policy: %w", err)
		}
		fmt.Printf("Updated alert policy [%s].\n", result.Name)
		return nil
	}

	// Read-modify-write.
	policy, err := svc.Projects.AlertPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting alert policy: %w", err)
	}

	var updateFields []string
	if flagPolUpdateEnabled {
		policy.Enabled = true
		updateFields = append(updateFields, "enabled")
	}
	if flagPolUpdateNoEnabled {
		policy.Enabled = false
		updateFields = append(updateFields, "enabled")
	}
	if len(flagPolAddChannels) > 0 {
		policy.NotificationChannels = append(policy.NotificationChannels, flagPolAddChannels...)
		updateFields = append(updateFields, "notification_channels")
	}
	if len(flagPolRemoveChannels) > 0 {
		removeSet := make(map[string]bool)
		for _, ch := range flagPolRemoveChannels {
			removeSet[ch] = true
		}
		var kept []string
		for _, ch := range policy.NotificationChannels {
			if !removeSet[ch] {
				kept = append(kept, ch)
			}
		}
		policy.NotificationChannels = kept
		updateFields = append(updateFields, "notification_channels")
	}

	if len(updateFields) == 0 && flagPolUpdateFields == "" {
		return fmt.Errorf("no update flags specified")
	}

	mask := flagPolUpdateFields
	if mask == "" {
		mask = strings.Join(updateFields, ",")
	}

	result, err := svc.Projects.AlertPolicies.Patch(name, policy).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating alert policy: %w", err)
	}

	fmt.Printf("Updated alert policy [%s].\n", result.Name)
	return nil
}

// --- policies delete (#182) ---

func runMonitoringPoliciesDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Printf("You are about to delete alert policy [%s].\n", args[0])
		fmt.Print("Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/alertPolicies/%s", project, args[0])
	if _, err := svc.Projects.AlertPolicies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting alert policy: %w", err)
	}

	fmt.Printf("Deleted alert policy [%s].\n", args[0])
	return nil
}

// --- channels list (#183) ---

func runMonitoringChannelsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var allChannels []*monitoring.NotificationChannel
	pageToken := ""
	for {
		call := svc.Projects.NotificationChannels.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagChanListFilter != "" {
			call = call.Filter(flagChanListFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notification channels: %w", err)
		}
		allChannels = append(allChannels, resp.NotificationChannels...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagChanListURI {
		for _, ch := range allChannels {
			fmt.Println(ch.Name)
		}
		return nil
	}

	if flagChanListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allChannels)
	}

	fmt.Printf("%-60s %-15s %-10s %s\n", "NAME", "TYPE", "ENABLED", "DISPLAY_NAME")
	for _, ch := range allChannels {
		fmt.Printf("%-60s %-15s %-10t %s\n", ch.Name, ch.Type, ch.Enabled, ch.DisplayName)
	}
	return nil
}

// --- channels describe (#184) ---

func runMonitoringChannelsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/notificationChannels/%s", project, args[0])
	ch, err := svc.Projects.NotificationChannels.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notification channel: %w", err)
	}

	return formatOutput(ch, "")
}

// --- channels create (#184) ---

func runMonitoringChannelsCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	ch := &monitoring.NotificationChannel{
		Type: flagChanType,
	}
	if flagChanDisplayName != "" {
		ch.DisplayName = flagChanDisplayName
	}
	if len(flagChanLabels) > 0 {
		ch.Labels = flagChanLabels
	}

	result, err := svc.Projects.NotificationChannels.Create(fmt.Sprintf("projects/%s", project), ch).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notification channel: %w", err)
	}

	fmt.Printf("Created notification channel [%s].\n", result.Name)
	return nil
}

// --- channels update (#184) ---

func runMonitoringChannelsUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/notificationChannels/%s", project, args[0])

	ch, err := svc.Projects.NotificationChannels.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting notification channel: %w", err)
	}

	var updateFields []string
	if flagChanUpdateDisplayName != "" {
		ch.DisplayName = flagChanUpdateDisplayName
		updateFields = append(updateFields, "display_name")
	}
	if len(flagChanUpdateLabels) > 0 {
		if ch.Labels == nil {
			ch.Labels = make(map[string]string)
		}
		for k, v := range flagChanUpdateLabels {
			ch.Labels[k] = v
		}
		updateFields = append(updateFields, "labels")
	}

	mask := flagChanUpdateFields
	if mask == "" && len(updateFields) > 0 {
		mask = strings.Join(updateFields, ",")
	}

	result, err := svc.Projects.NotificationChannels.Patch(name, ch).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating notification channel: %w", err)
	}

	fmt.Printf("Updated notification channel [%s].\n", result.Name)
	return nil
}

// --- channels delete (#184) ---

func runMonitoringChannelsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Printf("You are about to delete notification channel [%s].\n", args[0])
		fmt.Print("Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/notificationChannels/%s", project, args[0])
	if _, err := svc.Projects.NotificationChannels.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting notification channel: %w", err)
	}

	fmt.Printf("Deleted notification channel [%s].\n", args[0])
	return nil
}
