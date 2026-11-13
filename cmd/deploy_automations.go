package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy automations (#1525) ---

func (s *deployCmds) initDeployAutomations() {
	all := []*cobra.Command{s.deployAutoDeleteCmd, s.deployAutoDescribeCmd, s.deployAutoExportCmd, s.deployAutoListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDeployAutoRegion, "region", "", "Cloud Deploy region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDeployAutoPipeline, "delivery-pipeline", "",
			"Delivery pipeline (required)")
		_ = c.MarkFlagRequired("delivery-pipeline")
		c.Flags().StringVar(&s.flagDeployAutoFormat, "format", "", "Output format")
	}
	s.deployAutoListCmd.Flags().Int64Var(&s.flagDeployAutoPageSize, "page-size", 0, "Maximum results per page")

	s.deployAutoCmd.AddCommand(all...)
	s.deployCmd.AddCommand(s.deployAutoCmd)
}

func (s *deployCmds) deployAutoParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return deployChild("deliveryPipelines", s.flagDeployAutoPipeline, deployLocationParent(project, s.flagDeployAutoRegion)), nil
}

func (s *deployCmds) deployAutoName(id string) (string, error) {
	parent, err := s.deployAutoParent()
	if err != nil {
		return "", err
	}
	return deployChild("automations", id, parent), nil
}

func (s *deployCmds) runDeployAutoDelete(cmd *cobra.Command, args []string) error {
	name, err := s.deployAutoName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DeliveryPipelines.Automations.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting automation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for automation [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDeployAutoFormat)
}

func (s *deployCmds) runDeployAutoDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.deployAutoName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Automations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing automation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDeployAutoFormat)
}

func (s *deployCmds) runDeployAutoExport(cmd *cobra.Command, args []string) error {
	name, err := s.deployAutoName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DeliveryPipelines.Automations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting automation: %w", err)
	}
	format := s.flagDeployAutoFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, got, format)
}

func (s *deployCmds) runDeployAutoList(cmd *cobra.Command, args []string) error {
	parent, err := s.deployAutoParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudDeploy(ctx)
	if err != nil {
		return err
	}
	var all []*clouddeploy.Automation
	pageToken := ""
	for {
		call := svc.Projects.Locations.DeliveryPipelines.Automations.List(parent).Context(ctx)
		if s.flagDeployAutoPageSize > 0 {
			call = call.PageSize(s.flagDeployAutoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing automations: %w", err)
		}
		all = append(all, resp.Automations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDeployAutoFormat)
}
