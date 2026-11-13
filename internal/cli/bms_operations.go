package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud bms operations (#1229) ---

func (s *bmsCmds) initBmsOperations() {
	all := []*cobra.Command{s.bmsOpDescribeCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBmsOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagBmsOpFormat, "format", "", "Output format")
	}

	s.bmsOperationsCmd.AddCommand(all...)
	s.bmsCmd.AddCommand(s.bmsOperationsCmd)
}

func (s *bmsCmds) bmsOpName(id string) (string, error) {
	return s.bmsResource(s.flagBmsOpLocation, "operations", id)
}

func (s *bmsCmds) runBmsOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.bmsOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsOpFormat)
}
