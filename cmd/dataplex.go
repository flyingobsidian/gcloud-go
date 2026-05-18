package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataplex "google.golang.org/api/dataplex/v1"
)

var dataplexCmd = &cobra.Command{
	Use:   "dataplex",
	Short: "Manage Cloud Dataplex",
}

var dataplexDatascansCmd = &cobra.Command{
	Use:   "datascans",
	Short: "Manage Dataplex data scans",
}

// --- datascans run ---

var dataplexDatascansRunCmd = &cobra.Command{
	Use:   "run DATASCAN_ID",
	Short: "Run an on-demand data scan",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataplexDatascansRun,
}

// --- datascans jobs ---

var dataplexDatascansJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage data scan jobs",
}

var dataplexDatascansJobsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List data scan jobs",
	Args:  cobra.NoArgs,
	RunE:  runDataplexDatascansJobsList,
}

var dataplexDatascansJobsDescribeCmd = &cobra.Command{
	Use:   "describe JOB_ID",
	Short: "Describe a data scan job",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataplexDatascansJobsDescribe,
}

var (
	flagDataplexLocation  string
	flagDataplexDatascan  string
	flagDataplexJobFormat string
	flagDataplexJobView   string
)

func init() {
	dataplexDatascansRunCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	dataplexDatascansRunCmd.Flags().StringVar(&flagDataplexDatascan, "datascan", "", "Data scan ID (for jobs commands)")

	dataplexDatascansJobsListCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	dataplexDatascansJobsListCmd.Flags().StringVar(&flagDataplexDatascan, "datascan", "", "Data scan ID")
	dataplexDatascansJobsListCmd.Flags().StringVar(&flagDataplexJobFormat, "format", "", "Output format (e.g. json)")

	dataplexDatascansJobsDescribeCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	dataplexDatascansJobsDescribeCmd.Flags().StringVar(&flagDataplexDatascan, "datascan", "", "Data scan ID")
	dataplexDatascansJobsDescribeCmd.Flags().StringVar(&flagDataplexJobView, "view", "", "Job view (BASIC or FULL)")

	dataplexDatascansJobsCmd.AddCommand(dataplexDatascansJobsListCmd)
	dataplexDatascansJobsCmd.AddCommand(dataplexDatascansJobsDescribeCmd)
	dataplexDatascansCmd.AddCommand(dataplexDatascansRunCmd)
	dataplexDatascansCmd.AddCommand(dataplexDatascansJobsCmd)
	dataplexCmd.AddCommand(dataplexDatascansCmd)
	rootCmd.AddCommand(dataplexCmd)
}

func resolveDataplexLocation() (string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", err
	}
	location := flagDataplexLocation
	if location == "" {
		_, region, err := resolveRegion()
		if err != nil || region == "" {
			return "", "", fmt.Errorf("--location is required")
		}
		location = region
	}
	return project, location, nil
}

func datascanName(project, location, datascanID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/dataScans/%s", project, location, datascanID)
}

func runDataplexDatascansRun(cmd *cobra.Command, args []string) error {
	project, location, err := resolveDataplexLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.DataplexService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := datascanName(project, location, args[0])
	resp, err := svc.Projects.Locations.DataScans.Run(name, &dataplex.GoogleCloudDataplexV1RunDataScanRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running data scan: %w", err)
	}

	return formatOutput(resp, "")
}

func runDataplexDatascansJobsList(cmd *cobra.Command, args []string) error {
	project, location, err := resolveDataplexLocation()
	if err != nil {
		return err
	}

	datascanID := flagDataplexDatascan
	if datascanID == "" {
		return fmt.Errorf("--datascan is required")
	}

	ctx := context.Background()
	svc, err := gcp.DataplexService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := datascanName(project, location, datascanID)

	var allJobs []*dataplex.GoogleCloudDataplexV1DataScanJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.DataScans.Jobs.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing data scan jobs: %w", err)
		}
		allJobs = append(allJobs, resp.DataScanJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagDataplexJobFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allJobs)
	}

	fmt.Printf("%-60s %-15s %s\n", "NAME", "STATE", "START_TIME")
	for _, j := range allJobs {
		fmt.Printf("%-60s %-15s %s\n", j.Name, j.State, j.StartTime)
	}
	return nil
}

func runDataplexDatascansJobsDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := resolveDataplexLocation()
	if err != nil {
		return err
	}

	datascanID := flagDataplexDatascan
	if datascanID == "" {
		return fmt.Errorf("--datascan is required")
	}

	ctx := context.Background()
	svc, err := gcp.DataplexService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s/jobs/%s", datascanName(project, location, datascanID), args[0])
	call := svc.Projects.Locations.DataScans.Jobs.Get(name).Context(ctx)
	if flagDataplexJobView != "" {
		call = call.View(flagDataplexJobView)
	}
	job, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing data scan job: %w", err)
	}

	return formatOutput(job, "")
}
