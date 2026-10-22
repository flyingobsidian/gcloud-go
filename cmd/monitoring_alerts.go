package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	monitoring "google.golang.org/api/monitoring/v3"
)

// --- gcloud monitoring alerts (GA in gcloud-python 589.0.0, #1919) ---

var monitoringAlertsCmd = &cobra.Command{Use: "alerts", Short: "Manage Cloud Monitoring alerts"}

var monitoringAlertsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List alerts for a project",
	Args:  cobra.NoArgs,
	RunE:  runMonitoringAlertsList,
}

var monitoringAlertsDescribeCmd = &cobra.Command{
	Use:   "describe ALERT",
	Short: "Describe an alert",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonitoringAlertsDescribe,
}

var (
	flagMonAlertsFormat   string
	flagMonAlertsFilter   string
	flagMonAlertsSortBy   []string
	flagMonAlertsPageSize int64
	flagMonAlertsLimit    int
	flagMonAlertsURI      bool
)

func init() {
	for _, c := range []*cobra.Command{monitoringAlertsListCmd, monitoringAlertsDescribeCmd} {
		c.Flags().StringVar(&flagMonAlertsFormat, "format", "", "Output format (default yaml)")
	}
	monitoringAlertsListCmd.Flags().StringVar(&flagMonAlertsFilter, "filter", "",
		"Server-side filter expression, e.g. state=OPEN")
	monitoringAlertsListCmd.Flags().StringSliceVar(&flagMonAlertsSortBy, "sort-by", nil,
		"Comma-separated list of fields to order by (sent to the API as orderBy)")
	monitoringAlertsListCmd.Flags().Int64Var(&flagMonAlertsPageSize, "page-size", 1000,
		"Maximum number of alerts to request")
	monitoringAlertsListCmd.Flags().IntVar(&flagMonAlertsLimit, "limit", 0,
		"Maximum number of alerts to list (0 = no limit)")
	monitoringAlertsListCmd.Flags().BoolVar(&flagMonAlertsURI, "uri", false, "Print resource names")

	monitoringAlertsCmd.AddCommand(monitoringAlertsListCmd, monitoringAlertsDescribeCmd)
	monitoringCmd.AddCommand(monitoringAlertsCmd)
}

// monitoringAlertName expands a short alert id to projects/P/alerts/ID;
// fully qualified names pass through.
func monitoringAlertName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/alerts/%s", project, id)
}

func runMonitoringAlertsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// Like gcloud-python, filter and order are applied server-side and only
	// a single page of --page-size alerts is fetched.
	call := svc.Projects.Alerts.List("projects/" + project).PageSize(flagMonAlertsPageSize).Context(ctx)
	if flagMonAlertsFilter != "" {
		call = call.Filter(flagMonAlertsFilter)
	}
	if len(flagMonAlertsSortBy) > 0 {
		call = call.OrderBy(strings.Join(flagMonAlertsSortBy, ","))
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing alerts: %w", err)
	}
	alerts := resp.Alerts
	if flagMonAlertsLimit > 0 && len(alerts) > flagMonAlertsLimit {
		alerts = alerts[:flagMonAlertsLimit]
	}
	return emitMonitoringAlerts(os.Stdout, alerts, flagMonAlertsFormat, flagMonAlertsURI)
}

// emitMonitoringAlerts prints alerts as yaml records separated by "---"
// (the gcloud-python default), as resource names with --uri, or per --format.
func emitMonitoringAlerts(w io.Writer, alerts []*monitoring.Alert, format string, uri bool) error {
	if uri {
		for _, a := range alerts {
			if _, err := fmt.Fprintln(w, a.Name); err != nil {
				return err
			}
		}
		return nil
	}
	if format != "" && format != "yaml" {
		if alerts == nil {
			alerts = []*monitoring.Alert{}
		}
		return emitFormattedTo(w, alerts, format)
	}
	for i, a := range alerts {
		if i > 0 {
			if _, err := fmt.Fprintln(w, "---"); err != nil {
				return err
			}
		}
		if err := emitYAML(w, a); err != nil {
			return err
		}
	}
	return nil
}

func runMonitoringAlertsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}
	alert, err := svc.Projects.Alerts.Get(monitoringAlertName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing alert: %w", err)
	}
	return emitFormatted(alert, flagMonAlertsFormat)
}
