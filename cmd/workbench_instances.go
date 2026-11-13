package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	notebooks "google.golang.org/api/notebooks/v2"
)

func (g *workbenchCmds) initWorkbenchInstances() {
	all := []*cobra.Command{
		g.wbiCreateCmd, g.wbiDeleteCmd, g.wbiDescribeCmd, g.wbiListCmd, g.wbiUpdateCmd,
		g.wbiStartCmd, g.wbiStopCmd, g.wbiResetCmd, g.wbiDiagnoseCmd, g.wbiGetConfigCmd,
		g.wbiCheckUpgCmd, g.wbiUpgradeCmd, g.wbiRollbackCmd, g.wbiRestoreCmd, g.wbiResizeDiskCmd,
		g.wbiGetIamCmd, g.wbiSetIamCmd, g.wbiAddIamCmd, g.wbiRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagWBILocation, "location", "", "Location containing the instance (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.wbiCreateCmd, g.wbiUpdateCmd} {
		c.Flags().StringVar(&g.flagWBIConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Instance message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.wbiUpdateCmd.Flags().StringVar(&g.flagWBIUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{
		g.wbiCreateCmd, g.wbiDeleteCmd, g.wbiUpdateCmd, g.wbiStartCmd, g.wbiStopCmd,
		g.wbiResetCmd, g.wbiDiagnoseCmd, g.wbiUpgradeCmd, g.wbiRollbackCmd, g.wbiRestoreCmd,
		g.wbiResizeDiskCmd,
	} {
		c.Flags().BoolVar(&g.flagWBIAsync, "async", false, "Return the long-running operation without waiting")
	}
	g.wbiDescribeCmd.Flags().StringVar(&g.flagWBIFormat, "format", "", "Output format")
	g.wbiListCmd.Flags().StringVar(&g.flagWBIFormat, "format", "", "Output format")
	g.wbiGetConfigCmd.Flags().StringVar(&g.flagWBIFormat, "format", "", "Output format")
	g.wbiCheckUpgCmd.Flags().StringVar(&g.flagWBIFormat, "format", "", "Output format")
	g.wbiGetIamCmd.Flags().StringVar(&g.flagWBIFormat, "format", "", "Output format")

	g.wbiResizeDiskCmd.Flags().Int64Var(&g.flagWBIDiskSize, "new-size", 0,
		"New disk size in GB (required)")
	g.wbiResizeDiskCmd.Flags().StringVar(&g.flagWBIDiskType, "disk-type", "DATA_DISK",
		"Which disk to resize: DATA_DISK or BOOT_DISK")
	_ = g.wbiResizeDiskCmd.MarkFlagRequired("new-size")

	g.wbiRollbackCmd.Flags().StringVar(&g.flagWBISnapshot, "target-snapshot", "",
		"Fully qualified snapshot resource name to rollback to (required)")
	g.wbiRollbackCmd.Flags().StringVar(&g.flagWBIRevision, "revision-id", "",
		"Revision ID to rollback to")
	_ = g.wbiRollbackCmd.MarkFlagRequired("target-snapshot")

	g.wbiRestoreCmd.Flags().StringVar(&g.flagWBIRestoreSource, "snapshot", "",
		"Fully qualified snapshot resource name to restore from")
	g.wbiRestoreCmd.Flags().StringVar(&g.flagWBIRestoreTime, "snapshot-time", "",
		"Point-in-time snapshot timestamp (RFC3339)")

	g.wbiUpgradeCmd.Flags().StringVar(&g.flagWBIImageFamily, "image-family", "",
		"Compute Engine image family resource name to upgrade to (e.g. projects/PROJECT/global/images/family/FAMILY); "+
			"upgrades to the latest image in the specified family. If omitted, upgrades to the latest image in the instance's current family.")

	for _, c := range []*cobra.Command{g.wbiAddIamCmd, g.wbiRemoveIamCmd} {
		c.Flags().StringVar(&g.flagWBIIamMember, "member", "",
			"IAM member, e.g. user:foo@example.com (required)")
		c.Flags().StringVar(&g.flagWBIIamRole, "role", "",
			"IAM role, e.g. roles/notebooks.viewer (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}

	g.workbenchInstancesCmd.AddCommand(all...)
	g.workbenchCmd.AddCommand(g.workbenchInstancesCmd)
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

func (g *workbenchCmds) wbiFinishOp(ctx context.Context, svc *notebooks.Service, op *notebooks.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(g.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(g.inv.stdout, op, "")
	}
	final, err := wbiWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(g.inv.stdout, final.Response, "")
	}
	return nil
}

func (g *workbenchCmds) runWBICreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	inst := &notebooks.Instance{}
	if err := loadYAMLOrJSONInto(g.flagWBIConfigFile, inst); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(fmt.Sprintf("projects/%s/locations/%s", project, g.flagWBILocation), inst).
		InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return g.wbiFinishOp(ctx, svc, op, "Create instance", args[0], g.flagWBIAsync)
}

func (g *workbenchCmds) runWBIDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(wbiName(args[0], project, g.flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return g.wbiFinishOp(ctx, svc, op, "Delete instance", args[0], g.flagWBIAsync)
}

func (g *workbenchCmds) runWBIDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	inst, err := svc.Projects.Locations.Instances.Get(wbiName(args[0], project, g.flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, inst, g.flagWBIFormat)
}

func (g *workbenchCmds) runWBIList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.List(fmt.Sprintf("projects/%s/locations/%s", project, g.flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if g.flagWBIFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Instances, g.flagWBIFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Instances), "%-40s %s\n", "NAME", "STATE")
	for _, i := range resp.Instances {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(i.Name), i.State)
	}
	return nil
}

func (g *workbenchCmds) runWBIUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	inst := &notebooks.Instance{}
	if err := loadYAMLOrJSONInto(g.flagWBIConfigFile, inst); err != nil {
		return err
	}
	mask := g.flagWBIUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Patch(wbiName(args[0], project, g.flagWBILocation), inst).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return g.wbiFinishOp(ctx, svc, op, "Update instance", args[0], g.flagWBIAsync)
}

func (g *workbenchCmds) wbiInvokeAction(name string, verb string, action func(ctx context.Context, svc *notebooks.Service, resourceName string) (*notebooks.Operation, error)) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	op, err := action(ctx, svc, wbiName(name, project, g.flagWBILocation))
	if err != nil {
		return fmt.Errorf("%s: %w", strings.ToLower(verb), err)
	}
	return g.wbiFinishOp(ctx, svc, op, verb, name, g.flagWBIAsync)
}

func (g *workbenchCmds) runWBIStart(cmd *cobra.Command, args []string) error {
	return g.wbiInvokeAction(args[0], "Start instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Start(name, &notebooks.StartInstanceRequest{}).Context(ctx).Do()
	})
}

func (g *workbenchCmds) runWBIStop(cmd *cobra.Command, args []string) error {
	return g.wbiInvokeAction(args[0], "Stop instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Stop(name, &notebooks.StopInstanceRequest{}).Context(ctx).Do()
	})
}

func (g *workbenchCmds) runWBIReset(cmd *cobra.Command, args []string) error {
	return g.wbiInvokeAction(args[0], "Reset instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Reset(name, &notebooks.ResetInstanceRequest{}).Context(ctx).Do()
	})
}

func (g *workbenchCmds) runWBIDiagnose(cmd *cobra.Command, args []string) error {
	return g.wbiInvokeAction(args[0], "Diagnose instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Diagnose(name, &notebooks.DiagnoseInstanceRequest{}).Context(ctx).Do()
	})
}

func (g *workbenchCmds) runWBIUpgrade(cmd *cobra.Command, args []string) error {
	req := &notebooks.UpgradeInstanceRequest{ImageFamily: g.flagWBIImageFamily}
	return g.wbiInvokeAction(args[0], "Upgrade instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Upgrade(name, req).Context(ctx).Do()
	})
}

func (g *workbenchCmds) runWBIRollback(cmd *cobra.Command, args []string) error {
	req := &notebooks.RollbackInstanceRequest{TargetSnapshot: g.flagWBISnapshot, RevisionId: g.flagWBIRevision}
	return g.wbiInvokeAction(args[0], "Rollback instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Rollback(name, req).Context(ctx).Do()
	})
}

func (g *workbenchCmds) runWBIRestore(cmd *cobra.Command, args []string) error {
	req := &notebooks.RestoreInstanceRequest{Snapshot: &notebooks.Snapshot{ProjectId: "", SnapshotId: g.flagWBIRestoreSource}}
	return g.wbiInvokeAction(args[0], "Restore instance", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.Restore(name, req).Context(ctx).Do()
	})
}

func (g *workbenchCmds) runWBIResizeDisk(cmd *cobra.Command, args []string) error {
	req := &notebooks.ResizeDiskRequest{}
	switch strings.ToUpper(g.flagWBIDiskType) {
	case "BOOT_DISK":
		req.BootDisk = &notebooks.BootDisk{DiskSizeGb: g.flagWBIDiskSize}
	default:
		req.DataDisk = &notebooks.DataDisk{DiskSizeGb: g.flagWBIDiskSize}
	}
	return g.wbiInvokeAction(args[0], "Resize disk", func(ctx context.Context, svc *notebooks.Service, name string) (*notebooks.Operation, error) {
		return svc.Projects.Locations.Instances.ResizeDisk(name, req).Context(ctx).Do()
	})
}

func (g *workbenchCmds) runWBIGetConfig(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.Instances.GetConfig(fmt.Sprintf("projects/%s/locations/%s/instances", project, g.flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance config: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, cfg, g.flagWBIFormat)
}

func (g *workbenchCmds) runWBICheckUpgrade(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Instances.CheckUpgradability(wbiName(args[0], project, g.flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking upgradability: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagWBIFormat)
}

func (g *workbenchCmds) runWBIGetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(wbiName(args[0], project, g.flagWBILocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, policy, g.flagWBIFormat)
}

func (g *workbenchCmds) runWBISetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	policy := &notebooks.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	req := &notebooks.SetIamPolicyRequest{Policy: policy}
	got, err := svc.Projects.Locations.Instances.SetIamPolicy(wbiName(args[0], project, g.flagWBILocation), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *workbenchCmds) runWBIAddIam(cmd *cobra.Command, args []string) error {
	return g.wbiModifyIam(args[0], func(p *notebooks.Policy) {
		for _, b := range p.Bindings {
			if b.Role == g.flagWBIIamRole {
				for _, m := range b.Members {
					if m == g.flagWBIIamMember {
						return
					}
				}
				b.Members = append(b.Members, g.flagWBIIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &notebooks.Binding{Role: g.flagWBIIamRole, Members: []string{g.flagWBIIamMember}})
	})
}

func (g *workbenchCmds) runWBIRemoveIam(cmd *cobra.Command, args []string) error {
	return g.wbiModifyIam(args[0], func(p *notebooks.Policy) {
		for _, b := range p.Bindings {
			if b.Role != g.flagWBIIamRole {
				continue
			}
			filtered := b.Members[:0]
			for _, m := range b.Members {
				if m != g.flagWBIIamMember {
					filtered = append(filtered, m)
				}
			}
			b.Members = filtered
		}
	})
}

func (g *workbenchCmds) wbiModifyIam(name string, mutate func(*notebooks.Policy)) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().Notebooks(ctx)
	if err != nil {
		return err
	}
	resourceName := wbiName(name, project, g.flagWBILocation)
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(resourceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Instances.SetIamPolicy(resourceName, &notebooks.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}
