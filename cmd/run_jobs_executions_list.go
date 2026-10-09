package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	runv1 "google.golang.org/api/run/v1"
)

// runJobsExecDefaultPageSize matches gcloud's ListExecutions page size.
const runJobsExecDefaultPageSize = 100

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

// runJobsSortAndLimit sorts executions as gcloud does, then keeps at most
// limit of them (limit <= 0 means no limit).
func runJobsSortAndLimit(execs []*runv1.Execution, limit int64) []*runv1.Execution {
	runJobsSortExecutions(execs)
	if limit > 0 && int64(len(execs)) > limit {
		return execs[:limit]
	}
	return execs
}

func runJobsExecList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV1Service(ctx, flagAccount, flagRunJobsRegion)
	if err != nil {
		return err
	}
	pageSize := flagRunJobsExecPageSize
	if pageSize <= 0 {
		pageSize = runJobsExecDefaultPageSize
	}
	selector := runJobsExecLabelSelector(flagRunJobsExecJob)
	page := func(ctx context.Context, continueToken string) (*runv1.ListExecutionsResponse, error) {
		call := svc.Namespaces.Executions.List("namespaces/" + project).Context(ctx).Limit(pageSize)
		if selector != "" {
			call = call.LabelSelector(selector)
		}
		if continueToken != "" {
			call = call.Continue(continueToken)
		}
		return call.Do()
	}
	all, err := runJobsListV1Executions(ctx, page)
	if err != nil {
		return fmt.Errorf("listing executions: %w", err)
	}
	return emitFormatted(runJobsSortAndLimit(all, flagRunJobsExecLimit), flagRunJobsFormat)
}
