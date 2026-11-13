package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	sourcerepo "google.golang.org/api/sourcerepo/v1"
)

// --- gcloud source repos (#1154) ---

func (s *sourceCmds) initSourceRepos() {
	all := []*cobra.Command{
		s.sourceRepoCreateCmd, s.sourceRepoDeleteCmd, s.sourceRepoDescribeCmd,
		s.sourceRepoListCmd, s.sourceRepoUpdateCmd, s.sourceRepoGetIamCmd, s.sourceRepoSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSourceRepoFormat, "format", "", "Output format")
	}
	s.sourceRepoCreateCmd.Flags().StringVar(&s.flagSourceRepoConfigFile, "config-file", "", "YAML/JSON file with the Repo body (optional; a minimal repo is created if omitted)")
	s.sourceRepoListCmd.Flags().Int64Var(&s.flagSourceRepoPageSize, "page-size", 0, "Maximum results per page")
	s.sourceRepoUpdateCmd.Flags().StringVar(&s.flagSourceRepoConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.sourceRepoUpdateCmd.MarkFlagRequired("config-file")
	s.sourceRepoUpdateCmd.Flags().StringVar(&s.flagSourceRepoUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.sourceReposCmd.AddCommand(all...)
	s.sourceCmd.AddCommand(s.sourceReposCmd)
}

func (s *sourceCmds) runSourceRepoCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.sourceProjectName()
	if err != nil {
		return err
	}
	name, err := s.sourceRepoName(args[0])
	if err != nil {
		return err
	}
	body := &sourcerepo.Repo{}
	if s.flagSourceRepoConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSourceRepoConfigFile, body); err != nil {
			return err
		}
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Repos.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating repo: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created repo [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagSourceRepoFormat)
}

func (s *sourceCmds) runSourceRepoDelete(cmd *cobra.Command, args []string) error {
	name, err := s.sourceRepoName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Repos.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting repo: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted repo [%s].\n", args[0])
	return nil
}

func (s *sourceCmds) runSourceRepoDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.sourceRepoName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Repos.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing repo: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSourceRepoFormat)
}

func (s *sourceCmds) runSourceRepoList(cmd *cobra.Command, args []string) error {
	parent, err := s.sourceProjectName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
	if err != nil {
		return err
	}
	var all []*sourcerepo.Repo
	pageToken := ""
	for {
		call := svc.Projects.Repos.List(parent).Context(ctx)
		if s.flagSourceRepoPageSize > 0 {
			call = call.PageSize(s.flagSourceRepoPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSourceRepoFormat)
}

func (s *sourceCmds) runSourceRepoUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.sourceRepoName(args[0])
	if err != nil {
		return err
	}
	body := &sourcerepo.Repo{}
	if err := loadYAMLOrJSONInto(s.flagSourceRepoConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagSourceRepoUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Updated repo [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagSourceRepoFormat)
}

func (s *sourceCmds) runSourceRepoGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.sourceRepoName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Repos.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagSourceRepoFormat)
}

func (s *sourceCmds) runSourceRepoSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.sourceRepoName(args[0])
	if err != nil {
		return err
	}
	policy := &sourcerepo.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Repos.SetIamPolicy(name, &sourcerepo.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for repo [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagSourceRepoFormat)
}
