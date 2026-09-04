package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

// TestSccManageServicesSubgroup checks the surface: `scc manage services`
// exists under `scc manage`, exposes describe/list/update, and nothing else.
// Without this test a renamed or missing subcommand would slip through
// unnoticed until a user tried to run it.
func TestSccManageServicesSubgroup(t *testing.T) {
	manage := sccSubgroup("manage")
	if manage == nil {
		t.Fatal("scc manage missing")
	}
	// The list of subgroups also expands with services, so re-check it.
	assertSubcommands(t, manage, []string{"settings", "services"})
	services := sccNestedSubgroup("manage", "services")
	if services == nil {
		t.Fatal("scc manage services missing")
	}
	assertExactSubcommands(t, services, []string{"describe", "list", "update"})
}

func TestSccResolveServiceName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		// Full name.
		{"security-health-analytics", "security-health-analytics"},
		// Uppercased full name.
		{"Security-Health-Analytics", "security-health-analytics"},
		// Abbreviation.
		{"sha", "security-health-analytics"},
		// The gcloud-python 572.0.0 addition (#1820).
		{"external-exposure", "external-exposure"},
		{"ee", "external-exposure"},
		// Whitespace tolerated.
		{"  vm-manager  ", "vm-manager"},
		{"vmm", "vm-manager"},
		// Multi-hyphen abbreviations.
		{"ec2-va", "ec2-vulnerability-assessment"},
		// Unknown.
		{"totally-made-up", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := sccResolveServiceName(tc.in); got != tc.want {
			t.Errorf("sccResolveServiceName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestSccSupportedServicesIncludesExternalExposure is the direct regression
// test for #1820: the gcloud-python 572.0.0 tuple entry must appear in our
// mirror.
func TestSccSupportedServicesIncludesExternalExposure(t *testing.T) {
	found := false
	for _, s := range sccSupportedServices {
		if s.Name == "external-exposure" {
			if s.Abbrev != "ee" {
				t.Errorf("external-exposure abbreviation = %q, want %q", s.Abbrev, "ee")
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("external-exposure missing from sccSupportedServices (#1820)")
	}
}

func TestSccManageParseEnablementState(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"ENABLED", "ENABLED", false},
		{"enabled", "ENABLED", false},
		{"Disabled", "DISABLED", false},
		{"inherited", "INHERITED", false},
		// Empty stays empty (unset flag).
		{"", "", false},
		// Unknown values rejected.
		{"muted", "", true},
		{"unknown", "", true},
	}
	for _, tc := range cases {
		got, err := sccManageParseEnablementState(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("sccManageParseEnablementState(%q) succeeded, want error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("sccManageParseEnablementState(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("sccManageParseEnablementState(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSccManageBuildUpdateMask(t *testing.T) {
	cases := []struct {
		state, modules bool
		want           string
		wantErr        bool
	}{
		{true, true, "intended_enablement_state,modules", false},
		{true, false, "intended_enablement_state", false},
		{false, true, "modules", false},
		{false, false, "", true},
	}
	for _, tc := range cases {
		got, err := sccManageBuildUpdateMask(tc.state, tc.modules)
		if tc.wantErr {
			if err == nil {
				t.Errorf("sccManageBuildUpdateMask(%v,%v) succeeded, want error", tc.state, tc.modules)
			}
			continue
		}
		if err != nil {
			t.Errorf("sccManageBuildUpdateMask(%v,%v): %v", tc.state, tc.modules, err)
		}
		if got != tc.want {
			t.Errorf("sccManageBuildUpdateMask(%v,%v) = %q, want %q", tc.state, tc.modules, got, tc.want)
		}
	}
}

// TestSccManageServicesParent covers the four ways of supplying scope:
// --organization, --folder, --project, and --parent, plus the mutex rule
// between --parent and the others.
func TestSccManageServicesParent(t *testing.T) {
	saveOrg, saveFolder, saveProj, saveParent :=
		flagSccOrg, flagSccFolder, flagSccProject, flagSccManageParent
	defer func() {
		flagSccOrg, flagSccFolder, flagSccProject, flagSccManageParent =
			saveOrg, saveFolder, saveProj, saveParent
	}()

	// --organization
	flagSccOrg = "1057127509270"
	flagSccFolder, flagSccProject, flagSccManageParent = "", "", ""
	got, err := sccManageServicesParent()
	if err != nil {
		t.Fatalf("--organization: %v", err)
	}
	if got != "organizations/1057127509270/locations/global" {
		t.Errorf("--organization parent = %q", got)
	}

	// --folder
	flagSccOrg, flagSccFolder, flagSccProject, flagSccManageParent = "", "42", "", ""
	got, err = sccManageServicesParent()
	if err != nil {
		t.Fatalf("--folder: %v", err)
	}
	if got != "folders/42/locations/global" {
		t.Errorf("--folder parent = %q", got)
	}

	// --project (id only)
	flagSccOrg, flagSccFolder, flagSccProject, flagSccManageParent = "", "", "my-proj", ""
	got, err = sccManageServicesParent()
	if err != nil {
		t.Fatalf("--project: %v", err)
	}
	if got != "projects/my-proj/locations/global" {
		t.Errorf("--project parent = %q", got)
	}

	// --parent (organizations/…)
	flagSccOrg, flagSccFolder, flagSccProject, flagSccManageParent = "", "", "", "organizations/9"
	got, err = sccManageServicesParent()
	if err != nil {
		t.Fatalf("--parent org: %v", err)
	}
	if got != "organizations/9/locations/global" {
		t.Errorf("--parent org parent = %q", got)
	}

	// --parent with trailing slash trimmed.
	flagSccManageParent = "folders/12/"
	flagSccOrg, flagSccFolder, flagSccProject = "", "", ""
	got, err = sccManageServicesParent()
	if err != nil {
		t.Fatalf("--parent folder w/ trailing slash: %v", err)
	}
	if got != "folders/12/locations/global" {
		t.Errorf("--parent folder w/ trailing slash = %q", got)
	}

	// --parent bad shape.
	flagSccManageParent = "widgets/1"
	if _, err := sccManageServicesParent(); err == nil {
		t.Error("--parent=widgets/1: expected error")
	}

	// --parent + --organization: mutex.
	flagSccManageParent = "organizations/1"
	flagSccOrg = "2"
	if _, err := sccManageServicesParent(); err == nil ||
		!strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutex error, got %v", err)
	}
}

func TestSccManageParseFilterModules(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"WEB_UI_ENABLED", []string{"WEB_UI_ENABLED"}},
		{"WEB_UI_ENABLED,API_KEY_NOT_ROTATED",
			[]string{"WEB_UI_ENABLED", "API_KEY_NOT_ROTATED"}},
		{"[WEB_UI_ENABLED, API_KEY_NOT_ROTATED]",
			[]string{"WEB_UI_ENABLED", "API_KEY_NOT_ROTATED"}},
	}
	for _, tc := range cases {
		got := sccManageParseFilterModules(tc.in)
		if len(tc.want) == 0 {
			if got != nil {
				t.Errorf("sccManageParseFilterModules(%q) = %v, want nil", tc.in, got)
			}
			continue
		}
		if len(got) != len(tc.want) {
			t.Errorf("sccManageParseFilterModules(%q) = %v, want %v", tc.in, got, tc.want)
		}
		for _, name := range tc.want {
			if !got[name] {
				t.Errorf("sccManageParseFilterModules(%q) missing %q", tc.in, name)
			}
		}
	}
}

func TestSccManageLoadModuleConfig(t *testing.T) {
	dir := t.TempDir()

	// YAML input using the snake_case form the Python examples ship with.
	yamlPath := filepath.Join(dir, "modules.yaml")
	if err := os.WriteFile(yamlPath, []byte(
		"DISK_CMEK_DISABLED:\n  intended_enablement_state: disabled\n"+
			"SQL_WEAK_ROOT_PASSWORD:\n  intended_enablement_state: ENABLED\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := sccManageLoadModuleConfig(yamlPath)
	if err != nil {
		t.Fatalf("yaml: %v", err)
	}
	// Keys should have been rewritten to intendedEnablementState and values
	// upper-cased.
	disk, _ := got["DISK_CMEK_DISABLED"].(map[string]any)
	if disk == nil || disk["intendedEnablementState"] != "DISABLED" {
		t.Errorf("YAML disk config = %v", disk)
	}
	sql, _ := got["SQL_WEAK_ROOT_PASSWORD"].(map[string]any)
	if sql == nil || sql["intendedEnablementState"] != "ENABLED" {
		t.Errorf("YAML sql config = %v", sql)
	}

	// JSON with the camelCase key -- should pass through unchanged.
	jsonPath := filepath.Join(dir, "modules.json")
	if err := os.WriteFile(jsonPath, []byte(
		`{"MOD_A": {"intendedEnablementState": "inherited"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = sccManageLoadModuleConfig(jsonPath)
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	m, _ := got["MOD_A"].(map[string]any)
	if m == nil || m["intendedEnablementState"] != "INHERITED" {
		t.Errorf("JSON mod A config = %v", m)
	}

	// Empty path: nil (nothing to send).
	if got, err := sccManageLoadModuleConfig(""); err != nil || got != nil {
		t.Errorf("empty path: got %v, %v", got, err)
	}

	// Invalid enablement value inside the file surfaces as an error naming
	// the module so the caller can find it.
	badPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(badPath, []byte(
		"MOD_A:\n  intended_enablement_state: MUTED\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := sccManageLoadModuleConfig(badPath); err == nil ||
		!strings.Contains(err.Error(), "MOD_A") {
		t.Errorf("expected MOD_A error, got %v", err)
	}
}

// sccManageServicesTestSetup wires the shared REST client at an httptest
// server, injects a synthetic OAuth token, and resets the manage-services
// flag globals when the test ends.
func sccManageServicesTestSetup(t *testing.T, server *httptest.Server) {
	t.Helper()

	origRest := sccManageRest
	sccManageRestClientForTest(server.URL)
	t.Cleanup(func() { sccManageRest = origRest })

	restoreTok := SetRestTokenSourceForTest(
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"}))
	t.Cleanup(restoreTok)

	restoreQP := SetRestUserProjectForTest("qp-manage")
	t.Cleanup(restoreQP)

	saveOrg, saveFolder, saveProj, saveParent :=
		flagSccOrg, flagSccFolder, flagSccProject, flagSccManageParent
	saveState, saveModuleFile, saveValidate :=
		flagSccManageEnablementState, flagSccManageModuleConfigFile, flagSccManageValidateOnly
	saveFilter, saveFormat, savePage :=
		flagSccManageFilterModules, flagSccFormat, flagSccPageSize
	t.Cleanup(func() {
		flagSccOrg, flagSccFolder, flagSccProject, flagSccManageParent =
			saveOrg, saveFolder, saveProj, saveParent
		flagSccManageEnablementState, flagSccManageModuleConfigFile, flagSccManageValidateOnly =
			saveState, saveModuleFile, saveValidate
		flagSccManageFilterModules, flagSccFormat, flagSccPageSize =
			saveFilter, saveFormat, savePage
	})
	// Start from a clean slate so a leaking global from an earlier test
	// doesn't affect the parent resolution here.
	flagSccOrg, flagSccFolder, flagSccProject, flagSccManageParent = "", "", "", ""
	flagSccManageEnablementState, flagSccManageModuleConfigFile, flagSccManageValidateOnly = "", "", false
	flagSccManageFilterModules, flagSccFormat, flagSccPageSize = "", "", 0
}

func TestSccManageServicesDescribe(t *testing.T) {
	var gotPath, gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":                     "organizations/1/locations/global/securityCenterServices/security-health-analytics",
			"intendedEnablementState":  "ENABLED",
			"effectiveEnablementState": "ENABLED",
			"modules": map[string]any{
				"DISK_CMEK_DISABLED":     map[string]any{"intendedEnablementState": "ENABLED"},
				"SQL_WEAK_ROOT_PASSWORD": map[string]any{"intendedEnablementState": "DISABLED"},
			},
		})
	}))
	defer server.Close()
	sccManageServicesTestSetup(t, server)

	// Abbreviation input should be normalised to the full name in the URL.
	flagSccOrg = "1"
	flagSccFormat = "json"
	out := captureStdout(t, func() {
		if err := runSccManageServicesDescribe(sccManageServicesDescribeCmd, []string{"sha"}); err != nil {
			t.Fatalf("runSccManageServicesDescribe: %v", err)
		}
	})

	if gotMethod != http.MethodGet {
		t.Errorf("method = %s, want GET", gotMethod)
	}
	want := "/organizations/1/locations/global/securityCenterServices/security-health-analytics"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	for _, need := range []string{"intendedEnablementState", "DISK_CMEK_DISABLED"} {
		if !strings.Contains(out, need) {
			t.Errorf("output missing %q; got:\n%s", need, out)
		}
	}
}

// TestSccManageServicesDescribeFilterModules exercises --filter-modules,
// which reduces the response to a modules map holding only the named keys.
func TestSccManageServicesDescribeFilterModules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"modules": map[string]any{
				"MOD_KEEP":  map[string]any{"intendedEnablementState": "ENABLED"},
				"MOD_DROP":  map[string]any{"intendedEnablementState": "DISABLED"},
				"MOD_OTHER": map[string]any{"intendedEnablementState": "INHERITED"},
			},
		})
	}))
	defer server.Close()
	sccManageServicesTestSetup(t, server)

	flagSccOrg = "1"
	flagSccFormat = "json"
	flagSccManageFilterModules = "MOD_KEEP,MOD_OTHER"

	out := captureStdout(t, func() {
		if err := runSccManageServicesDescribe(sccManageServicesDescribeCmd,
			[]string{"security-health-analytics"}); err != nil {
			t.Fatalf("runSccManageServicesDescribe: %v", err)
		}
	})
	if !strings.Contains(out, "MOD_KEEP") || !strings.Contains(out, "MOD_OTHER") {
		t.Errorf("expected MOD_KEEP and MOD_OTHER in output; got:\n%s", out)
	}
	if strings.Contains(out, "MOD_DROP") {
		t.Errorf("did not expect MOD_DROP in output; got:\n%s", out)
	}
}

// TestSccManageServicesDescribeRejectsUnknownService checks that describe
// bails out before making an HTTP request when the service isn't in the
// allowlist. The reference test would otherwise wait for the server's 404.
func TestSccManageServicesDescribeRejectsUnknownService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("HTTP server should not be reached for an invalid service")
	}))
	defer server.Close()
	sccManageServicesTestSetup(t, server)
	flagSccOrg = "1"

	err := runSccManageServicesDescribe(sccManageServicesDescribeCmd, []string{"not-a-service"})
	if err == nil || !strings.Contains(err.Error(), "invalid service name") {
		t.Errorf("expected invalid-service-name error, got %v", err)
	}
}

func TestSccManageServicesList(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"securityCenterServices": []any{
				map[string]any{
					"name":                     "folders/42/locations/global/securityCenterServices/security-health-analytics",
					"intendedEnablementState":  "ENABLED",
					"effectiveEnablementState": "ENABLED",
				},
				map[string]any{
					"name":                     "folders/42/locations/global/securityCenterServices/event-threat-detection",
					"intendedEnablementState":  "INHERITED",
					"effectiveEnablementState": "DISABLED",
				},
			},
		})
	}))
	defer server.Close()
	sccManageServicesTestSetup(t, server)

	flagSccFolder = "42"
	out := captureStdout(t, func() {
		if err := runSccManageServicesList(sccManageServicesListCmd, nil); err != nil {
			t.Fatalf("runSccManageServicesList: %v", err)
		}
	})
	want := "/folders/42/locations/global/securityCenterServices"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	// Default table: header + two rows -> at least the service basenames
	// and their intended states must show up.
	for _, s := range []string{"NAME", "INTENDED_STATE", "EFFECTIVE_STATE",
		"security-health-analytics", "ENABLED",
		"event-threat-detection", "INHERITED"} {
		if !strings.Contains(out, s) {
			t.Errorf("output missing %q; got:\n%s", s, out)
		}
	}
}

func TestSccManageServicesUpdate(t *testing.T) {
	var gotPath, gotMethod, gotQuery string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod, gotQuery = r.URL.Path, r.Method, r.URL.RawQuery
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":                    gotBody["name"],
			"intendedEnablementState": "ENABLED",
		})
	}))
	defer server.Close()
	sccManageServicesTestSetup(t, server)

	// Take the prompt out of the way: in a non-interactive test context
	// promptYesNo(_, true) already defaults to yes, but flagQuiet keeps
	// it explicit and independent of tty detection quirks.
	saveQuiet := flagQuiet
	flagQuiet = true
	defer func() { flagQuiet = saveQuiet }()

	// Module config on disk so the load path is exercised as well.
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "modules.yaml")
	if err := os.WriteFile(modulePath, []byte(
		"DISK_CMEK_DISABLED:\n  intended_enablement_state: enabled\n"), 0600); err != nil {
		t.Fatal(err)
	}

	flagSccOrg = "1"
	flagSccManageEnablementState = "enabled"
	flagSccManageModuleConfigFile = modulePath
	flagSccFormat = "json"

	if err := runSccManageServicesUpdate(sccManageServicesUpdateCmd, []string{"sha"}); err != nil {
		t.Fatalf("runSccManageServicesUpdate: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", gotMethod)
	}
	wantPath := "/organizations/1/locations/global/securityCenterServices/security-health-analytics"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if !strings.Contains(gotQuery, "updateMask=intended_enablement_state%2Cmodules") {
		t.Errorf("updateMask missing (both flags set); got query %q", gotQuery)
	}
	if strings.Contains(gotQuery, "validateOnly") {
		t.Errorf("validateOnly should be absent when the flag is unset; got query %q", gotQuery)
	}
	if gotBody["intendedEnablementState"] != "ENABLED" {
		t.Errorf("intendedEnablementState = %v, want ENABLED", gotBody["intendedEnablementState"])
	}
	modules, _ := gotBody["modules"].(map[string]any)
	if modules == nil {
		t.Fatalf("modules missing from body: %v", gotBody)
	}
	disk, _ := modules["DISK_CMEK_DISABLED"].(map[string]any)
	if disk == nil || disk["intendedEnablementState"] != "ENABLED" {
		t.Errorf("modules[DISK_CMEK_DISABLED] = %v, want intendedEnablementState=ENABLED", disk)
	}
}

// TestSccManageServicesUpdateValidateOnly checks the --validate-only path:
// the request must carry validateOnly=true and skip the confirmation
// prompt entirely, so it works fine without --quiet.
func TestSccManageServicesUpdateValidateOnly(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "ok"})
	}))
	defer server.Close()
	sccManageServicesTestSetup(t, server)

	flagSccOrg = "1"
	flagSccManageEnablementState = "disabled"
	flagSccManageValidateOnly = true
	flagSccFormat = "json"

	if err := runSccManageServicesUpdate(sccManageServicesUpdateCmd, []string{"etd"}); err != nil {
		t.Fatalf("runSccManageServicesUpdate: %v", err)
	}
	if !strings.Contains(gotQuery, "validateOnly=true") {
		t.Errorf("expected validateOnly=true in query, got %q", gotQuery)
	}
	if !strings.Contains(gotQuery, "updateMask=intended_enablement_state") {
		t.Errorf("expected updateMask=intended_enablement_state, got %q", gotQuery)
	}
}

// TestSccManageServicesUpdateRequiresAField mirrors the Python parser: at
// least one of --enablement-state or --module-config-file must be supplied,
// otherwise the caller has asked for a Patch with an empty mask.
func TestSccManageServicesUpdateRequiresAField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("HTTP server should not be reached without --enablement-state or --module-config-file")
	}))
	defer server.Close()
	sccManageServicesTestSetup(t, server)

	flagSccOrg = "1"

	err := runSccManageServicesUpdate(sccManageServicesUpdateCmd, []string{"sha"})
	if err == nil || !strings.Contains(err.Error(), "--enablement-state") {
		t.Errorf("expected update-mask error naming --enablement-state, got %v", err)
	}
}
