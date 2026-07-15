package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging metrics (#916) ---
//
// Logs-based metrics live only in a project (Metrics.projects.metrics in the
// underlying API). --organization/--folder/--billing-account are rejected.

var loggingMetricsCmd = &cobra.Command{Use: "metrics", Short: "Manage logs-based metrics"}

var (
	flagLogMetricDescription string
	flagLogMetricFilter      string
	flagLogMetricBucketName  string
)

var (
	loggingMetricsCreateCmd = &cobra.Command{
		Use: "create METRIC", Short: "Create a logs-based metric",
		Args: cobra.ExactArgs(1), RunE: runLogMetricCreate,
	}
	loggingMetricsDeleteCmd = &cobra.Command{
		Use: "delete METRIC", Short: "Delete a logs-based metric",
		Args: cobra.ExactArgs(1), RunE: runLogMetricDelete,
	}
	loggingMetricsDescribeCmd = &cobra.Command{
		Use: "describe METRIC", Short: "Describe a logs-based metric",
		Args: cobra.ExactArgs(1), RunE: runLogMetricDescribe,
	}
	loggingMetricsListCmd = &cobra.Command{
		Use: "list", Short: "List logs-based metrics",
		Args: cobra.NoArgs, RunE: runLogMetricList,
	}
	loggingMetricsUpdateCmd = &cobra.Command{
		Use: "update METRIC", Short: "Update a logs-based metric",
		Args: cobra.ExactArgs(1), RunE: runLogMetricUpdate,
	}
)

func metricProjectParent() (string, error) {
	parent, err := loggingParent()
	if err != nil {
		return "", err
	}
	if loggingScope(parent) != "projects" {
		return "", fmt.Errorf("logs-based metrics are only supported at project scope")
	}
	return parent, nil
}

func metricFromFlags(body *logging.LogMetric, id string) {
	if body.Name == "" {
		body.Name = id
	}
	if flagLogMetricDescription != "" {
		body.Description = flagLogMetricDescription
	}
	if flagLogMetricFilter != "" {
		body.Filter = flagLogMetricFilter
	}
	if flagLogMetricBucketName != "" {
		body.BucketName = flagLogMetricBucketName
	}
}

func runLogMetricCreate(cmd *cobra.Command, args []string) error {
	parent, err := metricProjectParent()
	if err != nil {
		return err
	}
	body := &logging.LogMetric{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	metricFromFlags(body, args[0])
	if body.Filter == "" {
		return fmt.Errorf("--log-filter is required (or provide `filter` via --config-file)")
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Metrics.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating log metric: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogMetricDelete(cmd *cobra.Command, args []string) error {
	parent, err := metricProjectParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "metrics", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Metrics.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting log metric: %w", err)
	}
	fmt.Printf("Deleted metric [%s].\n", args[0])
	return nil
}

func runLogMetricDescribe(cmd *cobra.Command, args []string) error {
	parent, err := metricProjectParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "metrics", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Metrics.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing log metric: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogMetricList(cmd *cobra.Command, args []string) error {
	parent, err := metricProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.LogMetric
	pageToken := ""
	for {
		call := svc.Projects.Metrics.List(parent).Context(ctx)
		if flagLogPageSize > 0 {
			call = call.PageSize(flagLogPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing log metrics: %w", err)
		}
		all = append(all, resp.Metrics...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "FILTER")
	for _, m := range all {
		fmt.Printf("%-40s %s\n", m.Name, m.Filter)
	}
	return nil
}

func runLogMetricUpdate(cmd *cobra.Command, args []string) error {
	parent, err := metricProjectParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "metrics", args[0])
	body := &logging.LogMetric{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	metricFromFlags(body, args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Metrics.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating log metric: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func init() {
	all := []*cobra.Command{loggingMetricsCreateCmd, loggingMetricsDeleteCmd, loggingMetricsDescribeCmd,
		loggingMetricsListCmd, loggingMetricsUpdateCmd}
	addLogScopeFlags(all...)
	addLogFormatFlag(loggingMetricsCreateCmd, loggingMetricsDescribeCmd, loggingMetricsListCmd, loggingMetricsUpdateCmd)
	addLogPageSizeFlag(loggingMetricsListCmd)
	for _, c := range []*cobra.Command{loggingMetricsCreateCmd, loggingMetricsUpdateCmd} {
		c.Flags().StringVar(&flagLogMetricDescription, "description", "", "A textual description for the metric")
		c.Flags().StringVar(&flagLogMetricFilter, "log-filter", "", "Advanced logs filter")
		c.Flags().StringVar(&flagLogMetricBucketName, "bucket-name", "", "Log bucket that owns the metric")
		c.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogMetric body")
	}
	loggingMetricsCmd.AddCommand(all...)
	loggingCmd.AddCommand(loggingMetricsCmd)
}
