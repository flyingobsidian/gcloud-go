package cli

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/flyingobsidian/gcloud-go/internal/config"
)

// Every top-level command group runs in-process, concurrently, with its
// output going only to the run's own streams (#2070).
func TestRunInProcessEveryGroup(t *testing.T) {
	env := config.MapEnv(map[string]string{"CLOUDSDK_CONFIG": t.TempDir()}, "/nonexistent")
	var wg sync.WaitGroup
	for _, c := range newRootCmd(legacyTestInvocation()).Commands() {
		name := c.Name()
		if name == "help" || name == "completion" {
			continue
		}
		wg.Go(func() {
			var stdout, stderr bytes.Buffer
			err := RunInProcess(context.Background(), InProcess{
				Args: []string{name, "--help"}, Stdout: &stdout, Stderr: &stderr, Env: env,
			})
			if errors.Is(err, ErrNotInProcess) {
				t.Errorf("%s: not run in-process", name)
				return
			}
			if err != nil {
				t.Errorf("%s --help: %v", name, err)
				return
			}
			if !bytes.Contains(stdout.Bytes(), []byte("gcloud "+name)) {
				t.Errorf("%s --help: stdout lacks the usage line:\n%s", name, stdout.String())
			}
		})
	}
	wg.Wait()
}
