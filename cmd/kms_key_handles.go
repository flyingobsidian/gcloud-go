package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms key-handles (#1105) ---

var kmsKeyHandlesCmd = &cobra.Command{
	Use:   "key-handles",
	Short: "Manage Cloud KMS KeyHandle resources",
}

var (
	flagKmsKHLocation   string
	flagKmsKHFormat     string
	flagKmsKHFilter     string
	flagKmsKHPageSize   int64
	flagKmsKHConfigFile string
	flagKmsKHHandleID   string
)

var kmsKHCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a KeyHandle (returns a long-running Operation)",
	Args:  cobra.NoArgs,
	RunE:  runKmsKHCreate,
}

var kmsKHDescribeCmd = &cobra.Command{
	Use:   "describe HANDLE",
	Short: "Describe a KeyHandle",
	Args:  cobra.ExactArgs(1),
	RunE:  runKmsKHDescribe,
}

var kmsKHListCmd = &cobra.Command{
	Use:   "list",
	Short: "List KeyHandles in a location",
	Args:  cobra.NoArgs,
	RunE:  runKmsKHList,
}

func init() {
	for _, c := range []*cobra.Command{kmsKHCreateCmd, kmsKHDescribeCmd, kmsKHListCmd} {
		c.Flags().StringVar(&flagKmsKHLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagKmsKHFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}

	kmsKHCreateCmd.Flags().StringVar(&flagKmsKHConfigFile, "config-file", "", "YAML/JSON body for the KeyHandle (required)")
	kmsKHCreateCmd.Flags().StringVar(&flagKmsKHHandleID, "key-handle-id", "", "Optional KeyHandle id; defaults to a server-generated UUID")
	_ = kmsKHCreateCmd.MarkFlagRequired("config-file")

	kmsKHListCmd.Flags().StringVar(&flagKmsKHFilter, "filter", "", "Filter expression")
	kmsKHListCmd.Flags().Int64Var(&flagKmsKHPageSize, "page-size", 0, "Page size")

	kmsKeyHandlesCmd.AddCommand(kmsKHCreateCmd, kmsKHDescribeCmd, kmsKHListCmd)
	kmsCmd.AddCommand(kmsKeyHandlesCmd)
}

func kmsKHName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/keyHandles"
	return kmsFullName(parent, raw)
}

func runKmsKHCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.KeyHandle{}
	if err := loadYAMLOrJSONInto(flagKmsKHConfigFile, body); err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsKHLocation)
	call := svc.Projects.Locations.KeyHandles.Create(parent, body).Context(ctx)
	if flagKmsKHHandleID != "" {
		call = call.KeyHandleId(flagKmsKHHandleID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating key handle: %w", err)
	}
	return emitFormatted(op, flagKmsKHFormat)
}

func runKmsKHDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsKHName(project, flagKmsKHLocation, args[0])
	out, err := svc.Projects.Locations.KeyHandles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing key handle: %w", err)
	}
	return emitFormatted(out, flagKmsKHFormat)
}

func runKmsKHList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, flagKmsKHLocation)
	var all []*cloudkms.KeyHandle
	token := ""
	for {
		call := svc.Projects.Locations.KeyHandles.List(parent).Context(ctx)
		if flagKmsKHFilter != "" {
			call = call.Filter(flagKmsKHFilter)
		}
		if flagKmsKHPageSize > 0 {
			call = call.PageSize(flagKmsKHPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing key handles: %w", err)
		}
		all = append(all, resp.KeyHandles...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsKHFormat)
}
