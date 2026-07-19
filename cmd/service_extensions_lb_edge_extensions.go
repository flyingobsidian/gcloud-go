package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions lb-edge-extensions (#1043) ---

var seLbEdgeExtensionsCmd = &cobra.Command{Use: "lb-edge-extensions", Short: "Manage LbEdgeExtension resources"}

var (
	flagSeLbEdgeLocation    string
	flagSeLbEdgeFormat      string
	flagSeLbEdgeDestination string
	flagSeLbEdgeSource      string
	flagSeLbEdgePageSize    int64
)

var (
	seLbEdgeDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an LbEdgeExtension",
		Args: cobra.ExactArgs(1), RunE: runSeLbEdgeDelete,
	}
	seLbEdgeDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an LbEdgeExtension",
		Args: cobra.ExactArgs(1), RunE: runSeLbEdgeDescribe,
	}
	seLbEdgeExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an LbEdgeExtension to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeLbEdgeExport,
	}
	seLbEdgeImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an LbEdgeExtension from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeLbEdgeImport,
	}
	seLbEdgeListCmd = &cobra.Command{
		Use: "list", Short: "List LbEdgeExtensions in a location",
		Args: cobra.NoArgs, RunE: runSeLbEdgeList,
	}
)

func init() {
	all := []*cobra.Command{
		seLbEdgeDeleteCmd, seLbEdgeDescribeCmd,
		seLbEdgeExportCmd, seLbEdgeImportCmd,
		seLbEdgeListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSeLbEdgeLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagSeLbEdgeFormat, "format", "", "Output format")
	}
	nsBindExportFlags(seLbEdgeExportCmd, &flagSeLbEdgeDestination)
	nsBindImportFlags(seLbEdgeImportCmd, &flagSeLbEdgeSource)
	seLbEdgeListCmd.Flags().Int64Var(&flagSeLbEdgePageSize, "page-size", 0, "Maximum results per page")

	seLbEdgeExtensionsCmd.AddCommand(all...)
	serviceExtensionsCmd.AddCommand(seLbEdgeExtensionsCmd)
}

func seLbEdgeName(id string) (string, error) {
	return seResource(flagSeLbEdgeLocation, "lbEdgeExtensions", id)
}

func runSeLbEdgeDelete(cmd *cobra.Command, args []string) error {
	name, err := seLbEdgeName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.LbEdgeExtensions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting lb edge extension: %w", err)
	}
	fmt.Printf("Delete lb edge extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeLbEdgeFormat)
}

func runSeLbEdgeDescribe(cmd *cobra.Command, args []string) error {
	name, err := seLbEdgeName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbEdgeExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing lb edge extension: %w", err)
	}
	return emitFormatted(got, flagSeLbEdgeFormat)
}

func runSeLbEdgeExport(cmd *cobra.Command, args []string) error {
	name, err := seLbEdgeName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbEdgeExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting lb edge extension: %w", err)
	}
	return saveAsYAML(flagSeLbEdgeDestination, got)
}

func runSeLbEdgeImport(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeLbEdgeLocation)
	if err != nil {
		return err
	}
	name, err := seLbEdgeName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.LbEdgeExtension{}
	if err := loadYAMLOrJSONInto(flagSeLbEdgeSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.LbEdgeExtensions.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.LbEdgeExtensions.Create(parent, body).LbEdgeExtensionId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating lb edge extension: %w", err)
		}
		fmt.Printf("Create lb edge extension [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagSeLbEdgeFormat)
	}
	op, err := svc.Projects.Locations.LbEdgeExtensions.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating lb edge extension: %w", err)
	}
	fmt.Printf("Update lb edge extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeLbEdgeFormat)
}

func runSeLbEdgeList(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeLbEdgeLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.LbEdgeExtension
	pageToken := ""
	for {
		call := svc.Projects.Locations.LbEdgeExtensions.List(parent).Context(ctx)
		if flagSeLbEdgePageSize > 0 {
			call = call.PageSize(flagSeLbEdgePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing lb edge extensions: %w", err)
		}
		all = append(all, resp.LbEdgeExtensions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSeLbEdgeFormat)
}
