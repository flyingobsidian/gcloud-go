package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	monitoring "google.golang.org/api/monitoring/v3"
)

// --- gcloud monitoring alerts (GA in gcloud-python 589.0.0, #1919) ---

func (g *monitoringCmds) initMonitoringAlerts() {
	for _, c := range []*cobra.Command{g.monitoringAlertsListCmd, g.monitoringAlertsDescribeCmd} {
		c.Flags().StringVar(&g.flagMonAlertsFormat, "format", "", "Output format (default yaml)")
	}
	g.monitoringAlertsListCmd.Flags().StringVar(&g.flagMonAlertsFilter, "filter", "",
		"Server-side filter expression, e.g. state=OPEN")
	g.monitoringAlertsListCmd.Flags().StringSliceVar(&g.flagMonAlertsSortBy, "sort-by", nil,
		"Comma-separated list of fields to order by (sent to the API as orderBy)")
	g.monitoringAlertsListCmd.Flags().Int64Var(&g.flagMonAlertsPageSize, "page-size", 1000,
		"Maximum number of alerts to request")
	g.monitoringAlertsListCmd.Flags().IntVar(&g.flagMonAlertsLimit, "limit", 0,
		"Maximum number of alerts to list (0 = no limit)")
	g.monitoringAlertsListCmd.Flags().BoolVar(&g.flagMonAlertsURI, "uri", false, "Print resource names")

	g.monitoringAlertsCmd.AddCommand(g.monitoringAlertsListCmd, g.monitoringAlertsDescribeCmd)
	g.monitoringCmd.AddCommand(g.monitoringAlertsCmd)
}

// monitoringAlertName expands a short alert id to projects/P/alerts/ID;
// fully qualified names pass through.
func monitoringAlertName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/alerts/%s", project, id)
}

func (g *monitoringCmds) runMonitoringAlertsList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}
	// Like gcloud-python, filter and order are applied server-side and only
	// a single page of --page-size alerts is fetched.
	call := svc.Projects.Alerts.List("projects/" + project).PageSize(g.flagMonAlertsPageSize).Context(ctx)
	if g.flagMonAlertsFilter != "" {
		call = call.Filter(g.flagMonAlertsFilter)
	}
	if len(g.flagMonAlertsSortBy) > 0 {
		call = call.OrderBy(strings.Join(g.flagMonAlertsSortBy, ","))
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing alerts: %w", err)
	}
	alerts := resp.Alerts
	if g.flagMonAlertsLimit > 0 && len(alerts) > g.flagMonAlertsLimit {
		alerts = alerts[:g.flagMonAlertsLimit]
	}
	return emitMonitoringAlerts(g.inv.stdout, alerts, g.flagMonAlertsFormat, g.flagMonAlertsURI)
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

func (g *monitoringCmds) runMonitoringAlertsDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}
	alert, err := svc.Projects.Alerts.Get(monitoringAlertName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing alert: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, alert, g.flagMonAlertsFormat)
}
