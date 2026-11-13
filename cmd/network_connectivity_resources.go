package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	networkconnectivity "google.golang.org/api/networkconnectivity/v1"
)

// --- internal-ranges (#902) ---

func (grp *networkConnectivityCmds) registerNCInternalRanges(root *cobra.Command) {
	crud := &ncCRUD[networkconnectivity.InternalRange]{
		s:     grp,
		group: "internal-ranges", singular: "internal range", collection: "internalRanges",
		parentFn: grp.ncProjectParent,
		createFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *networkconnectivity.InternalRange, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.InternalRanges.Create(parent, body).InternalRangeId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.InternalRanges.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networkconnectivity.Service, name string) (*networkconnectivity.InternalRange, error) {
			return svc.Projects.Locations.InternalRanges.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*networkconnectivity.InternalRange, string, error) {
			c := svc.Projects.Locations.InternalRanges.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			if filter != "" {
				c = c.Filter(filter)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.InternalRanges, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networkconnectivity.Service, name string, body *networkconnectivity.InternalRange, mask, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.InternalRanges.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(r *networkconnectivity.InternalRange) string { return ncBasename(r.Name) },
		secondaryCol:    func(r *networkconnectivity.InternalRange) string { return r.IpCidrRange },
		secondaryHeader: "IP_CIDR_RANGE",
	}
	crud.build(root, "internal-ranges", "Manage internal ranges", grp.addNCLocationFlag, true)
}

// --- policy-based-routes (#907) ---

func (grp *networkConnectivityCmds) registerNCPolicyBasedRoutes(root *cobra.Command) {
	crud := &ncCRUD[networkconnectivity.PolicyBasedRoute]{
		s:     grp,
		group: "policy-based-routes", singular: "policy based route", collection: "global/policyBasedRoutes",
		parentFn: grp.ncGlobalParent,
		createFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *networkconnectivity.PolicyBasedRoute, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Global.PolicyBasedRoutes.Create(parent, body).PolicyBasedRouteId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Global.PolicyBasedRoutes.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networkconnectivity.Service, name string) (*networkconnectivity.PolicyBasedRoute, error) {
			return svc.Projects.Locations.Global.PolicyBasedRoutes.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*networkconnectivity.PolicyBasedRoute, string, error) {
			c := svc.Projects.Locations.Global.PolicyBasedRoutes.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			if filter != "" {
				c = c.Filter(filter)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.PolicyBasedRoutes, r.NextPageToken, nil
		},
		nameCol:         func(r *networkconnectivity.PolicyBasedRoute) string { return ncBasename(r.Name) },
		secondaryCol:    func(r *networkconnectivity.PolicyBasedRoute) string { return r.Network },
		secondaryHeader: "NETWORK",
	}
	// Policy-based routes are global; no --location flag.
	crud.build(root, "policy-based-routes", "Manage policy-based routes", func(_ ...*cobra.Command) {}, false)
}

// --- regional-endpoints (#908) ---

func (grp *networkConnectivityCmds) registerNCRegionalEndpoints(root *cobra.Command) {
	crud := &ncCRUD[networkconnectivity.RegionalEndpoint]{
		s:     grp,
		group: "regional-endpoints", singular: "regional endpoint", collection: "regionalEndpoints",
		parentFn: grp.ncProjectParent,
		createFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *networkconnectivity.RegionalEndpoint, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.RegionalEndpoints.Create(parent, body).RegionalEndpointId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.RegionalEndpoints.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networkconnectivity.Service, name string) (*networkconnectivity.RegionalEndpoint, error) {
			return svc.Projects.Locations.RegionalEndpoints.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*networkconnectivity.RegionalEndpoint, string, error) {
			c := svc.Projects.Locations.RegionalEndpoints.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			if filter != "" {
				c = c.Filter(filter)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.RegionalEndpoints, r.NextPageToken, nil
		},
		nameCol:         func(r *networkconnectivity.RegionalEndpoint) string { return ncBasename(r.Name) },
		secondaryCol:    func(r *networkconnectivity.RegionalEndpoint) string { return r.Address },
		secondaryHeader: "ADDRESS",
	}
	crud.build(root, "regional-endpoints", "Manage regional endpoints", grp.addNCLocationFlag, false)
}

// --- service-connection-policies (#909) ---

func (grp *networkConnectivityCmds) registerNCServiceConnectionPolicies(root *cobra.Command) {
	crud := &ncCRUD[networkconnectivity.ServiceConnectionPolicy]{
		s:     grp,
		group: "service-connection-policies", singular: "service connection policy", collection: "serviceConnectionPolicies",
		parentFn: grp.ncProjectParent,
		createFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *networkconnectivity.ServiceConnectionPolicy, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.ServiceConnectionPolicies.Create(parent, body).ServiceConnectionPolicyId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.ServiceConnectionPolicies.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networkconnectivity.Service, name string) (*networkconnectivity.ServiceConnectionPolicy, error) {
			return svc.Projects.Locations.ServiceConnectionPolicies.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*networkconnectivity.ServiceConnectionPolicy, string, error) {
			c := svc.Projects.Locations.ServiceConnectionPolicies.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			if filter != "" {
				c = c.Filter(filter)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.ServiceConnectionPolicies, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networkconnectivity.Service, name string, body *networkconnectivity.ServiceConnectionPolicy, mask, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.ServiceConnectionPolicies.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(r *networkconnectivity.ServiceConnectionPolicy) string { return ncBasename(r.Name) },
		secondaryCol:    func(r *networkconnectivity.ServiceConnectionPolicy) string { return r.ServiceClass },
		secondaryHeader: "SERVICE_CLASS",
	}
	crud.build(root, "service-connection-policies", "Manage service connection policies", grp.addNCLocationFlag, true)
}

// --- multicloud-data-transfer-configs (#904) ---

func (grp *networkConnectivityCmds) registerNCMulticloudDataTransferConfigs(root *cobra.Command) {
	crud := &ncCRUD[networkconnectivity.MulticloudDataTransferConfig]{
		s:     grp,
		group: "multicloud-data-transfer-configs", singular: "multicloud data transfer config", collection: "multicloudDataTransferConfigs",
		parentFn: grp.ncProjectParent,
		createFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *networkconnectivity.MulticloudDataTransferConfig, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.MulticloudDataTransferConfigs.Create(parent, body).MulticloudDataTransferConfigId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.MulticloudDataTransferConfigs.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networkconnectivity.Service, name string) (*networkconnectivity.MulticloudDataTransferConfig, error) {
			return svc.Projects.Locations.MulticloudDataTransferConfigs.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*networkconnectivity.MulticloudDataTransferConfig, string, error) {
			c := svc.Projects.Locations.MulticloudDataTransferConfigs.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			if filter != "" {
				c = c.Filter(filter)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.MulticloudDataTransferConfigs, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networkconnectivity.Service, name string, body *networkconnectivity.MulticloudDataTransferConfig, mask, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.MulticloudDataTransferConfigs.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol: func(r *networkconnectivity.MulticloudDataTransferConfig) string { return ncBasename(r.Name) },
		secondaryCol: func(r *networkconnectivity.MulticloudDataTransferConfig) string {
			return fmt.Sprintf("%d", r.DestinationsCount)
		},
		secondaryHeader: "DESTINATIONS",
	}
	crud.build(root, "multicloud-data-transfer-configs", "Manage multicloud data transfer configs", grp.addNCLocationFlag, true)
}

// --- multicloud-data-transfer-supported-services (#905) ---

func (grp *networkConnectivityCmds) registerNCMulticloudDataTransferSupportedServices(root *cobra.Command) {
	g := &cobra.Command{Use: "multicloud-data-transfer-supported-services", Short: "Manage multicloud data transfer supported services"}

	describe := &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a supported service",
		Args: cobra.ExactArgs(1), RunE: grp.runNCSupportedServiceDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: "List supported services in a location",
		Args: cobra.NoArgs, RunE: grp.runNCSupportedServiceList,
	}
	grp.addNCLocationFlag(describe, list)
	grp.addNCFormatFlag(describe, list)

	g.AddCommand(describe, list)
	root.AddCommand(g)
}

func (grp *networkConnectivityCmds) runNCSupportedServiceDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	parent, err := grp.ncProjectParent()
	if err != nil {
		return err
	}
	name := ncChild(parent, "multicloudDataTransferSupportedServices", args[0])
	got, err := svc.Projects.Locations.MulticloudDataTransferSupportedServices.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing supported service: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, got, grp.flagNCFormat)
}

func (grp *networkConnectivityCmds) runNCSupportedServiceList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	parent, err := grp.ncProjectParent()
	if err != nil {
		return err
	}
	var all []*networkconnectivity.MulticloudDataTransferSupportedService
	pageToken := ""
	for {
		call := svc.Projects.Locations.MulticloudDataTransferSupportedServices.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing supported services: %w", err)
		}
		all = append(all, resp.MulticloudDataTransferSupportedServices...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagNCFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagNCFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-50s\n", "NAME")
	for _, s := range all {
		fmt.Fprintf(grp.inv.stdout, "%-50s\n", ncBasename(s.Name))
	}
	return nil
}

// --- transports (#911) ---

func (grp *networkConnectivityCmds) registerNCTransports(root *cobra.Command) {
	crud := &ncCRUD[networkconnectivity.Transport]{
		s:     grp,
		group: "transports", singular: "transport", collection: "transports",
		parentFn: grp.ncProjectParent,
		createFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *networkconnectivity.Transport, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Transports.Create(parent, body).TransportId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Transports.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networkconnectivity.Service, name string) (*networkconnectivity.Transport, error) {
			return svc.Projects.Locations.Transports.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*networkconnectivity.Transport, string, error) {
			c := svc.Projects.Locations.Transports.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			if filter != "" {
				c = c.Filter(filter)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.Transports, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networkconnectivity.Service, name string, body *networkconnectivity.Transport, mask, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error) {
			c := svc.Projects.Locations.Transports.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(r *networkconnectivity.Transport) string { return ncBasename(r.Name) },
		secondaryCol:    func(r *networkconnectivity.Transport) string { return r.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "transports", "Manage transports", grp.addNCLocationFlag, true)
}
