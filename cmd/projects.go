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
)

// --- gcloud projects (#375, #515) ---

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage Google Cloud projects",
}

var projectDescribeCmd = &cobra.Command{
	Use:   "describe PROJECT_ID_OR_NUMBER",
	Short: "Show metadata for a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectDescribe,
}

var flagProjectDescribeFormat string

var projectCreateCmd = &cobra.Command{
	Use:   "create PROJECT_ID",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectCreate,
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete PROJECT_ID_OR_NUMBER",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectDelete,
}

var projectUndeleteCmd = &cobra.Command{
	Use:   "undelete PROJECT_ID_OR_NUMBER",
	Short: "Undelete a project marked for deletion",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectUndelete,
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update PROJECT_ID_OR_NUMBER",
	Short: "Update the name of a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectUpdate,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects accessible by the active account",
	Args:  cobra.NoArgs,
	RunE:  runProjectList,
}

var projectGetAncestorsCmd = &cobra.Command{
	Use:   "get-ancestors PROJECT_ID_OR_NUMBER",
	Short: "Get the ancestors for a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectGetAncestors,
}

var projectGetIamPolicyCmd = &cobra.Command{
	Use:   "get-iam-policy PROJECT_ID_OR_NUMBER",
	Short: "Get IAM policy for a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectGetIamPolicy,
}

var projectGetAncestorsIamPolicyCmd = &cobra.Command{
	Use:   "get-ancestors-iam-policy PROJECT_ID_OR_NUMBER",
	Short: "Get IAM policies for a project and its ancestors",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectGetAncestorsIamPolicy,
}

var projectSetIamPolicyCmd = &cobra.Command{
	Use:   "set-iam-policy PROJECT_ID_OR_NUMBER POLICY_FILE",
	Short: "Set IAM policy for a project",
	Args:  cobra.ExactArgs(2),
	RunE:  runProjectSetIamPolicy,
}

var projectAddIamBindingCmd = &cobra.Command{
	Use:   "add-iam-policy-binding PROJECT_ID_OR_NUMBER",
	Short: "Add IAM policy binding for a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectAddIamBinding,
}

var projectRemoveIamBindingCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding PROJECT_ID_OR_NUMBER",
	Short: "Remove IAM policy binding for a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectRemoveIamBinding,
}

var (
	flagProjectCreateName    string
	flagProjectCreateFolder  string
	flagProjectCreateOrg     string
	flagProjectCreateLabels  map[string]string
	flagProjectUpdateName    string
	flagProjectListFolder    string
	flagProjectListOrg       string
	flagProjectListFilter    string
	flagProjectListPageSize  int64
	flagProjectListLimit     int64
	flagProjectListFormat    string
	flagProjectListURI       bool
	flagProjectListShowDel   bool
	flagProjectIamMember     string
	flagProjectIamRole       string
	flagProjectIamCondExpr   string
	flagProjectIamCondTitle  string
	flagProjectIamCondDesc   string
	flagProjectIamAllCond    bool
)

func init() {
	projectDescribeCmd.Flags().StringVar(&flagProjectDescribeFormat, "format", "", "Output format: yaml (default), json, csv(FIELDS), table(FIELDS), text(FIELDS), value(FIELDS), config(FIELDS), get(FIELD)")

	projectCreateCmd.Flags().StringVar(&flagProjectCreateName, "name", "", "Display name for the new project (defaults to the project ID)")
	projectCreateCmd.Flags().StringVar(&flagProjectCreateFolder, "folder", "", "Parent folder ID (mutually exclusive with --organization)")
	projectCreateCmd.Flags().StringVar(&flagProjectCreateOrg, "organization", "", "Parent organization ID (mutually exclusive with --folder)")
	projectCreateCmd.Flags().StringToStringVar(&flagProjectCreateLabels, "labels", nil, "Comma-separated key=value label pairs")

	projectUpdateCmd.Flags().StringVar(&flagProjectUpdateName, "name", "", "New display name for the project (required)")
	projectUpdateCmd.MarkFlagRequired("name")

	projectListCmd.Flags().StringVar(&flagProjectListFolder, "folder", "", "List projects under this parent folder")
	projectListCmd.Flags().StringVar(&flagProjectListOrg, "organization", "", "List projects under this parent organization")
	projectListCmd.Flags().StringVar(&flagProjectListFilter, "filter", "", "Search query expression (e.g. state:ACTIVE)")
	projectListCmd.Flags().Int64Var(&flagProjectListPageSize, "page-size", 0, "Page size for API pagination")
	projectListCmd.Flags().Int64Var(&flagProjectListLimit, "limit", 0, "Maximum number of projects to list (0 = no limit)")
	projectListCmd.Flags().StringVar(&flagProjectListFormat, "format", "", "Output format (json, yaml, or table)")
	projectListCmd.Flags().BoolVar(&flagProjectListURI, "uri", false, "Print resource names only")
	projectListCmd.Flags().BoolVar(&flagProjectListShowDel, "show-deleted", false, "Include deleted projects (only when listing under a parent)")

	for _, c := range []*cobra.Command{projectAddIamBindingCmd, projectRemoveIamBindingCmd} {
		c.Flags().StringVar(&flagProjectIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&flagProjectIamRole, "role", "", "IAM role to bind (e.g. roles/browser) (required)")
		c.Flags().StringVar(&flagProjectIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&flagProjectIamCondTitle, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&flagProjectIamCondDesc, "condition-description", "", "Description for a conditional binding")
		c.MarkFlagRequired("member")
		c.MarkFlagRequired("role")
	}
	projectRemoveIamBindingCmd.Flags().BoolVar(&flagProjectIamAllCond, "all", false, "Remove the member from all bindings for the role, regardless of condition")

	projectsCmd.AddCommand(
		projectAddIamBindingCmd,
		projectCreateCmd,
		projectDeleteCmd,
		projectDescribeCmd,
		projectGetAncestorsCmd,
		projectGetAncestorsIamPolicyCmd,
		projectGetIamPolicyCmd,
		projectListCmd,
		projectRemoveIamBindingCmd,
		projectSetIamPolicyCmd,
		projectUndeleteCmd,
		projectUpdateCmd,
	)
	rootCmd.AddCommand(projectsCmd)
}

// projectResourceName returns the fully qualified project name, accepting
// either a bare project ID or number, or the "projects/..." form.
func projectResourceName(id string) string {
	return "projects/" + strings.TrimPrefix(id, "projects/")
}

// projectParent converts --folder or --organization flags into the
// corresponding "folders/..." or "organizations/..." parent value. An empty
// result means no parent (project has no parent resource).
func projectParent(folder, org string) (string, error) {
	if folder != "" && org != "" {
		return "", fmt.Errorf("specify only one of --folder or --organization")
	}
	if folder != "" {
		return "folders/" + strings.TrimPrefix(folder, "folders/"), nil
	}
	if org != "" {
		return "organizations/" + strings.TrimPrefix(org, "organizations/"), nil
	}
	return "", nil
}

func runProjectDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	project, err := svc.Projects.Get(projectResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project: %w", err)
	}
	return emitFormatted(project, flagProjectDescribeFormat)
}

func runProjectCreate(cmd *cobra.Command, args []string) error {
	parent, err := projectParent(flagProjectCreateFolder, flagProjectCreateOrg)
	if err != nil {
		return err
	}

	id := args[0]
	displayName := flagProjectCreateName
	if displayName == "" {
		displayName = id
	}

	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	op, err := svc.Projects.Create(&crm.Project{
		ProjectId:   id,
		DisplayName: displayName,
		Parent:      parent,
		Labels:      flagProjectCreateLabels,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating project: %w", err)
	}
	fmt.Printf("Create project in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Delete(projectResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}
	fmt.Printf("Delete project in progress (operation: %s).\n", op.Name)
	fmt.Fprintf(os.Stderr, "You can undo this operation for a limited period by running:\n"+
		"    gcloud projects undelete %s\n", strings.TrimPrefix(args[0], "projects/"))
	return yamlEncode(op)
}

func runProjectUndelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Undelete(projectResourceName(args[0]), &crm.UndeleteProjectRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting project: %w", err)
	}
	fmt.Printf("Undelete project in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runProjectUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Patch(projectResourceName(args[0]), &crm.Project{
		DisplayName: flagProjectUpdateName,
	}).UpdateMask("display_name").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}
	fmt.Printf("Update project in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runProjectList(cmd *cobra.Command, args []string) error {
	parent, err := projectParent(flagProjectListFolder, flagProjectListOrg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	all, err := listProjects(ctx, svc, parent, flagProjectListFilter,
		flagProjectListPageSize, flagProjectListLimit, flagProjectListShowDel)
	if err != nil {
		return err
	}

	if flagProjectListURI {
		for _, p := range all {
			fmt.Println(p.Name)
		}
		return nil
	}

	switch flagProjectListFormat {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(all)
	case "yaml":
		return yamlEncode(all)
	}

	fmt.Printf("%-30s %-30s %-20s %s\n", "PROJECT_ID", "NAME", "PROJECT_NUMBER", "STATE")
	for _, p := range all {
		fmt.Printf("%-30s %-30s %-20s %s\n", p.ProjectId, p.DisplayName, path.Base(p.Name), p.State)
	}
	return nil
}

// listProjects pages through v3.Projects.List(parent=...) when a parent is
// given, and v3.Projects.Search(query=...) otherwise. limit caps the total
// number of returned projects (0 = no cap).
func listProjects(ctx context.Context, svc *crm.Service, parent, query string, pageSize, limit int64, showDeleted bool) ([]*crm.Project, error) {
	var all []*crm.Project
	pageToken := ""
	for {
		var (
			projects []*crm.Project
			nextTok  string
			err      error
		)
		if parent != "" {
			call := svc.Projects.List().Parent(parent).ShowDeleted(showDeleted).Context(ctx)
			if pageSize > 0 {
				call = call.PageSize(pageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, e := call.Do()
			if e != nil {
				return nil, fmt.Errorf("listing projects: %w", e)
			}
			projects, nextTok, err = resp.Projects, resp.NextPageToken, nil
		} else {
			call := svc.Projects.Search().Context(ctx)
			if query != "" {
				call = call.Query(query)
			}
			if pageSize > 0 {
				call = call.PageSize(pageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, e := call.Do()
			if e != nil {
				return nil, fmt.Errorf("searching projects: %w", e)
			}
			projects, nextTok, err = resp.Projects, resp.NextPageToken, nil
		}
		if err != nil {
			return nil, err
		}
		all = append(all, projects...)
		if limit > 0 && int64(len(all)) >= limit {
			return all[:limit], nil
		}
		if nextTok == "" {
			return all, nil
		}
		pageToken = nextTok
	}
}

func runProjectGetAncestors(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	ancestors, err := projectAncestors(ctx, svc, args[0])
	if err != nil {
		return err
	}

	fmt.Printf("%-40s %s\n", "ID", "TYPE")
	for _, a := range ancestors {
		fmt.Printf("%-40s %s\n", a.id, a.kind)
	}
	return nil
}

// ancestor is a single entry in the project ancestry chain: the project itself
// first, followed by its parent folder(s) and terminating organization.
type ancestor struct {
	id, kind string
}

// projectAncestors walks the parent chain from a project up through folders
// to the terminating organization (or until parent is empty), returning the
// chain in order from the project outward.
func projectAncestors(ctx context.Context, svc *crm.Service, projectID string) ([]ancestor, error) {
	project, err := svc.Projects.Get(projectResourceName(projectID)).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("describing project: %w", err)
	}

	chain := []ancestor{{id: project.ProjectId, kind: "project"}}
	parent := project.Parent
	for parent != "" {
		switch {
		case strings.HasPrefix(parent, "organizations/"):
			chain = append(chain, ancestor{id: strings.TrimPrefix(parent, "organizations/"), kind: "organization"})
			return chain, nil
		case strings.HasPrefix(parent, "folders/"):
			id := strings.TrimPrefix(parent, "folders/")
			chain = append(chain, ancestor{id: id, kind: "folder"})
			folder, err := svc.Folders.Get(parent).Context(ctx).Do()
			if err != nil {
				return nil, fmt.Errorf("describing folder %s: %w", parent, err)
			}
			parent = folder.Parent
		default:
			return nil, fmt.Errorf("unrecognized parent resource: %s", parent)
		}
	}
	return chain, nil
}

func runProjectGetIamPolicy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.GetIamPolicy(projectResourceName(args[0]), &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return yamlEncode(policy)
}

func runProjectGetAncestorsIamPolicy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	ancestors, err := projectAncestors(ctx, svc, args[0])
	if err != nil {
		return err
	}

	type ancestorPolicy struct {
		ID     string      `json:"id" yaml:"id"`
		Type   string      `json:"type" yaml:"type"`
		Policy *crm.Policy `json:"policy" yaml:"policy"`
	}

	results := make([]ancestorPolicy, 0, len(ancestors))
	for _, a := range ancestors {
		var (
			policy *crm.Policy
			err    error
		)
		req := &crm.GetIamPolicyRequest{Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3}}
		switch a.kind {
		case "project":
			policy, err = svc.Projects.GetIamPolicy(projectResourceName(a.id), req).Context(ctx).Do()
		case "folder":
			policy, err = svc.Folders.GetIamPolicy("folders/"+a.id, req).Context(ctx).Do()
		case "organization":
			policy, err = svc.Organizations.GetIamPolicy("organizations/"+a.id, req).Context(ctx).Do()
		}
		if err != nil {
			return fmt.Errorf("getting IAM policy for %s %s: %w", a.kind, a.id, err)
		}
		results = append(results, ancestorPolicy{ID: a.id, Type: a.kind, Policy: policy})
	}
	return yamlEncode(results)
}

func runProjectSetIamPolicy(cmd *cobra.Command, args []string) error {
	policy, err := parsePolicyFile(args[1])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy.Version = 3
	updated, err := svc.Projects.SetIamPolicy(projectResourceName(args[0]), &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for project [%s].\n", strings.TrimPrefix(args[0], "projects/"))
	return yamlEncode(updated)
}

func runProjectAddIamBinding(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resource := projectResourceName(args[0])
	policy, err := svc.Projects.GetIamPolicy(resource, &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	addBindingToPolicy(policy, flagProjectIamRole, flagProjectIamMember,
		rmBuildCondition(flagProjectIamCondExpr, flagProjectIamCondTitle, flagProjectIamCondDesc))
	policy.Version = 3

	updated, err := svc.Projects.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for project [%s].\n", strings.TrimPrefix(args[0], "projects/"))
	return yamlEncode(updated)
}

func runProjectRemoveIamBinding(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resource := projectResourceName(args[0])
	policy, err := svc.Projects.GetIamPolicy(resource, &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	if !removeBindingFromPolicy(policy, flagProjectIamRole, flagProjectIamMember,
		rmBuildCondition(flagProjectIamCondExpr, flagProjectIamCondTitle, flagProjectIamCondDesc),
		flagProjectIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagProjectIamRole, flagProjectIamMember)
	}

	updated, err := svc.Projects.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for project [%s].\n", strings.TrimPrefix(args[0], "projects/"))
	return yamlEncode(updated)
}
