package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage managed-folders (#1242) ---

var storageManagedFoldersCmd = &cobra.Command{Use: "managed-folders", Short: "Manage managed folders inside a bucket"}

var (
	flagStMfBucket   string
	flagStMfFormat   string
	flagStMfPageSize int64
	flagStMfPrefix   string
)

var (
	storageMfCreateCmd = &cobra.Command{
		Use: "create MANAGED_FOLDER", Short: "Create a managed folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: runStMfCreate,
	}
	storageMfDeleteCmd = &cobra.Command{
		Use: "delete MANAGED_FOLDER", Short: "Delete a managed folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: runStMfDelete,
	}
	storageMfDescribeCmd = &cobra.Command{
		Use: "describe MANAGED_FOLDER", Short: "Describe a managed folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: runStMfDescribe,
	}
	storageMfListCmd = &cobra.Command{
		Use: "list", Short: "List managed folders in a bucket",
		Args: cobra.NoArgs, RunE: runStMfList,
	}
	storageMfGetIamCmd = &cobra.Command{
		Use: "get-iam-policy MANAGED_FOLDER", Short: "Get the IAM policy for a managed folder",
		Args: cobra.ExactArgs(1), RunE: runStMfGetIam,
	}
	storageMfSetIamCmd = &cobra.Command{
		Use: "set-iam-policy MANAGED_FOLDER POLICY_FILE", Short: "Set the IAM policy for a managed folder",
		Args: cobra.ExactArgs(2), RunE: runStMfSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		storageMfCreateCmd, storageMfDeleteCmd, storageMfDescribeCmd,
		storageMfListCmd, storageMfGetIamCmd, storageMfSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagStMfBucket, "bucket", "", "Bucket that owns the managed folder (required)")
		_ = c.MarkFlagRequired("bucket")
		c.Flags().StringVar(&flagStMfFormat, "format", "", "Output format")
	}
	storageMfListCmd.Flags().StringVar(&flagStMfPrefix, "prefix", "", "Only list managed folders with this prefix")
	storageMfListCmd.Flags().Int64Var(&flagStMfPageSize, "page-size", 0, "Maximum results per page")

	storageManagedFoldersCmd.AddCommand(all...)
	storageCmd.AddCommand(storageManagedFoldersCmd)
}

func runStMfCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ManagedFolders.Insert(flagStMfBucket, &storage.ManagedFolder{Name: args[0]}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating managed folder: %w", err)
	}
	fmt.Printf("Created managed folder [%s] in bucket [%s].\n", args[0], flagStMfBucket)
	return emitFormatted(got, flagStMfFormat)
}

func runStMfDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if err := svc.ManagedFolders.Delete(flagStMfBucket, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting managed folder: %w", err)
	}
	fmt.Printf("Deleted managed folder [%s] in bucket [%s].\n", args[0], flagStMfBucket)
	return nil
}

func runStMfDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.ManagedFolders.Get(flagStMfBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing managed folder: %w", err)
	}
	return emitFormatted(got, flagStMfFormat)
}

func runStMfList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*storage.ManagedFolder
	pageToken := ""
	for {
		call := svc.ManagedFolders.List(flagStMfBucket).Context(ctx)
		if flagStMfPrefix != "" {
			call = call.Prefix(flagStMfPrefix)
		}
		if flagStMfPageSize > 0 {
			call = call.PageSize(flagStMfPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing managed folders: %w", err)
		}
		all = append(all, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagStMfFormat)
}

func runStMfGetIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.ManagedFolders.GetIamPolicy(flagStMfBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagStMfFormat)
}

func runStMfSetIam(cmd *cobra.Command, args []string) error {
	policy := &storage.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.ManagedFolders.SetIamPolicy(flagStMfBucket, args[0], policy).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for managed folder [%s].\n", args[0])
	return emitFormatted(updated, flagStMfFormat)
}
