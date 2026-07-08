package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	orgpolicy "google.golang.org/api/orgpolicy/v2"
	"gopkg.in/yaml.v3"
)

var orgPoliciesCmd = &cobra.Command{
	Use:   "org-policies",
	Short: "Manage Cloud Organization Policies",
}

var orgPolicyDescribeCmd = &cobra.Command{
	Use:   "describe CONSTRAINT",
	Short: "Show the policy for a constraint on a resource",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgPolicyDescribe,
}

var orgPolicyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organization policies applied to a resource",
	Args:  cobra.NoArgs,
	RunE:  runOrgPolicyList,
}

var orgPolicyDeleteCmd = &cobra.Command{
	Use:   "delete CONSTRAINT",
	Short: "Delete an organization policy",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgPolicyDelete,
}

var orgPolicySetCmd = &cobra.Command{
	Use:   "set-policy POLICY_FILE",
	Short: "Set an organization policy from a JSON or YAML file",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgPolicySet,
}

var orgPolicyEnableEnforceCmd = &cobra.Command{
	Use:   "enable-enforce CONSTRAINT",
	Short: "Set a boolean constraint's enforce value to true",
	Args:  cobra.ExactArgs(1),
	RunE:  makeOrgPolicyEnforce(true),
}

var orgPolicyDisableEnforceCmd = &cobra.Command{
	Use:   "disable-enforce CONSTRAINT",
	Short: "Set a boolean constraint's enforce value to false",
	Args:  cobra.ExactArgs(1),
	RunE:  makeOrgPolicyEnforce(false),
}

var (
	flagOrgPolicyProject     string
	flagOrgPolicyFolder      string
	flagOrgPolicyOrg         string
	flagOrgPolicyEffective   bool
	flagOrgPolicyListPageSize int64
	flagOrgPolicyListLimit    int64
)

func init() {
	for _, c := range []*cobra.Command{
		orgPolicyDescribeCmd, orgPolicyListCmd, orgPolicyDeleteCmd,
		orgPolicyEnableEnforceCmd, orgPolicyDisableEnforceCmd,
	} {
		c.Flags().StringVar(&flagOrgPolicyProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&flagOrgPolicyFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&flagOrgPolicyOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}
	orgPolicyDescribeCmd.Flags().BoolVar(&flagOrgPolicyEffective, "effective", false, "Return the effective policy (including inherited)")
	orgPolicyListCmd.Flags().Int64Var(&flagOrgPolicyListPageSize, "page-size", 0, "Page size for API pagination")
	orgPolicyListCmd.Flags().Int64Var(&flagOrgPolicyListLimit, "limit", 0, "Maximum number of policies to list (0 = no limit)")

	orgPoliciesCmd.AddCommand(
		orgPolicyDescribeCmd,
		orgPolicyListCmd,
		orgPolicyDeleteCmd,
		orgPolicySetCmd,
		orgPolicyEnableEnforceCmd,
		orgPolicyDisableEnforceCmd,
	)
	resourceManagerCmd.AddCommand(orgPoliciesCmd)
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

func runOrgPolicyDescribe(cmd *cobra.Command, args []string) error {
	resource, err := resolveOrgPolicyResource(flagOrgPolicyProject, flagOrgPolicyFolder, flagOrgPolicyOrg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OrgPolicyService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := policyResourceName(resource, args[0])
	if flagOrgPolicyEffective {
		policy, err := describeEffectivePolicy(ctx, svc, name)
		if err != nil {
			return err
		}
		return yamlEncode(policy)
	}
	policy, err := describePolicy(ctx, svc, name)
	if err != nil {
		return err
	}
	return yamlEncode(policy)
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

func runOrgPolicyList(cmd *cobra.Command, args []string) error {
	resource, err := resolveOrgPolicyResource(flagOrgPolicyProject, flagOrgPolicyFolder, flagOrgPolicyOrg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OrgPolicyService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*orgpolicy.GoogleCloudOrgpolicyV2Policy
	pageToken := ""
	for {
		resp, err := listPoliciesPage(ctx, svc, resource, pageToken, flagOrgPolicyListPageSize)
		if err != nil {
			return fmt.Errorf("listing policies: %w", err)
		}
		all = append(all, resp.Policies...)
		if flagOrgPolicyListLimit > 0 && int64(len(all)) >= flagOrgPolicyListLimit {
			all = all[:flagOrgPolicyListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	fmt.Printf("%-70s %s\n", "CONSTRAINT", "ETAG")
	for _, p := range all {
		fmt.Printf("%-70s %s\n", p.Name, p.Etag)
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

func runOrgPolicyDelete(cmd *cobra.Command, args []string) error {
	resource, err := resolveOrgPolicyResource(flagOrgPolicyProject, flagOrgPolicyFolder, flagOrgPolicyOrg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OrgPolicyService(ctx, flagAccount)
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
	fmt.Fprintf(os.Stderr, "Deleted policy [%s].\n", name)
	return nil
}

func runOrgPolicySet(cmd *cobra.Command, args []string) error {
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

	ctx := context.Background()
	svc, err := gcp.OrgPolicyService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Try to update first; if the policy does not exist yet, create it.
	updated, err := patchPolicy(ctx, svc, policy)
	if err == nil {
		return yamlEncode(updated)
	}
	created, cErr := createPolicy(ctx, svc, policy)
	if cErr != nil {
		return fmt.Errorf("setting policy: %w (update failed with: %v)", cErr, err)
	}
	return yamlEncode(created)
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

func makeOrgPolicyEnforce(enforce bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		resource, err := resolveOrgPolicyResource(flagOrgPolicyProject, flagOrgPolicyFolder, flagOrgPolicyOrg)
		if err != nil {
			return err
		}
		ctx := context.Background()
		svc, err := gcp.OrgPolicyService(ctx, flagAccount)
		if err != nil {
			return err
		}
		name := policyResourceName(resource, args[0])

		existing, getErr := describePolicy(ctx, svc, name)
		policy := &orgpolicy.GoogleCloudOrgpolicyV2Policy{Name: name}
		if getErr == nil && existing != nil {
			policy = existing
		}
		rule := &orgpolicy.GoogleCloudOrgpolicyV2PolicySpecPolicyRule{}
		if enforce {
			rule.Enforce = true
			rule.ForceSendFields = []string{"Enforce"}
		} else {
			rule.Enforce = false
			rule.ForceSendFields = []string{"Enforce"}
		}
		policy.Spec = &orgpolicy.GoogleCloudOrgpolicyV2PolicySpec{
			Rules: []*orgpolicy.GoogleCloudOrgpolicyV2PolicySpecPolicyRule{rule},
		}

		if getErr == nil {
			updated, err := patchPolicy(ctx, svc, policy)
			if err != nil {
				return fmt.Errorf("updating policy: %w", err)
			}
			return yamlEncode(updated)
		}
		created, err := createPolicy(ctx, svc, policy)
		if err != nil {
			return fmt.Errorf("creating policy: %w", err)
		}
		return yamlEncode(created)
	}
}
