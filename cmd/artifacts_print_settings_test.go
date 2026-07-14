package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

func TestArtifactsPrintSettingsSubcommandsRegistered(t *testing.T) {
	// print-settings must be a real group (not a stub command) with the four
	// language subcommands exposed by gcloud-python.
	var ps *cobra.Command
	for _, c := range artifactsCmd.Commands() {
		if c.Name() == "print-settings" {
			ps = c
			break
		}
	}
	if ps == nil {
		t.Fatal("artifacts print-settings not registered")
	}
	want := []string{"gradle", "mvn", "npm", "python"}
	got := map[string]bool{}
	for _, c := range ps.Commands() {
		got[c.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("artifacts print-settings %s subcommand not registered", name)
		}
	}
}

func TestApplyReplacements(t *testing.T) {
	tmpl := "hello {name}, your key is {key}"
	got := applyReplacements(tmpl, map[string]string{
		"{name}": "world",
		"{key}":  "abc",
	})
	if got != "hello world, your key is abc" {
		t.Errorf("got %q", got)
	}
}

func TestApplyReplacementsPreservesLiteralValueWithBraces(t *testing.T) {
	// Values that contain { ... } tokens must not be re-expanded.
	tmpl := "prefix {password} suffix"
	got := applyReplacements(tmpl, map[string]string{
		"{password}": "{location}",
	})
	if got != "prefix {location} suffix" {
		t.Errorf("got %q, want value to be inserted verbatim without re-expansion", got)
	}
}

func TestApplyReplacementsUnknownTokenLeftIntact(t *testing.T) {
	got := applyReplacements("{a}{b}{a}", map[string]string{"{a}": "X"})
	if got != "X{b}X" {
		t.Errorf("got %q, want X{b}X", got)
	}
}

func TestRenderGradleSnippetNoSAGa(t *testing.T) {
	in := &printSettingsInputs{project: "my-proj", location: "us-central1", repository: "my-repo"}
	got := renderGradleSnippet(in, nil, "")
	if !strings.Contains(got, `url "artifactregistry://us-central1-maven.pkg.dev/my-proj/my-repo"`) {
		t.Errorf("gradle snippet missing expected repo URL: %s", got)
	}
	if !strings.Contains(got, `version "`+artifactRegistryExtensionVersion+`"`) {
		t.Errorf("gradle snippet missing extension version %q", artifactRegistryExtensionVersion)
	}
}

func TestRenderGradleSnippetWithSA(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	got := renderGradleSnippet(in, nil, "SA_CREDS_B64")
	if !strings.Contains(got, `def artifactRegistryMavenSecret = "SA_CREDS_B64"`) {
		t.Errorf("gradle snippet missing SA secret line: %s", got)
	}
	if !strings.Contains(got, `username = "_json_key_base64"`) {
		t.Errorf("gradle snippet missing SA username: %s", got)
	}
	if !strings.Contains(got, `url "https://us-maven.pkg.dev/p/r"`) {
		t.Errorf("gradle snippet missing SA URL scheme: %s", got)
	}
}

func TestRenderGradleSnippetSnapshotPolicy(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	cfg := &artifactregistry.MavenRepositoryConfig{VersionPolicy: "SNAPSHOT"}
	got := renderGradleSnippet(in, cfg, "")
	if !strings.Contains(got, `def snapshotURL = "artifactregistry://us-maven.pkg.dev/p/r"`) {
		t.Errorf("gradle SNAPSHOT snippet missing snapshotURL: %s", got)
	}
	if !strings.Contains(got, `<Paste release URL here>`) {
		t.Errorf("gradle SNAPSHOT snippet missing release placeholder: %s", got)
	}
}

func TestRenderMvnSnippetNoSA(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	got := renderMavenSnippet(in, nil, "")
	if !strings.Contains(got, `<url>artifactregistry://us-maven.pkg.dev/p/r</url>`) {
		t.Errorf("mvn snippet missing artifactregistry scheme URL: %s", got)
	}
	if !strings.Contains(got, `<version>`+artifactRegistryExtensionVersion+`</version>`) {
		t.Errorf("mvn snippet missing extension version: %s", got)
	}
}

func TestRenderMvnSnippetWithSA(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	got := renderMavenSnippet(in, nil, "SA_CREDS_B64")
	if !strings.Contains(got, `<url>https://us-maven.pkg.dev/p/r</url>`) {
		t.Errorf("mvn SA snippet missing https URL: %s", got)
	}
	if !strings.Contains(got, `<password>SA_CREDS_B64</password>`) {
		t.Errorf("mvn SA snippet missing password: %s", got)
	}
	if !strings.Contains(got, `<username>_json_key_base64</username>`) {
		t.Errorf("mvn SA snippet missing username: %s", got)
	}
}

func TestRenderMvnSnippetReleasePolicy(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	cfg := &artifactregistry.MavenRepositoryConfig{VersionPolicy: "RELEASE"}
	got := renderMavenSnippet(in, cfg, "")
	if strings.Contains(got, `<snapshotRepository>`) {
		t.Errorf("mvn RELEASE snippet unexpectedly contains snapshotRepository: %s", got)
	}
	if !strings.Contains(got, `<snapshots>
        <enabled>false</enabled>
      </snapshots>`) {
		t.Errorf("mvn RELEASE snippet should disable snapshots: %s", got)
	}
}

func TestRenderNpmSnippetNoSA(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	got := renderNpmSnippet(in, "", "")
	if !strings.Contains(got, "registry=https://us-npm.pkg.dev/p/r/") {
		t.Errorf("npm snippet missing registry line: %s", got)
	}
	if !strings.Contains(got, "//us-npm.pkg.dev/p/r/:always-auth=true") {
		t.Errorf("npm snippet missing always-auth: %s", got)
	}
	if strings.Contains(got, "_password") {
		t.Errorf("npm no-SA snippet must not contain _password: %s", got)
	}
}

func TestRenderNpmSnippetWithScope(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	got := renderNpmSnippet(in, "@acme", "")
	if !strings.Contains(got, "@acme:registry=https://us-npm.pkg.dev/p/r/") {
		t.Errorf("npm snippet missing scoped registry: %s", got)
	}
}

func TestRenderNpmSnippetWithSA(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	got := renderNpmSnippet(in, "", "PWD")
	if !strings.Contains(got, `//us-npm.pkg.dev/p/r/:_password="PWD"`) {
		t.Errorf("npm SA snippet missing password: %s", got)
	}
	if !strings.Contains(got, "//us-npm.pkg.dev/p/r/:username=_json_key_base64") {
		t.Errorf("npm SA snippet missing username: %s", got)
	}
}

func TestRenderPythonSnippetNoSA(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	got := renderPythonSnippet(in, "")
	if !strings.Contains(got, "[r]") {
		t.Errorf("python snippet missing [repo] section: %s", got)
	}
	if !strings.Contains(got, "repository: https://us-python.pkg.dev/p/r/") {
		t.Errorf("python snippet missing repository URL: %s", got)
	}
	if !strings.Contains(got, "extra-index-url = https://us-python.pkg.dev/p/r/simple/") {
		t.Errorf("python snippet missing pip extra-index-url: %s", got)
	}
	if strings.Contains(got, "_json_key_base64") {
		t.Errorf("python no-SA snippet must not include _json_key_base64: %s", got)
	}
}

func TestRenderPythonSnippetWithSA(t *testing.T) {
	in := &printSettingsInputs{project: "p", location: "us", repository: "r"}
	got := renderPythonSnippet(in, "PWD")
	if !strings.Contains(got, "username: _json_key_base64") {
		t.Errorf("python SA snippet missing username: %s", got)
	}
	if !strings.Contains(got, "password: PWD") {
		t.Errorf("python SA snippet missing password: %s", got)
	}
	if !strings.Contains(got, "extra-index-url = https://_json_key_base64:PWD@us-python.pkg.dev/p/r/simple/") {
		t.Errorf("python SA snippet missing embedded creds in pip URL: %s", got)
	}
}

func TestSelectGradleTemplateMatrix(t *testing.T) {
	tests := []struct {
		policy string
		hasSA  bool
		want   string
	}{
		{"", false, gradleNoServiceAccountTemplate},
		{"", true, gradleServiceAccountTemplate},
		{"SNAPSHOT", false, gradleNoServiceAccountSnapshotTemplate},
		{"SNAPSHOT", true, gradleServiceAccountSnapshotTemplate},
		{"RELEASE", false, gradleNoServiceAccountReleaseTemplate},
		{"RELEASE", true, gradleServiceAccountReleaseTemplate},
		{"VERSION_POLICY_UNSPECIFIED", false, gradleNoServiceAccountTemplate},
	}
	for _, tt := range tests {
		var cfg *artifactregistry.MavenRepositoryConfig
		if tt.policy != "" {
			cfg = &artifactregistry.MavenRepositoryConfig{VersionPolicy: tt.policy}
		}
		sa := ""
		if tt.hasSA {
			sa = "x"
		}
		if got := selectGradleTemplate(cfg, sa); got != tt.want {
			t.Errorf("selectGradleTemplate(policy=%q, hasSA=%v) picked wrong template", tt.policy, tt.hasSA)
		}
	}
}

func TestSelectMavenTemplateMatrix(t *testing.T) {
	tests := []struct {
		policy string
		hasSA  bool
		want   string
	}{
		{"", false, mvnNoServiceAccountTemplate},
		{"", true, mvnServiceAccountTemplate},
		{"SNAPSHOT", false, mvnNoServiceAccountSnapshotTemplate},
		{"SNAPSHOT", true, mvnServiceAccountSnapshotTemplate},
		{"RELEASE", false, mvnNoServiceAccountReleaseTemplate},
		{"RELEASE", true, mvnServiceAccountReleaseTemplate},
	}
	for _, tt := range tests {
		var cfg *artifactregistry.MavenRepositoryConfig
		if tt.policy != "" {
			cfg = &artifactregistry.MavenRepositoryConfig{VersionPolicy: tt.policy}
		}
		sa := ""
		if tt.hasSA {
			sa = "x"
		}
		if got := selectMavenTemplate(cfg, sa); got != tt.want {
			t.Errorf("selectMavenTemplate(policy=%q, hasSA=%v) picked wrong template", tt.policy, tt.hasSA)
		}
	}
}
