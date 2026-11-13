package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services gateways (#989) ---

func (s *networkServicesCmds) initNetworkServicesGateways() {
	all := []*cobra.Command{
		s.networkServicesGwDeleteCmd, s.networkServicesGwDescribeCmd,
		s.networkServicesGwExportCmd, s.networkServicesGwImportCmd,
		s.networkServicesGwListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsGwLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsGwFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesGwExportCmd, &s.flagNsGwDestination)
	nsBindImportFlags(s.networkServicesGwImportCmd, &s.flagNsGwSource)
	s.networkServicesGwListCmd.Flags().Int64Var(&s.flagNsGwPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesGatewaysCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesGatewaysCmd)
}

func (s *networkServicesCmds) nsGwName(id string) (string, error) {
	return s.nsResourceName(s.flagNsGwLocation, "gateways", id)
}

func (s *networkServicesCmds) runNsGwDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsGwName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Gateways.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting gateway: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsGwFormat)
}

func (s *networkServicesCmds) runNsGwDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsGwName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Gateways.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing gateway: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsGwFormat)
}

func (s *networkServicesCmds) runNsGwExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsGwName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Gateways.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting gateway: %w", err)
	}
	return saveAsYAML(s.flagNsGwDestination, got)
}

func (s *networkServicesCmds) runNsGwImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsGwLocation)
	if err != nil {
		return err
	}
	name, err := s.nsGwName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.Gateway{}
	if err := loadYAMLOrJSONInto(s.flagNsGwSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Gateways.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.Gateways.Create(parent, body).GatewayId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating gateway: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsGwFormat)
	}
	op, err := svc.Projects.Locations.Gateways.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating gateway: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsGwFormat)
}

func (s *networkServicesCmds) runNsGwList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsGwLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.Gateway
	pageToken := ""
	for {
		call := svc.Projects.Locations.Gateways.List(parent).Context(ctx)
		if s.flagNsGwPageSize > 0 {
			call = call.PageSize(s.flagNsGwPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing gateways: %w", err)
		}
		all = append(all, resp.Gateways...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsGwFormat)
}
