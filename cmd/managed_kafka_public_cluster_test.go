package cmd

import (
	"testing"

	managedkafka "google.golang.org/api/managedkafka/v1"
)

func TestMKPublicClusterFlagsRegistered(t *testing.T) {
	for _, c := range []string{"create", "update"} {
		cmd := findSub(mkClustersCmd, c)
		if cmd == nil {
			t.Fatalf("managed-kafka clusters %s missing", c)
		}
		for _, f := range []string{"public-cluster", "no-public-cluster", "allowed-source-ip-ranges"} {
			if cmd.Flags().Lookup(f) == nil {
				t.Errorf("--%s missing on clusters %s", f, c)
			}
		}
	}
}

func TestMKApplyPublicClusterCreate(t *testing.T) {
	cluster := &managedkafka.Cluster{GcpConfig: &managedkafka.GcpConfig{
		AccessConfig: &managedkafka.AccessConfig{NetworkConfigs: []*managedkafka.NetworkConfig{{Subnet: "s"}}},
	}}
	if err := mkApplyPublicClusterCreate(cluster, true, []string{"203.0.113.0/24"}); err != nil {
		t.Fatal(err)
	}
	ac := cluster.GcpConfig.AccessConfig
	if len(ac.NetworkConfigs) != 1 || ac.PublicClusterConfig == nil || ac.PublicClusterConfig.AllowedSourceIpRanges[0] != "203.0.113.0/24" {
		t.Errorf("accessConfig = %+v", ac)
	}
	if err := mkApplyPublicClusterCreate(&managedkafka.Cluster{}, true, nil); err == nil {
		t.Error("expected error for --public-cluster without ranges")
	}
	if err := mkApplyPublicClusterCreate(&managedkafka.Cluster{}, false, []string{"1.2.3.0/24"}); err == nil {
		t.Error("expected error for ranges without --public-cluster")
	}
	empty := &managedkafka.Cluster{}
	if err := mkApplyPublicClusterCreate(empty, false, nil); err != nil || empty.GcpConfig != nil {
		t.Errorf("no flags: %+v, %v", empty, err)
	}
}

func TestMKApplyPublicClusterUpdate(t *testing.T) {
	cluster := &managedkafka.Cluster{}
	mask, err := mkApplyPublicClusterUpdate(cluster, false, false, []string{"1.2.3.0/24"}, "labels")
	if err != nil || mask != "labels,"+mkPublicClusterMask || cluster.GcpConfig.AccessConfig.PublicClusterConfig == nil {
		t.Errorf("ranges: mask=%q err=%v cluster=%+v", mask, err, cluster)
	}
	mask, err = mkApplyPublicClusterUpdate(cluster, false, true, nil, "")
	if err != nil || mask != mkPublicClusterMask || cluster.GcpConfig.AccessConfig.PublicClusterConfig != nil {
		t.Errorf("no-public: mask=%q err=%v", mask, err)
	}
	if mask, err := mkApplyPublicClusterUpdate(&managedkafka.Cluster{}, false, false, nil, "labels"); err != nil || mask != "labels" {
		t.Errorf("no flags: mask=%q err=%v", mask, err)
	}
	if _, err := mkApplyPublicClusterUpdate(&managedkafka.Cluster{}, true, false, nil, ""); err == nil {
		t.Error("expected error for --public-cluster without ranges")
	}
	if _, err := mkApplyPublicClusterUpdate(&managedkafka.Cluster{}, false, true, []string{"1.2.3.0/24"}, ""); err == nil {
		t.Error("expected error for --no-public-cluster with ranges")
	}
}
