package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp backup-vaults (#1199) ---

// --- backups subgroup ---

func (s *netappCmds) initNetappBackupVaults() {
	// backup-vaults command flags
	bvAll := []*cobra.Command{s.netappBVCreateCmd, s.netappBVDeleteCmd, s.netappBVDescribeCmd, s.netappBVListCmd, s.netappBVUpdateCmd}
	for _, c := range bvAll {
		c.Flags().StringVar(&s.flagNetAppBVLocation, "location", "", "Location for the backup vault (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppBVFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappBVCreateCmd, s.netappBVUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppBVConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the BackupVault body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappBVUpdateCmd.Flags().StringVar(&s.flagNetAppBVUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappBVListCmd.Flags().StringVar(&s.flagNetAppBVFilter, "filter", "", "Server-side filter expression")
	s.netappBVListCmd.Flags().Int64Var(&s.flagNetAppBVPageSize, "page-size", 0, "Maximum number of results per page")

	s.netappBVCmd.AddCommand(bvAll...)

	// backups subgroup
	bkAll := []*cobra.Command{s.netappBackupCreateCmd, s.netappBackupDeleteCmd, s.netappBackupDescribeCmd, s.netappBackupListCmd, s.netappBackupUpdateCmd}
	for _, c := range bkAll {
		c.Flags().StringVar(&s.flagNetAppBVLocation, "location", "", "Location for the backup (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppBackupVault, "backup-vault", "", "Backup vault that owns the backup (required)")
		_ = c.MarkFlagRequired("backup-vault")
		c.Flags().StringVar(&s.flagNetAppBVFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappBackupCreateCmd, s.netappBackupUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppBackupConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Backup body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappBackupUpdateCmd.Flags().StringVar(&s.flagNetAppBackupUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappBackupListCmd.Flags().StringVar(&s.flagNetAppBackupFilter, "filter", "", "Server-side filter expression")
	s.netappBackupListCmd.Flags().Int64Var(&s.flagNetAppBackupPageSize, "page-size", 0, "Maximum number of results per page")
	s.netappBackupCreateCmd.Flags().StringVar(&s.flagNetAppBackupOntapSrc, "ontap-source", "",
		"ONTAP source for the backup, as a comma-separated key=value list. "+
			"Keys: storage-pool (required), volume-uuid (required), snapshot-uuid (optional).")

	s.netappBackupsCmd.AddCommand(bkAll...)
	s.netappBVCmd.AddCommand(s.netappBackupsCmd)
	s.netappCmd.AddCommand(s.netappBVCmd)
}

func (s *netappCmds) netappBVParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, s.flagNetAppBVLocation), nil
}

func (s *netappCmds) netappBVName(id string) (string, error) {
	parent, err := s.netappBVParent()
	if err != nil {
		return "", err
	}
	return netappChild("backupVaults", id, parent), nil
}

func (s *netappCmds) netappBackupParent() (string, error) {
	bv, err := s.netappBVName(s.flagNetAppBackupVault)
	if err != nil {
		return "", err
	}
	return bv, nil
}

func (s *netappCmds) netappBackupName(id string) (string, error) {
	parent, err := s.netappBackupParent()
	if err != nil {
		return "", err
	}
	return netappChild("backups", id, parent), nil
}

func (s *netappCmds) runNetAppBVCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappBVParent()
	if err != nil {
		return err
	}
	body := &netapp.BackupVault{}
	if err := loadYAMLOrJSONInto(s.flagNetAppBVConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupVaults.Create(parent, body).BackupVaultId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup vault: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for backup vault [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBVFormat)
}

func (s *netappCmds) runNetAppBVDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappBVName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupVaults.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup vault: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for backup vault [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBVFormat)
}

func (s *netappCmds) runNetAppBVDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappBVName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.BackupVaults.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup vault: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppBVFormat)
}

func (s *netappCmds) runNetAppBVList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappBVParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.BackupVault
	pageToken := ""
	for {
		call := svc.Projects.Locations.BackupVaults.List(parent).Context(ctx)
		if s.flagNetAppBVFilter != "" {
			call = call.Filter(s.flagNetAppBVFilter)
		}
		if s.flagNetAppBVPageSize > 0 {
			call = call.PageSize(s.flagNetAppBVPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppBVFormat)
}

func (s *netappCmds) runNetAppBVUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappBVName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.BackupVault{}
	if err := loadYAMLOrJSONInto(s.flagNetAppBVConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppBVUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Update request issued for backup vault [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBVFormat)
}

func (s *netappCmds) runNetAppBackupCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappBackupParent()
	if err != nil {
		return err
	}
	body := &netapp.Backup{}
	if err := loadYAMLOrJSONInto(s.flagNetAppBackupConfigFile, body); err != nil {
		return err
	}
	if s.flagNetAppBackupOntapSrc != "" {
		os, perr := parseNetAppOntapSource(s.flagNetAppBackupOntapSrc)
		if perr != nil {
			return perr
		}
		body.OntapSource = os
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupVaults.Backups.Create(parent, body).BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for backup [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBVFormat)
}

func parseNetAppOntapSource(spec string) (*netapp.OntapSource, error) {
	kv := parseKeyValueSpec(spec)
	os := &netapp.OntapSource{
		StoragePool:  kv["storage-pool"],
		VolumeUuid:   kv["volume-uuid"],
		SnapshotUuid: kv["snapshot-uuid"],
	}
	if os.StoragePool == "" || os.VolumeUuid == "" {
		return nil, fmt.Errorf("--ontap-source requires storage-pool and volume-uuid keys")
	}
	return os, nil
}

func (s *netappCmds) runNetAppBackupDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappBackupName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupVaults.Backups.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for backup [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBVFormat)
}

func (s *netappCmds) runNetAppBackupDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappBackupName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.BackupVaults.Backups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppBVFormat)
}

func (s *netappCmds) runNetAppBackupList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappBackupParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.Backup
	pageToken := ""
	for {
		call := svc.Projects.Locations.BackupVaults.Backups.List(parent).Context(ctx)
		if s.flagNetAppBackupFilter != "" {
			call = call.Filter(s.flagNetAppBackupFilter)
		}
		if s.flagNetAppBackupPageSize > 0 {
			call = call.PageSize(s.flagNetAppBackupPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppBVFormat)
}

func (s *netappCmds) runNetAppBackupUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappBackupName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.Backup{}
	if err := loadYAMLOrJSONInto(s.flagNetAppBackupConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppBackupUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Update request issued for backup [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBVFormat)
}
