package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	filestore "google.golang.org/api/file/v1"
)

// --- gcloud filestore (#944-#949) ---

var filestoreCmd = &cobra.Command{
	Use:   "filestore",
	Short: "Manage Filestore",
}

// Common flags shared across filestore subcommands. Filestore resources live
// under a location (either a region or a zone); users identify them either
// with --location plus a short id, or by passing the fully-qualified resource
// name.
var (
	flagFSLocation    string
	flagFSFile        string
	flagFSDescription string
	flagFSLabels      map[string]string
	flagFSFilter      string
	flagFSOrderBy     string
	flagFSPageSize    int64
	flagFSLimit       int64
	// Backup-specific fields.
	flagFSBackupSourceInstance  string
	flagFSBackupSourceFileShare string
	flagFSBackupSourceBackup    string
	// Instance-specific fields.
	flagFSInstanceTier             string
	flagFSInstanceNetwork          string
	flagFSInstanceFileShareName    string
	flagFSInstanceFileShareCapGB   int64
	flagFSInstanceSourceBackup     string
	flagFSInstanceTargetSnapshotID string
	flagFSInstanceRestoreFileShare string
)

// --- Backups ---

var filestoreBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Filestore backups"}

var filestoreBackupsCreateCmd = &cobra.Command{
	Use:   "create BACKUP",
	Short: "Create a Filestore backup",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSBackupsCreate,
}

var filestoreBackupsDeleteCmd = &cobra.Command{
	Use:   "delete BACKUP",
	Short: "Delete a Filestore backup",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSBackupsDelete,
}

var filestoreBackupsDescribeCmd = &cobra.Command{
	Use:   "describe BACKUP",
	Short: "Describe a Filestore backup",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSBackupsDescribe,
}

var filestoreBackupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Filestore backups",
	Args:  cobra.NoArgs,
	RunE:  runFSBackupsList,
}

var filestoreBackupsUpdateCmd = &cobra.Command{
	Use:   "update BACKUP",
	Short: "Update a Filestore backup",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSBackupsUpdate,
}

// --- Instances ---

var filestoreInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Filestore instances"}

var filestoreInstancesCreateCmd = &cobra.Command{
	Use:   "create INSTANCE",
	Short: "Create a Filestore instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSInstancesCreate,
}

var filestoreInstancesDeleteCmd = &cobra.Command{
	Use:   "delete INSTANCE",
	Short: "Delete a Filestore instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSInstancesDelete,
}

var filestoreInstancesDescribeCmd = &cobra.Command{
	Use:   "describe INSTANCE",
	Short: "Describe a Filestore instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSInstancesDescribe,
}

var filestoreInstancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Filestore instances",
	Args:  cobra.NoArgs,
	RunE:  runFSInstancesList,
}

var filestoreInstancesUpdateCmd = &cobra.Command{
	Use:   "update INSTANCE",
	Short: "Update a Filestore instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSInstancesUpdate,
}

var filestoreInstancesRestoreCmd = &cobra.Command{
	Use:   "restore INSTANCE",
	Short: "Restore a Filestore instance from a backup",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSInstancesRestore,
}

var filestoreInstancesRevertCmd = &cobra.Command{
	Use:   "revert INSTANCE",
	Short: "Revert a Filestore instance to a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSInstancesRevert,
}

// --- Locations ---

var filestoreLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Filestore locations"}

var filestoreLocationsDescribeCmd = &cobra.Command{
	Use:   "describe LOCATION",
	Short: "Describe a Filestore location",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSLocationsDescribe,
}

var filestoreLocationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Filestore locations",
	Args:  cobra.NoArgs,
	RunE:  runFSLocationsList,
}

// --- Operations ---

var filestoreOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Filestore operations"}

var filestoreOperationsCancelCmd = &cobra.Command{
	Use:   "cancel OPERATION",
	Short: "Cancel a Filestore operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSOperationsCancel,
}

var filestoreOperationsDeleteCmd = &cobra.Command{
	Use:   "delete OPERATION",
	Short: "Delete a Filestore operation record",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSOperationsDelete,
}

var filestoreOperationsDescribeCmd = &cobra.Command{
	Use:   "describe OPERATION",
	Short: "Describe a Filestore operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSOperationsDescribe,
}

var filestoreOperationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Filestore operations",
	Args:  cobra.NoArgs,
	RunE:  runFSOperationsList,
}

var filestoreOperationsWaitCmd = &cobra.Command{
	Use:   "wait OPERATION",
	Short: "Poll a Filestore operation until it completes",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSOperationsWait,
}

// --- Regions / Zones ---
//
// gcloud python exposes both `filestore regions list` and `filestore zones
// list`, filtered subsets of the locations list. We render them the same way.

var filestoreRegionsCmd = &cobra.Command{Use: "regions", Short: "List Filestore regions"}
var filestoreRegionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Filestore regions",
	Args:  cobra.NoArgs,
	RunE:  runFSRegionsList,
}

var filestoreZonesCmd = &cobra.Command{Use: "zones", Short: "List Filestore zones"}
var filestoreZonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Filestore zones",
	Args:  cobra.NoArgs,
	RunE:  runFSZonesList,
}

func init() {
	// Backup flags.
	for _, c := range []*cobra.Command{filestoreBackupsCreateCmd, filestoreBackupsDeleteCmd, filestoreBackupsDescribeCmd, filestoreBackupsUpdateCmd} {
		c.Flags().StringVar(&flagFSLocation, "region", "", "Region containing the backup")
	}
	filestoreBackupsCreateCmd.Flags().StringVar(&flagFSDescription, "description", "", "Backup description")
	filestoreBackupsCreateCmd.Flags().StringToStringVar(&flagFSLabels, "labels", nil, "Labels (key=value)")
	filestoreBackupsCreateCmd.Flags().StringVar(&flagFSBackupSourceInstance, "source-instance", "", "Source instance resource name")
	filestoreBackupsCreateCmd.Flags().StringVar(&flagFSBackupSourceFileShare, "source-file-share", "", "Source file share")
	filestoreBackupsCreateCmd.Flags().StringVar(&flagFSFile, "config-from-file", "", "YAML/JSON file with the backup spec")
	filestoreBackupsListCmd.Flags().StringVar(&flagFSLocation, "region", "", "Region to list backups in (defaults to all with `-`)")
	filestoreBackupsListCmd.Flags().StringVar(&flagFSFilter, "filter", "", "Server-side filter expression")
	filestoreBackupsListCmd.Flags().StringVar(&flagFSOrderBy, "order-by", "", "Server-side ordering expression")
	filestoreBackupsListCmd.Flags().Int64Var(&flagFSPageSize, "page-size", 0, "Number of results per page")
	filestoreBackupsListCmd.Flags().Int64Var(&flagFSLimit, "limit", 0, "Maximum number of results to return")
	filestoreBackupsUpdateCmd.Flags().StringVar(&flagFSDescription, "description", "", "New description")
	filestoreBackupsUpdateCmd.Flags().StringToStringVar(&flagFSLabels, "labels", nil, "New labels (replaces the existing set)")
	filestoreBackupsUpdateCmd.Flags().StringVar(&flagFSFile, "config-from-file", "", "YAML/JSON file with the backup patch")

	filestoreBackupsCmd.AddCommand(filestoreBackupsCreateCmd, filestoreBackupsDeleteCmd,
		filestoreBackupsDescribeCmd, filestoreBackupsListCmd, filestoreBackupsUpdateCmd)

	// Instance flags. Filestore uses zonal or regional locations; we call the
	// flag --zone for parity with gcloud python but let --region alias it.
	for _, c := range []*cobra.Command{
		filestoreInstancesCreateCmd, filestoreInstancesDeleteCmd, filestoreInstancesDescribeCmd,
		filestoreInstancesUpdateCmd, filestoreInstancesRestoreCmd, filestoreInstancesRevertCmd,
	} {
		c.Flags().StringVar(&flagFSLocation, "zone", "", "Zone (or region) containing the instance")
	}
	filestoreInstancesCreateCmd.Flags().StringVar(&flagFSInstanceTier, "tier", "", "Service tier (e.g. STANDARD, PREMIUM, ENTERPRISE, ZONAL, REGIONAL)")
	filestoreInstancesCreateCmd.Flags().StringVar(&flagFSInstanceNetwork, "network", "", "VPC network to attach to (name=NAME[,reserved-ip-range=CIDR])")
	filestoreInstancesCreateCmd.Flags().StringVar(&flagFSInstanceFileShareName, "file-share-name", "", "Name of the file share")
	filestoreInstancesCreateCmd.Flags().Int64Var(&flagFSInstanceFileShareCapGB, "file-share-capacity-gb", 0, "File share capacity in GB")
	filestoreInstancesCreateCmd.Flags().StringVar(&flagFSDescription, "description", "", "Instance description")
	filestoreInstancesCreateCmd.Flags().StringToStringVar(&flagFSLabels, "labels", nil, "Labels (key=value)")
	filestoreInstancesCreateCmd.Flags().StringVar(&flagFSFile, "config-from-file", "", "YAML/JSON file with the instance spec")
	filestoreInstancesListCmd.Flags().StringVar(&flagFSLocation, "zone", "", "Zone or region to list instances in (`-` for all)")
	filestoreInstancesListCmd.Flags().StringVar(&flagFSFilter, "filter", "", "Server-side filter expression")
	filestoreInstancesListCmd.Flags().StringVar(&flagFSOrderBy, "order-by", "", "Server-side ordering expression")
	filestoreInstancesListCmd.Flags().Int64Var(&flagFSPageSize, "page-size", 0, "Number of results per page")
	filestoreInstancesListCmd.Flags().Int64Var(&flagFSLimit, "limit", 0, "Maximum number of results to return")
	filestoreInstancesUpdateCmd.Flags().StringVar(&flagFSDescription, "description", "", "New description")
	filestoreInstancesUpdateCmd.Flags().StringToStringVar(&flagFSLabels, "labels", nil, "Labels (key=value)")
	filestoreInstancesUpdateCmd.Flags().StringVar(&flagFSFile, "config-from-file", "", "YAML/JSON file with the instance patch")
	filestoreInstancesRestoreCmd.Flags().StringVar(&flagFSInstanceSourceBackup, "source-backup", "", "Backup resource name (required)")
	filestoreInstancesRestoreCmd.Flags().StringVar(&flagFSInstanceRestoreFileShare, "file-share", "", "File share on the instance to restore into (required)")
	filestoreInstancesRestoreCmd.MarkFlagRequired("source-backup")
	filestoreInstancesRestoreCmd.MarkFlagRequired("file-share")
	filestoreInstancesRevertCmd.Flags().StringVar(&flagFSInstanceTargetSnapshotID, "snapshot", "", "Snapshot short id to revert to (required)")
	filestoreInstancesRevertCmd.MarkFlagRequired("snapshot")
	filestoreInstancesCmd.AddCommand(
		filestoreInstancesCreateCmd, filestoreInstancesDeleteCmd, filestoreInstancesDescribeCmd,
		filestoreInstancesListCmd, filestoreInstancesUpdateCmd, filestoreInstancesRestoreCmd,
		filestoreInstancesRevertCmd,
	)

	// Locations.
	filestoreLocationsListCmd.Flags().Int64Var(&flagFSPageSize, "page-size", 0, "Number of results per page")
	filestoreLocationsListCmd.Flags().Int64Var(&flagFSLimit, "limit", 0, "Maximum number of results to return")
	filestoreLocationsListCmd.Flags().StringVar(&flagFSFilter, "filter", "", "Server-side filter expression")
	filestoreLocationsCmd.AddCommand(filestoreLocationsDescribeCmd, filestoreLocationsListCmd)

	// Operations.
	for _, c := range []*cobra.Command{
		filestoreOperationsCancelCmd, filestoreOperationsDeleteCmd,
		filestoreOperationsDescribeCmd, filestoreOperationsWaitCmd,
	} {
		c.Flags().StringVar(&flagFSLocation, "location", "", "Location containing the operation")
	}
	filestoreOperationsListCmd.Flags().StringVar(&flagFSLocation, "location", "-", "Location to list operations in (`-` for all)")
	filestoreOperationsListCmd.Flags().StringVar(&flagFSFilter, "filter", "", "Server-side filter expression")
	filestoreOperationsListCmd.Flags().Int64Var(&flagFSPageSize, "page-size", 0, "Number of results per page")
	filestoreOperationsListCmd.Flags().Int64Var(&flagFSLimit, "limit", 0, "Maximum number of results to return")
	filestoreOperationsCmd.AddCommand(
		filestoreOperationsCancelCmd, filestoreOperationsDeleteCmd,
		filestoreOperationsDescribeCmd, filestoreOperationsListCmd, filestoreOperationsWaitCmd,
	)

	// Regions & zones lists share flags for filtering.
	for _, c := range []*cobra.Command{filestoreRegionsListCmd, filestoreZonesListCmd} {
		c.Flags().StringVar(&flagFSFilter, "filter", "", "Server-side filter expression")
		c.Flags().Int64Var(&flagFSLimit, "limit", 0, "Maximum number of results to return")
	}
	filestoreRegionsCmd.AddCommand(filestoreRegionsListCmd)
	filestoreZonesCmd.AddCommand(filestoreZonesListCmd)

	filestoreCmd.AddCommand(
		filestoreBackupsCmd, filestoreInstancesCmd, filestoreLocationsCmd,
		filestoreOperationsCmd, filestoreRegionsCmd, filestoreZonesCmd,
	)
	rootCmd.AddCommand(filestoreCmd)
}

// --- Helpers ---

func fsRequireLocation() (string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", err
	}
	if flagFSLocation == "" {
		return "", "", fmt.Errorf("--zone/--region/--location is required")
	}
	return project, flagFSLocation, nil
}

func fsParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// fsResourceName qualifies a possibly-short id (e.g. "my-instance") into a full
// resource path (`projects/.../locations/.../<kind>/<id>`). If the id is
// already fully qualified it is returned unchanged.
func fsResourceName(id, project, location, kind string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", fsParent(project, location), kind, id)
}

// --- Backups impl ---

func runFSBackupsCreate(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	backup := &filestore.Backup{}
	if flagFSFile != "" {
		if err := loadYAMLOrJSONInto(flagFSFile, backup); err != nil {
			return err
		}
	}
	if flagFSDescription != "" {
		backup.Description = flagFSDescription
	}
	if len(flagFSLabels) > 0 {
		backup.Labels = flagFSLabels
	}
	if flagFSBackupSourceInstance != "" {
		backup.SourceInstance = flagFSBackupSourceInstance
	}
	if flagFSBackupSourceFileShare != "" {
		backup.SourceFileShare = flagFSBackupSourceFileShare
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Create(fsParent(project, location), backup).
		BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Printf("Create request issued for backup [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runFSBackupsDelete(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Delete(fsResourceName(args[0], project, location, "backups")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Printf("Delete request issued for backup [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runFSBackupsDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Backups.Get(fsResourceName(args[0], project, location, "backups")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, "")
}

func runFSBackupsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	location := flagFSLocation
	if location == "" {
		location = "-"
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*filestore.Backup
	pageToken := ""
	for {
		call := svc.Projects.Locations.Backups.List(fsParent(project, location)).Context(ctx)
		if flagFSPageSize > 0 {
			call = call.PageSize(flagFSPageSize)
		}
		if flagFSFilter != "" {
			call = call.Filter(flagFSFilter)
		}
		if flagFSOrderBy != "" {
			call = call.OrderBy(flagFSOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backups: %w", err)
		}
		all = append(all, resp.Backups...)
		if flagFSLimit > 0 && int64(len(all)) >= flagFSLimit {
			all = all[:flagFSLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runFSBackupsUpdate(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	body := &filestore.Backup{}
	if flagFSFile != "" {
		if err := loadYAMLOrJSONInto(flagFSFile, body); err != nil {
			return err
		}
	}
	var mask []string
	if flagFSDescription != "" {
		body.Description = flagFSDescription
		mask = append(mask, "description")
	}
	if len(flagFSLabels) > 0 {
		body.Labels = flagFSLabels
		mask = append(mask, "labels")
	}
	if flagFSFile != "" && len(mask) == 0 {
		mask = nonEmptyJSONFields(body)
	}
	if len(mask) == 0 {
		return fmt.Errorf("nothing to update: pass --description, --labels, or --config-from-file")
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Patch(fsResourceName(args[0], project, location, "backups"), body).
		UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	fmt.Printf("Update request issued for backup [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Instances impl ---

func runFSInstancesCreate(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	inst := &filestore.Instance{}
	if flagFSFile != "" {
		if err := loadYAMLOrJSONInto(flagFSFile, inst); err != nil {
			return err
		}
	}
	if flagFSInstanceTier != "" {
		inst.Tier = flagFSInstanceTier
	}
	if flagFSDescription != "" {
		inst.Description = flagFSDescription
	}
	if len(flagFSLabels) > 0 {
		inst.Labels = flagFSLabels
	}
	if flagFSInstanceNetwork != "" {
		net := &filestore.NetworkConfig{Network: flagFSInstanceNetwork}
		inst.Networks = append(inst.Networks, net)
	}
	if flagFSInstanceFileShareName != "" || flagFSInstanceFileShareCapGB > 0 {
		share := &filestore.FileShareConfig{
			Name:       flagFSInstanceFileShareName,
			CapacityGb: flagFSInstanceFileShareCapGB,
		}
		inst.FileShares = append(inst.FileShares, share)
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(fsParent(project, location), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Printf("Create request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runFSInstancesDelete(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(fsResourceName(args[0], project, location, "instances")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Printf("Delete request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runFSInstancesDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	inst, err := svc.Projects.Locations.Instances.Get(fsResourceName(args[0], project, location, "instances")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(inst, "")
}

func runFSInstancesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	location := flagFSLocation
	if location == "" {
		location = "-"
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*filestore.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(fsParent(project, location)).Context(ctx)
		if flagFSPageSize > 0 {
			call = call.PageSize(flagFSPageSize)
		}
		if flagFSFilter != "" {
			call = call.Filter(flagFSFilter)
		}
		if flagFSOrderBy != "" {
			call = call.OrderBy(flagFSOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if flagFSLimit > 0 && int64(len(all)) >= flagFSLimit {
			all = all[:flagFSLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runFSInstancesUpdate(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	body := &filestore.Instance{}
	if flagFSFile != "" {
		if err := loadYAMLOrJSONInto(flagFSFile, body); err != nil {
			return err
		}
	}
	var mask []string
	if flagFSDescription != "" {
		body.Description = flagFSDescription
		mask = append(mask, "description")
	}
	if len(flagFSLabels) > 0 {
		body.Labels = flagFSLabels
		mask = append(mask, "labels")
	}
	if flagFSFile != "" && len(mask) == 0 {
		mask = nonEmptyJSONFields(body)
	}
	if len(mask) == 0 {
		return fmt.Errorf("nothing to update: pass --description, --labels, or --config-from-file")
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Patch(fsResourceName(args[0], project, location, "instances"), body).
		UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	fmt.Printf("Update request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runFSInstancesRestore(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &filestore.RestoreInstanceRequest{
		FileShare:    flagFSInstanceRestoreFileShare,
		SourceBackup: flagFSInstanceSourceBackup,
	}
	op, err := svc.Projects.Locations.Instances.Restore(fsResourceName(args[0], project, location, "instances"), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring instance: %w", err)
	}
	fmt.Printf("Restore request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runFSInstancesRevert(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &filestore.RevertInstanceRequest{TargetSnapshotId: flagFSInstanceTargetSnapshotID}
	op, err := svc.Projects.Locations.Instances.Revert(fsResourceName(args[0], project, location, "instances"), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reverting instance: %w", err)
	}
	fmt.Printf("Revert request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Locations impl ---

func runFSLocationsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(fsParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(loc, "")
}

func runFSLocationsList(cmd *cobra.Command, args []string) error {
	locs, err := fsListLocations()
	if err != nil {
		return err
	}
	return emitFormatted(locs, "")
}

// fsListLocations paginates the whole locations list, honoring the shared
// --filter/--page-size/--limit flags.
func fsListLocations() ([]*filestore.Location, error) {
	project, err := resolveProject()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	var all []*filestore.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagFSPageSize > 0 {
			call = call.PageSize(flagFSPageSize)
		}
		if flagFSFilter != "" {
			call = call.Filter(flagFSFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if flagFSLimit > 0 && int64(len(all)) >= flagFSLimit {
			all = all[:flagFSLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return all, nil
}

// --- Operations impl ---

func runFSOperationsCancel(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(fsResourceName(args[0], project, location, "operations"), &filestore.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runFSOperationsDelete(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(fsResourceName(args[0], project, location, "operations")).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runFSOperationsDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(fsResourceName(args[0], project, location, "operations")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, "")
}

func runFSOperationsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	location := flagFSLocation
	if location == "" {
		location = "-"
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*filestore.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(fsParent(project, location)).Context(ctx)
		if flagFSPageSize > 0 {
			call = call.PageSize(flagFSPageSize)
		}
		if flagFSFilter != "" {
			call = call.Filter(flagFSFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if flagFSLimit > 0 && int64(len(all)) >= flagFSLimit {
			all = all[:flagFSLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runFSOperationsWait(cmd *cobra.Command, args []string) error {
	project, location, err := fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FilestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := fsResourceName(args[0], project, location, "operations")
	for {
		op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation %s failed: %s", name, op.Error.Message)
			}
			return emitFormatted(op, "")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// --- Regions / Zones impl ---

// runFSRegionsList lists locations whose metadata indicates a region (the
// service exposes both regions and zones under the same endpoint; regions are
// those without an availability-zones-style suffix like "-a"/"-b"/"-c").
func runFSRegionsList(cmd *cobra.Command, args []string) error {
	locs, err := fsListLocations()
	if err != nil {
		return err
	}
	regions := make([]*filestore.Location, 0, len(locs))
	for _, l := range locs {
		if !isZoneLocationID(l.LocationId) {
			regions = append(regions, l)
		}
	}
	return emitFormatted(regions, "")
}

func runFSZonesList(cmd *cobra.Command, args []string) error {
	locs, err := fsListLocations()
	if err != nil {
		return err
	}
	zones := make([]*filestore.Location, 0, len(locs))
	for _, l := range locs {
		if isZoneLocationID(l.LocationId) {
			zones = append(zones, l)
		}
	}
	return emitFormatted(zones, "")
}

// isZoneLocationID reports whether the location id looks like a zone (i.e.
// ends with -a..-z after a dash). Filestore returns both regions
// ("us-central1") and zones ("us-central1-a") from the same list endpoint.
func isZoneLocationID(id string) bool {
	if len(id) < 2 || id[len(id)-2] != '-' {
		return false
	}
	last := id[len(id)-1]
	return last >= 'a' && last <= 'z'
}
