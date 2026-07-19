package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks instances (#1062) ---

var notebooksInstCmd = &cobra.Command{Use: "instances", Short: "Manage notebook instances"}

var (
	flagNotebooksInstLocation   string
	flagNotebooksInstFormat     string
	flagNotebooksInstConfigFile string
	flagNotebooksInstUpdateMask string
	flagNotebooksInstPageSize   int64

	flagNotebooksInstIamMember   string
	flagNotebooksInstIamRole     string
	flagNotebooksInstIamCondExpr string
	flagNotebooksInstIamCondT    string
	flagNotebooksInstIamCondD    string
	flagNotebooksInstIamAllCond  bool
)

var (
	notebooksInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstCreate,
	}
	notebooksInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstDelete,
	}
	notebooksInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstDescribe,
	}
	notebooksInstListCmd = &cobra.Command{
		Use: "list", Short: "List notebook instances",
		Args: cobra.NoArgs, RunE: runNotebooksInstList,
	}
	notebooksInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a notebook instance configuration",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstUpdate,
	}
	notebooksInstAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding INSTANCE", Short: "Add an IAM policy binding to a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstAddIam,
	}
	notebooksInstRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding INSTANCE", Short: "Remove an IAM policy binding from a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstRemoveIam,
	}
	notebooksInstGetIamCmd = &cobra.Command{
		Use: "get-iam-policy INSTANCE", Short: "Get the IAM policy for a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstGetIam,
	}
	notebooksInstSetIamCmd = &cobra.Command{
		Use: "set-iam-policy INSTANCE POLICY_FILE", Short: "Set the IAM policy for a notebook instance",
		Args: cobra.ExactArgs(2), RunE: runNotebooksInstSetIam,
	}
	notebooksInstDiagnoseCmd = &cobra.Command{
		Use: "diagnose INSTANCE", Short: "Run diagnostics for a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstDiagnose,
	}
	notebooksInstGetHealthCmd = &cobra.Command{
		Use: "get-health INSTANCE", Short: "Get the health state of a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstGetHealth,
	}
	notebooksInstIsUpgradeableCmd = &cobra.Command{
		Use: "is-upgradeable INSTANCE", Short: "Check whether a notebook instance can be upgraded",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstIsUpgradeable,
	}
	notebooksInstMigrateCmd = &cobra.Command{
		Use: "migrate INSTANCE", Short: "Migrate a notebook instance to the Workbench Instances API",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstMigrate,
	}
	notebooksInstRegisterCmd = &cobra.Command{
		Use: "register INSTANCE", Short: "Register a legacy notebook instance under Managed Notebooks",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstRegister,
	}
	notebooksInstResetCmd = &cobra.Command{
		Use: "reset INSTANCE", Short: "Reset a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstReset,
	}
	notebooksInstRollbackCmd = &cobra.Command{
		Use: "rollback INSTANCE", Short: "Rollback a notebook instance to a prior VM image",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstRollback,
	}
	notebooksInstStartCmd = &cobra.Command{
		Use: "start INSTANCE", Short: "Start a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstStart,
	}
	notebooksInstStopCmd = &cobra.Command{
		Use: "stop INSTANCE", Short: "Stop a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstStop,
	}
	notebooksInstUpgradeCmd = &cobra.Command{
		Use: "upgrade INSTANCE", Short: "Upgrade a notebook instance",
		Args: cobra.ExactArgs(1), RunE: runNotebooksInstUpgrade,
	}
)

func init() {
	all := []*cobra.Command{
		notebooksInstCreateCmd, notebooksInstDeleteCmd, notebooksInstDescribeCmd,
		notebooksInstListCmd, notebooksInstUpdateCmd,
		notebooksInstAddIamCmd, notebooksInstRemoveIamCmd,
		notebooksInstGetIamCmd, notebooksInstSetIamCmd,
		notebooksInstDiagnoseCmd, notebooksInstGetHealthCmd, notebooksInstIsUpgradeableCmd,
		notebooksInstMigrateCmd, notebooksInstRegisterCmd, notebooksInstResetCmd,
		notebooksInstRollbackCmd, notebooksInstStartCmd, notebooksInstStopCmd,
		notebooksInstUpgradeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNotebooksInstLocation, "location", "", "Notebook location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNotebooksInstFormat, "format", "", "Output format")
	}
	// Endpoints that consume a --config-file (required):
	for _, c := range []*cobra.Command{
		notebooksInstCreateCmd, notebooksInstUpdateCmd, notebooksInstDiagnoseCmd,
		notebooksInstRegisterCmd, notebooksInstRollbackCmd,
	} {
		c.Flags().StringVar(&flagNotebooksInstConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	// Migrate takes an optional --config-file (the request body may be empty).
	notebooksInstMigrateCmd.Flags().StringVar(&flagNotebooksInstConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the MigrateInstanceRequest body (optional)")
	notebooksInstUpdateCmd.Flags().StringVar(&flagNotebooksInstUpdateMask, "update-mask", "",
		"Comma-separated list of InstanceConfig fields to update (defaults to every populated field)")
	notebooksInstListCmd.Flags().Int64Var(&flagNotebooksInstPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{notebooksInstAddIamCmd, notebooksInstRemoveIamCmd} {
		notebooksIamMemberFlags(c, &flagNotebooksInstIamMember, &flagNotebooksInstIamRole,
			&flagNotebooksInstIamCondExpr, &flagNotebooksInstIamCondT, &flagNotebooksInstIamCondD)
	}
	notebooksInstRemoveIamCmd.Flags().BoolVar(&flagNotebooksInstIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	notebooksInstCmd.AddCommand(all...)
	notebooksCmd.AddCommand(notebooksInstCmd)
}

func notebooksInstParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagNotebooksInstLocation), nil
}

func notebooksInstName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := notebooksInstParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", parent, id), nil
}

func runNotebooksInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := notebooksInstParent()
	if err != nil {
		return err
	}
	body := &notebooksv1.Instance{}
	if err := loadYAMLOrJSONInto(flagNotebooksInstConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(parent, body).InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notebook instance: %w", err)
	}
	fmt.Printf("Create request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstDelete(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook instance: %w", err)
	}
	fmt.Printf("Delete request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notebook instance: %w", err)
	}
	return emitFormatted(got, flagNotebooksInstFormat)
}

func runNotebooksInstList(cmd *cobra.Command, args []string) error {
	parent, err := notebooksInstParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*notebooksv1.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(parent).Context(ctx)
		if flagNotebooksInstPageSize > 0 {
			call = call.PageSize(flagNotebooksInstPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNotebooksInstFormat)
}

// runNotebooksInstUpdate calls the UpdateConfig endpoint (a PATCH against
// v1/{name}:updateConfig) with an InstanceConfig loaded from --config-file.
// The v1 API does not expose a top-level Instance PATCH; UpdateConfig is the
// generalized configuration-update endpoint.
func runNotebooksInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	// Allow the config file to be either a bare InstanceConfig or a full
	// UpdateInstanceConfigRequest ({"config": {...}}).
	req := &notebooksv1.UpdateInstanceConfigRequest{}
	if err := loadYAMLOrJSONInto(flagNotebooksInstConfigFile, req); err != nil {
		return err
	}
	if req.Config == nil {
		ic := &notebooksv1.InstanceConfig{}
		if err := loadYAMLOrJSONInto(flagNotebooksInstConfigFile, ic); err != nil {
			return err
		}
		req.Config = ic
	}
	_ = flagNotebooksInstUpdateMask // accepted but the UpdateConfig endpoint infers the mask from the request body
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.UpdateConfig(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating notebook instance: %w", err)
	}
	fmt.Printf("Update request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstGetIam(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagNotebooksInstFormat)
}

func runNotebooksInstSetIam(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	policy := &notebooksv1.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Instances.SetIamPolicy(name,
		&notebooksv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	notebooksUpdatedIam(fmt.Sprintf("notebook instance [%s]", args[0]))
	return emitFormatted(updated, flagNotebooksInstFormat)
}

func runNotebooksInstAddIam(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	notebooksIamAddBinding(policy, flagNotebooksInstIamRole, flagNotebooksInstIamMember,
		notebooksIamBuildCondition(flagNotebooksInstIamCondExpr, flagNotebooksInstIamCondT, flagNotebooksInstIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Instances.SetIamPolicy(name,
		&notebooksv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	notebooksUpdatedIam(fmt.Sprintf("notebook instance [%s]", args[0]))
	return emitFormatted(updated, flagNotebooksInstFormat)
}

func runNotebooksInstRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !notebooksIamRemoveBinding(policy, flagNotebooksInstIamRole, flagNotebooksInstIamMember,
		notebooksIamBuildCondition(flagNotebooksInstIamCondExpr, flagNotebooksInstIamCondT, flagNotebooksInstIamCondD),
		flagNotebooksInstIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			flagNotebooksInstIamRole, flagNotebooksInstIamMember)
	}
	updated, err := svc.Projects.Locations.Instances.SetIamPolicy(name,
		&notebooksv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	notebooksUpdatedIam(fmt.Sprintf("notebook instance [%s]", args[0]))
	return emitFormatted(updated, flagNotebooksInstFormat)
}

func runNotebooksInstDiagnose(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.DiagnoseInstanceRequest{}
	if err := loadYAMLOrJSONInto(flagNotebooksInstConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Diagnose(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("diagnosing notebook instance: %w", err)
	}
	fmt.Printf("Diagnose request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstGetHealth(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.GetInstanceHealth(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting notebook instance health: %w", err)
	}
	return emitFormatted(got, flagNotebooksInstFormat)
}

func runNotebooksInstIsUpgradeable(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.IsUpgradeable(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking notebook instance upgradeability: %w", err)
	}
	return emitFormatted(got, flagNotebooksInstFormat)
}

func runNotebooksInstMigrate(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.MigrateInstanceRequest{}
	if flagNotebooksInstConfigFile != "" {
		if _, err := os.Stat(flagNotebooksInstConfigFile); err == nil {
			if err := loadYAMLOrJSONInto(flagNotebooksInstConfigFile, req); err != nil {
				return err
			}
		}
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Migrate(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating notebook instance: %w", err)
	}
	fmt.Printf("Migrate request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstRegister(cmd *cobra.Command, args []string) error {
	parent, err := notebooksInstParent()
	if err != nil {
		return err
	}
	req := &notebooksv1.RegisterInstanceRequest{}
	if err := loadYAMLOrJSONInto(flagNotebooksInstConfigFile, req); err != nil {
		return err
	}
	if req.InstanceId == "" {
		req.InstanceId = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Register(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("registering notebook instance: %w", err)
	}
	fmt.Printf("Register request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstReset(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Reset(name, &notebooksv1.ResetInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting notebook instance: %w", err)
	}
	fmt.Printf("Reset request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstRollback(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.RollbackInstanceRequest{}
	if err := loadYAMLOrJSONInto(flagNotebooksInstConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Rollback(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back notebook instance: %w", err)
	}
	fmt.Printf("Rollback request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstStart(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Start(name, &notebooksv1.StartInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting notebook instance: %w", err)
	}
	fmt.Printf("Start request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstStop(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Stop(name, &notebooksv1.StopInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping notebook instance: %w", err)
	}
	fmt.Printf("Stop request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}

func runNotebooksInstUpgrade(cmd *cobra.Command, args []string) error {
	name, err := notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Upgrade(name, &notebooksv1.UpgradeInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upgrading notebook instance: %w", err)
	}
	fmt.Printf("Upgrade request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksInstFormat)
}
