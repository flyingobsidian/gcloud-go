package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// --- gcloud oracle-database operations (#1277) ---

var oracleOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Oracle Database long-running operations",
}

var (
	flagOdbOpLocation string
	flagOdbOpFormat   string
	flagOdbOpFilter   string
	flagOdbOpPageSize int64
)

var (
	oracleOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an Oracle Database operation",
		Args: cobra.ExactArgs(1), RunE: runOracleOpCancel,
	}
	oracleOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an Oracle Database operation",
		Args: cobra.ExactArgs(1), RunE: runOracleOpDelete,
	}
	oracleOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an Oracle Database operation",
		Args: cobra.ExactArgs(1), RunE: runOracleOpDescribe,
	}
	oracleOpListCmd = &cobra.Command{
		Use: "list", Short: "List Oracle Database operations in a location",
		Args: cobra.NoArgs, RunE: runOracleOpList,
	}
)

func init() {
	all := []*cobra.Command{oracleOpCancelCmd, oracleOpDeleteCmd, oracleOpDescribeCmd, oracleOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagOdbOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagOdbOpFormat, "format", "", "Output format")
	}
	oracleOpListCmd.Flags().StringVar(&flagOdbOpFilter, "filter", "", "Server-side filter expression")
	oracleOpListCmd.Flags().Int64Var(&flagOdbOpPageSize, "page-size", 0, "Maximum results per page")

	oracleOperationsCmd.AddCommand(all...)
	oracleDatabaseCmd.AddCommand(oracleOperationsCmd)
}

func oracleOpName(id string) (string, error) {
	return odbResource(flagOdbOpLocation, "operations", id)
}

func runOracleOpCancel(cmd *cobra.Command, args []string) error {
	name, err := oracleOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &oracledatabase.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runOracleOpDelete(cmd *cobra.Command, args []string) error {
	name, err := oracleOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runOracleOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := oracleOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagOdbOpFormat)
}

func runOracleOpList(cmd *cobra.Command, args []string) error {
	parent, err := odbLocationParent(flagOdbOpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.OracleDatabaseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*oracledatabase.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagOdbOpFilter != "" {
			call = call.Filter(flagOdbOpFilter)
		}
		if flagOdbOpPageSize > 0 {
			call = call.PageSize(flagOdbOpPageSize)
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
	return emitFormatted(all, flagOdbOpFormat)
}
