package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms ekm-connections (#1102) ---

func (s *kmsCmds) initKmsEkmConnections() {
	all := []*cobra.Command{
		s.kmsEkmConnCreateCmd, s.kmsEkmConnDescribeCmd, s.kmsEkmConnListCmd, s.kmsEkmConnUpdateCmd,
		s.kmsEkmConnGetIamCmd, s.kmsEkmConnSetIamCmd, s.kmsEkmConnAddIamCmd, s.kmsEkmConnRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagKmsEkmConnLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsEkmConnFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}

	s.kmsEkmConnCreateCmd.Flags().StringVar(&s.flagKmsEkmConnConfigFile, "config-file", "", "YAML/JSON body for the EkmConnection (required)")
	_ = s.kmsEkmConnCreateCmd.MarkFlagRequired("config-file")

	s.kmsEkmConnListCmd.Flags().StringVar(&s.flagKmsEkmConnFilter, "filter", "", "Filter expression")
	s.kmsEkmConnListCmd.Flags().Int64Var(&s.flagKmsEkmConnPageSize, "page-size", 0, "Page size")

	s.kmsEkmConnUpdateCmd.Flags().StringVar(&s.flagKmsEkmConnConfigFile, "config-file", "", "YAML/JSON body for the EkmConnection update (required)")
	s.kmsEkmConnUpdateCmd.Flags().StringVar(&s.flagKmsEkmConnMask, "update-mask", "", "Fields to update; defaults to populated fields")
	_ = s.kmsEkmConnUpdateCmd.MarkFlagRequired("config-file")

	kmsIamMemberFlags(s.kmsEkmConnAddIamCmd, &s.flagKmsEkmConnIamMember, &s.flagKmsEkmConnIamRole,
		&s.flagKmsEkmConnIamCondExpr, &s.flagKmsEkmConnIamCondTitle, &s.flagKmsEkmConnIamCondDesc)
	kmsIamMemberFlags(s.kmsEkmConnRemoveIamCmd, &s.flagKmsEkmConnIamMember, &s.flagKmsEkmConnIamRole,
		&s.flagKmsEkmConnIamCondExpr, &s.flagKmsEkmConnIamCondTitle, &s.flagKmsEkmConnIamCondDesc)
	s.kmsEkmConnRemoveIamCmd.Flags().BoolVar(&s.flagKmsEkmConnAllConds, "all", false, "Match bindings for the role across all conditions")

	s.kmsEkmConnectionsCmd.AddCommand(all...)
	s.kmsCmd.AddCommand(s.kmsEkmConnectionsCmd)
}

func kmsEkmConnName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/ekmConnections"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsEkmConnCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.EkmConnection{}
	if err := loadYAMLOrJSONInto(s.flagKmsEkmConnConfigFile, body); err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsEkmConnLocation)
	out, err := svc.Projects.Locations.EkmConnections.Create(parent, body).
		EkmConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating EKM connection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmConnFormat)
}

func (s *kmsCmds) runKmsEkmConnDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsEkmConnName(project, s.flagKmsEkmConnLocation, args[0])
	out, err := svc.Projects.Locations.EkmConnections.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing EKM connection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmConnFormat)
}

func (s *kmsCmds) runKmsEkmConnList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsEkmConnLocation)
	var all []*cloudkms.EkmConnection
	token := ""
	for {
		call := svc.Projects.Locations.EkmConnections.List(parent).Context(ctx)
		if s.flagKmsEkmConnFilter != "" {
			call = call.Filter(s.flagKmsEkmConnFilter)
		}
		if s.flagKmsEkmConnPageSize > 0 {
			call = call.PageSize(s.flagKmsEkmConnPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsEkmConnFormat)
}

func (s *kmsCmds) runKmsEkmConnUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.EkmConnection{}
	if err := loadYAMLOrJSONInto(s.flagKmsEkmConnConfigFile, body); err != nil {
		return err
	}
	mask := s.flagKmsEkmConnMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := kmsEkmConnName(project, s.flagKmsEkmConnLocation, args[0])
	call := svc.Projects.Locations.EkmConnections.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating EKM connection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmConnFormat)
}

func (s *kmsCmds) runKmsEkmConnGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsEkmConnName(project, s.flagKmsEkmConnLocation, args[0])
	pol, err := svc.Projects.Locations.EkmConnections.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, pol, s.flagKmsEkmConnFormat)
}

func (s *kmsCmds) runKmsEkmConnSetIam(cmd *cobra.Command, args []string) error {
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
	name := kmsEkmConnName(project, s.flagKmsEkmConnLocation, args[0])
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.EkmConnections.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmConnFormat)
}

func (s *kmsCmds) runKmsEkmConnAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsEkmConnName(project, s.flagKmsEkmConnLocation, args[0])
	pol, err := svc.Projects.Locations.EkmConnections.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsEkmConnIamCondExpr, s.flagKmsEkmConnIamCondTitle, s.flagKmsEkmConnIamCondDesc)
	kmsIamAddBinding(pol, s.flagKmsEkmConnIamRole, s.flagKmsEkmConnIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.EkmConnections.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmConnFormat)
}

func (s *kmsCmds) runKmsEkmConnRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsEkmConnName(project, s.flagKmsEkmConnLocation, args[0])
	pol, err := svc.Projects.Locations.EkmConnections.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsEkmConnIamCondExpr, s.flagKmsEkmConnIamCondTitle, s.flagKmsEkmConnIamCondDesc)
	if !kmsIamRemoveBinding(pol, s.flagKmsEkmConnIamRole, s.flagKmsEkmConnIamMember, cond, s.flagKmsEkmConnAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.EkmConnections.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsEkmConnFormat)
}
