package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns policies (#1539) ---

var dnsPolCmd = &cobra.Command{Use: "policies", Short: "Manage Cloud DNS policies"}

var (
	flagDNSPolFormat     string
	flagDNSPolConfigFile string
	flagDNSPolMaxResults int64
)

var (
	dnsPolCreateCmd = &cobra.Command{
		Use: "create POLICY", Short: "Create a DNS policy",
		Args: cobra.ExactArgs(1), RunE: runDNSPolCreate,
	}
	dnsPolDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete a DNS policy",
		Args: cobra.ExactArgs(1), RunE: runDNSPolDelete,
	}
	dnsPolDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe a DNS policy",
		Args: cobra.ExactArgs(1), RunE: runDNSPolDescribe,
	}
	dnsPolListCmd = &cobra.Command{
		Use: "list", Short: "List DNS policies",
		Args: cobra.NoArgs, RunE: runDNSPolList,
	}
	dnsPolUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update a DNS policy",
		Args: cobra.ExactArgs(1), RunE: runDNSPolUpdate,
	}
)

func init() {
	all := []*cobra.Command{dnsPolCreateCmd, dnsPolDeleteCmd, dnsPolDescribeCmd, dnsPolListCmd, dnsPolUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDNSPolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dnsPolCreateCmd, dnsPolUpdateCmd} {
		c.Flags().StringVar(&flagDNSPolConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Policy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dnsPolListCmd.Flags().Int64Var(&flagDNSPolMaxResults, "limit", 0, "Maximum results per page")

	dnsPolCmd.AddCommand(all...)
	dnsCmd.AddCommand(dnsPolCmd)
}

func runDNSPolCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dns.Policy{}
	if err := loadYAMLOrJSONInto(flagDNSPolConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Policies.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DNS policy: %w", err)
	}
	fmt.Printf("Created DNS policy [%s].\n", args[0])
	return emitFormatted(got, flagDNSPolFormat)
}

func runDNSPolDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if err := svc.Policies.Delete(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting DNS policy: %w", err)
	}
	fmt.Printf("Deleted DNS policy [%s].\n", args[0])
	return nil
}

func runDNSPolDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Policies.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS policy: %w", err)
	}
	return emitFormatted(got, flagDNSPolFormat)
}

func runDNSPolList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.Policy
	pageToken := ""
	for {
		call := svc.Policies.List(project).Context(ctx)
		if flagDNSPolMaxResults > 0 {
			call = call.MaxResults(flagDNSPolMaxResults)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing DNS policies: %w", err)
		}
		all = append(all, resp.Policies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDNSPolFormat)
}

func runDNSPolUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dns.Policy{}
	if err := loadYAMLOrJSONInto(flagDNSPolConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Policies.Patch(project, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating DNS policy: %w", err)
	}
	fmt.Printf("Updated DNS policy [%s].\n", args[0])
	return emitFormatted(got, flagDNSPolFormat)
}
