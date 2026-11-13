package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap tcp dest-groups (#1068) ---

func (s *iapCmds) initIapTcpDestGroups() {
	all := []*cobra.Command{
		s.iapDGCreateCmd, s.iapDGDeleteCmd, s.iapDGDescribeCmd, s.iapDGListCmd, s.iapDGUpdateCmd,
		s.iapDGGetIamCmd, s.iapDGSetIamCmd, s.iapDGAddIamCmd, s.iapDGRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagIapDGLocation, "location", "", "IAP TCP tunnel location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagIapDGFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.iapDGCreateCmd, s.iapDGUpdateCmd} {
		c.Flags().StringVar(&s.flagIapDGConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the TunnelDestGroup body")
		c.Flags().StringSliceVar(&s.flagIapDGCidrs, "cidrs", nil, "Comma-separated list of CIDR blocks")
		c.Flags().StringSliceVar(&s.flagIapDGFqdns, "fqdns", nil, "Comma-separated list of FQDNs")
	}
	s.iapDGUpdateCmd.Flags().StringVar(&s.flagIapDGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.iapDGListCmd.Flags().Int64Var(&s.flagIapDGPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{s.iapDGAddIamCmd, s.iapDGRemoveIamCmd} {
		iapIamMemberFlags(c, &s.flagIapDGIamMember, &s.flagIapDGIamRole,
			&s.flagIapDGIamCondExpr, &s.flagIapDGIamCondT, &s.flagIapDGIamCondD)
	}
	s.iapDGRemoveIamCmd.Flags().BoolVar(&s.flagIapDGIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.iapDestGroupsCmd.AddCommand(all...)
	s.iapTcpCmd.AddCommand(s.iapDestGroupsCmd)
}

func (s *iapCmds) iapDGParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/iap_tunnel/locations/%s", project, s.flagIapDGLocation), nil
}

func (s *iapCmds) iapDGName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.iapDGParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/destGroups/%s", parent, id), nil
}

func (s *iapCmds) iapDGBodyFromFlags() (*iap.TunnelDestGroup, error) {
	body := &iap.TunnelDestGroup{}
	if s.flagIapDGConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagIapDGConfigFile, body); err != nil {
			return nil, err
		}
	}
	if len(s.flagIapDGCidrs) > 0 {
		body.Cidrs = s.flagIapDGCidrs
	}
	if len(s.flagIapDGFqdns) > 0 {
		body.Fqdns = s.flagIapDGFqdns
	}
	return body, nil
}

func (s *iapCmds) runIapDGCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.iapDGParent()
	if err != nil {
		return err
	}
	body, err := s.iapDGBodyFromFlags()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.IapTunnel.Locations.DestGroups.Create(parent, body).TunnelDestGroupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating IAP TCP destination group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created IAP TCP destination group [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagIapDGFormat)
}

func (s *iapCmds) runIapDGDelete(cmd *cobra.Command, args []string) error {
	name, err := s.iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.IapTunnel.Locations.DestGroups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting IAP TCP destination group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted IAP TCP destination group [%s].\n", args[0])
	return nil
}

func (s *iapCmds) runIapDGDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.IapTunnel.Locations.DestGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing IAP TCP destination group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIapDGFormat)
}

func (s *iapCmds) runIapDGList(cmd *cobra.Command, args []string) error {
	parent, err := s.iapDGParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	var all []*iap.TunnelDestGroup
	pageToken := ""
	for {
		call := svc.Projects.IapTunnel.Locations.DestGroups.List(parent).Context(ctx)
		if s.flagIapDGPageSize > 0 {
			call = call.PageSize(s.flagIapDGPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing IAP TCP destination groups: %w", err)
		}
		all = append(all, resp.TunnelDestGroups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagIapDGFormat)
}

func (s *iapCmds) runIapDGUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.iapDGName(args[0])
	if err != nil {
		return err
	}
	body, err := s.iapDGBodyFromFlags()
	if err != nil {
		return err
	}
	mask := s.flagIapDGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.IapTunnel.Locations.DestGroups.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating IAP TCP destination group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated IAP TCP destination group [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagIapDGFormat)
}

// The IAP TCP destination-groups service does not expose per-resource IAM
// methods; instead policies are attached to the resource via the shared v1
// IAM endpoints, addressed by the destination-group's full resource name.

func (s *iapCmds) runIapDGGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.V1.GetIamPolicy(name, &iap.GetIamPolicyRequest{
		Options: &iap.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagIapDGFormat)
}

func (s *iapCmds) runIapDGSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.iapDGName(args[0])
	if err != nil {
		return err
	}
	policy := &iap.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.V1.SetIamPolicy(name, &iap.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.iapUpdatedIam(fmt.Sprintf("IAP TCP destination group [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagIapDGFormat)
}

func (s *iapCmds) runIapDGAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.V1.GetIamPolicy(name, &iap.GetIamPolicyRequest{
		Options: &iap.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	iapIamAddBinding(policy, s.flagIapDGIamRole, s.flagIapDGIamMember,
		iapIamBuildCondition(s.flagIapDGIamCondExpr, s.flagIapDGIamCondT, s.flagIapDGIamCondD))
	policy.Version = 3
	updated, err := svc.V1.SetIamPolicy(name, &iap.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.iapUpdatedIam(fmt.Sprintf("IAP TCP destination group [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagIapDGFormat)
}

func (s *iapCmds) runIapDGRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.V1.GetIamPolicy(name, &iap.GetIamPolicyRequest{
		Options: &iap.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !iapIamRemoveBinding(policy, s.flagIapDGIamRole, s.flagIapDGIamMember,
		iapIamBuildCondition(s.flagIapDGIamCondExpr, s.flagIapDGIamCondT, s.flagIapDGIamCondD), s.flagIapDGIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			s.flagIapDGIamRole, s.flagIapDGIamMember)
	}
	updated, err := svc.V1.SetIamPolicy(name, &iap.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.iapUpdatedIam(fmt.Sprintf("IAP TCP destination group [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagIapDGFormat)
}
