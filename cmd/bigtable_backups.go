package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable backups (#1303) ---

var bigtableBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Bigtable backups"}

var (
	flagBtBkInstance   string
	flagBtBkCluster    string
	flagBtBkFormat     string
	flagBtBkConfigFile string
	flagBtBkUpdateMask string
	flagBtBkPageSize   int64
	flagBtBkFilter     string
	flagBtBkOrderBy    string
	flagBtBkDestTable  string
)

var (
	bigtableBkCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a Bigtable backup",
		Args: cobra.ExactArgs(1), RunE: runBtBkCreate,
	}
	bigtableBkDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a Bigtable backup",
		Args: cobra.ExactArgs(1), RunE: runBtBkDelete,
	}
	bigtableBkDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a Bigtable backup",
		Args: cobra.ExactArgs(1), RunE: runBtBkDescribe,
	}
	bigtableBkListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable backups on a cluster",
		Args: cobra.NoArgs, RunE: runBtBkList,
	}
	bigtableBkUpdateCmd = &cobra.Command{
		Use: "update BACKUP", Short: "Update a Bigtable backup",
		Args: cobra.ExactArgs(1), RunE: runBtBkUpdate,
	}
	bigtableBkRestoreCmd = &cobra.Command{
		Use: "restore BACKUP", Short: "Restore a Bigtable backup to a new table",
		Args: cobra.ExactArgs(1), RunE: runBtBkRestore,
	}
)

func init() {
	all := []*cobra.Command{
		bigtableBkCreateCmd, bigtableBkDeleteCmd, bigtableBkDescribeCmd,
		bigtableBkListCmd, bigtableBkUpdateCmd, bigtableBkRestoreCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBtBkInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagBtBkCluster, "cluster", "", "Bigtable cluster ID (required)")
		_ = c.MarkFlagRequired("cluster")
		c.Flags().StringVar(&flagBtBkFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableBkCreateCmd, bigtableBkUpdateCmd} {
		c.Flags().StringVar(&flagBtBkConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Backup body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	bigtableBkUpdateCmd.Flags().StringVar(&flagBtBkUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	bigtableBkListCmd.Flags().Int64Var(&flagBtBkPageSize, "page-size", 0, "Maximum results per page")
	bigtableBkListCmd.Flags().StringVar(&flagBtBkFilter, "filter", "", "Backup filter expression")
	bigtableBkListCmd.Flags().StringVar(&flagBtBkOrderBy, "order-by", "", "Server-side ordering expression")
	bigtableBkRestoreCmd.Flags().StringVar(&flagBtBkDestTable, "destination-table", "",
		"ID of the new table to create from the backup (required)")
	_ = bigtableBkRestoreCmd.MarkFlagRequired("destination-table")

	bigtableBackupsCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableBackupsCmd)
}

func btBkName(id string) (string, error) {
	parent, err := btClusterName(flagBtBkInstance, flagBtBkCluster)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backups/%s", parent, id), nil
}

func runBtBkCreate(cmd *cobra.Command, args []string) error {
	parent, err := btClusterName(flagBtBkInstance, flagBtBkCluster)
	if err != nil {
		return err
	}
	body := &bigtableadmin.Backup{}
	if err := loadYAMLOrJSONInto(flagBtBkConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Clusters.Backups.Create(parent, body).BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Printf("Create request issued for backup [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtBkFormat)
}

func runBtBkDelete(cmd *cobra.Command, args []string) error {
	name, err := btBkName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Clusters.Backups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Printf("Deleted backup [%s].\n", args[0])
	return nil
}

func runBtBkDescribe(cmd *cobra.Command, args []string) error {
	name, err := btBkName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Clusters.Backups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagBtBkFormat)
}

func runBtBkList(cmd *cobra.Command, args []string) error {
	parent, err := btClusterName(flagBtBkInstance, flagBtBkCluster)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.Backup
	pageToken := ""
	for {
		call := svc.Projects.Instances.Clusters.Backups.List(parent).Context(ctx)
		if flagBtBkPageSize > 0 {
			call = call.PageSize(flagBtBkPageSize)
		}
		if flagBtBkFilter != "" {
			call = call.Filter(flagBtBkFilter)
		}
		if flagBtBkOrderBy != "" {
			call = call.OrderBy(flagBtBkOrderBy)
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
	return emitFormatted(all, flagBtBkFormat)
}

func runBtBkUpdate(cmd *cobra.Command, args []string) error {
	name, err := btBkName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.Backup{}
	if err := loadYAMLOrJSONInto(flagBtBkConfigFile, body); err != nil {
		return err
	}
	mask := flagBtBkUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
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
	fmt.Printf("Updated backup [%s].\n", args[0])
	return emitFormatted(got, flagBtBkFormat)
}

func runBtBkRestore(cmd *cobra.Command, args []string) error {
	backup, err := btBkName(args[0])
	if err != nil {
		return err
	}
	instance, err := btInstanceName(flagBtBkInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.Restore(instance, &bigtableadmin.RestoreTableRequest{
		Backup:  backup,
		TableId: flagBtBkDestTable,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring backup: %w", err)
	}
	fmt.Printf("Restore request issued for table [%s] (operation: %s).\n", flagBtBkDestTable, op.Name)
	return emitFormatted(op, flagBtBkFormat)
}
