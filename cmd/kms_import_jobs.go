package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms import-jobs (#1103) ---

func (s *kmsCmds) initKmsImportJobs() {
	all := []*cobra.Command{
		s.kmsImpCreateCmd, s.kmsImpDescribeCmd, s.kmsImpListCmd,
		s.kmsImpGetIamCmd, s.kmsImpSetIamCmd, s.kmsImpAddIamCmd, s.kmsImpRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagKmsImpLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsImpKeyring, "keyring", "", "Parent key ring (required)")
		c.Flags().StringVar(&s.flagKmsImpFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("keyring")
	}
	s.kmsImpCreateCmd.Flags().StringVar(&s.flagKmsImpConfigFile, "config-file", "", "YAML/JSON body for the ImportJob (required)")
	_ = s.kmsImpCreateCmd.MarkFlagRequired("config-file")

	s.kmsImpListCmd.Flags().StringVar(&s.flagKmsImpFilter, "filter", "", "Filter expression")
	s.kmsImpListCmd.Flags().Int64Var(&s.flagKmsImpPageSize, "page-size", 0, "Page size")

	kmsIamMemberFlags(s.kmsImpAddIamCmd, &s.flagKmsImpIamMember, &s.flagKmsImpIamRole,
		&s.flagKmsImpIamCondExpr, &s.flagKmsImpIamCondTitle, &s.flagKmsImpIamCondDesc)
	kmsIamMemberFlags(s.kmsImpRemoveIamCmd, &s.flagKmsImpIamMember, &s.flagKmsImpIamRole,
		&s.flagKmsImpIamCondExpr, &s.flagKmsImpIamCondTitle, &s.flagKmsImpIamCondDesc)
	s.kmsImpRemoveIamCmd.Flags().BoolVar(&s.flagKmsImpAllConds, "all", false, "Match bindings for the role across all conditions")

	s.kmsImportJobsCmd.AddCommand(all...)
	s.kmsCmd.AddCommand(s.kmsImportJobsCmd)
}

func kmsImpJobName(project, location, keyring, raw string) string {
	parent := kmsKeyringParent(project, location, keyring) + "/importJobs"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsImpCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.ImportJob{}
	if err := loadYAMLOrJSONInto(s.flagKmsImpConfigFile, body); err != nil {
		return err
	}
	parent := kmsKeyringParent(project, s.flagKmsImpLocation, s.flagKmsImpKeyring)
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.Create(parent, body).
		ImportJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating import job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsImpFormat)
}

func (s *kmsCmds) runKmsImpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsImpJobName(project, s.flagKmsImpLocation, s.flagKmsImpKeyring, args[0])
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing import job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsImpFormat)
}

func (s *kmsCmds) runKmsImpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsKeyringParent(project, s.flagKmsImpLocation, s.flagKmsImpKeyring)
	var all []*cloudkms.ImportJob
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.ImportJobs.List(parent).Context(ctx)
		if s.flagKmsImpFilter != "" {
			call = call.Filter(s.flagKmsImpFilter)
		}
		if s.flagKmsImpPageSize > 0 {
			call = call.PageSize(s.flagKmsImpPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsImpFormat)
}

func (s *kmsCmds) runKmsImpGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsImpJobName(project, s.flagKmsImpLocation, s.flagKmsImpKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.ImportJobs.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, pol, s.flagKmsImpFormat)
}

func (s *kmsCmds) runKmsImpSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	policy := &cloudkms.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	name := kmsImpJobName(project, s.flagKmsImpLocation, s.flagKmsImpKeyring, args[0])
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsImpFormat)
}

func (s *kmsCmds) runKmsImpAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsImpJobName(project, s.flagKmsImpLocation, s.flagKmsImpKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.ImportJobs.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsImpIamCondExpr, s.flagKmsImpIamCondTitle, s.flagKmsImpIamCondDesc)
	kmsIamAddBinding(pol, s.flagKmsImpIamRole, s.flagKmsImpIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsImpFormat)
}

func (s *kmsCmds) runKmsImpRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsImpJobName(project, s.flagKmsImpLocation, s.flagKmsImpKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.ImportJobs.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsImpIamCondExpr, s.flagKmsImpIamCondTitle, s.flagKmsImpIamCondDesc)
	if !kmsIamRemoveBinding(pol, s.flagKmsImpIamRole, s.flagKmsImpIamMember, cond, s.flagKmsImpAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.ImportJobs.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsImpFormat)
}
