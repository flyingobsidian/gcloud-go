package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
)

// --- gcloud container clusters (#1136) ---
//
// Wires the core cluster CRUD verbs against the container v1 client.
// The many specialty verbs on ClustersService (get-credentials, resize,
// upgrade, set-*-config, ...) are intentionally left for follow-up
// changes.

var containerClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage GKE clusters"}

var (
	flagCtnClLocation   string
	flagCtnClFormat     string
	flagCtnClConfigFile string
)

var (
	containerClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a GKE cluster (loads CreateClusterRequest.cluster from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCtnClCreate,
	}
	containerClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a GKE cluster",
		Args: cobra.ExactArgs(1), RunE: runCtnClDelete,
	}
	containerClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a GKE cluster",
		Args: cobra.ExactArgs(1), RunE: runCtnClDescribe,
	}
	containerClListCmd = &cobra.Command{
		Use: "list", Short: "List GKE clusters in a location",
		Args: cobra.NoArgs, RunE: runCtnClList,
	}
	containerClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a GKE cluster (loads UpdateClusterRequest.update from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCtnClUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		containerClCreateCmd, containerClDeleteCmd, containerClDescribeCmd,
		containerClListCmd, containerClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagCtnClLocation, "location", "", "Cluster location (region or zone) (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagCtnClFormat, "format", "", "Output format")
	}
	containerClCreateCmd.Flags().StringVar(&flagCtnClConfigFile, "config-file", "", "YAML/JSON file with a container.Cluster body (required)")
	_ = containerClCreateCmd.MarkFlagRequired("config-file")
	containerClUpdateCmd.Flags().StringVar(&flagCtnClConfigFile, "config-file", "", "YAML/JSON file with a container.ClusterUpdate body (required)")
	_ = containerClUpdateCmd.MarkFlagRequired("config-file")

	containerClustersCmd.AddCommand(all...)
	containerCmd.AddCommand(containerClustersCmd)
}

func ctnParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagCtnClLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagCtnClLocation), nil
}

func ctnClusterName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := ctnParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/clusters/%s", parent, id), nil
}

func runCtnClCreate(cmd *cobra.Command, args []string) error {
	parent, err := ctnParent()
	if err != nil {
		return err
	}
	cluster := &container.Cluster{}
	if err := loadYAMLOrJSONInto(flagCtnClConfigFile, cluster); err != nil {
		return err
	}
	cluster.Name = args[0]
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Create(parent, &container.CreateClusterRequest{Cluster: cluster}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	fmt.Printf("Create cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCtnClFormat)
}

func runCtnClDelete(cmd *cobra.Command, args []string) error {
	name, err := ctnClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	fmt.Printf("Delete cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCtnClFormat)
}

func runCtnClDescribe(cmd *cobra.Command, args []string) error {
	name, err := ctnClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormatted(got, flagCtnClFormat)
}

func runCtnClList(cmd *cobra.Command, args []string) error {
	parent, err := ctnParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	return emitFormatted(resp.Clusters, flagCtnClFormat)
}

func runCtnClUpdate(cmd *cobra.Command, args []string) error {
	name, err := ctnClusterName(args[0])
	if err != nil {
		return err
	}
	update := &container.ClusterUpdate{}
	if err := loadYAMLOrJSONInto(flagCtnClConfigFile, update); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Update(name, &container.UpdateClusterRequest{Update: update}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	fmt.Printf("Update cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCtnClFormat)
}
