package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	aiplatformbeta "google.golang.org/api/aiplatform/v1beta1"
)

// --- gcloud ai model-garden (#1456) ---

func (s *aiCmds) initAiModelGarden() {
	all := []*cobra.Command{s.aiMGListCmd, s.aiMGDescribeCmd}
	for _, c := range all {
		// Model Garden is not per-region, but --region is required so we can
		// route through a regional aiplatform endpoint.
		c.Flags().StringVar(&s.flagAIMGRegion, "region", "", "Region for endpoint routing (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAIMGFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagAIMGPublisher, "publisher", "google",
			"Publisher whose models to browse (default: google)")
		c.Flags().StringVar(&s.flagAIMGLanguage, "language-code", "",
			"BCP-47 language code for model text metadata (default: en)")
		c.Flags().StringVar(&s.flagAIMGView, "view", "",
			"PublisherModel view (PUBLISHER_MODEL_VIEW_BASIC | PUBLISHER_MODEL_VIEW_FULL)")
	}
	s.aiMGListCmd.Flags().StringVar(&s.flagAIMGFilter, "filter", "", "Server-side filter expression")
	s.aiMGListCmd.Flags().StringVar(&s.flagAIMGOrderBy, "order-by", "", "Order-by expression")
	s.aiMGListCmd.Flags().Int64Var(&s.flagAIMGPageSize, "page-size", 0, "Maximum results per page")
	s.aiMGListCmd.Flags().BoolVar(&s.flagAIMGListAll, "list-all-versions", false,
		"List all versions of each publisher model")

	s.aiMGModelsCmd.AddCommand(all...)
	s.aiMGCmd.AddCommand(s.aiMGModelsCmd)
	s.aiCmd.AddCommand(s.aiMGCmd)
}

func (s *aiCmds) runAIMGList(cmd *cobra.Command, args []string) error {
	if s.flagAIMGPublisher == "" {
		s.flagAIMGPublisher = "google"
	}
	if s.flagAIMGRegion == "" {
		return fmt.Errorf("--region is required")
	}
	parent := fmt.Sprintf("publishers/%s", s.flagAIMGPublisher)
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatformBeta(ctx, s.flagAIMGRegion)
	if err != nil {
		return err
	}
	var all []*aiplatformbeta.GoogleCloudAiplatformV1beta1PublisherModel
	pageToken := ""
	for {
		call := svc.Publishers.Models.List(parent).Context(ctx)
		if s.flagAIMGFilter != "" {
			call = call.Filter(s.flagAIMGFilter)
		}
		if s.flagAIMGOrderBy != "" {
			call = call.OrderBy(s.flagAIMGOrderBy)
		}
		if s.flagAIMGPageSize > 0 {
			call = call.PageSize(s.flagAIMGPageSize)
		}
		if s.flagAIMGListAll {
			call = call.ListAllVersions(true)
		}
		if s.flagAIMGLanguage != "" {
			call = call.LanguageCode(s.flagAIMGLanguage)
		}
		if s.flagAIMGView != "" {
			call = call.View(s.flagAIMGView)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing publisher models: %w", err)
		}
		all = append(all, resp.PublisherModels...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIMGFormat)
}

func (s *aiCmds) runAIMGDescribe(cmd *cobra.Command, args []string) error {
	if s.flagAIMGPublisher == "" {
		s.flagAIMGPublisher = "google"
	}
	if s.flagAIMGRegion == "" {
		return fmt.Errorf("--region is required")
	}
	name := args[0]
	if !strings.HasPrefix(name, "publishers/") {
		name = fmt.Sprintf("publishers/%s/models/%s", s.flagAIMGPublisher, name)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatformBeta(ctx, s.flagAIMGRegion)
	if err != nil {
		return err
	}
	call := svc.Publishers.Models.Get(name).Context(ctx)
	if s.flagAIMGLanguage != "" {
		call = call.LanguageCode(s.flagAIMGLanguage)
	}
	if s.flagAIMGView != "" {
		call = call.View(s.flagAIMGView)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing publisher model: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIMGFormat)
}
