package cmd

import (
	"context"

	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// registerNSFirewall wires up firewall-endpoints (organization scope) and
// firewall-endpoint-associations (project scope).
func registerNSFirewall(root *cobra.Command) {
	registerNSFirewallEndpoints(root)
	registerNSFirewallEndpointAssociations(root)
}

// firewall-endpoints are organization-scoped (issue #824).
func registerNSFirewallEndpoints(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.FirewallEndpoint]{
		group: "firewall-endpoints", singular: "firewall endpoint", collection: "firewallEndpoints",
		parentFn: nsOrgParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.FirewallEndpoint, requestID string) (*networksecurity.Operation, error) {
			c := svc.Organizations.Locations.FirewallEndpoints.Create(parent, body).FirewallEndpointId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Organizations.Locations.FirewallEndpoints.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.FirewallEndpoint, error) {
			return svc.Organizations.Locations.FirewallEndpoints.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.FirewallEndpoint, string, error) {
			c := svc.Organizations.Locations.FirewallEndpoints.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.FirewallEndpoints, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.FirewallEndpoint, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Organizations.Locations.FirewallEndpoints.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(f *networksecurity.FirewallEndpoint) string { return nsBasename(f.Name) },
		secondaryCol:    func(f *networksecurity.FirewallEndpoint) string { return f.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "firewall-endpoints", "Manage firewall endpoints", addNSOrgFlags)
}

// firewall-endpoint-associations are project-scoped (issue #823).
func registerNSFirewallEndpointAssociations(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.FirewallEndpointAssociation]{
		group: "firewall-endpoint-associations", singular: "firewall endpoint association", collection: "firewallEndpointAssociations",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.FirewallEndpointAssociation, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.FirewallEndpointAssociations.Create(parent, body).FirewallEndpointAssociationId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.FirewallEndpointAssociations.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.FirewallEndpointAssociation, error) {
			return svc.Projects.Locations.FirewallEndpointAssociations.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.FirewallEndpointAssociation, string, error) {
			c := svc.Projects.Locations.FirewallEndpointAssociations.List(parent).Context(ctx)
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
			return r.FirewallEndpointAssociations, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.FirewallEndpointAssociation, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.FirewallEndpointAssociations.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(a *networksecurity.FirewallEndpointAssociation) string { return nsBasename(a.Name) },
		secondaryCol:    func(a *networksecurity.FirewallEndpointAssociation) string { return a.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "firewall-endpoint-associations", "Manage firewall endpoint associations", addNSLocationFlag)
}
