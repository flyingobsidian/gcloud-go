package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	storagebatchoperations "google.golang.org/api/storagebatchoperations/v1"
)

// --- gcloud storage batch-operations (#1236) ---

var storageBatchOperationsCmd = &cobra.Command{Use: "batch-operations", Short: "Manage long-running batch storage jobs"}
var storageBatchOpsJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage batch storage jobs"}
var storageBatchOpsBucketOpsCmd = &cobra.Command{Use: "bucket-operations", Short: "Inspect per-bucket sub-operations of a batch job"}

var (
	flagStBoLocation   string
	flagStBoFormat     string
	flagStBoConfigFile string
	flagStBoJob        string
	flagStBoPageSize   int64
)

var (
	// jobs
	storageBatchOpsJobsCancelCmd = &cobra.Command{
		Use: "cancel JOB", Short: "Cancel a batch storage job",
		Args: cobra.ExactArgs(1), RunE: runStBoJobCancel,
	}
	storageBatchOpsJobsCreateCmd = &cobra.Command{
		Use: "create JOB", Short: "Create a batch storage job",
		Args: cobra.ExactArgs(1), RunE: runStBoJobCreate,
	}
	storageBatchOpsJobsDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a batch storage job",
		Args: cobra.ExactArgs(1), RunE: runStBoJobDelete,
	}
	storageBatchOpsJobsDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a batch storage job",
		Args: cobra.ExactArgs(1), RunE: runStBoJobDescribe,
	}
	storageBatchOpsJobsListCmd = &cobra.Command{
		Use: "list", Short: "List batch storage jobs in a location",
		Args: cobra.NoArgs, RunE: runStBoJobList,
	}

	// bucket-operations
	storageBatchOpsBucketOpsDescribeCmd = &cobra.Command{
		Use: "describe BUCKET_OPERATION", Short: "Describe a per-bucket sub-operation of a batch job",
		Args: cobra.ExactArgs(1), RunE: runStBoBucketOpDescribe,
	}
	storageBatchOpsBucketOpsListCmd = &cobra.Command{
		Use: "list", Short: "List per-bucket sub-operations of a batch job",
		Args: cobra.NoArgs, RunE: runStBoBucketOpList,
	}
)

func init() {
	jobs := []*cobra.Command{
		storageBatchOpsJobsCancelCmd, storageBatchOpsJobsCreateCmd,
		storageBatchOpsJobsDeleteCmd, storageBatchOpsJobsDescribeCmd,
		storageBatchOpsJobsListCmd,
	}
	bucketOps := []*cobra.Command{
		storageBatchOpsBucketOpsDescribeCmd, storageBatchOpsBucketOpsListCmd,
	}
	all := append(append([]*cobra.Command{}, jobs...), bucketOps...)
	for _, c := range all {
		c.Flags().StringVar(&flagStBoLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagStBoFormat, "format", "", "Output format")
	}
	storageBatchOpsJobsCreateCmd.Flags().StringVar(&flagStBoConfigFile, "config-file", "", "YAML/JSON file with the Job body (required)")
	_ = storageBatchOpsJobsCreateCmd.MarkFlagRequired("config-file")
	storageBatchOpsJobsListCmd.Flags().Int64Var(&flagStBoPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range bucketOps {
		c.Flags().StringVar(&flagStBoJob, "job", "", "Parent batch job id (required)")
		_ = c.MarkFlagRequired("job")
	}
	storageBatchOpsBucketOpsListCmd.Flags().Int64Var(&flagStBoPageSize, "page-size", 0, "Maximum results per page")

	storageBatchOpsJobsCmd.AddCommand(jobs...)
	storageBatchOpsBucketOpsCmd.AddCommand(bucketOps...)
	storageBatchOperationsCmd.AddCommand(storageBatchOpsJobsCmd, storageBatchOpsBucketOpsCmd)
	storageCmd.AddCommand(storageBatchOperationsCmd)
}

func stBoLocationParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagStBoLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagStBoLocation), nil
}

func stBoJobName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := stBoLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/jobs/%s", parent, id), nil
}

func stBoBucketOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	job, err := stBoJobName(flagStBoJob)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/bucketOperations/%s", job, id), nil
}

func runStBoJobCancel(cmd *cobra.Command, args []string) error {
	name, err := stBoJobName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageBatchOperationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Cancel(name, &storagebatchoperations.CancelJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling job: %w", err)
	}
	fmt.Printf("Cancelled batch job [%s].\n", args[0])
	return nil
}

func runStBoJobCreate(cmd *cobra.Command, args []string) error {
	parent, err := stBoLocationParent()
	if err != nil {
		return err
	}
	body := &storagebatchoperations.Job{}
	if err := loadYAMLOrJSONInto(flagStBoConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageBatchOperationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Create(parent, body).JobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	fmt.Printf("Create batch job [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagStBoFormat)
}

func runStBoJobDelete(cmd *cobra.Command, args []string) error {
	name, err := stBoJobName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageBatchOperationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Printf("Deleted batch job [%s].\n", args[0])
	return nil
}

func runStBoJobDescribe(cmd *cobra.Command, args []string) error {
	name, err := stBoJobName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageBatchOperationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormatted(got, flagStBoFormat)
}

func runStBoJobList(cmd *cobra.Command, args []string) error {
	parent, err := stBoLocationParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageBatchOperationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*storagebatchoperations.Job
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.List(parent).Context(ctx)
		if flagStBoPageSize > 0 {
			call = call.PageSize(flagStBoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing jobs: %w", err)
		}
		all = append(all, resp.Jobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagStBoFormat)
}

func runStBoBucketOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := stBoBucketOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageBatchOperationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.BucketOperations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing bucket-operation: %w", err)
	}
	return emitFormatted(got, flagStBoFormat)
}

func runStBoBucketOpList(cmd *cobra.Command, args []string) error {
	parent, err := stBoJobName(flagStBoJob)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageBatchOperationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*storagebatchoperations.BucketOperation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.BucketOperations.List(parent).Context(ctx)
		if flagStBoPageSize > 0 {
			call = call.PageSize(flagStBoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing bucket-operations: %w", err)
		}
		all = append(all, resp.BucketOperations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagStBoFormat)
}
