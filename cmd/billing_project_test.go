package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/flyingobsidian/gcloud-go/internal/config"
)

// TestResolveBillingProjectPriority walks the four-step priority chain from
// #1740 by swapping CLOUDSDK_CONFIG at a per-case tmpdir and toggling
// flagBillingProject.
func TestResolveBillingProjectPriority(t *testing.T) {
	saveFlag := flagBillingProject
	saveEnvCfg := os.Getenv("CLOUDSDK_CONFIG")
	defer func() {
		flagBillingProject = saveFlag
		if saveEnvCfg == "" {
			os.Unsetenv("CLOUDSDK_CONFIG")
		} else {
			os.Setenv("CLOUDSDK_CONFIG", saveEnvCfg)
		}
	}()

	// Fresh gcloud config dir for each subtest.
	newDir := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		t.Setenv("CLOUDSDK_CONFIG", dir)
		return dir
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
		flagBillingProject = "flag-project"
		if got := resolveBillingProject(); got != "flag-project" {
			t.Errorf("got %q, want flag-project", got)
		}
	})

	t.Run("config wins over ADC when no flag", func(t *testing.T) {
		dir := newDir(t)
		writeCfg(t, dir, "cfg-project")
		writeADC(t, dir, "adc-project")
		flagBillingProject = ""
		if got := resolveBillingProject(); got != "cfg-project" {
			t.Errorf("got %q, want cfg-project", got)
		}
	})

	t.Run("ADC used when only ADC is set", func(t *testing.T) {
		dir := newDir(t)
		writeADC(t, dir, "adc-project")
		flagBillingProject = ""
		if got := resolveBillingProject(); got != "adc-project" {
			t.Errorf("got %q, want adc-project", got)
		}
	})

	t.Run("empty when nothing is set", func(t *testing.T) {
		_ = newDir(t)
		flagBillingProject = ""
		if got := resolveBillingProject(); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

// TestBillingConfigSetGet exercises the [billing] section round-trip via
// getPropertyValue / setPropertyValue. Regression for the "unrecognized
// section: billing" error noted in the #1740 report.
func TestBillingConfigSetGet(t *testing.T) {
	props := &config.Properties{}

	// Set and read back via the same helpers used by `gcloud config set`.
	if err := setPropertyValue(props, "billing", "quota_project", "qp-set"); err != nil {
		t.Fatalf("setPropertyValue billing/quota_project: %v", err)
	}
	if got := props.Billing.QuotaProject; got != "qp-set" {
		t.Errorf("Billing.QuotaProject = %q, want qp-set", got)
	}
	got, err := getPropertyValue(props, "billing", "quota_project")
	if err != nil {
		t.Fatalf("getPropertyValue billing/quota_project: %v", err)
	}
	if got != "qp-set" {
		t.Errorf("get billing/quota_project = %q, want qp-set", got)
	}

	// Unknown key under billing should error clearly (not "unrecognized section").
	if _, err := getPropertyValue(props, "billing", "bogus"); err == nil ||
		err.Error() == "unrecognized section: billing" {
		t.Errorf("expected billing/bogus to error as an unrecognized property, got: %v", err)
	}
}
