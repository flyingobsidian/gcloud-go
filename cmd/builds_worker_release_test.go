package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
)

func TestApplyBuildsWPWorkerRelease(t *testing.T) {
	for _, c := range []struct {
		name string
		cmd  *cobra.Command
	}{{"create", buildsWPCreateCmd}, {"update", buildsWPUpdateCmd}} {
		cmd := c.cmd
		f := cmd.Flags().Lookup("worker-release")
		if f == nil {
			t.Fatalf("worker-pools %s missing --worker-release", c.name)
		}

		wp := &cloudbuild.WorkerPool{}
		if applyBuildsWPWorkerRelease(cmd, wp) || wp.PrivatePoolV1Config != nil {
			t.Errorf("%s: unset flag modified the body", c.name)
		}

		if err := cmd.Flags().Set("worker-release", "stable"); err != nil {
			t.Fatal(err)
		}
		wp = &cloudbuild.WorkerPool{PrivatePoolV1Config: &cloudbuild.PrivatePoolV1Config{
			WorkerConfig: &cloudbuild.WorkerConfig{MachineType: "e2-medium", WorkerRelease: "rapid"},
		}}
		if !applyBuildsWPWorkerRelease(cmd, wp) {
			t.Errorf("%s: set flag not applied", c.name)
		}
		wc := wp.PrivatePoolV1Config.WorkerConfig
		if wc.WorkerRelease != "stable" || wc.MachineType != "e2-medium" {
			t.Errorf("%s: workerConfig = %+v", c.name, wc)
		}
		f.Changed = false
		flagBuildsWorkerRelease = ""
	}
}

func TestBuildsWPUpdateMask(t *testing.T) {
	cases := []struct {
		fields []string
		wr     bool
		want   string
	}{
		{[]string{"displayName"}, false, "displayName"},
		{[]string{"displayName"}, true, "displayName,privatePoolV1Config.workerConfig.workerRelease"},
		{[]string{"privatePoolV1Config"}, true, "privatePoolV1Config"},
		{nil, true, "privatePoolV1Config.workerConfig.workerRelease"},
	}
	for _, c := range cases {
		if got := buildsWPUpdateMask(c.fields, c.wr); got != c.want {
			t.Errorf("buildsWPUpdateMask(%v, %v) = %q, want %q", c.fields, c.wr, got, c.want)
		}
	}
}
