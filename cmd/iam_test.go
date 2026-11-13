package cmd

import (
	"encoding/json"
	"os"
	"testing"
)

func TestRunCreateCredConfig(t *testing.T) {
	grp := newIamCmds(legacyTestInvocation())
	outFile := t.TempDir() + "/cred-config.json"

	// Simulate the command.
	grp.flagOutputFile = outFile
	grp.flagServiceAccount = "sa@project.iam.gserviceaccount.com"
	grp.flagCredentialSourceFile = "/var/run/token"
	grp.flagCredentialSourceType = "text"
	grp.flagCredentialSourceFieldName = ""
	grp.flagSubjectTokenType = "urn:ietf:params:oauth:token-type:jwt"
	grp.flagServiceAccountTokenLifetime = 3600
	grp.flagAws = false
	grp.flagCredentialSourceURL = ""
	grp.flagExecutableCommand = ""

	err := grp.runCreateCredConfig(nil, []string{
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
	grp := newIamCmds(legacyTestInvocation())
	// Reset flags.
	grp.flagSubjectTokenType = ""
	grp.flagAws = false

	got := grp.resolveSubjectTokenType()
	if got != "urn:ietf:params:oauth:token-type:jwt" {
		t.Errorf("default = %q", got)
	}

	grp.flagAws = true
	got = grp.resolveSubjectTokenType()
	if got != "urn:ietf:params:aws:token-type:aws4_request" {
		t.Errorf("aws = %q", got)
	}

	grp.flagSubjectTokenType = "custom"
	got = grp.resolveSubjectTokenType()
	if got != "custom" {
		t.Errorf("custom = %q", got)
	}

	// Reset.
	grp.flagSubjectTokenType = ""
	grp.flagAws = false
}

func TestBuildCredentialSourceFile(t *testing.T) {
	grp := newIamCmds(legacyTestInvocation())
	grp.flagCredentialSourceFile = "/path/to/token"
	grp.flagCredentialSourceURL = ""
	grp.flagExecutableCommand = ""
	grp.flagAws = false
	grp.flagCredentialSourceType = "json"
	grp.flagCredentialSourceFieldName = "access_token"

	src, err := grp.buildCredentialSource()
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
	grp.flagCredentialSourceFile = ""
	grp.flagCredentialSourceType = ""
	grp.flagCredentialSourceFieldName = ""
}

func TestBuildCredentialSourceNone(t *testing.T) {
	grp := newIamCmds(legacyTestInvocation())
	grp.flagCredentialSourceFile = ""
	grp.flagCredentialSourceURL = ""
	grp.flagExecutableCommand = ""
	grp.flagAws = false

	_, err := grp.buildCredentialSource()
	if err == nil {
		t.Error("expected error when no source specified")
	}
}
