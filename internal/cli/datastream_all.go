package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

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

func (g *datastreamCmds) dsFinishOp(ctx context.Context, svc *datastream.Service, op *datastream.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(g.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(g.inv.stdout, op, "")
	}
	final, err := dsWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(g.inv.stdout, final.Response, "")
	}
	return nil
}

// --- locations ---

func (g *datastreamCmds) initDatastreamAll() {
	g.dsLocDescribeCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsLocListCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsLocationsCmd.AddCommand(g.dsLocDescribeCmd, g.dsLocListCmd)
	g.datastreamCmd.AddCommand(g.dsLocationsCmd)
}

func (g *datastreamCmds) runDSLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(dsLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, loc, g.flagDSFormat)
}

func (g *datastreamCmds) runDSLocList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if g.flagDSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Locations, g.flagDSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Locations), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(g.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- operations ---

func (g *datastreamCmds) initDatastreamAll2() {
	for _, c := range []*cobra.Command{g.dsOpCancelCmd, g.dsOpDeleteCmd, g.dsOpDescribeCmd, g.dsOpListCmd} {
		c.Flags().StringVar(&g.flagDSOpLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	g.dsOpDescribeCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsOpListCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsOperationsCmd.AddCommand(g.dsOpCancelCmd, g.dsOpDeleteCmd, g.dsOpDescribeCmd, g.dsOpListCmd)
	g.datastreamCmd.AddCommand(g.dsOperationsCmd)
}

func dsOpName(id, project, location string) string {
	return dsChildName("operations", id, dsLocationParent(project, location))
}

func (g *datastreamCmds) runDSOpCancel(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(dsOpName(args[0], project, g.flagDSOpLocation), &datastream.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (g *datastreamCmds) runDSOpDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(dsOpName(args[0], project, g.flagDSOpLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (g *datastreamCmds) runDSOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(dsOpName(args[0], project, g.flagDSOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagDSFormat)
}

func (g *datastreamCmds) runDSOpList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(dsLocationParent(project, g.flagDSOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if g.flagDSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Operations, g.flagDSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(g.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- connection-profiles ---

func (g *datastreamCmds) initDatastreamAll3() {
	all := []*cobra.Command{g.dsCPCreateCmd, g.dsCPDeleteCmd, g.dsCPDescribeCmd, g.dsCPDiscoverCmd, g.dsCPListCmd, g.dsCPUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagDSCPLocation, "location", "", "Location containing the connection profile (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.dsCPCreateCmd, g.dsCPUpdateCmd, g.dsCPDiscoverCmd} {
		c.Flags().StringVar(&g.flagDSCPConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the ConnectionProfile / DiscoverConnectionProfileRequest body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.dsCPUpdateCmd.Flags().StringVar(&g.flagDSCPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.dsCPCreateCmd, g.dsCPDeleteCmd, g.dsCPUpdateCmd} {
		c.Flags().BoolVar(&g.flagDSCPAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.dsCPDescribeCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsCPListCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")

	g.dsConnProfilesCmd.AddCommand(all...)
	g.datastreamCmd.AddCommand(g.dsConnProfilesCmd)
}

func dsCPName(id, project, location string) string {
	return dsChildName("connectionProfiles", id, dsLocationParent(project, location))
}

func (g *datastreamCmds) runDSCPCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	cp := &datastream.ConnectionProfile{}
	if err := loadYAMLOrJSONInto(g.flagDSCPConfigFile, cp); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectionProfiles.Create(dsLocationParent(project, g.flagDSCPLocation), cp).
		ConnectionProfileId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connection profile: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Create connection profile", args[0], g.flagDSCPAsync)
}

func (g *datastreamCmds) runDSCPDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectionProfiles.Delete(dsCPName(args[0], project, g.flagDSCPLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connection profile: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Delete connection profile", args[0], g.flagDSCPAsync)
}

func (g *datastreamCmds) runDSCPDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectionProfiles.Get(dsCPName(args[0], project, g.flagDSCPLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connection profile: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagDSFormat)
}

func (g *datastreamCmds) runDSCPDiscover(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	req := &datastream.DiscoverConnectionProfileRequest{}
	if err := loadYAMLOrJSONInto(g.flagDSCPConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectionProfiles.Discover(dsLocationParent(project, g.flagDSCPLocation), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("discovering connection profile: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *datastreamCmds) runDSCPList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.ConnectionProfiles.List(dsLocationParent(project, g.flagDSCPLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing connection profiles: %w", err)
	}
	if g.flagDSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.ConnectionProfiles, g.flagDSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.ConnectionProfiles), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, p := range resp.ConnectionProfiles {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.DisplayName)
	}
	return nil
}

func (g *datastreamCmds) runDSCPUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	cp := &datastream.ConnectionProfile{}
	if err := loadYAMLOrJSONInto(g.flagDSCPConfigFile, cp); err != nil {
		return err
	}
	mask := g.flagDSCPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cp))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectionProfiles.Patch(dsCPName(args[0], project, g.flagDSCPLocation), cp).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connection profile: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Update connection profile", args[0], g.flagDSCPAsync)
}

// --- private-connections ---

func (g *datastreamCmds) initDatastreamAll4() {
	all := []*cobra.Command{g.dsPCCreateCmd, g.dsPCDeleteCmd, g.dsPCDescribeCmd, g.dsPCListCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagDSPCLocation, "location", "", "Location containing the private connection (required)")
		_ = c.MarkFlagRequired("location")
	}
	g.dsPCCreateCmd.Flags().StringVar(&g.flagDSPCConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the PrivateConnection body (required)")
	_ = g.dsPCCreateCmd.MarkFlagRequired("config-file")
	for _, c := range []*cobra.Command{g.dsPCCreateCmd, g.dsPCDeleteCmd} {
		c.Flags().BoolVar(&g.flagDSPCAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.dsPCDescribeCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsPCListCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")

	g.dsPrivateConnsCmd.AddCommand(all...)
	g.datastreamCmd.AddCommand(g.dsPrivateConnsCmd)
}

func dsPCName(id, project, location string) string {
	return dsChildName("privateConnections", id, dsLocationParent(project, location))
}

func (g *datastreamCmds) runDSPCCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	pc := &datastream.PrivateConnection{}
	if err := loadYAMLOrJSONInto(g.flagDSPCConfigFile, pc); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Create(dsLocationParent(project, g.flagDSPCLocation), pc).
		PrivateConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating private connection: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Create private connection", args[0], g.flagDSPCAsync)
}

func (g *datastreamCmds) runDSPCDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Delete(dsPCName(args[0], project, g.flagDSPCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting private connection: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Delete private connection", args[0], g.flagDSPCAsync)
}

func (g *datastreamCmds) runDSPCDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PrivateConnections.Get(dsPCName(args[0], project, g.flagDSPCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing private connection: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagDSFormat)
}

func (g *datastreamCmds) runDSPCList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.PrivateConnections.List(dsLocationParent(project, g.flagDSPCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing private connections: %w", err)
	}
	if g.flagDSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.PrivateConnections, g.flagDSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.PrivateConnections), "%-40s %s\n", "NAME", "STATE")
	for _, pc := range resp.PrivateConnections {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(pc.Name), pc.State)
	}
	return nil
}

// --- routes (under private-connections) ---

func (g *datastreamCmds) initDatastreamAll5() {
	all := []*cobra.Command{g.dsRouteCreateCmd, g.dsRouteDeleteCmd, g.dsRouteDescribeCmd, g.dsRouteListCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagDSRouteLocation, "location", "", "Location containing the private connection (required)")
		c.Flags().StringVar(&g.flagDSRoutePC, "private-connection", "", "Private connection containing the route (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("private-connection")
	}
	g.dsRouteCreateCmd.Flags().StringVar(&g.flagDSRouteConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Route message body (required)")
	_ = g.dsRouteCreateCmd.MarkFlagRequired("config-file")
	for _, c := range []*cobra.Command{g.dsRouteCreateCmd, g.dsRouteDeleteCmd} {
		c.Flags().BoolVar(&g.flagDSRouteAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.dsRouteDescribeCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsRouteListCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")

	g.dsRoutesCmd.AddCommand(all...)
	g.datastreamCmd.AddCommand(g.dsRoutesCmd)
}

func dsRouteParent(project, location, pc string) string {
	return fmt.Sprintf("%s/privateConnections/%s", dsLocationParent(project, location), pc)
}

func dsRouteName(id, project, location, pc string) string {
	return dsChildName("routes", id, dsRouteParent(project, location, pc))
}

func (g *datastreamCmds) runDSRouteCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	route := &datastream.Route{}
	if err := loadYAMLOrJSONInto(g.flagDSRouteConfigFile, route); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Routes.Create(dsRouteParent(project, g.flagDSRouteLocation, g.flagDSRoutePC), route).
		RouteId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating route: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Create route", args[0], g.flagDSRouteAsync)
}

func (g *datastreamCmds) runDSRouteDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Routes.Delete(dsRouteName(args[0], project, g.flagDSRouteLocation, g.flagDSRoutePC)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting route: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Delete route", args[0], g.flagDSRouteAsync)
}

func (g *datastreamCmds) runDSRouteDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PrivateConnections.Routes.Get(dsRouteName(args[0], project, g.flagDSRouteLocation, g.flagDSRoutePC)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing route: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagDSFormat)
}

func (g *datastreamCmds) runDSRouteList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.PrivateConnections.Routes.List(dsRouteParent(project, g.flagDSRouteLocation, g.flagDSRoutePC)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing routes: %w", err)
	}
	if g.flagDSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Routes, g.flagDSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Routes), "%-40s %s\n", "NAME", "DESTINATION_ADDRESS")
	for _, r := range resp.Routes {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(r.Name), r.DestinationAddress)
	}
	return nil
}

// --- streams ---

func (g *datastreamCmds) initDatastreamAll6() {
	all := []*cobra.Command{g.dsStreamCreateCmd, g.dsStreamDeleteCmd, g.dsStreamDescribeCmd, g.dsStreamListCmd, g.dsStreamUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagDSStreamLocation, "location", "", "Location containing the stream (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.dsStreamCreateCmd, g.dsStreamUpdateCmd} {
		c.Flags().StringVar(&g.flagDSStreamConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Stream message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.dsStreamUpdateCmd.Flags().StringVar(&g.flagDSStreamUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.dsStreamCreateCmd, g.dsStreamDeleteCmd, g.dsStreamUpdateCmd} {
		c.Flags().BoolVar(&g.flagDSStreamAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.dsStreamDescribeCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsStreamListCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")

	g.dsStreamsCmd.AddCommand(all...)
	g.datastreamCmd.AddCommand(g.dsStreamsCmd)
}

func dsStreamName(id, project, location string) string {
	return dsChildName("streams", id, dsLocationParent(project, location))
}

func (g *datastreamCmds) runDSStreamCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &datastream.Stream{}
	if err := loadYAMLOrJSONInto(g.flagDSStreamConfigFile, s); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Streams.Create(dsLocationParent(project, g.flagDSStreamLocation), s).
		StreamId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating stream: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Create stream", args[0], g.flagDSStreamAsync)
}

func (g *datastreamCmds) runDSStreamDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Streams.Delete(dsStreamName(args[0], project, g.flagDSStreamLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting stream: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Delete stream", args[0], g.flagDSStreamAsync)
}

func (g *datastreamCmds) runDSStreamDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Get(dsStreamName(args[0], project, g.flagDSStreamLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing stream: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagDSFormat)
}

func (g *datastreamCmds) runDSStreamList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Streams.List(dsLocationParent(project, g.flagDSStreamLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing streams: %w", err)
	}
	if g.flagDSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Streams, g.flagDSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Streams), "%-40s %s\n", "NAME", "STATE")
	for _, s := range resp.Streams {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(s.Name), s.State)
	}
	return nil
}

func (g *datastreamCmds) runDSStreamUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &datastream.Stream{}
	if err := loadYAMLOrJSONInto(g.flagDSStreamConfigFile, s); err != nil {
		return err
	}
	mask := g.flagDSStreamUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Streams.Patch(dsStreamName(args[0], project, g.flagDSStreamLocation), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating stream: %w", err)
	}
	return g.dsFinishOp(ctx, svc, op, "Update stream", args[0], g.flagDSStreamAsync)
}

// --- objects ---

func (g *datastreamCmds) initDatastreamAll7() {
	all := []*cobra.Command{g.dsObjDescribeCmd, g.dsObjListCmd, g.dsObjLookupCmd, g.dsObjStartBackfillCmd, g.dsObjStopBackfillCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagDSObjLocation, "location", "", "Location containing the stream (required)")
		c.Flags().StringVar(&g.flagDSObjStream, "stream", "", "Stream containing the object (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("stream")
	}
	g.dsObjLookupCmd.Flags().StringVar(&g.flagDSObjConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the LookupStreamObjectRequest body (required)")
	_ = g.dsObjLookupCmd.MarkFlagRequired("config-file")
	g.dsObjStartBackfillCmd.Flags().StringVar(&g.flagDSObjSQLWhere, "sql-where-clause", "",
		"Optional SQL WHERE clause (without the WHERE keyword) that limits the backfill to matching rows; only supported for SQL sources")
	g.dsObjDescribeCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")
	g.dsObjListCmd.Flags().StringVar(&g.flagDSFormat, "format", "", "Output format")

	g.dsObjectsCmd.AddCommand(all...)
	g.datastreamCmd.AddCommand(g.dsObjectsCmd)
}

func dsObjParent(project, location, stream string) string {
	return fmt.Sprintf("%s/streams/%s", dsLocationParent(project, location), stream)
}

func dsObjName(id, project, location, stream string) string {
	return dsChildName("objects", id, dsObjParent(project, location, stream))
}

func (g *datastreamCmds) runDSObjDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Objects.Get(dsObjName(args[0], project, g.flagDSObjLocation, g.flagDSObjStream)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing stream object: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagDSFormat)
}

func (g *datastreamCmds) runDSObjList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Streams.Objects.List(dsObjParent(project, g.flagDSObjLocation, g.flagDSObjStream)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing stream objects: %w", err)
	}
	if g.flagDSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.StreamObjects, g.flagDSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.StreamObjects), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, o := range resp.StreamObjects {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(o.Name), o.DisplayName)
	}
	return nil
}

func (g *datastreamCmds) runDSObjLookup(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	req := &datastream.LookupStreamObjectRequest{}
	if err := loadYAMLOrJSONInto(g.flagDSObjConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Objects.Lookup(dsObjParent(project, g.flagDSObjLocation, g.flagDSObjStream), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up stream object: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagDSFormat)
}

func (g *datastreamCmds) runDSObjStartBackfill(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	req := &datastream.StartBackfillJobRequest{}
	if g.flagDSObjSQLWhere != "" {
		req.EventFilter = &datastream.EventFilter{SqlWhereClause: g.flagDSObjSQLWhere}
	}
	got, err := svc.Projects.Locations.Streams.Objects.StartBackfillJob(dsObjName(args[0], project, g.flagDSObjLocation, g.flagDSObjStream), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting backfill: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *datastreamCmds) runDSObjStopBackfill(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Datastream(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Streams.Objects.StopBackfillJob(dsObjName(args[0], project, g.flagDSObjLocation, g.flagDSObjStream), &datastream.StopBackfillJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping backfill: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}
