package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	composer "google.golang.org/api/composer/v1"
)

// --- gcloud composer operations (#1503) ---

var composerOpCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Composer operations"}

var (
	flagComposerOpLocation string
	flagComposerOpFormat   string
	flagComposerOpFilter   string
	flagComposerOpPageSize int64
	flagComposerOpTimeout  time.Duration
)

var (
	composerOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Composer operation record",
		Args: cobra.ExactArgs(1), RunE: runComposerOpDelete,
	}
	composerOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Composer operation",
		Args: cobra.ExactArgs(1), RunE: runComposerOpDescribe,
	}
	composerOpListCmd = &cobra.Command{
		Use: "list", Short: "List Composer operations",
		Args: cobra.NoArgs, RunE: runComposerOpList,
	}
	composerOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Composer operation to complete",
		Args: cobra.ExactArgs(1), RunE: runComposerOpWait,
	}
)

func init() {
	all := []*cobra.Command{composerOpDeleteCmd, composerOpDescribeCmd, composerOpListCmd, composerOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagComposerOpLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagComposerOpFormat, "format", "", "Output format")
	}
	composerOpListCmd.Flags().StringVar(&flagComposerOpFilter, "filter", "", "Server-side filter expression")
	composerOpListCmd.Flags().Int64Var(&flagComposerOpPageSize, "page-size", 0, "Maximum results per page")
	composerOpWaitCmd.Flags().DurationVar(&flagComposerOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait before returning (e.g. 15m, 1h)")

	composerOpCmd.AddCommand(all...)
	composerCmd.AddCommand(composerOpCmd)
}

func composerOpParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return composerLocationParent(project, flagComposerOpLocation), nil
}

func composerOpName(id string) (string, error) {
	parent, err := composerOpParent()
	if err != nil {
		return "", err
	}
	if len(id) > 0 && id[0] == '/' {
		return id[1:], nil
	}
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func runComposerOpDelete(cmd *cobra.Command, args []string) error {
	name, err := composerOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation %s.\n", args[0])
	return nil
}

func runComposerOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := composerOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagComposerOpFormat)
}

func runComposerOpList(cmd *cobra.Command, args []string) error {
	parent, err := composerOpParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*composer.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagComposerOpFilter != "" {
			call = call.Filter(flagComposerOpFilter)
		}
		if flagComposerOpPageSize > 0 {
			call = call.PageSize(flagComposerOpPageSize)
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
	if flagComposerOpFormat != "" {
		return emitFormatted(all, flagComposerOpFormat)
	}
	fmt.Printf("%-50s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-50s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func runComposerOpWait(cmd *cobra.Command, args []string) error {
	name, err := composerOpName(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), flagComposerOpTimeout)
	defer cancel()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	backoff := 2 * time.Second
	for {
		op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation failed: %s", op.Error.Message)
			}
			fmt.Printf("Operation %s completed.\n", args[0])
			return emitFormatted(op, flagComposerOpFormat)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for operation %s: %w", args[0], ctx.Err())
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff = time.Duration(float64(backoff) * 1.5)
		}
	}
}
