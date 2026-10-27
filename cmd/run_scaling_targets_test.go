package cmd

import (
	"slices"
	"testing"

	runv2 "google.golang.org/api/run/v2"
)

func TestParseRunUtilization(t *testing.T) {
	for _, tc := range []struct {
		in      string
		want    float64
		max     float64
		restore bool
		bad     bool
	}{
		{"0.6", 0.6, 0.90, false, false},
		{"0.904", 0.9, 0.95, false, false},
		{"disabled", 0, 0.90, false, false},
		{"default", 0, 0.90, true, false},
		{"0.95", 0, 0.90, false, true},
		{"0.95", 0.95, 0.95, false, false},
		{"0.05", 0, 0.90, false, true},
		{"abc", 0, 0.90, false, true},
	} {
		got, restore, err := parseRunUtilization(tc.in, tc.max)
		if tc.bad {
			if err == nil {
				t.Errorf("%q: expected error", tc.in)
			}
			continue
		}
		if err != nil || got != tc.want || restore != tc.restore {
			t.Errorf("%q: got %v %v %v", tc.in, got, restore, err)
		}
	}
}

func TestApplyRunScalingTargets(t *testing.T) {
	tmpl := &runv2.GoogleCloudRunV2RevisionTemplate{Scaling: &runv2.GoogleCloudRunV2RevisionScaling{ConcurrencyUtilization: 0.5}}
	if err := applyRunScalingTargets(tmpl, "disabled", "default"); err != nil {
		t.Fatal(err)
	}
	sc := tmpl.Scaling
	if sc.CpuUtilization != 0 || !slices.Contains(sc.ForceSendFields, "CpuUtilization") {
		t.Errorf("disabled cpu: %+v", sc)
	}
	if sc.ConcurrencyUtilization != 0 || slices.Contains(sc.ForceSendFields, "ConcurrencyUtilization") {
		t.Errorf("default concurrency: %+v", sc)
	}
	if err := applyRunScalingTargets(tmpl, "0.7", ""); err != nil {
		t.Fatal(err)
	}
	if sc.CpuUtilization != 0.7 || slices.Contains(sc.ForceSendFields, "CpuUtilization") {
		t.Errorf("cpu 0.7: %+v", sc)
	}
	if runDeployCmd.Flags().Lookup("scaling-cpu-target") == nil || runDeployCmd.Flags().Lookup("scaling-concurrency-target") == nil {
		t.Error("scaling target flags missing on run deploy")
	}
}
