package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp storage-pools (#1204) ---

func (s *netappCmds) initNetappStoragePools() {
	all := []*cobra.Command{
		s.netappSPCreateCmd, s.netappSPDeleteCmd, s.netappSPDescribeCmd,
		s.netappSPListCmd, s.netappSPUpdateCmd, s.netappSPSwitchCmd,
		s.netappSPUpdateBackupCfgCmd, s.netappSPListBackupCfgsCmd, s.netappSPRestoreVolumeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNetAppSPLocation, "location", "", "Location for the storage pool (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppSPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappSPCreateCmd, s.netappSPUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppSPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the StoragePool body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappSPUpdateCmd.Flags().StringVar(&s.flagNetAppSPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappSPListCmd.Flags().StringVar(&s.flagNetAppSPFilter, "filter", "", "Server-side filter expression")
	s.netappSPListCmd.Flags().Int64Var(&s.flagNetAppSPPageSize, "page-size", 0, "Maximum number of results per page")

	// update-backup-config flags
	s.netappSPUpdateBackupCfgCmd.Flags().StringVar(&s.flagNetAppSPBackupCfgVolumeUUID, "volume-uuid", "",
		"UUID of the ONTAP-mode volume to update backup config for (required)")
	_ = s.netappSPUpdateBackupCfgCmd.MarkFlagRequired("volume-uuid")
	s.netappSPUpdateBackupCfgCmd.Flags().StringVar(&s.flagNetAppSPBackupCfgVault, "backup-vault", "",
		"Backup vault resource name")
	s.netappSPUpdateBackupCfgCmd.Flags().StringSliceVar(&s.flagNetAppSPBackupCfgPolicies, "backup-policies", nil,
		"Comma-separated list of backup policy resource names")
	s.netappSPUpdateBackupCfgCmd.Flags().StringVar(&s.flagNetAppSPBackupCfgSchedule, "enable-scheduled-backups", "",
		"Whether scheduled backups are enabled on the volume (true|false)")
	s.netappSPUpdateBackupCfgCmd.Flags().StringVar(&s.flagNetAppSPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every specified field)")

	// restore-volume flags
	s.netappSPRestoreVolumeCmd.Flags().StringVar(&s.flagNetAppSPRestoreBackup, "backup", "",
		"Full resource name of the backup to restore from (required)")
	_ = s.netappSPRestoreVolumeCmd.MarkFlagRequired("backup")
	s.netappSPRestoreVolumeCmd.Flags().StringVar(&s.flagNetAppSPRestoreVolumeUUID, "volume-uuid", "",
		"UUID of the ONTAP-mode volume to restore to (required)")
	_ = s.netappSPRestoreVolumeCmd.MarkFlagRequired("volume-uuid")
	s.netappSPRestoreVolumeCmd.Flags().StringVar(&s.flagNetAppSPRestoreDestPath, "restore-destination-path", "",
		"Absolute directory path in the destination volume for selective file restore")
	s.netappSPRestoreVolumeCmd.Flags().StringSliceVar(&s.flagNetAppSPRestoreFileList, "file-list", nil,
		"Comma-separated list of absolute file paths to restore (for selective file restore)")

	s.netappSPCmd.AddCommand(all...)
	s.netappCmd.AddCommand(s.netappSPCmd)
}

func (s *netappCmds) netappSPParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, s.flagNetAppSPLocation), nil
}

func (s *netappCmds) netappSPName(id string) (string, error) {
	parent, err := s.netappSPParent()
	if err != nil {
		return "", err
	}
	return netappChild("storagePools", id, parent), nil
}

func (s *netappCmds) runNetAppSPCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappSPParent()
	if err != nil {
		return err
	}
	body := &netapp.StoragePool{}
	if err := loadYAMLOrJSONInto(s.flagNetAppSPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.Create(parent, body).StoragePoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating storage pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppSPFormat)
}

func (s *netappCmds) runNetAppSPDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappSPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting storage pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppSPFormat)
}

func (s *netappCmds) runNetAppSPDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappSPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.StoragePools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing storage pool: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppSPFormat)
}

func (s *netappCmds) runNetAppSPList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappSPParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.StoragePool
	pageToken := ""
	for {
		call := svc.Projects.Locations.StoragePools.List(parent).Context(ctx)
		if s.flagNetAppSPFilter != "" {
			call = call.Filter(s.flagNetAppSPFilter)
		}
		if s.flagNetAppSPPageSize > 0 {
			call = call.PageSize(s.flagNetAppSPPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppSPFormat)
}

func (s *netappCmds) runNetAppSPUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappSPName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.StoragePool{}
	if err := loadYAMLOrJSONInto(s.flagNetAppSPConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppSPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Update request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppSPFormat)
}

func parseBoolFlag(v string) (bool, error) {
	b, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(v)))
	if err != nil {
		return false, fmt.Errorf("invalid boolean value %q", v)
	}
	return b, nil
}

func (s *netappCmds) runNetAppSPUpdateBackupConfig(cmd *cobra.Command, args []string) error {
	name, err := s.netappSPName(args[0])
	if err != nil {
		return err
	}
	cfg := &netapp.BackupConfig{}
	var fields []string
	if s.flagNetAppSPBackupCfgVault != "" {
		cfg.BackupVault = s.flagNetAppSPBackupCfgVault
		fields = append(fields, "backup_config.backup_vault")
	}
	if len(s.flagNetAppSPBackupCfgPolicies) > 0 {
		cfg.BackupPolicies = s.flagNetAppSPBackupCfgPolicies
		fields = append(fields, "backup_config.backup_policies")
	}
	if s.flagNetAppSPBackupCfgSchedule != "" {
		enable, perr := parseBoolFlag(s.flagNetAppSPBackupCfgSchedule)
		if perr != nil {
			return fmt.Errorf("--enable-scheduled-backups: %w", perr)
		}
		cfg.ScheduledBackupEnabled = enable
		cfg.ForceSendFields = append(cfg.ForceSendFields, "ScheduledBackupEnabled")
		fields = append(fields, "backup_config.scheduled_backup_enabled")
	}
	mask := s.flagNetAppSPUpdateMask
	if mask == "" {
		mask = joinMask(fields)
	}
	req := &netapp.UpdateBackupConfigRequest{
		BackupConfig: cfg,
		VolumeUuid:   s.flagNetAppSPBackupCfgVolumeUUID,
		UpdateMask:   mask,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.UpdateBackupConfig(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating backup config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update-backup-config request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppSPFormat)
}

func (s *netappCmds) runNetAppSPListBackupConfigs(cmd *cobra.Command, args []string) error {
	parent, err := s.netappSPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.VolumeBackupConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.StoragePools.BackupConfigs.List(parent).Context(ctx)
		if s.flagNetAppSPFilter != "" {
			call = call.Filter(s.flagNetAppSPFilter)
		}
		if s.flagNetAppSPPageSize > 0 {
			call = call.PageSize(s.flagNetAppSPPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppSPFormat)
}

func (s *netappCmds) runNetAppSPRestoreVolume(cmd *cobra.Command, args []string) error {
	name, err := s.netappSPName(args[0])
	if err != nil {
		return err
	}
	req := &netapp.RestoreVolumeRequest{
		BackupSource: &netapp.BackupSource{
			Backup:   s.flagNetAppSPRestoreBackup,
			FileList: s.flagNetAppSPRestoreFileList,
		},
		OntapVolumeTarget: &netapp.OntapVolumeTarget{
			VolumeUuid:             s.flagNetAppSPRestoreVolumeUUID,
			RestoreDestinationPath: s.flagNetAppSPRestoreDestPath,
		},
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.RestoreVolume(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring volume in storage pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restore-volume request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppSPFormat)
}

func (s *netappCmds) runNetAppSPSwitch(cmd *cobra.Command, args []string) error {
	name, err := s.netappSPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.StoragePools.Switch(name, &netapp.SwitchActiveReplicaZoneRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("switching storage pool zone: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Switch request issued for storage pool [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppSPFormat)
}
