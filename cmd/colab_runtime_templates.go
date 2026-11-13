package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab runtime-templates (#1499) ---

func (s *colabCmds) initColabRuntimeTemplates() {
	all := []*cobra.Command{
		s.colabRTTCreateCmd, s.colabRTTDeleteCmd, s.colabRTTDescribeCmd, s.colabRTTListCmd,
		s.colabRTTGetIamCmd, s.colabRTTSetIamCmd, s.colabRTTAddIamCmd, s.colabRTTRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagColabRTTRegion, "region", "", "Region where the runtime template lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagColabRTTFormat, "format", "", "Output format")
	}
	s.colabRTTCreateCmd.Flags().StringVar(&s.flagColabRTTConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the NotebookRuntimeTemplate body (required)")
	_ = s.colabRTTCreateCmd.MarkFlagRequired("config-file")
	s.colabRTTListCmd.Flags().StringVar(&s.flagColabRTTFilter, "filter", "", "Server-side filter expression")
	s.colabRTTListCmd.Flags().StringVar(&s.flagColabRTTOrderBy, "order-by", "", "Order-by expression")
	s.colabRTTListCmd.Flags().Int64Var(&s.flagColabRTTPageSize, "page-size", 0, "Maximum results per page")
	s.colabRTTListCmd.Flags().StringVar(&s.flagColabRTTReadMask, "read-mask", "", "Field mask for reads")

	for _, c := range []*cobra.Command{s.colabRTTAddIamCmd, s.colabRTTRemoveIamCmd} {
		colabIamMemberFlags(c, &s.flagColabRTTIamMember, &s.flagColabRTTIamRole,
			&s.flagColabRTTIamCondExpr, &s.flagColabRTTIamCondT, &s.flagColabRTTIamCondD)
	}
	s.colabRTTRemoveIamCmd.Flags().BoolVar(&s.flagColabRTTIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.colabRTTCmd.AddCommand(all...)
	s.colabCmd.AddCommand(s.colabRTTCmd)
}

func (s *colabCmds) colabRTTParent() (string, error) {
	return s.colabParent(s.flagColabRTTRegion)
}

func (s *colabCmds) colabRTTName(id string) (string, error) {
	parent, err := s.colabRTTParent()
	if err != nil {
		return "", err
	}
	return colabChild("notebookRuntimeTemplates", id, parent), nil
}

func (s *colabCmds) runColabRTTCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.colabRTTParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1NotebookRuntimeTemplate{}
	if err := loadYAMLOrJSONInto(s.flagColabRTTConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabRTTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimeTemplates.Create(parent, body).NotebookRuntimeTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating runtime template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for runtime template [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabRTTFormat)
}

func (s *colabCmds) runColabRTTDelete(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabRTTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimeTemplates.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting runtime template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for runtime template [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabRTTFormat)
}

func (s *colabCmds) runColabRTTDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabRTTRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NotebookRuntimeTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing runtime template: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagColabRTTFormat)
}

func (s *colabCmds) runColabRTTList(cmd *cobra.Command, args []string) error {
	parent, err := s.colabRTTParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabRTTRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1NotebookRuntimeTemplate
	pageToken := ""
	for {
		call := svc.Projects.Locations.NotebookRuntimeTemplates.List(parent).Context(ctx)
		if s.flagColabRTTFilter != "" {
			call = call.Filter(s.flagColabRTTFilter)
		}
		if s.flagColabRTTOrderBy != "" {
			call = call.OrderBy(s.flagColabRTTOrderBy)
		}
		if s.flagColabRTTPageSize > 0 {
			call = call.PageSize(s.flagColabRTTPageSize)
		}
		if s.flagColabRTTReadMask != "" {
			call = call.ReadMask(s.flagColabRTTReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing runtime templates: %w", err)
		}
		all = append(all, resp.NotebookRuntimeTemplates...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagColabRTTFormat)
}

func (s *colabCmds) runColabRTTGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabRTTRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.NotebookRuntimeTemplates.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagColabRTTFormat)
}

func (s *colabCmds) runColabRTTSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTTName(args[0])
	if err != nil {
		return err
	}
	policy := &aiplatform.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabRTTRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.NotebookRuntimeTemplates.SetIamPolicy(name, &aiplatform.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.writeUpdatedIam(fmt.Sprintf("runtime template [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagColabRTTFormat)
}

func (s *colabCmds) runColabRTTAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabRTTRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.NotebookRuntimeTemplates.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	colabAddBinding(policy, s.flagColabRTTIamRole, s.flagColabRTTIamMember,
		colabBuildCondition(s.flagColabRTTIamCondExpr, s.flagColabRTTIamCondT, s.flagColabRTTIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.NotebookRuntimeTemplates.SetIamPolicy(name, &aiplatform.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.writeUpdatedIam(fmt.Sprintf("runtime template [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagColabRTTFormat)
}

func (s *colabCmds) runColabRTTRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabRTTRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.NotebookRuntimeTemplates.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !colabRemoveBinding(policy, s.flagColabRTTIamRole, s.flagColabRTTIamMember,
		colabBuildCondition(s.flagColabRTTIamCondExpr, s.flagColabRTTIamCondT, s.flagColabRTTIamCondD), s.flagColabRTTIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagColabRTTIamRole, s.flagColabRTTIamMember)
	}
	updated, err := svc.Projects.Locations.NotebookRuntimeTemplates.SetIamPolicy(name, &aiplatform.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.writeUpdatedIam(fmt.Sprintf("runtime template [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagColabRTTFormat)
}
