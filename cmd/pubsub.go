package cmd

import "github.com/spf13/cobra"

// --- gcloud pubsub (#377) ---

func (g *pubsubCmds) initPubsub() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(g.pubsubCmd, "schemas", "Manage Pub/Sub schemas", append(crud, "commit", "revisions", "validate-schema", "validate-message")...)
	registerStubGroup(g.pubsubCmd, "snapshots", "Manage Pub/Sub snapshots", crud...)

}

// pubsubCmds holds the commands and flag values of one invocation's "pubsub"
// command group, so invocations don't share state (#2070).
type pubsubCmds struct {
	inv *invocation

	pubsubCmd                  *cobra.Command
	pubsubLiteOperationsCmd    *cobra.Command
	flagPSLOpLocation          string
	flagPSLOpFormat            string
	flagPSLOpFilter            string
	flagPSLOpSubscription      string
	flagPSLOpPageSize          int64
	flagPSLOpDone              bool
	flagPSLOpDoneSet           bool
	pubsubLiteOpDescribeCmd    *cobra.Command
	pubsubLiteOpListCmd        *cobra.Command
	pubsubLiteReservationsCmd  *cobra.Command
	flagPSLResLocation         string
	flagPSLResThroughput       int64
	flagPSLResFormat           string
	flagPSLResPageSize         int64
	pubsubLiteResCreateCmd     *cobra.Command
	pubsubLiteResDeleteCmd     *cobra.Command
	pubsubLiteResDescribeCmd   *cobra.Command
	pubsubLiteResListCmd       *cobra.Command
	pubsubLiteResListTopicsCmd *cobra.Command
	pubsubLiteResUpdateCmd     *cobra.Command
	pubsubLiteSubsCmd          *cobra.Command
	flagPSLSubLocation         string
	flagPSLSubFormat           string
	flagPSLSubConfigFile       string
	flagPSLSubUpdateMask       string
	flagPSLSubPageSize         int64
	flagPSLSubPartition        int64
	flagPSLSubOffset           int64
	pubsubLiteSubAckUpToCmd    *cobra.Command
	pubsubLiteSubCreateCmd     *cobra.Command
	pubsubLiteSubDeleteCmd     *cobra.Command
	pubsubLiteSubDescribeCmd   *cobra.Command
	pubsubLiteSubListCmd       *cobra.Command
	pubsubLiteSubSeekCmd       *cobra.Command
	pubsubLiteSubSubscribeCmd  *cobra.Command
	pubsubLiteSubUpdateCmd     *cobra.Command
	pubsubLiteTopicsCmd        *cobra.Command
	flagPSLTopicLocation       string
	flagPSLTopicFormat         string
	flagPSLTopicConfigFile     string
	flagPSLTopicUpdateMask     string
	flagPSLTopicPageSize       int64
	pubsubLiteTopicCreateCmd   *cobra.Command
	pubsubLiteTopicDeleteCmd   *cobra.Command
	pubsubLiteTopicDescribeCmd *cobra.Command
	pubsubLiteTopicListCmd     *cobra.Command
	pubsubLiteTopicListSubsCmd *cobra.Command
	pubsubLiteTopicPublishCmd  *cobra.Command
	pubsubLiteTopicUpdateCmd   *cobra.Command
	pubsubMessageTransformsCmd *cobra.Command
	flagPMTMessage             string
	flagPMTAttributes          []string
	flagPMTMessageTransforms   string
	flagPMTTopic               string
	flagPMTSubscription        string
	flagPMTTransformFile       string
	flagPMTFormat              string
	pubsubMTTestCmd            *cobra.Command
	pubsubMTValidateCmd        *cobra.Command
	flagPSSchemaType           string
	flagPSSchemaDefinition     string
	flagPSSchemaDefFile        string
	flagPSSchemaMsgFile        string
	flagPSSchemaEncoding       string
	flagPSSchemaRevision       string
	flagPSSnapshotSub          string
	pubsubSchemasCmd           *cobra.Command
	psSchemaCreateCmd          *cobra.Command
	psSchemaDeleteCmd          *cobra.Command
	psSchemaDeleteRevCmd       *cobra.Command
	psSchemaDescribeCmd        *cobra.Command
	psSchemaListCmd            *cobra.Command
	psSchemaListRevsCmd        *cobra.Command
	psSchemaCommitCmd          *cobra.Command
	psSchemaRollbackCmd        *cobra.Command
	psSchemaValidateSchemaCmd  *cobra.Command
	psSchemaValidateMessageCmd *cobra.Command
	pubsubSnapshotsCmd         *cobra.Command
	psSnapCreateCmd            *cobra.Command
	psSnapDeleteCmd            *cobra.Command
	psSnapDescribeCmd          *cobra.Command
	psSnapListCmd              *cobra.Command
	flagPSConfigFile           string
	flagPSUpdateMask           string
	flagPSFormat               string
	flagPSIamMember            string
	flagPSIamRole              string
	// topics publish
	flagPSPublishMessage    string
	flagPSPublishAttributes map[string]string
	flagPSPublishOrderKey   string
	// subscriptions ack / seek / modify
	flagPSAckDeadline  int64
	flagPSMaxMessages  int64
	flagPSSeekTime     string
	flagPSSeekSnapshot string
	flagPSPushEndpoint string
	flagPSPushAudience string
	// subscriptions create hints
	flagPSSubTopic         string
	pubsubTopicsCmd        *cobra.Command
	psTopCreateCmd         *cobra.Command
	psTopDeleteCmd         *cobra.Command
	psTopDescribeCmd       *cobra.Command
	psTopListCmd           *cobra.Command
	psTopUpdateCmd         *cobra.Command
	psTopPublishCmd        *cobra.Command
	psTopListSubsCmd       *cobra.Command
	psTopDetachSubCmd      *cobra.Command
	psTopGetIamCmd         *cobra.Command
	psTopSetIamCmd         *cobra.Command
	psTopAddIamCmd         *cobra.Command
	psTopRemoveIamCmd      *cobra.Command
	pubsubSubscriptionsCmd *cobra.Command
	psSubCreateCmd         *cobra.Command
	psSubDeleteCmd         *cobra.Command
	psSubDescribeCmd       *cobra.Command
	psSubListCmd           *cobra.Command
	psSubUpdateCmd         *cobra.Command
	psSubAckCmd            *cobra.Command
	psSubPullCmd           *cobra.Command
	psSubSeekCmd           *cobra.Command
	psSubModifyAckCmd      *cobra.Command
	psSubModifyPushCmd     *cobra.Command
	psSubGetIamCmd         *cobra.Command
	psSubSetIamCmd         *cobra.Command
	psSubAddIamCmd         *cobra.Command
	psSubRemoveIamCmd      *cobra.Command
}

// newPubsubCmd returns the "pubsub" command group for inv.
func newPubsubCmd(inv *invocation) *cobra.Command {
	return newPubsubCmds(inv).pubsubCmd
}

// newPubsubCmds builds the "pubsub" command group for inv and returns its state.
func newPubsubCmds(inv *invocation) *pubsubCmds {
	g := &pubsubCmds{inv: inv}
	g.pubsubCmd = &cobra.Command{Use: "pubsub", Short: "Manage Cloud Pub/Sub"}
	g.pubsubLiteOperationsCmd = &cobra.Command{
		Use:   "lite-operations",
		Short: "Manage Pub/Sub Lite operations",
	}
	g.pubsubLiteOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Pub/Sub Lite operation",
		Args: cobra.ExactArgs(1), RunE: g.runPSLOpDescribe,
	}
	g.pubsubLiteOpListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub Lite operations in a location",
		Args: cobra.NoArgs, RunE: g.runPSLOpList,
	}
	g.pubsubLiteReservationsCmd = &cobra.Command{
		Use:   "lite-reservations",
		Short: "Manage Pub/Sub Lite reservations",
	}
	g.pubsubLiteResCreateCmd = &cobra.Command{
		Use: "create RESERVATION", Short: "Create a Pub/Sub Lite reservation",
		Args: cobra.ExactArgs(1), RunE: g.runPSLResCreate,
	}
	g.pubsubLiteResDeleteCmd = &cobra.Command{
		Use: "delete RESERVATION", Short: "Delete a Pub/Sub Lite reservation",
		Args: cobra.ExactArgs(1), RunE: g.runPSLResDelete,
	}
	g.pubsubLiteResDescribeCmd = &cobra.Command{
		Use: "describe RESERVATION", Short: "Describe a Pub/Sub Lite reservation",
		Args: cobra.ExactArgs(1), RunE: g.runPSLResDescribe,
	}
	g.pubsubLiteResListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub Lite reservations in a location",
		Args: cobra.NoArgs, RunE: g.runPSLResList,
	}
	g.pubsubLiteResListTopicsCmd = &cobra.Command{
		Use: "list-topics RESERVATION", Short: "List Pub/Sub Lite topics attached to a reservation",
		Args: cobra.ExactArgs(1), RunE: g.runPSLResListTopics,
	}
	g.pubsubLiteResUpdateCmd = &cobra.Command{
		Use: "update RESERVATION", Short: "Update a Pub/Sub Lite reservation",
		Args: cobra.ExactArgs(1), RunE: g.runPSLResUpdate,
	}
	g.pubsubLiteSubsCmd = &cobra.Command{
		Use:   "lite-subscriptions",
		Short: "Manage Pub/Sub Lite subscriptions",
	}
	g.pubsubLiteSubAckUpToCmd = &cobra.Command{
		Use: "ack-up-to SUBSCRIPTION", Short: "Commit the cursor for a subscription partition",
		Args: cobra.ExactArgs(1), RunE: g.runPSLSubAckUpTo,
	}
	g.pubsubLiteSubCreateCmd = &cobra.Command{
		Use: "create SUBSCRIPTION", Short: "Create a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSLSubCreate,
	}
	g.pubsubLiteSubDeleteCmd = &cobra.Command{
		Use: "delete SUBSCRIPTION", Short: "Delete a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSLSubDelete,
	}
	g.pubsubLiteSubDescribeCmd = &cobra.Command{
		Use: "describe SUBSCRIPTION", Short: "Describe a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSLSubDescribe,
	}
	g.pubsubLiteSubListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub Lite subscriptions in a location",
		Args: cobra.NoArgs, RunE: g.runPSLSubList,
	}
	g.pubsubLiteSubSeekCmd = &cobra.Command{
		Use: "seek SUBSCRIPTION", Short: "Seek a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSLSubSeek,
	}
	g.pubsubLiteSubSubscribeCmd = &cobra.Command{
		Use: "subscribe SUBSCRIPTION", Short: "Subscribe to a Pub/Sub Lite subscription (unsupported)",
		Args: cobra.ExactArgs(1), RunE: runPSLSubSubscribe,
	}
	g.pubsubLiteSubUpdateCmd = &cobra.Command{
		Use: "update SUBSCRIPTION", Short: "Update a Pub/Sub Lite subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSLSubUpdate,
	}
	g.pubsubLiteTopicsCmd = &cobra.Command{
		Use:   "lite-topics",
		Short: "Manage Pub/Sub Lite topics",
	}
	g.pubsubLiteTopicCreateCmd = &cobra.Command{
		Use: "create TOPIC", Short: "Create a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSLTopicCreate,
	}
	g.pubsubLiteTopicDeleteCmd = &cobra.Command{
		Use: "delete TOPIC", Short: "Delete a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSLTopicDelete,
	}
	g.pubsubLiteTopicDescribeCmd = &cobra.Command{
		Use: "describe TOPIC", Short: "Describe a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSLTopicDescribe,
	}
	g.pubsubLiteTopicListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub Lite topics in a location",
		Args: cobra.NoArgs, RunE: g.runPSLTopicList,
	}
	g.pubsubLiteTopicListSubsCmd = &cobra.Command{
		Use: "list-subscriptions TOPIC", Short: "List subscriptions attached to a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSLTopicListSubs,
	}
	g.pubsubLiteTopicPublishCmd = &cobra.Command{
		Use: "publish TOPIC", Short: "Publish to a Pub/Sub Lite topic (unsupported)",
		Args: cobra.ExactArgs(1), RunE: runPSLTopicPublish,
	}
	g.pubsubLiteTopicUpdateCmd = &cobra.Command{
		Use: "update TOPIC", Short: "Update a Pub/Sub Lite topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSLTopicUpdate,
	}
	g.pubsubMessageTransformsCmd = &cobra.Command{
		Use:   "message-transforms",
		Short: "Manage Cloud Pub/Sub message transforms",
	}
	g.pubsubMTTestCmd = &cobra.Command{
		Use: "test", Short: "Test message transforms against a given message",
		Args: cobra.NoArgs, RunE: g.runPMTTest,
	}
	g.pubsubMTValidateCmd = &cobra.Command{
		Use: "validate", Short: "Validate a message transform",
		Args: cobra.NoArgs, RunE: g.runPMTValidate,
	}
	g.pubsubSchemasCmd = &cobra.Command{Use: "schemas", Short: "Manage Pub/Sub schemas"}
	g.psSchemaCreateCmd = &cobra.Command{
		Use: "create SCHEMA", Short: "Create a Pub/Sub schema",
		Args: cobra.ExactArgs(1), RunE: g.runPSSchemaCreate,
	}
	g.psSchemaDeleteCmd = &cobra.Command{
		Use: "delete SCHEMA", Short: "Delete a Pub/Sub schema",
		Args: cobra.ExactArgs(1), RunE: g.runPSSchemaDelete,
	}
	g.psSchemaDeleteRevCmd = &cobra.Command{
		Use: "delete-revision SCHEMA", Short: "Delete a schema revision (--revision required)",
		Args: cobra.ExactArgs(1), RunE: g.runPSSchemaDeleteRev,
	}
	g.psSchemaDescribeCmd = &cobra.Command{
		Use: "describe SCHEMA", Short: "Describe a Pub/Sub schema",
		Args: cobra.ExactArgs(1), RunE: g.runPSSchemaDescribe,
	}
	g.psSchemaListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub schemas",
		Args: cobra.NoArgs, RunE: g.runPSSchemaList,
	}
	g.psSchemaListRevsCmd = &cobra.Command{
		Use: "list-revisions SCHEMA", Short: "List revisions of a Pub/Sub schema",
		Args: cobra.ExactArgs(1), RunE: g.runPSSchemaListRevs,
	}
	g.psSchemaCommitCmd = &cobra.Command{
		Use: "commit SCHEMA", Short: "Commit a new revision of a schema",
		Args: cobra.ExactArgs(1), RunE: g.runPSSchemaCommit,
	}
	g.psSchemaRollbackCmd = &cobra.Command{
		Use: "rollback SCHEMA", Short: "Roll back a schema to a given revision",
		Args: cobra.ExactArgs(1), RunE: g.runPSSchemaRollback,
	}
	g.psSchemaValidateSchemaCmd = &cobra.Command{
		Use: "validate-schema", Short: "Validate a schema definition without persisting it",
		Args: cobra.NoArgs, RunE: g.runPSSchemaValidateSchema,
	}
	g.psSchemaValidateMessageCmd = &cobra.Command{
		Use: "validate-message SCHEMA", Short: "Validate a message against a schema",
		Args: cobra.ExactArgs(1), RunE: g.runPSSchemaValidateMessage,
	}
	g.pubsubSnapshotsCmd = &cobra.Command{Use: "snapshots", Short: "Manage Pub/Sub snapshots"}
	g.psSnapCreateCmd = &cobra.Command{
		Use: "create SNAPSHOT", Short: "Create a Pub/Sub snapshot",
		Args: cobra.ExactArgs(1), RunE: g.runPSSnapCreate,
	}
	g.psSnapDeleteCmd = &cobra.Command{
		Use: "delete SNAPSHOT", Short: "Delete a Pub/Sub snapshot",
		Args: cobra.ExactArgs(1), RunE: g.runPSSnapDelete,
	}
	g.psSnapDescribeCmd = &cobra.Command{
		Use: "describe SNAPSHOT", Short: "Describe a Pub/Sub snapshot",
		Args: cobra.ExactArgs(1), RunE: g.runPSSnapDescribe,
	}
	g.psSnapListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub snapshots",
		Args: cobra.NoArgs, RunE: g.runPSSnapList,
	}
	g.pubsubTopicsCmd = &cobra.Command{Use: "topics", Short: "Manage Pub/Sub topics"}
	g.psTopCreateCmd = &cobra.Command{
		Use: "create TOPIC", Short: "Create a Pub/Sub topic (optionally from a --config-file)",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopCreate,
	}
	g.psTopDeleteCmd = &cobra.Command{
		Use: "delete TOPIC", Short: "Delete a Pub/Sub topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopDelete,
	}
	g.psTopDescribeCmd = &cobra.Command{
		Use: "describe TOPIC", Short: "Describe a Pub/Sub topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopDescribe,
	}
	g.psTopListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub topics in the project",
		Args: cobra.NoArgs, RunE: g.runPSTopList,
	}
	g.psTopUpdateCmd = &cobra.Command{
		Use: "update TOPIC", Short: "Update a Pub/Sub topic from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopUpdate,
	}
	g.psTopPublishCmd = &cobra.Command{
		Use: "publish TOPIC", Short: "Publish a message to a Pub/Sub topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopPublish,
	}
	g.psTopListSubsCmd = &cobra.Command{
		Use: "list-subscriptions TOPIC", Short: "List subscriptions attached to a topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopListSubs,
	}
	g.psTopDetachSubCmd = &cobra.Command{
		Use: "detach-subscription SUBSCRIPTION", Short: "Detach a subscription from its topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopDetachSub,
	}
	g.psTopGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TOPIC", Short: "Get the IAM policy for a topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopGetIam,
	}
	g.psTopSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TOPIC POLICY_FILE", Short: "Replace the IAM policy for a topic",
		Args: cobra.ExactArgs(2), RunE: g.runPSTopSetIam,
	}
	g.psTopAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TOPIC", Short: "Add an IAM binding to a topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopAddIam,
	}
	g.psTopRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TOPIC", Short: "Remove an IAM binding from a topic",
		Args: cobra.ExactArgs(1), RunE: g.runPSTopRemoveIam,
	}
	g.pubsubSubscriptionsCmd = &cobra.Command{Use: "subscriptions", Short: "Manage Pub/Sub subscriptions"}
	g.psSubCreateCmd = &cobra.Command{
		Use: "create SUBSCRIPTION", Short: "Create a Pub/Sub subscription (optionally from a --config-file)",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubCreate,
	}
	g.psSubDeleteCmd = &cobra.Command{
		Use: "delete SUBSCRIPTION", Short: "Delete a Pub/Sub subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubDelete,
	}
	g.psSubDescribeCmd = &cobra.Command{
		Use: "describe SUBSCRIPTION", Short: "Describe a Pub/Sub subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubDescribe,
	}
	g.psSubListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub subscriptions in the project",
		Args: cobra.NoArgs, RunE: g.runPSSubList,
	}
	g.psSubUpdateCmd = &cobra.Command{
		Use: "update SUBSCRIPTION", Short: "Update a subscription from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubUpdate,
	}
	g.psSubAckCmd = &cobra.Command{
		Use: "ack SUBSCRIPTION ACK_ID [ACK_ID ...]", Short: "Acknowledge messages on a subscription",
		Args: cobra.MinimumNArgs(2), RunE: g.runPSSubAck,
	}
	g.psSubPullCmd = &cobra.Command{
		Use: "pull SUBSCRIPTION", Short: "Pull messages from a subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubPull,
	}
	g.psSubSeekCmd = &cobra.Command{
		Use: "seek SUBSCRIPTION", Short: "Seek a subscription to a time or snapshot",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubSeek,
	}
	g.psSubModifyAckCmd = &cobra.Command{
		Use:   "modify-message-ack-deadline SUBSCRIPTION ACK_ID [ACK_ID ...]",
		Short: "Modify the ack deadline for outstanding messages",
		Args:  cobra.MinimumNArgs(2), RunE: g.runPSSubModifyAck,
	}
	g.psSubModifyPushCmd = &cobra.Command{
		Use: "modify-push-config SUBSCRIPTION", Short: "Modify the push configuration for a subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubModifyPush,
	}
	g.psSubGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SUBSCRIPTION", Short: "Get the IAM policy for a subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubGetIam,
	}
	g.psSubSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SUBSCRIPTION POLICY_FILE", Short: "Replace the IAM policy for a subscription",
		Args: cobra.ExactArgs(2), RunE: g.runPSSubSetIam,
	}
	g.psSubAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SUBSCRIPTION", Short: "Add an IAM binding to a subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubAddIam,
	}
	g.psSubRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SUBSCRIPTION", Short: "Remove an IAM binding from a subscription",
		Args: cobra.ExactArgs(1), RunE: g.runPSSubRemoveIam,
	}
	g.initPubsub()
	g.initPubsubLiteOperations()
	g.initPubsubLiteReservations()
	g.initPubsubLiteSubscriptions()
	g.initPubsubLiteTopics()
	g.initPubsubMessageTransforms()
	g.initPubsubSchemasSnapshots()
	g.initPubsubTopicsSubs()
	return g
}
