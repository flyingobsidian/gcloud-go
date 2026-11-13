package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

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

func (g *apphubCmds) ahFinishOp(ctx context.Context, svc *apphub.APIService, op *apphub.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(g.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(g.inv.stdout, op, "")
	}
	final, err := ahWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(g.inv.stdout, final.Response, "")
	}
	return nil
}

// --- locations ---

func (g *apphubCmds) initApphubAll() {
	g.ahLocDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.ahLocListCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.apphubLocationsCmd.AddCommand(g.ahLocDescribeCmd, g.ahLocListCmd)
	g.apphubCmd.AddCommand(g.apphubLocationsCmd)
}

func (g *apphubCmds) runAHLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(ahLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHFormat)
}

func (g *apphubCmds) runAHLocList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if g.flagAHFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Locations, g.flagAHFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Locations), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(g.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- operations ---

func (g *apphubCmds) initApphubAll2() {
	for _, c := range []*cobra.Command{g.ahOpCancelCmd, g.ahOpDeleteCmd, g.ahOpDescribeCmd, g.ahOpListCmd} {
		c.Flags().StringVar(&g.flagAHLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	g.ahOpDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.ahOpListCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.apphubOperationsCmd.AddCommand(g.ahOpCancelCmd, g.ahOpDeleteCmd, g.ahOpDescribeCmd, g.ahOpListCmd)
	g.apphubCmd.AddCommand(g.apphubOperationsCmd)
}

func ahOpName(id, project, location string) string {
	return ahChild("operations", id, ahLocationParent(project, location))
}

func (g *apphubCmds) runAHOpCancel(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(ahOpName(args[0], project, g.flagAHLocation), &apphub.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (g *apphubCmds) runAHOpDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(ahOpName(args[0], project, g.flagAHLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (g *apphubCmds) runAHOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(ahOpName(args[0], project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagAHFormat)
}

func (g *apphubCmds) runAHOpList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(ahLocationParent(project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if g.flagAHFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Operations, g.flagAHFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(g.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- boundary ---

func (g *apphubCmds) initApphubAll3() {
	for _, c := range []*cobra.Command{g.ahBoundaryDescribeCmd, g.ahBoundaryUpdateCmd} {
		c.Flags().StringVar(&g.flagAHLocation, "location", "", "Location containing the boundary (required)")
		_ = c.MarkFlagRequired("location")
	}
	g.ahBoundaryUpdateCmd.Flags().StringVar(&g.flagAHConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Boundary message body (required)")
	_ = g.ahBoundaryUpdateCmd.MarkFlagRequired("config-file")
	g.ahBoundaryUpdateCmd.Flags().StringVar(&g.flagAHUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	g.ahBoundaryUpdateCmd.Flags().BoolVar(&g.flagAHAsync, "async", false, "Return the long-running operation without waiting")
	g.ahBoundaryDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")

	g.apphubBoundaryCmd.AddCommand(g.ahBoundaryDescribeCmd, g.ahBoundaryUpdateCmd)
	g.apphubCmd.AddCommand(g.apphubBoundaryCmd)
}

func ahBoundaryName(project, location string) string {
	return fmt.Sprintf("%s/boundary", ahLocationParent(project, location))
}

func (g *apphubCmds) runAHBoundaryDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GetBoundary(ahBoundaryName(project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing boundary: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHFormat)
}

func (g *apphubCmds) runAHBoundaryUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	b := &apphub.Boundary{}
	if err := loadYAMLOrJSONInto(g.flagAHConfigFile, b); err != nil {
		return err
	}
	mask := g.flagAHUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(b))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.UpdateBoundary(ahBoundaryName(project, g.flagAHLocation), b).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating boundary: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Update boundary", g.flagAHLocation, g.flagAHAsync)
}

// --- applications (with services + workloads subgroups) ---

func (g *apphubCmds) initApphubAll4() {
	all := []*cobra.Command{g.ahAppCreateCmd, g.ahAppDeleteCmd, g.ahAppDescribeCmd, g.ahAppListCmd, g.ahAppUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagAHLocation, "location", "", "Location containing the application (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.ahAppCreateCmd, g.ahAppUpdateCmd} {
		c.Flags().StringVar(&g.flagAHConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Application message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.ahAppUpdateCmd.Flags().StringVar(&g.flagAHUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.ahAppCreateCmd, g.ahAppDeleteCmd, g.ahAppUpdateCmd} {
		c.Flags().BoolVar(&g.flagAHAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.ahAppDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.ahAppListCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")

	g.apphubApplicationsCmd.AddCommand(all...)
	g.registerAHAppServices(g.apphubApplicationsCmd)
	g.registerAHAppWorkloads(g.apphubApplicationsCmd)
	g.apphubCmd.AddCommand(g.apphubApplicationsCmd)
}

func ahAppName(id, project, location string) string {
	return ahChild("applications", id, ahLocationParent(project, location))
}

func (g *apphubCmds) runAHAppCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	app := &apphub.Application{}
	if err := loadYAMLOrJSONInto(g.flagAHConfigFile, app); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Create(ahLocationParent(project, g.flagAHLocation), app).
		ApplicationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating application: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Create application", args[0], g.flagAHAsync)
}

func (g *apphubCmds) runAHAppDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Delete(ahAppName(args[0], project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Delete application", args[0], g.flagAHAsync)
}

func (g *apphubCmds) runAHAppDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Applications.Get(ahAppName(args[0], project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing application: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHFormat)
}

func (g *apphubCmds) runAHAppList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Applications.List(ahLocationParent(project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing applications: %w", err)
	}
	if g.flagAHFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Applications, g.flagAHFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Applications), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, a := range resp.Applications {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(a.Name), a.DisplayName)
	}
	return nil
}

func (g *apphubCmds) runAHAppUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	app := &apphub.Application{}
	if err := loadYAMLOrJSONInto(g.flagAHConfigFile, app); err != nil {
		return err
	}
	mask := g.flagAHUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(app))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Patch(ahAppName(args[0], project, g.flagAHLocation), app).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Update application", args[0], g.flagAHAsync)
}

// --- applications services ---

func (g *apphubCmds) registerAHAppServices(parent *cobra.Command) {
	all := []*cobra.Command{g.ahAppSvcCreateCmd, g.ahAppSvcDeleteCmd, g.ahAppSvcDescribeCmd, g.ahAppSvcListCmd, g.ahAppSvcUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagAHLocation, "location", "", "Location containing the application (required)")
		c.Flags().StringVar(&g.flagAHAppID, "application", "", "Application containing the service (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("application")
	}
	for _, c := range []*cobra.Command{g.ahAppSvcCreateCmd, g.ahAppSvcUpdateCmd} {
		c.Flags().StringVar(&g.flagAHConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Service message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.ahAppSvcUpdateCmd.Flags().StringVar(&g.flagAHUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.ahAppSvcCreateCmd, g.ahAppSvcDeleteCmd, g.ahAppSvcUpdateCmd} {
		c.Flags().BoolVar(&g.flagAHAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.ahAppSvcDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.ahAppSvcListCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")

	g.apphubAppServicesCmd.AddCommand(all...)
	parent.AddCommand(g.apphubAppServicesCmd)
}

func ahAppSvcParent(project, location, app string) string {
	return ahAppName(app, project, location)
}

func ahAppSvcName(id, project, location, app string) string {
	return ahChild("services", id, ahAppSvcParent(project, location, app))
}

func (g *apphubCmds) runAHAppSvcCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &apphub.Service{}
	if err := loadYAMLOrJSONInto(g.flagAHConfigFile, s); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Services.Create(ahAppSvcParent(project, g.flagAHLocation, g.flagAHAppID), s).
		ServiceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Create service", args[0], g.flagAHAsync)
}

func (g *apphubCmds) runAHAppSvcDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Services.Delete(ahAppSvcName(args[0], project, g.flagAHLocation, g.flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Delete service", args[0], g.flagAHAsync)
}

func (g *apphubCmds) runAHAppSvcDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Applications.Services.Get(ahAppSvcName(args[0], project, g.flagAHLocation, g.flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHFormat)
}

func (g *apphubCmds) runAHAppSvcList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Applications.Services.List(ahAppSvcParent(project, g.flagAHLocation, g.flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	if g.flagAHFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Services, g.flagAHFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Services), "%-40s\n", "NAME")
	for _, s := range resp.Services {
		fmt.Fprintln(g.inv.stdout, path.Base(s.Name))
	}
	return nil
}

func (g *apphubCmds) runAHAppSvcUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &apphub.Service{}
	if err := loadYAMLOrJSONInto(g.flagAHConfigFile, s); err != nil {
		return err
	}
	mask := g.flagAHUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Services.Patch(ahAppSvcName(args[0], project, g.flagAHLocation, g.flagAHAppID), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Update service", args[0], g.flagAHAsync)
}

// --- applications workloads ---

func (g *apphubCmds) registerAHAppWorkloads(parent *cobra.Command) {
	all := []*cobra.Command{g.ahAppWLCreateCmd, g.ahAppWLDeleteCmd, g.ahAppWLDescribeCmd, g.ahAppWLListCmd, g.ahAppWLUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagAHLocation, "location", "", "Location containing the application (required)")
		c.Flags().StringVar(&g.flagAHAppID, "application", "", "Application containing the workload (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("application")
	}
	for _, c := range []*cobra.Command{g.ahAppWLCreateCmd, g.ahAppWLUpdateCmd} {
		c.Flags().StringVar(&g.flagAHConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Workload message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.ahAppWLUpdateCmd.Flags().StringVar(&g.flagAHUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.ahAppWLCreateCmd, g.ahAppWLDeleteCmd, g.ahAppWLUpdateCmd} {
		c.Flags().BoolVar(&g.flagAHAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.ahAppWLDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.ahAppWLListCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")

	g.apphubAppWorkloadsCmd.AddCommand(all...)
	parent.AddCommand(g.apphubAppWorkloadsCmd)
}

func ahAppWLName(id, project, location, app string) string {
	return ahChild("workloads", id, ahAppSvcParent(project, location, app))
}

func (g *apphubCmds) runAHAppWLCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	wl := &apphub.Workload{}
	if err := loadYAMLOrJSONInto(g.flagAHConfigFile, wl); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Workloads.Create(ahAppSvcParent(project, g.flagAHLocation, g.flagAHAppID), wl).
		WorkloadId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workload: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Create workload", args[0], g.flagAHAsync)
}

func (g *apphubCmds) runAHAppWLDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Workloads.Delete(ahAppWLName(args[0], project, g.flagAHLocation, g.flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workload: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Delete workload", args[0], g.flagAHAsync)
}

func (g *apphubCmds) runAHAppWLDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Applications.Workloads.Get(ahAppWLName(args[0], project, g.flagAHLocation, g.flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workload: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHFormat)
}

func (g *apphubCmds) runAHAppWLList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Applications.Workloads.List(ahAppSvcParent(project, g.flagAHLocation, g.flagAHAppID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing workloads: %w", err)
	}
	if g.flagAHFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Workloads, g.flagAHFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Workloads), "%-40s\n", "NAME")
	for _, w := range resp.Workloads {
		fmt.Fprintln(g.inv.stdout, path.Base(w.Name))
	}
	return nil
}

func (g *apphubCmds) runAHAppWLUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	wl := &apphub.Workload{}
	if err := loadYAMLOrJSONInto(g.flagAHConfigFile, wl); err != nil {
		return err
	}
	mask := g.flagAHUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(wl))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Applications.Workloads.Patch(ahAppWLName(args[0], project, g.flagAHLocation, g.flagAHAppID), wl).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating workload: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Update workload", args[0], g.flagAHAsync)
}

// --- discovered-services ---

func (g *apphubCmds) initApphubAll5() {
	for _, c := range []*cobra.Command{g.ahDSDescribeCmd, g.ahDSListCmd, g.ahDSLookupCmd} {
		c.Flags().StringVar(&g.flagAHLocation, "location", "", "Location containing the discovered service (required)")
		_ = c.MarkFlagRequired("location")
	}
	g.ahDSLookupCmd.Flags().StringVar(&g.flagAHDSURI, "uri", "", "URI to look up (required)")
	_ = g.ahDSLookupCmd.MarkFlagRequired("uri")
	g.ahDSDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.ahDSListCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")

	g.apphubDiscoveredServicesCmd.AddCommand(g.ahDSDescribeCmd, g.ahDSListCmd, g.ahDSLookupCmd)
	g.apphubCmd.AddCommand(g.apphubDiscoveredServicesCmd)
}

func ahDSName(id, project, location string) string {
	return ahChild("discoveredServices", id, ahLocationParent(project, location))
}

func (g *apphubCmds) runAHDSDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DiscoveredServices.Get(ahDSName(args[0], project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing discovered service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHFormat)
}

func (g *apphubCmds) runAHDSList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DiscoveredServices.List(ahLocationParent(project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing discovered services: %w", err)
	}
	if g.flagAHFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.DiscoveredServices, g.flagAHFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.DiscoveredServices), "%-40s\n", "NAME")
	for _, ds := range resp.DiscoveredServices {
		fmt.Fprintln(g.inv.stdout, path.Base(ds.Name))
	}
	return nil
}

func (g *apphubCmds) runAHDSLookup(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DiscoveredServices.Lookup(ahLocationParent(project, g.flagAHLocation)).Uri(g.flagAHDSURI).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up discovered service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagAHFormat)
}

// --- discovered-workloads ---

func (g *apphubCmds) initApphubAll6() {
	for _, c := range []*cobra.Command{g.ahDWDescribeCmd, g.ahDWListCmd, g.ahDWLookupCmd} {
		c.Flags().StringVar(&g.flagAHLocation, "location", "", "Location containing the discovered workload (required)")
		_ = c.MarkFlagRequired("location")
	}
	g.ahDWLookupCmd.Flags().StringVar(&g.flagAHDWURI, "uri", "", "URI to look up (required)")
	_ = g.ahDWLookupCmd.MarkFlagRequired("uri")
	g.ahDWDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.ahDWListCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")

	g.apphubDiscoveredWorkloadsCmd.AddCommand(g.ahDWDescribeCmd, g.ahDWListCmd, g.ahDWLookupCmd)
	g.apphubCmd.AddCommand(g.apphubDiscoveredWorkloadsCmd)
}

func ahDWName(id, project, location string) string {
	return ahChild("discoveredWorkloads", id, ahLocationParent(project, location))
}

func (g *apphubCmds) runAHDWDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DiscoveredWorkloads.Get(ahDWName(args[0], project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing discovered workload: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHFormat)
}

func (g *apphubCmds) runAHDWList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DiscoveredWorkloads.List(ahLocationParent(project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing discovered workloads: %w", err)
	}
	if g.flagAHFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.DiscoveredWorkloads, g.flagAHFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.DiscoveredWorkloads), "%-40s\n", "NAME")
	for _, dw := range resp.DiscoveredWorkloads {
		fmt.Fprintln(g.inv.stdout, path.Base(dw.Name))
	}
	return nil
}

func (g *apphubCmds) runAHDWLookup(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DiscoveredWorkloads.Lookup(ahLocationParent(project, g.flagAHLocation)).Uri(g.flagAHDWURI).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up discovered workload: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagAHFormat)
}

// --- service-projects ---

func (g *apphubCmds) initApphubAll7() {
	all := []*cobra.Command{g.ahSPAddCmd, g.ahSPRemoveCmd, g.ahSPDescribeCmd, g.ahSPListCmd, g.ahSPLookupCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagAHLocation, "location", "", "Location containing the service-project attachment (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.ahSPAddCmd, g.ahSPRemoveCmd} {
		c.Flags().BoolVar(&g.flagAHAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.ahSPDescribeCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")
	g.ahSPListCmd.Flags().StringVar(&g.flagAHFormat, "format", "", "Output format")

	g.apphubServiceProjectsCmd.AddCommand(all...)
	g.apphubCmd.AddCommand(g.apphubServiceProjectsCmd)
}

func ahSPName(id, project, location string) string {
	return ahChild("serviceProjectAttachments", id, ahLocationParent(project, location))
}

func (g *apphubCmds) runAHSPAdd(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	attach := &apphub.ServiceProjectAttachment{
		ServiceProject: fmt.Sprintf("projects/%s", args[0]),
	}
	op, err := svc.Projects.Locations.ServiceProjectAttachments.Create(ahLocationParent(project, g.flagAHLocation), attach).
		ServiceProjectAttachmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding service project: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Add service project", args[0], g.flagAHAsync)
}

func (g *apphubCmds) runAHSPRemove(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceProjectAttachments.Delete(ahSPName(args[0], project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing service project: %w", err)
	}
	return g.ahFinishOp(ctx, svc, op, "Remove service project", args[0], g.flagAHAsync)
}

func (g *apphubCmds) runAHSPDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceProjectAttachments.Get(ahSPName(args[0], project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service project attachment: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAHFormat)
}

func (g *apphubCmds) runAHSPList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.ServiceProjectAttachments.List(ahLocationParent(project, g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing service project attachments: %w", err)
	}
	if g.flagAHFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.ServiceProjectAttachments, g.flagAHFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.ServiceProjectAttachments), "%-40s %s\n", "NAME", "SERVICE_PROJECT")
	for _, sp := range resp.ServiceProjectAttachments {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(sp.Name), sp.ServiceProject)
	}
	return nil
}

func (g *apphubCmds) runAHSPLookup(cmd *cobra.Command, args []string) error {
	if _, err := g.inv.project(); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppHub(ctx)
	if err != nil {
		return err
	}
	// LookupServiceProjectAttachment takes the service project resource name
	// directly ("projects/PROJECT/locations/…"), not the App Hub host project.
	resp, err := svc.Projects.Locations.LookupServiceProjectAttachment(fmt.Sprintf("projects/%s/locations/%s", args[0], g.flagAHLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up service project attachment: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagAHFormat)
}
