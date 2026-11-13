package cli

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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// --- gcloud pubsub message-transforms (#1175) ---
//
// The Pub/Sub v1 REST API exposes projects:testMessageTransforms and
// projects:validateMessageTransform, but the generated Go client does not
// (yet) surface them, so we talk directly to pubsub.googleapis.com.

func (g *pubsubCmds) initPubsubMessageTransforms() {
	g.pubsubMTTestCmd.Flags().StringVar(&g.flagPMTMessage, "message", "",
		"Message body to test (utf-8 string)")
	g.pubsubMTTestCmd.Flags().StringSliceVar(&g.flagPMTAttributes, "attribute", nil,
		"Message attributes as KEY=VALUE (may be repeated)")
	g.pubsubMTTestCmd.Flags().StringVar(&g.flagPMTMessageTransforms, "message-transforms-file", "",
		"Path to a YAML/JSON file containing message transforms to test")
	g.pubsubMTTestCmd.Flags().StringVar(&g.flagPMTTopic, "topic", "",
		"Topic whose transforms to test against")
	g.pubsubMTTestCmd.Flags().StringVar(&g.flagPMTSubscription, "subscription", "",
		"Subscription whose transforms to test against")
	g.pubsubMTTestCmd.Flags().StringVar(&g.flagPMTFormat, "format", "", "Output format")

	g.pubsubMTValidateCmd.Flags().StringVar(&g.flagPMTTransformFile, "message-transform-file", "",
		"Path to a YAML/JSON file containing a single message transform to validate (required)")
	_ = g.pubsubMTValidateCmd.MarkFlagRequired("message-transform-file")

	g.pubsubMessageTransformsCmd.AddCommand(g.pubsubMTTestCmd, g.pubsubMTValidateCmd)
	g.pubsubCmd.AddCommand(g.pubsubMessageTransformsCmd)
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

func (g *pubsubCmds) pmtAttributesFromFlag() (map[string]string, error) {
	if len(g.flagPMTAttributes) == 0 {
		return nil, nil
	}
	out := map[string]string{}
	for _, kv := range g.flagPMTAttributes {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("--attribute %q must be KEY=VALUE", kv)
		}
		out[k] = v
	}
	return out, nil
}

func (g *pubsubCmds) pmtTopicRef(project string) string {
	if g.flagPMTTopic == "" {
		return ""
	}
	if strings.HasPrefix(g.flagPMTTopic, "projects/") {
		return g.flagPMTTopic
	}
	return fmt.Sprintf("projects/%s/topics/%s", project, g.flagPMTTopic)
}

func (g *pubsubCmds) pmtSubscriptionRef(project string) string {
	if g.flagPMTSubscription == "" {
		return ""
	}
	if strings.HasPrefix(g.flagPMTSubscription, "projects/") {
		return g.flagPMTSubscription
	}
	return fmt.Sprintf("projects/%s/subscriptions/%s", project, g.flagPMTSubscription)
}

func (g *pubsubCmds) runPMTTest(cmd *cobra.Command, args []string) error {
	if g.flagPMTMessage == "" && len(g.flagPMTAttributes) == 0 {
		return fmt.Errorf("must specify --message, one or more --attribute, or both")
	}
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	attrs, err := g.pmtAttributesFromFlag()
	if err != nil {
		return err
	}
	message := map[string]any{}
	if g.flagPMTMessage != "" {
		message["data"] = base64.StdEncoding.EncodeToString([]byte(g.flagPMTMessage))
	}
	if attrs != nil {
		message["attributes"] = attrs
	}
	body := map[string]any{"message": message}
	if g.flagPMTMessageTransforms != "" {
		transforms, err := pmtLoadTransforms(g.flagPMTMessageTransforms)
		if err != nil {
			return err
		}
		body["messageTransforms"] = map[string]any{"messageTransforms": transforms}
	}
	if ref := g.pmtTopicRef(project); ref != "" {
		body["topic"] = ref
	}
	if ref := g.pmtSubscriptionRef(project); ref != "" {
		body["subscription"] = ref
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	return g.pubsubHTTP(ctx, http.MethodPost,
		fmt.Sprintf("projects/%s:testMessageTransforms", project), payload, g.flagPMTFormat)
}

func (g *pubsubCmds) runPMTValidate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	transforms, err := pmtLoadTransforms(g.flagPMTTransformFile)
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
	ctx := cmd.Context()
	if err := g.pubsubHTTP(ctx, http.MethodPost,
		fmt.Sprintf("projects/%s:validateMessageTransform", project), payload, ""); err != nil {
		return err
	}
	fmt.Fprintln(g.inv.stdout, "Message transform is valid.")
	return nil
}

// pubsubHTTP posts an authenticated JSON request against
// https://pubsub.googleapis.com/v1 and emits the decoded body.
func (g *pubsubCmds) pubsubHTTP(ctx context.Context, method, name string, payload []byte, format string) error {
	ts, err := g.inv.clients().PlatformTokenSource(ctx)
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
		_, werr := g.inv.stdout.Write(respBody)
		return werr
	}
	return emitFormattedTo(g.inv.stdout, out, format)
}
