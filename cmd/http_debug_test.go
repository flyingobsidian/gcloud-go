package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestHTTPDebugWriter(t *testing.T) {
	saveVerbosity, saveLogHTTP := flagVerbosity, flagLogHTTP
	defer func() { flagVerbosity, flagLogHTTP = saveVerbosity, saveLogHTTP }()

	cases := []struct {
		verbosity string
		logHTTP   bool
		want      bool
	}{
		{"warning", false, false},
		{"info", false, false},
		{"debug", false, true},
		{"DEBUG", false, true},
		{"warning", true, true},
	}
	for _, tc := range cases {
		flagVerbosity, flagLogHTTP = tc.verbosity, tc.logHTTP
		if got := httpDebugWriter() != nil; got != tc.want {
			t.Errorf("httpDebugWriter() enabled = %v for --verbosity=%s --log-http=%v, want %v",
				got, tc.verbosity, tc.logHTTP, tc.want)
		}
	}
}

// TestSccFindingsListV2LogsCurl checks that --verbosity debug prints the REST
// call `scc findings list` makes, including the headers that carry the
// credentials and the billing project.
func TestSccFindingsListV2LogsCurl(t *testing.T) {
	grp := newSccCmds(legacyTestInvocation())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"listFindingsResults": []any{}})
	}))
	defer server.Close()

	orig := grp.sccV2Rest
	defer func() { grp.sccV2Rest = orig }()
	grp.sccV2RestClientForTest(server.URL)

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	var debug bytes.Buffer
	grp.inv.stderr, grp.inv.flags.verbosity = &debug, "debug"

	saveOrg, saveSrc, saveLoc, saveFmt, saveFilter := grp.flagSccOrg, grp.flagSccFindingSource, grp.flagSccFindingLocation, grp.flagSccFormat, grp.flagSccFilter
	defer func() {
		grp.flagSccOrg, grp.flagSccFindingSource, grp.flagSccFindingLocation, grp.flagSccFormat, grp.flagSccFilter =
			saveOrg, saveSrc, saveLoc, saveFmt, saveFilter
	}()
	grp.flagSccOrg = "1"
	grp.flagSccFindingSource = "-"
	withSccLocation(t, grp.sccFindingsListCmd, "eu")
	grp.flagSccFilter = "state=\"ACTIVE\""
	grp.flagSccFormat = ""

	_ = captureStdout(t, func() {
		if err := grp.runSccFindingsList(grp.sccFindingsListCmd, nil); err != nil {
			t.Fatalf("runSccFindingsList: %v", err)
		}
	})

	logged := debug.String()
	for _, want := range []string{
		"curl -X GET '" + server.URL + "/organizations/1/sources/-/locations/eu/findings?",
		"filter=state",
		`-H "Authorization: Bearer $(gcloud auth print-access-token)"`,
		"-H 'X-Goog-User-Project: qp-123'",
	} {
		if !strings.Contains(logged, want) {
			t.Errorf("debug output missing %q; got:\n%s", want, logged)
		}
	}
	if strings.Contains(logged, "test-token") {
		t.Error("debug output leaked the access token")
	}
}

// TestSccFindingsListV2QuietByDefault guards against curl output appearing
// without --verbosity debug / --log-http.
func TestSccFindingsListV2QuietByDefault(t *testing.T) {
	grp := newSccCmds(legacyTestInvocation())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"listFindingsResults": []any{}})
	}))
	defer server.Close()

	orig := grp.sccV2Rest
	defer func() { grp.sccV2Rest = orig }()
	grp.sccV2RestClientForTest(server.URL)

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	var debug bytes.Buffer
	grp.inv.stderr, grp.inv.flags.verbosity, grp.inv.flags.logHTTP = &debug, "warning", false

	saveOrg := grp.flagSccOrg
	defer func() { grp.flagSccOrg = saveOrg }()
	grp.flagSccOrg = "1"
	withSccLocation(t, grp.sccFindingsListCmd, "eu")

	_ = captureStdout(t, func() {
		if err := grp.runSccFindingsList(grp.sccFindingsListCmd, nil); err != nil {
			t.Fatalf("runSccFindingsList: %v", err)
		}
	})

	if debug.Len() != 0 {
		t.Errorf("expected no debug output at default verbosity, got:\n%s", debug.String())
	}
}
