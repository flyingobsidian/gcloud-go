package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager perimeters (#1444) ---

var acmPMCmd = &cobra.Command{Use: "perimeters", Short: "Manage service perimeters"}

var (
	flagACMPMFormat     string
	flagACMPMPolicy     string
	flagACMPMConfigFile string
	flagACMPMSourceFile string
	flagACMPMPageSize   int64
	flagACMPMEtag       string
)

var (
	acmPMCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create a service perimeter",
		Args: cobra.ExactArgs(1), RunE: runACMPMCreate,
	}
	acmPMDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete a service perimeter",
		Args: cobra.ExactArgs(1), RunE: runACMPMDelete,
	}
	acmPMDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe a service perimeter",
		Args: cobra.ExactArgs(1), RunE: runACMPMDescribe,
	}
	acmPMListCmd = &cobra.Command{
		Use: "list", Short: "List service perimeters",
		Args: cobra.NoArgs, RunE: runACMPMList,
	}
	acmPMUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update a service perimeter",
		Args: cobra.ExactArgs(1), RunE: runACMPMUpdate,
	}
	acmPMReplaceAllCmd = &cobra.Command{
		Use: "replace-all", Short: "Replace all service perimeters in a policy",
		Args: cobra.NoArgs, RunE: runACMPMReplaceAll,
	}
)

var acmPMDryRunCmd = &cobra.Command{Use: "dry-run", Short: "Manage dry-run configuration for a perimeter"}

var (
	acmPMDryCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create the dry-run spec for a perimeter",
		Args: cobra.ExactArgs(1), RunE: runACMPMDryCreate,
	}
	acmPMDryDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe the dry-run configuration for a perimeter",
		Args: cobra.ExactArgs(1), RunE: runACMPMDryDescribe,
	}
	acmPMDryEnforceCmd = &cobra.Command{
		Use: "enforce", Short: "Commit dry-run specs across all perimeters in a policy",
		Args: cobra.NoArgs, RunE: runACMPMDryEnforce,
	}
	acmPMDryResetCmd = &cobra.Command{
		Use: "reset NAME", Short: "Reset the dry-run spec for a perimeter",
		Args: cobra.ExactArgs(1), RunE: runACMPMDryDrop,
	}
	acmPMDryDropCmd = &cobra.Command{
		Use: "drop NAME", Short: "Drop the dry-run spec for a perimeter",
		Args: cobra.ExactArgs(1), RunE: runACMPMDryDrop,
	}
)

func init() {
	all := []*cobra.Command{
		acmPMCreateCmd, acmPMDeleteCmd, acmPMDescribeCmd, acmPMListCmd, acmPMUpdateCmd, acmPMReplaceAllCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagACMPMFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagACMPMPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	for _, c := range []*cobra.Command{acmPMCreateCmd, acmPMUpdateCmd} {
		c.Flags().StringVar(&flagACMPMConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ServicePerimeter body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	acmPMReplaceAllCmd.Flags().StringVar(&flagACMPMSourceFile, "source-file", "",
		"Path to a YAML/JSON file containing a list of ServicePerimeter bodies (required)")
	_ = acmPMReplaceAllCmd.MarkFlagRequired("source-file")
	acmPMReplaceAllCmd.Flags().StringVar(&flagACMPMEtag, "etag", "", "Expected etag of the parent AccessPolicy")
	acmPMListCmd.Flags().Int64Var(&flagACMPMPageSize, "page-size", 0, "Maximum results per page")

	dryAll := []*cobra.Command{
		acmPMDryCreateCmd, acmPMDryDescribeCmd, acmPMDryEnforceCmd, acmPMDryResetCmd, acmPMDryDropCmd,
	}
	for _, c := range dryAll {
		c.Flags().StringVar(&flagACMPMFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagACMPMPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	acmPMDryCreateCmd.Flags().StringVar(&flagACMPMConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the dry-run ServicePerimeter body (required)")
	_ = acmPMDryCreateCmd.MarkFlagRequired("config-file")
	acmPMDryEnforceCmd.Flags().StringVar(&flagACMPMEtag, "etag", "", "Expected etag of the parent AccessPolicy")

	acmPMDryRunCmd.AddCommand(dryAll...)
	acmPMCmd.AddCommand(all...)
	acmPMCmd.AddCommand(acmPMDryRunCmd)
	accessContextManagerCmd.AddCommand(acmPMCmd)
}

func acmPMResource(policy, name string) string {
	return fmt.Sprintf("%s/servicePerimeters/%s", acmPolicyResource(policy), name)
}

func runACMPMCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.ServicePerimeter{}
	if err := loadYAMLOrJSONInto(flagACMPMConfigFile, body); err != nil {
		return err
	}
	body.Name = acmPMResource(flagACMPMPolicy, args[0])
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.ServicePerimeters.Create(acmPolicyResource(flagACMPMPolicy), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service perimeter: %w", err)
	}
	fmt.Printf("Create request issued for service perimeter [%s].\n", args[0])
	return emitFormatted(op, flagACMPMFormat)
}

func runACMPMDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.ServicePerimeters.Delete(acmPMResource(flagACMPMPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service perimeter: %w", err)
	}
	fmt.Printf("Delete request issued for service perimeter [%s].\n", args[0])
	return emitFormatted(op, flagACMPMFormat)
}

func runACMPMDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.ServicePerimeters.Get(acmPMResource(flagACMPMPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service perimeter: %w", err)
	}
	return emitFormatted(got, flagACMPMFormat)
}

func runACMPMList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.ServicePerimeter
	pageToken := ""
	for {
		call := svc.AccessPolicies.ServicePerimeters.List(acmPolicyResource(flagACMPMPolicy)).Context(ctx)
		if flagACMPMPageSize > 0 {
			call = call.PageSize(flagACMPMPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service perimeters: %w", err)
		}
		all = append(all, resp.ServicePerimeters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagACMPMFormat)
}

func runACMPMUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.ServicePerimeter{}
	if err := loadYAMLOrJSONInto(flagACMPMConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.AccessPolicies.ServicePerimeters.Patch(acmPMResource(flagACMPMPolicy, args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service perimeter: %w", err)
	}
	fmt.Printf("Update request issued for service perimeter [%s].\n", args[0])
	return emitFormatted(op, flagACMPMFormat)
}

func runACMPMReplaceAll(cmd *cobra.Command, args []string) error {
	var perimeters []*accesscontextmanager.ServicePerimeter
	if err := loadYAMLOrJSONInto(flagACMPMSourceFile, &perimeters); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &accesscontextmanager.ReplaceServicePerimetersRequest{
		ServicePerimeters: perimeters,
		Etag:              flagACMPMEtag,
	}
	op, err := svc.AccessPolicies.ServicePerimeters.ReplaceAll(acmPolicyResource(flagACMPMPolicy), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing service perimeters: %w", err)
	}
	fmt.Println("Replace-all request issued for service perimeters.")
	return emitFormatted(op, flagACMPMFormat)
}

func runACMPMDryCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.ServicePerimeter{}
	if err := loadYAMLOrJSONInto(flagACMPMConfigFile, body); err != nil {
		return err
	}
	body.UseExplicitDryRunSpec = true
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.ServicePerimeters.
		Patch(acmPMResource(flagACMPMPolicy, args[0]), body).
		UpdateMask("spec,useExplicitDryRunSpec").
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating dry-run spec: %w", err)
	}
	fmt.Printf("Dry-run spec created for service perimeter [%s].\n", args[0])
	return emitFormatted(op, flagACMPMFormat)
}

func runACMPMDryDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.ServicePerimeters.Get(acmPMResource(flagACMPMPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service perimeter: %w", err)
	}
	view := struct {
		Name                  string                                    `json:"name"`
		Status                *accesscontextmanager.ServicePerimeterConfig `json:"status,omitempty"`
		Spec                  *accesscontextmanager.ServicePerimeterConfig `json:"spec,omitempty"`
		UseExplicitDryRunSpec bool                                      `json:"useExplicitDryRunSpec"`
	}{
		Name:                  got.Name,
		Status:                got.Status,
		Spec:                  got.Spec,
		UseExplicitDryRunSpec: got.UseExplicitDryRunSpec,
	}
	return emitFormatted(view, flagACMPMFormat)
}

func runACMPMDryEnforce(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &accesscontextmanager.CommitServicePerimetersRequest{Etag: flagACMPMEtag}
	op, err := svc.AccessPolicies.ServicePerimeters.Commit(acmPolicyResource(flagACMPMPolicy), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing dry-run specs: %w", err)
	}
	fmt.Println("Enforce request issued for dry-run specs.")
	return emitFormatted(op, flagACMPMFormat)
}

func runACMPMDryDrop(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &accesscontextmanager.ServicePerimeter{
		UseExplicitDryRunSpec: false,
		ForceSendFields:       []string{"UseExplicitDryRunSpec"},
		NullFields:            []string{"Spec"},
	}
	op, err := svc.AccessPolicies.ServicePerimeters.
		Patch(acmPMResource(flagACMPMPolicy, args[0]), body).
		UpdateMask("spec,useExplicitDryRunSpec").
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("clearing dry-run spec: %w", err)
	}
	fmt.Printf("Dry-run spec cleared for service perimeter [%s].\n", args[0])
	return emitFormatted(op, flagACMPMFormat)
}
