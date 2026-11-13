package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud storage service-agent (#1245) ---

func (s *storageCmds) initStorageServiceAgent() {
	s.storageSaDescribeCmd.Flags().StringVar(&s.flagStSaFormat, "format", "", "Output format")

	s.storageServiceAgentCmd.AddCommand(s.storageSaDescribeCmd)
	s.storageCmd.AddCommand(s.storageServiceAgentCmd)
}

func (s *storageCmds) runStSaDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.ServiceAccount.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service agent: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStSaFormat)
}
