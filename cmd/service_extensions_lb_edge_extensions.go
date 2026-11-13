package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions lb-edge-extensions (#1043) ---

func (s *serviceExtensionsCmds) initServiceExtensionsLbEdgeExtensions() {
	all := []*cobra.Command{
		s.seLbEdgeDeleteCmd, s.seLbEdgeDescribeCmd,
		s.seLbEdgeExportCmd, s.seLbEdgeImportCmd,
		s.seLbEdgeListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSeLbEdgeLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagSeLbEdgeFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.seLbEdgeExportCmd, &s.flagSeLbEdgeDestination)
	nsBindImportFlags(s.seLbEdgeImportCmd, &s.flagSeLbEdgeSource)
	s.seLbEdgeListCmd.Flags().Int64Var(&s.flagSeLbEdgePageSize, "page-size", 0, "Maximum results per page")

	s.seLbEdgeExtensionsCmd.AddCommand(all...)
	s.serviceExtensionsCmd.AddCommand(s.seLbEdgeExtensionsCmd)
}

func (s *serviceExtensionsCmds) seLbEdgeName(id string) (string, error) {
	return s.seResource(s.flagSeLbEdgeLocation, "lbEdgeExtensions", id)
}

func (s *serviceExtensionsCmds) runSeLbEdgeDelete(cmd *cobra.Command, args []string) error {
	name, err := s.seLbEdgeName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.LbEdgeExtensions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting lb edge extension: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete lb edge extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeLbEdgeFormat)
}

func (s *serviceExtensionsCmds) runSeLbEdgeDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.seLbEdgeName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbEdgeExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing lb edge extension: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSeLbEdgeFormat)
}

func (s *serviceExtensionsCmds) runSeLbEdgeExport(cmd *cobra.Command, args []string) error {
	name, err := s.seLbEdgeName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbEdgeExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting lb edge extension: %w", err)
	}
	return saveAsYAML(s.flagSeLbEdgeDestination, got)
}

func (s *serviceExtensionsCmds) runSeLbEdgeImport(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeLbEdgeLocation)
	if err != nil {
		return err
	}
	name, err := s.seLbEdgeName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.LbEdgeExtension{}
	if err := loadYAMLOrJSONInto(s.flagSeLbEdgeSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.LbEdgeExtensions.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.LbEdgeExtensions.Create(parent, body).LbEdgeExtensionId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating lb edge extension: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create lb edge extension [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagSeLbEdgeFormat)
	}
	op, err := svc.Projects.Locations.LbEdgeExtensions.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating lb edge extension: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update lb edge extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeLbEdgeFormat)
}

func (s *serviceExtensionsCmds) runSeLbEdgeList(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeLbEdgeLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.LbEdgeExtension
	pageToken := ""
	for {
		call := svc.Projects.Locations.LbEdgeExtensions.List(parent).Context(ctx)
		if s.flagSeLbEdgePageSize > 0 {
			call = call.PageSize(s.flagSeLbEdgePageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSeLbEdgeFormat)
}
