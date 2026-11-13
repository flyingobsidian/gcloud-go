package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	monitoringv1 "google.golang.org/api/monitoring/v1"
	monitoring "google.golang.org/api/monitoring/v3"
)

// --- gcloud monitoring dashboards (#951) & uptime (#952) ---

// --- Dashboards ---

// --- Uptime ---

func (g *monitoringCmds) initMonitoringDashboardsUptime() {
	g.monitoringDashboardsCreateCmd.Flags().StringVar(&g.flagMonDashFile, "config-from-file", "", "YAML/JSON file with the Dashboard definition (required)")
	g.monitoringDashboardsCreateCmd.Flags().StringVar(&g.flagMonDashDisplay, "display-name", "", "Override the display name in the config file")
	g.monitoringDashboardsCreateCmd.MarkFlagRequired("config-from-file")

	g.monitoringDashboardsListCmd.Flags().StringVar(&g.flagMonDashFilter, "filter", "", "Server-side filter expression")
	g.monitoringDashboardsListCmd.Flags().Int64Var(&g.flagMonDashPage, "page-size", 0, "Number of results per page")
	g.monitoringDashboardsListCmd.Flags().Int64Var(&g.flagMonDashLimit, "limit", 0, "Maximum number of results to return")

	g.monitoringDashboardsUpdateCmd.Flags().StringVar(&g.flagMonDashFile, "config-from-file", "", "YAML/JSON file with the Dashboard patch (required)")
	g.monitoringDashboardsUpdateCmd.Flags().StringVar(&g.flagMonDashDisplay, "display-name", "", "Override the display name in the config file")
	g.monitoringDashboardsUpdateCmd.MarkFlagRequired("config-from-file")

	g.monitoringDashboardsCmd.AddCommand(
		g.monitoringDashboardsCreateCmd, g.monitoringDashboardsDeleteCmd,
		g.monitoringDashboardsDescribeCmd, g.monitoringDashboardsListCmd,
		g.monitoringDashboardsUpdateCmd,
	)

	// Uptime create/update: users may pass a config file OR use short flags for
	// common cases (http/https/tcp against a URL/host).
	for _, c := range []*cobra.Command{g.monitoringUptimeCreateCmd, g.monitoringUptimeUpdateCmd} {
		c.Flags().StringVar(&g.flagMonUCFile, "config-from-file", "", "YAML/JSON file with the UptimeCheckConfig")
		c.Flags().StringVar(&g.flagMonUCDisplayName, "display-name", "", "Display name")
		c.Flags().StringVar(&g.flagMonUCResourceType, "resource-type", "uptime-url", "Monitored resource type (uptime-url, gce-instance, aws-ec2-instance, ...)")
		c.Flags().StringVar(&g.flagMonUCHost, "host", "", "Host or URL to check")
		c.Flags().StringVar(&g.flagMonUCPath, "path", "", "HTTP path")
		c.Flags().Int64Var(&g.flagMonUCPort, "port", 0, "TCP or HTTP port")
		c.Flags().StringVar(&g.flagMonUCProtocol, "protocol", "http", "Protocol (http, https, tcp)")
		c.Flags().StringVar(&g.flagMonUCPeriod, "period", "", "Check period in seconds (60, 300, 600, 900); e.g. 60s")
		c.Flags().StringVar(&g.flagMonUCTimeout, "timeout", "", "Timeout in seconds; e.g. 10s")
	}
	g.monitoringUptimeListCmd.Flags().StringVar(&g.flagMonUCFilter, "filter", "", "Server-side filter expression")
	g.monitoringUptimeListCmd.Flags().Int64Var(&g.flagMonUCPage, "page-size", 0, "Number of results per page")
	g.monitoringUptimeListCmd.Flags().Int64Var(&g.flagMonUCLimit, "limit", 0, "Maximum number of results to return")

	g.monitoringUptimeCmd.AddCommand(
		g.monitoringUptimeCreateCmd, g.monitoringUptimeDeleteCmd,
		g.monitoringUptimeDescribeCmd, g.monitoringUptimeListCmd, g.monitoringUptimeUpdateCmd,
	)

	g.monitoringCmd.AddCommand(g.monitoringDashboardsCmd, g.monitoringUptimeCmd)
}

// --- Dashboards impl ---

func dashboardName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/dashboards/%s", project, id)
}

func (g *monitoringCmds) runMonDashCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	dash := &monitoringv1.Dashboard{}
	if err := loadYAMLOrJSONInto(g.flagMonDashFile, dash); err != nil {
		return err
	}
	if g.flagMonDashDisplay != "" {
		dash.DisplayName = g.flagMonDashDisplay
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().MonitoringV1(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Dashboards.Create(fmt.Sprintf("projects/%s", project), dash).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating dashboard: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Created dashboard [%s].\n", got.Name)
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *monitoringCmds) runMonDashDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().MonitoringV1(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Dashboards.Delete(dashboardName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting dashboard: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted dashboard [%s].\n", args[0])
	return nil
}

func (g *monitoringCmds) runMonDashDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().MonitoringV1(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Dashboards.Get(dashboardName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing dashboard: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *monitoringCmds) runMonDashList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().MonitoringV1(ctx)
	if err != nil {
		return err
	}
	var all []*monitoringv1.Dashboard
	pageToken := ""
	for {
		call := svc.Projects.Dashboards.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if g.flagMonDashPage > 0 {
			call = call.PageSize(g.flagMonDashPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing dashboards: %w", err)
		}
		all = append(all, resp.Dashboards...)
		if g.flagMonDashLimit > 0 && int64(len(all)) >= g.flagMonDashLimit {
			all = all[:g.flagMonDashLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	// Server-side dashboards.list does not accept --filter, so we honor it
	// client-side by dropping anything whose display name doesn't contain the
	// substring (case-insensitive), matching the behavior of gcloud python.
	if g.flagMonDashFilter != "" {
		needle := strings.ToLower(g.flagMonDashFilter)
		filtered := all[:0]
		for _, d := range all {
			if strings.Contains(strings.ToLower(d.DisplayName), needle) {
				filtered = append(filtered, d)
			}
		}
		all = filtered
	}
	return emitFormattedTo(g.inv.stdout, all, "")
}

func (g *monitoringCmds) runMonDashUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	dash := &monitoringv1.Dashboard{}
	if err := loadYAMLOrJSONInto(g.flagMonDashFile, dash); err != nil {
		return err
	}
	if g.flagMonDashDisplay != "" {
		dash.DisplayName = g.flagMonDashDisplay
	}
	name := dashboardName(project, args[0])
	// Dashboards.Patch expects the resource's Name to match the URL.
	dash.Name = name
	ctx := cmd.Context()
	svc, err := g.inv.clients().MonitoringV1(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Dashboards.Patch(name, dash).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating dashboard: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Updated dashboard [%s].\n", got.Name)
	return emitFormattedTo(g.inv.stdout, got, "")
}

// --- Uptime impl ---

func uptimeName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/uptimeCheckConfigs/%s", project, id)
}

// buildUptimeConfig assembles an UptimeCheckConfig from a base (which may have
// been loaded from --config-from-file) and the short-form command-line flags.
// Only fields whose flags are explicitly set are overwritten so callers can
// mix a config file with a couple of overrides.
func (g *monitoringCmds) buildUptimeConfig(base *monitoring.UptimeCheckConfig, defaultDisplay string) *monitoring.UptimeCheckConfig {
	uc := base
	if uc == nil {
		uc = &monitoring.UptimeCheckConfig{}
	}
	if g.flagMonUCDisplayName != "" {
		uc.DisplayName = g.flagMonUCDisplayName
	} else if uc.DisplayName == "" && defaultDisplay != "" {
		uc.DisplayName = defaultDisplay
	}
	if g.flagMonUCPeriod != "" {
		uc.Period = g.flagMonUCPeriod
	}
	if g.flagMonUCTimeout != "" {
		uc.Timeout = g.flagMonUCTimeout
	}
	if g.flagMonUCHost != "" || g.flagMonUCPath != "" || g.flagMonUCPort > 0 || g.flagMonUCProtocol != "" {
		mrType := monitoredResourceType(g.flagMonUCResourceType)
		if g.flagMonUCHost != "" {
			uc.MonitoredResource = &monitoring.MonitoredResource{
				Type:   mrType,
				Labels: map[string]string{"host": g.flagMonUCHost},
			}
		}
		switch strings.ToLower(g.flagMonUCProtocol) {
		case "tcp":
			uc.TcpCheck = &monitoring.TcpCheck{Port: g.flagMonUCPort}
		case "https":
			uc.HttpCheck = &monitoring.HttpCheck{
				Path:   g.flagMonUCPath,
				Port:   g.flagMonUCPort,
				UseSsl: true,
			}
		default:
			uc.HttpCheck = &monitoring.HttpCheck{
				Path: g.flagMonUCPath,
				Port: g.flagMonUCPort,
			}
		}
	}
	return uc
}

func monitoredResourceType(kind string) string {
	// gcloud accepts short kebab-case names; the API wants the underscored
	// resource-type identifier.
	switch strings.ToLower(kind) {
	case "uptime-url", "":
		return "uptime_url"
	case "gce-instance":
		return "gce_instance"
	case "aws-ec2-instance":
		return "aws_ec2_instance"
	case "aws-elb-load-balancer":
		return "aws_elb_load_balancer"
	case "gae-app":
		return "gae_app"
	case "k8s-service":
		return "k8s_service"
	case "servicedirectory-service":
		return "servicedirectory_service"
	case "cloud-run-revision":
		return "cloud_run_revision"
	}
	return strings.ReplaceAll(strings.ToLower(kind), "-", "_")
}

func (g *monitoringCmds) runMonUCCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	var base *monitoring.UptimeCheckConfig
	if g.flagMonUCFile != "" {
		base = &monitoring.UptimeCheckConfig{}
		if err := loadYAMLOrJSONInto(g.flagMonUCFile, base); err != nil {
			return err
		}
	}
	uc := g.buildUptimeConfig(base, args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.UptimeCheckConfigs.Create(fmt.Sprintf("projects/%s", project), uc).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating uptime check: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Created uptime check [%s].\n", got.Name)
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *monitoringCmds) runMonUCDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.UptimeCheckConfigs.Delete(uptimeName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting uptime check: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted uptime check [%s].\n", args[0])
	return nil
}

func (g *monitoringCmds) runMonUCDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.UptimeCheckConfigs.Get(uptimeName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing uptime check: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *monitoringCmds) runMonUCList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}
	var all []*monitoring.UptimeCheckConfig
	pageToken := ""
	for {
		call := svc.Projects.UptimeCheckConfigs.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if g.flagMonUCPage > 0 {
			call = call.PageSize(g.flagMonUCPage)
		}
		if g.flagMonUCFilter != "" {
			call = call.Filter(g.flagMonUCFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing uptime checks: %w", err)
		}
		all = append(all, resp.UptimeCheckConfigs...)
		if g.flagMonUCLimit > 0 && int64(len(all)) >= g.flagMonUCLimit {
			all = all[:g.flagMonUCLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, "")
}

func (g *monitoringCmds) runMonUCUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	name := uptimeName(project, args[0])
	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}
	// Fetch the existing config so we can build an update mask from the fields
	// the user actually set on the command line.
	var base *monitoring.UptimeCheckConfig
	if g.flagMonUCFile != "" {
		base = &monitoring.UptimeCheckConfig{}
		if err := loadYAMLOrJSONInto(g.flagMonUCFile, base); err != nil {
			return err
		}
	}
	uc := g.buildUptimeConfig(base, "")
	uc.Name = name

	var mask []string
	if g.flagMonUCDisplayName != "" {
		mask = append(mask, "display_name")
	}
	if g.flagMonUCPeriod != "" {
		mask = append(mask, "period")
	}
	if g.flagMonUCTimeout != "" {
		mask = append(mask, "timeout")
	}
	if uc.HttpCheck != nil {
		mask = append(mask, "http_check")
	}
	if uc.TcpCheck != nil {
		mask = append(mask, "tcp_check")
	}
	if uc.MonitoredResource != nil {
		mask = append(mask, "monitored_resource")
	}
	call := svc.Projects.UptimeCheckConfigs.Patch(name, uc).Context(ctx)
	if len(mask) > 0 {
		call = call.UpdateMask(strings.Join(mask, ","))
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating uptime check: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Updated uptime check [%s].\n", got.Name)
	return emitFormattedTo(g.inv.stdout, got, "")
}
