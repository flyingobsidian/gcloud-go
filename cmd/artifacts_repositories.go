package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts repositories (#1079) ---

var artifactsRepositoriesCmd = &cobra.Command{
	Use:   "repositories",
	Short: "Manage Artifact Registry repositories",
}

var (
	flagArtRepoLocation    string
	flagArtRepoFormat      string
	flagArtRepoFilter      string
	flagArtRepoPageSize    int64
	flagArtRepoConfigFile  string
	flagArtRepoMask        string
	flagArtRepoAsync       bool
	flagArtRepoAllConds    bool
	flagArtRepoPolicyNames []string

	flagArtRepoIamMember    string
	flagArtRepoIamRole      string
	flagArtRepoIamCondExpr  string
	flagArtRepoIamCondTitle string
	flagArtRepoIamCondDesc  string

	// --mode passthrough, mirroring gcloud-python's CLI vocabulary.
	// NONE / STANDARD-REPOSITORY / VIRTUAL-REPOSITORY / REMOTE-REPOSITORY /
	// CONNECTOR-REPOSITORY (the last one is REMOTE_REPOSITORY on the wire).
	flagArtRepoMode string
)

// artRepoModeChoices maps the user-facing --mode value to the wire enum value.
// CONNECTOR-REPOSITORY is remote-mode on the wire but presented as a separate
// choice (gcloud-python 577.0.0).
var artRepoModeChoices = map[string]string{
	"NONE":                 "MODE_UNSPECIFIED",
	"STANDARD-REPOSITORY":  "STANDARD_REPOSITORY",
	"VIRTUAL-REPOSITORY":   "VIRTUAL_REPOSITORY",
	"REMOTE-REPOSITORY":    "REMOTE_REPOSITORY",
	"CONNECTOR-REPOSITORY": "REMOTE_REPOSITORY",
}

var artifactsRepositoriesCreateCmd = &cobra.Command{
	Use:   "create REPOSITORY",
	Short: "Create a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesCreate,
}

var artifactsRepositoriesDeleteCmd = &cobra.Command{
	Use:   "delete REPOSITORY",
	Short: "Delete a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesDelete,
}

var artifactsRepositoriesDescribeCmd = &cobra.Command{
	Use:   "describe REPOSITORY",
	Short: "Describe a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesDescribe,
}

var artifactsRepositoriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List repositories in a location",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsRepositoriesList,
}

var artifactsRepositoriesUpdateCmd = &cobra.Command{
	Use:   "update REPOSITORY",
	Short: "Update a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesUpdate,
}

var artifactsRepositoriesGetIamCmd = &cobra.Command{
	Use:   "get-iam-policy REPOSITORY",
	Short: "Get the IAM policy for a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesGetIam,
}

var artifactsRepositoriesSetIamCmd = &cobra.Command{
	Use:   "set-iam-policy REPOSITORY POLICY_FILE",
	Short: "Set the IAM policy for a repository",
	Args:  cobra.ExactArgs(2),
	RunE:  runArtifactsRepositoriesSetIam,
}

var artifactsRepositoriesAddIamCmd = &cobra.Command{
	Use:   "add-iam-policy-binding REPOSITORY",
	Short: "Add an IAM policy binding on a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesAddIam,
}

var artifactsRepositoriesRemoveIamCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding REPOSITORY",
	Short: "Remove an IAM policy binding from a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesRemoveIam,
}

var artifactsRepositoriesListCleanupPoliciesCmd = &cobra.Command{
	Use:   "list-cleanup-policies REPOSITORY",
	Short: "List cleanup policies on a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesListCleanup,
}

var artifactsRepositoriesSetCleanupPoliciesCmd = &cobra.Command{
	Use:   "set-cleanup-policies REPOSITORY",
	Short: "Replace cleanup policies on a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesSetCleanup,
}

var artifactsRepositoriesDeleteCleanupPoliciesCmd = &cobra.Command{
	Use:   "delete-cleanup-policies REPOSITORY",
	Short: "Delete cleanup policies from a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRepositoriesDeleteCleanup,
}

func init() {
	all := []*cobra.Command{
		artifactsRepositoriesCreateCmd, artifactsRepositoriesDeleteCmd, artifactsRepositoriesDescribeCmd,
		artifactsRepositoriesListCmd, artifactsRepositoriesUpdateCmd,
		artifactsRepositoriesGetIamCmd, artifactsRepositoriesSetIamCmd,
		artifactsRepositoriesAddIamCmd, artifactsRepositoriesRemoveIamCmd,
		artifactsRepositoriesListCleanupPoliciesCmd, artifactsRepositoriesSetCleanupPoliciesCmd,
		artifactsRepositoriesDeleteCleanupPoliciesCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagArtRepoLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtRepoFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	artifactsRepositoriesCreateCmd.Flags().StringVar(&flagArtRepoConfigFile, "config-file", "", "YAML/JSON body for the Repository (required)")
	artifactsRepositoriesCreateCmd.Flags().BoolVar(&flagArtRepoAsync, "async", false, "Return without waiting for operation")
	artifactsRepositoriesCreateCmd.Flags().StringVar(&flagArtRepoMode, "mode", "",
		"Repository mode override: NONE, STANDARD-REPOSITORY, VIRTUAL-REPOSITORY, "+
			"REMOTE-REPOSITORY, or CONNECTOR-REPOSITORY (fetches from upstream without caching). "+
			"When set, overrides mode in --config-file.")
	_ = artifactsRepositoriesCreateCmd.MarkFlagRequired("config-file")

	artifactsRepositoriesDeleteCmd.Flags().BoolVar(&flagArtRepoAsync, "async", false, "Return without waiting for operation")
	artifactsRepositoriesListCmd.Flags().StringVar(&flagArtRepoFilter, "filter", "", "Filter expression")
	artifactsRepositoriesListCmd.Flags().Int64Var(&flagArtRepoPageSize, "page-size", 0, "Page size")

	artifactsRepositoriesUpdateCmd.Flags().StringVar(&flagArtRepoConfigFile, "config-file", "", "YAML/JSON body for the Repository update (required)")
	artifactsRepositoriesUpdateCmd.Flags().StringVar(&flagArtRepoMask, "update-mask", "", "Fields to update")
	_ = artifactsRepositoriesUpdateCmd.MarkFlagRequired("config-file")

	artifactsRepositoriesSetCleanupPoliciesCmd.Flags().StringVar(&flagArtRepoConfigFile, "config-file", "", "YAML/JSON file mapping policy id -> CleanupPolicy (required)")
	_ = artifactsRepositoriesSetCleanupPoliciesCmd.MarkFlagRequired("config-file")

	artifactsRepositoriesDeleteCleanupPoliciesCmd.Flags().StringSliceVar(&flagArtRepoPolicyNames, "policy-names", nil, "Cleanup policy IDs to delete (required)")
	_ = artifactsRepositoriesDeleteCleanupPoliciesCmd.MarkFlagRequired("policy-names")

	artIamMemberFlags(artifactsRepositoriesAddIamCmd, &flagArtRepoIamMember, &flagArtRepoIamRole,
		&flagArtRepoIamCondExpr, &flagArtRepoIamCondTitle, &flagArtRepoIamCondDesc)
	artIamMemberFlags(artifactsRepositoriesRemoveIamCmd, &flagArtRepoIamMember, &flagArtRepoIamRole,
		&flagArtRepoIamCondExpr, &flagArtRepoIamCondTitle, &flagArtRepoIamCondDesc)
	artifactsRepositoriesRemoveIamCmd.Flags().BoolVar(&flagArtRepoAllConds, "all", false, "Match bindings for the role across all conditions")

	artifactsRepositoriesCmd.AddCommand(all...)
	artifactsCmd.AddCommand(artifactsRepositoriesCmd)
}

func artRepoName(project, location, raw string) string {
	return artFullName(artLocationParent(project, location)+"/repositories", raw)
}

// artRepoParseName splits a fully-qualified repository name into its
// (project, location, id) parts. It tolerates leading/trailing whitespace and
// returns false when the shape does not match.
func artRepoParseName(name string) (project, location, id string, ok bool) {
	parts := strings.Split(strings.TrimSpace(name), "/")
	if len(parts) != 6 || parts[0] != "projects" || parts[2] != "locations" || parts[4] != "repositories" {
		return "", "", "", false
	}
	return parts[1], parts[3], parts[5], true
}

// artRepoApplyRegistryURI mirrors gcloud-python's
// AddRegistryBaseToRepositoryInfo hook (gcloud-python 575.0.0): if the server
// didn't populate RegistryUri, synthesize it from the repo name + format and
// print "Registry URL: ..." to stderr for the user.
func artRepoApplyRegistryURI(r *artifactregistry.Repository) {
	if r == nil || r.Name == "" || r.RegistryUri != "" {
		return
	}
	project, location, id, ok := artRepoParseName(r.Name)
	if !ok {
		return
	}
	// Colon-separated legacy project IDs are shown as "project/id" in URLs.
	project = strings.ReplaceAll(project, ":", "/")
	r.RegistryUri = fmt.Sprintf("%s-%s.pkg.dev/%s/%s",
		location, strings.ToLower(r.Format), project, id)
	fmt.Fprintf(os.Stderr, "Registry URL: %s\n", r.RegistryUri)
}

// artRepoApplyMode overrides body.Mode using the --mode flag, translating the
// user-facing choice to the wire enum value. CONNECTOR-REPOSITORY is
// REMOTE_REPOSITORY on the wire.
func artRepoApplyMode(body *artifactregistry.Repository, flag string) error {
	if flag == "" {
		return nil
	}
	up := strings.ToUpper(flag)
	wire, ok := artRepoModeChoices[up]
	if !ok {
		return fmt.Errorf("invalid --mode %q: expected one of NONE, STANDARD-REPOSITORY, "+
			"VIRTUAL-REPOSITORY, REMOTE-REPOSITORY, CONNECTOR-REPOSITORY", flag)
	}
	body.Mode = wire
	return nil
}

func runArtifactsRepositoriesCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.Repository{}
	if err := loadYAMLOrJSONInto(flagArtRepoConfigFile, body); err != nil {
		return err
	}
	if err := artRepoApplyMode(body, flagArtRepoMode); err != nil {
		return err
	}
	parent := artLocationParent(project, flagArtRepoLocation)
	op, err := svc.Projects.Locations.Repositories.Create(parent, body).
		RepositoryId(args[0]).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating repository: %w", err)
	}
	if flagArtRepoAsync {
		fmt.Fprintf(os.Stderr, "Create started. Operation: %s\n", op.Name)
		return emitFormatted(op, flagArtRepoFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	// Best-effort: fetch the freshly-created repo so we can print the Registry
	// URL, matching gcloud-python's post-create hook.
	name := artRepoName(project, flagArtRepoLocation, args[0])
	if repo, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do(); err == nil {
		artRepoApplyRegistryURI(repo)
		return emitFormatted(repo, flagArtRepoFormat)
	}
	return emitFormatted(final, flagArtRepoFormat)
}

func runArtifactsRepositoriesDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	op, err := svc.Projects.Locations.Repositories.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting repository: %w", err)
	}
	if flagArtRepoAsync {
		fmt.Fprintf(os.Stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormatted(op, flagArtRepoFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtRepoFormat)
}

func runArtifactsRepositoriesDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	r, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing repository: %w", err)
	}
	artRepoApplyRegistryURI(r)
	return emitFormatted(r, flagArtRepoFormat)
}

func runArtifactsRepositoriesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artLocationParent(project, flagArtRepoLocation)
	var all []*artifactregistry.Repository
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.List(parent).Context(ctx)
		if flagArtRepoFilter != "" {
			call = call.Filter(flagArtRepoFilter)
		}
		if flagArtRepoPageSize > 0 {
			call = call.PageSize(flagArtRepoPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing repositories: %w", err)
		}
		all = append(all, resp.Repositories...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagArtRepoFormat)
}

func runArtifactsRepositoriesUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.Repository{}
	if err := loadYAMLOrJSONInto(flagArtRepoConfigFile, body); err != nil {
		return err
	}
	mask := flagArtRepoMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	call := svc.Projects.Locations.Repositories.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating repository: %w", err)
	}
	return emitFormatted(out, flagArtRepoFormat)
}

func runArtifactsRepositoriesGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	pol, err := svc.Projects.Locations.Repositories.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(pol, flagArtRepoFormat)
}

func runArtifactsRepositoriesSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy := &artifactregistry.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	req := &artifactregistry.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.Repositories.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	artUpdatedIam(name)
	return emitFormatted(out, flagArtRepoFormat)
}

func runArtifactsRepositoriesAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	pol, err := svc.Projects.Locations.Repositories.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := artIamBuildCondition(flagArtRepoIamCondExpr, flagArtRepoIamCondTitle, flagArtRepoIamCondDesc)
	artIamAddBinding(pol, flagArtRepoIamRole, flagArtRepoIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &artifactregistry.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.Repositories.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	artUpdatedIam(name)
	return emitFormatted(out, flagArtRepoFormat)
}

func runArtifactsRepositoriesRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	pol, err := svc.Projects.Locations.Repositories.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := artIamBuildCondition(flagArtRepoIamCondExpr, flagArtRepoIamCondTitle, flagArtRepoIamCondDesc)
	if !artIamRemoveBinding(pol, flagArtRepoIamRole, flagArtRepoIamMember, cond, flagArtRepoAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &artifactregistry.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.Repositories.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	artUpdatedIam(name)
	return emitFormatted(out, flagArtRepoFormat)
}

func runArtifactsRepositoriesListCleanup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	r, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting repository: %w", err)
	}
	return emitFormatted(r.CleanupPolicies, flagArtRepoFormat)
}

func runArtifactsRepositoriesSetCleanup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	r, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting repository: %w", err)
	}
	policies := map[string]artifactregistry.CleanupPolicy{}
	if err := loadYAMLOrJSONInto(flagArtRepoConfigFile, &policies); err != nil {
		return err
	}
	r.CleanupPolicies = policies
	out, err := svc.Projects.Locations.Repositories.Patch(name, r).
		UpdateMask("cleanupPolicies").
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting cleanup policies: %w", err)
	}
	return emitFormatted(out, flagArtRepoFormat)
}

func runArtifactsRepositoriesDeleteCleanup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRepoName(project, flagArtRepoLocation, args[0])
	r, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting repository: %w", err)
	}
	if r.CleanupPolicies == nil {
		r.CleanupPolicies = map[string]artifactregistry.CleanupPolicy{}
	}
	for _, id := range flagArtRepoPolicyNames {
		delete(r.CleanupPolicies, id)
	}
	out, err := svc.Projects.Locations.Repositories.Patch(name, r).
		UpdateMask("cleanupPolicies").
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cleanup policies: %w", err)
	}
	return emitFormatted(out, flagArtRepoFormat)
}
