package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	memcache "google.golang.org/api/memcache/v1"
)

// --- gcloud memcache (#354) ---

var memcacheCmd = &cobra.Command{Use: "memcache", Short: "Manage Memorystore for Memcached"}

func mcLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func mcChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// --- regions ---

var memcacheRegionsCmd = &cobra.Command{Use: "regions", Short: "Explore Memorystore for Memcached regions"}

var (
	mcRegionDescribeCmd = &cobra.Command{
		Use: "describe REGION", Short: "Describe a region",
		Args: cobra.ExactArgs(1), RunE: runMCRegionDescribe,
	}
	mcRegionListCmd = &cobra.Command{
		Use: "list", Short: "List regions",
		Args: cobra.NoArgs, RunE: runMCRegionList,
	}
)

var flagMCFormat string

// --- instances ---

var memcacheInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Memcached instances"}

var (
	mcInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMCInstCreate,
	}
	mcInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an instance",
		Args: cobra.ExactArgs(1), RunE: runMCInstDelete,
	}
	mcInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an instance",
		Args: cobra.ExactArgs(1), RunE: runMCInstDescribe,
	}
	mcInstListCmd = &cobra.Command{
		Use: "list", Short: "List instances in a region",
		Args: cobra.NoArgs, RunE: runMCInstList,
	}
	mcInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runMCInstUpdate,
	}
	mcInstUpgradeCmd = &cobra.Command{
		Use: "upgrade INSTANCE", Short: "Upgrade an instance (--config-file with UpgradeInstanceRequest)",
		Args: cobra.ExactArgs(1), RunE: runMCInstUpgrade,
	}
	mcInstApplyParamsCmd = &cobra.Command{
		Use: "apply-parameters INSTANCE", Short: "Apply pending parameter updates (--node-ids, --apply-all)",
		Args: cobra.ExactArgs(1), RunE: runMCInstApplyParams,
	}
)

var (
	flagMCRegion      string
	flagMCConfigFile  string
	flagMCUpdateMask  string
	flagMCAsync       bool
	flagMCNodeIDs     []string
	flagMCApplyAll    bool
)

// --- operations ---

var memcacheOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Memcache operations"}

var (
	mcOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: runMCOpCancel,
	}
	mcOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runMCOpDescribe,
	}
	mcOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a region",
		Args: cobra.NoArgs, RunE: runMCOpList,
	}
)

var flagMCOpRegion string

func init() {
	// regions
	mcRegionDescribeCmd.Flags().StringVar(&flagMCFormat, "format", "", "Output format")
	mcRegionListCmd.Flags().StringVar(&flagMCFormat, "format", "", "Output format")
	memcacheRegionsCmd.AddCommand(mcRegionDescribeCmd, mcRegionListCmd)
	memcacheCmd.AddCommand(memcacheRegionsCmd)

	// instances
	instAll := []*cobra.Command{mcInstCreateCmd, mcInstDeleteCmd, mcInstDescribeCmd, mcInstListCmd,
		mcInstUpdateCmd, mcInstUpgradeCmd, mcInstApplyParamsCmd}
	for _, c := range instAll {
		c.Flags().StringVar(&flagMCRegion, "region", "", "Region containing the instance (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{mcInstCreateCmd, mcInstUpdateCmd, mcInstUpgradeCmd} {
		c.Flags().StringVar(&flagMCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Instance / UpgradeInstanceRequest body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	mcInstUpdateCmd.Flags().StringVar(&flagMCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{mcInstCreateCmd, mcInstDeleteCmd, mcInstUpdateCmd, mcInstUpgradeCmd, mcInstApplyParamsCmd} {
		c.Flags().BoolVar(&flagMCAsync, "async", false, "Return the long-running operation without waiting")
	}
	mcInstDescribeCmd.Flags().StringVar(&flagMCFormat, "format", "", "Output format")
	mcInstListCmd.Flags().StringVar(&flagMCFormat, "format", "", "Output format")
	mcInstApplyParamsCmd.Flags().StringSliceVar(&flagMCNodeIDs, "node-ids", nil,
		"Comma-separated list of node IDs to apply the pending parameters to")
	mcInstApplyParamsCmd.Flags().BoolVar(&flagMCApplyAll, "apply-all", false, "Apply to all nodes")
	memcacheInstancesCmd.AddCommand(instAll...)
	memcacheCmd.AddCommand(memcacheInstancesCmd)

	// operations
	for _, c := range []*cobra.Command{mcOpCancelCmd, mcOpDescribeCmd, mcOpListCmd} {
		c.Flags().StringVar(&flagMCOpRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	mcOpDescribeCmd.Flags().StringVar(&flagMCFormat, "format", "", "Output format")
	mcOpListCmd.Flags().StringVar(&flagMCFormat, "format", "", "Output format")
	memcacheOperationsCmd.AddCommand(mcOpCancelCmd, mcOpDescribeCmd, mcOpListCmd)
	memcacheCmd.AddCommand(memcacheOperationsCmd)

	rootCmd.AddCommand(memcacheCmd)
}

// helper to poll & finish LROs.
func mcFinishOp(ctx context.Context, svc *memcache.Service, op *memcache.Operation, verb, name string, async bool) error {
	if async {
		fmt.Printf("%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	fmt.Printf("%s [%s] completed.\n", verb, name)
	if op.Response != nil {
		return emitFormatted(op.Response, "")
	}
	return nil
}

// --- regions impl ---

func runMCRegionDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(mcLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing region: %w", err)
	}
	return emitFormatted(loc, flagMCFormat)
}

func runMCRegionList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing regions: %w", err)
	}
	if flagMCFormat != "" {
		return emitFormatted(resp.Locations, flagMCFormat)
	}
	fmt.Printf("%-20s %s\n", "REGION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- instances impl ---

func mcInstName(id, project, region string) string {
	return mcChild("instances", id, mcLocationParent(project, region))
}

func runMCInstCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &memcache.Instance{}
	if err := loadYAMLOrJSONInto(flagMCConfigFile, inst); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(mcLocationParent(project, flagMCRegion), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return mcFinishOp(ctx, svc, op, "Create instance", args[0], flagMCAsync)
}

func runMCInstDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(mcInstName(args[0], project, flagMCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return mcFinishOp(ctx, svc, op, "Delete instance", args[0], flagMCAsync)
}

func runMCInstDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Get(mcInstName(args[0], project, flagMCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagMCFormat)
}

func runMCInstList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.List(mcLocationParent(project, flagMCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if flagMCFormat != "" {
		return emitFormatted(resp.Instances, flagMCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, i := range resp.Instances {
		fmt.Printf("%-40s %s\n", path.Base(i.Name), i.State)
	}
	return nil
}

func runMCInstUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &memcache.Instance{}
	if err := loadYAMLOrJSONInto(flagMCConfigFile, inst); err != nil {
		return err
	}
	mask := flagMCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Patch(mcInstName(args[0], project, flagMCRegion), inst).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return mcFinishOp(ctx, svc, op, "Update instance", args[0], flagMCAsync)
}

func runMCInstUpgrade(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &memcache.GoogleCloudMemcacheV1UpgradeInstanceRequest{}
	if err := loadYAMLOrJSONInto(flagMCConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Upgrade(mcInstName(args[0], project, flagMCRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upgrading instance: %w", err)
	}
	return mcFinishOp(ctx, svc, op, "Upgrade instance", args[0], flagMCAsync)
}

func runMCInstApplyParams(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &memcache.ApplyParametersRequest{NodeIds: flagMCNodeIDs, ApplyAll: flagMCApplyAll}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.ApplyParameters(mcInstName(args[0], project, flagMCRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("applying parameters: %w", err)
	}
	return mcFinishOp(ctx, svc, op, "Apply parameters", args[0], flagMCAsync)
}

// --- operations impl ---

func mcOpName(id, project, region string) string {
	return mcChild("operations", id, mcLocationParent(project, region))
}

func runMCOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(mcOpName(args[0], project, flagMCOpRegion), &memcache.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runMCOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(mcOpName(args[0], project, flagMCOpRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagMCFormat)
}

func runMCOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MemcacheService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(mcLocationParent(project, flagMCOpRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagMCFormat != "" {
		return emitFormatted(resp.Operations, flagMCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
