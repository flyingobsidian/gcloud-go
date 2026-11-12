package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
)

// newTestInvocation returns an invocation that reads stdin and the
// variables in vars, and writes to the returned buffers. Its gcloud
// configuration directory is a fresh temporary directory unless vars sets
// CLOUDSDK_CONFIG.
func newTestInvocation(t *testing.T, vars map[string]string, stdin string) (*invocation, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	all := map[string]string{"CLOUDSDK_CONFIG": t.TempDir()}
	for k, v := range vars {
		all[k] = v
	}
	env := config.MapEnv(all, "/nonexistent")
	var stdout, stderr bytes.Buffer
	return &invocation{
		stdin:  strings.NewReader(stdin),
		stdout: &stdout,
		stderr: &stderr,
		env:    env,
		tokens: auth.NewTokenProvider(env),
	}, &stdout, &stderr
}

func TestRootCmdInvocationsAreIndependent(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			inv, stdout, _ := newTestInvocation(t, nil, "")
			root := newRootCmd(inv, false)
			root.AddCommand(&cobra.Command{
				Use: "probe",
				RunE: func(c *cobra.Command, _ []string) error {
					project, err := inv.project()
					if err != nil {
						return err
					}
					fmt.Fprintf(inv.stdout, "%s %s", project, inv.flags.format)
					return nil
				},
			})
			want := fmt.Sprintf("project-%02d json%d", i, i)
			root.SetArgs([]string{"probe", "--project", fmt.Sprintf("project-%02d", i), "--format", fmt.Sprintf("json%d", i)})
			if err := root.ExecuteContext(context.Background()); err != nil {
				t.Error(err)
				return
			}
			if got := stdout.String(); got != want {
				t.Errorf("output = %q, want %q", got, want)
			}
		})
	}
	wg.Wait()
}

func TestInvocationProjectResolution(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "configurations"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configurations", "config_default"),
		[]byte("[core]\nproject = from-config\n[compute]\nzone = config-zone\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLOUDSDK_CORE_PROJECT", "from-process-env")

	tests := []struct {
		name string
		flag string
		vars map[string]string
		want string
	}{
		{"config", "", map[string]string{"CLOUDSDK_CONFIG": dir}, "from-config"},
		{"env", "", map[string]string{"CLOUDSDK_CONFIG": dir, "CLOUDSDK_CORE_PROJECT": "from-env"}, "from-env"},
		{"flag", "from-flag", map[string]string{"CLOUDSDK_CONFIG": dir, "CLOUDSDK_CORE_PROJECT": "from-env"}, "from-flag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, _, _ := newTestInvocation(t, tt.vars, "")
			inv.flags.project = tt.flag
			got, err := inv.project()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("project() = %q, want %q", got, tt.want)
			}
			_, zone, err := inv.projectZone()
			if err != nil {
				t.Fatal(err)
			}
			if zone != "config-zone" {
				t.Errorf("zone = %q, want %q", zone, "config-zone")
			}
		})
	}
}

func TestInvocationProjectMissing(t *testing.T) {
	inv, _, _ := newTestInvocation(t, nil, "typed-project\n")
	if _, err := inv.project(); err == nil {
		t.Error("project() succeeded with no project and no terminal, want error")
	}

	inv.interactive = true
	got, err := inv.project()
	if err != nil {
		t.Fatal(err)
	}
	if got != "typed-project" {
		t.Errorf("project() = %q, want %q", got, "typed-project")
	}
	if out := inv.stdout.(*bytes.Buffer).String(); out != "Enter project: " {
		t.Errorf("prompt = %q, want %q", out, "Enter project: ")
	}
}

func TestInvocationHTTPDebugWriter(t *testing.T) {
	inv, _, stderr := newTestInvocation(t, nil, "")
	if w := inv.httpDebugWriter(); w != nil {
		t.Errorf("httpDebugWriter() = %v with logging off, want nil", w)
	}
	if c := inv.clients(); c.DebugOut != nil {
		t.Errorf("clients().DebugOut = %v with logging off, want nil", c.DebugOut)
	}
	inv.flags.verbosity = "DEBUG"
	if w := inv.httpDebugWriter(); w != stderr {
		t.Errorf("httpDebugWriter() = %v with --verbosity=debug, want stderr", w)
	}
	inv.flags.verbosity = "warning"
	inv.flags.logHTTP = true
	inv.flags.account = "user@example.com"
	c := inv.clients()
	if c.DebugOut != stderr || c.Account != "user@example.com" || c.Tokens != inv.tokens {
		t.Errorf("clients() = %+v, want stderr logging, the --account and the invocation's tokens", c)
	}
}

func TestRootCmdSetsLegacyGlobals(t *testing.T) {
	saved := flagProject
	t.Cleanup(func() { flagProject = saved })

	inv, _, _ := newTestInvocation(t, nil, "")
	root := newRootCmd(inv, true)
	var seen string
	probe := &cobra.Command{
		Use: "legacy-probe",
		Run: func(*cobra.Command, []string) { seen = flagProject },
	}
	root.AddCommand(probe)
	root.SetArgs([]string{"legacy-probe", "--project", "legacy-project"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if seen != "legacy-project" {
		t.Errorf("flagProject = %q in a legacy command, want %q", seen, "legacy-project")
	}
}

func TestRootCmdAdoptsLegacyGroupsOnlyWhenAsked(t *testing.T) {
	inv, _, _ := newTestInvocation(t, nil, "")
	has := func(root *cobra.Command) bool {
		for _, c := range root.Commands() {
			if c.Name() == "secrets" {
				return true
			}
		}
		return false
	}
	if has(newRootCmd(inv, false)) {
		t.Error("root without legacy groups has the secrets group")
	}
	if !has(newRootCmd(inv, true)) {
		t.Error("root with legacy groups lacks the secrets group")
	}
}
