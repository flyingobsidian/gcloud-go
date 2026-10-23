package cmd

import (
	"fmt"

	managedkafka "google.golang.org/api/managedkafka/v1"
)

// --- Managed Kafka public-cluster flags (#1933) ---
//
// --public-cluster / --no-public-cluster and --allowed-source-ip-ranges were
// promoted to GA in gcloud-python 584.0.0 on `managed-kafka clusters create`
// and `update`. They populate gcpConfig.accessConfig.publicClusterConfig, as
// util.MapSubnetsToNetworkConfig (create) and
// util.PrepareUpdateWithPublicCluster (update) do.

const mkPublicClusterMask = "gcpConfig.accessConfig.publicClusterConfig"

var (
	flagMKPublicCluster   bool
	flagMKNoPublicCluster bool
	flagMKAllowedRanges   []string
)

func mkSetPublicClusterConfig(cluster *managedkafka.Cluster, ranges []string) {
	if cluster.GcpConfig == nil {
		cluster.GcpConfig = &managedkafka.GcpConfig{}
	}
	if cluster.GcpConfig.AccessConfig == nil {
		cluster.GcpConfig.AccessConfig = &managedkafka.AccessConfig{}
	}
	cluster.GcpConfig.AccessConfig.PublicClusterConfig = &managedkafka.PublicClusterConfig{
		AllowedSourceIpRanges: ranges,
	}
}

// mkApplyPublicClusterCreate applies the flags to a cluster being created.
func mkApplyPublicClusterCreate(cluster *managedkafka.Cluster, public bool, ranges []string) error {
	if public {
		if len(ranges) == 0 {
			return fmt.Errorf("Invalid value for [--allowed-source-ip-ranges]: Must specify at least one IPv4 CIDR range when public cluster is enabled.")
		}
		mkSetPublicClusterConfig(cluster, ranges)
		return nil
	}
	if len(ranges) > 0 {
		return fmt.Errorf("Invalid value for [--public-cluster]: Must enable public cluster to specify allowed source IP ranges.")
	}
	return nil
}

// mkApplyPublicClusterUpdate applies the flags to an update request and
// returns the update mask with the public cluster config path appended when
// it changes. noPublic clears the config.
func mkApplyPublicClusterUpdate(cluster *managedkafka.Cluster, public, noPublic bool, ranges []string, mask string) (string, error) {
	if !public && !noPublic && len(ranges) == 0 {
		return mask, nil
	}
	appendMask := func() string {
		if mask == "" {
			return mkPublicClusterMask
		}
		return mask + "," + mkPublicClusterMask
	}
	if noPublic {
		if len(ranges) > 0 {
			return "", fmt.Errorf("Invalid value for [--allowed-source-ip-ranges]: Cannot specify allowed source IP ranges when disabling public cluster.")
		}
		if cluster.GcpConfig != nil && cluster.GcpConfig.AccessConfig != nil {
			cluster.GcpConfig.AccessConfig.PublicClusterConfig = nil
		}
		return appendMask(), nil
	}
	if len(ranges) == 0 {
		return "", fmt.Errorf("Invalid value for [--allowed-source-ip-ranges]: Must specify at least one IPv4 CIDR range when public cluster is enabled.")
	}
	mkSetPublicClusterConfig(cluster, ranges)
	return appendMask(), nil
}
