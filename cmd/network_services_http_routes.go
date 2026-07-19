package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services http-routes (#991) ---

var networkServicesHttpRoutesCmd = &cobra.Command{Use: "http-routes", Short: "Manage HTTP routes"}

var (
	flagNsHrLocation    string
	flagNsHrFormat      string
	flagNsHrDestination string
	flagNsHrSource      string
	flagNsHrPageSize    int64
)

var (
	networkServicesHrDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete an HTTP route",
		Args: cobra.ExactArgs(1), RunE: runNsHrDelete,
	}
	networkServicesHrDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe an HTTP route",
		Args: cobra.ExactArgs(1), RunE: runNsHrDescribe,
	}
	networkServicesHrExportCmd = &cobra.Command{
		Use: "export ROUTE", Short: "Export an HTTP route to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsHrExport,
	}
	networkServicesHrImportCmd = &cobra.Command{
		Use: "import ROUTE", Short: "Import an HTTP route from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsHrImport,
	}
	networkServicesHrListCmd = &cobra.Command{
		Use: "list", Short: "List HTTP routes in a location",
		Args: cobra.NoArgs, RunE: runNsHrList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesHrDeleteCmd, networkServicesHrDescribeCmd,
		networkServicesHrExportCmd, networkServicesHrImportCmd,
		networkServicesHrListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsHrLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsHrFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesHrExportCmd, &flagNsHrDestination)
	nsBindImportFlags(networkServicesHrImportCmd, &flagNsHrSource)
	networkServicesHrListCmd.Flags().Int64Var(&flagNsHrPageSize, "page-size", 0, "Maximum results per page")

	networkServicesHttpRoutesCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesHttpRoutesCmd)
}

func nsHrName(id string) (string, error) {
	return nsResourceName(flagNsHrLocation, "httpRoutes", id)
}

func runNsHrDelete(cmd *cobra.Command, args []string) error {
	name, err := nsHrName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.HttpRoutes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting HTTP route: %w", err)
	}
	fmt.Printf("Delete HTTP route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsHrFormat)
}

func runNsHrDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsHrName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HttpRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing HTTP route: %w", err)
	}
	return emitFormatted(got, flagNsHrFormat)
}

func runNsHrExport(cmd *cobra.Command, args []string) error {
	name, err := nsHrName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HttpRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting HTTP route: %w", err)
	}
	return saveAsYAML(flagNsHrDestination, got)
}

func runNsHrImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsHrLocation)
	if err != nil {
		return err
	}
	name, err := nsHrName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.HttpRoute{}
	if err := loadYAMLOrJSONInto(flagNsHrSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.HttpRoutes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.HttpRoutes.Create(parent, body).HttpRouteId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating HTTP route: %w", err)
		}
		fmt.Printf("Create HTTP route [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsHrFormat)
	}
	op, err := svc.Projects.Locations.HttpRoutes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating HTTP route: %w", err)
	}
	fmt.Printf("Update HTTP route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsHrFormat)
}

func runNsHrList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsHrLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.HttpRoute
	pageToken := ""
	for {
		call := svc.Projects.Locations.HttpRoutes.List(parent).Context(ctx)
		if flagNsHrPageSize > 0 {
			call = call.PageSize(flagNsHrPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing HTTP routes: %w", err)
		}
		all = append(all, resp.HttpRoutes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsHrFormat)
}
