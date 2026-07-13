package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

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

// --- datascans describe ---

var dataplexDatascansDescribeCmd = &cobra.Command{
	Use:   "describe DATASCAN_ID",
	Short: "Describe a data scan",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataplexDatascansDescribe,
}

// --- datascans list ---

var dataplexDatascansListCmd = &cobra.Command{
	Use:   "list",
	Short: "List data scans",
	Args:  cobra.NoArgs,
	RunE:  runDataplexDatascansList,
}

var (
	flagDataplexListFormat string
	flagDataplexListURI    bool
)

// --- datascans create ---

var dataplexDatascansCreateCmd = &cobra.Command{
	Use:   "create DATASCAN_ID",
	Short: "Create a data scan",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataplexDatascansCreate,
}

var (
	flagDataplexDataSource string
	flagDataplexScanType   string
)

// --- datascans delete ---

var dataplexDatascansDeleteCmd = &cobra.Command{
	Use:   "delete DATASCAN_ID",
	Short: "Delete a data scan",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataplexDatascansDelete,
}

var (
	flagDataplexLocation  string
	flagDataplexDatascan  string
	flagDataplexJobFormat string
	flagDataplexJobView   string
	flagDataplexJobURI    bool
)

func init() {
	dataplexDatascansRunCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	dataplexDatascansRunCmd.Flags().StringVar(&flagDataplexDatascan, "datascan", "", "Data scan ID (for jobs commands)")

	dataplexDatascansJobsListCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	dataplexDatascansJobsListCmd.Flags().StringVar(&flagDataplexDatascan, "datascan", "", "Data scan ID")
	dataplexDatascansJobsListCmd.Flags().StringVar(&flagDataplexJobFormat, "format", "", "Output format (e.g. json)")
	dataplexDatascansJobsListCmd.Flags().BoolVar(&flagDataplexJobURI, "uri", false, "Print resource names")

	dataplexDatascansJobsDescribeCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	dataplexDatascansJobsDescribeCmd.Flags().StringVar(&flagDataplexDatascan, "datascan", "", "Data scan ID")
	dataplexDatascansJobsDescribeCmd.Flags().StringVar(&flagDataplexJobView, "view", "", "Job view (BASIC or FULL)")

	dataplexDatascansDescribeCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	dataplexDatascansListCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scans")
	dataplexDatascansListCmd.Flags().StringVar(&flagDataplexListFormat, "format", "", "Output format (e.g. json)")
	dataplexDatascansListCmd.Flags().BoolVar(&flagDataplexListURI, "uri", false, "Print resource names")
	dataplexDatascansCreateCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	dataplexDatascansCreateCmd.Flags().StringVar(&flagDataplexDataSource, "data-source", "", "Data source entity (required)")
	dataplexDatascansCreateCmd.MarkFlagRequired("data-source")
	dataplexDatascansCreateCmd.Flags().StringVar(&flagDataplexScanType, "type", "data-quality", "Scan type (data-quality or data-profile)")
	dataplexDatascansDeleteCmd.Flags().StringVar(&flagDataplexLocation, "location", "", "Location of the data scan")
	// --quiet is provided by the global persistent flag

	dataplexDatascansJobsCmd.AddCommand(dataplexDatascansJobsListCmd)
	dataplexDatascansJobsCmd.AddCommand(dataplexDatascansJobsDescribeCmd)
	dataplexDatascansCmd.AddCommand(dataplexDatascansRunCmd)
	dataplexDatascansCmd.AddCommand(dataplexDatascansDescribeCmd)
	dataplexDatascansCmd.AddCommand(dataplexDatascansListCmd)
	dataplexDatascansCmd.AddCommand(dataplexDatascansCreateCmd)
	dataplexDatascansCmd.AddCommand(dataplexDatascansDeleteCmd)
	dataplexDatascansCmd.AddCommand(dataplexDatascansJobsCmd)
	dataplexCmd.AddCommand(dataplexDatascansCmd)

	// gcloud-python dataplex subgroups not yet implemented (#541).
	for name, short := range map[string]string{
		"aspect-types":     "Manage aspect types",
		"assets":           "Manage Dataplex assets",
		"context":          "Manage Dataplex context resources",
		"encryption-config": "Manage Dataplex encryption config",
		"entries":          "Manage Dataplex entries",
		"entry-groups":     "Manage entry groups",
		"entry-types":      "Manage entry types",
		"glossaries":       "Manage Dataplex glossaries",
		"lakes":            "Manage Dataplex lakes",
		"metadata-jobs":    "Manage metadata jobs",
		"tasks":            "Manage Dataplex tasks",
		"zones":            "Manage Dataplex zones",
	} {
		registerStubGroup(dataplexCmd, name, short, "list", "describe")
	}

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

	if flagDataplexJobURI {
		for _, job := range allJobs {
			fmt.Println(job.Name)
		}
		return nil
	}

	if flagDataplexJobFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allJobs)
	}

	fmt.Printf("%-40s %-15s %-25s %-25s\n", "JOB_ID", "STATE", "START_TIME", "END_TIME")
	for _, j := range allJobs {
		fmt.Printf("%-40s %-15s %-25s %-25s\n", path.Base(j.Name), j.State, j.StartTime, j.EndTime)
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
		call = call.View(strings.ToUpper(flagDataplexJobView))
	}
	job, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing data scan job: %w", err)
	}

	return formatOutput(job, "")
}

func runDataplexDatascansDescribe(cmd *cobra.Command, args []string) error {
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
	scan, err := svc.Projects.Locations.DataScans.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing data scan: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(scan)
}

func runDataplexDatascansList(cmd *cobra.Command, args []string) error {
	project, location, err := resolveDataplexLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.DataplexService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	var allScans []*dataplex.GoogleCloudDataplexV1DataScan
	pageToken := ""
	for {
		call := svc.Projects.Locations.DataScans.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing data scans: %w", err)
		}
		allScans = append(allScans, resp.DataScans...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagDataplexListURI {
		for _, ds := range allScans {
			fmt.Println(ds.Name)
		}
		return nil
	}

	if flagDataplexListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allScans)
	}

	fmt.Printf("%-30s %-15s %-25s %-20s\n", "NAME", "STATE", "CREATE_TIME", "TYPE")
	for _, s := range allScans {
		scanType := ""
		if s.DataQualitySpec != nil {
			scanType = "DATA_QUALITY"
		} else if s.DataProfileSpec != nil {
			scanType = "DATA_PROFILE"
		}
		fmt.Printf("%-30s %-15s %-25s %-20s\n", path.Base(s.Name), s.State, s.CreateTime, scanType)
	}
	return nil
}

func runDataplexDatascansCreate(cmd *cobra.Command, args []string) error {
	project, location, err := resolveDataplexLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.DataplexService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	scan := &dataplex.GoogleCloudDataplexV1DataScan{
		Data: &dataplex.GoogleCloudDataplexV1DataSource{
			Entity: flagDataplexDataSource,
		},
	}

	switch strings.ToLower(flagDataplexScanType) {
	case "data-quality":
		scan.DataQualitySpec = &dataplex.GoogleCloudDataplexV1DataQualitySpec{}
	case "data-profile":
		scan.DataProfileSpec = &dataplex.GoogleCloudDataplexV1DataProfileSpec{}
	default:
		return fmt.Errorf("unsupported scan type: %s (use data-quality or data-profile)", flagDataplexScanType)
	}

	op, err := svc.Projects.Locations.DataScans.Create(parent, scan).DataScanId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating data scan: %w", err)
	}

	fmt.Printf("Created data scan [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

func runDataplexDatascansDelete(cmd *cobra.Command, args []string) error {
	project, location, err := resolveDataplexLocation()
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Printf("You are about to delete data scan [%s]. This action cannot be undone.\n", args[0])
		fmt.Print("Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := gcp.DataplexService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := datascanName(project, location, args[0])
	if _, err := svc.Projects.Locations.DataScans.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting data scan: %w", err)
	}

	fmt.Printf("Deleted data scan [%s].\n", args[0])
	return nil
}
