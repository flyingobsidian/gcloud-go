package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services gateways (#989) ---

var networkServicesGatewaysCmd = &cobra.Command{Use: "gateways", Short: "Manage gateways"}

var (
	flagNsGwLocation    string
	flagNsGwFormat      string
	flagNsGwDestination string
	flagNsGwSource      string
	flagNsGwPageSize    int64
)

var (
	networkServicesGwDeleteCmd = &cobra.Command{
		Use: "delete GATEWAY", Short: "Delete a gateway",
		Args: cobra.ExactArgs(1), RunE: runNsGwDelete,
	}
	networkServicesGwDescribeCmd = &cobra.Command{
		Use: "describe GATEWAY", Short: "Describe a gateway",
		Args: cobra.ExactArgs(1), RunE: runNsGwDescribe,
	}
	networkServicesGwExportCmd = &cobra.Command{
		Use: "export GATEWAY", Short: "Export a gateway to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsGwExport,
	}
	networkServicesGwImportCmd = &cobra.Command{
		Use: "import GATEWAY", Short: "Import a gateway from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsGwImport,
	}
	networkServicesGwListCmd = &cobra.Command{
		Use: "list", Short: "List gateways in a location",
		Args: cobra.NoArgs, RunE: runNsGwList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesGwDeleteCmd, networkServicesGwDescribeCmd,
		networkServicesGwExportCmd, networkServicesGwImportCmd,
		networkServicesGwListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsGwLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsGwFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesGwExportCmd, &flagNsGwDestination)
	nsBindImportFlags(networkServicesGwImportCmd, &flagNsGwSource)
	networkServicesGwListCmd.Flags().Int64Var(&flagNsGwPageSize, "page-size", 0, "Maximum results per page")

	networkServicesGatewaysCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesGatewaysCmd)
}

func nsGwName(id string) (string, error) {
	return nsResourceName(flagNsGwLocation, "gateways", id)
}

func runNsGwDelete(cmd *cobra.Command, args []string) error {
	name, err := nsGwName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Gateways.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting gateway: %w", err)
	}
	fmt.Printf("Delete gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsGwFormat)
}

func runNsGwDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsGwName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Gateways.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing gateway: %w", err)
	}
	return emitFormatted(got, flagNsGwFormat)
}

func runNsGwExport(cmd *cobra.Command, args []string) error {
	name, err := nsGwName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Gateways.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting gateway: %w", err)
	}
	return saveAsYAML(flagNsGwDestination, got)
}

func runNsGwImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsGwLocation)
	if err != nil {
		return err
	}
	name, err := nsGwName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.Gateway{}
	if err := loadYAMLOrJSONInto(flagNsGwSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Gateways.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.Gateways.Create(parent, body).GatewayId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating gateway: %w", err)
		}
		fmt.Printf("Create gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsGwFormat)
	}
	op, err := svc.Projects.Locations.Gateways.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating gateway: %w", err)
	}
	fmt.Printf("Update gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsGwFormat)
}

func runNsGwList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsGwLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.Gateway
	pageToken := ""
	for {
		call := svc.Projects.Locations.Gateways.List(parent).Context(ctx)
		if flagNsGwPageSize > 0 {
			call = call.PageSize(flagNsGwPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing gateways: %w", err)
		}
		all = append(all, resp.Gateways...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsGwFormat)
}
