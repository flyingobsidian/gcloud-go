package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	cloudbuild2 "google.golang.org/api/cloudbuild/v2"
)

// --- gcloud builds (#312) ---

func buildsLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func buildsChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// --- v1 op polling ---

func buildsV1WaitOp(ctx context.Context, svc *cloudbuild.Service, op *cloudbuild.Operation) (*cloudbuild.Operation, error) {
	for !op.Done {
		got, err := svc.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (s *buildsCmds) buildsV1FinishOp(ctx context.Context, svc *cloudbuild.Service, op *cloudbuild.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := buildsV1WaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// --- v2 op polling ---

func buildsV2WaitOp(ctx context.Context, svc *cloudbuild2.Service, op *cloudbuild2.Operation) (*cloudbuild2.Operation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (s *buildsCmds) buildsV2FinishOp(ctx context.Context, svc *cloudbuild2.Service, op *cloudbuild2.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := buildsV2WaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// --- connections (v2) ---

// --- repositories (v2, nested under connections) ---

// --- triggers (v1) ---

// --- worker-pools (v1) ---

func (s *buildsCmds) initBuilds() {
	// connections
	connAll := []*cobra.Command{s.buildsConnCreateCmd, s.buildsConnDeleteCmd, s.buildsConnDescribeCmd, s.buildsConnListCmd,
		s.buildsConnUpdateCmd, s.buildsConnGetIamCmd, s.buildsConnSetIamCmd, s.buildsConnAddIamCmd, s.buildsConnRemoveIamCmd}
	for _, c := range connAll {
		c.Flags().StringVar(&s.flagBuildsRegion, "region", "", "Region containing the connection (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.buildsConnCreateCmd, s.buildsConnUpdateCmd} {
		c.Flags().StringVar(&s.flagBuildsConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Connection body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.buildsConnUpdateCmd.Flags().StringVar(&s.flagBuildsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.buildsConnCreateCmd, s.buildsConnDeleteCmd, s.buildsConnUpdateCmd} {
		c.Flags().BoolVar(&s.flagBuildsAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{s.buildsConnDescribeCmd, s.buildsConnListCmd, s.buildsConnGetIamCmd} {
		c.Flags().StringVar(&s.flagBuildsFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.buildsConnAddIamCmd, s.buildsConnRemoveIamCmd} {
		c.Flags().StringVar(&s.flagBuildsIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&s.flagBuildsIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	s.buildsConnectionsCmd.AddCommand(connAll...)
	s.buildsCmd.AddCommand(s.buildsConnectionsCmd)

	// repositories
	repoAll := []*cobra.Command{s.buildsRepoCreateCmd, s.buildsRepoDeleteCmd, s.buildsRepoDescribeCmd, s.buildsRepoListCmd}
	for _, c := range repoAll {
		c.Flags().StringVar(&s.flagBuildsRegion, "region", "", "Region containing the repository (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagBuildsConnection, "connection", "",
			"Connection containing the repository (required)")
		_ = c.MarkFlagRequired("connection")
	}
	s.buildsRepoCreateCmd.Flags().StringVar(&s.flagBuildsConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Repository body (required)")
	_ = s.buildsRepoCreateCmd.MarkFlagRequired("config-file")
	for _, c := range []*cobra.Command{s.buildsRepoCreateCmd, s.buildsRepoDeleteCmd} {
		c.Flags().BoolVar(&s.flagBuildsAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{s.buildsRepoDescribeCmd, s.buildsRepoListCmd} {
		c.Flags().StringVar(&s.flagBuildsFormat, "format", "", "Output format")
	}
	s.buildsRepositoriesCmd.AddCommand(repoAll...)
	s.buildsCmd.AddCommand(s.buildsRepositoriesCmd)

	// triggers
	trigAll := []*cobra.Command{s.buildsTrigCreateCmd, s.buildsTrigDeleteCmd, s.buildsTrigDescribeCmd,
		s.buildsTrigImportCmd, s.buildsTrigListCmd, s.buildsTrigRunCmd, s.buildsTrigUpdateCmd}
	for _, c := range trigAll {
		c.Flags().StringVar(&s.flagBuildsRegion, "region", "global", "Region containing the trigger")
	}
	for _, c := range []*cobra.Command{s.buildsTrigCreateCmd, s.buildsTrigUpdateCmd} {
		c.Flags().StringVar(&s.flagBuildsConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the BuildTrigger body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.buildsTrigImportCmd.Flags().StringVar(&s.flagBuildsConfigFile, "source-file", "",
		"Path to a JSON/YAML file with the BuildTrigger body (required)")
	_ = s.buildsTrigImportCmd.MarkFlagRequired("source-file")
	s.buildsTrigUpdateCmd.Flags().StringVar(&s.flagBuildsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.buildsTrigDescribeCmd, s.buildsTrigListCmd} {
		c.Flags().StringVar(&s.flagBuildsFormat, "format", "", "Output format")
	}
	s.buildsTrigRunCmd.Flags().StringVar(&s.flagBuildsBranch, "branch", "", "Branch to run")
	s.buildsTrigRunCmd.Flags().StringVar(&s.flagBuildsTag, "tag", "", "Tag to run")
	s.buildsTrigRunCmd.Flags().StringVar(&s.flagBuildsCommit, "commit-sha", "", "Commit SHA to run")
	s.buildsTrigRunCmd.Flags().StringVar(&s.flagBuildsRepoName, "repo-name", "", "Repo name to run")
	s.buildsTrigRunCmd.Flags().StringVar(&s.flagBuildsRepoDir, "dir", "", "Repository directory")
	s.buildsTrigRunCmd.Flags().BoolVar(&s.flagBuildsAsync, "async", false, "Return the long-running operation without waiting")
	s.buildsTriggersCmd.AddCommand(trigAll...)
	s.buildsCmd.AddCommand(s.buildsTriggersCmd)

	// worker-pools
	wpAll := []*cobra.Command{s.buildsWPCreateCmd, s.buildsWPDeleteCmd, s.buildsWPDescribeCmd, s.buildsWPListCmd, s.buildsWPUpdateCmd}
	for _, c := range wpAll {
		c.Flags().StringVar(&s.flagBuildsRegion, "region", "", "Region containing the worker pool (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.buildsWPCreateCmd, s.buildsWPUpdateCmd} {
		c.Flags().StringVar(&s.flagBuildsConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the WorkerPool body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.buildsWPUpdateCmd.Flags().StringVar(&s.flagBuildsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.buildsWPCreateCmd, s.buildsWPUpdateCmd} {
		c.Flags().StringVar(&s.flagBuildsWorkerRelease, "worker-release", "",
			"Specifies the release channel or version to use for the build worker image. "+
				"Accepts release channel names (`rapid`, `regular`, `stable`) or specific "+
				"release versions (for example, `2026.09`, `legacy`). See "+
				"https://docs.cloud.google.com/build/docs/release-channels.")
	}
	for _, c := range []*cobra.Command{s.buildsWPCreateCmd, s.buildsWPDeleteCmd, s.buildsWPUpdateCmd} {
		c.Flags().BoolVar(&s.flagBuildsAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{s.buildsWPDescribeCmd, s.buildsWPListCmd} {
		c.Flags().StringVar(&s.flagBuildsFormat, "format", "", "Output format")
	}
	s.buildsWorkerPoolsCmd.AddCommand(wpAll...)
	s.buildsCmd.AddCommand(s.buildsWorkerPoolsCmd)

	// existing stubs for top-level verbs (submit, cancel, etc.) still register below
	for _, name := range []string{"cancel", "describe", "get-default-service-account", "log", "submit"} {
		registerStubCommand(s.buildsCmd, name, "Not yet implemented")
	}

}

// --- connections impl ---

func buildsConnName(id, project, region string) string {
	return buildsChild("connections", id, buildsLocationParent(project, region))
}

func (s *buildsCmds) runBuildsConnCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &cloudbuild2.Connection{}
	if err := loadYAMLOrJSONInto(s.flagBuildsConfigFile, c); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Create(buildsLocationParent(project, s.flagBuildsRegion), c).
		ConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connection: %w", err)
	}
	return s.buildsV2FinishOp(ctx, svc, op, "Create connection", args[0], s.flagBuildsAsync)
}

func (s *buildsCmds) runBuildsConnDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Delete(buildsConnName(args[0], project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connection: %w", err)
	}
	return s.buildsV2FinishOp(ctx, svc, op, "Delete connection", args[0], s.flagBuildsAsync)
}

func (s *buildsCmds) runBuildsConnDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Connections.Get(buildsConnName(args[0], project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBuildsFormat)
}

func (s *buildsCmds) runBuildsConnList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Connections.List(buildsLocationParent(project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing connections: %w", err)
	}
	if s.flagBuildsFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Connections, s.flagBuildsFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Connections), "%-40s\n", "NAME")
	for _, c := range resp.Connections {
		fmt.Fprintln(s.inv.stdout, path.Base(c.Name))
	}
	return nil
}

func (s *buildsCmds) runBuildsConnUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &cloudbuild2.Connection{}
	if err := loadYAMLOrJSONInto(s.flagBuildsConfigFile, c); err != nil {
		return err
	}
	mask := s.flagBuildsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Patch(buildsConnName(args[0], project, s.flagBuildsRegion), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connection: %w", err)
	}
	return s.buildsV2FinishOp(ctx, svc, op, "Update connection", args[0], s.flagBuildsAsync)
}

func (s *buildsCmds) runBuildsConnGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Connections.GetIamPolicy(buildsConnName(args[0], project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagBuildsFormat)
}

func (s *buildsCmds) runBuildsConnSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &cloudbuild2.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Connections.SetIamPolicy(buildsConnName(args[0], project, s.flagBuildsRegion), &cloudbuild2.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *buildsCmds) runBuildsConnAddIam(cmd *cobra.Command, args []string) error {
	return s.buildsConnModifyIam(args[0], func(p *cloudbuild2.Policy) {
		for _, b := range p.Bindings {
			if b.Role == s.flagBuildsIamRole {
				for _, m := range b.Members {
					if m == s.flagBuildsIamMember {
						return
					}
				}
				b.Members = append(b.Members, s.flagBuildsIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &cloudbuild2.Binding{Role: s.flagBuildsIamRole, Members: []string{s.flagBuildsIamMember}})
	})
}

func (s *buildsCmds) runBuildsConnRemoveIam(cmd *cobra.Command, args []string) error {
	return s.buildsConnModifyIam(args[0], func(p *cloudbuild2.Policy) {
		for _, b := range p.Bindings {
			if b.Role != s.flagBuildsIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != s.flagBuildsIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func (s *buildsCmds) buildsConnModifyIam(name string, mutate func(*cloudbuild2.Policy)) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	resource := buildsConnName(name, project, s.flagBuildsRegion)
	policy, err := svc.Projects.Locations.Connections.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Connections.SetIamPolicy(resource, &cloudbuild2.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- repositories impl ---

func buildsRepoParent(project, region, connection string) string {
	return buildsConnName(connection, project, region)
}

func buildsRepoName(id, project, region, connection string) string {
	return buildsChild("repositories", id, buildsRepoParent(project, region, connection))
}

func (s *buildsCmds) runBuildsRepoCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	r := &cloudbuild2.Repository{}
	if err := loadYAMLOrJSONInto(s.flagBuildsConfigFile, r); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Repositories.Create(
		buildsRepoParent(project, s.flagBuildsRegion, s.flagBuildsConnection), r).
		RepositoryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating repository: %w", err)
	}
	return s.buildsV2FinishOp(ctx, svc, op, "Create repository", args[0], s.flagBuildsAsync)
}

func (s *buildsCmds) runBuildsRepoDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Repositories.Delete(
		buildsRepoName(args[0], project, s.flagBuildsRegion, s.flagBuildsConnection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting repository: %w", err)
	}
	return s.buildsV2FinishOp(ctx, svc, op, "Delete repository", args[0], s.flagBuildsAsync)
}

func (s *buildsCmds) runBuildsRepoDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Connections.Repositories.Get(
		buildsRepoName(args[0], project, s.flagBuildsRegion, s.flagBuildsConnection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing repository: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBuildsFormat)
}

func (s *buildsCmds) runBuildsRepoList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuildV2(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Connections.Repositories.List(
		buildsRepoParent(project, s.flagBuildsRegion, s.flagBuildsConnection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing repositories: %w", err)
	}
	if s.flagBuildsFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Repositories, s.flagBuildsFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Repositories), "%-40s %s\n", "NAME", "REMOTE_URI")
	for _, r := range resp.Repositories {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(r.Name), r.RemoteUri)
	}
	return nil
}

// --- triggers impl (v1) ---

func buildsTrigName(id, project, region string) string {
	return buildsChild("triggers", id, buildsLocationParent(project, region))
}

func (s *buildsCmds) runBuildsTrigCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	t := &cloudbuild.BuildTrigger{}
	if err := loadYAMLOrJSONInto(s.flagBuildsConfigFile, t); err != nil {
		return err
	}
	if t.Name == "" {
		t.Name = args[0]
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Triggers.Create(buildsLocationParent(project, s.flagBuildsRegion), t).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating trigger: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *buildsCmds) runBuildsTrigDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Triggers.Delete(buildsTrigName(args[0], project, s.flagBuildsRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting trigger: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted trigger [%s].\n", args[0])
	return nil
}

func (s *buildsCmds) runBuildsTrigDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Triggers.Get(buildsTrigName(args[0], project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing trigger: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBuildsFormat)
}

func (s *buildsCmds) runBuildsTrigImport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	t := &cloudbuild.BuildTrigger{}
	if err := loadYAMLOrJSONInto(s.flagBuildsConfigFile, t); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Triggers.Create(buildsLocationParent(project, s.flagBuildsRegion), t).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing trigger: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *buildsCmds) runBuildsTrigList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Triggers.List(buildsLocationParent(project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing triggers: %w", err)
	}
	if s.flagBuildsFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Triggers, s.flagBuildsFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Triggers), "%-40s %s\n", "NAME", "DESCRIPTION")
	for _, t := range resp.Triggers {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", t.Name, t.Description)
	}
	return nil
}

func (s *buildsCmds) runBuildsTrigRun(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &cloudbuild.RunBuildTriggerRequest{}
	if s.flagBuildsBranch != "" || s.flagBuildsTag != "" || s.flagBuildsCommit != "" ||
		s.flagBuildsRepoName != "" || s.flagBuildsRepoDir != "" {
		req.Source = &cloudbuild.RepoSource{
			BranchName: s.flagBuildsBranch,
			TagName:    s.flagBuildsTag,
			CommitSha:  s.flagBuildsCommit,
			RepoName:   s.flagBuildsRepoName,
			Dir:        s.flagBuildsRepoDir,
		}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Triggers.Run(buildsTrigName(args[0], project, s.flagBuildsRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running trigger: %w", err)
	}
	return s.buildsV1FinishOp(ctx, svc, op, "Run trigger", args[0], s.flagBuildsAsync)
}

func (s *buildsCmds) runBuildsTrigUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	t := &cloudbuild.BuildTrigger{}
	if err := loadYAMLOrJSONInto(s.flagBuildsConfigFile, t); err != nil {
		return err
	}
	mask := s.flagBuildsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(t))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Triggers.Patch(buildsTrigName(args[0], project, s.flagBuildsRegion), t).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating trigger: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- worker-pools impl (v1) ---

func buildsWPName(id, project, region string) string {
	return buildsChild("workerPools", id, buildsLocationParent(project, region))
}

// buildsWPWorkerReleaseMask is the update-mask path for --worker-release.
const buildsWPWorkerReleaseMask = "privatePoolV1Config.workerConfig.workerRelease"

// applyBuildsWPWorkerRelease sets the worker release on the pool body when
// --worker-release was given, overriding any value from --config-file.
func (s *buildsCmds) applyBuildsWPWorkerRelease(cmd *cobra.Command, wp *cloudbuild.WorkerPool) bool {
	if !cmd.Flags().Changed("worker-release") {
		return false
	}
	if wp.PrivatePoolV1Config == nil {
		wp.PrivatePoolV1Config = &cloudbuild.PrivatePoolV1Config{}
	}
	if wp.PrivatePoolV1Config.WorkerConfig == nil {
		wp.PrivatePoolV1Config.WorkerConfig = &cloudbuild.WorkerConfig{}
	}
	wp.PrivatePoolV1Config.WorkerConfig.WorkerRelease = s.flagBuildsWorkerRelease
	return true
}

// buildsWPUpdateMask returns the update mask: --update-mask or the body's
// populated top-level fields, plus the worker-release path when that flag is
// set and not already covered by privatePoolV1Config.
func buildsWPUpdateMask(fields []string, workerRelease bool) string {
	if workerRelease {
		covered := false
		for _, f := range fields {
			switch f {
			case "privatePoolV1Config", "privatePoolV1Config.workerConfig", buildsWPWorkerReleaseMask:
				covered = true
			}
		}
		if !covered {
			fields = append(fields, buildsWPWorkerReleaseMask)
		}
	}
	return joinMask(fields)
}

func (s *buildsCmds) runBuildsWPCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	wp := &cloudbuild.WorkerPool{}
	if err := loadYAMLOrJSONInto(s.flagBuildsConfigFile, wp); err != nil {
		return err
	}
	s.applyBuildsWPWorkerRelease(cmd, wp)
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Create(buildsLocationParent(project, s.flagBuildsRegion), wp).
		WorkerPoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating worker pool: %w", err)
	}
	return s.buildsV1FinishOp(ctx, svc, op, "Create worker pool", args[0], s.flagBuildsAsync)
}

func (s *buildsCmds) runBuildsWPDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Delete(buildsWPName(args[0], project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting worker pool: %w", err)
	}
	return s.buildsV1FinishOp(ctx, svc, op, "Delete worker pool", args[0], s.flagBuildsAsync)
}

func (s *buildsCmds) runBuildsWPDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkerPools.Get(buildsWPName(args[0], project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing worker pool: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBuildsFormat)
}

func (s *buildsCmds) runBuildsWPList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.WorkerPools.List(buildsLocationParent(project, s.flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing worker pools: %w", err)
	}
	if s.flagBuildsFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.WorkerPools, s.flagBuildsFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.WorkerPools), "%-40s %s\n", "NAME", "STATE")
	for _, wp := range resp.WorkerPools {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(wp.Name), wp.State)
	}
	return nil
}

func (s *buildsCmds) runBuildsWPUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	wp := &cloudbuild.WorkerPool{}
	if err := loadYAMLOrJSONInto(s.flagBuildsConfigFile, wp); err != nil {
		return err
	}
	var fields []string
	if s.flagBuildsUpdateMask != "" {
		fields = strings.Split(s.flagBuildsUpdateMask, ",")
	} else {
		fields = nonEmptyJSONFields(wp)
	}
	mask := buildsWPUpdateMask(fields, s.applyBuildsWPWorkerRelease(cmd, wp))
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBuild(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Patch(buildsWPName(args[0], project, s.flagBuildsRegion), wp).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating worker pool: %w", err)
	}
	return s.buildsV1FinishOp(ctx, svc, op, "Update worker pool", args[0], s.flagBuildsAsync)
}

// buildsCmds holds the commands and flag values of one invocation's "builds"
// command group, so invocations don't share state (#2070).
type buildsCmds struct {
	inv *invocation

	buildsCmd            *cobra.Command
	flagBuildsRegion     string
	flagBuildsConfigFile string
	flagBuildsUpdateMask string
	flagBuildsFormat     string
	flagBuildsAsync      bool
	flagBuildsConnection string
	flagBuildsIamMember  string
	flagBuildsIamRole    string
	// triggers run
	flagBuildsBranch      string
	flagBuildsTag         string
	flagBuildsCommit      string
	flagBuildsRepoName    string
	flagBuildsRepoDir     string
	buildsConnectionsCmd  *cobra.Command
	buildsConnCreateCmd   *cobra.Command
	buildsConnDeleteCmd   *cobra.Command
	buildsConnDescribeCmd *cobra.Command
	buildsConnListCmd     *cobra.Command
	buildsConnUpdateCmd   *cobra.Command
	buildsConnGetIamCmd   *cobra.Command
	buildsConnSetIamCmd   *cobra.Command
	buildsConnAddIamCmd   *cobra.Command
	buildsRepositoriesCmd *cobra.Command
	buildsRepoCreateCmd   *cobra.Command
	buildsRepoDeleteCmd   *cobra.Command
	buildsRepoDescribeCmd *cobra.Command
	buildsRepoListCmd     *cobra.Command
	buildsTriggersCmd     *cobra.Command
	buildsTrigCreateCmd   *cobra.Command
	buildsTrigDeleteCmd   *cobra.Command
	buildsTrigDescribeCmd *cobra.Command
	buildsTrigImportCmd   *cobra.Command
	buildsTrigListCmd     *cobra.Command
	buildsTrigRunCmd      *cobra.Command
	buildsTrigUpdateCmd   *cobra.Command
	buildsWorkerPoolsCmd  *cobra.Command
	buildsWPCreateCmd     *cobra.Command
	buildsWPDeleteCmd     *cobra.Command
	buildsWPDescribeCmd   *cobra.Command
	buildsWPListCmd       *cobra.Command
	buildsWPUpdateCmd     *cobra.Command
	// remove-iam-policy-binding for connections
	buildsConnRemoveIamCmd *cobra.Command
	// flagBuildsWorkerRelease sets privatePoolV1Config.workerConfig.workerRelease
	// on `worker-pools create|update` (gcloud-python 587.0.0).
	flagBuildsWorkerRelease string
	flagBuildsListRegion    string
	flagBuildsListOngoing   bool
	flagBuildsListLimit     int64
	flagBuildsListPageSize  int64
}

// newBuildsCmd returns the "builds" command group for inv.
func newBuildsCmd(inv *invocation) *cobra.Command {
	return newBuildsCmds(inv).buildsCmd
}

// newBuildsCmds builds the "builds" command group for inv and returns its state.
func newBuildsCmds(inv *invocation) *buildsCmds {
	s := &buildsCmds{inv: inv}
	s.buildsCmd = &cobra.Command{Use: "builds", Short: "Manage Cloud Build"}
	s.buildsConnectionsCmd = &cobra.Command{Use: "connections", Short: "Manage Cloud Build connections"}
	s.buildsConnCreateCmd = &cobra.Command{
		Use: "create CONNECTION", Short: "Create a connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsConnCreate,
	}
	s.buildsConnDeleteCmd = &cobra.Command{
		Use: "delete CONNECTION", Short: "Delete a connection",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsConnDelete,
	}
	s.buildsConnDescribeCmd = &cobra.Command{
		Use: "describe CONNECTION", Short: "Describe a connection",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsConnDescribe,
	}
	s.buildsConnListCmd = &cobra.Command{
		Use: "list", Short: "List connections in a region",
		Args: cobra.NoArgs, RunE: s.runBuildsConnList,
	}
	s.buildsConnUpdateCmd = &cobra.Command{
		Use: "update CONNECTION", Short: "Update a connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsConnUpdate,
	}
	s.buildsConnGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CONNECTION", Short: "Get the IAM policy for a connection",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsConnGetIam,
	}
	s.buildsConnSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CONNECTION POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: s.runBuildsConnSetIam,
	}
	s.buildsConnAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding CONNECTION", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsConnAddIam,
	}
	s.buildsRepositoriesCmd = &cobra.Command{Use: "repositories", Short: "Manage repositories on a Cloud Build connection"}
	s.buildsRepoCreateCmd = &cobra.Command{
		Use: "create REPOSITORY", Short: "Create a repository under a connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsRepoCreate,
	}
	s.buildsRepoDeleteCmd = &cobra.Command{
		Use: "delete REPOSITORY", Short: "Delete a repository",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsRepoDelete,
	}
	s.buildsRepoDescribeCmd = &cobra.Command{
		Use: "describe REPOSITORY", Short: "Describe a repository",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsRepoDescribe,
	}
	s.buildsRepoListCmd = &cobra.Command{
		Use: "list", Short: "List repositories on a connection",
		Args: cobra.NoArgs, RunE: s.runBuildsRepoList,
	}
	s.buildsTriggersCmd = &cobra.Command{Use: "triggers", Short: "Manage Cloud Build triggers"}
	s.buildsTrigCreateCmd = &cobra.Command{
		Use: "create TRIGGER", Short: "Create a build trigger from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsTrigCreate,
	}
	s.buildsTrigDeleteCmd = &cobra.Command{
		Use: "delete TRIGGER", Short: "Delete a build trigger",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsTrigDelete,
	}
	s.buildsTrigDescribeCmd = &cobra.Command{
		Use: "describe TRIGGER", Short: "Describe a build trigger",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsTrigDescribe,
	}
	s.buildsTrigImportCmd = &cobra.Command{
		Use: "import", Short: "Import a build trigger from a --source-file",
		Args: cobra.NoArgs, RunE: s.runBuildsTrigImport,
	}
	s.buildsTrigListCmd = &cobra.Command{
		Use: "list", Short: "List build triggers in a region",
		Args: cobra.NoArgs, RunE: s.runBuildsTrigList,
	}
	s.buildsTrigRunCmd = &cobra.Command{
		Use: "run TRIGGER", Short: "Run a build trigger",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsTrigRun,
	}
	s.buildsTrigUpdateCmd = &cobra.Command{
		Use: "update TRIGGER", Short: "Update a build trigger from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsTrigUpdate,
	}
	s.buildsWorkerPoolsCmd = &cobra.Command{Use: "worker-pools", Short: "Manage Cloud Build worker pools"}
	s.buildsWPCreateCmd = &cobra.Command{
		Use: "create WORKER_POOL", Short: "Create a worker pool from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsWPCreate,
	}
	s.buildsWPDeleteCmd = &cobra.Command{
		Use: "delete WORKER_POOL", Short: "Delete a worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsWPDelete,
	}
	s.buildsWPDescribeCmd = &cobra.Command{
		Use: "describe WORKER_POOL", Short: "Describe a worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsWPDescribe,
	}
	s.buildsWPListCmd = &cobra.Command{
		Use: "list", Short: "List worker pools in a region",
		Args: cobra.NoArgs, RunE: s.runBuildsWPList,
	}
	s.buildsWPUpdateCmd = &cobra.Command{
		Use: "update WORKER_POOL", Short: "Update a worker pool from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsWPUpdate,
	}
	s.buildsConnRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding CONNECTION", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: s.runBuildsConnRemoveIam,
	}
	s.initBuilds()
	s.initBuildsList()
	return s
}
