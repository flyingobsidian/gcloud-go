package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	filestore "google.golang.org/api/file/v1"
)

// --- gcloud filestore (#944-#949) ---

// Common flags shared across filestore subcommands. Filestore resources live
// under a location (either a region or a zone); users identify them either
// with --location plus a short id, or by passing the fully-qualified resource
// name.

// --- Backups ---

// --- Instances ---

// --- Locations ---

// --- Operations ---

// --- Regions / Zones ---
//
// gcloud python exposes both `filestore regions list` and `filestore zones
// list`, filtered subsets of the locations list. We render them the same way.

func (s *filestoreCmds) initFilestore() {
	// Backup flags.
	for _, c := range []*cobra.Command{s.filestoreBackupsCreateCmd, s.filestoreBackupsDeleteCmd, s.filestoreBackupsDescribeCmd, s.filestoreBackupsUpdateCmd} {
		c.Flags().StringVar(&s.flagFSLocation, "region", "", "Region containing the backup")
	}
	s.filestoreBackupsCreateCmd.Flags().StringVar(&s.flagFSDescription, "description", "", "Backup description")
	s.filestoreBackupsCreateCmd.Flags().StringToStringVar(&s.flagFSLabels, "labels", nil, "Labels (key=value)")
	s.filestoreBackupsCreateCmd.Flags().StringVar(&s.flagFSBackupSourceInstance, "source-instance", "", "Source instance resource name")
	s.filestoreBackupsCreateCmd.Flags().StringVar(&s.flagFSBackupSourceFileShare, "source-file-share", "", "Source file share")
	s.filestoreBackupsCreateCmd.Flags().StringVar(&s.flagFSFile, "config-from-file", "", "YAML/JSON file with the backup spec")
	s.filestoreBackupsListCmd.Flags().StringVar(&s.flagFSLocation, "region", "", "Region to list backups in (defaults to all with `-`)")
	s.filestoreBackupsListCmd.Flags().StringVar(&s.flagFSFilter, "filter", "", "Server-side filter expression")
	s.filestoreBackupsListCmd.Flags().StringVar(&s.flagFSOrderBy, "order-by", "", "Server-side ordering expression")
	s.filestoreBackupsListCmd.Flags().Int64Var(&s.flagFSPageSize, "page-size", 0, "Number of results per page")
	s.filestoreBackupsListCmd.Flags().Int64Var(&s.flagFSLimit, "limit", 0, "Maximum number of results to return")
	s.filestoreBackupsUpdateCmd.Flags().StringVar(&s.flagFSDescription, "description", "", "New description")
	s.filestoreBackupsUpdateCmd.Flags().StringToStringVar(&s.flagFSLabels, "labels", nil, "New labels (replaces the existing set)")
	s.filestoreBackupsUpdateCmd.Flags().StringVar(&s.flagFSFile, "config-from-file", "", "YAML/JSON file with the backup patch")

	s.filestoreBackupsCmd.AddCommand(s.filestoreBackupsCreateCmd, s.filestoreBackupsDeleteCmd,
		s.filestoreBackupsDescribeCmd, s.filestoreBackupsListCmd, s.filestoreBackupsUpdateCmd)

	// Instance flags. Filestore uses zonal or regional locations; we call the
	// flag --zone for parity with gcloud python but let --region alias it.
	for _, c := range []*cobra.Command{
		s.filestoreInstancesCreateCmd, s.filestoreInstancesDeleteCmd, s.filestoreInstancesDescribeCmd,
		s.filestoreInstancesUpdateCmd, s.filestoreInstancesRestoreCmd, s.filestoreInstancesRevertCmd,
	} {
		c.Flags().StringVar(&s.flagFSLocation, "zone", "", "Zone (or region) containing the instance")
	}
	s.filestoreInstancesCreateCmd.Flags().StringVar(&s.flagFSInstanceTier, "tier", "", "Service tier (e.g. STANDARD, PREMIUM, ENTERPRISE, ZONAL, REGIONAL)")
	s.filestoreInstancesCreateCmd.Flags().StringVar(&s.flagFSInstanceNetwork, "network", "", "Network config: name=NAME[,reserved-ip-range=RANGE][,connect-mode=MODE][,address-mode=MODE][,psc-endpoint-project=PROJECT][,psc-requested-ip-address=IP]")
	s.filestoreInstancesCreateCmd.Flags().StringVar(&s.flagFSInstanceFileShareName, "file-share-name", "", "Name of the file share")
	s.filestoreInstancesCreateCmd.Flags().Int64Var(&s.flagFSInstanceFileShareCapGB, "file-share-capacity-gb", 0, "File share capacity in GB")
	s.filestoreInstancesCreateCmd.Flags().StringVar(&s.flagFSDescription, "description", "", "Instance description")
	s.filestoreInstancesCreateCmd.Flags().StringToStringVar(&s.flagFSLabels, "labels", nil, "Labels (key=value)")
	s.filestoreInstancesCreateCmd.Flags().StringVar(&s.flagFSFile, "config-from-file", "", "YAML/JSON file with the instance spec")
	s.filestoreInstancesListCmd.Flags().StringVar(&s.flagFSLocation, "zone", "", "Zone or region to list instances in (`-` for all)")
	s.filestoreInstancesListCmd.Flags().StringVar(&s.flagFSFilter, "filter", "", "Server-side filter expression")
	s.filestoreInstancesListCmd.Flags().StringVar(&s.flagFSOrderBy, "order-by", "", "Server-side ordering expression")
	s.filestoreInstancesListCmd.Flags().Int64Var(&s.flagFSPageSize, "page-size", 0, "Number of results per page")
	s.filestoreInstancesListCmd.Flags().Int64Var(&s.flagFSLimit, "limit", 0, "Maximum number of results to return")
	s.filestoreInstancesUpdateCmd.Flags().StringVar(&s.flagFSDescription, "description", "", "New description")
	s.filestoreInstancesUpdateCmd.Flags().StringToStringVar(&s.flagFSLabels, "labels", nil, "Labels (key=value)")
	s.filestoreInstancesUpdateCmd.Flags().StringVar(&s.flagFSFile, "config-from-file", "", "YAML/JSON file with the instance patch")
	s.filestoreInstancesRestoreCmd.Flags().StringVar(&s.flagFSInstanceSourceBackup, "source-backup", "", "Backup resource name (required)")
	s.filestoreInstancesRestoreCmd.Flags().StringVar(&s.flagFSInstanceRestoreFileShare, "file-share", "", "File share on the instance to restore into (required)")
	s.filestoreInstancesRestoreCmd.MarkFlagRequired("source-backup")
	s.filestoreInstancesRestoreCmd.MarkFlagRequired("file-share")
	s.filestoreInstancesRevertCmd.Flags().StringVar(&s.flagFSInstanceTargetSnapshotID, "snapshot", "", "Snapshot short id to revert to (required)")
	s.filestoreInstancesRevertCmd.MarkFlagRequired("snapshot")
	s.filestoreInstancesCmd.AddCommand(
		s.filestoreInstancesCreateCmd, s.filestoreInstancesDeleteCmd, s.filestoreInstancesDescribeCmd,
		s.filestoreInstancesListCmd, s.filestoreInstancesUpdateCmd, s.filestoreInstancesRestoreCmd,
		s.filestoreInstancesRevertCmd,
	)

	// Locations.
	s.filestoreLocationsListCmd.Flags().Int64Var(&s.flagFSPageSize, "page-size", 0, "Number of results per page")
	s.filestoreLocationsListCmd.Flags().Int64Var(&s.flagFSLimit, "limit", 0, "Maximum number of results to return")
	s.filestoreLocationsListCmd.Flags().StringVar(&s.flagFSFilter, "filter", "", "Server-side filter expression")
	s.filestoreLocationsCmd.AddCommand(s.filestoreLocationsDescribeCmd, s.filestoreLocationsListCmd)

	// Operations.
	for _, c := range []*cobra.Command{
		s.filestoreOperationsCancelCmd, s.filestoreOperationsDeleteCmd,
		s.filestoreOperationsDescribeCmd, s.filestoreOperationsWaitCmd,
	} {
		c.Flags().StringVar(&s.flagFSLocation, "location", "", "Location containing the operation")
	}
	s.filestoreOperationsListCmd.Flags().StringVar(&s.flagFSLocation, "location", "-", "Location to list operations in (`-` for all)")
	s.filestoreOperationsListCmd.Flags().StringVar(&s.flagFSFilter, "filter", "", "Server-side filter expression")
	s.filestoreOperationsListCmd.Flags().Int64Var(&s.flagFSPageSize, "page-size", 0, "Number of results per page")
	s.filestoreOperationsListCmd.Flags().Int64Var(&s.flagFSLimit, "limit", 0, "Maximum number of results to return")
	s.filestoreOperationsCmd.AddCommand(
		s.filestoreOperationsCancelCmd, s.filestoreOperationsDeleteCmd,
		s.filestoreOperationsDescribeCmd, s.filestoreOperationsListCmd, s.filestoreOperationsWaitCmd,
	)

	// Regions & zones lists share flags for filtering.
	for _, c := range []*cobra.Command{s.filestoreRegionsListCmd, s.filestoreZonesListCmd} {
		c.Flags().StringVar(&s.flagFSFilter, "filter", "", "Server-side filter expression")
		c.Flags().Int64Var(&s.flagFSLimit, "limit", 0, "Maximum number of results to return")
	}
	s.filestoreRegionsCmd.AddCommand(s.filestoreRegionsListCmd)
	s.filestoreZonesCmd.AddCommand(s.filestoreZonesListCmd)

	s.filestoreCmd.AddCommand(
		s.filestoreBackupsCmd, s.filestoreInstancesCmd, s.filestoreLocationsCmd,
		s.filestoreOperationsCmd, s.filestoreRegionsCmd, s.filestoreZonesCmd,
	)

}

// --- Helpers ---

func (s *filestoreCmds) fsRequireLocation() (string, string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", "", err
	}
	if s.flagFSLocation == "" {
		return "", "", fmt.Errorf("--zone/--region/--location is required")
	}
	return project, s.flagFSLocation, nil
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

func (s *filestoreCmds) runFSBackupsCreate(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	backup := &filestore.Backup{}
	if s.flagFSFile != "" {
		if err := loadYAMLOrJSONInto(s.flagFSFile, backup); err != nil {
			return err
		}
	}
	if s.flagFSDescription != "" {
		backup.Description = s.flagFSDescription
	}
	if len(s.flagFSLabels) > 0 {
		backup.Labels = s.flagFSLabels
	}
	if s.flagFSBackupSourceInstance != "" {
		backup.SourceInstance = s.flagFSBackupSourceInstance
	}
	if s.flagFSBackupSourceFileShare != "" {
		backup.SourceFileShare = s.flagFSBackupSourceFileShare
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Create(fsParent(project, location), backup).
		BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for backup [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *filestoreCmds) runFSBackupsDelete(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Delete(fsResourceName(args[0], project, location, "backups")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for backup [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *filestoreCmds) runFSBackupsDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Backups.Get(fsResourceName(args[0], project, location, "backups")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *filestoreCmds) runFSBackupsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	location := s.flagFSLocation
	if location == "" {
		location = "-"
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	var all []*filestore.Backup
	pageToken := ""
	for {
		call := svc.Projects.Locations.Backups.List(fsParent(project, location)).Context(ctx)
		if s.flagFSPageSize > 0 {
			call = call.PageSize(s.flagFSPageSize)
		}
		if s.flagFSFilter != "" {
			call = call.Filter(s.flagFSFilter)
		}
		if s.flagFSOrderBy != "" {
			call = call.OrderBy(s.flagFSOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backups: %w", err)
		}
		all = append(all, resp.Backups...)
		if s.flagFSLimit > 0 && int64(len(all)) >= s.flagFSLimit {
			all = all[:s.flagFSLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *filestoreCmds) runFSBackupsUpdate(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	body := &filestore.Backup{}
	if s.flagFSFile != "" {
		if err := loadYAMLOrJSONInto(s.flagFSFile, body); err != nil {
			return err
		}
	}
	var mask []string
	if s.flagFSDescription != "" {
		body.Description = s.flagFSDescription
		mask = append(mask, "description")
	}
	if len(s.flagFSLabels) > 0 {
		body.Labels = s.flagFSLabels
		mask = append(mask, "labels")
	}
	if s.flagFSFile != "" && len(mask) == 0 {
		mask = nonEmptyJSONFields(body)
	}
	if len(mask) == 0 {
		return fmt.Errorf("nothing to update: pass --description, --labels, or --config-from-file")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Backups.Patch(fsResourceName(args[0], project, location, "backups"), body).
		UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for backup [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Instances impl ---

func (s *filestoreCmds) runFSInstancesCreate(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	inst := &filestore.Instance{}
	if s.flagFSFile != "" {
		if err := loadYAMLOrJSONInto(s.flagFSFile, inst); err != nil {
			return err
		}
	}
	if s.flagFSInstanceTier != "" {
		inst.Tier = s.flagFSInstanceTier
	}
	if s.flagFSDescription != "" {
		inst.Description = s.flagFSDescription
	}
	if len(s.flagFSLabels) > 0 {
		inst.Labels = s.flagFSLabels
	}
	if s.flagFSInstanceNetwork != "" {
		net, err := fsParseNetwork(s.flagFSInstanceNetwork)
		if err != nil {
			return err
		}
		inst.Networks = append(inst.Networks, net)
	}
	if s.flagFSInstanceFileShareName != "" || s.flagFSInstanceFileShareCapGB > 0 {
		share := &filestore.FileShareConfig{
			Name:       s.flagFSInstanceFileShareName,
			CapacityGb: s.flagFSInstanceFileShareCapGB,
		}
		inst.FileShares = append(inst.FileShares, share)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(fsParent(project, location), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *filestoreCmds) runFSInstancesDelete(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(fsResourceName(args[0], project, location, "instances")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *filestoreCmds) runFSInstancesDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	inst, err := svc.Projects.Locations.Instances.Get(fsResourceName(args[0], project, location, "instances")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, inst, "")
}

func (s *filestoreCmds) runFSInstancesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	location := s.flagFSLocation
	if location == "" {
		location = "-"
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	var all []*filestore.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(fsParent(project, location)).Context(ctx)
		if s.flagFSPageSize > 0 {
			call = call.PageSize(s.flagFSPageSize)
		}
		if s.flagFSFilter != "" {
			call = call.Filter(s.flagFSFilter)
		}
		if s.flagFSOrderBy != "" {
			call = call.OrderBy(s.flagFSOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if s.flagFSLimit > 0 && int64(len(all)) >= s.flagFSLimit {
			all = all[:s.flagFSLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *filestoreCmds) runFSInstancesUpdate(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	body := &filestore.Instance{}
	if s.flagFSFile != "" {
		if err := loadYAMLOrJSONInto(s.flagFSFile, body); err != nil {
			return err
		}
	}
	var mask []string
	if s.flagFSDescription != "" {
		body.Description = s.flagFSDescription
		mask = append(mask, "description")
	}
	if len(s.flagFSLabels) > 0 {
		body.Labels = s.flagFSLabels
		mask = append(mask, "labels")
	}
	if s.flagFSFile != "" && len(mask) == 0 {
		mask = nonEmptyJSONFields(body)
	}
	if len(mask) == 0 {
		return fmt.Errorf("nothing to update: pass --description, --labels, or --config-from-file")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Patch(fsResourceName(args[0], project, location, "instances"), body).
		UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *filestoreCmds) runFSInstancesRestore(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	req := &filestore.RestoreInstanceRequest{
		FileShare:    s.flagFSInstanceRestoreFileShare,
		SourceBackup: s.flagFSInstanceSourceBackup,
	}
	op, err := svc.Projects.Locations.Instances.Restore(fsResourceName(args[0], project, location, "instances"), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restore request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *filestoreCmds) runFSInstancesRevert(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	req := &filestore.RevertInstanceRequest{TargetSnapshotId: s.flagFSInstanceTargetSnapshotID}
	op, err := svc.Projects.Locations.Instances.Revert(fsResourceName(args[0], project, location, "instances"), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reverting instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Revert request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Locations impl ---

func (s *filestoreCmds) runFSLocationsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(fsParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, loc, "")
}

func (s *filestoreCmds) runFSLocationsList(cmd *cobra.Command, args []string) error {
	locs, err := s.fsListLocations()
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, locs, "")
}

// fsListLocations paginates the whole locations list, honoring the shared
// --filter/--page-size/--limit flags.
// fsNetworkKeys are the --network dict keys accepted by gcloud-python's v1
// track (network_arg_spec_beta_v1).
func fsNetworkKeys() []string {
	return []string{
		"address-mode", "connect-mode", "name", "psc-endpoint-project",
		"psc-requested-ip-address", "reserved-ip-range",
	}
}

// fsParseNetwork builds a NetworkConfig from the --network dict, mirroring
// FilestoreClient.MakeNetworkConfig. A bare value without '=' is taken as
// the network name.
func fsParseNetwork(spec string) (*filestore.NetworkConfig, error) {
	vals := map[string]string{}
	if !strings.Contains(spec, "=") {
		vals["name"] = spec
	} else {
		for _, part := range strings.Split(spec, ",") {
			k, v, ok := strings.Cut(part, "=")
			if !ok {
				return nil, fmt.Errorf("argument --network: Bad syntax for dict arg: [%s]", part)
			}
			valid := false
			for _, key := range fsNetworkKeys() {
				if k == key {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("argument --network: valid keys are [%s]; received: %s",
					strings.Join(fsNetworkKeys(), ", "), k)
			}
			vals[k] = v
		}
	}
	net := &filestore.NetworkConfig{
		Network:         vals["name"],
		ReservedIpRange: vals["reserved-ip-range"],
	}
	if m, ok := vals["address-mode"]; ok {
		if m != "MODE_IPV4" && m != "MODE_IPV6" {
			return nil, fmt.Errorf("[%s] is not a valid address-mode. Must be one of MODE_IPV4 or MODE_IPV6.", m)
		}
		net.Modes = []string{m}
	}
	connectMode := "DIRECT_PEERING"
	if m, ok := vals["connect-mode"]; ok {
		connectMode = m
	}
	switch connectMode {
	case "DIRECT_PEERING", "PRIVATE_SERVICE_ACCESS", "PRIVATE_SERVICE_CONNECT":
	default:
		return nil, fmt.Errorf("[%s] is not a valid connect-mode. Must be one of DIRECT_PEERING or PRIVATE_SERVICE_ACCESS.", connectMode)
	}
	net.ConnectMode = connectMode
	if p, ip := vals["psc-endpoint-project"], vals["psc-requested-ip-address"]; p != "" || ip != "" {
		net.PscConfig = &filestore.PscConfig{EndpointProject: p, RequestedIpAddress: ip}
	}
	return net, nil
}

func (s *filestoreCmds) fsListLocations() ([]*filestore.Location, error) {
	project, err := s.inv.project()
	if err != nil {
		return nil, err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return nil, err
	}
	var all []*filestore.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if s.flagFSPageSize > 0 {
			call = call.PageSize(s.flagFSPageSize)
		}
		if s.flagFSFilter != "" {
			call = call.Filter(s.flagFSFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if s.flagFSLimit > 0 && int64(len(all)) >= s.flagFSLimit {
			all = all[:s.flagFSLimit]
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

func (s *filestoreCmds) runFSOperationsCancel(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(fsResourceName(args[0], project, location, "operations"), &filestore.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *filestoreCmds) runFSOperationsDelete(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(fsResourceName(args[0], project, location, "operations")).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *filestoreCmds) runFSOperationsDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(fsResourceName(args[0], project, location, "operations")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, "")
}

func (s *filestoreCmds) runFSOperationsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	location := s.flagFSLocation
	if location == "" {
		location = "-"
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
	if err != nil {
		return err
	}
	var all []*filestore.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(fsParent(project, location)).Context(ctx)
		if s.flagFSPageSize > 0 {
			call = call.PageSize(s.flagFSPageSize)
		}
		if s.flagFSFilter != "" {
			call = call.Filter(s.flagFSFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if s.flagFSLimit > 0 && int64(len(all)) >= s.flagFSLimit {
			all = all[:s.flagFSLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *filestoreCmds) runFSOperationsWait(cmd *cobra.Command, args []string) error {
	project, location, err := s.fsRequireLocation()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Filestore(ctx)
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
			return emitFormattedTo(s.inv.stdout, op, "")
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
func (s *filestoreCmds) runFSRegionsList(cmd *cobra.Command, args []string) error {
	locs, err := s.fsListLocations()
	if err != nil {
		return err
	}
	regions := make([]*filestore.Location, 0, len(locs))
	for _, l := range locs {
		if !isZoneLocationID(l.LocationId) {
			regions = append(regions, l)
		}
	}
	return emitFormattedTo(s.inv.stdout, regions, "")
}

func (s *filestoreCmds) runFSZonesList(cmd *cobra.Command, args []string) error {
	locs, err := s.fsListLocations()
	if err != nil {
		return err
	}
	zones := make([]*filestore.Location, 0, len(locs))
	for _, l := range locs {
		if isZoneLocationID(l.LocationId) {
			zones = append(zones, l)
		}
	}
	return emitFormattedTo(s.inv.stdout, zones, "")
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

// filestoreCmds holds the commands and flag values of one invocation's "filestore"
// command group, so invocations don't share state (#2070).
type filestoreCmds struct {
	inv *invocation

	filestoreCmd      *cobra.Command
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
	// Instance-specific fields.
	flagFSInstanceTier             string
	flagFSInstanceNetwork          string
	flagFSInstanceFileShareName    string
	flagFSInstanceFileShareCapGB   int64
	flagFSInstanceSourceBackup     string
	flagFSInstanceTargetSnapshotID string
	flagFSInstanceRestoreFileShare string
	filestoreBackupsCmd            *cobra.Command
	filestoreBackupsCreateCmd      *cobra.Command
	filestoreBackupsDeleteCmd      *cobra.Command
	filestoreBackupsDescribeCmd    *cobra.Command
	filestoreBackupsListCmd        *cobra.Command
	filestoreBackupsUpdateCmd      *cobra.Command
	filestoreInstancesCmd          *cobra.Command
	filestoreInstancesCreateCmd    *cobra.Command
	filestoreInstancesDeleteCmd    *cobra.Command
	filestoreInstancesDescribeCmd  *cobra.Command
	filestoreInstancesListCmd      *cobra.Command
	filestoreInstancesUpdateCmd    *cobra.Command
	filestoreInstancesRestoreCmd   *cobra.Command
	filestoreInstancesRevertCmd    *cobra.Command
	filestoreLocationsCmd          *cobra.Command
	filestoreLocationsDescribeCmd  *cobra.Command
	filestoreLocationsListCmd      *cobra.Command
	filestoreOperationsCmd         *cobra.Command
	filestoreOperationsCancelCmd   *cobra.Command
	filestoreOperationsDeleteCmd   *cobra.Command
	filestoreOperationsDescribeCmd *cobra.Command
	filestoreOperationsListCmd     *cobra.Command
	filestoreOperationsWaitCmd     *cobra.Command
	filestoreRegionsCmd            *cobra.Command
	filestoreRegionsListCmd        *cobra.Command
	filestoreZonesCmd              *cobra.Command
	filestoreZonesListCmd          *cobra.Command
}

// newFilestoreCmd returns the "filestore" command group for inv.
func newFilestoreCmd(inv *invocation) *cobra.Command {
	return newFilestoreCmds(inv).filestoreCmd
}

// newFilestoreCmds builds the "filestore" command group for inv and returns its state.
func newFilestoreCmds(inv *invocation) *filestoreCmds {
	s := &filestoreCmds{inv: inv}
	s.filestoreCmd = &cobra.Command{
		Use:   "filestore",
		Short: "Manage Filestore",
	}
	s.filestoreBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Filestore backups"}
	s.filestoreBackupsCreateCmd = &cobra.Command{
		Use:   "create BACKUP",
		Short: "Create a Filestore backup",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSBackupsCreate,
	}
	s.filestoreBackupsDeleteCmd = &cobra.Command{
		Use:   "delete BACKUP",
		Short: "Delete a Filestore backup",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSBackupsDelete,
	}
	s.filestoreBackupsDescribeCmd = &cobra.Command{
		Use:   "describe BACKUP",
		Short: "Describe a Filestore backup",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSBackupsDescribe,
	}
	s.filestoreBackupsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Filestore backups",
		Args:  cobra.NoArgs,
		RunE:  s.runFSBackupsList,
	}
	s.filestoreBackupsUpdateCmd = &cobra.Command{
		Use:   "update BACKUP",
		Short: "Update a Filestore backup",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSBackupsUpdate,
	}
	s.filestoreInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Filestore instances"}
	s.filestoreInstancesCreateCmd = &cobra.Command{
		Use:   "create INSTANCE",
		Short: "Create a Filestore instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSInstancesCreate,
	}
	s.filestoreInstancesDeleteCmd = &cobra.Command{
		Use:   "delete INSTANCE",
		Short: "Delete a Filestore instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSInstancesDelete,
	}
	s.filestoreInstancesDescribeCmd = &cobra.Command{
		Use:   "describe INSTANCE",
		Short: "Describe a Filestore instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSInstancesDescribe,
	}
	s.filestoreInstancesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Filestore instances",
		Args:  cobra.NoArgs,
		RunE:  s.runFSInstancesList,
	}
	s.filestoreInstancesUpdateCmd = &cobra.Command{
		Use:   "update INSTANCE",
		Short: "Update a Filestore instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSInstancesUpdate,
	}
	s.filestoreInstancesRestoreCmd = &cobra.Command{
		Use:   "restore INSTANCE",
		Short: "Restore a Filestore instance from a backup",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSInstancesRestore,
	}
	s.filestoreInstancesRevertCmd = &cobra.Command{
		Use:   "revert INSTANCE",
		Short: "Revert a Filestore instance to a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSInstancesRevert,
	}
	s.filestoreLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Filestore locations"}
	s.filestoreLocationsDescribeCmd = &cobra.Command{
		Use:   "describe LOCATION",
		Short: "Describe a Filestore location",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSLocationsDescribe,
	}
	s.filestoreLocationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Filestore locations",
		Args:  cobra.NoArgs,
		RunE:  s.runFSLocationsList,
	}
	s.filestoreOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Filestore operations"}
	s.filestoreOperationsCancelCmd = &cobra.Command{
		Use:   "cancel OPERATION",
		Short: "Cancel a Filestore operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSOperationsCancel,
	}
	s.filestoreOperationsDeleteCmd = &cobra.Command{
		Use:   "delete OPERATION",
		Short: "Delete a Filestore operation record",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSOperationsDelete,
	}
	s.filestoreOperationsDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe a Filestore operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSOperationsDescribe,
	}
	s.filestoreOperationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Filestore operations",
		Args:  cobra.NoArgs,
		RunE:  s.runFSOperationsList,
	}
	s.filestoreOperationsWaitCmd = &cobra.Command{
		Use:   "wait OPERATION",
		Short: "Poll a Filestore operation until it completes",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runFSOperationsWait,
	}
	s.filestoreRegionsCmd = &cobra.Command{Use: "regions", Short: "List Filestore regions"}
	s.filestoreRegionsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Filestore regions",
		Args:  cobra.NoArgs,
		RunE:  s.runFSRegionsList,
	}
	s.filestoreZonesCmd = &cobra.Command{Use: "zones", Short: "List Filestore zones"}
	s.filestoreZonesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Filestore zones",
		Args:  cobra.NoArgs,
		RunE:  s.runFSZonesList,
	}
	s.initFilestore()
	return s
}
