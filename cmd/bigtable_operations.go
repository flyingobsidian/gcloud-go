package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable operations (#1488) ---

var bigtableOpCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Bigtable operations"}

var (
	flagBTOpFormat   string
	flagBTOpFilter   string
	flagBTOpPageSize int64
)

var (
	bigtableOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Bigtable operation",
		Args: cobra.ExactArgs(1), RunE: runBTOpDescribe,
	}
	bigtableOpListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable operations for the current project",
		Args: cobra.NoArgs, RunE: runBTOpList,
	}
)

func init() {
	for _, c := range []*cobra.Command{bigtableOpDescribeCmd, bigtableOpListCmd} {
		c.Flags().StringVar(&flagBTOpFormat, "format", "", "Output format")
	}
	bigtableOpListCmd.Flags().StringVar(&flagBTOpFilter, "filter", "", "Server-side filter expression")
	bigtableOpListCmd.Flags().Int64Var(&flagBTOpPageSize, "page-size", 0, "Maximum results per page")

	bigtableOpCmd.AddCommand(bigtableOpDescribeCmd, bigtableOpListCmd)
	bigtableCmd.AddCommand(bigtableOpCmd)
}

func runBTOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagBTOpFormat)
}

func runBTOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("operations/projects/%s", project)
	var all []*bigtableadmin.Operation
	pageToken := ""
	for {
		call := svc.Operations.Projects.Operations.List(name).Context(ctx)
		if flagBTOpFilter != "" {
			call = call.Filter(flagBTOpFilter)
		}
		if flagBTOpPageSize > 0 {
			call = call.PageSize(flagBTOpPageSize)
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
	if flagBTOpFormat != "" {
		return emitFormatted(all, flagBTOpFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
