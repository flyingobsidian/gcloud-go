package cmd

import (
	"context"
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
