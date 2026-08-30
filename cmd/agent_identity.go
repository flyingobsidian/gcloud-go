package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
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

var agentIdentityCmd = &cobra.Command{
	Use:   "agent-identity",
	Short: "Manage Agent Identity resources",
}

// --- shared flags ---

var (
	flagAILocation                       string
	flagAIFormat                         string
	flagAIFilter                         string
	flagAIConfigFile                     string
	flagAIUpdateMask                     string
	flagAIAuthProviderID                 string
	flagAIThreeLeggedOAuthDefContinueURI string
	flagAIIamMember                      string
	flagAIIamRole                        string
	flagAIIamCondExpr                    string
	flagAIIamCondTitle                   string
	flagAIIamCondDesc                    string
	flagAIIamPermissions                 []string
	flagAIIamAllConds                    bool
)

// --- agent-identity auth-providers ---

var agentIdentityAuthProvidersCmd = &cobra.Command{
	Use:   "auth-providers",
	Short: "Manage Agent Identity auth providers",
}

var (
	aiAPCreateCmd = &cobra.Command{
		Use: "create AUTH_PROVIDER", Short: "Create an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPCreate,
	}
	aiAPDeleteCmd = &cobra.Command{
		Use: "delete AUTH_PROVIDER", Short: "Delete an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPDelete,
	}
	aiAPDescribeCmd = &cobra.Command{
		Use: "describe AUTH_PROVIDER", Short: "Describe an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPDescribe,
	}
	aiAPListCmd = &cobra.Command{
		Use: "list", Short: "List auth providers in a location",
		Args: cobra.NoArgs, RunE: runAIAPList,
	}
	aiAPUpdateCmd = &cobra.Command{
		Use: "update AUTH_PROVIDER", Short: "Update an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPUpdate,
	}
	aiAPEnableCmd = &cobra.Command{
		Use: "enable AUTH_PROVIDER", Short: "Enable an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPEnable,
	}
	aiAPDisableCmd = &cobra.Command{
		Use: "disable AUTH_PROVIDER", Short: "Disable an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPDisable,
	}
	aiAPUndeleteCmd = &cobra.Command{
		Use: "undelete AUTH_PROVIDER", Short: "Undelete an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPUndelete,
	}
	aiAPGetIamCmd = &cobra.Command{
		Use: "get-iam-policy AUTH_PROVIDER", Short: "Get IAM policy for an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPGetIam,
	}
	aiAPSetIamCmd = &cobra.Command{
		Use: "set-iam-policy AUTH_PROVIDER POLICY_FILE", Short: "Set IAM policy on an auth provider",
		Args: cobra.ExactArgs(2), RunE: runAIAPSetIam,
	}
	aiAPAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding AUTH_PROVIDER", Short: "Add an IAM policy binding on an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPAddIam,
	}
	aiAPRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding AUTH_PROVIDER", Short: "Remove an IAM policy binding from an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPRemoveIam,
	}
	aiAPTestIamCmd = &cobra.Command{
		Use: "test-iam-permissions AUTH_PROVIDER", Short: "Test IAM permissions on an auth provider",
		Args: cobra.ExactArgs(1), RunE: runAIAPTestIam,
	}
)

// --- agent-identity access-summaries ---

var agentIdentityAccessSummariesCmd = &cobra.Command{
	Use:   "access-summaries",
	Short: "View Agent Identity access summaries",
}

var (
	aiASDescribeCmd = &cobra.Command{
		Use: "describe ACCESS_SUMMARY", Short: "Describe an access summary",
		Args: cobra.ExactArgs(1), RunE: runAIASDescribe,
	}
	aiASListCmd = &cobra.Command{
		Use: "list", Short: "List access summaries in a location",
		Args: cobra.NoArgs, RunE: runAIASList,
	}
)

func init() {
	// Every command in this group is location-scoped.
	locCmds := []*cobra.Command{
		aiAPCreateCmd, aiAPDeleteCmd, aiAPDescribeCmd, aiAPListCmd, aiAPUpdateCmd,
		aiAPEnableCmd, aiAPDisableCmd, aiAPUndeleteCmd,
		aiAPGetIamCmd, aiAPSetIamCmd, aiAPAddIamCmd, aiAPRemoveIamCmd, aiAPTestIamCmd,
		aiASDescribeCmd, aiASListCmd,
	}
	for _, c := range locCmds {
		c.Flags().StringVar(&flagAILocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagAIFormat, "format", "", "Output format")
	}

	// list flags
	aiAPListCmd.Flags().StringVar(&flagAIFilter, "filter", "", "Server-side list filter")
	aiASListCmd.Flags().StringVar(&flagAIFilter, "filter", "", "Server-side list filter")

	// create/update take a --config-file for the AuthProvider body plus the
	// new (579.0.0) --three-legged-oauth-default-continue-uri override.
	for _, c := range []*cobra.Command{aiAPCreateCmd, aiAPUpdateCmd} {
		c.Flags().StringVar(&flagAIConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AuthProvider body")
		c.Flags().StringVar(&flagAIThreeLeggedOAuthDefContinueURI,
			"three-legged-oauth-default-continue-uri", "",
			"Default continue URI for the ThreeLeggedOAuth flow")
	}
	aiAPUpdateCmd.Flags().StringVar(&flagAIUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	// add/remove IAM binding flags.
	for _, c := range []*cobra.Command{aiAPAddIamCmd, aiAPRemoveIamCmd} {
		c.Flags().StringVar(&flagAIIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagAIIamRole, "role", "", "IAM role to bind (required)")
		c.Flags().StringVar(&flagAIIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&flagAIIamCondTitle, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&flagAIIamCondDesc, "condition-description", "", "Description for a conditional binding")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	aiAPRemoveIamCmd.Flags().BoolVar(&flagAIIamAllConds, "all", false,
		"Remove matching bindings across all conditions, not just the given one")

	// test-iam-permissions
	aiAPTestIamCmd.Flags().StringSliceVar(&flagAIIamPermissions, "permissions", nil,
		"Comma-separated list of permissions to test (required)")
	_ = aiAPTestIamCmd.MarkFlagRequired("permissions")

	agentIdentityAuthProvidersCmd.AddCommand(
		aiAPCreateCmd, aiAPDeleteCmd, aiAPDescribeCmd, aiAPListCmd, aiAPUpdateCmd,
		aiAPEnableCmd, aiAPDisableCmd, aiAPUndeleteCmd,
		aiAPGetIamCmd, aiAPSetIamCmd, aiAPAddIamCmd, aiAPRemoveIamCmd, aiAPTestIamCmd,
	)
	agentIdentityAccessSummariesCmd.AddCommand(aiASDescribeCmd, aiASListCmd)
	agentIdentityCmd.AddCommand(agentIdentityAuthProvidersCmd, agentIdentityAccessSummariesCmd)
	rootCmd.AddCommand(agentIdentityCmd)
}

// --- helpers ---

func aiParentLocation() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagAILocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagAILocation), nil
}

func aiAuthProviderName(id string) (string, error) {
	parent, err := aiParentLocation()
	if err != nil {
		return "", err
	}
	return parent + "/authProviders/" + id, nil
}

func aiAccessSummaryName(id string) (string, error) {
	parent, err := aiParentLocation()
	if err != nil {
		return "", err
	}
	return parent + "/accessSummaries/" + id, nil
}

// aiIamBuildCondition returns an agentidentity.Expr or nil if no condition
// fields are set.
func aiIamBuildCondition() *agentidentity.Expr {
	if flagAIIamCondExpr == "" && flagAIIamCondTitle == "" && flagAIIamCondDesc == "" {
		return nil
	}
	return &agentidentity.Expr{
		Expression:  flagAIIamCondExpr,
		Title:       flagAIIamCondTitle,
		Description: flagAIIamCondDesc,
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

func runAIAPCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiParentLocation()
	if err != nil {
		return err
	}
	ap := &agentidentity.AuthProvider{}
	if flagAIConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagAIConfigFile, ap); err != nil {
			return err
		}
	}
	aiApplyThreeLeggedContinueURI(ap, flagAIThreeLeggedOAuthDefContinueURI)
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Create(parent, ap).
		AuthProviderId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating auth provider: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created auth provider [%s].\n", args[0])
	return emitFormatted(got, flagAIFormat)
}

func runAIAPDelete(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.AuthProviders.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting auth provider: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deleted auth provider [%s].\n", args[0])
	return nil
}

func runAIAPDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing auth provider: %w", err)
	}
	return emitFormatted(got, flagAIFormat)
}

func runAIAPList(cmd *cobra.Command, args []string) error {
	parent, err := aiParentLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*agentidentity.AuthProvider
	pageToken := ""
	for {
		call := svc.Projects.Locations.AuthProviders.List(parent).Context(ctx)
		if flagAIFilter != "" {
			call = call.Filter(flagAIFilter)
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
	if flagAIFormat != "" {
		return emitFormatted(all, flagAIFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, p := range all {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

func runAIAPUpdate(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ap := &agentidentity.AuthProvider{}
	if flagAIConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagAIConfigFile, ap); err != nil {
			return err
		}
	}
	aiApplyThreeLeggedContinueURI(ap, flagAIThreeLeggedOAuthDefContinueURI)
	mask := flagAIUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ap))
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
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
	fmt.Fprintf(os.Stderr, "Updated auth provider [%s].\n", args[0])
	return emitFormatted(got, flagAIFormat)
}

func runAIAPEnable(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Enable(name,
		&agentidentity.EnableAuthProviderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling auth provider: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Enabled auth provider [%s].\n", args[0])
	return emitFormatted(got, flagAIFormat)
}

func runAIAPDisable(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Disable(name,
		&agentidentity.DisableAuthProviderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling auth provider: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Disabled auth provider [%s].\n", args[0])
	return emitFormatted(got, flagAIFormat)
}

func runAIAPUndelete(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthProviders.Undelete(name,
		&agentidentity.UndeleteAuthProviderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting auth provider: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Undeleted auth provider [%s].\n", args[0])
	return emitFormatted(got, flagAIFormat)
}

// --- IAM ---

func runAIAPGetIam(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.AuthProviders.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(pol, flagAIFormat)
}

func runAIAPSetIam(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	pol := &agentidentity.Policy{}
	if err := loadYAMLOrJSONInto(args[1], pol); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	out, err := svc.Projects.Locations.AuthProviders.SetIamPolicy(name,
		&agentidentity.SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", name)
	return emitFormatted(out, flagAIFormat)
}

func runAIAPAddIam(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.AuthProviders.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := aiIamBuildCondition()
	aiIamAddBinding(pol, flagAIIamRole, flagAIIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	out, err := svc.Projects.Locations.AuthProviders.SetIamPolicy(name,
		&agentidentity.SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", name)
	return emitFormatted(out, flagAIFormat)
}

func runAIAPRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.AuthProviders.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := aiIamBuildCondition()
	if !aiIamRemoveBinding(pol, flagAIIamRole, flagAIIamMember, cond, flagAIIamAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	out, err := svc.Projects.Locations.AuthProviders.SetIamPolicy(name,
		&agentidentity.SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", name)
	return emitFormatted(out, flagAIFormat)
}

func runAIAPTestIam(cmd *cobra.Command, args []string) error {
	name, err := aiAuthProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.AuthProviders.TestIamPermissions(name,
		&agentidentity.TestIamPermissionsRequest{Permissions: flagAIIamPermissions}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("testing IAM permissions: %w", err)
	}
	return emitFormatted(resp, flagAIFormat)
}

// --- access-summaries ---

func runAIASDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiAccessSummaryName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AccessSummaries.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing access summary: %w", err)
	}
	return emitFormatted(got, flagAIFormat)
}

func runAIASList(cmd *cobra.Command, args []string) error {
	parent, err := aiParentLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentIdentityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*agentidentity.AccessSummary
	pageToken := ""
	for {
		call := svc.Projects.Locations.AccessSummaries.List(parent).Context(ctx)
		if flagAIFilter != "" {
			call = call.Filter(flagAIFilter)
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
	if flagAIFormat != "" {
		return emitFormatted(all, flagAIFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, s := range all {
		fmt.Printf("%-40s\n", path.Base(s.Name))
	}
	return nil
}
