package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	agentidentity "google.golang.org/api/agentidentity/v1"
)

// --- gcloud agent-identity (#1749) ---
//
// Agent Identity surface. Ports the auth-providers and access-summaries
// promotions/additions shipped in gcloud-python 574.0.0–579.0.0:
//   - 574.0.0: `agent-identity auth-providers` IAM subcommands
//     (get/set/add-binding/remove-binding/test-permissions)
//   - 577.0.0: promoted `agent-identity auth-providers` and
//     `agent-identity access-summaries` to GA
//   - 579.0.0: added `--three-legged-oauth-default-continue-uri` to
//     `agent-identity auth-providers create` and `update`
//
// gcloud-go tracks GA. Backed by the agentidentity v1 SDK.

// --- shared flags ---

// --- agent-identity auth-providers ---

// --- agent-identity access-summaries ---

func (g *agentIdentityCmds) initAgentIdentity() {
	// Every command in this group is location-scoped.
	locCmds := []*cobra.Command{
		g.aiAPCreateCmd, g.aiAPDeleteCmd, g.aiAPDescribeCmd, g.aiAPListCmd, g.aiAPUpdateCmd,
		g.aiAPEnableCmd, g.aiAPDisableCmd, g.aiAPUndeleteCmd,
		g.aiAPGetIamCmd, g.aiAPSetIamCmd, g.aiAPAddIamCmd, g.aiAPRemoveIamCmd, g.aiAPTestIamCmd,
		g.aiASDescribeCmd, g.aiASListCmd,
	}
	for _, c := range locCmds {
		c.Flags().StringVar(&g.flagAILocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagAIFormat, "format", "", "Output format")
	}

	// list flags
	g.aiAPListCmd.Flags().StringVar(&g.flagAIFilter, "filter", "", "Server-side list filter")
	g.aiASListCmd.Flags().StringVar(&g.flagAIFilter, "filter", "", "Server-side list filter")

	// create/update take a --config-file for the AuthProvider body plus the
	// new (579.0.0) --three-legged-oauth-default-continue-uri override.
	for _, c := range []*cobra.Command{g.aiAPCreateCmd, g.aiAPUpdateCmd} {
		c.Flags().StringVar(&g.flagAIConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AuthProvider body")
		c.Flags().StringVar(&g.flagAIThreeLeggedOAuthDefContinueURI,
			"three-legged-oauth-default-continue-uri", "",
			"Default continue URI for the ThreeLeggedOAuth flow")
	}
	g.aiAPUpdateCmd.Flags().StringVar(&g.flagAIUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	// add/remove IAM binding flags.
	for _, c := range []*cobra.Command{g.aiAPAddIamCmd, g.aiAPRemoveIamCmd} {
		c.Flags().StringVar(&g.flagAIIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&g.flagAIIamRole, "role", "", "IAM role to bind (required)")
		c.Flags().StringVar(&g.flagAIIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&g.flagAIIamCondTitle, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&g.flagAIIamCondDesc, "condition-description", "", "Description for a conditional binding")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	g.aiAPRemoveIamCmd.Flags().BoolVar(&g.flagAIIamAllConds, "all", false,
		"Remove matching bindings across all conditions, not just the given one")

	// test-iam-permissions
	g.aiAPTestIamCmd.Flags().StringSliceVar(&g.flagAIIamPermissions, "permissions", nil,
		"Comma-separated list of permissions to test (required)")
	_ = g.aiAPTestIamCmd.MarkFlagRequired("permissions")

	g.agentIdentityAuthProvidersCmd.AddCommand(
		g.aiAPCreateCmd, g.aiAPDeleteCmd, g.aiAPDescribeCmd, g.aiAPListCmd, g.aiAPUpdateCmd,
		g.aiAPEnableCmd, g.aiAPDisableCmd, g.aiAPUndeleteCmd,
		g.aiAPGetIamCmd, g.aiAPSetIamCmd, g.aiAPAddIamCmd, g.aiAPRemoveIamCmd, g.aiAPTestIamCmd,
	)
	g.agentIdentityAccessSummariesCmd.AddCommand(g.aiASDescribeCmd, g.aiASListCmd)
	g.agentIdentityCmd.AddCommand(g.agentIdentityAuthProvidersCmd, g.agentIdentityAccessSummariesCmd)

}

// --- helpers ---

func (g *agentIdentityCmds) aiParentLocation() (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	if g.flagAILocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, g.flagAILocation), nil
}

func (g *agentIdentityCmds) aiAuthProviderName(id string) (string, error) {
	parent, err := g.aiParentLocation()
	if err != nil {
		return "", err
	}
	return parent + "/authProviders/" + id, nil
}

func (g *agentIdentityCmds) aiAccessSummaryName(id string) (string, error) {
	parent, err := g.aiParentLocation()
	if err != nil {
		return "", err
	}
	return parent + "/accessSummaries/" + id, nil
}

// aiIamBuildCondition returns an agentidentity.Expr or nil if no condition
// fields are set.
func (g *agentIdentityCmds) aiIamBuildCondition() *agentidentity.Expr {
	if g.flagAIIamCondExpr == "" && g.flagAIIamCondTitle == "" && g.flagAIIamCondDesc == "" {
		return nil
	}
	return &agentidentity.Expr{
		Expression:  g.flagAIIamCondExpr,
		Title:       g.flagAIIamCondTitle,
		Description: g.flagAIIamCondDesc,
	}
}

func aiIamCondsEqual(a, b *agentidentity.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func aiIamAddBinding(pol *agentidentity.Policy, role, member string, cond *agentidentity.Expr) {
	for _, b := range pol.Bindings {
		if b.Role != role || !aiIamCondsEqual(b.Condition, cond) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	pol.Bindings = append(pol.Bindings, &agentidentity.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func aiIamRemoveBinding(pol *agentidentity.Policy, role, member string, cond *agentidentity.Expr, allConds bool) bool {
	changed := false
	kept := pol.Bindings[:0]
	for _, b := range pol.Bindings {
		if b.Role != role || (!allConds && !aiIamCondsEqual(b.Condition, cond)) {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				continue
			}
			newMembers = append(newMembers, m)
		}
		if len(newMembers) != len(b.Members) {
			changed = true
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	pol.Bindings = kept
	return changed
}

// aiApplyThreeLeggedContinueURI stamps the given URI onto the AuthProvider
// body's AuthProviderTypeParams.ThreeLeggedOauth.DefaultContinueUri, creating
// the nested structs as needed.
func aiApplyThreeLeggedContinueURI(ap *agentidentity.AuthProvider, uri string) {
	if uri == "" {
		return
	}
	if ap.AuthProviderTypeParams == nil {
		ap.AuthProviderTypeParams = &agentidentity.AuthProviderTypeParams{}
	}
	if ap.AuthProviderTypeParams.ThreeLeggedOauth == nil {
		ap.AuthProviderTypeParams.ThreeLeggedOauth = &agentidentity.ThreeLeggedOAuth{}
	}
	ap.AuthProviderTypeParams.ThreeLeggedOauth.DefaultContinueUri = uri
}

// --- auth-providers implementations ---

func (g *agentIdentityCmds) runAIAPCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.aiParentLocation()
	if err != nil {
		return err
	}
	ap := &agentidentity.AuthProvider{}
	if g.flagAIConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagAIConfigFile, ap); err != nil {
			return err
		}
	}
	aiApplyThreeLeggedContinueURI(ap, g.flagAIThreeLeggedOAuthDefContinueURI)
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Create(parent, ap).
		AuthProviderId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating auth provider: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Created auth provider [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPDelete(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.AuthProviders.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting auth provider: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Deleted auth provider [%s].\n", args[0])
	return nil
}

func (g *agentIdentityCmds) runAIAPDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing auth provider: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPList(cmd *cobra.Command, args []string) error {
	parent, err := g.aiParentLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	var all []*agentidentity.AuthProvider
	pageToken := ""
	for {
		call := svc.Projects.Locations.AuthProviders.List(parent).Context(ctx)
		if g.flagAIFilter != "" {
			call = call.Filter(g.flagAIFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing auth providers: %w", err)
		}
		all = append(all, resp.AuthProviders...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagAIFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAIFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "STATE")
	for _, p := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

func (g *agentIdentityCmds) runAIAPUpdate(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ap := &agentidentity.AuthProvider{}
	if g.flagAIConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagAIConfigFile, ap); err != nil {
			return err
		}
	}
	aiApplyThreeLeggedContinueURI(ap, g.flagAIThreeLeggedOAuthDefContinueURI)
	mask := g.flagAIUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ap))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.AuthProviders.Patch(name, ap).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating auth provider: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Updated auth provider [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPEnable(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Enable(name,
		&agentidentity.EnableAuthProviderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling auth provider: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Enabled auth provider [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPDisable(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Disable(name,
		&agentidentity.DisableAuthProviderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling auth provider: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Disabled auth provider [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPUndelete(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Undelete(name,
		&agentidentity.UndeleteAuthProviderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting auth provider: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Undeleted auth provider [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagAIFormat)
}

// --- IAM ---

func (g *agentIdentityCmds) runAIAPGetIam(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.AuthProviders.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, pol, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPSetIam(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	pol := &agentidentity.Policy{}
	if err := loadYAMLOrJSONInto(args[1], pol); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	out, err := svc.Projects.Locations.AuthProviders.SetIamPolicy(name,
		&agentidentity.SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Updated IAM policy for %s.\n", name)
	return emitFormattedTo(g.inv.stdout, out, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPAddIam(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.AuthProviders.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := g.aiIamBuildCondition()
	aiIamAddBinding(pol, g.flagAIIamRole, g.flagAIIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	out, err := svc.Projects.Locations.AuthProviders.SetIamPolicy(name,
		&agentidentity.SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Updated IAM policy for %s.\n", name)
	return emitFormattedTo(g.inv.stdout, out, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.AuthProviders.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := g.aiIamBuildCondition()
	if !aiIamRemoveBinding(pol, g.flagAIIamRole, g.flagAIIamMember, cond, g.flagAIIamAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	out, err := svc.Projects.Locations.AuthProviders.SetIamPolicy(name,
		&agentidentity.SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Updated IAM policy for %s.\n", name)
	return emitFormattedTo(g.inv.stdout, out, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIAPTestIam(cmd *cobra.Command, args []string) error {
	name, err := g.aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.AuthProviders.TestIamPermissions(name,
		&agentidentity.TestIamPermissionsRequest{Permissions: g.flagAIIamPermissions}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("testing IAM permissions: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagAIFormat)
}

// --- access-summaries ---

func (g *agentIdentityCmds) runAIASDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.aiAccessSummaryName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AccessSummaries.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing access summary: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAIFormat)
}

func (g *agentIdentityCmds) runAIASList(cmd *cobra.Command, args []string) error {
	parent, err := g.aiParentLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentIdentity(ctx)
	if err != nil {
		return err
	}
	var all []*agentidentity.AccessSummary
	pageToken := ""
	for {
		call := svc.Projects.Locations.AccessSummaries.List(parent).Context(ctx)
		if g.flagAIFilter != "" {
			call = call.Filter(g.flagAIFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing access summaries: %w", err)
		}
		all = append(all, resp.AccessSummaries...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagAIFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAIFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s\n", "NAME")
	for _, s := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s\n", path.Base(s.Name))
	}
	return nil
}

// agentIdentityCmds holds the commands and flag values of one invocation's "agentIdentity"
// command group, so invocations don't share state (#2070).
type agentIdentityCmds struct {
	inv *invocation

	agentIdentityCmd                     *cobra.Command
	flagAILocation                       string
	flagAIFormat                         string
	flagAIFilter                         string
	flagAIConfigFile                     string
	flagAIUpdateMask                     string
	flagAIThreeLeggedOAuthDefContinueURI string
	flagAIIamMember                      string
	flagAIIamRole                        string
	flagAIIamCondExpr                    string
	flagAIIamCondTitle                   string
	flagAIIamCondDesc                    string
	flagAIIamPermissions                 []string
	flagAIIamAllConds                    bool
	agentIdentityAuthProvidersCmd        *cobra.Command
	aiAPCreateCmd                        *cobra.Command
	aiAPDeleteCmd                        *cobra.Command
	aiAPDescribeCmd                      *cobra.Command
	aiAPListCmd                          *cobra.Command
	aiAPUpdateCmd                        *cobra.Command
	aiAPEnableCmd                        *cobra.Command
	aiAPDisableCmd                       *cobra.Command
	aiAPUndeleteCmd                      *cobra.Command
	aiAPGetIamCmd                        *cobra.Command
	aiAPSetIamCmd                        *cobra.Command
	aiAPAddIamCmd                        *cobra.Command
	aiAPRemoveIamCmd                     *cobra.Command
	aiAPTestIamCmd                       *cobra.Command
	agentIdentityAccessSummariesCmd      *cobra.Command
	aiASDescribeCmd                      *cobra.Command
	aiASListCmd                          *cobra.Command
	flagAIAuthzLocation                  string
	flagAIAuthzAuthProvider              string
	flagAIAuthzEndpointMode              string
	flagAIAuthzMember                    string
	flagAIAuthzRole                      string
	flagAIAuthzPermissions               []string
	flagAIAuthzLimit                     int64
	flagAIAuthzPageSize                  int64
	flagAIAuthzSortBy                    []string
	agentIdentityAuthorizationsCmd       *cobra.Command
	aiAuthzGetIamCmd                     *cobra.Command
	aiAuthzSetIamCmd                     *cobra.Command
	aiAuthzAddIamCmd                     *cobra.Command
	aiAuthzRemoveIamCmd                  *cobra.Command
	aiAuthzTestIamCmd                    *cobra.Command
}

// newAgentIdentityCmd returns the "agentIdentity" command group for inv.
func newAgentIdentityCmd(inv *invocation) *cobra.Command {
	return newAgentIdentityCmds(inv).agentIdentityCmd
}

// newAgentIdentityCmds builds the "agentIdentity" command group for inv and returns its state.
func newAgentIdentityCmds(inv *invocation) *agentIdentityCmds {
	g := &agentIdentityCmds{inv: inv}
	g.agentIdentityCmd = &cobra.Command{
		Use:   "agent-identity",
		Short: "Manage Agent Identity resources",
	}
	g.agentIdentityAuthProvidersCmd = &cobra.Command{
		Use:   "auth-providers",
		Short: "Manage Agent Identity auth providers",
	}
	g.aiAPCreateCmd = &cobra.Command{
		Use: "create AUTH_PROVIDER", Short: "Create an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPCreate,
	}
	g.aiAPDeleteCmd = &cobra.Command{
		Use: "delete AUTH_PROVIDER", Short: "Delete an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPDelete,
	}
	g.aiAPDescribeCmd = &cobra.Command{
		Use: "describe AUTH_PROVIDER", Short: "Describe an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPDescribe,
	}
	g.aiAPListCmd = &cobra.Command{
		Use: "list", Short: "List auth providers in a location",
		Args: cobra.NoArgs, RunE: g.runAIAPList,
	}
	g.aiAPUpdateCmd = &cobra.Command{
		Use: "update AUTH_PROVIDER", Short: "Update an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPUpdate,
	}
	g.aiAPEnableCmd = &cobra.Command{
		Use: "enable AUTH_PROVIDER", Short: "Enable an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPEnable,
	}
	g.aiAPDisableCmd = &cobra.Command{
		Use: "disable AUTH_PROVIDER", Short: "Disable an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPDisable,
	}
	g.aiAPUndeleteCmd = &cobra.Command{
		Use: "undelete AUTH_PROVIDER", Short: "Undelete an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPUndelete,
	}
	g.aiAPGetIamCmd = &cobra.Command{
		Use: "get-iam-policy AUTH_PROVIDER", Short: "Get IAM policy for an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPGetIam,
	}
	g.aiAPSetIamCmd = &cobra.Command{
		Use: "set-iam-policy AUTH_PROVIDER POLICY_FILE", Short: "Set IAM policy on an auth provider",
		Args: cobra.ExactArgs(2), RunE: g.runAIAPSetIam,
	}
	g.aiAPAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding AUTH_PROVIDER", Short: "Add an IAM policy binding on an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPAddIam,
	}
	g.aiAPRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding AUTH_PROVIDER", Short: "Remove an IAM policy binding from an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPRemoveIam,
	}
	g.aiAPTestIamCmd = &cobra.Command{
		Use: "test-iam-permissions AUTH_PROVIDER", Short: "Test IAM permissions on an auth provider",
		Args: cobra.ExactArgs(1), RunE: g.runAIAPTestIam,
	}
	g.agentIdentityAccessSummariesCmd = &cobra.Command{
		Use:   "access-summaries",
		Short: "View Agent Identity access summaries",
	}
	g.aiASDescribeCmd = &cobra.Command{
		Use: "describe ACCESS_SUMMARY", Short: "Describe an access summary",
		Args: cobra.ExactArgs(1), RunE: g.runAIASDescribe,
	}
	g.aiASListCmd = &cobra.Command{
		Use: "list", Short: "List access summaries in a location",
		Args: cobra.NoArgs, RunE: g.runAIASList,
	}
	g.agentIdentityAuthorizationsCmd = &cobra.Command{
		Use:   "authorizations",
		Short: "Manage Authorization resources",
	}
	g.aiAuthzGetIamCmd = &cobra.Command{
		Use: "get-iam-policy AUTHORIZATION", Short: "Get the IAM policy for an Agent Identity authorization",
		Args: cobra.ExactArgs(1), RunE: g.runAIAuthzGetIam,
	}
	g.aiAuthzSetIamCmd = &cobra.Command{
		Use: "set-iam-policy AUTHORIZATION POLICY_FILE", Short: "Set the IAM policy for an Agent Identity authorization",
		Args: cobra.ExactArgs(2), RunE: g.runAIAuthzSetIam,
	}
	g.aiAuthzAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding AUTHORIZATION", Short: "Add IAM policy binding to an Agent Identity authorization",
		Args: cobra.ExactArgs(1), RunE: g.runAIAuthzAddIam,
	}
	g.aiAuthzRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding AUTHORIZATION", Short: "Remove IAM policy binding from an Agent Identity authorization",
		Args: cobra.ExactArgs(1), RunE: g.runAIAuthzRemoveIam,
	}
	g.aiAuthzTestIamCmd = &cobra.Command{
		Use: "test-iam-permissions AUTHORIZATION", Short: "Test IAM permissions for an Agent Identity authorization",
		Args: cobra.ExactArgs(1), RunE: g.runAIAuthzTestIam,
	}
	g.initAgentIdentity()
	g.initAgentIdentityAuthorizations()
	return g
}
