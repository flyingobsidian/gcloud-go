package cmd

import (
	"encoding/json"
	"os"
	"testing"
)

func TestRunCreateCredConfig(t *testing.T) {
	outFile := t.TempDir() + "/cred-config.json"

	// Simulate the command.
	flagOutputFile = outFile
	flagServiceAccount = "sa@project.iam.gserviceaccount.com"
	flagCredentialSourceFile = "/var/run/token"
	flagCredentialSourceType = "text"
	flagCredentialSourceFieldName = ""
	flagSubjectTokenType = "urn:ietf:params:oauth:token-type:jwt"
	flagServiceAccountTokenLifetime = 3600
	flagAws = false
	flagCredentialSourceURL = ""
	flagExecutableCommand = ""

	err := runCreateCredConfig(nil, []string{
		"//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/pool/providers/prov",
	})
	if err != nil {
		t.Fatalf("runCreateCredConfig() error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg["type"] != "external_account" {
		t.Errorf("type = %v, want external_account", cfg["type"])
	}
	if cfg["token_url"] != "https://sts.googleapis.com/v1/token" {
		t.Errorf("token_url = %v", cfg["token_url"])
	}
	if cfg["service_account_impersonation_url"] == nil {
		t.Error("expected service_account_impersonation_url")
	}

	credSource, ok := cfg["credential_source"].(map[string]any)
	if !ok {
		t.Fatal("credential_source missing or wrong type")
	}
	if credSource["file"] != "/var/run/token" {
		t.Errorf("credential_source.file = %v, want /var/run/token", credSource["file"])
	}
}

func TestResolveSubjectTokenType(t *testing.T) {
	// Reset flags.
	flagSubjectTokenType = ""
	flagAws = false

	got := resolveSubjectTokenType()
	if got != "urn:ietf:params:oauth:token-type:jwt" {
		t.Errorf("default = %q", got)
	}

	flagAws = true
	got = resolveSubjectTokenType()
	if got != "urn:ietf:params:aws:token-type:aws4_request" {
		t.Errorf("aws = %q", got)
	}

	flagSubjectTokenType = "custom"
	got = resolveSubjectTokenType()
	if got != "custom" {
		t.Errorf("custom = %q", got)
	}

	// Reset.
	flagSubjectTokenType = ""
	flagAws = false
}

func TestBuildCredentialSourceFile(t *testing.T) {
	flagCredentialSourceFile = "/path/to/token"
	flagCredentialSourceURL = ""
	flagExecutableCommand = ""
	flagAws = false
	flagCredentialSourceType = "json"
	flagCredentialSourceFieldName = "access_token"

	src, err := buildCredentialSource()
	if err != nil {
		t.Fatal(err)
	}

	if src["file"] != "/path/to/token" {
		t.Errorf("file = %v", src["file"])
	}
	format, ok := src["format"].(map[string]any)
	if !ok {
		t.Fatal("format missing")
	}
	if format["type"] != "json" {
		t.Errorf("format.type = %v", format["type"])
	}
	if format["subject_token_field_name"] != "access_token" {
		t.Errorf("format.subject_token_field_name = %v", format["subject_token_field_name"])
	}

	// Reset.
	flagCredentialSourceFile = ""
	flagCredentialSourceType = ""
	flagCredentialSourceFieldName = ""
}

func TestBuildCredentialSourceNone(t *testing.T) {
	flagCredentialSourceFile = ""
	flagCredentialSourceURL = ""
	flagExecutableCommand = ""
	flagAws = false

	_, err := buildCredentialSource()
	if err == nil {
		t.Error("expected error when no source specified")
	}
}
