package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns operations (#1538) ---

var dnsOpCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud DNS operations"}

var (
	flagDNSOpZone       string
	flagDNSOpFormat     string
	flagDNSOpMaxResults int64
)

var (
	dnsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Cloud DNS operation",
		Args: cobra.ExactArgs(1), RunE: runDNSOpDescribe,
	}
	dnsOpListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud DNS operations for a managed zone",
		Args: cobra.NoArgs, RunE: runDNSOpList,
	}
)

func init() {
	for _, c := range []*cobra.Command{dnsOpDescribeCmd, dnsOpListCmd} {
		c.Flags().StringVar(&flagDNSOpZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&flagDNSOpFormat, "format", "", "Output format")
	}
	dnsOpListCmd.Flags().Int64Var(&flagDNSOpMaxResults, "limit", 0, "Maximum results per page")

	dnsOpCmd.AddCommand(dnsOpDescribeCmd, dnsOpListCmd)
	dnsCmd.AddCommand(dnsOpCmd)
}

func runDNSOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ManagedZoneOperations.Get(project, flagDNSOpZone, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS operation: %w", err)
	}
	return emitFormatted(got, flagDNSOpFormat)
}

func runDNSOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.Operation
	pageToken := ""
	for {
		call := svc.ManagedZoneOperations.List(project, flagDNSOpZone).Context(ctx)
		if flagDNSOpMaxResults > 0 {
			call = call.MaxResults(flagDNSOpMaxResults)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing DNS operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDNSOpFormat)
}
