package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services grpc-routes (#990) ---

func (s *networkServicesCmds) initNetworkServicesGrpcRoutes() {
	all := []*cobra.Command{
		s.networkServicesGrDeleteCmd, s.networkServicesGrDescribeCmd,
		s.networkServicesGrExportCmd, s.networkServicesGrImportCmd,
		s.networkServicesGrListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsGrLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsGrFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesGrExportCmd, &s.flagNsGrDestination)
	nsBindImportFlags(s.networkServicesGrImportCmd, &s.flagNsGrSource)
	s.networkServicesGrListCmd.Flags().Int64Var(&s.flagNsGrPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesGrpcRoutesCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesGrpcRoutesCmd)
}

func (s *networkServicesCmds) nsGrName(id string) (string, error) {
	return s.nsResourceName(s.flagNsGrLocation, "grpcRoutes", id)
}

func (s *networkServicesCmds) runNsGrDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsGrName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GrpcRoutes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting gRPC route: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete gRPC route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsGrFormat)
}

func (s *networkServicesCmds) runNsGrDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsGrName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GrpcRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing gRPC route: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsGrFormat)
}

func (s *networkServicesCmds) runNsGrExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsGrName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GrpcRoutes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting gRPC route: %w", err)
	}
	return saveAsYAML(s.flagNsGrDestination, got)
}

func (s *networkServicesCmds) runNsGrImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsGrLocation)
	if err != nil {
		return err
	}
	name, err := s.nsGrName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.GrpcRoute{}
	if err := loadYAMLOrJSONInto(s.flagNsGrSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.GrpcRoutes.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.GrpcRoutes.Create(parent, body).GrpcRouteId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating gRPC route: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create gRPC route [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsGrFormat)
	}
	op, err := svc.Projects.Locations.GrpcRoutes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating gRPC route: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update gRPC route [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsGrFormat)
}

func (s *networkServicesCmds) runNsGrList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsGrLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.GrpcRoute
	pageToken := ""
	for {
		call := svc.Projects.Locations.GrpcRoutes.List(parent).Context(ctx)
		if s.flagNsGrPageSize > 0 {
			call = call.PageSize(s.flagNsGrPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing gRPC routes: %w", err)
		}
		all = append(all, resp.GrpcRoutes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsGrFormat)
}
