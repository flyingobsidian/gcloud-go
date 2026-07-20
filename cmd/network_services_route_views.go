package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services route-views (#1001) ---
//
// Route views are nested under either a Gateway or a Mesh. Commands
// require exactly one of --gateway or --mesh to select the parent.

var networkServicesRouteViewsCmd = &cobra.Command{Use: "route-views", Short: "View route views for a Gateway or Mesh"}

var (
	flagNsRvLocation string
	flagNsRvGateway  string
	flagNsRvMesh     string
	flagNsRvFormat   string
	flagNsRvPageSize int64
)

var (
	networkServicesRvDescribeCmd = &cobra.Command{
		Use: "describe ROUTE_VIEW", Short: "Describe a route view for a Gateway or Mesh",
		Args: cobra.ExactArgs(1), RunE: runNsRvDescribe,
	}
	networkServicesRvListCmd = &cobra.Command{
		Use: "list", Short: "List route views for a Gateway or Mesh",
		Args: cobra.NoArgs, RunE: runNsRvList,
	}
)

func init() {
	for _, c := range []*cobra.Command{networkServicesRvDescribeCmd, networkServicesRvListCmd} {
		c.Flags().StringVar(&flagNsRvLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsRvGateway, "gateway", "", "Name of the parent Gateway (mutually exclusive with --mesh)")
		c.Flags().StringVar(&flagNsRvMesh, "mesh", "", "Name of the parent Mesh (mutually exclusive with --gateway)")
		c.Flags().StringVar(&flagNsRvFormat, "format", "", "Output format")
	}
	networkServicesRvListCmd.Flags().Int64Var(&flagNsRvPageSize, "page-size", 0, "Maximum results per page")

	networkServicesRouteViewsCmd.AddCommand(networkServicesRvDescribeCmd, networkServicesRvListCmd)
	networkServicesCmd.AddCommand(networkServicesRouteViewsCmd)
}

func nsRvParent() (string, string, error) {
	if (flagNsRvGateway == "" && flagNsRvMesh == "") || (flagNsRvGateway != "" && flagNsRvMesh != "") {
		return "", "", fmt.Errorf("exactly one of --gateway or --mesh is required")
	}
	loc, err := nsLocationParent(flagNsRvLocation)
	if err != nil {
		return "", "", err
	}
	if flagNsRvGateway != "" {
		return fmt.Sprintf("%s/gateways/%s", loc, flagNsRvGateway), "gateway", nil
	}
	return fmt.Sprintf("%s/meshes/%s", loc, flagNsRvMesh), "mesh", nil
}

func runNsRvDescribe(cmd *cobra.Command, args []string) error {
	parent, kind, err := nsRvParent()
	if err != nil {
		return err
	}
	name := parent + "/routeViews/" + args[0]
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if kind == "gateway" {
		got, err := svc.Projects.Locations.Gateways.RouteViews.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing gateway route view: %w", err)
		}
		return emitFormatted(got, flagNsRvFormat)
	}
	got, err := svc.Projects.Locations.Meshes.RouteViews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing mesh route view: %w", err)
	}
	return emitFormatted(got, flagNsRvFormat)
}

func runNsRvList(cmd *cobra.Command, args []string) error {
	parent, kind, err := nsRvParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if kind == "gateway" {
		var all []*networkservices.GatewayRouteView
		pageToken := ""
		for {
			call := svc.Projects.Locations.Gateways.RouteViews.List(parent).Context(ctx)
			if flagNsRvPageSize > 0 {
				call = call.PageSize(flagNsRvPageSize)
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
		return emitFormatted(all, flagNsRvFormat)
	}
	var all []*networkservices.MeshRouteView
	pageToken := ""
	for {
		call := svc.Projects.Locations.Meshes.RouteViews.List(parent).Context(ctx)
		if flagNsRvPageSize > 0 {
			call = call.PageSize(flagNsRvPageSize)
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
	return emitFormatted(all, flagNsRvFormat)
}
