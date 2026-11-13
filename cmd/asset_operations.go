package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (s *assetCmds) initAssetOperations() {
	s.assetOperationsCmd.AddCommand(s.assetOperationDescribeCmd)
	s.assetCmd.AddCommand(s.assetOperationsCmd)
}

func (s *assetCmds) runAssetOperationDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return s.inv.yaml(op)
}
