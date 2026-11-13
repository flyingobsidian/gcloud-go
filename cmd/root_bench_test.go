package cmd

import (
	"context"
	"io"
	"testing"

	"github.com/flyingobsidian/gcloud-go/internal/config"
)

// BenchmarkNewRootCmd measures building the full command tree, which every
// invocation does (#2063).
func BenchmarkNewRootCmd(b *testing.B) {
	inv, _, _ := newTestInvocation(b, nil, "")
	b.ReportAllocs()
	for b.Loop() {
		newRootCmd(inv)
	}
}

// BenchmarkRunInProcessVersion measures a whole in-process run of a cheap
// command, tree building included.
func BenchmarkRunInProcessVersion(b *testing.B) {
	env := config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": b.TempDir()}, "/nonexistent")
	b.ReportAllocs()
	for b.Loop() {
		if err := RunInProcess(context.Background(), InProcess{Args: []string{"version"}, Stdout: io.Discard, Env: env}); err != nil {
			b.Fatal(err)
		}
	}
}
