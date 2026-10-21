package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	compute "google.golang.org/api/compute/v1"
)

// fakeDisksLister serves fixed disks and records which fetchers ran.
func fakeDisksLister(calls *[]string) computeLister[*compute.Disk] {
	zonal := &compute.Disk{Name: "z", Zone: "zones/europe-west2-a"}
	regional := &compute.Disk{Name: "r", Region: "regions/europe-west2"}
	l := disksLister
	l.aggregated = func(context.Context, *compute.Service, string) ([]*compute.Disk, error) {
		*calls = append(*calls, "aggregated")
		return []*compute.Disk{zonal, regional}, nil
	}
	l.zonalList = func(_ context.Context, _ *compute.Service, _, zone string) ([]*compute.Disk, error) {
		*calls = append(*calls, "zone:"+zone)
		return []*compute.Disk{zonal}, nil
	}
	l.regionalList = func(_ context.Context, _ *compute.Service, _, region string) ([]*compute.Disk, error) {
		*calls = append(*calls, "region:"+region)
		return []*compute.Disk{regional}, nil
	}
	return l
}

func TestComputeListerScopes(t *testing.T) {
	cases := []struct {
		name      string
		req       computeListRequest
		wantCalls string
		wantNames string
	}{
		{"default aggregates", computeListRequest{}, "aggregated", "z,r"},
		{"empty --zones keeps zonal", computeListRequest{zonesSet: true}, "aggregated", "z"},
		{"empty --regions keeps regional", computeListRequest{regionsSet: true}, "aggregated", "r"},
		{"named zones", computeListRequest{zonesSet: true, zones: []string{"a", "b"}}, "zone:a,zone:b", "z,z"},
		{"named regions", computeListRequest{regionsSet: true, regions: []string{"x"}}, "region:x", "r"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			got, err := fakeDisksLister(&calls).list(context.Background(), nil, "p", tc.req)
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, d := range got {
				names = append(names, d.Name)
			}
			if strings.Join(calls, ",") != tc.wantCalls || strings.Join(names, ",") != tc.wantNames {
				t.Errorf("calls %v names %v, want %s / %s", calls, names, tc.wantCalls, tc.wantNames)
			}
		})
	}
}

func TestComputeListerGlobalOnly(t *testing.T) {
	l := firewallRulesLister
	l.globalList = func(context.Context, *compute.Service, string) ([]*compute.Firewall, error) { return nil, nil }
	got, err := l.list(context.Background(), nil, "p", computeListRequest{})
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("got %v, %v; want empty non-nil list", got, err)
	}
}

func TestComputeListerError(t *testing.T) {
	var calls []string
	l := fakeDisksLister(&calls)
	l.aggregated = func(context.Context, *compute.Service, string) ([]*compute.Disk, error) { return nil, errors.New("boom") }
	if _, err := l.list(context.Background(), nil, "p", computeListRequest{}); err == nil || !strings.Contains(err.Error(), "listing disks: boom") {
		t.Errorf("err = %v", err)
	}
}

func TestComputeListerEmit(t *testing.T) {
	disks := []*compute.Disk{{Name: "d1", Zone: "https://x/projects/p/zones/europe-west2-a", SizeGb: 10,
		Type: "https://x/projects/p/zones/europe-west2-a/diskTypes/pd-balanced", Status: "READY", SelfLink: "https://x/d1"}}

	var out, errOut bytes.Buffer
	if err := disksLister.emit(&out, &errOut, disks, "", false); err != nil {
		t.Fatal(err)
	}
	want := "NAME  LOCATION        LOCATION_SCOPE  SIZE_GB  TYPE         STATUS\n" +
		"d1    europe-west2-a  zone            10       pd-balanced  READY\n"
	if out.String() != want {
		t.Errorf("table =\n%s\nwant\n%s", out.String(), want)
	}

	out.Reset()
	if err := disksLister.emit(&out, &errOut, disks, "", true); err != nil || out.String() != "https://x/d1\n" {
		t.Errorf("--uri output %q, err %v", out.String(), err)
	}

	out.Reset()
	if err := disksLister.emit(&out, &errOut, []*compute.Disk{}, "json", false); err != nil || strings.TrimSpace(out.String()) != "[]" {
		t.Errorf("empty json %q, err %v", out.String(), err)
	}

	out.Reset()
	errOut.Reset()
	if err := disksLister.emit(&out, &errOut, nil, "", false); err != nil || out.Len() != 0 || errOut.String() != "Listed 0 items.\n" {
		t.Errorf("empty table: stdout %q stderr %q err %v", out.String(), errOut.String(), err)
	}
}

func TestComputeListCmdFlags(t *testing.T) {
	cases := []struct {
		group     string
		has, lack []string
	}{
		{"disks", []string{"zones", "regions", "uri"}, []string{"global"}},
		{"instance-templates", []string{"regions", "global"}, []string{"zones"}},
		{"backend-services", []string{"regions", "global"}, []string{"zones"}},
		{"health-checks", []string{"regions", "global"}, []string{"zones"}},
		{"url-maps", []string{"regions", "global"}, []string{"zones"}},
		{"ssl-certificates", []string{"regions", "global"}, []string{"zones"}},
		{"routers", []string{"regions"}, []string{"zones", "global"}},
		{"firewall-rules", []string{"uri"}, []string{"zones", "regions", "global"}},
		{"snapshots", []string{"uri"}, []string{"zones", "regions", "global"}},
	}
	for _, tc := range cases {
		group, _, err := computeCmd.Find([]string{tc.group, "list"})
		if err != nil || group.Name() != "list" || group.RunE == nil {
			t.Errorf("compute %s list not registered: %v", tc.group, err)
			continue
		}
		for _, f := range tc.has {
			if group.Flags().Lookup(f) == nil {
				t.Errorf("compute %s list missing --%s", tc.group, f)
			}
		}
		for _, f := range tc.lack {
			if group.Flags().Lookup(f) != nil {
				t.Errorf("compute %s list should not have --%s", tc.group, f)
			}
		}
	}
}

func TestComputeTransforms(t *testing.T) {
	base := "https://compute.googleapis.com/compute/v1/projects/p/"
	cases := []struct{ name, got, want string }{
		{"scope zonal disk", computeScope(base + "zones/us-central1-a/disks/d1"), "us-central1-a/disks/d1"},
		{"scope global", computeScope(base + "global/images/i"), "i"},
		{"scoped suffix zonal", computeScopedSuffix(base + "zones/us-central1-a/instanceGroups/ig"), "us-central1-a/instanceGroups/ig"},
		{"scoped suffix global", computeScopedSuffix(base + "global/backendBuckets/b"), "b"},
		{"type suffix", computeTypeSuffix(base + "global/backendServices/web"), "backendServices/web"},
		{"machine type", computeMachineType("e2-medium"), "e2-medium"},
		{"custom machine type", computeMachineType("n2-custom-4-16384"), "custom (n2, 4 vCPU, 16.00 GiB)"},
		{"firewall ports", computeFirewallRule("tcp", []string{"22", "80-81"}), "tcp:22,tcp:80-81"},
		{"firewall no ports", computeFirewallRule("icmp", nil), "icmp"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestComputeListTableRows(t *testing.T) {
	fw := firewallRulesLister.tableRow(&compute.Firewall{
		Name: "allow-ssh", Network: "https://x/global/networks/default", Direction: "INGRESS", Priority: 1000,
		Allowed: []*compute.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"22"}}, {IPProtocol: "icmp"}},
	})
	if fw["allow"] != "tcp:22,icmp" || fw["network"] != "default" || fw["priority"] != "1000" || fw["disabled"] != "false" {
		t.Errorf("firewall row = %v", fw)
	}
	r := routersLister.tableRow(&compute.Router{Name: "r", Region: "https://x/regions/europe-west2"})
	if r["network"] != "N/A" || r["region"] != "europe-west2" {
		t.Errorf("router row = %v", r)
	}
	it := instanceTemplatesLister.tableRow(&compute.InstanceTemplate{Name: "t", Properties: &compute.InstanceProperties{
		MachineType: "e2-small", Scheduling: &compute.Scheduling{Preemptible: true}}})
	if it["machineType"] != "e2-small" || it["preemptible"] != "true" {
		t.Errorf("instance template row = %v", it)
	}
	bs := backendServicesLister.tableRow(&compute.BackendService{Name: "b", Protocol: "HTTP", Backends: []*compute.Backend{
		{Group: "https://x/projects/p/zones/z2/instanceGroups/b"}, {Group: "https://x/projects/p/zones/z1/instanceGroups/a"}}})
	if bs["backends"] != "z1/instanceGroups/a,z2/instanceGroups/b" {
		t.Errorf("backend service row = %v", bs)
	}
	um := urlMapsLister.tableRow(&compute.UrlMap{Name: "u", DefaultService: "https://x/projects/p/global/backendServices/web"})
	if um["defaultService"] != "backendServices/web" {
		t.Errorf("url map row = %v", um)
	}
	ssl := sslCertificatesLister.tableRow(&compute.SslCertificate{Name: "c", Type: "MANAGED", Managed: &compute.SslCertificateManagedSslCertificate{
		Status: "ACTIVE", DomainStatus: map[string]string{"b.example.com": "ACTIVE", "a.example.com": "PROVISIONING"}}})
	if ssl["managedStatus"] != "ACTIVE" || ssl["domainStatus"] != "a.example.com: PROVISIONING, b.example.com: ACTIVE" {
		t.Errorf("ssl certificate row = %v", ssl)
	}
}
