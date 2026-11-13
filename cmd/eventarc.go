package cmd

import "github.com/spf13/cobra"

// --- gcloud eventarc (#338) ---

func (grp *eventarcCmds) initEventarc() {
	// gcloud-python's `eventarc --help` groups: audit-logs-provider,
	// channel-connections, channels, enrollments, gke-destinations,
	// google-api-sources, google-channels, kafka-sources, locations,
	// message-buses, pipelines, providers, triggers. No `attributes`; that
	// stub was invented and has been removed (#1720). Every other subgroup
	// is registered from its own eventarc_*.go file.

}

// eventarcCmds holds the commands and flag values of one invocation's "eventarc"
// command group, so invocations don't share state (#2070).
type eventarcCmds struct {
	inv *invocation

	eventarcCmd                          *cobra.Command
	eventarcAuditLogsProviderCmd         *cobra.Command
	eventarcAuditLogsMethodNamesCmd      *cobra.Command
	eventarcAuditLogsServiceNamesCmd     *cobra.Command
	eventarcAuditLogsMethodNamesListCmd  *cobra.Command
	eventarcAuditLogsServiceNamesListCmd *cobra.Command
	flagEventarcAuditLogsServiceName     string
	eventarcChannelConnCmd               *cobra.Command
	evCCCreateCmd                        *cobra.Command
	evCCDeleteCmd                        *cobra.Command
	evCCDescribeCmd                      *cobra.Command
	evCCListCmd                          *cobra.Command
	flagEvCCLocation                     string
	flagEvCCConfigFile                   string
	flagEvCCFormat                       string
	flagEvCCAsync                        bool
	flagEvCCListLimit                    int64
	flagEvCCListPage                     int64
	flagEvCCListURI                      bool
	eventarcChannelsCmd                  *cobra.Command
	evChanCreateCmd                      *cobra.Command
	evChanDeleteCmd                      *cobra.Command
	evChanDescribeCmd                    *cobra.Command
	evChanListCmd                        *cobra.Command
	evChanUpdateCmd                      *cobra.Command
	flagEvChanLocation                   string
	flagEvChanConfigFile                 string
	flagEvChanUpdateMask                 string
	flagEvChanFormat                     string
	flagEvChanAsync                      bool
	flagEvChanListLimit                  int64
	flagEvChanListPage                   int64
	flagEvChanListURI                    bool
	eventarcEnrollmentsCmd               *cobra.Command
	evEnrCreateCmd                       *cobra.Command
	evEnrDeleteCmd                       *cobra.Command
	evEnrDescribeCmd                     *cobra.Command
	evEnrListCmd                         *cobra.Command
	evEnrUpdateCmd                       *cobra.Command
	flagEvEnrLocation                    string
	flagEvEnrConfigFile                  string
	flagEvEnrUpdateMask                  string
	flagEvEnrFormat                      string
	flagEvEnrAsync                       bool
	flagEvEnrListLimit                   int64
	flagEvEnrListPage                    int64
	flagEvEnrListFilter                  string
	flagEvEnrListURI                     bool
	eventarcGASCmd                       *cobra.Command
	evGASCreateCmd                       *cobra.Command
	evGASDeleteCmd                       *cobra.Command
	evGASDescribeCmd                     *cobra.Command
	evGASListCmd                         *cobra.Command
	evGASUpdateCmd                       *cobra.Command
	flagEvGASLocation                    string
	flagEvGASConfigFile                  string
	flagEvGASUpdateMask                  string
	flagEvGASFormat                      string
	flagEvGASAsync                       bool
	flagEvGASListLimit                   int64
	flagEvGASListPage                    int64
	flagEvGASListFilter                  string
	flagEvGASListURI                     bool
	eventarcGoogleChannelsCmd            *cobra.Command
	evGChanDescribeCmd                   *cobra.Command
	evGChanUpdateCmd                     *cobra.Command
	flagEvGChanLocation                  string
	flagEvGChanConfigFile                string
	flagEvGChanUpdateMask                string
	flagEvGChanFormat                    string
	eventarcLocationsCmd                 *cobra.Command
	eventarcLocationsListCmd             *cobra.Command
	eventarcProvidersCmd                 *cobra.Command
	evProvDescribeCmd                    *cobra.Command
	evProvListCmd                        *cobra.Command
	flagEvLocFormat                      string
	flagEvLocListLimit                   int64
	flagEvLocListPage                    int64
	flagEvProvLocation                   string
	flagEvProvFormat                     string
	flagEvProvListLimit                  int64
	flagEvProvListPage                   int64
	flagEvProvListFilter                 string
	flagEvProvListURI                    bool
	eventarcMessageBusesCmd              *cobra.Command
	evMBCreateCmd                        *cobra.Command
	evMBDeleteCmd                        *cobra.Command
	evMBDescribeCmd                      *cobra.Command
	evMBListCmd                          *cobra.Command
	evMBUpdateCmd                        *cobra.Command
	evMBListEnrCmd                       *cobra.Command
	evMBPublishCmd                       *cobra.Command
	flagEvMBLocation                     string
	flagEvMBConfigFile                   string
	flagEvMBUpdateMask                   string
	flagEvMBFormat                       string
	flagEvMBAsync                        bool
	flagEvMBListLimit                    int64
	flagEvMBListPage                     int64
	flagEvMBListFilter                   string
	flagEvMBListURI                      bool
	flagEvMBPublishEventID               string
	flagEvMBPublishType                  string
	flagEvMBPublishSource                string
	flagEvMBPublishData                  string
	flagEvMBPublishAttrs                 []string
	flagEvMBPublishJSON                  string
	flagEvMBPublishEnroll                string
	eventarcPipelinesCmd                 *cobra.Command
	evPipCreateCmd                       *cobra.Command
	evPipDeleteCmd                       *cobra.Command
	evPipDescribeCmd                     *cobra.Command
	evPipListCmd                         *cobra.Command
	evPipUpdateCmd                       *cobra.Command
	flagEvPipLocation                    string
	flagEvPipConfigFile                  string
	flagEvPipUpdateMask                  string
	flagEvPipFormat                      string
	flagEvPipAsync                       bool
	flagEvPipListLimit                   int64
	flagEvPipListPage                    int64
	flagEvPipListFilter                  string
	flagEvPipListURI                     bool
	eventarcTriggersCmd                  *cobra.Command
	evTrgCreateCmd                       *cobra.Command
	evTrgDeleteCmd                       *cobra.Command
	evTrgDescribeCmd                     *cobra.Command
	evTrgListCmd                         *cobra.Command
	evTrgUpdateCmd                       *cobra.Command
	flagEvTrgLocation                    string
	flagEvTrgConfigFile                  string
	flagEvTrgUpdateMask                  string
	flagEvTrgFormat                      string
	flagEvTrgAsync                       bool
	flagEvTrgListLimit                   int64
	flagEvTrgListPage                    int64
	flagEvTrgListFilter                  string
	flagEvTrgListURI                     bool
}

// newEventarcCmd returns the "eventarc" command group for inv.
func newEventarcCmd(inv *invocation) *cobra.Command {
	return newEventarcCmds(inv).eventarcCmd
}

// newEventarcCmds builds the "eventarc" command group for inv and returns its state.
func newEventarcCmds(inv *invocation) *eventarcCmds {
	grp := &eventarcCmds{inv: inv}
	grp.eventarcCmd = &cobra.Command{Use: "eventarc", Short: "Manage Eventarc"}
	grp.eventarcAuditLogsProviderCmd = &cobra.Command{
		Use:   "audit-logs-provider",
		Short: "Explore audit log providers for Eventarc",
		Long: `Explore provider serviceNames and methodNames for event type
google.cloud.audit.log.v1.written in Eventarc.`,
	}
	grp.eventarcAuditLogsMethodNamesCmd = &cobra.Command{
		Use:   "method-names",
		Short: "Explore values for the methodName attribute",
	}
	grp.eventarcAuditLogsServiceNamesCmd = &cobra.Command{
		Use:   "service-names",
		Short: "Explore values for the serviceName attribute",
	}
	grp.eventarcAuditLogsMethodNamesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List values for the methodName attribute",
		Args:  cobra.NoArgs,
		RunE:  grp.runEventarcAuditLogsMethodNamesList,
	}
	grp.eventarcAuditLogsServiceNamesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List values for the serviceName attribute",
		Args:  cobra.NoArgs,
		RunE:  grp.runEventarcAuditLogsServiceNamesList,
	}
	grp.eventarcChannelConnCmd = &cobra.Command{
		Use:   "channel-connections",
		Short: "Manage Eventarc channel connections",
	}
	grp.evCCCreateCmd = &cobra.Command{
		Use: "create CHANNEL_CONNECTION", Short: "Create a channel connection from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvCCCreate,
	}
	grp.evCCDeleteCmd = &cobra.Command{
		Use: "delete CHANNEL_CONNECTION", Short: "Delete a channel connection",
		Args: cobra.ExactArgs(1), RunE: grp.runEvCCDelete,
	}
	grp.evCCDescribeCmd = &cobra.Command{
		Use: "describe CHANNEL_CONNECTION", Short: "Describe a channel connection",
		Args: cobra.ExactArgs(1), RunE: grp.runEvCCDescribe,
	}
	grp.evCCListCmd = &cobra.Command{
		Use: "list", Short: "List channel connections in a location",
		Args: cobra.NoArgs, RunE: grp.runEvCCList,
	}
	grp.eventarcChannelsCmd = &cobra.Command{
		Use:   "channels",
		Short: "Manage Eventarc channels",
	}
	grp.evChanCreateCmd = &cobra.Command{
		Use: "create CHANNEL", Short: "Create an Eventarc channel from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvChanCreate,
	}
	grp.evChanDeleteCmd = &cobra.Command{
		Use: "delete CHANNEL", Short: "Delete an Eventarc channel",
		Args: cobra.ExactArgs(1), RunE: grp.runEvChanDelete,
	}
	grp.evChanDescribeCmd = &cobra.Command{
		Use: "describe CHANNEL", Short: "Describe an Eventarc channel",
		Args: cobra.ExactArgs(1), RunE: grp.runEvChanDescribe,
	}
	grp.evChanListCmd = &cobra.Command{
		Use: "list", Short: "List Eventarc channels in a location",
		Args: cobra.NoArgs, RunE: grp.runEvChanList,
	}
	grp.evChanUpdateCmd = &cobra.Command{
		Use: "update CHANNEL", Short: "Update an Eventarc channel from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvChanUpdate,
	}
	grp.eventarcEnrollmentsCmd = &cobra.Command{
		Use:   "enrollments",
		Short: "Manage Eventarc enrollments",
	}
	grp.evEnrCreateCmd = &cobra.Command{
		Use: "create ENROLLMENT", Short: "Create an enrollment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvEnrCreate,
	}
	grp.evEnrDeleteCmd = &cobra.Command{
		Use: "delete ENROLLMENT", Short: "Delete an enrollment",
		Args: cobra.ExactArgs(1), RunE: grp.runEvEnrDelete,
	}
	grp.evEnrDescribeCmd = &cobra.Command{
		Use: "describe ENROLLMENT", Short: "Describe an enrollment",
		Args: cobra.ExactArgs(1), RunE: grp.runEvEnrDescribe,
	}
	grp.evEnrListCmd = &cobra.Command{
		Use: "list", Short: "List enrollments in a location",
		Args: cobra.NoArgs, RunE: grp.runEvEnrList,
	}
	grp.evEnrUpdateCmd = &cobra.Command{
		Use: "update ENROLLMENT", Short: "Update an enrollment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvEnrUpdate,
	}
	grp.eventarcGASCmd = &cobra.Command{
		Use:   "google-api-sources",
		Short: "Manage Eventarc Google API sources",
	}
	grp.evGASCreateCmd = &cobra.Command{
		Use: "create GOOGLE_API_SOURCE", Short: "Create a Google API source from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvGASCreate,
	}
	grp.evGASDeleteCmd = &cobra.Command{
		Use: "delete GOOGLE_API_SOURCE", Short: "Delete a Google API source",
		Args: cobra.ExactArgs(1), RunE: grp.runEvGASDelete,
	}
	grp.evGASDescribeCmd = &cobra.Command{
		Use: "describe GOOGLE_API_SOURCE", Short: "Describe a Google API source",
		Args: cobra.ExactArgs(1), RunE: grp.runEvGASDescribe,
	}
	grp.evGASListCmd = &cobra.Command{
		Use: "list", Short: "List Google API sources in a location",
		Args: cobra.NoArgs, RunE: grp.runEvGASList,
	}
	grp.evGASUpdateCmd = &cobra.Command{
		Use: "update GOOGLE_API_SOURCE", Short: "Update a Google API source from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvGASUpdate,
	}
	grp.eventarcGoogleChannelsCmd = &cobra.Command{
		Use:   "google-channels",
		Short: "Manage the Google-managed Eventarc channel configuration",
	}
	grp.evGChanDescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the Google Channel configuration",
		Args:  cobra.NoArgs,
		RunE:  grp.runEvGChanDescribe,
	}
	grp.evGChanUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the Google Channel configuration from a --config-file",
		Args:  cobra.NoArgs,
		RunE:  grp.runEvGChanUpdate,
	}
	grp.eventarcLocationsCmd = &cobra.Command{
		Use:   "locations",
		Short: "Explore Eventarc locations",
	}
	grp.eventarcLocationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Eventarc-enabled locations for the project",
		Args:  cobra.NoArgs,
		RunE:  grp.runEventarcLocationsList,
	}
	grp.eventarcProvidersCmd = &cobra.Command{
		Use:   "providers",
		Short: "Explore Eventarc event providers",
	}
	grp.evProvDescribeCmd = &cobra.Command{
		Use:   "describe PROVIDER",
		Short: "Describe an event provider",
		Args:  cobra.ExactArgs(1),
		RunE:  grp.runEvProvDescribe,
	}
	grp.evProvListCmd = &cobra.Command{
		Use:   "list",
		Short: "List event providers in a location",
		Args:  cobra.NoArgs,
		RunE:  grp.runEvProvList,
	}
	grp.eventarcMessageBusesCmd = &cobra.Command{
		Use:   "message-buses",
		Short: "Manage Eventarc message buses",
	}
	grp.evMBCreateCmd = &cobra.Command{
		Use: "create MESSAGE_BUS", Short: "Create a message bus from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvMBCreate,
	}
	grp.evMBDeleteCmd = &cobra.Command{
		Use: "delete MESSAGE_BUS", Short: "Delete a message bus",
		Args: cobra.ExactArgs(1), RunE: grp.runEvMBDelete,
	}
	grp.evMBDescribeCmd = &cobra.Command{
		Use: "describe MESSAGE_BUS", Short: "Describe a message bus",
		Args: cobra.ExactArgs(1), RunE: grp.runEvMBDescribe,
	}
	grp.evMBListCmd = &cobra.Command{
		Use: "list", Short: "List message buses in a location",
		Args: cobra.NoArgs, RunE: grp.runEvMBList,
	}
	grp.evMBUpdateCmd = &cobra.Command{
		Use: "update MESSAGE_BUS", Short: "Update a message bus from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvMBUpdate,
	}
	grp.evMBListEnrCmd = &cobra.Command{
		Use: "list-enrollments MESSAGE_BUS", Short: "List enrollments attached to a message bus",
		Args: cobra.ExactArgs(1), RunE: grp.runEvMBListEnrollments,
	}
	grp.evMBPublishCmd = &cobra.Command{
		Use: "publish MESSAGE_BUS", Short: "Publish an event to a message bus",
		Args: cobra.ExactArgs(1), RunE: grp.runEvMBPublish,
	}
	grp.eventarcPipelinesCmd = &cobra.Command{
		Use:   "pipelines",
		Short: "Manage Eventarc pipelines",
	}
	grp.evPipCreateCmd = &cobra.Command{
		Use: "create PIPELINE", Short: "Create a pipeline from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvPipCreate,
	}
	grp.evPipDeleteCmd = &cobra.Command{
		Use: "delete PIPELINE", Short: "Delete a pipeline",
		Args: cobra.ExactArgs(1), RunE: grp.runEvPipDelete,
	}
	grp.evPipDescribeCmd = &cobra.Command{
		Use: "describe PIPELINE", Short: "Describe a pipeline",
		Args: cobra.ExactArgs(1), RunE: grp.runEvPipDescribe,
	}
	grp.evPipListCmd = &cobra.Command{
		Use: "list", Short: "List pipelines in a location",
		Args: cobra.NoArgs, RunE: grp.runEvPipList,
	}
	grp.evPipUpdateCmd = &cobra.Command{
		Use: "update PIPELINE", Short: "Update a pipeline from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvPipUpdate,
	}
	grp.eventarcTriggersCmd = &cobra.Command{
		Use:   "triggers",
		Short: "Manage Eventarc triggers",
	}
	grp.evTrgCreateCmd = &cobra.Command{
		Use: "create TRIGGER", Short: "Create a trigger from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvTrgCreate,
	}
	grp.evTrgDeleteCmd = &cobra.Command{
		Use: "delete TRIGGER", Short: "Delete a trigger",
		Args: cobra.ExactArgs(1), RunE: grp.runEvTrgDelete,
	}
	grp.evTrgDescribeCmd = &cobra.Command{
		Use: "describe TRIGGER", Short: "Describe a trigger",
		Args: cobra.ExactArgs(1), RunE: grp.runEvTrgDescribe,
	}
	grp.evTrgListCmd = &cobra.Command{
		Use: "list", Short: "List triggers in a location",
		Args: cobra.NoArgs, RunE: grp.runEvTrgList,
	}
	grp.evTrgUpdateCmd = &cobra.Command{
		Use: "update TRIGGER", Short: "Update a trigger from a --config-file",
		Args: cobra.ExactArgs(1), RunE: grp.runEvTrgUpdate,
	}
	grp.initEventarc()
	grp.initEventarcAuditLogsProvider()
	grp.initEventarcChannelConnections()
	grp.initEventarcChannels()
	grp.initEventarcEnrollments()
	grp.initEventarcGoogleApiSources()
	grp.initEventarcGoogleChannels()
	grp.initEventarcLocationsProviders()
	grp.initEventarcMessageBuses()
	grp.initEventarcPipelines()
	grp.initEventarcTriggers()
	return grp
}
