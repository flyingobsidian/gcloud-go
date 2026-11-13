package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms ekm-config (#1101) ---

func (s *kmsCmds) initKmsEkmConfig() {
	all := []*cobra.Command{
		s.kmsEkmCfgDescribeCmd, s.kmsEkmCfgUpdateCmd,
		s.kmsEkmCfgGetIamCmd, s.kmsEkmCfgSetIamCmd,
		s.kmsEkmCfgAddIamCmd, s.kmsEkmCfgRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagKmsEkmCfgLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsEkmCfgFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}

	s.kmsEkmCfgUpdateCmd.Flags().StringVar(&s.flagKmsEkmCfgConfigFile, "config-file", "", "YAML/JSON body for the EkmConfig (required)")
	s.kmsEkmCfgUpdateCmd.Flags().StringVar(&s.flagKmsEkmCfgMask, "update-mask", "", "Fields to update; defaults to populated fields")
	_ = s.kmsEkmCfgUpdateCmd.MarkFlagRequired("config-file")

	kmsIamMemberFlags(s.kmsEkmCfgAddIamCmd, &s.flagKmsEkmCfgIamMember, &s.flagKmsEkmCfgIamRole,
		&s.flagKmsEkmCfgIamCondExpr, &s.flagKmsEkmCfgIamCondTitle, &s.flagKmsEkmCfgIamCondDesc)
	kmsIamMemberFlags(s.kmsEkmCfgRemoveIamCmd, &s.flagKmsEkmCfgIamMember, &s.flagKmsEkmCfgIamRole,
		&s.flagKmsEkmCfgIamCondExpr, &s.flagKmsEkmCfgIamCondTitle, &s.flagKmsEkmCfgIamCondDesc)
	s.kmsEkmCfgRemoveIamCmd.Flags().BoolVar(&s.flagKmsEkmCfgAllConds, "all", false, "Match bindings for the role across all conditions")

	s.kmsEkmConfigCmd.AddCommand(all...)
	s.kmsCmd.AddCommand(s.kmsEkmConfigCmd)
}

func kmsEkmConfigName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/ekmConfig", project, location)
}

func (s *kmsCmds) runKmsEkmCfgDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsEkmConfigName(project, s.flagKmsEkmCfgLocation)
	out, err := svc.Projects.Locations.GetEkmConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing EkmConfig: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmCfgFormat)
}

func (s *kmsCmds) runKmsEkmCfgUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.EkmConfig{}
	if err := loadYAMLOrJSONInto(s.flagKmsEkmCfgConfigFile, body); err != nil {
		return err
	}
	mask := s.flagKmsEkmCfgMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := kmsEkmConfigName(project, s.flagKmsEkmCfgLocation)
	call := svc.Projects.Locations.UpdateEkmConfig(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating EkmConfig: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmCfgFormat)
}

func (s *kmsCmds) runKmsEkmCfgGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsEkmConfigName(project, s.flagKmsEkmCfgLocation)
	pol, err := svc.Projects.Locations.EkmConfig.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, pol, s.flagKmsEkmCfgFormat)
}

func (s *kmsCmds) runKmsEkmCfgSetIam(cmd *cobra.Command, args []string) error {
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
	if err := loadYAMLOrJSONInto(args[0], policy); err != nil {
		return err
	}
	name := kmsEkmConfigName(project, s.flagKmsEkmCfgLocation)
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.EkmConfig.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmCfgFormat)
}

func (s *kmsCmds) runKmsEkmCfgAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsEkmConfigName(project, s.flagKmsEkmCfgLocation)
	pol, err := svc.Projects.Locations.EkmConfig.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsEkmCfgIamCondExpr, s.flagKmsEkmCfgIamCondTitle, s.flagKmsEkmCfgIamCondDesc)
	kmsIamAddBinding(pol, s.flagKmsEkmCfgIamRole, s.flagKmsEkmCfgIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.EkmConfig.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmCfgFormat)
}

func (s *kmsCmds) runKmsEkmCfgRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsEkmConfigName(project, s.flagKmsEkmCfgLocation)
	pol, err := svc.Projects.Locations.EkmConfig.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsEkmCfgIamCondExpr, s.flagKmsEkmCfgIamCondTitle, s.flagKmsEkmCfgIamCondDesc)
	if !kmsIamRemoveBinding(pol, s.flagKmsEkmCfgIamRole, s.flagKmsEkmCfgIamMember, cond, s.flagKmsEkmCfgAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.EkmConfig.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmCfgFormat)
}
