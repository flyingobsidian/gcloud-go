package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions wasm-plugin-versions (#1046) ---

func (s *serviceExtensionsCmds) initServiceExtensionsWasmPluginVersions() {
	all := []*cobra.Command{
		s.seWasmVerCreateCmd, s.seWasmVerDeleteCmd,
		s.seWasmVerDescribeCmd, s.seWasmVerListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSeWasmVerLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagSeWasmVerPlugin, "plugin", "", "Parent WasmPlugin id (required)")
		_ = c.MarkFlagRequired("plugin")
		c.Flags().StringVar(&s.flagSeWasmVerFormat, "format", "", "Output format")
	}
	s.seWasmVerCreateCmd.Flags().StringVar(&s.flagSeWasmVerConfigFile, "config-file", "", "YAML/JSON file with the WasmPluginVersion body (required)")
	_ = s.seWasmVerCreateCmd.MarkFlagRequired("config-file")
	s.seWasmVerListCmd.Flags().Int64Var(&s.flagSeWasmVerPageSize, "page-size", 0, "Maximum results per page")

	s.seWasmPluginVersionsCmd.AddCommand(all...)
	s.serviceExtensionsCmd.AddCommand(s.seWasmPluginVersionsCmd)
}

func (s *serviceExtensionsCmds) seWasmVerParent() (string, error) {
	parent, err := s.seLocationParent(s.flagSeWasmVerLocation)
	if err != nil {
		return "", err
	}
	if s.flagSeWasmVerPlugin == "" {
		return "", fmt.Errorf("--plugin is required")
	}
	return parent + "/wasmPlugins/" + s.flagSeWasmVerPlugin, nil
}

func (s *serviceExtensionsCmds) seWasmVerName(id string) (string, error) {
	parent, err := s.seWasmVerParent()
	if err != nil {
		return "", err
	}
	return parent + "/versions/" + id, nil
}

func (s *serviceExtensionsCmds) runSeWasmVerCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.seWasmVerParent()
	if err != nil {
		return err
	}
	body := &networkservices.WasmPluginVersion{}
	if err := loadYAMLOrJSONInto(s.flagSeWasmVerConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WasmPlugins.Versions.Create(parent, body).WasmPluginVersionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating wasm plugin version: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create wasm plugin version [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeWasmVerFormat)
}

func (s *serviceExtensionsCmds) runSeWasmVerDelete(cmd *cobra.Command, args []string) error {
	name, err := s.seWasmVerName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WasmPlugins.Versions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting wasm plugin version: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete wasm plugin version [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeWasmVerFormat)
}

func (s *serviceExtensionsCmds) runSeWasmVerDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.seWasmVerName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WasmPlugins.Versions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing wasm plugin version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSeWasmVerFormat)
}

func (s *serviceExtensionsCmds) runSeWasmVerList(cmd *cobra.Command, args []string) error {
	parent, err := s.seWasmVerParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.WasmPluginVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.WasmPlugins.Versions.List(parent).Context(ctx)
		if s.flagSeWasmVerPageSize > 0 {
			call = call.PageSize(s.flagSeWasmVerPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSeWasmVerFormat)
}
