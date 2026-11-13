package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func (g *servicesCmds) initServicesOperations() {
	g.servicesOperationWaitCmd.Flags().DurationVar(&g.flagServicesOpWaitInterval, "poll-interval", 2*time.Second, "Interval between poll attempts")
	g.servicesOperationsCmd.AddCommand(g.servicesOperationDescribeCmd, g.servicesOperationWaitCmd)
	g.servicesCmd.AddCommand(g.servicesOperationsCmd)
}

func (g *servicesCmds) runServicesOperationDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceUsage(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return g.inv.yaml(op)
}

func (g *servicesCmds) runServicesOperationWait(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceUsage(ctx)
	if err != nil {
		return err
	}
	for {
		op, err := svc.Operations.Get(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			return g.inv.yaml(op)
		}
		time.Sleep(g.flagServicesOpWaitInterval)
	}
}
