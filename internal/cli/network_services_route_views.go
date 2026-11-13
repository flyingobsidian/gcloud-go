package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services route-views (#1001) ---
//
// Route views are nested under either a Gateway or a Mesh. Commands
// require exactly one of --gateway or --mesh to select the parent.

func (s *networkServicesCmds) initNetworkServicesRouteViews() {
	for _, c := range []*cobra.Command{s.networkServicesRvDescribeCmd, s.networkServicesRvListCmd} {
		c.Flags().StringVar(&s.flagNsRvLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsRvGateway, "gateway", "", "Name of the parent Gateway (mutually exclusive with --mesh)")
		c.Flags().StringVar(&s.flagNsRvMesh, "mesh", "", "Name of the parent Mesh (mutually exclusive with --gateway)")
		c.Flags().StringVar(&s.flagNsRvFormat, "format", "", "Output format")
	}
	s.networkServicesRvListCmd.Flags().Int64Var(&s.flagNsRvPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesRouteViewsCmd.AddCommand(s.networkServicesRvDescribeCmd, s.networkServicesRvListCmd)
	s.networkServicesCmd.AddCommand(s.networkServicesRouteViewsCmd)
}

func (s *networkServicesCmds) nsRvParent() (string, string, error) {
	if (s.flagNsRvGateway == "" && s.flagNsRvMesh == "") || (s.flagNsRvGateway != "" && s.flagNsRvMesh != "") {
		return "", "", fmt.Errorf("exactly one of --gateway or --mesh is required")
	}
	loc, err := s.nsLocationParent(s.flagNsRvLocation)
	if err != nil {
		return "", "", err
	}
	if s.flagNsRvGateway != "" {
		return fmt.Sprintf("%s/gateways/%s", loc, s.flagNsRvGateway), "gateway", nil
	}
	return fmt.Sprintf("%s/meshes/%s", loc, s.flagNsRvMesh), "mesh", nil
}

func (s *networkServicesCmds) runNsRvDescribe(cmd *cobra.Command, args []string) error {
	parent, kind, err := s.nsRvParent()
	if err != nil {
		return err
	}
	name := parent + "/routeViews/" + args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if kind == "gateway" {
		got, err := svc.Projects.Locations.Gateways.RouteViews.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing gateway route view: %w", err)
		}
		return emitFormattedTo(s.inv.stdout, got, s.flagNsRvFormat)
	}
	got, err := svc.Projects.Locations.Meshes.RouteViews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing mesh route view: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsRvFormat)
}

func (s *networkServicesCmds) runNsRvList(cmd *cobra.Command, args []string) error {
	parent, kind, err := s.nsRvParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if kind == "gateway" {
		var all []*networkservices.GatewayRouteView
		pageToken := ""
		for {
			call := svc.Projects.Locations.Gateways.RouteViews.List(parent).Context(ctx)
			if s.flagNsRvPageSize > 0 {
				call = call.PageSize(s.flagNsRvPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing gateway route views: %w", err)
			}
			all = append(all, resp.GatewayRouteViews...)
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		return emitFormattedTo(s.inv.stdout, all, s.flagNsRvFormat)
	}
	var all []*networkservices.MeshRouteView
	pageToken := ""
	for {
		call := svc.Projects.Locations.Meshes.RouteViews.List(parent).Context(ctx)
		if s.flagNsRvPageSize > 0 {
			call = call.PageSize(s.flagNsRvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing mesh route views: %w", err)
		}
		all = append(all, resp.MeshRouteViews...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsRvFormat)
}
