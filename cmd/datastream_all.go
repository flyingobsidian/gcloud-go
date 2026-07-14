package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datastream "google.golang.org/api/datastream/v1"
)

func dsLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func dsChildName(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func dsWaitOp(ctx context.Context, svc *datastream.Service, op *datastream.Operation) (*datastream.Operation, error) {
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

func dsFinishOp(ctx context.Context, svc *datastream.Service, op *datastream.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := dsWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// --- locations ---

var dsLocationsCmd = &cobra.Command{Use: "locations", Short: "Explore Datastream locations"}

var (
	dsLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Datastream location",
		Args: cobra.ExactArgs(1), RunE: runDSLocDescribe,
	}
	dsLocListCmd = &cobra.Command{
		Use: "list", Short: "List Datastream locations for the project",
		Args: cobra.NoArgs, RunE: runDSLocList,
	}
)

var flagDSFormat string

func init() {
	dsLocDescribeCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsLocListCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsLocationsCmd.AddCommand(dsLocDescribeCmd, dsLocListCmd)
	datastreamCmd.AddCommand(dsLocationsCmd)
}

func runDSLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(dsLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(loc, flagDSFormat)
}

func runDSLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if flagDSFormat != "" {
		return emitFormatted(resp.Locations, flagDSFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- operations ---

var dsOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Datastream operations"}

var (
	dsOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Datastream operation",
		Args: cobra.ExactArgs(1), RunE: runDSOpCancel,
	}
	dsOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Datastream operation",
		Args: cobra.ExactArgs(1), RunE: runDSOpDelete,
	}
	dsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Datastream operation",
		Args: cobra.ExactArgs(1), RunE: runDSOpDescribe,
	}
	dsOpListCmd = &cobra.Command{
		Use: "list", Short: "List Datastream operations in a location",
		Args: cobra.NoArgs, RunE: runDSOpList,
	}
)

var flagDSOpLocation string

func init() {
	for _, c := range []*cobra.Command{dsOpCancelCmd, dsOpDeleteCmd, dsOpDescribeCmd, dsOpListCmd} {
		c.Flags().StringVar(&flagDSOpLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	dsOpDescribeCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsOpListCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsOperationsCmd.AddCommand(dsOpCancelCmd, dsOpDeleteCmd, dsOpDescribeCmd, dsOpListCmd)
	datastreamCmd.AddCommand(dsOperationsCmd)
}

func dsOpName(id, project, location string) string {
	return dsChildName("operations", id, dsLocationParent(project, location))
}

func runDSOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(dsOpName(args[0], project, flagDSOpLocation), &datastream.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runDSOpDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(dsOpName(args[0], project, flagDSOpLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runDSOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(dsOpName(args[0], project, flagDSOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagDSFormat)
}

func runDSOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(dsLocationParent(project, flagDSOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagDSFormat != "" {
		return emitFormatted(resp.Operations, flagDSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- connection-profiles ---

var dsConnProfilesCmd = &cobra.Command{Use: "connection-profiles", Short: "Manage Datastream connection profiles"}

var (
	dsCPCreateCmd = &cobra.Command{
		Use: "create PROFILE", Short: "Create a connection profile from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDSCPCreate,
	}
	dsCPDeleteCmd = &cobra.Command{
		Use: "delete PROFILE", Short: "Delete a connection profile",
		Args: cobra.ExactArgs(1), RunE: runDSCPDelete,
	}
	dsCPDescribeCmd = &cobra.Command{
		Use: "describe PROFILE", Short: "Describe a connection profile",
		Args: cobra.ExactArgs(1), RunE: runDSCPDescribe,
	}
	dsCPDiscoverCmd = &cobra.Command{
		Use: "discover PROFILE", Short: "Discover the schema of a connection profile from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDSCPDiscover,
	}
	dsCPListCmd = &cobra.Command{
		Use: "list", Short: "List connection profiles in a location",
		Args: cobra.NoArgs, RunE: runDSCPList,
	}
	dsCPUpdateCmd = &cobra.Command{
		Use: "update PROFILE", Short: "Update a connection profile from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDSCPUpdate,
	}
)

var (
	flagDSCPLocation   string
	flagDSCPConfigFile string
	flagDSCPUpdateMask string
	flagDSCPAsync      bool
)

func init() {
	all := []*cobra.Command{dsCPCreateCmd, dsCPDeleteCmd, dsCPDescribeCmd, dsCPDiscoverCmd, dsCPListCmd, dsCPUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDSCPLocation, "location", "", "Location containing the connection profile (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{dsCPCreateCmd, dsCPUpdateCmd, dsCPDiscoverCmd} {
		c.Flags().StringVar(&flagDSCPConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the ConnectionProfile / DiscoverConnectionProfileRequest body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dsCPUpdateCmd.Flags().StringVar(&flagDSCPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{dsCPCreateCmd, dsCPDeleteCmd, dsCPUpdateCmd} {
		c.Flags().BoolVar(&flagDSCPAsync, "async", false, "Return the long-running operation without waiting")
	}
	dsCPDescribeCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsCPListCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")

	dsConnProfilesCmd.AddCommand(all...)
	datastreamCmd.AddCommand(dsConnProfilesCmd)
}

func dsCPName(id, project, location string) string {
	return dsChildName("connectionProfiles", id, dsLocationParent(project, location))
}

func runDSCPCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cp := &datastream.ConnectionProfile{}
	if err := loadYAMLOrJSONInto(flagDSCPConfigFile, cp); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectionProfiles.Create(dsLocationParent(project, flagDSCPLocation), cp).
		ConnectionProfileId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connection profile: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Create connection profile", args[0], flagDSCPAsync)
}

func runDSCPDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectionProfiles.Delete(dsCPName(args[0], project, flagDSCPLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connection profile: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Delete connection profile", args[0], flagDSCPAsync)
}

func runDSCPDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectionProfiles.Get(dsCPName(args[0], project, flagDSCPLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connection profile: %w", err)
	}
	return emitFormatted(got, flagDSFormat)
}

func runDSCPDiscover(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &datastream.DiscoverConnectionProfileRequest{}
	if err := loadYAMLOrJSONInto(flagDSCPConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectionProfiles.Discover(dsLocationParent(project, flagDSCPLocation), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("discovering connection profile: %w", err)
	}
	return emitFormatted(got, "")
}

func runDSCPList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.ConnectionProfiles.List(dsLocationParent(project, flagDSCPLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing connection profiles: %w", err)
	}
	if flagDSFormat != "" {
		return emitFormatted(resp.ConnectionProfiles, flagDSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, p := range resp.ConnectionProfiles {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.DisplayName)
	}
	return nil
}

func runDSCPUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cp := &datastream.ConnectionProfile{}
	if err := loadYAMLOrJSONInto(flagDSCPConfigFile, cp); err != nil {
		return err
	}
	mask := flagDSCPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cp))
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectionProfiles.Patch(dsCPName(args[0], project, flagDSCPLocation), cp).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connection profile: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Update connection profile", args[0], flagDSCPAsync)
}

// --- private-connections ---

var dsPrivateConnsCmd = &cobra.Command{Use: "private-connections", Short: "Manage Datastream private connections"}

var (
	dsPCCreateCmd = &cobra.Command{
		Use: "create PRIVATE_CONNECTION", Short: "Create a private connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDSPCCreate,
	}
	dsPCDeleteCmd = &cobra.Command{
		Use: "delete PRIVATE_CONNECTION", Short: "Delete a private connection",
		Args: cobra.ExactArgs(1), RunE: runDSPCDelete,
	}
	dsPCDescribeCmd = &cobra.Command{
		Use: "describe PRIVATE_CONNECTION", Short: "Describe a private connection",
		Args: cobra.ExactArgs(1), RunE: runDSPCDescribe,
	}
	dsPCListCmd = &cobra.Command{
		Use: "list", Short: "List private connections in a location",
		Args: cobra.NoArgs, RunE: runDSPCList,
	}
)

var (
	flagDSPCLocation   string
	flagDSPCConfigFile string
	flagDSPCAsync      bool
)

func init() {
	all := []*cobra.Command{dsPCCreateCmd, dsPCDeleteCmd, dsPCDescribeCmd, dsPCListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDSPCLocation, "location", "", "Location containing the private connection (required)")
		_ = c.MarkFlagRequired("location")
	}
	dsPCCreateCmd.Flags().StringVar(&flagDSPCConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the PrivateConnection body (required)")
	_ = dsPCCreateCmd.MarkFlagRequired("config-file")
	for _, c := range []*cobra.Command{dsPCCreateCmd, dsPCDeleteCmd} {
		c.Flags().BoolVar(&flagDSPCAsync, "async", false, "Return the long-running operation without waiting")
	}
	dsPCDescribeCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsPCListCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")

	dsPrivateConnsCmd.AddCommand(all...)
	datastreamCmd.AddCommand(dsPrivateConnsCmd)
}

func dsPCName(id, project, location string) string {
	return dsChildName("privateConnections", id, dsLocationParent(project, location))
}

func runDSPCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	pc := &datastream.PrivateConnection{}
	if err := loadYAMLOrJSONInto(flagDSPCConfigFile, pc); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Create(dsLocationParent(project, flagDSPCLocation), pc).
		PrivateConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating private connection: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Create private connection", args[0], flagDSPCAsync)
}

func runDSPCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Delete(dsPCName(args[0], project, flagDSPCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting private connection: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Delete private connection", args[0], flagDSPCAsync)
}

func runDSPCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PrivateConnections.Get(dsPCName(args[0], project, flagDSPCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing private connection: %w", err)
	}
	return emitFormatted(got, flagDSFormat)
}

func runDSPCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.PrivateConnections.List(dsLocationParent(project, flagDSPCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing private connections: %w", err)
	}
	if flagDSFormat != "" {
		return emitFormatted(resp.PrivateConnections, flagDSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, pc := range resp.PrivateConnections {
		fmt.Printf("%-40s %s\n", path.Base(pc.Name), pc.State)
	}
	return nil
}

// --- routes (under private-connections) ---

var dsRoutesCmd = &cobra.Command{Use: "routes", Short: "Manage Datastream routes (within a private connection)"}

var (
	dsRouteCreateCmd = &cobra.Command{
		Use: "create ROUTE", Short: "Create a route from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDSRouteCreate,
	}
	dsRouteDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete a route",
		Args: cobra.ExactArgs(1), RunE: runDSRouteDelete,
	}
	dsRouteDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe a route",
		Args: cobra.ExactArgs(1), RunE: runDSRouteDescribe,
	}
	dsRouteListCmd = &cobra.Command{
		Use: "list", Short: "List routes for a private connection",
		Args: cobra.NoArgs, RunE: runDSRouteList,
	}
)

var (
	flagDSRouteLocation   string
	flagDSRoutePC         string
	flagDSRouteConfigFile string
	flagDSRouteAsync      bool
)

func init() {
	all := []*cobra.Command{dsRouteCreateCmd, dsRouteDeleteCmd, dsRouteDescribeCmd, dsRouteListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDSRouteLocation, "location", "", "Location containing the private connection (required)")
		c.Flags().StringVar(&flagDSRoutePC, "private-connection", "", "Private connection containing the route (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("private-connection")
	}
	dsRouteCreateCmd.Flags().StringVar(&flagDSRouteConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Route message body (required)")
	_ = dsRouteCreateCmd.MarkFlagRequired("config-file")
	for _, c := range []*cobra.Command{dsRouteCreateCmd, dsRouteDeleteCmd} {
		c.Flags().BoolVar(&flagDSRouteAsync, "async", false, "Return the long-running operation without waiting")
	}
	dsRouteDescribeCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsRouteListCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")

	dsRoutesCmd.AddCommand(all...)
	datastreamCmd.AddCommand(dsRoutesCmd)
}

func dsRouteParent(project, location, pc string) string {
	return fmt.Sprintf("%s/privateConnections/%s", dsLocationParent(project, location), pc)
}

func dsRouteName(id, project, location, pc string) string {
	return dsChildName("routes", id, dsRouteParent(project, location, pc))
}

func runDSRouteCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	route := &datastream.Route{}
	if err := loadYAMLOrJSONInto(flagDSRouteConfigFile, route); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Routes.Create(dsRouteParent(project, flagDSRouteLocation, flagDSRoutePC), route).
		RouteId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating route: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Create route", args[0], flagDSRouteAsync)
}

func runDSRouteDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Routes.Delete(dsRouteName(args[0], project, flagDSRouteLocation, flagDSRoutePC)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting route: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Delete route", args[0], flagDSRouteAsync)
}

func runDSRouteDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PrivateConnections.Routes.Get(dsRouteName(args[0], project, flagDSRouteLocation, flagDSRoutePC)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing route: %w", err)
	}
	return emitFormatted(got, flagDSFormat)
}

func runDSRouteList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.PrivateConnections.Routes.List(dsRouteParent(project, flagDSRouteLocation, flagDSRoutePC)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing routes: %w", err)
	}
	if flagDSFormat != "" {
		return emitFormatted(resp.Routes, flagDSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DESTINATION_ADDRESS")
	for _, r := range resp.Routes {
		fmt.Printf("%-40s %s\n", path.Base(r.Name), r.DestinationAddress)
	}
	return nil
}

// --- streams ---

var dsStreamsCmd = &cobra.Command{Use: "streams", Short: "Manage Datastream streams"}

var (
	dsStreamCreateCmd = &cobra.Command{
		Use: "create STREAM", Short: "Create a stream from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDSStreamCreate,
	}
	dsStreamDeleteCmd = &cobra.Command{
		Use: "delete STREAM", Short: "Delete a stream",
		Args: cobra.ExactArgs(1), RunE: runDSStreamDelete,
	}
	dsStreamDescribeCmd = &cobra.Command{
		Use: "describe STREAM", Short: "Describe a stream",
		Args: cobra.ExactArgs(1), RunE: runDSStreamDescribe,
	}
	dsStreamListCmd = &cobra.Command{
		Use: "list", Short: "List streams in a location",
		Args: cobra.NoArgs, RunE: runDSStreamList,
	}
	dsStreamUpdateCmd = &cobra.Command{
		Use: "update STREAM", Short: "Update a stream from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDSStreamUpdate,
	}
)

var (
	flagDSStreamLocation   string
	flagDSStreamConfigFile string
	flagDSStreamUpdateMask string
	flagDSStreamAsync      bool
)

func init() {
	all := []*cobra.Command{dsStreamCreateCmd, dsStreamDeleteCmd, dsStreamDescribeCmd, dsStreamListCmd, dsStreamUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDSStreamLocation, "location", "", "Location containing the stream (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{dsStreamCreateCmd, dsStreamUpdateCmd} {
		c.Flags().StringVar(&flagDSStreamConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Stream message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dsStreamUpdateCmd.Flags().StringVar(&flagDSStreamUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{dsStreamCreateCmd, dsStreamDeleteCmd, dsStreamUpdateCmd} {
		c.Flags().BoolVar(&flagDSStreamAsync, "async", false, "Return the long-running operation without waiting")
	}
	dsStreamDescribeCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsStreamListCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")

	dsStreamsCmd.AddCommand(all...)
	datastreamCmd.AddCommand(dsStreamsCmd)
}

func dsStreamName(id, project, location string) string {
	return dsChildName("streams", id, dsLocationParent(project, location))
}

func runDSStreamCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &datastream.Stream{}
	if err := loadYAMLOrJSONInto(flagDSStreamConfigFile, s); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Streams.Create(dsLocationParent(project, flagDSStreamLocation), s).
		StreamId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating stream: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Create stream", args[0], flagDSStreamAsync)
}

func runDSStreamDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Streams.Delete(dsStreamName(args[0], project, flagDSStreamLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting stream: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Delete stream", args[0], flagDSStreamAsync)
}

func runDSStreamDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Get(dsStreamName(args[0], project, flagDSStreamLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing stream: %w", err)
	}
	return emitFormatted(got, flagDSFormat)
}

func runDSStreamList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Streams.List(dsLocationParent(project, flagDSStreamLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing streams: %w", err)
	}
	if flagDSFormat != "" {
		return emitFormatted(resp.Streams, flagDSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, s := range resp.Streams {
		fmt.Printf("%-40s %s\n", path.Base(s.Name), s.State)
	}
	return nil
}

func runDSStreamUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &datastream.Stream{}
	if err := loadYAMLOrJSONInto(flagDSStreamConfigFile, s); err != nil {
		return err
	}
	mask := flagDSStreamUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Streams.Patch(dsStreamName(args[0], project, flagDSStreamLocation), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating stream: %w", err)
	}
	return dsFinishOp(ctx, svc, op, "Update stream", args[0], flagDSStreamAsync)
}

// --- objects ---

var dsObjectsCmd = &cobra.Command{Use: "objects", Short: "Manage Datastream stream objects"}

var (
	dsObjDescribeCmd = &cobra.Command{
		Use: "describe OBJECT", Short: "Describe a stream object",
		Args: cobra.ExactArgs(1), RunE: runDSObjDescribe,
	}
	dsObjListCmd = &cobra.Command{
		Use: "list", Short: "List stream objects for a stream",
		Args: cobra.NoArgs, RunE: runDSObjList,
	}
	dsObjLookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up a stream object by source identifier (--config-file)",
		Args: cobra.NoArgs, RunE: runDSObjLookup,
	}
	dsObjStartBackfillCmd = &cobra.Command{
		Use: "start-backfill OBJECT", Short: "Start a backfill for a stream object",
		Args: cobra.ExactArgs(1), RunE: runDSObjStartBackfill,
	}
	dsObjStopBackfillCmd = &cobra.Command{
		Use: "stop-backfill OBJECT", Short: "Stop a backfill for a stream object",
		Args: cobra.ExactArgs(1), RunE: runDSObjStopBackfill,
	}
)

var (
	flagDSObjLocation   string
	flagDSObjStream     string
	flagDSObjConfigFile string
)

func init() {
	all := []*cobra.Command{dsObjDescribeCmd, dsObjListCmd, dsObjLookupCmd, dsObjStartBackfillCmd, dsObjStopBackfillCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDSObjLocation, "location", "", "Location containing the stream (required)")
		c.Flags().StringVar(&flagDSObjStream, "stream", "", "Stream containing the object (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("stream")
	}
	dsObjLookupCmd.Flags().StringVar(&flagDSObjConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the LookupStreamObjectRequest body (required)")
	_ = dsObjLookupCmd.MarkFlagRequired("config-file")
	dsObjDescribeCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")
	dsObjListCmd.Flags().StringVar(&flagDSFormat, "format", "", "Output format")

	dsObjectsCmd.AddCommand(all...)
	datastreamCmd.AddCommand(dsObjectsCmd)
}

func dsObjParent(project, location, stream string) string {
	return fmt.Sprintf("%s/streams/%s", dsLocationParent(project, location), stream)
}

func dsObjName(id, project, location, stream string) string {
	return dsChildName("objects", id, dsObjParent(project, location, stream))
}

func runDSObjDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Objects.Get(dsObjName(args[0], project, flagDSObjLocation, flagDSObjStream)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing stream object: %w", err)
	}
	return emitFormatted(got, flagDSFormat)
}

func runDSObjList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Streams.Objects.List(dsObjParent(project, flagDSObjLocation, flagDSObjStream)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing stream objects: %w", err)
	}
	if flagDSFormat != "" {
		return emitFormatted(resp.StreamObjects, flagDSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, o := range resp.StreamObjects {
		fmt.Printf("%-40s %s\n", path.Base(o.Name), o.DisplayName)
	}
	return nil
}

func runDSObjLookup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &datastream.LookupStreamObjectRequest{}
	if err := loadYAMLOrJSONInto(flagDSObjConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Objects.Lookup(dsObjParent(project, flagDSObjLocation, flagDSObjStream), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up stream object: %w", err)
	}
	return emitFormatted(got, flagDSFormat)
}

func runDSObjStartBackfill(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Objects.StartBackfillJob(dsObjName(args[0], project, flagDSObjLocation, flagDSObjStream), &datastream.StartBackfillJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting backfill: %w", err)
	}
	return emitFormatted(got, "")
}

func runDSObjStopBackfill(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DatastreamService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Objects.StopBackfillJob(dsObjName(args[0], project, flagDSObjLocation, flagDSObjStream), &datastream.StopBackfillJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping backfill: %w", err)
	}
	return emitFormatted(got, "")
}
