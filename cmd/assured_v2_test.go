package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// assuredFreshCmd returns a throwaway command with the create/update flags
// registered, so Changed() reflects only the flags set by the test.
func assuredFreshCmd(t *testing.T, flags ...string) *cobra.Command {
	t.Helper()
	c := &cobra.Command{Use: "x"}
	for _, f := range []string{"compliance-regime", "billing-account", "display-name", "framework",
		"framework-major-revision", "target-project", "target-folder", "new-project-config",
		"new-folder-config", "cloud-control-config", "cloud-control-configs-from-file",
		"cmek-key-ring", "cmek-dedicated-project-config", "description", "workload-id",
		"external-identifier", "async", "etag"} {
		c.Flags().String(f, "", "")
	}
	for _, f := range flags {
		if err := c.Flags().Set(f, "x"); err != nil {
			t.Fatal(err)
		}
	}
	return c
}

func TestAssuredValidateCreateFlags(t *testing.T) {
	cases := []struct {
		version string
		flags   []string
		args    []string
		wantErr string
	}{
		{"v1", []string{"display-name", "compliance-regime", "billing-account"}, nil, ""},
		{"v1", []string{"display-name", "compliance-regime"}, nil, "[--billing-account]: The billing account is required for V1 workloads."},
		{"v1", []string{"display-name", "compliance-regime", "billing-account", "framework"}, nil,
			"Invalid value for [--framework]: Flag cannot be used with --version=v1. Use '--compliance-regime' when creating V1 workloads."},
		{"v1", []string{"display-name", "compliance-regime", "billing-account"}, []string{"w"}, "Positional workload ID is only supported in V2"},
		{"v2", []string{"framework", "target-project"}, []string{"w"}, ""},
		{"v2", []string{"target-project"}, nil, "[--framework]: The compliance framework is required for V2 workloads."},
		{"v2", []string{"framework"}, nil, "One of --target-project, --target-folder, --new-project-config, or --new-folder-config"},
		{"v2", []string{"framework", "target-folder", "display-name"}, nil,
			"Invalid value for [--display-name]: Flag cannot be used with --version=v2."},
		{"v2", []string{"framework", "target-folder", "workload-id"}, []string{"w"}, "At most one of WORKLOAD_ID"},
	}
	for _, tc := range cases {
		err := assuredValidateCreateFlags(assuredFreshCmd(t, tc.flags...), tc.args, tc.version)
		if tc.wantErr == "" && err != nil {
			t.Errorf("%s %v: unexpected error %v", tc.version, tc.flags, err)
		}
		if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
			t.Errorf("%s %v: error %v, want %q", tc.version, tc.flags, err, tc.wantErr)
		}
	}
}

func TestAssuredValidateUpdateAndDeleteFlags(t *testing.T) {
	if err := assuredValidateUpdateFlags(assuredFreshCmd(t, "framework-major-revision"), "v2"); err == nil ||
		!strings.Contains(err.Error(), "[--cloud-control-config]") {
		t.Errorf("v2 update without controls: %v", err)
	}
	if err := assuredValidateUpdateFlags(assuredFreshCmd(t, "cloud-control-config", "async"), "v2"); err != nil {
		t.Errorf("v2 update: %v", err)
	}
	if err := assuredValidateUpdateFlags(assuredFreshCmd(t, "async", "display-name"), "v1"); err == nil ||
		!strings.Contains(err.Error(), "[--async]") {
		t.Errorf("v1 update with --async: %v", err)
	}
	if err := assuredValidateUpdateFlags(assuredFreshCmd(t), "v1"); err == nil {
		t.Error("v1 update with no fields: expected error")
	}
	if err := assuredValidateDeleteFlags(assuredFreshCmd(t, "async"), "v1"); err == nil {
		t.Error("v1 delete with --async: expected error")
	}
	if err := assuredValidateDeleteFlags(assuredFreshCmd(t, "async"), "v2"); err != nil {
		t.Errorf("v2 delete with --async: %v", err)
	}
}

func TestAssuredCloudControlConfigs(t *testing.T) {
	t.Cleanup(func() { flagAssuredCloudControlCfg, flagAssuredCloudControlFile = nil, "" })

	flagAssuredCloudControlCfg = []string{"cloud_control=c1,revision=2,enforcement_mode=detective"}
	got, err := assuredCloudControlConfigs()
	if err != nil {
		t.Fatal(err)
	}
	want := []map[string]any{{"cloudControl": "c1", "cloudControlRevision": int64(2), "enforcementMode": "DETECTIVE"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("inline: got %v, want %v", got, want)
	}

	f := filepath.Join(t.TempDir(), "controls.yaml")
	src := `- cloudControl: c2
  cloudControlRevision: 1
  enforcement-mode: preventive
  parameters:
    flag: true
    count: 3
    regions: [us, eu]
    choice: {name: opt, value: x}
`
	if err := os.WriteFile(f, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	flagAssuredCloudControlCfg, flagAssuredCloudControlFile = nil, f
	got, err = assuredCloudControlConfigs()
	if err != nil {
		t.Fatal(err)
	}
	want = []map[string]any{{
		"cloudControl": "c2", "cloudControlRevision": int64(1), "enforcementMode": "PREVENTIVE",
		"parameters": []map[string]any{
			{"name": "choice", "parameterValue": map[string]any{"oneofValue": map[string]any{
				"name": "opt", "parameterValue": map[string]any{"stringValue": "x"}}}},
			{"name": "count", "parameterValue": map[string]any{"numberValue": float64(3)}},
			{"name": "flag", "parameterValue": map[string]any{"boolValue": true}},
			{"name": "regions", "parameterValue": map[string]any{"stringListValue": map[string]any{"values": []string{"us", "eu"}}}},
		},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("file: got %v\nwant %v", got, want)
	}

	flagAssuredCloudControlFile = ""
	flagAssuredCloudControlCfg = []string{"cloud_control=c1,revision=2,enforcement_mode=bogus"}
	if _, err := assuredCloudControlConfigs(); err == nil || !strings.Contains(err.Error(), "Invalid enforcement mode: bogus") {
		t.Errorf("bad mode: %v", err)
	}
}

func TestAssuredResourceAndCmekConfig(t *testing.T) {
	t.Cleanup(func() {
		flagAssuredTargetProject, flagAssuredNewProjectConfig = "", nil
		flagAssuredCmekKeyRing, flagAssuredCmekDedicated = "", nil
	})
	flagAssuredNewProjectConfig = map[string]string{"parent": "folders/1", "display_name": "P", "billing_account": "billingAccounts/B"}
	rc, err := assuredResourceConfig()
	if err != nil || !reflect.DeepEqual(rc, map[string]any{"newProjectConfig": map[string]any{
		"parent": "folders/1", "displayName": "P", "billingAccount": "billingAccounts/B"}}) {
		t.Errorf("new project config = %v, %v", rc, err)
	}
	flagAssuredNewProjectConfig = map[string]string{"parent": "folders/1"}
	if _, err := assuredResourceConfig(); err == nil {
		t.Error("expected error for incomplete new-project-config")
	}

	flagAssuredCmekKeyRing = "projects/k/locations/us/keyRings/r"
	flagAssuredCmekDedicated = map[string]string{"billing_account": "billingAccounts/B"}
	if _, err := assuredCmekConfig(false); err == nil {
		t.Error("expected error for dedicated CMEK project on non-folder target")
	}
	cfg, err := assuredCmekConfig(true)
	if err != nil || !reflect.DeepEqual(cfg, map[string]any{"keyRingId": flagAssuredCmekKeyRing,
		"dedicatedProjectConfig": map[string]any{"billingAccount": "billingAccounts/B"}}) {
		t.Errorf("cmek config = %v, %v", cfg, err)
	}
}

func TestEmitAssuredV2List(t *testing.T) {
	var out, errOut bytes.Buffer
	all := []map[string]any{{
		"name": "organizations/1/locations/us/workloads/w1", "targetResourceDisplayName": "T",
		"computedTargetResource": "projects/9", "framework": map[string]any{"framework": "fedramp"},
		"state": "ACTIVE", "createTime": "2026-01-01T00:00:00Z",
	}}
	if err := emitAssuredV2List(&out, &errOut, all, ""); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "ID") ||
		strings.Join(strings.Fields(lines[0]), " ") != "ID DISPLAY_NAME TARGET_RESOURCE FRAMEWORK STATE CREATE_TIME" ||
		strings.Join(strings.Fields(lines[1]), " ") != "w1 T projects/9 fedramp ACTIVE 2026-01-01T00:00:00Z" {
		t.Errorf("table = %q", out.String())
	}
	out.Reset()
	if err := emitAssuredV2List(&out, &errOut, nil, ""); err != nil || errOut.String() != "Listed 0 items.\n" {
		t.Errorf("empty: %v %q", err, errOut.String())
	}
}

func TestAssuredWorkloadRef(t *testing.T) {
	t.Cleanup(func() { flagAssuredOrg, flagAssuredLocation = "", "" })
	if _, err := assuredWorkloadRef("w"); err == nil {
		t.Error("expected error without --organization/--location")
	}
	full := "organizations/1/locations/us-central1/workloads/w"
	if got, err := assuredWorkloadRef(full); err != nil || got != full {
		t.Errorf("full name = %q, %v", got, err)
	}
	if got := assuredRegionOf(full); got != "us-central1" {
		t.Errorf("assuredRegionOf = %q", got)
	}
}
