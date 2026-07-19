package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions wasm-plugin-versions (#1046) ---

var seWasmPluginVersionsCmd = &cobra.Command{Use: "wasm-plugin-versions", Short: "Manage WasmPluginVersion resources"}

var (
	flagSeWasmVerLocation   string
	flagSeWasmVerPlugin     string
	flagSeWasmVerFormat     string
	flagSeWasmVerConfigFile string
	flagSeWasmVerPageSize   int64
)

var (
	seWasmVerCreateCmd = &cobra.Command{
		Use: "create VERSION", Short: "Create a WasmPluginVersion",
		Args: cobra.ExactArgs(1), RunE: runSeWasmVerCreate,
	}
	seWasmVerDeleteCmd = &cobra.Command{
		Use: "delete VERSION", Short: "Delete a WasmPluginVersion",
		Args: cobra.ExactArgs(1), RunE: runSeWasmVerDelete,
	}
	seWasmVerDescribeCmd = &cobra.Command{
		Use: "describe VERSION", Short: "Describe a WasmPluginVersion",
		Args: cobra.ExactArgs(1), RunE: runSeWasmVerDescribe,
	}
	seWasmVerListCmd = &cobra.Command{
		Use: "list", Short: "List WasmPluginVersions under a WasmPlugin",
		Args: cobra.NoArgs, RunE: runSeWasmVerList,
	}
)

func init() {
	all := []*cobra.Command{
		seWasmVerCreateCmd, seWasmVerDeleteCmd,
		seWasmVerDescribeCmd, seWasmVerListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSeWasmVerLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagSeWasmVerPlugin, "plugin", "", "Parent WasmPlugin id (required)")
		_ = c.MarkFlagRequired("plugin")
		c.Flags().StringVar(&flagSeWasmVerFormat, "format", "", "Output format")
	}
	seWasmVerCreateCmd.Flags().StringVar(&flagSeWasmVerConfigFile, "config-file", "", "YAML/JSON file with the WasmPluginVersion body (required)")
	_ = seWasmVerCreateCmd.MarkFlagRequired("config-file")
	seWasmVerListCmd.Flags().Int64Var(&flagSeWasmVerPageSize, "page-size", 0, "Maximum results per page")

	seWasmPluginVersionsCmd.AddCommand(all...)
	serviceExtensionsCmd.AddCommand(seWasmPluginVersionsCmd)
}

func seWasmVerParent() (string, error) {
	parent, err := seLocationParent(flagSeWasmVerLocation)
	if err != nil {
		return "", err
	}
	if flagSeWasmVerPlugin == "" {
		return "", fmt.Errorf("--plugin is required")
	}
	return parent + "/wasmPlugins/" + flagSeWasmVerPlugin, nil
}

func seWasmVerName(id string) (string, error) {
	parent, err := seWasmVerParent()
	if err != nil {
		return "", err
	}
	return parent + "/versions/" + id, nil
}

func runSeWasmVerCreate(cmd *cobra.Command, args []string) error {
	parent, err := seWasmVerParent()
	if err != nil {
		return err
	}
	body := &networkservices.WasmPluginVersion{}
	if err := loadYAMLOrJSONInto(flagSeWasmVerConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WasmPlugins.Versions.Create(parent, body).WasmPluginVersionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating wasm plugin version: %w", err)
	}
	fmt.Printf("Create wasm plugin version [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeWasmVerFormat)
}

func runSeWasmVerDelete(cmd *cobra.Command, args []string) error {
	name, err := seWasmVerName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WasmPlugins.Versions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting wasm plugin version: %w", err)
	}
	fmt.Printf("Delete wasm plugin version [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeWasmVerFormat)
}

func runSeWasmVerDescribe(cmd *cobra.Command, args []string) error {
	name, err := seWasmVerName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WasmPlugins.Versions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing wasm plugin version: %w", err)
	}
	return emitFormatted(got, flagSeWasmVerFormat)
}

func runSeWasmVerList(cmd *cobra.Command, args []string) error {
	parent, err := seWasmVerParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.WasmPluginVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.WasmPlugins.Versions.List(parent).Context(ctx)
		if flagSeWasmVerPageSize > 0 {
			call = call.PageSize(flagSeWasmVerPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing wasm plugin versions: %w", err)
		}
		all = append(all, resp.WasmPluginVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSeWasmVerFormat)
}
