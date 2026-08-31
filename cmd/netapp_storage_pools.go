package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp storage-pools (#1204) ---

var netappSPCmd = &cobra.Command{Use: "storage-pools", Short: "Manage NetApp storage pools"}

var (
	flagNetAppSPLocation   string
	flagNetAppSPConfigFile string
	flagNetAppSPUpdateMask string
	flagNetAppSPFormat     string
	flagNetAppSPFilter     string
	flagNetAppSPPageSize   int64

	// update-backup-config
	flagNetAppSPBackupCfgVolumeUUID string
	flagNetAppSPBackupCfgVault      string
	flagNetAppSPBackupCfgPolicies   []string
	flagNetAppSPBackupCfgSchedule   string

	// restore-volume
	flagNetAppSPRestoreBackup      string
	flagNetAppSPRestoreVolumeUUID  string
	flagNetAppSPRestoreDestPath    string
	flagNetAppSPRestoreFileList    []string
)

var (
	netappSPCreateCmd = &cobra.Command{
		Use: "create STORAGE_POOL", Short: "Create a storage pool",
		Args: cobra.ExactArgs(1), RunE: runNetAppSPCreate,
	}
	netappSPDeleteCmd = &cobra.Command{
		Use: "delete STORAGE_POOL", Short: "Delete a storage pool",
		Args: cobra.ExactArgs(1), RunE: runNetAppSPDelete,
	}
	netappSPDescribeCmd = &cobra.Command{
		Use: "describe STORAGE_POOL", Short: "Describe a storage pool",
		Args: cobra.ExactArgs(1), RunE: runNetAppSPDescribe,
	}
	netappSPListCmd = &cobra.Command{
		Use: "list", Short: "List storage pools",
		Args: cobra.NoArgs, RunE: runNetAppSPList,
	}
	netappSPUpdateCmd = &cobra.Command{
		Use: "update STORAGE_POOL", Short: "Update a storage pool",
		Args: cobra.ExactArgs(1), RunE: runNetAppSPUpdate,
	}
	netappSPSwitchCmd = &cobra.Command{
		Use: "switch STORAGE_POOL", Short: "Switch the active zone of a Regional Flex storage pool",
		Args: cobra.ExactArgs(1), RunE: runNetAppSPSwitch,
	}
	netappSPUpdateBackupCfgCmd = &cobra.Command{
		Use: "update-backup-config STORAGE_POOL", Short: "Update the backup configuration for a volume in an ONTAP-mode storage pool",
		Args: cobra.ExactArgs(1), RunE: runNetAppSPUpdateBackupConfig,
	}
	netappSPListBackupCfgsCmd = &cobra.Command{
		Use: "list-backup-configs STORAGE_POOL", Short: "List backup configurations for all volumes in an ONTAP-mode storage pool",
		Args: cobra.ExactArgs(1), RunE: runNetAppSPListBackupConfigs,
	}
	netappSPRestoreVolumeCmd = &cobra.Command{
		Use: "restore-volume STORAGE_POOL", Short: "Restore a backup to a volume in an ONTAP-mode storage pool",
		Args: cobra.ExactArgs(1), RunE: runNetAppSPRestoreVolume,
	}
)

func init() {
	all := []*cobra.Command{
		netappSPCreateCmd, netappSPDeleteCmd, netappSPDescribeCmd,
		netappSPListCmd, netappSPUpdateCmd, netappSPSwitchCmd,
		netappSPUpdateBackupCfgCmd, netappSPListBackupCfgsCmd, netappSPRestoreVolumeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNetAppSPLocation, "location", "", "Location for the storage pool (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppSPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappSPCreateCmd, netappSPUpdateCmd} {
		c.Flags().StringVar(&flagNetAppSPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the StoragePool body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappSPUpdateCmd.Flags().StringVar(&flagNetAppSPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappSPListCmd.Flags().StringVar(&flagNetAppSPFilter, "filter", "", "Server-side filter expression")
	netappSPListCmd.Flags().Int64Var(&flagNetAppSPPageSize, "page-size", 0, "Maximum number of results per page")

	// update-backup-config flags
	netappSPUpdateBackupCfgCmd.Flags().StringVar(&flagNetAppSPBackupCfgVolumeUUID, "volume-uuid", "",
		"UUID of the ONTAP-mode volume to update backup config for (required)")
	_ = netappSPUpdateBackupCfgCmd.MarkFlagRequired("volume-uuid")
	netappSPUpdateBackupCfgCmd.Flags().StringVar(&flagNetAppSPBackupCfgVault, "backup-vault", "",
		"Backup vault resource name")
	netappSPUpdateBackupCfgCmd.Flags().StringSliceVar(&flagNetAppSPBackupCfgPolicies, "backup-policies", nil,
		"Comma-separated list of backup policy resource names")
	netappSPUpdateBackupCfgCmd.Flags().StringVar(&flagNetAppSPBackupCfgSchedule, "enable-scheduled-backups", "",
		"Whether scheduled backups are enabled on the volume (true|false)")
	netappSPUpdateBackupCfgCmd.Flags().StringVar(&flagNetAppSPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every specified field)")

	// restore-volume flags
	netappSPRestoreVolumeCmd.Flags().StringVar(&flagNetAppSPRestoreBackup, "backup", "",
		"Full resource name of the backup to restore from (required)")
	_ = netappSPRestoreVolumeCmd.MarkFlagRequired("backup")
	netappSPRestoreVolumeCmd.Flags().StringVar(&flagNetAppSPRestoreVolumeUUID, "volume-uuid", "",
		"UUID of the ONTAP-mode volume to restore to (required)")
	_ = netappSPRestoreVolumeCmd.MarkFlagRequired("volume-uuid")
	netappSPRestoreVolumeCmd.Flags().StringVar(&flagNetAppSPRestoreDestPath, "restore-destination-path", "",
		"Absolute directory path in the destination volume for selective file restore")
	netappSPRestoreVolumeCmd.Flags().StringSliceVar(&flagNetAppSPRestoreFileList, "file-list", nil,
		"Comma-separated list of absolute file paths to restore (for selective file restore)")

	netappSPCmd.AddCommand(all...)
	netappCmd.AddCommand(netappSPCmd)
}

func netappSPParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, flagNetAppSPLocation), nil
}

func netappSPName(id string) (string, error) {
	parent, err := netappSPParent()
	if err != nil {
		return "", err
	}
	return netappChild("storagePools", id, parent), nil
}

func runNetAppSPCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappSPParent()
	if err != nil {
		return err
	}
	body := &netapp.StoragePool{}
	if err := loadYAMLOrJSONInto(flagNetAppSPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.Create(parent, body).StoragePoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating storage pool: %w", err)
	}
	fmt.Printf("Create request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppSPFormat)
}

func runNetAppSPDelete(cmd *cobra.Command, args []string) error {
	name, err := netappSPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting storage pool: %w", err)
	}
	fmt.Printf("Delete request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppSPFormat)
}

func runNetAppSPDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappSPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.StoragePools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing storage pool: %w", err)
	}
	return emitFormatted(got, flagNetAppSPFormat)
}

func runNetAppSPList(cmd *cobra.Command, args []string) error {
	parent, err := netappSPParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.StoragePool
	pageToken := ""
	for {
		call := svc.Projects.Locations.StoragePools.List(parent).Context(ctx)
		if flagNetAppSPFilter != "" {
			call = call.Filter(flagNetAppSPFilter)
		}
		if flagNetAppSPPageSize > 0 {
			call = call.PageSize(flagNetAppSPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing storage pools: %w", err)
		}
		all = append(all, resp.StoragePools...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNetAppSPFormat)
}

func runNetAppSPUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappSPName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.StoragePool{}
	if err := loadYAMLOrJSONInto(flagNetAppSPConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppSPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.StoragePools.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating storage pool: %w", err)
	}
	fmt.Printf("Update request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppSPFormat)
}

func parseBoolFlag(v string) (bool, error) {
	b, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(v)))
	if err != nil {
		return false, fmt.Errorf("invalid boolean value %q", v)
	}
	return b, nil
}

func runNetAppSPUpdateBackupConfig(cmd *cobra.Command, args []string) error {
	name, err := netappSPName(args[0])
	if err != nil {
		return err
	}
	cfg := &netapp.BackupConfig{}
	var fields []string
	if flagNetAppSPBackupCfgVault != "" {
		cfg.BackupVault = flagNetAppSPBackupCfgVault
		fields = append(fields, "backup_config.backup_vault")
	}
	if len(flagNetAppSPBackupCfgPolicies) > 0 {
		cfg.BackupPolicies = flagNetAppSPBackupCfgPolicies
		fields = append(fields, "backup_config.backup_policies")
	}
	if flagNetAppSPBackupCfgSchedule != "" {
		enable, perr := parseBoolFlag(flagNetAppSPBackupCfgSchedule)
		if perr != nil {
			return fmt.Errorf("--enable-scheduled-backups: %w", perr)
		}
		cfg.ScheduledBackupEnabled = enable
		cfg.ForceSendFields = append(cfg.ForceSendFields, "ScheduledBackupEnabled")
		fields = append(fields, "backup_config.scheduled_backup_enabled")
	}
	mask := flagNetAppSPUpdateMask
	if mask == "" {
		mask = joinMask(fields)
	}
	req := &netapp.UpdateBackupConfigRequest{
		BackupConfig: cfg,
		VolumeUuid:   flagNetAppSPBackupCfgVolumeUUID,
		UpdateMask:   mask,
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.UpdateBackupConfig(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating backup config: %w", err)
	}
	fmt.Printf("Update-backup-config request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppSPFormat)
}

func runNetAppSPListBackupConfigs(cmd *cobra.Command, args []string) error {
	parent, err := netappSPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.VolumeBackupConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.StoragePools.BackupConfigs.List(parent).Context(ctx)
		if flagNetAppSPFilter != "" {
			call = call.Filter(flagNetAppSPFilter)
		}
		if flagNetAppSPPageSize > 0 {
			call = call.PageSize(flagNetAppSPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backup configs: %w", err)
		}
		all = append(all, resp.VolumeBackupConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNetAppSPFormat)
}

func runNetAppSPRestoreVolume(cmd *cobra.Command, args []string) error {
	name, err := netappSPName(args[0])
	if err != nil {
		return err
	}
	req := &netapp.RestoreVolumeRequest{
		BackupSource: &netapp.BackupSource{
			Backup:   flagNetAppSPRestoreBackup,
			FileList: flagNetAppSPRestoreFileList,
		},
		OntapVolumeTarget: &netapp.OntapVolumeTarget{
			VolumeUuid:             flagNetAppSPRestoreVolumeUUID,
			RestoreDestinationPath: flagNetAppSPRestoreDestPath,
		},
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.RestoreVolume(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring volume in storage pool: %w", err)
	}
	fmt.Printf("Restore-volume request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppSPFormat)
}

func runNetAppSPSwitch(cmd *cobra.Command, args []string) error {
	name, err := netappSPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.Switch(name, &netapp.SwitchActiveReplicaZoneRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("switching storage pool zone: %w", err)
	}
	fmt.Printf("Switch request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppSPFormat)
}
