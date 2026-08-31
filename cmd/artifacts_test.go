package cmd

import (
	"testing"

	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

func TestParseDpkgOutput(t *testing.T) {
	output := "base-files\t12.4+deb12u5\tamd64\nlibc6\t2.36-9+deb12u4\tamd64\n"
	pkgs := parseDpkgOutput(output)
	if len(pkgs) != 2 {
		t.Fatalf("got %d packages, want 2", len(pkgs))
	}
	if pkgs[0].Package != "base-files" || pkgs[0].Version != "12.4+deb12u5" {
		t.Errorf("pkg[0] = %s %s, want base-files 12.4+deb12u5", pkgs[0].Package, pkgs[0].Version)
	}
	if pkgs[0].PackageType != "DEBIAN" {
		t.Errorf("pkg[0].PackageType = %q, want DEBIAN", pkgs[0].PackageType)
	}
	if pkgs[0].Architecture != "amd64" {
		t.Errorf("pkg[0].Architecture = %q, want amd64", pkgs[0].Architecture)
	}
}

func TestParseRpmOutput(t *testing.T) {
	output := "bash\t5.2.15-3.fc38\tx86_64\ncurl\t8.0.1-2.fc38\tx86_64\n"
	pkgs := parseRpmOutput(output)
	if len(pkgs) != 2 {
		t.Fatalf("got %d packages, want 2", len(pkgs))
	}
	if pkgs[0].Package != "bash" || pkgs[0].Version != "5.2.15-3.fc38" {
		t.Errorf("pkg[0] = %s %s, want bash 5.2.15-3.fc38", pkgs[0].Package, pkgs[0].Version)
	}
	if pkgs[0].PackageType != "RPM" {
		t.Errorf("pkg[0].PackageType = %q, want RPM", pkgs[0].PackageType)
	}
}

func TestParseApkOutput(t *testing.T) {
	output := "alpine-baselayout-3.4.3-r1\nbusybox-1.36.1-r2\nmusl-1.2.4-r2\n"
	pkgs := parseApkOutput(output)
	if len(pkgs) != 3 {
		t.Fatalf("got %d packages, want 3", len(pkgs))
	}
	if pkgs[0].Package != "alpine-baselayout" || pkgs[0].Version != "3.4.3-r1" {
		t.Errorf("pkg[0] = %q %q, want alpine-baselayout 3.4.3-r1", pkgs[0].Package, pkgs[0].Version)
	}
	if pkgs[0].PackageType != "APK" {
		t.Errorf("pkg[0].PackageType = %q, want APK", pkgs[0].PackageType)
	}
}

func TestSplitApkPackage(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantVer  string
	}{
		{"busybox-1.36.1-r2", "busybox", "1.36.1-r2"},
		{"alpine-baselayout-3.4.3-r1", "alpine-baselayout", "3.4.3-r1"},
		{"musl-1.2.4-r2", "musl", "1.2.4-r2"},
		{"libcrypto3-3.1.4-r5", "libcrypto3", "3.1.4-r5"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, ver := splitApkPackage(tt.input)
			if name != tt.wantName || ver != tt.wantVer {
				t.Errorf("splitApkPackage(%q) = (%q, %q), want (%q, %q)",
					tt.input, name, ver, tt.wantName, tt.wantVer)
			}
		})
	}
}

func TestParseDpkgOutputEmpty(t *testing.T) {
	pkgs := parseDpkgOutput("")
	if len(pkgs) != 0 {
		t.Errorf("got %d packages for empty input, want 0", len(pkgs))
	}
}

func TestArtRepoParseName(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		project  string
		location string
		id       string
		ok       bool
	}{
		{"valid", "projects/proj/locations/us/repositories/repo", "proj", "us", "repo", true},
		{"trimmed", "  projects/p/locations/us-central1/repositories/r  ", "p", "us-central1", "r", true},
		{"too short", "projects/p/locations/us", "", "", "", false},
		{"wrong prefix", "orgs/p/locations/us/repositories/r", "", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, l, id, ok := artRepoParseName(tc.input)
			if p != tc.project || l != tc.location || id != tc.id || ok != tc.ok {
				t.Errorf("artRepoParseName(%q) = (%q, %q, %q, %v), want (%q, %q, %q, %v)",
					tc.input, p, l, id, ok, tc.project, tc.location, tc.id, tc.ok)
			}
		})
	}
}

func TestArtRepoApplyRegistryURI(t *testing.T) {
	// Server-populated URI is preserved.
	r := &artifactregistry.Repository{
		Name:        "projects/p/locations/us/repositories/r",
		Format:      "DOCKER",
		RegistryUri: "server-set.example.com/p/r",
	}
	artRepoApplyRegistryURI(r)
	if r.RegistryUri != "server-set.example.com/p/r" {
		t.Errorf("existing RegistryUri clobbered: %q", r.RegistryUri)
	}

	// Missing URI is synthesized from name + format.
	r = &artifactregistry.Repository{
		Name:   "projects/p/locations/us-central1/repositories/my-repo",
		Format: "DOCKER",
	}
	artRepoApplyRegistryURI(r)
	if got, want := r.RegistryUri, "us-central1-docker.pkg.dev/p/my-repo"; got != want {
		t.Errorf("RegistryUri = %q, want %q", got, want)
	}

	// Legacy project IDs (with a colon) split into project/id in URLs.
	r = &artifactregistry.Repository{
		Name:   "projects/my:proj/locations/us/repositories/r",
		Format: "MAVEN",
	}
	artRepoApplyRegistryURI(r)
	if got, want := r.RegistryUri, "us-maven.pkg.dev/my/proj/r"; got != want {
		t.Errorf("RegistryUri = %q, want %q", got, want)
	}

	// nil is a no-op.
	artRepoApplyRegistryURI(nil)
}

func TestArtRepoApplyMode(t *testing.T) {
	cases := []struct {
		name    string
		flag    string
		want    string
		wantErr bool
	}{
		{"empty leaves untouched", "", "PREEXISTING", false},
		{"connector maps to remote", "CONNECTOR-REPOSITORY", "REMOTE_REPOSITORY", false},
		{"case insensitive", "standard-repository", "STANDARD_REPOSITORY", false},
		{"none maps to unspecified", "NONE", "MODE_UNSPECIFIED", false},
		{"unknown errors", "bogus", "PREEXISTING", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := &artifactregistry.Repository{Mode: "PREEXISTING"}
			err := artRepoApplyMode(body, tc.flag)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if body.Mode != tc.want {
				t.Errorf("Mode = %q, want %q", body.Mode, tc.want)
			}
		})
	}
}

func TestArtifactsRepositoriesCreateHasModeFlag(t *testing.T) {
	if artifactsRepositoriesCreateCmd.Flag("mode") == nil {
		t.Fatal("artifacts repositories create missing --mode flag")
	}
}

func TestArtifactsDockerImagesListHasLimitAndPageSize(t *testing.T) {
	for _, name := range []string{"limit", "page-size"} {
		if artifactsDockerImagesListCmd.Flag(name) == nil {
			t.Errorf("artifacts docker images list missing --%s flag", name)
		}
	}
}

// TestArtifactsStubSubcommandsRegistered checks the subgroups added in #537
// are present on the artifacts command.
func TestArtifactsStubSubcommandsRegistered(t *testing.T) {
	want := []string{
		"apt", "attachments", "docker", "files", "generic", "go",
		"image-streaming-cache", "locations", "operations", "packages",
		"print-settings", "projects", "repositories", "rules", "sbom",
		"settings", "tags", "versions", "vpcsc-config", "vulnerabilities", "yum",
	}
	got := map[string]bool{}
	for _, c := range artifactsCmd.Commands() {
		got[c.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("artifacts subcommand %q not registered", name)
		}
	}
}
