package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	securitycenter "google.golang.org/api/securitycenter/v1"
)

// --- gcloud scc (#381, #805-#818) ---
//
// Real API-backed implementations for Security Command Center v1
// (securitycenter.googleapis.com) and Security Posture v1
// (securityposture.googleapis.com; see scc_posture.go).

// Common flags shared across all scc groups. Callers must set exactly one of
// --organization, --folder, or --project; if none are set the resolved active
// project is used.
var (
	flagSccLocation string
)

// --- scope resolution ---

// sccResolveParent returns the SCC scope parent as
// "organizations/{id}", "folders/{id}", or "projects/{id}". If no scope flag
// is set, it falls back to the active gcloud project.
func (g *sccCmds) sccResolveParent() (string, error) {
	set := 0
	for _, v := range []string{g.flagSccOrg, g.flagSccFolder, g.flagSccProject} {
		if v != "" {
			set++
		}
	}
	if set > 1 {
		return "", fmt.Errorf("--organization, --folder, and --project are mutually exclusive")
	}
	if g.flagSccOrg != "" {
		return "organizations/" + strings.TrimPrefix(g.flagSccOrg, "organizations/"), nil
	}
	if g.flagSccFolder != "" {
		return "folders/" + strings.TrimPrefix(g.flagSccFolder, "folders/"), nil
	}
	if g.flagSccProject != "" {
		return "projects/" + strings.TrimPrefix(g.flagSccProject, "projects/"), nil
	}
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

// sccResolveParentFromArg returns the SCC scope parent given an optional
// positional argument. If arg is non-empty it takes precedence over the
// --organization/--folder/--project flags (matching gcloud-python, which
// accepts the parent as a positional). A bare ID (e.g. "1234567890") is
// treated as an organization to preserve gcloud-python compatibility.
func (g *sccCmds) sccResolveParentFromArg(arg string) (string, error) {
	if arg == "" {
		return g.sccResolveParent()
	}
	if i := strings.Index(arg, "/"); i > 0 {
		kind := arg[:i]
		switch kind {
		case "organizations", "folders", "projects":
			return arg, nil
		default:
			return "", fmt.Errorf("invalid parent %q; expected organizations|folders|projects/{id}", arg)
		}
	}
	// Bare ID: default to organizations (matches gcloud-python's positional
	// parent handling when no resource type is specified).
	return "organizations/" + arg, nil
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
func (g *sccCmds) sccAddScopeFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagSccOrg, "organization", "", "Organization ID or fully-qualified organizations/{id}")
		c.Flags().StringVar(&g.flagSccFolder, "folder", "", "Folder ID or fully-qualified folders/{id}")
		c.Flags().StringVar(&g.flagSccProject, "project", "", "Project ID or fully-qualified projects/{id}")
	}
}

func (g *sccCmds) sccAddFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagSccFormat, "format", "", "Output format")
	}
}

func (g *sccCmds) sccAddFilterFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagSccFilter, "filter", "", "Server-side list filter")
	}
}

func (g *sccCmds) sccAddPageSizeFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().Int64Var(&g.flagSccPageSize, "page-size", 0, "Page size for list requests")
	}
}

// --- assets ---

// --- bqexports ---

// --- custom-modules ---

// --- findings ---

// sccLocationFlagHelp documents --location on every findings command that
// routes between the two API versions.
const sccLocationFlagHelp = "Location of the resource. Passing --location routes to the SCC V2 API via REST, " +
	"including --location=global; omitting it uses the SCC V1 API (deprecated server-side). " +
	"Organizations without data residency controls must use \"global\"."

// --- manage ---

// --- muteconfigs ---

// --- notifications ---

// --- operations ---

// --- sources ---

func (g *sccCmds) initScc() {
	// --- assets flags ---
	g.sccAddScopeFlags(g.sccAssetsDescribeCmd, g.sccAssetsGroupCmd, g.sccAssetsListCmd,
		g.sccAssetsRunDiscoveryCmd, g.sccAssetsUpdateMarksCmd)
	g.sccAddFormatFlag(g.sccAssetsDescribeCmd, g.sccAssetsGroupCmd, g.sccAssetsListCmd, g.sccAssetsUpdateMarksCmd)
	g.sccAddFilterFlag(g.sccAssetsGroupCmd, g.sccAssetsListCmd)
	g.sccAddPageSizeFlag(g.sccAssetsGroupCmd, g.sccAssetsListCmd)
	g.sccAssetsDescribeCmd.Flags().StringVar(&g.flagSccAssetsResourceName, "resource-name", "",
		"Full resource name of the asset to describe (e.g. //storage.googleapis.com/my-bucket)")
	g.sccAssetsDescribeCmd.Flags().StringVar(&g.flagSccAssetsAssetID, "asset", "", "Numeric SCC asset id")
	g.sccAssetsListCmd.Flags().StringVar(&g.flagSccAssetsCompareDuration, "compare-duration", "", "Compare-duration for state-change results")
	g.sccAssetsListCmd.Flags().StringVar(&g.flagSccAssetsReadTime, "read-time", "", "Read time (RFC3339)")
	g.sccAssetsListCmd.Flags().StringVar(&g.flagSccAssetsFieldMask, "field-mask", "", "Field mask for the response")
	g.sccAssetsListCmd.Flags().StringVar(&g.flagSccOrderBy, "order-by", "", "Server-side ordering expression")
	g.sccAssetsGroupCmd.Flags().StringVar(&g.flagSccAssetsGroupBy, "group-by", "", "Comma-separated list of fields to group by (required)")
	_ = g.sccAssetsGroupCmd.MarkFlagRequired("group-by")
	g.sccAssetsUpdateMarksCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the SecurityMarks body (required)")
	_ = g.sccAssetsUpdateMarksCmd.MarkFlagRequired("config-file")
	g.sccAssetsUpdateMarksCmd.Flags().StringVar(&g.flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.sccAssetsCmd.AddCommand(g.sccAssetsDescribeCmd, g.sccAssetsGroupCmd, g.sccAssetsListCmd,
		g.sccAssetsRunDiscoveryCmd, g.sccAssetsUpdateMarksCmd)
	g.sccCmd.AddCommand(g.sccAssetsCmd)

	// --- bqexports flags ---
	g.sccAddScopeFlags(g.sccBQExportsCreateCmd, g.sccBQExportsDeleteCmd, g.sccBQExportsDescribeCmd,
		g.sccBQExportsListCmd, g.sccBQExportsUpdateCmd)
	g.sccAddFormatFlag(g.sccBQExportsDescribeCmd, g.sccBQExportsListCmd)
	g.sccAddPageSizeFlag(g.sccBQExportsListCmd)
	g.sccBQExportsCreateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the BigQueryExport body (required)")
	_ = g.sccBQExportsCreateCmd.MarkFlagRequired("config-file")
	g.sccBQExportsUpdateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the BigQueryExport body (required)")
	_ = g.sccBQExportsUpdateCmd.MarkFlagRequired("config-file")
	g.sccBQExportsUpdateCmd.Flags().StringVar(&g.flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.sccBQExportsCmd.AddCommand(g.sccBQExportsCreateCmd, g.sccBQExportsDeleteCmd, g.sccBQExportsDescribeCmd,
		g.sccBQExportsListCmd, g.sccBQExportsUpdateCmd)
	g.sccCmd.AddCommand(g.sccBQExportsCmd)

	// --- custom-modules flags ---
	etdCmds := []*cobra.Command{g.sccEtdCreateCmd, g.sccEtdDeleteCmd, g.sccEtdDescribeCmd,
		g.sccEtdDescribeEffectiveCmd, g.sccEtdListCmd, g.sccEtdListDescendantCmd,
		g.sccEtdListEffectiveCmd, g.sccEtdUpdateCmd}
	shaCmds := []*cobra.Command{g.sccShaCreateCmd, g.sccShaDeleteCmd, g.sccShaDescribeCmd,
		g.sccShaDescribeEffectiveCmd, g.sccShaListCmd, g.sccShaListDescendantCmd,
		g.sccShaListEffectiveCmd, g.sccShaUpdateCmd}
	g.sccAddScopeFlags(etdCmds...)
	g.sccAddScopeFlags(shaCmds...)
	g.sccAddFormatFlag(g.sccEtdDescribeCmd, g.sccEtdDescribeEffectiveCmd, g.sccEtdListCmd,
		g.sccEtdListDescendantCmd, g.sccEtdListEffectiveCmd,
		g.sccShaDescribeCmd, g.sccShaDescribeEffectiveCmd, g.sccShaListCmd,
		g.sccShaListDescendantCmd, g.sccShaListEffectiveCmd)
	g.sccAddPageSizeFlag(g.sccEtdListCmd, g.sccEtdListDescendantCmd, g.sccEtdListEffectiveCmd,
		g.sccShaListCmd, g.sccShaListDescendantCmd, g.sccShaListEffectiveCmd)
	for _, c := range []*cobra.Command{g.sccEtdCreateCmd, g.sccEtdUpdateCmd,
		g.sccShaCreateCmd, g.sccShaUpdateCmd} {
		c.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the custom module body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{g.sccEtdUpdateCmd, g.sccShaUpdateCmd} {
		c.Flags().StringVar(&g.flagSccUpdateMask, "update-mask", "",
			"Comma-separated list of fields to update")
	}
	g.sccCustomModulesEtdCmd.AddCommand(etdCmds...)
	g.sccCustomModulesShaCmd.AddCommand(shaCmds...)
	g.sccCustomModulesCmd.AddCommand(g.sccCustomModulesEtdCmd, g.sccCustomModulesShaCmd)
	g.sccCmd.AddCommand(g.sccCustomModulesCmd)

	// --- findings flags ---
	findAll := []*cobra.Command{g.sccFindingsBulkMuteCmd, g.sccFindingsCreateCmd, g.sccFindingsExportBQCmd,
		g.sccFindingsGroupCmd, g.sccFindingsListCmd, g.sccFindingsListMarksCmd, g.sccFindingsSetMuteCmd,
		g.sccFindingsUpdateCmd, g.sccFindingsUpdateMarksCmd}
	g.sccAddScopeFlags(findAll...)
	g.sccAddFormatFlag(g.sccFindingsCreateCmd, g.sccFindingsExportBQCmd, g.sccFindingsGroupCmd,
		g.sccFindingsListCmd, g.sccFindingsListMarksCmd, g.sccFindingsSetMuteCmd,
		g.sccFindingsUpdateCmd, g.sccFindingsUpdateMarksCmd)
	for _, c := range []*cobra.Command{g.sccFindingsCreateCmd, g.sccFindingsExportBQCmd,
		g.sccFindingsGroupCmd, g.sccFindingsListCmd, g.sccFindingsListMarksCmd, g.sccFindingsSetMuteCmd,
		g.sccFindingsUpdateCmd, g.sccFindingsUpdateMarksCmd} {
		c.Flags().StringVar(&g.flagSccFindingSource, "source", "-",
			"Source id (or fully-qualified sources/{id}). Defaults to \"-\" (all sources)")
	}
	g.sccFindingsBulkMuteCmd.Flags().StringVar(&g.flagSccFindingBulkMuteFilter, "filter", "",
		"Filter expression identifying findings to mute (required)")
	_ = g.sccFindingsBulkMuteCmd.MarkFlagRequired("filter")
	g.sccFindingsBulkMuteCmd.Flags().StringVar(&g.flagSccFindingBulkMuteState, "mute-state", "muted",
		"Desired mute state of the findings (muted|undefined)")
	g.sccFindingsBulkMuteCmd.Flags().StringVar(&g.flagSccFindingLocation, "location", "global",
		sccLocationFlagHelp)
	g.sccFindingsCreateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Finding body (required)")
	_ = g.sccFindingsCreateCmd.MarkFlagRequired("config-file")
	g.sccFindingsExportBQCmd.Flags().StringVar(&g.flagSccFindingDataset, "dataset", "",
		"BigQuery dataset to export findings to, as projects/{project}/datasets/{dataset} (required)")
	_ = g.sccFindingsExportBQCmd.MarkFlagRequired("dataset")
	g.sccFindingsExportBQCmd.Flags().StringVar(&g.flagSccFindingLocation, "location", "global",
		sccLocationFlagHelp)
	g.sccFindingsListCmd.Flags().StringVar(&g.flagSccFilter, "filter", "", "Server-side list filter")
	g.sccFindingsListCmd.Flags().StringVar(&g.flagSccOrderBy, "order-by", "", "Server-side ordering expression")
	g.sccFindingsListCmd.Flags().StringVar(&g.flagSccFindingFieldMask, "field-mask", "", "Response field mask")
	g.sccFindingsListCmd.Flags().StringVar(&g.flagSccFindingCompareDur, "compare-duration", "", "Compare-duration for state-change results")
	g.sccFindingsListCmd.Flags().StringVar(&g.flagSccFindingReadTime, "read-time", "", "Read time (RFC3339)")
	g.sccFindingsListCmd.Flags().Int64Var(&g.flagSccPageSize, "page-size", 0,
		"Results per page. Unset lets the server choose: SCC findings.list defaults to 10, maximum 1000.")
	g.sccFindingsListCmd.Flags().StringVar(&g.flagSccFindingLocation, "location", "global",
		sccLocationFlagHelp)
	g.sccFindingsGroupCmd.Flags().StringVar(&g.flagSccFindingGroupBy, "group-by", "",
		"Comma-separated list of fields to group by (required)")
	_ = g.sccFindingsGroupCmd.MarkFlagRequired("group-by")
	g.sccFindingsGroupCmd.Flags().StringVar(&g.flagSccFilter, "filter", "", "Server-side filter")
	g.sccFindingsGroupCmd.Flags().StringVar(&g.flagSccFindingCompareDur, "compare-duration", "", "Compare-duration for state-change results")
	g.sccFindingsGroupCmd.Flags().StringVar(&g.flagSccFindingReadTime, "read-time", "", "Read time (RFC3339)")
	g.sccFindingsGroupCmd.Flags().Int64Var(&g.flagSccPageSize, "page-size", 0, "Page size for group requests")
	g.sccFindingsSetMuteCmd.Flags().StringVar(&g.flagSccFindingMuteState, "mute", "",
		"Mute state (MUTE_UNSPECIFIED|MUTED|UNMUTED|UNDEFINED) (required)")
	_ = g.sccFindingsSetMuteCmd.MarkFlagRequired("mute")
	g.sccFindingsUpdateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Finding body (required)")
	_ = g.sccFindingsUpdateCmd.MarkFlagRequired("config-file")
	g.sccFindingsUpdateCmd.Flags().StringVar(&g.flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.sccFindingsUpdateMarksCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the SecurityMarks body (required)")
	_ = g.sccFindingsUpdateMarksCmd.MarkFlagRequired("config-file")
	g.sccFindingsUpdateMarksCmd.Flags().StringVar(&g.flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.sccFindingsCmd.AddCommand(findAll...)
	g.sccCmd.AddCommand(g.sccFindingsCmd)

	// --- manage flags ---
	g.sccManageSettingsDescribeCmd.Flags().StringVar(&g.flagSccOrg, "organization", "",
		"Organization ID (required)")
	_ = g.sccManageSettingsDescribeCmd.MarkFlagRequired("organization")
	g.sccManageSettingsDescribeCmd.Flags().StringVar(&g.flagSccFormat, "format", "", "Output format")
	g.sccManageSettingsUpdateCmd.Flags().StringVar(&g.flagSccOrg, "organization", "",
		"Organization ID (required)")
	_ = g.sccManageSettingsUpdateCmd.MarkFlagRequired("organization")
	g.sccManageSettingsUpdateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the OrganizationSettings body (required)")
	_ = g.sccManageSettingsUpdateCmd.MarkFlagRequired("config-file")
	g.sccManageSettingsUpdateCmd.Flags().StringVar(&g.flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.sccManageSettingsCmd.AddCommand(g.sccManageSettingsDescribeCmd, g.sccManageSettingsUpdateCmd)
	g.sccManageCmd.AddCommand(g.sccManageSettingsCmd)
	g.sccCmd.AddCommand(g.sccManageCmd)

	// --- muteconfigs flags ---
	muteAll := []*cobra.Command{g.sccMuteCreateCmd, g.sccMuteDeleteCmd, g.sccMuteDescribeCmd,
		g.sccMuteListCmd, g.sccMuteUpdateCmd}
	g.sccAddScopeFlags(muteAll...)
	g.sccAddFormatFlag(g.sccMuteDescribeCmd, g.sccMuteListCmd)
	g.sccAddPageSizeFlag(g.sccMuteListCmd)
	g.sccMuteCreateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the MuteConfig body (required)")
	_ = g.sccMuteCreateCmd.MarkFlagRequired("config-file")
	g.sccMuteUpdateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the MuteConfig body (required)")
	_ = g.sccMuteUpdateCmd.MarkFlagRequired("config-file")
	g.sccMuteUpdateCmd.Flags().StringVar(&g.flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.sccMuteConfigsCmd.AddCommand(muteAll...)
	g.sccCmd.AddCommand(g.sccMuteConfigsCmd)

	// --- notifications flags ---
	notifyAll := []*cobra.Command{g.sccNotifyCreateCmd, g.sccNotifyDeleteCmd, g.sccNotifyDescribeCmd,
		g.sccNotifyListCmd, g.sccNotifyUpdateCmd}
	g.sccAddScopeFlags(notifyAll...)
	g.sccAddFormatFlag(g.sccNotifyDescribeCmd, g.sccNotifyListCmd)
	g.sccAddPageSizeFlag(g.sccNotifyListCmd)
	g.sccNotifyCreateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the NotificationConfig body (required)")
	_ = g.sccNotifyCreateCmd.MarkFlagRequired("config-file")
	g.sccNotifyUpdateCmd.Flags().StringVar(&g.flagSccConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the NotificationConfig body (required)")
	_ = g.sccNotifyUpdateCmd.MarkFlagRequired("config-file")
	g.sccNotifyUpdateCmd.Flags().StringVar(&g.flagSccUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	g.sccNotificationsCmd.AddCommand(notifyAll...)
	g.sccCmd.AddCommand(g.sccNotificationsCmd)

	// --- operations flags (organization-scoped only in v1) ---
	for _, c := range []*cobra.Command{g.sccOpDescribeCmd, g.sccOpDeleteCmd, g.sccOpListCmd} {
		c.Flags().StringVar(&g.flagSccOrg, "organization", "",
			"Organization ID (required unless the operation name is fully-qualified)")
	}
	g.sccAddFormatFlag(g.sccOpDescribeCmd, g.sccOpListCmd)
	g.sccAddFilterFlag(g.sccOpListCmd)
	g.sccAddPageSizeFlag(g.sccOpListCmd)
	g.sccOperationsCmd.AddCommand(g.sccOpDescribeCmd, g.sccOpDeleteCmd, g.sccOpListCmd)
	g.sccCmd.AddCommand(g.sccOperationsCmd)

	// --- sources flags ---
	g.sccAddScopeFlags(g.sccSourcesDescribeCmd, g.sccSourcesListCmd)
	g.sccAddFormatFlag(g.sccSourcesDescribeCmd, g.sccSourcesListCmd)
	g.sccAddPageSizeFlag(g.sccSourcesListCmd)
	g.sccSourcesCmd.AddCommand(g.sccSourcesDescribeCmd, g.sccSourcesListCmd)
	g.sccCmd.AddCommand(g.sccSourcesCmd)

	// Register scc_posture.go subgroups (see init in that file).

}

// --- helpers for scope-dispatched API calls ---

// sccClient returns an initialised securitycenter service.
func (g *sccCmds) sccClient(ctx context.Context) (*securitycenter.Service, error) {
	return g.inv.clients().SecurityCenter(ctx)
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

func (g *sccCmds) runSccAssetsDescribe(cmd *cobra.Command, args []string) error {
	// Describe is implemented client-side by listing with a filter that pins
	// the specific asset. This mirrors gcloud's DescribeAssetReqHook behavior.
	parent := args[0]
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	filter := g.flagSccFilter
	switch {
	case g.flagSccAssetsResourceName != "":
		filter = fmt.Sprintf("securityCenterProperties.resourceName=%q", g.flagSccAssetsResourceName)
	case g.flagSccAssetsAssetID != "":
		filter = fmt.Sprintf("name : %q", "/assets/"+g.flagSccAssetsAssetID)
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
	return emitFormattedTo(g.inv.stdout, assets[0], g.flagSccFormat)
}

func (g *sccCmds) runSccAssetsGroup(cmd *cobra.Command, args []string) error {
	parent := args[0]
	req := &securitycenter.GroupAssetsRequest{
		GroupBy:  g.flagSccAssetsGroupBy,
		Filter:   g.flagSccFilter,
		PageSize: g.flagSccPageSize,
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, resp, g.flagSccFormat)
}

func (g *sccCmds) runSccAssetsList(cmd *cobra.Command, args []string) error {
	parent := args[0]
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
			if g.flagSccFilter != "" {
				call = call.Filter(g.flagSccFilter)
			}
			if g.flagSccOrderBy != "" {
				call = call.OrderBy(g.flagSccOrderBy)
			}
			if g.flagSccAssetsCompareDuration != "" {
				call = call.CompareDuration(g.flagSccAssetsCompareDuration)
			}
			if g.flagSccAssetsReadTime != "" {
				call = call.ReadTime(g.flagSccAssetsReadTime)
			}
			if g.flagSccAssetsFieldMask != "" {
				call = call.FieldMask(g.flagSccAssetsFieldMask)
			}
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccFilter != "" {
				call = call.Filter(g.flagSccFilter)
			}
			if g.flagSccOrderBy != "" {
				call = call.OrderBy(g.flagSccOrderBy)
			}
			if g.flagSccAssetsCompareDuration != "" {
				call = call.CompareDuration(g.flagSccAssetsCompareDuration)
			}
			if g.flagSccAssetsReadTime != "" {
				call = call.ReadTime(g.flagSccAssetsReadTime)
			}
			if g.flagSccAssetsFieldMask != "" {
				call = call.FieldMask(g.flagSccAssetsFieldMask)
			}
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccFilter != "" {
				call = call.Filter(g.flagSccFilter)
			}
			if g.flagSccOrderBy != "" {
				call = call.OrderBy(g.flagSccOrderBy)
			}
			if g.flagSccAssetsCompareDuration != "" {
				call = call.CompareDuration(g.flagSccAssetsCompareDuration)
			}
			if g.flagSccAssetsReadTime != "" {
				call = call.ReadTime(g.flagSccAssetsReadTime)
			}
			if g.flagSccAssetsFieldMask != "" {
				call = call.FieldMask(g.flagSccAssetsFieldMask)
			}
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagSccFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-60s %s\n", "NAME", "RESOURCE_TYPE")
	for _, a := range all {
		if a.Asset != nil {
			name := a.Asset.Name
			var rtype string
			if a.Asset.SecurityCenterProperties != nil {
				rtype = a.Asset.SecurityCenterProperties.ResourceType
			}
			fmt.Fprintf(g.inv.stdout, "%-60s %s\n", path.Base(name), rtype)
		}
	}
	return nil
}

func (g *sccCmds) runSccAssetsRunDiscovery(cmd *cobra.Command, args []string) error {
	org := args[0]
	if !strings.HasPrefix(org, "organizations/") {
		org = "organizations/" + org
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Assets.RunDiscovery(org, &securitycenter.RunAssetDiscoveryRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running discovery: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagSccFormat)
}

func (g *sccCmds) runSccAssetsUpdateMarks(cmd *cobra.Command, args []string) error {
	name := args[0]
	// Expect assetName in the form <parent>/assets/{id}/securityMarks or the
	// full asset resource name; append /securityMarks if missing.
	if !strings.HasSuffix(name, "/securityMarks") {
		name = strings.TrimSuffix(name, "/") + "/securityMarks"
	}
	body := &securitycenter.SecurityMarks{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

// --- bqexports impl ---

func (g *sccCmds) runSccBQExportCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.GoogleCloudSecuritycenterV1BigQueryExport{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccBQExportDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "bigQueryExports", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Deleted BigQuery export [%s].\n", args[0])
	return nil
}

func (g *sccCmds) runSccBQExportDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "bigQueryExports", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccBQExportList(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var all []*securitycenter.GoogleCloudSecuritycenterV1BigQueryExport
	pageToken := ""
	for {
		var (
			page []*securitycenter.GoogleCloudSecuritycenterV1BigQueryExport
			next string
			cerr error
		)
		switch scope {
		case "organizations":
			call := svc.Organizations.BigQueryExports.List(parent).Context(ctx)
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagSccFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DATASET")
	for _, e := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(e.Name), e.Dataset)
	}
	return nil
}

func (g *sccCmds) runSccBQExportUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "bigQueryExports", args[0])
	body := &securitycenter.GoogleCloudSecuritycenterV1BigQueryExport{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

// --- custom-modules etd impl ---

func etdParent(parent string) string {
	return parent + "/eventThreatDetectionSettings"
}

func (g *sccCmds) runSccEtdCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.EventThreatDetectionCustomModule{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	if body.DisplayName == "" {
		body.DisplayName = args[0]
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccEtdDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(etdParent(parent), "customModules", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Deleted ETD custom module [%s].\n", args[0])
	return nil
}

func (g *sccCmds) runSccEtdDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(etdParent(parent), "customModules", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccEtdDescribeEffective(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(etdParent(parent), "effectiveCustomModules", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccEtdList(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	settings := etdParent(parent)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.EventThreatDetectionCustomModule
	switch scope {
	case "organizations":
		call := svc.Organizations.EventThreatDetectionSettings.CustomModules.List(settings).Context(ctx)
		if g.flagSccPageSize > 0 {
			call = call.PageSize(g.flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	case "folders":
		call := svc.Folders.EventThreatDetectionSettings.CustomModules.List(settings).Context(ctx)
		if g.flagSccPageSize > 0 {
			call = call.PageSize(g.flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	case "projects":
		call := svc.Projects.EventThreatDetectionSettings.CustomModules.List(settings).Context(ctx)
		if g.flagSccPageSize > 0 {
			call = call.PageSize(g.flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ETD custom modules: %w", err)
		}
		results = resp.EventThreatDetectionCustomModules
	}
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, results, g.flagSccFormat)
	}
	g.printModulesEtd(results)
	return nil
}

func (g *sccCmds) runSccEtdListDescendant(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	settings := etdParent(parent)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, results, g.flagSccFormat)
	}
	g.printModulesEtd(results)
	return nil
}

func (g *sccCmds) runSccEtdListEffective(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	settings := etdParent(parent)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, results, g.flagSccFormat)
}

func (g *sccCmds) runSccEtdUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(etdParent(parent), "customModules", args[0])
	body := &securitycenter.EventThreatDetectionCustomModule{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) printModulesEtd(modules []*securitycenter.EventThreatDetectionCustomModule) {
	printTableHeadingTo(g.inv.stdout, len(modules), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, m := range modules {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(m.Name), m.DisplayName)
	}
}

// --- custom-modules sha impl ---

func shaParent(parent string) string {
	return parent + "/securityHealthAnalyticsSettings"
}

func (g *sccCmds) runSccShaCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	if body.DisplayName == "" {
		body.DisplayName = args[0]
	}
	settings := shaParent(parent)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccShaDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(shaParent(parent), "customModules", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Deleted SHA custom module [%s].\n", args[0])
	return nil
}

func (g *sccCmds) runSccShaDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(shaParent(parent), "customModules", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccShaDescribeEffective(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(shaParent(parent), "effectiveCustomModules", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccShaList(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	settings := shaParent(parent)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	var results []*securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule
	switch scope {
	case "organizations":
		call := svc.Organizations.SecurityHealthAnalyticsSettings.CustomModules.List(settings).Context(ctx)
		if g.flagSccPageSize > 0 {
			call = call.PageSize(g.flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	case "folders":
		call := svc.Folders.SecurityHealthAnalyticsSettings.CustomModules.List(settings).Context(ctx)
		if g.flagSccPageSize > 0 {
			call = call.PageSize(g.flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	case "projects":
		call := svc.Projects.SecurityHealthAnalyticsSettings.CustomModules.List(settings).Context(ctx)
		if g.flagSccPageSize > 0 {
			call = call.PageSize(g.flagSccPageSize)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing SHA custom modules: %w", err)
		}
		results = resp.SecurityHealthAnalyticsCustomModules
	}
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, results, g.flagSccFormat)
	}
	g.printModulesSha(results)
	return nil
}

func (g *sccCmds) runSccShaListDescendant(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	settings := shaParent(parent)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, results, g.flagSccFormat)
	}
	g.printModulesSha(results)
	return nil
}

func (g *sccCmds) runSccShaListEffective(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	settings := shaParent(parent)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, results, g.flagSccFormat)
}

func (g *sccCmds) runSccShaUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(shaParent(parent), "customModules", args[0])
	body := &securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) printModulesSha(modules []*securitycenter.GoogleCloudSecuritycenterV1SecurityHealthAnalyticsCustomModule) {
	printTableHeadingTo(g.inv.stdout, len(modules), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, m := range modules {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(m.Name), m.DisplayName)
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

func (g *sccCmds) runSccFindingsBulkMute(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	// Same V1/V2 routing as findings list: an explicit --location selects V2.
	if sccLocationSpecified(cmd) {
		return g.sccV2BulkMuteFindings(parent)
	}
	muteState, err := sccMuteStateEnum(g.flagSccFindingBulkMuteState)
	if err != nil {
		return err
	}
	req := &securitycenter.BulkMuteFindingsRequest{
		Filter:    g.flagSccFindingBulkMuteFilter,
		MuteState: muteState,
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, op, g.flagSccFormat)
}

func (g *sccCmds) runSccFindingsCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	scope, _ := splitScope(parent)
	if scope != "organizations" {
		return fmt.Errorf("findings.create is only supported for --organization scope")
	}
	sourceParent := findingsSourceParent(parent, g.flagSccFindingSource)
	body := &securitycenter.Finding{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Sources.Findings.Create(sourceParent, body).FindingId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating finding: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

// runSccFindingsExportBQ exports findings to BigQuery. gcloud-python pins this
// command to the V2 API regardless of --location (surface/scc/findings/
// export_to_bigquery.py sets version = "v2"), so there is no V1 branch here.
func (g *sccCmds) runSccFindingsExportBQ(cmd *cobra.Command, args []string) error {
	pos := ""
	if len(args) == 1 {
		pos = args[0]
	}
	parent, err := g.sccResolveParentFromArg(pos)
	if err != nil {
		return err
	}
	return g.sccV2ExportFindings(parent)
}

func (g *sccCmds) runSccFindingsGroup(cmd *cobra.Command, args []string) error {
	pos := ""
	if len(args) == 1 {
		pos = args[0]
	}
	parent, err := g.sccResolveParentFromArg(pos)
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, g.flagSccFindingSource)
	req := &securitycenter.GroupFindingsRequest{
		GroupBy:         g.flagSccFindingGroupBy,
		Filter:          g.flagSccFilter,
		CompareDuration: g.flagSccFindingCompareDur,
		ReadTime:        g.flagSccFindingReadTime,
		PageSize:        g.flagSccPageSize,
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, resp, g.flagSccFormat)
}

// sccLocationSpecified reports whether the caller passed --location, which is
// what selects the SCC V2 API. cobra always supplies the command; a nil one
// means no command line was parsed, so no location was given.
func sccLocationSpecified(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	f := cmd.Flags().Lookup("location")
	return f != nil && f.Changed
}

func (g *sccCmds) runSccFindingsList(cmd *cobra.Command, args []string) error {
	pos := ""
	if len(args) == 1 {
		pos = args[0]
	}
	parent, err := g.sccResolveParentFromArg(pos)
	if err != nil {
		return err
	}
	// --location routes to the SCC V2 API. gcloud-python routes on whether
	// --location was *specified*, not on the value it was given: any explicit
	// location selects V2, "global" included (IsLocationSpecified /
	// GetVersionFromArguments in command_lib/scc/util.py). Routing on the
	// value instead made --location=global unreachable on V2, which is the
	// only location an organization without data residency controls may use:
	// regional endpoints there fail with DRZ_LOCATION_MISMATCH.
	//
	// The V1 findings API is deprecated (returns INVALID_ARGUMENT
	// server-side), so most users need V2. Since google.golang.org/api@v0.289.0
	// doesn't ship a generated securitycenter/v2 client, V2 is implemented via
	// direct REST calls (see sccV2ListFindings).
	if sccLocationSpecified(cmd) {
		return g.sccV2ListFindings(parent)
	}
	sourceParent := findingsSourceParent(parent, g.flagSccFindingSource)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
			if g.flagSccFilter != "" {
				call = call.Filter(g.flagSccFilter)
			}
			if g.flagSccOrderBy != "" {
				call = call.OrderBy(g.flagSccOrderBy)
			}
			if g.flagSccFindingFieldMask != "" {
				call = call.FieldMask(g.flagSccFindingFieldMask)
			}
			if g.flagSccFindingCompareDur != "" {
				call = call.CompareDuration(g.flagSccFindingCompareDur)
			}
			if g.flagSccFindingReadTime != "" {
				call = call.ReadTime(g.flagSccFindingReadTime)
			}
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccFilter != "" {
				call = call.Filter(g.flagSccFilter)
			}
			if g.flagSccOrderBy != "" {
				call = call.OrderBy(g.flagSccOrderBy)
			}
			if g.flagSccFindingFieldMask != "" {
				call = call.FieldMask(g.flagSccFindingFieldMask)
			}
			if g.flagSccFindingCompareDur != "" {
				call = call.CompareDuration(g.flagSccFindingCompareDur)
			}
			if g.flagSccFindingReadTime != "" {
				call = call.ReadTime(g.flagSccFindingReadTime)
			}
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccFilter != "" {
				call = call.Filter(g.flagSccFilter)
			}
			if g.flagSccOrderBy != "" {
				call = call.OrderBy(g.flagSccOrderBy)
			}
			if g.flagSccFindingFieldMask != "" {
				call = call.FieldMask(g.flagSccFindingFieldMask)
			}
			if g.flagSccFindingCompareDur != "" {
				call = call.CompareDuration(g.flagSccFindingCompareDur)
			}
			if g.flagSccFindingReadTime != "" {
				call = call.ReadTime(g.flagSccFindingReadTime)
			}
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagSccFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %-10s %s\n", "NAME", "STATE", "CATEGORY")
	for _, r := range all {
		if r.Finding == nil {
			continue
		}
		fmt.Fprintf(g.inv.stdout, "%-40s %-10s %s\n", path.Base(r.Finding.Name), r.Finding.State, r.Finding.Category)
	}
	return nil
}

func (g *sccCmds) runSccFindingsListMarks(cmd *cobra.Command, args []string) error {
	// Marks are embedded on the Finding; we fetch the finding via the list
	// endpoint (filtering by name) and emit its SecurityMarks.
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, g.flagSccFindingSource)
	findingName := sccQualifyChild(sourceParent, "findings", args[0])
	filter := fmt.Sprintf("name=%q", findingName)
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, results[0].Finding.SecurityMarks, g.flagSccFormat)
}

func (g *sccCmds) runSccFindingsSetMute(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, g.flagSccFindingSource)
	name := sccQualifyChild(sourceParent, "findings", args[0])
	req := &securitycenter.SetMuteRequest{Mute: g.flagSccFindingMuteState}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccFindingsUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, g.flagSccFindingSource)
	name := sccQualifyChild(sourceParent, "findings", args[0])
	body := &securitycenter.Finding{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccFindingsUpdateMarks(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	sourceParent := findingsSourceParent(parent, g.flagSccFindingSource)
	name := sccQualifyChild(sourceParent, "findings", args[0]) + "/securityMarks"
	body := &securitycenter.SecurityMarks{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

// --- manage settings impl ---

func (g *sccCmds) runSccManageSettingsDescribe(cmd *cobra.Command, args []string) error {
	name := "organizations/" + strings.TrimPrefix(g.flagSccOrg, "organizations/") + "/organizationSettings"
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.GetOrganizationSettings(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing organization settings: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccManageSettingsUpdate(cmd *cobra.Command, args []string) error {
	name := "organizations/" + strings.TrimPrefix(g.flagSccOrg, "organizations/") + "/organizationSettings"
	body := &securitycenter.OrganizationSettings{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.UpdateOrganizationSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating organization settings: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

// --- muteconfigs impl ---

func (g *sccCmds) runSccMuteCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.GoogleCloudSecuritycenterV1MuteConfig{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccMuteDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "muteConfigs", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Deleted mute config [%s].\n", args[0])
	return nil
}

func (g *sccCmds) runSccMuteDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "muteConfigs", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccMuteList(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagSccFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, m := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(m.Name), m.DisplayName)
	}
	return nil
}

func (g *sccCmds) runSccMuteUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "muteConfigs", args[0])
	body := &securitycenter.GoogleCloudSecuritycenterV1MuteConfig{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

// --- notifications impl ---

func (g *sccCmds) runSccNotifyCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	body := &securitycenter.NotificationConfig{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccNotifyDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "notificationConfigs", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Deleted notification config [%s].\n", args[0])
	return nil
}

func (g *sccCmds) runSccNotifyDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "notificationConfigs", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccNotifyList(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagSccFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "PUBSUB_TOPIC")
	for _, n := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(n.Name), n.PubsubTopic)
	}
	return nil
}

func (g *sccCmds) runSccNotifyUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "notificationConfigs", args[0])
	body := &securitycenter.NotificationConfig{}
	if err := loadYAMLOrJSONInto(g.flagSccConfigFile, body); err != nil {
		return err
	}
	mask := g.flagSccUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

// --- operations impl ---

// SCC v1 Operations endpoints are only exposed under organizations/. Callers
// must supply --organization or pass a fully-qualified operation resource
// name.
func (g *sccCmds) sccOperationsParent() (string, error) {
	if g.flagSccOrg != "" {
		return "organizations/" + strings.TrimPrefix(g.flagSccOrg, "organizations/"), nil
	}
	return "", fmt.Errorf("--organization is required for scc operations")
}

func (g *sccCmds) runSccOpDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	if !strings.Contains(name, "/operations/") {
		parent, err := g.sccOperationsParent()
		if err != nil {
			return err
		}
		name = parent + "/operations/" + name
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccOpDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	if !strings.Contains(name, "/operations/") {
		parent, err := g.sccOperationsParent()
		if err != nil {
			return err
		}
		name = parent + "/operations/" + name
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (g *sccCmds) runSccOpList(cmd *cobra.Command, args []string) error {
	parent, err := g.sccOperationsParent()
	if err != nil {
		return err
	}
	name := parent + "/operations"
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Operations.List(name).Context(ctx)
	if g.flagSccFilter != "" {
		call = call.Filter(g.flagSccFilter)
	}
	if g.flagSccPageSize > 0 {
		call = call.PageSize(g.flagSccPageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Operations, g.flagSccFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(g.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- sources impl ---

func (g *sccCmds) runSccSourcesDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	name := sccQualifyChild(parent, "sources", args[0])
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
	if err != nil {
		return err
	}
	// Only Organizations.Sources.Get exists in v1.
	got, err := svc.Organizations.Sources.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing source: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSccFormat)
}

func (g *sccCmds) runSccSourcesList(cmd *cobra.Command, args []string) error {
	parent, err := g.sccResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.sccClient(ctx)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
			if g.flagSccPageSize > 0 {
				call = call.PageSize(g.flagSccPageSize)
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
	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, sources, g.flagSccFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(sources), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, s := range sources {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(s.Name), s.DisplayName)
	}
	return nil
}

// sccCmds holds the commands and flag values of one invocation's "scc"
// command group, so invocations don't share state (#2070).
type sccCmds struct {
	inv *invocation

	sccCmd            *cobra.Command
	flagSccOrg        string
	flagSccFolder     string
	flagSccProject    string
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
	flagSccFindingBulkMuteFilter string
	flagSccFindingBulkMuteState  string
	flagSccFindingDataset        string
	flagSccFindingGroupBy        string
	flagSccFindingCompareDur     string
	flagSccFindingReadTime       string
	flagSccFindingFieldMask      string
	flagSccFindingLocation       string
	sccAssetsCmd                 *cobra.Command
	sccAssetsDescribeCmd         *cobra.Command
	sccAssetsGroupCmd            *cobra.Command
	sccAssetsListCmd             *cobra.Command
	sccAssetsRunDiscoveryCmd     *cobra.Command
	sccAssetsUpdateMarksCmd      *cobra.Command
	sccBQExportsCmd              *cobra.Command
	sccBQExportsCreateCmd        *cobra.Command
	sccBQExportsDeleteCmd        *cobra.Command
	sccBQExportsDescribeCmd      *cobra.Command
	sccBQExportsListCmd          *cobra.Command
	sccBQExportsUpdateCmd        *cobra.Command
	sccCustomModulesCmd          *cobra.Command
	// custom-modules etd
	sccCustomModulesEtdCmd     *cobra.Command
	sccEtdCreateCmd            *cobra.Command
	sccEtdDeleteCmd            *cobra.Command
	sccEtdDescribeCmd          *cobra.Command
	sccEtdDescribeEffectiveCmd *cobra.Command
	sccEtdListCmd              *cobra.Command
	sccEtdListDescendantCmd    *cobra.Command
	sccEtdListEffectiveCmd     *cobra.Command
	sccEtdUpdateCmd            *cobra.Command
	// custom-modules sha
	sccCustomModulesShaCmd       *cobra.Command
	sccShaCreateCmd              *cobra.Command
	sccShaDeleteCmd              *cobra.Command
	sccShaDescribeCmd            *cobra.Command
	sccShaDescribeEffectiveCmd   *cobra.Command
	sccShaListCmd                *cobra.Command
	sccShaListDescendantCmd      *cobra.Command
	sccShaListEffectiveCmd       *cobra.Command
	sccShaUpdateCmd              *cobra.Command
	sccFindingsCmd               *cobra.Command
	sccFindingsBulkMuteCmd       *cobra.Command
	sccFindingsCreateCmd         *cobra.Command
	sccFindingsExportBQCmd       *cobra.Command
	sccFindingsGroupCmd          *cobra.Command
	sccFindingsListCmd           *cobra.Command
	sccFindingsListMarksCmd      *cobra.Command
	sccFindingsSetMuteCmd        *cobra.Command
	sccFindingsUpdateCmd         *cobra.Command
	sccFindingsUpdateMarksCmd    *cobra.Command
	sccManageCmd                 *cobra.Command
	sccManageSettingsCmd         *cobra.Command
	sccManageSettingsDescribeCmd *cobra.Command
	sccManageSettingsUpdateCmd   *cobra.Command
	sccMuteConfigsCmd            *cobra.Command
	sccMuteCreateCmd             *cobra.Command
	sccMuteDeleteCmd             *cobra.Command
	sccMuteDescribeCmd           *cobra.Command
	sccMuteListCmd               *cobra.Command
	sccMuteUpdateCmd             *cobra.Command
	sccNotificationsCmd          *cobra.Command
	sccNotifyCreateCmd           *cobra.Command
	sccNotifyDeleteCmd           *cobra.Command
	sccNotifyDescribeCmd         *cobra.Command
	sccNotifyListCmd             *cobra.Command
	sccNotifyUpdateCmd           *cobra.Command
	sccOperationsCmd             *cobra.Command
	sccOpDescribeCmd             *cobra.Command
	sccOpDeleteCmd               *cobra.Command
	sccOpListCmd                 *cobra.Command
	sccSourcesCmd                *cobra.Command
	sccSourcesDescribeCmd        *cobra.Command
	sccSourcesListCmd            *cobra.Command
	sccManageRest                *restClient
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
	flagSccManageValidateOnly    bool
	sccManageServicesCmd         *cobra.Command
	sccManageServicesDescribeCmd *cobra.Command
	sccManageServicesListCmd     *cobra.Command
	sccManageServicesUpdateCmd   *cobra.Command
	flagPostureOrg               string
	flagPostureLocation          string
	flagPostureFormat            string
	flagPostureFilter            string
	flagPostureConfigFile        string
	flagPostureUpdateMask        string
	flagPostureAsync             bool
	flagPostureRevision          string
	flagPostureExtractWL         string
	flagPosturePageSize          int64
	// iac-validation-reports create
	flagIaCConfigFile        string
	sccPosturesCmd           *cobra.Command
	postureCreateCmd         *cobra.Command
	postureDeleteCmd         *cobra.Command
	postureDescribeCmd       *cobra.Command
	postureExtractCmd        *cobra.Command
	postureListCmd           *cobra.Command
	postureListRevisionsCmd  *cobra.Command
	postureUpdateCmd         *cobra.Command
	sccPostureDeploymentsCmd *cobra.Command
	pdCreateCmd              *cobra.Command
	pdDeleteCmd              *cobra.Command
	pdDescribeCmd            *cobra.Command
	pdListCmd                *cobra.Command
	pdUpdateCmd              *cobra.Command
	sccPostureTemplatesCmd   *cobra.Command
	ptDescribeCmd            *cobra.Command
	ptListCmd                *cobra.Command
	sccPostureOperationsCmd  *cobra.Command
	poCancelCmd              *cobra.Command
	poDeleteCmd              *cobra.Command
	poDescribeCmd            *cobra.Command
	poListCmd                *cobra.Command
	sccIaCReportsCmd         *cobra.Command
	iacCreateCmd             *cobra.Command
	iacDescribeCmd           *cobra.Command
	iacListCmd               *cobra.Command
	// sccV2Rest is the REST client for the SCC V2 API. google.golang.org/api
	// (through v0.293.0 at the time of writing) does not ship a generated
	// securitycenter/v2 client, so V2 endpoints are called directly. Requests
	// automatically pick up the caller's ADC / --account credentials via the
	// shared restClient.
	sccV2Rest *restClient
}

// newSccCmd returns the "scc" command group for inv.
func newSccCmd(inv *invocation) *cobra.Command {
	return newSccCmds(inv).sccCmd
}

// newSccCmds builds the "scc" command group for inv and returns its state.
func newSccCmds(inv *invocation) *sccCmds {
	g := &sccCmds{inv: inv}
	g.sccCmd = &cobra.Command{Use: "scc", Short: "Manage Security Command Center"}
	g.sccAssetsCmd = &cobra.Command{Use: "assets", Short: "Manage Security Command Center assets"}
	g.sccAssetsDescribeCmd = &cobra.Command{
		Use: "describe PARENT", Short: "Describe an asset",
		Args: cobra.ExactArgs(1), RunE: g.runSccAssetsDescribe,
	}
	g.sccAssetsGroupCmd = &cobra.Command{
		Use: "group PARENT", Short: "Group assets by fields",
		Args: cobra.ExactArgs(1), RunE: g.runSccAssetsGroup,
	}
	g.sccAssetsListCmd = &cobra.Command{
		Use: "list PARENT", Short: "List Security Command Center assets",
		Args: cobra.ExactArgs(1), RunE: g.runSccAssetsList,
	}
	g.sccAssetsRunDiscoveryCmd = &cobra.Command{
		Use: "run-discovery ORGANIZATION", Short: "Run asset discovery for an organization",
		Args: cobra.ExactArgs(1), RunE: g.runSccAssetsRunDiscovery,
	}
	g.sccAssetsUpdateMarksCmd = &cobra.Command{
		Use: "update-security-marks ASSET_NAME", Short: "Update security marks on an asset",
		Args: cobra.ExactArgs(1), RunE: g.runSccAssetsUpdateMarks,
	}
	g.sccBQExportsCmd = &cobra.Command{Use: "bqexports", Short: "Manage BigQuery exports"}
	g.sccBQExportsCreateCmd = &cobra.Command{
		Use: "create EXPORT", Short: "Create a BigQuery export from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccBQExportCreate,
	}
	g.sccBQExportsDeleteCmd = &cobra.Command{
		Use: "delete EXPORT", Short: "Delete a BigQuery export",
		Args: cobra.ExactArgs(1), RunE: g.runSccBQExportDelete,
	}
	g.sccBQExportsDescribeCmd = &cobra.Command{
		Use: "describe EXPORT", Short: "Describe a BigQuery export",
		Args: cobra.ExactArgs(1), RunE: g.runSccBQExportDescribe,
	}
	g.sccBQExportsListCmd = &cobra.Command{
		Use: "list", Short: "List BigQuery exports",
		Args: cobra.NoArgs, RunE: g.runSccBQExportList,
	}
	g.sccBQExportsUpdateCmd = &cobra.Command{
		Use: "update EXPORT", Short: "Update a BigQuery export from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccBQExportUpdate,
	}
	g.sccCustomModulesCmd = &cobra.Command{Use: "custom-modules", Short: "Manage SCC custom modules"}
	g.sccCustomModulesEtdCmd = &cobra.Command{Use: "etd", Short: "Manage Event Threat Detection custom modules"}
	g.sccEtdCreateCmd = &cobra.Command{
		Use: "create MODULE", Short: "Create an ETD custom module from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccEtdCreate,
	}
	g.sccEtdDeleteCmd = &cobra.Command{
		Use: "delete MODULE", Short: "Delete an ETD custom module",
		Args: cobra.ExactArgs(1), RunE: g.runSccEtdDelete,
	}
	g.sccEtdDescribeCmd = &cobra.Command{
		Use: "describe MODULE", Short: "Describe an ETD custom module",
		Args: cobra.ExactArgs(1), RunE: g.runSccEtdDescribe,
	}
	g.sccEtdDescribeEffectiveCmd = &cobra.Command{
		Use: "describe-effective MODULE", Short: "Describe an effective ETD custom module",
		Args: cobra.ExactArgs(1), RunE: g.runSccEtdDescribeEffective,
	}
	g.sccEtdListCmd = &cobra.Command{
		Use: "list", Short: "List ETD custom modules",
		Args: cobra.NoArgs, RunE: g.runSccEtdList,
	}
	g.sccEtdListDescendantCmd = &cobra.Command{
		Use: "list-descendant", Short: "List ETD custom modules and descendants",
		Args: cobra.NoArgs, RunE: g.runSccEtdListDescendant,
	}
	g.sccEtdListEffectiveCmd = &cobra.Command{
		Use: "list-effective", Short: "List effective ETD custom modules",
		Args: cobra.NoArgs, RunE: g.runSccEtdListEffective,
	}
	g.sccEtdUpdateCmd = &cobra.Command{
		Use: "update MODULE", Short: "Update an ETD custom module from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccEtdUpdate,
	}
	g.sccCustomModulesShaCmd = &cobra.Command{Use: "sha", Short: "Manage Security Health Analytics custom modules"}
	g.sccShaCreateCmd = &cobra.Command{
		Use: "create MODULE", Short: "Create a SHA custom module from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccShaCreate,
	}
	g.sccShaDeleteCmd = &cobra.Command{
		Use: "delete MODULE", Short: "Delete a SHA custom module",
		Args: cobra.ExactArgs(1), RunE: g.runSccShaDelete,
	}
	g.sccShaDescribeCmd = &cobra.Command{
		Use: "describe MODULE", Short: "Describe a SHA custom module",
		Args: cobra.ExactArgs(1), RunE: g.runSccShaDescribe,
	}
	g.sccShaDescribeEffectiveCmd = &cobra.Command{
		Use: "describe-effective MODULE", Short: "Describe an effective SHA custom module",
		Args: cobra.ExactArgs(1), RunE: g.runSccShaDescribeEffective,
	}
	g.sccShaListCmd = &cobra.Command{
		Use: "list", Short: "List SHA custom modules",
		Args: cobra.NoArgs, RunE: g.runSccShaList,
	}
	g.sccShaListDescendantCmd = &cobra.Command{
		Use: "list-descendant", Short: "List SHA custom modules and descendants",
		Args: cobra.NoArgs, RunE: g.runSccShaListDescendant,
	}
	g.sccShaListEffectiveCmd = &cobra.Command{
		Use: "list-effective", Short: "List effective SHA custom modules",
		Args: cobra.NoArgs, RunE: g.runSccShaListEffective,
	}
	g.sccShaUpdateCmd = &cobra.Command{
		Use: "update MODULE", Short: "Update a SHA custom module from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccShaUpdate,
	}
	g.sccFindingsCmd = &cobra.Command{Use: "findings", Short: "Manage findings"}
	g.sccFindingsBulkMuteCmd = &cobra.Command{
		Use: "bulk-mute", Short: "Bulk-mute findings that match a filter",
		Args: cobra.NoArgs, RunE: g.runSccFindingsBulkMute,
	}
	g.sccFindingsCreateCmd = &cobra.Command{
		Use: "create FINDING_ID", Short: "Create a finding from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccFindingsCreate,
	}
	g.sccFindingsExportBQCmd = &cobra.Command{
		Use: "export-to-bigquery [PARENT]", Short: "Export findings to a BigQuery dataset",
		Args: cobra.MaximumNArgs(1), RunE: g.runSccFindingsExportBQ,
	}
	g.sccFindingsGroupCmd = &cobra.Command{
		Use: "group [PARENT]", Short: "Group findings by fields",
		Args: cobra.MaximumNArgs(1), RunE: g.runSccFindingsGroup,
	}
	g.sccFindingsListCmd = &cobra.Command{
		Use: "list [PARENT]", Short: "List findings",
		Args: cobra.MaximumNArgs(1), RunE: g.runSccFindingsList,
	}
	g.sccFindingsListMarksCmd = &cobra.Command{
		Use: "list-marks FINDING", Short: "List security marks on a finding",
		Args: cobra.ExactArgs(1), RunE: g.runSccFindingsListMarks,
	}
	g.sccFindingsSetMuteCmd = &cobra.Command{
		Use: "set-mute FINDING", Short: "Set the mute state of a finding",
		Args: cobra.ExactArgs(1), RunE: g.runSccFindingsSetMute,
	}
	g.sccFindingsUpdateCmd = &cobra.Command{
		Use: "update FINDING", Short: "Update a finding from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccFindingsUpdate,
	}
	g.sccFindingsUpdateMarksCmd = &cobra.Command{
		Use: "update-marks FINDING", Short: "Update security marks on a finding",
		Args: cobra.ExactArgs(1), RunE: g.runSccFindingsUpdateMarks,
	}
	g.sccManageCmd = &cobra.Command{Use: "manage", Short: "Manage Security Command Center settings"}
	g.sccManageSettingsCmd = &cobra.Command{Use: "settings", Short: "Manage organization SCC settings"}
	g.sccManageSettingsDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe an organization's SCC settings",
		Args: cobra.NoArgs, RunE: g.runSccManageSettingsDescribe,
	}
	g.sccManageSettingsUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update an organization's SCC settings from a --config-file",
		Args: cobra.NoArgs, RunE: g.runSccManageSettingsUpdate,
	}
	g.sccMuteConfigsCmd = &cobra.Command{Use: "muteconfigs", Short: "Manage mute configs"}
	g.sccMuteCreateCmd = &cobra.Command{
		Use: "create MUTE_CONFIG", Short: "Create a mute config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccMuteCreate,
	}
	g.sccMuteDeleteCmd = &cobra.Command{
		Use: "delete MUTE_CONFIG", Short: "Delete a mute config",
		Args: cobra.ExactArgs(1), RunE: g.runSccMuteDelete,
	}
	g.sccMuteDescribeCmd = &cobra.Command{
		Use: "describe MUTE_CONFIG", Short: "Describe a mute config",
		Args: cobra.ExactArgs(1), RunE: g.runSccMuteDescribe,
	}
	g.sccMuteListCmd = &cobra.Command{
		Use: "list", Short: "List mute configs",
		Args: cobra.NoArgs, RunE: g.runSccMuteList,
	}
	g.sccMuteUpdateCmd = &cobra.Command{
		Use: "update MUTE_CONFIG", Short: "Update a mute config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccMuteUpdate,
	}
	g.sccNotificationsCmd = &cobra.Command{Use: "notifications", Short: "Manage notifications"}
	g.sccNotifyCreateCmd = &cobra.Command{
		Use: "create NOTIFICATION_CONFIG", Short: "Create a notification config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccNotifyCreate,
	}
	g.sccNotifyDeleteCmd = &cobra.Command{
		Use: "delete NOTIFICATION_CONFIG", Short: "Delete a notification config",
		Args: cobra.ExactArgs(1), RunE: g.runSccNotifyDelete,
	}
	g.sccNotifyDescribeCmd = &cobra.Command{
		Use: "describe NOTIFICATION_CONFIG", Short: "Describe a notification config",
		Args: cobra.ExactArgs(1), RunE: g.runSccNotifyDescribe,
	}
	g.sccNotifyListCmd = &cobra.Command{
		Use: "list", Short: "List notification configs",
		Args: cobra.NoArgs, RunE: g.runSccNotifyList,
	}
	g.sccNotifyUpdateCmd = &cobra.Command{
		Use: "update NOTIFICATION_CONFIG", Short: "Update a notification config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSccNotifyUpdate,
	}
	g.sccOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Security Command Center operations"}
	g.sccOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an SCC operation",
		Args: cobra.ExactArgs(1), RunE: g.runSccOpDescribe,
	}
	g.sccOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an SCC operation",
		Args: cobra.ExactArgs(1), RunE: g.runSccOpDelete,
	}
	g.sccOpListCmd = &cobra.Command{
		Use: "list", Short: "List SCC operations",
		Args: cobra.NoArgs, RunE: g.runSccOpList,
	}
	g.sccSourcesCmd = &cobra.Command{Use: "sources", Short: "Manage Security Command Center sources"}
	g.sccSourcesDescribeCmd = &cobra.Command{
		Use: "describe SOURCE", Short: "Describe a Security Command Center source",
		Args: cobra.ExactArgs(1), RunE: g.runSccSourcesDescribe,
	}
	g.sccSourcesListCmd = &cobra.Command{
		Use: "list", Short: "List Security Command Center sources",
		Args: cobra.NoArgs, RunE: g.runSccSourcesList,
	}
	g.sccManageRest = g.inv.restClient("https://securitycentermanagement.googleapis.com/v1")
	g.sccManageServicesCmd = &cobra.Command{
		Use:   "services",
		Short: "Manage Security Command Center services",
	}
	g.sccManageServicesDescribeCmd = &cobra.Command{
		Use:   "describe SERVICE_NAME",
		Short: "Describe a Security Command Center service",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSccManageServicesDescribe,
	}
	g.sccManageServicesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Security Command Center services",
		Args:  cobra.NoArgs,
		RunE:  g.runSccManageServicesList,
	}
	g.sccManageServicesUpdateCmd = &cobra.Command{
		Use:   "update SERVICE_NAME",
		Short: "Update a Security Command Center service",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSccManageServicesUpdate,
	}
	g.sccPosturesCmd = &cobra.Command{Use: "postures", Short: "Manage Security Command Center postures"}
	g.postureCreateCmd = &cobra.Command{
		Use: "create POSTURE", Short: "Create a posture from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runPostureCreate,
	}
	g.postureDeleteCmd = &cobra.Command{
		Use: "delete POSTURE", Short: "Delete a posture",
		Args: cobra.ExactArgs(1), RunE: g.runPostureDelete,
	}
	g.postureDescribeCmd = &cobra.Command{
		Use: "describe POSTURE", Short: "Describe a posture",
		Args: cobra.ExactArgs(1), RunE: g.runPostureDescribe,
	}
	g.postureExtractCmd = &cobra.Command{
		Use: "extract POSTURE_ID", Short: "Extract a posture from a workload",
		Args: cobra.ExactArgs(1), RunE: g.runPostureExtract,
	}
	g.postureListCmd = &cobra.Command{
		Use: "list", Short: "List postures",
		Args: cobra.NoArgs, RunE: g.runPostureList,
	}
	g.postureListRevisionsCmd = &cobra.Command{
		Use: "list-revisions POSTURE", Short: "List revisions of a posture",
		Args: cobra.ExactArgs(1), RunE: g.runPostureListRevisions,
	}
	g.postureUpdateCmd = &cobra.Command{
		Use: "update POSTURE", Short: "Update a posture from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runPostureUpdate,
	}
	g.sccPostureDeploymentsCmd = &cobra.Command{Use: "posture-deployments", Short: "Manage posture deployments"}
	g.pdCreateCmd = &cobra.Command{
		Use: "create DEPLOYMENT", Short: "Create a posture deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runPDCreate,
	}
	g.pdDeleteCmd = &cobra.Command{
		Use: "delete DEPLOYMENT", Short: "Delete a posture deployment",
		Args: cobra.ExactArgs(1), RunE: g.runPDDelete,
	}
	g.pdDescribeCmd = &cobra.Command{
		Use: "describe DEPLOYMENT", Short: "Describe a posture deployment",
		Args: cobra.ExactArgs(1), RunE: g.runPDDescribe,
	}
	g.pdListCmd = &cobra.Command{
		Use: "list", Short: "List posture deployments",
		Args: cobra.NoArgs, RunE: g.runPDList,
	}
	g.pdUpdateCmd = &cobra.Command{
		Use: "update DEPLOYMENT", Short: "Update a posture deployment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runPDUpdate,
	}
	g.sccPostureTemplatesCmd = &cobra.Command{Use: "posture-templates", Short: "Manage posture templates"}
	g.ptDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a posture template",
		Args: cobra.ExactArgs(1), RunE: g.runPTDescribe,
	}
	g.ptListCmd = &cobra.Command{
		Use: "list", Short: "List posture templates",
		Args: cobra.NoArgs, RunE: g.runPTList,
	}
	g.sccPostureOperationsCmd = &cobra.Command{Use: "posture-operations", Short: "Manage posture operations"}
	g.poCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a posture operation",
		Args: cobra.ExactArgs(1), RunE: g.runPOCancel,
	}
	g.poDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a posture operation",
		Args: cobra.ExactArgs(1), RunE: g.runPODelete,
	}
	g.poDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a posture operation",
		Args: cobra.ExactArgs(1), RunE: g.runPODescribe,
	}
	g.poListCmd = &cobra.Command{
		Use: "list", Short: "List posture operations",
		Args: cobra.NoArgs, RunE: g.runPOList,
	}
	g.sccIaCReportsCmd = &cobra.Command{Use: "iac-validation-reports", Short: "Manage IaC validation reports"}
	g.iacCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an IaC validation report from a --config-file",
		Args: cobra.NoArgs, RunE: g.runIaCCreate,
	}
	g.iacDescribeCmd = &cobra.Command{
		Use: "describe REPORT", Short: "Describe an IaC validation report",
		Args: cobra.ExactArgs(1), RunE: g.runIaCDescribe,
	}
	g.iacListCmd = &cobra.Command{
		Use: "list", Short: "List IaC validation reports",
		Args: cobra.NoArgs, RunE: g.runIaCList,
	}
	g.sccV2Rest = g.inv.restClient("https://securitycenter.googleapis.com/v2")
	g.initScc()
	g.initSccManageServices()
	g.initSccPosture()
	return g
}
