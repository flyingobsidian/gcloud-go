package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms keyrings (#1106) ---

var kmsKeyringsCmd = &cobra.Command{
	Use:   "keyrings",
	Short: "Manage Cloud KMS key rings",
}

var (
	flagKmsKRLocation   string
	flagKmsKRFormat     string
	flagKmsKRFilter     string
	flagKmsKRPageSize   int64
	flagKmsKRConfigFile string
	flagKmsKRAllConds   bool

	flagKmsKRIamMember    string
	flagKmsKRIamRole      string
	flagKmsKRIamCondExpr  string
	flagKmsKRIamCondTitle string
	flagKmsKRIamCondDesc  string
)

var kmsKRCreateCmd = &cobra.Command{
	Use:   "create KEYRING",
	Short: "Create a key ring",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKRCreate,
}

var kmsKRDescribeCmd = &cobra.Command{
	Use:   "describe KEYRING",
	Short: "Describe a key ring",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKRDescribe,
}

var kmsKRListCmd = &cobra.Command{
	Use:   "list",
	Short: "List key rings in a location",
	Args:  cobra.NoArgs,
	RunE:  runKmsKRList,
}

var kmsKRGetIamCmd = &cobra.Command{
	Use:   "get-iam-policy KEYRING",
	Short: "Get the IAM policy for a key ring",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKRGetIam,
}

var kmsKRSetIamCmd = &cobra.Command{
	Use:   "set-iam-policy KEYRING POLICY_FILE",
	Short: "Set the IAM policy for a key ring",
	Args:  cobra.ExactArgs(2),
	RunE:  runKmsKRSetIam,
}

var kmsKRAddIamCmd = &cobra.Command{
	Use:   "add-iam-policy-binding KEYRING",
	Short: "Add an IAM policy binding on a key ring",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKRAddIam,
}

var kmsKRRemoveIamCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding KEYRING",
	Short: "Remove an IAM policy binding from a key ring",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKRRemoveIam,
}

func init() {
	all := []*cobra.Command{
		kmsKRCreateCmd, kmsKRDescribeCmd, kmsKRListCmd,
		kmsKRGetIamCmd, kmsKRSetIamCmd, kmsKRAddIamCmd, kmsKRRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagKmsKRLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsKRFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	kmsKRCreateCmd.Flags().StringVar(&flagKmsKRConfigFile, "config-file", "", "YAML/JSON body for the KeyRing (optional; empty is valid)")

	kmsKRListCmd.Flags().StringVar(&flagKmsKRFilter, "filter", "", "Filter expression")
	kmsKRListCmd.Flags().Int64Var(&flagKmsKRPageSize, "page-size", 0, "Page size")

	kmsIamMemberFlags(kmsKRAddIamCmd, &flagKmsKRIamMember, &flagKmsKRIamRole,
		&flagKmsKRIamCondExpr, &flagKmsKRIamCondTitle, &flagKmsKRIamCondDesc)
	kmsIamMemberFlags(kmsKRRemoveIamCmd, &flagKmsKRIamMember, &flagKmsKRIamRole,
		&flagKmsKRIamCondExpr, &flagKmsKRIamCondTitle, &flagKmsKRIamCondDesc)
	kmsKRRemoveIamCmd.Flags().BoolVar(&flagKmsKRAllConds, "all", false, "Match bindings for the role across all conditions")

	kmsKeyringsCmd.AddCommand(all...)
	kmsCmd.AddCommand(kmsKeyringsCmd)
}

func kmsKRName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/keyRings"
	return kmsFullName(parent, raw)
}

func runKmsKRCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.KeyRing{}
	if flagKmsKRConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagKmsKRConfigFile, body); err != nil {
			return err
		}
	}
	parent := kmsLocationParent(project, flagKmsKRLocation)
	out, err := svc.Projects.Locations.KeyRings.Create(parent, body).
		KeyRingId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating key ring: %w", err)
	}
	return emitFormatted(out, flagKmsKRFormat)
}

func runKmsKRDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKRName(project, flagKmsKRLocation, args[0])
	out, err := svc.Projects.Locations.KeyRings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing key ring: %w", err)
	}
	return emitFormatted(out, flagKmsKRFormat)
}

func runKmsKRList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsKRLocation)
	var all []*cloudkms.KeyRing
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.List(parent).Context(ctx)
		if flagKmsKRFilter != "" {
			call = call.Filter(flagKmsKRFilter)
		}
		if flagKmsKRPageSize > 0 {
			call = call.PageSize(flagKmsKRPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing key rings: %w", err)
		}
		all = append(all, resp.KeyRings...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsKRFormat)
}

func runKmsKRGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKRName(project, flagKmsKRLocation, args[0])
	pol, err := svc.Projects.Locations.KeyRings.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(pol, flagKmsKRFormat)
}

func runKmsKRSetIam(cmd *cobra.Command, args []string) error {
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
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	name := kmsKRName(project, flagKmsKRLocation, args[0])
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.KeyRings.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsKRFormat)
}

func runKmsKRAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKRName(project, flagKmsKRLocation, args[0])
	pol, err := svc.Projects.Locations.KeyRings.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsKRIamCondExpr, flagKmsKRIamCondTitle, flagKmsKRIamCondDesc)
	kmsIamAddBinding(pol, flagKmsKRIamRole, flagKmsKRIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsKRFormat)
}

func runKmsKRRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKRName(project, flagKmsKRLocation, args[0])
	pol, err := svc.Projects.Locations.KeyRings.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsKRIamCondExpr, flagKmsKRIamCondTitle, flagKmsKRIamCondDesc)
	if !kmsIamRemoveBinding(pol, flagKmsKRIamRole, flagKmsKRIamMember, cond, flagKmsKRAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsKRFormat)
}
