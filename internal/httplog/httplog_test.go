package httplog

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCurlGET(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet,
		"https://securitycenter.googleapis.com/v1/organizations/1234/sources/-/findings?filter=state%3D%22ACTIVE%22", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer ya29.super-secret")
	req.Header.Set("X-Goog-User-Project", "my-billing-project")

	got := Curl(req)
	want := `curl -X GET 'https://securitycenter.googleapis.com/v1/organizations/1234/sources/-/findings?filter=state%3D%22ACTIVE%22'` +
		` -H "Authorization: Bearer $(gcloud auth print-access-token)"` +
		` -H 'X-Goog-User-Project: my-billing-project'`
	if got != want {
		t.Errorf("Curl() =\n%s\nwant:\n%s", got, want)
	}
	if strings.Contains(got, "super-secret") {
		t.Error("Curl() leaked the access token")
	}
}

func TestCurlPOSTBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://example.googleapis.com/v1/things",
		strings.NewReader(`{"name":"it's here"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	got := Curl(req)
	want := `curl -X POST 'https://example.googleapis.com/v1/things'` +
		` -H 'Content-Type: application/json'` +
		` --data-binary '{"name":"it'\''s here"}'`
	if got != want {
		t.Errorf("Curl() =\n%s\nwant:\n%s", got, want)
	}
}

func TestTransportLogsAndForwards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Goog-User-Project") != "my-billing-project" {
			t.Errorf("header not forwarded: %q", r.Header.Get("X-Goog-User-Project"))
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	var out bytes.Buffer
	client := &http.Client{Transport: NewTransport(nil, &out)}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/v1/findings", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Goog-User-Project", "my-billing-project")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	logged := out.String()
	if !strings.HasPrefix(logged, "curl -X GET '"+srv.URL+"/v1/findings'") {
		t.Errorf("unexpected log line: %q", logged)
	}
	if !strings.HasSuffix(logged, "\n") {
		t.Error("log line is not newline terminated")
	}
}

func TestTransportNilOutIsSilent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	client := &http.Client{Transport: NewTransport(nil, nil)}
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}
