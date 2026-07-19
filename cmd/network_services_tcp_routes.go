package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services tcp-routes (#1004) ---

var networkServicesTcpRoutesCmd = &cobra.Command{Use: "tcp-routes", Short: "Manage TCP routes"}

var (
	flagNsTcLocation    string
	flagNsTcFormat      string
	flagNsTcDestination string
	flagNsTcSource      string
	flagNsTcPageSize    int64
)

var (
	networkServicesTcDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete a TCP route",
		Args: cobra.ExactArgs(1), RunE: runNsTcDelete,
	}
	networkServicesTcDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe a TCP route",
		Args: cobra.ExactArgs(1), RunE: runNsTcDescribe,
	}
	networkServicesTcExportCmd = &cobra.Command{
		Use: "export ROUTE", Short: "Export a TCP route to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsTcExport,
	}
	networkServicesTcImportCmd = &cobra.Command{
		Use: "import ROUTE", Short: "Import a TCP route from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsTcImport,
	}
	networkServicesTcListCmd = &cobra.Command{
		Use: "list", Short: "List TCP routes in a location",
		Args: cobra.NoArgs, RunE: runNsTcList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesTcDeleteCmd, networkServicesTcDescribeCmd,
		networkServicesTcExportCmd, networkServicesTcImportCmd,
		networkServicesTcListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsTcLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsTcFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesTcExportCmd, &flagNsTcDestination)
	nsBindImportFlags(networkServicesTcImportCmd, &flagNsTcSource)
	networkServicesTcListCmd.Flags().Int64Var(&flagNsTcPageSize, "page-size", 0, "Maximum results per page")

	networkServicesTcpRoutesCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesTcpRoutesCmd)
}

func nsTcName(id string) (string, error) {
	return nsResourceName(flagNsTcLocation, "tcpRoutes", id)
}

func runNsTcDelete(cmd *cobra.Command, args []string) error {
	name, err := nsTcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TcpRoutes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting TCP route: %w", err)
	}
	fmt.Printf("Delete TCP route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsTcFormat)
}

func runNsTcDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsTcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TcpRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing TCP route: %w", err)
	}
	return emitFormatted(got, flagNsTcFormat)
}

func runNsTcExport(cmd *cobra.Command, args []string) error {
	name, err := nsTcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TcpRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting TCP route: %w", err)
	}
	return saveAsYAML(flagNsTcDestination, got)
}

func runNsTcImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsTcLocation)
	if err != nil {
		return err
	}
	name, err := nsTcName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.TcpRoute{}
	if err := loadYAMLOrJSONInto(flagNsTcSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.TcpRoutes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.TcpRoutes.Create(parent, body).TcpRouteId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating TCP route: %w", err)
		}
		fmt.Printf("Create TCP route [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsTcFormat)
	}
	op, err := svc.Projects.Locations.TcpRoutes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating TCP route: %w", err)
	}
	fmt.Printf("Update TCP route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsTcFormat)
}

func runNsTcList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsTcLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.TcpRoute
	pageToken := ""
	for {
		call := svc.Projects.Locations.TcpRoutes.List(parent).Context(ctx)
		if flagNsTcPageSize > 0 {
			call = call.PageSize(flagNsTcPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing TCP routes: %w", err)
		}
		all = append(all, resp.TcpRoutes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsTcFormat)
}
