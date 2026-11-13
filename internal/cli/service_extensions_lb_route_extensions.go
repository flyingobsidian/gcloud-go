package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions lb-route-extensions (#1044) ---

func (s *serviceExtensionsCmds) initServiceExtensionsLbRouteExtensions() {
	all := []*cobra.Command{
		s.seLbRouteDeleteCmd, s.seLbRouteDescribeCmd,
		s.seLbRouteExportCmd, s.seLbRouteImportCmd,
		s.seLbRouteListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSeLbRouteLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagSeLbRouteFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.seLbRouteExportCmd, &s.flagSeLbRouteDestination)
	nsBindImportFlags(s.seLbRouteImportCmd, &s.flagSeLbRouteSource)
	s.seLbRouteListCmd.Flags().Int64Var(&s.flagSeLbRoutePageSize, "page-size", 0, "Maximum results per page")

	s.seLbRouteExtensionsCmd.AddCommand(all...)
	s.serviceExtensionsCmd.AddCommand(s.seLbRouteExtensionsCmd)
}

func (s *serviceExtensionsCmds) seLbRouteName(id string) (string, error) {
	return s.seResource(s.flagSeLbRouteLocation, "lbRouteExtensions", id)
}

func (s *serviceExtensionsCmds) runSeLbRouteDelete(cmd *cobra.Command, args []string) error {
	name, err := s.seLbRouteName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.LbRouteExtensions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting lb route extension: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete lb route extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeLbRouteFormat)
}

func (s *serviceExtensionsCmds) runSeLbRouteDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.seLbRouteName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbRouteExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing lb route extension: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSeLbRouteFormat)
}

func (s *serviceExtensionsCmds) runSeLbRouteExport(cmd *cobra.Command, args []string) error {
	name, err := s.seLbRouteName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbRouteExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting lb route extension: %w", err)
	}
	return saveAsYAML(s.flagSeLbRouteDestination, got)
}

func (s *serviceExtensionsCmds) runSeLbRouteImport(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeLbRouteLocation)
	if err != nil {
		return err
	}
	name, err := s.seLbRouteName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.LbRouteExtension{}
	if err := loadYAMLOrJSONInto(s.flagSeLbRouteSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.LbRouteExtensions.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.LbRouteExtensions.Create(parent, body).LbRouteExtensionId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating lb route extension: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create lb route extension [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagSeLbRouteFormat)
	}
	op, err := svc.Projects.Locations.LbRouteExtensions.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating lb route extension: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update lb route extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeLbRouteFormat)
}

func (s *serviceExtensionsCmds) runSeLbRouteList(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeLbRouteLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.LbRouteExtension
	pageToken := ""
	for {
		call := svc.Projects.Locations.LbRouteExtensions.List(parent).Context(ctx)
		if s.flagSeLbRoutePageSize > 0 {
			call = call.PageSize(s.flagSeLbRoutePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing lb route extensions: %w", err)
		}
		all = append(all, resp.LbRouteExtensions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagSeLbRouteFormat)
}
