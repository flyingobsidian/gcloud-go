package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	sourcerepo "google.golang.org/api/sourcerepo/v1"
)

// --- gcloud source repos (#1154) ---

var sourceReposCmd = &cobra.Command{Use: "repos", Short: "Manage Cloud Source Repositories"}

var (
	flagSourceRepoFormat     string
	flagSourceRepoConfigFile string
	flagSourceRepoUpdateMask string
	flagSourceRepoPageSize   int64
)

var (
	sourceRepoCreateCmd = &cobra.Command{
		Use: "create REPO", Short: "Create a source repository",
		Args: cobra.ExactArgs(1), RunE: runSourceRepoCreate,
	}
	sourceRepoDeleteCmd = &cobra.Command{
		Use: "delete REPO", Short: "Delete a source repository",
		Args: cobra.ExactArgs(1), RunE: runSourceRepoDelete,
	}
	sourceRepoDescribeCmd = &cobra.Command{
		Use: "describe REPO", Short: "Describe a source repository",
		Args: cobra.ExactArgs(1), RunE: runSourceRepoDescribe,
	}
	sourceRepoListCmd = &cobra.Command{
		Use: "list", Short: "List source repositories in the current project",
		Args: cobra.NoArgs, RunE: runSourceRepoList,
	}
	sourceRepoUpdateCmd = &cobra.Command{
		Use: "update REPO", Short: "Update a source repository",
		Args: cobra.ExactArgs(1), RunE: runSourceRepoUpdate,
	}
	sourceRepoGetIamCmd = &cobra.Command{
		Use: "get-iam-policy REPO", Short: "Get the IAM policy for a source repository",
		Args: cobra.ExactArgs(1), RunE: runSourceRepoGetIam,
	}
	sourceRepoSetIamCmd = &cobra.Command{
		Use: "set-iam-policy REPO POLICY_FILE", Short: "Set the IAM policy for a source repository",
		Args: cobra.ExactArgs(2), RunE: runSourceRepoSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		sourceRepoCreateCmd, sourceRepoDeleteCmd, sourceRepoDescribeCmd,
		sourceRepoListCmd, sourceRepoUpdateCmd, sourceRepoGetIamCmd, sourceRepoSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSourceRepoFormat, "format", "", "Output format")
	}
	sourceRepoCreateCmd.Flags().StringVar(&flagSourceRepoConfigFile, "config-file", "", "YAML/JSON file with the Repo body (optional; a minimal repo is created if omitted)")
	sourceRepoListCmd.Flags().Int64Var(&flagSourceRepoPageSize, "page-size", 0, "Maximum results per page")
	sourceRepoUpdateCmd.Flags().StringVar(&flagSourceRepoConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = sourceRepoUpdateCmd.MarkFlagRequired("config-file")
	sourceRepoUpdateCmd.Flags().StringVar(&flagSourceRepoUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	sourceReposCmd.AddCommand(all...)
	sourceCmd.AddCommand(sourceReposCmd)
}

func runSourceRepoCreate(cmd *cobra.Command, args []string) error {
	parent, err := sourceProjectName()
	if err != nil {
		return err
	}
	name, err := sourceRepoName(args[0])
	if err != nil {
		return err
	}
	body := &sourcerepo.Repo{}
	if flagSourceRepoConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSourceRepoConfigFile, body); err != nil {
			return err
		}
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Repos.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating repo: %w", err)
	}
	fmt.Printf("Created repo [%s].\n", args[0])
	return emitFormatted(got, flagSourceRepoFormat)
}

func runSourceRepoDelete(cmd *cobra.Command, args []string) error {
	name, err := sourceRepoName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Repos.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting repo: %w", err)
	}
	fmt.Printf("Deleted repo [%s].\n", args[0])
	return nil
}

func runSourceRepoDescribe(cmd *cobra.Command, args []string) error {
	name, err := sourceRepoName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Repos.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing repo: %w", err)
	}
	return emitFormatted(got, flagSourceRepoFormat)
}

func runSourceRepoList(cmd *cobra.Command, args []string) error {
	parent, err := sourceProjectName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*sourcerepo.Repo
	pageToken := ""
	for {
		call := svc.Projects.Repos.List(parent).Context(ctx)
		if flagSourceRepoPageSize > 0 {
			call = call.PageSize(flagSourceRepoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing repos: %w", err)
		}
		all = append(all, resp.Repos...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSourceRepoFormat)
}

func runSourceRepoUpdate(cmd *cobra.Command, args []string) error {
	name, err := sourceRepoName(args[0])
	if err != nil {
		return err
	}
	body := &sourcerepo.Repo{}
	if err := loadYAMLOrJSONInto(flagSourceRepoConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagSourceRepoUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Repos.Patch(name, &sourcerepo.UpdateRepoRequest{
		Repo:       body,
		UpdateMask: mask,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating repo: %w", err)
	}
	fmt.Printf("Updated repo [%s].\n", args[0])
	return emitFormatted(got, flagSourceRepoFormat)
}

func runSourceRepoGetIam(cmd *cobra.Command, args []string) error {
	name, err := sourceRepoName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Repos.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagSourceRepoFormat)
}

func runSourceRepoSetIam(cmd *cobra.Command, args []string) error {
	name, err := sourceRepoName(args[0])
	if err != nil {
		return err
	}
	policy := &sourcerepo.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Repos.SetIamPolicy(name, &sourcerepo.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for repo [%s].\n", args[0])
	return emitFormatted(updated, flagSourceRepoFormat)
}
