package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms import-jobs (#1103) ---

var kmsImportJobsCmd = &cobra.Command{
	Use:   "import-jobs",
	Short: "Manage Cloud KMS import jobs",
}

var (
	flagKmsImpLocation   string
	flagKmsImpKeyring    string
	flagKmsImpFormat     string
	flagKmsImpFilter     string
	flagKmsImpPageSize   int64
	flagKmsImpConfigFile string
	flagKmsImpAllConds   bool

	flagKmsImpIamMember    string
	flagKmsImpIamRole      string
	flagKmsImpIamCondExpr  string
	flagKmsImpIamCondTitle string
	flagKmsImpIamCondDesc  string
)

var kmsImpCreateCmd = &cobra.Command{
	Use:   "create JOB",
	Short: "Create an import job",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsImpCreate,
}

var kmsImpDescribeCmd = &cobra.Command{
	Use:   "describe JOB",
	Short: "Describe an import job",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsImpDescribe,
}

var kmsImpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List import jobs in a keyring",
	Args:  cobra.NoArgs,
	RunE:  runKmsImpList,
}

var kmsImpGetIamCmd = &cobra.Command{
	Use:   "get-iam-policy JOB",
	Short: "Get the IAM policy for an import job",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsImpGetIam,
}

var kmsImpSetIamCmd = &cobra.Command{
	Use:   "set-iam-policy JOB POLICY_FILE",
	Short: "Set the IAM policy for an import job",
	Args:  cobra.ExactArgs(2),
	RunE:  runKmsImpSetIam,
}

var kmsImpAddIamCmd = &cobra.Command{
	Use:   "add-iam-policy-binding JOB",
	Short: "Add an IAM policy binding on an import job",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsImpAddIam,
}

var kmsImpRemoveIamCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding JOB",
	Short: "Remove an IAM policy binding from an import job",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsImpRemoveIam,
}

func init() {
	all := []*cobra.Command{
		kmsImpCreateCmd, kmsImpDescribeCmd, kmsImpListCmd,
		kmsImpGetIamCmd, kmsImpSetIamCmd, kmsImpAddIamCmd, kmsImpRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagKmsImpLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsImpKeyring, "keyring", "", "Parent key ring (required)")
		c.Flags().StringVar(&flagKmsImpFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("keyring")
	}
	kmsImpCreateCmd.Flags().StringVar(&flagKmsImpConfigFile, "config-file", "", "YAML/JSON body for the ImportJob (required)")
	_ = kmsImpCreateCmd.MarkFlagRequired("config-file")

	kmsImpListCmd.Flags().StringVar(&flagKmsImpFilter, "filter", "", "Filter expression")
	kmsImpListCmd.Flags().Int64Var(&flagKmsImpPageSize, "page-size", 0, "Page size")

	kmsIamMemberFlags(kmsImpAddIamCmd, &flagKmsImpIamMember, &flagKmsImpIamRole,
		&flagKmsImpIamCondExpr, &flagKmsImpIamCondTitle, &flagKmsImpIamCondDesc)
	kmsIamMemberFlags(kmsImpRemoveIamCmd, &flagKmsImpIamMember, &flagKmsImpIamRole,
		&flagKmsImpIamCondExpr, &flagKmsImpIamCondTitle, &flagKmsImpIamCondDesc)
	kmsImpRemoveIamCmd.Flags().BoolVar(&flagKmsImpAllConds, "all", false, "Match bindings for the role across all conditions")

	kmsImportJobsCmd.AddCommand(all...)
	kmsCmd.AddCommand(kmsImportJobsCmd)
}

func kmsImpJobName(project, location, keyring, raw string) string {
	parent := kmsKeyringParent(project, location, keyring) + "/importJobs"
	return kmsFullName(parent, raw)
}

func runKmsImpCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.ImportJob{}
	if err := loadYAMLOrJSONInto(flagKmsImpConfigFile, body); err != nil {
		return err
	}
	parent := kmsKeyringParent(project, flagKmsImpLocation, flagKmsImpKeyring)
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.Create(parent, body).
		ImportJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating import job: %w", err)
	}
	return emitFormatted(out, flagKmsImpFormat)
}

func runKmsImpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsImpJobName(project, flagKmsImpLocation, flagKmsImpKeyring, args[0])
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing import job: %w", err)
	}
	return emitFormatted(out, flagKmsImpFormat)
}

func runKmsImpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsKeyringParent(project, flagKmsImpLocation, flagKmsImpKeyring)
	var all []*cloudkms.ImportJob
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.ImportJobs.List(parent).Context(ctx)
		if flagKmsImpFilter != "" {
			call = call.Filter(flagKmsImpFilter)
		}
		if flagKmsImpPageSize > 0 {
			call = call.PageSize(flagKmsImpPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing import jobs: %w", err)
		}
		all = append(all, resp.ImportJobs...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsImpFormat)
}

func runKmsImpGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsImpJobName(project, flagKmsImpLocation, flagKmsImpKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.ImportJobs.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(pol, flagKmsImpFormat)
}

func runKmsImpSetIam(cmd *cobra.Command, args []string) error {
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
	name := kmsImpJobName(project, flagKmsImpLocation, flagKmsImpKeyring, args[0])
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsImpFormat)
}

func runKmsImpAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsImpJobName(project, flagKmsImpLocation, flagKmsImpKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.ImportJobs.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsImpIamCondExpr, flagKmsImpIamCondTitle, flagKmsImpIamCondDesc)
	kmsIamAddBinding(pol, flagKmsImpIamRole, flagKmsImpIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsImpFormat)
}

func runKmsImpRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsImpJobName(project, flagKmsImpLocation, flagKmsImpKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.ImportJobs.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(flagKmsImpIamCondExpr, flagKmsImpIamCondTitle, flagKmsImpIamCondDesc)
	if !kmsIamRemoveBinding(pol, flagKmsImpIamRole, flagKmsImpIamMember, cond, flagKmsImpAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	kmsUpdatedIam(name)
	return emitFormatted(out, flagKmsImpFormat)
}
