package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform models (#984) ---

func (s *aiPlatformCmds) initAiPlatformModels() {
	all := []*cobra.Command{
		s.aiPlatformModelsCreateCmd, s.aiPlatformModelsDeleteCmd, s.aiPlatformModelsDescribeCmd,
		s.aiPlatformModelsListCmd, s.aiPlatformModelsUpdateCmd, s.aiPlatformModelsGetIamCmd,
		s.aiPlatformModelsSetIamCmd, s.aiPlatformModelsAddIamCmd, s.aiPlatformModelsRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIPlatformModelsFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.aiPlatformModelsCreateCmd, s.aiPlatformModelsUpdateCmd} {
		c.Flags().StringVar(&s.flagAIPlatformModelsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Model body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.aiPlatformModelsUpdateCmd.Flags().StringVar(&s.flagAIPlatformModelsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")

	s.aiPlatformModelsListCmd.Flags().Int64Var(&s.flagAIPlatformModelsPageSize, "page-size", 0, "Maximum results per page")
	s.aiPlatformModelsListCmd.Flags().StringVar(&s.flagAIPlatformModelsFilter, "filter", "", "List filter expression")

	for _, c := range []*cobra.Command{s.aiPlatformModelsAddIamCmd, s.aiPlatformModelsRemoveIamCmd} {
		mlIamFlags(c, &s.flagAIPlatformModelsIamMember, &s.flagAIPlatformModelsIamRole,
			&s.flagAIPlatformModelsIamCondExpr, &s.flagAIPlatformModelsIamCondT, &s.flagAIPlatformModelsIamCondD)
	}
	s.aiPlatformModelsRemoveIamCmd.Flags().BoolVar(&s.flagAIPlatformModelsIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.aiPlatformModelsCmd.AddCommand(all...)
	s.aiPlatformCmd.AddCommand(s.aiPlatformModelsCmd)
}

func (s *aiPlatformCmds) runAIPlatformModelsCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Model{}
	if err := loadYAMLOrJSONInto(s.flagAIPlatformModelsConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Models.Create(mlProjectPath(project), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating model: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created model [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPlatformModelsFormat)
}

func (s *aiPlatformCmds) runAIPlatformModelsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Models.Delete(mlModelName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting model: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for model [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagAIPlatformModelsFormat)
}

func (s *aiPlatformCmds) runAIPlatformModelsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Models.Get(mlModelName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing model: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPlatformModelsFormat)
}

func (s *aiPlatformCmds) runAIPlatformModelsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	var all []*ml.GoogleCloudMlV1__Model
	pageToken := ""
	for {
		call := svc.Projects.Models.List(mlProjectPath(project)).Context(ctx)
		if s.flagAIPlatformModelsPageSize > 0 {
			call = call.PageSize(s.flagAIPlatformModelsPageSize)
		}
		if s.flagAIPlatformModelsFilter != "" {
			call = call.Filter(s.flagAIPlatformModelsFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing models: %w", err)
		}
		all = append(all, resp.Models...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIPlatformModelsFormat)
}

func (s *aiPlatformCmds) runAIPlatformModelsUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Model{}
	if err := loadYAMLOrJSONInto(s.flagAIPlatformModelsConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAIPlatformModelsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Models.Patch(mlModelName(project, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating model: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for model [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagAIPlatformModelsFormat)
}

func (s *aiPlatformCmds) runAIPlatformModelsGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Models.GetIamPolicy(mlModelName(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagAIPlatformModelsFormat)
}

func (s *aiPlatformCmds) runAIPlatformModelsSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &ml.GoogleIamV1__Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Models.SetIamPolicy(mlModelName(project, args[0]),
		&ml.GoogleIamV1__SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.mlUpdatedIam(fmt.Sprintf("model [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagAIPlatformModelsFormat)
}

func (s *aiPlatformCmds) runAIPlatformModelsAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	resource := mlModelName(project, args[0])
	policy, err := svc.Projects.Models.GetIamPolicy(resource).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mlAddBinding(policy, s.flagAIPlatformModelsIamRole, s.flagAIPlatformModelsIamMember,
		mlBuildCondition(s.flagAIPlatformModelsIamCondExpr, s.flagAIPlatformModelsIamCondT, s.flagAIPlatformModelsIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Models.SetIamPolicy(resource,
		&ml.GoogleIamV1__SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.mlUpdatedIam(fmt.Sprintf("model [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagAIPlatformModelsFormat)
}

func (s *aiPlatformCmds) runAIPlatformModelsRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	resource := mlModelName(project, args[0])
	policy, err := svc.Projects.Models.GetIamPolicy(resource).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !mlRemoveBinding(policy, s.flagAIPlatformModelsIamRole, s.flagAIPlatformModelsIamMember,
		mlBuildCondition(s.flagAIPlatformModelsIamCondExpr, s.flagAIPlatformModelsIamCondT, s.flagAIPlatformModelsIamCondD),
		s.flagAIPlatformModelsIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			s.flagAIPlatformModelsIamRole, s.flagAIPlatformModelsIamMember)
	}
	updated, err := svc.Projects.Models.SetIamPolicy(resource,
		&ml.GoogleIamV1__SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.mlUpdatedIam(fmt.Sprintf("model [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagAIPlatformModelsFormat)
}
