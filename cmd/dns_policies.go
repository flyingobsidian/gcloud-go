package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns policies (#1539) ---

func (s *dnsCmds) initDnsPolicies() {
	all := []*cobra.Command{s.dnsPolCreateCmd, s.dnsPolDeleteCmd, s.dnsPolDescribeCmd, s.dnsPolListCmd, s.dnsPolUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDNSPolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dnsPolCreateCmd, s.dnsPolUpdateCmd} {
		c.Flags().StringVar(&s.flagDNSPolConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Policy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dnsPolListCmd.Flags().Int64Var(&s.flagDNSPolMaxResults, "limit", 0, "Maximum results per page")

	s.dnsPolCmd.AddCommand(all...)
	s.dnsCmd.AddCommand(s.dnsPolCmd)
}

func (s *dnsCmds) runDNSPolCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dns.Policy{}
	if err := loadYAMLOrJSONInto(s.flagDNSPolConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Policies.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DNS policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created DNS policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSPolFormat)
}

func (s *dnsCmds) runDNSPolDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	if err := svc.Policies.Delete(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting DNS policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted DNS policy [%s].\n", args[0])
	return nil
}

func (s *dnsCmds) runDNSPolDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Policies.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSPolFormat)
}

func (s *dnsCmds) runDNSPolList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	var all []*dns.Policy
	pageToken := ""
	for {
		call := svc.Policies.List(project).Context(ctx)
		if s.flagDNSPolMaxResults > 0 {
			call = call.MaxResults(s.flagDNSPolMaxResults)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDNSPolFormat)
}

func (s *dnsCmds) runDNSPolUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dns.Policy{}
	if err := loadYAMLOrJSONInto(s.flagDNSPolConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Policies.Patch(project, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating DNS policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated DNS policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSPolFormat)
}
