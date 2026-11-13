package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services http-routes (#991) ---

func (s *networkServicesCmds) initNetworkServicesHttpRoutes() {
	all := []*cobra.Command{
		s.networkServicesHrDeleteCmd, s.networkServicesHrDescribeCmd,
		s.networkServicesHrExportCmd, s.networkServicesHrImportCmd,
		s.networkServicesHrListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsHrLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsHrFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesHrExportCmd, &s.flagNsHrDestination)
	nsBindImportFlags(s.networkServicesHrImportCmd, &s.flagNsHrSource)
	s.networkServicesHrListCmd.Flags().Int64Var(&s.flagNsHrPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesHttpRoutesCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesHttpRoutesCmd)
}

func (s *networkServicesCmds) nsHrName(id string) (string, error) {
	return s.nsResourceName(s.flagNsHrLocation, "httpRoutes", id)
}

func (s *networkServicesCmds) runNsHrDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsHrName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.HttpRoutes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting HTTP route: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete HTTP route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsHrFormat)
}

func (s *networkServicesCmds) runNsHrDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsHrName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HttpRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing HTTP route: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsHrFormat)
}

func (s *networkServicesCmds) runNsHrExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsHrName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HttpRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting HTTP route: %w", err)
	}
	return saveAsYAML(s.flagNsHrDestination, got)
}

func (s *networkServicesCmds) runNsHrImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsHrLocation)
	if err != nil {
		return err
	}
	name, err := s.nsHrName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.HttpRoute{}
	if err := loadYAMLOrJSONInto(s.flagNsHrSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.HttpRoutes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.HttpRoutes.Create(parent, body).HttpRouteId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating HTTP route: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create HTTP route [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsHrFormat)
	}
	op, err := svc.Projects.Locations.HttpRoutes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating HTTP route: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update HTTP route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsHrFormat)
}

func (s *networkServicesCmds) runNsHrList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsHrLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.HttpRoute
	pageToken := ""
	for {
		call := svc.Projects.Locations.HttpRoutes.List(parent).Context(ctx)
		if s.flagNsHrPageSize > 0 {
			call = call.PageSize(s.flagNsHrPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing HTTP routes: %w", err)
		}
		all = append(all, resp.HttpRoutes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsHrFormat)
}
