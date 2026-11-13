package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner backup-schedules (#1214) ---

func (s *spannerCmds) initSpannerBackupSchedules() {
	all := []*cobra.Command{
		s.spannerBsCreateCmd, s.spannerBsDeleteCmd, s.spannerBsDescribeCmd,
		s.spannerBsListCmd, s.spannerBsUpdateCmd, s.spannerBsGetIamCmd, s.spannerBsSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSpBsInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagSpBsDatabase, "database", "", "Spanner database (required)")
		_ = c.MarkFlagRequired("database")
		c.Flags().StringVar(&s.flagSpBsFormat, "format", "", "Output format")
	}
	s.spannerBsCreateCmd.Flags().StringVar(&s.flagSpBsConfigFile, "config-file", "", "YAML/JSON file with the BackupSchedule body (required)")
	_ = s.spannerBsCreateCmd.MarkFlagRequired("config-file")
	s.spannerBsListCmd.Flags().Int64Var(&s.flagSpBsPageSize, "page-size", 0, "Maximum results per page")
	s.spannerBsUpdateCmd.Flags().StringVar(&s.flagSpBsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.spannerBsUpdateCmd.MarkFlagRequired("config-file")
	s.spannerBsUpdateCmd.Flags().StringVar(&s.flagSpBsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.spannerBackupSchedulesCmd.AddCommand(all...)
	s.spannerCmd.AddCommand(s.spannerBackupSchedulesCmd)
}

func (s *spannerCmds) spBsName(id string) (string, error) {
	return s.spannerBackupSchedule(s.flagSpBsInstance, s.flagSpBsDatabase, id)
}

func (s *spannerCmds) runSpBsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerDatabase(s.flagSpBsInstance, s.flagSpBsDatabase)
	if err != nil {
		return err
	}
	body := &spanner.BackupSchedule{}
	if err := loadYAMLOrJSONInto(s.flagSpBsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Databases.BackupSchedules.Create(parent, body).BackupScheduleId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup schedule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created backup schedule [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagSpBsFormat)
}

func (s *spannerCmds) runSpBsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.spBsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Databases.BackupSchedules.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup schedule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted backup schedule [%s].\n", args[0])
	return nil
}

func (s *spannerCmds) runSpBsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.spBsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Databases.BackupSchedules.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup schedule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSpBsFormat)
}

func (s *spannerCmds) runSpBsList(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerDatabase(s.flagSpBsInstance, s.flagSpBsDatabase)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.BackupSchedule
	pageToken := ""
	for {
		call := svc.Projects.Instances.Databases.BackupSchedules.List(parent).Context(ctx)
		if s.flagSpBsPageSize > 0 {
			call = call.PageSize(s.flagSpBsPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSpBsFormat)
}

func (s *spannerCmds) runSpBsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.spBsName(args[0])
	if err != nil {
		return err
	}
	body := &spanner.BackupSchedule{}
	if err := loadYAMLOrJSONInto(s.flagSpBsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagSpBsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Updated backup schedule [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagSpBsFormat)
}

func (s *spannerCmds) runSpBsGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.spBsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Databases.BackupSchedules.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagSpBsFormat)
}

func (s *spannerCmds) runSpBsSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.spBsName(args[0])
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
	updated, err := svc.Projects.Instances.Databases.BackupSchedules.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for backup schedule [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpBsFormat)
}
