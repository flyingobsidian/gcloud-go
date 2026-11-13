package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage objects (#1243) ---
//
// Metadata-only object operations. Insert (upload) and Get with media
// download are intentionally left out — the streaming/media path is a
// larger surface and belongs in its own change.

func (s *storageCmds) initStorageObjects() {
	all := []*cobra.Command{
		s.storageObjectsDescribeCmd, s.storageObjectsListCmd,
		s.storageObjectsUpdateCmd, s.storageObjectsComposeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagStObjBucket, "bucket", "", "Bucket that owns the object (required)")
		_ = c.MarkFlagRequired("bucket")
		c.Flags().StringVar(&s.flagStObjFormat, "format", "", "Output format")
	}
	s.storageObjectsListCmd.Flags().StringVar(&s.flagStObjPrefix, "prefix", "", "Only list objects with this prefix")
	s.storageObjectsListCmd.Flags().StringVar(&s.flagStObjDelimiter, "delimiter", "", "Delimiter for hierarchical listing")
	s.storageObjectsListCmd.Flags().Int64Var(&s.flagStObjPageSize, "page-size", 0, "Maximum results per page")
	s.storageObjectsUpdateCmd.Flags().StringVar(&s.flagStObjConfigFile, "config-file", "", "YAML/JSON file with the Object body (required)")
	_ = s.storageObjectsUpdateCmd.MarkFlagRequired("config-file")
	s.storageObjectsComposeCmd.Flags().StringVar(&s.flagStObjConfigFile, "config-file", "", "YAML/JSON file with the ComposeRequest body (required)")
	_ = s.storageObjectsComposeCmd.MarkFlagRequired("config-file")

	s.storageObjectsCmd.AddCommand(all...)
	s.storageCmd.AddCommand(s.storageObjectsCmd)
}

func (s *storageCmds) runStObjDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Objects.Get(s.flagStObjBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing object: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStObjFormat)
}

func (s *storageCmds) runStObjList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	var all []*storage.Object
	pageToken := ""
	for {
		call := svc.Objects.List(s.flagStObjBucket).Context(ctx)
		if s.flagStObjPrefix != "" {
			call = call.Prefix(s.flagStObjPrefix)
		}
		if s.flagStObjDelimiter != "" {
			call = call.Delimiter(s.flagStObjDelimiter)
		}
		if s.flagStObjPageSize > 0 {
			call = call.MaxResults(s.flagStObjPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing objects: %w", err)
		}
		all = append(all, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagStObjFormat)
}

func (s *storageCmds) runStObjUpdate(cmd *cobra.Command, args []string) error {
	body := &storage.Object{}
	if err := loadYAMLOrJSONInto(s.flagStObjConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Objects.Patch(s.flagStObjBucket, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating object: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated object [%s] in bucket [%s].\n", args[0], s.flagStObjBucket)
	return emitFormattedTo(s.inv.stdout, got, s.flagStObjFormat)
}

func (s *storageCmds) runStObjCompose(cmd *cobra.Command, args []string) error {
	body := &storage.ComposeRequest{}
	if err := loadYAMLOrJSONInto(s.flagStObjConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Objects.Compose(s.flagStObjBucket, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("composing object: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Composed object [%s] in bucket [%s].\n", args[0], s.flagStObjBucket)
	return emitFormattedTo(s.inv.stdout, got, s.flagStObjFormat)
}
