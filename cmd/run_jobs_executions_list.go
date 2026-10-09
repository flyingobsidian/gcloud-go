package cmd

import (
	"context"
	"sort"
	"strings"

	runv1 "google.golang.org/api/run/v1"
)

// `gcloud run jobs executions list` uses the Cloud Run v1 (Knative-style)
// API: it lists namespaces/{project}/executions on the regional endpoint,
// filtered to one job by label, and prints apiVersion/kind/metadata/spec/
// status objects. The v2 API returns a different shape, so this command uses
// v1 to match the reference Python gcloud.

// runJobsExecJobLabel is the label Cloud Run sets on executions to name their job.
const runJobsExecJobLabel = "run.googleapis.com/job"

// runJobsExecLabelSelector returns the v1 label selector limiting executions
// to job, or "" to list executions of every job.
func runJobsExecLabelSelector(job string) string {
	if job == "" {
		return ""
	}
	return runJobsExecJobLabel + " = " + job
}

// runJobsExecPager fetches one page of v1 executions for the continue token.
type runJobsExecPager func(ctx context.Context, continueToken string) (*runv1.ListExecutionsResponse, error)

// runJobsListV1Executions pages through every v1 execution. It never returns
// a nil slice, so an empty result renders as [] in JSON. All pages are read
// because gcloud sorts the full list before applying --limit.
func runJobsListV1Executions(ctx context.Context, page runJobsExecPager) ([]*runv1.Execution, error) {
	all := []*runv1.Execution{}
	continueToken := ""
	for {
		resp, err := page(ctx, continueToken)
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Items...)
		if resp.Metadata == nil || resp.Metadata.Continue == "" {
			return all, nil
		}
		continueToken = resp.Metadata.Continue
	}
}

// runJobsExecCondition returns the status condition of the given type, or nil.
func runJobsExecCondition(e *runv1.Execution, condType string) *runv1.GoogleCloudRunV1Condition {
	if e.Status == nil {
		return nil
	}
	for _, c := range e.Status.Conditions {
		if c.Type == condType {
			return c
		}
	}
	return nil
}

// runJobsExecSortKey mirrors gcloud's _ByStartAndCreationTime: whether the
// execution is unstarted, and the time it started (or was created, if
// unstarted or the start time is unknown).
func runJobsExecSortKey(e *runv1.Execution) (bool, string) {
	started := runJobsExecCondition(e, "Started")
	unstarted := started == nil || conditionStatusUnknown(started.Status)
	t := ""
	if e.Metadata != nil {
		t = e.Metadata.CreationTimestamp
	}
	if started != nil && started.LastTransitionTime != "" {
		t = started.LastTransitionTime
	}
	return unstarted, t
}

// conditionStatusUnknown reports whether a condition status is neither
// "True" nor "False", which gcloud treats as unset.
func conditionStatusUnknown(status string) bool {
	s := strings.ToLower(status)
	return s != "true" && s != "false"
}

func runJobsV1ExecName(e *runv1.Execution) string {
	if e.Metadata == nil {
		return ""
	}
	return e.Metadata.Name
}

// runJobsSortExecutions sorts executions as gcloud does: unstarted
// executions first, then newest first by start (or creation) time, with
// ties in name order.
func runJobsSortExecutions(execs []*runv1.Execution) {
	sort.SliceStable(execs, func(i, j int) bool {
		return runJobsV1ExecName(execs[i]) < runJobsV1ExecName(execs[j])
	})
	sort.SliceStable(execs, func(i, j int) bool {
		ui, ti := runJobsExecSortKey(execs[i])
		uj, tj := runJobsExecSortKey(execs[j])
		if ui != uj {
			return ui
		}
		return ti > tj
	})
}
