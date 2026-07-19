package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner backups (#1206) ---

var spannerBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Cloud Spanner backups"}

var (
	flagSpBackupsInstance     string
	flagSpBackupsFormat       string
	flagSpBackupsDatabase     string
	flagSpBackupsConfigFile   string
	flagSpBackupsUpdateMask   string
	flagSpBackupsFilter       string
	flagSpBackupsPageSize     int64
	flagSpBackupsDestBackupID string
	flagSpBackupsSourceBackup string
	flagSpBackupsExpireTime   string
	flagSpBackupsIamMember    string
	flagSpBackupsIamRole      string
	flagSpBackupsIamCondExpr  string
	flagSpBackupsIamCondT     string
	flagSpBackupsIamCondD     string
	flagSpBackupsIamAllCond   bool
)

var (
	spannerBackupsCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: runSpBackupCreate,
	}
	spannerBackupsDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: runSpBackupDelete,
	}
	spannerBackupsDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: runSpBackupDescribe,
	}
	spannerBackupsListCmd = &cobra.Command{
		Use: "list", Short: "List Spanner backups",
		Args: cobra.NoArgs, RunE: runSpBackupList,
	}
	spannerBackupsCopyCmd = &cobra.Command{
		Use: "copy", Short: "Copy a Spanner backup",
		Args: cobra.NoArgs, RunE: runSpBackupCopy,
	}
	spannerBackupsUpdateMetaCmd = &cobra.Command{
		Use: "update-metadata BACKUP", Short: "Update metadata (expire-time, version-time) of a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: runSpBackupUpdateMeta,
	}
	spannerBackupsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy BACKUP", Short: "Get the IAM policy for a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: runSpBackupGetIam,
	}
	spannerBackupsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy BACKUP POLICY_FILE", Short: "Set the IAM policy for a Spanner backup",
		Args: cobra.ExactArgs(2), RunE: runSpBackupSetIam,
	}
	spannerBackupsAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding BACKUP", Short: "Add an IAM policy binding to a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: runSpBackupAddIam,
	}
	spannerBackupsRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding BACKUP", Short: "Remove an IAM policy binding from a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: runSpBackupRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		spannerBackupsCreateCmd, spannerBackupsDeleteCmd, spannerBackupsDescribeCmd,
		spannerBackupsListCmd, spannerBackupsCopyCmd, spannerBackupsUpdateMetaCmd,
		spannerBackupsGetIamCmd, spannerBackupsSetIamCmd, spannerBackupsAddIamCmd, spannerBackupsRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSpBackupsInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagSpBackupsFormat, "format", "", "Output format")
	}
	spannerBackupsCreateCmd.Flags().StringVar(&flagSpBackupsDatabase, "database", "", "Source database (required)")
	_ = spannerBackupsCreateCmd.MarkFlagRequired("database")
	spannerBackupsCreateCmd.Flags().StringVar(&flagSpBackupsConfigFile, "config-file", "", "YAML/JSON file for the Backup body")

	spannerBackupsListCmd.Flags().StringVar(&flagSpBackupsFilter, "filter", "", "Server-side filter expression")
	spannerBackupsListCmd.Flags().Int64Var(&flagSpBackupsPageSize, "page-size", 0, "Maximum results per page")

	spannerBackupsCopyCmd.Flags().StringVar(&flagSpBackupsDestBackupID, "dest-backup-id", "", "ID of the destination backup (required)")
	_ = spannerBackupsCopyCmd.MarkFlagRequired("dest-backup-id")
	spannerBackupsCopyCmd.Flags().StringVar(&flagSpBackupsSourceBackup, "source-backup", "", "Fully qualified source backup name or ID (required)")
	_ = spannerBackupsCopyCmd.MarkFlagRequired("source-backup")
	spannerBackupsCopyCmd.Flags().StringVar(&flagSpBackupsExpireTime, "expire-time", "", "Expiration time for the copy (RFC3339, required)")
	_ = spannerBackupsCopyCmd.MarkFlagRequired("expire-time")

	spannerBackupsUpdateMetaCmd.Flags().StringVar(&flagSpBackupsConfigFile, "config-file", "", "YAML/JSON file with Backup metadata (expireTime, versionTime)")
	spannerBackupsUpdateMetaCmd.Flags().StringVar(&flagSpBackupsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	for _, c := range []*cobra.Command{spannerBackupsAddIamCmd, spannerBackupsRemoveIamCmd} {
		spIamMemberFlags(c, &flagSpBackupsIamMember, &flagSpBackupsIamRole,
			&flagSpBackupsIamCondExpr, &flagSpBackupsIamCondT, &flagSpBackupsIamCondD)
	}
	spannerBackupsRemoveIamCmd.Flags().BoolVar(&flagSpBackupsIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	spannerBackupsCmd.AddCommand(all...)
	spannerCmd.AddCommand(spannerBackupsCmd)
}

func runSpBackupCreate(cmd *cobra.Command, args []string) error {
	parent, err := spannerInstance(flagSpBackupsInstance)
	if err != nil {
		return err
	}
	backup := &spanner.Backup{}
	if flagSpBackupsConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSpBackupsConfigFile, backup); err != nil {
			return err
		}
	}
	if backup.Database == "" {
		dbName, err := spannerDatabase(flagSpBackupsInstance, flagSpBackupsDatabase)
		if err != nil {
			return err
		}
		backup.Database = dbName
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Backups.Create(parent, backup).BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Printf("Create backup [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSpBackupsFormat)
}

func runSpBackupDelete(cmd *cobra.Command, args []string) error {
	name, err := spannerBackup(flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Backups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Printf("Deleted backup [%s].\n", args[0])
	return nil
}

func runSpBackupDescribe(cmd *cobra.Command, args []string) error {
	name, err := spannerBackup(flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Backups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagSpBackupsFormat)
}

func runSpBackupList(cmd *cobra.Command, args []string) error {
	parent, err := spannerInstance(flagSpBackupsInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.Backup
	pageToken := ""
	for {
		call := svc.Projects.Instances.Backups.List(parent).Context(ctx)
		if flagSpBackupsFilter != "" {
			call = call.Filter(flagSpBackupsFilter)
		}
		if flagSpBackupsPageSize > 0 {
			call = call.PageSize(flagSpBackupsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backups: %w", err)
		}
		all = append(all, resp.Backups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpBackupsFormat)
}

func runSpBackupCopy(cmd *cobra.Command, args []string) error {
	parent, err := spannerInstance(flagSpBackupsInstance)
	if err != nil {
		return err
	}
	src, err := spannerBackup(flagSpBackupsInstance, flagSpBackupsSourceBackup)
	if err != nil {
		return err
	}
	body := &spanner.CopyBackupRequest{
		BackupId:     flagSpBackupsDestBackupID,
		SourceBackup: src,
		ExpireTime:   flagSpBackupsExpireTime,
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Backups.Copy(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("copying backup: %w", err)
	}
	fmt.Printf("Copy backup [%s] initiated (operation: %s).\n", flagSpBackupsDestBackupID, op.Name)
	return emitFormatted(op, flagSpBackupsFormat)
}

func runSpBackupUpdateMeta(cmd *cobra.Command, args []string) error {
	name, err := spannerBackup(flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	body := &spanner.Backup{}
	if flagSpBackupsConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSpBackupsConfigFile, body); err != nil {
			return err
		}
	}
	body.Name = name
	mask := flagSpBackupsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Backups.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup metadata: %w", err)
	}
	return emitFormatted(got, flagSpBackupsFormat)
}

func runSpBackupGetIam(cmd *cobra.Command, args []string) error {
	name, err := spannerBackup(flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Backups.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagSpBackupsFormat)
}

func runSpBackupSetIam(cmd *cobra.Command, args []string) error {
	name, err := spannerBackup(flagSpBackupsInstance, args[0])
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
	updated, err := svc.Projects.Instances.Backups.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("backup [%s]", args[0]))
	return emitFormatted(updated, flagSpBackupsFormat)
}

func runSpBackupAddIam(cmd *cobra.Command, args []string) error {
	name, err := spannerBackup(flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Backups.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	spIamAddBinding(policy, flagSpBackupsIamRole, flagSpBackupsIamMember,
		spIamBuildCondition(flagSpBackupsIamCondExpr, flagSpBackupsIamCondT, flagSpBackupsIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Instances.Backups.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("backup [%s]", args[0]))
	return emitFormatted(updated, flagSpBackupsFormat)
}

func runSpBackupRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := spannerBackup(flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Backups.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !spIamRemoveBinding(policy, flagSpBackupsIamRole, flagSpBackupsIamMember,
		spIamBuildCondition(flagSpBackupsIamCondExpr, flagSpBackupsIamCondT, flagSpBackupsIamCondD), flagSpBackupsIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagSpBackupsIamRole, flagSpBackupsIamMember)
	}
	updated, err := svc.Projects.Instances.Backups.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("backup [%s]", args[0]))
	return emitFormatted(updated, flagSpBackupsFormat)
}
