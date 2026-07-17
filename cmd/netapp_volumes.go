package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp volumes (#1205) ---

var netappVolCmd = &cobra.Command{Use: "volumes", Short: "Manage NetApp volumes"}

var (
	flagNetAppVolLocation   string
	flagNetAppVolConfigFile string
	flagNetAppVolUpdateMask string
	flagNetAppVolFormat     string
	flagNetAppVolFilter     string
	flagNetAppVolPageSize   int64
	flagNetAppVolForce      bool
	flagNetAppVolSnapshotID string
)

var (
	netappVolCreateCmd = &cobra.Command{
		Use: "create VOLUME", Short: "Create a volume",
		Args: cobra.ExactArgs(1), RunE: runNetAppVolCreate,
	}
	netappVolDeleteCmd = &cobra.Command{
		Use: "delete VOLUME", Short: "Delete a volume",
		Args: cobra.ExactArgs(1), RunE: runNetAppVolDelete,
	}
	netappVolDescribeCmd = &cobra.Command{
		Use: "describe VOLUME", Short: "Describe a volume",
		Args: cobra.ExactArgs(1), RunE: runNetAppVolDescribe,
	}
	netappVolListCmd = &cobra.Command{
		Use: "list", Short: "List volumes",
		Args: cobra.NoArgs, RunE: runNetAppVolList,
	}
	netappVolUpdateCmd = &cobra.Command{
		Use: "update VOLUME", Short: "Update a volume",
		Args: cobra.ExactArgs(1), RunE: runNetAppVolUpdate,
	}
	netappVolRevertCmd = &cobra.Command{
		Use: "revert VOLUME", Short: "Revert a volume to a snapshot",
		Args: cobra.ExactArgs(1), RunE: runNetAppVolRevert,
	}
)

// --- replications subgroup ---

var netappReplCmd = &cobra.Command{Use: "replications", Short: "Manage NetApp volume replications"}

var (
	flagNetAppReplVolume     string
	flagNetAppReplConfigFile string
	flagNetAppReplUpdateMask string
	flagNetAppReplFilter     string
	flagNetAppReplPageSize   int64
)

var (
	netappReplCreateCmd = &cobra.Command{
		Use: "create REPLICATION", Short: "Create a replication",
		Args: cobra.ExactArgs(1), RunE: runNetAppReplCreate,
	}
	netappReplDeleteCmd = &cobra.Command{
		Use: "delete REPLICATION", Short: "Delete a replication",
		Args: cobra.ExactArgs(1), RunE: runNetAppReplDelete,
	}
	netappReplDescribeCmd = &cobra.Command{
		Use: "describe REPLICATION", Short: "Describe a replication",
		Args: cobra.ExactArgs(1), RunE: runNetAppReplDescribe,
	}
	netappReplListCmd = &cobra.Command{
		Use: "list", Short: "List replications for a volume",
		Args: cobra.NoArgs, RunE: runNetAppReplList,
	}
	netappReplUpdateCmd = &cobra.Command{
		Use: "update REPLICATION", Short: "Update a replication",
		Args: cobra.ExactArgs(1), RunE: runNetAppReplUpdate,
	}
	netappReplResumeCmd = &cobra.Command{
		Use: "resume REPLICATION", Short: "Resume a replication",
		Args: cobra.ExactArgs(1), RunE: runNetAppReplResume,
	}
	netappReplReverseCmd = &cobra.Command{
		Use: "reverse-direction REPLICATION", Short: "Reverse a replication's direction",
		Args: cobra.ExactArgs(1), RunE: runNetAppReplReverse,
	}
	netappReplStopCmd = &cobra.Command{
		Use: "stop REPLICATION", Short: "Stop a replication",
		Args: cobra.ExactArgs(1), RunE: runNetAppReplStop,
	}
	netappReplSyncCmd = &cobra.Command{
		Use: "sync REPLICATION", Short: "Sync a replication",
		Args: cobra.ExactArgs(1), RunE: runNetAppReplSync,
	}
)

// --- snapshots subgroup ---

var netappSnapCmd = &cobra.Command{Use: "snapshots", Short: "Manage NetApp volume snapshots"}

var (
	flagNetAppSnapVolume     string
	flagNetAppSnapConfigFile string
	flagNetAppSnapUpdateMask string
	flagNetAppSnapFilter     string
	flagNetAppSnapPageSize   int64
)

var (
	netappSnapCreateCmd = &cobra.Command{
		Use: "create SNAPSHOT", Short: "Create a snapshot",
		Args: cobra.ExactArgs(1), RunE: runNetAppSnapCreate,
	}
	netappSnapDeleteCmd = &cobra.Command{
		Use: "delete SNAPSHOT", Short: "Delete a snapshot",
		Args: cobra.ExactArgs(1), RunE: runNetAppSnapDelete,
	}
	netappSnapDescribeCmd = &cobra.Command{
		Use: "describe SNAPSHOT", Short: "Describe a snapshot",
		Args: cobra.ExactArgs(1), RunE: runNetAppSnapDescribe,
	}
	netappSnapListCmd = &cobra.Command{
		Use: "list", Short: "List snapshots for a volume",
		Args: cobra.NoArgs, RunE: runNetAppSnapList,
	}
	netappSnapUpdateCmd = &cobra.Command{
		Use: "update SNAPSHOT", Short: "Update a snapshot",
		Args: cobra.ExactArgs(1), RunE: runNetAppSnapUpdate,
	}
)

func init() {
	// volumes top-level flags
	volAll := []*cobra.Command{
		netappVolCreateCmd, netappVolDeleteCmd, netappVolDescribeCmd,
		netappVolListCmd, netappVolUpdateCmd, netappVolRevertCmd,
	}
	for _, c := range volAll {
		c.Flags().StringVar(&flagNetAppVolLocation, "location", "", "Location for the volume (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppVolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappVolCreateCmd, netappVolUpdateCmd} {
		c.Flags().StringVar(&flagNetAppVolConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Volume body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappVolUpdateCmd.Flags().StringVar(&flagNetAppVolUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappVolListCmd.Flags().StringVar(&flagNetAppVolFilter, "filter", "", "Server-side filter expression")
	netappVolListCmd.Flags().Int64Var(&flagNetAppVolPageSize, "page-size", 0, "Maximum number of results per page")
	netappVolDeleteCmd.Flags().BoolVar(&flagNetAppVolForce, "force", false, "Force delete non-empty volume")
	netappVolRevertCmd.Flags().StringVar(&flagNetAppVolSnapshotID, "snapshot", "",
		"Snapshot resource ID to revert to (required)")
	_ = netappVolRevertCmd.MarkFlagRequired("snapshot")

	netappVolCmd.AddCommand(volAll...)

	// replications
	replAll := []*cobra.Command{
		netappReplCreateCmd, netappReplDeleteCmd, netappReplDescribeCmd,
		netappReplListCmd, netappReplUpdateCmd, netappReplResumeCmd,
		netappReplReverseCmd, netappReplStopCmd, netappReplSyncCmd,
	}
	for _, c := range replAll {
		c.Flags().StringVar(&flagNetAppVolLocation, "location", "", "Location for the replication (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppReplVolume, "volume", "", "Volume that owns the replication (required)")
		_ = c.MarkFlagRequired("volume")
		c.Flags().StringVar(&flagNetAppVolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappReplCreateCmd, netappReplUpdateCmd} {
		c.Flags().StringVar(&flagNetAppReplConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Replication body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappReplUpdateCmd.Flags().StringVar(&flagNetAppReplUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappReplListCmd.Flags().StringVar(&flagNetAppReplFilter, "filter", "", "Server-side filter expression")
	netappReplListCmd.Flags().Int64Var(&flagNetAppReplPageSize, "page-size", 0, "Maximum number of results per page")

	netappReplCmd.AddCommand(replAll...)
	netappVolCmd.AddCommand(netappReplCmd)

	// snapshots
	snapAll := []*cobra.Command{
		netappSnapCreateCmd, netappSnapDeleteCmd, netappSnapDescribeCmd,
		netappSnapListCmd, netappSnapUpdateCmd,
	}
	for _, c := range snapAll {
		c.Flags().StringVar(&flagNetAppVolLocation, "location", "", "Location for the snapshot (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppSnapVolume, "volume", "", "Volume that owns the snapshot (required)")
		_ = c.MarkFlagRequired("volume")
		c.Flags().StringVar(&flagNetAppVolFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappSnapCreateCmd, netappSnapUpdateCmd} {
		c.Flags().StringVar(&flagNetAppSnapConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Snapshot body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappSnapUpdateCmd.Flags().StringVar(&flagNetAppSnapUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappSnapListCmd.Flags().StringVar(&flagNetAppSnapFilter, "filter", "", "Server-side filter expression")
	netappSnapListCmd.Flags().Int64Var(&flagNetAppSnapPageSize, "page-size", 0, "Maximum number of results per page")

	netappSnapCmd.AddCommand(snapAll...)
	netappVolCmd.AddCommand(netappSnapCmd)

	netappCmd.AddCommand(netappVolCmd)
}

func netappVolParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, flagNetAppVolLocation), nil
}

func netappVolName(id string) (string, error) {
	parent, err := netappVolParent()
	if err != nil {
		return "", err
	}
	return netappChild("volumes", id, parent), nil
}

func runNetAppVolCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappVolParent()
	if err != nil {
		return err
	}
	body := &netapp.Volume{}
	if err := loadYAMLOrJSONInto(flagNetAppVolConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Create(parent, body).VolumeId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating volume: %w", err)
	}
	fmt.Printf("Create request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppVolDelete(cmd *cobra.Command, args []string) error {
	name, err := netappVolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Volumes.Delete(name).Context(ctx)
	if flagNetAppVolForce {
		call = call.Force(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting volume: %w", err)
	}
	fmt.Printf("Delete request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppVolDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappVolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing volume: %w", err)
	}
	return emitFormatted(got, flagNetAppVolFormat)
}

func runNetAppVolList(cmd *cobra.Command, args []string) error {
	parent, err := netappVolParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.Volume
	pageToken := ""
	for {
		call := svc.Projects.Locations.Volumes.List(parent).Context(ctx)
		if flagNetAppVolFilter != "" {
			call = call.Filter(flagNetAppVolFilter)
		}
		if flagNetAppVolPageSize > 0 {
			call = call.PageSize(flagNetAppVolPageSize)
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
	return emitFormatted(all, flagNetAppVolFormat)
}

func runNetAppVolUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappVolName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.Volume{}
	if err := loadYAMLOrJSONInto(flagNetAppVolConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppVolUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
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
	fmt.Printf("Update request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppVolRevert(cmd *cobra.Command, args []string) error {
	name, err := netappVolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Revert(name, &netapp.RevertVolumeRequest{
		SnapshotId: flagNetAppVolSnapshotID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reverting volume: %w", err)
	}
	fmt.Printf("Revert request issued for volume [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func netappReplParent() (string, error) {
	vol, err := netappVolName(flagNetAppReplVolume)
	if err != nil {
		return "", err
	}
	return vol, nil
}

func netappReplName(id string) (string, error) {
	parent, err := netappReplParent()
	if err != nil {
		return "", err
	}
	return netappChild("replications", id, parent), nil
}

func runNetAppReplCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappReplParent()
	if err != nil {
		return err
	}
	body := &netapp.Replication{}
	if err := loadYAMLOrJSONInto(flagNetAppReplConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Create(parent, body).ReplicationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating replication: %w", err)
	}
	fmt.Printf("Create request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppReplDelete(cmd *cobra.Command, args []string) error {
	name, err := netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting replication: %w", err)
	}
	fmt.Printf("Delete request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppReplDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Replications.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing replication: %w", err)
	}
	return emitFormatted(got, flagNetAppVolFormat)
}

func runNetAppReplList(cmd *cobra.Command, args []string) error {
	parent, err := netappReplParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.Replication
	pageToken := ""
	for {
		call := svc.Projects.Locations.Volumes.Replications.List(parent).Context(ctx)
		if flagNetAppReplFilter != "" {
			call = call.Filter(flagNetAppReplFilter)
		}
		if flagNetAppReplPageSize > 0 {
			call = call.PageSize(flagNetAppReplPageSize)
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
	return emitFormatted(all, flagNetAppVolFormat)
}

func runNetAppReplUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappReplName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.Replication{}
	if err := loadYAMLOrJSONInto(flagNetAppReplConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppReplUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
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
	fmt.Printf("Update request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppReplResume(cmd *cobra.Command, args []string) error {
	name, err := netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Resume(name, &netapp.ResumeReplicationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resuming replication: %w", err)
	}
	fmt.Printf("Resume request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppReplReverse(cmd *cobra.Command, args []string) error {
	name, err := netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.ReverseDirection(name, &netapp.ReverseReplicationDirectionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reversing replication direction: %w", err)
	}
	fmt.Printf("Reverse-direction request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppReplStop(cmd *cobra.Command, args []string) error {
	name, err := netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Stop(name, &netapp.StopReplicationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping replication: %w", err)
	}
	fmt.Printf("Stop request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppReplSync(cmd *cobra.Command, args []string) error {
	name, err := netappReplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Replications.Sync(name, &netapp.SyncReplicationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("syncing replication: %w", err)
	}
	fmt.Printf("Sync request issued for replication [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func netappSnapParent() (string, error) {
	vol, err := netappVolName(flagNetAppSnapVolume)
	if err != nil {
		return "", err
	}
	return vol, nil
}

func netappSnapName(id string) (string, error) {
	parent, err := netappSnapParent()
	if err != nil {
		return "", err
	}
	return netappChild("snapshots", id, parent), nil
}

func runNetAppSnapCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappSnapParent()
	if err != nil {
		return err
	}
	body := &netapp.Snapshot{}
	if err := loadYAMLOrJSONInto(flagNetAppSnapConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Snapshots.Create(parent, body).SnapshotId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating snapshot: %w", err)
	}
	fmt.Printf("Create request issued for snapshot [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppSnapDelete(cmd *cobra.Command, args []string) error {
	name, err := netappSnapName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Snapshots.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting snapshot: %w", err)
	}
	fmt.Printf("Delete request issued for snapshot [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}

func runNetAppSnapDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappSnapName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Snapshots.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing snapshot: %w", err)
	}
	return emitFormatted(got, flagNetAppVolFormat)
}

func runNetAppSnapList(cmd *cobra.Command, args []string) error {
	parent, err := netappSnapParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.Snapshot
	pageToken := ""
	for {
		call := svc.Projects.Locations.Volumes.Snapshots.List(parent).Context(ctx)
		if flagNetAppSnapFilter != "" {
			call = call.Filter(flagNetAppSnapFilter)
		}
		if flagNetAppSnapPageSize > 0 {
			call = call.PageSize(flagNetAppSnapPageSize)
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
	return emitFormatted(all, flagNetAppVolFormat)
}

func runNetAppSnapUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappSnapName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.Snapshot{}
	if err := loadYAMLOrJSONInto(flagNetAppSnapConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppSnapUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
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
	fmt.Printf("Update request issued for snapshot [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppVolFormat)
}
