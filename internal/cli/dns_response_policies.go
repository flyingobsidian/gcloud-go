package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns response-policies (#1542) ---

func (s *dnsCmds) initDnsResponsePolicies() {
	base := []*cobra.Command{s.dnsRPCreateCmd, s.dnsRPDeleteCmd, s.dnsRPDescribeCmd, s.dnsRPListCmd, s.dnsRPUpdateCmd}
	for _, c := range base {
		c.Flags().StringVar(&s.flagDNSRPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dnsRPCreateCmd, s.dnsRPUpdateCmd} {
		c.Flags().StringVar(&s.flagDNSRPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ResponsePolicy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dnsRPListCmd.Flags().Int64Var(&s.flagDNSRPMaxResults, "limit", 0, "Maximum results per page")
	s.dnsRPCmd.AddCommand(base...)

	rules := []*cobra.Command{
		s.dnsRPRuleCreateCmd, s.dnsRPRuleDeleteCmd, s.dnsRPRuleDescribeCmd, s.dnsRPRuleListCmd, s.dnsRPRuleUpdateCmd,
	}
	for _, c := range rules {
		c.Flags().StringVar(&s.flagDNSRPRulePolicy, "response-policy", "",
			"Response policy that owns the rule (required)")
		_ = c.MarkFlagRequired("response-policy")
		c.Flags().StringVar(&s.flagDNSRPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dnsRPRuleCreateCmd, s.dnsRPRuleUpdateCmd} {
		c.Flags().StringVar(&s.flagDNSRPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ResponsePolicyRule body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dnsRPRuleListCmd.Flags().Int64Var(&s.flagDNSRPMaxResults, "limit", 0, "Maximum results per page")
	s.dnsRPRulesCmd.AddCommand(rules...)
	s.dnsRPCmd.AddCommand(s.dnsRPRulesCmd)

	s.dnsCmd.AddCommand(s.dnsRPCmd)
}

func (s *dnsCmds) runDNSRPCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dns.ResponsePolicy{}
	if err := loadYAMLOrJSONInto(s.flagDNSRPConfigFile, body); err != nil {
		return err
	}
	body.ResponsePolicyName = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicies.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating response policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created response policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRPFormat)
}

func (s *dnsCmds) runDNSRPDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	if err := svc.ResponsePolicies.Delete(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting response policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted response policy [%s].\n", args[0])
	return nil
}

func (s *dnsCmds) runDNSRPDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicies.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing response policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRPFormat)
}

func (s *dnsCmds) runDNSRPList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.ResponsePolicy
	pageToken := ""
	for {
		call := svc.ResponsePolicies.List(project).Context(ctx)
		if s.flagDNSRPMaxResults > 0 {
			call = call.MaxResults(s.flagDNSRPMaxResults)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing response policies: %w", err)
		}
		all = append(all, resp.ResponsePolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDNSRPFormat)
}

func (s *dnsCmds) runDNSRPUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dns.ResponsePolicy{}
	if err := loadYAMLOrJSONInto(s.flagDNSRPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicies.Patch(project, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating response policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated response policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRPFormat)
}

func (s *dnsCmds) runDNSRPRuleCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dns.ResponsePolicyRule{}
	if err := loadYAMLOrJSONInto(s.flagDNSRPConfigFile, body); err != nil {
		return err
	}
	body.RuleName = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicyRules.Create(project, s.flagDNSRPRulePolicy, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating response policy rule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created response policy rule [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRPFormat)
}

func (s *dnsCmds) runDNSRPRuleDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	if err := svc.ResponsePolicyRules.Delete(project, s.flagDNSRPRulePolicy, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting response policy rule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted response policy rule [%s].\n", args[0])
	return nil
}

func (s *dnsCmds) runDNSRPRuleDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicyRules.Get(project, s.flagDNSRPRulePolicy, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing response policy rule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRPFormat)
}

func (s *dnsCmds) runDNSRPRuleList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.ResponsePolicyRule
	pageToken := ""
	for {
		call := svc.ResponsePolicyRules.List(project, s.flagDNSRPRulePolicy).Context(ctx)
		if s.flagDNSRPMaxResults > 0 {
			call = call.MaxResults(s.flagDNSRPMaxResults)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing response policy rules: %w", err)
		}
		all = append(all, resp.ResponsePolicyRules...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDNSRPFormat)
}

func (s *dnsCmds) runDNSRPRuleUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dns.ResponsePolicyRule{}
	if err := loadYAMLOrJSONInto(s.flagDNSRPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicyRules.Patch(project, s.flagDNSRPRulePolicy, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating response policy rule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated response policy rule [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSRPFormat)
}
