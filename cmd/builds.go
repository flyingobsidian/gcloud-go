package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	cloudbuild2 "google.golang.org/api/cloudbuild/v2"
)

// --- gcloud builds (#312) ---

var buildsCmd = &cobra.Command{Use: "builds", Short: "Manage Cloud Build"}

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

func buildsV1FinishOp(ctx context.Context, svc *cloudbuild.Service, op *cloudbuild.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := buildsV1WaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
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

func buildsV2FinishOp(ctx context.Context, svc *cloudbuild2.Service, op *cloudbuild2.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := buildsV2WaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

var (
	flagBuildsRegion     string
	flagBuildsConfigFile string
	flagBuildsUpdateMask string
	flagBuildsFormat     string
	flagBuildsAsync      bool
	flagBuildsConnection string
	flagBuildsIamMember  string
	flagBuildsIamRole    string
	// triggers run
	flagBuildsBranch   string
	flagBuildsTag      string
	flagBuildsCommit   string
	flagBuildsRepoName string
	flagBuildsRepoDir  string
)

// --- connections (v2) ---

var buildsConnectionsCmd = &cobra.Command{Use: "connections", Short: "Manage Cloud Build connections"}

var (
	buildsConnCreateCmd = &cobra.Command{
		Use: "create CONNECTION", Short: "Create a connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBuildsConnCreate,
	}
	buildsConnDeleteCmd = &cobra.Command{
		Use: "delete CONNECTION", Short: "Delete a connection",
		Args: cobra.ExactArgs(1), RunE: runBuildsConnDelete,
	}
	buildsConnDescribeCmd = &cobra.Command{
		Use: "describe CONNECTION", Short: "Describe a connection",
		Args: cobra.ExactArgs(1), RunE: runBuildsConnDescribe,
	}
	buildsConnListCmd = &cobra.Command{
		Use: "list", Short: "List connections in a region",
		Args: cobra.NoArgs, RunE: runBuildsConnList,
	}
	buildsConnUpdateCmd = &cobra.Command{
		Use: "update CONNECTION", Short: "Update a connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBuildsConnUpdate,
	}
	buildsConnGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CONNECTION", Short: "Get the IAM policy for a connection",
		Args: cobra.ExactArgs(1), RunE: runBuildsConnGetIam,
	}
	buildsConnSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CONNECTION POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runBuildsConnSetIam,
	}
	buildsConnAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding CONNECTION", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runBuildsConnAddIam,
	}
)

// --- repositories (v2, nested under connections) ---

var buildsRepositoriesCmd = &cobra.Command{Use: "repositories", Short: "Manage repositories on a Cloud Build connection"}

var (
	buildsRepoCreateCmd = &cobra.Command{
		Use: "create REPOSITORY", Short: "Create a repository under a connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBuildsRepoCreate,
	}
	buildsRepoDeleteCmd = &cobra.Command{
		Use: "delete REPOSITORY", Short: "Delete a repository",
		Args: cobra.ExactArgs(1), RunE: runBuildsRepoDelete,
	}
	buildsRepoDescribeCmd = &cobra.Command{
		Use: "describe REPOSITORY", Short: "Describe a repository",
		Args: cobra.ExactArgs(1), RunE: runBuildsRepoDescribe,
	}
	buildsRepoListCmd = &cobra.Command{
		Use: "list", Short: "List repositories on a connection",
		Args: cobra.NoArgs, RunE: runBuildsRepoList,
	}
)

// --- triggers (v1) ---

var buildsTriggersCmd = &cobra.Command{Use: "triggers", Short: "Manage Cloud Build triggers"}

var (
	buildsTrigCreateCmd = &cobra.Command{
		Use: "create TRIGGER", Short: "Create a build trigger from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBuildsTrigCreate,
	}
	buildsTrigDeleteCmd = &cobra.Command{
		Use: "delete TRIGGER", Short: "Delete a build trigger",
		Args: cobra.ExactArgs(1), RunE: runBuildsTrigDelete,
	}
	buildsTrigDescribeCmd = &cobra.Command{
		Use: "describe TRIGGER", Short: "Describe a build trigger",
		Args: cobra.ExactArgs(1), RunE: runBuildsTrigDescribe,
	}
	buildsTrigImportCmd = &cobra.Command{
		Use: "import", Short: "Import a build trigger from a --source-file",
		Args: cobra.NoArgs, RunE: runBuildsTrigImport,
	}
	buildsTrigListCmd = &cobra.Command{
		Use: "list", Short: "List build triggers in a region",
		Args: cobra.NoArgs, RunE: runBuildsTrigList,
	}
	buildsTrigRunCmd = &cobra.Command{
		Use: "run TRIGGER", Short: "Run a build trigger",
		Args: cobra.ExactArgs(1), RunE: runBuildsTrigRun,
	}
	buildsTrigUpdateCmd = &cobra.Command{
		Use: "update TRIGGER", Short: "Update a build trigger from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBuildsTrigUpdate,
	}
)

// --- worker-pools (v1) ---

var buildsWorkerPoolsCmd = &cobra.Command{Use: "worker-pools", Short: "Manage Cloud Build worker pools"}

var (
	buildsWPCreateCmd = &cobra.Command{
		Use: "create WORKER_POOL", Short: "Create a worker pool from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBuildsWPCreate,
	}
	buildsWPDeleteCmd = &cobra.Command{
		Use: "delete WORKER_POOL", Short: "Delete a worker pool",
		Args: cobra.ExactArgs(1), RunE: runBuildsWPDelete,
	}
	buildsWPDescribeCmd = &cobra.Command{
		Use: "describe WORKER_POOL", Short: "Describe a worker pool",
		Args: cobra.ExactArgs(1), RunE: runBuildsWPDescribe,
	}
	buildsWPListCmd = &cobra.Command{
		Use: "list", Short: "List worker pools in a region",
		Args: cobra.NoArgs, RunE: runBuildsWPList,
	}
	buildsWPUpdateCmd = &cobra.Command{
		Use: "update WORKER_POOL", Short: "Update a worker pool from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBuildsWPUpdate,
	}
)

// remove-iam-policy-binding for connections
var buildsConnRemoveIamCmd = &cobra.Command{
	Use: "remove-iam-policy-binding CONNECTION", Short: "Remove an IAM binding",
	Args: cobra.ExactArgs(1), RunE: runBuildsConnRemoveIam,
}

func init() {
	// connections
	connAll := []*cobra.Command{buildsConnCreateCmd, buildsConnDeleteCmd, buildsConnDescribeCmd, buildsConnListCmd,
		buildsConnUpdateCmd, buildsConnGetIamCmd, buildsConnSetIamCmd, buildsConnAddIamCmd, buildsConnRemoveIamCmd}
	for _, c := range connAll {
		c.Flags().StringVar(&flagBuildsRegion, "region", "", "Region containing the connection (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{buildsConnCreateCmd, buildsConnUpdateCmd} {
		c.Flags().StringVar(&flagBuildsConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Connection body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	buildsConnUpdateCmd.Flags().StringVar(&flagBuildsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{buildsConnCreateCmd, buildsConnDeleteCmd, buildsConnUpdateCmd} {
		c.Flags().BoolVar(&flagBuildsAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{buildsConnDescribeCmd, buildsConnListCmd, buildsConnGetIamCmd} {
		c.Flags().StringVar(&flagBuildsFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{buildsConnAddIamCmd, buildsConnRemoveIamCmd} {
		c.Flags().StringVar(&flagBuildsIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagBuildsIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	buildsConnectionsCmd.AddCommand(connAll...)
	buildsCmd.AddCommand(buildsConnectionsCmd)

	// repositories
	repoAll := []*cobra.Command{buildsRepoCreateCmd, buildsRepoDeleteCmd, buildsRepoDescribeCmd, buildsRepoListCmd}
	for _, c := range repoAll {
		c.Flags().StringVar(&flagBuildsRegion, "region", "", "Region containing the repository (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagBuildsConnection, "connection", "",
			"Connection containing the repository (required)")
		_ = c.MarkFlagRequired("connection")
	}
	buildsRepoCreateCmd.Flags().StringVar(&flagBuildsConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Repository body (required)")
	_ = buildsRepoCreateCmd.MarkFlagRequired("config-file")
	for _, c := range []*cobra.Command{buildsRepoCreateCmd, buildsRepoDeleteCmd} {
		c.Flags().BoolVar(&flagBuildsAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{buildsRepoDescribeCmd, buildsRepoListCmd} {
		c.Flags().StringVar(&flagBuildsFormat, "format", "", "Output format")
	}
	buildsRepositoriesCmd.AddCommand(repoAll...)
	buildsCmd.AddCommand(buildsRepositoriesCmd)

	// triggers
	trigAll := []*cobra.Command{buildsTrigCreateCmd, buildsTrigDeleteCmd, buildsTrigDescribeCmd,
		buildsTrigImportCmd, buildsTrigListCmd, buildsTrigRunCmd, buildsTrigUpdateCmd}
	for _, c := range trigAll {
		c.Flags().StringVar(&flagBuildsRegion, "region", "global", "Region containing the trigger")
	}
	for _, c := range []*cobra.Command{buildsTrigCreateCmd, buildsTrigUpdateCmd} {
		c.Flags().StringVar(&flagBuildsConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the BuildTrigger body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	buildsTrigImportCmd.Flags().StringVar(&flagBuildsConfigFile, "source-file", "",
		"Path to a JSON/YAML file with the BuildTrigger body (required)")
	_ = buildsTrigImportCmd.MarkFlagRequired("source-file")
	buildsTrigUpdateCmd.Flags().StringVar(&flagBuildsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{buildsTrigDescribeCmd, buildsTrigListCmd} {
		c.Flags().StringVar(&flagBuildsFormat, "format", "", "Output format")
	}
	buildsTrigRunCmd.Flags().StringVar(&flagBuildsBranch, "branch", "", "Branch to run")
	buildsTrigRunCmd.Flags().StringVar(&flagBuildsTag, "tag", "", "Tag to run")
	buildsTrigRunCmd.Flags().StringVar(&flagBuildsCommit, "commit-sha", "", "Commit SHA to run")
	buildsTrigRunCmd.Flags().StringVar(&flagBuildsRepoName, "repo-name", "", "Repo name to run")
	buildsTrigRunCmd.Flags().StringVar(&flagBuildsRepoDir, "dir", "", "Repository directory")
	buildsTrigRunCmd.Flags().BoolVar(&flagBuildsAsync, "async", false, "Return the long-running operation without waiting")
	buildsTriggersCmd.AddCommand(trigAll...)
	buildsCmd.AddCommand(buildsTriggersCmd)

	// worker-pools
	wpAll := []*cobra.Command{buildsWPCreateCmd, buildsWPDeleteCmd, buildsWPDescribeCmd, buildsWPListCmd, buildsWPUpdateCmd}
	for _, c := range wpAll {
		c.Flags().StringVar(&flagBuildsRegion, "region", "", "Region containing the worker pool (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{buildsWPCreateCmd, buildsWPUpdateCmd} {
		c.Flags().StringVar(&flagBuildsConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the WorkerPool body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	buildsWPUpdateCmd.Flags().StringVar(&flagBuildsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{buildsWPCreateCmd, buildsWPDeleteCmd, buildsWPUpdateCmd} {
		c.Flags().BoolVar(&flagBuildsAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{buildsWPDescribeCmd, buildsWPListCmd} {
		c.Flags().StringVar(&flagBuildsFormat, "format", "", "Output format")
	}
	buildsWorkerPoolsCmd.AddCommand(wpAll...)
	buildsCmd.AddCommand(buildsWorkerPoolsCmd)

	// existing stubs for top-level verbs (submit, cancel, etc.) still register below
	for _, name := range []string{"cancel", "describe", "get-default-service-account", "list", "log", "submit"} {
		registerStubCommand(buildsCmd, name, "Not yet implemented")
	}

	rootCmd.AddCommand(buildsCmd)
}

// --- connections impl ---

func buildsConnName(id, project, region string) string {
	return buildsChild("connections", id, buildsLocationParent(project, region))
}

func runBuildsConnCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &cloudbuild2.Connection{}
	if err := loadYAMLOrJSONInto(flagBuildsConfigFile, c); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Create(buildsLocationParent(project, flagBuildsRegion), c).
		ConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connection: %w", err)
	}
	return buildsV2FinishOp(ctx, svc, op, "Create connection", args[0], flagBuildsAsync)
}

func runBuildsConnDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Delete(buildsConnName(args[0], project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connection: %w", err)
	}
	return buildsV2FinishOp(ctx, svc, op, "Delete connection", args[0], flagBuildsAsync)
}

func runBuildsConnDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Connections.Get(buildsConnName(args[0], project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connection: %w", err)
	}
	return emitFormatted(got, flagBuildsFormat)
}

func runBuildsConnList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Connections.List(buildsLocationParent(project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing connections: %w", err)
	}
	if flagBuildsFormat != "" {
		return emitFormatted(resp.Connections, flagBuildsFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, c := range resp.Connections {
		fmt.Println(path.Base(c.Name))
	}
	return nil
}

func runBuildsConnUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &cloudbuild2.Connection{}
	if err := loadYAMLOrJSONInto(flagBuildsConfigFile, c); err != nil {
		return err
	}
	mask := flagBuildsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Patch(buildsConnName(args[0], project, flagBuildsRegion), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connection: %w", err)
	}
	return buildsV2FinishOp(ctx, svc, op, "Update connection", args[0], flagBuildsAsync)
}

func runBuildsConnGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Connections.GetIamPolicy(buildsConnName(args[0], project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagBuildsFormat)
}

func runBuildsConnSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &cloudbuild2.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Connections.SetIamPolicy(buildsConnName(args[0], project, flagBuildsRegion), &cloudbuild2.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runBuildsConnAddIam(cmd *cobra.Command, args []string) error {
	return buildsConnModifyIam(args[0], func(p *cloudbuild2.Policy) {
		for _, b := range p.Bindings {
			if b.Role == flagBuildsIamRole {
				for _, m := range b.Members {
					if m == flagBuildsIamMember {
						return
					}
				}
				b.Members = append(b.Members, flagBuildsIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &cloudbuild2.Binding{Role: flagBuildsIamRole, Members: []string{flagBuildsIamMember}})
	})
}

func runBuildsConnRemoveIam(cmd *cobra.Command, args []string) error {
	return buildsConnModifyIam(args[0], func(p *cloudbuild2.Policy) {
		for _, b := range p.Bindings {
			if b.Role != flagBuildsIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != flagBuildsIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func buildsConnModifyIam(name string, mutate func(*cloudbuild2.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := buildsConnName(name, project, flagBuildsRegion)
	policy, err := svc.Projects.Locations.Connections.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Connections.SetIamPolicy(resource, &cloudbuild2.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// --- repositories impl ---

func buildsRepoParent(project, region, connection string) string {
	return buildsConnName(connection, project, region)
}

func buildsRepoName(id, project, region, connection string) string {
	return buildsChild("repositories", id, buildsRepoParent(project, region, connection))
}

func runBuildsRepoCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	r := &cloudbuild2.Repository{}
	if err := loadYAMLOrJSONInto(flagBuildsConfigFile, r); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Repositories.Create(
		buildsRepoParent(project, flagBuildsRegion, flagBuildsConnection), r).
		RepositoryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating repository: %w", err)
	}
	return buildsV2FinishOp(ctx, svc, op, "Create repository", args[0], flagBuildsAsync)
}

func runBuildsRepoDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Repositories.Delete(
		buildsRepoName(args[0], project, flagBuildsRegion, flagBuildsConnection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting repository: %w", err)
	}
	return buildsV2FinishOp(ctx, svc, op, "Delete repository", args[0], flagBuildsAsync)
}

func runBuildsRepoDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Connections.Repositories.Get(
		buildsRepoName(args[0], project, flagBuildsRegion, flagBuildsConnection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing repository: %w", err)
	}
	return emitFormatted(got, flagBuildsFormat)
}

func runBuildsRepoList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Connections.Repositories.List(
		buildsRepoParent(project, flagBuildsRegion, flagBuildsConnection)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing repositories: %w", err)
	}
	if flagBuildsFormat != "" {
		return emitFormatted(resp.Repositories, flagBuildsFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "REMOTE_URI")
	for _, r := range resp.Repositories {
		fmt.Printf("%-40s %s\n", path.Base(r.Name), r.RemoteUri)
	}
	return nil
}

// --- triggers impl (v1) ---

func buildsTrigName(id, project, region string) string {
	return buildsChild("triggers", id, buildsLocationParent(project, region))
}

func runBuildsTrigCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	t := &cloudbuild.BuildTrigger{}
	if err := loadYAMLOrJSONInto(flagBuildsConfigFile, t); err != nil {
		return err
	}
	if t.Name == "" {
		t.Name = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Triggers.Create(buildsLocationParent(project, flagBuildsRegion), t).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating trigger: %w", err)
	}
	return emitFormatted(got, "")
}

func runBuildsTrigDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Triggers.Delete(buildsTrigName(args[0], project, flagBuildsRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting trigger: %w", err)
	}
	fmt.Printf("Deleted trigger [%s].\n", args[0])
	return nil
}

func runBuildsTrigDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Triggers.Get(buildsTrigName(args[0], project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing trigger: %w", err)
	}
	return emitFormatted(got, flagBuildsFormat)
}

func runBuildsTrigImport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	t := &cloudbuild.BuildTrigger{}
	if err := loadYAMLOrJSONInto(flagBuildsConfigFile, t); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Triggers.Create(buildsLocationParent(project, flagBuildsRegion), t).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing trigger: %w", err)
	}
	return emitFormatted(got, "")
}

func runBuildsTrigList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Triggers.List(buildsLocationParent(project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing triggers: %w", err)
	}
	if flagBuildsFormat != "" {
		return emitFormatted(resp.Triggers, flagBuildsFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DESCRIPTION")
	for _, t := range resp.Triggers {
		fmt.Printf("%-40s %s\n", t.Name, t.Description)
	}
	return nil
}

func runBuildsTrigRun(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &cloudbuild.RunBuildTriggerRequest{}
	if flagBuildsBranch != "" || flagBuildsTag != "" || flagBuildsCommit != "" ||
		flagBuildsRepoName != "" || flagBuildsRepoDir != "" {
		req.Source = &cloudbuild.RepoSource{
			BranchName: flagBuildsBranch,
			TagName:    flagBuildsTag,
			CommitSha:  flagBuildsCommit,
			RepoName:   flagBuildsRepoName,
			Dir:        flagBuildsRepoDir,
		}
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Triggers.Run(buildsTrigName(args[0], project, flagBuildsRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running trigger: %w", err)
	}
	return buildsV1FinishOp(ctx, svc, op, "Run trigger", args[0], flagBuildsAsync)
}

func runBuildsTrigUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	t := &cloudbuild.BuildTrigger{}
	if err := loadYAMLOrJSONInto(flagBuildsConfigFile, t); err != nil {
		return err
	}
	mask := flagBuildsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(t))
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Triggers.Patch(buildsTrigName(args[0], project, flagBuildsRegion), t).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating trigger: %w", err)
	}
	return emitFormatted(got, "")
}

// --- worker-pools impl (v1) ---

func buildsWPName(id, project, region string) string {
	return buildsChild("workerPools", id, buildsLocationParent(project, region))
}

func runBuildsWPCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	wp := &cloudbuild.WorkerPool{}
	if err := loadYAMLOrJSONInto(flagBuildsConfigFile, wp); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Create(buildsLocationParent(project, flagBuildsRegion), wp).
		WorkerPoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating worker pool: %w", err)
	}
	return buildsV1FinishOp(ctx, svc, op, "Create worker pool", args[0], flagBuildsAsync)
}

func runBuildsWPDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Delete(buildsWPName(args[0], project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting worker pool: %w", err)
	}
	return buildsV1FinishOp(ctx, svc, op, "Delete worker pool", args[0], flagBuildsAsync)
}

func runBuildsWPDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkerPools.Get(buildsWPName(args[0], project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing worker pool: %w", err)
	}
	return emitFormatted(got, flagBuildsFormat)
}

func runBuildsWPList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.WorkerPools.List(buildsLocationParent(project, flagBuildsRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing worker pools: %w", err)
	}
	if flagBuildsFormat != "" {
		return emitFormatted(resp.WorkerPools, flagBuildsFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, wp := range resp.WorkerPools {
		fmt.Printf("%-40s %s\n", path.Base(wp.Name), wp.State)
	}
	return nil
}

func runBuildsWPUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	wp := &cloudbuild.WorkerPool{}
	if err := loadYAMLOrJSONInto(flagBuildsConfigFile, wp); err != nil {
		return err
	}
	mask := flagBuildsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(wp))
	}
	ctx := context.Background()
	svc, err := gcp.CloudBuildService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkerPools.Patch(buildsWPName(args[0], project, flagBuildsRegion), wp).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating worker pool: %w", err)
	}
	return buildsV1FinishOp(ctx, svc, op, "Update worker pool", args[0], flagBuildsAsync)
}
