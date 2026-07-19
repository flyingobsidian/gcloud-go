package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions wasm-plugins (#1047) ---

var seWasmPluginsCmd = &cobra.Command{Use: "wasm-plugins", Short: "Manage WasmPlugin resources"}

var (
	flagSeWasmLocation    string
	flagSeWasmFormat      string
	flagSeWasmDestination string
	flagSeWasmSource      string
	flagSeWasmPageSize    int64
)

var (
	seWasmDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete a WasmPlugin",
		Args: cobra.ExactArgs(1), RunE: runSeWasmDelete,
	}
	seWasmDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a WasmPlugin",
		Args: cobra.ExactArgs(1), RunE: runSeWasmDescribe,
	}
	seWasmExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export a WasmPlugin to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeWasmExport,
	}
	seWasmImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import a WasmPlugin from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeWasmImport,
	}
	seWasmListCmd = &cobra.Command{
		Use: "list", Short: "List WasmPlugins in a location",
		Args: cobra.NoArgs, RunE: runSeWasmList,
	}
)

func init() {
	all := []*cobra.Command{
		seWasmDeleteCmd, seWasmDescribeCmd,
		seWasmExportCmd, seWasmImportCmd,
		seWasmListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSeWasmLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagSeWasmFormat, "format", "", "Output format")
	}
	nsBindExportFlags(seWasmExportCmd, &flagSeWasmDestination)
	nsBindImportFlags(seWasmImportCmd, &flagSeWasmSource)
	seWasmListCmd.Flags().Int64Var(&flagSeWasmPageSize, "page-size", 0, "Maximum results per page")

	seWasmPluginsCmd.AddCommand(all...)
	serviceExtensionsCmd.AddCommand(seWasmPluginsCmd)
}

func seWasmName(id string) (string, error) {
	return seResource(flagSeWasmLocation, "wasmPlugins", id)
}

func runSeWasmDelete(cmd *cobra.Command, args []string) error {
	name, err := seWasmName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WasmPlugins.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting wasm plugin: %w", err)
	}
	fmt.Printf("Delete wasm plugin [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeWasmFormat)
}

func runSeWasmDescribe(cmd *cobra.Command, args []string) error {
	name, err := seWasmName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WasmPlugins.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing wasm plugin: %w", err)
	}
	return emitFormatted(got, flagSeWasmFormat)
}

func runSeWasmExport(cmd *cobra.Command, args []string) error {
	name, err := seWasmName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WasmPlugins.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting wasm plugin: %w", err)
	}
	return saveAsYAML(flagSeWasmDestination, got)
}

func runSeWasmImport(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeWasmLocation)
	if err != nil {
		return err
	}
	name, err := seWasmName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.WasmPlugin{}
	if err := loadYAMLOrJSONInto(flagSeWasmSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.WasmPlugins.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.WasmPlugins.Create(parent, body).WasmPluginId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating wasm plugin: %w", err)
		}
		fmt.Printf("Create wasm plugin [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagSeWasmFormat)
	}
	op, err := svc.Projects.Locations.WasmPlugins.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating wasm plugin: %w", err)
	}
	fmt.Printf("Update wasm plugin [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeWasmFormat)
}

func runSeWasmList(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeWasmLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.WasmPlugin
	pageToken := ""
	for {
		call := svc.Projects.Locations.WasmPlugins.List(parent).Context(ctx)
		if flagSeWasmPageSize > 0 {
			call = call.PageSize(flagSeWasmPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing wasm plugins: %w", err)
		}
		all = append(all, resp.WasmPlugins...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSeWasmFormat)
}
