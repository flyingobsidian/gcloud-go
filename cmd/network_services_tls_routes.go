package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services tls-routes (#1005) ---

var networkServicesTlsRoutesCmd = &cobra.Command{Use: "tls-routes", Short: "Manage TLS routes"}

var (
	flagNsTlLocation    string
	flagNsTlFormat      string
	flagNsTlDestination string
	flagNsTlSource      string
	flagNsTlPageSize    int64
)

var (
	networkServicesTlDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete a TLS route",
		Args: cobra.ExactArgs(1), RunE: runNsTlDelete,
	}
	networkServicesTlDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe a TLS route",
		Args: cobra.ExactArgs(1), RunE: runNsTlDescribe,
	}
	networkServicesTlExportCmd = &cobra.Command{
		Use: "export ROUTE", Short: "Export a TLS route to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsTlExport,
	}
	networkServicesTlImportCmd = &cobra.Command{
		Use: "import ROUTE", Short: "Import a TLS route from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsTlImport,
	}
	networkServicesTlListCmd = &cobra.Command{
		Use: "list", Short: "List TLS routes in a location",
		Args: cobra.NoArgs, RunE: runNsTlList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesTlDeleteCmd, networkServicesTlDescribeCmd,
		networkServicesTlExportCmd, networkServicesTlImportCmd,
		networkServicesTlListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsTlLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsTlFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesTlExportCmd, &flagNsTlDestination)
	nsBindImportFlags(networkServicesTlImportCmd, &flagNsTlSource)
	networkServicesTlListCmd.Flags().Int64Var(&flagNsTlPageSize, "page-size", 0, "Maximum results per page")

	networkServicesTlsRoutesCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesTlsRoutesCmd)
}

func nsTlName(id string) (string, error) {
	return nsResourceName(flagNsTlLocation, "tlsRoutes", id)
}

func runNsTlDelete(cmd *cobra.Command, args []string) error {
	name, err := nsTlName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TlsRoutes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting TLS route: %w", err)
	}
	fmt.Printf("Delete TLS route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsTlFormat)
}

func runNsTlDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsTlName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TlsRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing TLS route: %w", err)
	}
	return emitFormatted(got, flagNsTlFormat)
}

func runNsTlExport(cmd *cobra.Command, args []string) error {
	name, err := nsTlName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TlsRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting TLS route: %w", err)
	}
	return saveAsYAML(flagNsTlDestination, got)
}

func runNsTlImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsTlLocation)
	if err != nil {
		return err
	}
	name, err := nsTlName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.TlsRoute{}
	if err := loadYAMLOrJSONInto(flagNsTlSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.TlsRoutes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.TlsRoutes.Create(parent, body).TlsRouteId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating TLS route: %w", err)
		}
		fmt.Printf("Create TLS route [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsTlFormat)
	}
	op, err := svc.Projects.Locations.TlsRoutes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating TLS route: %w", err)
	}
	fmt.Printf("Update TLS route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsTlFormat)
}

func runNsTlList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsTlLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.TlsRoute
	pageToken := ""
	for {
		call := svc.Projects.Locations.TlsRoutes.List(parent).Context(ctx)
		if flagNsTlPageSize > 0 {
			call = call.PageSize(flagNsTlPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing TLS routes: %w", err)
		}
		all = append(all, resp.TlsRoutes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsTlFormat)
}
