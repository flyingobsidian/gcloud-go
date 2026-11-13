package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services meshes (#992) ---

func (s *networkServicesCmds) initNetworkServicesMeshes() {
	all := []*cobra.Command{
		s.networkServicesMeshDeleteCmd, s.networkServicesMeshDescribeCmd,
		s.networkServicesMeshExportCmd, s.networkServicesMeshImportCmd,
		s.networkServicesMeshListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsMeshLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsMeshFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesMeshExportCmd, &s.flagNsMeshDestination)
	nsBindImportFlags(s.networkServicesMeshImportCmd, &s.flagNsMeshSource)
	s.networkServicesMeshListCmd.Flags().Int64Var(&s.flagNsMeshPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesMeshesCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesMeshesCmd)
}

func (s *networkServicesCmds) nsMeshName(id string) (string, error) {
	return s.nsResourceName(s.flagNsMeshLocation, "meshes", id)
}

func (s *networkServicesCmds) runNsMeshDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsMeshName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Meshes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting mesh: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete mesh [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsMeshFormat)
}

func (s *networkServicesCmds) runNsMeshDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsMeshName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Meshes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing mesh: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsMeshFormat)
}

func (s *networkServicesCmds) runNsMeshExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsMeshName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Meshes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting mesh: %w", err)
	}
	return saveAsYAML(s.flagNsMeshDestination, got)
}

func (s *networkServicesCmds) runNsMeshImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsMeshLocation)
	if err != nil {
		return err
	}
	name, err := s.nsMeshName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.Mesh{}
	if err := loadYAMLOrJSONInto(s.flagNsMeshSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Meshes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.Meshes.Create(parent, body).MeshId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating mesh: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create mesh [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsMeshFormat)
	}
	op, err := svc.Projects.Locations.Meshes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating mesh: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update mesh [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsMeshFormat)
}

func (s *networkServicesCmds) runNsMeshList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsMeshLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.Mesh
	pageToken := ""
	for {
		call := svc.Projects.Locations.Meshes.List(parent).Context(ctx)
		if s.flagNsMeshPageSize > 0 {
			call = call.PageSize(s.flagNsMeshPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagNsMeshFormat)
}
