package cmd

import (
	"context"
	"fmt"

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
)

func init() {
	all := []*cobra.Command{
		netappSPCreateCmd, netappSPDeleteCmd, netappSPDescribeCmd,
		netappSPListCmd, netappSPUpdateCmd, netappSPSwitchCmd,
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
