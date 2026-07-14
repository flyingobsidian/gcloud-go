package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	recommender "google.golang.org/api/recommender/v1"
)

// --- gcloud recommender (#379) ---

var recommenderCmd = &cobra.Command{Use: "recommender", Short: "Manage Cloud Recommender"}

var (
	flagRecLocation    string
	flagRecInsightType string
	flagRecRecommender string
	flagRecEtag        string
	flagRecFormat      string
	flagRecConfigFile  string
	flagRecStateMeta   []string
	flagRecUpdateMask  string
)

// --- insights ---

var recInsightsCmd = &cobra.Command{Use: "insights", Short: "Manage recommender insights"}

var (
	recInsDescribeCmd = &cobra.Command{
		Use: "describe INSIGHT", Short: "Describe an insight",
		Args: cobra.ExactArgs(1), RunE: runRecInsDescribe,
	}
	recInsListCmd = &cobra.Command{
		Use: "list", Short: "List insights",
		Args: cobra.NoArgs, RunE: runRecInsList,
	}
	recInsMarkAcceptedCmd = &cobra.Command{
		Use: "mark-accepted INSIGHT", Short: "Mark an insight as accepted",
		Args: cobra.ExactArgs(1), RunE: runRecInsMark("accepted"),
	}
	recInsMarkActiveCmd = &cobra.Command{
		Use: "mark-active INSIGHT", Short: "Mark an insight as active",
		Args: cobra.ExactArgs(1), RunE: runRecInsMark("active"),
	}
	recInsMarkDismissedCmd = &cobra.Command{
		Use: "mark-dismissed INSIGHT", Short: "Mark an insight as dismissed",
		Args: cobra.ExactArgs(1), RunE: runRecInsMark("dismissed"),
	}
)

// --- recommendations ---

var recRecommendationsCmd = &cobra.Command{Use: "recommendations", Short: "Manage recommendations"}

var (
	recRecDescribeCmd = &cobra.Command{
		Use: "describe RECOMMENDATION", Short: "Describe a recommendation",
		Args: cobra.ExactArgs(1), RunE: runRecRecDescribe,
	}
	recRecListCmd = &cobra.Command{
		Use: "list", Short: "List recommendations",
		Args: cobra.NoArgs, RunE: runRecRecList,
	}
	recRecMarkActiveCmd = &cobra.Command{
		Use: "mark-active RECOMMENDATION", Short: "Mark a recommendation as active",
		Args: cobra.ExactArgs(1), RunE: runRecRecMark("active"),
	}
	recRecMarkClaimedCmd = &cobra.Command{
		Use: "mark-claimed RECOMMENDATION", Short: "Mark a recommendation as claimed",
		Args: cobra.ExactArgs(1), RunE: runRecRecMark("claimed"),
	}
	recRecMarkDismissedCmd = &cobra.Command{
		Use: "mark-dismissed RECOMMENDATION", Short: "Mark a recommendation as dismissed",
		Args: cobra.ExactArgs(1), RunE: runRecRecMark("dismissed"),
	}
	recRecMarkFailedCmd = &cobra.Command{
		Use: "mark-failed RECOMMENDATION", Short: "Mark a recommendation as failed",
		Args: cobra.ExactArgs(1), RunE: runRecRecMark("failed"),
	}
	recRecMarkSucceededCmd = &cobra.Command{
		Use: "mark-succeeded RECOMMENDATION", Short: "Mark a recommendation as succeeded",
		Args: cobra.ExactArgs(1), RunE: runRecRecMark("succeeded"),
	}
)

// --- insight-type-config ---

var recInsightTypeConfigCmd = &cobra.Command{Use: "insight-type-config", Short: "Manage insight-type configuration"}

var (
	recITCDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe insight-type config",
		Args: cobra.NoArgs, RunE: runRecITCDescribe,
	}
	recITCUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update insight-type config from --config-file",
		Args: cobra.NoArgs, RunE: runRecITCUpdate,
	}
)

// --- recommender-config ---

var recRecommenderConfigCmd = &cobra.Command{Use: "recommender-config", Short: "Manage recommender configuration"}

var (
	recRCDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe recommender config",
		Args: cobra.NoArgs, RunE: runRecRCDescribe,
	}
	recRCUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update recommender config from --config-file",
		Args: cobra.NoArgs, RunE: runRecRCUpdate,
	}
)

func init() {
	// insights
	for _, c := range []*cobra.Command{recInsDescribeCmd, recInsListCmd, recInsMarkAcceptedCmd, recInsMarkActiveCmd, recInsMarkDismissedCmd} {
		c.Flags().StringVar(&flagRecLocation, "location", "global", "Location containing the insight")
		c.Flags().StringVar(&flagRecInsightType, "insight-type", "", "Insight type ID (required)")
		_ = c.MarkFlagRequired("insight-type")
	}
	recInsDescribeCmd.Flags().StringVar(&flagRecFormat, "format", "", "Output format")
	recInsListCmd.Flags().StringVar(&flagRecFormat, "format", "", "Output format")
	for _, c := range []*cobra.Command{recInsMarkAcceptedCmd, recInsMarkActiveCmd, recInsMarkDismissedCmd} {
		c.Flags().StringVar(&flagRecEtag, "etag", "", "Insight etag (required for state changes)")
		c.Flags().StringSliceVar(&flagRecStateMeta, "state-metadata", nil, "State metadata as KEY=VALUE (repeatable)")
	}
	recInsightsCmd.AddCommand(recInsDescribeCmd, recInsListCmd, recInsMarkAcceptedCmd, recInsMarkActiveCmd, recInsMarkDismissedCmd)
	recommenderCmd.AddCommand(recInsightsCmd)

	// recommendations
	for _, c := range []*cobra.Command{recRecDescribeCmd, recRecListCmd, recRecMarkActiveCmd, recRecMarkClaimedCmd, recRecMarkDismissedCmd, recRecMarkFailedCmd, recRecMarkSucceededCmd} {
		c.Flags().StringVar(&flagRecLocation, "location", "global", "Location containing the recommendation")
		c.Flags().StringVar(&flagRecRecommender, "recommender", "", "Recommender ID (required)")
		_ = c.MarkFlagRequired("recommender")
	}
	recRecDescribeCmd.Flags().StringVar(&flagRecFormat, "format", "", "Output format")
	recRecListCmd.Flags().StringVar(&flagRecFormat, "format", "", "Output format")
	for _, c := range []*cobra.Command{recRecMarkActiveCmd, recRecMarkClaimedCmd, recRecMarkDismissedCmd, recRecMarkFailedCmd, recRecMarkSucceededCmd} {
		c.Flags().StringVar(&flagRecEtag, "etag", "", "Recommendation etag (required for state changes)")
		c.Flags().StringSliceVar(&flagRecStateMeta, "state-metadata", nil, "State metadata as KEY=VALUE (repeatable)")
	}
	recRecommendationsCmd.AddCommand(recRecDescribeCmd, recRecListCmd, recRecMarkActiveCmd, recRecMarkClaimedCmd, recRecMarkDismissedCmd, recRecMarkFailedCmd, recRecMarkSucceededCmd)
	recommenderCmd.AddCommand(recRecommendationsCmd)

	// insight-type-config
	for _, c := range []*cobra.Command{recITCDescribeCmd, recITCUpdateCmd} {
		c.Flags().StringVar(&flagRecLocation, "location", "global", "Location containing the config")
		c.Flags().StringVar(&flagRecInsightType, "insight-type", "", "Insight type ID (required)")
		_ = c.MarkFlagRequired("insight-type")
	}
	recITCDescribeCmd.Flags().StringVar(&flagRecFormat, "format", "", "Output format")
	recITCUpdateCmd.Flags().StringVar(&flagRecConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the InsightTypeConfig message body (required)")
	_ = recITCUpdateCmd.MarkFlagRequired("config-file")
	recITCUpdateCmd.Flags().StringVar(&flagRecUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	recInsightTypeConfigCmd.AddCommand(recITCDescribeCmd, recITCUpdateCmd)
	recommenderCmd.AddCommand(recInsightTypeConfigCmd)

	// recommender-config
	for _, c := range []*cobra.Command{recRCDescribeCmd, recRCUpdateCmd} {
		c.Flags().StringVar(&flagRecLocation, "location", "global", "Location containing the config")
		c.Flags().StringVar(&flagRecRecommender, "recommender", "", "Recommender ID (required)")
		_ = c.MarkFlagRequired("recommender")
	}
	recRCDescribeCmd.Flags().StringVar(&flagRecFormat, "format", "", "Output format")
	recRCUpdateCmd.Flags().StringVar(&flagRecConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the RecommenderConfig message body (required)")
	_ = recRCUpdateCmd.MarkFlagRequired("config-file")
	recRCUpdateCmd.Flags().StringVar(&flagRecUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	recRecommenderConfigCmd.AddCommand(recRCDescribeCmd, recRCUpdateCmd)
	recommenderCmd.AddCommand(recRecommenderConfigCmd)

	rootCmd.AddCommand(recommenderCmd)
}

// stateMetadataMap parses --state-metadata KEY=VALUE list.
func stateMetadataMap() (map[string]string, error) {
	if len(flagRecStateMeta) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(flagRecStateMeta))
	for _, kv := range flagRecStateMeta {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("--state-metadata value %q must be KEY=VALUE", kv)
		}
		m[k] = v
	}
	return m, nil
}

// --- insights impl ---

func recInsightParent(project, location, insightType string) string {
	return fmt.Sprintf("projects/%s/locations/%s/insightTypes/%s", project, location, insightType)
}

func recInsightName(id, project, location, insightType string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/insights/%s", recInsightParent(project, location, insightType), id)
}

func runRecInsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RecommenderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.InsightTypes.Insights.Get(recInsightName(args[0], project, flagRecLocation, flagRecInsightType)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing insight: %w", err)
	}
	return emitFormatted(got, flagRecFormat)
}

func runRecInsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RecommenderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.InsightTypes.Insights.List(recInsightParent(project, flagRecLocation, flagRecInsightType)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing insights: %w", err)
	}
	if flagRecFormat != "" {
		return emitFormatted(resp.Insights, flagRecFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "CATEGORY")
	for _, i := range resp.Insights {
		fmt.Printf("%-40s %s\n", path.Base(i.Name), i.Category)
	}
	return nil
}

// runRecInsMark handles both accepted (via Go client) and active/dismissed
// (via raw HTTP because the v1 Go client omits them).
func runRecInsMark(state string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := resolveProject()
		if err != nil {
			return err
		}
		meta, err := stateMetadataMap()
		if err != nil {
			return err
		}
		ctx := context.Background()
		name := recInsightName(args[0], project, flagRecLocation, flagRecInsightType)
		if state == "accepted" {
			svc, err := gcp.RecommenderService(ctx, flagAccount)
			if err != nil {
				return err
			}
			req := &recommender.GoogleCloudRecommenderV1MarkInsightAcceptedRequest{
				Etag:          flagRecEtag,
				StateMetadata: meta,
			}
			got, err := svc.Projects.Locations.InsightTypes.Insights.MarkAccepted(name, req).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking insight accepted: %w", err)
			}
			return emitFormatted(got, "")
		}
		body := map[string]any{"etag": flagRecEtag}
		if len(meta) > 0 {
			body["stateMetadata"] = meta
		}
		verb := "markActive"
		if state == "dismissed" {
			verb = "markDismissed"
		}
		return recRawMark(ctx, name, verb, body)
	}
}

// --- recommendations impl ---

func recRecommendationParent(project, location, rec string) string {
	return fmt.Sprintf("projects/%s/locations/%s/recommenders/%s", project, location, rec)
}

func recRecommendationName(id, project, location, rec string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/recommendations/%s", recRecommendationParent(project, location, rec), id)
}

func runRecRecDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RecommenderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Recommenders.Recommendations.Get(recRecommendationName(args[0], project, flagRecLocation, flagRecRecommender)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing recommendation: %w", err)
	}
	return emitFormatted(got, flagRecFormat)
}

func runRecRecList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RecommenderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Recommenders.Recommendations.List(recRecommendationParent(project, flagRecLocation, flagRecRecommender)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing recommendations: %w", err)
	}
	if flagRecFormat != "" {
		return emitFormatted(resp.Recommendations, flagRecFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "PRIORITY")
	for _, r := range resp.Recommendations {
		fmt.Printf("%-40s %s\n", path.Base(r.Name), r.Priority)
	}
	return nil
}

func runRecRecMark(state string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := resolveProject()
		if err != nil {
			return err
		}
		meta, err := stateMetadataMap()
		if err != nil {
			return err
		}
		ctx := context.Background()
		name := recRecommendationName(args[0], project, flagRecLocation, flagRecRecommender)
		switch state {
		case "claimed":
			svc, err := gcp.RecommenderService(ctx, flagAccount)
			if err != nil {
				return err
			}
			got, err := svc.Projects.Locations.Recommenders.Recommendations.MarkClaimed(name,
				&recommender.GoogleCloudRecommenderV1MarkRecommendationClaimedRequest{Etag: flagRecEtag, StateMetadata: meta}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking recommendation claimed: %w", err)
			}
			return emitFormatted(got, "")
		case "dismissed":
			svc, err := gcp.RecommenderService(ctx, flagAccount)
			if err != nil {
				return err
			}
			got, err := svc.Projects.Locations.Recommenders.Recommendations.MarkDismissed(name,
				&recommender.GoogleCloudRecommenderV1MarkRecommendationDismissedRequest{Etag: flagRecEtag}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking recommendation dismissed: %w", err)
			}
			return emitFormatted(got, "")
		case "failed":
			svc, err := gcp.RecommenderService(ctx, flagAccount)
			if err != nil {
				return err
			}
			got, err := svc.Projects.Locations.Recommenders.Recommendations.MarkFailed(name,
				&recommender.GoogleCloudRecommenderV1MarkRecommendationFailedRequest{Etag: flagRecEtag, StateMetadata: meta}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking recommendation failed: %w", err)
			}
			return emitFormatted(got, "")
		case "succeeded":
			svc, err := gcp.RecommenderService(ctx, flagAccount)
			if err != nil {
				return err
			}
			got, err := svc.Projects.Locations.Recommenders.Recommendations.MarkSucceeded(name,
				&recommender.GoogleCloudRecommenderV1MarkRecommendationSucceededRequest{Etag: flagRecEtag, StateMetadata: meta}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking recommendation succeeded: %w", err)
			}
			return emitFormatted(got, "")
		case "active":
			body := map[string]any{"etag": flagRecEtag}
			if len(meta) > 0 {
				body["stateMetadata"] = meta
			}
			return recRawMark(ctx, name, "markActive", body)
		}
		return fmt.Errorf("unknown state %q", state)
	}
}

// recRawMark hits the recommender REST API directly for mark verbs missing
// from the Go client (markActive on both insights and recommendations, and
// markDismissed on insights).
func recRawMark(ctx context.Context, name, verb string, body map[string]any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ts, err := gcp.PlatformTokenSource(ctx, flagAccount)
	if err != nil {
		return err
	}
	tok, err := ts.Token()
	if err != nil {
		return fmt.Errorf("obtaining access token: %w", err)
	}
	url := fmt.Sprintf("https://recommender.googleapis.com/v1/%s:%s", name, verb)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	tok.SetAuthHeader(req)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", verb, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%s: HTTP %d: %s", verb, resp.StatusCode, string(respBody))
	}
	var out any
	if len(respBody) > 0 {
		_ = json.Unmarshal(respBody, &out)
	}
	return emitFormatted(out, "")
}

// --- insight-type-config impl ---

func recITCName(project, location, insightType string) string {
	return fmt.Sprintf("%s/config", recInsightParent(project, location, insightType))
}

func runRecITCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RecommenderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.InsightTypes.GetConfig(recITCName(project, flagRecLocation, flagRecInsightType)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing insight-type config: %w", err)
	}
	return emitFormatted(got, flagRecFormat)
}

func runRecITCUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cfg := &recommender.GoogleCloudRecommenderV1InsightTypeConfig{}
	if err := loadYAMLOrJSONInto(flagRecConfigFile, cfg); err != nil {
		return err
	}
	mask := flagRecUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cfg))
	}
	ctx := context.Background()
	svc, err := gcp.RecommenderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.InsightTypes.UpdateConfig(recITCName(project, flagRecLocation, flagRecInsightType), cfg).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating insight-type config: %w", err)
	}
	return emitFormatted(got, "")
}

// --- recommender-config impl ---

func recRCName(project, location, rec string) string {
	return fmt.Sprintf("%s/config", recRecommendationParent(project, location, rec))
}

func runRecRCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RecommenderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Recommenders.GetConfig(recRCName(project, flagRecLocation, flagRecRecommender)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing recommender config: %w", err)
	}
	return emitFormatted(got, flagRecFormat)
}

func runRecRCUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cfg := &recommender.GoogleCloudRecommenderV1RecommenderConfig{}
	if err := loadYAMLOrJSONInto(flagRecConfigFile, cfg); err != nil {
		return err
	}
	mask := flagRecUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cfg))
	}
	ctx := context.Background()
	svc, err := gcp.RecommenderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Recommenders.UpdateConfig(recRCName(project, flagRecLocation, flagRecRecommender), cfg).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating recommender config: %w", err)
	}
	return emitFormatted(got, "")
}
