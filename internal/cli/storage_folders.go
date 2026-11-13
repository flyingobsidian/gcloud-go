package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage folders (#1237) ---
//
// Manages hierarchical-namespace folders inside a bucket.

func (s *storageCmds) initStorageFolders() {
	all := []*cobra.Command{
		s.storageFoldersCreateCmd, s.storageFoldersDeleteCmd, s.storageFoldersDescribeCmd,
		s.storageFoldersListCmd, s.storageFoldersRenameCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagStFolBucket, "bucket", "", "Bucket that owns the folder (required)")
		_ = c.MarkFlagRequired("bucket")
		c.Flags().StringVar(&s.flagStFolFormat, "format", "", "Output format")
	}
	s.storageFoldersListCmd.Flags().StringVar(&s.flagStFolPrefix, "prefix", "", "Only list folders with this prefix")
	s.storageFoldersListCmd.Flags().Int64Var(&s.flagStFolPageSize, "page-size", 0, "Maximum results per page")
	s.storageFoldersRenameCmd.Flags().StringVar(&s.flagStFolDest, "destination-folder", "", "New folder name within the same bucket (required)")
	_ = s.storageFoldersRenameCmd.MarkFlagRequired("destination-folder")

	s.storageFoldersCmd.AddCommand(all...)
	s.storageCmd.AddCommand(s.storageFoldersCmd)
}

func (s *storageCmds) runStFolCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Folders.Insert(s.flagStFolBucket, &storage.Folder{Name: args[0]}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating folder: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created folder [%s] in bucket [%s].\n", args[0], s.flagStFolBucket)
	return emitFormattedTo(s.inv.stdout, got, s.flagStFolFormat)
}

func (s *storageCmds) runStFolDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	if err := svc.Folders.Delete(s.flagStFolBucket, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting folder: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted folder [%s] in bucket [%s].\n", args[0], s.flagStFolBucket)
	return nil
}

func (s *storageCmds) runStFolDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Folders.Get(s.flagStFolBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing folder: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStFolFormat)
}

func (s *storageCmds) runStFolList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	var all []*storage.Folder
	pageToken := ""
	for {
		call := svc.Folders.List(s.flagStFolBucket).Context(ctx)
		if s.flagStFolPrefix != "" {
			call = call.Prefix(s.flagStFolPrefix)
		}
		if s.flagStFolPageSize > 0 {
			call = call.PageSize(s.flagStFolPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagStFolFormat)
}

func (s *storageCmds) runStFolRename(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Folders.Rename(s.flagStFolBucket, args[0], s.flagStFolDest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming folder: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Rename folder [%s] -> [%s] initiated (operation: %s).\n", args[0], s.flagStFolDest, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagStFolFormat)
}
