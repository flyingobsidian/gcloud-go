package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai persistent-resources (#1460) ---

func (s *aiCmds) initAiPersistentResources() {
	all := []*cobra.Command{
		s.aiPRCreateCmd, s.aiPRDeleteCmd, s.aiPRDescribeCmd, s.aiPRListCmd, s.aiPRRebootCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIPRRegion, "region", "", "Region where the persistent resource lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAIPRFormat, "format", "", "Output format")
	}
	s.aiPRCreateCmd.Flags().StringVar(&s.flagAIPRConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the PersistentResource body (required)")
	_ = s.aiPRCreateCmd.MarkFlagRequired("config-file")
	s.aiPRCreateCmd.Flags().StringVar(&s.flagAIPRPRId, "persistent-resource-id", "",
		"Caller-supplied persistent resource ID (required)")
	_ = s.aiPRCreateCmd.MarkFlagRequired("persistent-resource-id")
	s.aiPRListCmd.Flags().Int64Var(&s.flagAIPRPageSize, "page-size", 0, "Maximum results per page")

	s.aiPRCmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiPRCmd)
}

func (s *aiCmds) aiPRParent() (string, error) { return s.aiParent(s.flagAIPRRegion) }

func (s *aiCmds) aiPRName(id string) (string, error) {
	parent, err := s.aiPRParent()
	if err != nil {
		return "", err
	}
	return aiChild("persistentResources", id, parent), nil
}

func (s *aiCmds) runAIPRCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiPRParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1PersistentResource{}
	if err := loadYAMLOrJSONInto(s.flagAIPRConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIPRRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PersistentResources.Create(parent, body).
		PersistentResourceId(s.flagAIPRPRId).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating persistent resource: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Create request issued for persistent resource [%s] (operation: %s).\n", s.flagAIPRPRId, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIPRFormat)
}

func (s *aiCmds) runAIPRDelete(cmd *cobra.Command, args []string) error {
	name, err := s.aiPRName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIPRRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PersistentResources.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting persistent resource: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Delete request issued for persistent resource [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIPRFormat)
}

func (s *aiCmds) runAIPRDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiPRName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIPRRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PersistentResources.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing persistent resource: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPRFormat)
}

func (s *aiCmds) runAIPRList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiPRParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIPRRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1PersistentResource
	pageToken := ""
	for {
		call := svc.Projects.Locations.PersistentResources.List(parent).Context(ctx)
		if s.flagAIPRPageSize > 0 {
			call = call.PageSize(s.flagAIPRPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing persistent resources: %w", err)
		}
		all = append(all, resp.PersistentResources...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIPRFormat)
}

func (s *aiCmds) runAIPRReboot(cmd *cobra.Command, args []string) error {
	name, err := s.aiPRName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIPRRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PersistentResources.Reboot(name,
		&aiplatform.GoogleCloudAiplatformV1RebootPersistentResourceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rebooting persistent resource: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Reboot request issued for persistent resource [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIPRFormat)
}
