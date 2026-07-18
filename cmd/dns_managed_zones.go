package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns managed-zones (#1537) ---

var dnsMZCmd = &cobra.Command{Use: "managed-zones", Short: "Manage managed zones"}

var (
	flagDNSMZFormat     string
	flagDNSMZConfigFile string
	flagDNSMZMaxResults int64
	flagDNSMZDnsName    string

	flagDNSMZIamMember   string
	flagDNSMZIamRole     string
	flagDNSMZIamCondExpr string
	flagDNSMZIamCondT    string
	flagDNSMZIamCondD    string
	flagDNSMZIamAllCond  bool
)

var (
	dnsMZCreateCmd = &cobra.Command{
		Use: "create ZONE", Short: "Create a managed zone",
		Args: cobra.ExactArgs(1), RunE: runDNSMZCreate,
	}
	dnsMZDeleteCmd = &cobra.Command{
		Use: "delete ZONE", Short: "Delete a managed zone",
		Args: cobra.ExactArgs(1), RunE: runDNSMZDelete,
	}
	dnsMZDescribeCmd = &cobra.Command{
		Use: "describe ZONE", Short: "Describe a managed zone",
		Args: cobra.ExactArgs(1), RunE: runDNSMZDescribe,
	}
	dnsMZListCmd = &cobra.Command{
		Use: "list", Short: "List managed zones",
		Args: cobra.NoArgs, RunE: runDNSMZList,
	}
	dnsMZUpdateCmd = &cobra.Command{
		Use: "update ZONE", Short: "Update a managed zone",
		Args: cobra.ExactArgs(1), RunE: runDNSMZUpdate,
	}
	dnsMZGetIamCmd = &cobra.Command{
		Use: "get-iam-policy ZONE", Short: "Get the IAM policy for a managed zone",
		Args: cobra.ExactArgs(1), RunE: runDNSMZGetIam,
	}
	dnsMZSetIamCmd = &cobra.Command{
		Use: "set-iam-policy ZONE POLICY_FILE", Short: "Set the IAM policy for a managed zone",
		Args: cobra.ExactArgs(2), RunE: runDNSMZSetIam,
	}
	dnsMZAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding ZONE", Short: "Add an IAM policy binding to a managed zone",
		Args: cobra.ExactArgs(1), RunE: runDNSMZAddIam,
	}
	dnsMZRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding ZONE", Short: "Remove an IAM policy binding from a managed zone",
		Args: cobra.ExactArgs(1), RunE: runDNSMZRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		dnsMZCreateCmd, dnsMZDeleteCmd, dnsMZDescribeCmd, dnsMZListCmd, dnsMZUpdateCmd,
		dnsMZGetIamCmd, dnsMZSetIamCmd, dnsMZAddIamCmd, dnsMZRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDNSMZFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dnsMZCreateCmd, dnsMZUpdateCmd} {
		c.Flags().StringVar(&flagDNSMZConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ManagedZone body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dnsMZListCmd.Flags().Int64Var(&flagDNSMZMaxResults, "limit", 0, "Maximum results per page")
	dnsMZListCmd.Flags().StringVar(&flagDNSMZDnsName, "dns-name", "", "Filter by DNS name suffix")
	for _, c := range []*cobra.Command{dnsMZAddIamCmd, dnsMZRemoveIamCmd} {
		dnsIamMemberFlags(c, &flagDNSMZIamMember, &flagDNSMZIamRole,
			&flagDNSMZIamCondExpr, &flagDNSMZIamCondT, &flagDNSMZIamCondD)
	}
	dnsMZRemoveIamCmd.Flags().BoolVar(&flagDNSMZIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	dnsMZCmd.AddCommand(all...)
	dnsCmd.AddCommand(dnsMZCmd)
}

func dnsMZResourceName(project, zone string) string {
	return fmt.Sprintf("projects/%s/managedZones/%s", project, zone)
}

func runDNSMZCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dns.ManagedZone{}
	if err := loadYAMLOrJSONInto(flagDNSMZConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ManagedZones.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating managed zone: %w", err)
	}
	fmt.Printf("Created managed zone [%s].\n", args[0])
	return emitFormatted(got, flagDNSMZFormat)
}

func runDNSMZDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if err := svc.ManagedZones.Delete(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting managed zone: %w", err)
	}
	fmt.Printf("Deleted managed zone [%s].\n", args[0])
	return nil
}

func runDNSMZDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ManagedZones.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing managed zone: %w", err)
	}
	return emitFormatted(got, flagDNSMZFormat)
}

func runDNSMZList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.ManagedZone
	pageToken := ""
	for {
		call := svc.ManagedZones.List(project).Context(ctx)
		if flagDNSMZMaxResults > 0 {
			call = call.MaxResults(flagDNSMZMaxResults)
		}
		if flagDNSMZDnsName != "" {
			call = call.DnsName(flagDNSMZDnsName)
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
	return emitFormatted(all, flagDNSMZFormat)
}

func runDNSMZUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dns.ManagedZone{}
	if err := loadYAMLOrJSONInto(flagDNSMZConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.ManagedZones.Patch(project, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating managed zone: %w", err)
	}
	fmt.Printf("Update request issued for managed zone [%s].\n", args[0])
	return emitFormatted(op, flagDNSMZFormat)
}

func runDNSMZGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
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
	return emitFormatted(policy, flagDNSMZFormat)
}

func runDNSMZSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &dns.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.ManagedZones.SetIamPolicy(dnsMZResourceName(project, args[0]),
		&dns.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dnsUpdatedIam(fmt.Sprintf("managed zone [%s]", args[0]))
	return emitFormatted(updated, flagDNSMZFormat)
}

func runDNSMZAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
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
	dnsAddBinding(policy, flagDNSMZIamRole, flagDNSMZIamMember,
		dnsBuildCondition(flagDNSMZIamCondExpr, flagDNSMZIamCondT, flagDNSMZIamCondD))
	policy.Version = 3
	updated, err := svc.ManagedZones.SetIamPolicy(resource, &dns.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dnsUpdatedIam(fmt.Sprintf("managed zone [%s]", args[0]))
	return emitFormatted(updated, flagDNSMZFormat)
}

func runDNSMZRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
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
	if !dnsRemoveBinding(policy, flagDNSMZIamRole, flagDNSMZIamMember,
		dnsBuildCondition(flagDNSMZIamCondExpr, flagDNSMZIamCondT, flagDNSMZIamCondD), flagDNSMZIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDNSMZIamRole, flagDNSMZIamMember)
	}
	updated, err := svc.ManagedZones.SetIamPolicy(resource, &dns.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dnsUpdatedIam(fmt.Sprintf("managed zone [%s]", args[0]))
	return emitFormatted(updated, flagDNSMZFormat)
}
