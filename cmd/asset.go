package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud asset (#282) ---

func (s *assetCmds) initAsset() {

}

// resolveAssetScope returns a fully qualified scope for asset commands,
// accepting exactly one of --project, --folder, or --organization.
func resolveAssetScope(project, folder, org string) (string, error) {
	set := 0
	if project != "" {
		set++
	}
	if folder != "" {
		set++
	}
	if org != "" {
		set++
	}
	if set == 0 {
		return "", fmt.Errorf("one of --project, --folder, or --organization is required")
	}
	if set > 1 {
		return "", fmt.Errorf("specify only one of --project, --folder, or --organization")
	}
	switch {
	case project != "":
		return "projects/" + project, nil
	case folder != "":
		return "folders/" + strings.TrimPrefix(folder, "folders/"), nil
	default:
		return "organizations/" + strings.TrimPrefix(org, "organizations/"), nil
	}
}

// assetCmds holds the commands and flag values of one invocation's "asset"
// command group, so invocations don't share state (#2070).
type assetCmds struct {
	inv *invocation

	assetCmd                                   *cobra.Command
	assetAnalyzeIamPolicyCmd                   *cobra.Command
	assetAnalyzeIamPolicyLongCmd               *cobra.Command
	assetAnalyzeMoveCmd                        *cobra.Command
	assetAnalyzeOrgPoliciesCmd                 *cobra.Command
	assetAnalyzeOrgPolicyGovernedAssetsCmd     *cobra.Command
	assetAnalyzeOrgPolicyGovernedContainersCmd *cobra.Command
	// analyze-iam-policy / -longrunning
	flagAssetAnalyzeProject          string
	flagAssetAnalyzeFolder           string
	flagAssetAnalyzeOrg              string
	flagAssetAnalyzeFullResourceName string
	flagAssetAnalyzeIdentity         string
	flagAssetAnalyzePermissions      []string
	flagAssetAnalyzeRoles            []string
	flagAssetAnalyzeExpandGroups     bool
	flagAssetAnalyzeExpandRoles      bool
	flagAssetAnalyzeExpandResources  bool
	flagAssetAnalyzeGroupEdges       bool
	flagAssetAnalyzeResourceEdges    bool
	flagAssetAnalyzeSAImp            bool
	flagAssetAnalyzeExecTimeout      string
	flagAssetAnalyzeSavedQuery       string
	flagAssetAnalyzeAccessTime       string
	// longrunning output
	flagAssetAnalyzeGCSUri      string
	flagAssetAnalyzeBQDataset   string
	flagAssetAnalyzeBQTable     string
	flagAssetAnalyzeBQPartKey   string
	flagAssetAnalyzeBQWriteDisp string
	// analyze-move
	flagAssetMoveDest string
	flagAssetMoveView string
	// analyze-org-policies*
	flagAssetOPScope              string
	flagAssetOPConstraint         string
	flagAssetOPFilter             string
	flagAssetOPPageSize           int64
	flagAssetOPLimit              int64
	flagAssetOPFormat             string
	assetListCmd                  *cobra.Command
	assetGetHistoryCmd            *cobra.Command
	assetQueryCmd                 *cobra.Command
	assetExportCmd                *cobra.Command
	assetGetEffectiveIamPolicyCmd *cobra.Command
	assetSearchAllResourcesCmd    *cobra.Command
	assetSearchAllIamPoliciesCmd  *cobra.Command
	// scope selection reused by all top-level commands
	flagAssetProject string
	flagAssetFolder  string
	flagAssetOrg     string
	// list / get-history
	flagAssetAssetTypes    []string
	flagAssetContentType   string
	flagAssetSnapshotTime  string
	flagAssetRelationships []string
	flagAssetListPageSize  int64
	flagAssetListLimit     int64
	flagAssetListFormat    string
	flagAssetAssetNames    []string
	flagAssetHistoryStart  string
	flagAssetHistoryEnd    string
	// query
	flagAssetQueryStatement string
	flagAssetQueryJobRef    string
	flagAssetQueryTimeout   string
	flagAssetQueryPageSize  int64
	flagAssetQueryPageToken string
	// export
	flagAssetExportGCS       string
	flagAssetExportGCSPrefix string
	flagAssetExportBQDataset string
	flagAssetExportBQTable   string
	flagAssetExportBQForce   bool
	// get-effective-iam-policy
	flagAssetEffectiveNames []string
	// search-all-*
	flagAssetSearchScope        string
	flagAssetSearchQuery        string
	flagAssetSearchOrderBy      string
	flagAssetSearchPageSize     int64
	flagAssetSearchLimit        int64
	flagAssetSearchReadMask     string
	assetFeedsCmd               *cobra.Command
	assetFeedCreateCmd          *cobra.Command
	assetFeedDeleteCmd          *cobra.Command
	assetFeedDescribeCmd        *cobra.Command
	assetFeedListCmd            *cobra.Command
	assetFeedUpdateCmd          *cobra.Command
	flagAssetFeedProject        string
	flagAssetFeedFolder         string
	flagAssetFeedOrg            string
	flagAssetFeedAssetNames     []string
	flagAssetFeedAssetTypes     []string
	flagAssetFeedContentType    string
	flagAssetFeedPubsubTopic    string
	flagAssetFeedConditionExpr  string
	flagAssetFeedConditionTitle string
	flagAssetFeedConditionDesc  string
	flagAssetFeedRelationships  []string
	flagAssetFeedListFormat     string
	assetOperationsCmd          *cobra.Command
	assetOperationDescribeCmd   *cobra.Command
	assetSavedQueriesCmd        *cobra.Command
	assetSavedQueryCreateCmd    *cobra.Command
	assetSavedQueryDeleteCmd    *cobra.Command
	assetSavedQueryDescribeCmd  *cobra.Command
	assetSavedQueryListCmd      *cobra.Command
	assetSavedQueryUpdateCmd    *cobra.Command
	flagAssetSQProject          string
	flagAssetSQFolder           string
	flagAssetSQOrg              string
	flagAssetSQDescription      string
	flagAssetSQQueryFile        string
	flagAssetSQLabels           map[string]string
	flagAssetSQListFilter       string
	flagAssetSQListFormat       string
}

// newAssetCmd returns the "asset" command group for inv.
func newAssetCmd(inv *invocation) *cobra.Command {
	return newAssetCmds(inv).assetCmd
}

// newAssetCmds builds the "asset" command group for inv and returns its state.
func newAssetCmds(inv *invocation) *assetCmds {
	s := &assetCmds{inv: inv}
	s.assetCmd = &cobra.Command{
		Use:   "asset",
		Short: "Manage Cloud Asset Inventory",
	}
	s.assetAnalyzeIamPolicyCmd = &cobra.Command{
		Use:   "analyze-iam-policy",
		Short: "Analyze IAM policies that match a request",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetAnalyzeIamPolicy,
	}
	s.assetAnalyzeIamPolicyLongCmd = &cobra.Command{
		Use:   "analyze-iam-policy-longrunning",
		Short: "Analyze IAM policies asynchronously and write results to GCS or BigQuery",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetAnalyzeIamPolicyLong,
	}
	s.assetAnalyzeMoveCmd = &cobra.Command{
		Use:   "analyze-move RESOURCE",
		Short: "Analyze a resource move",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetAnalyzeMove,
	}
	s.assetAnalyzeOrgPoliciesCmd = &cobra.Command{
		Use:   "analyze-org-policies",
		Short: "Analyze organization policies under a scope",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetAnalyzeOrgPolicies,
	}
	s.assetAnalyzeOrgPolicyGovernedAssetsCmd = &cobra.Command{
		Use:   "analyze-org-policy-governed-assets",
		Short: "Analyze organization-policy-governed assets under a scope",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetAnalyzeOrgPolicyGovernedAssets,
	}
	s.assetAnalyzeOrgPolicyGovernedContainersCmd = &cobra.Command{
		Use:   "analyze-org-policy-governed-containers",
		Short: "Analyze organization-policy-governed containers under a scope",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetAnalyzeOrgPolicyGovernedContainers,
	}
	s.assetListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Cloud assets under a parent",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetList,
	}
	s.assetGetHistoryCmd = &cobra.Command{
		Use:   "get-history",
		Short: "Get the update history of assets within a time window",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetGetHistory,
	}
	s.assetQueryCmd = &cobra.Command{
		Use:   "query",
		Short: "Query Cloud assets with a SQL-like statement",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetQuery,
	}
	s.assetExportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export Cloud assets to Cloud Storage or BigQuery",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetExport,
	}
	s.assetGetEffectiveIamPolicyCmd = &cobra.Command{
		Use:   "get-effective-iam-policy",
		Short: "Get effective IAM policies for a batch of resources",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetGetEffectiveIamPolicy,
	}
	s.assetSearchAllResourcesCmd = &cobra.Command{
		Use:   "search-all-resources",
		Short: "Search all Cloud resources within an accessible scope",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetSearchAllResources,
	}
	s.assetSearchAllIamPoliciesCmd = &cobra.Command{
		Use:   "search-all-iam-policies",
		Short: "Search all IAM policies within an accessible scope",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetSearchAllIamPolicies,
	}
	s.assetFeedsCmd = &cobra.Command{
		Use:   "feeds",
		Short: "Manage Cloud Asset Inventory feeds",
	}
	s.assetFeedCreateCmd = &cobra.Command{
		Use:   "create FEED_ID",
		Short: "Create an asset feed",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetFeedCreate,
	}
	s.assetFeedDeleteCmd = &cobra.Command{
		Use:   "delete FEED_ID",
		Short: "Delete an asset feed",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetFeedDelete,
	}
	s.assetFeedDescribeCmd = &cobra.Command{
		Use:   "describe FEED_ID",
		Short: "Describe an asset feed",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetFeedDescribe,
	}
	s.assetFeedListCmd = &cobra.Command{
		Use:   "list",
		Short: "List asset feeds under a parent",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetFeedList,
	}
	s.assetFeedUpdateCmd = &cobra.Command{
		Use:   "update FEED_ID",
		Short: "Update an asset feed",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetFeedUpdate,
	}
	s.assetOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Cloud Asset Inventory operations",
	}
	s.assetOperationDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION_NAME",
		Short: "Describe a Cloud Asset Inventory operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetOperationDescribe,
	}
	s.assetSavedQueriesCmd = &cobra.Command{
		Use:   "saved-queries",
		Short: "Manage Cloud Asset Inventory saved queries",
	}
	s.assetSavedQueryCreateCmd = &cobra.Command{
		Use:   "create QUERY_ID",
		Short: "Create a saved query",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetSavedQueryCreate,
	}
	s.assetSavedQueryDeleteCmd = &cobra.Command{
		Use:   "delete QUERY_ID",
		Short: "Delete a saved query",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetSavedQueryDelete,
	}
	s.assetSavedQueryDescribeCmd = &cobra.Command{
		Use:   "describe QUERY_ID",
		Short: "Describe a saved query",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetSavedQueryDescribe,
	}
	s.assetSavedQueryListCmd = &cobra.Command{
		Use:   "list",
		Short: "List saved queries",
		Args:  cobra.NoArgs,
		RunE:  s.runAssetSavedQueryList,
	}
	s.assetSavedQueryUpdateCmd = &cobra.Command{
		Use:   "update QUERY_ID",
		Short: "Update a saved query",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAssetSavedQueryUpdate,
	}
	s.initAsset()
	s.initAssetAnalyze()
	s.initAssetCommands()
	s.initAssetFeeds()
	s.initAssetOperations()
	s.initAssetSavedQueries()
	return s
}
