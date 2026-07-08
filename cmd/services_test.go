package cmd

import (
	"reflect"
	"testing"
)

func TestServiceResourceName(t *testing.T) {
	cases := []struct{ project, service, want string }{
		{"my-project", "compute.googleapis.com", "projects/my-project/services/compute.googleapis.com"},
		{"my-project", "projects/other/services/foo", "projects/other/services/foo"},
	}
	for _, c := range cases {
		if got := serviceResourceName(c.project, c.service); got != c.want {
			t.Errorf("serviceResourceName(%q,%q) = %q, want %q", c.project, c.service, got, c.want)
		}
	}
}

func TestServicesListFilter(t *testing.T) {
	cases := []struct {
		enabled, avail bool
		want           string
	}{
		{false, false, "state:ENABLED"},
		{true, false, "state:ENABLED"},
		{false, true, ""},
		{true, true, "state:ENABLED"},
	}
	for _, c := range cases {
		flagServicesEnabled = c.enabled
		flagServicesAvail = c.avail
		if got := servicesListFilter(); got != c.want {
			t.Errorf("servicesListFilter(enabled=%v,avail=%v) = %q, want %q", c.enabled, c.avail, got, c.want)
		}
	}
	flagServicesEnabled = false
	flagServicesAvail = false
}

func TestAPIKeyName(t *testing.T) {
	cases := []struct{ project, id, want string }{
		{"my-project", "abc123", "projects/my-project/locations/global/keys/abc123"},
		{"my-project", "projects/other/locations/global/keys/foo", "projects/other/locations/global/keys/foo"},
	}
	for _, c := range cases {
		if got := apiKeyName(c.project, c.id); got != c.want {
			t.Errorf("apiKeyName(%q,%q) = %q, want %q", c.project, c.id, got, c.want)
		}
	}
}

func TestAPIKeyParent(t *testing.T) {
	if got := apiKeyParent("my-project"); got != "projects/my-project/locations/global" {
		t.Errorf("apiKeyParent = %q", got)
	}
}

func TestParseAPITargets(t *testing.T) {
	if parseAPITargets(nil) != nil {
		t.Error("expected nil for empty input")
	}
	got := parseAPITargets([]string{"compute.googleapis.com", "storage.googleapis.com:get:list"})
	if len(got) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(got))
	}
	if got[0].Service != "compute.googleapis.com" || len(got[0].Methods) != 0 {
		t.Errorf("first target = %+v", got[0])
	}
	if got[1].Service != "storage.googleapis.com" || !reflect.DeepEqual(got[1].Methods, []string{"get", "list"}) {
		t.Errorf("second target = %+v", got[1])
	}
}

func TestBuildAPIKeyRestrictionsNil(t *testing.T) {
	flagAPIKeyAllowedIPs = nil
	flagAPIKeyAllowedReferrers = nil
	flagAPIKeyAllowedBundleIDs = nil
	flagAPIKeyAllowedAndroidPackages = nil
	flagAPIKeyAllowedAndroidSHAs = nil
	flagAPIKeyAPITargets = nil
	if buildAPIKeyRestrictions() != nil {
		t.Error("expected nil when no restriction flags set")
	}
}

func TestBuildAPIKeyRestrictionsPopulated(t *testing.T) {
	flagAPIKeyAllowedIPs = []string{"1.2.3.4"}
	flagAPIKeyAllowedReferrers = []string{"example.com/*"}
	flagAPIKeyAllowedBundleIDs = []string{"com.example.app"}
	flagAPIKeyAllowedAndroidPackages = []string{"com.example.android"}
	flagAPIKeyAllowedAndroidSHAs = []string{"AB:CD:EF"}
	flagAPIKeyAPITargets = []string{"compute.googleapis.com:instances.get"}
	t.Cleanup(func() {
		flagAPIKeyAllowedIPs = nil
		flagAPIKeyAllowedReferrers = nil
		flagAPIKeyAllowedBundleIDs = nil
		flagAPIKeyAllowedAndroidPackages = nil
		flagAPIKeyAllowedAndroidSHAs = nil
		flagAPIKeyAPITargets = nil
	})

	r := buildAPIKeyRestrictions()
	if r == nil {
		t.Fatal("expected non-nil restrictions")
	}
	if r.ServerKeyRestrictions == nil || len(r.ServerKeyRestrictions.AllowedIps) != 1 {
		t.Errorf("ServerKeyRestrictions = %+v", r.ServerKeyRestrictions)
	}
	if r.BrowserKeyRestrictions == nil || len(r.BrowserKeyRestrictions.AllowedReferrers) != 1 {
		t.Errorf("BrowserKeyRestrictions = %+v", r.BrowserKeyRestrictions)
	}
	if r.IosKeyRestrictions == nil || len(r.IosKeyRestrictions.AllowedBundleIds) != 1 {
		t.Errorf("IosKeyRestrictions = %+v", r.IosKeyRestrictions)
	}
	if r.AndroidKeyRestrictions == nil || len(r.AndroidKeyRestrictions.AllowedApplications) != 1 ||
		r.AndroidKeyRestrictions.AllowedApplications[0].PackageName != "com.example.android" ||
		r.AndroidKeyRestrictions.AllowedApplications[0].Sha1Fingerprint != "AB:CD:EF" {
		t.Errorf("AndroidKeyRestrictions = %+v", r.AndroidKeyRestrictions)
	}
	if len(r.ApiTargets) != 1 || r.ApiTargets[0].Service != "compute.googleapis.com" {
		t.Errorf("ApiTargets = %+v", r.ApiTargets)
	}
}

func TestServiceParent(t *testing.T) {
	if got := serviceParent("servicenetworking.googleapis.com"); got != "services/servicenetworking.googleapis.com" {
		t.Errorf("serviceParent = %q", got)
	}
	if got := serviceParent("services/x"); got != "services/x" {
		t.Errorf("passthrough failed: %q", got)
	}
}

func TestConsumerNetwork(t *testing.T) {
	if got := consumerNetwork("my-project", "default"); got != "projects/my-project/global/networks/default" {
		t.Errorf("consumerNetwork = %q", got)
	}
	if got := consumerNetwork("my-project", "projects/other/global/networks/foo"); got != "projects/other/global/networks/foo" {
		t.Errorf("passthrough failed: %q", got)
	}
}

func TestPeeredDNSParent(t *testing.T) {
	got := peeredDNSParent("servicenetworking.googleapis.com", "12345", "default")
	want := "services/servicenetworking.googleapis.com/projects/12345/global/networks/default"
	if got != want {
		t.Errorf("peeredDNSParent = %q, want %q", got, want)
	}
	got = peeredDNSParent("services/x", "12345", "default")
	want = "services/x/projects/12345/global/networks/default"
	if got != want {
		t.Errorf("with services/ prefix = %q, want %q", got, want)
	}
}
