package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services grpc-routes (#990) ---

var networkServicesGrpcRoutesCmd = &cobra.Command{Use: "grpc-routes", Short: "Manage gRPC routes"}

var (
	flagNsGrLocation    string
	flagNsGrFormat      string
	flagNsGrDestination string
	flagNsGrSource      string
	flagNsGrPageSize    int64
)

var (
	networkServicesGrDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete a gRPC route",
		Args: cobra.ExactArgs(1), RunE: runNsGrDelete,
	}
	networkServicesGrDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe a gRPC route",
		Args: cobra.ExactArgs(1), RunE: runNsGrDescribe,
	}
	networkServicesGrExportCmd = &cobra.Command{
		Use: "export ROUTE", Short: "Export a gRPC route to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsGrExport,
	}
	networkServicesGrImportCmd = &cobra.Command{
		Use: "import ROUTE", Short: "Import a gRPC route from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsGrImport,
	}
	networkServicesGrListCmd = &cobra.Command{
		Use: "list", Short: "List gRPC routes in a location",
		Args: cobra.NoArgs, RunE: runNsGrList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesGrDeleteCmd, networkServicesGrDescribeCmd,
		networkServicesGrExportCmd, networkServicesGrImportCmd,
		networkServicesGrListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsGrLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsGrFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesGrExportCmd, &flagNsGrDestination)
	nsBindImportFlags(networkServicesGrImportCmd, &flagNsGrSource)
	networkServicesGrListCmd.Flags().Int64Var(&flagNsGrPageSize, "page-size", 0, "Maximum results per page")

	networkServicesGrpcRoutesCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesGrpcRoutesCmd)
}

func nsGrName(id string) (string, error) {
	return nsResourceName(flagNsGrLocation, "grpcRoutes", id)
}

func runNsGrDelete(cmd *cobra.Command, args []string) error {
	name, err := nsGrName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GrpcRoutes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting gRPC route: %w", err)
	}
	fmt.Printf("Delete gRPC route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsGrFormat)
}

func runNsGrDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsGrName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GrpcRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing gRPC route: %w", err)
	}
	return emitFormatted(got, flagNsGrFormat)
}

func runNsGrExport(cmd *cobra.Command, args []string) error {
	name, err := nsGrName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GrpcRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting gRPC route: %w", err)
	}
	return saveAsYAML(flagNsGrDestination, got)
}

func runNsGrImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsGrLocation)
	if err != nil {
		return err
	}
	name, err := nsGrName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.GrpcRoute{}
	if err := loadYAMLOrJSONInto(flagNsGrSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.GrpcRoutes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.GrpcRoutes.Create(parent, body).GrpcRouteId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating gRPC route: %w", err)
		}
		fmt.Printf("Create gRPC route [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsGrFormat)
	}
	op, err := svc.Projects.Locations.GrpcRoutes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating gRPC route: %w", err)
	}
	fmt.Printf("Update gRPC route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsGrFormat)
}

func runNsGrList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsGrLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.GrpcRoute
	pageToken := ""
	for {
		call := svc.Projects.Locations.GrpcRoutes.List(parent).Context(ctx)
		if flagNsGrPageSize > 0 {
			call = call.PageSize(flagNsGrPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing gRPC routes: %w", err)
		}
		all = append(all, resp.GrpcRoutes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsGrFormat)
}
