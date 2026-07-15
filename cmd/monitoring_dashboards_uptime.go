package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	monitoringv1 "google.golang.org/api/monitoring/v1"
	monitoring "google.golang.org/api/monitoring/v3"
)

// --- gcloud monitoring dashboards (#951) & uptime (#952) ---

// Common flags for the dashboards/uptime surface.
var (
	flagMonDashFile    string
	flagMonDashFilter  string
	flagMonDashPage    int64
	flagMonDashLimit   int64
	flagMonDashDisplay string

	flagMonUCFile         string
	flagMonUCDisplayName  string
	flagMonUCResourceType string
	flagMonUCHost         string
	flagMonUCPath         string
	flagMonUCPort         int64
	flagMonUCProtocol     string
	flagMonUCPeriod       string
	flagMonUCTimeout      string
	flagMonUCFilter       string
	flagMonUCPage         int64
	flagMonUCLimit        int64
)

// --- Dashboards ---

var monitoringDashboardsCmd = &cobra.Command{
	Use:   "dashboards",
	Short: "Manage Cloud Monitoring custom dashboards",
}

var monitoringDashboardsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a dashboard from a JSON/YAML definition",
	Args:  cobra.NoArgs,
	RunE:  runMonDashCreate,
}

var monitoringDashboardsDeleteCmd = &cobra.Command{
	Use:   "delete DASHBOARD",
	Short: "Delete a dashboard",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonDashDelete,
}

var monitoringDashboardsDescribeCmd = &cobra.Command{
	Use:   "describe DASHBOARD",
	Short: "Describe a dashboard",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonDashDescribe,
}

var monitoringDashboardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dashboards",
	Args:  cobra.NoArgs,
	RunE:  runMonDashList,
}

var monitoringDashboardsUpdateCmd = &cobra.Command{
	Use:   "update DASHBOARD",
	Short: "Update a dashboard from a JSON/YAML definition",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonDashUpdate,
}

// --- Uptime ---

var monitoringUptimeCmd = &cobra.Command{
	Use:   "uptime",
	Short: "Manage Cloud Monitoring uptime check configurations",
}

var monitoringUptimeCreateCmd = &cobra.Command{
	Use:   "create DISPLAY_NAME",
	Short: "Create an uptime check",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonUCCreate,
}

var monitoringUptimeDeleteCmd = &cobra.Command{
	Use:   "delete UPTIME_CHECK",
	Short: "Delete an uptime check",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonUCDelete,
}

var monitoringUptimeDescribeCmd = &cobra.Command{
	Use:   "describe UPTIME_CHECK",
	Short: "Describe an uptime check",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonUCDescribe,
}

var monitoringUptimeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List uptime checks",
	Args:  cobra.NoArgs,
	RunE:  runMonUCList,
}

var monitoringUptimeUpdateCmd = &cobra.Command{
	Use:   "update UPTIME_CHECK",
	Short: "Update an uptime check",
	Args:  cobra.ExactArgs(1),
	RunE:  runMonUCUpdate,
}

func init() {
	monitoringDashboardsCreateCmd.Flags().StringVar(&flagMonDashFile, "config-from-file", "", "YAML/JSON file with the Dashboard definition (required)")
	monitoringDashboardsCreateCmd.Flags().StringVar(&flagMonDashDisplay, "display-name", "", "Override the display name in the config file")
	monitoringDashboardsCreateCmd.MarkFlagRequired("config-from-file")

	monitoringDashboardsListCmd.Flags().StringVar(&flagMonDashFilter, "filter", "", "Server-side filter expression")
	monitoringDashboardsListCmd.Flags().Int64Var(&flagMonDashPage, "page-size", 0, "Number of results per page")
	monitoringDashboardsListCmd.Flags().Int64Var(&flagMonDashLimit, "limit", 0, "Maximum number of results to return")

	monitoringDashboardsUpdateCmd.Flags().StringVar(&flagMonDashFile, "config-from-file", "", "YAML/JSON file with the Dashboard patch (required)")
	monitoringDashboardsUpdateCmd.Flags().StringVar(&flagMonDashDisplay, "display-name", "", "Override the display name in the config file")
	monitoringDashboardsUpdateCmd.MarkFlagRequired("config-from-file")

	monitoringDashboardsCmd.AddCommand(
		monitoringDashboardsCreateCmd, monitoringDashboardsDeleteCmd,
		monitoringDashboardsDescribeCmd, monitoringDashboardsListCmd,
		monitoringDashboardsUpdateCmd,
	)

	// Uptime create/update: users may pass a config file OR use short flags for
	// common cases (http/https/tcp against a URL/host).
	for _, c := range []*cobra.Command{monitoringUptimeCreateCmd, monitoringUptimeUpdateCmd} {
		c.Flags().StringVar(&flagMonUCFile, "config-from-file", "", "YAML/JSON file with the UptimeCheckConfig")
		c.Flags().StringVar(&flagMonUCDisplayName, "display-name", "", "Display name")
		c.Flags().StringVar(&flagMonUCResourceType, "resource-type", "uptime-url", "Monitored resource type (uptime-url, gce-instance, aws-ec2-instance, ...)")
		c.Flags().StringVar(&flagMonUCHost, "host", "", "Host or URL to check")
		c.Flags().StringVar(&flagMonUCPath, "path", "", "HTTP path")
		c.Flags().Int64Var(&flagMonUCPort, "port", 0, "TCP or HTTP port")
		c.Flags().StringVar(&flagMonUCProtocol, "protocol", "http", "Protocol (http, https, tcp)")
		c.Flags().StringVar(&flagMonUCPeriod, "period", "", "Check period in seconds (60, 300, 600, 900); e.g. 60s")
		c.Flags().StringVar(&flagMonUCTimeout, "timeout", "", "Timeout in seconds; e.g. 10s")
	}
	monitoringUptimeListCmd.Flags().StringVar(&flagMonUCFilter, "filter", "", "Server-side filter expression")
	monitoringUptimeListCmd.Flags().Int64Var(&flagMonUCPage, "page-size", 0, "Number of results per page")
	monitoringUptimeListCmd.Flags().Int64Var(&flagMonUCLimit, "limit", 0, "Maximum number of results to return")

	monitoringUptimeCmd.AddCommand(
		monitoringUptimeCreateCmd, monitoringUptimeDeleteCmd,
		monitoringUptimeDescribeCmd, monitoringUptimeListCmd, monitoringUptimeUpdateCmd,
	)

	monitoringCmd.AddCommand(monitoringDashboardsCmd, monitoringUptimeCmd)
}

// --- Dashboards impl ---

func dashboardName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/dashboards/%s", project, id)
}

func runMonDashCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	dash := &monitoringv1.Dashboard{}
	if err := loadYAMLOrJSONInto(flagMonDashFile, dash); err != nil {
		return err
	}
	if flagMonDashDisplay != "" {
		dash.DisplayName = flagMonDashDisplay
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Dashboards.Create(fmt.Sprintf("projects/%s", project), dash).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating dashboard: %w", err)
	}
	fmt.Printf("Created dashboard [%s].\n", got.Name)
	return emitFormatted(got, "")
}

func runMonDashDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Dashboards.Delete(dashboardName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting dashboard: %w", err)
	}
	fmt.Printf("Deleted dashboard [%s].\n", args[0])
	return nil
}

func runMonDashDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Dashboards.Get(dashboardName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing dashboard: %w", err)
	}
	return emitFormatted(got, "")
}

func runMonDashList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*monitoringv1.Dashboard
	pageToken := ""
	for {
		call := svc.Projects.Dashboards.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagMonDashPage > 0 {
			call = call.PageSize(flagMonDashPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing dashboards: %w", err)
		}
		all = append(all, resp.Dashboards...)
		if flagMonDashLimit > 0 && int64(len(all)) >= flagMonDashLimit {
			all = all[:flagMonDashLimit]
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
	if flagMonDashFilter != "" {
		needle := strings.ToLower(flagMonDashFilter)
		filtered := all[:0]
		for _, d := range all {
			if strings.Contains(strings.ToLower(d.DisplayName), needle) {
				filtered = append(filtered, d)
			}
		}
		all = filtered
	}
	return emitFormatted(all, "")
}

func runMonDashUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	dash := &monitoringv1.Dashboard{}
	if err := loadYAMLOrJSONInto(flagMonDashFile, dash); err != nil {
		return err
	}
	if flagMonDashDisplay != "" {
		dash.DisplayName = flagMonDashDisplay
	}
	name := dashboardName(project, args[0])
	// Dashboards.Patch expects the resource's Name to match the URL.
	dash.Name = name
	ctx := context.Background()
	svc, err := gcp.MonitoringV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Dashboards.Patch(name, dash).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating dashboard: %w", err)
	}
	fmt.Printf("Updated dashboard [%s].\n", got.Name)
	return emitFormatted(got, "")
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
func buildUptimeConfig(base *monitoring.UptimeCheckConfig, defaultDisplay string) *monitoring.UptimeCheckConfig {
	uc := base
	if uc == nil {
		uc = &monitoring.UptimeCheckConfig{}
	}
	if flagMonUCDisplayName != "" {
		uc.DisplayName = flagMonUCDisplayName
	} else if uc.DisplayName == "" && defaultDisplay != "" {
		uc.DisplayName = defaultDisplay
	}
	if flagMonUCPeriod != "" {
		uc.Period = flagMonUCPeriod
	}
	if flagMonUCTimeout != "" {
		uc.Timeout = flagMonUCTimeout
	}
	if flagMonUCHost != "" || flagMonUCPath != "" || flagMonUCPort > 0 || flagMonUCProtocol != "" {
		mrType := monitoredResourceType(flagMonUCResourceType)
		if flagMonUCHost != "" {
			uc.MonitoredResource = &monitoring.MonitoredResource{
				Type:   mrType,
				Labels: map[string]string{"host": flagMonUCHost},
			}
		}
		switch strings.ToLower(flagMonUCProtocol) {
		case "tcp":
			uc.TcpCheck = &monitoring.TcpCheck{Port: flagMonUCPort}
		case "https":
			uc.HttpCheck = &monitoring.HttpCheck{
				Path:   flagMonUCPath,
				Port:   flagMonUCPort,
				UseSsl: true,
			}
		default:
			uc.HttpCheck = &monitoring.HttpCheck{
				Path: flagMonUCPath,
				Port: flagMonUCPort,
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

func runMonUCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var base *monitoring.UptimeCheckConfig
	if flagMonUCFile != "" {
		base = &monitoring.UptimeCheckConfig{}
		if err := loadYAMLOrJSONInto(flagMonUCFile, base); err != nil {
			return err
		}
	}
	uc := buildUptimeConfig(base, args[0])
	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.UptimeCheckConfigs.Create(fmt.Sprintf("projects/%s", project), uc).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating uptime check: %w", err)
	}
	fmt.Printf("Created uptime check [%s].\n", got.Name)
	return emitFormatted(got, "")
}

func runMonUCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.UptimeCheckConfigs.Delete(uptimeName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting uptime check: %w", err)
	}
	fmt.Printf("Deleted uptime check [%s].\n", args[0])
	return nil
}

func runMonUCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.UptimeCheckConfigs.Get(uptimeName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing uptime check: %w", err)
	}
	return emitFormatted(got, "")
}

func runMonUCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*monitoring.UptimeCheckConfig
	pageToken := ""
	for {
		call := svc.Projects.UptimeCheckConfigs.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagMonUCPage > 0 {
			call = call.PageSize(flagMonUCPage)
		}
		if flagMonUCFilter != "" {
			call = call.Filter(flagMonUCFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing uptime checks: %w", err)
		}
		all = append(all, resp.UptimeCheckConfigs...)
		if flagMonUCLimit > 0 && int64(len(all)) >= flagMonUCLimit {
			all = all[:flagMonUCLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runMonUCUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := uptimeName(project, args[0])
	ctx := context.Background()
	svc, err := gcp.MonitoringService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// Fetch the existing config so we can build an update mask from the fields
	// the user actually set on the command line.
	var base *monitoring.UptimeCheckConfig
	if flagMonUCFile != "" {
		base = &monitoring.UptimeCheckConfig{}
		if err := loadYAMLOrJSONInto(flagMonUCFile, base); err != nil {
			return err
		}
	}
	uc := buildUptimeConfig(base, "")
	uc.Name = name

	var mask []string
	if flagMonUCDisplayName != "" {
		mask = append(mask, "display_name")
	}
	if flagMonUCPeriod != "" {
		mask = append(mask, "period")
	}
	if flagMonUCTimeout != "" {
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
	fmt.Printf("Updated uptime check [%s].\n", got.Name)
	return emitFormatted(got, "")
}
