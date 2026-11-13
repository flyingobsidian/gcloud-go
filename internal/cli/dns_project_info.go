package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud dns project-info (#1540) ---

func (s *dnsCmds) initDnsProjectInfo() {
	s.dnsPIDescribeCmd.Flags().StringVar(&s.flagDNSPIFormat, "format", "", "Output format")
	s.dnsPICmd.AddCommand(s.dnsPIDescribeCmd)
	s.dnsCmd.AddCommand(s.dnsPICmd)
}

func (s *dnsCmds) runDNSPIDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DNS(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS project info: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDNSPIFormat)
}
