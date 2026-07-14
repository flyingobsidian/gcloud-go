package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

var dmOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Database Migration Service operations",
}

var dmOpDescribeCmd = &cobra.Command{
	Use:   "describe NAME",
	Short: "Describe a Database Migration operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMOpDescribe,
}

var dmOpDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete a Database Migration operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMOpDelete,
}

var dmOpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Database Migration operations in a region",
	Args:  cobra.NoArgs,
	RunE:  runDMOpList,
}

var (
	flagDMOpRegion       string
	flagDMOpFormat       string
	flagDMOpListFilter   string
	flagDMOpListPageSize int64
	flagDMOpListLimit    int64
	flagDMOpListURI      bool
)

func init() {
	for _, c := range []*cobra.Command{dmOpDescribeCmd, dmOpDeleteCmd, dmOpListCmd} {
		c.Flags().StringVar(&flagDMOpRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	dmOpDescribeCmd.Flags().StringVar(&flagDMOpFormat, "format", "", "Output format (yaml, json, table, ...)")

	dmOpListCmd.Flags().StringVar(&flagDMOpFormat, "format", "", "Output format (yaml, json, table, ...)")
	dmOpListCmd.Flags().StringVar(&flagDMOpListFilter, "filter", "", "Server-side filter expression")
	dmOpListCmd.Flags().Int64Var(&flagDMOpListPageSize, "page-size", 0, "Page size for API pagination")
	dmOpListCmd.Flags().Int64Var(&flagDMOpListLimit, "limit", 0, "Cap total results (0 = no cap)")
	dmOpListCmd.Flags().BoolVar(&flagDMOpListURI, "uri", false, "Print resource names only")

	dmOperationsCmd.AddCommand(dmOpDescribeCmd, dmOpDeleteCmd, dmOpListCmd)
	databaseMigrationCmd.AddCommand(dmOperationsCmd)
}

func dmOpResourceName(name, project, region string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, region, name)
}

func runDMOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(dmOpResourceName(args[0], project, flagDMOpRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagDMOpFormat)
}

func runDMOpDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(dmOpResourceName(args[0], project, flagDMOpRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runDMOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmParent(project, flagDMOpRegion)
	var all []*datamigration.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagDMOpListFilter != "" {
			call = call.Filter(flagDMOpListFilter)
		}
		if flagDMOpListPageSize > 0 {
			call = call.PageSize(flagDMOpListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if flagDMOpListLimit > 0 && int64(len(all)) >= flagDMOpListLimit {
			all = all[:flagDMOpListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDMOpListURI {
		for _, o := range all {
			fmt.Println(o.Name)
		}
		return nil
	}
	if flagDMOpFormat != "" {
		return emitFormatted(all, flagDMOpFormat)
	}
	fmt.Printf("%-40s %-10s %s\n", "NAME", "DONE", "TARGET")
	for _, o := range all {
		fmt.Printf("%-40s %-10v %s\n", path.Base(o.Name), o.Done, operationTarget(o))
	}
	return nil
}

// operationTarget returns the "target" resource name embedded in the operation
// metadata, if any. Metadata schemas vary; a best-effort JSON lookup keeps
// output human-readable without imposing a fixed schema.
func operationTarget(op *datamigration.Operation) string {
	if len(op.Metadata) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(op.Metadata, &m); err != nil {
		return ""
	}
	if t, ok := m["target"].(string); ok {
		return t
	}
	return ""
}
