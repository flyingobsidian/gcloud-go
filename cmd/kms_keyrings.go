package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms keyrings (#1106) ---

func (s *kmsCmds) initKmsKeyrings() {
	all := []*cobra.Command{
		s.kmsKRCreateCmd, s.kmsKRDeleteCmd, s.kmsKRDescribeCmd, s.kmsKRListCmd,
		s.kmsKRGetIamCmd, s.kmsKRSetIamCmd, s.kmsKRAddIamCmd, s.kmsKRRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagKmsKRLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsKRFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	s.kmsKRCreateCmd.Flags().StringVar(&s.flagKmsKRConfigFile, "config-file", "", "YAML/JSON body for the KeyRing (optional; empty is valid)")

	s.kmsKRListCmd.Flags().StringVar(&s.flagKmsKRFilter, "filter", "", "Filter expression")
	s.kmsKRListCmd.Flags().Int64Var(&s.flagKmsKRPageSize, "page-size", 0, "Page size")

	kmsIamMemberFlags(s.kmsKRAddIamCmd, &s.flagKmsKRIamMember, &s.flagKmsKRIamRole,
		&s.flagKmsKRIamCondExpr, &s.flagKmsKRIamCondTitle, &s.flagKmsKRIamCondDesc)
	kmsIamMemberFlags(s.kmsKRRemoveIamCmd, &s.flagKmsKRIamMember, &s.flagKmsKRIamRole,
		&s.flagKmsKRIamCondExpr, &s.flagKmsKRIamCondTitle, &s.flagKmsKRIamCondDesc)
	s.kmsKRRemoveIamCmd.Flags().BoolVar(&s.flagKmsKRAllConds, "all", false, "Match bindings for the role across all conditions")

	s.kmsKeyringsCmd.AddCommand(all...)
	s.kmsCmd.AddCommand(s.kmsKeyringsCmd)
}

func kmsKRName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/keyRings"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsKRCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.KeyRing{}
	if s.flagKmsKRConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagKmsKRConfigFile, body); err != nil {
			return err
		}
	}
	parent := kmsLocationParent(project, s.flagKmsKRLocation)
	out, err := svc.Projects.Locations.KeyRings.Create(parent, body).
		KeyRingId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating key ring: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKRFormat)
}

// runKmsKRDelete mirrors gcloud-python's keyrings delete (589.0.0): a
// DeleteCommand that issues the request and prints nothing unless --format.
func (s *kmsCmds) runKmsKRDelete(cmd *cobra.Command, args []string) error {
	if args[0] == "" {
		return fmt.Errorf("Invalid value for [keyring]: keyring id must be non-empty.")
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKRName(project, s.flagKmsKRLocation, args[0])
	op, err := svc.Projects.Locations.KeyRings.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting key ring: %w", err)
	}
	if s.flagKmsKRFormat != "" {
		return emitFormattedTo(s.inv.stdout, op, s.flagKmsKRFormat)
	}
	return nil
}

func (s *kmsCmds) runKmsKRDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKRName(project, s.flagKmsKRLocation, args[0])
	out, err := svc.Projects.Locations.KeyRings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing key ring: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKRFormat)
}

func (s *kmsCmds) runKmsKRList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsKRLocation)
	var all []*cloudkms.KeyRing
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.List(parent).Context(ctx)
		if s.flagKmsKRFilter != "" {
			call = call.Filter(s.flagKmsKRFilter)
		}
		if s.flagKmsKRPageSize > 0 {
			call = call.PageSize(s.flagKmsKRPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsKRFormat)
}

func (s *kmsCmds) runKmsKRGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKRName(project, s.flagKmsKRLocation, args[0])
	pol, err := svc.Projects.Locations.KeyRings.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, pol, s.flagKmsKRFormat)
}

func (s *kmsCmds) runKmsKRSetIam(cmd *cobra.Command, args []string) error {
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
	name := kmsKRName(project, s.flagKmsKRLocation, args[0])
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.KeyRings.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKRFormat)
}

func (s *kmsCmds) runKmsKRAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKRName(project, s.flagKmsKRLocation, args[0])
	pol, err := svc.Projects.Locations.KeyRings.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsKRIamCondExpr, s.flagKmsKRIamCondTitle, s.flagKmsKRIamCondDesc)
	kmsIamAddBinding(pol, s.flagKmsKRIamRole, s.flagKmsKRIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKRFormat)
}

func (s *kmsCmds) runKmsKRRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKRName(project, s.flagKmsKRLocation, args[0])
	pol, err := svc.Projects.Locations.KeyRings.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsKRIamCondExpr, s.flagKmsKRIamCondTitle, s.flagKmsKRIamCondDesc)
	if !kmsIamRemoveBinding(pol, s.flagKmsKRIamRole, s.flagKmsKRIamMember, cond, s.flagKmsKRAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKRFormat)
}
