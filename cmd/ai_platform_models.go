package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform models (#984) ---

var aiPlatformModelsCmd = &cobra.Command{Use: "models", Short: "Manage AI Platform models"}

var (
	flagAIPlatformModelsFormat     string
	flagAIPlatformModelsConfigFile string
	flagAIPlatformModelsPageSize   int64
	flagAIPlatformModelsFilter     string
	flagAIPlatformModelsUpdateMask string

	flagAIPlatformModelsIamMember   string
	flagAIPlatformModelsIamRole     string
	flagAIPlatformModelsIamCondExpr string
	flagAIPlatformModelsIamCondT    string
	flagAIPlatformModelsIamCondD    string
	flagAIPlatformModelsIamAllCond  bool
)

var (
	aiPlatformModelsCreateCmd = &cobra.Command{
		Use: "create MODEL", Short: "Create an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformModelsCreate,
	}
	aiPlatformModelsDeleteCmd = &cobra.Command{
		Use: "delete MODEL", Short: "Delete an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformModelsDelete,
	}
	aiPlatformModelsDescribeCmd = &cobra.Command{
		Use: "describe MODEL", Short: "Describe an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformModelsDescribe,
	}
	aiPlatformModelsListCmd = &cobra.Command{
		Use: "list", Short: "List AI Platform models",
		Args: cobra.NoArgs, RunE: runAIPlatformModelsList,
	}
	aiPlatformModelsUpdateCmd = &cobra.Command{
		Use: "update MODEL", Short: "Update an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformModelsUpdate,
	}
	aiPlatformModelsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy MODEL", Short: "Get the IAM policy for an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformModelsGetIam,
	}
	aiPlatformModelsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy MODEL POLICY_FILE", Short: "Set the IAM policy for an AI Platform model",
		Args: cobra.ExactArgs(2), RunE: runAIPlatformModelsSetIam,
	}
	aiPlatformModelsAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding MODEL", Short: "Add an IAM policy binding to an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformModelsAddIam,
	}
	aiPlatformModelsRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding MODEL", Short: "Remove an IAM policy binding from an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformModelsRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		aiPlatformModelsCreateCmd, aiPlatformModelsDeleteCmd, aiPlatformModelsDescribeCmd,
		aiPlatformModelsListCmd, aiPlatformModelsUpdateCmd, aiPlatformModelsGetIamCmd,
		aiPlatformModelsSetIamCmd, aiPlatformModelsAddIamCmd, aiPlatformModelsRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIPlatformModelsFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{aiPlatformModelsCreateCmd, aiPlatformModelsUpdateCmd} {
		c.Flags().StringVar(&flagAIPlatformModelsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Model body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	aiPlatformModelsUpdateCmd.Flags().StringVar(&flagAIPlatformModelsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")

	aiPlatformModelsListCmd.Flags().Int64Var(&flagAIPlatformModelsPageSize, "page-size", 0, "Maximum results per page")
	aiPlatformModelsListCmd.Flags().StringVar(&flagAIPlatformModelsFilter, "filter", "", "List filter expression")

	for _, c := range []*cobra.Command{aiPlatformModelsAddIamCmd, aiPlatformModelsRemoveIamCmd} {
		mlIamFlags(c, &flagAIPlatformModelsIamMember, &flagAIPlatformModelsIamRole,
			&flagAIPlatformModelsIamCondExpr, &flagAIPlatformModelsIamCondT, &flagAIPlatformModelsIamCondD)
	}
	aiPlatformModelsRemoveIamCmd.Flags().BoolVar(&flagAIPlatformModelsIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	aiPlatformModelsCmd.AddCommand(all...)
	aiPlatformCmd.AddCommand(aiPlatformModelsCmd)
}

func runAIPlatformModelsCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Model{}
	if err := loadYAMLOrJSONInto(flagAIPlatformModelsConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Models.Create(mlProjectPath(project), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating model: %w", err)
	}
	fmt.Printf("Created model [%s].\n", args[0])
	return emitFormatted(got, flagAIPlatformModelsFormat)
}

func runAIPlatformModelsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Models.Delete(mlModelName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting model: %w", err)
	}
	fmt.Printf("Delete request issued for model [%s].\n", args[0])
	return emitFormatted(op, flagAIPlatformModelsFormat)
}

func runAIPlatformModelsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Models.Get(mlModelName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing model: %w", err)
	}
	return emitFormatted(got, flagAIPlatformModelsFormat)
}

func runAIPlatformModelsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*ml.GoogleCloudMlV1__Model
	pageToken := ""
	for {
		call := svc.Projects.Models.List(mlProjectPath(project)).Context(ctx)
		if flagAIPlatformModelsPageSize > 0 {
			call = call.PageSize(flagAIPlatformModelsPageSize)
		}
		if flagAIPlatformModelsFilter != "" {
			call = call.Filter(flagAIPlatformModelsFilter)
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
	return emitFormatted(all, flagAIPlatformModelsFormat)
}

func runAIPlatformModelsUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Model{}
	if err := loadYAMLOrJSONInto(flagAIPlatformModelsConfigFile, body); err != nil {
		return err
	}
	mask := flagAIPlatformModelsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
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
	fmt.Printf("Update request issued for model [%s].\n", args[0])
	return emitFormatted(op, flagAIPlatformModelsFormat)
}

func runAIPlatformModelsGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Models.GetIamPolicy(mlModelName(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagAIPlatformModelsFormat)
}

func runAIPlatformModelsSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &ml.GoogleIamV1__Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Models.SetIamPolicy(mlModelName(project, args[0]),
		&ml.GoogleIamV1__SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	mlUpdatedIam(fmt.Sprintf("model [%s]", args[0]))
	return emitFormatted(updated, flagAIPlatformModelsFormat)
}

func runAIPlatformModelsAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := mlModelName(project, args[0])
	policy, err := svc.Projects.Models.GetIamPolicy(resource).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mlAddBinding(policy, flagAIPlatformModelsIamRole, flagAIPlatformModelsIamMember,
		mlBuildCondition(flagAIPlatformModelsIamCondExpr, flagAIPlatformModelsIamCondT, flagAIPlatformModelsIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Models.SetIamPolicy(resource,
		&ml.GoogleIamV1__SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	mlUpdatedIam(fmt.Sprintf("model [%s]", args[0]))
	return emitFormatted(updated, flagAIPlatformModelsFormat)
}

func runAIPlatformModelsRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := mlModelName(project, args[0])
	policy, err := svc.Projects.Models.GetIamPolicy(resource).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !mlRemoveBinding(policy, flagAIPlatformModelsIamRole, flagAIPlatformModelsIamMember,
		mlBuildCondition(flagAIPlatformModelsIamCondExpr, flagAIPlatformModelsIamCondT, flagAIPlatformModelsIamCondD),
		flagAIPlatformModelsIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			flagAIPlatformModelsIamRole, flagAIPlatformModelsIamMember)
	}
	updated, err := svc.Projects.Models.SetIamPolicy(resource,
		&ml.GoogleIamV1__SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	mlUpdatedIam(fmt.Sprintf("model [%s]", args[0]))
	return emitFormatted(updated, flagAIPlatformModelsFormat)
}
