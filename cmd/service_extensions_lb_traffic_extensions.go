package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions lb-traffic-extensions (#1045) ---

var seLbTrafficExtensionsCmd = &cobra.Command{Use: "lb-traffic-extensions", Short: "Manage LbTrafficExtension resources"}

var (
	flagSeLbTrafficLocation    string
	flagSeLbTrafficFormat      string
	flagSeLbTrafficDestination string
	flagSeLbTrafficSource      string
	flagSeLbTrafficPageSize    int64
)

var (
	seLbTrafficDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an LbTrafficExtension",
		Args: cobra.ExactArgs(1), RunE: runSeLbTrafficDelete,
	}
	seLbTrafficDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an LbTrafficExtension",
		Args: cobra.ExactArgs(1), RunE: runSeLbTrafficDescribe,
	}
	seLbTrafficExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an LbTrafficExtension to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeLbTrafficExport,
	}
	seLbTrafficImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an LbTrafficExtension from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeLbTrafficImport,
	}
	seLbTrafficListCmd = &cobra.Command{
		Use: "list", Short: "List LbTrafficExtensions in a location",
		Args: cobra.NoArgs, RunE: runSeLbTrafficList,
	}
)

func init() {
	all := []*cobra.Command{
		seLbTrafficDeleteCmd, seLbTrafficDescribeCmd,
		seLbTrafficExportCmd, seLbTrafficImportCmd,
		seLbTrafficListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSeLbTrafficLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagSeLbTrafficFormat, "format", "", "Output format")
	}
	nsBindExportFlags(seLbTrafficExportCmd, &flagSeLbTrafficDestination)
	nsBindImportFlags(seLbTrafficImportCmd, &flagSeLbTrafficSource)
	seLbTrafficListCmd.Flags().Int64Var(&flagSeLbTrafficPageSize, "page-size", 0, "Maximum results per page")

	seLbTrafficExtensionsCmd.AddCommand(all...)
	serviceExtensionsCmd.AddCommand(seLbTrafficExtensionsCmd)
}

func seLbTrafficName(id string) (string, error) {
	return seResource(flagSeLbTrafficLocation, "lbTrafficExtensions", id)
}

func runSeLbTrafficDelete(cmd *cobra.Command, args []string) error {
	name, err := seLbTrafficName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.LbTrafficExtensions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting lb traffic extension: %w", err)
	}
	fmt.Printf("Delete lb traffic extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeLbTrafficFormat)
}

func runSeLbTrafficDescribe(cmd *cobra.Command, args []string) error {
	name, err := seLbTrafficName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbTrafficExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing lb traffic extension: %w", err)
	}
	return emitFormatted(got, flagSeLbTrafficFormat)
}

func runSeLbTrafficExport(cmd *cobra.Command, args []string) error {
	name, err := seLbTrafficName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbTrafficExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting lb traffic extension: %w", err)
	}
	return saveAsYAML(flagSeLbTrafficDestination, got)
}

func runSeLbTrafficImport(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeLbTrafficLocation)
	if err != nil {
		return err
	}
	name, err := seLbTrafficName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.LbTrafficExtension{}
	if err := loadYAMLOrJSONInto(flagSeLbTrafficSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.LbTrafficExtensions.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.LbTrafficExtensions.Create(parent, body).LbTrafficExtensionId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating lb traffic extension: %w", err)
		}
		fmt.Printf("Create lb traffic extension [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagSeLbTrafficFormat)
	}
	op, err := svc.Projects.Locations.LbTrafficExtensions.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating lb traffic extension: %w", err)
	}
	fmt.Printf("Update lb traffic extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeLbTrafficFormat)
}

func runSeLbTrafficList(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeLbTrafficLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.LbTrafficExtension
	pageToken := ""
	for {
		call := svc.Projects.Locations.LbTrafficExtensions.List(parent).Context(ctx)
		if flagSeLbTrafficPageSize > 0 {
			call = call.PageSize(flagSeLbTrafficPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing lb traffic extensions: %w", err)
		}
		all = append(all, resp.LbTrafficExtensions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSeLbTrafficFormat)
}
