package cmd

import (
	"context"

	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// registerNSProfiles wires up security-profiles and security-profile-groups.
// Both are organization-scoped (issues #836, #837).
func registerNSProfiles(root *cobra.Command) {
	registerNSSecurityProfiles(root)
	registerNSSecurityProfileGroups(root)
}

func registerNSSecurityProfiles(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.SecurityProfile]{
		group: "security-profiles", singular: "security profile", collection: "securityProfiles",
		parentFn: nsOrgParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.SecurityProfile, requestID string) (*networksecurity.Operation, error) {
			return svc.Organizations.Locations.SecurityProfiles.Create(parent, body).SecurityProfileId(id).Context(ctx).Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Organizations.Locations.SecurityProfiles.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.SecurityProfile, error) {
			return svc.Organizations.Locations.SecurityProfiles.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.SecurityProfile, string, error) {
			c := svc.Organizations.Locations.SecurityProfiles.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.SecurityProfiles, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.SecurityProfile, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Organizations.Locations.SecurityProfiles.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol:         func(p *networksecurity.SecurityProfile) string { return nsBasename(p.Name) },
		secondaryCol:    func(p *networksecurity.SecurityProfile) string { return p.Type },
		secondaryHeader: "TYPE",
	}
	crud.build(root, "security-profiles", "Manage security profiles", addNSOrgFlags)
}

func registerNSSecurityProfileGroups(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.SecurityProfileGroup]{
		group: "security-profile-groups", singular: "security profile group", collection: "securityProfileGroups",
		parentFn: nsOrgParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.SecurityProfileGroup, requestID string) (*networksecurity.Operation, error) {
			return svc.Organizations.Locations.SecurityProfileGroups.Create(parent, body).SecurityProfileGroupId(id).Context(ctx).Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Organizations.Locations.SecurityProfileGroups.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.SecurityProfileGroup, error) {
			return svc.Organizations.Locations.SecurityProfileGroups.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.SecurityProfileGroup, string, error) {
			c := svc.Organizations.Locations.SecurityProfileGroups.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.SecurityProfileGroups, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.SecurityProfileGroup, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Organizations.Locations.SecurityProfileGroups.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol: func(g *networksecurity.SecurityProfileGroup) string { return nsBasename(g.Name) },
	}
	crud.build(root, "security-profile-groups", "Manage security profile groups", addNSOrgFlags)
}
