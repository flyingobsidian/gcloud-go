package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab executions (#1497) ---

var colabExecCmd = &cobra.Command{Use: "executions", Short: "Manage Colab Enterprise notebook executions"}

var (
	flagColabExecRegion     string
	flagColabExecFormat     string
	flagColabExecConfigFile string
	flagColabExecFilter     string
	flagColabExecOrderBy    string
	flagColabExecPageSize   int64
	flagColabExecView       string
)

var (
	colabExecCreateCmd = &cobra.Command{
		Use: "create EXECUTION", Short: "Create a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: runColabExecCreate,
	}
	colabExecDeleteCmd = &cobra.Command{
		Use: "delete EXECUTION", Short: "Delete a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: runColabExecDelete,
	}
	colabExecDescribeCmd = &cobra.Command{
		Use: "describe EXECUTION", Short: "Describe a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: runColabExecDescribe,
	}
	colabExecListCmd = &cobra.Command{
		Use: "list", Short: "List notebook execution jobs",
		Args: cobra.NoArgs, RunE: runColabExecList,
	}
)

func init() {
	all := []*cobra.Command{colabExecCreateCmd, colabExecDeleteCmd, colabExecDescribeCmd, colabExecListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagColabExecRegion, "region", "", "Region where the execution job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagColabExecFormat, "format", "", "Output format")
	}
	colabExecCreateCmd.Flags().StringVar(&flagColabExecConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the NotebookExecutionJob body (required)")
	_ = colabExecCreateCmd.MarkFlagRequired("config-file")
	colabExecListCmd.Flags().StringVar(&flagColabExecFilter, "filter", "", "Server-side filter expression")
	colabExecListCmd.Flags().StringVar(&flagColabExecOrderBy, "order-by", "", "Order-by expression")
	colabExecListCmd.Flags().Int64Var(&flagColabExecPageSize, "page-size", 0, "Maximum results per page")
	colabExecListCmd.Flags().StringVar(&flagColabExecView, "view", "", "NotebookExecutionJobView")

	colabExecCmd.AddCommand(all...)
	colabCmd.AddCommand(colabExecCmd)
}

func colabExecParent() (string, error) {
	return colabParent(flagColabExecRegion)
}

func colabExecName(id string) (string, error) {
	parent, err := colabExecParent()
	if err != nil {
		return "", err
	}
	return colabChild("notebookExecutionJobs", id, parent), nil
}

func runColabExecCreate(cmd *cobra.Command, args []string) error {
	parent, err := colabExecParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1NotebookExecutionJob{}
	if err := loadYAMLOrJSONInto(flagColabExecConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabExecRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookExecutionJobs.Create(parent, body).NotebookExecutionJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notebook execution job: %w", err)
	}
	fmt.Printf("Create request issued for execution [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabExecFormat)
}

func runColabExecDelete(cmd *cobra.Command, args []string) error {
	name, err := colabExecName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabExecRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookExecutionJobs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook execution job: %w", err)
	}
	fmt.Printf("Delete request issued for execution [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabExecFormat)
}

func runColabExecDescribe(cmd *cobra.Command, args []string) error {
	name, err := colabExecName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabExecRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.NotebookExecutionJobs.Get(name).Context(ctx)
	if flagColabExecView != "" {
		call = call.View(flagColabExecView)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing notebook execution job: %w", err)
	}
	return emitFormatted(got, flagColabExecFormat)
}

func runColabExecList(cmd *cobra.Command, args []string) error {
	parent, err := colabExecParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabExecRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1NotebookExecutionJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.NotebookExecutionJobs.List(parent).Context(ctx)
		if flagColabExecFilter != "" {
			call = call.Filter(flagColabExecFilter)
		}
		if flagColabExecOrderBy != "" {
			call = call.OrderBy(flagColabExecOrderBy)
		}
		if flagColabExecPageSize > 0 {
			call = call.PageSize(flagColabExecPageSize)
		}
		if flagColabExecView != "" {
			call = call.View(flagColabExecView)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook execution jobs: %w", err)
		}
		all = append(all, resp.NotebookExecutionJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagColabExecFormat)
}
