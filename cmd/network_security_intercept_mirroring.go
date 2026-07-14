package cmd

import (
	"context"

	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// registerNSInterceptMirroring wires up all 8 intercept-* and mirroring-*
// subgroups (issues #826-#831 minus mirroring-endpoints which is not in the
// v0.279.0 Go SDK, and #845/#846 for the endpoint-group-associations).
func registerNSInterceptMirroring(root *cobra.Command) {
	registerNSInterceptDeploymentGroups(root)
	registerNSInterceptDeployments(root)
	registerNSInterceptEndpointGroups(root)
	registerNSInterceptEndpointGroupAssociations(root)
	registerNSMirroringDeploymentGroups(root)
	registerNSMirroringDeployments(root)
	registerNSMirroringEndpointGroups(root)
	registerNSMirroringEndpointGroupAssociations(root)
}

func registerNSInterceptDeploymentGroups(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.InterceptDeploymentGroup]{
		group: "intercept-deployment-groups", singular: "intercept deployment group", collection: "interceptDeploymentGroups",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.InterceptDeploymentGroup, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptDeploymentGroups.Create(parent, body).InterceptDeploymentGroupId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptDeploymentGroups.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.InterceptDeploymentGroup, error) {
			return svc.Projects.Locations.InterceptDeploymentGroups.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.InterceptDeploymentGroup, string, error) {
			c := svc.Projects.Locations.InterceptDeploymentGroups.List(parent).Context(ctx)
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
			return r.InterceptDeploymentGroups, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.InterceptDeploymentGroup, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptDeploymentGroups.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(g *networksecurity.InterceptDeploymentGroup) string { return nsBasename(g.Name) },
		secondaryCol:    func(g *networksecurity.InterceptDeploymentGroup) string { return g.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "intercept-deployment-groups", "Manage intercept deployment groups", addNSLocationFlag)
}

func registerNSInterceptDeployments(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.InterceptDeployment]{
		group: "intercept-deployments", singular: "intercept deployment", collection: "interceptDeployments",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.InterceptDeployment, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptDeployments.Create(parent, body).InterceptDeploymentId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptDeployments.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.InterceptDeployment, error) {
			return svc.Projects.Locations.InterceptDeployments.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.InterceptDeployment, string, error) {
			c := svc.Projects.Locations.InterceptDeployments.List(parent).Context(ctx)
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
			return r.InterceptDeployments, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.InterceptDeployment, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptDeployments.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(d *networksecurity.InterceptDeployment) string { return nsBasename(d.Name) },
		secondaryCol:    func(d *networksecurity.InterceptDeployment) string { return d.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "intercept-deployments", "Manage intercept deployments", addNSLocationFlag)
}

func registerNSInterceptEndpointGroups(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.InterceptEndpointGroup]{
		group: "intercept-endpoint-groups", singular: "intercept endpoint group", collection: "interceptEndpointGroups",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.InterceptEndpointGroup, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptEndpointGroups.Create(parent, body).InterceptEndpointGroupId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptEndpointGroups.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.InterceptEndpointGroup, error) {
			return svc.Projects.Locations.InterceptEndpointGroups.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.InterceptEndpointGroup, string, error) {
			c := svc.Projects.Locations.InterceptEndpointGroups.List(parent).Context(ctx)
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
			return r.InterceptEndpointGroups, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.InterceptEndpointGroup, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptEndpointGroups.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(g *networksecurity.InterceptEndpointGroup) string { return nsBasename(g.Name) },
		secondaryCol:    func(g *networksecurity.InterceptEndpointGroup) string { return g.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "intercept-endpoint-groups", "Manage intercept endpoint groups", addNSLocationFlag)
}

func registerNSInterceptEndpointGroupAssociations(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.InterceptEndpointGroupAssociation]{
		group: "intercept-endpoint-group-associations", singular: "intercept endpoint group association", collection: "interceptEndpointGroupAssociations",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.InterceptEndpointGroupAssociation, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptEndpointGroupAssociations.Create(parent, body).InterceptEndpointGroupAssociationId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptEndpointGroupAssociations.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.InterceptEndpointGroupAssociation, error) {
			return svc.Projects.Locations.InterceptEndpointGroupAssociations.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.InterceptEndpointGroupAssociation, string, error) {
			c := svc.Projects.Locations.InterceptEndpointGroupAssociations.List(parent).Context(ctx)
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
			return r.InterceptEndpointGroupAssociations, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.InterceptEndpointGroupAssociation, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.InterceptEndpointGroupAssociations.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(a *networksecurity.InterceptEndpointGroupAssociation) string { return nsBasename(a.Name) },
		secondaryCol:    func(a *networksecurity.InterceptEndpointGroupAssociation) string { return a.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "intercept-endpoint-group-associations", "Manage intercept endpoint group associations", addNSLocationFlag)
}

func registerNSMirroringDeploymentGroups(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.MirroringDeploymentGroup]{
		group: "mirroring-deployment-groups", singular: "mirroring deployment group", collection: "mirroringDeploymentGroups",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.MirroringDeploymentGroup, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringDeploymentGroups.Create(parent, body).MirroringDeploymentGroupId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringDeploymentGroups.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.MirroringDeploymentGroup, error) {
			return svc.Projects.Locations.MirroringDeploymentGroups.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.MirroringDeploymentGroup, string, error) {
			c := svc.Projects.Locations.MirroringDeploymentGroups.List(parent).Context(ctx)
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
			return r.MirroringDeploymentGroups, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.MirroringDeploymentGroup, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringDeploymentGroups.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(g *networksecurity.MirroringDeploymentGroup) string { return nsBasename(g.Name) },
		secondaryCol:    func(g *networksecurity.MirroringDeploymentGroup) string { return g.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "mirroring-deployment-groups", "Manage mirroring deployment groups", addNSLocationFlag)
}

func registerNSMirroringDeployments(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.MirroringDeployment]{
		group: "mirroring-deployments", singular: "mirroring deployment", collection: "mirroringDeployments",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.MirroringDeployment, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringDeployments.Create(parent, body).MirroringDeploymentId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringDeployments.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.MirroringDeployment, error) {
			return svc.Projects.Locations.MirroringDeployments.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.MirroringDeployment, string, error) {
			c := svc.Projects.Locations.MirroringDeployments.List(parent).Context(ctx)
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
			return r.MirroringDeployments, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.MirroringDeployment, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringDeployments.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(d *networksecurity.MirroringDeployment) string { return nsBasename(d.Name) },
		secondaryCol:    func(d *networksecurity.MirroringDeployment) string { return d.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "mirroring-deployments", "Manage mirroring deployments", addNSLocationFlag)
}

func registerNSMirroringEndpointGroups(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.MirroringEndpointGroup]{
		group: "mirroring-endpoint-groups", singular: "mirroring endpoint group", collection: "mirroringEndpointGroups",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.MirroringEndpointGroup, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringEndpointGroups.Create(parent, body).MirroringEndpointGroupId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringEndpointGroups.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.MirroringEndpointGroup, error) {
			return svc.Projects.Locations.MirroringEndpointGroups.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.MirroringEndpointGroup, string, error) {
			c := svc.Projects.Locations.MirroringEndpointGroups.List(parent).Context(ctx)
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
			return r.MirroringEndpointGroups, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.MirroringEndpointGroup, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringEndpointGroups.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(g *networksecurity.MirroringEndpointGroup) string { return nsBasename(g.Name) },
		secondaryCol:    func(g *networksecurity.MirroringEndpointGroup) string { return g.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "mirroring-endpoint-groups", "Manage mirroring endpoint groups", addNSLocationFlag)
}

func registerNSMirroringEndpointGroupAssociations(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.MirroringEndpointGroupAssociation]{
		group: "mirroring-endpoint-group-associations", singular: "mirroring endpoint group association", collection: "mirroringEndpointGroupAssociations",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.MirroringEndpointGroupAssociation, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringEndpointGroupAssociations.Create(parent, body).MirroringEndpointGroupAssociationId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringEndpointGroupAssociations.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.MirroringEndpointGroupAssociation, error) {
			return svc.Projects.Locations.MirroringEndpointGroupAssociations.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.MirroringEndpointGroupAssociation, string, error) {
			c := svc.Projects.Locations.MirroringEndpointGroupAssociations.List(parent).Context(ctx)
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
			return r.MirroringEndpointGroupAssociations, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.MirroringEndpointGroupAssociation, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.MirroringEndpointGroupAssociations.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(a *networksecurity.MirroringEndpointGroupAssociation) string { return nsBasename(a.Name) },
		secondaryCol:    func(a *networksecurity.MirroringEndpointGroupAssociation) string { return a.State },
		secondaryHeader: "STATE",
	}
	crud.build(root, "mirroring-endpoint-group-associations", "Manage mirroring endpoint group associations", addNSLocationFlag)
}
