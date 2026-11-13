package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

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

func (s *managedKafkaCmds) mkFinishOp(ctx context.Context, svc *managedkafka.Service, op *managedkafka.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := mkWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// --- clusters ---

func (s *managedKafkaCmds) initManagedKafkaAll() {
	all := []*cobra.Command{s.mkClusterCreateCmd, s.mkClusterDeleteCmd, s.mkClusterDescribeCmd, s.mkClusterListCmd, s.mkClusterUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMKLocation, "location", "", "Location containing the cluster (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{s.mkClusterCreateCmd, s.mkClusterUpdateCmd} {
		c.Flags().StringVar(&s.flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Cluster message body (required unless --broker-disk-size-gib is set)")
		// GA in 582.0.0; the 576.0.0 helper accepts byte-unit suffixes (Ki/Mi/Gi/Ti and their
		// SI variants) and normalises them to GiB.
		c.Flags().StringVar(&s.flagMKBrokerDiskGiB, "broker-disk-size-gib", "",
			"Disk size for each broker in GiB (accepts byte-unit suffixes such as 100Gi or 2Ti)")
		c.Flags().BoolVar(&s.flagMKPublicCluster, "public-cluster", false,
			"Enable a public cluster. If disabled, public cluster config is cleared")
		c.Flags().BoolVar(&s.flagMKNoPublicCluster, "no-public-cluster", false, "Disable a public cluster")
		c.MarkFlagsMutuallyExclusive("public-cluster", "no-public-cluster")
		c.Flags().StringSliceVar(&s.flagMKAllowedRanges, "allowed-source-ip-ranges", nil,
			"Comma-separated IPv4 CIDR ranges allowed to connect to the public cluster (requires --public-cluster)")
	}
	s.mkClusterUpdateCmd.Flags().StringVar(&s.flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.mkClusterCreateCmd, s.mkClusterDeleteCmd, s.mkClusterUpdateCmd} {
		c.Flags().BoolVar(&s.flagMKAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.mkClusterDescribeCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")
	s.mkClusterListCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")

	s.mkClustersCmd.AddCommand(all...)
	s.managedKafkaCmd.AddCommand(s.mkClustersCmd)
}

func (s *managedKafkaCmds) runMKClusterCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if s.flagMKConfigFile == "" {
		return fmt.Errorf("--config-file is required")
	}
	cluster := &managedkafka.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, cluster); err != nil {
		return err
	}
	if err := s.applyBrokerDiskFlag(cluster); err != nil {
		return err
	}
	if err := mkApplyPublicClusterCreate(cluster, s.flagMKPublicCluster, s.flagMKAllowedRanges); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Create(mkLocationParent(project, s.flagMKLocation), cluster).
		ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	return s.mkFinishOp(ctx, svc, op, "Create cluster", args[0], s.flagMKAsync)
}

func (s *managedKafkaCmds) runMKClusterDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Delete(mkClusterParent(project, s.flagMKLocation, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	return s.mkFinishOp(ctx, svc, op, "Delete cluster", args[0], s.flagMKAsync)
}

func (s *managedKafkaCmds) runMKClusterDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Get(mkClusterParent(project, s.flagMKLocation, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMKFormat)
}

func (s *managedKafkaCmds) runMKClusterList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.List(mkLocationParent(project, s.flagMKLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}
	if s.flagMKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Clusters, s.flagMKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Clusters), "%-40s %s\n", "NAME", "STATE")
	for _, c := range resp.Clusters {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(c.Name), c.State)
	}
	return nil
}

func (s *managedKafkaCmds) runMKClusterUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	publicFlags := s.flagMKPublicCluster || s.flagMKNoPublicCluster || len(s.flagMKAllowedRanges) > 0
	if s.flagMKConfigFile == "" && s.flagMKBrokerDiskGiB == "" && !publicFlags {
		return fmt.Errorf("one of --config-file, --broker-disk-size-gib, --public-cluster, --no-public-cluster or --allowed-source-ip-ranges is required")
	}
	cluster := &managedkafka.Cluster{}
	if s.flagMKConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagMKConfigFile, cluster); err != nil {
			return err
		}
	}
	if err := s.applyBrokerDiskFlag(cluster); err != nil {
		return err
	}
	mask := s.flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cluster))
		if s.flagMKBrokerDiskGiB != "" && !strings.Contains(mask, "brokerCapacityConfig") {
			if mask != "" {
				mask += ","
			}
			mask += "brokerCapacityConfig.diskSizeGib"
		}
	}
	mask, err = mkApplyPublicClusterUpdate(cluster, s.flagMKPublicCluster, s.flagMKNoPublicCluster, s.flagMKAllowedRanges, mask)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Clusters.Patch(mkClusterParent(project, s.flagMKLocation, args[0]), cluster).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	return s.mkFinishOp(ctx, svc, op, "Update cluster", args[0], s.flagMKAsync)
}

// --- acls ---

func (s *managedKafkaCmds) initManagedKafkaAll2() {
	all := []*cobra.Command{s.mkAclCreateCmd, s.mkAclDeleteCmd, s.mkAclDescribeCmd, s.mkAclListCmd, s.mkAclUpdateCmd, s.mkAclAddEntryCmd, s.mkAclRemoveEntryCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMKLocation, "location", "", "Location containing the cluster (required)")
		c.Flags().StringVar(&s.flagMKCluster, "cluster", "", "Cluster containing the ACL (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("cluster")
	}
	for _, c := range []*cobra.Command{s.mkAclCreateCmd, s.mkAclUpdateCmd, s.mkAclAddEntryCmd, s.mkAclRemoveEntryCmd} {
		c.Flags().StringVar(&s.flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Acl or AclEntry message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.mkAclUpdateCmd.Flags().StringVar(&s.flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.mkAclDescribeCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")
	s.mkAclListCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")

	s.mkAclsCmd.AddCommand(all...)
	s.managedKafkaCmd.AddCommand(s.mkAclsCmd)
}

func mkAclName(id, project, location, cluster string) string {
	return mkChild("acls", id, mkClusterParent(project, location, cluster))
}

func (s *managedKafkaCmds) runMKAclCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	acl := &managedkafka.Acl{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, acl); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.Create(mkClusterParent(project, s.flagMKLocation, s.flagMKCluster), acl).
		AclId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ACL: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *managedKafkaCmds) runMKAclDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Clusters.Acls.Delete(mkAclName(args[0], project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting ACL: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted ACL [%s].\n", args[0])
	return nil
}

func (s *managedKafkaCmds) runMKAclDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.Get(mkAclName(args[0], project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ACL: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMKFormat)
}

func (s *managedKafkaCmds) runMKAclList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.Acls.List(mkClusterParent(project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing ACLs: %w", err)
	}
	if s.flagMKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Acls, s.flagMKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Acls), "%-40s %s\n", "NAME", "RESOURCE_TYPE")
	for _, a := range resp.Acls {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(a.Name), a.ResourceType)
	}
	return nil
}

func (s *managedKafkaCmds) runMKAclUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	acl := &managedkafka.Acl{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, acl); err != nil {
		return err
	}
	mask := s.flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(acl))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.Patch(mkAclName(args[0], project, s.flagMKLocation, s.flagMKCluster), acl).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating ACL: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *managedKafkaCmds) runMKAclAddEntry(cmd *cobra.Command, args []string) error {
	entry := &managedkafka.AclEntry{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, entry); err != nil {
		return err
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.AddAclEntry(mkAclName(args[0], project, s.flagMKLocation, s.flagMKCluster), entry).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding ACL entry: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *managedKafkaCmds) runMKAclRemoveEntry(cmd *cobra.Command, args []string) error {
	entry := &managedkafka.AclEntry{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, entry); err != nil {
		return err
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Acls.RemoveAclEntry(mkAclName(args[0], project, s.flagMKLocation, s.flagMKCluster), entry).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing ACL entry: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- topics ---

func (s *managedKafkaCmds) initManagedKafkaAll3() {
	all := []*cobra.Command{s.mkTopicCreateCmd, s.mkTopicDeleteCmd, s.mkTopicDescribeCmd, s.mkTopicListCmd, s.mkTopicUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMKLocation, "location", "", "Location containing the cluster (required)")
		c.Flags().StringVar(&s.flagMKCluster, "cluster", "", "Cluster containing the topic (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("cluster")
	}
	for _, c := range []*cobra.Command{s.mkTopicCreateCmd, s.mkTopicUpdateCmd} {
		c.Flags().StringVar(&s.flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Topic message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.mkTopicUpdateCmd.Flags().StringVar(&s.flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.mkTopicDescribeCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")
	s.mkTopicListCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")

	s.mkTopicsCmd.AddCommand(all...)
	s.managedKafkaCmd.AddCommand(s.mkTopicsCmd)
}

func mkTopicName(id, project, location, cluster string) string {
	return mkChild("topics", id, mkClusterParent(project, location, cluster))
}

func (s *managedKafkaCmds) runMKTopicCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	topic := &managedkafka.Topic{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, topic); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Topics.Create(mkClusterParent(project, s.flagMKLocation, s.flagMKCluster), topic).
		TopicId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating topic: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *managedKafkaCmds) runMKTopicDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Clusters.Topics.Delete(mkTopicName(args[0], project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting topic: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted topic [%s].\n", args[0])
	return nil
}

func (s *managedKafkaCmds) runMKTopicDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Topics.Get(mkTopicName(args[0], project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing topic: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMKFormat)
}

func (s *managedKafkaCmds) runMKTopicList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.Topics.List(mkClusterParent(project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing topics: %w", err)
	}
	if s.flagMKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Topics, s.flagMKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Topics), "%-40s %s\n", "NAME", "PARTITION_COUNT")
	for _, t := range resp.Topics {
		fmt.Fprintf(s.inv.stdout, "%-40s %d\n", path.Base(t.Name), t.PartitionCount)
	}
	return nil
}

func (s *managedKafkaCmds) runMKTopicUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	topic := &managedkafka.Topic{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, topic); err != nil {
		return err
	}
	mask := s.flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(topic))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.Topics.Patch(mkTopicName(args[0], project, s.flagMKLocation, s.flagMKCluster), topic).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating topic: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- consumer-groups ---

func (s *managedKafkaCmds) initManagedKafkaAll4() {
	all := []*cobra.Command{s.mkCGDeleteCmd, s.mkCGDescribeCmd, s.mkCGListCmd, s.mkCGUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMKLocation, "location", "", "Location containing the cluster (required)")
		c.Flags().StringVar(&s.flagMKCluster, "cluster", "", "Cluster containing the consumer group (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("cluster")
	}
	s.mkCGUpdateCmd.Flags().StringVar(&s.flagMKConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ConsumerGroup message body (required)")
	_ = s.mkCGUpdateCmd.MarkFlagRequired("config-file")
	s.mkCGUpdateCmd.Flags().StringVar(&s.flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.mkCGDescribeCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")
	s.mkCGListCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")

	s.mkConsumerGroupsCmd.AddCommand(all...)
	s.managedKafkaCmd.AddCommand(s.mkConsumerGroupsCmd)
}

func mkCGName(id, project, location, cluster string) string {
	return mkChild("consumerGroups", id, mkClusterParent(project, location, cluster))
}

func (s *managedKafkaCmds) runMKCGDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Clusters.ConsumerGroups.Delete(mkCGName(args[0], project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting consumer group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted consumer group [%s].\n", args[0])
	return nil
}

func (s *managedKafkaCmds) runMKCGDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.ConsumerGroups.Get(mkCGName(args[0], project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing consumer group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMKFormat)
}

func (s *managedKafkaCmds) runMKCGList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Clusters.ConsumerGroups.List(mkClusterParent(project, s.flagMKLocation, s.flagMKCluster)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing consumer groups: %w", err)
	}
	if s.flagMKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.ConsumerGroups, s.flagMKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.ConsumerGroups), "%-40s\n", "NAME")
	for _, g := range resp.ConsumerGroups {
		fmt.Fprintln(s.inv.stdout, path.Base(g.Name))
	}
	return nil
}

func (s *managedKafkaCmds) runMKCGUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	g := &managedkafka.ConsumerGroup{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, g); err != nil {
		return err
	}
	mask := s.flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(g))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Clusters.ConsumerGroups.Patch(mkCGName(args[0], project, s.flagMKLocation, s.flagMKCluster), g).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating consumer group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- connect-clusters ---

func (s *managedKafkaCmds) initManagedKafkaAll5() {
	all := []*cobra.Command{s.mkCCCreateCmd, s.mkCCDeleteCmd, s.mkCCDescribeCmd, s.mkCCListCmd, s.mkCCUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMKLocation, "location", "", "Location containing the Connect cluster (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{s.mkCCCreateCmd, s.mkCCUpdateCmd} {
		c.Flags().StringVar(&s.flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the ConnectCluster message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.mkCCUpdateCmd.Flags().StringVar(&s.flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.mkCCCreateCmd, s.mkCCDeleteCmd, s.mkCCUpdateCmd} {
		c.Flags().BoolVar(&s.flagMKAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.mkCCDescribeCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")
	s.mkCCListCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")

	s.mkConnectClustersCmd.AddCommand(all...)
	s.managedKafkaCmd.AddCommand(s.mkConnectClustersCmd)
}

func (s *managedKafkaCmds) runMKCCCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cc := &managedkafka.ConnectCluster{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, cc); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectClusters.Create(mkLocationParent(project, s.flagMKLocation), cc).
		ConnectClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connect cluster: %w", err)
	}
	return s.mkFinishOp(ctx, svc, op, "Create connect cluster", args[0], s.flagMKAsync)
}

func (s *managedKafkaCmds) runMKCCDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectClusters.Delete(mkConnectClusterParent(project, s.flagMKLocation, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connect cluster: %w", err)
	}
	return s.mkFinishOp(ctx, svc, op, "Delete connect cluster", args[0], s.flagMKAsync)
}

func (s *managedKafkaCmds) runMKCCDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectClusters.Get(mkConnectClusterParent(project, s.flagMKLocation, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connect cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMKFormat)
}

func (s *managedKafkaCmds) runMKCCList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.ConnectClusters.List(mkLocationParent(project, s.flagMKLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing connect clusters: %w", err)
	}
	if s.flagMKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.ConnectClusters, s.flagMKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.ConnectClusters), "%-40s %s\n", "NAME", "STATE")
	for _, c := range resp.ConnectClusters {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(c.Name), c.State)
	}
	return nil
}

func (s *managedKafkaCmds) runMKCCUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cc := &managedkafka.ConnectCluster{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, cc); err != nil {
		return err
	}
	mask := s.flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cc))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConnectClusters.Patch(mkConnectClusterParent(project, s.flagMKLocation, args[0]), cc).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connect cluster: %w", err)
	}
	return s.mkFinishOp(ctx, svc, op, "Update connect cluster", args[0], s.flagMKAsync)
}

// --- connectors ---

func (s *managedKafkaCmds) initManagedKafkaAll6() {
	all := []*cobra.Command{s.mkConnCreateCmd, s.mkConnDeleteCmd, s.mkConnDescribeCmd, s.mkConnListCmd, s.mkConnUpdateCmd,
		s.mkConnPauseCmd, s.mkConnResumeCmd, s.mkConnRestartCmd, s.mkConnStopCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMKLocation, "location", "", "Location containing the Connect cluster (required)")
		c.Flags().StringVar(&s.flagMKConnectClust, "connect-cluster", "", "Connect cluster containing the connector (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("connect-cluster")
	}
	for _, c := range []*cobra.Command{s.mkConnCreateCmd, s.mkConnUpdateCmd} {
		c.Flags().StringVar(&s.flagMKConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Connector message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.mkConnUpdateCmd.Flags().StringVar(&s.flagMKUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.mkConnDescribeCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")
	s.mkConnListCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")

	s.mkConnectorsCmd.AddCommand(all...)
	s.managedKafkaCmd.AddCommand(s.mkConnectorsCmd)
}

func mkConnName(id, project, location, connectCluster string) string {
	return mkChild("connectors", id, mkConnectClusterParent(project, location, connectCluster))
}

func (s *managedKafkaCmds) runMKConnCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	conn := &managedkafka.Connector{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, conn); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectClusters.Connectors.Create(mkConnectClusterParent(project, s.flagMKLocation, s.flagMKConnectClust), conn).
		ConnectorId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connector: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *managedKafkaCmds) runMKConnDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ConnectClusters.Connectors.Delete(mkConnName(args[0], project, s.flagMKLocation, s.flagMKConnectClust)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting connector: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted connector [%s].\n", args[0])
	return nil
}

func (s *managedKafkaCmds) runMKConnDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectClusters.Connectors.Get(mkConnName(args[0], project, s.flagMKLocation, s.flagMKConnectClust)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connector: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMKFormat)
}

func (s *managedKafkaCmds) runMKConnList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.ConnectClusters.Connectors.List(mkConnectClusterParent(project, s.flagMKLocation, s.flagMKConnectClust)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing connectors: %w", err)
	}
	if s.flagMKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Connectors, s.flagMKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Connectors), "%-40s\n", "NAME")
	for _, c := range resp.Connectors {
		fmt.Fprintln(s.inv.stdout, path.Base(c.Name))
	}
	return nil
}

func (s *managedKafkaCmds) runMKConnUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	conn := &managedkafka.Connector{}
	if err := loadYAMLOrJSONInto(s.flagMKConfigFile, conn); err != nil {
		return err
	}
	mask := s.flagMKUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(conn))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ConnectClusters.Connectors.Patch(mkConnName(args[0], project, s.flagMKLocation, s.flagMKConnectClust), conn).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating connector: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *managedKafkaCmds) runMKConnAction(verb string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := s.inv.project()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		svc, err := s.inv.clients().ManagedKafka(ctx)
		if err != nil {
			return err
		}
		name := mkConnName(args[0], project, s.flagMKLocation, s.flagMKConnectClust)
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
		fmt.Fprintf(s.inv.stdout, "%s connector [%s].\n", verb, args[0])
		return nil
	}
}

// --- operations ---

func (s *managedKafkaCmds) initManagedKafkaAll7() {
	for _, c := range []*cobra.Command{s.mkOpDescribeCmd, s.mkOpListCmd} {
		c.Flags().StringVar(&s.flagMKLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	s.mkOpDescribeCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")
	s.mkOpListCmd.Flags().StringVar(&s.flagMKFormat, "format", "", "Output format")

	s.mkOperationsCmd.AddCommand(s.mkOpDescribeCmd, s.mkOpListCmd)
	s.managedKafkaCmd.AddCommand(s.mkOperationsCmd)
}

func mkOpName(id, project, location string) string {
	return mkChild("operations", id, mkLocationParent(project, location))
}

func (s *managedKafkaCmds) runMKOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(mkOpName(args[0], project, s.flagMKLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagMKFormat)
}

func (s *managedKafkaCmds) runMKOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedKafka(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(mkLocationParent(project, s.flagMKLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagMKFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagMKFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// silence linter about the placeholder var (used by ACL commands).
