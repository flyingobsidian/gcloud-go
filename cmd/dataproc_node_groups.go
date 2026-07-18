package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc node-groups (#1514) ---

var dpNGCmd = &cobra.Command{Use: "node-groups", Short: "Manage Dataproc cluster node groups"}

var (
	flagDPNGRegion    string
	flagDPNGCluster   string
	flagDPNGFormat    string
	flagDPNGSize      int64
	flagDPNGGraceful  string
	flagDPNGRequestID string
)

var (
	dpNGDescribeCmd = &cobra.Command{
		Use: "describe NODE_GROUP", Short: "Describe a Dataproc node group",
		Args: cobra.ExactArgs(1), RunE: runDPNGDescribe,
	}
	dpNGResizeCmd = &cobra.Command{
		Use: "resize NODE_GROUP", Short: "Resize a Dataproc node group",
		Args: cobra.ExactArgs(1), RunE: runDPNGResize,
	}
)

func init() {
	all := []*cobra.Command{dpNGDescribeCmd, dpNGResizeCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDPNGRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDPNGCluster, "cluster", "", "Cluster that owns the node group (required)")
		_ = c.MarkFlagRequired("cluster")
		c.Flags().StringVar(&flagDPNGFormat, "format", "", "Output format")
	}
	dpNGResizeCmd.Flags().Int64Var(&flagDPNGSize, "size", 0, "Target size for the node group (required)")
	_ = dpNGResizeCmd.MarkFlagRequired("size")
	dpNGResizeCmd.Flags().StringVar(&flagDPNGGraceful, "graceful-decommission-timeout", "",
		"Timeout for graceful decommission (e.g. 10m)")
	dpNGResizeCmd.Flags().StringVar(&flagDPNGRequestID, "request-id", "",
		"Optional idempotency ID")

	dpNGCmd.AddCommand(all...)
	dataprocCmd.AddCommand(dpNGCmd)
}

func dpNGResourceName(project, region, cluster, id string) string {
	return fmt.Sprintf("projects/%s/regions/%s/clusters/%s/nodeGroups/%s", project, region, cluster, id)
}

func runDPNGDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := dpNGResourceName(project, flagDPNGRegion, flagDPNGCluster, args[0])
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPNGRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Clusters.NodeGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing node group: %w", err)
	}
	return emitFormatted(got, flagDPNGFormat)
}

func runDPNGResize(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := dpNGResourceName(project, flagDPNGRegion, flagDPNGCluster, args[0])
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPNGRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Clusters.NodeGroups.Resize(name, &dataproc.ResizeNodeGroupRequest{
		Size:                        flagDPNGSize,
		GracefulDecommissionTimeout: flagDPNGGraceful,
		RequestId:                   flagDPNGRequestID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resizing node group: %w", err)
	}
	fmt.Printf("Resize request issued for node group [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPNGFormat)
}
