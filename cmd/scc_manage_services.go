package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// --- gcloud scc manage services (#1820) ---
//
// This subgroup wraps the Security Center Management API
// (securitycentermanagement.googleapis.com/v1), which is a separate API
// from the SCC data plane used by `scc findings` etc.: there is no V2, so
// the V1/V2 routing rule that governs other scc subcommands does not
// apply here.
//
// The Python reference (v582.0.0) lives at:
//   * lib/surface/scc/manage/services/{describe,list,update}.py
//   * lib/googlecloudsdk/command_lib/scc/manage/{constants,flags,parsing}.py
//   * lib/googlecloudsdk/api_lib/scc/manage/services/clients.py
//
// Location is hardcoded to "global" upstream and here for parity.

var sccManageRest = newRESTClient("https://securitycentermanagement.googleapis.com/v1")

// sccManageRestClientForTest lets tests point the manage-services client at
// an httptest.Server. Callers restore the previous value themselves.
func sccManageRestClientForTest(endpoint string) {
	sccManageRest = newRESTClient(endpoint)
}

// sccService pairs the canonical service short-name with its optional
// abbreviation, mirroring gcloud-python's SecurityCenterService dataclass.
// Both names are accepted from user input and resolved to the canonical
// short name before the API call.
type sccService struct {
	Name   string
	Abbrev string
}

// sccSupportedServices is the client-side allowlist of Security Command
// Center services that `scc manage services` accepts. It mirrors
// gcloud-python's SUPPORTED_SERVICES tuple in
// lib/googlecloudsdk/command_lib/scc/manage/constants.py at v582.0.0.
//
// Keep this in sync with upstream; new services land here as a one-liner
// per release.
var sccSupportedServices = []sccService{
	{Name: "security-health-analytics", Abbrev: "sha"},
	{Name: "event-threat-detection", Abbrev: "etd"},
	{Name: "container-threat-detection", Abbrev: "ctd"},
	{Name: "vm-threat-detection", Abbrev: "vmtd"},
	{Name: "web-security-scanner", Abbrev: "wss"},
	{Name: "vm-threat-detection-aws", Abbrev: "vmtd-aws"},
	{Name: "cloud-run-threat-detection", Abbrev: "crtd"},
	// external-exposure (ee) was added upstream at gcloud-python 572.0.0.
	{Name: "external-exposure", Abbrev: "ee"},
	{Name: "vm-manager", Abbrev: "vmm"},
	{Name: "ec2-vulnerability-assessment", Abbrev: "ec2-va"},
	{Name: "gce-vulnerability-assessment", Abbrev: "gce-va"},
	{Name: "azure-vulnerability-assessment", Abbrev: "azure-va"},
	{Name: "notebook-security-scanner", Abbrev: "nss"},
	{Name: "agent-engine-threat-detection", Abbrev: "aetd"},
}

// sccResolveServiceName maps a full name or abbreviation (case-insensitive)
// to the canonical short name the API accepts, or returns "" when the input
// matches nothing. This mirrors the SERVICE_INVENTORY lookup in
// gcloud-python parsing.GetServiceNameFromArgs.
func sccResolveServiceName(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	if lower == "" {
		return ""
	}
	for _, s := range sccSupportedServices {
		if s.Name == lower {
			return s.Name
		}
		if s.Abbrev != "" && s.Abbrev == lower {
			return s.Name
		}
	}
	return ""
}

// sccSupportedServicesForHelp returns the supported services rendered for
// --help text, alphabetised for stable output, with abbreviations shown in
// parentheses when present.
func sccSupportedServicesForHelp() string {
	names := make([]string, 0, len(sccSupportedServices))
	for _, s := range sccSupportedServices {
		if s.Abbrev != "" {
			names = append(names, fmt.Sprintf("%s (%s)", s.Name, s.Abbrev))
		} else {
			names = append(names, s.Name)
		}
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// --- flags / scope ---

var (
	// flagSccManageParent is the shared --parent flag for scc manage services
	// commands. It is registered separately from --organization/--folder/
	// --project (which come from sccAddScopeFlags) so `--parent=folders/1`
	// works as an alternative one-flag form.
	flagSccManageParent string

	// flagSccManageFilterModules restricts describe output to a comma-separated
	// list of module names (SCREAMING_SNAKE_CASE). Empty means no filter.
	flagSccManageFilterModules string

	// flagSccManageEnablementState is --enablement-state on update, one of
	// ENABLED / DISABLED / INHERITED (case-insensitive). Empty means unset.
	flagSccManageEnablementState string

	// flagSccManageModuleConfigFile is --module-config-file on update, a
	// path to a JSON or YAML file describing per-module intended state.
	flagSccManageModuleConfigFile string

	// flagSccManageValidateOnly is --validate-only on update: perform IAM
	// checks and validation without mutating the service.
	flagSccManageValidateOnly bool
)

// sccManageServicesParent resolves the parent for the manage services
// commands and appends /locations/global, matching gcloud-python which
// hardcodes the location. It accepts one of --parent, --organization,
// --folder, or --project (mutually exclusive; falls back to the active
// project when none are supplied).
func sccManageServicesParent() (string, error) {
	// --parent takes precedence and must be a fully-qualified resource
	// name. Mixing --parent with any of --organization/--folder/--project
	// is rejected so the caller gets a clear error instead of one flag
	// silently winning.
	if flagSccManageParent != "" {
		if flagSccOrg != "" || flagSccFolder != "" || flagSccProject != "" {
			return "", fmt.Errorf(
				"--parent is mutually exclusive with --organization, --folder, and --project")
		}
		trimmed := strings.TrimSuffix(flagSccManageParent, "/")
		kind := strings.SplitN(trimmed, "/", 2)[0]
		switch kind {
		case "organizations", "folders", "projects":
			// Nothing else to do -- the parent already names the resource
			// type and its id.
		default:
			return "", fmt.Errorf(
				"--parent must be one of organizations|folders|projects/{id}, got %q",
				flagSccManageParent)
		}
		return trimmed + "/locations/global", nil
	}
	parent, err := sccResolveParent()
	if err != nil {
		return "", err
	}
	return parent + "/locations/global", nil
}

// sccManageServiceResourceName returns the fully-qualified resource name
// for a Security Center Service under the resolved parent, using the
// user-supplied short-name or abbreviation.
func sccManageServiceResourceName(userInput string) (string, error) {
	canonical := sccResolveServiceName(userInput)
	if canonical == "" {
		return "", fmt.Errorf(
			"invalid service name %q. Supported: %s",
			userInput, sccSupportedServicesForHelp())
	}
	parent, err := sccManageServicesParent()
	if err != nil {
		return "", err
	}
	return parent + "/securityCenterServices/" + canonical, nil
}

// sccManageParseEnablementState normalises the --enablement-state input,
// returning "" when unset. Matches the Python parser: ENABLED, DISABLED,
// or INHERITED (case-insensitive).
func sccManageParseEnablementState(state string) (string, error) {
	if state == "" {
		return "", nil
	}
	upper := strings.ToUpper(strings.TrimSpace(state))
	switch upper {
	case "ENABLED", "DISABLED", "INHERITED":
		return upper, nil
	default:
		return "", fmt.Errorf(
			`--enablement-state must be one of ENABLED, DISABLED, or INHERITED, got %q`,
			state)
	}
}

// sccManageLoadModuleConfig reads the --module-config-file at path (JSON or
// YAML) and returns the modules map ready to place into the API request.
//
// The upstream YAML uses snake_case keys (`intended_enablement_state`) but
// the API expects camelCase. We accept either form to spare users a mental
// map and rewrite the known field on the way in. The enablement value is
// also upper-cased so `enabled` etc. work.
func sccManageLoadModuleConfig(path string) (map[string]any, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading --module-config-file: %w", err)
	}
	var generic any
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return nil, fmt.Errorf("parsing --module-config-file: %w", err)
	}
	normalised, ok := convertYAMLKeys(generic).(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"--module-config-file must be a map of module name to settings, got %T",
			generic)
	}
	out := make(map[string]any, len(normalised))
	for moduleName, raw := range normalised {
		settings, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"module %q in --module-config-file must be an object, got %T",
				moduleName, raw)
		}
		module := make(map[string]any, len(settings))
		for key, val := range settings {
			apiKey := key
			// Accept both snake_case (mirrors the Python YAML example) and
			// camelCase (the raw API field name) so users don't have to
			// remember which form the file expects.
			if apiKey == "intended_enablement_state" {
				apiKey = "intendedEnablementState"
			}
			if apiKey == "intendedEnablementState" {
				strVal, ok := val.(string)
				if !ok {
					return nil, fmt.Errorf(
						"module %q: intendedEnablementState must be a string, got %T",
						moduleName, val)
				}
				enum, err := sccManageParseEnablementState(strVal)
				if err != nil {
					return nil, fmt.Errorf("module %q: %w", moduleName, err)
				}
				val = enum
			}
			module[apiKey] = val
		}
		out[moduleName] = module
	}
	return out, nil
}

// sccManageParseFilterModules splits --filter-modules into a set of module
// names, tolerating the `[MOD_A, MOD_B]` shape gcloud-python accepts
// alongside plain comma-separated lists.
func sccManageParseFilterModules(input string) map[string]bool {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}
	trimmed = strings.Trim(trimmed, "[]")
	parts := strings.Split(trimmed, ",")
	set := make(map[string]bool, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			set[p] = true
		}
	}
	return set
}

// sccManageBuildUpdateMask synthesises the FieldMask for a service Patch
// call from which flags the caller set. Mirrors
// CreateUpdateMaskFromArgsForService in gcloud-python.
func sccManageBuildUpdateMask(hasState, hasModuleConfig bool) (string, error) {
	switch {
	case hasState && hasModuleConfig:
		return "intended_enablement_state,modules", nil
	case hasState:
		return "intended_enablement_state", nil
	case hasModuleConfig:
		return "modules", nil
	default:
		return "", fmt.Errorf(
			"either --enablement-state or --module-config-file (or both) is required")
	}
}

// --- commands ---

var sccManageServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage Security Command Center services",
}

var (
	sccManageServicesDescribeCmd = &cobra.Command{
		Use:   "describe SERVICE_NAME",
		Short: "Describe a Security Command Center service",
		Args:  cobra.ExactArgs(1),
		RunE:  runSccManageServicesDescribe,
	}
	sccManageServicesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Security Command Center services",
		Args:  cobra.NoArgs,
		RunE:  runSccManageServicesList,
	}
	sccManageServicesUpdateCmd = &cobra.Command{
		Use:   "update SERVICE_NAME",
		Short: "Update a Security Command Center service",
		Args:  cobra.ExactArgs(1),
		RunE:  runSccManageServicesUpdate,
	}
)

// sccAddManageServicesScopeFlags registers --parent and the shared
// --organization/--folder/--project scope flags on the given commands, so
// callers can supply exactly one.
func sccAddManageServicesScopeFlags(cmds ...*cobra.Command) {
	sccAddScopeFlags(cmds...)
	for _, c := range cmds {
		c.Flags().StringVar(&flagSccManageParent, "parent", "",
			"Fully-qualified parent (organizations/{id}, folders/{id}, or projects/{id}). "+
				"Mutually exclusive with --organization/--folder/--project.")
	}
}

// serviceNameHelp is the shared help suffix for the SERVICE_NAME positional
// so users see the supported list without having to leave the CLI.
func serviceNameHelp() string {
	return "The service name in lowercase-hyphenated form (e.g. security-health-analytics), " +
		"or its abbreviation (e.g. sha). Supported: " + sccSupportedServicesForHelp() + "."
}

// The init() wires the new subgroup into the existing `scc manage` command.
// This runs after all package-level vars are initialised, so sccManageCmd
// (defined in scc.go) is guaranteed to exist. The other file's init() also
// mutates sccManageCmd (to add `settings`); both mutations are order-
// independent because cobra just appends to a slice.
func init() {
	sccAddManageServicesScopeFlags(
		sccManageServicesDescribeCmd,
		sccManageServicesListCmd,
		sccManageServicesUpdateCmd,
	)
	sccAddFormatFlag(
		sccManageServicesDescribeCmd,
		sccManageServicesListCmd,
		sccManageServicesUpdateCmd,
	)
	sccAddPageSizeFlag(sccManageServicesListCmd)

	sccManageServicesDescribeCmd.Flags().StringVar(&flagSccManageFilterModules,
		"filter-modules", "",
		"If provided, only print module information for the listed modules. "+
			"Comma-separated (or [MOD_A,MOD_B]) list of SCREAMING_SNAKE_CASE "+
			"module names (e.g. WEB_UI_ENABLED,API_KEY_NOT_ROTATED).")

	sccManageServicesUpdateCmd.Flags().StringVar(&flagSccManageEnablementState,
		"enablement-state", "",
		"Sets the enablement state of the Security Center service. "+
			"One of ENABLED, DISABLED, or INHERITED. INHERITED is only valid "+
			"at project or folder scope. At least one of --enablement-state or "+
			"--module-config-file is required.")
	sccManageServicesUpdateCmd.Flags().StringVar(&flagSccManageModuleConfigFile,
		"module-config-file", "",
		"Path to a JSON or YAML file mapping module name to its intended "+
			"enablement state. Keys are module names (e.g. DISK_CMEK_DISABLED), "+
			"values are objects with an `intended_enablement_state` (or "+
			"`intendedEnablementState`) field of ENABLED, DISABLED, or INHERITED.")
	sccManageServicesUpdateCmd.Flags().BoolVar(&flagSccManageValidateOnly,
		"validate-only", false,
		"If set, the request is validated (including IAM checks) but no "+
			"action is taken.")

	sccManageServicesDescribeCmd.Long = "Describe a Security Command Center service. " +
		"Resolves INHERITED enablement states to ENABLED or DISABLED for " +
		"services at ancestor levels.\n\n" +
		"Positional: SERVICE_NAME. " + serviceNameHelp()
	sccManageServicesUpdateCmd.Long = "Update the enablement state of a Security Command " +
		"Center service and its modules for the specified organization, " +
		"folder, or project.\n\n" +
		"Positional: SERVICE_NAME. " + serviceNameHelp()

	sccManageServicesCmd.AddCommand(
		sccManageServicesDescribeCmd,
		sccManageServicesListCmd,
		sccManageServicesUpdateCmd,
	)
	sccManageCmd.AddCommand(sccManageServicesCmd)
}

// --- run functions ---

func runSccManageServicesDescribe(cmd *cobra.Command, args []string) error {
	name, err := sccManageServiceResourceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var raw map[string]any
	if err := sccManageRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &raw); err != nil {
		return fmt.Errorf("describing security center service: %w", err)
	}
	filter := sccManageParseFilterModules(flagSccManageFilterModules)
	if len(filter) == 0 {
		return emitFormatted(raw, flagSccFormat)
	}
	// Filter down to the requested modules, preserving the map shape the
	// full response uses so downstream tools that key off "modules" keep
	// working.
	modules, _ := raw["modules"].(map[string]any)
	filtered := make(map[string]any, len(filter))
	for key, value := range modules {
		if filter[key] {
			filtered[key] = value
		}
	}
	return emitFormatted(map[string]any{"modules": filtered}, flagSccFormat)
}

func runSccManageServicesList(cmd *cobra.Command, args []string) error {
	parent, err := sccManageServicesParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	services, err := sccManageRest.paginate(
		ctx, "/"+parent+"/securityCenterServices",
		url.Values{}, "securityCenterServices", flagSccPageSize)
	if err != nil {
		return fmt.Errorf("listing security center services: %w", err)
	}
	if flagSccFormat != "" {
		return emitFormatted(services, flagSccFormat)
	}
	// Default table matches the shape the other scc list commands use:
	// short name + intended state + effective state.
	fmt.Printf("%-40s %-24s %s\n", "NAME", "INTENDED_STATE", "EFFECTIVE_STATE")
	for _, s := range services {
		name, _ := s["name"].(string)
		intended, _ := s["intendedEnablementState"].(string)
		effective, _ := s["effectiveEnablementState"].(string)
		fmt.Printf("%-40s %-24s %s\n", path.Base(name), intended, effective)
	}
	return nil
}

func runSccManageServicesUpdate(cmd *cobra.Command, args []string) error {
	name, err := sccManageServiceResourceName(args[0])
	if err != nil {
		return err
	}
	state, err := sccManageParseEnablementState(flagSccManageEnablementState)
	if err != nil {
		return err
	}
	modules, err := sccManageLoadModuleConfig(flagSccManageModuleConfigFile)
	if err != nil {
		return err
	}
	mask, err := sccManageBuildUpdateMask(state != "", modules != nil)
	if err != nil {
		return err
	}

	// Match gcloud-python: prompt before mutating unless --validate-only
	// makes the call a dry-run. Non-interactive/--quiet callers accept the
	// default (proceed), matching PromptContinue's behavior upstream.
	if !flagSccManageValidateOnly {
		msg := fmt.Sprintf(
			"Are you sure you want to update the Security Center Service %s?",
			args[0])
		if !promptYesNo(msg, true) {
			return fmt.Errorf("aborted by user")
		}
	}

	body := map[string]any{"name": name}
	if state != "" {
		body["intendedEnablementState"] = state
	}
	if modules != nil {
		body["modules"] = modules
	}

	query := url.Values{}
	query.Set("updateMask", mask)
	if flagSccManageValidateOnly {
		query.Set("validateOnly", "true")
	}

	ctx := context.Background()
	var resp map[string]any
	if err := sccManageRest.do(ctx, http.MethodPatch, "/"+name, query, body, &resp); err != nil {
		return fmt.Errorf("updating security center service: %w", err)
	}
	if flagSccManageValidateOnly {
		fmt.Fprintln(os.Stderr, "Request is valid.")
	} else {
		fmt.Fprintf(os.Stderr, "Updated security center service [%s].\n", name)
	}
	return emitFormatted(resp, flagSccFormat)
}
