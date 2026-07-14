package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

var firestoreOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Firestore admin operations",
}

var (
	fsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Firestore admin operation",
		Args: cobra.ExactArgs(1), RunE: runFSOpDescribe,
	}
	fsOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a running Firestore admin operation",
		Args: cobra.ExactArgs(1), RunE: runFSOpCancel,
	}
	fsOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a completed Firestore admin operation",
		Args: cobra.ExactArgs(1), RunE: runFSOpDelete,
	}
	fsOpListCmd = &cobra.Command{
		Use: "list", Short: "List Firestore admin operations for a database",
		Args: cobra.NoArgs, RunE: runFSOpList,
	}
)

var (
	flagFSOpDatabase string
	flagFSOpFormat   string
	flagFSOpFilter   string
	flagFSOpPageSize int64
	flagFSOpLimit    int64
)

func init() {
	for _, c := range []*cobra.Command{fsOpDescribeCmd, fsOpCancelCmd, fsOpDeleteCmd, fsOpListCmd} {
		firestoreAddDatabaseFlag(c, &flagFSOpDatabase, false)
	}
	fsOpDescribeCmd.Flags().StringVar(&flagFSOpFormat, "format", "", "Output format")
	fsOpListCmd.Flags().StringVar(&flagFSOpFormat, "format", "", "Output format")
	fsOpListCmd.Flags().StringVar(&flagFSOpFilter, "filter", "", "Server-side filter expression")
	fsOpListCmd.Flags().Int64Var(&flagFSOpPageSize, "page-size", 0, "Page size")
	fsOpListCmd.Flags().Int64Var(&flagFSOpLimit, "limit", 0, "Cap total results (0 = no cap)")

	firestoreOperationsCmd.AddCommand(fsOpCancelCmd, fsOpDeleteCmd, fsOpDescribeCmd, fsOpListCmd)
	firestoreCmd.AddCommand(firestoreOperationsCmd)
}

func firestoreOpName(id, project, db string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/operations/%s", firestoreDatabaseName(project, defaultDB(db)), id)
}

func defaultDB(db string) string {
	if db == "" {
		return "(default)"
	}
	return db
}

func runFSOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Operations.Get(firestoreOpName(args[0], project, flagFSOpDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagFSOpFormat)
}

func runFSOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.Operations.Cancel(firestoreOpName(args[0], project, flagFSOpDatabase), &firestore.GoogleLongrunningCancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runFSOpDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.Operations.Delete(firestoreOpName(args[0], project, flagFSOpDatabase)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runFSOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := firestoreDatabaseName(project, defaultDB(flagFSOpDatabase))
	var all []*firestore.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Projects.Databases.Operations.List(parent).Context(ctx)
		if flagFSOpFilter != "" {
			call = call.Filter(flagFSOpFilter)
		}
		if flagFSOpPageSize > 0 {
			call = call.PageSize(flagFSOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if flagFSOpLimit > 0 && int64(len(all)) >= flagFSOpLimit {
			all = all[:flagFSOpLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagFSOpFormat != "" {
		return emitFormatted(all, flagFSOpFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
