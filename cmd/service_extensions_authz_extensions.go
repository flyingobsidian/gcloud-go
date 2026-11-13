package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions authz-extensions (#1042) ---

func (s *serviceExtensionsCmds) initServiceExtensionsAuthzExtensions() {
	all := []*cobra.Command{
		s.seAuthzDeleteCmd, s.seAuthzDescribeCmd,
		s.seAuthzExportCmd, s.seAuthzImportCmd,
		s.seAuthzListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSeAuthzLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagSeAuthzFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.seAuthzExportCmd, &s.flagSeAuthzDestination)
	nsBindImportFlags(s.seAuthzImportCmd, &s.flagSeAuthzSource)
	s.seAuthzListCmd.Flags().Int64Var(&s.flagSeAuthzPageSize, "page-size", 0, "Maximum results per page")

	s.seAuthzExtensionsCmd.AddCommand(all...)
	s.serviceExtensionsCmd.AddCommand(s.seAuthzExtensionsCmd)
}

func (s *serviceExtensionsCmds) seAuthzName(id string) (string, error) {
	return s.seResource(s.flagSeAuthzLocation, "authzExtensions", id)
}

func (s *serviceExtensionsCmds) runSeAuthzDelete(cmd *cobra.Command, args []string) error {
	name, err := s.seAuthzName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AuthzExtensions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting authz extension: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete authz extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeAuthzFormat)
}

func (s *serviceExtensionsCmds) runSeAuthzDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.seAuthzName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthzExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing authz extension: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSeAuthzFormat)
}

func (s *serviceExtensionsCmds) runSeAuthzExport(cmd *cobra.Command, args []string) error {
	name, err := s.seAuthzName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthzExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting authz extension: %w", err)
	}
	return saveAsYAML(s.flagSeAuthzDestination, got)
}

func (s *serviceExtensionsCmds) runSeAuthzImport(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeAuthzLocation)
	if err != nil {
		return err
	}
	name, err := s.seAuthzName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.AuthzExtension{}
	if err := loadYAMLOrJSONInto(s.flagSeAuthzSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.AuthzExtensions.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.AuthzExtensions.Create(parent, body).AuthzExtensionId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating authz extension: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create authz extension [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagSeAuthzFormat)
	}
	op, err := svc.Projects.Locations.AuthzExtensions.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating authz extension: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update authz extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSeAuthzFormat)
}

func (s *serviceExtensionsCmds) runSeAuthzList(cmd *cobra.Command, args []string) error {
	parent, err := s.seLocationParent(s.flagSeAuthzLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.AuthzExtension
	pageToken := ""
	for {
		call := svc.Projects.Locations.AuthzExtensions.List(parent).Context(ctx)
		if s.flagSeAuthzPageSize > 0 {
			call = call.PageSize(s.flagSeAuthzPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing authz extensions: %w", err)
		}
		all = append(all, resp.AuthzExtensions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagSeAuthzFormat)
}
