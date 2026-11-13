package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	agentidentityalpha "google.golang.org/api/agentidentity/v1alpha"
	"google.golang.org/api/googleapi"
	"gopkg.in/yaml.v3"
)

// --- gcloud agent-identity auth-providers authorizations (#1900) ---
//
// Ports the 588.0.0 addition of the authorization IAM commands
// (get-iam-policy, set-iam-policy, add-iam-policy-binding,
// remove-iam-policy-binding, test-iam-permissions). gcloud-python ships these
// on the alpha/beta tracks only, backed by the agentidentity v1alpha API; v1
// does not expose IAM methods on authorizations, so gcloud-go uses v1alpha.
//
// Output goes through the global --format/--filter/--flatten pipeline. The
// IAM spec in the reference YAML sets enable_condition: false, so the binding
// commands take no condition flags.

const (
	aiEndpointModeGlobal            = "global"
	aiEndpointModeRegional          = "regional"
	aiEndpointModeRegionalPreferred = "regional-preferred"
)

func (g *agentIdentityCmds) initAgentIdentityAuthorizations() {
	all := []*cobra.Command{g.aiAuthzGetIamCmd, g.aiAuthzSetIamCmd, g.aiAuthzAddIamCmd, g.aiAuthzRemoveIamCmd, g.aiAuthzTestIamCmd}
	for _, c := range all {
		// --location and --auth-provider are resource-argument fallthroughs:
		// they are only needed when AUTHORIZATION is not a full resource name.
		c.Flags().StringVar(&g.flagAIAuthzLocation, "location", "", "The location id of the authorization resource")
		c.Flags().StringVar(&g.flagAIAuthzAuthProvider, "auth-provider", "", "The authProvider id of the authorization resource")
		c.Flags().StringVar(&g.flagAIAuthzEndpointMode, "endpoint-mode", "",
			"Endpoint mode: global (default), regional, or regional-preferred")
	}

	// get-iam-policy is a ListCommand in gcloud-python. The policy is a single
	// resource, so --limit/--page-size/--sort-by are accepted for parity but
	// do not change the output.
	g.aiAuthzGetIamCmd.Flags().Int64Var(&g.flagAIAuthzLimit, "limit", 0, "Maximum number of resources to list")
	g.aiAuthzGetIamCmd.Flags().Int64Var(&g.flagAIAuthzPageSize, "page-size", 0, "Number of resources per page")
	g.aiAuthzGetIamCmd.Flags().StringSliceVar(&g.flagAIAuthzSortBy, "sort-by", nil, "Comma-separated list of resource field keys to sort by")

	for _, c := range []*cobra.Command{g.aiAuthzAddIamCmd, g.aiAuthzRemoveIamCmd} {
		c.Flags().StringVar(&g.flagAIAuthzMember, "member", "", "The principal to add/remove the binding for (required)")
		c.Flags().StringVar(&g.flagAIAuthzRole, "role", "", "The role to add/remove the principal from (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}

	g.aiAuthzTestIamCmd.Flags().StringSliceVar(&g.flagAIAuthzPermissions, "permissions", nil, "The permissions to test (required)")
	_ = g.aiAuthzTestIamCmd.MarkFlagRequired("permissions")

	g.agentIdentityAuthorizationsCmd.AddCommand(all...)
	g.agentIdentityAuthProvidersCmd.AddCommand(g.agentIdentityAuthorizationsCmd)
}

// --- helpers ---

// aiAuthorizationName resolves the AUTHORIZATION positional into a full
// resource name. A full `projects/.../authorizations/ID` name is used as-is;
// otherwise --location and --auth-provider (and the active project) fill in
// the parent.
func (g *agentIdentityCmds) aiAuthorizationName(arg string) (string, error) {
	if strings.HasPrefix(arg, "projects/") {
		parts := strings.Split(arg, "/")
		if len(parts) != 8 || parts[2] != "locations" || parts[4] != "authProviders" || parts[6] != "authorizations" {
			return "", fmt.Errorf("invalid authorization resource name [%s]: expected projects/PROJECT/locations/LOCATION/authProviders/AUTH_PROVIDER/authorizations/AUTHORIZATION", arg)
		}
		return arg, nil
	}
	if g.flagAIAuthzLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	if g.flagAIAuthzAuthProvider == "" {
		return "", fmt.Errorf("--auth-provider is required")
	}
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/authProviders/%s/authorizations/%s",
		project, g.flagAIAuthzLocation, g.flagAIAuthzAuthProvider, arg), nil
}

// aiAuthzEndpoint returns the base URL override for the given endpoint mode
// and location, or "" for the global endpoint. It mirrors gcloud-python's
// _ShouldUseRegionalEndpoints for a command annotated
// regional_endpoint_compatibility: SUPPORTED. The 589.0.0 API map lists no
// known regional endpoints for agentidentity, so regional-preferred always
// resolves to global, while an explicit regional request is let through to
// the {service}.{region}.rep.googleapis.com form.
func aiAuthzEndpoint(mode, location string) (string, error) {
	switch mode {
	case "", aiEndpointModeGlobal, aiEndpointModeRegionalPreferred:
		return "", nil
	case aiEndpointModeRegional:
		return fmt.Sprintf("https://agentidentity.%s.rep.googleapis.com/", location), nil
	default:
		return "", fmt.Errorf("argument --endpoint-mode: Invalid choice: '%s'. Valid choices are [%s, %s, %s]",
			mode, aiEndpointModeGlobal, aiEndpointModeRegional, aiEndpointModeRegionalPreferred)
	}
}

// aiAuthzService resolves the authorization name and returns a v1alpha client
// pointed at the endpoint selected by --endpoint-mode.
func (g *agentIdentityCmds) aiAuthzService(ctx context.Context, arg string) (*agentidentityalpha.Service, string, error) {
	name, err := g.aiAuthorizationName(arg)
	if err != nil {
		return nil, "", err
	}
	endpoint, err := aiAuthzEndpoint(g.flagAIAuthzEndpointMode, strings.Split(name, "/")[3])
	if err != nil {
		return nil, "", err
	}
	svc, err := g.inv.clients().AgentIdentityAlpha(ctx, endpoint)
	if err != nil {
		return nil, "", err
	}
	return svc, name, nil
}

// aiAuthzAddBinding mirrors iam_util.AddBindingToIamPolicy: the member is
// appended to the first binding for role, or a new binding is created. It is
// a no-op when the member already holds the role.
func aiAuthzAddBinding(pol *agentidentityalpha.Policy, role, member string) {
	for _, b := range pol.Bindings {
		if b.Role != role {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
	}
	for _, b := range pol.Bindings {
		if b.Role == role {
			b.Members = append(b.Members, member)
			return
		}
	}
	pol.Bindings = append(pol.Bindings, &agentidentityalpha.Binding{Role: role, Members: []string{member}})
}

// aiAuthzRemoveBinding mirrors iam_util.RemoveBindingFromIamPolicy: the
// member is removed from the first binding for role that contains it, then
// empty bindings are dropped.
func aiAuthzRemoveBinding(pol *agentidentityalpha.Policy, role, member string) error {
	found := false
	for _, b := range pol.Bindings {
		if b.Role != role {
			continue
		}
		for i, m := range b.Members {
			if m == member {
				b.Members = append(b.Members[:i], b.Members[i+1:]...)
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return errors.New("Policy binding with the specified principal and role not found!")
	}
	kept := pol.Bindings[:0]
	for _, b := range pol.Bindings {
		if len(b.Members) > 0 {
			kept = append(kept, b)
		}
	}
	pol.Bindings = kept
	return nil
}

// aiAuthzPolicyUpdateMask mirrors iam_util.ConstructUpdateMaskFromPolicy: the
// sorted, comma-joined top-level keys of the policy file.
func aiAuthzPolicyUpdateMask(policyFile string) (string, error) {
	data, err := os.ReadFile(policyFile)
	if err != nil {
		return "", fmt.Errorf("reading policy file: %w", err)
	}
	var top map[string]any
	if err := yaml.Unmarshal(data, &top); err != nil {
		return "", fmt.Errorf("parsing policy file: %w", err)
	}
	keys := make([]string, 0, len(top))
	for k := range top {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ","), nil
}

// aiAuthzLintHint prints gcloud-python's lint-condition hint when a policy
// write is rejected with HTTP 400.
func (g *agentIdentityCmds) aiAuthzLintHint(err error, binding bool) {
	var gerr *googleapi.Error
	if !errors.As(err, &gerr) || gerr.Code != http.StatusBadRequest {
		return
	}
	if binding {
		fmt.Fprintln(g.inv.stderr, `ERROR: Policy modification failed. For a binding with condition, run "gcloud alpha iam policies lint-condition" to identify issues in condition.`)
		return
	}
	fmt.Fprintln(g.inv.stderr, `ERROR: Policy modification failed. For bindings with conditions, run "gcloud alpha iam policies lint-condition" to identify issues in conditions.`)
}

func (g *agentIdentityCmds) aiAuthzLogSetIamPolicy(name string) {
	fmt.Fprintf(g.inv.stderr, "Updated IAM policy for authorization [%s].\n", path.Base(name))
}

// --- implementations ---

func (g *agentIdentityCmds) runAIAuthzGetIam(cmd *cobra.Command, args []string) error {
	for _, f := range []string{"limit", "page-size"} {
		if cmd.Flags().Changed(f) {
			if v, _ := cmd.Flags().GetInt64(f); v < 1 {
				return fmt.Errorf("argument --%s: Value must be greater than or equal to 1; received: %d", f, v)
			}
		}
	}
	ctx := cmd.Context()
	svc, name, err := g.aiAuthzService(ctx, args[0])
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.AuthProviders.Authorizations.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return g.inv.emitWithPipeline(pol)
}

func (g *agentIdentityCmds) runAIAuthzSetIam(cmd *cobra.Command, args []string) error {
	pol := &agentidentityalpha.Policy{}
	if err := loadYAMLOrJSONInto(args[1], pol); err != nil {
		return err
	}
	mask, err := aiAuthzPolicyUpdateMask(args[1])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, name, err := g.aiAuthzService(ctx, args[0])
	if err != nil {
		return err
	}
	out, err := svc.Projects.Locations.AuthProviders.Authorizations.SetIamPolicy(name,
		&agentidentityalpha.SetIamPolicyRequest{Policy: pol, UpdateMask: mask}).Context(ctx).Do()
	if err != nil {
		g.aiAuthzLintHint(err, false)
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	g.aiAuthzLogSetIamPolicy(name)
	return emitFormattedTo(g.inv.stdout, out, g.inv.flags.format)
}

func (g *agentIdentityCmds) runAIAuthzAddIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, name, err := g.aiAuthzService(ctx, args[0])
	if err != nil {
		return err
	}
	authz := svc.Projects.Locations.AuthProviders.Authorizations
	pol, err := authz.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	aiAuthzAddBinding(pol, g.flagAIAuthzRole, g.flagAIAuthzMember)
	out, err := authz.SetIamPolicy(name, &agentidentityalpha.SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		g.aiAuthzLintHint(err, true)
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	g.aiAuthzLogSetIamPolicy(name)
	return emitFormattedTo(g.inv.stdout, out, g.inv.flags.format)
}

func (g *agentIdentityCmds) runAIAuthzRemoveIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, name, err := g.aiAuthzService(ctx, args[0])
	if err != nil {
		return err
	}
	authz := svc.Projects.Locations.AuthProviders.Authorizations
	pol, err := authz.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if err := aiAuthzRemoveBinding(pol, g.flagAIAuthzRole, g.flagAIAuthzMember); err != nil {
		return err
	}
	out, err := authz.SetIamPolicy(name, &agentidentityalpha.SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	g.aiAuthzLogSetIamPolicy(name)
	return emitFormattedTo(g.inv.stdout, out, g.inv.flags.format)
}

func (g *agentIdentityCmds) runAIAuthzTestIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, name, err := g.aiAuthzService(ctx, args[0])
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.AuthProviders.Authorizations.TestIamPermissions(name,
		&agentidentityalpha.TestIamPermissionsRequest{Permissions: g.flagAIAuthzPermissions}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("testing IAM permissions: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.inv.flags.format)
}
