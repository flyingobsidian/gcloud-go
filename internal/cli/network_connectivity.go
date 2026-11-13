package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	networkconnectivity "google.golang.org/api/networkconnectivity/v1"
)

// --- gcloud network-connectivity (#901-#911) ---
//
// Every subgroup here is backed by the google.golang.org/api/networkconnectivity
// service (v1). Layout mirrors internal/cli/network_security.go:
//
//   - a small builder for the shared --location / --config-file / --async / --format
//     flags,
//   - a generic ncCRUD binding that turns a resource-specific closure set into
//     the standard create/delete/describe/list/update subcommands,
//   - LRO create/delete/patch calls block by default and wait via ncWaitOp.

// ncProjectParent returns projects/PROJECT/locations/LOCATION for the current
// project and --location.
func (grp *networkConnectivityCmds) ncProjectParent() (string, error) {
	project, err := grp.inv.project()
	if err != nil {
		return "", err
	}
	if grp.flagNCLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, grp.flagNCLocation), nil
}

// ncGlobalParent returns projects/PROJECT/locations/global.
func (grp *networkConnectivityCmds) ncGlobalParent() (string, error) {
	project, err := grp.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/global", project), nil
}

// ncChild appends "collection/id" to parent unless id is already a fully
// qualified resource name.
func ncChild(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func ncBasename(name string) string { return path.Base(name) }

func (grp *networkConnectivityCmds) ncResolveMask(body any) string {
	if grp.flagNCUpdateMask != "" {
		return grp.flagNCUpdateMask
	}
	return joinMask(nonEmptyJSONFields(body))
}

// ncWaitOp polls a long-running operation until it completes.
func ncWaitOp(ctx context.Context, svc *networkconnectivity.Service, op *networkconnectivity.GoogleLongrunningOperation) (*networkconnectivity.GoogleLongrunningOperation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (grp *networkConnectivityCmds) ncFinishOp(ctx context.Context, svc *networkconnectivity.Service, op *networkconnectivity.GoogleLongrunningOperation, verb, name string) error {
	if grp.flagNCAsync {
		fmt.Fprintf(grp.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(grp.inv.stdout, op, "")
	}
	final, err := ncWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(grp.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(grp.inv.stdout, final.Response, "")
	}
	return nil
}

func (grp *networkConnectivityCmds) addNCLocationFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&grp.flagNCLocation, "location", "", "Location (required)")
	}
}

func (grp *networkConnectivityCmds) addNCGlobalLocationFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&grp.flagNCLocation, "location", "global", "Location (defaults to global)")
	}
}

func (grp *networkConnectivityCmds) addNCFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&grp.flagNCFormat, "format", "", "Output format")
	}
}

func (grp *networkConnectivityCmds) addNCFilterFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&grp.flagNCFilter, "filter", "", "Server-side list filter")
	}
}

func (grp *networkConnectivityCmds) addNCAsyncFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().BoolVar(&grp.flagNCAsync, "async", false, "Do not wait for the operation to complete")
	}
}

func (grp *networkConnectivityCmds) addNCRequestIDFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&grp.flagNCRequestID, "request-id", "", "Optional idempotency request ID")
	}
}

func (grp *networkConnectivityCmds) addNCCreateConfigFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&grp.flagNCConfigFile, "config-file", "", "Path to a JSON/YAML file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
}

func (grp *networkConnectivityCmds) addNCUpdateConfigFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&grp.flagNCConfigFile, "config-file", "", "Path to a JSON/YAML file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
		c.Flags().StringVar(&grp.flagNCUpdateMask, "update-mask", "", "Comma-separated list of fields to update (defaults to every populated field)")
	}
}

// --- operations (#906) ---

func (grp *networkConnectivityCmds) ncResolveOperationName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := grp.ncProjectParent()
	if err != nil {
		return "", err
	}
	return ncChild(parent, "operations", id), nil
}

func (grp *networkConnectivityCmds) runNCOpCancel(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	name, err := grp.ncResolveOperationName(args[0])
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &networkconnectivity.GoogleLongrunningCancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Cancel request issued for operation %s.\n", args[0])
	return nil
}

func (grp *networkConnectivityCmds) runNCOpDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	name, err := grp.ncResolveOperationName(args[0])
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Deleted operation %s.\n", args[0])
	return nil
}

func (grp *networkConnectivityCmds) runNCOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	name, err := grp.ncResolveOperationName(args[0])
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, op, grp.flagNCFormat)
}

func (grp *networkConnectivityCmds) runNCOpList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	parent, err := grp.ncProjectParent()
	if err != nil {
		return err
	}
	var all []*networkconnectivity.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if grp.flagNCFilter != "" {
			call = call.Filter(grp.flagNCFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagNCFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagNCFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-60s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(grp.inv.stdout, "%-60s %v\n", ncBasename(o.Name), o.Done)
	}
	return nil
}

// --- locations (#903) ---

func (grp *networkConnectivityCmds) runNCLocDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	name := args[0]
	if !strings.HasPrefix(name, "projects/") {
		name = fmt.Sprintf("projects/%s/locations/%s", project, name)
	}
	loc, err := svc.Projects.Locations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, loc, grp.flagNCFormat)
}

func (grp *networkConnectivityCmds) runNCLocList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := grp.inv.clients().NetworkConnectivity(ctx)
	if err != nil {
		return err
	}
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s", project)
	var all []*networkconnectivity.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(parent).Context(ctx)
		if grp.flagNCFilter != "" {
			call = call.Filter(grp.flagNCFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagNCFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagNCFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-30s %s\n", "NAME", "LOCATION_ID")
	for _, l := range all {
		fmt.Fprintf(grp.inv.stdout, "%-30s %s\n", ncBasename(l.Name), l.LocationId)
	}
	return nil
}

func (grp *networkConnectivityCmds) initNetworkConnectivity() {
	// operations
	for _, c := range []*cobra.Command{grp.ncOpCancelCmd, grp.ncOpDeleteCmd, grp.ncOpDescribeCmd, grp.ncOpListCmd} {
		c.Flags().StringVar(&grp.flagNCLocation, "location", "", "Location (required)")
	}
	grp.addNCFormatFlag(grp.ncOpDescribeCmd, grp.ncOpListCmd)
	grp.addNCFilterFlag(grp.ncOpListCmd)
	grp.networkConnectivityOperationsCmd.AddCommand(grp.ncOpCancelCmd, grp.ncOpDeleteCmd, grp.ncOpDescribeCmd, grp.ncOpListCmd)
	grp.networkConnectivityCmd.AddCommand(grp.networkConnectivityOperationsCmd)

	// locations
	grp.addNCFormatFlag(grp.ncLocDescribeCmd, grp.ncLocListCmd)
	grp.addNCFilterFlag(grp.ncLocListCmd)
	grp.networkConnectivityLocationsCmd.AddCommand(grp.ncLocDescribeCmd, grp.ncLocListCmd)
	grp.networkConnectivityCmd.AddCommand(grp.networkConnectivityLocationsCmd)

	grp.registerNCHubs(grp.networkConnectivityCmd)
	grp.registerNCSpokes(grp.networkConnectivityCmd)
	grp.registerNCInternalRanges(grp.networkConnectivityCmd)
	grp.registerNCPolicyBasedRoutes(grp.networkConnectivityCmd)
	grp.registerNCRegionalEndpoints(grp.networkConnectivityCmd)
	grp.registerNCServiceConnectionPolicies(grp.networkConnectivityCmd)
	grp.registerNCMulticloudDataTransferConfigs(grp.networkConnectivityCmd)
	grp.registerNCMulticloudDataTransferSupportedServices(grp.networkConnectivityCmd)
	grp.registerNCTransports(grp.networkConnectivityCmd)

}

// networkConnectivityCmds holds the commands and flag values of one invocation's "networkConnectivity"
// command group, so invocations don't share state (#2070).
type networkConnectivityCmds struct {
	inv *invocation

	networkConnectivityCmd           *cobra.Command
	flagNCLocation                   string
	flagNCFormat                     string
	flagNCFilter                     string
	flagNCConfigFile                 string
	flagNCUpdateMask                 string
	flagNCAsync                      bool
	flagNCRequestID                  string
	flagNCHub                        string
	flagNCSpokeURI                   string
	flagNCDetails                    string
	networkConnectivityOperationsCmd *cobra.Command
	ncOpCancelCmd                    *cobra.Command
	ncOpDeleteCmd                    *cobra.Command
	ncOpDescribeCmd                  *cobra.Command
	ncOpListCmd                      *cobra.Command
	networkConnectivityLocationsCmd  *cobra.Command
	ncLocDescribeCmd                 *cobra.Command
	ncLocListCmd                     *cobra.Command
}

// newNetworkConnectivityCmd returns the "networkConnectivity" command group for inv.
func newNetworkConnectivityCmd(inv *invocation) *cobra.Command {
	return newNetworkConnectivityCmds(inv).networkConnectivityCmd
}

// newNetworkConnectivityCmds builds the "networkConnectivity" command group for inv and returns its state.
func newNetworkConnectivityCmds(inv *invocation) *networkConnectivityCmds {
	grp := &networkConnectivityCmds{inv: inv}
	grp.networkConnectivityCmd = &cobra.Command{Use: "network-connectivity", Short: "Manage Network Connectivity"}
	grp.networkConnectivityOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Network Connectivity operations"}
	grp.ncOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: grp.runNCOpCancel,
	}
	grp.ncOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: grp.runNCOpDelete,
	}
	grp.ncOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: grp.runNCOpDescribe,
	}
	grp.ncOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: grp.runNCOpList,
	}
	grp.networkConnectivityLocationsCmd = &cobra.Command{Use: "locations", Short: "Get Network Connectivity locations"}
	grp.ncLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a location",
		Args: cobra.ExactArgs(1), RunE: grp.runNCLocDescribe,
	}
	grp.ncLocListCmd = &cobra.Command{
		Use: "list", Short: "List Network Connectivity locations",
		Args: cobra.NoArgs, RunE: grp.runNCLocList,
	}
	grp.initNetworkConnectivity()
	return grp
}
