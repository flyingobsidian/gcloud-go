package cmd

import "github.com/spf13/cobra"

// --- gcloud managed-kafka (#353) ---

var managedKafkaCmd = &cobra.Command{Use: "managed-kafka", Short: "Manage Managed Service for Apache Kafka (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(managedKafkaCmd, "acls", "Manage ACLs", append(crud, "add-acl-entry", "remove-acl-entry")...)
	registerStubGroup(managedKafkaCmd, "clusters", "Manage clusters", crud...)
	registerStubGroup(managedKafkaCmd, "connect-clusters", "Manage connect clusters", crud...)
	registerStubGroup(managedKafkaCmd, "connectors", "Manage connectors", append(crud, "pause", "resume", "restart", "stop")...)
	registerStubGroup(managedKafkaCmd, "consumer-groups", "Manage consumer groups", crud...)
	registerStubGroup(managedKafkaCmd, "operations", "Manage operations", "cancel", "describe", "list")
	registerStubGroup(managedKafkaCmd, "topics", "Manage topics", crud...)
	rootCmd.AddCommand(managedKafkaCmd)
}
