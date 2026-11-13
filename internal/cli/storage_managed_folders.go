package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage managed-folders (#1242) ---

func (s *storageCmds) initStorageManagedFolders() {
	all := []*cobra.Command{
		s.storageMfCreateCmd, s.storageMfDeleteCmd, s.storageMfDescribeCmd,
		s.storageMfListCmd, s.storageMfGetIamCmd, s.storageMfSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagStMfBucket, "bucket", "", "Bucket that owns the managed folder (required)")
		_ = c.MarkFlagRequired("bucket")
		c.Flags().StringVar(&s.flagStMfFormat, "format", "", "Output format")
	}
	s.storageMfListCmd.Flags().StringVar(&s.flagStMfPrefix, "prefix", "", "Only list managed folders with this prefix")
	s.storageMfListCmd.Flags().Int64Var(&s.flagStMfPageSize, "page-size", 0, "Maximum results per page")

	s.storageManagedFoldersCmd.AddCommand(all...)
	s.storageCmd.AddCommand(s.storageManagedFoldersCmd)
}

func (s *storageCmds) runStMfCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ManagedFolders.Insert(s.flagStMfBucket, &storage.ManagedFolder{Name: args[0]}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating managed folder: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created managed folder [%s] in bucket [%s].\n", args[0], s.flagStMfBucket)
	return emitFormattedTo(s.inv.stdout, got, s.flagStMfFormat)
}

func (s *storageCmds) runStMfDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	if err := svc.ManagedFolders.Delete(s.flagStMfBucket, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting managed folder: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted managed folder [%s] in bucket [%s].\n", args[0], s.flagStMfBucket)
	return nil
}

func (s *storageCmds) runStMfDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.ManagedFolders.Get(s.flagStMfBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing managed folder: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStMfFormat)
}

func (s *storageCmds) runStMfList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	var all []*storage.ManagedFolder
	pageToken := ""
	for {
		call := svc.ManagedFolders.List(s.flagStMfBucket).Context(ctx)
		if s.flagStMfPrefix != "" {
			call = call.Prefix(s.flagStMfPrefix)
		}
		if s.flagStMfPageSize > 0 {
			call = call.PageSize(s.flagStMfPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagStMfFormat)
}

func (s *storageCmds) runStMfGetIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.ManagedFolders.GetIamPolicy(s.flagStMfBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagStMfFormat)
}

func (s *storageCmds) runStMfSetIam(cmd *cobra.Command, args []string) error {
	policy := &storage.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.ManagedFolders.SetIamPolicy(s.flagStMfBucket, args[0], policy).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for managed folder [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagStMfFormat)
}
