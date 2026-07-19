package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services meshes (#992) ---

var networkServicesMeshesCmd = &cobra.Command{Use: "meshes", Short: "Manage service meshes"}

var (
	flagNsMeshLocation    string
	flagNsMeshFormat      string
	flagNsMeshDestination string
	flagNsMeshSource      string
	flagNsMeshPageSize    int64
)

var (
	networkServicesMeshDeleteCmd = &cobra.Command{
		Use: "delete MESH", Short: "Delete a service mesh",
		Args: cobra.ExactArgs(1), RunE: runNsMeshDelete,
	}
	networkServicesMeshDescribeCmd = &cobra.Command{
		Use: "describe MESH", Short: "Describe a service mesh",
		Args: cobra.ExactArgs(1), RunE: runNsMeshDescribe,
	}
	networkServicesMeshExportCmd = &cobra.Command{
		Use: "export MESH", Short: "Export a service mesh to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsMeshExport,
	}
	networkServicesMeshImportCmd = &cobra.Command{
		Use: "import MESH", Short: "Import a service mesh from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsMeshImport,
	}
	networkServicesMeshListCmd = &cobra.Command{
		Use: "list", Short: "List service meshes in a location",
		Args: cobra.NoArgs, RunE: runNsMeshList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesMeshDeleteCmd, networkServicesMeshDescribeCmd,
		networkServicesMeshExportCmd, networkServicesMeshImportCmd,
		networkServicesMeshListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsMeshLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsMeshFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesMeshExportCmd, &flagNsMeshDestination)
	nsBindImportFlags(networkServicesMeshImportCmd, &flagNsMeshSource)
	networkServicesMeshListCmd.Flags().Int64Var(&flagNsMeshPageSize, "page-size", 0, "Maximum results per page")

	networkServicesMeshesCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesMeshesCmd)
}

func nsMeshName(id string) (string, error) {
	return nsResourceName(flagNsMeshLocation, "meshes", id)
}

func runNsMeshDelete(cmd *cobra.Command, args []string) error {
	name, err := nsMeshName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Meshes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting mesh: %w", err)
	}
	fmt.Printf("Delete mesh [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsMeshFormat)
}

func runNsMeshDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsMeshName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Meshes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing mesh: %w", err)
	}
	return emitFormatted(got, flagNsMeshFormat)
}

func runNsMeshExport(cmd *cobra.Command, args []string) error {
	name, err := nsMeshName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Meshes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting mesh: %w", err)
	}
	return saveAsYAML(flagNsMeshDestination, got)
}

func runNsMeshImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsMeshLocation)
	if err != nil {
		return err
	}
	name, err := nsMeshName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.Mesh{}
	if err := loadYAMLOrJSONInto(flagNsMeshSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Meshes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.Meshes.Create(parent, body).MeshId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating mesh: %w", err)
		}
		fmt.Printf("Create mesh [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsMeshFormat)
	}
	op, err := svc.Projects.Locations.Meshes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating mesh: %w", err)
	}
	fmt.Printf("Update mesh [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsMeshFormat)
}

func runNsMeshList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsMeshLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.Mesh
	pageToken := ""
	for {
		call := svc.Projects.Locations.Meshes.List(parent).Context(ctx)
		if flagNsMeshPageSize > 0 {
			call = call.PageSize(flagNsMeshPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing meshes: %w", err)
		}
		all = append(all, resp.Meshes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsMeshFormat)
}
