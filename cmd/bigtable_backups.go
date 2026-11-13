package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable backups (#1303) ---

func (s *bigtableCmds) initBigtableBackups() {
	all := []*cobra.Command{
		s.bigtableBkCreateCmd, s.bigtableBkDeleteCmd, s.bigtableBkDescribeCmd,
		s.bigtableBkListCmd, s.bigtableBkUpdateCmd, s.bigtableBkRestoreCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBtBkInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagBtBkCluster, "cluster", "", "Bigtable cluster ID (required)")
		_ = c.MarkFlagRequired("cluster")
		c.Flags().StringVar(&s.flagBtBkFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bigtableBkCreateCmd, s.bigtableBkUpdateCmd} {
		c.Flags().StringVar(&s.flagBtBkConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Backup body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bigtableBkUpdateCmd.Flags().StringVar(&s.flagBtBkUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.bigtableBkListCmd.Flags().Int64Var(&s.flagBtBkPageSize, "page-size", 0, "Maximum results per page")
	s.bigtableBkListCmd.Flags().StringVar(&s.flagBtBkFilter, "filter", "", "Backup filter expression")
	s.bigtableBkListCmd.Flags().StringVar(&s.flagBtBkOrderBy, "order-by", "", "Server-side ordering expression")
	s.bigtableBkRestoreCmd.Flags().StringVar(&s.flagBtBkDestTable, "destination-table", "",
		"ID of the new table to create from the backup (required)")
	_ = s.bigtableBkRestoreCmd.MarkFlagRequired("destination-table")

	s.bigtableBackupsCmd.AddCommand(all...)
	s.bigtableCmd.AddCommand(s.bigtableBackupsCmd)
}

func (s *bigtableCmds) btBkName(id string) (string, error) {
	parent, err := s.btClusterName(s.flagBtBkInstance, s.flagBtBkCluster)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backups/%s", parent, id), nil
}

func (s *bigtableCmds) runBtBkCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.btClusterName(s.flagBtBkInstance, s.flagBtBkCluster)
	if err != nil {
		return err
	}
	body := &bigtableadmin.Backup{}
	if err := loadYAMLOrJSONInto(s.flagBtBkConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Clusters.Backups.Create(parent, body).BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for backup [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtBkFormat)
}

func (s *bigtableCmds) runBtBkDelete(cmd *cobra.Command, args []string) error {
	name, err := s.btBkName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Clusters.Backups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted backup [%s].\n", args[0])
	return nil
}

func (s *bigtableCmds) runBtBkDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.btBkName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Clusters.Backups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBtBkFormat)
}

func (s *bigtableCmds) runBtBkList(cmd *cobra.Command, args []string) error {
	parent, err := s.btClusterName(s.flagBtBkInstance, s.flagBtBkCluster)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.Backup
	pageToken := ""
	for {
		call := svc.Projects.Instances.Clusters.Backups.List(parent).Context(ctx)
		if s.flagBtBkPageSize > 0 {
			call = call.PageSize(s.flagBtBkPageSize)
		}
		if s.flagBtBkFilter != "" {
			call = call.Filter(s.flagBtBkFilter)
		}
		if s.flagBtBkOrderBy != "" {
			call = call.OrderBy(s.flagBtBkOrderBy)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagBtBkFormat)
}

func (s *bigtableCmds) runBtBkUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.btBkName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.Backup{}
	if err := loadYAMLOrJSONInto(s.flagBtBkConfigFile, body); err != nil {
		return err
	}
	mask := s.flagBtBkUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Clusters.Backups.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated backup [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBtBkFormat)
}

func (s *bigtableCmds) runBtBkRestore(cmd *cobra.Command, args []string) error {
	backup, err := s.btBkName(args[0])
	if err != nil {
		return err
	}
	instance, err := s.btInstanceName(s.flagBtBkInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.Restore(instance, &bigtableadmin.RestoreTableRequest{
		Backup:  backup,
		TableId: s.flagBtBkDestTable,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restore request issued for table [%s] (operation: %s).\n", s.flagBtBkDestTable, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBtBkFormat)
}
