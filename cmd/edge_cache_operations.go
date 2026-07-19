package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud edge-cache operations (#1058) ---

var edgeCacheOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Media CDN Edge Cache long-running operations"}

var (
	flagEdgeCacheOperationsLocation string
	flagEdgeCacheOperationsFormat   string
	flagEdgeCacheOperationsPageSize int64
)

var (
	edgeCacheOperationsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an Edge Cache long-running operation",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheOperationsDescribe,
	}
	edgeCacheOperationsListCmd = &cobra.Command{
		Use: "list", Short: "List Edge Cache long-running operations",
		Args: cobra.NoArgs, RunE: runEdgeCacheOperationsList,
	}
)

func init() {
	all := []*cobra.Command{edgeCacheOperationsDescribeCmd, edgeCacheOperationsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagEdgeCacheOperationsLocation, "location", "", "Edge Cache location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagEdgeCacheOperationsFormat, "format", "", "Output format")
	}
	edgeCacheOperationsListCmd.Flags().Int64Var(&flagEdgeCacheOperationsPageSize, "page-size", 0, "Maximum results per page")

	edgeCacheOperationsCmd.AddCommand(all...)
	edgeCacheCmd.AddCommand(edgeCacheOperationsCmd)
}

func edgeCacheOperationsParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagEdgeCacheOperationsLocation), nil
}

func edgeCacheOperationsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := edgeCacheOperationsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func runEdgeCacheOperationsDescribe(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheOperationsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing edge-cache operation: %w", err)
	}
	return emitFormatted(got, flagEdgeCacheOperationsFormat)
}

func runEdgeCacheOperationsList(cmd *cobra.Command, args []string) error {
	parent, err := edgeCacheOperationsParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := edgeCacheRest.paginate(ctx, "/"+parent+"/operations", nil, "operations", flagEdgeCacheOperationsPageSize)
	if err != nil {
		return fmt.Errorf("listing edge-cache operations: %w", err)
	}
	return emitFormatted(items, flagEdgeCacheOperationsFormat)
}
