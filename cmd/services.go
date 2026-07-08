package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	su "google.golang.org/api/serviceusage/v1"
)

// --- gcloud services (#284) ---

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage Cloud services",
}

var servicesEnableCmd = &cobra.Command{
	Use:   "enable SERVICE_NAME [SERVICE_NAME ...]",
	Short: "Enable one or more services for a project",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runServicesEnable,
}

var servicesDisableCmd = &cobra.Command{
	Use:   "disable SERVICE_NAME [SERVICE_NAME ...]",
	Short: "Disable one or more services for a project",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runServicesDisable,
}

var servicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List services for a project",
	Args:  cobra.NoArgs,
	RunE:  runServicesList,
}

var (
	flagServicesEnabled bool
	flagServicesAvail   bool
	flagServicesForce   bool
	flagServicesFormat  string
	flagServicesPage    int64
	flagServicesLimit   int64
)

func init() {
	servicesDisableCmd.Flags().BoolVar(&flagServicesForce, "force", false, "Also disable enabled dependent services")

	servicesListCmd.Flags().BoolVar(&flagServicesEnabled, "enabled", false, "Return services which are enabled (default)")
	servicesListCmd.Flags().BoolVar(&flagServicesAvail, "available", false, "Return services which are available to be enabled on the project")
	servicesListCmd.Flags().StringVar(&flagServicesFormat, "format", "", "Output format (json, yaml, or table)")
	servicesListCmd.Flags().Int64Var(&flagServicesPage, "page-size", 0, "Page size for API pagination")
	servicesListCmd.Flags().Int64Var(&flagServicesLimit, "limit", 0, "Maximum number of services to list (0 = no limit)")

	servicesCmd.AddCommand(servicesEnableCmd, servicesDisableCmd, servicesListCmd)
	rootCmd.AddCommand(servicesCmd)
}

// serviceResourceName builds the fully qualified name for a service enable/
// disable call: `projects/{project}/services/{serviceName}`. Accepts either
// the bare API service name (e.g. `compute.googleapis.com`) or a fully
// qualified resource name.
func serviceResourceName(project, service string) string {
	if strings.HasPrefix(service, "projects/") {
		return service
	}
	return fmt.Sprintf("projects/%s/services/%s", project, service)
}

func runServicesEnable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceUsageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		op, err := svc.Services.Enable(serviceResourceName(project, args[0]), &su.EnableServiceRequest{}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("enabling service: %w", err)
		}
		fmt.Printf("Enable service in progress (operation: %s).\n", op.Name)
		return yamlEncode(op)
	}
	op, err := svc.Services.BatchEnable("projects/"+project, &su.BatchEnableServicesRequest{
		ServiceIds: args,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling services: %w", err)
	}
	fmt.Printf("Batch enable in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runServicesDisable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceUsageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	req := &su.DisableServiceRequest{DisableDependentServices: flagServicesForce}
	// Disable does not have a batch endpoint; iterate serially so a partial
	// failure surfaces on the offending service.
	for _, s := range args {
		op, err := svc.Services.Disable(serviceResourceName(project, s), req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("disabling service %q: %w", s, err)
		}
		fmt.Printf("Disable %s in progress (operation: %s).\n", s, op.Name)
	}
	return nil
}

// servicesListFilter maps the --enabled / --available toggles to the API's
// state filter. --enabled is the default when neither is specified.
func servicesListFilter() string {
	if flagServicesAvail && !flagServicesEnabled {
		// gcloud Python treats --available as "not enabled" and returns the
		// full catalogue; the Service Usage API achieves this with no filter.
		return ""
	}
	return "state:ENABLED"
}

func runServicesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceUsageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	filter := servicesListFilter()
	var all []*su.GoogleApiServiceusageV1Service
	pageToken := ""
	for {
		call := svc.Services.List("projects/" + project).Context(ctx)
		if filter != "" {
			call = call.Filter(filter)
		}
		if flagServicesPage > 0 {
			call = call.PageSize(flagServicesPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing services: %w", err)
		}
		all = append(all, resp.Services...)
		if flagServicesLimit > 0 && int64(len(all)) >= flagServicesLimit {
			all = all[:flagServicesLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return printListResults(all, flagServicesFormat, func() {
		fmt.Printf("%-60s %-40s %s\n", "NAME", "TITLE", "STATE")
		for _, s := range all {
			title := ""
			if s.Config != nil {
				title = s.Config.Title
			}
			fmt.Printf("%-60s %-40s %s\n", s.Name, title, s.State)
		}
	})
}
