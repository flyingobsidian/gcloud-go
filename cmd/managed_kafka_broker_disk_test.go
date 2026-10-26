package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	managedkafka "google.golang.org/api/managedkafka/v1"
)

func TestApplyBrokerDiskFlag(t *testing.T) {
	cases := []struct {
		name    string
		flag    string
		cluster *managedkafka.Cluster
		want    *managedkafka.BrokerCapacityConfig
		wantErr bool
	}{
		{name: "unset leaves cluster untouched", flag: "", cluster: &managedkafka.Cluster{}, want: nil},
		{name: "creates config when nil", flag: "500", cluster: &managedkafka.Cluster{}, want: &managedkafka.BrokerCapacityConfig{DiskSizeGib: 500}},
		{name: "parses units", flag: "1TiB", cluster: &managedkafka.Cluster{}, want: &managedkafka.BrokerCapacityConfig{DiskSizeGib: 1024}},
		{
			name:    "overrides config-file value",
			flag:    "200Gi",
			cluster: &managedkafka.Cluster{BrokerCapacityConfig: &managedkafka.BrokerCapacityConfig{DiskSizeGib: 100}},
			want:    &managedkafka.BrokerCapacityConfig{DiskSizeGib: 200},
		},
		{name: "invalid value", flag: "1.5Gi", cluster: &managedkafka.Cluster{}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			old := flagMKBrokerDiskGiB
			flagMKBrokerDiskGiB = tc.flag
			t.Cleanup(func() { flagMKBrokerDiskGiB = old })

			err := applyBrokerDiskFlag(tc.cluster)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("applyBrokerDiskFlag() = nil, want error")
				}
				if tc.cluster.BrokerCapacityConfig != nil {
					t.Errorf("BrokerCapacityConfig = %+v, want nil on error", tc.cluster.BrokerCapacityConfig)
				}
				return
			}
			if err != nil {
				t.Fatalf("applyBrokerDiskFlag() error: %v", err)
			}
			got := tc.cluster.BrokerCapacityConfig
			if tc.want == nil {
				if got != nil {
					t.Errorf("BrokerCapacityConfig = %+v, want nil", got)
				}
				return
			}
			if got == nil || got.DiskSizeGib != tc.want.DiskSizeGib {
				t.Errorf("BrokerCapacityConfig = %+v, want DiskSizeGib %d", got, tc.want.DiskSizeGib)
			}
		})
	}
}

func TestApplyBrokerDiskFlagRequestBody(t *testing.T) {
	old := flagMKBrokerDiskGiB
	flagMKBrokerDiskGiB = "0"
	t.Cleanup(func() { flagMKBrokerDiskGiB = old })

	cluster := &managedkafka.Cluster{}
	if err := applyBrokerDiskFlag(cluster); err != nil {
		t.Fatalf("applyBrokerDiskFlag() error: %v", err)
	}
	raw, err := json.Marshal(cluster)
	if err != nil {
		t.Fatal(err)
	}
	if want := `"brokerCapacityConfig":{"diskSizeGib":"0"}`; !strings.Contains(string(raw), want) {
		t.Errorf("request body %s does not contain %s", raw, want)
	}
	if got := joinMask(nonEmptyJSONFields(cluster)); got != "brokerCapacityConfig" {
		t.Errorf("derived update mask = %q, want %q", got, "brokerCapacityConfig")
	}
}

func TestParseBrokerDiskSizeGiB(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{in: "100", want: 100},
		{in: "0", want: 0},
		{in: "100Gi", want: 100},
		{in: "100GiB", want: 100},
		{in: "1Ti", want: 1024},
		{in: "1TiB", want: 1024},
		{in: "  200Gi ", want: 200},
		// Fractional GiB values are rejected.
		{in: "1.5Gi", wantErr: true},
		{in: "500Mi", wantErr: true},
		// Unknown units.
		{in: "10Xi", wantErr: true},
		// Empty / non-numeric.
		{in: "", wantErr: true},
		{in: "Gi", wantErr: true},
		{in: "-5", wantErr: true},
	}
	for _, tc := range cases {
		got, err := parseBrokerDiskSizeGiB(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseBrokerDiskSizeGiB(%q) = %d, want error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseBrokerDiskSizeGiB(%q) error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseBrokerDiskSizeGiB(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestMKClusterCommandsHaveBrokerDiskFlag(t *testing.T) {
	cases := []struct {
		path []string
	}{
		{path: []string{"managed-kafka", "clusters", "create"}},
		{path: []string{"managed-kafka", "clusters", "update"}},
	}
	for _, tc := range cases {
		var cur = findRootSub(tc.path[0])
		for _, n := range tc.path[1:] {
			cur = findSub(cur, n)
		}
		if cur == nil {
			t.Fatalf("%v missing", tc.path)
		}
		if cur.Flags().Lookup("broker-disk-size-gib") == nil {
			t.Errorf("%v missing --broker-disk-size-gib flag", tc.path)
		}
	}
}
