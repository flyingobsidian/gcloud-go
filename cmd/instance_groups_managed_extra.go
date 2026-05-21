package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- managed create ---

var managedCreateCmd = &cobra.Command{
	Use:   "create INSTANCE_GROUP",
	Short: "Create a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedCreate,
}

var (
	flagManagedTemplate     string
	flagManagedSize         int64
	flagManagedCreateRegion string
)

// --- managed delete ---

var managedDeleteCmd = &cobra.Command{
	Use:   "delete INSTANCE_GROUP",
	Short: "Delete a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedDelete,
}

var flagManagedDeleteRegion string

// --- managed list ---

var managedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List managed instance groups",
	Args:  cobra.NoArgs,
	RunE:  runManagedList,
}

var flagManagedListAllFormat string

// --- managed update ---

var managedUpdateCmd = &cobra.Command{
	Use:   "update INSTANCE_GROUP",
	Short: "Update a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedUpdate,
}

var (
	flagManagedUpdateRegion string
	flagManagedUpdateTemplate string
)

// --- managed set-instance-template ---

var managedSetTemplateCmd = &cobra.Command{
	Use:   "set-instance-template INSTANCE_GROUP",
	Short: "Set the instance template for a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedSetTemplate,
}

var (
	flagManagedSetTemplateRegion   string
	flagManagedSetTemplateName     string
)

// --- managed recreate-instances ---

var managedRecreateCmd = &cobra.Command{
	Use:   "recreate-instances INSTANCE_GROUP",
	Short: "Recreate instances in a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedRecreateInstances,
}

var (
	flagManagedRecreateRegion    string
	flagManagedRecreateInstances []string
)

// --- managed set-autoscaling ---

var managedSetAutoscalingCmd = &cobra.Command{
	Use:   "set-autoscaling INSTANCE_GROUP",
	Short: "Set autoscaling for a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedSetAutoscaling,
}

var (
	flagAutoscalingRegion    string
	flagAutoscalingMin       int64
	flagAutoscalingMax       int64
	flagAutoscalingCPUTarget float64
)

// --- managed stop-autoscaling ---

var managedStopAutoscalingCmd = &cobra.Command{
	Use:   "stop-autoscaling INSTANCE_GROUP",
	Short: "Stop autoscaling for a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedStopAutoscaling,
}

var flagStopAutoscalingRegion string

func init() {
	// create
	managedCreateCmd.Flags().StringVar(&flagManagedTemplate, "template", "", "Instance template name")
	managedCreateCmd.MarkFlagRequired("template")
	managedCreateCmd.Flags().Int64Var(&flagManagedSize, "size", 1, "Target size")
	managedCreateCmd.Flags().StringVar(&flagManagedCreateRegion, "region", "", "Region for regional MIG")

	// delete
	managedDeleteCmd.Flags().StringVar(&flagManagedDeleteRegion, "region", "", "Region")

	// list
	managedListCmd.Flags().StringVar(&flagManagedListAllFormat, "format", "", "Output format (e.g. json)")

	// update
	managedUpdateCmd.Flags().StringVar(&flagManagedUpdateRegion, "region", "", "Region")
	managedUpdateCmd.Flags().StringVar(&flagManagedUpdateTemplate, "template", "", "New instance template")

	// set-instance-template
	managedSetTemplateCmd.Flags().StringVar(&flagManagedSetTemplateRegion, "region", "", "Region")
	managedSetTemplateCmd.Flags().StringVar(&flagManagedSetTemplateName, "template", "", "Instance template name")
	managedSetTemplateCmd.MarkFlagRequired("template")

	// recreate-instances
	managedRecreateCmd.Flags().StringVar(&flagManagedRecreateRegion, "region", "", "Region")
	managedRecreateCmd.Flags().StringSliceVar(&flagManagedRecreateInstances, "instances", nil, "Instance URLs to recreate")
	managedRecreateCmd.MarkFlagRequired("instances")

	// set-autoscaling
	managedSetAutoscalingCmd.Flags().StringVar(&flagAutoscalingRegion, "region", "", "Region")
	managedSetAutoscalingCmd.Flags().Int64Var(&flagAutoscalingMin, "min-num-replicas", 1, "Minimum number of replicas")
	managedSetAutoscalingCmd.Flags().Int64Var(&flagAutoscalingMax, "max-num-replicas", 0, "Maximum number of replicas")
	managedSetAutoscalingCmd.MarkFlagRequired("max-num-replicas")
	managedSetAutoscalingCmd.Flags().Float64Var(&flagAutoscalingCPUTarget, "target-cpu-utilization", 0.6, "Target CPU utilization (0.0-1.0)")

	// stop-autoscaling
	managedStopAutoscalingCmd.Flags().StringVar(&flagStopAutoscalingRegion, "region", "", "Region")

	managedCmd.AddCommand(managedCreateCmd)
	managedCmd.AddCommand(managedDeleteCmd)
	managedCmd.AddCommand(managedListCmd)
	managedCmd.AddCommand(managedUpdateCmd)
	managedCmd.AddCommand(managedSetTemplateCmd)
	managedCmd.AddCommand(managedRecreateCmd)
	managedCmd.AddCommand(managedSetAutoscalingCmd)
	managedCmd.AddCommand(managedStopAutoscalingCmd)
}

func resolveTemplateURL(project, template string) string {
	return fmt.Sprintf("projects/%s/global/instanceTemplates/%s", project, template)
}

func runManagedCreate(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		props, loadErr := loadProps()
		if loadErr != nil {
			return err
		}
		project = resolveProjectOnly(props)
		if project == "" {
			return err
		}
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	mig := &compute.InstanceGroupManager{
		Name:             group,
		InstanceTemplate: resolveTemplateURL(project, flagManagedTemplate),
		TargetSize:       flagManagedSize,
	}

	if flagManagedCreateRegion != "" {
		op, err := svc.RegionInstanceGroupManagers.Insert(project, flagManagedCreateRegion, mig).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating managed instance group: %w", err)
		}
		fmt.Printf("Create operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	op, err := svc.InstanceGroupManagers.Insert(project, zone, mig).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating managed instance group: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Created managed instance group [%s].\n", group)
	return nil
}

func runManagedDelete(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		props, loadErr := loadProps()
		if loadErr != nil {
			return err
		}
		project = resolveProjectOnly(props)
		if project == "" {
			return err
		}
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	if flagManagedDeleteRegion != "" {
		op, err := svc.RegionInstanceGroupManagers.Delete(project, flagManagedDeleteRegion, group).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting managed instance group: %w", err)
		}
		fmt.Printf("Delete operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	op, err := svc.InstanceGroupManagers.Delete(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting managed instance group: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Deleted managed instance group [%s].\n", group)
	return nil
}

func runManagedList(cmd *cobra.Command, args []string) error {
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var migs []*compute.InstanceGroupManager
	pageToken := ""
	for {
		call := svc.InstanceGroupManagers.List(project, zone).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing managed instance groups: %w", err)
		}
		migs = append(migs, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagManagedListAllFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(migs)
	}

	fmt.Printf("%-30s %-15s %-10s %-30s\n", "NAME", "ZONE", "SIZE", "TEMPLATE")
	for _, m := range migs {
		tmpl := path.Base(m.InstanceTemplate)
		fmt.Printf("%-30s %-15s %-10d %-30s\n", m.Name, zone, m.TargetSize, tmpl)
	}
	return nil
}

func runManagedUpdate(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	if flagManagedUpdateRegion != "" {
		mig, err := svc.RegionInstanceGroupManagers.Get(project, flagManagedUpdateRegion, group).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting MIG: %w", err)
		}
		if flagManagedUpdateTemplate != "" {
			mig.InstanceTemplate = resolveTemplateURL(project, flagManagedUpdateTemplate)
		}
		op, err := svc.RegionInstanceGroupManagers.Patch(project, flagManagedUpdateRegion, group, mig).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating MIG: %w", err)
		}
		fmt.Printf("Update operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	mig, err := svc.InstanceGroupManagers.Get(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting MIG: %w", err)
	}
	if flagManagedUpdateTemplate != "" {
		mig.InstanceTemplate = resolveTemplateURL(project, flagManagedUpdateTemplate)
	}
	op, err := svc.InstanceGroupManagers.Patch(project, zone, group, mig).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating MIG: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Updated managed instance group [%s].\n", group)
	return nil
}

func runManagedSetTemplate(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	templateURL := resolveTemplateURL(project, flagManagedSetTemplateName)
	req := &compute.InstanceGroupManagersSetInstanceTemplateRequest{
		InstanceTemplate: templateURL,
	}

	if flagManagedSetTemplateRegion != "" {
		op, err := svc.RegionInstanceGroupManagers.SetInstanceTemplate(project, flagManagedSetTemplateRegion, group, req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("setting instance template: %w", err)
		}
		fmt.Printf("Set-template operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	op, err := svc.InstanceGroupManagers.SetInstanceTemplate(project, zone, group, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting instance template: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Set instance template for [%s].\n", group)
	return nil
}

func runManagedRecreateInstances(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	req := &compute.InstanceGroupManagersRecreateInstancesRequest{
		Instances: flagManagedRecreateInstances,
	}

	if flagManagedRecreateRegion != "" {
		reqR := &compute.RegionInstanceGroupManagersRecreateRequest{
			Instances: flagManagedRecreateInstances,
		}
		op, err := svc.RegionInstanceGroupManagers.RecreateInstances(project, flagManagedRecreateRegion, group, reqR).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("recreating instances: %w", err)
		}
		fmt.Printf("Recreate operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	op, err := svc.InstanceGroupManagers.RecreateInstances(project, zone, group, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("recreating instances: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Recreated instances in [%s].\n", group)
	return nil
}

func runManagedSetAutoscaling(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	autoscaler := &compute.Autoscaler{
		Name:   group + "-autoscaler",
		Target: "", // will be set below
		AutoscalingPolicy: &compute.AutoscalingPolicy{
			MinNumReplicas: flagAutoscalingMin,
			MaxNumReplicas: flagAutoscalingMax,
			CpuUtilization: &compute.AutoscalingPolicyCpuUtilization{
				UtilizationTarget: flagAutoscalingCPUTarget,
			},
		},
	}

	if flagAutoscalingRegion != "" {
		autoscaler.Target = fmt.Sprintf("projects/%s/regions/%s/instanceGroupManagers/%s", project, flagAutoscalingRegion, group)
		op, err := svc.RegionAutoscalers.Insert(project, flagAutoscalingRegion, autoscaler).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("setting autoscaling: %w", err)
		}
		fmt.Printf("Autoscaling operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	autoscaler.Target = fmt.Sprintf("projects/%s/zones/%s/instanceGroupManagers/%s", project, zone, group)
	op, err := svc.Autoscalers.Insert(project, zone, autoscaler).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting autoscaling: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Set autoscaling for [%s].\n", group)
	return nil
}

func runManagedStopAutoscaling(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	autoscalerName := group + "-autoscaler"

	if flagStopAutoscalingRegion != "" {
		op, err := svc.RegionAutoscalers.Delete(project, flagStopAutoscalingRegion, autoscalerName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("stopping autoscaling: %w", err)
		}
		fmt.Printf("Stop-autoscaling operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	op, err := svc.Autoscalers.Delete(project, zone, autoscalerName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping autoscaling: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Stopped autoscaling for [%s].\n", group)
	return nil
}
