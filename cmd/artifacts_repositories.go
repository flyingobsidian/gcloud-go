package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts repositories (#1079) ---

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

func (s *artifactsCmds) initArtifactsRepositories() {
	all := []*cobra.Command{
		s.artifactsRepositoriesCreateCmd, s.artifactsRepositoriesDeleteCmd, s.artifactsRepositoriesDescribeCmd,
		s.artifactsRepositoriesListCmd, s.artifactsRepositoriesUpdateCmd,
		s.artifactsRepositoriesGetIamCmd, s.artifactsRepositoriesSetIamCmd,
		s.artifactsRepositoriesAddIamCmd, s.artifactsRepositoriesRemoveIamCmd,
		s.artifactsRepositoriesListCleanupPoliciesCmd, s.artifactsRepositoriesSetCleanupPoliciesCmd,
		s.artifactsRepositoriesDeleteCleanupPoliciesCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagArtRepoLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtRepoFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	s.artifactsRepositoriesCreateCmd.Flags().StringVar(&s.flagArtRepoConfigFile, "config-file", "", "YAML/JSON body for the Repository (required)")
	s.artifactsRepositoriesCreateCmd.Flags().BoolVar(&s.flagArtRepoAsync, "async", false, "Return without waiting for operation")
	s.artifactsRepositoriesCreateCmd.Flags().StringVar(&s.flagArtRepoMode, "mode", "",
		"Repository mode override: NONE, STANDARD-REPOSITORY, VIRTUAL-REPOSITORY, "+
			"REMOTE-REPOSITORY, or CONNECTOR-REPOSITORY (fetches from upstream without caching). "+
			"When set, overrides mode in --config-file.")
	_ = s.artifactsRepositoriesCreateCmd.MarkFlagRequired("config-file")

	s.artifactsRepositoriesDeleteCmd.Flags().BoolVar(&s.flagArtRepoAsync, "async", false, "Return without waiting for operation")
	s.artifactsRepositoriesListCmd.Flags().StringVar(&s.flagArtRepoFilter, "filter", "", "Filter expression")
	s.artifactsRepositoriesListCmd.Flags().Int64Var(&s.flagArtRepoPageSize, "page-size", 0, "Page size")

	s.artifactsRepositoriesUpdateCmd.Flags().StringVar(&s.flagArtRepoConfigFile, "config-file", "", "YAML/JSON body for the Repository update (required)")
	s.artifactsRepositoriesUpdateCmd.Flags().StringVar(&s.flagArtRepoMask, "update-mask", "", "Fields to update")
	_ = s.artifactsRepositoriesUpdateCmd.MarkFlagRequired("config-file")

	s.artifactsRepositoriesSetCleanupPoliciesCmd.Flags().StringVar(&s.flagArtRepoConfigFile, "config-file", "", "YAML/JSON file mapping policy id -> CleanupPolicy (required)")
	_ = s.artifactsRepositoriesSetCleanupPoliciesCmd.MarkFlagRequired("config-file")

	s.artifactsRepositoriesDeleteCleanupPoliciesCmd.Flags().StringSliceVar(&s.flagArtRepoPolicyNames, "policy-names", nil, "Cleanup policy IDs to delete (required)")
	_ = s.artifactsRepositoriesDeleteCleanupPoliciesCmd.MarkFlagRequired("policy-names")

	artIamMemberFlags(s.artifactsRepositoriesAddIamCmd, &s.flagArtRepoIamMember, &s.flagArtRepoIamRole,
		&s.flagArtRepoIamCondExpr, &s.flagArtRepoIamCondTitle, &s.flagArtRepoIamCondDesc)
	artIamMemberFlags(s.artifactsRepositoriesRemoveIamCmd, &s.flagArtRepoIamMember, &s.flagArtRepoIamRole,
		&s.flagArtRepoIamCondExpr, &s.flagArtRepoIamCondTitle, &s.flagArtRepoIamCondDesc)
	s.artifactsRepositoriesRemoveIamCmd.Flags().BoolVar(&s.flagArtRepoAllConds, "all", false, "Match bindings for the role across all conditions")

	s.artifactsRepositoriesCmd.AddCommand(all...)
	s.artifactsCmd.AddCommand(s.artifactsRepositoriesCmd)
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
func (s *artifactsCmds) artRepoApplyRegistryURI(r *artifactregistry.Repository) {
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
	fmt.Fprintf(s.inv.stderr, "Registry URL: %s\n", r.RegistryUri)
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

func (s *artifactsCmds) runArtifactsRepositoriesCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.Repository{}
	if err := loadYAMLOrJSONInto(s.flagArtRepoConfigFile, body); err != nil {
		return err
	}
	if err := artRepoApplyMode(body, s.flagArtRepoMode); err != nil {
		return err
	}
	parent := artLocationParent(project, s.flagArtRepoLocation)
	op, err := svc.Projects.Locations.Repositories.Create(parent, body).
		RepositoryId(args[0]).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating repository: %w", err)
	}
	if s.flagArtRepoAsync {
		fmt.Fprintf(s.inv.stderr, "Create started. Operation: %s\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagArtRepoFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	// Best-effort: fetch the freshly-created repo so we can print the Registry
	// URL, matching gcloud-python's post-create hook.
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	if repo, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do(); err == nil {
		s.artRepoApplyRegistryURI(repo)
		return emitFormattedTo(s.inv.stdout, repo, s.flagArtRepoFormat)
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	op, err := svc.Projects.Locations.Repositories.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting repository: %w", err)
	}
	if s.flagArtRepoAsync {
		fmt.Fprintf(s.inv.stderr, "Delete started. Operation: %s\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagArtRepoFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	r, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing repository: %w", err)
	}
	s.artRepoApplyRegistryURI(r)
	return emitFormattedTo(s.inv.stdout, r, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artLocationParent(project, s.flagArtRepoLocation)
	var all []*artifactregistry.Repository
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.List(parent).Context(ctx)
		if s.flagArtRepoFilter != "" {
			call = call.Filter(s.flagArtRepoFilter)
		}
		if s.flagArtRepoPageSize > 0 {
			call = call.PageSize(s.flagArtRepoPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.Repository{}
	if err := loadYAMLOrJSONInto(s.flagArtRepoConfigFile, body); err != nil {
		return err
	}
	mask := s.flagArtRepoMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	call := svc.Projects.Locations.Repositories.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating repository: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	pol, err := svc.Projects.Locations.Repositories.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, pol, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	policy := &artifactregistry.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	req := &artifactregistry.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.Repositories.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.artUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	pol, err := svc.Projects.Locations.Repositories.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := artIamBuildCondition(s.flagArtRepoIamCondExpr, s.flagArtRepoIamCondTitle, s.flagArtRepoIamCondDesc)
	artIamAddBinding(pol, s.flagArtRepoIamRole, s.flagArtRepoIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &artifactregistry.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.Repositories.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	s.artUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	pol, err := svc.Projects.Locations.Repositories.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := artIamBuildCondition(s.flagArtRepoIamCondExpr, s.flagArtRepoIamCondTitle, s.flagArtRepoIamCondDesc)
	if !artIamRemoveBinding(pol, s.flagArtRepoIamRole, s.flagArtRepoIamMember, cond, s.flagArtRepoAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &artifactregistry.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.Repositories.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	s.artUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesListCleanup(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	r, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting repository: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, r.CleanupPolicies, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesSetCleanup(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	r, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting repository: %w", err)
	}
	policies := map[string]artifactregistry.CleanupPolicy{}
	if err := loadYAMLOrJSONInto(s.flagArtRepoConfigFile, &policies); err != nil {
		return err
	}
	r.CleanupPolicies = policies
	out, err := svc.Projects.Locations.Repositories.Patch(name, r).
		UpdateMask("cleanupPolicies").
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting cleanup policies: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRepoFormat)
}

func (s *artifactsCmds) runArtifactsRepositoriesDeleteCleanup(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRepoName(project, s.flagArtRepoLocation, args[0])
	r, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting repository: %w", err)
	}
	if r.CleanupPolicies == nil {
		r.CleanupPolicies = map[string]artifactregistry.CleanupPolicy{}
	}
	for _, id := range s.flagArtRepoPolicyNames {
		delete(r.CleanupPolicies, id)
	}
	out, err := svc.Projects.Locations.Repositories.Patch(name, r).
		UpdateMask("cleanupPolicies").
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cleanup policies: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRepoFormat)
}
