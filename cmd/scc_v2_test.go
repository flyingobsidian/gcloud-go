package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

// TestSccFindingsListV2 spins up an httptest.Server acting as SCC V2 and
// checks that runSccFindingsList sends the right request and renders the
// response through the standard table branch.
func TestSccFindingsListV2(t *testing.T) {
	var gotPath, gotQuery, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
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

	// Save + reset flags between subtests.
	saveOrg, saveSrc, saveLoc, saveFmt, saveFilter := flagSccOrg, flagSccFindingSource, flagSccFindingLocation, flagSccFormat, flagSccFilter
	defer func() {
		flagSccOrg, flagSccFindingSource, flagSccFindingLocation, flagSccFormat, flagSccFilter =
			saveOrg, saveSrc, saveLoc, saveFmt, saveFilter
	}()
	flagSccOrg = "1"
	flagSccFindingSource = "-"
	flagSccFindingLocation = "eu"
	flagSccFilter = "state=\"ACTIVE\""
	flagSccFormat = ""

	out := captureStdout(t, func() {
		if err := runSccFindingsList(nil, nil); err != nil {
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
	for _, want := range []string{"NAME", "STATE", "CATEGORY", "f-abc", "OPEN_FIREWALL", "f-def"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got:\n%s", want, out)
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
