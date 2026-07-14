package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// gcpNS is a convenience wrapper for the networksecurity v1 client used
// across the network-security subgroup files.
func gcpNS(ctx context.Context) (*networksecurity.Service, error) {
	return gcp.NetworkSecurityService(ctx, flagAccount)
}

// registerNSPolicies wires up every project-scoped policy / config resource:
// authorization-policies, authz-policies, backend-authentication-configs,
// client-tls-policies, dns-threat-detectors, server-tls-policies,
// tls-inspection-policies, url-lists, and gateway-security-policies (with its
// rules subgroup).
func registerNSPolicies(root *cobra.Command) {
	registerNSAuthorizationPolicies(root)
	registerNSAuthzPolicies(root)
	registerNSBackendAuthenticationConfigs(root)
	registerNSClientTlsPolicies(root)
	registerNSDnsThreatDetectors(root)
	registerNSServerTlsPolicies(root)
	registerNSTlsInspectionPolicies(root)
	registerNSUrlLists(root)
	registerNSGatewaySecurityPolicies(root)
}

func registerNSAuthorizationPolicies(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.AuthorizationPolicy]{
		group: "authorization-policies", singular: "authorization policy", collection: "authorizationPolicies",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.AuthorizationPolicy, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.AuthorizationPolicies.Create(parent, body).AuthorizationPolicyId(id).Context(ctx)
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.AuthorizationPolicies.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.AuthorizationPolicy, error) {
			return svc.Projects.Locations.AuthorizationPolicies.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.AuthorizationPolicy, string, error) {
			c := svc.Projects.Locations.AuthorizationPolicies.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.AuthorizationPolicies, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.AuthorizationPolicy, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.AuthorizationPolicies.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol:         func(p *networksecurity.AuthorizationPolicy) string { return nsBasename(p.Name) },
		secondaryCol:    func(p *networksecurity.AuthorizationPolicy) string { return p.Action },
		secondaryHeader: "ACTION",
	}
	crud.build(root, "authorization-policies", "Manage authorization policies", addNSLocationFlag)
}

func registerNSAuthzPolicies(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.AuthzPolicy]{
		group: "authz-policies", singular: "authz policy", collection: "authzPolicies",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.AuthzPolicy, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.AuthzPolicies.Create(parent, body).AuthzPolicyId(id).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.AuthzPolicies.Delete(name).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.AuthzPolicy, error) {
			return svc.Projects.Locations.AuthzPolicies.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.AuthzPolicy, string, error) {
			c := svc.Projects.Locations.AuthzPolicies.List(parent).Context(ctx)
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
			return r.AuthzPolicies, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.AuthzPolicy, mask, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.AuthzPolicies.Patch(name, body).UpdateMask(mask).Context(ctx)
			if requestID != "" {
				c = c.RequestId(requestID)
			}
			return c.Do()
		},
		nameCol:         func(p *networksecurity.AuthzPolicy) string { return nsBasename(p.Name) },
		secondaryCol:    func(p *networksecurity.AuthzPolicy) string { return p.Action },
		secondaryHeader: "ACTION",
	}
	crud.build(root, "authz-policies", "Manage authz policies", addNSLocationFlag)
}

func registerNSBackendAuthenticationConfigs(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.BackendAuthenticationConfig]{
		group: "backend-authentication-configs", singular: "backend authentication config", collection: "backendAuthenticationConfigs",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.BackendAuthenticationConfig, requestID string) (*networksecurity.Operation, error) {
			c := svc.Projects.Locations.BackendAuthenticationConfigs.Create(parent, body).BackendAuthenticationConfigId(id).Context(ctx)
			return c.Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.BackendAuthenticationConfigs.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.BackendAuthenticationConfig, error) {
			return svc.Projects.Locations.BackendAuthenticationConfigs.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.BackendAuthenticationConfig, string, error) {
			c := svc.Projects.Locations.BackendAuthenticationConfigs.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.BackendAuthenticationConfigs, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.BackendAuthenticationConfig, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.BackendAuthenticationConfigs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol: func(p *networksecurity.BackendAuthenticationConfig) string { return nsBasename(p.Name) },
	}
	crud.build(root, "backend-authentication-configs", "Manage backend authentication configs", addNSLocationFlag)
}

func registerNSClientTlsPolicies(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.ClientTlsPolicy]{
		group: "client-tls-policies", singular: "client TLS policy", collection: "clientTlsPolicies",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.ClientTlsPolicy, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.ClientTlsPolicies.Create(parent, body).ClientTlsPolicyId(id).Context(ctx).Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.ClientTlsPolicies.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.ClientTlsPolicy, error) {
			return svc.Projects.Locations.ClientTlsPolicies.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.ClientTlsPolicy, string, error) {
			c := svc.Projects.Locations.ClientTlsPolicies.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.ClientTlsPolicies, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.ClientTlsPolicy, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.ClientTlsPolicies.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol: func(p *networksecurity.ClientTlsPolicy) string { return nsBasename(p.Name) },
		secondaryCol: func(p *networksecurity.ClientTlsPolicy) string { return p.Sni },
		secondaryHeader: "SNI",
	}
	crud.build(root, "client-tls-policies", "Manage client TLS policies", addNSLocationFlag)
}

// DnsThreatDetectors is unusual: Create/Delete/Patch are synchronous and
// return the resource (or Empty) directly rather than an LRO. It is wired
// with dedicated cobra RunEs instead of the shared nsCRUD framework.
func registerNSDnsThreatDetectors(root *cobra.Command) {
	group := &cobra.Command{Use: "dns-threat-detectors", Short: "Manage DNS threat detectors"}
	create := &cobra.Command{
		Use: "create DETECTOR", Short: "Create a DNS threat detector from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runNSDNSCreate,
	}
	del := &cobra.Command{
		Use: "delete DETECTOR", Short: "Delete a DNS threat detector",
		Args: cobra.ExactArgs(1), RunE: runNSDNSDelete,
	}
	desc := &cobra.Command{
		Use: "describe DETECTOR", Short: "Describe a DNS threat detector",
		Args: cobra.ExactArgs(1), RunE: runNSDNSDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: "List DNS threat detectors in a location",
		Args: cobra.NoArgs, RunE: runNSDNSList,
	}
	upd := &cobra.Command{
		Use: "update DETECTOR", Short: "Update a DNS threat detector from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runNSDNSUpdate,
	}
	addNSLocationFlag(create, del, desc, list, upd)
	addNSFormatFlag(desc, list)
	addNSCreateConfigFlag(create)
	addNSUpdateConfigFlag(upd)
	group.AddCommand(create, del, desc, list, upd)
	root.AddCommand(group)
}

func runNSDNSCreate(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecurity.DnsThreatDetector{}
	if err := loadYAMLOrJSONInto(flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcpNS(ctx)
	if err != nil {
		return err
	}
	res, err := svc.Projects.Locations.DnsThreatDetectors.Create(parent, body).DnsThreatDetectorId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DNS threat detector: %w", err)
	}
	return emitFormatted(res, "")
}

func runNSDNSDelete(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcpNS(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "dnsThreatDetectors", args[0])
	if _, err := svc.Projects.Locations.DnsThreatDetectors.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting DNS threat detector: %w", err)
	}
	fmt.Printf("Deleted DNS threat detector %s.\n", args[0])
	return nil
}

func runNSDNSDescribe(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcpNS(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "dnsThreatDetectors", args[0])
	got, err := svc.Projects.Locations.DnsThreatDetectors.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS threat detector: %w", err)
	}
	return emitFormatted(got, flagNSFormat)
}

func runNSDNSList(_ *cobra.Command, _ []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcpNS(ctx)
	if err != nil {
		return err
	}
	var all []*networksecurity.DnsThreatDetector
	pageToken := ""
	for {
		c := svc.Projects.Locations.DnsThreatDetectors.List(parent).Context(ctx)
		if pageToken != "" {
			c = c.PageToken(pageToken)
		}
		resp, err := c.Do()
		if err != nil {
			return fmt.Errorf("listing DNS threat detectors: %w", err)
		}
		all = append(all, resp.DnsThreatDetectors...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagNSFormat != "" {
		return emitFormatted(all, flagNSFormat)
	}
	fmt.Printf("%-50s\n", "NAME")
	for _, d := range all {
		fmt.Printf("%-50s\n", nsBasename(d.Name))
	}
	return nil
}

func runNSDNSUpdate(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecurity.DnsThreatDetector{}
	if err := loadYAMLOrJSONInto(flagNSConfigFile, body); err != nil {
		return err
	}
	mask := nsResolveMask(body)
	ctx := context.Background()
	svc, err := gcpNS(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "dnsThreatDetectors", args[0])
	res, err := svc.Projects.Locations.DnsThreatDetectors.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating DNS threat detector: %w", err)
	}
	return emitFormatted(res, "")
}

func registerNSServerTlsPolicies(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.ServerTlsPolicy]{
		group: "server-tls-policies", singular: "server TLS policy", collection: "serverTlsPolicies",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.ServerTlsPolicy, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.ServerTlsPolicies.Create(parent, body).ServerTlsPolicyId(id).Context(ctx).Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.ServerTlsPolicies.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.ServerTlsPolicy, error) {
			return svc.Projects.Locations.ServerTlsPolicies.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.ServerTlsPolicy, string, error) {
			c := svc.Projects.Locations.ServerTlsPolicies.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.ServerTlsPolicies, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.ServerTlsPolicy, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.ServerTlsPolicies.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol: func(p *networksecurity.ServerTlsPolicy) string { return nsBasename(p.Name) },
	}
	crud.build(root, "server-tls-policies", "Manage server TLS policies", addNSLocationFlag)
}

func registerNSTlsInspectionPolicies(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.TlsInspectionPolicy]{
		group: "tls-inspection-policies", singular: "TLS inspection policy", collection: "tlsInspectionPolicies",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.TlsInspectionPolicy, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.TlsInspectionPolicies.Create(parent, body).TlsInspectionPolicyId(id).Context(ctx).Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.TlsInspectionPolicies.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.TlsInspectionPolicy, error) {
			return svc.Projects.Locations.TlsInspectionPolicies.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.TlsInspectionPolicy, string, error) {
			c := svc.Projects.Locations.TlsInspectionPolicies.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.TlsInspectionPolicies, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.TlsInspectionPolicy, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.TlsInspectionPolicies.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol: func(p *networksecurity.TlsInspectionPolicy) string { return nsBasename(p.Name) },
	}
	crud.build(root, "tls-inspection-policies", "Manage TLS inspection policies", addNSLocationFlag)
}

func registerNSUrlLists(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.UrlList]{
		group: "url-lists", singular: "URL list", collection: "urlLists",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.UrlList, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.UrlLists.Create(parent, body).UrlListId(id).Context(ctx).Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.UrlLists.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.UrlList, error) {
			return svc.Projects.Locations.UrlLists.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.UrlList, string, error) {
			c := svc.Projects.Locations.UrlLists.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.UrlLists, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.UrlList, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.UrlLists.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol: func(p *networksecurity.UrlList) string { return nsBasename(p.Name) },
	}
	crud.build(root, "url-lists", "Manage URL lists", addNSLocationFlag)
}

func registerNSGatewaySecurityPolicies(root *cobra.Command) {
	crud := &nsCRUD[networksecurity.GatewaySecurityPolicy]{
		group: "gateway-security-policies", singular: "gateway security policy", collection: "gatewaySecurityPolicies",
		parentFn: nsProjectParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.GatewaySecurityPolicy, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.GatewaySecurityPolicies.Create(parent, body).GatewaySecurityPolicyId(id).Context(ctx).Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.GatewaySecurityPolicies.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.GatewaySecurityPolicy, error) {
			return svc.Projects.Locations.GatewaySecurityPolicies.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.GatewaySecurityPolicy, string, error) {
			c := svc.Projects.Locations.GatewaySecurityPolicies.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.GatewaySecurityPolicies, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.GatewaySecurityPolicy, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.GatewaySecurityPolicies.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol: func(p *networksecurity.GatewaySecurityPolicy) string { return nsBasename(p.Name) },
	}
	group := crud.build(root, "gateway-security-policies", "Manage gateway security policies", addNSLocationFlag)

	// Rules subgroup lives under a gateway policy: names must include the policy
	// ID as a prefix, so parent = projects/PROJECT/locations/LOC/gatewaySecurityPolicies/POLICY.
	rulesCRUD := &nsCRUD[networksecurity.GatewaySecurityPolicyRule]{
		group: "rules", singular: "gateway security policy rule", collection: "rules",
		parentFn: nsGatewayPolicyParent,
		createFn: func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *networksecurity.GatewaySecurityPolicyRule, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.GatewaySecurityPolicies.Rules.Create(parent, body).GatewaySecurityPolicyRuleId(id).Context(ctx).Do()
		},
		deleteFn: func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.GatewaySecurityPolicies.Rules.Delete(name).Context(ctx).Do()
		},
		getFn: func(ctx context.Context, svc *networksecurity.Service, name string) (*networksecurity.GatewaySecurityPolicyRule, error) {
			return svc.Projects.Locations.GatewaySecurityPolicies.Rules.Get(name).Context(ctx).Do()
		},
		listFn: func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*networksecurity.GatewaySecurityPolicyRule, string, error) {
			c := svc.Projects.Locations.GatewaySecurityPolicies.Rules.List(parent).Context(ctx)
			if pageToken != "" {
				c = c.PageToken(pageToken)
			}
			r, err := c.Do()
			if err != nil {
				return nil, "", err
			}
			return r.GatewaySecurityPolicyRules, r.NextPageToken, nil
		},
		patchFn: func(ctx context.Context, svc *networksecurity.Service, name string, body *networksecurity.GatewaySecurityPolicyRule, mask, requestID string) (*networksecurity.Operation, error) {
			return svc.Projects.Locations.GatewaySecurityPolicies.Rules.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
		},
		nameCol:         func(r *networksecurity.GatewaySecurityPolicyRule) string { return nsBasename(r.Name) },
		secondaryCol:    func(r *networksecurity.GatewaySecurityPolicyRule) string { return r.BasicProfile },
		secondaryHeader: "PROFILE",
	}
	rulesGroup := rulesCRUD.build(group, "rules", "Manage gateway security policy rules", addNSLocationFlag)
	// Add --policy on every rules subcommand.
	for _, c := range rulesGroup.Commands() {
		c.Flags().StringVar(&flagNSGatewayPolicy, "policy", "", "Parent gateway security policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
}

// flagNSGatewayPolicy identifies the parent gateway security policy for rules
// subcommands.
var flagNSGatewayPolicy string

func nsGatewayPolicyParent() (string, error) {
	parent, err := nsProjectParent()
	if err != nil {
		return "", err
	}
	if flagNSGatewayPolicy == "" {
		return "", errFlagRequired("policy")
	}
	return parent + "/gatewaySecurityPolicies/" + flagNSGatewayPolicy, nil
}

func errFlagRequired(name string) error {
	return fmt.Errorf("--%s is required", name)
}
