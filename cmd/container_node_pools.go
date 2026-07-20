package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
)

// --- gcloud container node-pools (#1140) ---

var containerNodePoolsCmd = &cobra.Command{Use: "node-pools", Short: "Manage GKE node pools"}

var (
	flagCtnNpLocation   string
	flagCtnNpCluster    string
	flagCtnNpFormat     string
	flagCtnNpConfigFile string
)

var (
	containerNpCreateCmd = &cobra.Command{
		Use: "create NODE_POOL", Short: "Create a node pool (loads CreateNodePoolRequest.nodePool from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCtnNpCreate,
	}
	containerNpDeleteCmd = &cobra.Command{
		Use: "delete NODE_POOL", Short: "Delete a node pool",
		Args: cobra.ExactArgs(1), RunE: runCtnNpDelete,
	}
	containerNpDescribeCmd = &cobra.Command{
		Use: "describe NODE_POOL", Short: "Describe a node pool",
		Args: cobra.ExactArgs(1), RunE: runCtnNpDescribe,
	}
	containerNpListCmd = &cobra.Command{
		Use: "list", Short: "List node pools on a cluster",
		Args: cobra.NoArgs, RunE: runCtnNpList,
	}
	containerNpUpdateCmd = &cobra.Command{
		Use: "update NODE_POOL", Short: "Update a node pool (loads UpdateNodePoolRequest from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCtnNpUpdate,
	}
	containerNpRollbackCmd = &cobra.Command{
		Use: "rollback NODE_POOL", Short: "Roll back a previously-Aborted or Failed NodePool upgrade",
		Args: cobra.ExactArgs(1), RunE: runCtnNpRollback,
	}
)

func init() {
	all := []*cobra.Command{
		containerNpCreateCmd, containerNpDeleteCmd, containerNpDescribeCmd,
		containerNpListCmd, containerNpUpdateCmd, containerNpRollbackCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagCtnNpLocation, "location", "", "Cluster location (region or zone) (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagCtnNpCluster, "cluster", "", "Cluster that owns the node pool (required)")
		_ = c.MarkFlagRequired("cluster")
		c.Flags().StringVar(&flagCtnNpFormat, "format", "", "Output format")
	}
	containerNpCreateCmd.Flags().StringVar(&flagCtnNpConfigFile, "config-file", "", "YAML/JSON file with a container.NodePool body (required)")
	_ = containerNpCreateCmd.MarkFlagRequired("config-file")
	containerNpUpdateCmd.Flags().StringVar(&flagCtnNpConfigFile, "config-file", "", "YAML/JSON file with an UpdateNodePoolRequest body (required)")
	_ = containerNpUpdateCmd.MarkFlagRequired("config-file")

	containerNodePoolsCmd.AddCommand(all...)
	containerCmd.AddCommand(containerNodePoolsCmd)
}

func ctnNpClusterName() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagCtnNpLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	if flagCtnNpCluster == "" {
		return "", fmt.Errorf("--cluster is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, flagCtnNpLocation, flagCtnNpCluster), nil
}

func ctnNpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	cluster, err := ctnNpClusterName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/nodePools/%s", cluster, id), nil
}

func runCtnNpCreate(cmd *cobra.Command, args []string) error {
	parent, err := ctnNpClusterName()
	if err != nil {
		return err
	}
	np := &container.NodePool{}
	if err := loadYAMLOrJSONInto(flagCtnNpConfigFile, np); err != nil {
		return err
	}
	np.Name = args[0]
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.NodePools.Create(parent, &container.CreateNodePoolRequest{NodePool: np}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating node pool: %w", err)
	}
	fmt.Printf("Create node pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCtnNpFormat)
}

func runCtnNpDelete(cmd *cobra.Command, args []string) error {
	name, err := ctnNpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.NodePools.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting node pool: %w", err)
	}
	fmt.Printf("Delete node pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCtnNpFormat)
}

func runCtnNpDescribe(cmd *cobra.Command, args []string) error {
	name, err := ctnNpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.NodePools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing node pool: %w", err)
	}
	return emitFormatted(got, flagCtnNpFormat)
}

func runCtnNpList(cmd *cobra.Command, args []string) error {
	parent, err := ctnNpClusterName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.NodePools.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing node pools: %w", err)
	}
	return emitFormatted(resp.NodePools, flagCtnNpFormat)
}

func runCtnNpUpdate(cmd *cobra.Command, args []string) error {
	name, err := ctnNpName(args[0])
	if err != nil {
		return err
	}
	body := &container.UpdateNodePoolRequest{}
	if err := loadYAMLOrJSONInto(flagCtnNpConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.NodePools.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating node pool: %w", err)
	}
	fmt.Printf("Update node pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCtnNpFormat)
}

func runCtnNpRollback(cmd *cobra.Command, args []string) error {
	name, err := ctnNpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.NodePools.Rollback(name, &container.RollbackNodePoolUpgradeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back node pool: %w", err)
	}
	fmt.Printf("Rollback node pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCtnNpFormat)
}
