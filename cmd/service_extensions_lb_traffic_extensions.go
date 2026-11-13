package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions lb-traffic-extensions (#1045) ---

func (s *serviceExtensionsCmds) initServiceExtensionsLbTrafficExtensions() {
	all := []*cobra.Command{
		s.seLbTrafficDeleteCmd, s.seLbTrafficDescribeCmd,
		s.seLbTrafficExportCmd, s.seLbTrafficImportCmd,
		s.seLbTrafficListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSeLbTrafficLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagSeLbTrafficFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.seLbTrafficExportCmd, &s.flagSeLbTrafficDestination)
	nsBindImportFlags(s.seLbTrafficImportCmd, &s.flagSeLbTrafficSource)
	s.seLbTrafficListCmd.Flags().Int64Var(&s.flagSeLbTrafficPageSize, "page-size", 0, "Maximum results per page")

	s.seLbTrafficExtensionsCmd.AddCommand(all...)
	s.serviceExtensionsCmd.AddCommand(s.seLbTrafficExtensionsCmd)
}

func (s *serviceExtensionsCmds) seLbTrafficName(id string) (string, error) {
	return s.seResource(s.flagSeLbTrafficLocation, "lbTrafficExtensions", id)
}

func (s *serviceExtensionsCmds) runSeLbTrafficDelete(cmd *cobra.Command, args []string) error {
	name, err := s.seLbTrafficName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.LbTrafficExtensions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting lb traffic extension: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete lb traffic extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeLbTrafficFormat)
}

func (s *serviceExtensionsCmds) runSeLbTrafficDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.seLbTrafficName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbTrafficExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing lb traffic extension: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSeLbTrafficFormat)
}

func (s *serviceExtensionsCmds) runSeLbTrafficExport(cmd *cobra.Command, args []string) error {
	name, err := s.seLbTrafficName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LbTrafficExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting lb traffic extension: %w", err)
	}
	return saveAsYAML(s.flagSeLbTrafficDestination, got)
}

func (s *serviceExtensionsCmds) runSeLbTrafficImport(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeLbTrafficLocation)
	if err != nil {
		return err
	}
	name, err := s.seLbTrafficName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.LbTrafficExtension{}
	if err := loadYAMLOrJSONInto(s.flagSeLbTrafficSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.LbTrafficExtensions.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.LbTrafficExtensions.Create(parent, body).LbTrafficExtensionId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating lb traffic extension: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create lb traffic extension [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagSeLbTrafficFormat)
	}
	op, err := svc.Projects.Locations.LbTrafficExtensions.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating lb traffic extension: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update lb traffic extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeLbTrafficFormat)
}

func (s *serviceExtensionsCmds) runSeLbTrafficList(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeLbTrafficLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.LbTrafficExtension
	pageToken := ""
	for {
		call := svc.Projects.Locations.LbTrafficExtensions.List(parent).Context(ctx)
		if s.flagSeLbTrafficPageSize > 0 {
			call = call.PageSize(s.flagSeLbTrafficPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing lb traffic extensions: %w", err)
		}
		all = append(all, resp.LbTrafficExtensions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagSeLbTrafficFormat)
}
