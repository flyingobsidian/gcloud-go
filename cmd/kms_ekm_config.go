package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms ekm-config (#1101) ---

var kmsEkmConfigCmd = &cobra.Command{
	Use:   "ekm-config",
	Short: "Manage per-location EkmConfig resources",
}

var (
	flagKmsEkmCfgLocation   string
	flagKmsEkmCfgFormat     string
	flagKmsEkmCfgConfigFile string
	flagKmsEkmCfgMask       string
	flagKmsEkmCfgAllConds   bool

	flagKmsEkmCfgIamMember    string
	flagKmsEkmCfgIamRole      string
	flagKmsEkmCfgIamCondExpr  string
	flagKmsEkmCfgIamCondTitle string
	flagKmsEkmCfgIamCondDesc  string
)

var kmsEkmCfgDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe the EkmConfig for a location",
	Args:  cobra.NoArgs,
	RunE:  runKmsEkmCfgDescribe,
}

var kmsEkmCfgUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the EkmConfig for a location",
	Args:  cobra.NoArgs,
	RunE:  runKmsEkmCfgUpdate,
}

var kmsEkmCfgGetIamCmd = &cobra.Command{
	Use:   "get-iam-policy",
	Short: "Get the IAM policy for the EkmConfig",
	Args:  cobra.NoArgs,
	RunE:  runKmsEkmCfgGetIam,
}

var kmsEkmCfgSetIamCmd = &cobra.Command{
	Use:   "set-iam-policy POLICY_FILE",
	Short: "Set the IAM policy for the EkmConfig",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsEkmCfgSetIam,
}

var kmsEkmCfgAddIamCmd = &cobra.Command{
	Use:   "add-iam-policy-binding",
	Short: "Add an IAM policy binding on the EkmConfig",
	Args:  cobra.NoArgs,
	RunE:  runKmsEkmCfgAddIam,
}

var kmsEkmCfgRemoveIamCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding",
	Short: "Remove an IAM policy binding from the EkmConfig",
	Args:  cobra.NoArgs,
	RunE:  runKmsEkmCfgRemoveIam,
}

func init() {
	all := []*cobra.Command{
		kmsEkmCfgDescribeCmd, kmsEkmCfgUpdateCmd,
		kmsEkmCfgGetIamCmd, kmsEkmCfgSetIamCmd,
		kmsEkmCfgAddIamCmd, kmsEkmCfgRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagKmsEkmCfgLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsEkmCfgFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}

	kmsEkmCfgUpdateCmd.Flags().StringVar(&flagKmsEkmCfgConfigFile, "config-file", "", "YAML/JSON body for the EkmConfig (required)")
	kmsEkmCfgUpdateCmd.Flags().StringVar(&flagKmsEkmCfgMask, "update-mask", "", "Fields to update; defaults to populated fields")
	_ = kmsEkmCfgUpdateCmd.MarkFlagRequired("config-file")

	kmsIamMemberFlags(kmsEkmCfgAddIamCmd, &flagKmsEkmCfgIamMember, &flagKmsEkmCfgIamRole,
		&flagKmsEkmCfgIamCondExpr, &flagKmsEkmCfgIamCondTitle, &flagKmsEkmCfgIamCondDesc)
	kmsIamMemberFlags(kmsEkmCfgRemoveIamCmd, &flagKmsEkmCfgIamMember, &flagKmsEkmCfgIamRole,
		&flagKmsEkmCfgIamCondExpr, &flagKmsEkmCfgIamCondTitle, &flagKmsEkmCfgIamCondDesc)
	kmsEkmCfgRemoveIamCmd.Flags().BoolVar(&flagKmsEkmCfgAllConds, "all", false, "Match bindings for the role across all conditions")

	kmsEkmConfigCmd.AddCommand(all...)
	kmsCmd.AddCommand(kmsEkmConfigCmd)
}

func kmsEkmConfigName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/ekmConfig", project, location)
}

func runKmsEkmCfgDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsEkmConfigName(project, flagKmsEkmCfgLocation)
	out, err := svc.Projects.Locations.GetEkmConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing EkmConfig: %w", err)
	}
	return emitFormatted(out, flagKmsEkmCfgFormat)
}

func runKmsEkmCfgUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.EkmConfig{}
	if err := loadYAMLOrJSONInto(flagKmsEkmCfgConfigFile, body); err != nil {
		return err
	}
	mask := flagKmsEkmCfgMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := kmsEkmConfigName(project, flagKmsEkmCfgLocation)
	call := svc.Projects.Locations.UpdateEkmConfig(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating EkmConfig: %w", err)
	}
	return emitFormatted(out, flagKmsEkmCfgFormat)
}

func runKmsEkmCfgGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsEkmConfigName(project, flagKmsEkmCfgLocation)
	pol, err := svc.Projects.Locations.EkmConfig.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(pol, flagKmsEkmCfgFormat)
}

func runKmsEkmCfgSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy := &cloudkms.Policy{}
	if err := loadYAMLOrJSONInto(args[0], policy); err != nil {
		return err
	}
	name := kmsEkmConfigName(project, flagKmsEkmCfgLocation)
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.EkmConfig.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsEkmCfgFormat)
}

func runKmsEkmCfgAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsEkmConfigName(project, flagKmsEkmCfgLocation)
	pol, err := svc.Projects.Locations.EkmConfig.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsEkmCfgIamCondExpr, flagKmsEkmCfgIamCondTitle, flagKmsEkmCfgIamCondDesc)
	kmsIamAddBinding(pol, flagKmsEkmCfgIamRole, flagKmsEkmCfgIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.EkmConfig.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsEkmCfgFormat)
}

func runKmsEkmCfgRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsEkmConfigName(project, flagKmsEkmCfgLocation)
	pol, err := svc.Projects.Locations.EkmConfig.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsEkmCfgIamCondExpr, flagKmsEkmCfgIamCondTitle, flagKmsEkmCfgIamCondDesc)
	if !kmsIamRemoveBinding(pol, flagKmsEkmCfgIamRole, flagKmsEkmCfgIamMember, cond, flagKmsEkmCfgAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.EkmConfig.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsEkmCfgFormat)
}
