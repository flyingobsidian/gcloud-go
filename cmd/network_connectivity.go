package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkconnectivity "google.golang.org/api/networkconnectivity/v1"
)

// --- gcloud network-connectivity (#901-#911) ---
//
// Every subgroup here is backed by the google.golang.org/api/networkconnectivity
// service (v1). Layout mirrors cmd/network_security.go:
//
//   - a small builder for the shared --location / --config-file / --async / --format
//     flags,
//   - a generic ncCRUD binding that turns a resource-specific closure set into
//     the standard create/delete/describe/list/update subcommands,
//   - LRO create/delete/patch calls block by default and wait via ncWaitOp.

var networkConnectivityCmd = &cobra.Command{Use: "network-connectivity", Short: "Manage Network Connectivity"}

// shared network-connectivity flags
var (
	flagNCLocation   string
	flagNCFormat     string
	flagNCFilter     string
	flagNCConfigFile string
	flagNCUpdateMask string
	flagNCAsync      bool
	flagNCRequestID  string
	flagNCHub        string
	flagNCSpokeURI   string
	flagNCDetails    string
)

// ncProjectParent returns projects/PROJECT/locations/LOCATION for the current
// project and --location.
func ncProjectParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagNCLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagNCLocation), nil
}

// ncGlobalParent returns projects/PROJECT/locations/global.
func ncGlobalParent() (string, error) {
	project, err := resolveProject()
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

func ncResolveMask(body any) string {
	if flagNCUpdateMask != "" {
		return flagNCUpdateMask
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

func ncFinishOp(ctx context.Context, svc *networkconnectivity.Service, op *networkconnectivity.GoogleLongrunningOperation, verb, name string) error {
	if flagNCAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := ncWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

func addNCLocationFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNCLocation, "location", "", "Location (required)")
	}
}

func addNCGlobalLocationFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNCLocation, "location", "global", "Location (defaults to global)")
	}
}

func addNCFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNCFormat, "format", "", "Output format")
	}
}

func addNCFilterFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNCFilter, "filter", "", "Server-side list filter")
	}
}

func addNCAsyncFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().BoolVar(&flagNCAsync, "async", false, "Do not wait for the operation to complete")
	}
}

func addNCRequestIDFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNCRequestID, "request-id", "", "Optional idempotency request ID")
	}
}

func addNCCreateConfigFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNCConfigFile, "config-file", "", "Path to a JSON/YAML file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
}

func addNCUpdateConfigFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagNCConfigFile, "config-file", "", "Path to a JSON/YAML file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
		c.Flags().StringVar(&flagNCUpdateMask, "update-mask", "", "Comma-separated list of fields to update (defaults to every populated field)")
	}
}

// --- operations (#906) ---

var networkConnectivityOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Network Connectivity operations"}

var (
	ncOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: runNCOpCancel,
	}
	ncOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: runNCOpDelete,
	}
	ncOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runNCOpDescribe,
	}
	ncOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: runNCOpList,
	}
)

func ncResolveOperationName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := ncProjectParent()
	if err != nil {
		return "", err
	}
	return ncChild(parent, "operations", id), nil
}

func runNCOpCancel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := ncResolveOperationName(args[0])
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &networkconnectivity.GoogleLongrunningCancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Cancel request issued for operation %s.\n", args[0])
	return nil
}

func runNCOpDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := ncResolveOperationName(args[0])
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Deleted operation %s.\n", args[0])
	return nil
}

func runNCOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := ncResolveOperationName(args[0])
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagNCFormat)
}

func runNCOpList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent, err := ncProjectParent()
	if err != nil {
		return err
	}
	var all []*networkconnectivity.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagNCFilter != "" {
			call = call.Filter(flagNCFilter)
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
	if flagNCFormat != "" {
		return emitFormatted(all, flagNCFormat)
	}
	fmt.Printf("%-60s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-60s %v\n", ncBasename(o.Name), o.Done)
	}
	return nil
}

// --- locations (#903) ---

var networkConnectivityLocationsCmd = &cobra.Command{Use: "locations", Short: "Get Network Connectivity locations"}

var (
	ncLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a location",
		Args: cobra.ExactArgs(1), RunE: runNCLocDescribe,
	}
	ncLocListCmd = &cobra.Command{
		Use: "list", Short: "List Network Connectivity locations",
		Args: cobra.NoArgs, RunE: runNCLocList,
	}
)

func runNCLocDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	project, err := resolveProject()
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
	return emitFormatted(loc, flagNCFormat)
}

func runNCLocList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s", project)
	var all []*networkconnectivity.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(parent).Context(ctx)
		if flagNCFilter != "" {
			call = call.Filter(flagNCFilter)
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
	if flagNCFormat != "" {
		return emitFormatted(all, flagNCFormat)
	}
	fmt.Printf("%-30s %s\n", "NAME", "LOCATION_ID")
	for _, l := range all {
		fmt.Printf("%-30s %s\n", ncBasename(l.Name), l.LocationId)
	}
	return nil
}

func init() {
	// operations
	for _, c := range []*cobra.Command{ncOpCancelCmd, ncOpDeleteCmd, ncOpDescribeCmd, ncOpListCmd} {
		c.Flags().StringVar(&flagNCLocation, "location", "", "Location (required)")
	}
	addNCFormatFlag(ncOpDescribeCmd, ncOpListCmd)
	addNCFilterFlag(ncOpListCmd)
	networkConnectivityOperationsCmd.AddCommand(ncOpCancelCmd, ncOpDeleteCmd, ncOpDescribeCmd, ncOpListCmd)
	networkConnectivityCmd.AddCommand(networkConnectivityOperationsCmd)

	// locations
	addNCFormatFlag(ncLocDescribeCmd, ncLocListCmd)
	addNCFilterFlag(ncLocListCmd)
	networkConnectivityLocationsCmd.AddCommand(ncLocDescribeCmd, ncLocListCmd)
	networkConnectivityCmd.AddCommand(networkConnectivityLocationsCmd)

	registerNCHubs(networkConnectivityCmd)
	registerNCSpokes(networkConnectivityCmd)
	registerNCInternalRanges(networkConnectivityCmd)
	registerNCPolicyBasedRoutes(networkConnectivityCmd)
	registerNCRegionalEndpoints(networkConnectivityCmd)
	registerNCServiceConnectionPolicies(networkConnectivityCmd)
	registerNCMulticloudDataTransferConfigs(networkConnectivityCmd)
	registerNCMulticloudDataTransferSupportedServices(networkConnectivityCmd)
	registerNCTransports(networkConnectivityCmd)

	rootCmd.AddCommand(networkConnectivityCmd)
}
