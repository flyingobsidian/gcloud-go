package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
	flagDataflowRegion     string
	flagDataflowListFormat string
	flagDataflowListFilter string
	flagDataflowListStatus string
)

func init() {
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowRegion, "region", "", "Region of the jobs")
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowListFormat, "format", "", "Output format (e.g. json)")
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowListFilter, "filter", "", "Filter expression")
	dataflowJobsListCmd.Flags().StringVar(&flagDataflowListStatus, "status", "", "Filter by job status (active, all, terminated)")

	dataflowJobsDescribeCmd.Flags().StringVar(&flagDataflowRegion, "region", "", "Region of the job")
	dataflowJobsCancelCmd.Flags().StringVar(&flagDataflowRegion, "region", "", "Region of the job")

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

	call := svc.Projects.Locations.Jobs.List(project, region).Context(ctx)
	if flagDataflowListFilter != "" {
		call = call.Filter(flagDataflowListFilter)
	}
	if flagDataflowListStatus != "" {
		call = call.Filter(flagDataflowListStatus)
	}

	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing dataflow jobs: %w", err)
	}

	if flagDataflowListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp.Jobs)
	}

	fmt.Printf("%-40s %-15s %-20s %-10s\n", "JOB_ID", "NAME", "TYPE", "STATE")
	for _, j := range resp.Jobs {
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

	job, err := svc.Projects.Locations.Jobs.Get(project, region, args[0]).Context(ctx).Do()
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

	// Cancel by updating the job's requested state to JOB_STATE_CANCELLED.
	_, err = svc.Projects.Locations.Jobs.Update(project, region, args[0], &dataflow.Job{
		RequestedState: "JOB_STATE_CANCELLED",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling dataflow job: %w", err)
	}

	fmt.Printf("Cancelled job [%s].\n", args[0])
	return nil
}
