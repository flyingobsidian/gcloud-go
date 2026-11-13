package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns operations (#1538) ---

func (s *dnsCmds) initDnsOperations() {
	for _, c := range []*cobra.Command{s.dnsOpDescribeCmd, s.dnsOpListCmd} {
		c.Flags().StringVar(&s.flagDNSOpZone, "zone", "", "Managed zone name (required)")
		_ = c.MarkFlagRequired("zone")
		c.Flags().StringVar(&s.flagDNSOpFormat, "format", "", "Output format")
	}
	s.dnsOpListCmd.Flags().Int64Var(&s.flagDNSOpMaxResults, "limit", 0, "Maximum results per page")

	s.dnsOpCmd.AddCommand(s.dnsOpDescribeCmd, s.dnsOpListCmd)
	s.dnsCmd.AddCommand(s.dnsOpCmd)
}

func (s *dnsCmds) runDNSOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ManagedZoneOperations.Get(project, s.flagDNSOpZone, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSOpFormat)
}

func (s *dnsCmds) runDNSOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.Operation
	pageToken := ""
	for {
		call := svc.ManagedZoneOperations.List(project, s.flagDNSOpZone).Context(ctx)
		if s.flagDNSOpMaxResults > 0 {
			call = call.MaxResults(s.flagDNSOpMaxResults)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDNSOpFormat)
}
