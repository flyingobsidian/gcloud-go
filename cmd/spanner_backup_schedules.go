package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner backup-schedules (#1214) ---

var spannerBackupSchedulesCmd = &cobra.Command{Use: "backup-schedules", Short: "Manage Spanner backup schedules"}

var (
	flagSpBsInstance   string
	flagSpBsDatabase   string
	flagSpBsFormat     string
	flagSpBsConfigFile string
	flagSpBsUpdateMask string
	flagSpBsPageSize   int64
)

var (
	spannerBsCreateCmd = &cobra.Command{
		Use: "create SCHEDULE", Short: "Create a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: runSpBsCreate,
	}
	spannerBsDeleteCmd = &cobra.Command{
		Use: "delete SCHEDULE", Short: "Delete a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: runSpBsDelete,
	}
	spannerBsDescribeCmd = &cobra.Command{
		Use: "describe SCHEDULE", Short: "Describe a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: runSpBsDescribe,
	}
	spannerBsListCmd = &cobra.Command{
		Use: "list", Short: "List Spanner backup schedules on a database",
		Args: cobra.NoArgs, RunE: runSpBsList,
	}
	spannerBsUpdateCmd = &cobra.Command{
		Use: "update SCHEDULE", Short: "Update a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: runSpBsUpdate,
	}
	spannerBsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SCHEDULE", Short: "Get the IAM policy for a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: runSpBsGetIam,
	}
	spannerBsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SCHEDULE POLICY_FILE", Short: "Set the IAM policy for a Spanner backup schedule",
		Args: cobra.ExactArgs(2), RunE: runSpBsSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		spannerBsCreateCmd, spannerBsDeleteCmd, spannerBsDescribeCmd,
		spannerBsListCmd, spannerBsUpdateCmd, spannerBsGetIamCmd, spannerBsSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSpBsInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagSpBsDatabase, "database", "", "Spanner database (required)")
		_ = c.MarkFlagRequired("database")
		c.Flags().StringVar(&flagSpBsFormat, "format", "", "Output format")
	}
	spannerBsCreateCmd.Flags().StringVar(&flagSpBsConfigFile, "config-file", "", "YAML/JSON file with the BackupSchedule body (required)")
	_ = spannerBsCreateCmd.MarkFlagRequired("config-file")
	spannerBsListCmd.Flags().Int64Var(&flagSpBsPageSize, "page-size", 0, "Maximum results per page")
	spannerBsUpdateCmd.Flags().StringVar(&flagSpBsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = spannerBsUpdateCmd.MarkFlagRequired("config-file")
	spannerBsUpdateCmd.Flags().StringVar(&flagSpBsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	spannerBackupSchedulesCmd.AddCommand(all...)
	spannerCmd.AddCommand(spannerBackupSchedulesCmd)
}

func spBsName(id string) (string, error) {
	return spannerBackupSchedule(flagSpBsInstance, flagSpBsDatabase, id)
}

func runSpBsCreate(cmd *cobra.Command, args []string) error {
	parent, err := spannerDatabase(flagSpBsInstance, flagSpBsDatabase)
	if err != nil {
		return err
	}
	body := &spanner.BackupSchedule{}
	if err := loadYAMLOrJSONInto(flagSpBsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Databases.BackupSchedules.Create(parent, body).BackupScheduleId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup schedule: %w", err)
	}
	fmt.Printf("Created backup schedule [%s].\n", args[0])
	return emitFormatted(got, flagSpBsFormat)
}

func runSpBsDelete(cmd *cobra.Command, args []string) error {
	name, err := spBsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Databases.BackupSchedules.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup schedule: %w", err)
	}
	fmt.Printf("Deleted backup schedule [%s].\n", args[0])
	return nil
}

func runSpBsDescribe(cmd *cobra.Command, args []string) error {
	name, err := spBsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Databases.BackupSchedules.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup schedule: %w", err)
	}
	return emitFormatted(got, flagSpBsFormat)
}

func runSpBsList(cmd *cobra.Command, args []string) error {
	parent, err := spannerDatabase(flagSpBsInstance, flagSpBsDatabase)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.BackupSchedule
	pageToken := ""
	for {
		call := svc.Projects.Instances.Databases.BackupSchedules.List(parent).Context(ctx)
		if flagSpBsPageSize > 0 {
			call = call.PageSize(flagSpBsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backup schedules: %w", err)
		}
		all = append(all, resp.BackupSchedules...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpBsFormat)
}

func runSpBsUpdate(cmd *cobra.Command, args []string) error {
	name, err := spBsName(args[0])
	if err != nil {
		return err
	}
	body := &spanner.BackupSchedule{}
	if err := loadYAMLOrJSONInto(flagSpBsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagSpBsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Databases.BackupSchedules.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup schedule: %w", err)
	}
	fmt.Printf("Updated backup schedule [%s].\n", args[0])
	return emitFormatted(got, flagSpBsFormat)
}

func runSpBsGetIam(cmd *cobra.Command, args []string) error {
	name, err := spBsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Databases.BackupSchedules.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagSpBsFormat)
}

func runSpBsSetIam(cmd *cobra.Command, args []string) error {
	name, err := spBsName(args[0])
	if err != nil {
		return err
	}
	policy := &spanner.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Instances.Databases.BackupSchedules.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for backup schedule [%s].\n", args[0])
	return emitFormatted(updated, flagSpBsFormat)
}
