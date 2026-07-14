package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apphub "google.golang.org/api/apphub/v1"
)

func ahLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func ahChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func ahWaitOp(ctx context.Context, svc *apphub.APIService, op *apphub.Operation) (*apphub.Operation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func ahFinishOp(ctx context.Context, svc *apphub.APIService, op *apphub.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := ahWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

var (
	flagAHLocation   string
	flagAHConfigFile string
	flagAHUpdateMask string
	flagAHFormat     string
	flagAHAsync      bool
)

// --- locations ---

var apphubLocationsCmd = &cobra.Command{Use: "locations", Short: "Explore App Hub locations"}

var (
	ahLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe an App Hub location",
		Args: cobra.ExactArgs(1), RunE: runAHLocDescribe,
	}
	ahLocListCmd = &cobra.Command{
		Use: "list", Short: "List App Hub locations",
		Args: cobra.NoArgs, RunE: runAHLocList,
	}
)

func init() {
	ahLocDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	ahLocListCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	apphubLocationsCmd.AddCommand(ahLocDescribeCmd, ahLocListCmd)
	apphubCmd.AddCommand(apphubLocationsCmd)
}

func runAHLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(ahLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(got, flagAHFormat)
}

func runAHLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if flagAHFormat != "" {
		return emitFormatted(resp.Locations, flagAHFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- operations ---

var apphubOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage App Hub operations"}

var (
	ahOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an App Hub operation",
		Args: cobra.ExactArgs(1), RunE: runAHOpCancel,
	}
	ahOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an App Hub operation",
		Args: cobra.ExactArgs(1), RunE: runAHOpDelete,
	}
	ahOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an App Hub operation",
		Args: cobra.ExactArgs(1), RunE: runAHOpDescribe,
	}
	ahOpListCmd = &cobra.Command{
		Use: "list", Short: "List App Hub operations in a location",
		Args: cobra.NoArgs, RunE: runAHOpList,
	}
)

func init() {
	for _, c := range []*cobra.Command{ahOpCancelCmd, ahOpDeleteCmd, ahOpDescribeCmd, ahOpListCmd} {
		c.Flags().StringVar(&flagAHLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	ahOpDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	ahOpListCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	apphubOperationsCmd.AddCommand(ahOpCancelCmd, ahOpDeleteCmd, ahOpDescribeCmd, ahOpListCmd)
	apphubCmd.AddCommand(apphubOperationsCmd)
}

func ahOpName(id, project, location string) string {
	return ahChild("operations", id, ahLocationParent(project, location))
}

func runAHOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(ahOpName(args[0], project, flagAHLocation), &apphub.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runAHOpDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(ahOpName(args[0], project, flagAHLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runAHOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(ahOpName(args[0], project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagAHFormat)
}

func runAHOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(ahLocationParent(project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagAHFormat != "" {
		return emitFormatted(resp.Operations, flagAHFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- boundary ---

var apphubBoundaryCmd = &cobra.Command{Use: "boundary", Short: "Manage App Hub location boundaries"}

var (
	ahBoundaryDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the boundary for a location",
		Args: cobra.NoArgs, RunE: runAHBoundaryDescribe,
	}
	ahBoundaryUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update the boundary from a --config-file",
		Args: cobra.NoArgs, RunE: runAHBoundaryUpdate,
	}
)

func init() {
	for _, c := range []*cobra.Command{ahBoundaryDescribeCmd, ahBoundaryUpdateCmd} {
		c.Flags().StringVar(&flagAHLocation, "location", "", "Location containing the boundary (required)")
		_ = c.MarkFlagRequired("location")
	}
	ahBoundaryUpdateCmd.Flags().StringVar(&flagAHConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Boundary message body (required)")
	_ = ahBoundaryUpdateCmd.MarkFlagRequired("config-file")
	ahBoundaryUpdateCmd.Flags().StringVar(&flagAHUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	ahBoundaryUpdateCmd.Flags().BoolVar(&flagAHAsync, "async", false, "Return the long-running operation without waiting")
	ahBoundaryDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")

	apphubBoundaryCmd.AddCommand(ahBoundaryDescribeCmd, ahBoundaryUpdateCmd)
	apphubCmd.AddCommand(apphubBoundaryCmd)
}

func ahBoundaryName(project, location string) string {
	return fmt.Sprintf("%s/boundary", ahLocationParent(project, location))
}

func runAHBoundaryDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GetBoundary(ahBoundaryName(project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing boundary: %w", err)
	}
	return emitFormatted(got, flagAHFormat)
}

func runAHBoundaryUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	b := &apphub.Boundary{}
	if err := loadYAMLOrJSONInto(flagAHConfigFile, b); err != nil {
		return err
	}
	mask := flagAHUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(b))
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.UpdateBoundary(ahBoundaryName(project, flagAHLocation), b).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating boundary: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Update boundary", flagAHLocation, flagAHAsync)
}

// --- applications (with services + workloads subgroups) ---

var apphubApplicationsCmd = &cobra.Command{Use: "applications", Short: "Manage App Hub applications"}

var (
	ahAppCreateCmd = &cobra.Command{
		Use: "create APPLICATION", Short: "Create an application from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAHAppCreate,
	}
	ahAppDeleteCmd = &cobra.Command{
		Use: "delete APPLICATION", Short: "Delete an application",
		Args: cobra.ExactArgs(1), RunE: runAHAppDelete,
	}
	ahAppDescribeCmd = &cobra.Command{
		Use: "describe APPLICATION", Short: "Describe an application",
		Args: cobra.ExactArgs(1), RunE: runAHAppDescribe,
	}
	ahAppListCmd = &cobra.Command{
		Use: "list", Short: "List applications in a location",
		Args: cobra.NoArgs, RunE: runAHAppList,
	}
	ahAppUpdateCmd = &cobra.Command{
		Use: "update APPLICATION", Short: "Update an application from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAHAppUpdate,
	}
)

func init() {
	all := []*cobra.Command{ahAppCreateCmd, ahAppDeleteCmd, ahAppDescribeCmd, ahAppListCmd, ahAppUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagAHLocation, "location", "", "Location containing the application (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{ahAppCreateCmd, ahAppUpdateCmd} {
		c.Flags().StringVar(&flagAHConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Application message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	ahAppUpdateCmd.Flags().StringVar(&flagAHUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{ahAppCreateCmd, ahAppDeleteCmd, ahAppUpdateCmd} {
		c.Flags().BoolVar(&flagAHAsync, "async", false, "Return the long-running operation without waiting")
	}
	ahAppDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	ahAppListCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")

	apphubApplicationsCmd.AddCommand(all...)
	registerAHAppServices(apphubApplicationsCmd)
	registerAHAppWorkloads(apphubApplicationsCmd)
	apphubCmd.AddCommand(apphubApplicationsCmd)
}

func ahAppName(id, project, location string) string {
	return ahChild("applications", id, ahLocationParent(project, location))
}

func runAHAppCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	app := &apphub.Application{}
	if err := loadYAMLOrJSONInto(flagAHConfigFile, app); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Create(ahLocationParent(project, flagAHLocation), app).
		ApplicationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating application: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Create application", args[0], flagAHAsync)
}

func runAHAppDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Delete(ahAppName(args[0], project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Delete application", args[0], flagAHAsync)
}

func runAHAppDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Applications.Get(ahAppName(args[0], project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing application: %w", err)
	}
	return emitFormatted(got, flagAHFormat)
}

func runAHAppList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Applications.List(ahLocationParent(project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing applications: %w", err)
	}
	if flagAHFormat != "" {
		return emitFormatted(resp.Applications, flagAHFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, a := range resp.Applications {
		fmt.Printf("%-40s %s\n", path.Base(a.Name), a.DisplayName)
	}
	return nil
}

func runAHAppUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	app := &apphub.Application{}
	if err := loadYAMLOrJSONInto(flagAHConfigFile, app); err != nil {
		return err
	}
	mask := flagAHUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(app))
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Patch(ahAppName(args[0], project, flagAHLocation), app).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Update application", args[0], flagAHAsync)
}

// --- applications services ---

var apphubAppServicesCmd = &cobra.Command{Use: "services", Short: "Manage application services"}

var (
	flagAHAppID string
)

var (
	ahAppSvcCreateCmd = &cobra.Command{
		Use: "create SERVICE", Short: "Create a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAHAppSvcCreate,
	}
	ahAppSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a service",
		Args: cobra.ExactArgs(1), RunE: runAHAppSvcDelete,
	}
	ahAppSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: runAHAppSvcDescribe,
	}
	ahAppSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services in an application",
		Args: cobra.NoArgs, RunE: runAHAppSvcList,
	}
	ahAppSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAHAppSvcUpdate,
	}
)

func registerAHAppServices(parent *cobra.Command) {
	all := []*cobra.Command{ahAppSvcCreateCmd, ahAppSvcDeleteCmd, ahAppSvcDescribeCmd, ahAppSvcListCmd, ahAppSvcUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagAHLocation, "location", "", "Location containing the application (required)")
		c.Flags().StringVar(&flagAHAppID, "application", "", "Application containing the service (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("application")
	}
	for _, c := range []*cobra.Command{ahAppSvcCreateCmd, ahAppSvcUpdateCmd} {
		c.Flags().StringVar(&flagAHConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Service message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	ahAppSvcUpdateCmd.Flags().StringVar(&flagAHUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{ahAppSvcCreateCmd, ahAppSvcDeleteCmd, ahAppSvcUpdateCmd} {
		c.Flags().BoolVar(&flagAHAsync, "async", false, "Return the long-running operation without waiting")
	}
	ahAppSvcDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	ahAppSvcListCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")

	apphubAppServicesCmd.AddCommand(all...)
	parent.AddCommand(apphubAppServicesCmd)
}

func ahAppSvcParent(project, location, app string) string {
	return ahAppName(app, project, location)
}

func ahAppSvcName(id, project, location, app string) string {
	return ahChild("services", id, ahAppSvcParent(project, location, app))
}

func runAHAppSvcCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &apphub.Service{}
	if err := loadYAMLOrJSONInto(flagAHConfigFile, s); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Services.Create(ahAppSvcParent(project, flagAHLocation, flagAHAppID), s).
		ServiceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Create service", args[0], flagAHAsync)
}

func runAHAppSvcDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Services.Delete(ahAppSvcName(args[0], project, flagAHLocation, flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Delete service", args[0], flagAHAsync)
}

func runAHAppSvcDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Applications.Services.Get(ahAppSvcName(args[0], project, flagAHLocation, flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormatted(got, flagAHFormat)
}

func runAHAppSvcList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Applications.Services.List(ahAppSvcParent(project, flagAHLocation, flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	if flagAHFormat != "" {
		return emitFormatted(resp.Services, flagAHFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, s := range resp.Services {
		fmt.Println(path.Base(s.Name))
	}
	return nil
}

func runAHAppSvcUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &apphub.Service{}
	if err := loadYAMLOrJSONInto(flagAHConfigFile, s); err != nil {
		return err
	}
	mask := flagAHUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Services.Patch(ahAppSvcName(args[0], project, flagAHLocation, flagAHAppID), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Update service", args[0], flagAHAsync)
}

// --- applications workloads ---

var apphubAppWorkloadsCmd = &cobra.Command{Use: "workloads", Short: "Manage application workloads"}

var (
	ahAppWLCreateCmd = &cobra.Command{
		Use: "create WORKLOAD", Short: "Create a workload from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAHAppWLCreate,
	}
	ahAppWLDeleteCmd = &cobra.Command{
		Use: "delete WORKLOAD", Short: "Delete a workload",
		Args: cobra.ExactArgs(1), RunE: runAHAppWLDelete,
	}
	ahAppWLDescribeCmd = &cobra.Command{
		Use: "describe WORKLOAD", Short: "Describe a workload",
		Args: cobra.ExactArgs(1), RunE: runAHAppWLDescribe,
	}
	ahAppWLListCmd = &cobra.Command{
		Use: "list", Short: "List workloads in an application",
		Args: cobra.NoArgs, RunE: runAHAppWLList,
	}
	ahAppWLUpdateCmd = &cobra.Command{
		Use: "update WORKLOAD", Short: "Update a workload from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAHAppWLUpdate,
	}
)

func registerAHAppWorkloads(parent *cobra.Command) {
	all := []*cobra.Command{ahAppWLCreateCmd, ahAppWLDeleteCmd, ahAppWLDescribeCmd, ahAppWLListCmd, ahAppWLUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagAHLocation, "location", "", "Location containing the application (required)")
		c.Flags().StringVar(&flagAHAppID, "application", "", "Application containing the workload (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("application")
	}
	for _, c := range []*cobra.Command{ahAppWLCreateCmd, ahAppWLUpdateCmd} {
		c.Flags().StringVar(&flagAHConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Workload message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	ahAppWLUpdateCmd.Flags().StringVar(&flagAHUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{ahAppWLCreateCmd, ahAppWLDeleteCmd, ahAppWLUpdateCmd} {
		c.Flags().BoolVar(&flagAHAsync, "async", false, "Return the long-running operation without waiting")
	}
	ahAppWLDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	ahAppWLListCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")

	apphubAppWorkloadsCmd.AddCommand(all...)
	parent.AddCommand(apphubAppWorkloadsCmd)
}

func ahAppWLName(id, project, location, app string) string {
	return ahChild("workloads", id, ahAppSvcParent(project, location, app))
}

func runAHAppWLCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	wl := &apphub.Workload{}
	if err := loadYAMLOrJSONInto(flagAHConfigFile, wl); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Workloads.Create(ahAppSvcParent(project, flagAHLocation, flagAHAppID), wl).
		WorkloadId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workload: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Create workload", args[0], flagAHAsync)
}

func runAHAppWLDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Workloads.Delete(ahAppWLName(args[0], project, flagAHLocation, flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workload: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Delete workload", args[0], flagAHAsync)
}

func runAHAppWLDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Applications.Workloads.Get(ahAppWLName(args[0], project, flagAHLocation, flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workload: %w", err)
	}
	return emitFormatted(got, flagAHFormat)
}

func runAHAppWLList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Applications.Workloads.List(ahAppSvcParent(project, flagAHLocation, flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing workloads: %w", err)
	}
	if flagAHFormat != "" {
		return emitFormatted(resp.Workloads, flagAHFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, w := range resp.Workloads {
		fmt.Println(path.Base(w.Name))
	}
	return nil
}

func runAHAppWLUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	wl := &apphub.Workload{}
	if err := loadYAMLOrJSONInto(flagAHConfigFile, wl); err != nil {
		return err
	}
	mask := flagAHUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(wl))
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Workloads.Patch(ahAppWLName(args[0], project, flagAHLocation, flagAHAppID), wl).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating workload: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Update workload", args[0], flagAHAsync)
}

// --- discovered-services ---

var apphubDiscoveredServicesCmd = &cobra.Command{Use: "discovered-services", Short: "View discovered App Hub services"}

var (
	ahDSDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a discovered service",
		Args: cobra.ExactArgs(1), RunE: runAHDSDescribe,
	}
	ahDSListCmd = &cobra.Command{
		Use: "list", Short: "List discovered services in a location",
		Args: cobra.NoArgs, RunE: runAHDSList,
	}
	ahDSLookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up a discovered service by URI",
		Args: cobra.NoArgs, RunE: runAHDSLookup,
	}
)

var flagAHDSURI string

func init() {
	for _, c := range []*cobra.Command{ahDSDescribeCmd, ahDSListCmd, ahDSLookupCmd} {
		c.Flags().StringVar(&flagAHLocation, "location", "", "Location containing the discovered service (required)")
		_ = c.MarkFlagRequired("location")
	}
	ahDSLookupCmd.Flags().StringVar(&flagAHDSURI, "uri", "", "URI to look up (required)")
	_ = ahDSLookupCmd.MarkFlagRequired("uri")
	ahDSDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	ahDSListCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")

	apphubDiscoveredServicesCmd.AddCommand(ahDSDescribeCmd, ahDSListCmd, ahDSLookupCmd)
	apphubCmd.AddCommand(apphubDiscoveredServicesCmd)
}

func ahDSName(id, project, location string) string {
	return ahChild("discoveredServices", id, ahLocationParent(project, location))
}

func runAHDSDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DiscoveredServices.Get(ahDSName(args[0], project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing discovered service: %w", err)
	}
	return emitFormatted(got, flagAHFormat)
}

func runAHDSList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DiscoveredServices.List(ahLocationParent(project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing discovered services: %w", err)
	}
	if flagAHFormat != "" {
		return emitFormatted(resp.DiscoveredServices, flagAHFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, ds := range resp.DiscoveredServices {
		fmt.Println(path.Base(ds.Name))
	}
	return nil
}

func runAHDSLookup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DiscoveredServices.Lookup(ahLocationParent(project, flagAHLocation)).Uri(flagAHDSURI).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up discovered service: %w", err)
	}
	return emitFormatted(resp, flagAHFormat)
}

// --- discovered-workloads ---

var apphubDiscoveredWorkloadsCmd = &cobra.Command{Use: "discovered-workloads", Short: "View discovered App Hub workloads"}

var (
	ahDWDescribeCmd = &cobra.Command{
		Use: "describe WORKLOAD", Short: "Describe a discovered workload",
		Args: cobra.ExactArgs(1), RunE: runAHDWDescribe,
	}
	ahDWListCmd = &cobra.Command{
		Use: "list", Short: "List discovered workloads in a location",
		Args: cobra.NoArgs, RunE: runAHDWList,
	}
	ahDWLookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up a discovered workload by URI",
		Args: cobra.NoArgs, RunE: runAHDWLookup,
	}
)

var flagAHDWURI string

func init() {
	for _, c := range []*cobra.Command{ahDWDescribeCmd, ahDWListCmd, ahDWLookupCmd} {
		c.Flags().StringVar(&flagAHLocation, "location", "", "Location containing the discovered workload (required)")
		_ = c.MarkFlagRequired("location")
	}
	ahDWLookupCmd.Flags().StringVar(&flagAHDWURI, "uri", "", "URI to look up (required)")
	_ = ahDWLookupCmd.MarkFlagRequired("uri")
	ahDWDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	ahDWListCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")

	apphubDiscoveredWorkloadsCmd.AddCommand(ahDWDescribeCmd, ahDWListCmd, ahDWLookupCmd)
	apphubCmd.AddCommand(apphubDiscoveredWorkloadsCmd)
}

func ahDWName(id, project, location string) string {
	return ahChild("discoveredWorkloads", id, ahLocationParent(project, location))
}

func runAHDWDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DiscoveredWorkloads.Get(ahDWName(args[0], project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing discovered workload: %w", err)
	}
	return emitFormatted(got, flagAHFormat)
}

func runAHDWList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DiscoveredWorkloads.List(ahLocationParent(project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing discovered workloads: %w", err)
	}
	if flagAHFormat != "" {
		return emitFormatted(resp.DiscoveredWorkloads, flagAHFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, dw := range resp.DiscoveredWorkloads {
		fmt.Println(path.Base(dw.Name))
	}
	return nil
}

func runAHDWLookup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DiscoveredWorkloads.Lookup(ahLocationParent(project, flagAHLocation)).Uri(flagAHDWURI).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up discovered workload: %w", err)
	}
	return emitFormatted(resp, flagAHFormat)
}

// --- service-projects ---

var apphubServiceProjectsCmd = &cobra.Command{Use: "service-projects", Short: "Manage App Hub service project attachments"}

var (
	ahSPAddCmd = &cobra.Command{
		Use: "add SERVICE_PROJECT", Short: "Attach a service project",
		Args: cobra.ExactArgs(1), RunE: runAHSPAdd,
	}
	ahSPRemoveCmd = &cobra.Command{
		Use: "remove SERVICE_PROJECT", Short: "Detach a service project",
		Args: cobra.ExactArgs(1), RunE: runAHSPRemove,
	}
	ahSPDescribeCmd = &cobra.Command{
		Use: "describe SERVICE_PROJECT", Short: "Describe a service project attachment",
		Args: cobra.ExactArgs(1), RunE: runAHSPDescribe,
	}
	ahSPListCmd = &cobra.Command{
		Use: "list", Short: "List service project attachments in a location",
		Args: cobra.NoArgs, RunE: runAHSPList,
	}
	ahSPLookupCmd = &cobra.Command{
		Use: "lookup SERVICE_PROJECT", Short: "Look up a service project attachment by project ID",
		Args: cobra.ExactArgs(1), RunE: runAHSPLookup,
	}
)

func init() {
	all := []*cobra.Command{ahSPAddCmd, ahSPRemoveCmd, ahSPDescribeCmd, ahSPListCmd, ahSPLookupCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagAHLocation, "location", "", "Location containing the service-project attachment (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{ahSPAddCmd, ahSPRemoveCmd} {
		c.Flags().BoolVar(&flagAHAsync, "async", false, "Return the long-running operation without waiting")
	}
	ahSPDescribeCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")
	ahSPListCmd.Flags().StringVar(&flagAHFormat, "format", "", "Output format")

	apphubServiceProjectsCmd.AddCommand(all...)
	apphubCmd.AddCommand(apphubServiceProjectsCmd)
}

func ahSPName(id, project, location string) string {
	return ahChild("serviceProjectAttachments", id, ahLocationParent(project, location))
}

func runAHSPAdd(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	attach := &apphub.ServiceProjectAttachment{
		ServiceProject: fmt.Sprintf("projects/%s", args[0]),
	}
	op, err := svc.Projects.Locations.ServiceProjectAttachments.Create(ahLocationParent(project, flagAHLocation), attach).
		ServiceProjectAttachmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding service project: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Add service project", args[0], flagAHAsync)
}

func runAHSPRemove(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceProjectAttachments.Delete(ahSPName(args[0], project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing service project: %w", err)
	}
	return ahFinishOp(ctx, svc, op, "Remove service project", args[0], flagAHAsync)
}

func runAHSPDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceProjectAttachments.Get(ahSPName(args[0], project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service project attachment: %w", err)
	}
	return emitFormatted(got, flagAHFormat)
}

func runAHSPList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.ServiceProjectAttachments.List(ahLocationParent(project, flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing service project attachments: %w", err)
	}
	if flagAHFormat != "" {
		return emitFormatted(resp.ServiceProjectAttachments, flagAHFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "SERVICE_PROJECT")
	for _, sp := range resp.ServiceProjectAttachments {
		fmt.Printf("%-40s %s\n", path.Base(sp.Name), sp.ServiceProject)
	}
	return nil
}

func runAHSPLookup(cmd *cobra.Command, args []string) error {
	if _, err := resolveProject(); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// LookupServiceProjectAttachment takes the service project resource name
	// directly ("projects/PROJECT/locations/…"), not the App Hub host project.
	resp, err := svc.Projects.Locations.LookupServiceProjectAttachment(fmt.Sprintf("projects/%s/locations/%s", args[0], flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up service project attachment: %w", err)
	}
	return emitFormatted(resp, flagAHFormat)
}
