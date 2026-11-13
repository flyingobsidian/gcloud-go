package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	memcache "google.golang.org/api/memcache/v1"
)

// --- gcloud memcache (#354) ---

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

// --- instances ---

// --- operations ---

func (s *memcacheCmds) initMemcache() {
	// regions
	s.mcRegionDescribeCmd.Flags().StringVar(&s.flagMCFormat, "format", "", "Output format")
	s.mcRegionListCmd.Flags().StringVar(&s.flagMCFormat, "format", "", "Output format")
	s.memcacheRegionsCmd.AddCommand(s.mcRegionDescribeCmd, s.mcRegionListCmd)
	s.memcacheCmd.AddCommand(s.memcacheRegionsCmd)

	// instances
	instAll := []*cobra.Command{s.mcInstCreateCmd, s.mcInstDeleteCmd, s.mcInstDescribeCmd, s.mcInstListCmd,
		s.mcInstUpdateCmd, s.mcInstUpgradeCmd, s.mcInstApplyParamsCmd}
	for _, c := range instAll {
		c.Flags().StringVar(&s.flagMCRegion, "region", "", "Region containing the instance (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.mcInstCreateCmd, s.mcInstUpdateCmd, s.mcInstUpgradeCmd} {
		c.Flags().StringVar(&s.flagMCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Instance / UpgradeInstanceRequest body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.mcInstUpdateCmd.Flags().StringVar(&s.flagMCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.mcInstCreateCmd, s.mcInstDeleteCmd, s.mcInstUpdateCmd, s.mcInstUpgradeCmd, s.mcInstApplyParamsCmd} {
		c.Flags().BoolVar(&s.flagMCAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.mcInstDescribeCmd.Flags().StringVar(&s.flagMCFormat, "format", "", "Output format")
	s.mcInstListCmd.Flags().StringVar(&s.flagMCFormat, "format", "", "Output format")
	s.mcInstApplyParamsCmd.Flags().StringSliceVar(&s.flagMCNodeIDs, "node-ids", nil,
		"Comma-separated list of node IDs to apply the pending parameters to")
	s.mcInstApplyParamsCmd.Flags().BoolVar(&s.flagMCApplyAll, "apply-all", false, "Apply to all nodes")
	s.memcacheInstancesCmd.AddCommand(instAll...)
	s.memcacheCmd.AddCommand(s.memcacheInstancesCmd)

	// operations
	for _, c := range []*cobra.Command{s.mcOpCancelCmd, s.mcOpDescribeCmd, s.mcOpListCmd} {
		c.Flags().StringVar(&s.flagMCOpRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	s.mcOpDescribeCmd.Flags().StringVar(&s.flagMCFormat, "format", "", "Output format")
	s.mcOpListCmd.Flags().StringVar(&s.flagMCFormat, "format", "", "Output format")
	s.memcacheOperationsCmd.AddCommand(s.mcOpCancelCmd, s.mcOpDescribeCmd, s.mcOpListCmd)
	s.memcacheCmd.AddCommand(s.memcacheOperationsCmd)

}

// helper to poll & finish LROs.
func (s *memcacheCmds) mcFinishOp(ctx context.Context, svc *memcache.Service, op *memcache.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stdout, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
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
	fmt.Fprintf(s.inv.stdout, "%s [%s] completed.\n", verb, name)
	if op.Response != nil {
		return emitFormattedTo(s.inv.stdout, op.Response, "")
	}
	return nil
}

// --- regions impl ---

func (s *memcacheCmds) runMCRegionDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(mcLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing region: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, loc, s.flagMCFormat)
}

func (s *memcacheCmds) runMCRegionList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing regions: %w", err)
	}
	if s.flagMCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Locations, s.flagMCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Locations), "%-20s %s\n", "REGION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(s.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- instances impl ---

func mcInstName(id, project, region string) string {
	return mcChild("instances", id, mcLocationParent(project, region))
}

func (s *memcacheCmds) runMCInstCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	inst := &memcache.Instance{}
	if err := loadYAMLOrJSONInto(s.flagMCConfigFile, inst); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(mcLocationParent(project, s.flagMCRegion), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return s.mcFinishOp(ctx, svc, op, "Create instance", args[0], s.flagMCAsync)
}

func (s *memcacheCmds) runMCInstDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(mcInstName(args[0], project, s.flagMCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return s.mcFinishOp(ctx, svc, op, "Delete instance", args[0], s.flagMCAsync)
}

func (s *memcacheCmds) runMCInstDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Get(mcInstName(args[0], project, s.flagMCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMCFormat)
}

func (s *memcacheCmds) runMCInstList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.List(mcLocationParent(project, s.flagMCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if s.flagMCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Instances, s.flagMCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Instances), "%-40s %s\n", "NAME", "STATE")
	for _, i := range resp.Instances {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(i.Name), i.State)
	}
	return nil
}

func (s *memcacheCmds) runMCInstUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	inst := &memcache.Instance{}
	if err := loadYAMLOrJSONInto(s.flagMCConfigFile, inst); err != nil {
		return err
	}
	mask := s.flagMCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Patch(mcInstName(args[0], project, s.flagMCRegion), inst).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return s.mcFinishOp(ctx, svc, op, "Update instance", args[0], s.flagMCAsync)
}

func (s *memcacheCmds) runMCInstUpgrade(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &memcache.GoogleCloudMemcacheV1UpgradeInstanceRequest{}
	if err := loadYAMLOrJSONInto(s.flagMCConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Upgrade(mcInstName(args[0], project, s.flagMCRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upgrading instance: %w", err)
	}
	return s.mcFinishOp(ctx, svc, op, "Upgrade instance", args[0], s.flagMCAsync)
}

func (s *memcacheCmds) runMCInstApplyParams(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &memcache.ApplyParametersRequest{NodeIds: s.flagMCNodeIDs, ApplyAll: s.flagMCApplyAll}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.ApplyParameters(mcInstName(args[0], project, s.flagMCRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("applying parameters: %w", err)
	}
	return s.mcFinishOp(ctx, svc, op, "Apply parameters", args[0], s.flagMCAsync)
}

// --- operations impl ---

func mcOpName(id, project, region string) string {
	return mcChild("operations", id, mcLocationParent(project, region))
}

func (s *memcacheCmds) runMCOpCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(mcOpName(args[0], project, s.flagMCOpRegion), &memcache.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *memcacheCmds) runMCOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(mcOpName(args[0], project, s.flagMCOpRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagMCFormat)
}

func (s *memcacheCmds) runMCOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Memcache(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(mcLocationParent(project, s.flagMCOpRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagMCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagMCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// memcacheCmds holds the commands and flag values of one invocation's "memcache"
// command group, so invocations don't share state (#2070).
type memcacheCmds struct {
	inv *invocation

	memcacheCmd           *cobra.Command
	memcacheRegionsCmd    *cobra.Command
	mcRegionDescribeCmd   *cobra.Command
	mcRegionListCmd       *cobra.Command
	flagMCFormat          string
	memcacheInstancesCmd  *cobra.Command
	mcInstCreateCmd       *cobra.Command
	mcInstDeleteCmd       *cobra.Command
	mcInstDescribeCmd     *cobra.Command
	mcInstListCmd         *cobra.Command
	mcInstUpdateCmd       *cobra.Command
	mcInstUpgradeCmd      *cobra.Command
	mcInstApplyParamsCmd  *cobra.Command
	flagMCRegion          string
	flagMCConfigFile      string
	flagMCUpdateMask      string
	flagMCAsync           bool
	flagMCNodeIDs         []string
	flagMCApplyAll        bool
	memcacheOperationsCmd *cobra.Command
	mcOpCancelCmd         *cobra.Command
	mcOpDescribeCmd       *cobra.Command
	mcOpListCmd           *cobra.Command
	flagMCOpRegion        string
}

// newMemcacheCmd returns the "memcache" command group for inv.
func newMemcacheCmd(inv *invocation) *cobra.Command {
	return newMemcacheCmds(inv).memcacheCmd
}

// newMemcacheCmds builds the "memcache" command group for inv and returns its state.
func newMemcacheCmds(inv *invocation) *memcacheCmds {
	s := &memcacheCmds{inv: inv}
	s.memcacheCmd = &cobra.Command{Use: "memcache", Short: "Manage Memorystore for Memcached"}
	s.memcacheRegionsCmd = &cobra.Command{Use: "regions", Short: "Explore Memorystore for Memcached regions"}
	s.mcRegionDescribeCmd = &cobra.Command{
		Use: "describe REGION", Short: "Describe a region",
		Args: cobra.ExactArgs(1), RunE: s.runMCRegionDescribe,
	}
	s.mcRegionListCmd = &cobra.Command{
		Use: "list", Short: "List regions",
		Args: cobra.NoArgs, RunE: s.runMCRegionList,
	}
	s.memcacheInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Memcached instances"}
	s.mcInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMCInstCreate,
	}
	s.mcInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an instance",
		Args: cobra.ExactArgs(1), RunE: s.runMCInstDelete,
	}
	s.mcInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an instance",
		Args: cobra.ExactArgs(1), RunE: s.runMCInstDescribe,
	}
	s.mcInstListCmd = &cobra.Command{
		Use: "list", Short: "List instances in a region",
		Args: cobra.NoArgs, RunE: s.runMCInstList,
	}
	s.mcInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update an instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runMCInstUpdate,
	}
	s.mcInstUpgradeCmd = &cobra.Command{
		Use: "upgrade INSTANCE", Short: "Upgrade an instance (--config-file with UpgradeInstanceRequest)",
		Args: cobra.ExactArgs(1), RunE: s.runMCInstUpgrade,
	}
	s.mcInstApplyParamsCmd = &cobra.Command{
		Use: "apply-parameters INSTANCE", Short: "Apply pending parameter updates (--node-ids, --apply-all)",
		Args: cobra.ExactArgs(1), RunE: s.runMCInstApplyParams,
	}
	s.memcacheOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Memcache operations"}
	s.mcOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: s.runMCOpCancel,
	}
	s.mcOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runMCOpDescribe,
	}
	s.mcOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a region",
		Args: cobra.NoArgs, RunE: s.runMCOpList,
	}
	s.initMemcache()
	return s
}
