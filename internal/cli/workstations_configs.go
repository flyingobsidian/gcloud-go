package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	workstations "google.golang.org/api/workstations/v1"
)

// --- gcloud workstations configs (#1260) ---

func (s *workstationsCmds) initWorkstationsConfigs() {
	all := []*cobra.Command{
		s.workstationsCfgCreateCmd, s.workstationsCfgDeleteCmd,
		s.workstationsCfgDescribeCmd, s.workstationsCfgListCmd,
		s.workstationsCfgUpdateCmd, s.workstationsCfgGetIamCmd,
		s.workstationsCfgSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagWsCfgLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagWsCfgCluster, "cluster", "", "Workstation cluster (required)")
		_ = c.MarkFlagRequired("cluster")
		c.Flags().StringVar(&s.flagWsCfgFormat, "format", "", "Output format")
	}
	s.workstationsCfgCreateCmd.Flags().StringVar(&s.flagWsCfgConfigFile, "config-file", "", "YAML/JSON file with the WorkstationConfig body (required)")
	_ = s.workstationsCfgCreateCmd.MarkFlagRequired("config-file")
	s.workstationsCfgListCmd.Flags().Int64Var(&s.flagWsCfgPageSize, "page-size", 0, "Maximum results per page")
	s.workstationsCfgUpdateCmd.Flags().StringVar(&s.flagWsCfgConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.workstationsCfgUpdateCmd.MarkFlagRequired("config-file")
	s.workstationsCfgUpdateCmd.Flags().StringVar(&s.flagWsCfgUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.workstationsConfigsCmd.AddCommand(all...)
	s.workstationsCmd.AddCommand(s.workstationsConfigsCmd)
}

func (s *workstationsCmds) runWsCfgCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.wsClusterName(s.flagWsCfgLocation, s.flagWsCfgCluster)
	if err != nil {
		return err
	}
	body := &workstations.WorkstationConfig{}
	if err := loadYAMLOrJSONInto(s.flagWsCfgConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.Create(parent, body).WorkstationConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workstation configuration: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create workstation configuration [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagWsCfgFormat)
}

func (s *workstationsCmds) runWsCfgDelete(cmd *cobra.Command, args []string) error {
	name, err := s.wsConfigName(s.flagWsCfgLocation, s.flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workstation configuration: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete workstation configuration [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagWsCfgFormat)
}

func (s *workstationsCmds) runWsCfgDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.wsConfigName(s.flagWsCfgLocation, s.flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workstation configuration: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagWsCfgFormat)
}

func (s *workstationsCmds) runWsCfgList(cmd *cobra.Command, args []string) error {
	parent, err := s.wsClusterName(s.flagWsCfgLocation, s.flagWsCfgCluster)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	var all []*workstations.WorkstationConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.List(parent).Context(ctx)
		if s.flagWsCfgPageSize > 0 {
			call = call.PageSize(s.flagWsCfgPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workstation configurations: %w", err)
		}
		all = append(all, resp.WorkstationConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagWsCfgFormat)
}

func (s *workstationsCmds) runWsCfgUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.wsConfigName(s.flagWsCfgLocation, s.flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	body := &workstations.WorkstationConfig{}
	if err := loadYAMLOrJSONInto(s.flagWsCfgConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagWsCfgUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating workstation configuration: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update workstation configuration [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagWsCfgFormat)
}

func (s *workstationsCmds) runWsCfgGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.wsConfigName(s.flagWsCfgLocation, s.flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagWsCfgFormat)
}

func (s *workstationsCmds) runWsCfgSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.wsConfigName(s.flagWsCfgLocation, s.flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	policy := &workstations.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.SetIamPolicy(name, &workstations.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for workstation configuration [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagWsCfgFormat)
}
