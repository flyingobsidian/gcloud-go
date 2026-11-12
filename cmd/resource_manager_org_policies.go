package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	orgpolicy "google.golang.org/api/orgpolicy/v2"
	"gopkg.in/yaml.v3"
)

// orgPoliciesFlags holds the flag values of the commands built by newOrgPoliciesCmd.
type orgPoliciesFlags struct {
	orgPolicyProject      string
	orgPolicyFolder       string
	orgPolicyOrg          string
	orgPolicyEffective    bool
	orgPolicyListPageSize int64
	orgPolicyListLimit    int64
}

func newOrgPoliciesCmd(inv *invocation) *cobra.Command {
	f := &orgPoliciesFlags{}

	orgPoliciesCmd := &cobra.Command{
		Use:   "org-policies",
		Short: "Manage Cloud Organization Policies",
	}

	orgPolicyDescribeCmd := &cobra.Command{
		Use:   "describe CONSTRAINT",
		Short: "Show the policy for a constraint on a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgPolicyDescribe(c.Context(), inv, f, c, args)
		},
	}

	orgPolicyListCmd := &cobra.Command{
		Use:   "list",
		Short: "List organization policies applied to a resource",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgPolicyList(c.Context(), inv, f, c, args)
		},
	}

	orgPolicyDeleteCmd := &cobra.Command{
		Use:   "delete CONSTRAINT",
		Short: "Delete an organization policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgPolicyDelete(c.Context(), inv, f, c, args)
		},
	}

	orgPolicySetCmd := &cobra.Command{
		Use:   "set-policy POLICY_FILE",
		Short: "Set an organization policy from a JSON or YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgPolicySet(c.Context(), inv, f, c, args)
		},
	}

	orgPolicyEnableEnforceCmd := newOrgPolicyEnforceCmd(inv, f, true)
	orgPolicyDisableEnforceCmd := newOrgPolicyEnforceCmd(inv, f, false)
	for _, c := range []*cobra.Command{
		orgPolicyDescribeCmd, orgPolicyListCmd, orgPolicyDeleteCmd,
		orgPolicyEnableEnforceCmd, orgPolicyDisableEnforceCmd,
	} {
		c.Flags().StringVar(&f.orgPolicyProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&f.orgPolicyFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&f.orgPolicyOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}
	orgPolicyDescribeCmd.Flags().BoolVar(&f.orgPolicyEffective, "effective", false, "Return the effective policy (including inherited)")
	orgPolicyListCmd.Flags().Int64Var(&f.orgPolicyListPageSize, "page-size", 0, "Page size for API pagination")
	orgPolicyListCmd.Flags().Int64Var(&f.orgPolicyListLimit, "limit", 0, "Maximum number of policies to list (0 = no limit)")

	orgPoliciesCmd.AddCommand(
		orgPolicyDescribeCmd,
		orgPolicyListCmd,
		orgPolicyDeleteCmd,
		orgPolicySetCmd,
		orgPolicyEnableEnforceCmd,
		orgPolicyDisableEnforceCmd,
	)
	return orgPoliciesCmd
}

// resolveOrgPolicyResource returns the resource identifier (e.g.
// "projects/123", "folders/456", "organizations/789") for policy operations.
func resolveOrgPolicyResource(project, folder, org string) (string, error) {
	set := 0
	if project != "" {
		set++
	}
	if folder != "" {
		set++
	}
	if org != "" {
		set++
	}
	if set == 0 {
		return "", fmt.Errorf("one of --project, --folder, or --organization is required")
	}
	if set > 1 {
		return "", fmt.Errorf("specify only one of --project, --folder, or --organization")
	}
	switch {
	case project != "":
		return "projects/" + project, nil
	case folder != "":
		return "folders/" + strings.TrimPrefix(folder, "folders/"), nil
	default:
		return "organizations/" + strings.TrimPrefix(org, "organizations/"), nil
	}
}

// policyResourceName builds the full name of an org-policy for the given
// resource and constraint. Accepts a bare constraint name (e.g.
// "constraints/compute.disableSerialPortAccess") or one already prefixed with
// the resource path.
func policyResourceName(resource, constraint string) string {
	if strings.Contains(constraint, "/policies/") {
		return constraint
	}
	constraint = strings.TrimPrefix(constraint, "constraints/")
	return resource + "/policies/" + constraint
}

func runOrgPolicyDescribe(ctx context.Context, inv *invocation, f *orgPoliciesFlags, cmd *cobra.Command, args []string) error {
	resource, err := resolveOrgPolicyResource(f.orgPolicyProject, f.orgPolicyFolder, f.orgPolicyOrg)
	if err != nil {
		return err
	}
	svc, err := inv.clients().OrgPolicy(ctx)
	if err != nil {
		return err
	}
	name := policyResourceName(resource, args[0])
	if f.orgPolicyEffective {
		policy, err := describeEffectivePolicy(ctx, svc, name)
		if err != nil {
			return err
		}
		return inv.yaml(policy)
	}
	policy, err := describePolicy(ctx, svc, name)
	if err != nil {
		return err
	}
	return inv.yaml(policy)
}

func describePolicy(ctx context.Context, svc *orgpolicy.Service, name string) (*orgpolicy.GoogleCloudOrgpolicyV2Policy, error) {
	switch {
	case strings.HasPrefix(name, "projects/"):
		return svc.Projects.Policies.Get(name).Context(ctx).Do()
	case strings.HasPrefix(name, "folders/"):
		return svc.Folders.Policies.Get(name).Context(ctx).Do()
	case strings.HasPrefix(name, "organizations/"):
		return svc.Organizations.Policies.Get(name).Context(ctx).Do()
	}
	return nil, fmt.Errorf("unrecognized policy resource: %s", name)
}

func describeEffectivePolicy(ctx context.Context, svc *orgpolicy.Service, name string) (*orgpolicy.GoogleCloudOrgpolicyV2Policy, error) {
	switch {
	case strings.HasPrefix(name, "projects/"):
		return svc.Projects.Policies.GetEffectivePolicy(name).Context(ctx).Do()
	case strings.HasPrefix(name, "folders/"):
		return svc.Folders.Policies.GetEffectivePolicy(name).Context(ctx).Do()
	case strings.HasPrefix(name, "organizations/"):
		return svc.Organizations.Policies.GetEffectivePolicy(name).Context(ctx).Do()
	}
	return nil, fmt.Errorf("unrecognized policy resource: %s", name)
}

func runOrgPolicyList(ctx context.Context, inv *invocation, f *orgPoliciesFlags, cmd *cobra.Command, args []string) error {
	resource, err := resolveOrgPolicyResource(f.orgPolicyProject, f.orgPolicyFolder, f.orgPolicyOrg)
	if err != nil {
		return err
	}
	svc, err := inv.clients().OrgPolicy(ctx)
	if err != nil {
		return err
	}

	var all []*orgpolicy.GoogleCloudOrgpolicyV2Policy
	pageToken := ""
	for {
		resp, err := listPoliciesPage(ctx, svc, resource, pageToken, f.orgPolicyListPageSize)
		if err != nil {
			return fmt.Errorf("listing policies: %w", err)
		}
		all = append(all, resp.Policies...)
		if f.orgPolicyListLimit > 0 && int64(len(all)) >= f.orgPolicyListLimit {
			all = all[:f.orgPolicyListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	printTableHeadingTo(inv.stdout, len(all), "%-70s %s\n", "CONSTRAINT", "ETAG")
	for _, p := range all {
		fmt.Fprintf(inv.stdout, "%-70s %s\n", p.Name, p.Etag)
	}
	return nil
}

func listPoliciesPage(ctx context.Context, svc *orgpolicy.Service, parent, pageToken string, pageSize int64) (*orgpolicy.GoogleCloudOrgpolicyV2ListPoliciesResponse, error) {
	switch {
	case strings.HasPrefix(parent, "projects/"):
		call := svc.Projects.Policies.List(parent).Context(ctx)
		if pageSize > 0 {
			call = call.PageSize(pageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		return call.Do()
	case strings.HasPrefix(parent, "folders/"):
		call := svc.Folders.Policies.List(parent).Context(ctx)
		if pageSize > 0 {
			call = call.PageSize(pageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		return call.Do()
	case strings.HasPrefix(parent, "organizations/"):
		call := svc.Organizations.Policies.List(parent).Context(ctx)
		if pageSize > 0 {
			call = call.PageSize(pageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		return call.Do()
	}
	return nil, fmt.Errorf("unrecognized parent: %s", parent)
}

func runOrgPolicyDelete(ctx context.Context, inv *invocation, f *orgPoliciesFlags, cmd *cobra.Command, args []string) error {
	resource, err := resolveOrgPolicyResource(f.orgPolicyProject, f.orgPolicyFolder, f.orgPolicyOrg)
	if err != nil {
		return err
	}
	svc, err := inv.clients().OrgPolicy(ctx)
	if err != nil {
		return err
	}
	name := policyResourceName(resource, args[0])
	switch {
	case strings.HasPrefix(name, "projects/"):
		_, err = svc.Projects.Policies.Delete(name).Context(ctx).Do()
	case strings.HasPrefix(name, "folders/"):
		_, err = svc.Folders.Policies.Delete(name).Context(ctx).Do()
	case strings.HasPrefix(name, "organizations/"):
		_, err = svc.Organizations.Policies.Delete(name).Context(ctx).Do()
	default:
		return fmt.Errorf("unrecognized policy resource: %s", name)
	}
	if err != nil {
		return fmt.Errorf("deleting policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Deleted policy [%s].\n", name)
	return nil
}

func runOrgPolicySet(ctx context.Context, inv *invocation, f *orgPoliciesFlags, cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading policy file: %w", err)
	}
	policy := &orgpolicy.GoogleCloudOrgpolicyV2Policy{}
	if err := yaml.Unmarshal(data, policy); err != nil {
		return fmt.Errorf("parsing policy file: %w", err)
	}
	if policy.Name == "" {
		return fmt.Errorf("policy file must include a name field")
	}

	svc, err := inv.clients().OrgPolicy(ctx)
	if err != nil {
		return err
	}

	// Try to update first; if the policy does not exist yet, create it.
	updated, err := patchPolicy(ctx, svc, policy)
	if err == nil {
		return inv.yaml(updated)
	}
	created, cErr := createPolicy(ctx, svc, policy)
	if cErr != nil {
		return fmt.Errorf("setting policy: %w (update failed with: %v)", cErr, err)
	}
	return inv.yaml(created)
}

func patchPolicy(ctx context.Context, svc *orgpolicy.Service, p *orgpolicy.GoogleCloudOrgpolicyV2Policy) (*orgpolicy.GoogleCloudOrgpolicyV2Policy, error) {
	switch {
	case strings.HasPrefix(p.Name, "projects/"):
		return svc.Projects.Policies.Patch(p.Name, p).Context(ctx).Do()
	case strings.HasPrefix(p.Name, "folders/"):
		return svc.Folders.Policies.Patch(p.Name, p).Context(ctx).Do()
	case strings.HasPrefix(p.Name, "organizations/"):
		return svc.Organizations.Policies.Patch(p.Name, p).Context(ctx).Do()
	}
	return nil, fmt.Errorf("unrecognized policy resource: %s", p.Name)
}

func createPolicy(ctx context.Context, svc *orgpolicy.Service, p *orgpolicy.GoogleCloudOrgpolicyV2Policy) (*orgpolicy.GoogleCloudOrgpolicyV2Policy, error) {
	parent := parentForPolicyName(p.Name)
	switch {
	case strings.HasPrefix(parent, "projects/"):
		return svc.Projects.Policies.Create(parent, p).Context(ctx).Do()
	case strings.HasPrefix(parent, "folders/"):
		return svc.Folders.Policies.Create(parent, p).Context(ctx).Do()
	case strings.HasPrefix(parent, "organizations/"):
		return svc.Organizations.Policies.Create(parent, p).Context(ctx).Do()
	}
	return nil, fmt.Errorf("unrecognized policy parent: %s", parent)
}

// parentForPolicyName extracts the parent resource from a policy name like
// "projects/123/policies/constraint.foo".
func parentForPolicyName(name string) string {
	idx := strings.Index(name, "/policies/")
	if idx < 0 {
		return name
	}
	return name[:idx]
}

// newOrgPolicyEnforceCmd returns enable-enforce when enforce is set, else
// disable-enforce.
func newOrgPolicyEnforceCmd(inv *invocation, f *orgPoliciesFlags, enforce bool) *cobra.Command {
	c := &cobra.Command{
		Use:   "enable-enforce CONSTRAINT",
		Short: "Set a boolean constraint's enforce value to true",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runOrgPolicyEnforce(c.Context(), inv, f, args[0], enforce)
		},
	}
	if !enforce {
		c.Use = "disable-enforce CONSTRAINT"
		c.Short = "Set a boolean constraint's enforce value to false"
	}
	return c
}

func runOrgPolicyEnforce(ctx context.Context, inv *invocation, f *orgPoliciesFlags, constraint string, enforce bool) error {
	resource, err := resolveOrgPolicyResource(f.orgPolicyProject, f.orgPolicyFolder, f.orgPolicyOrg)
	if err != nil {
		return err
	}
	svc, err := inv.clients().OrgPolicy(ctx)
	if err != nil {
		return err
	}
	name := policyResourceName(resource, constraint)

	existing, getErr := describePolicy(ctx, svc, name)
	policy := &orgpolicy.GoogleCloudOrgpolicyV2Policy{Name: name}
	if getErr == nil && existing != nil {
		policy = existing
	}
	rule := &orgpolicy.GoogleCloudOrgpolicyV2PolicySpecPolicyRule{
		Enforce:         enforce,
		ForceSendFields: []string{"Enforce"},
	}
	policy.Spec = &orgpolicy.GoogleCloudOrgpolicyV2PolicySpec{
		Rules: []*orgpolicy.GoogleCloudOrgpolicyV2PolicySpecPolicyRule{rule},
	}

	if getErr == nil {
		updated, err := patchPolicy(ctx, svc, policy)
		if err != nil {
			return fmt.Errorf("updating policy: %w", err)
		}
		return inv.yaml(updated)
	}
	created, err := createPolicy(ctx, svc, policy)
	if err != nil {
		return fmt.Errorf("creating policy: %w", err)
	}
	return inv.yaml(created)
}
