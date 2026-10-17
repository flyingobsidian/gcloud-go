package cmd

import (
	"encoding/json"
	"testing"

	aiplatform "google.golang.org/api/aiplatform/v1"
)

func TestAISGPDenialMessageFlag(t *testing.T) {
	for _, c := range []string{"create", "update"} {
		cmd, _, err := aiSGPCmd.Find([]string{c})
		if err != nil || cmd.Name() != c {
			t.Fatalf("semantic-governance-policies %s missing", c)
		}
		if cmd.Flags().Lookup("agent-response-denial-message") == nil {
			t.Errorf("%s: missing --agent-response-denial-message", c)
		}
	}
}

func TestAISGPSetDenialMessage(t *testing.T) {
	body := &aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy{DisplayName: "p"}
	aiSGPSetDenialMessage(body, "Not allowed")
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"agentResponseCustomization":{"denialMessage":"Not allowed"},"displayName":"p"}`
	if string(raw) != want {
		t.Errorf("got %s, want %s", raw, want)
	}

	// An explicitly empty message is still sent.
	empty := &aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy{}
	aiSGPSetDenialMessage(empty, "")
	raw, _ = json.Marshal(empty)
	if string(raw) != `{"agentResponseCustomization":{"denialMessage":""}}` {
		t.Errorf("empty message: got %s", raw)
	}
}

func TestAISGPUpdateMaskFor(t *testing.T) {
	tests := []struct {
		name      string
		body      *aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy
		denialSet bool
		want      string
	}{
		{"denial only", &aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy{}, true,
			"agentResponseCustomization.denialMessage"},
		{"config only", &aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy{
			NaturalLanguageConstraint: "c", Description: "d"}, false,
			"description,naturalLanguageConstraint"},
		{"config and denial", &aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy{
			Description: "d"}, true,
			"agentResponseCustomization.denialMessage,description"},
		{"config sets customization", &aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy{
			AgentResponseCustomization: &aiplatform.GoogleCloudAiplatformV1AgentResponseCustomization{DenialMessage: "x"}}, true,
			"agentResponseCustomization"},
	}
	for _, tt := range tests {
		if got := aiSGPUpdateMaskFor(tt.body, tt.denialSet); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
