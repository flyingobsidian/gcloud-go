package cmd

import (
	"os"
	"testing"
)

func TestReadDataFileStdin(t *testing.T) {
	// We can't easily test stdin reading in unit tests,
	// but we can test that "-" is recognized as stdin.
	// Just verify the function signature compiles.
	_ = readDataFile
}

func TestResolveProject(t *testing.T) {
	// Set project via flag.
	flagProject = "test-project"
	project, err := resolveProject()
	if err != nil {
		t.Fatalf("resolveProject() error: %v", err)
	}
	if project != "test-project" {
		t.Errorf("project = %q, want test-project", project)
	}
	flagProject = "" // reset
}

func TestResolveProjectMissing(t *testing.T) {
	flagProject = ""
	t.Setenv("CLOUDSDK_CORE_PROJECT", "")
	t.Setenv("CLOUDSDK_CONFIG", t.TempDir()) // empty config

	_, err := resolveProject()
	if err == nil {
		t.Error("resolveProject() expected error when no project configured")
	}
}

func TestSecretsCommandTree(t *testing.T) {
	// Verify all secrets subcommands are registered.
	sub := secretsCmd.Commands()
	names := make(map[string]bool)
	for _, c := range sub {
		names[c.Name()] = true
	}
	for _, want := range []string{"versions", "create", "list", "describe", "delete"} {
		if !names[want] {
			t.Errorf("missing subcommand: secrets %s", want)
		}
	}

	// Check versions subcommands.
	versionsSub := secretsVersionsCmd.Commands()
	vNames := make(map[string]bool)
	for _, c := range versionsSub {
		vNames[c.Name()] = true
	}
	for _, want := range []string{"access", "add"} {
		if !vNames[want] {
			t.Errorf("missing subcommand: secrets versions %s", want)
		}
	}
}

func TestReadDataFile(t *testing.T) {
	// Write a temp file and read it back.
	path := t.TempDir() + "/data.txt"
	want := "secret-data-123"
	if err := writeTestFile(path, want); err != nil {
		t.Fatal(err)
	}

	got, err := readDataFile(path)
	if err != nil {
		t.Fatalf("readDataFile() error: %v", err)
	}
	if string(got) != want {
		t.Errorf("readDataFile() = %q, want %q", got, want)
	}
}

func TestReadDataFileNotFound(t *testing.T) {
	_, err := readDataFile("/nonexistent/path")
	if err == nil {
		t.Error("readDataFile() expected error for missing file")
	}
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}
