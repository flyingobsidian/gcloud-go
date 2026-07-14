package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	securitycenter "google.golang.org/api/securitycenter/v1"
)

// --- gcloud scc (#381, #805-#818) ---
//
// Real API-backed implementations for Security Command Center v1
// (securitycenter.googleapis.com) and Security Posture v1
// (securityposture.googleapis.com; see scc_posture.go).

var sccCmd = &cobra.Command{Use: "scc", Short: "Manage Security Command Center"}

// Common flags shared across all scc groups. Callers must set exactly one of
// --organization, --folder, or --project; if none are set the resolved active
// project is used.
var (
	flagSccOrg        string
	flagSccFolder     string
	flagSccProject    string
	flagSccLocation   string
	flagSccFormat     string
	flagSccFilter     string
	flagSccOrderBy    string
	flagSccPageSize   int64
	flagSccConfigFile string
	flagSccUpdateMask string

	// assets flags
	flagSccAssetsResourceName    string
	flagSccAssetsAssetID         string
	flagSccAssetsCompareDuration string
	flagSccAssetsReadTime        string
	flagSccAssetsFieldMask       string
	flagSccAssetsGroupBy         string

	// findings flags
	flagSccFindingSource         string
	flagSccFindingMuteState      string
	flagSccFindingState          string
	flagSccFindingStartTime      string
	flagSccFindingBulkMuteFilter string
	flagSccFindingGroupBy        string
	flagSccFindingCompareDur     string
	flagSccFindingReadTime       string
	flagSccFindingFieldMask      string
)

// --- scope resolution ---

// sccResolveParent returns the SCC scope parent as
// "organizations/{id}", "folders/{id}", or "projects/{id}". If no scope flag
// is set, it falls back to the active gcloud project.
func sccResolveParent() (string, error) {
	set := 0
	for _, v := range []string{flagSccOrg, flagSccFolder, flagSccProject} {
		if v != "" {
			set++
		}
	}
	if set > 1 {
		return "", fmt.Errorf("--organization, --folder, and --project are mutually exclusive")
	}
	if flagSccOrg != "" {
		return "organizations/" + strings.TrimPrefix(flagSccOrg, "organizations/"), nil
	}
	if flagSccFolder != "" {
		return "folders/" + strings.TrimPrefix(flagSccFolder, "folders/"), nil
	}
	if flagSccProject != "" {
		return "projects/" + strings.TrimPrefix(flagSccProject, "projects/"), nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

// sccQualifyChild joins parent + collection + id (returning id unchanged if
// it is already a fully-qualified resource name).
func sccQualifyChild(parent, collection, id string) string {
	if strings.Contains(id, "/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// sccFlagStringVars registers the given --organization/--folder/--project set
// on cmd.
func sccAddScopeFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagSccOrg, "organization", "", "Organization ID or fully-qualified organizations/{id}")
		c.Flags().StringVar(&flagSccFolder, "folder", "", "Folder ID or fully-qualified folders/{id}")
		c.Flags().StringVar(&flagSccProject, "project", "", "Project ID or fully-qualified projects/{id}")
	}
}

func sccAddFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagSccFormat, "format", "", "Output format")
	}
}

func sccAddFilterFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagSccFilter, "filter", "", "Server-side list filter")
	}
}

func sccAddPageSizeFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().Int64Var(&flagSccPageSize, "page-size", 0, "Page size for list requests")
	}
}

// --- assets ---

var sccAssetsCmd = &cobra.Command{Use: "assets", Short: "Manage Security Command Center assets"}

var (
	sccAssetsDescribeCmd = &cobra.Command{
		Use: "describe PARENT", Short: "Describe an asset",
		Args: cobra.ExactArgs(1), RunE: runSccAssetsDescribe,
	}
	sccAssetsGroupCmd = &cobra.Command{
		Use: "group PARENT", Short: "Group assets by fields",
		Args: cobra.ExactArgs(1), RunE: runSccAssetsGroup,
	}
	sccAssetsListCmd = &cobra.Command{
		Use: "list PARENT", Short: "List Security Command Center assets",
		Args: cobra.ExactArgs(1), RunE: runSccAssetsList,
	}
	sccAssetsRunDiscoveryCmd = &cobra.Command{
		Use: "run-discovery ORGANIZATION", Short: "Run asset discovery for an organization",
		Args: cobra.ExactArgs(1), RunE: runSccAssetsRunDiscovery,
	}
	sccAssetsUpdateMarksCmd = &cobra.Command{
		Use: "update-security-marks ASSET_NAME", Short: "Update security marks on an asset",
		Args: cobra.ExactArgs(1), RunE: runSccAssetsUpdateMarks,
	}
)

// --- bqexports ---

var sccBQExportsCmd = &cobra.Command{Use: "bqexports", Short: "Manage BigQuery exports"}

var (
	sccBQExportsCreateCmd = &cobra.Command{
		Use: "create EXPORT", Short: "Create a BigQuery export from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccBQExportCreate,
	}
	sccBQExportsDeleteCmd = &cobra.Command{
		Use: "delete EXPORT", Short: "Delete a BigQuery export",
		Args: cobra.ExactArgs(1), RunE: runSccBQExportDelete,
	}
	sccBQExportsDescribeCmd = &cobra.Command{
		Use: "describe EXPORT", Short: "Describe a BigQuery export",
		Args: cobra.ExactArgs(1), RunE: runSccBQExportDescribe,
	}
	sccBQExportsListCmd = &cobra.Command{
		Use: "list", Short: "List BigQuery exports",
		Args: cobra.NoArgs, RunE: runSccBQExportList,
	}
	sccBQExportsUpdateCmd = &cobra.Command{
		Use: "update EXPORT", Short: "Update a BigQuery export from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccBQExportUpdate,
	}
)

// --- custom-modules ---

var sccCustomModulesCmd = &cobra.Command{Use: "custom-modules", Short: "Manage SCC custom modules"}

// custom-modules etd
var sccCustomModulesEtdCmd = &cobra.Command{Use: "etd", Short: "Manage Event Threat Detection custom modules"}

var (
	sccEtdCreateCmd = &cobra.Command{
		Use: "create MODULE", Short: "Create an ETD custom module from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccEtdCreate,
	}
	sccEtdDeleteCmd = &cobra.Command{
		Use: "delete MODULE", Short: "Delete an ETD custom module",
		Args: cobra.ExactArgs(1), RunE: runSccEtdDelete,
	}
	sccEtdDescribeCmd = &cobra.Command{
		Use: "describe MODULE", Short: "Describe an ETD custom module",
		Args: cobra.ExactArgs(1), RunE: runSccEtdDescribe,
	}
	sccEtdDescribeEffectiveCmd = &cobra.Command{
		Use: "describe-effective MODULE", Short: "Describe an effective ETD custom module",
		Args: cobra.ExactArgs(1), RunE: runSccEtdDescribeEffective,
	}
	sccEtdListCmd = &cobra.Command{
		Use: "list", Short: "List ETD custom modules",
		Args: cobra.NoArgs, RunE: runSccEtdList,
	}
	sccEtdListDescendantCmd = &cobra.Command{
		Use: "list-descendant", Short: "List ETD custom modules and descendants",
		Args: cobra.NoArgs, RunE: runSccEtdListDescendant,
	}
	sccEtdListEffectiveCmd = &cobra.Command{
		Use: "list-effective", Short: "List effective ETD custom modules",
		Args: cobra.NoArgs, RunE: runSccEtdListEffective,
	}
	sccEtdUpdateCmd = &cobra.Command{
		Use: "update MODULE", Short: "Update an ETD custom module from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccEtdUpdate,
	}
)

// custom-modules sha
var sccCustomModulesShaCmd = &cobra.Command{Use: "sha", Short: "Manage Security Health Analytics custom modules"}

var (
	sccShaCreateCmd = &cobra.Command{
		Use: "create MODULE", Short: "Create a SHA custom module from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccShaCreate,
	}
	sccShaDeleteCmd = &cobra.Command{
		Use: "delete MODULE", Short: "Delete a SHA custom module",
		Args: cobra.ExactArgs(1), RunE: runSccShaDelete,
	}
	sccShaDescribeCmd = &cobra.Command{
		Use: "describe MODULE", Short: "Describe a SHA custom module",
		Args: cobra.ExactArgs(1), RunE: runSccShaDescribe,
	}
	sccShaDescribeEffectiveCmd = &cobra.Command{
		Use: "describe-effective MODULE", Short: "Describe an effective SHA custom module",
		Args: cobra.ExactArgs(1), RunE: runSccShaDescribeEffective,
	}
	sccShaListCmd = &cobra.Command{
		Use: "list", Short: "List SHA custom modules",
		Args: cobra.NoArgs, RunE: runSccShaList,
	}
	sccShaListDescendantCmd = &cobra.Command{
		Use: "list-descendant", Short: "List SHA custom modules and descendants",
		Args: cobra.NoArgs, RunE: runSccShaListDescendant,
	}
	sccShaListEffectiveCmd = &cobra.Command{
		Use: "list-effective", Short: "List effective SHA custom modules",
		Args: cobra.NoArgs, RunE: runSccShaListEffective,
	}
	sccShaUpdateCmd = &cobra.Command{
		Use: "update MODULE", Short: "Update a SHA custom module from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccShaUpdate,
	}
)

// --- findings ---

var sccFindingsCmd = &cobra.Command{Use: "findings", Short: "Manage findings"}

var (
	sccFindingsBulkMuteCmd = &cobra.Command{
		Use: "bulk-mute", Short: "Bulk-mute findings that match a filter",
		Args: cobra.NoArgs, RunE: runSccFindingsBulkMute,
	}
	sccFindingsCreateCmd = &cobra.Command{
		Use: "create FINDING_ID", Short: "Create a finding from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccFindingsCreate,
	}
	sccFindingsGroupCmd = &cobra.Command{
		Use: "group", Short: "Group findings by fields",
		Args: cobra.NoArgs, RunE: runSccFindingsGroup,
	}
	sccFindingsListCmd = &cobra.Command{
		Use: "list", Short: "List findings",
		Args: cobra.NoArgs, RunE: runSccFindingsList,
	}
	sccFindingsListMarksCmd = &cobra.Command{
		Use: "list-marks FINDING", Short: "List security marks on a finding",
		Args: cobra.ExactArgs(1), RunE: runSccFindingsListMarks,
	}
	sccFindingsSetMuteCmd = &cobra.Command{
		Use: "set-mute FINDING", Short: "Set the mute state of a finding",
		Args: cobra.ExactArgs(1), RunE: runSccFindingsSetMute,
	}
	sccFindingsSetStateCmd = &cobra.Command{
		Use: "set-state FINDING", Short: "Set the state of a finding",
		Args: cobra.ExactArgs(1), RunE: runSccFindingsSetState,
	}
	sccFindingsUpdateCmd = &cobra.Command{
		Use: "update FINDING", Short: "Update a finding from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccFindingsUpdate,
	}
	sccFindingsUpdateMarksCmd = &cobra.Command{
		Use: "update-security-marks FINDING", Short: "Update security marks on a finding",
		Args: cobra.ExactArgs(1), RunE: runSccFindingsUpdateMarks,
	}
)

// --- manage ---

var sccManageCmd = &cobra.Command{Use: "manage", Short: "Manage Security Command Center settings"}

var sccManageSettingsCmd = &cobra.Command{Use: "settings", Short: "Manage organization SCC settings"}

var (
	sccManageSettingsDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe an organization's SCC settings",
		Args: cobra.NoArgs, RunE: runSccManageSettingsDescribe,
	}
	sccManageSettingsUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update an organization's SCC settings from a --config-file",
		Args: cobra.NoArgs, RunE: runSccManageSettingsUpdate,
	}
)

// --- muteconfigs ---

var sccMuteConfigsCmd = &cobra.Command{Use: "muteconfigs", Short: "Manage mute configs"}

var (
	sccMuteCreateCmd = &cobra.Command{
		Use: "create MUTE_CONFIG", Short: "Create a mute config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccMuteCreate,
	}
	sccMuteDeleteCmd = &cobra.Command{
		Use: "delete MUTE_CONFIG", Short: "Delete a mute config",
		Args: cobra.ExactArgs(1), RunE: runSccMuteDelete,
	}
	sccMuteDescribeCmd = &cobra.Command{
		Use: "describe MUTE_CONFIG", Short: "Describe a mute config",
		Args: cobra.ExactArgs(1), RunE: runSccMuteDescribe,
	}
	sccMuteListCmd = &cobra.Command{
		Use: "list", Short: "List mute configs",
		Args: cobra.NoArgs, RunE: runSccMuteList,
	}
	sccMuteUpdateCmd = &cobra.Command{
		Use: "update MUTE_CONFIG", Short: "Update a mute config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccMuteUpdate,
	}
)

// --- notifications ---

var sccNotificationsCmd = &cobra.Command{Use: "notifications", Short: "Manage notifications"}

var (
	sccNotifyCreateCmd = &cobra.Command{
		Use: "create NOTIFICATION_CONFIG", Short: "Create a notification config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccNotifyCreate,
	}
	sccNotifyDeleteCmd = &cobra.Command{
		Use: "delete NOTIFICATION_CONFIG", Short: "Delete a notification config",
		Args: cobra.ExactArgs(1), RunE: runSccNotifyDelete,
	}
	sccNotifyDescribeCmd = &cobra.Command{
		Use: "describe NOTIFICATION_CONFIG", Short: "Describe a notification config",
		Args: cobra.ExactArgs(1), RunE: runSccNotifyDescribe,
	}
	sccNotifyListCmd = &cobra.Command{
		Use: "list", Short: "List notification configs",
		Args: cobra.NoArgs, RunE: runSccNotifyList,
	}
	sccNotifyUpdateCmd = &cobra.Command{
		Use: "update NOTIFICATION_CONFIG", Short: "Update a notification config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSccNotifyUpdate,
	}
)

// --- operations ---

var sccOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Security Command Center operations"}

var (
	sccOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an SCC operation",
		Args: cobra.ExactArgs(1), RunE: runSccOpDescribe,
	}
	sccOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an SCC operation",
		Args: cobra.ExactArgs(1), RunE: runSccOpDelete,
	}
	sccOpListCmd = &cobra.Command{
		Use: "list", Short: "List SCC operations",
		Args: cobra.NoArgs, RunE: runSccOpList,
	}
)

// --- sources ---

var sccSourcesCmd = &cobra.Command{Use: "sources", Short: "Manage Security Command Center sources"}

var (
	sccSourcesDescribeCmd = &cobra.Command{
		Use: "describe SOURCE", Short: "Describe a Security Command Center source",
		Args: cobra.ExactArgs(1), RunE: runSccSourcesDescribe,
	}
	sccSourcesListCmd = &cobra.Command{
		Use: "list", Short: "List Security Command Center sources",
		Args: cobra.NoArgs, RunE: runSccSourcesList,
	}
)

func init() {
	// --- assets flags ---
	sccAddScopeFlags(sccAssetsDescribeCmd, sccAssetsGroupCmd, sccAssetsListCmd,
		sccAssetsRunDiscoveryCmd, sccAssetsUpdateMarksCmd)
	sccAddFormatFlag(sccAssetsDescribeCmd, sccAssetsGroupCmd, sccAssetsListCmd, sccAssetsUpdateMarksCmd)
	sccAddFilterFlag(sccAssetsGroupCmd, sccAssetsListCmd)
	sccAddPageSizeFlag(sccAssetsGroupCmd, sccAssetsListCmd)
	sccAssetsDescribeCmd.Flags().StringVar(&flagSccAssetsResourceName, "resource-name", "",
		"Full resource name of the asset to describe (e.g. //storage.googleapis.com/my-bucket)")
	sccAssetsDescribeCmd.Flags().StringVar(&flagSccAssetsAssetID, "asset", "", "Numeric SCC asset id")
	sccAssetsListCmd.Flags().StringVar(&flagSccAssetsCompareDuration, "compare-duration", "", "Compare-duration for state-change results")
	sccAssetsListCmd.Flags().StringVar(&flagSccAssetsReadTime, "read-time", "", "Read time (RFC3339)")
	sccAssetsListCmd.Flags().StringVar(&flagSccAssetsFieldMask, "field-mask", "", "Field mask for the response")
	sccAssetsListCmd.Flags().StringVar(&flagSccOrderBy, "order-by", "", "Server-side ordering expression")
	sccAssetsGroupCmd.Flags().StringVar(&flagSccAssetsGroupBy, "group-by", "", "Comma-separated list of fields to group by (required)")
	_ = sccAssetsGroupCmd.MarkFlagRequired("group-by")
	sccAssetsUpdateMarksCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the SecurityMarks body (required)")
	_ = sccAssetsUpdateMarksCmd.MarkFlagRequired("config-file")
	sccAssetsUpdateMarksCmd.Flags().StringVar(&flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	sccAssetsCmd.AddCommand(sccAssetsDescribeCmd, sccAssetsGroupCmd, sccAssetsListCmd,
		sccAssetsRunDiscoveryCmd, sccAssetsUpdateMarksCmd)
	sccCmd.AddCommand(sccAssetsCmd)

	// --- bqexports flags ---
	sccAddScopeFlags(sccBQExportsCreateCmd, sccBQExportsDeleteCmd, sccBQExportsDescribeCmd,
		sccBQExportsListCmd, sccBQExportsUpdateCmd)
	sccAddFormatFlag(sccBQExportsDescribeCmd, sccBQExportsListCmd)
	sccAddPageSizeFlag(sccBQExportsListCmd)
	sccBQExportsCreateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the BigQueryExport body (required)")
	_ = sccBQExportsCreateCmd.MarkFlagRequired("config-file")
	sccBQExportsUpdateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the BigQueryExport body (required)")
	_ = sccBQExportsUpdateCmd.MarkFlagRequired("config-file")
	sccBQExportsUpdateCmd.Flags().StringVar(&flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	sccBQExportsCmd.AddCommand(sccBQExportsCreateCmd, sccBQExportsDeleteCmd, sccBQExportsDescribeCmd,
		sccBQExportsListCmd, sccBQExportsUpdateCmd)
	sccCmd.AddCommand(sccBQExportsCmd)

	// --- custom-modules flags ---
	etdCmds := []*cobra.Command{sccEtdCreateCmd, sccEtdDeleteCmd, sccEtdDescribeCmd,
		sccEtdDescribeEffectiveCmd, sccEtdListCmd, sccEtdListDescendantCmd,
		sccEtdListEffectiveCmd, sccEtdUpdateCmd}
	shaCmds := []*cobra.Command{sccShaCreateCmd, sccShaDeleteCmd, sccShaDescribeCmd,
		sccShaDescribeEffectiveCmd, sccShaListCmd, sccShaListDescendantCmd,
		sccShaListEffectiveCmd, sccShaUpdateCmd}
	sccAddScopeFlags(etdCmds...)
	sccAddScopeFlags(shaCmds...)
	sccAddFormatFlag(sccEtdDescribeCmd, sccEtdDescribeEffectiveCmd, sccEtdListCmd,
		sccEtdListDescendantCmd, sccEtdListEffectiveCmd,
		sccShaDescribeCmd, sccShaDescribeEffectiveCmd, sccShaListCmd,
		sccShaListDescendantCmd, sccShaListEffectiveCmd)
	sccAddPageSizeFlag(sccEtdListCmd, sccEtdListDescendantCmd, sccEtdListEffectiveCmd,
		sccShaListCmd, sccShaListDescendantCmd, sccShaListEffectiveCmd)
	for _, c := range []*cobra.Command{sccEtdCreateCmd, sccEtdUpdateCmd,
		sccShaCreateCmd, sccShaUpdateCmd} {
		c.Flags().StringVar(&flagSccConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the custom module body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{sccEtdUpdateCmd, sccShaUpdateCmd} {
		c.Flags().StringVar(&flagSccUpdateMask, "update-mask", "",
			"Comma-separated list of fields to update")
	}
	sccCustomModulesEtdCmd.AddCommand(etdCmds...)
	sccCustomModulesShaCmd.AddCommand(shaCmds...)
	sccCustomModulesCmd.AddCommand(sccCustomModulesEtdCmd, sccCustomModulesShaCmd)
	sccCmd.AddCommand(sccCustomModulesCmd)

	// --- findings flags ---
	findAll := []*cobra.Command{sccFindingsBulkMuteCmd, sccFindingsCreateCmd, sccFindingsGroupCmd,
		sccFindingsListCmd, sccFindingsListMarksCmd, sccFindingsSetMuteCmd, sccFindingsSetStateCmd,
		sccFindingsUpdateCmd, sccFindingsUpdateMarksCmd}
	sccAddScopeFlags(findAll...)
	sccAddFormatFlag(sccFindingsCreateCmd, sccFindingsGroupCmd, sccFindingsListCmd,
		sccFindingsListMarksCmd, sccFindingsSetMuteCmd, sccFindingsSetStateCmd,
		sccFindingsUpdateCmd, sccFindingsUpdateMarksCmd)
	for _, c := range []*cobra.Command{sccFindingsCreateCmd, sccFindingsGroupCmd,
		sccFindingsListCmd, sccFindingsListMarksCmd, sccFindingsSetMuteCmd, sccFindingsSetStateCmd,
		sccFindingsUpdateCmd, sccFindingsUpdateMarksCmd} {
		c.Flags().StringVar(&flagSccFindingSource, "source", "-",
			"Source id (or fully-qualified sources/{id}). Defaults to \"-\" (all sources)")
	}
	sccFindingsBulkMuteCmd.Flags().StringVar(&flagSccFindingBulkMuteFilter, "filter", "",
		"Filter expression identifying findings to mute (required)")
	_ = sccFindingsBulkMuteCmd.MarkFlagRequired("filter")
	sccFindingsCreateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Finding body (required)")
	_ = sccFindingsCreateCmd.MarkFlagRequired("config-file")
	sccFindingsListCmd.Flags().StringVar(&flagSccFilter, "filter", "", "Server-side list filter")
	sccFindingsListCmd.Flags().StringVar(&flagSccOrderBy, "order-by", "", "Server-side ordering expression")
	sccFindingsListCmd.Flags().StringVar(&flagSccFindingFieldMask, "field-mask", "", "Response field mask")
	sccFindingsListCmd.Flags().StringVar(&flagSccFindingCompareDur, "compare-duration", "", "Compare-duration for state-change results")
	sccFindingsListCmd.Flags().StringVar(&flagSccFindingReadTime, "read-time", "", "Read time (RFC3339)")
	sccFindingsListCmd.Flags().Int64Var(&flagSccPageSize, "page-size", 0, "Page size for list requests")
	sccFindingsGroupCmd.Flags().StringVar(&flagSccFindingGroupBy, "group-by", "",
		"Comma-separated list of fields to group by (required)")
	_ = sccFindingsGroupCmd.MarkFlagRequired("group-by")
	sccFindingsGroupCmd.Flags().StringVar(&flagSccFilter, "filter", "", "Server-side filter")
	sccFindingsGroupCmd.Flags().StringVar(&flagSccFindingCompareDur, "compare-duration", "", "Compare-duration for state-change results")
	sccFindingsGroupCmd.Flags().StringVar(&flagSccFindingReadTime, "read-time", "", "Read time (RFC3339)")
	sccFindingsGroupCmd.Flags().Int64Var(&flagSccPageSize, "page-size", 0, "Page size for group requests")
	sccFindingsSetMuteCmd.Flags().StringVar(&flagSccFindingMuteState, "mute", "",
		"Mute state (MUTE_UNSPECIFIED|MUTED|UNMUTED|UNDEFINED) (required)")
	_ = sccFindingsSetMuteCmd.MarkFlagRequired("mute")
	sccFindingsSetStateCmd.Flags().StringVar(&flagSccFindingState, "state", "",
		"Finding state (STATE_UNSPECIFIED|ACTIVE|INACTIVE) (required)")
	_ = sccFindingsSetStateCmd.MarkFlagRequired("state")
	sccFindingsSetStateCmd.Flags().StringVar(&flagSccFindingStartTime, "start-time", "",
		"Time at which the updated state takes effect (RFC3339)")
	sccFindingsUpdateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Finding body (required)")
	_ = sccFindingsUpdateCmd.MarkFlagRequired("config-file")
	sccFindingsUpdateCmd.Flags().StringVar(&flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	sccFindingsUpdateMarksCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the SecurityMarks body (required)")
	_ = sccFindingsUpdateMarksCmd.MarkFlagRequired("config-file")
	sccFindingsUpdateMarksCmd.Flags().StringVar(&flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	sccFindingsCmd.AddCommand(findAll...)
	sccCmd.AddCommand(sccFindingsCmd)

	// --- manage flags ---
	sccManageSettingsDescribeCmd.Flags().StringVar(&flagSccOrg, "organization", "",
		"Organization ID (required)")
	_ = sccManageSettingsDescribeCmd.MarkFlagRequired("organization")
	sccManageSettingsDescribeCmd.Flags().StringVar(&flagSccFormat, "format", "", "Output format")
	sccManageSettingsUpdateCmd.Flags().StringVar(&flagSccOrg, "organization", "",
		"Organization ID (required)")
	_ = sccManageSettingsUpdateCmd.MarkFlagRequired("organization")
	sccManageSettingsUpdateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the OrganizationSettings body (required)")
	_ = sccManageSettingsUpdateCmd.MarkFlagRequired("config-file")
	sccManageSettingsUpdateCmd.Flags().StringVar(&flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	sccManageSettingsCmd.AddCommand(sccManageSettingsDescribeCmd, sccManageSettingsUpdateCmd)
	sccManageCmd.AddCommand(sccManageSettingsCmd)
	sccCmd.AddCommand(sccManageCmd)

	// --- muteconfigs flags ---
	muteAll := []*cobra.Command{sccMuteCreateCmd, sccMuteDeleteCmd, sccMuteDescribeCmd,
		sccMuteListCmd, sccMuteUpdateCmd}
	sccAddScopeFlags(muteAll...)
	sccAddFormatFlag(sccMuteDescribeCmd, sccMuteListCmd)
	sccAddPageSizeFlag(sccMuteListCmd)
	sccMuteCreateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the MuteConfig body (required)")
	_ = sccMuteCreateCmd.MarkFlagRequired("config-file")
	sccMuteUpdateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the MuteConfig body (required)")
	_ = sccMuteUpdateCmd.MarkFlagRequired("config-file")
	sccMuteUpdateCmd.Flags().StringVar(&flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	sccMuteConfigsCmd.AddCommand(muteAll...)
	sccCmd.AddCommand(sccMuteConfigsCmd)

	// --- notifications flags ---
	notifyAll := []*cobra.Command{sccNotifyCreateCmd, sccNotifyDeleteCmd, sccNotifyDescribeCmd,
		sccNotifyListCmd, sccNotifyUpdateCmd}
	sccAddScopeFlags(notifyAll...)
	sccAddFormatFlag(sccNotifyDescribeCmd, sccNotifyListCmd)
	sccAddPageSizeFlag(sccNotifyListCmd)
	sccNotifyCreateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the NotificationConfig body (required)")
	_ = sccNotifyCreateCmd.MarkFlagRequired("config-file")
	sccNotifyUpdateCmd.Flags().StringVar(&flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the NotificationConfig body (required)")
	_ = sccNotifyUpdateCmd.MarkFlagRequired("config-file")
	sccNotifyUpdateCmd.Flags().StringVar(&flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	sccNotificationsCmd.AddCommand(notifyAll...)
	sccCmd.AddCommand(sccNotificationsCmd)

	// --- operations flags (organization-scoped only in v1) ---
	for _, c := range []*cobra.Command{sccOpDescribeCmd, sccOpDeleteCmd, sccOpListCmd} {
		c.Flags().StringVar(&flagSccOrg, "organization", "",
			"Organization ID (required unless the operation name is fully-qualified)")
	}
	sccAddFormatFlag(sccOpDescribeCmd, sccOpListCmd)
	sccAddFilterFlag(sccOpListCmd)
	sccAddPageSizeFlag(sccOpListCmd)
	sccOperationsCmd.AddCommand(sccOpDescribeCmd, sccOpDeleteCmd, sccOpListCmd)
	sccCmd.AddCommand(sccOperationsCmd)

	// --- sources flags ---
	sccAddScopeFlags(sccSourcesDescribeCmd, sccSourcesListCmd)
	sccAddFormatFlag(sccSourcesDescribeCmd, sccSourcesListCmd)
	sccAddPageSizeFlag(sccSourcesListCmd)
	sccSourcesCmd.AddCommand(sccSourcesDescribeCmd, sccSourcesListCmd)
	sccCmd.AddCommand(sccSourcesCmd)

	// Register scc_posture.go subgroups (see init in that file).
	rootCmd.AddCommand(sccCmd)
}

// --- helpers for scope-dispatched API calls ---

// sccClient returns an initialised securitycenter service.
func sccClient(ctx context.Context) (*securitycenter.Service, error) {
	return gcp.SecurityCenterService(ctx, flagAccount)
}

// splitScope returns ("organizations"|"folders"|"projects", "{id}") from a
// parent like "organizations/123". Returns ("", "") for anything else.
func splitScope(parent string) (string, string) {
	parts := strings.SplitN(parent, "/", 3)
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// --- assets impl ---

func runSccAssetsDescribe(cmd *cobra.Command, args []string) error {
	// Describe is implemented client-side by listing with a filter that pins
	// the specific asset. This mirrors gcloud's DescribeAssetReqHook behavior.
	parent := args[0]
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	filter := flagSccFilter
	switch {
	case flagSccAssetsResourceName != "":
		filter = fmt.Sprintf("securityCenterProperties.resourceName=%q", flagSccAssetsResourceName)
	case flagSccAssetsAssetID != "":
		filter = fmt.Sprintf("name : %q", "/assets/"+flagSccAssetsAssetID)
	}
	scope, _ := splitScope(parent)
	var assets []*securitycenter.ListAssetsResult
	switch scope {
	case "organizations":
		call := svc.Organizations.Assets.List(parent).Context(ctx)
		if filter != "" {
			call = call.Filter(filter)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("describing asset: %w", err)
		}
		assets = resp.ListAssetsResults
	case "folders":
		call := svc.Folders.Assets.List(parent).Context(ctx)
		if filter != "" {
			call = call.Filter(filter)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("describing asset: %w", err)
		}
		assets = resp.ListAssetsResults
	case "projects":
		call := svc.Projects.Assets.List(parent).Context(ctx)
		if filter != "" {
			call = call.Filter(filter)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("describing asset: %w", err)
		}
		assets = resp.ListAssetsResults
	default:
		return fmt.Errorf("invalid parent %q; expected organizations|folders|projects/{id}", parent)
	}
	if len(assets) == 0 {
		return fmt.Errorf("no asset matched")
	}
	if len(assets) > 1 {
		return fmt.Errorf("filter matched %d assets; refine --resource-name/--asset", len(assets))
	}
	return emitFormatted(assets[0], flagSccFormat)
}

func runSccAssetsGroup(cmd *cobra.Command, args []string) error {
	parent := args[0]
	req := &securitycenter.GroupAssetsRequest{
		GroupBy:  flagSccAssetsGroupBy,
		Filter:   flagSccFilter,
		PageSize: flagSccPageSize,
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var resp any
	switch scope {
	case "organizations":
		resp, err = svc.Organizations.Assets.Group(parent, req).Context(ctx).Do()
	case "folders":
		resp, err = svc.Folders.Assets.Group(parent, req).Context(ctx).Do()
	case "projects":
		resp, err = svc.Projects.Assets.Group(parent, req).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("grouping assets: %w", err)
	}
	return emitFormatted(resp, flagSccFormat)
}

func runSccAssetsList(cmd *cobra.Command, args []string) error {
	parent := args[0]
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var all []*securitycenter.ListAssetsResult
	pageToken := ""
	for {
		var (
			results       []*securitycenter.ListAssetsResult
			nextPageToken string
		)
		switch scope {
		case "organizations":
			call := svc.Organizations.Assets.List(parent).Context(ctx)
			if flagSccFilter != "" {
				call = call.Filter(flagSccFilter)
			}
			if flagSccOrderBy != "" {
				call = call.OrderBy(flagSccOrderBy)
			}
			if flagSccAssetsCompareDuration != "" {
				call = call.CompareDuration(flagSccAssetsCompareDuration)
			}
			if flagSccAssetsReadTime != "" {
				call = call.ReadTime(flagSccAssetsReadTime)
			}
			if flagSccAssetsFieldMask != "" {
				call = call.FieldMask(flagSccAssetsFieldMask)
			}
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing assets: %w", err)
			}
			results = resp.ListAssetsResults
			nextPageToken = resp.NextPageToken
		case "folders":
			call := svc.Folders.Assets.List(parent).Context(ctx)
			if flagSccFilter != "" {
				call = call.Filter(flagSccFilter)
			}
			if flagSccOrderBy != "" {
				call = call.OrderBy(flagSccOrderBy)
			}
			if flagSccAssetsCompareDuration != "" {
				call = call.CompareDuration(flagSccAssetsCompareDuration)
			}
			if flagSccAssetsReadTime != "" {
				call = call.ReadTime(flagSccAssetsReadTime)
			}
			if flagSccAssetsFieldMask != "" {
				call = call.FieldMask(flagSccAssetsFieldMask)
			}
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing assets: %w", err)
			}
			results = resp.ListAssetsResults
			nextPageToken = resp.NextPageToken
		case "projects":
			call := svc.Projects.Assets.List(parent).Context(ctx)
			if flagSccFilter != "" {
				call = call.Filter(flagSccFilter)
			}
			if flagSccOrderBy != "" {
				call = call.OrderBy(flagSccOrderBy)
			}
			if flagSccAssetsCompareDuration != "" {
				call = call.CompareDuration(flagSccAssetsCompareDuration)
			}
			if flagSccAssetsReadTime != "" {
				call = call.ReadTime(flagSccAssetsReadTime)
			}
			if flagSccAssetsFieldMask != "" {
				call = call.FieldMask(flagSccAssetsFieldMask)
			}
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing assets: %w", err)
			}
			results = resp.ListAssetsResults
			nextPageToken = resp.NextPageToken
		default:
			return fmt.Errorf("invalid parent %q", parent)
		}
		all = append(all, results...)
		if nextPageToken == "" {
			break
		}
		pageToken = nextPageToken
	}
	if flagSccFormat != "" {
		return emitFormatted(all, flagSccFormat)
	}
	fmt.Printf("%-60s %s\n", "NAME", "RESOURCE_TYPE")
	for _, a := range all {
		if a.Asset != nil {
			name := a.Asset.Name
			var rtype string
			if a.Asset.SecurityCenterProperties != nil {
				rtype = a.Asset.SecurityCenterProperties.ResourceType
			}
			fmt.Printf("%-60s %s\n", path.Base(name), rtype)
		}
	}
	return nil
}

func runSccAssetsRunDiscovery(cmd *cobra.Command, args []string) error {
	org := args[0]
	if !strings.HasPrefix(org, "organizations/") {
		org = "organizations/" + org
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Assets.RunDiscovery(org, &securitycenter.RunAssetDiscoveryRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running discovery: %w", err)
	}
	return emitFormatted(op, flagSccFormat)
}

func runSccAssetsUpdateMarks(cmd *cobra.Command, args []string) error {
	name := args[0]
	// Expect assetName in the form <parent>/assets/{id}/securityMarks or the
	// full asset resource name; append /securityMarks if missing.
	if !strings.HasSuffix(name, "/securityMarks") {
		name = strings.TrimSuffix(name, "/") + "/securityMarks"
	}
	body := &securitycenter.SecurityMarks{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.SecurityMarks
	switch scope {
	case "organizations":
		got, err = svc.Organizations.Assets.UpdateSecurityMarks(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Assets.UpdateSecurityMarks(name, body).UpdateMask(mask).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.Assets.UpdateSecurityMarks(name, body).UpdateMask(mask).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid asset name %q", name)
	}
	if err != nil {
		return fmt.Errorf("updating security marks: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

// --- bqexports impl ---

func runSccBQExportCreate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.GoogleCloudSecuritycenterV1BigQueryExport{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var got *securitycenter.GoogleCloudSecuritycenterV1BigQueryExport
	switch scope {
	case "organizations":
		got, err = svc.Organizations.BigQueryExports.Create(parent, body).BigQueryExportId(args[0]).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.BigQueryExports.Create(parent, body).BigQueryExportId(args[0]).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.BigQueryExports.Create(parent, body).BigQueryExportId(args[0]).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("creating BigQuery export: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccBQExportDelete(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "bigQueryExports", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	switch scope {
	case "organizations":
		_, err = svc.Organizations.BigQueryExports.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.BigQueryExports.Delete(name).Context(ctx).Do()
	case "projects":
		_, err = svc.Projects.BigQueryExports.Delete(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("deleting BigQuery export: %w", err)
	}
	fmt.Printf("Deleted BigQuery export [%s].\n", args[0])
	return nil
}

func runSccBQExportDescribe(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "bigQueryExports", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.GoogleCloudSecuritycenterV1BigQueryExport
	switch scope {
	case "organizations":
		got, err = svc.Organizations.BigQueryExports.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.BigQueryExports.Get(name).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.BigQueryExports.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing BigQuery export: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccBQExportList(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var all []*securitycenter.GoogleCloudSecuritycenterV1BigQueryExport
	pageToken := ""
	for {
		var (
			page  []*securitycenter.GoogleCloudSecuritycenterV1BigQueryExport
			next  string
			cerr  error
		)
		switch scope {
		case "organizations":
			call := svc.Organizations.BigQueryExports.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, cerr2 := call.Do()
			cerr = cerr2
			if resp != nil {
				page = resp.BigQueryExports
				next = resp.NextPageToken
			}
		case "folders":
			call := svc.Folders.BigQueryExports.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, cerr2 := call.Do()
			cerr = cerr2
			if resp != nil {
				page = resp.BigQueryExports
				next = resp.NextPageToken
			}
		case "projects":
			call := svc.Projects.BigQueryExports.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, cerr2 := call.Do()
			cerr = cerr2
			if resp != nil {
				page = resp.BigQueryExports
				next = resp.NextPageToken
			}
		default:
			return fmt.Errorf("invalid parent %q", parent)
		}
		if cerr != nil {
			return fmt.Errorf("listing BigQuery exports: %w", cerr)
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if flagSccFormat != "" {
		return emitFormatted(all, flagSccFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DATASET")
	for _, e := range all {
		fmt.Printf("%-40s %s\n", path.Base(e.Name), e.Dataset)
	}
	return nil
}

func runSccBQExportUpdate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "bigQueryExports", args[0])
	body := &securitycenter.GoogleCloudSecuritycenterV1BigQueryExport{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.GoogleCloudSecuritycenterV1BigQueryExport
	switch scope {
	case "organizations":
		got, err = svc.Organizations.BigQueryExports.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.BigQueryExports.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.BigQueryExports.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("updating BigQuery export: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

// --- custom-modules etd impl ---

func etdParent(parent string) string {
	return parent + "/eventThreatDetectionSettings"
}

func runSccEtdCreate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.EventThreatDetectionCustomModule{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	if body.DisplayName == "" {
		body.DisplayName = args[0]
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	settings := etdParent(parent)
	var got *securitycenter.EventThreatDetectionCustomModule
	switch scope {
	case "organizations":
		got, err = svc.Organizations.EventThreatDetectionSettings.CustomModules.Create(settings, body).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.EventThreatDetectionSettings.CustomModules.Create(settings, body).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.EventThreatDetectionSettings.CustomModules.Create(settings, body).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("creating ETD custom module: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccEtdDelete(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(etdParent(parent), "customModules", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	switch scope {
	case "organizations":
		_, err = svc.Organizations.EventThreatDetectionSettings.CustomModules.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.EventThreatDetectionSettings.CustomModules.Delete(name).Context(ctx).Do()
	case "projects":
		_, err = svc.Projects.EventThreatDetectionSettings.CustomModules.Delete(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("deleting ETD custom module: %w", err)
	}
	fmt.Printf("Deleted ETD custom module [%s].\n", args[0])
	return nil
}

func runSccEtdDescribe(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(etdParent(parent), "customModules", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.EventThreatDetectionCustomModule
	switch scope {
	case "organizations":
		got, err = svc.Organizations.EventThreatDetectionSettings.CustomModules.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.EventThreatDetectionSettings.CustomModules.Get(name).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.EventThreatDetectionSettings.CustomModules.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing ETD custom module: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccEtdDescribeEffective(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(etdParent(parent), "effectiveCustomModules", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.EffectiveEventThreatDetectionCustomModule
	switch scope {
	case "organizations":
		got, err = svc.Organizations.EventThreatDetectionSettings.EffectiveCustomModules.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.EventThreatDetectionSettings.EffectiveCustomModules.Get(name).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.EventThreatDetectionSettings.EffectiveCustomModules.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing effective ETD custom module: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccEtdList(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	settings := etdParent(parent)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.EventThreatDetectionCustomModule
	switch scope {
	case "organizations":
		call := svc.Organizations.EventThreatDetectionSettings.CustomModules.List(settings).Context(ctx)
		if flagSccPageSize > 0 {
			call = call.PageSize(flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	case "folders":
		call := svc.Folders.EventThreatDetectionSettings.CustomModules.List(settings).Context(ctx)
		if flagSccPageSize > 0 {
			call = call.PageSize(flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	case "projects":
		call := svc.Projects.EventThreatDetectionSettings.CustomModules.List(settings).Context(ctx)
		if flagSccPageSize > 0 {
			call = call.PageSize(flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	}
	if flagSccFormat != "" {
		return emitFormatted(results, flagSccFormat)
	}
	printModulesEtd(results)
	return nil
}

func runSccEtdListDescendant(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	settings := etdParent(parent)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.EventThreatDetectionCustomModule
	switch scope {
	case "organizations":
		resp, err := svc.Organizations.EventThreatDetectionSettings.CustomModules.ListDescendant(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing descendant ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	case "folders":
		resp, err := svc.Folders.EventThreatDetectionSettings.CustomModules.ListDescendant(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing descendant ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	case "projects":
		resp, err := svc.Projects.EventThreatDetectionSettings.CustomModules.ListDescendant(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing descendant ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	}
	if flagSccFormat != "" {
		return emitFormatted(results, flagSccFormat)
	}
	printModulesEtd(results)
	return nil
}

func runSccEtdListEffective(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	settings := etdParent(parent)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.EffectiveEventThreatDetectionCustomModule
	switch scope {
	case "organizations":
		resp, err := svc.Organizations.EventThreatDetectionSettings.EffectiveCustomModules.List(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing effective ETD custom modules: %w", err)
		}
		results = resp.EffectiveEventThreatDetectionCustomModules
	case "folders":
		resp, err := svc.Folders.EventThreatDetectionSettings.EffectiveCustomModules.List(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing effective ETD custom modules: %w", err)
		}
		results = resp.EffectiveEventThreatDetectionCustomModules
	case "projects":
		resp, err := svc.Projects.EventThreatDetectionSettings.EffectiveCustomModules.List(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing effective ETD custom modules: %w", err)
		}
		results = resp.EffectiveEventThreatDetectionCustomModules
	}
	return emitFormatted(results, flagSccFormat)
}

func runSccEtdUpdate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(etdParent(parent), "customModules", args[0])
	body := &securitycenter.EventThreatDetectionCustomModule{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.EventThreatDetectionCustomModule
	switch scope {
	case "organizations":
		got, err = svc.Organizations.EventThreatDetectionSettings.CustomModules.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.EventThreatDetectionSettings.CustomModules.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.EventThreatDetectionSettings.CustomModules.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating ETD custom module: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func printModulesEtd(modules []*securitycenter.EventThreatDetectionCustomModule) {
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, m := range modules {
		fmt.Printf("%-40s %s\n", path.Base(m.Name), m.DisplayName)
	}
}

// --- custom-modules sha impl ---

func shaParent(parent string) string {
	return parent + "/securityHealthAnalyticsSettings"
}

func runSccShaCreate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	if body.DisplayName == "" {
		body.DisplayName = args[0]
	}
	settings := shaParent(parent)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var got *securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule
	switch scope {
	case "organizations":
		got, err = svc.Organizations.SecurityHealthAnalyticsSettings.CustomModules.Create(settings, body).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.SecurityHealthAnalyticsSettings.CustomModules.Create(settings, body).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.SecurityHealthAnalyticsSettings.CustomModules.Create(settings, body).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("creating SHA custom module: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccShaDelete(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(shaParent(parent), "customModules", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	switch scope {
	case "organizations":
		_, err = svc.Organizations.SecurityHealthAnalyticsSettings.CustomModules.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.SecurityHealthAnalyticsSettings.CustomModules.Delete(name).Context(ctx).Do()
	case "projects":
		_, err = svc.Projects.SecurityHealthAnalyticsSettings.CustomModules.Delete(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("deleting SHA custom module: %w", err)
	}
	fmt.Printf("Deleted SHA custom module [%s].\n", args[0])
	return nil
}

func runSccShaDescribe(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(shaParent(parent), "customModules", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule
	switch scope {
	case "organizations":
		got, err = svc.Organizations.SecurityHealthAnalyticsSettings.CustomModules.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.SecurityHealthAnalyticsSettings.CustomModules.Get(name).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.SecurityHealthAnalyticsSettings.CustomModules.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing SHA custom module: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccShaDescribeEffective(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(shaParent(parent), "effectiveCustomModules", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.GoogleCloudSecuritycenterV1EffectiveSecurityHealthAnalyticsCustomModule
	switch scope {
	case "organizations":
		got, err = svc.Organizations.SecurityHealthAnalyticsSettings.EffectiveCustomModules.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.SecurityHealthAnalyticsSettings.EffectiveCustomModules.Get(name).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.SecurityHealthAnalyticsSettings.EffectiveCustomModules.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing effective SHA custom module: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccShaList(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	settings := shaParent(parent)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule
	switch scope {
	case "organizations":
		call := svc.Organizations.SecurityHealthAnalyticsSettings.CustomModules.List(settings).Context(ctx)
		if flagSccPageSize > 0 {
			call = call.PageSize(flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	case "folders":
		call := svc.Folders.SecurityHealthAnalyticsSettings.CustomModules.List(settings).Context(ctx)
		if flagSccPageSize > 0 {
			call = call.PageSize(flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	case "projects":
		call := svc.Projects.SecurityHealthAnalyticsSettings.CustomModules.List(settings).Context(ctx)
		if flagSccPageSize > 0 {
			call = call.PageSize(flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	}
	if flagSccFormat != "" {
		return emitFormatted(results, flagSccFormat)
	}
	printModulesSha(results)
	return nil
}

func runSccShaListDescendant(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	settings := shaParent(parent)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule
	switch scope {
	case "organizations":
		resp, err := svc.Organizations.SecurityHealthAnalyticsSettings.CustomModules.ListDescendant(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing descendant SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	case "folders":
		resp, err := svc.Folders.SecurityHealthAnalyticsSettings.CustomModules.ListDescendant(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing descendant SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	case "projects":
		resp, err := svc.Projects.SecurityHealthAnalyticsSettings.CustomModules.ListDescendant(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing descendant SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	}
	if flagSccFormat != "" {
		return emitFormatted(results, flagSccFormat)
	}
	printModulesSha(results)
	return nil
}

func runSccShaListEffective(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	settings := shaParent(parent)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.GoogleCloudSecuritycenterV1EffectiveSecurityHealthAnalyticsCustomModule
	switch scope {
	case "organizations":
		resp, err := svc.Organizations.SecurityHealthAnalyticsSettings.EffectiveCustomModules.List(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing effective SHA custom modules: %w", err)
		}
		results = resp.EffectiveSecurityHealthAnalyticsCustomModules
	case "folders":
		resp, err := svc.Folders.SecurityHealthAnalyticsSettings.EffectiveCustomModules.List(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing effective SHA custom modules: %w", err)
		}
		results = resp.EffectiveSecurityHealthAnalyticsCustomModules
	case "projects":
		resp, err := svc.Projects.SecurityHealthAnalyticsSettings.EffectiveCustomModules.List(settings).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing effective SHA custom modules: %w", err)
		}
		results = resp.EffectiveSecurityHealthAnalyticsCustomModules
	}
	return emitFormatted(results, flagSccFormat)
}

func runSccShaUpdate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(shaParent(parent), "customModules", args[0])
	body := &securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule
	switch scope {
	case "organizations":
		got, err = svc.Organizations.SecurityHealthAnalyticsSettings.CustomModules.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.SecurityHealthAnalyticsSettings.CustomModules.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.SecurityHealthAnalyticsSettings.CustomModules.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating SHA custom module: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func printModulesSha(modules []*securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule) {
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, m := range modules {
		fmt.Printf("%-40s %s\n", path.Base(m.Name), m.DisplayName)
	}
}

// --- findings impl ---

func findingsSourceParent(parent, source string) string {
	if source == "" {
		source = "-"
	}
	if strings.Contains(source, "/") {
		return source
	}
	return parent + "/sources/" + source
}

func runSccFindingsBulkMute(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	req := &securitycenter.BulkMuteFindingsRequest{Filter: flagSccFindingBulkMuteFilter}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var op *securitycenter.Operation
	switch scope {
	case "organizations":
		op, err = svc.Organizations.Findings.BulkMute(parent, req).Context(ctx).Do()
	case "folders":
		op, err = svc.Folders.Findings.BulkMute(parent, req).Context(ctx).Do()
	case "projects":
		op, err = svc.Projects.Findings.BulkMute(parent, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("bulk muting findings: %w", err)
	}
	return emitFormatted(op, flagSccFormat)
}

func runSccFindingsCreate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	if scope != "organizations" {
		return fmt.Errorf("findings.create is only supported for --organization scope")
	}
	sourceParent := findingsSourceParent(parent, flagSccFindingSource)
	body := &securitycenter.Finding{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Sources.Findings.Create(sourceParent, body).FindingId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating finding: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccFindingsGroup(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, flagSccFindingSource)
	req := &securitycenter.GroupFindingsRequest{
		GroupBy:         flagSccFindingGroupBy,
		Filter:          flagSccFilter,
		CompareDuration: flagSccFindingCompareDur,
		ReadTime:        flagSccFindingReadTime,
		PageSize:        flagSccPageSize,
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var resp any
	switch scope {
	case "organizations":
		resp, err = svc.Organizations.Sources.Findings.Group(sourceParent, req).Context(ctx).Do()
	case "folders":
		resp, err = svc.Folders.Sources.Findings.Group(sourceParent, req).Context(ctx).Do()
	case "projects":
		resp, err = svc.Projects.Sources.Findings.Group(sourceParent, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("grouping findings: %w", err)
	}
	return emitFormatted(resp, flagSccFormat)
}

func runSccFindingsList(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, flagSccFindingSource)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var all []*securitycenter.ListFindingsResult
	pageToken := ""
	for {
		var (
			page []*securitycenter.ListFindingsResult
			next string
		)
		switch scope {
		case "organizations":
			call := svc.Organizations.Sources.Findings.List(sourceParent).Context(ctx)
			if flagSccFilter != "" {
				call = call.Filter(flagSccFilter)
			}
			if flagSccOrderBy != "" {
				call = call.OrderBy(flagSccOrderBy)
			}
			if flagSccFindingFieldMask != "" {
				call = call.FieldMask(flagSccFindingFieldMask)
			}
			if flagSccFindingCompareDur != "" {
				call = call.CompareDuration(flagSccFindingCompareDur)
			}
			if flagSccFindingReadTime != "" {
				call = call.ReadTime(flagSccFindingReadTime)
			}
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing findings: %w", err)
			}
			page = resp.ListFindingsResults
			next = resp.NextPageToken
		case "folders":
			call := svc.Folders.Sources.Findings.List(sourceParent).Context(ctx)
			if flagSccFilter != "" {
				call = call.Filter(flagSccFilter)
			}
			if flagSccOrderBy != "" {
				call = call.OrderBy(flagSccOrderBy)
			}
			if flagSccFindingFieldMask != "" {
				call = call.FieldMask(flagSccFindingFieldMask)
			}
			if flagSccFindingCompareDur != "" {
				call = call.CompareDuration(flagSccFindingCompareDur)
			}
			if flagSccFindingReadTime != "" {
				call = call.ReadTime(flagSccFindingReadTime)
			}
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing findings: %w", err)
			}
			page = resp.ListFindingsResults
			next = resp.NextPageToken
		case "projects":
			call := svc.Projects.Sources.Findings.List(sourceParent).Context(ctx)
			if flagSccFilter != "" {
				call = call.Filter(flagSccFilter)
			}
			if flagSccOrderBy != "" {
				call = call.OrderBy(flagSccOrderBy)
			}
			if flagSccFindingFieldMask != "" {
				call = call.FieldMask(flagSccFindingFieldMask)
			}
			if flagSccFindingCompareDur != "" {
				call = call.CompareDuration(flagSccFindingCompareDur)
			}
			if flagSccFindingReadTime != "" {
				call = call.ReadTime(flagSccFindingReadTime)
			}
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing findings: %w", err)
			}
			page = resp.ListFindingsResults
			next = resp.NextPageToken
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if flagSccFormat != "" {
		return emitFormatted(all, flagSccFormat)
	}
	fmt.Printf("%-40s %-10s %s\n", "NAME", "STATE", "CATEGORY")
	for _, r := range all {
		if r.Finding == nil {
			continue
		}
		fmt.Printf("%-40s %-10s %s\n", path.Base(r.Finding.Name), r.Finding.State, r.Finding.Category)
	}
	return nil
}

func runSccFindingsListMarks(cmd *cobra.Command, args []string) error {
	// Marks are embedded on the Finding; we fetch the finding via the list
	// endpoint (filtering by name) and emit its SecurityMarks.
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, flagSccFindingSource)
	findingName := sccQualifyChild(sourceParent, "findings", args[0])
	filter := fmt.Sprintf("name=%q", findingName)
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.ListFindingsResult
	switch scope {
	case "organizations":
		resp, err := svc.Organizations.Sources.Findings.List(sourceParent).Filter(filter).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("fetching finding: %w", err)
		}
		results = resp.ListFindingsResults
	case "folders":
		resp, err := svc.Folders.Sources.Findings.List(sourceParent).Filter(filter).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("fetching finding: %w", err)
		}
		results = resp.ListFindingsResults
	case "projects":
		resp, err := svc.Projects.Sources.Findings.List(sourceParent).Filter(filter).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("fetching finding: %w", err)
		}
		results = resp.ListFindingsResults
	}
	if len(results) == 0 || results[0].Finding == nil {
		return fmt.Errorf("finding %q not found", args[0])
	}
	return emitFormatted(results[0].Finding.SecurityMarks, flagSccFormat)
}

func runSccFindingsSetMute(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, flagSccFindingSource)
	name := sccQualifyChild(sourceParent, "findings", args[0])
	req := &securitycenter.SetMuteRequest{Mute: flagSccFindingMuteState}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var got *securitycenter.Finding
	switch scope {
	case "organizations":
		got, err = svc.Organizations.Sources.Findings.SetMute(name, req).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Sources.Findings.SetMute(name, req).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.Sources.Findings.SetMute(name, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("setting mute: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccFindingsSetState(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, flagSccFindingSource)
	name := sccQualifyChild(sourceParent, "findings", args[0])
	req := &securitycenter.SetFindingStateRequest{
		State:     flagSccFindingState,
		StartTime: flagSccFindingStartTime,
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var got *securitycenter.Finding
	switch scope {
	case "organizations":
		got, err = svc.Organizations.Sources.Findings.SetState(name, req).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Sources.Findings.SetState(name, req).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.Sources.Findings.SetState(name, req).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("setting finding state: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccFindingsUpdate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, flagSccFindingSource)
	name := sccQualifyChild(sourceParent, "findings", args[0])
	body := &securitycenter.Finding{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var got *securitycenter.Finding
	switch scope {
	case "organizations":
		got, err = svc.Organizations.Sources.Findings.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Sources.Findings.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.Sources.Findings.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating finding: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccFindingsUpdateMarks(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, flagSccFindingSource)
	name := sccQualifyChild(sourceParent, "findings", args[0]) + "/securityMarks"
	body := &securitycenter.SecurityMarks{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var got *securitycenter.SecurityMarks
	switch scope {
	case "organizations":
		got, err = svc.Organizations.Sources.Findings.UpdateSecurityMarks(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Sources.Findings.UpdateSecurityMarks(name, body).UpdateMask(mask).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.Sources.Findings.UpdateSecurityMarks(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating finding security marks: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

// --- manage settings impl ---

func runSccManageSettingsDescribe(cmd *cobra.Command, args []string) error {
	name := "organizations/" + strings.TrimPrefix(flagSccOrg, "organizations/") + "/organizationSettings"
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.GetOrganizationSettings(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing organization settings: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccManageSettingsUpdate(cmd *cobra.Command, args []string) error {
	name := "organizations/" + strings.TrimPrefix(flagSccOrg, "organizations/") + "/organizationSettings"
	body := &securitycenter.OrganizationSettings{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.UpdateOrganizationSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating organization settings: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

// --- muteconfigs impl ---

func runSccMuteCreate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.GoogleCloudSecuritycenterV1MuteConfig{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var got *securitycenter.GoogleCloudSecuritycenterV1MuteConfig
	switch scope {
	case "organizations":
		got, err = svc.Organizations.MuteConfigs.Create(parent, body).MuteConfigId(args[0]).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.MuteConfigs.Create(parent, body).MuteConfigId(args[0]).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.MuteConfigs.Create(parent, body).MuteConfigId(args[0]).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("creating mute config: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccMuteDelete(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "muteConfigs", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	switch scope {
	case "organizations":
		_, err = svc.Organizations.MuteConfigs.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.MuteConfigs.Delete(name).Context(ctx).Do()
	case "projects":
		_, err = svc.Projects.MuteConfigs.Delete(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("deleting mute config: %w", err)
	}
	fmt.Printf("Deleted mute config [%s].\n", args[0])
	return nil
}

func runSccMuteDescribe(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "muteConfigs", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.GoogleCloudSecuritycenterV1MuteConfig
	switch scope {
	case "organizations":
		got, err = svc.Organizations.MuteConfigs.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.MuteConfigs.Get(name).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.MuteConfigs.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing mute config: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccMuteList(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var all []*securitycenter.GoogleCloudSecuritycenterV1MuteConfig
	pageToken := ""
	for {
		var (
			page []*securitycenter.GoogleCloudSecuritycenterV1MuteConfig
			next string
		)
		switch scope {
		case "organizations":
			call := svc.Organizations.MuteConfigs.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing mute configs: %w", err)
			}
			page = resp.MuteConfigs
			next = resp.NextPageToken
		case "folders":
			call := svc.Folders.MuteConfigs.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing mute configs: %w", err)
			}
			page = resp.MuteConfigs
			next = resp.NextPageToken
		case "projects":
			call := svc.Projects.MuteConfigs.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing mute configs: %w", err)
			}
			page = resp.MuteConfigs
			next = resp.NextPageToken
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if flagSccFormat != "" {
		return emitFormatted(all, flagSccFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, m := range all {
		fmt.Printf("%-40s %s\n", path.Base(m.Name), m.DisplayName)
	}
	return nil
}

func runSccMuteUpdate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "muteConfigs", args[0])
	body := &securitycenter.GoogleCloudSecuritycenterV1MuteConfig{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.GoogleCloudSecuritycenterV1MuteConfig
	switch scope {
	case "organizations":
		got, err = svc.Organizations.MuteConfigs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.MuteConfigs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.MuteConfigs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating mute config: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

// --- notifications impl ---

func runSccNotifyCreate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.NotificationConfig{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var got *securitycenter.NotificationConfig
	switch scope {
	case "organizations":
		got, err = svc.Organizations.NotificationConfigs.Create(parent, body).ConfigId(args[0]).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.NotificationConfigs.Create(parent, body).ConfigId(args[0]).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.NotificationConfigs.Create(parent, body).ConfigId(args[0]).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("creating notification config: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccNotifyDelete(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "notificationConfigs", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	switch scope {
	case "organizations":
		_, err = svc.Organizations.NotificationConfigs.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.NotificationConfigs.Delete(name).Context(ctx).Do()
	case "projects":
		_, err = svc.Projects.NotificationConfigs.Delete(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("deleting notification config: %w", err)
	}
	fmt.Printf("Deleted notification config [%s].\n", args[0])
	return nil
}

func runSccNotifyDescribe(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "notificationConfigs", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.NotificationConfig
	switch scope {
	case "organizations":
		got, err = svc.Organizations.NotificationConfigs.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.NotificationConfigs.Get(name).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.NotificationConfigs.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing notification config: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccNotifyList(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var all []*securitycenter.NotificationConfig
	pageToken := ""
	for {
		var (
			page []*securitycenter.NotificationConfig
			next string
		)
		switch scope {
		case "organizations":
			call := svc.Organizations.NotificationConfigs.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing notification configs: %w", err)
			}
			page = resp.NotificationConfigs
			next = resp.NextPageToken
		case "folders":
			call := svc.Folders.NotificationConfigs.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing notification configs: %w", err)
			}
			page = resp.NotificationConfigs
			next = resp.NextPageToken
		case "projects":
			call := svc.Projects.NotificationConfigs.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing notification configs: %w", err)
			}
			page = resp.NotificationConfigs
			next = resp.NextPageToken
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if flagSccFormat != "" {
		return emitFormatted(all, flagSccFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "PUBSUB_TOPIC")
	for _, n := range all {
		fmt.Printf("%-40s %s\n", path.Base(n.Name), n.PubsubTopic)
	}
	return nil
}

func runSccNotifyUpdate(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "notificationConfigs", args[0])
	body := &securitycenter.NotificationConfig{}
	if err := loadYAMLOrJSONInto(flagSccConfigFile, body); err != nil {
		return err
	}
	mask := flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(name)
	var got *securitycenter.NotificationConfig
	switch scope {
	case "organizations":
		got, err = svc.Organizations.NotificationConfigs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.NotificationConfigs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	case "projects":
		got, err = svc.Projects.NotificationConfigs.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating notification config: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

// --- operations impl ---

// SCC v1 Operations endpoints are only exposed under organizations/. Callers
// must supply --organization or pass a fully-qualified operation resource
// name.
func sccOperationsParent() (string, error) {
	if flagSccOrg != "" {
		return "organizations/" + strings.TrimPrefix(flagSccOrg, "organizations/"), nil
	}
	return "", fmt.Errorf("--organization is required for scc operations")
}

func runSccOpDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	if !strings.Contains(name, "/operations/") {
		parent, err := sccOperationsParent()
		if err != nil {
			return err
		}
		name = parent + "/operations/" + name
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccOpDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	if !strings.Contains(name, "/operations/") {
		parent, err := sccOperationsParent()
		if err != nil {
			return err
		}
		name = parent + "/operations/" + name
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runSccOpList(cmd *cobra.Command, args []string) error {
	parent, err := sccOperationsParent()
	if err != nil {
		return err
	}
	name := parent + "/operations"
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Operations.List(name).Context(ctx)
	if flagSccFilter != "" {
		call = call.Filter(flagSccFilter)
	}
	if flagSccPageSize > 0 {
		call = call.PageSize(flagSccPageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagSccFormat != "" {
		return emitFormatted(resp.Operations, flagSccFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- sources impl ---

func runSccSourcesDescribe(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "sources", args[0])
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	// Only Organizations.Sources.Get exists in v1.
	got, err := svc.Organizations.Sources.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing source: %w", err)
	}
	return emitFormatted(got, flagSccFormat)
}

func runSccSourcesList(cmd *cobra.Command, args []string) error {
	parent, err := sccResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var sources []*securitycenter.Source
	pageToken := ""
	for {
		var (
			page []*securitycenter.Source
			next string
		)
		switch scope {
		case "organizations":
			call := svc.Organizations.Sources.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing sources: %w", err)
			}
			page = resp.Sources
			next = resp.NextPageToken
		case "folders":
			call := svc.Folders.Sources.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing sources: %w", err)
			}
			page = resp.Sources
			next = resp.NextPageToken
		case "projects":
			call := svc.Projects.Sources.List(parent).Context(ctx)
			if flagSccPageSize > 0 {
				call = call.PageSize(flagSccPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing sources: %w", err)
			}
			page = resp.Sources
			next = resp.NextPageToken
		}
		sources = append(sources, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if flagSccFormat != "" {
		return emitFormatted(sources, flagSccFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, s := range sources {
		fmt.Printf("%-40s %s\n", path.Base(s.Name), s.DisplayName)
	}
	return nil
}
