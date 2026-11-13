package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/flyingobsidian/gcloud-go/internal/config"
)

// TestResolveBillingProjectPriority walks the four-step priority chain from
// #1740 with a per-case gcloud config directory and --billing-project value.
func TestResolveBillingProjectPriority(t *testing.T) {
	// Fresh gcloud config dir for each subtest.
	newDir := func(t *testing.T) string {
		t.Helper()
		return t.TempDir()
	}
	billingProject := func(t *testing.T, dir, flag string) string {
		t.Helper()
		inv, _, _ := newTestInvocation(t, map[string]string{"CLOUDSDK_CONFIG": dir}, "")
		inv.flags.billingProject = flag
		return inv.billingProject()
	}

	writeADC := func(t *testing.T, dir, project string) {
		t.Helper()
		if project == "" {
			return
		}
		body, _ := json.Marshal(map[string]any{"quota_project_id": project})
		if err := os.WriteFile(filepath.Join(dir, "application_default_credentials.json"), body, 0600); err != nil {
			t.Fatalf("writing ADC: %v", err)
		}
	}
	writeCfg := func(t *testing.T, dir, project string) {
		t.Helper()
		if project == "" {
			return
		}
		confDir := filepath.Join(dir, "configurations")
		if err := os.MkdirAll(confDir, 0700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		body := "[billing]\nquota_project = " + project + "\n"
		if err := os.WriteFile(filepath.Join(confDir, "config_default"), []byte(body), 0600); err != nil {
			t.Fatalf("writing config: %v", err)
		}
	}

	t.Run("flag wins over config and ADC", func(t *testing.T) {
		dir := newDir(t)
		writeCfg(t, dir, "cfg-project")
		writeADC(t, dir, "adc-project")
		if got := billingProject(t, dir, "flag-project"); got != "flag-project" {
			t.Errorf("got %q, want flag-project", got)
		}
	})

	t.Run("config wins over ADC when no flag", func(t *testing.T) {
		dir := newDir(t)
		writeCfg(t, dir, "cfg-project")
		writeADC(t, dir, "adc-project")
		if got := billingProject(t, dir, ""); got != "cfg-project" {
			t.Errorf("got %q, want cfg-project", got)
		}
	})

	t.Run("ADC used when only ADC is set", func(t *testing.T) {
		dir := newDir(t)
		writeADC(t, dir, "adc-project")
		if got := billingProject(t, dir, ""); got != "adc-project" {
			t.Errorf("got %q, want adc-project", got)
		}
	})

	t.Run("empty when nothing is set", func(t *testing.T) {
		dir := newDir(t)
		if got := billingProject(t, dir, ""); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

// TestBillingConfigSetGet exercises the [billing] section round-trip via
// `gcloud config set`. Regression for the "unrecognized section: billing"
// error noted in the #1740 report.
func TestBillingConfigSetGet(t *testing.T) {
	grp := newConfigCmds(legacyTestInvocation())
	t.Setenv("CLOUDSDK_CONFIG", t.TempDir())
	t.Setenv("CLOUDSDK_BILLING_QUOTA_PROJECT", "")

	if err := grp.runConfigSet(grp.configSetCmd, []string{"billing/quota_project", "qp-set"}); err != nil {
		t.Fatalf("config set billing/quota_project: %v", err)
	}
	props, err := config.OSEnv().Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := props.Billing.QuotaProject; got != "qp-set" {
		t.Errorf("Billing.QuotaProject = %q, want qp-set", got)
	}
	if got, err := config.OSEnv().Property("billing", "quota_project"); err != nil || got != "qp-set" {
		t.Errorf("Property(billing, quota_project) = %q, %v; want qp-set", got, err)
	}
}
