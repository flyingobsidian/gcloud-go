package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cloudbuild "google.golang.org/api/cloudbuild/v1"
)

func TestBuildsListRegion(t *testing.T) {
	defer func() { flagBuildsListRegion = "" }()
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "default")
	t.Setenv("CLOUDSDK_BUILDS_REGION", "")

	check := func(name, want string) {
		t.Helper()
		got, err := buildsListRegion()
		if err != nil || got != want {
			t.Errorf("%s: got %q, %v; want %q", name, got, err, want)
		}
	}
	check("default", "global")

	if err := os.MkdirAll(filepath.Join(dir, "configurations"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configurations", "config_default"), []byte("[builds]\nregion = europe-west2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	check("property", "europe-west2")

	flagBuildsListRegion = "us-central1"
	check("flag", "us-central1")
}

func TestListBuilds(t *testing.T) {
	calls := 0
	page := func(_ context.Context, token string) (*cloudbuild.ListBuildsResponse, error) {
		calls++
		start := 0
		if token != "" {
			fmt.Sscanf(token, "%d", &start)
		}
		resp := &cloudbuild.ListBuildsResponse{}
		for i := start; i < 5 && i < start+2; i++ {
			resp.Builds = append(resp.Builds, &cloudbuild.Build{Id: fmt.Sprint(i)})
		}
		if start+2 < 5 {
			resp.NextPageToken = fmt.Sprint(start + 2)
		}
		return resp, nil
	}
	got, err := listBuilds(context.Background(), page, 3)
	if err != nil || len(got) != 3 || calls != 2 {
		t.Errorf("limit 3: got %d builds in %d calls, err %v", len(got), calls, err)
	}
	calls = 0
	if got, err = listBuilds(context.Background(), page, 0); err != nil || len(got) != 5 || calls != 3 {
		t.Errorf("no limit: got %d builds in %d calls, err %v", len(got), calls, err)
	}
}

func TestBuildsDuration(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	cases := []struct{ start, end, want string }{
		{"2026-10-21T10:00:00Z", "2026-10-21T10:01:05.4Z", "1M5S"},
		{"2026-10-21T10:00:00Z", "2026-10-21T11:00:00Z", "1H"},
		{"2026-10-20T10:00:00Z", "2026-10-21T12:00:03Z", "26H3S"},
		{"2026-10-21T10:00:00Z", "2026-10-21T10:00:00Z", ""},
		{"2026-10-21T11:59:30Z", "", "30S"},
		{"", "", "-"},
	}
	for _, tc := range cases {
		if got := buildsDuration(tc.start, tc.end, now); got != tc.want {
			t.Errorf("buildsDuration(%q, %q) = %q, want %q", tc.start, tc.end, got, tc.want)
		}
	}
}

func TestBuildsSourceAndImages(t *testing.T) {
	cases := []struct {
		src  *cloudbuild.Source
		want string
	}{
		{nil, "-"},
		{&cloudbuild.Source{StorageSource: &cloudbuild.StorageSource{Bucket: "b", Object: "src.tgz"}}, "gs://b/src.tgz"},
		{&cloudbuild.Source{RepoSource: &cloudbuild.RepoSource{BranchName: "main"}}, "default@main"},
		{&cloudbuild.Source{GitSource: &cloudbuild.GitSource{Url: "https://github.com/o/r", Revision: "abc"}}, "https://github.com/o/r@abc"},
		{&cloudbuild.Source{ConnectedRepository: &cloudbuild.ConnectedRepository{Repository: "repo", Revision: "v1"}}, "repo@v1"},
		{&cloudbuild.Source{}, "-"},
	}
	for _, tc := range cases {
		if got := buildsSource(tc.src); got != tc.want {
			t.Errorf("buildsSource(%+v) = %q, want %q", tc.src, got, tc.want)
		}
	}
	if got := buildsImages(&cloudbuild.Results{Images: []*cloudbuild.BuiltImage{{Name: "gcr.io/p/a"}, {Name: "gcr.io/p/b"}}}); got != "gcr.io/p/a (+1 more)" {
		t.Errorf("buildsImages = %q", got)
	}
	if got := buildsImages(nil); got != "-" {
		t.Errorf("buildsImages(nil) = %q", got)
	}
}

func TestEmitBuildsListTable(t *testing.T) {
	builds := []*cloudbuild.Build{{
		Id: "b1", CreateTime: "2026-10-21T09:59:58Z", StartTime: "2026-10-21T10:00:00Z", FinishTime: "2026-10-21T10:02:00Z",
		Source: &cloudbuild.Source{StorageSource: &cloudbuild.StorageSource{Bucket: "b", Object: "o.tgz"}},
		Status: "SUCCESS",
	}}
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Skip(err)
	}
	var out bytes.Buffer
	if err := emitBuildsList(&out, &bytes.Buffer{}, builds, "", time.Now(), london); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	want := []string{"ID CREATE_TIME DURATION SOURCE IMAGES STATUS", "b1 2026-10-21T10:59:58+01:00 2M gs://b/o.tgz - SUCCESS"}
	for i, w := range want {
		if i >= len(lines) || strings.Join(strings.Fields(lines[i]), " ") != w {
			t.Errorf("table =\n%s\nwant line %d %q", out.String(), i, w)
		}
	}
	var errOut bytes.Buffer
	out.Reset()
	if err := emitBuildsList(&out, &errOut, nil, "", time.Now(), london); err != nil || errOut.String() != "Listed 0 items.\n" {
		t.Errorf("empty: stderr %q err %v", errOut.String(), err)
	}
}
