package cli

import (
	"fmt"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- managed create ---

// --- managed delete ---

// --- managed list ---

// --- managed update ---

// --- managed set-instance-template ---

// --- managed recreate-instances ---

// --- managed set-autoscaling ---

// --- managed stop-autoscaling ---

func (s *computeCmds) initInstanceGroupsManagedExtra() {
	// create
	s.managedCreateCmd.Flags().StringVar(&s.flagManagedTemplate, "template", "", "Instance template name")
	s.managedCreateCmd.MarkFlagRequired("template")
	s.managedCreateCmd.Flags().Int64Var(&s.flagManagedSize, "size", 1, "Target size")
	s.managedCreateCmd.Flags().StringVar(&s.flagManagedCreateRegion, "region", "", "Region for regional MIG")

	// delete
	s.managedDeleteCmd.Flags().StringVar(&s.flagManagedDeleteRegion, "region", "", "Region")

	// list
	s.managedListCmd.Flags().BoolVar(&s.flagManagedListURI, "uri", false, "Print resource URIs")

	// update
	s.managedUpdateCmd.Flags().StringVar(&s.flagManagedUpdateRegion, "region", "", "Region")
	s.managedUpdateCmd.Flags().StringVar(&s.flagManagedUpdateTemplate, "template", "", "New instance template")

	// set-instance-template
	s.managedSetTemplateCmd.Flags().StringVar(&s.flagManagedSetTemplateRegion, "region", "", "Region")
	s.managedSetTemplateCmd.Flags().StringVar(&s.flagManagedSetTemplateName, "template", "", "Instance template name")
	s.managedSetTemplateCmd.MarkFlagRequired("template")

	// recreate-instances
	s.managedRecreateCmd.Flags().StringVar(&s.flagManagedRecreateRegion, "region", "", "Region")
	s.managedRecreateCmd.Flags().StringSliceVar(&s.flagManagedRecreateInstances, "instances", nil, "Instance URLs to recreate")
	s.managedRecreateCmd.MarkFlagRequired("instances")

	// set-autoscaling
	s.managedSetAutoscalingCmd.Flags().StringVar(&s.flagAutoscalingRegion, "region", "", "Region")
	s.managedSetAutoscalingCmd.Flags().Int64Var(&s.flagAutoscalingMin, "min-num-replicas", 1, "Minimum number of replicas")
	s.managedSetAutoscalingCmd.Flags().Int64Var(&s.flagAutoscalingMax, "max-num-replicas", 0, "Maximum number of replicas")
	s.managedSetAutoscalingCmd.MarkFlagRequired("max-num-replicas")
	s.managedSetAutoscalingCmd.Flags().Float64Var(&s.flagAutoscalingCPUTarget, "target-cpu-utilization", 0.6, "Target CPU utilization (0.0-1.0)")

	// stop-autoscaling
	s.managedStopAutoscalingCmd.Flags().StringVar(&s.flagStopAutoscalingRegion, "region", "", "Region")

	s.managedCmd.AddCommand(s.managedCreateCmd)
	s.managedCmd.AddCommand(s.managedDeleteCmd)
	s.managedCmd.AddCommand(s.managedListCmd)
	s.managedCmd.AddCommand(s.managedUpdateCmd)
	s.managedCmd.AddCommand(s.managedSetTemplateCmd)
	s.managedCmd.AddCommand(s.managedRecreateCmd)
	s.managedCmd.AddCommand(s.managedSetAutoscalingCmd)
	s.managedCmd.AddCommand(s.managedStopAutoscalingCmd)
}

func resolveTemplateURL(project, template string) string {
	return fmt.Sprintf("projects/%s/global/instanceTemplates/%s", project, template)
}

func (s *computeCmds) runManagedCreate(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		props, loadErr := s.loadProps()
		if loadErr != nil {
			return err
		}
		project = s.resolveProjectOnly(props)
		if project == "" {
			return err
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	mig := &compute.InstanceGroupManager{
		Name:             group,
		InstanceTemplate: resolveTemplateURL(project, s.flagManagedTemplate),
		TargetSize:       s.flagManagedSize,
	}

	if s.flagManagedCreateRegion != "" {
		op, err := svc.RegionInstanceGroupManagers.Insert(project, s.flagManagedCreateRegion, mig).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating managed instance group: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := s.inv.projectZone()
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
	fmt.Fprintf(s.inv.stdout, "Created managed instance group [%s].\n", group)
	return nil
}

func (s *computeCmds) runManagedDelete(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		props, loadErr := s.loadProps()
		if loadErr != nil {
			return err
		}
		project = s.resolveProjectOnly(props)
		if project == "" {
			return err
		}
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	if s.flagManagedDeleteRegion != "" {
		op, err := svc.RegionInstanceGroupManagers.Delete(project, s.flagManagedDeleteRegion, group).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting managed instance group: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Delete operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := s.inv.projectZone()
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
	fmt.Fprintf(s.inv.stdout, "Deleted managed instance group [%s].\n", group)
	return nil
}

func (s *computeCmds) runManagedUpdate(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	if s.flagManagedUpdateRegion != "" {
		mig, err := svc.RegionInstanceGroupManagers.Get(project, s.flagManagedUpdateRegion, group).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting MIG: %w", err)
		}
		if s.flagManagedUpdateTemplate != "" {
			mig.InstanceTemplate = resolveTemplateURL(project, s.flagManagedUpdateTemplate)
		}
		op, err := svc.RegionInstanceGroupManagers.Patch(project, s.flagManagedUpdateRegion, group, mig).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating MIG: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Update operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := s.inv.projectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	mig, err := svc.InstanceGroupManagers.Get(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting MIG: %w", err)
	}
	if s.flagManagedUpdateTemplate != "" {
		mig.InstanceTemplate = resolveTemplateURL(project, s.flagManagedUpdateTemplate)
	}
	op, err := svc.InstanceGroupManagers.Patch(project, zone, group, mig).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating MIG: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Updated managed instance group [%s].\n", group)
	return nil
}

func (s *computeCmds) runManagedSetTemplate(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	templateURL := resolveTemplateURL(project, s.flagManagedSetTemplateName)
	req := &compute.InstanceGroupManagersSetInstanceTemplateRequest{
		InstanceTemplate: templateURL,
	}

	if s.flagManagedSetTemplateRegion != "" {
		reqR := &compute.RegionInstanceGroupManagersSetTemplateRequest{
			InstanceTemplate: templateURL,
		}
		op, err := svc.RegionInstanceGroupManagers.SetInstanceTemplate(project, s.flagManagedSetTemplateRegion, group, reqR).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("setting instance template: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Set-template operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := s.inv.projectZone()
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
	fmt.Fprintf(s.inv.stdout, "Set instance template for [%s].\n", group)
	return nil
}

func (s *computeCmds) runManagedRecreateInstances(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	req := &compute.InstanceGroupManagersRecreateInstancesRequest{
		Instances: s.flagManagedRecreateInstances,
	}

	if s.flagManagedRecreateRegion != "" {
		reqR := &compute.RegionInstanceGroupManagersRecreateRequest{
			Instances: s.flagManagedRecreateInstances,
		}
		op, err := svc.RegionInstanceGroupManagers.RecreateInstances(project, s.flagManagedRecreateRegion, group, reqR).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("recreating instances: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Recreate operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := s.inv.projectZone()
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
	fmt.Fprintf(s.inv.stdout, "Recreated instances in [%s].\n", group)
	return nil
}

func (s *computeCmds) runManagedSetAutoscaling(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	autoscaler := &compute.Autoscaler{
		Name:   group + "-autoscaler",
		Target: "", // will be set below
		AutoscalingPolicy: &compute.AutoscalingPolicy{
			MinNumReplicas: s.flagAutoscalingMin,
			MaxNumReplicas: s.flagAutoscalingMax,
			CpuUtilization: &compute.AutoscalingPolicyCpuUtilization{
				UtilizationTarget: s.flagAutoscalingCPUTarget,
			},
		},
	}

	if s.flagAutoscalingRegion != "" {
		autoscaler.Target = fmt.Sprintf("projects/%s/regions/%s/instanceGroupManagers/%s", project, s.flagAutoscalingRegion, group)
		op, err := svc.RegionAutoscalers.Insert(project, s.flagAutoscalingRegion, autoscaler).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("setting autoscaling: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Autoscaling operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := s.inv.projectZone()
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
	fmt.Fprintf(s.inv.stdout, "Set autoscaling for [%s].\n", group)
	return nil
}

func (s *computeCmds) runManagedStopAutoscaling(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	autoscalerName := group + "-autoscaler"

	if s.flagStopAutoscalingRegion != "" {
		op, err := svc.RegionAutoscalers.Delete(project, s.flagStopAutoscalingRegion, autoscalerName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("stopping autoscaling: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Stop-autoscaling operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := s.inv.projectZone()
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
	fmt.Fprintf(s.inv.stdout, "Stopped autoscaling for [%s].\n", group)
	return nil
}
