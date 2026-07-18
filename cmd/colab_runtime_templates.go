package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab runtime-templates (#1499) ---

var colabRTTCmd = &cobra.Command{Use: "runtime-templates", Short: "Manage Colab Enterprise runtime templates"}

var (
	flagColabRTTRegion     string
	flagColabRTTFormat     string
	flagColabRTTConfigFile string
	flagColabRTTUpdateMask string
	flagColabRTTFilter     string
	flagColabRTTOrderBy    string
	flagColabRTTPageSize   int64
	flagColabRTTReadMask   string

	flagColabRTTIamMember   string
	flagColabRTTIamRole     string
	flagColabRTTIamCondExpr string
	flagColabRTTIamCondT    string
	flagColabRTTIamCondD    string
	flagColabRTTIamAllCond  bool
)

var (
	colabRTTCreateCmd = &cobra.Command{
		Use: "create RUNTIME_TEMPLATE", Short: "Create a runtime template",
		Args: cobra.ExactArgs(1), RunE: runColabRTTCreate,
	}
	colabRTTDeleteCmd = &cobra.Command{
		Use: "delete RUNTIME_TEMPLATE", Short: "Delete a runtime template",
		Args: cobra.ExactArgs(1), RunE: runColabRTTDelete,
	}
	colabRTTDescribeCmd = &cobra.Command{
		Use: "describe RUNTIME_TEMPLATE", Short: "Describe a runtime template",
		Args: cobra.ExactArgs(1), RunE: runColabRTTDescribe,
	}
	colabRTTListCmd = &cobra.Command{
		Use: "list", Short: "List runtime templates",
		Args: cobra.NoArgs, RunE: runColabRTTList,
	}
	colabRTTGetIamCmd = &cobra.Command{
		Use: "get-iam-policy RUNTIME_TEMPLATE", Short: "Get the IAM policy for a runtime template",
		Args: cobra.ExactArgs(1), RunE: runColabRTTGetIam,
	}
	colabRTTSetIamCmd = &cobra.Command{
		Use: "set-iam-policy RUNTIME_TEMPLATE POLICY_FILE",
		Short: "Set the IAM policy for a runtime template",
		Args:  cobra.ExactArgs(2), RunE: runColabRTTSetIam,
	}
	colabRTTAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding RUNTIME_TEMPLATE",
		Short: "Add an IAM policy binding to a runtime template",
		Args: cobra.ExactArgs(1), RunE: runColabRTTAddIam,
	}
	colabRTTRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding RUNTIME_TEMPLATE",
		Short: "Remove an IAM policy binding from a runtime template",
		Args: cobra.ExactArgs(1), RunE: runColabRTTRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		colabRTTCreateCmd, colabRTTDeleteCmd, colabRTTDescribeCmd, colabRTTListCmd,
		colabRTTGetIamCmd, colabRTTSetIamCmd, colabRTTAddIamCmd, colabRTTRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagColabRTTRegion, "region", "", "Region where the runtime template lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagColabRTTFormat, "format", "", "Output format")
	}
	colabRTTCreateCmd.Flags().StringVar(&flagColabRTTConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the NotebookRuntimeTemplate body (required)")
	_ = colabRTTCreateCmd.MarkFlagRequired("config-file")
	colabRTTListCmd.Flags().StringVar(&flagColabRTTFilter, "filter", "", "Server-side filter expression")
	colabRTTListCmd.Flags().StringVar(&flagColabRTTOrderBy, "order-by", "", "Order-by expression")
	colabRTTListCmd.Flags().Int64Var(&flagColabRTTPageSize, "page-size", 0, "Maximum results per page")
	colabRTTListCmd.Flags().StringVar(&flagColabRTTReadMask, "read-mask", "", "Field mask for reads")

	for _, c := range []*cobra.Command{colabRTTAddIamCmd, colabRTTRemoveIamCmd} {
		colabIamMemberFlags(c, &flagColabRTTIamMember, &flagColabRTTIamRole,
			&flagColabRTTIamCondExpr, &flagColabRTTIamCondT, &flagColabRTTIamCondD)
	}
	colabRTTRemoveIamCmd.Flags().BoolVar(&flagColabRTTIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	colabRTTCmd.AddCommand(all...)
	colabCmd.AddCommand(colabRTTCmd)
}

func colabRTTParent() (string, error) {
	return colabParent(flagColabRTTRegion)
}

func colabRTTName(id string) (string, error) {
	parent, err := colabRTTParent()
	if err != nil {
		return "", err
	}
	return colabChild("notebookRuntimeTemplates", id, parent), nil
}

func runColabRTTCreate(cmd *cobra.Command, args []string) error {
	parent, err := colabRTTParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1NotebookRuntimeTemplate{}
	if err := loadYAMLOrJSONInto(flagColabRTTConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimeTemplates.Create(parent, body).NotebookRuntimeTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating runtime template: %w", err)
	}
	fmt.Printf("Create request issued for runtime template [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabRTTFormat)
}

func runColabRTTDelete(cmd *cobra.Command, args []string) error {
	name, err := colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimeTemplates.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting runtime template: %w", err)
	}
	fmt.Printf("Delete request issued for runtime template [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabRTTFormat)
}

func runColabRTTDescribe(cmd *cobra.Command, args []string) error {
	name, err := colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTTRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NotebookRuntimeTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing runtime template: %w", err)
	}
	return emitFormatted(got, flagColabRTTFormat)
}

func runColabRTTList(cmd *cobra.Command, args []string) error {
	parent, err := colabRTTParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTTRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1NotebookRuntimeTemplate
	pageToken := ""
	for {
		call := svc.Projects.Locations.NotebookRuntimeTemplates.List(parent).Context(ctx)
		if flagColabRTTFilter != "" {
			call = call.Filter(flagColabRTTFilter)
		}
		if flagColabRTTOrderBy != "" {
			call = call.OrderBy(flagColabRTTOrderBy)
		}
		if flagColabRTTPageSize > 0 {
			call = call.PageSize(flagColabRTTPageSize)
		}
		if flagColabRTTReadMask != "" {
			call = call.ReadMask(flagColabRTTReadMask)
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
	return emitFormatted(all, flagColabRTTFormat)
}

func runColabRTTGetIam(cmd *cobra.Command, args []string) error {
	name, err := colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTTRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.NotebookRuntimeTemplates.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagColabRTTFormat)
}

func runColabRTTSetIam(cmd *cobra.Command, args []string) error {
	name, err := colabRTTName(args[0])
	if err != nil {
		return err
	}
	policy := &aiplatform.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTTRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.NotebookRuntimeTemplates.SetIamPolicy(name, &aiplatform.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	writeUpdatedIam(fmt.Sprintf("runtime template [%s]", args[0]))
	return emitFormatted(updated, flagColabRTTFormat)
}

func runColabRTTAddIam(cmd *cobra.Command, args []string) error {
	name, err := colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTTRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.NotebookRuntimeTemplates.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	colabAddBinding(policy, flagColabRTTIamRole, flagColabRTTIamMember,
		colabBuildCondition(flagColabRTTIamCondExpr, flagColabRTTIamCondT, flagColabRTTIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.NotebookRuntimeTemplates.SetIamPolicy(name, &aiplatform.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	writeUpdatedIam(fmt.Sprintf("runtime template [%s]", args[0]))
	return emitFormatted(updated, flagColabRTTFormat)
}

func runColabRTTRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := colabRTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTTRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.NotebookRuntimeTemplates.GetIamPolicy(name).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !colabRemoveBinding(policy, flagColabRTTIamRole, flagColabRTTIamMember,
		colabBuildCondition(flagColabRTTIamCondExpr, flagColabRTTIamCondT, flagColabRTTIamCondD), flagColabRTTIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagColabRTTIamRole, flagColabRTTIamMember)
	}
	updated, err := svc.Projects.Locations.NotebookRuntimeTemplates.SetIamPolicy(name, &aiplatform.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	writeUpdatedIam(fmt.Sprintf("runtime template [%s]", args[0]))
	return emitFormatted(updated, flagColabRTTFormat)
}
