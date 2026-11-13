package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
)

// --- gcloud container clusters (#1136) ---
//
// Wires the core cluster CRUD verbs against the container v1 client.
// The many specialty verbs on ClustersService (get-credentials, resize,
// upgrade, set-*-config, ...) are intentionally left for follow-up
// changes.

func (s *containerCmds) initContainerClusters() {
	all := []*cobra.Command{
		s.containerClCreateCmd, s.containerClDeleteCmd, s.containerClDescribeCmd,
		s.containerClListCmd, s.containerClUpdateCmd, s.containerClCompleteCpUpgradeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagCtnClFormat, "format", "", "Output format")
		if c == s.containerClListCmd {
			continue
		}
		c.Flags().StringVar(&s.flagCtnClLocation, "location", "", "Cluster location (region or zone) (required)")
		_ = c.MarkFlagRequired("location")
	}
	// Like gcloud, list queries every location unless one is given.
	s.containerClListCmd.Flags().StringVar(&s.flagCtnClLocation, "location", "", "Cluster location (region or zone); default all locations")
	s.containerClListCmd.Flags().StringVar(&s.flagCtnClRegion, "region", "", "Cluster region; default all locations")
	s.containerClListCmd.MarkFlagsMutuallyExclusive("location", "region")
	s.containerClCreateCmd.Flags().StringVar(&s.flagCtnClConfigFile, "config-file", "", "YAML/JSON file with a container.Cluster body (required)")
	_ = s.containerClCreateCmd.MarkFlagRequired("config-file")
	s.containerClUpdateCmd.Flags().StringVar(&s.flagCtnClConfigFile, "config-file", "", "YAML/JSON file with a container.ClusterUpdate body (required)")
	_ = s.containerClUpdateCmd.MarkFlagRequired("config-file")

	s.containerClustersCmd.AddCommand(all...)
	s.containerCmd.AddCommand(s.containerClustersCmd)
}

func (s *containerCmds) ctnParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagCtnClLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagCtnClLocation), nil
}

func (s *containerCmds) ctnClusterName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.ctnParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/clusters/%s", parent, id), nil
}

func (s *containerCmds) runCtnClCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.ctnParent()
	if err != nil {
		return err
	}
	cluster := &container.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagCtnClConfigFile, cluster); err != nil {
		return err
	}
	cluster.Name = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Create(parent, &container.CreateClusterRequest{Cluster: cluster}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCtnClFormat)
}

func (s *containerCmds) runCtnClDelete(cmd *cobra.Command, args []string) error {
	name, err := s.ctnClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCtnClFormat)
}

func (s *containerCmds) runCtnClDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.ctnClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCtnClFormat)
}

// ctnClListLocation returns the location `clusters list` queries: --location,
// --region or an explicit --zone, else "-" for all locations. As in gcloud,
// the compute/zone and compute/region properties are ignored.
func (s *containerCmds) ctnClListLocation() string {
	for _, loc := range []string{s.flagCtnClLocation, s.flagCtnClRegion, s.inv.flags.zone} {
		if loc != "" {
			return loc
		}
	}
	return "-"
}

func (s *containerCmds) runCtnClList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", project, s.ctnClListLocation())
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if len(resp.MissingZones) > 0 {
		fmt.Fprintf(s.inv.stderr, "WARNING: The following zones did not respond: %s. List results may be incomplete.\n",
			strings.Join(resp.MissingZones, ", "))
	}
	clusters := ctnSortClusters(resp.Clusters)
	if s.flagCtnClFormat != "" {
		return emitFormattedTo(s.inv.stdout, clusters, s.flagCtnClFormat)
	}
	if len(clusters) == 0 {
		fmt.Fprintln(s.inv.stderr, "Listed 0 items.")
		return nil
	}
	return emitFormattedTo(s.inv.stdout, ctnClTableRows(clusters), ctnClListFormat)
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

func (s *containerCmds) runCtnClUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.ctnClusterName(args[0])
	if err != nil {
		return err
	}
	update := &container.ClusterUpdate{}
	if err := loadYAMLOrJSONInto(s.flagCtnClConfigFile, update); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Update(name, &container.UpdateClusterRequest{Update: update}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCtnClFormat)
}

func (s *containerCmds) runCtnClCompleteCpUpgrade(cmd *cobra.Command, args []string) error {
	name, err := s.ctnClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.CompleteControlPlaneUpgrade(name, &container.CompleteControlPlaneUpgradeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("completing control plane upgrade: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Complete control plane upgrade for cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCtnClFormat)
}
