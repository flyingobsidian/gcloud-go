package gcloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

func testRunner(t *testing.T) *Runner {
	t.Helper()
	dir := t.TempDir()
	return NewRunner(Options{
		Getenv: func(key string) string {
			if key == "CLOUDSDK_CONFIG" {
				return dir
			}
			return ""
		},
		HomeDir: func() (string, error) { return dir, nil },
	})
}

func TestRunConcurrent(t *testing.T) {
	r := testRunner(t)
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			var stdout, stderr bytes.Buffer
			err := r.Run(context.Background(), Invocation{Args: []string{"--version"}, Stdout: &stdout, Stderr: &stderr})
			if err != nil {
				t.Error(err)
				return
			}
			if !strings.HasPrefix(stdout.String(), "gcloud-go ") || strings.Count(stdout.String(), "gcloud-go ") != 1 {
				t.Errorf("stdout = %q, want one version banner", stdout.String())
			}
			if stderr.Len() != 0 {
				t.Errorf("stderr = %q, want nothing", stderr.String())
			}
		})
	}
	wg.Wait()
}

func TestRunNotInProcess(t *testing.T) {
	r := testRunner(t)
	for _, args := range [][]string{
		{"secrets", "list"},
		{"--project", "p", "secrets", "list"},
		{"no-such-group"},
	} {
		if err := r.Run(context.Background(), Invocation{Args: args}); !errors.Is(err, ErrNotInProcess) {
			t.Errorf("Run(%q) = %v, want ErrNotInProcess", args, err)
		}
	}
}

func TestRunCommandError(t *testing.T) {
	r := testRunner(t)
	err := r.Run(context.Background(), Invocation{Args: []string{"--verbosity"}})
	var cerr *CommandError
	if !errors.As(err, &cerr) {
		t.Fatalf("Run = %v, want a *CommandError", err)
	}
	if cerr.Command != "gcloud" {
		t.Errorf("Command = %q, want %q", cerr.Command, "gcloud")
	}
	if want := "ERROR: (gcloud) flag needs an argument: --verbosity."; cerr.Error() != want {
		t.Errorf("Error() = %q, want %q", cerr.Error(), want)
	}
	if ExitCode(err) != 1 {
		t.Errorf("ExitCode = %d, want 1", ExitCode(err))
	}
}

func TestExitCode(t *testing.T) {
	if got := ExitCode(nil); got != 0 {
		t.Errorf("ExitCode(nil) = %d, want 0", got)
	}
	exitErr := exec.Command("sh", "-c", "exit 7").Run()
	if got := ExitCode(fmt.Errorf("ssh: %w", exitErr)); got != 7 {
		t.Errorf("ExitCode(exit 7) = %d, want 7", got)
	}
}

func TestVersion(t *testing.T) {
	if Version() == "" {
		t.Error("Version() is empty")
	}
}

func TestRunMigratedGroupInProcess(t *testing.T) {
	r := testRunner(t)
	var stderr bytes.Buffer
	err := r.Run(context.Background(), Invocation{Args: []string{"projects", "describe", "some-project"}, Stderr: &stderr})
	if errors.Is(err, ErrNotInProcess) {
		t.Fatal("projects describe was not run in-process")
	}
	var cerr *CommandError
	if !errors.As(err, &cerr) || cerr.Command != "gcloud.projects.describe" {
		t.Errorf("Run = %v, want a *CommandError from gcloud.projects.describe (there are no credentials)", err)
	}
}
