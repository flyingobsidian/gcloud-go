package cmd

import (
	"encoding/json"
	"os"
	"testing"
)

func resetIamWpCcFlags() {
	flagIamWpCcOutputFile = ""
	flagIamWpCcUserProject = ""
	flagIamWpCcSubjectTokenType = ""
	flagIamWpCcCredSourceFile = ""
	flagIamWpCcCredSourceURL = ""
	flagIamWpCcCredSourceHeaders = nil
	flagIamWpCcCredSourceType = ""
	flagIamWpCcCredSourceField = ""
	flagIamWpCcExecCommand = ""
	flagIamWpCcExecTimeoutMs = 30000
	flagIamWpCcExecOutputFile = ""
	flagIamWpCcServiceAccount = ""
	flagIamWpCcSATokenLifetime = 0
	flagIamWpCcStsLocation = ""
	flagIamWpCcUniverseDomain = ""
	flagIamWpCcEnableMtls = false
}

func TestNormalizeWorkforceAudience(t *testing.T) {
	cases := []struct {
		in, want string
		wantErr  bool
	}{
		{
			in:   "locations/global/workforcePools/my-pool/providers/my-prov",
			want: "locations/global/workforcePools/my-pool/providers/my-prov",
		},
		{
			in:   "//iam.googleapis.com/locations/global/workforcePools/my-pool/providers/my-prov",
			want: "locations/global/workforcePools/my-pool/providers/my-prov",
		},
		{
			in:   "my-pool/my-prov",
			want: "locations/global/workforcePools/my-pool/providers/my-prov",
		},
		{in: "onlyone", wantErr: true},
		{in: "a/b/c", wantErr: true},
	}
	for _, tc := range cases {
		got, err := normalizeWorkforceAudience(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("normalizeWorkforceAudience(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizeWorkforceAudience(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("normalizeWorkforceAudience(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestStsBaseURL(t *testing.T) {
	cases := []struct {
		universe, location string
		mtls               bool
		want               string
		wantErr            bool
	}{
		{universe: "googleapis.com", location: "", mtls: false, want: "https://sts.googleapis.com"},
		{universe: "googleapis.com", location: "global", mtls: false, want: "https://sts.googleapis.com"},
		{universe: "googleapis.com", location: "", mtls: true, want: "https://sts.mtls.googleapis.com"},
		{universe: "googleapis.com", location: "us-east1", mtls: false, want: "https://sts.us-east1.rep.googleapis.com"},
		{universe: "googleapis.com", location: "us-east1", mtls: true, wantErr: true},
	}
	for _, tc := range cases {
		got, err := stsBaseURL(tc.universe, tc.location, tc.mtls)
		if tc.wantErr {
			if err == nil {
				t.Errorf("stsBaseURL(%v) expected error", tc)
			}
			continue
		}
		if err != nil {
			t.Errorf("stsBaseURL(%v) unexpected error: %v", tc, err)
			continue
		}
		if got != tc.want {
			t.Errorf("stsBaseURL(%v) = %q, want %q", tc, got, tc.want)
		}
	}
}

func TestRunIamWpCreateCredConfigShortForm(t *testing.T) {
	resetIamWpCcFlags()
	defer resetIamWpCcFlags()

	out := t.TempDir() + "/wp-cred.json"
	flagIamWpCcOutputFile = out
	flagIamWpCcUserProject = "42"
	flagIamWpCcCredSourceFile = "/var/run/wpi"

	if err := runIamWpCreateCredConfig(nil, []string{"my-pool/my-prov"}); err != nil {
		t.Fatalf("runIamWpCreateCredConfig error: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg["type"] != "external_account" {
		t.Errorf("type = %v", cfg["type"])
	}
	if cfg["audience"] != "//iam.googleapis.com/locations/global/workforcePools/my-pool/providers/my-prov" {
		t.Errorf("audience = %v", cfg["audience"])
	}
	if cfg["subject_token_type"] != "urn:ietf:params:oauth:token-type:id_token" {
		t.Errorf("subject_token_type = %v", cfg["subject_token_type"])
	}
	if cfg["workforce_pool_user_project"] != "42" {
		t.Errorf("workforce_pool_user_project = %v", cfg["workforce_pool_user_project"])
	}
	if cfg["token_url"] != "https://sts.googleapis.com/v1/token" {
		t.Errorf("token_url = %v", cfg["token_url"])
	}
	if cfg["token_info_url"] != "https://sts.googleapis.com/v1/introspect" {
		t.Errorf("token_info_url = %v", cfg["token_info_url"])
	}
}

func TestRunIamWpCreateCredConfigStsLocation(t *testing.T) {
	resetIamWpCcFlags()
	defer resetIamWpCcFlags()

	out := t.TempDir() + "/wp-cred.json"
	flagIamWpCcOutputFile = out
	flagIamWpCcUserProject = "42"
	flagIamWpCcCredSourceFile = "/var/run/wpi"
	flagIamWpCcStsLocation = "us-east1"

	if err := runIamWpCreateCredConfig(nil, []string{"pool/prov"}); err != nil {
		t.Fatalf("error: %v", err)
	}
	data, _ := os.ReadFile(out)
	var cfg map[string]any
	_ = json.Unmarshal(data, &cfg)
	if cfg["token_url"] != "https://sts.us-east1.rep.googleapis.com/v1/token" {
		t.Errorf("token_url = %v", cfg["token_url"])
	}
}

func TestRunIamWpCreateLoginConfig(t *testing.T) {
	out := t.TempDir() + "/wp-login.json"
	flagIamWpLcOutputFile = out
	flagIamWpLcUniverseDomain = ""
	flagIamWpLcCloudWebDomain = ""
	flagIamWpLcEnableMtls = false
	flagIamWpLcActivate = false

	if err := runIamWpCreateLoginConfig(nil, []string{"pool/prov"}); err != nil {
		t.Fatalf("error: %v", err)
	}
	data, _ := os.ReadFile(out)
	var cfg map[string]any
	_ = json.Unmarshal(data, &cfg)
	if cfg["type"] != "external_account_authorized_user_login_config" {
		t.Errorf("type = %v", cfg["type"])
	}
	if cfg["audience"] != "//iam.googleapis.com/locations/global/workforcePools/pool/providers/prov" {
		t.Errorf("audience = %v", cfg["audience"])
	}
	if cfg["auth_url"] != "https://auth.cloud.google/authorize" {
		t.Errorf("auth_url = %v", cfg["auth_url"])
	}
	if cfg["token_url"] != "https://sts.googleapis.com/v1/oauthtoken" {
		t.Errorf("token_url = %v", cfg["token_url"])
	}
}

func TestRunCreateCredConfigX509Mtls(t *testing.T) {
	// X.509 cert path should imply mTLS on STS endpoints.
	flagOutputFile = t.TempDir() + "/cred.json"
	flagCredCertPath = "/tmp/cert.pem"
	flagCredCertKeyPath = "/tmp/key.pem"
	flagCredentialSourceFile = ""
	flagCredentialSourceURL = ""
	flagExecutableCommand = ""
	flagAws = false
	flagAzure = false
	flagServiceAccount = ""
	flagServiceAccountTokenLifetime = 0
	flagCredStsLocation = ""
	flagCredUniverseDomain = ""
	flagCredEnableMtls = false

	if err := runCreateCredConfig(nil, []string{"//iam.googleapis.com/x"}); err != nil {
		t.Fatalf("runCreateCredConfig error: %v", err)
	}
	data, _ := os.ReadFile(flagOutputFile)
	var cfg map[string]any
	_ = json.Unmarshal(data, &cfg)
	if cfg["token_url"] != "https://sts.mtls.googleapis.com/v1/token" {
		t.Errorf("token_url = %v", cfg["token_url"])
	}
	if cfg["subject_token_type"] != "urn:ietf:params:oauth:token-type:mtls" {
		t.Errorf("subject_token_type = %v", cfg["subject_token_type"])
	}
	// Reset shared flags used elsewhere.
	flagCredCertPath = ""
	flagCredCertKeyPath = ""
}
