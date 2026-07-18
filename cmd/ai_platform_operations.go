package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform operations (#985) ---

var aiPlatformOpsCmd = &cobra.Command{Use: "operations", Short: "Manage AI Platform long-running operations"}

var (
	flagAIPlatformOpsFormat    string
	flagAIPlatformOpsPageSize  int64
	flagAIPlatformOpsFilter    string
	flagAIPlatformOpsPollEvery time.Duration
	flagAIPlatformOpsTimeout   time.Duration
)

var (
	aiPlatformOpsCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an AI Platform operation",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformOpsCancel,
	}
	aiPlatformOpsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an AI Platform operation",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformOpsDescribe,
	}
	aiPlatformOpsListCmd = &cobra.Command{
		Use: "list", Short: "List AI Platform operations",
		Args: cobra.NoArgs, RunE: runAIPlatformOpsList,
	}
	aiPlatformOpsWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Poll an AI Platform operation until it completes",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformOpsWait,
	}
)

func init() {
	all := []*cobra.Command{
		aiPlatformOpsCancelCmd, aiPlatformOpsDescribeCmd, aiPlatformOpsListCmd, aiPlatformOpsWaitCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIPlatformOpsFormat, "format", "", "Output format")
	}
	aiPlatformOpsListCmd.Flags().Int64Var(&flagAIPlatformOpsPageSize, "page-size", 0, "Maximum results per page")
	aiPlatformOpsListCmd.Flags().StringVar(&flagAIPlatformOpsFilter, "filter", "", "List filter expression")

	aiPlatformOpsWaitCmd.Flags().DurationVar(&flagAIPlatformOpsPollEvery, "poll-interval", 5*time.Second,
		"Interval between operation polls")
	aiPlatformOpsWaitCmd.Flags().DurationVar(&flagAIPlatformOpsTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to complete")

	aiPlatformOpsCmd.AddCommand(all...)
	aiPlatformCmd.AddCommand(aiPlatformOpsCmd)
}

func runAIPlatformOpsCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Operations.Cancel(mlOperationName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runAIPlatformOpsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Operations.Get(mlOperationName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagAIPlatformOpsFormat)
}

func runAIPlatformOpsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*ml.GoogleLongrunning__Operation
	pageToken := ""
	for {
		call := svc.Projects.Operations.List(mlProjectPath(project)).Context(ctx)
		if flagAIPlatformOpsPageSize > 0 {
			call = call.PageSize(flagAIPlatformOpsPageSize)
		}
		if flagAIPlatformOpsFilter != "" {
			call = call.Filter(flagAIPlatformOpsFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAIPlatformOpsFormat)
}

func runAIPlatformOpsWait(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := mlOperationName(project, args[0])
	deadline := time.Now().Add(flagAIPlatformOpsTimeout)
	for {
		op, err := svc.Projects.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation %s failed: %s (code %d)", args[0], op.Error.Message, op.Error.Code)
			}
			return emitFormatted(op, flagAIPlatformOpsFormat)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for operation [%s]", flagAIPlatformOpsTimeout, args[0])
		}
		time.Sleep(flagAIPlatformOpsPollEvery)
	}
}
