package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkmanagement "google.golang.org/api/networkmanagement/v1"
)

// --- gcloud network-management (#953, #955, #956) ---
//
// Note: the "network-monitoring-providers" subgroup (#954) is deferred: the
// google.golang.org/api/networkmanagement/v1 client at v0.279.0 does not
// expose the underlying REST resource, so we keep it registered as a stub so
// users still see it in `--help`.

var networkManagementCmd = &cobra.Command{
	Use:   "network-management",
	Short: "Manage Network Management (Network Intelligence Center)",
}

// Common flags across the network-management surface. Connectivity tests and
// network-monitoring providers all live under `locations/global`, so we
// default there. VPC flow logs configs and operations are regional.
var (
	flagNMLocation string
	flagNMFile     string
	flagNMFilter   string
	flagNMOrderBy  string
	flagNMPageSize int64
	flagNMLimit    int64

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
	flagNMVFLDescription     string
	flagNMVFLInterconnectAtt string
	flagNMVFLVpnTunnel       string
	flagNMVFLSubnet          string
	flagNMVFLNetwork         string
	flagNMVFLAggInterval     string
	flagNMVFLFlowSampling    float64
	flagNMVFLMetadata        string
	flagNMVFLFilterExpr      string
	flagNMVFLState           string
	flagNMVFLLabels          map[string]string
)

// --- Connectivity tests ---

var netmgmtConnectivityTestsCmd = &cobra.Command{
	Use:   "connectivity-tests",
	Short: "Manage connectivity tests (Network Intelligence Center)",
}

var netmgmtCTCreateCmd = &cobra.Command{
	Use:   "create TEST",
	Short: "Create a connectivity test",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMCTCreate,
}

var netmgmtCTDeleteCmd = &cobra.Command{
	Use:   "delete TEST",
	Short: "Delete a connectivity test",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMCTDelete,
}

var netmgmtCTDescribeCmd = &cobra.Command{
	Use:   "describe TEST",
	Short: "Describe a connectivity test",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMCTDescribe,
}

var netmgmtCTListCmd = &cobra.Command{
	Use:   "list",
	Short: "List connectivity tests",
	Args:  cobra.NoArgs,
	RunE:  runNMCTList,
}

var netmgmtCTRerunCmd = &cobra.Command{
	Use:   "rerun TEST",
	Short: "Rerun a connectivity test",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMCTRerun,
}

var netmgmtCTUpdateCmd = &cobra.Command{
	Use:   "update TEST",
	Short: "Update a connectivity test",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMCTUpdate,
}

// --- Operations ---

var netmgmtOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Network Management operations",
}

var netmgmtOpsCancelCmd = &cobra.Command{
	Use:   "cancel OPERATION",
	Short: "Cancel an operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMOpsCancel,
}

var netmgmtOpsDeleteCmd = &cobra.Command{
	Use:   "delete OPERATION",
	Short: "Delete an operation record",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMOpsDelete,
}

var netmgmtOpsDescribeCmd = &cobra.Command{
	Use:   "describe OPERATION",
	Short: "Describe an operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMOpsDescribe,
}

var netmgmtOpsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List operations",
	Args:  cobra.NoArgs,
	RunE:  runNMOpsList,
}

var netmgmtOpsWaitCmd = &cobra.Command{
	Use:   "wait OPERATION",
	Short: "Poll an operation until it completes",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMOpsWait,
}

// --- VPC Flow Logs configs ---

var netmgmtVFLCmd = &cobra.Command{
	Use:   "vpc-flow-logs-configs",
	Short: "Manage VPC Flow Logs configs",
}

var netmgmtVFLCreateCmd = &cobra.Command{
	Use:   "create CONFIG",
	Short: "Create a VPC flow logs config",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMVFLCreate,
}

var netmgmtVFLDeleteCmd = &cobra.Command{
	Use:   "delete CONFIG",
	Short: "Delete a VPC flow logs config",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMVFLDelete,
}

var netmgmtVFLDescribeCmd = &cobra.Command{
	Use:   "describe CONFIG",
	Short: "Describe a VPC flow logs config",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMVFLDescribe,
}

var netmgmtVFLListCmd = &cobra.Command{
	Use:   "list",
	Short: "List VPC flow logs configs",
	Args:  cobra.NoArgs,
	RunE:  runNMVFLList,
}

var netmgmtVFLUpdateCmd = &cobra.Command{
	Use:   "update CONFIG",
	Short: "Update a VPC flow logs config",
	Args:  cobra.ExactArgs(1),
	RunE:  runNMVFLUpdate,
}

func init() {
	// Connectivity tests always live under locations/global; the flag exists
	// only for parity with other network-management resources.
	for _, c := range []*cobra.Command{
		netmgmtCTCreateCmd, netmgmtCTDeleteCmd, netmgmtCTDescribeCmd,
		netmgmtCTRerunCmd, netmgmtCTUpdateCmd,
	} {
		c.Flags().StringVar(&flagNMLocation, "location", "global", "Location containing the test")
	}
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMFile, "config-from-file", "", "YAML/JSON file with the ConnectivityTest spec")
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMTestDescription, "description", "", "Description of the test")
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMTestProtocol, "protocol", "", "IP protocol (e.g. TCP, UDP, ICMP)")
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMTestSourceIP, "source-ip-address", "", "Source IP")
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMTestSourceProj, "source-project", "", "Source project ID")
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMTestSourceNet, "source-network", "", "Source network URI")
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMTestDestIP, "destination-ip-address", "", "Destination IP")
	netmgmtCTCreateCmd.Flags().Int64Var(&flagNMTestDestPort, "destination-port", 0, "Destination port")
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMTestDestProj, "destination-project", "", "Destination project ID")
	netmgmtCTCreateCmd.Flags().StringVar(&flagNMTestDestNet, "destination-network", "", "Destination network URI")
	netmgmtCTCreateCmd.Flags().StringToStringVar(&flagNMTestLabels, "labels", nil, "Labels (key=value)")

	netmgmtCTUpdateCmd.Flags().StringVar(&flagNMFile, "config-from-file", "", "YAML/JSON file with the ConnectivityTest patch")
	netmgmtCTUpdateCmd.Flags().StringVar(&flagNMTestDescription, "description", "", "New description")
	netmgmtCTUpdateCmd.Flags().StringToStringVar(&flagNMTestLabels, "labels", nil, "Labels (replaces the existing set)")

	netmgmtCTListCmd.Flags().StringVar(&flagNMLocation, "location", "global", "Location to list tests in")
	netmgmtCTListCmd.Flags().StringVar(&flagNMFilter, "filter", "", "Server-side filter expression")
	netmgmtCTListCmd.Flags().StringVar(&flagNMOrderBy, "order-by", "", "Server-side ordering expression")
	netmgmtCTListCmd.Flags().Int64Var(&flagNMPageSize, "page-size", 0, "Number of results per page")
	netmgmtCTListCmd.Flags().Int64Var(&flagNMLimit, "limit", 0, "Maximum number of results to return")

	netmgmtConnectivityTestsCmd.AddCommand(
		netmgmtCTCreateCmd, netmgmtCTDeleteCmd, netmgmtCTDescribeCmd,
		netmgmtCTListCmd, netmgmtCTRerunCmd, netmgmtCTUpdateCmd,
	)

	// Operations under locations/global.
	for _, c := range []*cobra.Command{netmgmtOpsCancelCmd, netmgmtOpsDeleteCmd, netmgmtOpsDescribeCmd, netmgmtOpsWaitCmd} {
		c.Flags().StringVar(&flagNMLocation, "location", "global", "Location containing the operation")
	}
	netmgmtOpsListCmd.Flags().StringVar(&flagNMLocation, "location", "global", "Location to list operations in")
	netmgmtOpsListCmd.Flags().StringVar(&flagNMFilter, "filter", "", "Server-side filter expression")
	netmgmtOpsListCmd.Flags().Int64Var(&flagNMPageSize, "page-size", 0, "Number of results per page")
	netmgmtOpsListCmd.Flags().Int64Var(&flagNMLimit, "limit", 0, "Maximum number of results to return")
	netmgmtOperationsCmd.AddCommand(
		netmgmtOpsCancelCmd, netmgmtOpsDeleteCmd, netmgmtOpsDescribeCmd,
		netmgmtOpsListCmd, netmgmtOpsWaitCmd,
	)

	// VPC Flow Logs configs.
	for _, c := range []*cobra.Command{netmgmtVFLCreateCmd, netmgmtVFLDeleteCmd, netmgmtVFLDescribeCmd, netmgmtVFLUpdateCmd} {
		c.Flags().StringVar(&flagNMLocation, "location", "global", "Location containing the config")
	}
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMFile, "config-from-file", "", "YAML/JSON file with the VpcFlowLogsConfig spec")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLDescription, "description", "", "Description")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLInterconnectAtt, "interconnect-attachment", "", "Full resource path of an interconnect attachment to monitor")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLVpnTunnel, "vpn-tunnel", "", "Full resource path of a VPN tunnel to monitor")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLSubnet, "subnet", "", "Full resource path of a subnet to monitor")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLNetwork, "network", "", "Full resource path of a VPC network to monitor")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLAggInterval, "aggregation-interval", "", "Aggregation interval (e.g. INTERVAL_5_SEC)")
	netmgmtVFLCreateCmd.Flags().Float64Var(&flagNMVFLFlowSampling, "flow-sampling", 0, "Sampling rate (0..1)")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLMetadata, "metadata", "", "Metadata scope (INCLUDE_ALL_METADATA, EXCLUDE_ALL_METADATA, CUSTOM_METADATA)")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLFilterExpr, "filter-expr", "", "CEL filter to further select which flows are logged")
	netmgmtVFLCreateCmd.Flags().StringVar(&flagNMVFLState, "state", "", "State (ENABLED, DISABLED)")
	netmgmtVFLCreateCmd.Flags().StringToStringVar(&flagNMVFLLabels, "labels", nil, "Labels (key=value)")

	netmgmtVFLUpdateCmd.Flags().StringVar(&flagNMFile, "config-from-file", "", "YAML/JSON file with the VpcFlowLogsConfig patch")
	netmgmtVFLUpdateCmd.Flags().StringVar(&flagNMVFLDescription, "description", "", "New description")
	netmgmtVFLUpdateCmd.Flags().StringVar(&flagNMVFLAggInterval, "aggregation-interval", "", "Aggregation interval")
	netmgmtVFLUpdateCmd.Flags().Float64Var(&flagNMVFLFlowSampling, "flow-sampling", 0, "Sampling rate (0..1)")
	netmgmtVFLUpdateCmd.Flags().StringVar(&flagNMVFLMetadata, "metadata", "", "Metadata scope")
	netmgmtVFLUpdateCmd.Flags().StringVar(&flagNMVFLFilterExpr, "filter-expr", "", "CEL filter")
	netmgmtVFLUpdateCmd.Flags().StringVar(&flagNMVFLState, "state", "", "State")
	netmgmtVFLUpdateCmd.Flags().StringToStringVar(&flagNMVFLLabels, "labels", nil, "Labels (replaces the existing set)")

	netmgmtVFLListCmd.Flags().StringVar(&flagNMLocation, "location", "global", "Location to list configs in")
	netmgmtVFLListCmd.Flags().StringVar(&flagNMFilter, "filter", "", "Server-side filter expression")
	netmgmtVFLListCmd.Flags().StringVar(&flagNMOrderBy, "order-by", "", "Server-side ordering expression")
	netmgmtVFLListCmd.Flags().Int64Var(&flagNMPageSize, "page-size", 0, "Number of results per page")
	netmgmtVFLListCmd.Flags().Int64Var(&flagNMLimit, "limit", 0, "Maximum number of results to return")

	netmgmtVFLCmd.AddCommand(
		netmgmtVFLCreateCmd, netmgmtVFLDeleteCmd, netmgmtVFLDescribeCmd,
		netmgmtVFLListCmd, netmgmtVFLUpdateCmd,
	)

	// network-monitoring-providers (#954): not modeled by the Go client
	// google.golang.org/api/networkmanagement/v1 at v0.279.0. Keep as stubs so
	// the subgroup still appears in --help and users get a clear message.
	registerStubGroup(networkManagementCmd, "network-monitoring-providers",
		"(Not yet implemented) Manage network monitoring providers",
		"create", "delete", "describe", "list", "update",
	)

	networkManagementCmd.AddCommand(
		netmgmtConnectivityTestsCmd, netmgmtOperationsCmd, netmgmtVFLCmd,
	)
	rootCmd.AddCommand(networkManagementCmd)
}

// --- Helpers ---

func nmProject() (string, error) {
	return resolveProject()
}

func nmParent(project string) string {
	loc := flagNMLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, loc)
}

// nmQualify turns a bare id into a fully-qualified resource name under
// projects/PROJECT/locations/LOCATION/<kind>/. A caller-supplied full path
// (starting with "projects/") is returned unchanged.
func nmQualify(id, project, kind string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", nmParent(project), kind, id)
}

// --- Connectivity tests impl ---

func buildConnectivityTest(base *networkmanagement.ConnectivityTest) *networkmanagement.ConnectivityTest {
	t := base
	if t == nil {
		t = &networkmanagement.ConnectivityTest{}
	}
	if flagNMTestDescription != "" {
		t.Description = flagNMTestDescription
	}
	if flagNMTestProtocol != "" {
		t.Protocol = flagNMTestProtocol
	}
	if len(flagNMTestLabels) > 0 {
		t.Labels = flagNMTestLabels
	}
	if flagNMTestSourceIP != "" || flagNMTestSourceProj != "" || flagNMTestSourceNet != "" {
		if t.Source == nil {
			t.Source = &networkmanagement.Endpoint{}
		}
		if flagNMTestSourceIP != "" {
			t.Source.IpAddress = flagNMTestSourceIP
		}
		if flagNMTestSourceProj != "" {
			t.Source.ProjectId = flagNMTestSourceProj
		}
		if flagNMTestSourceNet != "" {
			t.Source.Network = flagNMTestSourceNet
		}
	}
	if flagNMTestDestIP != "" || flagNMTestDestPort > 0 || flagNMTestDestProj != "" || flagNMTestDestNet != "" {
		if t.Destination == nil {
			t.Destination = &networkmanagement.Endpoint{}
		}
		if flagNMTestDestIP != "" {
			t.Destination.IpAddress = flagNMTestDestIP
		}
		if flagNMTestDestPort > 0 {
			t.Destination.Port = flagNMTestDestPort
		}
		if flagNMTestDestProj != "" {
			t.Destination.ProjectId = flagNMTestDestProj
		}
		if flagNMTestDestNet != "" {
			t.Destination.Network = flagNMTestDestNet
		}
	}
	return t
}

func runNMCTCreate(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	var base *networkmanagement.ConnectivityTest
	if flagNMFile != "" {
		base = &networkmanagement.ConnectivityTest{}
		if err := loadYAMLOrJSONInto(flagNMFile, base); err != nil {
			return err
		}
	}
	test := buildConnectivityTest(base)
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.ConnectivityTests.Create(nmParent(project), test).TestId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connectivity test: %w", err)
	}
	fmt.Printf("Create request issued for connectivity test [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runNMCTDelete(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.ConnectivityTests.Delete(nmQualify(args[0], project, "connectivityTests")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connectivity test: %w", err)
	}
	fmt.Printf("Delete request issued for connectivity test [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runNMCTDescribe(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.ConnectivityTests.Get(nmQualify(args[0], project, "connectivityTests")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connectivity test: %w", err)
	}
	return emitFormatted(got, "")
}

func runNMCTList(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkmanagement.ConnectivityTest
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.ConnectivityTests.List(nmParent(project)).Context(ctx)
		if flagNMPageSize > 0 {
			call = call.PageSize(flagNMPageSize)
		}
		if flagNMFilter != "" {
			call = call.Filter(flagNMFilter)
		}
		if flagNMOrderBy != "" {
			call = call.OrderBy(flagNMOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing connectivity tests: %w", err)
		}
		all = append(all, resp.Resources...)
		if flagNMLimit > 0 && int64(len(all)) >= flagNMLimit {
			all = all[:flagNMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runNMCTRerun(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.ConnectivityTests.Rerun(nmQualify(args[0], project, "connectivityTests"), &networkmanagement.RerunConnectivityTestRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rerunning connectivity test: %w", err)
	}
	fmt.Printf("Rerun request issued for connectivity test [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runNMCTUpdate(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	var base *networkmanagement.ConnectivityTest
	if flagNMFile != "" {
		base = &networkmanagement.ConnectivityTest{}
		if err := loadYAMLOrJSONInto(flagNMFile, base); err != nil {
			return err
		}
	}
	test := buildConnectivityTest(base)
	var mask []string
	if flagNMTestDescription != "" {
		mask = append(mask, "description")
	}
	if flagNMTestProtocol != "" {
		mask = append(mask, "protocol")
	}
	if len(flagNMTestLabels) > 0 {
		mask = append(mask, "labels")
	}
	if flagNMTestSourceIP != "" || flagNMTestSourceProj != "" || flagNMTestSourceNet != "" {
		mask = append(mask, "source")
	}
	if flagNMTestDestIP != "" || flagNMTestDestPort > 0 || flagNMTestDestProj != "" || flagNMTestDestNet != "" {
		mask = append(mask, "destination")
	}
	if flagNMFile != "" && len(mask) == 0 {
		mask = nonEmptyJSONFields(test)
	}
	if len(mask) == 0 {
		return fmt.Errorf("nothing to update: pass one of --description, --protocol, --labels, --source-*, --destination-*, or --config-from-file")
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.ConnectivityTests.Patch(nmQualify(args[0], project, "connectivityTests"), test).
		UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connectivity test: %w", err)
	}
	fmt.Printf("Update request issued for connectivity test [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Operations impl ---

func runNMOpsCancel(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Global.Operations.Cancel(nmQualify(args[0], project, "operations"), &networkmanagement.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runNMOpsDelete(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Global.Operations.Delete(nmQualify(args[0], project, "operations")).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runNMOpsDescribe(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Operations.Get(nmQualify(args[0], project, "operations")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, "")
}

func runNMOpsList(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkmanagement.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Operations.List(nmParent(project)).Context(ctx)
		if flagNMPageSize > 0 {
			call = call.PageSize(flagNMPageSize)
		}
		if flagNMFilter != "" {
			call = call.Filter(flagNMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if flagNMLimit > 0 && int64(len(all)) >= flagNMLimit {
			all = all[:flagNMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runNMOpsWait(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nmQualify(args[0], project, "operations")
	for {
		op, err := svc.Projects.Locations.Global.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation %s failed: %s", name, op.Error.Message)
			}
			return emitFormatted(op, "")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// --- VPC Flow Logs configs impl ---

func buildVpcFlowLogsConfig(base *networkmanagement.VpcFlowLogsConfig) *networkmanagement.VpcFlowLogsConfig {
	c := base
	if c == nil {
		c = &networkmanagement.VpcFlowLogsConfig{}
	}
	if flagNMVFLDescription != "" {
		c.Description = flagNMVFLDescription
	}
	if flagNMVFLInterconnectAtt != "" {
		c.InterconnectAttachment = flagNMVFLInterconnectAtt
	}
	if flagNMVFLVpnTunnel != "" {
		c.VpnTunnel = flagNMVFLVpnTunnel
	}
	if flagNMVFLSubnet != "" {
		c.Subnet = flagNMVFLSubnet
	}
	if flagNMVFLNetwork != "" {
		c.Network = flagNMVFLNetwork
	}
	if flagNMVFLAggInterval != "" {
		c.AggregationInterval = flagNMVFLAggInterval
	}
	if flagNMVFLFlowSampling > 0 {
		c.FlowSampling = flagNMVFLFlowSampling
	}
	if flagNMVFLMetadata != "" {
		c.Metadata = flagNMVFLMetadata
	}
	if flagNMVFLFilterExpr != "" {
		c.FilterExpr = flagNMVFLFilterExpr
	}
	if flagNMVFLState != "" {
		c.State = flagNMVFLState
	}
	if len(flagNMVFLLabels) > 0 {
		c.Labels = flagNMVFLLabels
	}
	return c
}

func runNMVFLCreate(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	var base *networkmanagement.VpcFlowLogsConfig
	if flagNMFile != "" {
		base = &networkmanagement.VpcFlowLogsConfig{}
		if err := loadYAMLOrJSONInto(flagNMFile, base); err != nil {
			return err
		}
	}
	cfg := buildVpcFlowLogsConfig(base)
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VpcFlowLogsConfigs.Create(nmParent(project), cfg).VpcFlowLogsConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating VPC flow logs config: %w", err)
	}
	fmt.Printf("Create request issued for VPC flow logs config [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runNMVFLDelete(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VpcFlowLogsConfigs.Delete(nmQualify(args[0], project, "vpcFlowLogsConfigs")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting VPC flow logs config: %w", err)
	}
	fmt.Printf("Delete request issued for VPC flow logs config [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runNMVFLDescribe(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.VpcFlowLogsConfigs.Get(nmQualify(args[0], project, "vpcFlowLogsConfigs")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing VPC flow logs config: %w", err)
	}
	return emitFormatted(got, "")
}

func runNMVFLList(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkmanagement.VpcFlowLogsConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.VpcFlowLogsConfigs.List(nmParent(project)).Context(ctx)
		if flagNMPageSize > 0 {
			call = call.PageSize(flagNMPageSize)
		}
		if flagNMFilter != "" {
			call = call.Filter(flagNMFilter)
		}
		if flagNMOrderBy != "" {
			call = call.OrderBy(flagNMOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing VPC flow logs configs: %w", err)
		}
		all = append(all, resp.VpcFlowLogsConfigs...)
		if flagNMLimit > 0 && int64(len(all)) >= flagNMLimit {
			all = all[:flagNMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runNMVFLUpdate(cmd *cobra.Command, args []string) error {
	project, err := nmProject()
	if err != nil {
		return err
	}
	var base *networkmanagement.VpcFlowLogsConfig
	if flagNMFile != "" {
		base = &networkmanagement.VpcFlowLogsConfig{}
		if err := loadYAMLOrJSONInto(flagNMFile, base); err != nil {
			return err
		}
	}
	cfg := buildVpcFlowLogsConfig(base)
	var mask []string
	if flagNMVFLDescription != "" {
		mask = append(mask, "description")
	}
	if flagNMVFLAggInterval != "" {
		mask = append(mask, "aggregation_interval")
	}
	if flagNMVFLFlowSampling > 0 {
		mask = append(mask, "flow_sampling")
	}
	if flagNMVFLMetadata != "" {
		mask = append(mask, "metadata")
	}
	if flagNMVFLFilterExpr != "" {
		mask = append(mask, "filter_expr")
	}
	if flagNMVFLState != "" {
		mask = append(mask, "state")
	}
	if len(flagNMVFLLabels) > 0 {
		mask = append(mask, "labels")
	}
	if flagNMFile != "" && len(mask) == 0 {
		mask = nonEmptyJSONFields(cfg)
	}
	if len(mask) == 0 {
		return fmt.Errorf("nothing to update: pass at least one of --description, --aggregation-interval, --flow-sampling, --metadata, --filter-expr, --state, --labels, or --config-from-file")
	}
	ctx := context.Background()
	svc, err := gcp.NetworkManagementService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VpcFlowLogsConfigs.Patch(nmQualify(args[0], project, "vpcFlowLogsConfigs"), cfg).
		UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating VPC flow logs config: %w", err)
	}
	fmt.Printf("Update request issued for VPC flow logs config [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}
