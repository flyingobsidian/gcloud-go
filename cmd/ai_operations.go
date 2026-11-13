package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud ai operations (#1459) ---

func (s *aiCmds) initAiOperations() {
	s.aiOpsDescribeCmd.Flags().StringVar(&s.flagAIOpsRegion, "region", "", "Region where the operation lives (required)")
	_ = s.aiOpsDescribeCmd.MarkFlagRequired("region")
	s.aiOpsDescribeCmd.Flags().StringVar(&s.flagAIOpsFormat, "format", "", "Output format")
	s.aiOpsCmd.AddCommand(s.aiOpsDescribeCmd)
	s.aiCmd.AddCommand(s.aiOpsCmd)
}

func (s *aiCmds) aiOpsName(id string) (string, error) {
	parent, err := s.aiParent(s.flagAIOpsRegion)
	if err != nil {
		return "", err
	}
	return aiChild("operations", id, parent), nil
}

func (s *aiCmds) runAIOpsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiOpsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIOpsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIOpsFormat)
}
