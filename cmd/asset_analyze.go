package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	asset "google.golang.org/api/cloudasset/v1"
)

var assetAnalyzeIamPolicyCmd = &cobra.Command{
	Use:   "analyze-iam-policy",
	Short: "Analyze IAM policies that match a request",
	Args:  cobra.NoArgs,
	RunE:  runAssetAnalyzeIamPolicy,
}

var assetAnalyzeIamPolicyLongCmd = &cobra.Command{
	Use:   "analyze-iam-policy-longrunning",
	Short: "Analyze IAM policies asynchronously and write results to GCS or BigQuery",
	Args:  cobra.NoArgs,
	RunE:  runAssetAnalyzeIamPolicyLong,
}

var assetAnalyzeMoveCmd = &cobra.Command{
	Use:   "analyze-move RESOURCE",
	Short: "Analyze a resource move",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetAnalyzeMove,
}

var assetAnalyzeOrgPoliciesCmd = &cobra.Command{
	Use:   "analyze-org-policies",
	Short: "Analyze organization policies under a scope",
	Args:  cobra.NoArgs,
	RunE:  runAssetAnalyzeOrgPolicies,
}

var assetAnalyzeOrgPolicyGovernedAssetsCmd = &cobra.Command{
	Use:   "analyze-org-policy-governed-assets",
	Short: "Analyze organization-policy-governed assets under a scope",
	Args:  cobra.NoArgs,
	RunE:  runAssetAnalyzeOrgPolicyGovernedAssets,
}

var assetAnalyzeOrgPolicyGovernedContainersCmd = &cobra.Command{
	Use:   "analyze-org-policy-governed-containers",
	Short: "Analyze organization-policy-governed containers under a scope",
	Args:  cobra.NoArgs,
	RunE:  runAssetAnalyzeOrgPolicyGovernedContainers,
}

var (
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
	flagAssetAnalyzeGCSUri     string
	flagAssetAnalyzeBQDataset  string
	flagAssetAnalyzeBQTable    string
	flagAssetAnalyzeBQPartKey  string
	flagAssetAnalyzeBQWriteDisp string

	// analyze-move
	flagAssetMoveDest string
	flagAssetMoveView string

	// analyze-org-policies*
	flagAssetOPScope      string
	flagAssetOPConstraint string
	flagAssetOPFilter     string
	flagAssetOPPageSize   int64
	flagAssetOPLimit      int64
	flagAssetOPFormat     string
)

func init() {
	analyzeScopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&flagAssetAnalyzeProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&flagAssetAnalyzeFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&flagAssetAnalyzeOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}
	analyzeQueryFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&flagAssetAnalyzeFullResourceName, "full-resource-name", "", "Full resource name to analyze")
		c.Flags().StringVar(&flagAssetAnalyzeIdentity, "identity", "", "Identity to analyze (e.g. user:alice@example.com)")
		c.Flags().StringSliceVar(&flagAssetAnalyzePermissions, "permissions", nil, "Permissions to include in analysis")
		c.Flags().StringSliceVar(&flagAssetAnalyzeRoles, "roles", nil, "Roles to include in analysis")
		c.Flags().BoolVar(&flagAssetAnalyzeExpandGroups, "expand-groups", false, "Expand group memberships in results")
		c.Flags().BoolVar(&flagAssetAnalyzeExpandRoles, "expand-roles", false, "Expand role permissions in results")
		c.Flags().BoolVar(&flagAssetAnalyzeExpandResources, "expand-resources", false, "Expand descendant resources")
		c.Flags().BoolVar(&flagAssetAnalyzeGroupEdges, "output-group-edges", false, "Include group membership edges in results")
		c.Flags().BoolVar(&flagAssetAnalyzeResourceEdges, "output-resource-edges", false, "Include resource hierarchy edges in results")
		c.Flags().BoolVar(&flagAssetAnalyzeSAImp, "analyze-service-account-impersonation", false, "Include service account impersonation analysis")
		c.Flags().StringVar(&flagAssetAnalyzeSavedQuery, "saved-analysis-query", "", "Fully qualified saved analysis query resource name")
		c.Flags().StringVar(&flagAssetAnalyzeAccessTime, "access-time", "", "Hypothetical access time (RFC 3339) for IAM condition evaluation")
	}

	analyzeScopeFlags(assetAnalyzeIamPolicyCmd)
	analyzeQueryFlags(assetAnalyzeIamPolicyCmd)
	assetAnalyzeIamPolicyCmd.Flags().StringVar(&flagAssetAnalyzeExecTimeout, "execution-timeout", "", "Server-side timeout (e.g. 30s)")

	analyzeScopeFlags(assetAnalyzeIamPolicyLongCmd)
	analyzeQueryFlags(assetAnalyzeIamPolicyLongCmd)
	assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&flagAssetAnalyzeGCSUri, "gcs-output-path", "", "Cloud Storage URI (gs://bucket/object) for the results")
	assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&flagAssetAnalyzeBQDataset, "bigquery-dataset", "", "BigQuery dataset in projects/PROJECT/datasets/DATASET format")
	assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&flagAssetAnalyzeBQTable, "bigquery-table-prefix", "", "BigQuery table name prefix")
	assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&flagAssetAnalyzeBQPartKey, "bigquery-partition-key", "", "BigQuery partition key (REQUEST_TIME)")
	assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&flagAssetAnalyzeBQWriteDisp, "bigquery-write-disposition", "", "BigQuery write disposition (WRITE_TRUNCATE, WRITE_APPEND, WRITE_EMPTY)")

	assetAnalyzeMoveCmd.Flags().StringVar(&flagAssetMoveDest, "destination-parent", "", "Destination parent (e.g. organizations/123 or folders/456) (required)")
	assetAnalyzeMoveCmd.Flags().StringVar(&flagAssetMoveView, "view", "", "Analysis view: BASIC or FULL")
	assetAnalyzeMoveCmd.MarkFlagRequired("destination-parent")

	orgPolicyScopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&flagAssetOPScope, "scope", "", "Scope (e.g. organizations/123) — overrides --project/--folder/--organization")
		c.Flags().StringVar(&flagAssetAnalyzeProject, "project", "", "Project ID")
		c.Flags().StringVar(&flagAssetAnalyzeFolder, "folder", "", "Folder ID")
		c.Flags().StringVar(&flagAssetAnalyzeOrg, "organization", "", "Organization ID")
		c.Flags().StringVar(&flagAssetOPConstraint, "constraint", "", "Constraint name (required)")
		c.Flags().StringVar(&flagAssetOPFilter, "filter", "", "Result filter")
		c.Flags().Int64Var(&flagAssetOPPageSize, "page-size", 0, "Page size for API pagination")
		c.Flags().Int64Var(&flagAssetOPLimit, "limit", 0, "Maximum number of results (0 = no limit)")
		c.Flags().StringVar(&flagAssetOPFormat, "format", "", "Output format (json, yaml, or table)")
		c.MarkFlagRequired("constraint")
	}
	orgPolicyScopeFlags(assetAnalyzeOrgPoliciesCmd)
	orgPolicyScopeFlags(assetAnalyzeOrgPolicyGovernedAssetsCmd)
	orgPolicyScopeFlags(assetAnalyzeOrgPolicyGovernedContainersCmd)

	assetCmd.AddCommand(
		assetAnalyzeIamPolicyCmd,
		assetAnalyzeIamPolicyLongCmd,
		assetAnalyzeMoveCmd,
		assetAnalyzeOrgPoliciesCmd,
		assetAnalyzeOrgPolicyGovernedAssetsCmd,
		assetAnalyzeOrgPolicyGovernedContainersCmd,
	)
}

// analyzeScope resolves the scope for analyze-iam-policy and family.
func analyzeScope() (string, error) {
	return resolveAssetScope(flagAssetAnalyzeProject, flagAssetAnalyzeFolder, flagAssetAnalyzeOrg)
}

// orgPolicyAnalyzeScope resolves the scope for analyze-org-* commands,
// preferring an explicit --scope.
func orgPolicyAnalyzeScope() (string, error) {
	if flagAssetOPScope != "" {
		return flagAssetOPScope, nil
	}
	return analyzeScope()
}

func applyIamPolicyAnalysisFlags(c *asset.V1AnalyzeIamPolicyCall) *asset.V1AnalyzeIamPolicyCall {
	if flagAssetAnalyzeFullResourceName != "" {
		c = c.AnalysisQueryResourceSelectorFullResourceName(flagAssetAnalyzeFullResourceName)
	}
	if flagAssetAnalyzeIdentity != "" {
		c = c.AnalysisQueryIdentitySelectorIdentity(flagAssetAnalyzeIdentity)
	}
	if len(flagAssetAnalyzePermissions) > 0 {
		c = c.AnalysisQueryAccessSelectorPermissions(flagAssetAnalyzePermissions...)
	}
	if len(flagAssetAnalyzeRoles) > 0 {
		c = c.AnalysisQueryAccessSelectorRoles(flagAssetAnalyzeRoles...)
	}
	if flagAssetAnalyzeExpandGroups {
		c = c.AnalysisQueryOptionsExpandGroups(true)
	}
	if flagAssetAnalyzeExpandRoles {
		c = c.AnalysisQueryOptionsExpandRoles(true)
	}
	if flagAssetAnalyzeExpandResources {
		c = c.AnalysisQueryOptionsExpandResources(true)
	}
	if flagAssetAnalyzeGroupEdges {
		c = c.AnalysisQueryOptionsOutputGroupEdges(true)
	}
	if flagAssetAnalyzeResourceEdges {
		c = c.AnalysisQueryOptionsOutputResourceEdges(true)
	}
	if flagAssetAnalyzeSAImp {
		c = c.AnalysisQueryOptionsAnalyzeServiceAccountImpersonation(true)
	}
	if flagAssetAnalyzeSavedQuery != "" {
		c = c.SavedAnalysisQuery(flagAssetAnalyzeSavedQuery)
	}
	if flagAssetAnalyzeAccessTime != "" {
		c = c.AnalysisQueryConditionContextAccessTime(flagAssetAnalyzeAccessTime)
	}
	if flagAssetAnalyzeExecTimeout != "" {
		c = c.ExecutionTimeout(flagAssetAnalyzeExecTimeout)
	}
	return c
}

func runAssetAnalyzeIamPolicy(cmd *cobra.Command, args []string) error {
	scope, err := analyzeScope()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := applyIamPolicyAnalysisFlags(svc.V1.AnalyzeIamPolicy(scope)).Context(ctx)
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("analyzing IAM policy: %w", err)
	}
	return yamlEncode(resp)
}

// buildAnalyzeLongOutputConfig assembles the JSON-body destination for the
// long-running analyze request.
func buildAnalyzeLongOutputConfig() (*asset.IamPolicyAnalysisOutputConfig, error) {
	hasGCS := flagAssetAnalyzeGCSUri != ""
	hasBQ := flagAssetAnalyzeBQDataset != "" || flagAssetAnalyzeBQTable != ""
	if hasGCS && hasBQ {
		return nil, fmt.Errorf("specify either --gcs-output-path or the --bigquery-* flags, not both")
	}
	if !hasGCS && !hasBQ {
		return nil, fmt.Errorf("output destination is required: --gcs-output-path or --bigquery-dataset + --bigquery-table-prefix")
	}
	if hasGCS {
		return &asset.IamPolicyAnalysisOutputConfig{
			GcsDestination: &asset.GoogleCloudAssetV1GcsDestination{Uri: flagAssetAnalyzeGCSUri},
		}, nil
	}
	if flagAssetAnalyzeBQDataset == "" || flagAssetAnalyzeBQTable == "" {
		return nil, fmt.Errorf("both --bigquery-dataset and --bigquery-table-prefix are required for BigQuery output")
	}
	return &asset.IamPolicyAnalysisOutputConfig{
		BigqueryDestination: &asset.GoogleCloudAssetV1BigQueryDestination{
			Dataset:          flagAssetAnalyzeBQDataset,
			TablePrefix:      flagAssetAnalyzeBQTable,
			PartitionKey:     flagAssetAnalyzeBQPartKey,
			WriteDisposition: flagAssetAnalyzeBQWriteDisp,
		},
	}, nil
}

func buildIamPolicyAnalysisQuery(scope string) *asset.IamPolicyAnalysisQuery {
	q := &asset.IamPolicyAnalysisQuery{Scope: scope}
	if flagAssetAnalyzeFullResourceName != "" {
		q.ResourceSelector = &asset.ResourceSelector{FullResourceName: flagAssetAnalyzeFullResourceName}
	}
	if flagAssetAnalyzeIdentity != "" {
		q.IdentitySelector = &asset.IdentitySelector{Identity: flagAssetAnalyzeIdentity}
	}
	if len(flagAssetAnalyzePermissions) > 0 || len(flagAssetAnalyzeRoles) > 0 {
		q.AccessSelector = &asset.AccessSelector{
			Permissions: flagAssetAnalyzePermissions,
			Roles:       flagAssetAnalyzeRoles,
		}
	}
	opts := &asset.Options{}
	optsSet := false
	if flagAssetAnalyzeExpandGroups {
		opts.ExpandGroups, optsSet = true, true
	}
	if flagAssetAnalyzeExpandRoles {
		opts.ExpandRoles, optsSet = true, true
	}
	if flagAssetAnalyzeExpandResources {
		opts.ExpandResources, optsSet = true, true
	}
	if flagAssetAnalyzeGroupEdges {
		opts.OutputGroupEdges, optsSet = true, true
	}
	if flagAssetAnalyzeResourceEdges {
		opts.OutputResourceEdges, optsSet = true, true
	}
	if flagAssetAnalyzeSAImp {
		opts.AnalyzeServiceAccountImpersonation, optsSet = true, true
	}
	if optsSet {
		q.Options = opts
	}
	return q
}

func runAssetAnalyzeIamPolicyLong(cmd *cobra.Command, args []string) error {
	scope, err := analyzeScope()
	if err != nil {
		return err
	}
	outCfg, err := buildAnalyzeLongOutputConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &asset.AnalyzeIamPolicyLongrunningRequest{
		AnalysisQuery:      buildIamPolicyAnalysisQuery(scope),
		OutputConfig:       outCfg,
		SavedAnalysisQuery: flagAssetAnalyzeSavedQuery,
	}
	op, err := svc.V1.AnalyzeIamPolicyLongrunning(scope, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("analyzing IAM policy (long-running): %w", err)
	}
	fmt.Printf("Long-running analysis started: %s\n", op.Name)
	return yamlEncode(op)
}

func runAssetAnalyzeMove(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.V1.AnalyzeMove(args[0]).DestinationParent(flagAssetMoveDest).Context(ctx)
	if flagAssetMoveView != "" {
		call = call.View(flagAssetMoveView)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("analyzing move: %w", err)
	}
	return yamlEncode(resp)
}

func runAssetAnalyzeOrgPolicies(cmd *cobra.Command, args []string) error {
	scope, err := orgPolicyAnalyzeScope()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*asset.OrgPolicyResult
	pageToken := ""
	for {
		call := svc.V1.AnalyzeOrgPolicies(scope).Constraint(flagAssetOPConstraint).Context(ctx)
		if flagAssetOPFilter != "" {
			call = call.Filter(flagAssetOPFilter)
		}
		if flagAssetOPPageSize > 0 {
			call = call.PageSize(flagAssetOPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("analyzing org policies: %w", err)
		}
		all = append(all, resp.OrgPolicyResults...)
		if flagAssetOPLimit > 0 && int64(len(all)) >= flagAssetOPLimit {
			all = all[:flagAssetOPLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return printListResults(all, flagAssetOPFormat, func() {
		fmt.Printf("%-40s %s\n", "CONSTRAINT", "CONSOLIDATED_POLICY_ATTACHED_TO")
		for _, r := range all {
			attached := ""
			if r.ConsolidatedPolicy != nil {
				attached = r.ConsolidatedPolicy.AttachedResource
			}
			fmt.Printf("%-40s %s\n", flagAssetOPConstraint, attached)
		}
	})
}

func runAssetAnalyzeOrgPolicyGovernedAssets(cmd *cobra.Command, args []string) error {
	scope, err := orgPolicyAnalyzeScope()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*asset.GoogleCloudAssetV1AnalyzeOrgPolicyGovernedAssetsResponseGovernedAsset
	pageToken := ""
	for {
		call := svc.V1.AnalyzeOrgPolicyGovernedAssets(scope).Constraint(flagAssetOPConstraint).Context(ctx)
		if flagAssetOPFilter != "" {
			call = call.Filter(flagAssetOPFilter)
		}
		if flagAssetOPPageSize > 0 {
			call = call.PageSize(flagAssetOPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("analyzing governed assets: %w", err)
		}
		all = append(all, resp.GovernedAssets...)
		if flagAssetOPLimit > 0 && int64(len(all)) >= flagAssetOPLimit {
			all = all[:flagAssetOPLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return printListResults(all, flagAssetOPFormat, func() {
		fmt.Printf("%-60s\n", "GOVERNED_ASSET")
		for _, g := range all {
			if g.GovernedResource != nil {
				fmt.Printf("%-60s\n", g.GovernedResource.FullResourceName)
			} else if g.GovernedIamPolicy != nil {
				fmt.Printf("%-60s\n", g.GovernedIamPolicy.AttachedResource)
			}
		}
	})
}

func runAssetAnalyzeOrgPolicyGovernedContainers(cmd *cobra.Command, args []string) error {
	scope, err := orgPolicyAnalyzeScope()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*asset.GoogleCloudAssetV1GovernedContainer
	pageToken := ""
	for {
		call := svc.V1.AnalyzeOrgPolicyGovernedContainers(scope).Constraint(flagAssetOPConstraint).Context(ctx)
		if flagAssetOPFilter != "" {
			call = call.Filter(flagAssetOPFilter)
		}
		if flagAssetOPPageSize > 0 {
			call = call.PageSize(flagAssetOPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("analyzing governed containers: %w", err)
		}
		all = append(all, resp.GovernedContainers...)
		if flagAssetOPLimit > 0 && int64(len(all)) >= flagAssetOPLimit {
			all = all[:flagAssetOPLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return printListResults(all, flagAssetOPFormat, func() {
		fmt.Printf("%-60s %s\n", "NAME", "PARENT")
		for _, g := range all {
			fmt.Printf("%-60s %s\n", g.FullResourceName, g.Parent)
		}
	})
}
