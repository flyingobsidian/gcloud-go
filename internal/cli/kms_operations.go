package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud kms operations (#1109) ---

func (s *kmsCmds) initKmsOperations() {
	s.kmsOpDescribeCmd.Flags().StringVar(&s.flagKmsOpLocation, "location", "", "Location (required)")
	s.kmsOpDescribeCmd.Flags().StringVar(&s.flagKmsOpFormat, "format", "", "Output format")
	_ = s.kmsOpDescribeCmd.MarkFlagRequired("location")
	s.kmsOperationsCmd.AddCommand(s.kmsOpDescribeCmd)
	s.kmsCmd.AddCommand(s.kmsOperationsCmd)
}

func (s *kmsCmds) runKmsOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsOpLocation) + "/operations"
	name := kmsFullName(parent, args[0])
	out, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsOpFormat)
}
