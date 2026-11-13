package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp volumes (#1205) ---

// --- replications subgroup ---

// --- snapshots subgroup ---

func (s *netappCmds) initNetappVolumes() {
	// volumes top-level flags
	volAll := []*cobra.Command{
		s.netappVolCreateCmd, s.netappVolDeleteCmd, s.netappVolDescribeCmd,
		s.netappVolListCmd, s.netappVolUpdateCmd, s.netappVolRevertCmd,
		s.netappVolStartSplitCmd, s.netappVolGetSplitStatusCmd,
	}
	for _, c := range volAll {
		c.Flags().StringVar(&s.flagNetAppVolLocation, "location", "", "Location for the volume (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppVolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappVolCreateCmd, s.netappVolUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppVolConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Volume body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappVolUpdateCmd.Flags().StringVar(&s.flagNetAppVolUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappVolListCmd.Flags().StringVar(&s.flagNetAppVolFilter, "filter", "", "Server-side filter expression")
	s.netappVolListCmd.Flags().Int64Var(&s.flagNetAppVolPageSize, "page-size", 0, "Maximum number of results per page")
	s.netappVolDeleteCmd.Flags().BoolVar(&s.flagNetAppVolForce, "force", false, "Force delete non-empty volume")
	s.netappVolRevertCmd.Flags().StringVar(&s.flagNetAppVolSnapshotID, "snapshot", "",
		"Snapshot resource ID to revert to (required)")
	_ = s.netappVolRevertCmd.MarkFlagRequired("snapshot")

	s.netappVolCmd.AddCommand(volAll...)

	// replications
	replAll := []*cobra.Command{
		s.netappReplCreateCmd, s.netappReplDeleteCmd, s.netappReplDescribeCmd,
		s.netappReplListCmd, s.netappReplUpdateCmd, s.netappReplResumeCmd,
		s.netappReplReverseCmd, s.netappReplStopCmd, s.netappReplSyncCmd,
	}
	for _, c := range replAll {
		c.Flags().StringVar(&s.flagNetAppVolLocation, "location", "", "Location for the replication (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppReplVolume, "volume", "", "Volume that owns the replication (required)")
		_ = c.MarkFlagRequired("volume")
		c.Flags().StringVar(&s.flagNetAppVolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappReplCreateCmd, s.netappReplUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppReplConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Replication body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappReplUpdateCmd.Flags().StringVar(&s.flagNetAppReplUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappReplListCmd.Flags().StringVar(&s.flagNetAppReplFilter, "filter", "", "Server-side filter expression")
	s.netappReplListCmd.Flags().Int64Var(&s.flagNetAppReplPageSize, "page-size", 0, "Maximum number of results per page")

	s.netappReplCmd.AddCommand(replAll...)
	s.netappVolCmd.AddCommand(s.netappReplCmd)

	// snapshots
	snapAll := []*cobra.Command{
		s.netappSnapCreateCmd, s.netappSnapDeleteCmd, s.netappSnapDescribeCmd,
		s.netappSnapListCmd, s.netappSnapUpdateCmd,
	}
	for _, c := range snapAll {
		c.Flags().StringVar(&s.flagNetAppVolLocation, "location", "", "Location for the snapshot (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppSnapVolume, "volume", "", "Volume that owns the snapshot (required)")
		_ = c.MarkFlagRequired("volume")
		c.Flags().StringVar(&s.flagNetAppVolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappSnapCreateCmd, s.netappSnapUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppSnapConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Snapshot body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappSnapUpdateCmd.Flags().StringVar(&s.flagNetAppSnapUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappSnapListCmd.Flags().StringVar(&s.flagNetAppSnapFilter, "filter", "", "Server-side filter expression")
	s.netappSnapListCmd.Flags().Int64Var(&s.flagNetAppSnapPageSize, "page-size", 0, "Maximum number of results per page")

	s.netappSnapCmd.AddCommand(snapAll...)
	s.netappVolCmd.AddCommand(s.netappSnapCmd)

	s.netappCmd.AddCommand(s.netappVolCmd)
}

func (s *netappCmds) netappVolParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, s.flagNetAppVolLocation), nil
}

func (s *netappCmds) netappVolName(id string) (string, error) {
	parent, err := s.netappVolParent()
	if err != nil {
		return "", err
	}
	return netappChild("volumes", id, parent), nil
}

func (s *netappCmds) runNetAppVolCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappVolParent()
	if err != nil {
		return err
	}
	body := &netapp.Volume{}
	if err := loadYAMLOrJSONInto(s.flagNetAppVolConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Create(parent, body).VolumeId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating volume: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppVolDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappVolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Volumes.Delete(name).Context(ctx)
	if s.flagNetAppVolForce {
		call = call.Force(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting volume: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppVolDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappVolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing volume: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppVolList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappVolParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.Volume
	pageToken := ""
	for {
		call := svc.Projects.Locations.Volumes.List(parent).Context(ctx)
		if s.flagNetAppVolFilter != "" {
			call = call.Filter(s.flagNetAppVolFilter)
		}
		if s.flagNetAppVolPageSize > 0 {
			call = call.PageSize(s.flagNetAppVolPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing volumes: %w", err)
		}
		all = append(all, resp.Volumes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppVolUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappVolName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.Volume{}
	if err := loadYAMLOrJSONInto(s.flagNetAppVolConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppVolUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Volumes.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating volume: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppVolStartSplit(cmd *cobra.Command, args []string) error {
	name, err := s.netappVolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.StartSplit(name, &netapp.StartSplitRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting volume split: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start-split request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppVolGetSplitStatus(cmd *cobra.Command, args []string) error {
	name, err := s.netappVolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.GetSplitStatus(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting split status: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppVolRevert(cmd *cobra.Command, args []string) error {
	name, err := s.netappVolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Revert(name, &netapp.RevertVolumeRequest{
		SnapshotId: s.flagNetAppVolSnapshotID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reverting volume: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Revert request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) netappReplParent() (string, error) {
	vol, err := s.netappVolName(s.flagNetAppReplVolume)
	if err != nil {
		return "", err
	}
	return vol, nil
}

func (s *netappCmds) netappReplName(id string) (string, error) {
	parent, err := s.netappReplParent()
	if err != nil {
		return "", err
	}
	return netappChild("replications", id, parent), nil
}

func (s *netappCmds) runNetAppReplCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappReplParent()
	if err != nil {
		return err
	}
	body := &netapp.Replication{}
	if err := loadYAMLOrJSONInto(s.flagNetAppReplConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Create(parent, body).ReplicationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating replication: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppReplDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting replication: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppReplDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Replications.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing replication: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppReplList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappReplParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.Replication
	pageToken := ""
	for {
		call := svc.Projects.Locations.Volumes.Replications.List(parent).Context(ctx)
		if s.flagNetAppReplFilter != "" {
			call = call.Filter(s.flagNetAppReplFilter)
		}
		if s.flagNetAppReplPageSize > 0 {
			call = call.PageSize(s.flagNetAppReplPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing replications: %w", err)
		}
		all = append(all, resp.Replications...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppReplUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappReplName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.Replication{}
	if err := loadYAMLOrJSONInto(s.flagNetAppReplConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppReplUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Volumes.Replications.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating replication: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppReplResume(cmd *cobra.Command, args []string) error {
	name, err := s.netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Resume(name, &netapp.ResumeReplicationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming replication: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Resume request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppReplReverse(cmd *cobra.Command, args []string) error {
	name, err := s.netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.ReverseDirection(name, &netapp.ReverseReplicationDirectionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reversing replication direction: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Reverse-direction request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppReplStop(cmd *cobra.Command, args []string) error {
	name, err := s.netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Stop(name, &netapp.StopReplicationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping replication: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Stop request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppReplSync(cmd *cobra.Command, args []string) error {
	name, err := s.netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Sync(name, &netapp.SyncReplicationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("syncing replication: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Sync request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) netappSnapParent() (string, error) {
	vol, err := s.netappVolName(s.flagNetAppSnapVolume)
	if err != nil {
		return "", err
	}
	return vol, nil
}

func (s *netappCmds) netappSnapName(id string) (string, error) {
	parent, err := s.netappSnapParent()
	if err != nil {
		return "", err
	}
	return netappChild("snapshots", id, parent), nil
}

func (s *netappCmds) runNetAppSnapCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappSnapParent()
	if err != nil {
		return err
	}
	body := &netapp.Snapshot{}
	if err := loadYAMLOrJSONInto(s.flagNetAppSnapConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Snapshots.Create(parent, body).SnapshotId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating snapshot: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for snapshot [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppSnapDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappSnapName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Snapshots.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting snapshot: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for snapshot [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppSnapDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappSnapName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Snapshots.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing snapshot: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppSnapList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappSnapParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.Snapshot
	pageToken := ""
	for {
		call := svc.Projects.Locations.Volumes.Snapshots.List(parent).Context(ctx)
		if s.flagNetAppSnapFilter != "" {
			call = call.Filter(s.flagNetAppSnapFilter)
		}
		if s.flagNetAppSnapPageSize > 0 {
			call = call.PageSize(s.flagNetAppSnapPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing snapshots: %w", err)
		}
		all = append(all, resp.Snapshots...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppVolFormat)
}

func (s *netappCmds) runNetAppSnapUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappSnapName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.Snapshot{}
	if err := loadYAMLOrJSONInto(s.flagNetAppSnapConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppSnapUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Volumes.Snapshots.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating snapshot: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for snapshot [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppVolFormat)
}
