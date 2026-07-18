package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns dns-keys (#1536) ---

var dnsKeysCmd = &cobra.Command{Use: "dns-keys", Short: "Manage DNSKEY records"}

var (
	flagDNSKeyZone       string
	flagDNSKeyFormat     string
	flagDNSKeyMaxResults int64
	flagDNSKeyDigestType string
)

var (
	dnsKeysDescribeCmd = &cobra.Command{
		Use: "describe KEY_ID", Short: "Describe a DNSKEY",
		Args: cobra.ExactArgs(1), RunE: runDNSKeyDescribe,
	}
	dnsKeysListCmd = &cobra.Command{
		Use: "list", Short: "List DNSKEYs for a managed zone",
		Args: cobra.NoArgs, RunE: runDNSKeyList,
	}
)

func init() {
	for _, c := range []*cobra.Command{dnsKeysDescribeCmd, dnsKeysListCmd} {
		c.Flags().StringVar(&flagDNSKeyZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&flagDNSKeyFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagDNSKeyDigestType, "digest-type", "",
			"Digest type used when generating the DS record")
	}
	dnsKeysListCmd.Flags().Int64Var(&flagDNSKeyMaxResults, "limit", 0, "Maximum results per page")

	dnsKeysCmd.AddCommand(dnsKeysDescribeCmd, dnsKeysListCmd)
	dnsCmd.AddCommand(dnsKeysCmd)
}

func runDNSKeyDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.DnsKeys.Get(project, flagDNSKeyZone, args[0]).Context(ctx)
	if flagDNSKeyDigestType != "" {
		call = call.DigestType(flagDNSKeyDigestType)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing DNSKEY: %w", err)
	}
	return emitFormatted(got, flagDNSKeyFormat)
}

func runDNSKeyList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DNSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*dns.DnsKey
	pageToken := ""
	for {
		call := svc.DnsKeys.List(project, flagDNSKeyZone).Context(ctx)
		if flagDNSKeyMaxResults > 0 {
			call = call.MaxResults(flagDNSKeyMaxResults)
		}
		if flagDNSKeyDigestType != "" {
			call = call.DigestType(flagDNSKeyDigestType)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing DNSKEYs: %w", err)
		}
		all = append(all, resp.DnsKeys...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDNSKeyFormat)
}
