package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub plugins (#1164) ---

func (g *apihubCmds) initApihubPlugins() {
	pluginCmds := []*cobra.Command{
		g.apihubPluginsCreateCmd, g.apihubPluginsDeleteCmd, g.apihubPluginsDescribeCmd,
		g.apihubPluginsDisableCmd, g.apihubPluginsEnableCmd, g.apihubPluginsListCmd,
	}
	for _, c := range pluginCmds {
		c.Flags().StringVar(&g.flagAPPluginsLocation, "location", "",
			"Location that owns the plugin (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagAPPluginsFormat, "format", "", "Output format")
	}
	g.apihubPluginsCreateCmd.Flags().StringVar(&g.flagAPPluginsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Plugin body (required)")
	_ = g.apihubPluginsCreateCmd.MarkFlagRequired("config-file")
	g.apihubPluginsListCmd.Flags().StringVar(&g.flagAPPluginsFilter, "filter", "", "Server-side filter expression")
	g.apihubPluginsListCmd.Flags().Int64Var(&g.flagAPPluginsPageSize, "page-size", 0, "Maximum number of results per page")

	instCmds := []*cobra.Command{
		g.apihubPluginsInstCreateCmd, g.apihubPluginsInstDeleteCmd, g.apihubPluginsInstDescribeCmd,
		g.apihubPluginsInstListCmd, g.apihubPluginsInstUpdateCmd,
	}
	for _, c := range instCmds {
		c.Flags().StringVar(&g.flagAPPluginsLocation, "location", "",
			"Location that owns the plugin (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagAPPluginsFormat, "format", "", "Output format")
		c.Flags().StringVar(&g.flagAPPluginsPlugin, "plugin", "",
			"Parent plugin ID (required)")
		_ = c.MarkFlagRequired("plugin")
	}
	for _, c := range []*cobra.Command{g.apihubPluginsInstCreateCmd, g.apihubPluginsInstUpdateCmd} {
		c.Flags().StringVar(&g.flagAPPluginsInstanceCfg, "config-file", "",
			"Path to a YAML/JSON file with the PluginInstance body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.apihubPluginsInstUpdateCmd.Flags().StringVar(&g.flagAPPluginsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (default: top-level fields in --config-file)")
	g.apihubPluginsInstListCmd.Flags().StringVar(&g.flagAPPluginsFilter, "filter", "", "Server-side filter expression")
	g.apihubPluginsInstListCmd.Flags().Int64Var(&g.flagAPPluginsPageSize, "page-size", 0, "Maximum number of results per page")

	g.apihubPluginsCmd.AddCommand(pluginCmds...)
	g.apihubPluginsInstancesCmd.AddCommand(instCmds...)
	g.apihubPluginsCmd.AddCommand(g.apihubPluginsInstancesCmd)
	g.apihubCmd.AddCommand(g.apihubPluginsCmd)
}

func (g *apihubCmds) apihubPluginsParent() (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, g.flagAPPluginsLocation), nil
}

func (g *apihubCmds) apihubPluginName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := g.apihubPluginsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/plugins/%s", parent, id), nil
}

func (g *apihubCmds) apihubPluginParentForInstance() (string, error) {
	if strings.HasPrefix(g.flagAPPluginsPlugin, "projects/") {
		return g.flagAPPluginsPlugin, nil
	}
	return g.apihubPluginName(g.flagAPPluginsPlugin)
}

func (g *apihubCmds) apihubPluginInstanceName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	pluginName, err := g.apihubPluginParentForInstance()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", pluginName, id), nil
}

func (g *apihubCmds) runAPPluginCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubPluginsParent()
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Plugin{}
	if err := loadYAMLOrJSONInto(g.flagAPPluginsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Create(parent, body).
		PluginId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating plugin: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubPluginName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Plugins.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting plugin: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Delete request issued for plugin [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(g.inv.stdout, op, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubPluginName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing plugin: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginDisable(cmd *cobra.Command, args []string) error {
	name, err := g.apihubPluginName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Disable(name, &apihub.GoogleCloudApihubV1DisablePluginRequest{}).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling plugin: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginEnable(cmd *cobra.Command, args []string) error {
	name, err := g.apihubPluginName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Enable(name, &apihub.GoogleCloudApihubV1EnablePluginRequest{}).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling plugin: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubPluginsParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Plugin
	pageToken := ""
	for {
		call := svc.Projects.Locations.Plugins.List(parent).Context(ctx)
		if g.flagAPPluginsFilter != "" {
			call = call.Filter(g.flagAPPluginsFilter)
		}
		if g.flagAPPluginsPageSize > 0 {
			call = call.PageSize(g.flagAPPluginsPageSize)
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
	return emitFormattedTo(g.inv.stdout, all, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubPluginParentForInstance()
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1PluginInstance{}
	if err := loadYAMLOrJSONInto(g.flagAPPluginsInstanceCfg, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Plugins.Instances.Create(parent, body).
		PluginInstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating plugin instance: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Create request issued for plugin instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(g.inv.stdout, op, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginInstDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubPluginInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Plugins.Instances.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting plugin instance: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Delete request issued for plugin instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(g.inv.stdout, op, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubPluginInstanceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing plugin instance: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginInstList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubPluginParentForInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1PluginInstance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Plugins.Instances.List(parent).Context(ctx)
		if g.flagAPPluginsFilter != "" {
			call = call.Filter(g.flagAPPluginsFilter)
		}
		if g.flagAPPluginsPageSize > 0 {
			call = call.PageSize(g.flagAPPluginsPageSize)
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
	return emitFormattedTo(g.inv.stdout, all, g.flagAPPluginsFormat)
}

func (g *apihubCmds) runAPPluginInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := g.apihubPluginInstanceName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1PluginInstance{}
	if err := loadYAMLOrJSONInto(g.flagAPPluginsInstanceCfg, body); err != nil {
		return err
	}
	mask := g.flagAPPluginsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Plugins.Instances.Patch(name, body).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating plugin instance: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAPPluginsFormat)
}
