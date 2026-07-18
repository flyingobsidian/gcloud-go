package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns response-policies (#1542) ---

var dnsRPCmd = &cobra.Command{Use: "response-policies", Short: "Manage Cloud DNS response policies"}

var (
	flagDNSRPFormat     string
	flagDNSRPConfigFile string
	flagDNSRPMaxResults int64

	dnsRPRulesCmd = &cobra.Command{Use: "rules", Short: "Manage response policy rules"}

	flagDNSRPRulePolicy string
)

var (
	dnsRPCreateCmd = &cobra.Command{
		Use: "create RESPONSE_POLICY", Short: "Create a response policy",
		Args: cobra.ExactArgs(1), RunE: runDNSRPCreate,
	}
	dnsRPDeleteCmd = &cobra.Command{
		Use: "delete RESPONSE_POLICY", Short: "Delete a response policy",
		Args: cobra.ExactArgs(1), RunE: runDNSRPDelete,
	}
	dnsRPDescribeCmd = &cobra.Command{
		Use: "describe RESPONSE_POLICY", Short: "Describe a response policy",
		Args: cobra.ExactArgs(1), RunE: runDNSRPDescribe,
	}
	dnsRPListCmd = &cobra.Command{
		Use: "list", Short: "List response policies",
		Args: cobra.NoArgs, RunE: runDNSRPList,
	}
	dnsRPUpdateCmd = &cobra.Command{
		Use: "update RESPONSE_POLICY", Short: "Update a response policy",
		Args: cobra.ExactArgs(1), RunE: runDNSRPUpdate,
	}

	dnsRPRuleCreateCmd = &cobra.Command{
		Use: "create RULE", Short: "Create a response policy rule",
		Args: cobra.ExactArgs(1), RunE: runDNSRPRuleCreate,
	}
	dnsRPRuleDeleteCmd = &cobra.Command{
		Use: "delete RULE", Short: "Delete a response policy rule",
		Args: cobra.ExactArgs(1), RunE: runDNSRPRuleDelete,
	}
	dnsRPRuleDescribeCmd = &cobra.Command{
		Use: "describe RULE", Short: "Describe a response policy rule",
		Args: cobra.ExactArgs(1), RunE: runDNSRPRuleDescribe,
	}
	dnsRPRuleListCmd = &cobra.Command{
		Use: "list", Short: "List response policy rules",
		Args: cobra.NoArgs, RunE: runDNSRPRuleList,
	}
	dnsRPRuleUpdateCmd = &cobra.Command{
		Use: "update RULE", Short: "Update a response policy rule",
		Args: cobra.ExactArgs(1), RunE: runDNSRPRuleUpdate,
	}
)

func init() {
	base := []*cobra.Command{dnsRPCreateCmd, dnsRPDeleteCmd, dnsRPDescribeCmd, dnsRPListCmd, dnsRPUpdateCmd}
	for _, c := range base {
		c.Flags().StringVar(&flagDNSRPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dnsRPCreateCmd, dnsRPUpdateCmd} {
		c.Flags().StringVar(&flagDNSRPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ResponsePolicy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dnsRPListCmd.Flags().Int64Var(&flagDNSRPMaxResults, "limit", 0, "Maximum results per page")
	dnsRPCmd.AddCommand(base...)

	rules := []*cobra.Command{
		dnsRPRuleCreateCmd, dnsRPRuleDeleteCmd, dnsRPRuleDescribeCmd, dnsRPRuleListCmd, dnsRPRuleUpdateCmd,
	}
	for _, c := range rules {
		c.Flags().StringVar(&flagDNSRPRulePolicy, "response-policy", "",
			"Response policy that owns the rule (required)")
		_ = c.MarkFlagRequired("response-policy")
		c.Flags().StringVar(&flagDNSRPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dnsRPRuleCreateCmd, dnsRPRuleUpdateCmd} {
		c.Flags().StringVar(&flagDNSRPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ResponsePolicyRule body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dnsRPRuleListCmd.Flags().Int64Var(&flagDNSRPMaxResults, "limit", 0, "Maximum results per page")
	dnsRPRulesCmd.AddCommand(rules...)
	dnsRPCmd.AddCommand(dnsRPRulesCmd)

	dnsCmd.AddCommand(dnsRPCmd)
}

func runDNSRPCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dns.ResponsePolicy{}
	if err := loadYAMLOrJSONInto(flagDNSRPConfigFile, body); err != nil {
		return err
	}
	body.ResponsePolicyName = args[0]
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicies.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating response policy: %w", err)
	}
	fmt.Printf("Created response policy [%s].\n", args[0])
	return emitFormatted(got, flagDNSRPFormat)
}

func runDNSRPDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if err := svc.ResponsePolicies.Delete(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting response policy: %w", err)
	}
	fmt.Printf("Deleted response policy [%s].\n", args[0])
	return nil
}

func runDNSRPDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicies.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing response policy: %w", err)
	}
	return emitFormatted(got, flagDNSRPFormat)
}

func runDNSRPList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.ResponsePolicy
	pageToken := ""
	for {
		call := svc.ResponsePolicies.List(project).Context(ctx)
		if flagDNSRPMaxResults > 0 {
			call = call.MaxResults(flagDNSRPMaxResults)
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
	return emitFormatted(all, flagDNSRPFormat)
}

func runDNSRPUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dns.ResponsePolicy{}
	if err := loadYAMLOrJSONInto(flagDNSRPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicies.Patch(project, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating response policy: %w", err)
	}
	fmt.Printf("Updated response policy [%s].\n", args[0])
	return emitFormatted(got, flagDNSRPFormat)
}

func runDNSRPRuleCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dns.ResponsePolicyRule{}
	if err := loadYAMLOrJSONInto(flagDNSRPConfigFile, body); err != nil {
		return err
	}
	body.RuleName = args[0]
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicyRules.Create(project, flagDNSRPRulePolicy, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating response policy rule: %w", err)
	}
	fmt.Printf("Created response policy rule [%s].\n", args[0])
	return emitFormatted(got, flagDNSRPFormat)
}

func runDNSRPRuleDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if err := svc.ResponsePolicyRules.Delete(project, flagDNSRPRulePolicy, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting response policy rule: %w", err)
	}
	fmt.Printf("Deleted response policy rule [%s].\n", args[0])
	return nil
}

func runDNSRPRuleDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicyRules.Get(project, flagDNSRPRulePolicy, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing response policy rule: %w", err)
	}
	return emitFormatted(got, flagDNSRPFormat)
}

func runDNSRPRuleList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.ResponsePolicyRule
	pageToken := ""
	for {
		call := svc.ResponsePolicyRules.List(project, flagDNSRPRulePolicy).Context(ctx)
		if flagDNSRPMaxResults > 0 {
			call = call.MaxResults(flagDNSRPMaxResults)
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
	return emitFormatted(all, flagDNSRPFormat)
}

func runDNSRPRuleUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dns.ResponsePolicyRule{}
	if err := loadYAMLOrJSONInto(flagDNSRPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ResponsePolicyRules.Patch(project, flagDNSRPRulePolicy, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating response policy rule: %w", err)
	}
	fmt.Printf("Updated response policy rule [%s].\n", args[0])
	return emitFormatted(got, flagDNSRPFormat)
}
