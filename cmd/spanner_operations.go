package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner operations (#1211) ---

var spannerOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Spanner operations"}

var (
	flagSpOpInstance string
	flagSpOpDatabase string
	flagSpOpBackup   string
	flagSpOpFormat   string
	flagSpOpFilter   string
	flagSpOpPageSize int64
)

var (
	spannerOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Spanner operation",
		Args: cobra.ExactArgs(1), RunE: runSpOpCancel,
	}
	spannerOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Spanner operation",
		Args: cobra.ExactArgs(1), RunE: runSpOpDescribe,
	}
	spannerOpListCmd = &cobra.Command{
		Use: "list", Short: "List Spanner operations at the requested scope",
		Args: cobra.NoArgs, RunE: runSpOpList,
	}
)

func init() {
	all := []*cobra.Command{spannerOpCancelCmd, spannerOpDescribeCmd, spannerOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagSpOpInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagSpOpDatabase, "database", "",
			"Database whose operations are targeted (mutually exclusive with --backup)")
		c.Flags().StringVar(&flagSpOpBackup, "backup", "",
			"Backup whose operations are targeted (mutually exclusive with --database)")
		c.Flags().StringVar(&flagSpOpFormat, "format", "", "Output format")
	}
	spannerOpListCmd.Flags().StringVar(&flagSpOpFilter, "filter", "", "Server-side filter expression")
	spannerOpListCmd.Flags().Int64Var(&flagSpOpPageSize, "page-size", 0, "Maximum results per page")

	spannerOperationsCmd.AddCommand(all...)
	spannerCmd.AddCommand(spannerOperationsCmd)
}

// spannerOpParent returns the parent collection for an ops list. The scope is
// database-level if --database is set, backup-level if --backup, otherwise
// instance-level.
func spannerOpParent() (string, string, error) {
	if flagSpOpDatabase != "" && flagSpOpBackup != "" {
		return "", "", fmt.Errorf("--database and --backup are mutually exclusive")
	}
	if flagSpOpDatabase != "" {
		db, err := spannerDatabase(flagSpOpInstance, flagSpOpDatabase)
		return db, "database", err
	}
	if flagSpOpBackup != "" {
		bk, err := spannerBackup(flagSpOpInstance, flagSpOpBackup)
		return bk, "backup", err
	}
	inst, err := spannerInstance(flagSpOpInstance)
	return inst, "instance", err
}

// spannerOpFullName resolves an operation ID/URI to a fully qualified name,
// using --database/--backup/--instance to compose the parent when the input is
// a short ID.
func spannerOpFullName(id string) (string, string, error) {
	if strings.HasPrefix(id, "projects/") {
		switch {
		case strings.Contains(id, "/databases/"):
			return id, "database", nil
		case strings.Contains(id, "/backups/"):
			return id, "backup", nil
		default:
			return id, "instance", nil
		}
	}
	parent, scope, err := spannerOpParent()
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), scope, nil
}

func runSpOpCancel(cmd *cobra.Command, args []string) error {
	name, scope, err := spannerOpFullName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	switch scope {
	case "database":
		_, err = svc.Projects.Instances.Databases.Operations.Cancel(name).Context(ctx).Do()
	case "backup":
		_, err = svc.Projects.Instances.Backups.Operations.Cancel(name).Context(ctx).Do()
	default:
		_, err = svc.Projects.Instances.Operations.Cancel(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runSpOpDescribe(cmd *cobra.Command, args []string) error {
	name, scope, err := spannerOpFullName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var op *spanner.Operation
	switch scope {
	case "database":
		op, err = svc.Projects.Instances.Databases.Operations.Get(name).Context(ctx).Do()
	case "backup":
		op, err = svc.Projects.Instances.Backups.Operations.Get(name).Context(ctx).Do()
	default:
		op, err = svc.Projects.Instances.Operations.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagSpOpFormat)
}

func runSpOpList(cmd *cobra.Command, args []string) error {
	parent, scope, err := spannerOpParent()
	if err != nil {
		return err
	}
	// Operations list endpoints expect the collection "{parent}/operations".
	target := parent + "/operations"
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.Operation
	pageToken := ""
	for {
		var resp *spanner.ListOperationsResponse
		switch scope {
		case "database":
			call := svc.Projects.Instances.Databases.Operations.List(target).Context(ctx)
			if flagSpOpFilter != "" {
				call = call.Filter(flagSpOpFilter)
			}
			if flagSpOpPageSize > 0 {
				call = call.PageSize(flagSpOpPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		case "backup":
			call := svc.Projects.Instances.Backups.Operations.List(target).Context(ctx)
			if flagSpOpFilter != "" {
				call = call.Filter(flagSpOpFilter)
			}
			if flagSpOpPageSize > 0 {
				call = call.PageSize(flagSpOpPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		default:
			call := svc.Projects.Instances.Operations.List(target).Context(ctx)
			if flagSpOpFilter != "" {
				call = call.Filter(flagSpOpFilter)
			}
			if flagSpOpPageSize > 0 {
				call = call.PageSize(flagSpOpPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		}
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpOpFormat)
}
