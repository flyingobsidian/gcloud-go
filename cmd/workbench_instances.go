package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	notebooks "google.golang.org/api/notebooks/v2"
)

var workbenchInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage Vertex AI Workbench instances",
}

var (
	wbiCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Workbench instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runWBICreate,
	}
	wbiDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIDelete,
	}
	wbiDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIDescribe,
	}
	wbiListCmd = &cobra.Command{
		Use: "list", Short: "List Workbench instances in a location",
		Args: cobra.NoArgs, RunE: runWBIList,
	}
	wbiUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a Workbench instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runWBIUpdate,
	}
	wbiStartCmd = &cobra.Command{
		Use: "start INSTANCE", Short: "Start a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIStart,
	}
	wbiStopCmd = &cobra.Command{
		Use: "stop INSTANCE", Short: "Stop a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIStop,
	}
	wbiResetCmd = &cobra.Command{
		Use: "reset INSTANCE", Short: "Reset a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIReset,
	}
	wbiDiagnoseCmd = &cobra.Command{
		Use: "diagnose INSTANCE", Short: "Collect a diagnostic snapshot of a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIDiagnose,
	}
	wbiGetConfigCmd = &cobra.Command{
		Use: "get-config", Short: "Get the Workbench instance configuration for a location",
		Args: cobra.NoArgs, RunE: runWBIGetConfig,
	}
	wbiCheckUpgCmd = &cobra.Command{
		Use: "check-instance-upgradability INSTANCE", Short: "Check whether an instance can be upgraded",
		Args: cobra.ExactArgs(1), RunE: runWBICheckUpgrade,
	}
	wbiUpgradeCmd = &cobra.Command{
		Use: "upgrade INSTANCE", Short: "Upgrade a Workbench instance to the latest image",
		Args: cobra.ExactArgs(1), RunE: runWBIUpgrade,
	}
	wbiRollbackCmd = &cobra.Command{
		Use: "rollback INSTANCE", Short: "Roll back a Workbench instance to a previous snapshot",
		Args: cobra.ExactArgs(1), RunE: runWBIRollback,
	}
	wbiRestoreCmd = &cobra.Command{
		Use: "restore INSTANCE", Short: "Restore a Workbench instance from a snapshot",
		Args: cobra.ExactArgs(1), RunE: runWBIRestore,
	}
	wbiResizeDiskCmd = &cobra.Command{
		Use: "resize-disk INSTANCE", Short: "Resize a Workbench instance's disk",
		Args: cobra.ExactArgs(1), RunE: runWBIResizeDisk,
	}
	wbiGetIamCmd = &cobra.Command{
		Use: "get-iam-policy INSTANCE", Short: "Print the IAM policy for a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIGetIam,
	}
	wbiSetIamCmd = &cobra.Command{
		Use: "set-iam-policy INSTANCE POLICY_FILE", Short: "Replace the IAM policy for a Workbench instance",
		Args: cobra.ExactArgs(2), RunE: runWBISetIam,
	}
	wbiAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding INSTANCE", Short: "Add an IAM binding to a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIAddIam,
	}
	wbiRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding INSTANCE", Short: "Remove an IAM binding from a Workbench instance",
		Args: cobra.ExactArgs(1), RunE: runWBIRemoveIam,
	}
)

var (
	flagWBILocation      string
	flagWBIConfigFile    string
	flagWBIUpdateMask    string
	flagWBIFormat        string
	flagWBIAsync         bool
	flagWBIDiskSize      int64
	flagWBIDiskType      string
	flagWBISnapshot      string
	flagWBIRestoreSource string
	flagWBIRestoreTime   string
	flagWBIRevision      string
	flagWBIIamMember     string
	flagWBIIamRole       string
)

func init() {
	all := []*cobra.Command{
		wbiCreateCmd, wbiDeleteCmd, wbiDescribeCmd, wbiListCmd, wbiUpdateCmd,
		wbiStartCmd, wbiStopCmd, wbiResetCmd, wbiDiagnoseCmd, wbiGetConfigCmd,
		wbiCheckUpgCmd, wbiUpgradeCmd, wbiRollbackCmd, wbiRestoreCmd, wbiResizeDiskCmd,
		wbiGetIamCmd, wbiSetIamCmd, wbiAddIamCmd, wbiRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagWBILocation, "location", "", "Location containing the instance (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{wbiCreateCmd, wbiUpdateCmd} {
		c.Flags().StringVar(&flagWBIConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Instance message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	wbiUpdateCmd.Flags().StringVar(&flagWBIUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{
		wbiCreateCmd, wbiDeleteCmd, wbiUpdateCmd, wbiStartCmd, wbiStopCmd,
		wbiResetCmd, wbiDiagnoseCmd, wbiUpgradeCmd, wbiRollbackCmd, wbiRestoreCmd,
		wbiResizeDiskCmd,
	} {
		c.Flags().BoolVar(&flagWBIAsync, "async", false, "Return the long-running operation without waiting")
	}
	wbiDescribeCmd.Flags().StringVar(&flagWBIFormat, "format", "", "Output format")
	wbiListCmd.Flags().StringVar(&flagWBIFormat, "format", "", "Output format")
	wbiGetConfigCmd.Flags().StringVar(&flagWBIFormat, "format", "", "Output format")
	wbiCheckUpgCmd.Flags().StringVar(&flagWBIFormat, "format", "", "Output format")
	wbiGetIamCmd.Flags().StringVar(&flagWBIFormat, "format", "", "Output format")

	wbiResizeDiskCmd.Flags().Int64Var(&flagWBIDiskSize, "new-size", 0,
		"New disk size in GB (required)")
	wbiResizeDiskCmd.Flags().StringVar(&flagWBIDiskType, "disk-type", "DATA_DISK",
		"Which disk to resize: DATA_DISK or BOOT_DISK")
	_ = wbiResizeDiskCmd.MarkFlagRequired("new-size")

	wbiRollbackCmd.Flags().StringVar(&flagWBISnapshot, "target-snapshot", "",
		"Fully qualified snapshot resource name to rollback to (required)")
	wbiRollbackCmd.Flags().StringVar(&flagWBIRevision, "revision-id", "",
		"Revision ID to rollback to")
	_ = wbiRollbackCmd.MarkFlagRequired("target-snapshot")

	wbiRestoreCmd.Flags().StringVar(&flagWBIRestoreSource, "snapshot", "",
		"Fully qualified snapshot resource name to restore from")
	wbiRestoreCmd.Flags().StringVar(&flagWBIRestoreTime, "snapshot-time", "",
		"Point-in-time snapshot timestamp (RFC3339)")

	for _, c := range []*cobra.Command{wbiAddIamCmd, wbiRemoveIamCmd} {
		c.Flags().StringVar(&flagWBIIamMember, "member", "",
			"IAM member, e.g. user:foo@example.com (required)")
		c.Flags().StringVar(&flagWBIIamRole, "role", "",
			"IAM role, e.g. roles/notebooks.viewer (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}

	workbenchInstancesCmd.AddCommand(all...)
	workbenchCmd.AddCommand(workbenchInstancesCmd)
}

func wbiName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, location, id)
}

func wbiWaitOp(ctx context.Context, svc *notebooks.Service, op *notebooks.Operation) (*notebooks.Operation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func wbiFinishOp(ctx context.Context, svc *notebooks.Service, op *notebooks.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := wbiWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

func runWBICreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &notebooks.Instance{}
	if err := loadYAMLOrJSONInto(flagWBIConfigFile, inst); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(fmt.Sprintf("projects/%s/locations/%s", project, flagWBILocation), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return wbiFinishOp(ctx, svc, op, "Create instance", args[0], flagWBIAsync)
}

func runWBIDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(wbiName(args[0], project, flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return wbiFinishOp(ctx, svc, op, "Delete instance", args[0], flagWBIAsync)
}

func runWBIDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	inst, err := svc.Projects.Locations.Instances.Get(wbiName(args[0], project, flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(inst, flagWBIFormat)
}

func runWBIList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.List(fmt.Sprintf("projects/%s/locations/%s", project, flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if flagWBIFormat != "" {
		return emitFormatted(resp.Instances, flagWBIFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, i := range resp.Instances {
		fmt.Printf("%-40s %s\n", path.Base(i.Name), i.State)
	}
	return nil
}

func runWBIUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	inst := &notebooks.Instance{}
	if err := loadYAMLOrJSONInto(flagWBIConfigFile, inst); err != nil {
		return err
	}
	mask := flagWBIUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Patch(wbiName(args[0], project, flagWBILocation), inst).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return wbiFinishOp(ctx, svc, op, "Update instance", args[0], flagWBIAsync)
}

func wbiInvokeAction(name string, verb string, action func(ctx context.Context, svc *notebooks.Service, resourceName string) (*notebooks.Operation, error)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := action(ctx, svc, wbiName(name, project, flagWBILocation))
	if err != nil {
		return fmt.Errorf("%s: %w", strings.ToLower(verb), err)
	}
	return wbiFinishOp(ctx, svc, op, verb, name, flagWBIAsync)
}

func runWBIStart(cmd *cobra.Command, args []string) error {
	return wbiInvokeAction(args[0], "Start instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Start(name, &notebooks.StartInstanceRequest{}).Context(ctx).Do()
	})
}

func runWBIStop(cmd *cobra.Command, args []string) error {
	return wbiInvokeAction(args[0], "Stop instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Stop(name, &notebooks.StopInstanceRequest{}).Context(ctx).Do()
	})
}

func runWBIReset(cmd *cobra.Command, args []string) error {
	return wbiInvokeAction(args[0], "Reset instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Reset(name, &notebooks.ResetInstanceRequest{}).Context(ctx).Do()
	})
}

func runWBIDiagnose(cmd *cobra.Command, args []string) error {
	return wbiInvokeAction(args[0], "Diagnose instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Diagnose(name, &notebooks.DiagnoseInstanceRequest{}).Context(ctx).Do()
	})
}

func runWBIUpgrade(cmd *cobra.Command, args []string) error {
	return wbiInvokeAction(args[0], "Upgrade instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Upgrade(name, &notebooks.UpgradeInstanceRequest{}).Context(ctx).Do()
	})
}

func runWBIRollback(cmd *cobra.Command, args []string) error {
	req := &notebooks.RollbackInstanceRequest{TargetSnapshot: flagWBISnapshot, RevisionId: flagWBIRevision}
	return wbiInvokeAction(args[0], "Rollback instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Rollback(name, req).Context(ctx).Do()
	})
}

func runWBIRestore(cmd *cobra.Command, args []string) error {
	req := &notebooks.RestoreInstanceRequest{Snapshot: &notebooks.Snapshot{ProjectId: "", SnapshotId: flagWBIRestoreSource}}
	return wbiInvokeAction(args[0], "Restore instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Restore(name, req).Context(ctx).Do()
	})
}

func runWBIResizeDisk(cmd *cobra.Command, args []string) error {
	req := &notebooks.ResizeDiskRequest{}
	switch strings.ToUpper(flagWBIDiskType) {
	case "BOOT_DISK":
		req.BootDisk = &notebooks.BootDisk{DiskSizeGb: flagWBIDiskSize}
	default:
		req.DataDisk = &notebooks.DataDisk{DiskSizeGb: flagWBIDiskSize}
	}
	return wbiInvokeAction(args[0], "Resize disk", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.ResizeDisk(name, req).Context(ctx).Do()
	})
}

func runWBIGetConfig(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.Instances.GetConfig(fmt.Sprintf("projects/%s/locations/%s/instances", project, flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance config: %w", err)
	}
	return emitFormatted(cfg, flagWBIFormat)
}

func runWBICheckUpgrade(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.CheckUpgradability(wbiName(args[0], project, flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking upgradability: %w", err)
	}
	return emitFormatted(resp, flagWBIFormat)
}

func runWBIGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(wbiName(args[0], project, flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagWBIFormat)
}

func runWBISetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &notebooks.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &notebooks.SetIamPolicyRequest{Policy: policy}
	got, err := svc.Projects.Locations.Instances.SetIamPolicy(wbiName(args[0], project, flagWBILocation), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runWBIAddIam(cmd *cobra.Command, args []string) error {
	return wbiModifyIam(args[0], func(p *notebooks.Policy) {
		for _, b := range p.Bindings {
			if b.Role == flagWBIIamRole {
				for _, m := range b.Members {
					if m == flagWBIIamMember {
						return
					}
				}
				b.Members = append(b.Members, flagWBIIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &notebooks.Binding{Role: flagWBIIamRole, Members: []string{flagWBIIamMember}})
	})
}

func runWBIRemoveIam(cmd *cobra.Command, args []string) error {
	return wbiModifyIam(args[0], func(p *notebooks.Policy) {
		for _, b := range p.Bindings {
			if b.Role != flagWBIIamRole {
				continue
			}
			filtered := b.Members[:0]
			for _, m := range b.Members {
				if m != flagWBIIamMember {
					filtered = append(filtered, m)
				}
			}
			b.Members = filtered
		}
	})
}

func wbiModifyIam(name string, mutate func(*notebooks.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resourceName := wbiName(name, project, flagWBILocation)
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(resourceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Instances.SetIamPolicy(resourceName, &notebooks.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}
