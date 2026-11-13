package cmd

import "github.com/spf13/cobra"

// --- gcloud managed-kafka (#353) ---

func (s *managedKafkaCmds) initManagedKafka() {
	// All subgroups (acls, clusters, connect-clusters, connectors,
	// consumer-groups, operations, topics) are implemented in
	// managed_kafka_all.go.

}

// managedKafkaCmds holds the commands and flag values of one invocation's "managedKafka"
// command group, so invocations don't share state (#2070).
type managedKafkaCmds struct {
	inv *invocation

	managedKafkaCmd       *cobra.Command
	mkClustersCmd         *cobra.Command
	mkClusterCreateCmd    *cobra.Command
	mkClusterDeleteCmd    *cobra.Command
	mkClusterDescribeCmd  *cobra.Command
	mkClusterListCmd      *cobra.Command
	mkClusterUpdateCmd    *cobra.Command
	flagMKLocation        string
	flagMKConfigFile      string
	flagMKUpdateMask      string
	flagMKFormat          string
	flagMKAsync           bool
	flagMKCluster         string
	flagMKConnectClust    string
	flagMKBrokerDiskGiB   string
	mkAclsCmd             *cobra.Command
	mkAclCreateCmd        *cobra.Command
	mkAclDeleteCmd        *cobra.Command
	mkAclDescribeCmd      *cobra.Command
	mkAclListCmd          *cobra.Command
	mkAclUpdateCmd        *cobra.Command
	mkAclAddEntryCmd      *cobra.Command
	mkAclRemoveEntryCmd   *cobra.Command
	mkTopicsCmd           *cobra.Command
	mkTopicCreateCmd      *cobra.Command
	mkTopicDeleteCmd      *cobra.Command
	mkTopicDescribeCmd    *cobra.Command
	mkTopicListCmd        *cobra.Command
	mkTopicUpdateCmd      *cobra.Command
	mkConsumerGroupsCmd   *cobra.Command
	mkCGDeleteCmd         *cobra.Command
	mkCGDescribeCmd       *cobra.Command
	mkCGListCmd           *cobra.Command
	mkCGUpdateCmd         *cobra.Command
	mkConnectClustersCmd  *cobra.Command
	mkCCCreateCmd         *cobra.Command
	mkCCDeleteCmd         *cobra.Command
	mkCCDescribeCmd       *cobra.Command
	mkCCListCmd           *cobra.Command
	mkCCUpdateCmd         *cobra.Command
	mkConnectorsCmd       *cobra.Command
	mkConnCreateCmd       *cobra.Command
	mkConnDeleteCmd       *cobra.Command
	mkConnDescribeCmd     *cobra.Command
	mkConnListCmd         *cobra.Command
	mkConnUpdateCmd       *cobra.Command
	mkConnPauseCmd        *cobra.Command
	mkConnResumeCmd       *cobra.Command
	mkConnRestartCmd      *cobra.Command
	mkConnStopCmd         *cobra.Command
	mkOperationsCmd       *cobra.Command
	mkOpDescribeCmd       *cobra.Command
	mkOpListCmd           *cobra.Command
	flagMKPublicCluster   bool
	flagMKNoPublicCluster bool
	flagMKAllowedRanges   []string
}

// newManagedKafkaCmd returns the "managedKafka" command group for inv.
func newManagedKafkaCmd(inv *invocation) *cobra.Command {
	return newManagedKafkaCmds(inv).managedKafkaCmd
}

// newManagedKafkaCmds builds the "managedKafka" command group for inv and returns its state.
func newManagedKafkaCmds(inv *invocation) *managedKafkaCmds {
	s := &managedKafkaCmds{inv: inv}
	s.managedKafkaCmd = &cobra.Command{Use: "managed-kafka", Short: "Manage Managed Service for Apache Kafka"}
	s.mkClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage Managed Kafka clusters"}
	s.mkClusterCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKClusterCreate,
	}
	s.mkClusterDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a cluster",
		Args: cobra.ExactArgs(1), RunE: s.runMKClusterDelete,
	}
	s.mkClusterDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a cluster",
		Args: cobra.ExactArgs(1), RunE: s.runMKClusterDescribe,
	}
	s.mkClusterListCmd = &cobra.Command{
		Use: "list", Short: "List clusters in a location",
		Args: cobra.NoArgs, RunE: s.runMKClusterList,
	}
	s.mkClusterUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKClusterUpdate,
	}
	s.mkAclsCmd = &cobra.Command{Use: "acls", Short: "Manage Kafka ACLs"}
	s.mkAclCreateCmd = &cobra.Command{
		Use: "create ACL", Short: "Create an ACL from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKAclCreate,
	}
	s.mkAclDeleteCmd = &cobra.Command{
		Use: "delete ACL", Short: "Delete an ACL",
		Args: cobra.ExactArgs(1), RunE: s.runMKAclDelete,
	}
	s.mkAclDescribeCmd = &cobra.Command{
		Use: "describe ACL", Short: "Describe an ACL",
		Args: cobra.ExactArgs(1), RunE: s.runMKAclDescribe,
	}
	s.mkAclListCmd = &cobra.Command{
		Use: "list", Short: "List ACLs in a cluster",
		Args: cobra.NoArgs, RunE: s.runMKAclList,
	}
	s.mkAclUpdateCmd = &cobra.Command{
		Use: "update ACL", Short: "Update an ACL from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKAclUpdate,
	}
	s.mkAclAddEntryCmd = &cobra.Command{
		Use: "add-acl-entry ACL", Short: "Add an entry to an ACL from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKAclAddEntry,
	}
	s.mkAclRemoveEntryCmd = &cobra.Command{
		Use: "remove-acl-entry ACL", Short: "Remove an entry from an ACL from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKAclRemoveEntry,
	}
	s.mkTopicsCmd = &cobra.Command{Use: "topics", Short: "Manage Kafka topics"}
	s.mkTopicCreateCmd = &cobra.Command{
		Use: "create TOPIC", Short: "Create a topic from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKTopicCreate,
	}
	s.mkTopicDeleteCmd = &cobra.Command{
		Use: "delete TOPIC", Short: "Delete a topic",
		Args: cobra.ExactArgs(1), RunE: s.runMKTopicDelete,
	}
	s.mkTopicDescribeCmd = &cobra.Command{
		Use: "describe TOPIC", Short: "Describe a topic",
		Args: cobra.ExactArgs(1), RunE: s.runMKTopicDescribe,
	}
	s.mkTopicListCmd = &cobra.Command{
		Use: "list", Short: "List topics in a cluster",
		Args: cobra.NoArgs, RunE: s.runMKTopicList,
	}
	s.mkTopicUpdateCmd = &cobra.Command{
		Use: "update TOPIC", Short: "Update a topic from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKTopicUpdate,
	}
	s.mkConsumerGroupsCmd = &cobra.Command{Use: "consumer-groups", Short: "Manage Kafka consumer groups"}
	s.mkCGDeleteCmd = &cobra.Command{
		Use: "delete CONSUMER_GROUP", Short: "Delete a consumer group",
		Args: cobra.ExactArgs(1), RunE: s.runMKCGDelete,
	}
	s.mkCGDescribeCmd = &cobra.Command{
		Use: "describe CONSUMER_GROUP", Short: "Describe a consumer group",
		Args: cobra.ExactArgs(1), RunE: s.runMKCGDescribe,
	}
	s.mkCGListCmd = &cobra.Command{
		Use: "list", Short: "List consumer groups in a cluster",
		Args: cobra.NoArgs, RunE: s.runMKCGList,
	}
	s.mkCGUpdateCmd = &cobra.Command{
		Use: "update CONSUMER_GROUP", Short: "Update a consumer group from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKCGUpdate,
	}
	s.mkConnectClustersCmd = &cobra.Command{Use: "connect-clusters", Short: "Manage Kafka Connect clusters"}
	s.mkCCCreateCmd = &cobra.Command{
		Use: "create CONNECT_CLUSTER", Short: "Create a Connect cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKCCCreate,
	}
	s.mkCCDeleteCmd = &cobra.Command{
		Use: "delete CONNECT_CLUSTER", Short: "Delete a Connect cluster",
		Args: cobra.ExactArgs(1), RunE: s.runMKCCDelete,
	}
	s.mkCCDescribeCmd = &cobra.Command{
		Use: "describe CONNECT_CLUSTER", Short: "Describe a Connect cluster",
		Args: cobra.ExactArgs(1), RunE: s.runMKCCDescribe,
	}
	s.mkCCListCmd = &cobra.Command{
		Use: "list", Short: "List Connect clusters in a location",
		Args: cobra.NoArgs, RunE: s.runMKCCList,
	}
	s.mkCCUpdateCmd = &cobra.Command{
		Use: "update CONNECT_CLUSTER", Short: "Update a Connect cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKCCUpdate,
	}
	s.mkConnectorsCmd = &cobra.Command{Use: "connectors", Short: "Manage Kafka Connect connectors"}
	s.mkConnCreateCmd = &cobra.Command{
		Use: "create CONNECTOR", Short: "Create a connector from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKConnCreate,
	}
	s.mkConnDeleteCmd = &cobra.Command{
		Use: "delete CONNECTOR", Short: "Delete a connector",
		Args: cobra.ExactArgs(1), RunE: s.runMKConnDelete,
	}
	s.mkConnDescribeCmd = &cobra.Command{
		Use: "describe CONNECTOR", Short: "Describe a connector",
		Args: cobra.ExactArgs(1), RunE: s.runMKConnDescribe,
	}
	s.mkConnListCmd = &cobra.Command{
		Use: "list", Short: "List connectors in a Connect cluster",
		Args: cobra.NoArgs, RunE: s.runMKConnList,
	}
	s.mkConnUpdateCmd = &cobra.Command{
		Use: "update CONNECTOR", Short: "Update a connector from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMKConnUpdate,
	}
	s.mkConnPauseCmd = &cobra.Command{
		Use: "pause CONNECTOR", Short: "Pause a connector",
		Args: cobra.ExactArgs(1), RunE: s.runMKConnAction("Pause"),
	}
	s.mkConnResumeCmd = &cobra.Command{
		Use: "resume CONNECTOR", Short: "Resume a connector",
		Args: cobra.ExactArgs(1), RunE: s.runMKConnAction("Resume"),
	}
	s.mkConnRestartCmd = &cobra.Command{
		Use: "restart CONNECTOR", Short: "Restart a connector",
		Args: cobra.ExactArgs(1), RunE: s.runMKConnAction("Restart"),
	}
	s.mkConnStopCmd = &cobra.Command{
		Use: "stop CONNECTOR", Short: "Stop a connector",
		Args: cobra.ExactArgs(1), RunE: s.runMKConnAction("Stop"),
	}
	s.mkOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Managed Kafka operations"}
	s.mkOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runMKOpDescribe,
	}
	s.mkOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: s.runMKOpList,
	}
	s.initManagedKafka()
	s.initManagedKafkaAll()
	s.initManagedKafkaAll2()
	s.initManagedKafkaAll3()
	s.initManagedKafkaAll4()
	s.initManagedKafkaAll5()
	s.initManagedKafkaAll6()
	s.initManagedKafkaAll7()
	return s
}
