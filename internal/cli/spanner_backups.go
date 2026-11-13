package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner backups (#1206) ---

func (s *spannerCmds) initSpannerBackups() {
	all := []*cobra.Command{
		s.spannerBackupsCreateCmd, s.spannerBackupsDeleteCmd, s.spannerBackupsDescribeCmd,
		s.spannerBackupsListCmd, s.spannerBackupsCopyCmd, s.spannerBackupsUpdateMetaCmd,
		s.spannerBackupsGetIamCmd, s.spannerBackupsSetIamCmd, s.spannerBackupsAddIamCmd, s.spannerBackupsRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSpBackupsInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagSpBackupsFormat, "format", "", "Output format")
	}
	s.spannerBackupsCreateCmd.Flags().StringVar(&s.flagSpBackupsDatabase, "database", "", "Source database (required)")
	_ = s.spannerBackupsCreateCmd.MarkFlagRequired("database")
	s.spannerBackupsCreateCmd.Flags().StringVar(&s.flagSpBackupsConfigFile, "config-file", "", "YAML/JSON file for the Backup body")

	s.spannerBackupsListCmd.Flags().StringVar(&s.flagSpBackupsFilter, "filter", "", "Server-side filter expression")
	s.spannerBackupsListCmd.Flags().Int64Var(&s.flagSpBackupsPageSize, "page-size", 0, "Maximum results per page")

	s.spannerBackupsCopyCmd.Flags().StringVar(&s.flagSpBackupsDestBackupID, "dest-backup-id", "", "ID of the destination backup (required)")
	_ = s.spannerBackupsCopyCmd.MarkFlagRequired("dest-backup-id")
	s.spannerBackupsCopyCmd.Flags().StringVar(&s.flagSpBackupsSourceBackup, "source-backup", "", "Fully qualified source backup name or ID (required)")
	_ = s.spannerBackupsCopyCmd.MarkFlagRequired("source-backup")
	s.spannerBackupsCopyCmd.Flags().StringVar(&s.flagSpBackupsExpireTime, "expire-time", "", "Expiration time for the copy (RFC3339, required)")
	_ = s.spannerBackupsCopyCmd.MarkFlagRequired("expire-time")

	s.spannerBackupsUpdateMetaCmd.Flags().StringVar(&s.flagSpBackupsConfigFile, "config-file", "", "YAML/JSON file with Backup metadata (expireTime, versionTime)")
	s.spannerBackupsUpdateMetaCmd.Flags().StringVar(&s.flagSpBackupsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	for _, c := range []*cobra.Command{s.spannerBackupsAddIamCmd, s.spannerBackupsRemoveIamCmd} {
		spIamMemberFlags(c, &s.flagSpBackupsIamMember, &s.flagSpBackupsIamRole,
			&s.flagSpBackupsIamCondExpr, &s.flagSpBackupsIamCondT, &s.flagSpBackupsIamCondD)
	}
	s.spannerBackupsRemoveIamCmd.Flags().BoolVar(&s.flagSpBackupsIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.spannerBackupsCmd.AddCommand(all...)
	s.spannerCmd.AddCommand(s.spannerBackupsCmd)
}

func (s *spannerCmds) runSpBackupCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerInstance(s.flagSpBackupsInstance)
	if err != nil {
		return err
	}
	backup := &spanner.Backup{}
	if s.flagSpBackupsConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSpBackupsConfigFile, backup); err != nil {
			return err
		}
	}
	if backup.Database == "" {
		dbName, err := s.spannerDatabase(s.flagSpBackupsInstance, s.flagSpBackupsDatabase)
		if err != nil {
			return err
		}
		backup.Database = dbName
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Backups.Create(parent, backup).BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create backup [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSpBackupsFormat)
}

func (s *spannerCmds) runSpBackupDelete(cmd *cobra.Command, args []string) error {
	name, err := s.spannerBackup(s.flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Backups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted backup [%s].\n", args[0])
	return nil
}

func (s *spannerCmds) runSpBackupDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.spannerBackup(s.flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Backups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSpBackupsFormat)
}

func (s *spannerCmds) runSpBackupList(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerInstance(s.flagSpBackupsInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.Backup
	pageToken := ""
	for {
		call := svc.Projects.Instances.Backups.List(parent).Context(ctx)
		if s.flagSpBackupsFilter != "" {
			call = call.Filter(s.flagSpBackupsFilter)
		}
		if s.flagSpBackupsPageSize > 0 {
			call = call.PageSize(s.flagSpBackupsPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSpBackupsFormat)
}

func (s *spannerCmds) runSpBackupCopy(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerInstance(s.flagSpBackupsInstance)
	if err != nil {
		return err
	}
	src, err := s.spannerBackup(s.flagSpBackupsInstance, s.flagSpBackupsSourceBackup)
	if err != nil {
		return err
	}
	body := &spanner.CopyBackupRequest{
		BackupId:     s.flagSpBackupsDestBackupID,
		SourceBackup: src,
		ExpireTime:   s.flagSpBackupsExpireTime,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Backups.Copy(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("copying backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Copy backup [%s] initiated (operation: %s).\n", s.flagSpBackupsDestBackupID, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSpBackupsFormat)
}

func (s *spannerCmds) runSpBackupUpdateMeta(cmd *cobra.Command, args []string) error {
	name, err := s.spannerBackup(s.flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	body := &spanner.Backup{}
	if s.flagSpBackupsConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSpBackupsConfigFile, body); err != nil {
			return err
		}
	}
	body.Name = name
	mask := s.flagSpBackupsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
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
	return emitFormattedTo(s.inv.stdout, got, s.flagSpBackupsFormat)
}

func (s *spannerCmds) runSpBackupGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerBackup(s.flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Backups.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagSpBackupsFormat)
}

func (s *spannerCmds) runSpBackupSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerBackup(s.flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	policy := &spanner.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Instances.Backups.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("backup [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpBackupsFormat)
}

func (s *spannerCmds) runSpBackupAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerBackup(s.flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Backups.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	spIamAddBinding(policy, s.flagSpBackupsIamRole, s.flagSpBackupsIamMember,
		spIamBuildCondition(s.flagSpBackupsIamCondExpr, s.flagSpBackupsIamCondT, s.flagSpBackupsIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Instances.Backups.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("backup [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpBackupsFormat)
}

func (s *spannerCmds) runSpBackupRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerBackup(s.flagSpBackupsInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Backups.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !spIamRemoveBinding(policy, s.flagSpBackupsIamRole, s.flagSpBackupsIamMember,
		spIamBuildCondition(s.flagSpBackupsIamCondExpr, s.flagSpBackupsIamCondT, s.flagSpBackupsIamCondD), s.flagSpBackupsIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagSpBackupsIamRole, s.flagSpBackupsIamMember)
	}
	updated, err := svc.Projects.Instances.Backups.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("backup [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpBackupsFormat)
}
