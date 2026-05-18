package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	flagSnoozeDisplayName string
	flagSnoozeStartTime   string
	flagSnoozeEndTime     string
	flagSnoozePolicies    []string
	flagSnoozeFilter      string
	flagSnoozeFromFile    string
)

// --- snoozes describe ---

var monitoringSnoozesDescribeCmd = &cobra.Command{
	Use:   "describe SNOOZE_ID",
	Short: "Describe an alert snooze",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringSnoozesDescribe,
}

// --- snoozes list ---

var monitoringSnoozesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List alert snoozes",
	Args:  cobra.NoArgs,
	RunE:  runMonitoringSnoozesList,
}

var flagSnoozesListFormat string

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

func init() {
	monitoringPoliciesListCmd.Flags().StringVar(&flagMonPoliciesFormat, "format", "", "Output format (e.g. json)")
	monitoringPoliciesListCmd.Flags().StringVar(&flagMonPoliciesFilter, "filter", "", "Filter expression")
	monitoringPoliciesCmd.AddCommand(monitoringPoliciesListCmd)

	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeDisplayName, "display-name", "", "Display name for the snooze")
	monitoringSnoozesCreateCmd.MarkFlagRequired("display-name")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeStartTime, "start-time", "", "Start time (RFC3339)")
	monitoringSnoozesCreateCmd.MarkFlagRequired("start-time")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeEndTime, "end-time", "", "End time (RFC3339)")
	monitoringSnoozesCreateCmd.MarkFlagRequired("end-time")
	monitoringSnoozesCreateCmd.Flags().StringSliceVar(&flagSnoozePolicies, "criteria-policies", nil, "Alert policy resource names to snooze")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeFilter, "criteria-filter", "", "Filter for snooze criteria")
	monitoringSnoozesCreateCmd.Flags().StringVar(&flagSnoozeFromFile, "snooze-from-file", "", "JSON file containing snooze definition")
	monitoringSnoozesCmd.AddCommand(monitoringSnoozesCreateCmd)

	monitoringSnoozesListCmd.Flags().StringVar(&flagSnoozesListFormat, "format", "", "Output format (e.g. json)")
	monitoringSnoozesCmd.AddCommand(monitoringSnoozesListCmd)

	monitoringSnoozesCmd.AddCommand(monitoringSnoozesDescribeCmd)

	monitoringSnoozesUpdateCmd.Flags().StringVar(&flagSnoozeUpdateDisplayName, "display-name", "", "New display name")
	monitoringSnoozesUpdateCmd.Flags().StringVar(&flagSnoozeUpdateStartTime, "start-time", "", "New start time (RFC3339)")
	monitoringSnoozesUpdateCmd.Flags().StringVar(&flagSnoozeUpdateEndTime, "end-time", "", "New end time (RFC3339)")
	monitoringSnoozesUpdateCmd.Flags().StringVar(&flagSnoozeUpdateFromFile, "snooze-from-file", "", "JSON file containing snooze definition")
	monitoringSnoozesCmd.AddCommand(monitoringSnoozesUpdateCmd)

	monitoringSnoozesCmd.AddCommand(monitoringSnoozesCancelCmd)

	monitoringCmd.AddCommand(monitoringPoliciesCmd)
	monitoringCmd.AddCommand(monitoringSnoozesCmd)
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
		snooze = &monitoring.Snooze{
			DisplayName: flagSnoozeDisplayName,
			Interval: &monitoring.TimeInterval{
				StartTime: flagSnoozeStartTime,
				EndTime:   flagSnoozeEndTime,
			},
			Criteria: &monitoring.Criteria{
				Policies: flagSnoozePolicies,
				Filter:   flagSnoozeFilter,
			},
		}
	}

	result, err := svc.Projects.Snoozes.Create(fmt.Sprintf("projects/%s", project), snooze).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating snooze: %w", err)
	}

	fmt.Printf("Created snooze [%s].\n", result.Name)
	return nil
}

func snoozeName(project, snoozeID string) string {
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

	return formatOutput(snooze, "")
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

	if flagSnoozesListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allSnoozes)
	}

	fmt.Printf("%-60s %s\n", "NAME", "DISPLAY_NAME")
	for _, s := range allSnoozes {
		fmt.Printf("%-60s %s\n", s.Name, s.DisplayName)
	}
	return nil
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
