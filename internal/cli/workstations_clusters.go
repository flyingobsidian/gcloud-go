package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	workstations "google.golang.org/api/workstations/v1"
)

// --- gcloud workstations clusters (#1259) ---

func (s *workstationsCmds) initWorkstationsClusters() {
	all := []*cobra.Command{
		s.workstationsClCreateCmd, s.workstationsClDeleteCmd,
		s.workstationsClDescribeCmd, s.workstationsClListCmd,
		s.workstationsClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagWsClLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagWsClFormat, "format", "", "Output format")
	}
	s.workstationsClCreateCmd.Flags().StringVar(&s.flagWsClConfigFile, "config-file", "", "YAML/JSON file with the WorkstationCluster body (required)")
	_ = s.workstationsClCreateCmd.MarkFlagRequired("config-file")
	s.workstationsClListCmd.Flags().Int64Var(&s.flagWsClPageSize, "page-size", 0, "Maximum results per page")
	s.workstationsClUpdateCmd.Flags().StringVar(&s.flagWsClConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.workstationsClUpdateCmd.MarkFlagRequired("config-file")
	s.workstationsClUpdateCmd.Flags().StringVar(&s.flagWsClUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.workstationsClustersCmd.AddCommand(all...)
	s.workstationsCmd.AddCommand(s.workstationsClustersCmd)
}

func (s *workstationsCmds) runWsClCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.wsLocationParent(s.flagWsClLocation)
	if err != nil {
		return err
	}
	body := &workstations.WorkstationCluster{}
	if err := loadYAMLOrJSONInto(s.flagWsClConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkstationClusters.Create(parent, body).WorkstationClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workstation cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create workstation cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagWsClFormat)
}

func (s *workstationsCmds) runWsClDelete(cmd *cobra.Command, args []string) error {
	name, err := s.wsClusterName(s.flagWsClLocation, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkstationClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workstation cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete workstation cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagWsClFormat)
}

func (s *workstationsCmds) runWsClDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.wsClusterName(s.flagWsClLocation, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkstationClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workstation cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagWsClFormat)
}

func (s *workstationsCmds) runWsClList(cmd *cobra.Command, args []string) error {
	parent, err := s.wsLocationParent(s.flagWsClLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	var all []*workstations.WorkstationCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkstationClusters.List(parent).Context(ctx)
		if s.flagWsClPageSize > 0 {
			call = call.PageSize(s.flagWsClPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workstation clusters: %w", err)
		}
		all = append(all, resp.WorkstationClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagWsClFormat)
}

func (s *workstationsCmds) runWsClUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.wsClusterName(s.flagWsClLocation, args[0])
	if err != nil {
		return err
	}
	body := &workstations.WorkstationCluster{}
	if err := loadYAMLOrJSONInto(s.flagWsClConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagWsClUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Workstations(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.WorkstationClusters.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating workstation cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update workstation cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagWsClFormat)
}
