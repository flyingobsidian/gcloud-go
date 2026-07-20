package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage folders (#1237) ---
//
// Manages hierarchical-namespace folders inside a bucket.

var storageFoldersCmd = &cobra.Command{Use: "folders", Short: "Manage hierarchical-namespace folders"}

var (
	flagStFolBucket string
	flagStFolFormat string
	flagStFolPrefix string
	flagStFolPageSize int64
	flagStFolDest string
)

var (
	storageFoldersCreateCmd = &cobra.Command{
		Use: "create FOLDER", Short: "Create a folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: runStFolCreate,
	}
	storageFoldersDeleteCmd = &cobra.Command{
		Use: "delete FOLDER", Short: "Delete a folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: runStFolDelete,
	}
	storageFoldersDescribeCmd = &cobra.Command{
		Use: "describe FOLDER", Short: "Describe a folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: runStFolDescribe,
	}
	storageFoldersListCmd = &cobra.Command{
		Use: "list", Short: "List folders in a bucket",
		Args: cobra.NoArgs, RunE: runStFolList,
	}
	storageFoldersRenameCmd = &cobra.Command{
		Use: "rename SOURCE_FOLDER", Short: "Rename a folder within a bucket",
		Args: cobra.ExactArgs(1), RunE: runStFolRename,
	}
)

func init() {
	all := []*cobra.Command{
		storageFoldersCreateCmd, storageFoldersDeleteCmd, storageFoldersDescribeCmd,
		storageFoldersListCmd, storageFoldersRenameCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagStFolBucket, "bucket", "", "Bucket that owns the folder (required)")
		_ = c.MarkFlagRequired("bucket")
		c.Flags().StringVar(&flagStFolFormat, "format", "", "Output format")
	}
	storageFoldersListCmd.Flags().StringVar(&flagStFolPrefix, "prefix", "", "Only list folders with this prefix")
	storageFoldersListCmd.Flags().Int64Var(&flagStFolPageSize, "page-size", 0, "Maximum results per page")
	storageFoldersRenameCmd.Flags().StringVar(&flagStFolDest, "destination-folder", "", "New folder name within the same bucket (required)")
	_ = storageFoldersRenameCmd.MarkFlagRequired("destination-folder")

	storageFoldersCmd.AddCommand(all...)
	storageCmd.AddCommand(storageFoldersCmd)
}

func runStFolCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Folders.Insert(flagStFolBucket, &storage.Folder{Name: args[0]}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating folder: %w", err)
	}
	fmt.Printf("Created folder [%s] in bucket [%s].\n", args[0], flagStFolBucket)
	return emitFormatted(got, flagStFolFormat)
}

func runStFolDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if err := svc.Folders.Delete(flagStFolBucket, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting folder: %w", err)
	}
	fmt.Printf("Deleted folder [%s] in bucket [%s].\n", args[0], flagStFolBucket)
	return nil
}

func runStFolDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Folders.Get(flagStFolBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing folder: %w", err)
	}
	return emitFormatted(got, flagStFolFormat)
}

func runStFolList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*storage.Folder
	pageToken := ""
	for {
		call := svc.Folders.List(flagStFolBucket).Context(ctx)
		if flagStFolPrefix != "" {
			call = call.Prefix(flagStFolPrefix)
		}
		if flagStFolPageSize > 0 {
			call = call.PageSize(flagStFolPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing folders: %w", err)
		}
		all = append(all, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagStFolFormat)
}

func runStFolRename(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Folders.Rename(flagStFolBucket, args[0], flagStFolDest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming folder: %w", err)
	}
	fmt.Printf("Rename folder [%s] -> [%s] initiated (operation: %s).\n", args[0], flagStFolDest, op.Name)
	return emitFormatted(op, flagStFolFormat)
}
