package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataflow "google.golang.org/api/dataflow/v1b3"
)

var dataflowCmd = &cobra.Command{
	Use:   "dataflow",
	Short: "Manage Google Cloud Dataflow",
}

var dataflowJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage Dataflow jobs",
}

var dataflowJobsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Dataflow jobs",
	Args:  cobra.NoArgs,
	RunE:  runDataflowJobsList,
}

var dataflowJobsDescribeCmd = &cobra.Command{
	Use:   "describe JOB_ID",
	Short: "Describe a Dataflow job",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataflowJobsDescribe,
}

var dataflowJobsCancelCmd = &cobra.Command{
	Use:   "cancel JOB_ID",
	Short: "Cancel a running Dataflow job",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataflowJobsCancel,
}

var (
	flagDataflowRegion       string
	flagDataflowListFormat   string
	flagDataflowListFilter   string
	flagDataflowListStatus   string
	flagDataflowCreatedAfter string
	flagDataflowCreatedBefore string
	flagDataflowDescribeFull bool
	flagDataflowCancelForce  bool
)

func init() {
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowRegion, "region", "", "Region of the jobs")
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowListFormat, "format", "", "Output format (e.g. json)")
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowListFilter, "filter", "", "Client-side filter by job name (substring match)")
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowListStatus, "status", "", "Server-side status filter: active, all, or terminated")
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowCreatedAfter, "created-after", "", "Filter jobs created after this time (RFC 3339)")
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowCreatedBefore, "created-before", "", "Filter jobs created before this time (RFC 3339)")

	dataflowJobsDescribeCmd.Flags().StringVar(&flagDataflowRegion, "region", "", "Region of the job")
	dataflowJobsDescribeCmd.Flags().BoolVar(&flagDataflowDescribeFull, "full", false, "Show full job details including steps")

	dataflowJobsCancelCmd.Flags().StringVar(&flagDataflowRegion, "region", "", "Region of the job")
	dataflowJobsCancelCmd.Flags().BoolVar(&flagDataflowCancelForce, "force", false, "Force drain the job instead of cancelling")

	dataflowJobsCmd.AddCommand(dataflowJobsListCmd)
	dataflowJobsCmd.AddCommand(dataflowJobsDescribeCmd)
	dataflowJobsCmd.AddCommand(dataflowJobsCancelCmd)
	dataflowCmd.AddCommand(dataflowJobsCmd)
	rootCmd.AddCommand(dataflowCmd)
}

func resolveDataflowRegion() (string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", err
	}
	region := flagDataflowRegion
	if region == "" {
		_, r, err := resolveRegion()
		if err != nil || r == "" {
			return "", "", fmt.Errorf("--region is required")
		}
		region = r
	}
	return project, region, nil
}

func runDataflowJobsList(cmd *cobra.Command, args []string) error {
	project, region, err := resolveDataflowRegion()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Map --status to the API's filter enum (UNKNOWN, ALL, TERMINATED, ACTIVE).
	apiFilter := statusToAPIFilter(flagDataflowListStatus)

	var allJobs []*dataflow.Job
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.List(project, region).Context(ctx)
		if apiFilter != "" {
			call = call.Filter(apiFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing dataflow jobs: %w", err)
		}
		allJobs = append(allJobs, resp.Jobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	// Client-side filters.
	allJobs = filterDataflowJobs(allJobs)

	if flagDataflowListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allJobs)
	}

	fmt.Printf("%-40s %-15s %-20s %-10s\n", "JOB_ID", "NAME", "TYPE", "STATE")
	for _, j := range allJobs {
		fmt.Printf("%-40s %-15s %-20s %-10s\n", j.Id, j.Name, j.Type, j.CurrentState)
	}
	return nil
}

func runDataflowJobsDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := resolveDataflowRegion()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}

	call := svc.Projects.Locations.Jobs.Get(project, region, args[0]).Context(ctx)
	if flagDataflowDescribeFull {
		call = call.View("JOB_VIEW_ALL")
	}
	job, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing dataflow job: %w", err)
	}

	return formatOutput(job, "")
}

func runDataflowJobsCancel(cmd *cobra.Command, args []string) error {
	project, region, err := resolveDataflowRegion()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.DataflowService(ctx, flagAccount)
	if err != nil {
		return err
	}

	state := "JOB_STATE_CANCELLED"
	if flagDataflowCancelForce {
		state = "JOB_STATE_DRAINED"
	}
	_, err = svc.Projects.Locations.Jobs.Update(project, region, args[0], &dataflow.Job{
		RequestedState: state,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling dataflow job: %w", err)
	}

	fmt.Printf("Cancelled job [%s].\n", args[0])
	return nil
}

// statusToAPIFilter maps the user-facing --status flag to the Dataflow API
// filter enum value. The API accepts UNKNOWN, ALL, TERMINATED, or ACTIVE.
func statusToAPIFilter(status string) string {
	switch strings.ToLower(status) {
	case "active":
		return "ACTIVE"
	case "terminated":
		return "TERMINATED"
	case "all":
		return "ALL"
	case "":
		return ""
	default:
		return strings.ToUpper(status)
	}
}

// filterDataflowJobs applies client-side filters (--filter for name matching,
// --created-after/--created-before for time range) to the job list.
func filterDataflowJobs(jobs []*dataflow.Job) []*dataflow.Job {
	needsFilter := flagDataflowListFilter != "" ||
		flagDataflowCreatedAfter != "" ||
		flagDataflowCreatedBefore != ""
	if !needsFilter {
		return jobs
	}

	nameFilter := strings.ToLower(flagDataflowListFilter)

	var filtered []*dataflow.Job
	for _, j := range jobs {
		if nameFilter != "" && !strings.Contains(strings.ToLower(j.Name), nameFilter) {
			continue
		}
		if flagDataflowCreatedAfter != "" || flagDataflowCreatedBefore != "" {
			ct, err := time.Parse(time.RFC3339, j.CreateTime)
			if err != nil {
				filtered = append(filtered, j)
				continue
			}
			if flagDataflowCreatedAfter != "" {
				after, err := time.Parse(time.RFC3339, flagDataflowCreatedAfter)
				if err == nil && ct.Before(after) {
					continue
				}
			}
			if flagDataflowCreatedBefore != "" {
				before, err := time.Parse(time.RFC3339, flagDataflowCreatedBefore)
				if err == nil && ct.After(before) {
					continue
				}
			}
		}
		filtered = append(filtered, j)
	}
	return filtered
}
