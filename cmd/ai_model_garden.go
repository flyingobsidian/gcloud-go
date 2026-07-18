package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatformbeta "google.golang.org/api/aiplatform/v1beta1"
)

// --- gcloud ai model-garden (#1456) ---

var aiMGCmd = &cobra.Command{Use: "model-garden", Short: "Interact with Vertex AI Model Garden"}
var aiMGModelsCmd = &cobra.Command{Use: "models", Short: "Browse Model Garden publisher models"}

var (
	flagAIMGRegion    string
	flagAIMGFormat    string
	flagAIMGPublisher string
	flagAIMGFilter    string
	flagAIMGOrderBy   string
	flagAIMGPageSize  int64
	flagAIMGListAll   bool
	flagAIMGView      string
	flagAIMGLanguage  string
)

var (
	aiMGListCmd = &cobra.Command{
		Use: "list", Short: "List publisher models in Model Garden",
		Args: cobra.NoArgs, RunE: runAIMGList,
	}
	aiMGDescribeCmd = &cobra.Command{
		Use: "describe MODEL", Short: "Describe a publisher model in Model Garden",
		Args: cobra.ExactArgs(1), RunE: runAIMGDescribe,
	}
)

func init() {
	all := []*cobra.Command{aiMGListCmd, aiMGDescribeCmd}
	for _, c := range all {
		// Model Garden is not per-region, but --region is required so we can
		// route through a regional aiplatform endpoint.
		c.Flags().StringVar(&flagAIMGRegion, "region", "", "Region for endpoint routing (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAIMGFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagAIMGPublisher, "publisher", "google",
			"Publisher whose models to browse (default: google)")
		c.Flags().StringVar(&flagAIMGLanguage, "language-code", "",
			"BCP-47 language code for model text metadata (default: en)")
		c.Flags().StringVar(&flagAIMGView, "view", "",
			"PublisherModel view (PUBLISHER_MODEL_VIEW_BASIC | PUBLISHER_MODEL_VIEW_FULL)")
	}
	aiMGListCmd.Flags().StringVar(&flagAIMGFilter, "filter", "", "Server-side filter expression")
	aiMGListCmd.Flags().StringVar(&flagAIMGOrderBy, "order-by", "", "Order-by expression")
	aiMGListCmd.Flags().Int64Var(&flagAIMGPageSize, "page-size", 0, "Maximum results per page")
	aiMGListCmd.Flags().BoolVar(&flagAIMGListAll, "list-all-versions", false,
		"List all versions of each publisher model")

	aiMGModelsCmd.AddCommand(all...)
	aiMGCmd.AddCommand(aiMGModelsCmd)
	aiCmd.AddCommand(aiMGCmd)
}

func runAIMGList(cmd *cobra.Command, args []string) error {
	if flagAIMGPublisher == "" {
		flagAIMGPublisher = "google"
	}
	if flagAIMGRegion == "" {
		return fmt.Errorf("--region is required")
	}
	parent := fmt.Sprintf("publishers/%s", flagAIMGPublisher)
	ctx := context.Background()
	svc, err := gcp.AIPlatformBetaService(ctx, flagAccount, flagAIMGRegion)
	if err != nil {
		return err
	}
	var all []*aiplatformbeta.GoogleCloudAiplatformV1beta1PublisherModel
	pageToken := ""
	for {
		call := svc.Publishers.Models.List(parent).Context(ctx)
		if flagAIMGFilter != "" {
			call = call.Filter(flagAIMGFilter)
		}
		if flagAIMGOrderBy != "" {
			call = call.OrderBy(flagAIMGOrderBy)
		}
		if flagAIMGPageSize > 0 {
			call = call.PageSize(flagAIMGPageSize)
		}
		if flagAIMGListAll {
			call = call.ListAllVersions(true)
		}
		if flagAIMGLanguage != "" {
			call = call.LanguageCode(flagAIMGLanguage)
		}
		if flagAIMGView != "" {
			call = call.View(flagAIMGView)
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
	return emitFormatted(all, flagAIMGFormat)
}

func runAIMGDescribe(cmd *cobra.Command, args []string) error {
	if flagAIMGPublisher == "" {
		flagAIMGPublisher = "google"
	}
	if flagAIMGRegion == "" {
		return fmt.Errorf("--region is required")
	}
	name := args[0]
	if !strings.HasPrefix(name, "publishers/") {
		name = fmt.Sprintf("publishers/%s/models/%s", flagAIMGPublisher, name)
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformBetaService(ctx, flagAccount, flagAIMGRegion)
	if err != nil {
		return err
	}
	call := svc.Publishers.Models.Get(name).Context(ctx)
	if flagAIMGLanguage != "" {
		call = call.LanguageCode(flagAIMGLanguage)
	}
	if flagAIMGView != "" {
		call = call.View(flagAIMGView)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing publisher model: %w", err)
	}
	return emitFormatted(got, flagAIMGFormat)
}
