package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions lb-route-extensions (#1044) ---

var seLbRouteExtensionsCmd = &cobra.Command{Use: "lb-route-extensions", Short: "Manage LbRouteExtension resources"}

var (
	flagSeLbRouteLocation    string
	flagSeLbRouteFormat      string
	flagSeLbRouteDestination string
	flagSeLbRouteSource      string
	flagSeLbRoutePageSize    int64
)

var (
	seLbRouteDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an LbRouteExtension",
		Args: cobra.ExactArgs(1), RunE: runSeLbRouteDelete,
	}
	seLbRouteDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an LbRouteExtension",
		Args: cobra.ExactArgs(1), RunE: runSeLbRouteDescribe,
	}
	seLbRouteExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an LbRouteExtension to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeLbRouteExport,
	}
	seLbRouteImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an LbRouteExtension from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeLbRouteImport,
	}
	seLbRouteListCmd = &cobra.Command{
		Use: "list", Short: "List LbRouteExtensions in a location",
		Args: cobra.NoArgs, RunE: runSeLbRouteList,
	}
)

func init() {
	all := []*cobra.Command{
		seLbRouteDeleteCmd, seLbRouteDescribeCmd,
		seLbRouteExportCmd, seLbRouteImportCmd,
		seLbRouteListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSeLbRouteLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagSeLbRouteFormat, "format", "", "Output format")
	}
	nsBindExportFlags(seLbRouteExportCmd, &flagSeLbRouteDestination)
	nsBindImportFlags(seLbRouteImportCmd, &flagSeLbRouteSource)
	seLbRouteListCmd.Flags().Int64Var(&flagSeLbRoutePageSize, "page-size", 0, "Maximum results per page")

	seLbRouteExtensionsCmd.AddCommand(all...)
	serviceExtensionsCmd.AddCommand(seLbRouteExtensionsCmd)
}

func seLbRouteName(id string) (string, error) {
	return seResource(flagSeLbRouteLocation, "lbRouteExtensions", id)
}

func runSeLbRouteDelete(cmd *cobra.Command, args []string) error {
	name, err := seLbRouteName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.LbRouteExtensions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting lb route extension: %w", err)
	}
	fmt.Printf("Delete lb route extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeLbRouteFormat)
}

func runSeLbRouteDescribe(cmd *cobra.Command, args []string) error {
	name, err := seLbRouteName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbRouteExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing lb route extension: %w", err)
	}
	return emitFormatted(got, flagSeLbRouteFormat)
}

func runSeLbRouteExport(cmd *cobra.Command, args []string) error {
	name, err := seLbRouteName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbRouteExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting lb route extension: %w", err)
	}
	return saveAsYAML(flagSeLbRouteDestination, got)
}

func runSeLbRouteImport(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeLbRouteLocation)
	if err != nil {
		return err
	}
	name, err := seLbRouteName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.LbRouteExtension{}
	if err := loadYAMLOrJSONInto(flagSeLbRouteSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.LbRouteExtensions.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.LbRouteExtensions.Create(parent, body).LbRouteExtensionId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating lb route extension: %w", err)
		}
		fmt.Printf("Create lb route extension [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagSeLbRouteFormat)
	}
	op, err := svc.Projects.Locations.LbRouteExtensions.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating lb route extension: %w", err)
	}
	fmt.Printf("Update lb route extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeLbRouteFormat)
}

func runSeLbRouteList(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeLbRouteLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.LbRouteExtension
	pageToken := ""
	for {
		call := svc.Projects.Locations.LbRouteExtensions.List(parent).Context(ctx)
		if flagSeLbRoutePageSize > 0 {
			call = call.PageSize(flagSeLbRoutePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing lb route extensions: %w", err)
		}
		all = append(all, resp.LbRouteExtensions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSeLbRouteFormat)
}
