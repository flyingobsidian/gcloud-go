package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai endpoints (#1452) ---

func (s *aiCmds) initAiEndpoints() {
	all := []*cobra.Command{
		s.aiEPCreateCmd, s.aiEPDeleteCmd, s.aiEPDeployModelCmd, s.aiEPUndeployModelCmd,
		s.aiEPDescribeCmd, s.aiEPListCmd, s.aiEPUpdateCmd,
		s.aiEPPredictCmd, s.aiEPRawPredictCmd, s.aiEPExplainCmd,
		s.aiEPDirectPredictCmd, s.aiEPDirectRawPredictCmd,
		s.aiEPStreamDirectPredictCmd, s.aiEPStreamDirectRawPredictCmd, s.aiEPStreamRawPredictCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIEPRegion, "region", "", "Region where the endpoint lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAIEPFormat, "format", "", "Output format")
	}

	for _, c := range []*cobra.Command{
		s.aiEPCreateCmd, s.aiEPDeployModelCmd, s.aiEPUndeployModelCmd, s.aiEPUpdateCmd,
		s.aiEPPredictCmd, s.aiEPRawPredictCmd, s.aiEPExplainCmd,
		s.aiEPDirectPredictCmd, s.aiEPDirectRawPredictCmd,
		s.aiEPStreamDirectPredictCmd, s.aiEPStreamDirectRawPredictCmd, s.aiEPStreamRawPredictCmd,
	} {
		c.Flags().StringVar(&s.flagAIEPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}

	s.aiEPCreateCmd.Flags().StringVar(&s.flagAIEPEndpointID, "endpoint-id", "",
		"Optional caller-supplied endpoint ID (immutable)")
	s.aiEPUpdateCmd.Flags().StringVar(&s.flagAIEPUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")

	s.aiEPListCmd.Flags().StringVar(&s.flagAIEPFilter, "filter", "", "Server-side filter expression")
	s.aiEPListCmd.Flags().StringVar(&s.flagAIEPOrderBy, "order-by", "", "Order-by expression")
	s.aiEPListCmd.Flags().Int64Var(&s.flagAIEPPageSize, "page-size", 0, "Maximum results per page")
	s.aiEPListCmd.Flags().StringVar(&s.flagAIEPReadMask, "read-mask", "", "Field mask for reads")

	s.aiEndpointsCmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiEndpointsCmd)
}

func (s *aiCmds) aiEPParent() (string, error) { return s.aiParent(s.flagAIEPRegion) }

func (s *aiCmds) aiEPName(id string) (string, error) {
	parent, err := s.aiEPParent()
	if err != nil {
		return "", err
	}
	return aiChild("endpoints", id, parent), nil
}

func (s *aiCmds) runAIEPCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiEPParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Endpoint{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Endpoints.Create(parent, body).Context(ctx)
	if s.flagAIEPEndpointID != "" {
		call = call.EndpointId(s.flagAIEPEndpointID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Create request issued (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPDelete(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting endpoint: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Delete request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPDeployModel(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DeployModelRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.DeployModel(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deploying model: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Deploy request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPUndeployModel(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1UndeployModelRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.UndeployModel(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeploying model: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Undeploy request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Endpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiEPParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Endpoint
	pageToken := ""
	for {
		call := svc.Projects.Locations.Endpoints.List(parent).Context(ctx)
		if s.flagAIEPFilter != "" {
			call = call.Filter(s.flagAIEPFilter)
		}
		if s.flagAIEPOrderBy != "" {
			call = call.OrderBy(s.flagAIEPOrderBy)
		}
		if s.flagAIEPPageSize > 0 {
			call = call.PageSize(s.flagAIEPPageSize)
		}
		if s.flagAIEPReadMask != "" {
			call = call.ReadMask(s.flagAIEPReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing endpoints: %w", err)
		}
		all = append(all, resp.Endpoints...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Endpoint{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAIEPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Endpoints.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating endpoint: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPPredict(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1PredictRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.Predict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running predict: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPRawPredict(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1RawPredictRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.RawPredict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running raw-predict: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPExplain(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1ExplainRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.Explain(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running explain: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPDirectPredict(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DirectPredictRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.DirectPredict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running direct-predict: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPDirectRawPredict(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DirectRawPredictRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.DirectRawPredict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running direct-raw-predict: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIEPFormat)
}

func (s *aiCmds) runAIEPStreamRawPredict(cmd *cobra.Command, args []string) error {
	name, err := s.aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1StreamRawPredictRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.StreamRawPredict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running stream-raw-predict: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIEPFormat)
}
