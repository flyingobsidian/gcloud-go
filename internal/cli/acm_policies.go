package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager policies (#1445) ---

func (s *accessContextManagerCmds) initAcmPolicies() {
	all := []*cobra.Command{
		s.acmPOCreateCmd, s.acmPODeleteCmd, s.acmPODescribeCmd, s.acmPOListCmd, s.acmPOUpdateCmd,
		s.acmPOGetIamCmd, s.acmPOSetIamCmd, s.acmPOAddIamCmd, s.acmPORemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagACMPOFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.acmPOCreateCmd, s.acmPOListCmd} {
		c.Flags().StringVar(&s.flagACMPOOrganization, "organization", "", "Organization ID (required)")
		_ = c.MarkFlagRequired("organization")
	}
	for _, c := range []*cobra.Command{s.acmPOCreateCmd, s.acmPOUpdateCmd} {
		c.Flags().StringVar(&s.flagACMPOConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AccessPolicy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.acmPOListCmd.Flags().Int64Var(&s.flagACMPOPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{s.acmPOAddIamCmd, s.acmPORemoveIamCmd} {
		acmIamFlags(c, &s.flagACMPOIamMember, &s.flagACMPOIamRole,
			&s.flagACMPOIamCondExpr, &s.flagACMPOIamCondT, &s.flagACMPOIamCondD)
	}
	s.acmPORemoveIamCmd.Flags().BoolVar(&s.flagACMPOIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.acmPOCmd.AddCommand(all...)
	s.accessContextManagerCmd.AddCommand(s.acmPOCmd)
}

func (s *accessContextManagerCmds) runACMPOCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AccessPolicy{}
	if err := loadYAMLOrJSONInto(s.flagACMPOConfigFile, body); err != nil {
		return err
	}
	body.Parent = "organizations/" + s.flagACMPOOrganization
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.Create(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating access policy: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, "Create request issued for access policy.")
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPOFormat)
}

func (s *accessContextManagerCmds) runACMPODelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.Delete(acmPolicyResource(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting access policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for access policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPOFormat)
}

func (s *accessContextManagerCmds) runACMPODescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.Get(acmPolicyResource(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing access policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagACMPOFormat)
}

func (s *accessContextManagerCmds) runACMPOList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.AccessPolicy
	pageToken := ""
	for {
		call := svc.AccessPolicies.List().Parent("organizations/" + s.flagACMPOOrganization).Context(ctx)
		if s.flagACMPOPageSize > 0 {
			call = call.PageSize(s.flagACMPOPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing access policies: %w", err)
		}
		all = append(all, resp.AccessPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagACMPOFormat)
}

func (s *accessContextManagerCmds) runACMPOUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AccessPolicy{}
	if err := loadYAMLOrJSONInto(s.flagACMPOConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	call := svc.AccessPolicies.Patch(acmPolicyResource(args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating access policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for access policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPOFormat)
}

func (s *accessContextManagerCmds) runACMPOGetIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.AccessPolicies.GetIamPolicy(acmPolicyResource(args[0]),
		&accesscontextmanager.GetIamPolicyRequest{
			Options: &accesscontextmanager.GetPolicyOptions{RequestedPolicyVersion: 3},
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagACMPOFormat)
}

func (s *accessContextManagerCmds) runACMPOSetIam(cmd *cobra.Command, args []string) error {
	policy := &accesscontextmanager.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.AccessPolicies.SetIamPolicy(acmPolicyResource(args[0]),
		&accesscontextmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.acmUpdatedIam(fmt.Sprintf("access policy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagACMPOFormat)
}

func (s *accessContextManagerCmds) runACMPOAddIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	resource := acmPolicyResource(args[0])
	policy, err := svc.AccessPolicies.GetIamPolicy(resource, &accesscontextmanager.GetIamPolicyRequest{
		Options: &accesscontextmanager.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	acmAddBinding(policy, s.flagACMPOIamRole, s.flagACMPOIamMember,
		acmBuildCondition(s.flagACMPOIamCondExpr, s.flagACMPOIamCondT, s.flagACMPOIamCondD))
	policy.Version = 3
	updated, err := svc.AccessPolicies.SetIamPolicy(resource,
		&accesscontextmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.acmUpdatedIam(fmt.Sprintf("access policy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagACMPOFormat)
}

func (s *accessContextManagerCmds) runACMPORemoveIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	resource := acmPolicyResource(args[0])
	policy, err := svc.AccessPolicies.GetIamPolicy(resource, &accesscontextmanager.GetIamPolicyRequest{
		Options: &accesscontextmanager.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !acmRemoveBinding(policy, s.flagACMPOIamRole, s.flagACMPOIamMember,
		acmBuildCondition(s.flagACMPOIamCondExpr, s.flagACMPOIamCondT, s.flagACMPOIamCondD), s.flagACMPOIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagACMPOIamRole, s.flagACMPOIamMember)
	}
	updated, err := svc.AccessPolicies.SetIamPolicy(resource,
		&accesscontextmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.acmUpdatedIam(fmt.Sprintf("access policy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagACMPOFormat)
}
