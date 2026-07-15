package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// --- gcloud pubsub message-transforms (#1175) ---
//
// The Pub/Sub v1 REST API exposes projects:testMessageTransforms and
// projects:validateMessageTransform, but the generated Go client does not
// (yet) surface them, so we talk directly to pubsub.googleapis.com.

var pubsubMessageTransformsCmd = &cobra.Command{
	Use:   "message-transforms",
	Short: "Manage Cloud Pub/Sub message transforms",
}

var (
	flagPMTMessage           string
	flagPMTAttributes        []string
	flagPMTMessageTransforms string
	flagPMTTopic             string
	flagPMTSubscription      string
	flagPMTTransformFile     string
	flagPMTFormat            string
)

var (
	pubsubMTTestCmd = &cobra.Command{
		Use: "test", Short: "Test message transforms against a given message",
		Args: cobra.NoArgs, RunE: runPMTTest,
	}
	pubsubMTValidateCmd = &cobra.Command{
		Use: "validate", Short: "Validate a message transform",
		Args: cobra.NoArgs, RunE: runPMTValidate,
	}
)

func init() {
	pubsubMTTestCmd.Flags().StringVar(&flagPMTMessage, "message", "",
		"Message body to test (utf-8 string)")
	pubsubMTTestCmd.Flags().StringSliceVar(&flagPMTAttributes, "attribute", nil,
		"Message attributes as KEY=VALUE (may be repeated)")
	pubsubMTTestCmd.Flags().StringVar(&flagPMTMessageTransforms, "message-transforms-file", "",
		"Path to a YAML/JSON file containing message transforms to test")
	pubsubMTTestCmd.Flags().StringVar(&flagPMTTopic, "topic", "",
		"Topic whose transforms to test against")
	pubsubMTTestCmd.Flags().StringVar(&flagPMTSubscription, "subscription", "",
		"Subscription whose transforms to test against")
	pubsubMTTestCmd.Flags().StringVar(&flagPMTFormat, "format", "", "Output format")

	pubsubMTValidateCmd.Flags().StringVar(&flagPMTTransformFile, "message-transform-file", "",
		"Path to a YAML/JSON file containing a single message transform to validate (required)")
	_ = pubsubMTValidateCmd.MarkFlagRequired("message-transform-file")

	pubsubMessageTransformsCmd.AddCommand(pubsubMTTestCmd, pubsubMTValidateCmd)
	pubsubCmd.AddCommand(pubsubMessageTransformsCmd)
}

// pmtLoadTransforms reads a file containing either a single message transform
// or a list of them, and returns the normalised slice.
func pmtLoadTransforms(path string) ([]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	converted := convertYAMLKeys(raw)
	switch v := converted.(type) {
	case []any:
		return v, nil
	case map[string]any:
		if inner, ok := v["messageTransforms"]; ok {
			list, ok := inner.([]any)
			if !ok {
				return nil, fmt.Errorf("%s: messageTransforms must be a list", path)
			}
			return list, nil
		}
		return []any{v}, nil
	default:
		return nil, fmt.Errorf("%s: unexpected top-level YAML/JSON type", path)
	}
}

func pmtAttributesFromFlag() (map[string]string, error) {
	if len(flagPMTAttributes) == 0 {
		return nil, nil
	}
	out := map[string]string{}
	for _, kv := range flagPMTAttributes {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("--attribute %q must be KEY=VALUE", kv)
		}
		out[k] = v
	}
	return out, nil
}

func pmtTopicRef(project string) string {
	if flagPMTTopic == "" {
		return ""
	}
	if strings.HasPrefix(flagPMTTopic, "projects/") {
		return flagPMTTopic
	}
	return fmt.Sprintf("projects/%s/topics/%s", project, flagPMTTopic)
}

func pmtSubscriptionRef(project string) string {
	if flagPMTSubscription == "" {
		return ""
	}
	if strings.HasPrefix(flagPMTSubscription, "projects/") {
		return flagPMTSubscription
	}
	return fmt.Sprintf("projects/%s/subscriptions/%s", project, flagPMTSubscription)
}

func runPMTTest(cmd *cobra.Command, args []string) error {
	if flagPMTMessage == "" && len(flagPMTAttributes) == 0 {
		return fmt.Errorf("must specify --message, one or more --attribute, or both")
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	attrs, err := pmtAttributesFromFlag()
	if err != nil {
		return err
	}
	message := map[string]any{}
	if flagPMTMessage != "" {
		message["data"] = base64.StdEncoding.EncodeToString([]byte(flagPMTMessage))
	}
	if attrs != nil {
		message["attributes"] = attrs
	}
	body := map[string]any{"message": message}
	if flagPMTMessageTransforms != "" {
		transforms, err := pmtLoadTransforms(flagPMTMessageTransforms)
		if err != nil {
			return err
		}
		body["messageTransforms"] = map[string]any{"messageTransforms": transforms}
	}
	if ref := pmtTopicRef(project); ref != "" {
		body["topic"] = ref
	}
	if ref := pmtSubscriptionRef(project); ref != "" {
		body["subscription"] = ref
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return pubsubHTTP(ctx, http.MethodPost,
		fmt.Sprintf("projects/%s:testMessageTransforms", project), payload, flagPMTFormat)
}

func runPMTValidate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	transforms, err := pmtLoadTransforms(flagPMTTransformFile)
	if err != nil {
		return err
	}
	if len(transforms) != 1 {
		return fmt.Errorf("--message-transform-file must contain exactly one message transform")
	}
	body := map[string]any{"messageTransform": transforms[0]}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := pubsubHTTP(ctx, http.MethodPost,
		fmt.Sprintf("projects/%s:validateMessageTransform", project), payload, ""); err != nil {
		return err
	}
	fmt.Println("Message transform is valid.")
	return nil
}

// pubsubHTTP posts an authenticated JSON request against
// https://pubsub.googleapis.com/v1 and emits the decoded body.
func pubsubHTTP(ctx context.Context, method, name string, payload []byte, format string) error {
	ts, err := gcp.PlatformTokenSource(ctx, flagAccount)
	if err != nil {
		return err
	}
	tok, err := ts.Token()
	if err != nil {
		return fmt.Errorf("obtaining access token: %w", err)
	}
	target := fmt.Sprintf("https://pubsub.googleapis.com/v1/%s", name)
	var reqBody io.Reader
	if payload != nil {
		reqBody = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reqBody)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	tok.SetAuthHeader(req)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP %s: %w", method, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	if len(respBody) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal(respBody, &out); err != nil {
		_, werr := os.Stdout.Write(respBody)
		return werr
	}
	return emitFormatted(out, format)
}
