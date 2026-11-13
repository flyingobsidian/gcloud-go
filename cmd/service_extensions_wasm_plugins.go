package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions wasm-plugins (#1047) ---

func (s *serviceExtensionsCmds) initServiceExtensionsWasmPlugins() {
	all := []*cobra.Command{
		s.seWasmDeleteCmd, s.seWasmDescribeCmd,
		s.seWasmExportCmd, s.seWasmImportCmd,
		s.seWasmListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSeWasmLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagSeWasmFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.seWasmExportCmd, &s.flagSeWasmDestination)
	nsBindImportFlags(s.seWasmImportCmd, &s.flagSeWasmSource)
	s.seWasmListCmd.Flags().Int64Var(&s.flagSeWasmPageSize, "page-size", 0, "Maximum results per page")

	s.seWasmPluginsCmd.AddCommand(all...)
	s.serviceExtensionsCmd.AddCommand(s.seWasmPluginsCmd)
}

func (s *serviceExtensionsCmds) seWasmName(id string) (string, error) {
	return s.seResource(s.flagSeWasmLocation, "wasmPlugins", id)
}

func (s *serviceExtensionsCmds) runSeWasmDelete(cmd *cobra.Command, args []string) error {
	name, err := s.seWasmName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WasmPlugins.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting wasm plugin: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete wasm plugin [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeWasmFormat)
}

func (s *serviceExtensionsCmds) runSeWasmDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.seWasmName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WasmPlugins.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing wasm plugin: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSeWasmFormat)
}

func (s *serviceExtensionsCmds) runSeWasmExport(cmd *cobra.Command, args []string) error {
	name, err := s.seWasmName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WasmPlugins.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting wasm plugin: %w", err)
	}
	return saveAsYAML(s.flagSeWasmDestination, got)
}

func (s *serviceExtensionsCmds) runSeWasmImport(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeWasmLocation)
	if err != nil {
		return err
	}
	name, err := s.seWasmName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.WasmPlugin{}
	if err := loadYAMLOrJSONInto(s.flagSeWasmSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.WasmPlugins.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.WasmPlugins.Create(parent, body).WasmPluginId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating wasm plugin: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create wasm plugin [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagSeWasmFormat)
	}
	op, err := svc.Projects.Locations.WasmPlugins.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating wasm plugin: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update wasm plugin [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeWasmFormat)
}

func (s *serviceExtensionsCmds) runSeWasmList(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeWasmLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.WasmPlugin
	pageToken := ""
	for {
		call := svc.Projects.Locations.WasmPlugins.List(parent).Context(ctx)
		if s.flagSeWasmPageSize > 0 {
			call = call.PageSize(s.flagSeWasmPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSeWasmFormat)
}
