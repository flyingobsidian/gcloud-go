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

	"github.com/spf13/cobra"
	recommender "google.golang.org/api/recommender/v1"
)

// --- gcloud recommender (#379) ---

// --- insights ---

// --- recommendations ---

// --- insight-type-config ---

// --- recommender-config ---

func (s *recommenderCmds) initRecommender() {
	// insights
	for _, c := range []*cobra.Command{s.recInsDescribeCmd, s.recInsListCmd, s.recInsMarkAcceptedCmd, s.recInsMarkActiveCmd, s.recInsMarkDismissedCmd} {
		c.Flags().StringVar(&s.flagRecLocation, "location", "global", "Location containing the insight")
		c.Flags().StringVar(&s.flagRecInsightType, "insight-type", "", "Insight type ID (required)")
		_ = c.MarkFlagRequired("insight-type")
	}
	s.recInsDescribeCmd.Flags().StringVar(&s.flagRecFormat, "format", "", "Output format")
	s.recInsListCmd.Flags().StringVar(&s.flagRecFormat, "format", "", "Output format")
	for _, c := range []*cobra.Command{s.recInsMarkAcceptedCmd, s.recInsMarkActiveCmd, s.recInsMarkDismissedCmd} {
		c.Flags().StringVar(&s.flagRecEtag, "etag", "", "Insight etag (required for state changes)")
		c.Flags().StringSliceVar(&s.flagRecStateMeta, "state-metadata", nil, "State metadata as KEY=VALUE (repeatable)")
	}
	s.recInsightsCmd.AddCommand(s.recInsDescribeCmd, s.recInsListCmd, s.recInsMarkAcceptedCmd, s.recInsMarkActiveCmd, s.recInsMarkDismissedCmd)
	s.recommenderCmd.AddCommand(s.recInsightsCmd)

	// recommendations
	for _, c := range []*cobra.Command{s.recRecDescribeCmd, s.recRecListCmd, s.recRecMarkActiveCmd, s.recRecMarkClaimedCmd, s.recRecMarkDismissedCmd, s.recRecMarkFailedCmd, s.recRecMarkSucceededCmd} {
		c.Flags().StringVar(&s.flagRecLocation, "location", "global", "Location containing the recommendation")
		c.Flags().StringVar(&s.flagRecRecommender, "recommender", "", "Recommender ID (required)")
		_ = c.MarkFlagRequired("recommender")
	}
	s.recRecDescribeCmd.Flags().StringVar(&s.flagRecFormat, "format", "", "Output format")
	s.recRecListCmd.Flags().StringVar(&s.flagRecFormat, "format", "", "Output format")
	for _, c := range []*cobra.Command{s.recRecMarkActiveCmd, s.recRecMarkClaimedCmd, s.recRecMarkDismissedCmd, s.recRecMarkFailedCmd, s.recRecMarkSucceededCmd} {
		c.Flags().StringVar(&s.flagRecEtag, "etag", "", "Recommendation etag (required for state changes)")
		c.Flags().StringSliceVar(&s.flagRecStateMeta, "state-metadata", nil, "State metadata as KEY=VALUE (repeatable)")
	}
	s.recRecommendationsCmd.AddCommand(s.recRecDescribeCmd, s.recRecListCmd, s.recRecMarkActiveCmd, s.recRecMarkClaimedCmd, s.recRecMarkDismissedCmd, s.recRecMarkFailedCmd, s.recRecMarkSucceededCmd)
	s.recommenderCmd.AddCommand(s.recRecommendationsCmd)

	// insight-type-config
	for _, c := range []*cobra.Command{s.recITCDescribeCmd, s.recITCUpdateCmd} {
		c.Flags().StringVar(&s.flagRecLocation, "location", "global", "Location containing the config")
		c.Flags().StringVar(&s.flagRecInsightType, "insight-type", "", "Insight type ID (required)")
		_ = c.MarkFlagRequired("insight-type")
	}
	s.recITCDescribeCmd.Flags().StringVar(&s.flagRecFormat, "format", "", "Output format")
	s.recITCUpdateCmd.Flags().StringVar(&s.flagRecConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the InsightTypeConfig message body (required)")
	_ = s.recITCUpdateCmd.MarkFlagRequired("config-file")
	s.recITCUpdateCmd.Flags().StringVar(&s.flagRecUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.recInsightTypeConfigCmd.AddCommand(s.recITCDescribeCmd, s.recITCUpdateCmd)
	s.recommenderCmd.AddCommand(s.recInsightTypeConfigCmd)

	// recommender-config
	for _, c := range []*cobra.Command{s.recRCDescribeCmd, s.recRCUpdateCmd} {
		c.Flags().StringVar(&s.flagRecLocation, "location", "global", "Location containing the config")
		c.Flags().StringVar(&s.flagRecRecommender, "recommender", "", "Recommender ID (required)")
		_ = c.MarkFlagRequired("recommender")
	}
	s.recRCDescribeCmd.Flags().StringVar(&s.flagRecFormat, "format", "", "Output format")
	s.recRCUpdateCmd.Flags().StringVar(&s.flagRecConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the RecommenderConfig message body (required)")
	_ = s.recRCUpdateCmd.MarkFlagRequired("config-file")
	s.recRCUpdateCmd.Flags().StringVar(&s.flagRecUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.recRecommenderConfigCmd.AddCommand(s.recRCDescribeCmd, s.recRCUpdateCmd)
	s.recommenderCmd.AddCommand(s.recRecommenderConfigCmd)

}

// stateMetadataMap parses --state-metadata KEY=VALUE list.
func (s *recommenderCmds) stateMetadataMap() (map[string]string, error) {
	if len(s.flagRecStateMeta) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(s.flagRecStateMeta))
	for _, kv := range s.flagRecStateMeta {
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

func (s *recommenderCmds) runRecInsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Recommender(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.InsightTypes.Insights.Get(recInsightName(args[0], project, s.flagRecLocation, s.flagRecInsightType)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing insight: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRecFormat)
}

func (s *recommenderCmds) runRecInsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Recommender(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.InsightTypes.Insights.List(recInsightParent(project, s.flagRecLocation, s.flagRecInsightType)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing insights: %w", err)
	}
	if s.flagRecFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Insights, s.flagRecFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Insights), "%-40s %s\n", "NAME", "CATEGORY")
	for _, i := range resp.Insights {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(i.Name), i.Category)
	}
	return nil
}

// runRecInsMark handles both accepted (via Go client) and active/dismissed
// (via raw HTTP because the v1 Go client omits them).
func (s *recommenderCmds) runRecInsMark(state string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := s.inv.project()
		if err != nil {
			return err
		}
		meta, err := s.stateMetadataMap()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		name := recInsightName(args[0], project, s.flagRecLocation, s.flagRecInsightType)
		if state == "accepted" {
			svc, err := s.inv.clients().Recommender(ctx)
			if err != nil {
				return err
			}
			req := &recommender.GoogleCloudRecommenderV1MarkInsightAcceptedRequest{
				Etag:          s.flagRecEtag,
				StateMetadata: meta,
			}
			got, err := svc.Projects.Locations.InsightTypes.Insights.MarkAccepted(name, req).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking insight accepted: %w", err)
			}
			return emitFormattedTo(s.inv.stdout, got, "")
		}
		body := map[string]any{"etag": s.flagRecEtag}
		if len(meta) > 0 {
			body["stateMetadata"] = meta
		}
		verb := "markActive"
		if state == "dismissed" {
			verb = "markDismissed"
		}
		return s.recRawMark(ctx, name, verb, body)
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

func (s *recommenderCmds) runRecRecDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Recommender(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Recommenders.Recommendations.Get(recRecommendationName(args[0], project, s.flagRecLocation, s.flagRecRecommender)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing recommendation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRecFormat)
}

func (s *recommenderCmds) runRecRecList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Recommender(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Recommenders.Recommendations.List(recRecommendationParent(project, s.flagRecLocation, s.flagRecRecommender)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing recommendations: %w", err)
	}
	if s.flagRecFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Recommendations, s.flagRecFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Recommendations), "%-40s %s\n", "NAME", "PRIORITY")
	for _, r := range resp.Recommendations {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(r.Name), r.Priority)
	}
	return nil
}

func (s *recommenderCmds) runRecRecMark(state string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := s.inv.project()
		if err != nil {
			return err
		}
		meta, err := s.stateMetadataMap()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		name := recRecommendationName(args[0], project, s.flagRecLocation, s.flagRecRecommender)
		switch state {
		case "claimed":
			svc, err := s.inv.clients().Recommender(ctx)
			if err != nil {
				return err
			}
			got, err := svc.Projects.Locations.Recommenders.Recommendations.MarkClaimed(name,
				&recommender.GoogleCloudRecommenderV1MarkRecommendationClaimedRequest{Etag: s.flagRecEtag, StateMetadata: meta}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking recommendation claimed: %w", err)
			}
			return emitFormattedTo(s.inv.stdout, got, "")
		case "dismissed":
			svc, err := s.inv.clients().Recommender(ctx)
			if err != nil {
				return err
			}
			got, err := svc.Projects.Locations.Recommenders.Recommendations.MarkDismissed(name,
				&recommender.GoogleCloudRecommenderV1MarkRecommendationDismissedRequest{Etag: s.flagRecEtag}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking recommendation dismissed: %w", err)
			}
			return emitFormattedTo(s.inv.stdout, got, "")
		case "failed":
			svc, err := s.inv.clients().Recommender(ctx)
			if err != nil {
				return err
			}
			got, err := svc.Projects.Locations.Recommenders.Recommendations.MarkFailed(name,
				&recommender.GoogleCloudRecommenderV1MarkRecommendationFailedRequest{Etag: s.flagRecEtag, StateMetadata: meta}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking recommendation failed: %w", err)
			}
			return emitFormattedTo(s.inv.stdout, got, "")
		case "succeeded":
			svc, err := s.inv.clients().Recommender(ctx)
			if err != nil {
				return err
			}
			got, err := svc.Projects.Locations.Recommenders.Recommendations.MarkSucceeded(name,
				&recommender.GoogleCloudRecommenderV1MarkRecommendationSucceededRequest{Etag: s.flagRecEtag, StateMetadata: meta}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("marking recommendation succeeded: %w", err)
			}
			return emitFormattedTo(s.inv.stdout, got, "")
		case "active":
			body := map[string]any{"etag": s.flagRecEtag}
			if len(meta) > 0 {
				body["stateMetadata"] = meta
			}
			return s.recRawMark(ctx, name, "markActive", body)
		}
		return fmt.Errorf("unknown state %q", state)
	}
}

// recRawMark hits the recommender REST API directly for mark verbs missing
// from the Go client (markActive on both insights and recommendations, and
// markDismissed on insights).
func (s *recommenderCmds) recRawMark(ctx context.Context, name, verb string, body map[string]any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ts, err := s.inv.clients().PlatformTokenSource(ctx)
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
	return emitFormattedTo(s.inv.stdout, out, "")
}

// --- insight-type-config impl ---

func recITCName(project, location, insightType string) string {
	return fmt.Sprintf("%s/config", recInsightParent(project, location, insightType))
}

func (s *recommenderCmds) runRecITCDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Recommender(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.InsightTypes.GetConfig(recITCName(project, s.flagRecLocation, s.flagRecInsightType)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing insight-type config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRecFormat)
}

func (s *recommenderCmds) runRecITCUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cfg := &recommender.GoogleCloudRecommenderV1InsightTypeConfig{}
	if err := loadYAMLOrJSONInto(s.flagRecConfigFile, cfg); err != nil {
		return err
	}
	mask := s.flagRecUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cfg))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Recommender(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.InsightTypes.UpdateConfig(recITCName(project, s.flagRecLocation, s.flagRecInsightType), cfg).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating insight-type config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- recommender-config impl ---

func recRCName(project, location, rec string) string {
	return fmt.Sprintf("%s/config", recRecommendationParent(project, location, rec))
}

func (s *recommenderCmds) runRecRCDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Recommender(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Recommenders.GetConfig(recRCName(project, s.flagRecLocation, s.flagRecRecommender)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing recommender config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRecFormat)
}

func (s *recommenderCmds) runRecRCUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cfg := &recommender.GoogleCloudRecommenderV1RecommenderConfig{}
	if err := loadYAMLOrJSONInto(s.flagRecConfigFile, cfg); err != nil {
		return err
	}
	mask := s.flagRecUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cfg))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Recommender(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Recommenders.UpdateConfig(recRCName(project, s.flagRecLocation, s.flagRecRecommender), cfg).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating recommender config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// recommenderCmds holds the commands and flag values of one invocation's "recommender"
// command group, so invocations don't share state (#2070).
type recommenderCmds struct {
	inv *invocation

	recommenderCmd          *cobra.Command
	flagRecLocation         string
	flagRecInsightType      string
	flagRecRecommender      string
	flagRecEtag             string
	flagRecFormat           string
	flagRecConfigFile       string
	flagRecStateMeta        []string
	flagRecUpdateMask       string
	recInsightsCmd          *cobra.Command
	recInsDescribeCmd       *cobra.Command
	recInsListCmd           *cobra.Command
	recInsMarkAcceptedCmd   *cobra.Command
	recInsMarkActiveCmd     *cobra.Command
	recInsMarkDismissedCmd  *cobra.Command
	recRecommendationsCmd   *cobra.Command
	recRecDescribeCmd       *cobra.Command
	recRecListCmd           *cobra.Command
	recRecMarkActiveCmd     *cobra.Command
	recRecMarkClaimedCmd    *cobra.Command
	recRecMarkDismissedCmd  *cobra.Command
	recRecMarkFailedCmd     *cobra.Command
	recRecMarkSucceededCmd  *cobra.Command
	recInsightTypeConfigCmd *cobra.Command
	recITCDescribeCmd       *cobra.Command
	recITCUpdateCmd         *cobra.Command
	recRecommenderConfigCmd *cobra.Command
	recRCDescribeCmd        *cobra.Command
	recRCUpdateCmd          *cobra.Command
}

// newRecommenderCmd returns the "recommender" command group for inv.
func newRecommenderCmd(inv *invocation) *cobra.Command {
	return newRecommenderCmds(inv).recommenderCmd
}

// newRecommenderCmds builds the "recommender" command group for inv and returns its state.
func newRecommenderCmds(inv *invocation) *recommenderCmds {
	s := &recommenderCmds{inv: inv}
	s.recommenderCmd = &cobra.Command{Use: "recommender", Short: "Manage Cloud Recommender"}
	s.recInsightsCmd = &cobra.Command{Use: "insights", Short: "Manage recommender insights"}
	s.recInsDescribeCmd = &cobra.Command{
		Use: "describe INSIGHT", Short: "Describe an insight",
		Args: cobra.ExactArgs(1), RunE: s.runRecInsDescribe,
	}
	s.recInsListCmd = &cobra.Command{
		Use: "list", Short: "List insights",
		Args: cobra.NoArgs, RunE: s.runRecInsList,
	}
	s.recInsMarkAcceptedCmd = &cobra.Command{
		Use: "mark-accepted INSIGHT", Short: "Mark an insight as accepted",
		Args: cobra.ExactArgs(1), RunE: s.runRecInsMark("accepted"),
	}
	s.recInsMarkActiveCmd = &cobra.Command{
		Use: "mark-active INSIGHT", Short: "Mark an insight as active",
		Args: cobra.ExactArgs(1), RunE: s.runRecInsMark("active"),
	}
	s.recInsMarkDismissedCmd = &cobra.Command{
		Use: "mark-dismissed INSIGHT", Short: "Mark an insight as dismissed",
		Args: cobra.ExactArgs(1), RunE: s.runRecInsMark("dismissed"),
	}
	s.recRecommendationsCmd = &cobra.Command{Use: "recommendations", Short: "Manage recommendations"}
	s.recRecDescribeCmd = &cobra.Command{
		Use: "describe RECOMMENDATION", Short: "Describe a recommendation",
		Args: cobra.ExactArgs(1), RunE: s.runRecRecDescribe,
	}
	s.recRecListCmd = &cobra.Command{
		Use: "list", Short: "List recommendations",
		Args: cobra.NoArgs, RunE: s.runRecRecList,
	}
	s.recRecMarkActiveCmd = &cobra.Command{
		Use: "mark-active RECOMMENDATION", Short: "Mark a recommendation as active",
		Args: cobra.ExactArgs(1), RunE: s.runRecRecMark("active"),
	}
	s.recRecMarkClaimedCmd = &cobra.Command{
		Use: "mark-claimed RECOMMENDATION", Short: "Mark a recommendation as claimed",
		Args: cobra.ExactArgs(1), RunE: s.runRecRecMark("claimed"),
	}
	s.recRecMarkDismissedCmd = &cobra.Command{
		Use: "mark-dismissed RECOMMENDATION", Short: "Mark a recommendation as dismissed",
		Args: cobra.ExactArgs(1), RunE: s.runRecRecMark("dismissed"),
	}
	s.recRecMarkFailedCmd = &cobra.Command{
		Use: "mark-failed RECOMMENDATION", Short: "Mark a recommendation as failed",
		Args: cobra.ExactArgs(1), RunE: s.runRecRecMark("failed"),
	}
	s.recRecMarkSucceededCmd = &cobra.Command{
		Use: "mark-succeeded RECOMMENDATION", Short: "Mark a recommendation as succeeded",
		Args: cobra.ExactArgs(1), RunE: s.runRecRecMark("succeeded"),
	}
	s.recInsightTypeConfigCmd = &cobra.Command{Use: "insight-type-config", Short: "Manage insight-type configuration"}
	s.recITCDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe insight-type config",
		Args: cobra.NoArgs, RunE: s.runRecITCDescribe,
	}
	s.recITCUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update insight-type config from --config-file",
		Args: cobra.NoArgs, RunE: s.runRecITCUpdate,
	}
	s.recRecommenderConfigCmd = &cobra.Command{Use: "recommender-config", Short: "Manage recommender configuration"}
	s.recRCDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe recommender config",
		Args: cobra.NoArgs, RunE: s.runRecRCDescribe,
	}
	s.recRCUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update recommender config from --config-file",
		Args: cobra.NoArgs, RunE: s.runRecRCUpdate,
	}
	s.initRecommender()
	return s
}
