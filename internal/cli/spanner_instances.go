package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner instances (#1210) ---

func (s *spannerCmds) initSpannerInstances() {
	all := []*cobra.Command{
		s.spannerInstCreateCmd, s.spannerInstDeleteCmd, s.spannerInstDescribeCmd, s.spannerInstListCmd,
		s.spannerInstUpdateCmd, s.spannerInstMoveCmd, s.spannerInstGetLocsCmd,
		s.spannerInstGetIamCmd, s.spannerInstSetIamCmd, s.spannerInstAddIamCmd, s.spannerInstRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSpInstFormat, "format", "", "Output format")
	}

	s.spannerInstCreateCmd.Flags().StringVar(&s.flagSpInstConfigFile, "config-file", "", "YAML/JSON file with the Instance body")
	s.spannerInstCreateCmd.Flags().StringVar(&s.flagSpInstConfig, "config", "", "Instance config (short name or full URI)")
	s.spannerInstCreateCmd.Flags().Int64Var(&s.flagSpInstNodes, "nodes", 0, "Number of nodes")
	s.spannerInstCreateCmd.Flags().Int64Var(&s.flagSpInstPU, "processing-units", 0, "Processing units")
	s.spannerInstCreateCmd.Flags().StringVar(&s.flagSpInstDisplayName, "display-name", "", "Display name")
	s.spannerInstCreateCmd.Flags().StringToStringVar(&s.flagSpInstLabels, "labels", nil,
		"Instance labels as KEY=VALUE pairs")

	s.spannerInstUpdateCmd.Flags().StringVar(&s.flagSpInstConfigFile, "config-file", "", "YAML/JSON file with the Instance body")
	s.spannerInstUpdateCmd.Flags().StringVar(&s.flagSpInstUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.spannerInstUpdateCmd.Flags().Int64Var(&s.flagSpInstNodes, "nodes", 0, "Number of nodes")
	s.spannerInstUpdateCmd.Flags().Int64Var(&s.flagSpInstPU, "processing-units", 0, "Processing units")
	s.spannerInstUpdateCmd.Flags().StringVar(&s.flagSpInstDisplayName, "display-name", "", "Display name")
	s.spannerInstUpdateCmd.Flags().StringToStringVar(&s.flagSpInstUpdateLabels, "update-labels", nil,
		"Set or update instance labels as KEY=VALUE pairs")
	s.spannerInstUpdateCmd.Flags().StringSliceVar(&s.flagSpInstRemoveLabels, "remove-labels", nil,
		"Comma-separated list of label keys to remove")
	s.spannerInstUpdateCmd.Flags().BoolVar(&s.flagSpInstClearLabels, "clear-labels", false,
		"Remove all labels from the instance (mutually exclusive with --update-labels/--remove-labels)")

	s.spannerInstMoveCmd.Flags().StringVar(&s.flagSpInstTargetCfg, "target-config", "", "Target instance configuration (required)")
	_ = s.spannerInstMoveCmd.MarkFlagRequired("target-config")

	s.spannerInstListCmd.Flags().StringVar(&s.flagSpInstFilter, "filter", "", "Server-side filter expression")
	s.spannerInstListCmd.Flags().Int64Var(&s.flagSpInstPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{s.spannerInstAddIamCmd, s.spannerInstRemoveIamCmd} {
		spIamMemberFlags(c, &s.flagSpInstIamMember, &s.flagSpInstIamRole,
			&s.flagSpInstIamCondExpr, &s.flagSpInstIamCondT, &s.flagSpInstIamCondD)
	}
	s.spannerInstRemoveIamCmd.Flags().BoolVar(&s.flagSpInstIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.spannerInstancesCmd.AddCommand(all...)
	s.spannerCmd.AddCommand(s.spannerInstancesCmd)
}

// spannerInstBuild constructs an Instance from either the config file or the
// short flags; short flags override file fields when set.
func (s *spannerCmds) spannerInstBuild(existing *spanner.Instance) *spanner.Instance {
	if existing == nil {
		existing = &spanner.Instance{}
	}
	if s.flagSpInstConfig != "" {
		cfg, err := s.spannerInstanceConfig(s.flagSpInstConfig)
		if err == nil {
			existing.Config = cfg
		}
	}
	if s.flagSpInstNodes > 0 {
		existing.NodeCount = s.flagSpInstNodes
	}
	if s.flagSpInstPU > 0 {
		existing.ProcessingUnits = s.flagSpInstPU
	}
	if s.flagSpInstDisplayName != "" {
		existing.DisplayName = s.flagSpInstDisplayName
	}
	return existing
}

func (s *spannerCmds) runSpInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerProject()
	if err != nil {
		return err
	}
	inst := &spanner.Instance{}
	if s.flagSpInstConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSpInstConfigFile, inst); err != nil {
			return err
		}
	}
	inst = s.spannerInstBuild(inst)
	if len(s.flagSpInstLabels) > 0 {
		if inst.Labels == nil {
			inst.Labels = map[string]string{}
		}
		for k, v := range s.flagSpInstLabels {
			inst.Labels[k] = v
		}
	}
	body := &spanner.CreateInstanceRequest{Instance: inst, InstanceId: args[0]}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSpInstFormat)
}

func (s *spannerCmds) runSpInstDelete(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted instance [%s].\n", args[0])
	return nil
}

func (s *spannerCmds) runSpInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSpInstFormat)
}

func (s *spannerCmds) runSpInstList(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.Instance
	pageToken := ""
	for {
		call := svc.Projects.Instances.List(parent).Context(ctx)
		if s.flagSpInstFilter != "" {
			call = call.Filter(s.flagSpInstFilter)
		}
		if s.flagSpInstPageSize > 0 {
			call = call.PageSize(s.flagSpInstPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagSpInstFormat)
}

func (s *spannerCmds) runSpInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	if s.flagSpInstClearLabels && (len(s.flagSpInstUpdateLabels) > 0 || len(s.flagSpInstRemoveLabels) > 0) {
		return fmt.Errorf("--clear-labels is mutually exclusive with --update-labels/--remove-labels")
	}
	inst := &spanner.Instance{}
	if s.flagSpInstConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSpInstConfigFile, inst); err != nil {
			return err
		}
	}
	inst = s.spannerInstBuild(inst)
	inst.Name = name

	labelsChanged := false
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	if s.flagSpInstClearLabels {
		inst.Labels = map[string]string{}
		inst.ForceSendFields = append(inst.ForceSendFields, "Labels")
		labelsChanged = true
	} else if len(s.flagSpInstUpdateLabels) > 0 || len(s.flagSpInstRemoveLabels) > 0 {
		existing, gerr := svc.Projects.Instances.Get(name).Context(ctx).Do()
		if gerr != nil {
			return fmt.Errorf("fetching existing instance for label edit: %w", gerr)
		}
		labels := map[string]string{}
		for k, v := range existing.Labels {
			labels[k] = v
		}
		for k, v := range s.flagSpInstUpdateLabels {
			labels[k] = v
		}
		for _, k := range s.flagSpInstRemoveLabels {
			delete(labels, k)
		}
		inst.Labels = labels
		if len(labels) == 0 {
			inst.ForceSendFields = append(inst.ForceSendFields, "Labels")
		}
		labelsChanged = true
	}

	mask := s.flagSpInstUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(inst))
		if labelsChanged && !maskHasField(mask, "labels") {
			if mask == "" {
				mask = "labels"
			} else {
				mask += ",labels"
			}
		}
	}
	body := &spanner.UpdateInstanceRequest{Instance: inst, FieldMask: mask}
	op, err := svc.Projects.Instances.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagSpInstFormat)
}

// maskHasField reports whether comma-separated `mask` already contains `field`.
func maskHasField(mask, field string) bool {
	if mask == "" {
		return false
	}
	for _, part := range strings.Split(mask, ",") {
		if strings.TrimSpace(part) == field {
			return true
		}
	}
	return false
}

func (s *spannerCmds) runSpInstMove(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	target, err := s.spannerInstanceConfig(s.flagSpInstTargetCfg)
	if err != nil {
		return err
	}
	body := &spanner.MoveInstanceRequest{TargetConfig: target}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Move(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("moving instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Move instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSpInstFormat)
}

func (s *spannerCmds) runSpInstGetLocations(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	inst, err := svc.Projects.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	if inst.Config == "" {
		return fmt.Errorf("instance [%s] has no config", args[0])
	}
	cfg, err := svc.Projects.InstanceConfigs.Get(inst.Config).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, cfg.Replicas, s.flagSpInstFormat)
}

func (s *spannerCmds) runSpInstGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagSpInstFormat)
}

func (s *spannerCmds) runSpInstSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	policy := &spanner.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("instance [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpInstFormat)
}

func (s *spannerCmds) runSpInstAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	spIamAddBinding(policy, s.flagSpInstIamRole, s.flagSpInstIamMember,
		spIamBuildCondition(s.flagSpInstIamCondExpr, s.flagSpInstIamCondT, s.flagSpInstIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("instance [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpInstFormat)
}

func (s *spannerCmds) runSpInstRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstance(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !spIamRemoveBinding(policy, s.flagSpInstIamRole, s.flagSpInstIamMember,
		spIamBuildCondition(s.flagSpInstIamCondExpr, s.flagSpInstIamCondT, s.flagSpInstIamCondD), s.flagSpInstIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagSpInstIamRole, s.flagSpInstIamMember)
	}
	updated, err := svc.Projects.Instances.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("instance [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpInstFormat)
}
