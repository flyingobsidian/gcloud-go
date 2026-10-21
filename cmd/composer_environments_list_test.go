package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	composer "google.golang.org/api/composer/v1"
)

func TestComposerEnvListLocations(t *testing.T) {
	defer func() { flagComposerEnvListLocations, flagComposerEnvLocation = nil, "" }()
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	t.Setenv("CLOUDSDK_ACTIVE_CONFIG_NAME", "default")
	t.Setenv("CLOUDSDK_COMPOSER_LOCATION", "")

	if _, err := composerEnvListLocations(); !errors.Is(err, errComposerNoLocations) {
		t.Errorf("no locations: err = %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "configurations"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configurations", "config_default"), []byte("[composer]\nlocation = europe-west2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	check := func(name, want string) {
		t.Helper()
		got, err := composerEnvListLocations()
		if err != nil || strings.Join(got, ",") != want {
			t.Errorf("%s: got %v, %v; want %s", name, got, err, want)
		}
	}
	check("property", "europe-west2")
	flagComposerEnvLocation = "us-east1"
	check("hidden --location", "us-east1")
	flagComposerEnvListLocations = []string{"us-central1", "asia-east1"}
	check("--locations", "us-central1,asia-east1")
}

func TestListComposerEnvs(t *testing.T) {
	envs := map[string][]*composer.ListEnvironmentsResponse{
		"projects/p/locations/a": {
			{Environments: []*composer.Environment{{Name: "projects/p/locations/a/environments/one"}}, NextPageToken: "1"},
			{Environments: []*composer.Environment{{Name: "projects/p/locations/a/environments/two"}}},
		},
		"projects/p/locations/b": {
			{Environments: []*composer.Environment{{Name: "projects/p/locations/b/environments/three"}}},
		},
	}
	page := func(_ context.Context, parent, token string) (*composer.ListEnvironmentsResponse, error) {
		i := 0
		if token != "" {
			i = 1
		}
		return envs[parent][i], nil
	}
	names := func(es []*composer.Environment) string {
		var n []string
		for _, e := range es {
			n = append(n, e.Name[strings.LastIndex(e.Name, "/")+1:])
		}
		return strings.Join(n, ",")
	}
	parents := []string{"projects/p/locations/a", "projects/p/locations/b"}
	got, err := listComposerEnvs(context.Background(), parents, page, "", 0)
	if err != nil || names(got) != "one,two,three" {
		t.Errorf("all: %s, %v", names(got), err)
	}
	if got, err = listComposerEnvs(context.Background(), parents, page, "", 2); err != nil || names(got) != "one,two" {
		t.Errorf("limit 2: %s, %v", names(got), err)
	}
	if got, err = listComposerEnvs(context.Background(), parents, page, "three", 0); err != nil || names(got) != "three" {
		t.Errorf("filter: %s, %v", names(got), err)
	}
	empty := func(context.Context, string, string) (*composer.ListEnvironmentsResponse, error) {
		return &composer.ListEnvironmentsResponse{}, nil
	}
	if got, err = listComposerEnvs(context.Background(), parents, empty, "", 0); err != nil || got == nil || len(got) != 0 {
		t.Errorf("empty: %v, %v", got, err)
	}
}

func TestComposerEnvListFlags(t *testing.T) {
	if composerEnvListCmd.Flags().Lookup("locations") == nil {
		t.Fatal("composer environments list missing --locations")
	}
	if f := composerEnvListCmd.Flags().Lookup("location"); f == nil || !f.Hidden {
		t.Error("composer environments list --location should be a hidden alias")
	} else if _, required := f.Annotations["cobra_annotation_bash_completion_one_required_flag"]; required {
		t.Error("composer environments list --location should not be required")
	}
	if f := composerEnvDescribeCmd.Flags().Lookup("location"); f == nil || f.Annotations["cobra_annotation_bash_completion_one_required_flag"] == nil {
		t.Error("composer environments describe --location should stay required")
	}
}
