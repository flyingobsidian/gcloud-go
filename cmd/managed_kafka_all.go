package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	managedkafka "google.golang.org/api/managedkafka/v1"
)

// --- shared ---

func mkLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func mkClusterParent(project, location, cluster string) string {
	return fmt.Sprintf("%s/clusters/%s", mkLocationParent(project, location), cluster)
}

func mkConnectClusterParent(project, location, cluster string) string {
	return fmt.Sprintf("%s/connectClusters/%s", mkLocationParent(project, location), cluster)
}

func mkChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func mkWaitOp(ctx context.Context, svc *managedkafka.Service, op *managedkafka.Operation) (*managedkafka.Operation, error) {
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

func mkFinishOp(ctx context.Context, svc *managedkafka.Service, op *managedkafka.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := mkWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// --- clusters ---

var mkClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage Managed Kafka clusters"}

var (
	mkClusterCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKClusterCreate,
	}
	mkClusterDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a cluster",
		Args: cobra.ExactArgs(1), RunE: runMKClusterDelete,
	}
	mkClusterDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a cluster",
		Args: cobra.ExactArgs(1), RunE: runMKClusterDescribe,
	}
	mkClusterListCmd = &cobra.Command{
		Use: "list", Short: "List clusters in a location",
		Args: cobra.NoArgs, RunE: runMKClusterList,
	}
	mkClusterUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKClusterUpdate,
	}
)

var (
	flagMKLocation      string
	flagMKConfigFile    string
	flagMKUpdateMask    string
	flagMKFormat        string
	flagMKAsync         bool
	flagMKCluster       string
	flagMKConnectClust  string
	flagMKAclOp         string
)

func init() {
	all := []*cobra.Command{mkClusterCreateCmd, mkClusterDeleteCmd, mkClusterDescribeCmd, mkClusterListCmd, mkClusterUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMKLocation, "location", "", "Location containing the cluster (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{mkClusterCreateCmd, mkClusterUpdateCmd} {
		c.Flags().StringVar(&flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Cluster message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	mkClusterUpdateCmd.Flags().StringVar(&flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{mkClusterCreateCmd, mkClusterDeleteCmd, mkClusterUpdateCmd} {
		c.Flags().BoolVar(&flagMKAsync, "async", false, "Return the long-running operation without waiting")
	}
	mkClusterDescribeCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")
	mkClusterListCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")

	mkClustersCmd.AddCommand(all...)
	managedKafkaCmd.AddCommand(mkClustersCmd)
}

func runMKClusterCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cluster := &managedkafka.Cluster{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, cluster); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Create(mkLocationParent(project, flagMKLocation), cluster).
		ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	return mkFinishOp(ctx, svc, op, "Create cluster", args[0], flagMKAsync)
}

func runMKClusterDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(mkClusterParent(project, flagMKLocation, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return mkFinishOp(ctx, svc, op, "Delete cluster", args[0], flagMKAsync)
}

func runMKClusterDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Get(mkClusterParent(project, flagMKLocation, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormatted(got, flagMKFormat)
}

func runMKClusterList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(mkLocationParent(project, flagMKLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if flagMKFormat != "" {
		return emitFormatted(resp.Clusters, flagMKFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, c := range resp.Clusters {
		fmt.Printf("%-40s %s\n", path.Base(c.Name), c.State)
	}
	return nil
}

func runMKClusterUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cluster := &managedkafka.Cluster{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, cluster); err != nil {
		return err
	}
	mask := flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cluster))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Patch(mkClusterParent(project, flagMKLocation, args[0]), cluster).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	return mkFinishOp(ctx, svc, op, "Update cluster", args[0], flagMKAsync)
}

// --- acls ---

var mkAclsCmd = &cobra.Command{Use: "acls", Short: "Manage Kafka ACLs"}

var (
	mkAclCreateCmd = &cobra.Command{
		Use: "create ACL", Short: "Create an ACL from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKAclCreate,
	}
	mkAclDeleteCmd = &cobra.Command{
		Use: "delete ACL", Short: "Delete an ACL",
		Args: cobra.ExactArgs(1), RunE: runMKAclDelete,
	}
	mkAclDescribeCmd = &cobra.Command{
		Use: "describe ACL", Short: "Describe an ACL",
		Args: cobra.ExactArgs(1), RunE: runMKAclDescribe,
	}
	mkAclListCmd = &cobra.Command{
		Use: "list", Short: "List ACLs in a cluster",
		Args: cobra.NoArgs, RunE: runMKAclList,
	}
	mkAclUpdateCmd = &cobra.Command{
		Use: "update ACL", Short: "Update an ACL from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKAclUpdate,
	}
	mkAclAddEntryCmd = &cobra.Command{
		Use: "add-acl-entry ACL", Short: "Add an entry to an ACL from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKAclAddEntry,
	}
	mkAclRemoveEntryCmd = &cobra.Command{
		Use: "remove-acl-entry ACL", Short: "Remove an entry from an ACL from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKAclRemoveEntry,
	}
)

func init() {
	all := []*cobra.Command{mkAclCreateCmd, mkAclDeleteCmd, mkAclDescribeCmd, mkAclListCmd, mkAclUpdateCmd, mkAclAddEntryCmd, mkAclRemoveEntryCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMKLocation, "location", "", "Location containing the cluster (required)")
		c.Flags().StringVar(&flagMKCluster, "cluster", "", "Cluster containing the ACL (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("cluster")
	}
	for _, c := range []*cobra.Command{mkAclCreateCmd, mkAclUpdateCmd, mkAclAddEntryCmd, mkAclRemoveEntryCmd} {
		c.Flags().StringVar(&flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Acl or AclEntry message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	mkAclUpdateCmd.Flags().StringVar(&flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	mkAclDescribeCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")
	mkAclListCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")

	mkAclsCmd.AddCommand(all...)
	managedKafkaCmd.AddCommand(mkAclsCmd)
}

func mkAclName(id, project, location, cluster string) string {
	return mkChild("acls", id, mkClusterParent(project, location, cluster))
}

func runMKAclCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	acl := &managedkafka.Acl{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, acl); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.Create(mkClusterParent(project, flagMKLocation, flagMKCluster), acl).
		AclId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ACL: %w", err)
	}
	return emitFormatted(got, "")
}

func runMKAclDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Clusters.Acls.Delete(mkAclName(args[0], project, flagMKLocation, flagMKCluster)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting ACL: %w", err)
	}
	fmt.Printf("Deleted ACL [%s].\n", args[0])
	return nil
}

func runMKAclDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.Get(mkAclName(args[0], project, flagMKLocation, flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ACL: %w", err)
	}
	return emitFormatted(got, flagMKFormat)
}

func runMKAclList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.Acls.List(mkClusterParent(project, flagMKLocation, flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing ACLs: %w", err)
	}
	if flagMKFormat != "" {
		return emitFormatted(resp.Acls, flagMKFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "RESOURCE_TYPE")
	for _, a := range resp.Acls {
		fmt.Printf("%-40s %s\n", path.Base(a.Name), a.ResourceType)
	}
	return nil
}

func runMKAclUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	acl := &managedkafka.Acl{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, acl); err != nil {
		return err
	}
	mask := flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(acl))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.Patch(mkAclName(args[0], project, flagMKLocation, flagMKCluster), acl).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating ACL: %w", err)
	}
	return emitFormatted(got, "")
}

func runMKAclAddEntry(cmd *cobra.Command, args []string) error {
	entry := &managedkafka.AclEntry{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, entry); err != nil {
		return err
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.AddAclEntry(mkAclName(args[0], project, flagMKLocation, flagMKCluster), entry).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding ACL entry: %w", err)
	}
	return emitFormatted(got, "")
}

func runMKAclRemoveEntry(cmd *cobra.Command, args []string) error {
	entry := &managedkafka.AclEntry{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, entry); err != nil {
		return err
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.RemoveAclEntry(mkAclName(args[0], project, flagMKLocation, flagMKCluster), entry).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing ACL entry: %w", err)
	}
	return emitFormatted(got, "")
}

// --- topics ---

var mkTopicsCmd = &cobra.Command{Use: "topics", Short: "Manage Kafka topics"}

var (
	mkTopicCreateCmd = &cobra.Command{
		Use: "create TOPIC", Short: "Create a topic from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKTopicCreate,
	}
	mkTopicDeleteCmd = &cobra.Command{
		Use: "delete TOPIC", Short: "Delete a topic",
		Args: cobra.ExactArgs(1), RunE: runMKTopicDelete,
	}
	mkTopicDescribeCmd = &cobra.Command{
		Use: "describe TOPIC", Short: "Describe a topic",
		Args: cobra.ExactArgs(1), RunE: runMKTopicDescribe,
	}
	mkTopicListCmd = &cobra.Command{
		Use: "list", Short: "List topics in a cluster",
		Args: cobra.NoArgs, RunE: runMKTopicList,
	}
	mkTopicUpdateCmd = &cobra.Command{
		Use: "update TOPIC", Short: "Update a topic from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKTopicUpdate,
	}
)

func init() {
	all := []*cobra.Command{mkTopicCreateCmd, mkTopicDeleteCmd, mkTopicDescribeCmd, mkTopicListCmd, mkTopicUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMKLocation, "location", "", "Location containing the cluster (required)")
		c.Flags().StringVar(&flagMKCluster, "cluster", "", "Cluster containing the topic (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("cluster")
	}
	for _, c := range []*cobra.Command{mkTopicCreateCmd, mkTopicUpdateCmd} {
		c.Flags().StringVar(&flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Topic message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	mkTopicUpdateCmd.Flags().StringVar(&flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	mkTopicDescribeCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")
	mkTopicListCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")

	mkTopicsCmd.AddCommand(all...)
	managedKafkaCmd.AddCommand(mkTopicsCmd)
}

func mkTopicName(id, project, location, cluster string) string {
	return mkChild("topics", id, mkClusterParent(project, location, cluster))
}

func runMKTopicCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	topic := &managedkafka.Topic{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, topic); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Topics.Create(mkClusterParent(project, flagMKLocation, flagMKCluster), topic).
		TopicId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating topic: %w", err)
	}
	return emitFormatted(got, "")
}

func runMKTopicDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Clusters.Topics.Delete(mkTopicName(args[0], project, flagMKLocation, flagMKCluster)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting topic: %w", err)
	}
	fmt.Printf("Deleted topic [%s].\n", args[0])
	return nil
}

func runMKTopicDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Topics.Get(mkTopicName(args[0], project, flagMKLocation, flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing topic: %w", err)
	}
	return emitFormatted(got, flagMKFormat)
}

func runMKTopicList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.Topics.List(mkClusterParent(project, flagMKLocation, flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing topics: %w", err)
	}
	if flagMKFormat != "" {
		return emitFormatted(resp.Topics, flagMKFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "PARTITION_COUNT")
	for _, t := range resp.Topics {
		fmt.Printf("%-40s %d\n", path.Base(t.Name), t.PartitionCount)
	}
	return nil
}

func runMKTopicUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	topic := &managedkafka.Topic{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, topic); err != nil {
		return err
	}
	mask := flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(topic))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Topics.Patch(mkTopicName(args[0], project, flagMKLocation, flagMKCluster), topic).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating topic: %w", err)
	}
	return emitFormatted(got, "")
}

// --- consumer-groups ---

var mkConsumerGroupsCmd = &cobra.Command{Use: "consumer-groups", Short: "Manage Kafka consumer groups"}

var (
	mkCGDeleteCmd = &cobra.Command{
		Use: "delete CONSUMER_GROUP", Short: "Delete a consumer group",
		Args: cobra.ExactArgs(1), RunE: runMKCGDelete,
	}
	mkCGDescribeCmd = &cobra.Command{
		Use: "describe CONSUMER_GROUP", Short: "Describe a consumer group",
		Args: cobra.ExactArgs(1), RunE: runMKCGDescribe,
	}
	mkCGListCmd = &cobra.Command{
		Use: "list", Short: "List consumer groups in a cluster",
		Args: cobra.NoArgs, RunE: runMKCGList,
	}
	mkCGUpdateCmd = &cobra.Command{
		Use: "update CONSUMER_GROUP", Short: "Update a consumer group from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKCGUpdate,
	}
)

func init() {
	all := []*cobra.Command{mkCGDeleteCmd, mkCGDescribeCmd, mkCGListCmd, mkCGUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMKLocation, "location", "", "Location containing the cluster (required)")
		c.Flags().StringVar(&flagMKCluster, "cluster", "", "Cluster containing the consumer group (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("cluster")
	}
	mkCGUpdateCmd.Flags().StringVar(&flagMKConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ConsumerGroup message body (required)")
	_ = mkCGUpdateCmd.MarkFlagRequired("config-file")
	mkCGUpdateCmd.Flags().StringVar(&flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	mkCGDescribeCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")
	mkCGListCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")

	mkConsumerGroupsCmd.AddCommand(all...)
	managedKafkaCmd.AddCommand(mkConsumerGroupsCmd)
}

func mkCGName(id, project, location, cluster string) string {
	return mkChild("consumerGroups", id, mkClusterParent(project, location, cluster))
}

func runMKCGDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Clusters.ConsumerGroups.Delete(mkCGName(args[0], project, flagMKLocation, flagMKCluster)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting consumer group: %w", err)
	}
	fmt.Printf("Deleted consumer group [%s].\n", args[0])
	return nil
}

func runMKCGDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.ConsumerGroups.Get(mkCGName(args[0], project, flagMKLocation, flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing consumer group: %w", err)
	}
	return emitFormatted(got, flagMKFormat)
}

func runMKCGList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.ConsumerGroups.List(mkClusterParent(project, flagMKLocation, flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing consumer groups: %w", err)
	}
	if flagMKFormat != "" {
		return emitFormatted(resp.ConsumerGroups, flagMKFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, g := range resp.ConsumerGroups {
		fmt.Println(path.Base(g.Name))
	}
	return nil
}

func runMKCGUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	g := &managedkafka.ConsumerGroup{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, g); err != nil {
		return err
	}
	mask := flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(g))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.ConsumerGroups.Patch(mkCGName(args[0], project, flagMKLocation, flagMKCluster), g).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating consumer group: %w", err)
	}
	return emitFormatted(got, "")
}

// --- connect-clusters ---

var mkConnectClustersCmd = &cobra.Command{Use: "connect-clusters", Short: "Manage Kafka Connect clusters"}

var (
	mkCCCreateCmd = &cobra.Command{
		Use: "create CONNECT_CLUSTER", Short: "Create a Connect cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKCCCreate,
	}
	mkCCDeleteCmd = &cobra.Command{
		Use: "delete CONNECT_CLUSTER", Short: "Delete a Connect cluster",
		Args: cobra.ExactArgs(1), RunE: runMKCCDelete,
	}
	mkCCDescribeCmd = &cobra.Command{
		Use: "describe CONNECT_CLUSTER", Short: "Describe a Connect cluster",
		Args: cobra.ExactArgs(1), RunE: runMKCCDescribe,
	}
	mkCCListCmd = &cobra.Command{
		Use: "list", Short: "List Connect clusters in a location",
		Args: cobra.NoArgs, RunE: runMKCCList,
	}
	mkCCUpdateCmd = &cobra.Command{
		Use: "update CONNECT_CLUSTER", Short: "Update a Connect cluster from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKCCUpdate,
	}
)

func init() {
	all := []*cobra.Command{mkCCCreateCmd, mkCCDeleteCmd, mkCCDescribeCmd, mkCCListCmd, mkCCUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMKLocation, "location", "", "Location containing the Connect cluster (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{mkCCCreateCmd, mkCCUpdateCmd} {
		c.Flags().StringVar(&flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the ConnectCluster message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	mkCCUpdateCmd.Flags().StringVar(&flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{mkCCCreateCmd, mkCCDeleteCmd, mkCCUpdateCmd} {
		c.Flags().BoolVar(&flagMKAsync, "async", false, "Return the long-running operation without waiting")
	}
	mkCCDescribeCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")
	mkCCListCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")

	mkConnectClustersCmd.AddCommand(all...)
	managedKafkaCmd.AddCommand(mkConnectClustersCmd)
}

func runMKCCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cc := &managedkafka.ConnectCluster{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, cc); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectClusters.Create(mkLocationParent(project, flagMKLocation), cc).
		ConnectClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connect cluster: %w", err)
	}
	return mkFinishOp(ctx, svc, op, "Create connect cluster", args[0], flagMKAsync)
}

func runMKCCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectClusters.Delete(mkConnectClusterParent(project, flagMKLocation, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connect cluster: %w", err)
	}
	return mkFinishOp(ctx, svc, op, "Delete connect cluster", args[0], flagMKAsync)
}

func runMKCCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectClusters.Get(mkConnectClusterParent(project, flagMKLocation, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connect cluster: %w", err)
	}
	return emitFormatted(got, flagMKFormat)
}

func runMKCCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.ConnectClusters.List(mkLocationParent(project, flagMKLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing connect clusters: %w", err)
	}
	if flagMKFormat != "" {
		return emitFormatted(resp.ConnectClusters, flagMKFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, c := range resp.ConnectClusters {
		fmt.Printf("%-40s %s\n", path.Base(c.Name), c.State)
	}
	return nil
}

func runMKCCUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cc := &managedkafka.ConnectCluster{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, cc); err != nil {
		return err
	}
	mask := flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cc))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectClusters.Patch(mkConnectClusterParent(project, flagMKLocation, args[0]), cc).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connect cluster: %w", err)
	}
	return mkFinishOp(ctx, svc, op, "Update connect cluster", args[0], flagMKAsync)
}

// --- connectors ---

var mkConnectorsCmd = &cobra.Command{Use: "connectors", Short: "Manage Kafka Connect connectors"}

var (
	mkConnCreateCmd = &cobra.Command{
		Use: "create CONNECTOR", Short: "Create a connector from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKConnCreate,
	}
	mkConnDeleteCmd = &cobra.Command{
		Use: "delete CONNECTOR", Short: "Delete a connector",
		Args: cobra.ExactArgs(1), RunE: runMKConnDelete,
	}
	mkConnDescribeCmd = &cobra.Command{
		Use: "describe CONNECTOR", Short: "Describe a connector",
		Args: cobra.ExactArgs(1), RunE: runMKConnDescribe,
	}
	mkConnListCmd = &cobra.Command{
		Use: "list", Short: "List connectors in a Connect cluster",
		Args: cobra.NoArgs, RunE: runMKConnList,
	}
	mkConnUpdateCmd = &cobra.Command{
		Use: "update CONNECTOR", Short: "Update a connector from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMKConnUpdate,
	}
	mkConnPauseCmd = &cobra.Command{
		Use: "pause CONNECTOR", Short: "Pause a connector",
		Args: cobra.ExactArgs(1), RunE: runMKConnAction("Pause"),
	}
	mkConnResumeCmd = &cobra.Command{
		Use: "resume CONNECTOR", Short: "Resume a connector",
		Args: cobra.ExactArgs(1), RunE: runMKConnAction("Resume"),
	}
	mkConnRestartCmd = &cobra.Command{
		Use: "restart CONNECTOR", Short: "Restart a connector",
		Args: cobra.ExactArgs(1), RunE: runMKConnAction("Restart"),
	}
	mkConnStopCmd = &cobra.Command{
		Use: "stop CONNECTOR", Short: "Stop a connector",
		Args: cobra.ExactArgs(1), RunE: runMKConnAction("Stop"),
	}
)

func init() {
	all := []*cobra.Command{mkConnCreateCmd, mkConnDeleteCmd, mkConnDescribeCmd, mkConnListCmd, mkConnUpdateCmd,
		mkConnPauseCmd, mkConnResumeCmd, mkConnRestartCmd, mkConnStopCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMKLocation, "location", "", "Location containing the Connect cluster (required)")
		c.Flags().StringVar(&flagMKConnectClust, "connect-cluster", "", "Connect cluster containing the connector (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("connect-cluster")
	}
	for _, c := range []*cobra.Command{mkConnCreateCmd, mkConnUpdateCmd} {
		c.Flags().StringVar(&flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Connector message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	mkConnUpdateCmd.Flags().StringVar(&flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	mkConnDescribeCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")
	mkConnListCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")

	mkConnectorsCmd.AddCommand(all...)
	managedKafkaCmd.AddCommand(mkConnectorsCmd)
}

func mkConnName(id, project, location, connectCluster string) string {
	return mkChild("connectors", id, mkConnectClusterParent(project, location, connectCluster))
}

func runMKConnCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	conn := &managedkafka.Connector{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, conn); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectClusters.Connectors.Create(mkConnectClusterParent(project, flagMKLocation, flagMKConnectClust), conn).
		ConnectorId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connector: %w", err)
	}
	return emitFormatted(got, "")
}

func runMKConnDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ConnectClusters.Connectors.Delete(mkConnName(args[0], project, flagMKLocation, flagMKConnectClust)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting connector: %w", err)
	}
	fmt.Printf("Deleted connector [%s].\n", args[0])
	return nil
}

func runMKConnDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectClusters.Connectors.Get(mkConnName(args[0], project, flagMKLocation, flagMKConnectClust)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connector: %w", err)
	}
	return emitFormatted(got, flagMKFormat)
}

func runMKConnList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.ConnectClusters.Connectors.List(mkConnectClusterParent(project, flagMKLocation, flagMKConnectClust)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing connectors: %w", err)
	}
	if flagMKFormat != "" {
		return emitFormatted(resp.Connectors, flagMKFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, c := range resp.Connectors {
		fmt.Println(path.Base(c.Name))
	}
	return nil
}

func runMKConnUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	conn := &managedkafka.Connector{}
	if err := loadYAMLOrJSONInto(flagMKConfigFile, conn); err != nil {
		return err
	}
	mask := flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(conn))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectClusters.Connectors.Patch(mkConnName(args[0], project, flagMKLocation, flagMKConnectClust), conn).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connector: %w", err)
	}
	return emitFormatted(got, "")
}

func runMKConnAction(verb string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := resolveProject()
		if err != nil {
			return err
		}
		ctx := context.Background()
		svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
		if err != nil {
			return err
		}
		name := mkConnName(args[0], project, flagMKLocation, flagMKConnectClust)
		switch verb {
		case "Pause":
			if _, err := svc.Projects.Locations.ConnectClusters.Connectors.Pause(name, &managedkafka.PauseConnectorRequest{}).Context(ctx).Do(); err != nil {
				return fmt.Errorf("pausing connector: %w", err)
			}
		case "Resume":
			if _, err := svc.Projects.Locations.ConnectClusters.Connectors.Resume(name, &managedkafka.ResumeConnectorRequest{}).Context(ctx).Do(); err != nil {
				return fmt.Errorf("resuming connector: %w", err)
			}
		case "Restart":
			if _, err := svc.Projects.Locations.ConnectClusters.Connectors.Restart(name, &managedkafka.RestartConnectorRequest{}).Context(ctx).Do(); err != nil {
				return fmt.Errorf("restarting connector: %w", err)
			}
		case "Stop":
			if _, err := svc.Projects.Locations.ConnectClusters.Connectors.Stop(name, &managedkafka.StopConnectorRequest{}).Context(ctx).Do(); err != nil {
				return fmt.Errorf("stopping connector: %w", err)
			}
		}
		fmt.Printf("%s connector [%s].\n", verb, args[0])
		return nil
	}
}

// --- operations ---

var mkOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Managed Kafka operations"}

var (
	mkOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runMKOpDescribe,
	}
	mkOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: runMKOpList,
	}
)

func init() {
	for _, c := range []*cobra.Command{mkOpDescribeCmd, mkOpListCmd} {
		c.Flags().StringVar(&flagMKLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	mkOpDescribeCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")
	mkOpListCmd.Flags().StringVar(&flagMKFormat, "format", "", "Output format")

	mkOperationsCmd.AddCommand(mkOpDescribeCmd, mkOpListCmd)
	managedKafkaCmd.AddCommand(mkOperationsCmd)
}

func mkOpName(id, project, location string) string {
	return mkChild("operations", id, mkLocationParent(project, location))
}

func runMKOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(mkOpName(args[0], project, flagMKLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagMKFormat)
}

func runMKOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedKafkaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(mkLocationParent(project, flagMKLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagMKFormat != "" {
		return emitFormatted(resp.Operations, flagMKFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// silence linter about the placeholder var (used by ACL commands).
var _ = flagMKAclOp
