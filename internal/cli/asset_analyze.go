package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	asset "google.golang.org/api/cloudasset/v1"
)

func (s *assetCmds) initAssetAnalyze() {
	analyzeScopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&s.flagAssetAnalyzeProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&s.flagAssetAnalyzeFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&s.flagAssetAnalyzeOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}
	analyzeQueryFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&s.flagAssetAnalyzeFullResourceName, "full-resource-name", "", "Full resource name to analyze")
		c.Flags().StringVar(&s.flagAssetAnalyzeIdentity, "identity", "", "Identity to analyze (e.g. user:alice@example.com)")
		c.Flags().StringSliceVar(&s.flagAssetAnalyzePermissions, "permissions", nil, "Permissions to include in analysis")
		c.Flags().StringSliceVar(&s.flagAssetAnalyzeRoles, "roles", nil, "Roles to include in analysis")
		c.Flags().BoolVar(&s.flagAssetAnalyzeExpandGroups, "expand-groups", false, "Expand group memberships in results")
		c.Flags().BoolVar(&s.flagAssetAnalyzeExpandRoles, "expand-roles", false, "Expand role permissions in results")
		c.Flags().BoolVar(&s.flagAssetAnalyzeExpandResources, "expand-resources", false, "Expand descendant resources")
		c.Flags().BoolVar(&s.flagAssetAnalyzeGroupEdges, "output-group-edges", false, "Include group membership edges in results")
		c.Flags().BoolVar(&s.flagAssetAnalyzeResourceEdges, "output-resource-edges", false, "Include resource hierarchy edges in results")
		c.Flags().BoolVar(&s.flagAssetAnalyzeSAImp, "analyze-service-account-impersonation", false, "Include service account impersonation analysis")
		c.Flags().StringVar(&s.flagAssetAnalyzeSavedQuery, "saved-analysis-query", "", "Fully qualified saved analysis query resource name")
		c.Flags().StringVar(&s.flagAssetAnalyzeAccessTime, "access-time", "", "Hypothetical access time (RFC 3339) for IAM condition evaluation")
	}

	analyzeScopeFlags(s.assetAnalyzeIamPolicyCmd)
	analyzeQueryFlags(s.assetAnalyzeIamPolicyCmd)
	s.assetAnalyzeIamPolicyCmd.Flags().StringVar(&s.flagAssetAnalyzeExecTimeout, "execution-timeout", "", "Server-side timeout (e.g. 30s)")

	analyzeScopeFlags(s.assetAnalyzeIamPolicyLongCmd)
	analyzeQueryFlags(s.assetAnalyzeIamPolicyLongCmd)
	s.assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&s.flagAssetAnalyzeGCSUri, "gcs-output-path", "", "Cloud Storage URI (gs://bucket/object) for the results")
	s.assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&s.flagAssetAnalyzeBQDataset, "bigquery-dataset", "", "BigQuery dataset in projects/PROJECT/datasets/DATASET format")
	s.assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&s.flagAssetAnalyzeBQTable, "bigquery-table-prefix", "", "BigQuery table name prefix")
	s.assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&s.flagAssetAnalyzeBQPartKey, "bigquery-partition-key", "", "BigQuery partition key (REQUEST_TIME)")
	s.assetAnalyzeIamPolicyLongCmd.Flags().StringVar(&s.flagAssetAnalyzeBQWriteDisp, "bigquery-write-disposition", "", "BigQuery write disposition (WRITE_TRUNCATE, WRITE_APPEND, WRITE_EMPTY)")

	s.assetAnalyzeMoveCmd.Flags().StringVar(&s.flagAssetMoveDest, "destination-parent", "", "Destination parent (e.g. organizations/123 or folders/456) (required)")
	s.assetAnalyzeMoveCmd.Flags().StringVar(&s.flagAssetMoveView, "view", "", "Analysis view: BASIC or FULL")
	s.assetAnalyzeMoveCmd.MarkFlagRequired("destination-parent")

	orgPolicyScopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&s.flagAssetOPScope, "scope", "", "Scope (e.g. organizations/123) — overrides --project/--folder/--organization")
		c.Flags().StringVar(&s.flagAssetAnalyzeProject, "project", "", "Project ID")
		c.Flags().StringVar(&s.flagAssetAnalyzeFolder, "folder", "", "Folder ID")
		c.Flags().StringVar(&s.flagAssetAnalyzeOrg, "organization", "", "Organization ID")
		c.Flags().StringVar(&s.flagAssetOPConstraint, "constraint", "", "Constraint name (required)")
		c.Flags().StringVar(&s.flagAssetOPFilter, "filter", "", "Result filter")
		c.Flags().Int64Var(&s.flagAssetOPPageSize, "page-size", 0, "Page size for API pagination")
		c.Flags().Int64Var(&s.flagAssetOPLimit, "limit", 0, "Maximum number of results (0 = no limit)")
		c.Flags().StringVar(&s.flagAssetOPFormat, "format", "", "Output format (json, yaml, or table)")
		c.MarkFlagRequired("constraint")
	}
	orgPolicyScopeFlags(s.assetAnalyzeOrgPoliciesCmd)
	orgPolicyScopeFlags(s.assetAnalyzeOrgPolicyGovernedAssetsCmd)
	orgPolicyScopeFlags(s.assetAnalyzeOrgPolicyGovernedContainersCmd)

	s.assetCmd.AddCommand(
		s.assetAnalyzeIamPolicyCmd,
		s.assetAnalyzeIamPolicyLongCmd,
		s.assetAnalyzeMoveCmd,
		s.assetAnalyzeOrgPoliciesCmd,
		s.assetAnalyzeOrgPolicyGovernedAssetsCmd,
		s.assetAnalyzeOrgPolicyGovernedContainersCmd,
	)
}

// analyzeScope resolves the scope for analyze-iam-policy and family.
func (s *assetCmds) analyzeScope() (string, error) {
	return resolveAssetScope(s.flagAssetAnalyzeProject, s.flagAssetAnalyzeFolder, s.flagAssetAnalyzeOrg)
}

// orgPolicyAnalyzeScope resolves the scope for analyze-org-* commands,
// preferring an explicit --scope.
func (s *assetCmds) orgPolicyAnalyzeScope() (string, error) {
	if s.flagAssetOPScope != "" {
		return s.flagAssetOPScope, nil
	}
	return s.analyzeScope()
}

func (s *assetCmds) applyIamPolicyAnalysisFlags(c *asset.V1AnalyzeIamPolicyCall) *asset.V1AnalyzeIamPolicyCall {
	if s.flagAssetAnalyzeFullResourceName != "" {
		c = c.AnalysisQueryResourceSelectorFullResourceName(s.flagAssetAnalyzeFullResourceName)
	}
	if s.flagAssetAnalyzeIdentity != "" {
		c = c.AnalysisQueryIdentitySelectorIdentity(s.flagAssetAnalyzeIdentity)
	}
	if len(s.flagAssetAnalyzePermissions) > 0 {
		c = c.AnalysisQueryAccessSelectorPermissions(s.flagAssetAnalyzePermissions...)
	}
	if len(s.flagAssetAnalyzeRoles) > 0 {
		c = c.AnalysisQueryAccessSelectorRoles(s.flagAssetAnalyzeRoles...)
	}
	if s.flagAssetAnalyzeExpandGroups {
		c = c.AnalysisQueryOptionsExpandGroups(true)
	}
	if s.flagAssetAnalyzeExpandRoles {
		c = c.AnalysisQueryOptionsExpandRoles(true)
	}
	if s.flagAssetAnalyzeExpandResources {
		c = c.AnalysisQueryOptionsExpandResources(true)
	}
	if s.flagAssetAnalyzeGroupEdges {
		c = c.AnalysisQueryOptionsOutputGroupEdges(true)
	}
	if s.flagAssetAnalyzeResourceEdges {
		c = c.AnalysisQueryOptionsOutputResourceEdges(true)
	}
	if s.flagAssetAnalyzeSAImp {
		c = c.AnalysisQueryOptionsAnalyzeServiceAccountImpersonation(true)
	}
	if s.flagAssetAnalyzeSavedQuery != "" {
		c = c.SavedAnalysisQuery(s.flagAssetAnalyzeSavedQuery)
	}
	if s.flagAssetAnalyzeAccessTime != "" {
		c = c.AnalysisQueryConditionContextAccessTime(s.flagAssetAnalyzeAccessTime)
	}
	if s.flagAssetAnalyzeExecTimeout != "" {
		c = c.ExecutionTimeout(s.flagAssetAnalyzeExecTimeout)
	}
	return c
}

func (s *assetCmds) runAssetAnalyzeIamPolicy(cmd *cobra.Command, args []string) error {
	scope, err := s.analyzeScope()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	call := s.applyIamPolicyAnalysisFlags(svc.V1.AnalyzeIamPolicy(scope)).Context(ctx)
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("analyzing IAM policy: %w", err)
	}
	return s.inv.yaml(resp)
}

// buildAnalyzeLongOutputConfig assembles the JSON-body destination for the
// long-running analyze request.
func (s *assetCmds) buildAnalyzeLongOutputConfig() (*asset.IamPolicyAnalysisOutputConfig, error) {
	hasGCS := s.flagAssetAnalyzeGCSUri != ""
	hasBQ := s.flagAssetAnalyzeBQDataset != "" || s.flagAssetAnalyzeBQTable != ""
	if hasGCS && hasBQ {
		return nil, fmt.Errorf("specify either --gcs-output-path or the --bigquery-* flags, not both")
	}
	if !hasGCS && !hasBQ {
		return nil, fmt.Errorf("output destination is required: --gcs-output-path or --bigquery-dataset + --bigquery-table-prefix")
	}
	if hasGCS {
		return &asset.IamPolicyAnalysisOutputConfig{
			GcsDestination: &asset.GoogleCloudAssetV1GcsDestination{Uri: s.flagAssetAnalyzeGCSUri},
		}, nil
	}
	if s.flagAssetAnalyzeBQDataset == "" || s.flagAssetAnalyzeBQTable == "" {
		return nil, fmt.Errorf("both --bigquery-dataset and --bigquery-table-prefix are required for BigQuery output")
	}
	return &asset.IamPolicyAnalysisOutputConfig{
		BigqueryDestination: &asset.GoogleCloudAssetV1BigQueryDestination{
			Dataset:          s.flagAssetAnalyzeBQDataset,
			TablePrefix:      s.flagAssetAnalyzeBQTable,
			PartitionKey:     s.flagAssetAnalyzeBQPartKey,
			WriteDisposition: s.flagAssetAnalyzeBQWriteDisp,
		},
	}, nil
}

func (s *assetCmds) buildIamPolicyAnalysisQuery(scope string) *asset.IamPolicyAnalysisQuery {
	q := &asset.IamPolicyAnalysisQuery{Scope: scope}
	if s.flagAssetAnalyzeFullResourceName != "" {
		q.ResourceSelector = &asset.ResourceSelector{FullResourceName: s.flagAssetAnalyzeFullResourceName}
	}
	if s.flagAssetAnalyzeIdentity != "" {
		q.IdentitySelector = &asset.IdentitySelector{Identity: s.flagAssetAnalyzeIdentity}
	}
	if len(s.flagAssetAnalyzePermissions) > 0 || len(s.flagAssetAnalyzeRoles) > 0 {
		q.AccessSelector = &asset.AccessSelector{
			Permissions: s.flagAssetAnalyzePermissions,
			Roles:       s.flagAssetAnalyzeRoles,
		}
	}
	opts := &asset.Options{}
	optsSet := false
	if s.flagAssetAnalyzeExpandGroups {
		opts.ExpandGroups, optsSet = true, true
	}
	if s.flagAssetAnalyzeExpandRoles {
		opts.ExpandRoles, optsSet = true, true
	}
	if s.flagAssetAnalyzeExpandResources {
		opts.ExpandResources, optsSet = true, true
	}
	if s.flagAssetAnalyzeGroupEdges {
		opts.OutputGroupEdges, optsSet = true, true
	}
	if s.flagAssetAnalyzeResourceEdges {
		opts.OutputResourceEdges, optsSet = true, true
	}
	if s.flagAssetAnalyzeSAImp {
		opts.AnalyzeServiceAccountImpersonation, optsSet = true, true
	}
	if optsSet {
		q.Options = opts
	}
	return q
}

func (s *assetCmds) runAssetAnalyzeIamPolicyLong(cmd *cobra.Command, args []string) error {
	scope, err := s.analyzeScope()
	if err != nil {
		return err
	}
	outCfg, err := s.buildAnalyzeLongOutputConfig()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	req := &asset.AnalyzeIamPolicyLongrunningRequest{
		AnalysisQuery:      s.buildIamPolicyAnalysisQuery(scope),
		OutputConfig:       outCfg,
		SavedAnalysisQuery: s.flagAssetAnalyzeSavedQuery,
	}
	op, err := svc.V1.AnalyzeIamPolicyLongrunning(scope, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("analyzing IAM policy (long-running): %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Long-running analysis started: %s\n", op.Name)
	return s.inv.yaml(op)
}

func (s *assetCmds) runAssetAnalyzeMove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	call := svc.V1.AnalyzeMove(args[0]).DestinationParent(s.flagAssetMoveDest).Context(ctx)
	if s.flagAssetMoveView != "" {
		call = call.View(s.flagAssetMoveView)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("analyzing move: %w", err)
	}
	return s.inv.yaml(resp)
}

func (s *assetCmds) runAssetAnalyzeOrgPolicies(cmd *cobra.Command, args []string) error {
	scope, err := s.orgPolicyAnalyzeScope()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	var all []*asset.OrgPolicyResult
	pageToken := ""
	for {
		call := svc.V1.AnalyzeOrgPolicies(scope).Constraint(s.flagAssetOPConstraint).Context(ctx)
		if s.flagAssetOPFilter != "" {
			call = call.Filter(s.flagAssetOPFilter)
		}
		if s.flagAssetOPPageSize > 0 {
			call = call.PageSize(s.flagAssetOPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("analyzing org policies: %w", err)
		}
		all = append(all, resp.OrgPolicyResults...)
		if s.flagAssetOPLimit > 0 && int64(len(all)) >= s.flagAssetOPLimit {
			all = all[:s.flagAssetOPLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return s.inv.printListResults(all, s.flagAssetOPFormat, func() {
		printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "CONSTRAINT", "CONSOLIDATED_POLICY_ATTACHED_TO")
		for _, r := range all {
			attached := ""
			if r.ConsolidatedPolicy != nil {
				attached = r.ConsolidatedPolicy.AttachedResource
			}
			fmt.Fprintf(s.inv.stdout, "%-40s %s\n", s.flagAssetOPConstraint, attached)
		}
	})
}

func (s *assetCmds) runAssetAnalyzeOrgPolicyGovernedAssets(cmd *cobra.Command, args []string) error {
	scope, err := s.orgPolicyAnalyzeScope()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	var all []*asset.GoogleCloudAssetV1AnalyzeOrgPolicyGovernedAssetsResponseGovernedAsset
	pageToken := ""
	for {
		call := svc.V1.AnalyzeOrgPolicyGovernedAssets(scope).Constraint(s.flagAssetOPConstraint).Context(ctx)
		if s.flagAssetOPFilter != "" {
			call = call.Filter(s.flagAssetOPFilter)
		}
		if s.flagAssetOPPageSize > 0 {
			call = call.PageSize(s.flagAssetOPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("analyzing governed assets: %w", err)
		}
		all = append(all, resp.GovernedAssets...)
		if s.flagAssetOPLimit > 0 && int64(len(all)) >= s.flagAssetOPLimit {
			all = all[:s.flagAssetOPLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return s.inv.printListResults(all, s.flagAssetOPFormat, func() {
		printTableHeadingTo(s.inv.stdout, len(all), "%-60s\n", "GOVERNED_ASSET")
		for _, g := range all {
			if g.GovernedResource != nil {
				fmt.Fprintf(s.inv.stdout, "%-60s\n", g.GovernedResource.FullResourceName)
			} else if g.GovernedIamPolicy != nil {
				fmt.Fprintf(s.inv.stdout, "%-60s\n", g.GovernedIamPolicy.AttachedResource)
			}
		}
	})
}

func (s *assetCmds) runAssetAnalyzeOrgPolicyGovernedContainers(cmd *cobra.Command, args []string) error {
	scope, err := s.orgPolicyAnalyzeScope()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	var all []*asset.GoogleCloudAssetV1GovernedContainer
	pageToken := ""
	for {
		call := svc.V1.AnalyzeOrgPolicyGovernedContainers(scope).Constraint(s.flagAssetOPConstraint).Context(ctx)
		if s.flagAssetOPFilter != "" {
			call = call.Filter(s.flagAssetOPFilter)
		}
		if s.flagAssetOPPageSize > 0 {
			call = call.PageSize(s.flagAssetOPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("analyzing governed containers: %w", err)
		}
		all = append(all, resp.GovernedContainers...)
		if s.flagAssetOPLimit > 0 && int64(len(all)) >= s.flagAssetOPLimit {
			all = all[:s.flagAssetOPLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return s.inv.printListResults(all, s.flagAssetOPFormat, func() {
		printTableHeadingTo(s.inv.stdout, len(all), "%-60s %s\n", "NAME", "PARENT")
		for _, g := range all {
			fmt.Fprintf(s.inv.stdout, "%-60s %s\n", g.FullResourceName, g.Parent)
		}
	})
}
