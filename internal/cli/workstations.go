package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	workstations "google.golang.org/api/workstations/v1"
)

// --- gcloud workstations (#400) ---

func (s *workstationsCmds) initWorkstations() {
	// clusters (#1259) and configs (#1260) live in dedicated files. The
	// leaf commands on the individual workstations resource (create/delete/
	// describe/list/list-usable/ssh/start/start-tcp-tunnel/stop/update and
	// get-/set-iam-policy) are still stubs.
	for _, name := range []string{
		"create", "delete", "describe", "get-iam-policy", "list", "list-usable",
		"set-iam-policy", "ssh", "start", "start-tcp-tunnel", "stop", "update",
	} {
		registerStubCommand(s.workstationsCmd, name, "Not yet implemented")
	}

}

// wsRegionalEndpointRegions returns the regions with a known workstations
// regional endpoint in gcloud-python 589.0.0's API map.
func wsRegionalEndpointRegions() []string {
	return []string{
		"asia-south1", "europe-west6", "us-central1", "us-east1",
		"us-east4", "us-east5", "us-west1", "us-west4",
	}
}

// addWsEndpointModeFlag adds --endpoint-mode to c, as gcloud-python does for
// commands annotated regional_endpoint_compatibility: SUPPORTED (587.0.0).
func (s *workstationsCmds) addWsEndpointModeFlag(c *cobra.Command) {
	addEndpointModeFlag(c, &s.flagWsEndpointMode)
}

// wsEndpoint returns the workstations endpoint for location, or "" for the
// global endpoint (see regionalEndpoint).
func (s *workstationsCmds) wsEndpoint(location string) (string, error) {
	mode, err := s.inv.endpointMode(s.flagWsEndpointMode)
	if err != nil {
		return "", err
	}
	return regionalEndpoint(mode, "workstations", location, wsRegionalEndpointRegions())
}

// wsService returns a workstations client for the endpoint selected for the
// location of resource, a projects/P/locations/L/... name.
func (s *workstationsCmds) wsService(ctx context.Context, resource string) (*workstations.Service, error) {
	location := ""
	if parts := strings.Split(resource, "/"); len(parts) >= 4 && parts[2] == "locations" {
		location = parts[3]
	}
	endpoint, err := s.wsEndpoint(location)
	if err != nil {
		return nil, err
	}
	return s.inv.clients().Workstations(ctx, endpoint)
}

// wsLocationParent returns "projects/PROJECT/locations/LOCATION".
func (s *workstationsCmds) wsLocationParent(location string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// wsClusterName returns projects/.../locations/.../workstationClusters/ID.
func (s *workstationsCmds) wsClusterName(location, cluster string) (string, error) {
	if strings.HasPrefix(cluster, "projects/") {
		return cluster, nil
	}
	parent, err := s.wsLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/workstationClusters/%s", parent, cluster), nil
}

// wsConfigName returns projects/.../workstationClusters/CLUSTER/workstationConfigs/ID.
func (s *workstationsCmds) wsConfigName(location, cluster, config string) (string, error) {
	if strings.HasPrefix(config, "projects/") {
		return config, nil
	}
	c, err := s.wsClusterName(location, cluster)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/workstationConfigs/%s", c, config), nil
}

// workstationsCmds holds the commands and flag values of one invocation's "workstations"
// command group, so invocations don't share state (#2070).
type workstationsCmds struct {
	inv *invocation

	workstationsCmd            *cobra.Command
	flagWsEndpointMode         string
	workstationsClustersCmd    *cobra.Command
	flagWsClLocation           string
	flagWsClFormat             string
	flagWsClConfigFile         string
	flagWsClUpdateMask         string
	flagWsClPageSize           int64
	workstationsClCreateCmd    *cobra.Command
	workstationsClDeleteCmd    *cobra.Command
	workstationsClDescribeCmd  *cobra.Command
	workstationsClListCmd      *cobra.Command
	workstationsClUpdateCmd    *cobra.Command
	workstationsConfigsCmd     *cobra.Command
	flagWsCfgLocation          string
	flagWsCfgCluster           string
	flagWsCfgFormat            string
	flagWsCfgConfigFile        string
	flagWsCfgUpdateMask        string
	flagWsCfgPageSize          int64
	workstationsCfgCreateCmd   *cobra.Command
	workstationsCfgDeleteCmd   *cobra.Command
	workstationsCfgDescribeCmd *cobra.Command
	workstationsCfgListCmd     *cobra.Command
	workstationsCfgUpdateCmd   *cobra.Command
	workstationsCfgGetIamCmd   *cobra.Command
	workstationsCfgSetIamCmd   *cobra.Command
}

// newWorkstationsCmd returns the "workstations" command group for inv.
func newWorkstationsCmd(inv *invocation) *cobra.Command {
	return newWorkstationsCmds(inv).workstationsCmd
}

// newWorkstationsCmds builds the "workstations" command group for inv and returns its state.
func newWorkstationsCmds(inv *invocation) *workstationsCmds {
	s := &workstationsCmds{inv: inv}
	s.workstationsCmd = &cobra.Command{Use: "workstations", Short: "Manage Cloud Workstations"}
	s.workstationsClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage workstation clusters"}
	s.workstationsClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a workstation cluster",
		Args: cobra.ExactArgs(1), RunE: s.runWsClCreate,
	}
	s.workstationsClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a workstation cluster",
		Args: cobra.ExactArgs(1), RunE: s.runWsClDelete,
	}
	s.workstationsClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a workstation cluster",
		Args: cobra.ExactArgs(1), RunE: s.runWsClDescribe,
	}
	s.workstationsClListCmd = &cobra.Command{
		Use: "list", Short: "List workstation clusters in a location",
		Args: cobra.NoArgs, RunE: s.runWsClList,
	}
	s.workstationsClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a workstation cluster",
		Args: cobra.ExactArgs(1), RunE: s.runWsClUpdate,
	}
	s.workstationsConfigsCmd = &cobra.Command{Use: "configs", Short: "Manage workstation configurations"}
	s.workstationsCfgCreateCmd = &cobra.Command{
		Use: "create CONFIG", Short: "Create a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: s.runWsCfgCreate,
	}
	s.workstationsCfgDeleteCmd = &cobra.Command{
		Use: "delete CONFIG", Short: "Delete a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: s.runWsCfgDelete,
	}
	s.workstationsCfgDescribeCmd = &cobra.Command{
		Use: "describe CONFIG", Short: "Describe a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: s.runWsCfgDescribe,
	}
	s.workstationsCfgListCmd = &cobra.Command{
		Use: "list", Short: "List workstation configurations in a cluster",
		Args: cobra.NoArgs, RunE: s.runWsCfgList,
	}
	s.workstationsCfgUpdateCmd = &cobra.Command{
		Use: "update CONFIG", Short: "Update a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: s.runWsCfgUpdate,
	}
	s.workstationsCfgGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CONFIG", Short: "Get the IAM policy for a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: s.runWsCfgGetIam,
	}
	s.workstationsCfgSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CONFIG POLICY_FILE", Short: "Set the IAM policy for a workstation configuration",
		Args: cobra.ExactArgs(2), RunE: s.runWsCfgSetIam,
	}
	s.initWorkstations()
	s.initWorkstationsClusters()
	s.initWorkstationsConfigs()
	return s
}
