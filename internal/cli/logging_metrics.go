package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging metrics (#916) ---
//
// Logs-based metrics live only in a project (Metrics.projects.metrics in the
// underlying API). --organization/--folder/--billing-account are rejected.

func (g *loggingCmds) metricProjectParent() (string, error) {
	parent, err := g.loggingParent()
	if err != nil {
		return "", err
	}
	if loggingScope(parent) != "projects" {
		return "", fmt.Errorf("logs-based metrics are only supported at project scope")
	}
	return parent, nil
}

func (g *loggingCmds) metricFromFlags(body *logging.LogMetric, id string) {
	if body.Name == "" {
		body.Name = id
	}
	if g.flagLogMetricDescription != "" {
		body.Description = g.flagLogMetricDescription
	}
	if g.flagLogMetricFilter != "" {
		body.Filter = g.flagLogMetricFilter
	}
	if g.flagLogMetricBucketName != "" {
		body.BucketName = g.flagLogMetricBucketName
	}
}

func (g *loggingCmds) runLogMetricCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.metricProjectParent()
	if err != nil {
		return err
	}
	body := &logging.LogMetric{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	g.metricFromFlags(body, args[0])
	if body.Filter == "" {
		return fmt.Errorf("--log-filter is required (or provide `filter` via --config-file)")
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Metrics.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating log metric: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogMetricDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.metricProjectParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "metrics", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Metrics.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting log metric: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted metric [%s].\n", args[0])
	return nil
}

func (g *loggingCmds) runLogMetricDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.metricProjectParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "metrics", args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Metrics.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing log metric: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogMetricList(cmd *cobra.Command, args []string) error {
	parent, err := g.metricProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.LogMetric
	pageToken := ""
	for {
		call := svc.Projects.Metrics.List(parent).Context(ctx)
		if g.flagLogPageSize > 0 {
			call = call.PageSize(g.flagLogPageSize)
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
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "FILTER")
	for _, m := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", m.Name, m.Filter)
	}
	return nil
}

func (g *loggingCmds) runLogMetricUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.metricProjectParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "metrics", args[0])
	body := &logging.LogMetric{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	g.metricFromFlags(body, args[0])
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Metrics.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating log metric: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) initLoggingMetrics() {
	all := []*cobra.Command{g.loggingMetricsCreateCmd, g.loggingMetricsDeleteCmd, g.loggingMetricsDescribeCmd,
		g.loggingMetricsListCmd, g.loggingMetricsUpdateCmd}
	g.addLogScopeFlags(all...)
	g.addLogFormatFlag(g.loggingMetricsCreateCmd, g.loggingMetricsDescribeCmd, g.loggingMetricsListCmd, g.loggingMetricsUpdateCmd)
	g.addLogPageSizeFlag(g.loggingMetricsListCmd)
	for _, c := range []*cobra.Command{g.loggingMetricsCreateCmd, g.loggingMetricsUpdateCmd} {
		c.Flags().StringVar(&g.flagLogMetricDescription, "description", "", "A textual description for the metric")
		c.Flags().StringVar(&g.flagLogMetricFilter, "log-filter", "", "Advanced logs filter")
		c.Flags().StringVar(&g.flagLogMetricBucketName, "bucket-name", "", "Log bucket that owns the metric")
		c.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogMetric body")
	}
	g.loggingMetricsCmd.AddCommand(all...)
	g.loggingCmd.AddCommand(g.loggingMetricsCmd)
}
