package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// --- gcloud assured workloads --version=v2 (#1910, gcloud-python 583–589) ---
//
// gcloud-python routes `assured workloads create/describe/list/delete/update`
// to the Assured Workloads v2 API when --version=v2 is given (default v1).
// v2 is not in google.golang.org/api, so it is called over REST on the
// regional endpoint. v2 defines no operations service: long-running
// operations are polled through the v1 operations endpoint, and create/update
// then fetch the resulting workload with a v2 GET, as Python's
// CloudOperationPoller does.

var (
	flagAssuredVersion           string
	flagAssuredAsync             bool
	flagAssuredWorkloadID        string
	flagAssuredExternalID        string
	flagAssuredDescription       string
	flagAssuredTargetProject     string
	flagAssuredTargetFolder      string
	flagAssuredNewProjectConfig  map[string]string
	flagAssuredNewFolderConfig   map[string]string
	flagAssuredFramework         string
	flagAssuredFrameworkRevision int64
	flagAssuredCloudControlCfg   []string
	flagAssuredCloudControlFile  string
	flagAssuredCmekKeyRing       string
	flagAssuredCmekDedicated     map[string]string
)

// assuredVersionFlag pairs a flag with the remediation shown when it is used
// with the other API version (gcloud-python's V1_ONLY_*/V2_ONLY_* tables).
type assuredVersionFlag struct{ name, remediation string }

var (
	assuredV1OnlyCreateFlags = []assuredVersionFlag{
		{"compliance-regime", "Use '--framework' when creating V2 workloads."},
		{"billing-account", "Configure billing within '--new-project-config' or '--cloud-control-config' when creating V2 workloads."},
		{"display-name", "Configure folder/project display names within '--new-folder-config' or '--new-project-config'."},
	}
	assuredV2OnlyCreateFlags = []assuredVersionFlag{
		{"framework", "Use '--compliance-regime' when creating V1 workloads."},
		{"framework-major-revision", "Major revision is only applicable to V2 frameworks."},
		{"target-project", "Target resources are only supported in V2 workloads."},
		{"target-folder", "Target resources are only supported in V2 workloads."},
		{"new-project-config", "New project specifications are only supported in V2 workloads."},
		{"new-folder-config", "New folder specifications are only supported in V2 workloads."},
		{"cloud-control-config", "Use '--resource-settings' when creating V1 workloads."},
		{"cloud-control-configs-from-file", "Use '--resource-settings' when creating V1 workloads."},
		{"cmek-key-ring", "CMEK key rings are only supported in V2 workloads."},
		{"cmek-dedicated-project-config", "Dedicated CMEK projects are only supported in V2 workloads."},
		{"description", "Use '--display-name' when creating V1 workloads."},
		{"workload-id", "Explicit workload-id is only supported in V2 workloads."},
		{"async", "Asynchronous operations are only supported on V2 workloads."},
	}
	assuredV1OnlyUpdateFlags = []assuredVersionFlag{
		{"display-name", "Display name is only supported on V1 workloads."},
	}
	assuredV2OnlyUpdateFlags = []assuredVersionFlag{
		{"framework-major-revision", "Framework major revision is only supported on V2 workloads."},
		{"cloud-control-config", "Cloud control configs are only supported on V2 workloads."},
		{"cloud-control-configs-from-file", "Cloud control configs are only supported on V2 workloads."},
		{"async", "Asynchronous operations are only supported on V2 workloads."},
	}
)

func init() {
	for _, c := range []*cobra.Command{assuredWorkloadCreateCmd, assuredWorkloadDeleteCmd,
		assuredWorkloadDescribeCmd, assuredWorkloadListCmd, assuredWorkloadUpdateCmd} {
		c.Flags().StringVar(&flagAssuredVersion, "version", "v1",
			"Assured Workloads API version to use. Options are v1 and v2. Defaults to v1.")
	}
	for _, c := range []*cobra.Command{assuredWorkloadCreateCmd, assuredWorkloadDeleteCmd, assuredWorkloadUpdateCmd} {
		c.Flags().BoolVar(&flagAssuredAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete (V2).")
	}
	ccFlags := func(c *cobra.Command) {
		c.Flags().StringArrayVar(&flagAssuredCloudControlCfg, "cloud-control-config", nil,
			"Repeated inline control configuration, as PROPERTY=VALUE pairs (V2)")
		c.Flags().StringVar(&flagAssuredCloudControlFile, "cloud-control-configs-from-file", "",
			"Path to JSON/YAML file specifying multiple control configurations (V2)")
		c.MarkFlagsMutuallyExclusive("cloud-control-config", "cloud-control-configs-from-file")
		c.Flags().Int64Var(&flagAssuredFrameworkRevision, "framework-major-revision", 0,
			"Major revision ID of the compliance framework (V2)")
	}

	c := assuredWorkloadCreateCmd
	c.Use = "create [WORKLOAD_ID]"
	c.Args = cobra.MaximumNArgs(1)
	c.Flags().StringVar(&flagAssuredWorkloadID, "workload-id", "", "User-assigned unique identifier for the workload (V2)")
	c.Flags().StringVar(&flagAssuredExternalID, "external-identifier", "", "The external identifier of the Assured Workloads environment")
	c.MarkFlagsMutuallyExclusive("workload-id", "external-identifier")
	c.Flags().StringVar(&flagAssuredDescription, "description", "", "User-friendly description for the workload (V2)")
	c.Flags().StringVar(&flagAssuredTargetProject, "target-project", "",
		"Existing Project ID or project number to onboard into Assured Workloads (V2)")
	c.Flags().StringVar(&flagAssuredTargetFolder, "target-folder", "", "Existing Folder ID to onboard into Assured Workloads (V2)")
	c.Flags().StringToStringVar(&flagAssuredNewProjectConfig, "new-project-config", nil,
		"Specification for provisioning a new project: parent, display_name, billing_account (V2)")
	c.Flags().StringToStringVar(&flagAssuredNewFolderConfig, "new-folder-config", nil,
		"Specification for provisioning a new folder: parent, display_name (V2)")
	c.MarkFlagsMutuallyExclusive("target-project", "target-folder", "new-project-config", "new-folder-config")
	c.Flags().StringVar(&flagAssuredFramework, "framework", "",
		"Compliance framework resource name or ID, for example, fedramp-moderate (V2)")
	ccFlags(c)
	c.Flags().StringVar(&flagAssuredCmekKeyRing, "cmek-key-ring", "",
		"KMS Key Ring resource path for Customer-Managed Encryption Keys (V2)")
	c.Flags().StringToStringVar(&flagAssuredCmekDedicated, "cmek-dedicated-project-config", nil,
		"Dedicated KMS project configuration: project_id, billing_account, display_name (valid only with folder targets)")

	ccFlags(assuredWorkloadUpdateCmd)
}

// assuredAPIVersion validates and returns --version.
func assuredAPIVersion() (string, error) {
	switch flagAssuredVersion {
	case "", "v1":
		return "v1", nil
	case "v2":
		return "v2", nil
	}
	return "", fmt.Errorf("argument --version: Invalid choice: '%s'. Valid choices are [v1, v2].", flagAssuredVersion)
}

// assuredValidateVersionFlags rejects flags that belong to the other API
// version, as gcloud-python's _ValidateVersionExclusiveFlags does.
func assuredValidateVersionFlags(cmd *cobra.Command, version string, v1Only, v2Only []assuredVersionFlag) error {
	other, otherName := v2Only, "v1"
	if version == "v2" {
		other, otherName = v1Only, "v2"
	}
	for _, f := range other {
		if cmd.Flags().Changed(f.name) {
			return fmt.Errorf("Invalid value for [--%s]: Flag cannot be used with --version=%s. %s", f.name, otherName, f.remediation)
		}
	}
	return nil
}

func assuredValidateCreateFlags(cmd *cobra.Command, args []string, version string) error {
	if err := assuredValidateVersionFlags(cmd, version, assuredV1OnlyCreateFlags, assuredV2OnlyCreateFlags); err != nil {
		return err
	}
	if len(args) > 0 && (cmd.Flags().Changed("workload-id") || cmd.Flags().Changed("external-identifier")) {
		return fmt.Errorf("argument WORKLOAD_ID: At most one of WORKLOAD_ID | --external-identifier | --workload-id can be specified.")
	}
	changed := cmd.Flags().Changed
	if version == "v2" {
		if !changed("framework") {
			return fmt.Errorf("Missing required argument [--framework]: The compliance framework is required for V2 workloads.")
		}
		if !changed("target-project") && !changed("target-folder") && !changed("new-project-config") && !changed("new-folder-config") {
			return fmt.Errorf("Missing required argument [--target-project]: One of --target-project, --target-folder, --new-project-config, or --new-folder-config is required for V2 workloads.")
		}
		return nil
	}
	if len(args) > 0 {
		return fmt.Errorf("Invalid value for [workload_id]: Flag cannot be used with --version=v1. Positional workload ID is only supported in V2 workloads. Use '--external-identifier' for V1.")
	}
	for _, r := range []struct{ flag, what string }{
		{"display-name", "The display name"},
		{"compliance-regime", "The compliance regime"},
		{"billing-account", "The billing account"},
	} {
		if !changed(r.flag) {
			return fmt.Errorf("Missing required argument [--%s]: %s is required for V1 workloads.", r.flag, r.what)
		}
	}
	return nil
}

func assuredValidateUpdateFlags(cmd *cobra.Command, version string) error {
	if err := assuredValidateVersionFlags(cmd, version, assuredV1OnlyUpdateFlags, assuredV2OnlyUpdateFlags); err != nil {
		return err
	}
	changed := cmd.Flags().Changed
	if version == "v2" {
		if !changed("cloud-control-config") && !changed("cloud-control-configs-from-file") {
			return fmt.Errorf("Missing required argument [--cloud-control-config]: One of --cloud-control-config or --cloud-control-configs-from-file must be specified when updating V2 workloads.")
		}
		return nil
	}
	if !changed("display-name") && !changed("etag") {
		return fmt.Errorf("Missing required argument [--display-name]: At least one field must be specified for update")
	}
	return nil
}

func assuredValidateDeleteFlags(cmd *cobra.Command, version string) error {
	if version == "v1" && cmd.Flags().Changed("async") {
		return fmt.Errorf("Invalid value for [--async]: Flag cannot be used with --version=v1. Asynchronous deletion is only supported on V2 workloads.")
	}
	return nil
}

// assuredRegionalEndpoint returns the regional Assured Workloads endpoint
// (https://REGION-assuredworkloads.googleapis.com/), or the global one when
// region is empty, and logs it as gcloud-python does.
func assuredRegionalEndpoint(region string) string {
	ep := "https://assuredworkloads.googleapis.com/"
	if region != "" {
		ep = "https://" + region + "-assuredworkloads.googleapis.com/"
	}
	fmt.Fprintf(os.Stderr, "Using endpoint [%s]\n", ep)
	return ep
}

// assuredRegionOf returns LOCATION from organizations/O/locations/LOCATION/...
func assuredRegionOf(name string) string {
	parts := strings.Split(name, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "locations" {
			return parts[i+1]
		}
	}
	return ""
}

// assuredSplitArgDict parses one "k=v,k=v" ArgDict value.
func assuredSplitArgDict(s string) map[string]any {
	out := map[string]any{}
	for _, kv := range strings.Split(s, ",") {
		if kv == "" {
			continue
		}
		k, v, _ := strings.Cut(kv, "=")
		out[k] = v
	}
	return out
}

// assuredFirst returns the first non-empty string value among keys.
func assuredFirst(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil && v != "" {
			return v
		}
	}
	return nil
}

func assuredStr(m map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// assuredResourceConfig builds the v2 ResourceConfig (BuildResourceConfig).
func assuredResourceConfig() (map[string]any, error) {
	switch {
	case flagAssuredTargetProject != "":
		return map[string]any{"existingProject": flagAssuredTargetProject}, nil
	case flagAssuredTargetFolder != "":
		return map[string]any{"existingFolder": flagAssuredTargetFolder}, nil
	case flagAssuredNewProjectConfig != nil:
		m := flagAssuredNewProjectConfig
		parent := m["parent"]
		dn := assuredStr(m, "display_name", "displayName", "display-name")
		ba := assuredStr(m, "billing_account", "billingAccount", "billing-account")
		if parent == "" || dn == "" || ba == "" {
			return nil, fmt.Errorf("Invalid value for [new_project_config]: new_project_dict must contain parent, display_name, and billing_account.")
		}
		return map[string]any{"newProjectConfig": map[string]any{"parent": parent, "displayName": dn, "billingAccount": ba}}, nil
	case flagAssuredNewFolderConfig != nil:
		m := flagAssuredNewFolderConfig
		parent := m["parent"]
		dn := assuredStr(m, "display_name", "displayName", "display-name")
		if parent == "" || dn == "" {
			return nil, fmt.Errorf("Invalid value for [new_folder_config]: new_folder_dict must contain parent and display_name.")
		}
		return map[string]any{"newFolderConfig": map[string]any{"parent": parent, "displayName": dn}}, nil
	}
	return nil, fmt.Errorf("Invalid value for [resource_config]: Exactly one target resource configuration must be provided (one of target_project, target_folder, new_project_dict, or new_folder_dict).")
}

// assuredCmekConfig builds the v2 CmekConfig (BuildCmekConfig).
func assuredCmekConfig(isFolderTarget bool) (map[string]any, error) {
	if flagAssuredCmekKeyRing == "" {
		return nil, fmt.Errorf("Invalid value for [key_ring_id]: Key ring ID is required for CMEK configuration.")
	}
	cfg := map[string]any{"keyRingId": flagAssuredCmekKeyRing}
	if m := flagAssuredCmekDedicated; len(m) > 0 {
		if !isFolderTarget {
			return nil, fmt.Errorf("Invalid value for [dedicated_project_config]: Dedicated CMEK project configuration is only valid when the target resource is a folder.")
		}
		ba := assuredStr(m, "billing_account", "billingAccount", "billing-account")
		if ba == "" {
			return nil, fmt.Errorf("Invalid value for [billing_account]: Billing account is required for dedicated CMEK project configuration.")
		}
		dp := map[string]any{"billingAccount": ba}
		if v := assuredStr(m, "display_name", "displayName", "display-name"); v != "" {
			dp["displayName"] = v
		}
		if v := assuredStr(m, "project_id", "projectId", "project-id"); v != "" {
			dp["projectId"] = v
		}
		cfg["dedicatedProjectConfig"] = dp
	}
	return cfg, nil
}

// assuredParamValue converts a raw value into a v2 ParamValue
// (BuildParamValue).
func assuredParamValue(v any) (map[string]any, error) {
	switch t := v.(type) {
	case bool:
		return map[string]any{"boolValue": t}, nil
	case int:
		return map[string]any{"numberValue": float64(t)}, nil
	case int64:
		return map[string]any{"numberValue": float64(t)}, nil
	case float64:
		return map[string]any{"numberValue": t}, nil
	case string:
		return map[string]any{"stringValue": t}, nil
	case []any:
		vals := make([]string, len(t))
		for i, e := range t {
			vals[i] = fmt.Sprint(e)
		}
		return map[string]any{"stringListValue": map[string]any{"values": vals}}, nil
	case map[string]any:
		if name, ok := t["name"]; ok {
			sub, has := t["value"]
			if !has {
				sub, has = t["parameter_value"]
			}
			if !has {
				sub, has = t["parameterValue"]
			}
			if has {
				pv, err := assuredParamValue(sub)
				if err != nil {
					return nil, err
				}
				return map[string]any{"oneofValue": map[string]any{"name": name, "parameterValue": pv}}, nil
			}
		}
		for _, k := range [][2]string{{"stringValue", "string_value"}, {"boolValue", "bool_value"}} {
			if x := assuredFirst(t, k[0], k[1]); x != nil {
				return map[string]any{k[0]: x}, nil
			}
		}
		if x := assuredFirst(t, "numberValue", "number_value"); x != nil {
			f, err := strconv.ParseFloat(fmt.Sprint(x), 64)
			if err != nil {
				return nil, fmt.Errorf("Invalid value for [parameter_value]: %v", err)
			}
			return map[string]any{"numberValue": f}, nil
		}
		if x := assuredFirst(t, "stringListValue", "string_list_value"); x != nil {
			if m, ok := x.(map[string]any); ok {
				x = m["values"]
			}
			raw, _ := x.([]any)
			vals := make([]string, len(raw))
			for i, e := range raw {
				vals[i] = fmt.Sprint(e)
			}
			return map[string]any{"stringListValue": map[string]any{"values": vals}}, nil
		}
		return nil, fmt.Errorf("Invalid value for [parameter_value]: Unsupported parameter dict format: %v", t)
	}
	return nil, fmt.Errorf("Invalid value for [parameter_value]: Unsupported parameter value type: %T", v)
}

func assuredParameter(name string, v any) (map[string]any, error) {
	pv, err := assuredParamValue(v)
	if err != nil {
		return nil, err
	}
	return map[string]any{"name": name, "parameterValue": pv}, nil
}

// assuredCloudControlConfigs builds v2 CloudControlConfig objects from
// --cloud-control-config values or --cloud-control-configs-from-file
// (BuildCloudControlConfigs).
func assuredCloudControlConfigs() ([]map[string]any, error) {
	var items []any
	if flagAssuredCloudControlFile != "" {
		data, err := os.ReadFile(flagAssuredCloudControlFile)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", flagAssuredCloudControlFile, err)
		}
		var raw any
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", flagAssuredCloudControlFile, err)
		}
		switch t := raw.(type) {
		case map[string]any:
			items = []any{t}
		case []any:
			items = t
		default:
			return nil, fmt.Errorf("Invalid value for [cloud_control_configs]: Cloud control configs must be a list of configurations or a path to a YAML file.")
		}
	} else {
		for _, s := range flagAssuredCloudControlCfg {
			items = append(items, assuredSplitArgDict(s))
		}
	}

	var out []map[string]any
	for _, it := range items {
		item, ok := it.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("Invalid value for [cloud_control_configs]: Cloud control configs must be a list of configurations or a path to a YAML file.")
		}
		control := assuredFirst(item, "cloud_control", "cloudControl", "cloud-control")
		if control == nil {
			return nil, fmt.Errorf("Invalid value for [cloud_control]: Cloud control name is required.")
		}
		revision := assuredFirst(item, "cloud_control_revision", "cloudControlRevision", "cloud-control-revision", "revision")
		if revision == nil {
			return nil, fmt.Errorf("Invalid value for [cloud_control_revision]: Cloud control revision is required.")
		}
		mode := assuredFirst(item, "enforcement_mode", "enforcementMode", "enforcement-mode")
		if mode == nil {
			return nil, fmt.Errorf("Invalid value for [enforcement_mode]: Enforcement mode is required.")
		}
		modeStr := strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(fmt.Sprint(mode))), "-", "_")
		switch modeStr {
		case "ENFORCEMENT_MODE_UNSPECIFIED", "PREVENTIVE", "DETECTIVE", "AUDIT":
		default:
			return nil, fmt.Errorf("Invalid value for [enforcement_mode]: Invalid enforcement mode: %v. Valid options are PREVENTIVE, DETECTIVE, AUDIT.", mode)
		}

		var params []map[string]any
		switch p := item["parameters"].(type) {
		case map[string]any:
			keys := make([]string, 0, len(p))
			for k := range p {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				param, err := assuredParameter(k, p[k])
				if err != nil {
					return nil, err
				}
				params = append(params, param)
			}
		case []any:
			for _, e := range p {
				pm, ok := e.(map[string]any)
				if !ok {
					continue
				}
				name := pm["name"]
				val, has := pm["value"]
				if !has {
					val = assuredFirst(pm, "parameter_value", "parameterValue")
				}
				if name != nil && val != nil {
					param, err := assuredParameter(fmt.Sprint(name), val)
					if err != nil {
						return nil, err
					}
					params = append(params, param)
				} else if len(pm) == 1 {
					for k, v := range pm {
						param, err := assuredParameter(k, v)
						if err != nil {
							return nil, err
						}
						params = append(params, param)
					}
				}
			}
		}

		rev, err := strconv.ParseInt(strings.TrimSpace(fmt.Sprint(revision)), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Invalid value for [cloud_control_revision]: Cloud control revision must be an integer, got: %v.", revision)
		}
		cfg := map[string]any{
			"cloudControl":         control,
			"cloudControlRevision": rev,
			"enforcementMode":      modeStr,
		}
		if len(params) > 0 {
			cfg["parameters"] = params
		}
		out = append(out, cfg)
	}
	return out, nil
}

// assuredV2Clients returns REST clients for the regional v2 API and the v1
// operations service used to poll its long-running operations.
func assuredV2Clients(region string) (v2, ops *restClient) {
	ep := strings.TrimSuffix(assuredRegionalEndpoint(region), "/")
	return newRESTClient(ep + "/v2"), newRESTClient(ep + "/v1")
}

// assuredV2Wait polls a v2 operation via the v1 operations service, printing
// gcloud's "<progress>...done." line, and returns the finished operation.
func assuredV2Wait(ctx context.Context, ops *restClient, op map[string]any, progress string) (map[string]any, error) {
	name, _ := op["name"].(string)
	fmt.Fprintf(os.Stderr, "%s...", progress)
	if done, _ := op["done"].(bool); !done {
		got, err := ops.waitOperation(ctx, name, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed.")
			return nil, err
		}
		op = got
	}
	if e, ok := op["error"].(map[string]any); ok {
		fmt.Fprintln(os.Stderr, "failed.")
		return nil, fmt.Errorf("operation %s failed: %v", name, e["message"])
	}
	fmt.Fprintln(os.Stderr, "done.")
	return op, nil
}

// assuredV2Result fetches the workload named in a finished operation's
// response, as Python's CloudOperationPoller.GetResult does.
func assuredV2Result(ctx context.Context, v2 *restClient, op map[string]any) (map[string]any, error) {
	resp, _ := op["response"].(map[string]any)
	name, _ := resp["name"].(string)
	if name == "" {
		return resp, nil
	}
	var w map[string]any
	if err := v2.do(ctx, http.MethodGet, "/"+name, nil, nil, &w); err != nil {
		return nil, fmt.Errorf("describing workload: %w", err)
	}
	return w, nil
}

// assuredLogChange prints gcloud's log.CreatedResource / UpdatedResource /
// DeletedResource line for an Assured Workloads environment.
func assuredLogChange(op, name string, async bool) {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress for Assured Workloads environment [%s].\n",
			strings.ToUpper(op[:1])+op[1:], name)
		return
	}
	fmt.Fprintf(os.Stderr, "%sd Assured Workloads environment [%s].\n", strings.ToUpper(op[:1])+op[1:], name)
}

func runAssuredWorkloadCreateV2(cmd *cobra.Command, args []string) error {
	parent := assuredParent()
	body := map[string]any{}
	if flagAssuredCmekKeyRing != "" || len(flagAssuredCmekDedicated) > 0 {
		cmek, err := assuredCmekConfig(flagAssuredTargetFolder != "" || flagAssuredNewFolderConfig != nil)
		if err != nil {
			return err
		}
		body["cmekConfig"] = cmek
	}
	rc, err := assuredResourceConfig()
	if err != nil {
		return err
	}
	body["resourceConfig"] = rc
	fw := map[string]any{"framework": flagAssuredFramework}
	if cmd.Flags().Changed("framework-major-revision") {
		fw["majorRevisionId"] = flagAssuredFrameworkRevision
	}
	body["framework"] = fw
	if flagAssuredCloudControlFile != "" || len(flagAssuredCloudControlCfg) > 0 {
		ccs, err := assuredCloudControlConfigs()
		if err != nil {
			return err
		}
		if len(ccs) > 0 {
			body["cloudControlConfigs"] = ccs
		}
	}
	if flagAssuredDescription != "" {
		body["description"] = flagAssuredDescription
	}
	q := url.Values{}
	workloadID := flagAssuredWorkloadID
	if len(args) > 0 {
		workloadID = args[0]
	}
	if workloadID == "" {
		workloadID = flagAssuredExternalID
	}
	if workloadID != "" {
		q.Set("workloadId", workloadID)
	}

	ctx := context.Background()
	v2, ops := assuredV2Clients(flagAssuredLocation)
	var op map[string]any
	if err := v2.do(ctx, http.MethodPost, "/"+parent+"/workloads", q, body, &op); err != nil {
		return fmt.Errorf("creating workload: %w", err)
	}
	if flagAssuredAsync {
		if err := emitFormatted(op, ""); err != nil {
			return err
		}
		name, _ := op["name"].(string)
		assuredLogChange("create", name, true)
		return nil
	}
	done, err := assuredV2Wait(ctx, ops, op, "Creating Assured Workloads environment")
	if err != nil {
		return err
	}
	w, err := assuredV2Result(ctx, v2, done)
	if err != nil {
		return err
	}
	if err := emitFormatted(w, ""); err != nil {
		return err
	}
	if name, _ := w["name"].(string); name != "" {
		assuredLogChange("create", name, false)
	}
	return nil
}

func runAssuredWorkloadUpdateV2(cmd *cobra.Command, name string) error {
	body := map[string]any{}
	var mask []string
	if cmd.Flags().Changed("framework-major-revision") {
		body["framework"] = map[string]any{"majorRevisionId": flagAssuredFrameworkRevision}
		mask = append(mask, "framework.major_revision_id")
	}
	ccs, err := assuredCloudControlConfigs()
	if err != nil {
		return err
	}
	if ccs == nil {
		ccs = []map[string]any{}
	}
	body["cloudControlConfigs"] = ccs
	mask = append(mask, "cloud_control_configs")
	if cmd.Flags().Changed("etag") {
		body["etag"] = flagAssuredEtag
	}
	q := url.Values{}
	q.Set("updateMask", strings.Join(mask, ","))

	ctx := context.Background()
	v2, ops := assuredV2Clients(assuredRegionOf(name))
	var op map[string]any
	if err := v2.do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating workload: %w", err)
	}
	if flagAssuredAsync {
		if err := emitFormatted(op, ""); err != nil {
			return err
		}
		opName, _ := op["name"].(string)
		assuredLogChange("update", opName, true)
		return nil
	}
	done, err := assuredV2Wait(ctx, ops, op, "Updating Assured Workloads environment")
	if err != nil {
		return err
	}
	w, err := assuredV2Result(ctx, v2, done)
	if err != nil {
		return err
	}
	if err := emitFormatted(w, ""); err != nil {
		return err
	}
	if n, _ := w["name"].(string); n != "" {
		assuredLogChange("update", n, false)
	}
	return nil
}

func runAssuredWorkloadDeleteV2(name string) error {
	q := url.Values{}
	if flagAssuredEtag != "" {
		q.Set("etag", flagAssuredEtag)
	}
	ctx := context.Background()
	v2, ops := assuredV2Clients(assuredRegionOf(name))
	var op map[string]any
	if err := v2.do(ctx, http.MethodDelete, "/"+name, q, nil, &op); err != nil {
		return fmt.Errorf("deleting workload: %w", err)
	}
	if flagAssuredAsync {
		if err := emitFormatted(op, ""); err != nil {
			return err
		}
		assuredLogChange("delete", name, true)
		return nil
	}
	if _, err := assuredV2Wait(ctx, ops, op, "Deleting Assured Workloads environment"); err != nil {
		return err
	}
	assuredLogChange("delete", name, false)
	return nil
}

func runAssuredWorkloadDescribeV2(name string) error {
	v2, _ := assuredV2Clients(assuredRegionOf(name))
	var w map[string]any
	if err := v2.do(context.Background(), http.MethodGet, "/"+name, nil, nil, &w); err != nil {
		return fmt.Errorf("describing workload: %w", err)
	}
	return emitFormatted(w, "")
}

func runAssuredWorkloadListV2() error {
	filter := flagAssuredListFilter
	if filter == "" {
		filter = flagFilter
	}
	q := url.Values{}
	if filter != "" {
		q.Set("filter", filter)
	}
	v2, _ := assuredV2Clients(flagAssuredLocation)
	all, err := v2.paginate(context.Background(), "/"+assuredParent()+"/workloads", q, "workloads", flagAssuredListPageSize)
	if err != nil {
		return fmt.Errorf("listing workloads: %w", err)
	}
	if flagAssuredListLimit > 0 && int64(len(all)) > flagAssuredListLimit {
		all = all[:flagAssuredListLimit]
	}
	return emitAssuredV2List(os.Stdout, os.Stderr, all, flagAssuredListFormat)
}

// emitAssuredV2List renders v2 workloads with gcloud-python's _FORMAT_V2
// table unless --format is given.
func emitAssuredV2List(w, errW io.Writer, all []map[string]any, format string) error {
	if format != "" {
		if all == nil {
			all = []map[string]any{}
		}
		return emitFormattedTo(w, all, format)
	}
	if len(all) == 0 {
		_, err := fmt.Fprintln(errW, "Listed 0 items.")
		return err
	}
	rows := make([]map[string]any, len(all))
	for i, r := range all {
		name, _ := r["name"].(string)
		fw, _ := r["framework"].(map[string]any)
		rows[i] = map[string]any{
			"id":             name[strings.LastIndex(name, "/")+1:],
			"displayName":    r["targetResourceDisplayName"],
			"targetResource": r["computedTargetResource"],
			"framework":      fw["framework"],
			"state":          r["state"],
			"createTime":     r["createTime"],
		}
	}
	return emitFormattedTo(w, rows, "table(id,displayName,targetResource,framework,state,createTime)")
}
