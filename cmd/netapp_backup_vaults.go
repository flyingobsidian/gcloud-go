package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp backup-vaults (#1199) ---

var netappBVCmd = &cobra.Command{Use: "backup-vaults", Short: "Manage NetApp backup vaults"}

var (
	flagNetAppBVLocation   string
	flagNetAppBVConfigFile string
	flagNetAppBVUpdateMask string
	flagNetAppBVFormat     string
	flagNetAppBVFilter     string
	flagNetAppBVPageSize   int64
)

var (
	netappBVCreateCmd = &cobra.Command{
		Use: "create BACKUP_VAULT", Short: "Create a backup vault",
		Args: cobra.ExactArgs(1), RunE: runNetAppBVCreate,
	}
	netappBVDeleteCmd = &cobra.Command{
		Use: "delete BACKUP_VAULT", Short: "Delete a backup vault",
		Args: cobra.ExactArgs(1), RunE: runNetAppBVDelete,
	}
	netappBVDescribeCmd = &cobra.Command{
		Use: "describe BACKUP_VAULT", Short: "Describe a backup vault",
		Args: cobra.ExactArgs(1), RunE: runNetAppBVDescribe,
	}
	netappBVListCmd = &cobra.Command{
		Use: "list", Short: "List backup vaults",
		Args: cobra.NoArgs, RunE: runNetAppBVList,
	}
	netappBVUpdateCmd = &cobra.Command{
		Use: "update BACKUP_VAULT", Short: "Update a backup vault",
		Args: cobra.ExactArgs(1), RunE: runNetAppBVUpdate,
	}
)

// --- backups subgroup ---

var netappBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage NetApp backups within a backup vault"}

var (
	flagNetAppBackupVault      string
	flagNetAppBackupConfigFile string
	flagNetAppBackupUpdateMask string
	flagNetAppBackupFilter     string
	flagNetAppBackupPageSize   int64
)

var (
	netappBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a backup",
		Args: cobra.ExactArgs(1), RunE: runNetAppBackupCreate,
	}
	netappBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup",
		Args: cobra.ExactArgs(1), RunE: runNetAppBackupDelete,
	}
	netappBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup",
		Args: cobra.ExactArgs(1), RunE: runNetAppBackupDescribe,
	}
	netappBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups within a backup vault",
		Args: cobra.NoArgs, RunE: runNetAppBackupList,
	}
	netappBackupUpdateCmd = &cobra.Command{
		Use: "update BACKUP", Short: "Update a backup",
		Args: cobra.ExactArgs(1), RunE: runNetAppBackupUpdate,
	}
)

func init() {
	// backup-vaults command flags
	bvAll := []*cobra.Command{netappBVCreateCmd, netappBVDeleteCmd, netappBVDescribeCmd, netappBVListCmd, netappBVUpdateCmd}
	for _, c := range bvAll {
		c.Flags().StringVar(&flagNetAppBVLocation, "location", "", "Location for the backup vault (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppBVFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappBVCreateCmd, netappBVUpdateCmd} {
		c.Flags().StringVar(&flagNetAppBVConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the BackupVault body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappBVUpdateCmd.Flags().StringVar(&flagNetAppBVUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappBVListCmd.Flags().StringVar(&flagNetAppBVFilter, "filter", "", "Server-side filter expression")
	netappBVListCmd.Flags().Int64Var(&flagNetAppBVPageSize, "page-size", 0, "Maximum number of results per page")

	netappBVCmd.AddCommand(bvAll...)

	// backups subgroup
	bkAll := []*cobra.Command{netappBackupCreateCmd, netappBackupDeleteCmd, netappBackupDescribeCmd, netappBackupListCmd, netappBackupUpdateCmd}
	for _, c := range bkAll {
		c.Flags().StringVar(&flagNetAppBVLocation, "location", "", "Location for the backup (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppBackupVault, "backup-vault", "", "Backup vault that owns the backup (required)")
		_ = c.MarkFlagRequired("backup-vault")
		c.Flags().StringVar(&flagNetAppBVFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappBackupCreateCmd, netappBackupUpdateCmd} {
		c.Flags().StringVar(&flagNetAppBackupConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Backup body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappBackupUpdateCmd.Flags().StringVar(&flagNetAppBackupUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappBackupListCmd.Flags().StringVar(&flagNetAppBackupFilter, "filter", "", "Server-side filter expression")
	netappBackupListCmd.Flags().Int64Var(&flagNetAppBackupPageSize, "page-size", 0, "Maximum number of results per page")

	netappBackupsCmd.AddCommand(bkAll...)
	netappBVCmd.AddCommand(netappBackupsCmd)
	netappCmd.AddCommand(netappBVCmd)
}

func netappBVParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, flagNetAppBVLocation), nil
}

func netappBVName(id string) (string, error) {
	parent, err := netappBVParent()
	if err != nil {
		return "", err
	}
	return netappChild("backupVaults", id, parent), nil
}

func netappBackupParent() (string, error) {
	bv, err := netappBVName(flagNetAppBackupVault)
	if err != nil {
		return "", err
	}
	return bv, nil
}

func netappBackupName(id string) (string, error) {
	parent, err := netappBackupParent()
	if err != nil {
		return "", err
	}
	return netappChild("backups", id, parent), nil
}

func runNetAppBVCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappBVParent()
	if err != nil {
		return err
	}
	body := &netapp.BackupVault{}
	if err := loadYAMLOrJSONInto(flagNetAppBVConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupVaults.Create(parent, body).BackupVaultId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup vault: %w", err)
	}
	fmt.Printf("Create request issued for backup vault [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBVFormat)
}

func runNetAppBVDelete(cmd *cobra.Command, args []string) error {
	name, err := netappBVName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupVaults.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup vault: %w", err)
	}
	fmt.Printf("Delete request issued for backup vault [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBVFormat)
}

func runNetAppBVDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappBVName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.BackupVaults.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup vault: %w", err)
	}
	return emitFormatted(got, flagNetAppBVFormat)
}

func runNetAppBVList(cmd *cobra.Command, args []string) error {
	parent, err := netappBVParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.BackupVault
	pageToken := ""
	for {
		call := svc.Projects.Locations.BackupVaults.List(parent).Context(ctx)
		if flagNetAppBVFilter != "" {
			call = call.Filter(flagNetAppBVFilter)
		}
		if flagNetAppBVPageSize > 0 {
			call = call.PageSize(flagNetAppBVPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backup vaults: %w", err)
		}
		all = append(all, resp.BackupVaults...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNetAppBVFormat)
}

func runNetAppBVUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappBVName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.BackupVault{}
	if err := loadYAMLOrJSONInto(flagNetAppBVConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppBVUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.BackupVaults.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup vault: %w", err)
	}
	fmt.Printf("Update request issued for backup vault [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBVFormat)
}

func runNetAppBackupCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappBackupParent()
	if err != nil {
		return err
	}
	body := &netapp.Backup{}
	if err := loadYAMLOrJSONInto(flagNetAppBackupConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupVaults.Backups.Create(parent, body).BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Printf("Create request issued for backup [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBVFormat)
}

func runNetAppBackupDelete(cmd *cobra.Command, args []string) error {
	name, err := netappBackupName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupVaults.Backups.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Printf("Delete request issued for backup [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBVFormat)
}

func runNetAppBackupDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappBackupName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.BackupVaults.Backups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagNetAppBVFormat)
}

func runNetAppBackupList(cmd *cobra.Command, args []string) error {
	parent, err := netappBackupParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.Backup
	pageToken := ""
	for {
		call := svc.Projects.Locations.BackupVaults.Backups.List(parent).Context(ctx)
		if flagNetAppBackupFilter != "" {
			call = call.Filter(flagNetAppBackupFilter)
		}
		if flagNetAppBackupPageSize > 0 {
			call = call.PageSize(flagNetAppBackupPageSize)
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
	return emitFormatted(all, flagNetAppBVFormat)
}

func runNetAppBackupUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappBackupName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.Backup{}
	if err := loadYAMLOrJSONInto(flagNetAppBackupConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppBackupUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.BackupVaults.Backups.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	fmt.Printf("Update request issued for backup [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBVFormat)
}
