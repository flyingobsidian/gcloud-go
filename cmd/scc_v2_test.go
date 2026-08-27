package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func TestSccV2ParentPath(t *testing.T) {
	cases := []struct {
		parent, source, location, want string
	}{
		{"organizations/1", "", "us", "organizations/1/sources/-/locations/us"},
		{"organizations/1", "-", "us", "organizations/1/sources/-/locations/us"},
		{"organizations/1", "9876", "eu", "organizations/1/sources/9876/locations/eu"},
		{"folders/2", "3", "us-central1", "folders/2/sources/3/locations/us-central1"},
		{"projects/p", "", "eu", "projects/p/sources/-/locations/eu"},
		{"organizations/1", "sources/42", "us", "organizations/1/sources/42/locations/us"},
	}
	for _, tc := range cases {
		got := sccV2ParentPath(tc.parent, tc.source, tc.location)
		if got != tc.want {
			t.Errorf("sccV2ParentPath(%q,%q,%q) = %q, want %q", tc.parent, tc.source, tc.location, got, tc.want)
		}
	}
}

// withSccLocation sets --location on cmd the way parsing a command line
// would, so that the V1/V2 routing sees it as specified. Each command
// registers its own --location flag, so the routing only sees the one on the
// command being run. The flag is restored when the test ends.
func withSccLocation(t *testing.T, cmd *cobra.Command, location string) {
	t.Helper()
	f := cmd.Flags().Lookup("location")
	if f == nil {
		t.Fatalf("--location flag is not registered on %q", cmd.Name())
	}
	prev, prevChanged := f.Value.String(), f.Changed
	if err := cmd.Flags().Set("location", location); err != nil {
		t.Fatalf("setting --location=%s: %v", location, err)
	}
	t.Cleanup(func() {
		_ = f.Value.Set(prev)
		f.Changed = prevChanged
	})
}

// TestSccFindingsListV2 spins up an httptest.Server acting as SCC V2 and
// checks that runSccFindingsList sends the right request and renders the
// response through the standard table branch.
func TestSccFindingsListV2(t *testing.T) {
	var gotPath, gotQuery, gotAuth, gotUserProject string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		gotUserProject = r.Header.Get("X-Goog-User-Project")
		resp := map[string]any{
			"listFindingsResults": []any{
				map[string]any{
					"finding": map[string]any{
						"name":     "organizations/1/sources/-/locations/eu/findings/f-abc",
						"state":    "ACTIVE",
						"category": "OPEN_FIREWALL",
					},
				},
				map[string]any{
					"finding": map[string]any{
						"name":     "organizations/1/sources/-/locations/eu/findings/f-def",
						"state":    "INACTIVE",
						"category": "STALE_KEY",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Swap the V2 endpoint for our test server and restore afterwards.
	orig := sccV2Rest
	defer func() { sccV2Rest = orig }()
	sccV2RestClientForTest(server.URL)

	// Avoid the real auth path -- CI has no credentials. A synthetic static
	// token is enough because the test server ignores the value beyond
	// asserting the Bearer prefix.
	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()

	// #1740: sccV2ListFindings requires a resolvable billing project and
	// the shared restClient stamps it into X-Goog-User-Project.
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	// Save + reset flags between subtests.
	saveOrg, saveSrc, saveLoc, saveFmt, saveFilter := flagSccOrg, flagSccFindingSource, flagSccFindingLocation, flagSccFormat, flagSccFilter
	defer func() {
		flagSccOrg, flagSccFindingSource, flagSccFindingLocation, flagSccFormat, flagSccFilter =
			saveOrg, saveSrc, saveLoc, saveFmt, saveFilter
	}()
	flagSccOrg = "1"
	flagSccFindingSource = "-"
	withSccLocation(t, sccFindingsListCmd, "eu")
	flagSccFilter = "state=\"ACTIVE\""
	flagSccFormat = ""

	out := captureStdout(t, func() {
		if err := runSccFindingsList(sccFindingsListCmd, nil); err != nil {
			t.Fatalf("runSccFindingsList: %v", err)
		}
	})

	wantPath := "/organizations/1/sources/-/locations/eu/findings"
	if gotPath != wantPath {
		t.Errorf("request path = %q, want %q", gotPath, wantPath)
	}
	if !strings.Contains(gotQuery, "filter=") {
		t.Errorf("expected filter= in query, got %q", gotQuery)
	}
	if gotAuth == "" || !strings.HasPrefix(strings.ToLower(gotAuth), "bearer ") {
		t.Errorf("expected Bearer token in Authorization header, got %q", gotAuth)
	}
	if gotUserProject != "qp-123" {
		t.Errorf("expected X-Goog-User-Project=qp-123 (#1740), got %q", gotUserProject)
	}
	for _, want := range []string{"NAME", "STATE", "CATEGORY", "f-abc", "OPEN_FIREWALL", "f-def"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got:\n%s", want, out)
		}
	}
}

// TestSccFindingsListRoutesGlobalToV2 covers the case an organization without
// data residency controls needs: --location=global must reach the V2 API at
// locations/global, not fall back to V1. Regional endpoints are rejected there
// with DRZ_LOCATION_MISMATCH, so global is the only location such an
// organization can use.
func TestSccFindingsListRoutesGlobalToV2(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"listFindingsResults": []any{}})
	}))
	defer server.Close()

	orig := sccV2Rest
	defer func() { sccV2Rest = orig }()
	sccV2RestClientForTest(server.URL)

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	saveOrg, saveSrc, saveFmt := flagSccOrg, flagSccFindingSource, flagSccFormat
	defer func() { flagSccOrg, flagSccFindingSource, flagSccFormat = saveOrg, saveSrc, saveFmt }()
	flagSccOrg = "1057127509270"
	flagSccFindingSource = "-"
	flagSccFormat = ""
	withSccLocation(t, sccFindingsListCmd, "global")

	_ = captureStdout(t, func() {
		if err := runSccFindingsList(sccFindingsListCmd, nil); err != nil {
			t.Fatalf("runSccFindingsList: %v", err)
		}
	})

	want := "/organizations/1057127509270/sources/-/locations/global/findings"
	if gotPath != want {
		t.Errorf("request path = %q, want %q", gotPath, want)
	}
}

// TestSccLocationSpecified pins the routing rule itself: an explicit
// --location selects V2 whatever its value, and omitting it stays on V1.
func TestSccLocationSpecified(t *testing.T) {
	if sccLocationSpecified(sccFindingsListCmd) {
		t.Error("expected --location to start unspecified")
	}
	for _, location := range []string{"global", "eu", "us"} {
		withSccLocation(t, sccFindingsListCmd, location)
		if !sccLocationSpecified(sccFindingsListCmd) {
			t.Errorf("--location=%s should route to V2", location)
		}
	}
}

// TestSccBulkMuteV2 checks the V2 bulk-mute request: a POST to
// findings:bulkMute under the regionalized scope, carrying the filter and the
// mute state, with the LRO echoed back through --format.
func TestSccBulkMuteV2(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name": "organizations/1057127509270/operations/bulk-mute-op",
			"done": false,
		})
	}))
	defer server.Close()

	orig := sccV2Rest
	defer func() { sccV2Rest = orig }()
	sccV2RestClientForTest(server.URL)

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	saveOrg, saveFilter, saveState, saveFmt := flagSccOrg, flagSccFindingBulkMuteFilter, flagSccFindingBulkMuteState, flagSccFormat
	defer func() {
		flagSccOrg, flagSccFindingBulkMuteFilter, flagSccFindingBulkMuteState, flagSccFormat =
			saveOrg, saveFilter, saveState, saveFmt
	}()
	flagSccOrg = "1057127509270"
	flagSccFindingBulkMuteFilter = `category="XSS_SCRIPTING"`
	flagSccFindingBulkMuteState = "muted"
	flagSccFormat = "json"
	withSccLocation(t, sccFindingsBulkMuteCmd, "global")

	out := captureStdout(t, func() {
		if err := runSccFindingsBulkMute(sccFindingsBulkMuteCmd, nil); err != nil {
			t.Fatalf("runSccFindingsBulkMute: %v", err)
		}
	})

	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	wantPath := "/organizations/1057127509270/locations/global/findings:bulkMute"
	if gotPath != wantPath {
		t.Errorf("request path = %q, want %q", gotPath, wantPath)
	}
	if got := gotBody["filter"]; got != `category="XSS_SCRIPTING"` {
		t.Errorf("filter = %v, want the category filter", got)
	}
	if got := gotBody["muteState"]; got != "MUTED" {
		t.Errorf("muteState = %v, want MUTED", got)
	}
	if !strings.Contains(out, "bulk-mute-op") {
		t.Errorf("expected the operation name in the output, got:\n%s", out)
	}
}

// TestSccBulkMuteV2Undefined checks --mute-state=undefined, which unmutes.
func TestSccBulkMuteV2Undefined(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"done": true})
	}))
	defer server.Close()

	orig := sccV2Rest
	defer func() { sccV2Rest = orig }()
	sccV2RestClientForTest(server.URL)

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	saveOrg, saveFilter, saveState, saveFmt := flagSccOrg, flagSccFindingBulkMuteFilter, flagSccFindingBulkMuteState, flagSccFormat
	defer func() {
		flagSccOrg, flagSccFindingBulkMuteFilter, flagSccFindingBulkMuteState, flagSccFormat =
			saveOrg, saveFilter, saveState, saveFmt
	}()
	flagSccOrg = "1"
	flagSccFindingBulkMuteFilter = "state=\"ACTIVE\""
	flagSccFindingBulkMuteState = "undefined"
	flagSccFormat = "json"
	withSccLocation(t, sccFindingsBulkMuteCmd, "eu")

	_ = captureStdout(t, func() {
		if err := runSccFindingsBulkMute(sccFindingsBulkMuteCmd, nil); err != nil {
			t.Fatalf("runSccFindingsBulkMute: %v", err)
		}
	})

	if got := gotBody["muteState"]; got != "UNDEFINED" {
		t.Errorf("muteState = %v, want UNDEFINED", got)
	}
}

func TestSccMuteStateEnum(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"muted", "MUTED", false},
		{"MUTED", "MUTED", false},
		{"", "MUTED", false},
		{"undefined", "UNDEFINED", false},
		{"Undefined", "UNDEFINED", false},
		{"unmuted", "", true},
		{"nonsense", "", true},
	}
	for _, tc := range cases {
		got, err := sccMuteStateEnum(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("sccMuteStateEnum(%q) = %q, want an error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("sccMuteStateEnum(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("sccMuteStateEnum(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSccV2ScopePath(t *testing.T) {
	cases := []struct{ parent, location, want string }{
		{"organizations/1", "global", "organizations/1/locations/global"},
		{"folders/2", "eu", "folders/2/locations/eu"},
		{"projects/p", "locations/us", "projects/p/locations/us"},
	}
	for _, tc := range cases {
		if got := sccV2ScopePath(tc.parent, tc.location); got != tc.want {
			t.Errorf("sccV2ScopePath(%q,%q) = %q, want %q", tc.parent, tc.location, got, tc.want)
		}
	}
}

func TestSccV2RejectsUnsupportedFlags(t *testing.T) {
	// --compare-duration and --read-time are V1-only.
	saveLoc, saveDur, saveTime, saveOrg := flagSccFindingLocation, flagSccFindingCompareDur, flagSccFindingReadTime, flagSccOrg
	defer func() {
		flagSccFindingLocation, flagSccFindingCompareDur, flagSccFindingReadTime, flagSccOrg =
			saveLoc, saveDur, saveTime, saveOrg
	}()
	flagSccFindingLocation = "eu"
	flagSccOrg = "1"
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	flagSccFindingCompareDur = "1h"
	flagSccFindingReadTime = ""
	if err := sccV2ListFindings("organizations/1"); err == nil ||
		!strings.Contains(err.Error(), "--compare-duration") {
		t.Errorf("expected --compare-duration rejection, got %v", err)
	}

	flagSccFindingCompareDur = ""
	flagSccFindingReadTime = "2026-01-01T00:00:00Z"
	if err := sccV2ListFindings("organizations/1"); err == nil ||
		!strings.Contains(err.Error(), "--read-time") {
		t.Errorf("expected --read-time rejection, got %v", err)
	}
}

// TestSccV2RequiresBillingProject guards #1740: sccV2ListFindings must fail
// early with a clear message when no billing / quota project can be
// resolved.
func TestSccV2RequiresBillingProject(t *testing.T) {
	saveLoc, saveOrg := flagSccFindingLocation, flagSccOrg
	defer func() { flagSccFindingLocation, flagSccOrg = saveLoc, saveOrg }()
	flagSccFindingLocation = "eu"
	flagSccOrg = "1"

	restoreQP := SetRestUserProjectForTest("")
	defer restoreQP()

	err := sccV2ListFindings("organizations/1")
	if err == nil {
		t.Fatal("expected error when no billing project is resolvable")
	}
	msg := err.Error()
	for _, want := range []string{
		"quota (billing) project is required",
		"--billing-project",
		"billing/quota_project",
		"set-quota-project",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message missing hint %q; got: %s", want, msg)
		}
	}
}
