package cmd

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	runv2 "google.golang.org/api/run/v2"
)

// --- --scaling-cpu-target / --scaling-concurrency-target on run deploy (#1940) ---
//
// Promoted to GA in gcloud-python 587.0.0, which also lowered the CPU target
// cap from 0.95 to 0.90. Python (v1 API) sets or, with "default", deletes the
// run.googleapis.com/scaling-{cpu,concurrency}-target template annotations;
// gcloud-go deploys through the v2 API, which rejects run.googleapis.com
// annotations and carries the same values in template.scaling
// cpuUtilization / concurrencyUtilization.

var (
	flagRunDeployScalingCPUTarget         string
	flagRunDeployScalingConcurrencyTarget string
)

func init() {
	runDeployCmd.Flags().StringVar(&flagRunDeployScalingCPUTarget, "scaling-cpu-target", "",
		`CPU utilization target threshold for scaling up new instances (0.1-0.90, "default" to unset, "disabled" to disable)`)
	runDeployCmd.Flags().StringVar(&flagRunDeployScalingConcurrencyTarget, "scaling-concurrency-target", "",
		`Concurrency utilization target threshold for scaling up new instances (0.1-0.95, "default" to unset, "disabled" to disable)`)
}

// parseRunUtilization mirrors ServiceUtilizationValue: it returns the
// utilization (0 for "disabled"), or restoreDefault for "default". Values
// are rounded to two decimal places as the API does.
func parseRunUtilization(value string, max float64) (utilization float64, restoreDefault bool, err error) {
	if value == "default" {
		return 0, true, nil
	}
	if value == "disabled" {
		return 0, false, nil
	}
	u, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false, fmt.Errorf("Utilization value %s is not a decimal, 'default', or 'disabled'.", value)
	}
	if u != 0 && u < 0.1 {
		return 0, false, fmt.Errorf("Utilization value %s is less than 0.1.", value)
	}
	if u > max {
		return 0, false, fmt.Errorf("Utilization value %s is greater than %.2f.", value, max)
	}
	return math.Round(u*100) / 100, false, nil
}

// applyRunScalingTargets sets, disables (explicit 0) or clears the
// utilization targets on the revision template for the flags given.
func applyRunScalingTargets(tmpl *runv2.GoogleCloudRunV2RevisionTemplate, cpu, concurrency string) error {
	for _, t := range []struct {
		value, field string
		max          float64
	}{
		{cpu, "CpuUtilization", 0.90},
		{concurrency, "ConcurrencyUtilization", 0.95},
	} {
		if t.value == "" {
			continue
		}
		u, restore, err := parseRunUtilization(t.value, t.max)
		if err != nil {
			return err
		}
		if tmpl.Scaling == nil {
			tmpl.Scaling = &runv2.GoogleCloudRunV2RevisionScaling{}
		}
		sc := tmpl.Scaling
		if t.field == "CpuUtilization" {
			sc.CpuUtilization = u
		} else {
			sc.ConcurrencyUtilization = u
		}
		sc.ForceSendFields = slices.DeleteFunc(sc.ForceSendFields, func(f string) bool { return f == t.field })
		if !restore && u == 0 {
			sc.ForceSendFields = append(sc.ForceSendFields, t.field)
		}
	}
	return nil
}
