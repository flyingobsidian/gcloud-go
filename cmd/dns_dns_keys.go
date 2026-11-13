package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns dns-keys (#1536) ---

func (s *dnsCmds) initDnsDnsKeys() {
	for _, c := range []*cobra.Command{s.dnsKeysDescribeCmd, s.dnsKeysListCmd} {
		c.Flags().StringVar(&s.flagDNSKeyZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&s.flagDNSKeyFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagDNSKeyDigestType, "digest-type", "",
			"Digest type used when generating the DS record")
	}
	s.dnsKeysListCmd.Flags().Int64Var(&s.flagDNSKeyMaxResults, "limit", 0, "Maximum results per page")

	s.dnsKeysCmd.AddCommand(s.dnsKeysDescribeCmd, s.dnsKeysListCmd)
	s.dnsCmd.AddCommand(s.dnsKeysCmd)
}

func (s *dnsCmds) runDNSKeyDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	call := svc.DnsKeys.Get(project, s.flagDNSKeyZone, args[0]).Context(ctx)
	if s.flagDNSKeyDigestType != "" {
		call = call.DigestType(s.flagDNSKeyDigestType)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing DNSKEY: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSKeyFormat)
}

func (s *dnsCmds) runDNSKeyList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.DnsKey
	pageToken := ""
	for {
		call := svc.DnsKeys.List(project, s.flagDNSKeyZone).Context(ctx)
		if s.flagDNSKeyMaxResults > 0 {
			call = call.MaxResults(s.flagDNSKeyMaxResults)
		}
		if s.flagDNSKeyDigestType != "" {
			call = call.DigestType(s.flagDNSKeyDigestType)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDNSKeyFormat)
}
