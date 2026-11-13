package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
)

// --- gcloud container node-pools (#1140) ---

func (s *containerCmds) initContainerNodePools() {
	all := []*cobra.Command{
		s.containerNpCreateCmd, s.containerNpDeleteCmd, s.containerNpDescribeCmd,
		s.containerNpListCmd, s.containerNpUpdateCmd, s.containerNpRollbackCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagCtnNpLocation, "location", "", "Cluster location (region or zone) (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagCtnNpCluster, "cluster", "", "Cluster that owns the node pool (required)")
		_ = c.MarkFlagRequired("cluster")
		c.Flags().StringVar(&s.flagCtnNpFormat, "format", "", "Output format")
	}
	s.containerNpCreateCmd.Flags().StringVar(&s.flagCtnNpConfigFile, "config-file", "", "YAML/JSON file with a container.NodePool body (required)")
	_ = s.containerNpCreateCmd.MarkFlagRequired("config-file")
	s.containerNpUpdateCmd.Flags().StringVar(&s.flagCtnNpConfigFile, "config-file", "", "YAML/JSON file with an UpdateNodePoolRequest body (required)")
	_ = s.containerNpUpdateCmd.MarkFlagRequired("config-file")

	s.containerNodePoolsCmd.AddCommand(all...)
	s.containerCmd.AddCommand(s.containerNodePoolsCmd)
}

func (s *containerCmds) ctnNpClusterName() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagCtnNpLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	if s.flagCtnNpCluster == "" {
		return "", fmt.Errorf("--cluster is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, s.flagCtnNpLocation, s.flagCtnNpCluster), nil
}

func (s *containerCmds) ctnNpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	cluster, err := s.ctnNpClusterName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/nodePools/%s", cluster, id), nil
}

func (s *containerCmds) runCtnNpCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.ctnNpClusterName()
	if err != nil {
		return err
	}
	np := &container.NodePool{}
	if err := loadYAMLOrJSONInto(s.flagCtnNpConfigFile, np); err != nil {
		return err
	}
	np.Name = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.NodePools.Create(parent, &container.CreateNodePoolRequest{NodePool: np}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating node pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create node pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCtnNpFormat)
}

func (s *containerCmds) runCtnNpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.ctnNpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.NodePools.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting node pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete node pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCtnNpFormat)
}

func (s *containerCmds) runCtnNpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.ctnNpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.NodePools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing node pool: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCtnNpFormat)
}

func (s *containerCmds) runCtnNpList(cmd *cobra.Command, args []string) error {
	parent, err := s.ctnNpClusterName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.NodePools.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing node pools: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp.NodePools, s.flagCtnNpFormat)
}

func (s *containerCmds) runCtnNpUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.ctnNpName(args[0])
	if err != nil {
		return err
	}
	body := &container.UpdateNodePoolRequest{}
	if err := loadYAMLOrJSONInto(s.flagCtnNpConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.NodePools.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating node pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update node pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCtnNpFormat)
}

func (s *containerCmds) runCtnNpRollback(cmd *cobra.Command, args []string) error {
	name, err := s.ctnNpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.NodePools.Rollback(name, &container.RollbackNodePoolUpgradeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back node pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Rollback node pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCtnNpFormat)
}
