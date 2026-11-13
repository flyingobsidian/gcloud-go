package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai index-endpoints (#1454) ---

func (s *aiCmds) initAiIndexEndpoints() {
	all := []*cobra.Command{
		s.aiIECreateCmd, s.aiIEDeleteCmd, s.aiIEDeployIndexCmd, s.aiIEDescribeCmd, s.aiIEListCmd,
		s.aiIEMutateDeployedIndexCmd, s.aiIEUndeployIndexCmd, s.aiIEUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIIERegion, "region", "", "Region where the index endpoint lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAIIEFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.aiIECreateCmd, s.aiIEDeployIndexCmd, s.aiIEMutateDeployedIndexCmd,
		s.aiIEUndeployIndexCmd, s.aiIEUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagAIIEConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.aiIEUpdateCmd.Flags().StringVar(&s.flagAIIEUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")
	s.aiIEListCmd.Flags().StringVar(&s.flagAIIEFilter, "filter", "", "Server-side filter expression")
	s.aiIEListCmd.Flags().Int64Var(&s.flagAIIEPageSize, "page-size", 0, "Maximum results per page")
	s.aiIEListCmd.Flags().StringVar(&s.flagAIIEReadMask, "read-mask", "", "Field mask for reads")

	s.aiIECmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiIECmd)
}

func (s *aiCmds) aiIEParent() (string, error) { return s.aiParent(s.flagAIIERegion) }

func (s *aiCmds) aiIEName(id string) (string, error) {
	parent, err := s.aiIEParent()
	if err != nil {
		return "", err
	}
	return aiChild("indexEndpoints", id, parent), nil
}

func (s *aiCmds) runAIIECreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiIEParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1IndexEndpoint{}
	if err := loadYAMLOrJSONInto(s.flagAIIEConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating index endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Create request issued (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIIEFormat)
}

func (s *aiCmds) runAIIEDelete(cmd *cobra.Command, args []string) error {
	name, err := s.aiIEName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting index endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Delete request issued for index endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIIEFormat)
}

func (s *aiCmds) runAIIEDeployIndex(cmd *cobra.Command, args []string) error {
	name, err := s.aiIEName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DeployIndexRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIIEConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.DeployIndex(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deploying index: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Deploy request issued for index endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIIEFormat)
}

func (s *aiCmds) runAIIEDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiIEName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIERegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.IndexEndpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing index endpoint: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIIEFormat)
}

func (s *aiCmds) runAIIEList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiIEParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIERegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1IndexEndpoint
	pageToken := ""
	for {
		call := svc.Projects.Locations.IndexEndpoints.List(parent).Context(ctx)
		if s.flagAIIEFilter != "" {
			call = call.Filter(s.flagAIIEFilter)
		}
		if s.flagAIIEPageSize > 0 {
			call = call.PageSize(s.flagAIIEPageSize)
		}
		if s.flagAIIEReadMask != "" {
			call = call.ReadMask(s.flagAIIEReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing index endpoints: %w", err)
		}
		all = append(all, resp.IndexEndpoints...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIIEFormat)
}

func (s *aiCmds) runAIIEMutateDeployedIndex(cmd *cobra.Command, args []string) error {
	name, err := s.aiIEName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DeployedIndex{}
	if err := loadYAMLOrJSONInto(s.flagAIIEConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.MutateDeployedIndex(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("mutating deployed index: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Mutate request issued for index endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIIEFormat)
}

func (s *aiCmds) runAIIEUndeployIndex(cmd *cobra.Command, args []string) error {
	name, err := s.aiIEName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1UndeployIndexRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIIEConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.UndeployIndex(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeploying index: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Undeploy request issued for index endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIIEFormat)
}

func (s *aiCmds) runAIIEUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.aiIEName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1IndexEndpoint{}
	if err := loadYAMLOrJSONInto(s.flagAIIEConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAIIEUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIERegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.IndexEndpoints.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating index endpoint: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIIEFormat)
}
