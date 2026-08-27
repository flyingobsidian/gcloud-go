package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

// TestRestClientRefusesWithoutBillingProject checks that do() fails before it
// touches the network when no billing / quota project resolves, rather than
// sending a request that the server would reject with PERMISSION_DENIED.
func TestRestClientRefusesWithoutBillingProject(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("")
	defer restoreQP()

	c := newRESTClient(server.URL + "/v2")
	err := c.do(context.Background(), http.MethodGet, "/organizations/1/findings", nil, nil, nil)
	if err == nil {
		t.Fatal("expected do() to refuse the call with no billing project")
	}
	if called {
		t.Error("do() made the HTTP call despite having no billing project")
	}
	for _, want := range []string{
		"quota (billing) project is required",
		"--billing-project",
		"billing/quota_project",
		"set-quota-project",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error message missing hint %q; got: %s", want, err.Error())
		}
	}
}

// TestRestClientSendsBillingProject is the other half: with a project set the
// call goes out carrying it.
func TestRestClientSendsBillingProject(t *testing.T) {
	var gotUserProject string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserProject = r.Header.Get("X-Goog-User-Project")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	c := newRESTClient(server.URL + "/v2")
	if err := c.do(context.Background(), http.MethodGet, "/organizations/1/findings", nil, nil, nil); err != nil {
		t.Fatalf("do(): %v", err)
	}
	if gotUserProject != "qp-123" {
		t.Errorf("X-Goog-User-Project = %q, want qp-123", gotUserProject)
	}
}

// TestPaginateLogsProgress drives a three-page listing and checks the running
// counts, the totalSize taken from the response body, and that the last page
// is marked as such.
func TestPaginateLogsProgress(t *testing.T) {
	pages := []map[string]any{
		{"listFindingsResults": []any{map[string]any{"a": 1}, map[string]any{"a": 2}}, "totalSize": 5, "nextPageToken": "tok-2"},
		{"listFindingsResults": []any{map[string]any{"a": 3}, map[string]any{"a": 4}}, "totalSize": 5, "nextPageToken": "tok-3"},
		{"listFindingsResults": []any{map[string]any{"a": 5}}, "totalSize": 5},
	}
	var gotTokens []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("pageToken")
		gotTokens = append(gotTokens, token)
		i := 0
		switch token {
		case "tok-2":
			i = 1
		case "tok-3":
			i = 2
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pages[i])
	}))
	defer server.Close()

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	var debug bytes.Buffer
	saveOut, saveVerbosity := httpDebugOut, flagVerbosity
	httpDebugOut, flagVerbosity = &debug, "debug"
	defer func() { httpDebugOut, flagVerbosity = saveOut, saveVerbosity }()

	c := newRESTClient(server.URL + "/v2")
	all, err := c.paginate(context.Background(), "/organizations/1/findings", nil, "listFindingsResults", 2)
	if err != nil {
		t.Fatalf("paginate: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("collected %d results, want 5", len(all))
	}
	if len(gotTokens) != 3 {
		t.Fatalf("made %d requests, want 3", len(gotTokens))
	}

	logged := debug.String()
	for _, want := range []string{
		"# page 1: 2 listFindingsResults, 2 of 5 so far (more pages follow)",
		"# page 2: 2 listFindingsResults, 4 of 5 so far (more pages follow)",
		"# page 3: 1 listFindingsResults, 5 of 5 so far (last page)",
	} {
		if !strings.Contains(logged, want) {
			t.Errorf("debug output missing %q; got:\n%s", want, logged)
		}
	}
}

// TestPaginateOmitsAbsentTotalSize covers collections that do not report a
// grand total: the running count is still logged, without an " of N".
func TestPaginateOmitsAbsentTotalSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"a": 1}}})
	}))
	defer server.Close()

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	var debug bytes.Buffer
	saveOut, saveVerbosity := httpDebugOut, flagVerbosity
	httpDebugOut, flagVerbosity = &debug, "debug"
	defer func() { httpDebugOut, flagVerbosity = saveOut, saveVerbosity }()

	c := newRESTClient(server.URL + "/v1")
	if _, err := c.paginate(context.Background(), "/things", nil, "items", 0); err != nil {
		t.Fatalf("paginate: %v", err)
	}

	want := "# page 1: 1 items, 1 so far (last page)"
	if !strings.Contains(debug.String(), want) {
		t.Errorf("debug output missing %q; got:\n%s", want, debug.String())
	}
}

// TestPaginateQuietByDefault keeps the progress lines behind the debug flag.
func TestPaginateQuietByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	restore := SetRestTokenSourceForTest(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	defer restore()
	restoreQP := SetRestUserProjectForTest("qp-123")
	defer restoreQP()

	var debug bytes.Buffer
	saveOut, saveVerbosity, saveLogHTTP := httpDebugOut, flagVerbosity, flagLogHTTP
	httpDebugOut, flagVerbosity, flagLogHTTP = &debug, "warning", false
	defer func() { httpDebugOut, flagVerbosity, flagLogHTTP = saveOut, saveVerbosity, saveLogHTTP }()

	c := newRESTClient(server.URL + "/v1")
	if _, err := c.paginate(context.Background(), "/things", nil, "items", 0); err != nil {
		t.Fatalf("paginate: %v", err)
	}
	if debug.Len() != 0 {
		t.Errorf("expected no debug output at default verbosity, got:\n%s", debug.String())
	}
}

func TestEndpointHost(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://securitycenter.googleapis.com/v2", "securitycenter.googleapis.com"},
		{"https://biglake.googleapis.com/iceberg/v1/restcatalog/extensions", "biglake.googleapis.com"},
		{"http://127.0.0.1:8080/v1", "127.0.0.1:8080"},
		{"not-a-url/", "not-a-url"},
	}
	for _, tc := range cases {
		if got := endpointHost(tc.in); got != tc.want {
			t.Errorf("endpointHost(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestErrNoBillingProjectNamesAPI(t *testing.T) {
	err := errNoBillingProject("securitycenter.googleapis.com")
	if !strings.Contains(err.Error(), "securitycenter.googleapis.com") {
		t.Errorf("error does not name the API: %s", err.Error())
	}
}
