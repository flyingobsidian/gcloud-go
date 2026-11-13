package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services tls-routes (#1005) ---

func (s *networkServicesCmds) initNetworkServicesTlsRoutes() {
	all := []*cobra.Command{
		s.networkServicesTlDeleteCmd, s.networkServicesTlDescribeCmd,
		s.networkServicesTlExportCmd, s.networkServicesTlImportCmd,
		s.networkServicesTlListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsTlLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsTlFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesTlExportCmd, &s.flagNsTlDestination)
	nsBindImportFlags(s.networkServicesTlImportCmd, &s.flagNsTlSource)
	s.networkServicesTlListCmd.Flags().Int64Var(&s.flagNsTlPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesTlsRoutesCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesTlsRoutesCmd)
}

func (s *networkServicesCmds) nsTlName(id string) (string, error) {
	return s.nsResourceName(s.flagNsTlLocation, "tlsRoutes", id)
}

func (s *networkServicesCmds) runNsTlDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsTlName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TlsRoutes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting TLS route: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete TLS route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsTlFormat)
}

func (s *networkServicesCmds) runNsTlDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsTlName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TlsRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing TLS route: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsTlFormat)
}

func (s *networkServicesCmds) runNsTlExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsTlName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TlsRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting TLS route: %w", err)
	}
	return saveAsYAML(s.flagNsTlDestination, got)
}

func (s *networkServicesCmds) runNsTlImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsTlLocation)
	if err != nil {
		return err
	}
	name, err := s.nsTlName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.TlsRoute{}
	if err := loadYAMLOrJSONInto(s.flagNsTlSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.TlsRoutes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.TlsRoutes.Create(parent, body).TlsRouteId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating TLS route: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create TLS route [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsTlFormat)
	}
	op, err := svc.Projects.Locations.TlsRoutes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating TLS route: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update TLS route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsTlFormat)
}

func (s *networkServicesCmds) runNsTlList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsTlLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.TlsRoute
	pageToken := ""
	for {
		call := svc.Projects.Locations.TlsRoutes.List(parent).Context(ctx)
		if s.flagNsTlPageSize > 0 {
			call = call.PageSize(s.flagNsTlPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing TLS routes: %w", err)
		}
		all = append(all, resp.TlsRoutes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsTlFormat)
}
