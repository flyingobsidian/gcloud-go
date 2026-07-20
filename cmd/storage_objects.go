package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage objects (#1243) ---
//
// Metadata-only object operations. Insert (upload) and Get with media
// download are intentionally left out — the streaming/media path is a
// larger surface and belongs in its own change.

var storageObjectsCmd = &cobra.Command{Use: "objects", Short: "Manage Cloud Storage objects (metadata operations)"}

var (
	flagStObjBucket     string
	flagStObjFormat     string
	flagStObjConfigFile string
	flagStObjPrefix     string
	flagStObjDelimiter  string
	flagStObjPageSize   int64
)

var (
	storageObjectsDescribeCmd = &cobra.Command{
		Use: "describe OBJECT", Short: "Describe an object (metadata only)",
		Args: cobra.ExactArgs(1), RunE: runStObjDescribe,
	}
	storageObjectsListCmd = &cobra.Command{
		Use: "list", Short: "List objects in a bucket",
		Args: cobra.NoArgs, RunE: runStObjList,
	}
	storageObjectsUpdateCmd = &cobra.Command{
		Use: "update OBJECT", Short: "Update object metadata (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runStObjUpdate,
	}
	storageObjectsComposeCmd = &cobra.Command{
		Use: "compose DESTINATION_OBJECT", Short: "Compose an object from source objects (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runStObjCompose,
	}
)

func init() {
	all := []*cobra.Command{
		storageObjectsDescribeCmd, storageObjectsListCmd,
		storageObjectsUpdateCmd, storageObjectsComposeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagStObjBucket, "bucket", "", "Bucket that owns the object (required)")
		_ = c.MarkFlagRequired("bucket")
		c.Flags().StringVar(&flagStObjFormat, "format", "", "Output format")
	}
	storageObjectsListCmd.Flags().StringVar(&flagStObjPrefix, "prefix", "", "Only list objects with this prefix")
	storageObjectsListCmd.Flags().StringVar(&flagStObjDelimiter, "delimiter", "", "Delimiter for hierarchical listing")
	storageObjectsListCmd.Flags().Int64Var(&flagStObjPageSize, "page-size", 0, "Maximum results per page")
	storageObjectsUpdateCmd.Flags().StringVar(&flagStObjConfigFile, "config-file", "", "YAML/JSON file with the Object body (required)")
	_ = storageObjectsUpdateCmd.MarkFlagRequired("config-file")
	storageObjectsComposeCmd.Flags().StringVar(&flagStObjConfigFile, "config-file", "", "YAML/JSON file with the ComposeRequest body (required)")
	_ = storageObjectsComposeCmd.MarkFlagRequired("config-file")

	storageObjectsCmd.AddCommand(all...)
	storageCmd.AddCommand(storageObjectsCmd)
}

func runStObjDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Objects.Get(flagStObjBucket, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing object: %w", err)
	}
	return emitFormatted(got, flagStObjFormat)
}

func runStObjList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*storage.Object
	pageToken := ""
	for {
		call := svc.Objects.List(flagStObjBucket).Context(ctx)
		if flagStObjPrefix != "" {
			call = call.Prefix(flagStObjPrefix)
		}
		if flagStObjDelimiter != "" {
			call = call.Delimiter(flagStObjDelimiter)
		}
		if flagStObjPageSize > 0 {
			call = call.MaxResults(flagStObjPageSize)
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
	return emitFormatted(all, flagStObjFormat)
}

func runStObjUpdate(cmd *cobra.Command, args []string) error {
	body := &storage.Object{}
	if err := loadYAMLOrJSONInto(flagStObjConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Objects.Patch(flagStObjBucket, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating object: %w", err)
	}
	fmt.Printf("Updated object [%s] in bucket [%s].\n", args[0], flagStObjBucket)
	return emitFormatted(got, flagStObjFormat)
}

func runStObjCompose(cmd *cobra.Command, args []string) error {
	body := &storage.ComposeRequest{}
	if err := loadYAMLOrJSONInto(flagStObjConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Objects.Compose(flagStObjBucket, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("composing object: %w", err)
	}
	fmt.Printf("Composed object [%s] in bucket [%s].\n", args[0], flagStObjBucket)
	return emitFormatted(got, flagStObjFormat)
}
