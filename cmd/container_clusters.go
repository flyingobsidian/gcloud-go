package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
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
	flagCtnClRegion     string
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
		Use: "list", Short: "List GKE clusters in a location, or in all locations",
		Args: cobra.NoArgs, RunE: runCtnClList,
	}
	containerClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a GKE cluster (loads UpdateClusterRequest.update from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCtnClUpdate,
	}
	containerClCompleteCpUpgradeCmd = &cobra.Command{
		Use: "complete-control-plane-upgrade CLUSTER", Short: "Complete a rollbackable (two-step) control plane upgrade",
		Args: cobra.ExactArgs(1), RunE: runCtnClCompleteCpUpgrade,
	}
)

func init() {
	all := []*cobra.Command{
		containerClCreateCmd, containerClDeleteCmd, containerClDescribeCmd,
		containerClListCmd, containerClUpdateCmd, containerClCompleteCpUpgradeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagCtnClFormat, "format", "", "Output format")
		if c == containerClListCmd {
			continue
		}
		c.Flags().StringVar(&flagCtnClLocation, "location", "", "Cluster location (region or zone) (required)")
		_ = c.MarkFlagRequired("location")
	}
	// Like gcloud, list queries every location unless one is given.
	containerClListCmd.Flags().StringVar(&flagCtnClLocation, "location", "", "Cluster location (region or zone); default all locations")
	containerClListCmd.Flags().StringVar(&flagCtnClRegion, "region", "", "Cluster region; default all locations")
	containerClListCmd.MarkFlagsMutuallyExclusive("location", "region")
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

// ctnClListLocation returns the location `clusters list` queries: --location,
// --region or an explicit --zone, else "-" for all locations. As in gcloud,
// the compute/zone and compute/region properties are ignored.
func ctnClListLocation() string {
	for _, loc := range []string{flagCtnClLocation, flagCtnClRegion, flagZone} {
		if loc != "" {
			return loc
		}
	}
	return "-"
}

func runCtnClList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", project, ctnClListLocation())
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if len(resp.MissingZones) > 0 {
		fmt.Fprintf(os.Stderr, "WARNING: The following zones did not respond: %s. List results may be incomplete.\n",
			strings.Join(resp.MissingZones, ", "))
	}
	clusters := ctnSortClusters(resp.Clusters)
	if flagCtnClFormat != "" {
		return emitFormatted(clusters, flagCtnClFormat)
	}
	if len(clusters) == 0 {
		fmt.Fprintln(os.Stderr, "Listed 0 items.")
		return nil
	}
	return emitFormatted(ctnClTableRows(clusters), ctnClListFormat)
}

// ctnSortClusters sorts clusters by (zone, name) like gcloud, returning a
// non-nil slice so an empty list renders as [] in JSON.
func ctnSortClusters(clusters []*container.Cluster) []*container.Cluster {
	sorted := append([]*container.Cluster{}, clusters...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Zone != sorted[j].Zone {
			return sorted[i].Zone < sorted[j].Zone
		}
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// ctnClListFormat is gcloud's default `clusters list` table, over the rows
// built by ctnClTableRows.
const ctnClListFormat = "table(name,location,masterVersion,masterIp,machineType,nodeVersion,numNodes,status,stackType)"

// ctnClTableRows projects clusters onto the columns of gcloud's default table.
func ctnClTableRows(clusters []*container.Cluster) []map[string]any {
	rows := make([]map[string]any, 0, len(clusters))
	for _, c := range clusters {
		machineType, nodeVersion := "", ""
		if len(c.NodePools) > 0 {
			// gcloud hides currentNodeVersion for clusters without node
			// pools, where it is the cluster API version.
			nodeVersion = c.CurrentNodeVersion
			if c.NodePools[0].Config != nil {
				machineType = c.NodePools[0].Config.MachineType
			}
		}
		numNodes := c.CurrentNodeCount
		if numNodes == 0 {
			numNodes = c.InitialNodeCount
		}
		stackType := "IPV4"
		if c.IpAllocationPolicy != nil && c.IpAllocationPolicy.StackType != "" {
			stackType = c.IpAllocationPolicy.StackType
		}
		rows = append(rows, map[string]any{
			"name":          c.Name,
			"location":      c.Zone,
			"masterVersion": c.CurrentMasterVersion,
			"masterIp":      c.Endpoint,
			"machineType":   machineType,
			"nodeVersion":   nodeVersion,
			"numNodes":      numNodes,
			"status":        c.Status,
			"stackType":     stackType,
		})
	}
	return rows
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

func runCtnClCompleteCpUpgrade(cmd *cobra.Command, args []string) error {
	name, err := ctnClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.CompleteControlPlaneUpgrade(name, &container.CompleteControlPlaneUpgradeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("completing control plane upgrade: %w", err)
	}
	fmt.Printf("Complete control plane upgrade for cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCtnClFormat)
}
