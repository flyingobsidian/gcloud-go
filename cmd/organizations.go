package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
	"gopkg.in/yaml.v3"
)

// --- gcloud organizations (#279) ---

var organizationsCmd = &cobra.Command{
	Use:   "organizations",
	Short: "Manage Google Cloud organizations",
}

var orgDescribeCmd = &cobra.Command{
	Use:   "describe ORGANIZATION_ID",
	Short: "Show metadata for an organization",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgDescribe,
}

var orgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations accessible by the active account",
	Args:  cobra.NoArgs,
	RunE:  runOrgList,
}

var orgGetIamPolicyCmd = &cobra.Command{
	Use:   "get-iam-policy ORGANIZATION_ID",
	Short: "Get IAM policy for an organization",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgGetIamPolicy,
}

var orgSetIamPolicyCmd = &cobra.Command{
	Use:   "set-iam-policy ORGANIZATION_ID POLICY_FILE",
	Short: "Set IAM policy for an organization",
	Args:  cobra.ExactArgs(2),
	RunE:  runOrgSetIamPolicy,
}

var orgAddIamBindingCmd = &cobra.Command{
	Use:   "add-iam-policy-binding ORGANIZATION_ID",
	Short: "Add IAM policy binding for an organization",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgAddIamBinding,
}

var orgRemoveIamBindingCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding ORGANIZATION_ID",
	Short: "Remove IAM policy binding for an organization",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgRemoveIamBinding,
}

var (
	flagOrgListFilter    string
	flagOrgListPageSize  int64
	flagOrgListLimit     int64
	flagOrgListFormat    string
	flagOrgListURI       bool
	flagOrgIamMember     string
	flagOrgIamRole       string
	flagOrgIamCondExpr   string
	flagOrgIamCondTitle  string
	flagOrgIamCondDesc   string
	flagOrgIamAllCond    bool
)

func init() {
	orgListCmd.Flags().StringVar(&flagOrgListFilter, "filter", "", "Filter expression (e.g. domain:example.com)")
	orgListCmd.Flags().Int64Var(&flagOrgListPageSize, "page-size", 0, "Page size for API pagination")
	orgListCmd.Flags().Int64Var(&flagOrgListLimit, "limit", 0, "Maximum number of organizations to list (0 = no limit)")
	orgListCmd.Flags().StringVar(&flagOrgListFormat, "format", "", "Output format (json, yaml, or table)")
	orgListCmd.Flags().BoolVar(&flagOrgListURI, "uri", false, "Print resource names only")

	for _, c := range []*cobra.Command{orgAddIamBindingCmd, orgRemoveIamBindingCmd} {
		c.Flags().StringVar(&flagOrgIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&flagOrgIamRole, "role", "", "IAM role to bind (e.g. roles/browser) (required)")
		c.Flags().StringVar(&flagOrgIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&flagOrgIamCondTitle, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&flagOrgIamCondDesc, "condition-description", "", "Description for a conditional binding")
		c.MarkFlagRequired("member")
		c.MarkFlagRequired("role")
	}
	orgRemoveIamBindingCmd.Flags().BoolVar(&flagOrgIamAllCond, "all", false, "Remove the member from all bindings for the role, regardless of condition")

	organizationsCmd.AddCommand(orgDescribeCmd)
	organizationsCmd.AddCommand(orgListCmd)
	organizationsCmd.AddCommand(orgGetIamPolicyCmd)
	organizationsCmd.AddCommand(orgSetIamPolicyCmd)
	organizationsCmd.AddCommand(orgAddIamBindingCmd)
	organizationsCmd.AddCommand(orgRemoveIamBindingCmd)
	rootCmd.AddCommand(organizationsCmd)
}

// orgResourceName returns the full resource name for an organization,
// accepting either a bare numeric ID or the fully qualified form.
func orgResourceName(orgID string) string {
	orgID = strings.TrimPrefix(orgID, "organizations/")
	return "organizations/" + orgID
}

func runOrgDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	org, err := svc.Organizations.Get(orgResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing organization: %w", err)
	}
	return yamlEncode(org)
}

func runOrgList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*crm.Organization
	pageToken := ""
	for {
		call := svc.Organizations.Search().Context(ctx)
		if flagOrgListFilter != "" {
			call = call.Query(flagOrgListFilter)
		}
		if flagOrgListPageSize > 0 {
			call = call.PageSize(flagOrgListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing organizations: %w", err)
		}
		all = append(all, resp.Organizations...)
		if flagOrgListLimit > 0 && int64(len(all)) >= flagOrgListLimit {
			all = all[:flagOrgListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagOrgListURI {
		for _, o := range all {
			fmt.Println(o.Name)
		}
		return nil
	}

	switch flagOrgListFormat {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(all)
	case "yaml":
		return yamlEncode(all)
	}

	fmt.Printf("%-40s %-20s %s\n", "DISPLAY_NAME", "ID", "DIRECTORY_CUSTOMER_ID")
	for _, o := range all {
		fmt.Printf("%-40s %-20s %s\n", o.DisplayName, path.Base(o.Name), o.DirectoryCustomerId)
	}
	return nil
}

func runOrgGetIamPolicy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	req := &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}
	policy, err := svc.Organizations.GetIamPolicy(orgResourceName(args[0]), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return yamlEncode(policy)
}

// parsePolicyFile reads and decodes an IAM policy from JSON or YAML on disk.
func parsePolicyFile(pathname string) (*crm.Policy, error) {
	data, err := os.ReadFile(pathname)
	if err != nil {
		return nil, fmt.Errorf("reading policy file: %w", err)
	}
	policy := &crm.Policy{}
	// JSON is a subset of YAML for policy files, so yaml.Unmarshal handles both.
	if err := yaml.Unmarshal(data, policy); err != nil {
		return nil, fmt.Errorf("parsing policy file: %w", err)
	}
	return policy, nil
}

func runOrgSetIamPolicy(cmd *cobra.Command, args []string) error {
	policy, err := parsePolicyFile(args[1])
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	req := &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}
	updated, err := svc.Organizations.SetIamPolicy(orgResourceName(args[0]), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for organization [%s].\n", strings.TrimPrefix(orgResourceName(args[0]), "organizations/"))
	return yamlEncode(updated)
}

// buildCondition returns a *crm.Expr from the --condition-* flags, or nil if
// no condition flags were set.
func buildCondition() *crm.Expr {
	if flagOrgIamCondExpr == "" && flagOrgIamCondTitle == "" && flagOrgIamCondDesc == "" {
		return nil
	}
	return &crm.Expr{
		Expression:  flagOrgIamCondExpr,
		Title:       flagOrgIamCondTitle,
		Description: flagOrgIamCondDesc,
	}
}

// conditionsEqual reports whether two condition expressions describe the same
// binding for the purpose of matching add/remove operations.
func conditionsEqual(a, b *crm.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

// addBindingToPolicy adds member to the binding matching role and condition,
// creating the binding if none exists. Returns true if the policy changed.
func addBindingToPolicy(policy *crm.Policy, role, member string, condition *crm.Expr) bool {
	for _, b := range policy.Bindings {
		if b.Role != role || !conditionsEqual(b.Condition, condition) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return false
			}
		}
		b.Members = append(b.Members, member)
		return true
	}
	policy.Bindings = append(policy.Bindings, &crm.Binding{
		Role:      role,
		Members:   []string{member},
		Condition: condition,
	})
	return true
}

// removeBindingFromPolicy removes member from bindings matching role. If
// allConditions is true, matches every binding for the role; otherwise only
// the binding whose condition matches. Returns true if the policy changed.
func removeBindingFromPolicy(policy *crm.Policy, role, member string, condition *crm.Expr, allConditions bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConditions || conditionsEqual(b.Condition, condition))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				changed = true
				continue
			}
			newMembers = append(newMembers, m)
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

func runOrgAddIamBinding(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resource := orgResourceName(args[0])
	getReq := &crm.GetIamPolicyRequest{Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3}}
	policy, err := svc.Organizations.GetIamPolicy(resource, getReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	addBindingToPolicy(policy, flagOrgIamRole, flagOrgIamMember, buildCondition())
	policy.Version = 3

	updated, err := svc.Organizations.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for organization [%s].\n", strings.TrimPrefix(resource, "organizations/"))
	return yamlEncode(updated)
}

func runOrgRemoveIamBinding(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resource := orgResourceName(args[0])
	getReq := &crm.GetIamPolicyRequest{Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3}}
	policy, err := svc.Organizations.GetIamPolicy(resource, getReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	if !removeBindingFromPolicy(policy, flagOrgIamRole, flagOrgIamMember, buildCondition(), flagOrgIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagOrgIamRole, flagOrgIamMember)
	}

	updated, err := svc.Organizations.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for organization [%s].\n", strings.TrimPrefix(resource, "organizations/"))
	return yamlEncode(updated)
}
