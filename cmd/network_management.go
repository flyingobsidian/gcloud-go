package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	networkmanagement "google.golang.org/api/networkmanagement/v1"
)

// --- gcloud network-management (#953, #954, #955, #956) ---
//
// The "network-monitoring-providers" subgroup (#954) is not modeled by
// google.golang.org/api/networkmanagement/v1; it lives on the v1beta1 REST
// surface and is implemented via the shared restClient in
// network_management_network_monitoring_providers.go.

// netmgmtMonRest talks to the v1beta1 REST surface of
// networkmanagement.googleapis.com which exposes networkMonitoringProviders.
var netmgmtMonRest = newRESTClient("https://networkmanagement.googleapis.com/v1beta1")

// --- Connectivity tests ---

// --- Operations ---

// --- VPC Flow Logs configs ---

func (s *networkManagementCmds) initNetworkManagement() {
	// Connectivity tests always live under locations/global; the flag exists
	// only for parity with other network-management resources.
	for _, c := range []*cobra.Command{
		s.netmgmtCTCreateCmd, s.netmgmtCTDeleteCmd, s.netmgmtCTDescribeCmd,
		s.netmgmtCTRerunCmd, s.netmgmtCTUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagNMLocation, "location", "global", "Location containing the test")
	}
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMFile, "config-from-file", "", "YAML/JSON file with the ConnectivityTest spec")
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMTestDescription, "description", "", "Description of the test")
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMTestProtocol, "protocol", "", "IP protocol (e.g. TCP, UDP, ICMP)")
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMTestSourceIP, "source-ip-address", "", "Source IP")
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMTestSourceProj, "source-project", "", "Source project ID")
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMTestSourceNet, "source-network", "", "Source network URI")
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMTestDestIP, "destination-ip-address", "", "Destination IP")
	s.netmgmtCTCreateCmd.Flags().Int64Var(&s.flagNMTestDestPort, "destination-port", 0, "Destination port")
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMTestDestProj, "destination-project", "", "Destination project ID")
	s.netmgmtCTCreateCmd.Flags().StringVar(&s.flagNMTestDestNet, "destination-network", "", "Destination network URI")
	s.netmgmtCTCreateCmd.Flags().StringToStringVar(&s.flagNMTestLabels, "labels", nil, "Labels (key=value)")

	s.netmgmtCTUpdateCmd.Flags().StringVar(&s.flagNMFile, "config-from-file", "", "YAML/JSON file with the ConnectivityTest patch")
	s.netmgmtCTUpdateCmd.Flags().StringVar(&s.flagNMTestDescription, "description", "", "New description")
	s.netmgmtCTUpdateCmd.Flags().StringToStringVar(&s.flagNMTestLabels, "labels", nil, "Labels (replaces the existing set)")

	s.netmgmtCTListCmd.Flags().StringVar(&s.flagNMLocation, "location", "global", "Location to list tests in")
	s.netmgmtCTListCmd.Flags().StringVar(&s.flagNMFilter, "filter", "", "Server-side filter expression")
	s.netmgmtCTListCmd.Flags().StringVar(&s.flagNMOrderBy, "order-by", "", "Server-side ordering expression")
	s.netmgmtCTListCmd.Flags().Int64Var(&s.flagNMPageSize, "page-size", 0, "Number of results per page")
	s.netmgmtCTListCmd.Flags().Int64Var(&s.flagNMLimit, "limit", 0, "Maximum number of results to return")

	s.netmgmtConnectivityTestsCmd.AddCommand(
		s.netmgmtCTCreateCmd, s.netmgmtCTDeleteCmd, s.netmgmtCTDescribeCmd,
		s.netmgmtCTListCmd, s.netmgmtCTRerunCmd, s.netmgmtCTUpdateCmd,
	)

	// Operations under locations/global.
	for _, c := range []*cobra.Command{s.netmgmtOpsCancelCmd, s.netmgmtOpsDeleteCmd, s.netmgmtOpsDescribeCmd, s.netmgmtOpsWaitCmd} {
		c.Flags().StringVar(&s.flagNMLocation, "location", "global", "Location containing the operation")
	}
	s.netmgmtOpsListCmd.Flags().StringVar(&s.flagNMLocation, "location", "global", "Location to list operations in")
	s.netmgmtOpsListCmd.Flags().StringVar(&s.flagNMFilter, "filter", "", "Server-side filter expression")
	s.netmgmtOpsListCmd.Flags().Int64Var(&s.flagNMPageSize, "page-size", 0, "Number of results per page")
	s.netmgmtOpsListCmd.Flags().Int64Var(&s.flagNMLimit, "limit", 0, "Maximum number of results to return")
	s.netmgmtOperationsCmd.AddCommand(
		s.netmgmtOpsCancelCmd, s.netmgmtOpsDeleteCmd, s.netmgmtOpsDescribeCmd,
		s.netmgmtOpsListCmd, s.netmgmtOpsWaitCmd,
	)

	// VPC Flow Logs configs.
	for _, c := range []*cobra.Command{s.netmgmtVFLCreateCmd, s.netmgmtVFLDeleteCmd, s.netmgmtVFLDescribeCmd, s.netmgmtVFLUpdateCmd} {
		c.Flags().StringVar(&s.flagNMLocation, "location", "global", "Location containing the config")
	}
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMFile, "config-from-file", "", "YAML/JSON file with the VpcFlowLogsConfig spec")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLDescription, "description", "", "Description")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLInterconnectAtt, "interconnect-attachment", "", "Full resource path of an interconnect attachment to monitor")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLVpnTunnel, "vpn-tunnel", "", "Full resource path of a VPN tunnel to monitor")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLSubnet, "subnet", "", "Full resource path of a subnet to monitor")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLNetwork, "network", "", "Full resource path of a VPC network to monitor")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLAggInterval, "aggregation-interval", "", "Aggregation interval (e.g. INTERVAL_5_SEC)")
	s.netmgmtVFLCreateCmd.Flags().Float64Var(&s.flagNMVFLFlowSampling, "flow-sampling", 0, "Sampling rate (0..1)")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLMetadata, "metadata", "", "Metadata scope (INCLUDE_ALL_METADATA, EXCLUDE_ALL_METADATA, CUSTOM_METADATA)")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLFilterExpr, "filter-expr", "", "CEL filter to further select which flows are logged")
	s.netmgmtVFLCreateCmd.Flags().StringVar(&s.flagNMVFLState, "state", "", "State (ENABLED, DISABLED)")
	s.netmgmtVFLCreateCmd.Flags().StringToStringVar(&s.flagNMVFLLabels, "labels", nil, "Labels (key=value)")

	s.netmgmtVFLUpdateCmd.Flags().StringVar(&s.flagNMFile, "config-from-file", "", "YAML/JSON file with the VpcFlowLogsConfig patch")
	s.netmgmtVFLUpdateCmd.Flags().StringVar(&s.flagNMVFLDescription, "description", "", "New description")
	s.netmgmtVFLUpdateCmd.Flags().StringVar(&s.flagNMVFLAggInterval, "aggregation-interval", "", "Aggregation interval")
	s.netmgmtVFLUpdateCmd.Flags().Float64Var(&s.flagNMVFLFlowSampling, "flow-sampling", 0, "Sampling rate (0..1)")
	s.netmgmtVFLUpdateCmd.Flags().StringVar(&s.flagNMVFLMetadata, "metadata", "", "Metadata scope")
	s.netmgmtVFLUpdateCmd.Flags().StringVar(&s.flagNMVFLFilterExpr, "filter-expr", "", "CEL filter")
	s.netmgmtVFLUpdateCmd.Flags().StringVar(&s.flagNMVFLState, "state", "", "State")
	s.netmgmtVFLUpdateCmd.Flags().StringToStringVar(&s.flagNMVFLLabels, "labels", nil, "Labels (replaces the existing set)")

	s.netmgmtVFLListCmd.Flags().StringVar(&s.flagNMLocation, "location", "global", "Location to list configs in")
	s.netmgmtVFLListCmd.Flags().StringVar(&s.flagNMFilter, "filter", "", "Server-side filter expression")
	s.netmgmtVFLListCmd.Flags().StringVar(&s.flagNMOrderBy, "order-by", "", "Server-side ordering expression")
	s.netmgmtVFLListCmd.Flags().Int64Var(&s.flagNMPageSize, "page-size", 0, "Number of results per page")
	s.netmgmtVFLListCmd.Flags().Int64Var(&s.flagNMLimit, "limit", 0, "Maximum number of results to return")

	s.netmgmtVFLCmd.AddCommand(
		s.netmgmtVFLCreateCmd, s.netmgmtVFLDeleteCmd, s.netmgmtVFLDescribeCmd,
		s.netmgmtVFLListCmd, s.netmgmtVFLUpdateCmd,
	)

	s.networkManagementCmd.AddCommand(
		s.netmgmtConnectivityTestsCmd, s.netmgmtOperationsCmd, s.netmgmtVFLCmd,
	)

}

// --- Helpers ---

func (s *networkManagementCmds) nmProject() (string, error) {
	return s.inv.project()
}

func (s *networkManagementCmds) nmParent(project string) string {
	loc := s.flagNMLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, loc)
}

// nmQualify turns a bare id into a fully-qualified resource name under
// projects/PROJECT/locations/LOCATION/<kind>/. A caller-supplied full path
// (starting with "projects/") is returned unchanged.
func (s *networkManagementCmds) nmQualify(id, project, kind string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", s.nmParent(project), kind, id)
}

// --- Connectivity tests impl ---

func (s *networkManagementCmds) buildConnectivityTest(base *networkmanagement.ConnectivityTest) *networkmanagement.ConnectivityTest {
	t := base
	if t == nil {
		t = &networkmanagement.ConnectivityTest{}
	}
	if s.flagNMTestDescription != "" {
		t.Description = s.flagNMTestDescription
	}
	if s.flagNMTestProtocol != "" {
		t.Protocol = s.flagNMTestProtocol
	}
	if len(s.flagNMTestLabels) > 0 {
		t.Labels = s.flagNMTestLabels
	}
	if s.flagNMTestSourceIP != "" || s.flagNMTestSourceProj != "" || s.flagNMTestSourceNet != "" {
		if t.Source == nil {
			t.Source = &networkmanagement.Endpoint{}
		}
		if s.flagNMTestSourceIP != "" {
			t.Source.IpAddress = s.flagNMTestSourceIP
		}
		if s.flagNMTestSourceProj != "" {
			t.Source.ProjectId = s.flagNMTestSourceProj
		}
		if s.flagNMTestSourceNet != "" {
			t.Source.Network = s.flagNMTestSourceNet
		}
	}
	if s.flagNMTestDestIP != "" || s.flagNMTestDestPort > 0 || s.flagNMTestDestProj != "" || s.flagNMTestDestNet != "" {
		if t.Destination == nil {
			t.Destination = &networkmanagement.Endpoint{}
		}
		if s.flagNMTestDestIP != "" {
			t.Destination.IpAddress = s.flagNMTestDestIP
		}
		if s.flagNMTestDestPort > 0 {
			t.Destination.Port = s.flagNMTestDestPort
		}
		if s.flagNMTestDestProj != "" {
			t.Destination.ProjectId = s.flagNMTestDestProj
		}
		if s.flagNMTestDestNet != "" {
			t.Destination.Network = s.flagNMTestDestNet
		}
	}
	return t
}

func (s *networkManagementCmds) runNMCTCreate(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	var base *networkmanagement.ConnectivityTest
	if s.flagNMFile != "" {
		base = &networkmanagement.ConnectivityTest{}
		if err := loadYAMLOrJSONInto(s.flagNMFile, base); err != nil {
			return err
		}
	}
	test := s.buildConnectivityTest(base)
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.ConnectivityTests.Create(s.nmParent(project), test).TestId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connectivity test: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for connectivity test [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *networkManagementCmds) runNMCTDelete(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.ConnectivityTests.Delete(s.nmQualify(args[0], project, "connectivityTests")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connectivity test: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for connectivity test [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *networkManagementCmds) runNMCTDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.ConnectivityTests.Get(s.nmQualify(args[0], project, "connectivityTests")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connectivity test: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *networkManagementCmds) runNMCTList(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	var all []*networkmanagement.ConnectivityTest
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.ConnectivityTests.List(s.nmParent(project)).Context(ctx)
		if s.flagNMPageSize > 0 {
			call = call.PageSize(s.flagNMPageSize)
		}
		if s.flagNMFilter != "" {
			call = call.Filter(s.flagNMFilter)
		}
		if s.flagNMOrderBy != "" {
			call = call.OrderBy(s.flagNMOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing connectivity tests: %w", err)
		}
		all = append(all, resp.Resources...)
		if s.flagNMLimit > 0 && int64(len(all)) >= s.flagNMLimit {
			all = all[:s.flagNMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *networkManagementCmds) runNMCTRerun(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.ConnectivityTests.Rerun(s.nmQualify(args[0], project, "connectivityTests"), &networkmanagement.RerunConnectivityTestRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rerunning connectivity test: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Rerun request issued for connectivity test [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *networkManagementCmds) runNMCTUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	var base *networkmanagement.ConnectivityTest
	if s.flagNMFile != "" {
		base = &networkmanagement.ConnectivityTest{}
		if err := loadYAMLOrJSONInto(s.flagNMFile, base); err != nil {
			return err
		}
	}
	test := s.buildConnectivityTest(base)
	var mask []string
	if s.flagNMTestDescription != "" {
		mask = append(mask, "description")
	}
	if s.flagNMTestProtocol != "" {
		mask = append(mask, "protocol")
	}
	if len(s.flagNMTestLabels) > 0 {
		mask = append(mask, "labels")
	}
	if s.flagNMTestSourceIP != "" || s.flagNMTestSourceProj != "" || s.flagNMTestSourceNet != "" {
		mask = append(mask, "source")
	}
	if s.flagNMTestDestIP != "" || s.flagNMTestDestPort > 0 || s.flagNMTestDestProj != "" || s.flagNMTestDestNet != "" {
		mask = append(mask, "destination")
	}
	if s.flagNMFile != "" && len(mask) == 0 {
		mask = nonEmptyJSONFields(test)
	}
	if len(mask) == 0 {
		return fmt.Errorf("nothing to update: pass one of --description, --protocol, --labels, --source-*, --destination-*, or --config-from-file")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.ConnectivityTests.Patch(s.nmQualify(args[0], project, "connectivityTests"), test).
		UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connectivity test: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for connectivity test [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Operations impl ---

func (s *networkManagementCmds) runNMOpsCancel(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Global.Operations.Cancel(s.nmQualify(args[0], project, "operations"), &networkmanagement.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *networkManagementCmds) runNMOpsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Global.Operations.Delete(s.nmQualify(args[0], project, "operations")).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *networkManagementCmds) runNMOpsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Operations.Get(s.nmQualify(args[0], project, "operations")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, "")
}

func (s *networkManagementCmds) runNMOpsList(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	var all []*networkmanagement.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Operations.List(s.nmParent(project)).Context(ctx)
		if s.flagNMPageSize > 0 {
			call = call.PageSize(s.flagNMPageSize)
		}
		if s.flagNMFilter != "" {
			call = call.Filter(s.flagNMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if s.flagNMLimit > 0 && int64(len(all)) >= s.flagNMLimit {
			all = all[:s.flagNMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *networkManagementCmds) runNMOpsWait(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	name := s.nmQualify(args[0], project, "operations")
	for {
		op, err := svc.Projects.Locations.Global.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation %s failed: %s", name, op.Error.Message)
			}
			return emitFormattedTo(s.inv.stdout, op, "")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// --- VPC Flow Logs configs impl ---

func (s *networkManagementCmds) buildVpcFlowLogsConfig(base *networkmanagement.VpcFlowLogsConfig) *networkmanagement.VpcFlowLogsConfig {
	c := base
	if c == nil {
		c = &networkmanagement.VpcFlowLogsConfig{}
	}
	if s.flagNMVFLDescription != "" {
		c.Description = s.flagNMVFLDescription
	}
	if s.flagNMVFLInterconnectAtt != "" {
		c.InterconnectAttachment = s.flagNMVFLInterconnectAtt
	}
	if s.flagNMVFLVpnTunnel != "" {
		c.VpnTunnel = s.flagNMVFLVpnTunnel
	}
	if s.flagNMVFLSubnet != "" {
		c.Subnet = s.flagNMVFLSubnet
	}
	if s.flagNMVFLNetwork != "" {
		c.Network = s.flagNMVFLNetwork
	}
	if s.flagNMVFLAggInterval != "" {
		c.AggregationInterval = s.flagNMVFLAggInterval
	}
	if s.flagNMVFLFlowSampling > 0 {
		c.FlowSampling = s.flagNMVFLFlowSampling
	}
	if s.flagNMVFLMetadata != "" {
		c.Metadata = s.flagNMVFLMetadata
	}
	if s.flagNMVFLFilterExpr != "" {
		c.FilterExpr = s.flagNMVFLFilterExpr
	}
	if s.flagNMVFLState != "" {
		c.State = s.flagNMVFLState
	}
	if len(s.flagNMVFLLabels) > 0 {
		c.Labels = s.flagNMVFLLabels
	}
	return c
}

func (s *networkManagementCmds) runNMVFLCreate(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	var base *networkmanagement.VpcFlowLogsConfig
	if s.flagNMFile != "" {
		base = &networkmanagement.VpcFlowLogsConfig{}
		if err := loadYAMLOrJSONInto(s.flagNMFile, base); err != nil {
			return err
		}
	}
	cfg := s.buildVpcFlowLogsConfig(base)
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VpcFlowLogsConfigs.Create(s.nmParent(project), cfg).VpcFlowLogsConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating VPC flow logs config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for VPC flow logs config [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *networkManagementCmds) runNMVFLDelete(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VpcFlowLogsConfigs.Delete(s.nmQualify(args[0], project, "vpcFlowLogsConfigs")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting VPC flow logs config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for VPC flow logs config [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *networkManagementCmds) runNMVFLDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.VpcFlowLogsConfigs.Get(s.nmQualify(args[0], project, "vpcFlowLogsConfigs")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing VPC flow logs config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *networkManagementCmds) runNMVFLList(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	var all []*networkmanagement.VpcFlowLogsConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.VpcFlowLogsConfigs.List(s.nmParent(project)).Context(ctx)
		if s.flagNMPageSize > 0 {
			call = call.PageSize(s.flagNMPageSize)
		}
		if s.flagNMFilter != "" {
			call = call.Filter(s.flagNMFilter)
		}
		if s.flagNMOrderBy != "" {
			call = call.OrderBy(s.flagNMOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing VPC flow logs configs: %w", err)
		}
		all = append(all, resp.VpcFlowLogsConfigs...)
		if s.flagNMLimit > 0 && int64(len(all)) >= s.flagNMLimit {
			all = all[:s.flagNMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *networkManagementCmds) runNMVFLUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.nmProject()
	if err != nil {
		return err
	}
	var base *networkmanagement.VpcFlowLogsConfig
	if s.flagNMFile != "" {
		base = &networkmanagement.VpcFlowLogsConfig{}
		if err := loadYAMLOrJSONInto(s.flagNMFile, base); err != nil {
			return err
		}
	}
	cfg := s.buildVpcFlowLogsConfig(base)
	var mask []string
	if s.flagNMVFLDescription != "" {
		mask = append(mask, "description")
	}
	if s.flagNMVFLAggInterval != "" {
		mask = append(mask, "aggregation_interval")
	}
	if s.flagNMVFLFlowSampling > 0 {
		mask = append(mask, "flow_sampling")
	}
	if s.flagNMVFLMetadata != "" {
		mask = append(mask, "metadata")
	}
	if s.flagNMVFLFilterExpr != "" {
		mask = append(mask, "filter_expr")
	}
	if s.flagNMVFLState != "" {
		mask = append(mask, "state")
	}
	if len(s.flagNMVFLLabels) > 0 {
		mask = append(mask, "labels")
	}
	if s.flagNMFile != "" && len(mask) == 0 {
		mask = nonEmptyJSONFields(cfg)
	}
	if len(mask) == 0 {
		return fmt.Errorf("nothing to update: pass at least one of --description, --aggregation-interval, --flow-sampling, --metadata, --filter-expr, --state, --labels, or --config-from-file")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkManagement(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VpcFlowLogsConfigs.Patch(s.nmQualify(args[0], project, "vpcFlowLogsConfigs"), cfg).
		UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating VPC flow logs config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for VPC flow logs config [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// networkManagementCmds holds the commands and flag values of one invocation's "networkManagement"
// command group, so invocations don't share state (#2070).
type networkManagementCmds struct {
	inv *invocation

	networkManagementCmd *cobra.Command
	flagNMLocation       string
	flagNMFile           string
	flagNMFilter         string
	flagNMOrderBy        string
	flagNMPageSize       int64
	flagNMLimit          int64
	// Connectivity test short-form flags.
	flagNMTestDescription string
	flagNMTestProtocol    string
	flagNMTestSourceIP    string
	flagNMTestSourceProj  string
	flagNMTestSourceNet   string
	flagNMTestDestIP      string
	flagNMTestDestPort    int64
	flagNMTestDestProj    string
	flagNMTestDestNet     string
	flagNMTestLabels      map[string]string
	// VPC flow logs config short-form flags.
	flagNMVFLDescription        string
	flagNMVFLInterconnectAtt    string
	flagNMVFLVpnTunnel          string
	flagNMVFLSubnet             string
	flagNMVFLNetwork            string
	flagNMVFLAggInterval        string
	flagNMVFLFlowSampling       float64
	flagNMVFLMetadata           string
	flagNMVFLFilterExpr         string
	flagNMVFLState              string
	flagNMVFLLabels             map[string]string
	netmgmtConnectivityTestsCmd *cobra.Command
	netmgmtCTCreateCmd          *cobra.Command
	netmgmtCTDeleteCmd          *cobra.Command
	netmgmtCTDescribeCmd        *cobra.Command
	netmgmtCTListCmd            *cobra.Command
	netmgmtCTRerunCmd           *cobra.Command
	netmgmtCTUpdateCmd          *cobra.Command
	netmgmtOperationsCmd        *cobra.Command
	netmgmtOpsCancelCmd         *cobra.Command
	netmgmtOpsDeleteCmd         *cobra.Command
	netmgmtOpsDescribeCmd       *cobra.Command
	netmgmtOpsListCmd           *cobra.Command
	netmgmtOpsWaitCmd           *cobra.Command
	netmgmtVFLCmd               *cobra.Command
	netmgmtVFLCreateCmd         *cobra.Command
	netmgmtVFLDeleteCmd         *cobra.Command
	netmgmtVFLDescribeCmd       *cobra.Command
	netmgmtVFLListCmd           *cobra.Command
	netmgmtVFLUpdateCmd         *cobra.Command
	netmgmtMonCmd               *cobra.Command
	flagNMMonLocation           string
	flagNMMonFormat             string
	flagNMMonConfigFile         string
	flagNMMonPageSize           int64
	netmgmtMonCreateCmd         *cobra.Command
	netmgmtMonDeleteCmd         *cobra.Command
	netmgmtMonDescribeCmd       *cobra.Command
	netmgmtMonGenMPConfigCmd    *cobra.Command
	netmgmtMonGenAccessTokenCmd *cobra.Command
	netmgmtMonListCmd           *cobra.Command
}

// newNetworkManagementCmd returns the "networkManagement" command group for inv.
func newNetworkManagementCmd(inv *invocation) *cobra.Command {
	return newNetworkManagementCmds(inv).networkManagementCmd
}

// newNetworkManagementCmds builds the "networkManagement" command group for inv and returns its state.
func newNetworkManagementCmds(inv *invocation) *networkManagementCmds {
	s := &networkManagementCmds{inv: inv}
	s.networkManagementCmd = &cobra.Command{
		Use:   "network-management",
		Short: "Manage Network Management (Network Intelligence Center)",
	}
	s.netmgmtConnectivityTestsCmd = &cobra.Command{
		Use:   "connectivity-tests",
		Short: "Manage connectivity tests (Network Intelligence Center)",
	}
	s.netmgmtCTCreateCmd = &cobra.Command{
		Use:   "create TEST",
		Short: "Create a connectivity test",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMCTCreate,
	}
	s.netmgmtCTDeleteCmd = &cobra.Command{
		Use:   "delete TEST",
		Short: "Delete a connectivity test",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMCTDelete,
	}
	s.netmgmtCTDescribeCmd = &cobra.Command{
		Use:   "describe TEST",
		Short: "Describe a connectivity test",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMCTDescribe,
	}
	s.netmgmtCTListCmd = &cobra.Command{
		Use:   "list",
		Short: "List connectivity tests",
		Args:  cobra.NoArgs,
		RunE:  s.runNMCTList,
	}
	s.netmgmtCTRerunCmd = &cobra.Command{
		Use:   "rerun TEST",
		Short: "Rerun a connectivity test",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMCTRerun,
	}
	s.netmgmtCTUpdateCmd = &cobra.Command{
		Use:   "update TEST",
		Short: "Update a connectivity test",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMCTUpdate,
	}
	s.netmgmtOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Network Management operations",
	}
	s.netmgmtOpsCancelCmd = &cobra.Command{
		Use:   "cancel OPERATION",
		Short: "Cancel an operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMOpsCancel,
	}
	s.netmgmtOpsDeleteCmd = &cobra.Command{
		Use:   "delete OPERATION",
		Short: "Delete an operation record",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMOpsDelete,
	}
	s.netmgmtOpsDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe an operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMOpsDescribe,
	}
	s.netmgmtOpsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List operations",
		Args:  cobra.NoArgs,
		RunE:  s.runNMOpsList,
	}
	s.netmgmtOpsWaitCmd = &cobra.Command{
		Use:   "wait OPERATION",
		Short: "Poll an operation until it completes",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMOpsWait,
	}
	s.netmgmtVFLCmd = &cobra.Command{
		Use:   "vpc-flow-logs-configs",
		Short: "Manage VPC Flow Logs configs",
	}
	s.netmgmtVFLCreateCmd = &cobra.Command{
		Use:   "create CONFIG",
		Short: "Create a VPC flow logs config",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMVFLCreate,
	}
	s.netmgmtVFLDeleteCmd = &cobra.Command{
		Use:   "delete CONFIG",
		Short: "Delete a VPC flow logs config",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMVFLDelete,
	}
	s.netmgmtVFLDescribeCmd = &cobra.Command{
		Use:   "describe CONFIG",
		Short: "Describe a VPC flow logs config",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMVFLDescribe,
	}
	s.netmgmtVFLListCmd = &cobra.Command{
		Use:   "list",
		Short: "List VPC flow logs configs",
		Args:  cobra.NoArgs,
		RunE:  s.runNMVFLList,
	}
	s.netmgmtVFLUpdateCmd = &cobra.Command{
		Use:   "update CONFIG",
		Short: "Update a VPC flow logs config",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runNMVFLUpdate,
	}
	s.netmgmtMonCmd = &cobra.Command{
		Use:   "network-monitoring-providers",
		Short: "Manage Network Monitoring providers",
	}
	s.netmgmtMonCreateCmd = &cobra.Command{
		Use: "create PROVIDER", Short: "Create a network monitoring provider",
		Args: cobra.ExactArgs(1), RunE: s.runNMMonCreate,
	}
	s.netmgmtMonDeleteCmd = &cobra.Command{
		Use: "delete PROVIDER", Short: "Delete a network monitoring provider",
		Args: cobra.ExactArgs(1), RunE: s.runNMMonDelete,
	}
	s.netmgmtMonDescribeCmd = &cobra.Command{
		Use: "describe PROVIDER", Short: "Describe a network monitoring provider",
		Args: cobra.ExactArgs(1), RunE: s.runNMMonDescribe,
	}
	s.netmgmtMonGenMPConfigCmd = &cobra.Command{
		Use:   "generate-monitoring-point-config PROVIDER",
		Short: "Generate a monitoring point config for a network monitoring provider",
		Args:  cobra.ExactArgs(1), RunE: s.runNMMonGenMPConfig,
	}
	s.netmgmtMonGenAccessTokenCmd = &cobra.Command{
		Use:   "generate-provider-access-token PROVIDER",
		Short: "Generate a provider access token for a network monitoring provider",
		Args:  cobra.ExactArgs(1), RunE: s.runNMMonGenAccessToken,
	}
	s.netmgmtMonListCmd = &cobra.Command{
		Use: "list", Short: "List network monitoring providers",
		Args: cobra.NoArgs, RunE: s.runNMMonList,
	}
	s.initNetworkManagement()
	s.initNetworkManagementNetworkMonitoringProviders()
	return s
}
