package cmd

import (
	"bytes"
	"reflect"
	"testing"

	dns "google.golang.org/api/dns/v1"
)

func TestDNSOutboundEndpointsRegistered(t *testing.T) {
	c, _, err := dnsCmd.Find([]string{"outbound-endpoints"})
	if err != nil || c != dnsOutboundEndpointsCmd {
		t.Fatalf("dns outbound-endpoints not registered: %v", err)
	}
	assertSubcommands(t, c, []string{"create", "delete", "describe", "list", "update"})
	for _, f := range []string{"network", "subnetwork", "endpoint-ip", "labels", "async", "endpoint-mode"} {
		if dnsOECreateCmd.Flags().Lookup(f) == nil {
			t.Errorf("create missing --%s", f)
		}
	}
}

func TestDNSOEEndpoint(t *testing.T) {
	for mode, want := range map[string]string{
		"": "", "global": "", "regional-preferred": "",
		"regional": "https://dns.us-east1.rep.googleapis.com/",
	} {
		got, err := dnsOEEndpoint(mode, "us-east1")
		if err != nil || got != want {
			t.Errorf("%q: got %q, %v; want %q", mode, got, err, want)
		}
	}
	if _, err := dnsOEEndpoint("bogus", "us-east1"); err == nil {
		t.Error("expected error for invalid --endpoint-mode")
	}
}

func TestDNSOENames(t *testing.T) {
	parent := "projects/p/locations/us-east1"
	if got := dnsOEName(parent, "ep"); got != parent+"/outboundEndpoints/ep" {
		t.Errorf("short: %q", got)
	}
	if got := dnsOEName(parent, "projects/q/locations/l/outboundEndpoints/ep"); got != "projects/q/locations/l/outboundEndpoints/ep" {
		t.Errorf("full: %q", got)
	}
	if got := dnsOEFormatNetwork("net", "p"); got != "projects/p/global/networks/net" {
		t.Errorf("network: %q", got)
	}
	if got := dnsOEFormatSubnetwork("sub", "p", "us-east1"); got != "projects/p/regions/us-east1/subnetworks/sub" {
		t.Errorf("subnetwork: %q", got)
	}
	if got := dnsOEFormatSubnetwork("projects/x/regions/r/subnetworks/s", "p", "l"); got != "projects/x/regions/r/subnetworks/s" {
		t.Errorf("subnetwork passthrough: %q", got)
	}
}

func TestApplyLabelsDiff(t *testing.T) {
	cur := map[string]string{"a": "1", "b": "2"}
	cases := []struct {
		clear   bool
		remove  []string
		update  map[string]string
		want    map[string]string
		changed bool
	}{
		{false, nil, map[string]string{"a": "1"}, map[string]string{"a": "1", "b": "2"}, false},
		{false, nil, map[string]string{"c": "3"}, map[string]string{"a": "1", "b": "2", "c": "3"}, true},
		{false, []string{"b"}, map[string]string{"a": "9"}, map[string]string{"a": "9"}, true},
		{true, nil, map[string]string{"z": "0"}, map[string]string{"z": "0"}, true},
	}
	for i, c := range cases {
		got, changed := applyLabelsDiff(cur, c.clear, c.remove, c.update)
		if !reflect.DeepEqual(got, c.want) || changed != c.changed {
			t.Errorf("case %d: got %v, %v; want %v, %v", i, got, changed, c.want, c.changed)
		}
	}
}

func TestEmitDNSOEList(t *testing.T) {
	var out, errOut bytes.Buffer
	eps := []*dns.OutboundEndpoint{{
		Name:       "projects/p/locations/us-east1/outboundEndpoints/ep1",
		EndpointIp: "10.0.0.5",
		Network:    "projects/p/global/networks/net",
		Subnetwork: "projects/p/regions/us-east1/subnetworks/sub",
	}}
	if err := emitDNSOEList(&out, &errOut, eps, "", false); err != nil {
		t.Fatal(err)
	}
	want := "NAME  ENDPOINT_IP  NETWORK  SUBNETWORK  DESCRIPTION\n" +
		"ep1   10.0.0.5     net      sub         \n"
	if out.String() != want {
		t.Errorf("table:\n%q\nwant:\n%q", out.String(), want)
	}
	out.Reset()
	if err := emitDNSOEList(&out, &errOut, nil, "", false); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 || errOut.String() != "Listed 0 items.\n" {
		t.Errorf("empty: out=%q err=%q", out.String(), errOut.String())
	}
}
