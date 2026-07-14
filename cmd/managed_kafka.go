package cmd

import "github.com/spf13/cobra"

// --- gcloud managed-kafka (#353) ---

var managedKafkaCmd = &cobra.Command{Use: "managed-kafka", Short: "Manage Managed Service for Apache Kafka"}

func init() {
	// All subgroups (acls, clusters, connect-clusters, connectors,
	// consumer-groups, operations, topics) are implemented in
	// managed_kafka_all.go.
	rootCmd.AddCommand(managedKafkaCmd)
}
