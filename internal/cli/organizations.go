package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
	"gopkg.in/yaml.v3"
)

// --- gcloud organizations (#279) ---

func newOrganizationsCmd(inv *invocation) *cobra.Command {
	c := &cobra.Command{
		Use:   "organizations",
		Short: "Manage Google Cloud organizations",
	}
	c.AddCommand(
		newOrgDescribeCmd(inv),
		newOrgListCmd(inv),
		newOrgGetIamPolicyCmd(inv),
		newOrgSetIamPolicyCmd(inv),
		newOrgIamBindingCmd(inv, true),
		newOrgIamBindingCmd(inv, false),
	)
	return c
}

func newOrgDescribeCmd(inv *invocation) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ORGANIZATION_ID",
		Short: "Show metadata for an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgDescribe(c.Context(), inv, args[0])
		},
	}
}

type orgListOptions struct {
	filter, format  string
	pageSize, limit int64
	uri             bool
}

func newOrgListCmd(inv *invocation) *cobra.Command {
	var o orgListOptions
	c := &cobra.Command{
		Use:   "list",
		Short: "List organizations accessible by the active account",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgList(c.Context(), inv, o)
		},
	}
	c.Flags().StringVar(&o.filter, "filter", "", "Filter expression (e.g. domain:example.com)")
	c.Flags().Int64Var(&o.pageSize, "page-size", 0, "Page size for API pagination")
	c.Flags().Int64Var(&o.limit, "limit", 0, "Maximum number of organizations to list (0 = no limit)")
	c.Flags().StringVar(&o.format, "format", "", "Output format (json, yaml, or table)")
	c.Flags().BoolVar(&o.uri, "uri", false, "Print resource names only")
	return c
}

func newOrgGetIamPolicyCmd(inv *invocation) *cobra.Command {
	return &cobra.Command{
		Use:   "get-iam-policy ORGANIZATION_ID",
		Short: "Get IAM policy for an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgGetIamPolicy(c.Context(), inv, args[0])
		},
	}
}

func newOrgSetIamPolicyCmd(inv *invocation) *cobra.Command {
	return &cobra.Command{
		Use:   "set-iam-policy ORGANIZATION_ID POLICY_FILE",
		Short: "Set IAM policy for an organization",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgSetIamPolicy(c.Context(), inv, args[0], args[1])
		},
	}
}

// newOrgIamBindingCmd returns add-iam-policy-binding when add is set, else
// remove-iam-policy-binding.
func newOrgIamBindingCmd(inv *invocation, add bool) *cobra.Command {
	var o iamBindingOptions
	c := &cobra.Command{
		Use:   "add-iam-policy-binding ORGANIZATION_ID",
		Short: "Add IAM policy binding for an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgAddIamBinding(c.Context(), inv, args[0], o)
		},
	}
	if !add {
		c.Use = "remove-iam-policy-binding ORGANIZATION_ID"
		c.Short = "Remove IAM policy binding for an organization"
		c.RunE = func(c *cobra.Command, args []string) error {
			return runOrgRemoveIamBinding(c.Context(), inv, args[0], o)
		}
	}
	addIamBindingFlags(c, &o, !add)
	return c
}

// orgResourceName returns the full resource name for an organization,
// accepting either a bare numeric ID or the fully qualified form.
func orgResourceName(orgID string) string {
	orgID = strings.TrimPrefix(orgID, "organizations/")
	return "organizations/" + orgID
}

func runOrgDescribe(ctx context.Context, inv *invocation, id string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	org, err := svc.Organizations.Get(orgResourceName(id)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing organization: %w", err)
	}
	return inv.yaml(org)
}

func runOrgList(ctx context.Context, inv *invocation, o orgListOptions) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	var all []*crm.Organization
	pageToken := ""
	for {
		call := svc.Organizations.Search().Context(ctx)
		if o.filter != "" {
			call = call.Query(o.filter)
		}
		if o.pageSize > 0 {
			call = call.PageSize(o.pageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing organizations: %w", err)
		}
		all = append(all, resp.Organizations...)
		if o.limit > 0 && int64(len(all)) >= o.limit {
			all = all[:o.limit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if o.uri {
		for _, org := range all {
			fmt.Fprintln(inv.stdout, org.Name)
		}
		return nil
	}

	return printOrgList(inv.stdout, all, o.format)
}

// printOrgList writes orgs in the requested format. An empty result prints
// [] for json and nothing otherwise, matching gcloud-python.
func printOrgList(w io.Writer, orgs []*crm.Organization, format string) error {
	switch format {
	case "json":
		return emitJSON(w, orgs)
	case "yaml":
		return emitFormattedTo(w, orgs, "yaml")
	}

	if len(orgs) == 0 {
		return nil
	}
	printTableHeadingTo(w, len(orgs), "%-40s %-20s %s\n", "DISPLAY_NAME", "ID", "DIRECTORY_CUSTOMER_ID")
	for _, o := range orgs {
		fmt.Fprintf(w, "%-40s %-20s %s\n", o.DisplayName, path.Base(o.Name), o.DirectoryCustomerId)
	}
	return nil
}

func runOrgGetIamPolicy(ctx context.Context, inv *invocation, id string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	req := &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}
	policy, err := svc.Organizations.GetIamPolicy(orgResourceName(id), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return inv.yaml(policy)
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

func runOrgSetIamPolicy(ctx context.Context, inv *invocation, id, policyFile string) error {
	policy, err := parsePolicyFile(policyFile)
	if err != nil {
		return err
	}

	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	req := &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}
	updated, err := svc.Organizations.SetIamPolicy(orgResourceName(id), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Updated IAM policy for organization [%s].\n", strings.TrimPrefix(orgResourceName(id), "organizations/"))
	return inv.yaml(updated)
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

func runOrgAddIamBinding(ctx context.Context, inv *invocation, id string, o iamBindingOptions) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	resource := orgResourceName(id)
	getReq := &crm.GetIamPolicyRequest{Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3}}
	policy, err := svc.Organizations.GetIamPolicy(resource, getReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	addBindingToPolicy(policy, o.role, o.member, o.condition())
	policy.Version = 3

	updated, err := svc.Organizations.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Updated IAM policy for organization [%s].\n", strings.TrimPrefix(resource, "organizations/"))
	return inv.yaml(updated)
}

func runOrgRemoveIamBinding(ctx context.Context, inv *invocation, id string, o iamBindingOptions) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	resource := orgResourceName(id)
	getReq := &crm.GetIamPolicyRequest{Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3}}
	policy, err := svc.Organizations.GetIamPolicy(resource, getReq).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	if !removeBindingFromPolicy(policy, o.role, o.member, o.condition(), o.allConditions) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", o.role, o.member)
	}

	updated, err := svc.Organizations.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Updated IAM policy for organization [%s].\n", strings.TrimPrefix(resource, "organizations/"))
	return inv.yaml(updated)
}
