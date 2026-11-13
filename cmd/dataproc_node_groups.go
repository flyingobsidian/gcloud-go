package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc node-groups (#1514) ---

func (s *dataprocCmds) initDataprocNodeGroups() {
	all := []*cobra.Command{s.dpNGDescribeCmd, s.dpNGResizeCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDPNGRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDPNGCluster, "cluster", "", "Cluster that owns the node group (required)")
		_ = c.MarkFlagRequired("cluster")
		c.Flags().StringVar(&s.flagDPNGFormat, "format", "", "Output format")
	}
	s.dpNGResizeCmd.Flags().Int64Var(&s.flagDPNGSize, "size", 0, "Target size for the node group (required)")
	_ = s.dpNGResizeCmd.MarkFlagRequired("size")
	s.dpNGResizeCmd.Flags().StringVar(&s.flagDPNGGraceful, "graceful-decommission-timeout", "",
		"Timeout for graceful decommission (e.g. 10m)")
	s.dpNGResizeCmd.Flags().StringVar(&s.flagDPNGRequestID, "request-id", "",
		"Optional idempotency ID")

	s.dpNGCmd.AddCommand(all...)
	s.dataprocCmd.AddCommand(s.dpNGCmd)
}

func dpNGResourceName(project, region, cluster, id string) string {
	return fmt.Sprintf("projects/%s/regions/%s/clusters/%s/nodeGroups/%s", project, region, cluster, id)
}

func (s *dataprocCmds) runDPNGDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	name := dpNGResourceName(project, s.flagDPNGRegion, s.flagDPNGCluster, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPNGRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Clusters.NodeGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing node group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPNGFormat)
}

func (s *dataprocCmds) runDPNGResize(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	name := dpNGResourceName(project, s.flagDPNGRegion, s.flagDPNGCluster, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPNGRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Clusters.NodeGroups.Resize(name, &dataproc.ResizeNodeGroupRequest{
		Size:                        s.flagDPNGSize,
		GracefulDecommissionTimeout: s.flagDPNGGraceful,
		RequestId:                   s.flagDPNGRequestID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resizing node group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Resize request issued for node group [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPNGFormat)
}
