package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services tcp-routes (#1004) ---

func (s *networkServicesCmds) initNetworkServicesTcpRoutes() {
	all := []*cobra.Command{
		s.networkServicesTcDeleteCmd, s.networkServicesTcDescribeCmd,
		s.networkServicesTcExportCmd, s.networkServicesTcImportCmd,
		s.networkServicesTcListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsTcLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsTcFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesTcExportCmd, &s.flagNsTcDestination)
	nsBindImportFlags(s.networkServicesTcImportCmd, &s.flagNsTcSource)
	s.networkServicesTcListCmd.Flags().Int64Var(&s.flagNsTcPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesTcpRoutesCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesTcpRoutesCmd)
}

func (s *networkServicesCmds) nsTcName(id string) (string, error) {
	return s.nsResourceName(s.flagNsTcLocation, "tcpRoutes", id)
}

func (s *networkServicesCmds) runNsTcDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsTcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TcpRoutes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting TCP route: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete TCP route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsTcFormat)
}

func (s *networkServicesCmds) runNsTcDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsTcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TcpRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing TCP route: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsTcFormat)
}

func (s *networkServicesCmds) runNsTcExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsTcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TcpRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting TCP route: %w", err)
	}
	return saveAsYAML(s.flagNsTcDestination, got)
}

func (s *networkServicesCmds) runNsTcImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsTcLocation)
	if err != nil {
		return err
	}
	name, err := s.nsTcName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.TcpRoute{}
	if err := loadYAMLOrJSONInto(s.flagNsTcSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.TcpRoutes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.TcpRoutes.Create(parent, body).TcpRouteId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating TCP route: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create TCP route [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsTcFormat)
	}
	op, err := svc.Projects.Locations.TcpRoutes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating TCP route: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update TCP route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsTcFormat)
}

func (s *networkServicesCmds) runNsTcList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsTcLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.TcpRoute
	pageToken := ""
	for {
		call := svc.Projects.Locations.TcpRoutes.List(parent).Context(ctx)
		if s.flagNsTcPageSize > 0 {
			call = call.PageSize(s.flagNsTcPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing TCP routes: %w", err)
		}
		all = append(all, resp.TcpRoutes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsTcFormat)
}
