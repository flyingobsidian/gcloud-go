package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub plugins (#1164) ---

var apihubPluginsCmd = &cobra.Command{Use: "plugins", Short: "Manage API Hub plugins"}

var apihubPluginsInstancesCmd = &cobra.Command{
	Use: "instances", Short: "Manage API Hub plugin instances",
}

var (
	flagAPPluginsLocation    string
	flagAPPluginsFormat      string
	flagAPPluginsFilter      string
	flagAPPluginsPageSize    int64
	flagAPPluginsConfigFile  string
	flagAPPluginsPlugin      string
	flagAPPluginsUpdateMask  string
	flagAPPluginsInstanceCfg string
)

var (
	apihubPluginsCreateCmd = &cobra.Command{
		Use: "create PLUGIN", Short: "Create an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: runAPPluginCreate,
	}
	apihubPluginsDeleteCmd = &cobra.Command{
		Use: "delete PLUGIN", Short: "Delete an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: runAPPluginDelete,
	}
	apihubPluginsDescribeCmd = &cobra.Command{
		Use: "describe PLUGIN", Short: "Describe an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: runAPPluginDescribe,
	}
	apihubPluginsDisableCmd = &cobra.Command{
		Use: "disable PLUGIN", Short: "Disable an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: runAPPluginDisable,
	}
	apihubPluginsEnableCmd = &cobra.Command{
		Use: "enable PLUGIN", Short: "Enable an API Hub plugin",
		Args: cobra.ExactArgs(1), RunE: runAPPluginEnable,
	}
	apihubPluginsListCmd = &cobra.Command{
		Use: "list", Short: "List API Hub plugins in a location",
		Args: cobra.NoArgs, RunE: runAPPluginList,
	}

	apihubPluginsInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create an API Hub plugin instance",
		Args: cobra.ExactArgs(1), RunE: runAPPluginInstCreate,
	}
	apihubPluginsInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an API Hub plugin instance",
		Args: cobra.ExactArgs(1), RunE: runAPPluginInstDelete,
	}
	apihubPluginsInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an API Hub plugin instance",
		Args: cobra.ExactArgs(1), RunE: runAPPluginInstDescribe,
	}
	apihubPluginsInstListCmd = &cobra.Command{
		Use: "list", Short: "List API Hub plugin instances under a plugin",
		Args: cobra.NoArgs, RunE: runAPPluginInstList,
	}
	apihubPluginsInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update an API Hub plugin instance",
		Args: cobra.ExactArgs(1), RunE: runAPPluginInstUpdate,
	}
)

func init() {
	pluginCmds := []*cobra.Command{
		apihubPluginsCreateCmd, apihubPluginsDeleteCmd, apihubPluginsDescribeCmd,
		apihubPluginsDisableCmd, apihubPluginsEnableCmd, apihubPluginsListCmd,
	}
	for _, c := range pluginCmds {
		c.Flags().StringVar(&flagAPPluginsLocation, "location", "",
			"Location that owns the plugin (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagAPPluginsFormat, "format", "", "Output format")
	}
	apihubPluginsCreateCmd.Flags().StringVar(&flagAPPluginsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Plugin body (required)")
	_ = apihubPluginsCreateCmd.MarkFlagRequired("config-file")
	apihubPluginsListCmd.Flags().StringVar(&flagAPPluginsFilter, "filter", "", "Server-side filter expression")
	apihubPluginsListCmd.Flags().Int64Var(&flagAPPluginsPageSize, "page-size", 0, "Maximum number of results per page")

	instCmds := []*cobra.Command{
		apihubPluginsInstCreateCmd, apihubPluginsInstDeleteCmd, apihubPluginsInstDescribeCmd,
		apihubPluginsInstListCmd, apihubPluginsInstUpdateCmd,
	}
	for _, c := range instCmds {
		c.Flags().StringVar(&flagAPPluginsLocation, "location", "",
			"Location that owns the plugin (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagAPPluginsFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagAPPluginsPlugin, "plugin", "",
			"Parent plugin ID (required)")
		_ = c.MarkFlagRequired("plugin")
	}
	for _, c := range []*cobra.Command{apihubPluginsInstCreateCmd, apihubPluginsInstUpdateCmd} {
		c.Flags().StringVar(&flagAPPluginsInstanceCfg, "config-file", "",
			"Path to a YAML/JSON file with the PluginInstance body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	apihubPluginsInstUpdateCmd.Flags().StringVar(&flagAPPluginsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (default: top-level fields in --config-file)")
	apihubPluginsInstListCmd.Flags().StringVar(&flagAPPluginsFilter, "filter", "", "Server-side filter expression")
	apihubPluginsInstListCmd.Flags().Int64Var(&flagAPPluginsPageSize, "page-size", 0, "Maximum number of results per page")

	apihubPluginsCmd.AddCommand(pluginCmds...)
	apihubPluginsInstancesCmd.AddCommand(instCmds...)
	apihubPluginsCmd.AddCommand(apihubPluginsInstancesCmd)
	apihubCmd.AddCommand(apihubPluginsCmd)
}

func apihubPluginsParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagAPPluginsLocation), nil
}

func apihubPluginName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := apihubPluginsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/plugins/%s", parent, id), nil
}

func apihubPluginParentForInstance() (string, error) {
	if strings.HasPrefix(flagAPPluginsPlugin, "projects/") {
		return flagAPPluginsPlugin, nil
	}
	return apihubPluginName(flagAPPluginsPlugin)
}

func apihubPluginInstanceName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	pluginName, err := apihubPluginParentForInstance()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", pluginName, id), nil
}

func runAPPluginCreate(cmd *cobra.Command, args []string) error {
	parent, err := apihubPluginsParent()
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Plugin{}
	if err := loadYAMLOrJSONInto(flagAPPluginsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Create(parent, body).
		PluginId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating plugin: %w", err)
	}
	return emitFormatted(got, flagAPPluginsFormat)
}

func runAPPluginDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubPluginName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Plugins.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting plugin: %w", err)
	}
	fmt.Printf("Delete request issued for plugin [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAPPluginsFormat)
}

func runAPPluginDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubPluginName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing plugin: %w", err)
	}
	return emitFormatted(got, flagAPPluginsFormat)
}

func runAPPluginDisable(cmd *cobra.Command, args []string) error {
	name, err := apihubPluginName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Disable(name, &apihub.GoogleCloudApihubV1DisablePluginRequest{}).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling plugin: %w", err)
	}
	return emitFormatted(got, flagAPPluginsFormat)
}

func runAPPluginEnable(cmd *cobra.Command, args []string) error {
	name, err := apihubPluginName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Enable(name, &apihub.GoogleCloudApihubV1EnablePluginRequest{}).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling plugin: %w", err)
	}
	return emitFormatted(got, flagAPPluginsFormat)
}

func runAPPluginList(cmd *cobra.Command, args []string) error {
	parent, err := apihubPluginsParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Plugin
	pageToken := ""
	for {
		call := svc.Projects.Locations.Plugins.List(parent).Context(ctx)
		if flagAPPluginsFilter != "" {
			call = call.Filter(flagAPPluginsFilter)
		}
		if flagAPPluginsPageSize > 0 {
			call = call.PageSize(flagAPPluginsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing plugins: %w", err)
		}
		all = append(all, resp.Plugins...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAPPluginsFormat)
}

func runAPPluginInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := apihubPluginParentForInstance()
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1PluginInstance{}
	if err := loadYAMLOrJSONInto(flagAPPluginsInstanceCfg, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Plugins.Instances.Create(parent, body).
		PluginInstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating plugin instance: %w", err)
	}
	fmt.Printf("Create request issued for plugin instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAPPluginsFormat)
}

func runAPPluginInstDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubPluginInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Plugins.Instances.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting plugin instance: %w", err)
	}
	fmt.Printf("Delete request issued for plugin instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAPPluginsFormat)
}

func runAPPluginInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubPluginInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing plugin instance: %w", err)
	}
	return emitFormatted(got, flagAPPluginsFormat)
}

func runAPPluginInstList(cmd *cobra.Command, args []string) error {
	parent, err := apihubPluginParentForInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1PluginInstance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Plugins.Instances.List(parent).Context(ctx)
		if flagAPPluginsFilter != "" {
			call = call.Filter(flagAPPluginsFilter)
		}
		if flagAPPluginsPageSize > 0 {
			call = call.PageSize(flagAPPluginsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing plugin instances: %w", err)
		}
		all = append(all, resp.PluginInstances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAPPluginsFormat)
}

func runAPPluginInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := apihubPluginInstanceName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1PluginInstance{}
	if err := loadYAMLOrJSONInto(flagAPPluginsInstanceCfg, body); err != nil {
		return err
	}
	mask := flagAPPluginsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Instances.Patch(name, body).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating plugin instance: %w", err)
	}
	return emitFormatted(got, flagAPPluginsFormat)
}
