package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage hmac (#1238) ---

var storageHmacCmd = &cobra.Command{Use: "hmac", Short: "Manage HMAC keys for service accounts"}

var (
	flagStHmacFormat         string
	flagStHmacServiceAccount string
	flagStHmacState          string
)

var (
	storageHmacCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an HMAC key for a service account (loads request from --service-account)",
		Args: cobra.NoArgs, RunE: runStHmacCreate,
	}
	storageHmacDeleteCmd = &cobra.Command{
		Use: "delete ACCESS_ID", Short: "Delete an HMAC key",
		Args: cobra.ExactArgs(1), RunE: runStHmacDelete,
	}
	storageHmacDescribeCmd = &cobra.Command{
		Use: "describe ACCESS_ID", Short: "Describe an HMAC key",
		Args: cobra.ExactArgs(1), RunE: runStHmacDescribe,
	}
	storageHmacListCmd = &cobra.Command{
		Use: "list", Short: "List HMAC keys in the current project",
		Args: cobra.NoArgs, RunE: runStHmacList,
	}
	storageHmacUpdateCmd = &cobra.Command{
		Use: "update ACCESS_ID", Short: "Update an HMAC key state (ACTIVE or INACTIVE)",
		Args: cobra.ExactArgs(1), RunE: runStHmacUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		storageHmacCreateCmd, storageHmacDeleteCmd, storageHmacDescribeCmd,
		storageHmacListCmd, storageHmacUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagStHmacFormat, "format", "", "Output format")
	}
	storageHmacCreateCmd.Flags().StringVar(&flagStHmacServiceAccount, "service-account", "", "Service account email to create the key for (required)")
	_ = storageHmacCreateCmd.MarkFlagRequired("service-account")
	storageHmacListCmd.Flags().StringVar(&flagStHmacServiceAccount, "service-account", "", "Restrict list to a service account")
	storageHmacUpdateCmd.Flags().StringVar(&flagStHmacState, "state", "", "New state: ACTIVE or INACTIVE (required)")
	_ = storageHmacUpdateCmd.MarkFlagRequired("state")

	storageHmacCmd.AddCommand(all...)
	storageCmd.AddCommand(storageHmacCmd)
}

func runStHmacCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.HmacKeys.Create(project, flagStHmacServiceAccount).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating HMAC key: %w", err)
	}
	return emitFormatted(got, flagStHmacFormat)
}

func runStHmacDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if err := svc.Projects.HmacKeys.Delete(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting HMAC key: %w", err)
	}
	fmt.Printf("Deleted HMAC key [%s].\n", args[0])
	return nil
}

func runStHmacDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.HmacKeys.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing HMAC key: %w", err)
	}
	return emitFormatted(got, flagStHmacFormat)
}

func runStHmacList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*storage.HmacKeyMetadata
	pageToken := ""
	for {
		call := svc.Projects.HmacKeys.List(project).Context(ctx)
		if flagStHmacServiceAccount != "" {
			call = call.ServiceAccountEmail(flagStHmacServiceAccount)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing HMAC keys: %w", err)
		}
		all = append(all, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagStHmacFormat)
}

func runStHmacUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.HmacKeys.Update(project, args[0], &storage.HmacKeyMetadata{State: flagStHmacState}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating HMAC key: %w", err)
	}
	fmt.Printf("Updated HMAC key [%s].\n", args[0])
	return emitFormatted(got, flagStHmacFormat)
}
