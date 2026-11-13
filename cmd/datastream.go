package cmd

import "github.com/spf13/cobra"

// --- gcloud datastream (#326) ---

func (g *datastreamCmds) initDatastream() {
	// All subgroups (connection-profiles, locations, objects, operations,
	// private-connections, routes, streams) are implemented in
	// datastream_all.go.

}

// datastreamCmds holds the commands and flag values of one invocation's "datastream"
// command group, so invocations don't share state (#2070).
type datastreamCmds struct {
	inv *invocation

	datastreamCmd          *cobra.Command
	dsLocationsCmd         *cobra.Command
	dsLocDescribeCmd       *cobra.Command
	dsLocListCmd           *cobra.Command
	flagDSFormat           string
	dsOperationsCmd        *cobra.Command
	dsOpCancelCmd          *cobra.Command
	dsOpDeleteCmd          *cobra.Command
	dsOpDescribeCmd        *cobra.Command
	dsOpListCmd            *cobra.Command
	flagDSOpLocation       string
	dsConnProfilesCmd      *cobra.Command
	dsCPCreateCmd          *cobra.Command
	dsCPDeleteCmd          *cobra.Command
	dsCPDescribeCmd        *cobra.Command
	dsCPDiscoverCmd        *cobra.Command
	dsCPListCmd            *cobra.Command
	dsCPUpdateCmd          *cobra.Command
	flagDSCPLocation       string
	flagDSCPConfigFile     string
	flagDSCPUpdateMask     string
	flagDSCPAsync          bool
	dsPrivateConnsCmd      *cobra.Command
	dsPCCreateCmd          *cobra.Command
	dsPCDeleteCmd          *cobra.Command
	dsPCDescribeCmd        *cobra.Command
	dsPCListCmd            *cobra.Command
	flagDSPCLocation       string
	flagDSPCConfigFile     string
	flagDSPCAsync          bool
	dsRoutesCmd            *cobra.Command
	dsRouteCreateCmd       *cobra.Command
	dsRouteDeleteCmd       *cobra.Command
	dsRouteDescribeCmd     *cobra.Command
	dsRouteListCmd         *cobra.Command
	flagDSRouteLocation    string
	flagDSRoutePC          string
	flagDSRouteConfigFile  string
	flagDSRouteAsync       bool
	dsStreamsCmd           *cobra.Command
	dsStreamCreateCmd      *cobra.Command
	dsStreamDeleteCmd      *cobra.Command
	dsStreamDescribeCmd    *cobra.Command
	dsStreamListCmd        *cobra.Command
	dsStreamUpdateCmd      *cobra.Command
	flagDSStreamLocation   string
	flagDSStreamConfigFile string
	flagDSStreamUpdateMask string
	flagDSStreamAsync      bool
	dsObjectsCmd           *cobra.Command
	dsObjDescribeCmd       *cobra.Command
	dsObjListCmd           *cobra.Command
	dsObjLookupCmd         *cobra.Command
	dsObjStartBackfillCmd  *cobra.Command
	dsObjStopBackfillCmd   *cobra.Command
	flagDSObjLocation      string
	flagDSObjStream        string
	flagDSObjConfigFile    string
	flagDSObjSQLWhere      string
}

// newDatastreamCmd returns the "datastream" command group for inv.
func newDatastreamCmd(inv *invocation) *cobra.Command {
	return newDatastreamCmds(inv).datastreamCmd
}

// newDatastreamCmds builds the "datastream" command group for inv and returns its state.
func newDatastreamCmds(inv *invocation) *datastreamCmds {
	g := &datastreamCmds{inv: inv}
	g.datastreamCmd = &cobra.Command{Use: "datastream", Short: "Manage Datastream"}
	g.dsLocationsCmd = &cobra.Command{Use: "locations", Short: "Explore Datastream locations"}
	g.dsLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Datastream location",
		Args: cobra.ExactArgs(1), RunE: g.runDSLocDescribe,
	}
	g.dsLocListCmd = &cobra.Command{
		Use: "list", Short: "List Datastream locations for the project",
		Args: cobra.NoArgs, RunE: g.runDSLocList,
	}
	g.dsOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Datastream operations"}
	g.dsOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Datastream operation",
		Args: cobra.ExactArgs(1), RunE: g.runDSOpCancel,
	}
	g.dsOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Datastream operation",
		Args: cobra.ExactArgs(1), RunE: g.runDSOpDelete,
	}
	g.dsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Datastream operation",
		Args: cobra.ExactArgs(1), RunE: g.runDSOpDescribe,
	}
	g.dsOpListCmd = &cobra.Command{
		Use: "list", Short: "List Datastream operations in a location",
		Args: cobra.NoArgs, RunE: g.runDSOpList,
	}
	g.dsConnProfilesCmd = &cobra.Command{Use: "connection-profiles", Short: "Manage Datastream connection profiles"}
	g.dsCPCreateCmd = &cobra.Command{
		Use: "create PROFILE", Short: "Create a connection profile from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runDSCPCreate,
	}
	g.dsCPDeleteCmd = &cobra.Command{
		Use: "delete PROFILE", Short: "Delete a connection profile",
		Args: cobra.ExactArgs(1), RunE: g.runDSCPDelete,
	}
	g.dsCPDescribeCmd = &cobra.Command{
		Use: "describe PROFILE", Short: "Describe a connection profile",
		Args: cobra.ExactArgs(1), RunE: g.runDSCPDescribe,
	}
	g.dsCPDiscoverCmd = &cobra.Command{
		Use: "discover PROFILE", Short: "Discover the schema of a connection profile from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runDSCPDiscover,
	}
	g.dsCPListCmd = &cobra.Command{
		Use: "list", Short: "List connection profiles in a location",
		Args: cobra.NoArgs, RunE: g.runDSCPList,
	}
	g.dsCPUpdateCmd = &cobra.Command{
		Use: "update PROFILE", Short: "Update a connection profile from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runDSCPUpdate,
	}
	g.dsPrivateConnsCmd = &cobra.Command{Use: "private-connections", Short: "Manage Datastream private connections"}
	g.dsPCCreateCmd = &cobra.Command{
		Use: "create PRIVATE_CONNECTION", Short: "Create a private connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runDSPCCreate,
	}
	g.dsPCDeleteCmd = &cobra.Command{
		Use: "delete PRIVATE_CONNECTION", Short: "Delete a private connection",
		Args: cobra.ExactArgs(1), RunE: g.runDSPCDelete,
	}
	g.dsPCDescribeCmd = &cobra.Command{
		Use: "describe PRIVATE_CONNECTION", Short: "Describe a private connection",
		Args: cobra.ExactArgs(1), RunE: g.runDSPCDescribe,
	}
	g.dsPCListCmd = &cobra.Command{
		Use: "list", Short: "List private connections in a location",
		Args: cobra.NoArgs, RunE: g.runDSPCList,
	}
	g.dsRoutesCmd = &cobra.Command{Use: "routes", Short: "Manage Datastream routes (within a private connection)"}
	g.dsRouteCreateCmd = &cobra.Command{
		Use: "create ROUTE", Short: "Create a route from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runDSRouteCreate,
	}
	g.dsRouteDeleteCmd = &cobra.Command{
		Use: "delete ROUTE", Short: "Delete a route",
		Args: cobra.ExactArgs(1), RunE: g.runDSRouteDelete,
	}
	g.dsRouteDescribeCmd = &cobra.Command{
		Use: "describe ROUTE", Short: "Describe a route",
		Args: cobra.ExactArgs(1), RunE: g.runDSRouteDescribe,
	}
	g.dsRouteListCmd = &cobra.Command{
		Use: "list", Short: "List routes for a private connection",
		Args: cobra.NoArgs, RunE: g.runDSRouteList,
	}
	g.dsStreamsCmd = &cobra.Command{Use: "streams", Short: "Manage Datastream streams"}
	g.dsStreamCreateCmd = &cobra.Command{
		Use: "create STREAM", Short: "Create a stream from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runDSStreamCreate,
	}
	g.dsStreamDeleteCmd = &cobra.Command{
		Use: "delete STREAM", Short: "Delete a stream",
		Args: cobra.ExactArgs(1), RunE: g.runDSStreamDelete,
	}
	g.dsStreamDescribeCmd = &cobra.Command{
		Use: "describe STREAM", Short: "Describe a stream",
		Args: cobra.ExactArgs(1), RunE: g.runDSStreamDescribe,
	}
	g.dsStreamListCmd = &cobra.Command{
		Use: "list", Short: "List streams in a location",
		Args: cobra.NoArgs, RunE: g.runDSStreamList,
	}
	g.dsStreamUpdateCmd = &cobra.Command{
		Use: "update STREAM", Short: "Update a stream from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runDSStreamUpdate,
	}
	g.dsObjectsCmd = &cobra.Command{Use: "objects", Short: "Manage Datastream stream objects"}
	g.dsObjDescribeCmd = &cobra.Command{
		Use: "describe OBJECT", Short: "Describe a stream object",
		Args: cobra.ExactArgs(1), RunE: g.runDSObjDescribe,
	}
	g.dsObjListCmd = &cobra.Command{
		Use: "list", Short: "List stream objects for a stream",
		Args: cobra.NoArgs, RunE: g.runDSObjList,
	}
	g.dsObjLookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up a stream object by source identifier (--config-file)",
		Args: cobra.NoArgs, RunE: g.runDSObjLookup,
	}
	g.dsObjStartBackfillCmd = &cobra.Command{
		Use: "start-backfill OBJECT", Short: "Start a backfill for a stream object",
		Args: cobra.ExactArgs(1), RunE: g.runDSObjStartBackfill,
	}
	g.dsObjStopBackfillCmd = &cobra.Command{
		Use: "stop-backfill OBJECT", Short: "Stop a backfill for a stream object",
		Args: cobra.ExactArgs(1), RunE: g.runDSObjStopBackfill,
	}
	g.initDatastream()
	g.initDatastreamAll()
	g.initDatastreamAll2()
	g.initDatastreamAll3()
	g.initDatastreamAll4()
	g.initDatastreamAll5()
	g.initDatastreamAll6()
	g.initDatastreamAll7()
	return g
}
