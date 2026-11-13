package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy automation-runs (#1524) ---

func (s *deployCmds) initDeployAutomationRuns() {
	all := []*cobra.Command{s.deployARCancelCmd, s.deployARDescribeCmd, s.deployARListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployARRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployARPipeline, "delivery-pipeline", "",
			"Delivery pipeline (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&s.flagDeployARFormat, "format", "", "Output format")
	}
	s.deployARListCmd.Flags().Int64Var(&s.flagDeployARPageSize, "page-size", 0, "Maximum results per page")

	s.deployARCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployARCmd)
}

func (s *deployCmds) deployARParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return deployChild("deliveryPipelines", s.flagDeployARPipeline, deployLocationParent(project, s.flagDeployARRegion)), nil
}

func (s *deployCmds) deployARName(id string) (string, error) {
	parent, err := s.deployARParent()
	if err != nil {
		return "", err
	}
	return deployChild("automationRuns", id, parent), nil
}

func (s *deployCmds) runDeployARCancel(cmd *cobra.Command, args []string) error {
	name, err := s.deployARName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.DeliveryPipelines.AutomationRuns.Cancel(name, &clouddeploy.CancelAutomationRunRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling automation run: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled automation run [%s].\n", args[0])
	return nil
}

func (s *deployCmds) runDeployARDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployARName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.AutomationRuns.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing automation run: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployARFormat)
}

func (s *deployCmds) runDeployARList(cmd *cobra.Command, args []string) error {
	parent, err := s.deployARParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	var all []*clouddeploy.AutomationRun
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.AutomationRuns.List(parent).Context(ctx)
		if s.flagDeployARPageSize > 0 {
			call = call.PageSize(s.flagDeployARPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing automation runs: %w", err)
		}
		all = append(all, resp.AutomationRuns...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDeployARFormat)
}
