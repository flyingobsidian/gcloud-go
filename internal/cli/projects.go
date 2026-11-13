package cli

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gfilter"
	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
)

// --- gcloud projects (#375, #515) ---

func newProjectsCmd(inv *invocation) *cobra.Command {
	c := &cobra.Command{
		Use:   "projects",
		Short: "Manage Google Cloud projects",
	}
	c.AddCommand(
		newProjectIamBindingCmd(inv, true),
		newProjectCreateCmd(inv),
		newProjectDeleteCmd(inv),
		newProjectDescribeCmd(inv),
		newProjectGetAncestorsCmd(inv),
		newProjectGetAncestorsIamPolicyCmd(inv),
		newProjectGetIamPolicyCmd(inv),
		newProjectListCmd(inv),
		newProjectIamBindingCmd(inv, false),
		newProjectSetIamPolicyCmd(inv),
		newProjectUndeleteCmd(inv),
		newProjectUpdateCmd(inv),
	)
	return c
}

func newProjectDescribeCmd(inv *invocation) *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "describe PROJECT_ID_OR_NUMBER",
		Short: "Show metadata for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectDescribe(c.Context(), inv, args[0], format)
		},
	}
	c.Flags().StringVar(&format, "format", "", "Output format: yaml (default), json, csv(FIELDS), table(FIELDS), text(FIELDS), value(FIELDS), config(FIELDS), get(FIELD)")
	return c
}

type projectCreateOptions struct {
	name, folder, org string
	labels            map[string]string
}

func newProjectCreateCmd(inv *invocation) *cobra.Command {
	var o projectCreateOptions
	c := &cobra.Command{
		Use:   "create PROJECT_ID",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectCreate(c.Context(), inv, args[0], o)
		},
	}
	c.Flags().StringVar(&o.name, "name", "", "Display name for the new project (defaults to the project ID)")
	c.Flags().StringVar(&o.folder, "folder", "", "Parent folder ID (mutually exclusive with --organization)")
	c.Flags().StringVar(&o.org, "organization", "", "Parent organization ID (mutually exclusive with --folder)")
	c.Flags().StringToStringVar(&o.labels, "labels", nil, "Comma-separated key=value label pairs")
	return c
}

func newProjectDeleteCmd(inv *invocation) *cobra.Command {
	c := &cobra.Command{
		Use:   "delete PROJECT_ID_OR_NUMBER",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectDelete(c.Context(), inv, args[0])
		},
	}
	addProjectRecommendFlag(c)
	return c
}

func newProjectUndeleteCmd(inv *invocation) *cobra.Command {
	return &cobra.Command{
		Use:   "undelete PROJECT_ID_OR_NUMBER",
		Short: "Undelete a project marked for deletion",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectUndelete(c.Context(), inv, args[0])
		},
	}
}

func newProjectUpdateCmd(inv *invocation) *cobra.Command {
	var name string
	c := &cobra.Command{
		Use:   "update PROJECT_ID_OR_NUMBER",
		Short: "Update the name of a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectUpdate(c.Context(), inv, args[0], name)
		},
	}
	c.Flags().StringVar(&name, "name", "", "New display name for the project (required)")
	c.MarkFlagRequired("name")
	return c
}

type projectListOptions struct {
	folder, org, filter, format string
	pageSize, limit             int64
	uri, showDeleted            bool
}

func newProjectListCmd(inv *invocation) *cobra.Command {
	var o projectListOptions
	c := &cobra.Command{
		Use:   "list",
		Short: "List projects accessible by the active account",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectList(c.Context(), inv, o)
		},
	}
	c.Flags().StringVar(&o.folder, "folder", "", "List projects under this parent folder")
	c.Flags().StringVar(&o.org, "organization", "", "List projects under this parent organization")
	c.Flags().StringVar(&o.filter, "filter", "", "Search query expression (e.g. state:ACTIVE)")
	c.Flags().Int64Var(&o.pageSize, "page-size", 0, "Page size for API pagination")
	c.Flags().Int64Var(&o.limit, "limit", 0, "Maximum number of projects to list (0 = no limit)")
	c.Flags().StringVar(&o.format, "format", "", "Output format (json, yaml, or table)")
	c.Flags().BoolVar(&o.uri, "uri", false, "Print resource names only")
	c.Flags().BoolVar(&o.showDeleted, "show-deleted", false, "Include deleted projects (only when listing under a parent)")
	return c
}

func newProjectGetAncestorsCmd(inv *invocation) *cobra.Command {
	return &cobra.Command{
		Use:   "get-ancestors PROJECT_ID_OR_NUMBER",
		Short: "Get the ancestors for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectGetAncestors(c.Context(), inv, args[0])
		},
	}
}

func newProjectGetIamPolicyCmd(inv *invocation) *cobra.Command {
	return &cobra.Command{
		Use:   "get-iam-policy PROJECT_ID_OR_NUMBER",
		Short: "Get IAM policy for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectGetIamPolicy(c.Context(), inv, args[0])
		},
	}
}

func newProjectGetAncestorsIamPolicyCmd(inv *invocation) *cobra.Command {
	return &cobra.Command{
		Use:   "get-ancestors-iam-policy PROJECT_ID_OR_NUMBER",
		Short: "Get IAM policies for a project and its ancestors",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectGetAncestorsIamPolicy(c.Context(), inv, args[0])
		},
	}
}

func newProjectSetIamPolicyCmd(inv *invocation) *cobra.Command {
	return &cobra.Command{
		Use:   "set-iam-policy PROJECT_ID_OR_NUMBER POLICY_FILE",
		Short: "Set IAM policy for a project",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectSetIamPolicy(c.Context(), inv, args[0], args[1])
		},
	}
}

// iamBindingOptions are the flags of add-iam-policy-binding and
// remove-iam-policy-binding.
type iamBindingOptions struct {
	member, role                  string
	condExpr, condTitle, condDesc string
	allConditions                 bool
}

// condition returns the binding condition given by the --condition-* flags,
// or nil if none is set.
func (o *iamBindingOptions) condition() *crm.Expr {
	return rmBuildCondition(o.condExpr, o.condTitle, o.condDesc)
}

// addIamBindingFlags binds the add-/remove-iam-policy-binding flags of c to
// o; remove adds --all.
func addIamBindingFlags(c *cobra.Command, o *iamBindingOptions, remove bool) {
	c.Flags().StringVar(&o.member, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
	c.Flags().StringVar(&o.role, "role", "", "IAM role to bind (e.g. roles/browser) (required)")
	c.Flags().StringVar(&o.condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(&o.condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(&o.condDesc, "condition-description", "", "Description for a conditional binding")
	c.MarkFlagRequired("member")
	c.MarkFlagRequired("role")
	if remove {
		c.Flags().BoolVar(&o.allConditions, "all", false, "Remove the member from all bindings for the role, regardless of condition")
	}
}

// newProjectIamBindingCmd returns add-iam-policy-binding when add is set,
// else remove-iam-policy-binding.
func newProjectIamBindingCmd(inv *invocation, add bool) *cobra.Command {
	var o iamBindingOptions
	c := &cobra.Command{
		Use:   "add-iam-policy-binding PROJECT_ID_OR_NUMBER",
		Short: "Add IAM policy binding for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProjectAddIamBinding(c.Context(), inv, args[0], o)
		},
	}
	if !add {
		c.Use = "remove-iam-policy-binding PROJECT_ID_OR_NUMBER"
		c.Short = "Remove IAM policy binding for a project"
		c.RunE = func(c *cobra.Command, args []string) error {
			return runProjectRemoveIamBinding(c.Context(), inv, args[0], o)
		}
	}
	addIamBindingFlags(c, &o, !add)
	if !add {
		addProjectRecommendFlag(c)
	}
	return c
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

func runProjectDescribe(ctx context.Context, inv *invocation, id, format string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	project, err := svc.Projects.Get(projectResourceName(id)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project: %w", err)
	}
	return emitFormattedTo(inv.stdout, project, format)
}

func runProjectCreate(ctx context.Context, inv *invocation, id string, o projectCreateOptions) error {
	parent, err := projectParent(o.folder, o.org)
	if err != nil {
		return err
	}

	displayName := o.name
	if displayName == "" {
		displayName = id
	}

	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	op, err := svc.Projects.Create(&crm.Project{
		ProjectId:   id,
		DisplayName: displayName,
		Parent:      parent,
		Labels:      o.labels,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating project: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Create project in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runProjectDelete(ctx context.Context, inv *invocation, id string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Delete(projectResourceName(id)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Delete project in progress (operation: %s).\n", op.Name)
	fmt.Fprintf(inv.stderr, "You can undo this operation for a limited period by running:\n"+
		"    gcloud projects undelete %s\n", strings.TrimPrefix(id, "projects/"))
	return inv.yaml(op)
}

func runProjectUndelete(ctx context.Context, inv *invocation, id string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Undelete(projectResourceName(id), &crm.UndeleteProjectRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting project: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Undelete project in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runProjectUpdate(ctx context.Context, inv *invocation, id, name string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Patch(projectResourceName(id), &crm.Project{
		DisplayName: name,
	}).UpdateMask("display_name").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Update project in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runProjectList(ctx context.Context, inv *invocation, o projectListOptions) error {
	parent, err := projectParent(o.folder, o.org)
	if err != nil {
		return err
	}

	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	all, err := listProjects(ctx, svc, parent, o.filter, o.pageSize, o.limit, o.showDeleted)
	if err != nil {
		return err
	}

	if o.uri {
		for _, p := range all {
			fmt.Fprintln(inv.stdout, p.Name)
		}
		return nil
	}

	switch o.format {
	case "json":
		return inv.printJSON(all)
	case "yaml":
		return inv.yaml(all)
	}

	printTableHeadingTo(inv.stdout, len(all), "%-30s %-30s %-20s %s\n", "PROJECT_ID", "NAME", "PROJECT_NUMBER", "STATE")
	for _, p := range all {
		fmt.Fprintf(inv.stdout, "%-30s %-30s %-20s %s\n", p.ProjectId, p.DisplayName, path.Base(p.Name), p.State)
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

func runProjectGetAncestors(ctx context.Context, inv *invocation, id string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	ancestors, err := projectAncestors(ctx, svc, id)
	if err != nil {
		return err
	}

	printTableHeadingTo(inv.stdout, len(ancestors), "%-40s %s\n", "ID", "TYPE")
	for _, a := range ancestors {
		fmt.Fprintf(inv.stdout, "%-40s %s\n", a.id, a.kind)
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

func runProjectGetIamPolicy(ctx context.Context, inv *invocation, id string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.GetIamPolicy(projectResourceName(id), &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return inv.emitWithPipeline(policy)
}

// emitWithPipeline runs the standard gcloud output pipeline (--flatten →
// --filter → --format) against v. When none of those global flags is set, it
// falls back to the yaml dump that runProjectGetIamPolicy used before #1728.
func (inv *invocation) emitWithPipeline(v any) error {
	return emitPipeline(inv.stdout, v, inv.flags.flatten, inv.flags.filter, inv.flags.format)
}

// emitPipeline writes v to w through --flatten, --filter and --format.
func emitPipeline(w io.Writer, v any, flatten []string, filterExpr, format string) error {
	// Fast path: no pipeline modifiers → keep the historical yaml dump so
	// commands that don't opt into fancy output stay byte-identical.
	if len(flatten) == 0 && filterExpr == "" && format == "" {
		return emitFormattedTo(w, v, "yaml")
	}

	records, err := flattenRecords(v, flatten)
	if err != nil {
		return err
	}

	if filterExpr != "" {
		filter, err := gfilter.Compile(filterExpr)
		if err != nil {
			return fmt.Errorf("invalid --filter expression: %w", err)
		}
		kept := records[:0]
		for _, r := range records {
			if filter.Match(r) {
				kept = append(kept, r)
			}
		}
		records = kept
	}

	// If --flatten was applied, hand the record list to emitFormattedTo so
	// table/csv/etc. render one row per record. Without --flatten we pass
	// through the original single value so scalar top-level policies still
	// render sensibly.
	if len(flatten) > 0 {
		return emitFormattedTo(w, records, format)
	}
	return emitFormattedTo(w, v, format)
}

func runProjectGetAncestorsIamPolicy(ctx context.Context, inv *invocation, id string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	ancestors, err := projectAncestors(ctx, svc, id)
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
	return inv.yaml(results)
}

func runProjectSetIamPolicy(ctx context.Context, inv *invocation, id, policyFile string) error {
	policy, err := parsePolicyFile(policyFile)
	if err != nil {
		return err
	}
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	policy.Version = 3
	updated, err := svc.Projects.SetIamPolicy(projectResourceName(id), &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Updated IAM policy for project [%s].\n", strings.TrimPrefix(id, "projects/"))
	return inv.yaml(updated)
}

func runProjectAddIamBinding(ctx context.Context, inv *invocation, id string, o iamBindingOptions) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	resource := projectResourceName(id)
	policy, err := svc.Projects.GetIamPolicy(resource, &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	addBindingToPolicy(policy, o.role, o.member, o.condition())
	policy.Version = 3

	updated, err := svc.Projects.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Updated IAM policy for project [%s].\n", strings.TrimPrefix(id, "projects/"))
	return inv.yaml(updated)
}

func runProjectRemoveIamBinding(ctx context.Context, inv *invocation, id string, o iamBindingOptions) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	resource := projectResourceName(id)
	policy, err := svc.Projects.GetIamPolicy(resource, &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	if !removeBindingFromPolicy(policy, o.role, o.member,
		o.condition(), o.allConditions) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", o.role, o.member)
	}

	updated, err := svc.Projects.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Updated IAM policy for project [%s].\n", strings.TrimPrefix(id, "projects/"))
	return inv.yaml(updated)
}
