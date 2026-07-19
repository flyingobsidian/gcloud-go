package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms ekm-connections (#1102) ---

var kmsEkmConnectionsCmd = &cobra.Command{
	Use:   "ekm-connections",
	Short: "Manage Cloud KMS EKM connections",
}

var (
	flagKmsEkmConnLocation   string
	flagKmsEkmConnFormat     string
	flagKmsEkmConnFilter     string
	flagKmsEkmConnPageSize   int64
	flagKmsEkmConnConfigFile string
	flagKmsEkmConnMask       string
	flagKmsEkmConnAllConds   bool

	flagKmsEkmConnIamMember    string
	flagKmsEkmConnIamRole      string
	flagKmsEkmConnIamCondExpr  string
	flagKmsEkmConnIamCondTitle string
	flagKmsEkmConnIamCondDesc  string
)

var kmsEkmConnCreateCmd = &cobra.Command{
	Use:   "create CONNECTION",
	Short: "Create an EKM connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsEkmConnCreate,
}

var kmsEkmConnDescribeCmd = &cobra.Command{
	Use:   "describe CONNECTION",
	Short: "Describe an EKM connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsEkmConnDescribe,
}

var kmsEkmConnListCmd = &cobra.Command{
	Use:   "list",
	Short: "List EKM connections in a location",
	Args:  cobra.NoArgs,
	RunE:  runKmsEkmConnList,
}

var kmsEkmConnUpdateCmd = &cobra.Command{
	Use:   "update CONNECTION",
	Short: "Update an EKM connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsEkmConnUpdate,
}

var kmsEkmConnGetIamCmd = &cobra.Command{
	Use:   "get-iam-policy CONNECTION",
	Short: "Get the IAM policy for an EKM connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsEkmConnGetIam,
}

var kmsEkmConnSetIamCmd = &cobra.Command{
	Use:   "set-iam-policy CONNECTION POLICY_FILE",
	Short: "Set the IAM policy for an EKM connection",
	Args:  cobra.ExactArgs(2),
	RunE:  runKmsEkmConnSetIam,
}

var kmsEkmConnAddIamCmd = &cobra.Command{
	Use:   "add-iam-policy-binding CONNECTION",
	Short: "Add an IAM policy binding on an EKM connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsEkmConnAddIam,
}

var kmsEkmConnRemoveIamCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding CONNECTION",
	Short: "Remove an IAM policy binding from an EKM connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsEkmConnRemoveIam,
}

func init() {
	all := []*cobra.Command{
		kmsEkmConnCreateCmd, kmsEkmConnDescribeCmd, kmsEkmConnListCmd, kmsEkmConnUpdateCmd,
		kmsEkmConnGetIamCmd, kmsEkmConnSetIamCmd, kmsEkmConnAddIamCmd, kmsEkmConnRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagKmsEkmConnLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsEkmConnFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}

	kmsEkmConnCreateCmd.Flags().StringVar(&flagKmsEkmConnConfigFile, "config-file", "", "YAML/JSON body for the EkmConnection (required)")
	_ = kmsEkmConnCreateCmd.MarkFlagRequired("config-file")

	kmsEkmConnListCmd.Flags().StringVar(&flagKmsEkmConnFilter, "filter", "", "Filter expression")
	kmsEkmConnListCmd.Flags().Int64Var(&flagKmsEkmConnPageSize, "page-size", 0, "Page size")

	kmsEkmConnUpdateCmd.Flags().StringVar(&flagKmsEkmConnConfigFile, "config-file", "", "YAML/JSON body for the EkmConnection update (required)")
	kmsEkmConnUpdateCmd.Flags().StringVar(&flagKmsEkmConnMask, "update-mask", "", "Fields to update; defaults to populated fields")
	_ = kmsEkmConnUpdateCmd.MarkFlagRequired("config-file")

	kmsIamMemberFlags(kmsEkmConnAddIamCmd, &flagKmsEkmConnIamMember, &flagKmsEkmConnIamRole,
		&flagKmsEkmConnIamCondExpr, &flagKmsEkmConnIamCondTitle, &flagKmsEkmConnIamCondDesc)
	kmsIamMemberFlags(kmsEkmConnRemoveIamCmd, &flagKmsEkmConnIamMember, &flagKmsEkmConnIamRole,
		&flagKmsEkmConnIamCondExpr, &flagKmsEkmConnIamCondTitle, &flagKmsEkmConnIamCondDesc)
	kmsEkmConnRemoveIamCmd.Flags().BoolVar(&flagKmsEkmConnAllConds, "all", false, "Match bindings for the role across all conditions")

	kmsEkmConnectionsCmd.AddCommand(all...)
	kmsCmd.AddCommand(kmsEkmConnectionsCmd)
}

func kmsEkmConnName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/ekmConnections"
	return kmsFullName(parent, raw)
}

func runKmsEkmConnCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.EkmConnection{}
	if err := loadYAMLOrJSONInto(flagKmsEkmConnConfigFile, body); err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsEkmConnLocation)
	out, err := svc.Projects.Locations.EkmConnections.Create(parent, body).
		EkmConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating EKM connection: %w", err)
	}
	return emitFormatted(out, flagKmsEkmConnFormat)
}

func runKmsEkmConnDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsEkmConnName(project, flagKmsEkmConnLocation, args[0])
	out, err := svc.Projects.Locations.EkmConnections.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing EKM connection: %w", err)
	}
	return emitFormatted(out, flagKmsEkmConnFormat)
}

func runKmsEkmConnList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsEkmConnLocation)
	var all []*cloudkms.EkmConnection
	token := ""
	for {
		call := svc.Projects.Locations.EkmConnections.List(parent).Context(ctx)
		if flagKmsEkmConnFilter != "" {
			call = call.Filter(flagKmsEkmConnFilter)
		}
		if flagKmsEkmConnPageSize > 0 {
			call = call.PageSize(flagKmsEkmConnPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing EKM connections: %w", err)
		}
		all = append(all, resp.EkmConnections...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsEkmConnFormat)
}

func runKmsEkmConnUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.EkmConnection{}
	if err := loadYAMLOrJSONInto(flagKmsEkmConnConfigFile, body); err != nil {
		return err
	}
	mask := flagKmsEkmConnMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := kmsEkmConnName(project, flagKmsEkmConnLocation, args[0])
	call := svc.Projects.Locations.EkmConnections.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating EKM connection: %w", err)
	}
	return emitFormatted(out, flagKmsEkmConnFormat)
}

func runKmsEkmConnGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsEkmConnName(project, flagKmsEkmConnLocation, args[0])
	pol, err := svc.Projects.Locations.EkmConnections.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(pol, flagKmsEkmConnFormat)
}

func runKmsEkmConnSetIam(cmd *cobra.Command, args []string) error {
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
	name := kmsEkmConnName(project, flagKmsEkmConnLocation, args[0])
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.EkmConnections.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsEkmConnFormat)
}

func runKmsEkmConnAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsEkmConnName(project, flagKmsEkmConnLocation, args[0])
	pol, err := svc.Projects.Locations.EkmConnections.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsEkmConnIamCondExpr, flagKmsEkmConnIamCondTitle, flagKmsEkmConnIamCondDesc)
	kmsIamAddBinding(pol, flagKmsEkmConnIamRole, flagKmsEkmConnIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.EkmConnections.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsEkmConnFormat)
}

func runKmsEkmConnRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsEkmConnName(project, flagKmsEkmConnLocation, args[0])
	pol, err := svc.Projects.Locations.EkmConnections.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsEkmConnIamCondExpr, flagKmsEkmConnIamCondTitle, flagKmsEkmConnIamCondDesc)
	if !kmsIamRemoveBinding(pol, flagKmsEkmConnIamRole, flagKmsEkmConnIamMember, cond, flagKmsEkmConnAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.EkmConnections.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsEkmConnFormat)
}
