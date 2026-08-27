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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	var debug bytes.Buffer
	saveOut, saveVerbosity := httpDebugOut, flagVerbosity
	httpDebugOut, flagVerbosity = &debug, "debug"
	defer func() { httpDebugOut, flagVerbosity = saveOut, saveVerbosity }()

	saveOrg, saveSrc, saveLoc, saveFmt, saveFilter := flagSccOrg, flagSccFindingSource, flagSccFindingLocation, flagSccFormat, flagSccFilter
	defer func() {
		flagSccOrg, flagSccFindingSource, flagSccFindingLocation, flagSccFormat, flagSccFilter =
			saveOrg, saveSrc, saveLoc, saveFmt, saveFilter
	}()
	flagSccOrg = "1"
	flagSccFindingSource = "-"
	withSccLocation(t, "eu")
	flagSccFilter = "state=\"ACTIVE\""
	flagSccFormat = ""

	_ = captureStdout(t, func() {
		if err := runSccFindingsList(sccFindingsListCmd, nil); err != nil {
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	var debug bytes.Buffer
	saveOut, saveVerbosity, saveLogHTTP := httpDebugOut, flagVerbosity, flagLogHTTP
	httpDebugOut, flagVerbosity, flagLogHTTP = &debug, "warning", false
	defer func() { httpDebugOut, flagVerbosity, flagLogHTTP = saveOut, saveVerbosity, saveLogHTTP }()

	saveOrg := flagSccOrg
	defer func() { flagSccOrg = saveOrg }()
	flagSccOrg = "1"
	withSccLocation(t, "eu")

	_ = captureStdout(t, func() {
		if err := runSccFindingsList(sccFindingsListCmd, nil); err != nil {
			t.Fatalf("runSccFindingsList: %v", err)
		}
	})

	if debug.Len() != 0 {
		t.Errorf("expected no debug output at default verbosity, got:\n%s", debug.String())
	}
}
