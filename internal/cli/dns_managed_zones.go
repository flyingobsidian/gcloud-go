package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns managed-zones (#1537) ---

func (s *dnsCmds) initDnsManagedZones() {
	all := []*cobra.Command{
		s.dnsMZCreateCmd, s.dnsMZDeleteCmd, s.dnsMZDescribeCmd, s.dnsMZListCmd, s.dnsMZUpdateCmd,
		s.dnsMZGetIamCmd, s.dnsMZSetIamCmd, s.dnsMZAddIamCmd, s.dnsMZRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDNSMZFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dnsMZCreateCmd, s.dnsMZUpdateCmd} {
		c.Flags().StringVar(&s.flagDNSMZConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ManagedZone body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dnsMZListCmd.Flags().Int64Var(&s.flagDNSMZMaxResults, "limit", 0, "Maximum results per page")
	s.dnsMZListCmd.Flags().StringVar(&s.flagDNSMZDnsName, "dns-name", "", "Filter by DNS name suffix")
	for _, c := range []*cobra.Command{s.dnsMZAddIamCmd, s.dnsMZRemoveIamCmd} {
		dnsIamMemberFlags(c, &s.flagDNSMZIamMember, &s.flagDNSMZIamRole,
			&s.flagDNSMZIamCondExpr, &s.flagDNSMZIamCondT, &s.flagDNSMZIamCondD)
	}
	s.dnsMZRemoveIamCmd.Flags().BoolVar(&s.flagDNSMZIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.dnsMZCmd.AddCommand(all...)
	s.dnsCmd.AddCommand(s.dnsMZCmd)
}

func dnsMZResourceName(project, zone string) string {
	return fmt.Sprintf("projects/%s/managedZones/%s", project, zone)
}

func (s *dnsCmds) runDNSMZCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dns.ManagedZone{}
	if err := loadYAMLOrJSONInto(s.flagDNSMZConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ManagedZones.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating managed zone: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created managed zone [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSMZFormat)
}

func (s *dnsCmds) runDNSMZDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	if err := svc.ManagedZones.Delete(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting managed zone: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted managed zone [%s].\n", args[0])
	return nil
}

func (s *dnsCmds) runDNSMZDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ManagedZones.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing managed zone: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSMZFormat)
}

func (s *dnsCmds) runDNSMZList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.ManagedZone
	pageToken := ""
	for {
		call := svc.ManagedZones.List(project).Context(ctx)
		if s.flagDNSMZMaxResults > 0 {
			call = call.MaxResults(s.flagDNSMZMaxResults)
		}
		if s.flagDNSMZDnsName != "" {
			call = call.DnsName(s.flagDNSMZDnsName)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing managed zones: %w", err)
		}
		all = append(all, resp.ManagedZones...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDNSMZFormat)
}

func (s *dnsCmds) runDNSMZUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dns.ManagedZone{}
	if err := loadYAMLOrJSONInto(s.flagDNSMZConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	op, err := svc.ManagedZones.Patch(project, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating managed zone: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for managed zone [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagDNSMZFormat)
}

func (s *dnsCmds) runDNSMZGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.ManagedZones.GetIamPolicy(dnsMZResourceName(project, args[0]),
		&dns.GoogleIamV1GetIamPolicyRequest{
			Options: &dns.GoogleIamV1GetPolicyOptions{RequestedPolicyVersion: 3},
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDNSMZFormat)
}

func (s *dnsCmds) runDNSMZSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &dns.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.ManagedZones.SetIamPolicy(dnsMZResourceName(project, args[0]),
		&dns.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dnsUpdatedIam(fmt.Sprintf("managed zone [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDNSMZFormat)
}

func (s *dnsCmds) runDNSMZAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	resource := dnsMZResourceName(project, args[0])
	policy, err := svc.ManagedZones.GetIamPolicy(resource, &dns.GoogleIamV1GetIamPolicyRequest{
		Options: &dns.GoogleIamV1GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dnsAddBinding(policy, s.flagDNSMZIamRole, s.flagDNSMZIamMember,
		dnsBuildCondition(s.flagDNSMZIamCondExpr, s.flagDNSMZIamCondT, s.flagDNSMZIamCondD))
	policy.Version = 3
	updated, err := svc.ManagedZones.SetIamPolicy(resource, &dns.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dnsUpdatedIam(fmt.Sprintf("managed zone [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDNSMZFormat)
}

func (s *dnsCmds) runDNSMZRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	resource := dnsMZResourceName(project, args[0])
	policy, err := svc.ManagedZones.GetIamPolicy(resource, &dns.GoogleIamV1GetIamPolicyRequest{
		Options: &dns.GoogleIamV1GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dnsRemoveBinding(policy, s.flagDNSMZIamRole, s.flagDNSMZIamMember,
		dnsBuildCondition(s.flagDNSMZIamCondExpr, s.flagDNSMZIamCondT, s.flagDNSMZIamCondD), s.flagDNSMZIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDNSMZIamRole, s.flagDNSMZIamMember)
	}
	updated, err := svc.ManagedZones.SetIamPolicy(resource, &dns.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dnsUpdatedIam(fmt.Sprintf("managed zone [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDNSMZFormat)
}
