package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

var dmCWCmd = &cobra.Command{
	Use:   "conversion-workspaces",
	Short: "Manage Database Migration Service conversion workspaces",
}

var (
	dmCWCreateCmd = &cobra.Command{
		Use: "create WORKSPACE", Short: "Create a conversion workspace from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDMCWCreate,
	}
	dmCWDeleteCmd = &cobra.Command{
		Use: "delete WORKSPACE", Short: "Delete a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: runDMCWDelete,
	}
	dmCWDescribeCmd = &cobra.Command{
		Use: "describe WORKSPACE", Short: "Show details about a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: runDMCWDescribe,
	}
	dmCWUpdateCmd = &cobra.Command{
		Use: "update WORKSPACE", Short: "Update a conversion workspace from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDMCWUpdate,
	}
	dmCWListCmd = &cobra.Command{
		Use: "list", Short: "List conversion workspaces in a region",
		Args: cobra.NoArgs, RunE: runDMCWList,
	}
	dmCWApplyCmd = &cobra.Command{
		Use: "apply WORKSPACE", Short: "Apply the conversion workspace to the destination",
		Args: cobra.ExactArgs(1), RunE: runDMCWApply,
	}
	dmCWCommitCmd = &cobra.Command{
		Use: "commit WORKSPACE", Short: "Commit a conversion workspace revision",
		Args: cobra.ExactArgs(1), RunE: runDMCWCommit,
	}
	dmCWConvertCmd = &cobra.Command{
		Use: "convert WORKSPACE", Short: "Convert a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: runDMCWConvert,
	}
	dmCWRollbackCmd = &cobra.Command{
		Use: "rollback WORKSPACE", Short: "Rollback the conversion workspace to the previous commit",
		Args: cobra.ExactArgs(1), RunE: runDMCWRollback,
	}
	dmCWSeedCmd = &cobra.Command{
		Use: "seed WORKSPACE", Short: "Seed a conversion workspace from a connection profile",
		Args: cobra.ExactArgs(1), RunE: runDMCWSeed,
	}
	dmCWDescribeDDLsCmd = &cobra.Command{
		Use: "describe-ddls WORKSPACE", Short: "Describe the DDLs for a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: runDMCWDescribeDDLs,
	}
	dmCWDescribeEntitiesCmd = &cobra.Command{
		Use: "describe-entities WORKSPACE", Short: "Describe entities in a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: runDMCWDescribeEntities,
	}
	dmCWDescribeIssuesCmd = &cobra.Command{
		Use: "describe-issues WORKSPACE", Short: "Describe conversion issues in a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: runDMCWDescribeIssues,
	}
	dmCWImportRulesCmd = &cobra.Command{
		Use: "import-rules WORKSPACE", Short: "Import mapping rules into a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: runDMCWImportRules,
	}
	dmCWListBackgroundJobsCmd = &cobra.Command{
		Use: "list-background-jobs WORKSPACE", Short: "List background jobs for a conversion workspace",
		Args: cobra.ExactArgs(1), RunE: runDMCWListBackgroundJobs,
	}
	dmCWMappingRulesCmd = &cobra.Command{
		Use:   "mapping-rules",
		Short: "Manage mapping rules for a conversion workspace",
	}
	dmCWMRListCmd = &cobra.Command{
		Use: "list", Short: "List mapping rules in a conversion workspace",
		Args: cobra.NoArgs, RunE: runDMCWMRList,
	}
	dmCWMRDescribeCmd = &cobra.Command{
		Use: "describe RULE", Short: "Show a mapping rule",
		Args: cobra.ExactArgs(1), RunE: runDMCWMRDescribe,
	}
	dmCWMRDeleteCmd = &cobra.Command{
		Use: "delete RULE", Short: "Delete a mapping rule",
		Args: cobra.ExactArgs(1), RunE: runDMCWMRDelete,
	}
	dmCWMRCreateCmd = &cobra.Command{
		Use: "create RULE", Short: "Create a mapping rule from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDMCWMRCreate,
	}
)

var (
	flagDMCWRegion            string
	flagDMCWFormat            string
	flagDMCWConfigFile        string
	flagDMCWUpdateMask        string
	flagDMCWListPageSize      int64
	flagDMCWListLimit         int64
	flagDMCWListFilter        string
	flagDMCWListURI           bool
	flagDMCWAsync             bool
	flagDMCWAutoCommit        bool
	flagDMCWDryRun            bool
	flagDMCWFilter            string
	flagDMCWDestConnProfile   string
	flagDMCWSourceConnProfile string
	flagDMCWConnProfile       string
	flagDMCWCommitName        string
	flagDMCWConvertFullPath   bool
	flagDMCWDescribeCommitID  string
	flagDMCWDescribeTree      string
	flagDMCWDescribeView      string
	flagDMCWDescribeUncomm    bool
	flagDMCWDescribeDBName    string
	flagDMCWRulesFormat       string
	flagDMCWRulesFiles        []string
	flagDMCWWorkspace         string
)

func init() {
	// Top-level workspace commands.
	topCommands := []*cobra.Command{
		dmCWCreateCmd, dmCWDeleteCmd, dmCWDescribeCmd, dmCWUpdateCmd, dmCWListCmd,
		dmCWApplyCmd, dmCWCommitCmd, dmCWConvertCmd, dmCWRollbackCmd, dmCWSeedCmd,
		dmCWDescribeDDLsCmd, dmCWDescribeEntitiesCmd, dmCWDescribeIssuesCmd,
		dmCWImportRulesCmd, dmCWListBackgroundJobsCmd,
	}
	for _, c := range topCommands {
		c.Flags().StringVar(&flagDMCWRegion, "region", "", "Region containing the conversion workspace (required)")
		_ = c.MarkFlagRequired("region")
	}

	for _, c := range []*cobra.Command{dmCWCreateCmd, dmCWUpdateCmd} {
		c.Flags().StringVar(&flagDMCWConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the ConversionWorkspace message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dmCWUpdateCmd.Flags().StringVar(&flagDMCWUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	// LRO commands get --async.
	for _, c := range []*cobra.Command{
		dmCWCreateCmd, dmCWDeleteCmd, dmCWUpdateCmd, dmCWApplyCmd, dmCWCommitCmd,
		dmCWConvertCmd, dmCWRollbackCmd, dmCWSeedCmd, dmCWImportRulesCmd,
	} {
		c.Flags().BoolVar(&flagDMCWAsync, "async", false, "Return the long-running operation without waiting")
	}

	// apply / convert / seed common options.
	for _, c := range []*cobra.Command{dmCWApplyCmd, dmCWConvertCmd, dmCWSeedCmd} {
		c.Flags().BoolVar(&flagDMCWAutoCommit, "auto-commit", false, "Commit the workspace automatically after the operation")
		c.Flags().StringVar(&flagDMCWFilter, "filter", "", "AIP-160 style filter for entities to include")
	}
	dmCWApplyCmd.Flags().BoolVar(&flagDMCWDryRun, "dry-run", false,
		"Validate the apply without changing the destination database")
	dmCWApplyCmd.Flags().StringVar(&flagDMCWConnProfile, "connection-profile", "",
		"Fully qualified name of the destination connection profile")
	dmCWConvertCmd.Flags().BoolVar(&flagDMCWConvertFullPath, "convert-full-path", false,
		"Automatically convert the full entity path for each entity matching --filter")

	dmCWSeedCmd.Flags().StringVar(&flagDMCWSourceConnProfile, "source-connection-profile", "",
		"Fully qualified name of the source connection profile")
	dmCWSeedCmd.Flags().StringVar(&flagDMCWDestConnProfile, "destination-connection-profile", "",
		"Fully qualified name of the destination connection profile")

	dmCWCommitCmd.Flags().StringVar(&flagDMCWCommitName, "commit-name", "", "Optional name for the commit")

	// describe-entities / describe-ddls / describe-issues shared filters.
	for _, c := range []*cobra.Command{dmCWDescribeDDLsCmd, dmCWDescribeEntitiesCmd, dmCWDescribeIssuesCmd} {
		c.Flags().StringVar(&flagDMCWDescribeCommitID, "commit-id", "", "Describe entities as of this commit")
		c.Flags().StringVar(&flagDMCWDescribeTree, "tree", "", "Which entity tree to walk (SOURCE, DRAFT, DESTINATION)")
		c.Flags().BoolVar(&flagDMCWDescribeUncomm, "uncommitted", false, "Include uncommitted changes")
		c.Flags().StringVar(&flagDMCWDescribeDBName, "database", "", "Restrict entities to this database")
		c.Flags().StringVar(&flagDMCWFormat, "format", "", "Output format")
	}
	dmCWDescribeEntitiesCmd.Flags().StringVar(&flagDMCWDescribeView, "view", "",
		"Which entity view to return (BASIC or FULL)")

	dmCWImportRulesCmd.Flags().StringSliceVar(&flagDMCWRulesFiles, "file", nil,
		"Path(s) to rules files (repeat to import multiple files)")
	dmCWImportRulesCmd.Flags().StringVar(&flagDMCWRulesFormat, "rules-format", "IMPORT_RULES_FILE_FORMAT_HARBOUR_BRIDGE_SESSION_FILE",
		"Format of the rules content file")
	dmCWImportRulesCmd.Flags().BoolVar(&flagDMCWAutoCommit, "auto-commit", false,
		"Commit the workspace automatically after import")
	_ = dmCWImportRulesCmd.MarkFlagRequired("file")

	dmCWDescribeCmd.Flags().StringVar(&flagDMCWFormat, "format", "", "Output format")
	dmCWListCmd.Flags().StringVar(&flagDMCWFormat, "format", "", "Output format")
	dmCWListCmd.Flags().Int64Var(&flagDMCWListPageSize, "page-size", 0, "Page size")
	dmCWListCmd.Flags().Int64Var(&flagDMCWListLimit, "limit", 0, "Cap total results")
	dmCWListCmd.Flags().StringVar(&flagDMCWListFilter, "filter", "", "Server-side filter expression")
	dmCWListCmd.Flags().BoolVar(&flagDMCWListURI, "uri", false, "Print resource names only")

	dmCWListBackgroundJobsCmd.Flags().StringVar(&flagDMCWFormat, "format", "", "Output format")
	dmCWListBackgroundJobsCmd.Flags().Int64Var(&flagDMCWListLimit, "limit", 0, "Cap total results (0 = no cap)")

	dmCWMappingRulesCmd.AddCommand(dmCWMRCreateCmd, dmCWMRDeleteCmd, dmCWMRDescribeCmd, dmCWMRListCmd)
	for _, c := range dmCWMappingRulesCmd.Commands() {
		c.Flags().StringVar(&flagDMCWRegion, "region", "", "Region containing the conversion workspace (required)")
		c.Flags().StringVar(&flagDMCWWorkspace, "conversion-workspace", "",
			"Name of the parent conversion workspace (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("conversion-workspace")
	}
	dmCWMRListCmd.Flags().Int64Var(&flagDMCWListPageSize, "page-size", 0, "Page size")
	dmCWMRListCmd.Flags().Int64Var(&flagDMCWListLimit, "limit", 0, "Cap total results")
	dmCWMRListCmd.Flags().StringVar(&flagDMCWFormat, "format", "", "Output format")
	dmCWMRDescribeCmd.Flags().StringVar(&flagDMCWFormat, "format", "", "Output format")
	dmCWMRCreateCmd.Flags().StringVar(&flagDMCWConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the MappingRule message body (required)")
	_ = dmCWMRCreateCmd.MarkFlagRequired("config-file")

	dmCWCmd.AddCommand(topCommands...)
	dmCWCmd.AddCommand(dmCWMappingRulesCmd)
	databaseMigrationCmd.AddCommand(dmCWCmd)
}

func dmCWResourceName(name, project, region string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/conversionWorkspaces/%s", project, region, name)
}

func dmCWMRResourceName(name, project, region, workspace string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	parent := dmCWResourceName(workspace, project, region)
	return fmt.Sprintf("%s/mappingRules/%s", parent, name)
}

func runDMCWCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cw := &datamigration.ConversionWorkspace{}
	if err := loadYAMLOrJSONInto(flagDMCWConfigFile, cw); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Create(dmParent(project, flagDMCWRegion), cw).
		ConversionWorkspaceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating conversion workspace: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Create", args[0], flagDMCWAsync)
}

func runDMCWDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Delete(dmCWResourceName(args[0], project, flagDMCWRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting conversion workspace: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Delete", args[0], flagDMCWAsync)
}

func runDMCWDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cw, err := svc.Projects.Locations.ConversionWorkspaces.Get(dmCWResourceName(args[0], project, flagDMCWRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing conversion workspace: %w", err)
	}
	return emitFormatted(cw, flagDMCWFormat)
}

func runDMCWUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cw := &datamigration.ConversionWorkspace{}
	if err := loadYAMLOrJSONInto(flagDMCWConfigFile, cw); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	mask := flagDMCWUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cw))
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Patch(dmCWResourceName(args[0], project, flagDMCWRegion), cw).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating conversion workspace: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Update", args[0], flagDMCWAsync)
}

func runDMCWList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmParent(project, flagDMCWRegion)
	var all []*datamigration.ConversionWorkspace
	pageToken := ""
	for {
		call := svc.Projects.Locations.ConversionWorkspaces.List(parent).Context(ctx)
		if flagDMCWListFilter != "" {
			call = call.Filter(flagDMCWListFilter)
		}
		if flagDMCWListPageSize > 0 {
			call = call.PageSize(flagDMCWListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing conversion workspaces: %w", err)
		}
		all = append(all, resp.ConversionWorkspaces...)
		if flagDMCWListLimit > 0 && int64(len(all)) >= flagDMCWListLimit {
			all = all[:flagDMCWListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDMCWListURI {
		for _, c := range all {
			fmt.Println(c.Name)
		}
		return nil
	}
	if flagDMCWFormat != "" {
		return emitFormatted(all, flagDMCWFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, c := range all {
		fmt.Printf("%-40s %s\n", path.Base(c.Name), c.DisplayName)
	}
	return nil
}

func runDMCWApply(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &datamigration.ApplyConversionWorkspaceRequest{
		AutoCommit:        flagDMCWAutoCommit,
		ConnectionProfile: flagDMCWConnProfile,
		DryRun:            flagDMCWDryRun,
		Filter:            flagDMCWFilter,
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Apply(dmCWResourceName(args[0], project, flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("applying conversion workspace: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Apply", args[0], flagDMCWAsync)
}

func runDMCWCommit(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &datamigration.CommitConversionWorkspaceRequest{CommitName: flagDMCWCommitName}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Commit(dmCWResourceName(args[0], project, flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing conversion workspace: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Commit", args[0], flagDMCWAsync)
}

func runDMCWConvert(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &datamigration.ConvertConversionWorkspaceRequest{
		AutoCommit:      flagDMCWAutoCommit,
		ConvertFullPath: flagDMCWConvertFullPath,
		Filter:          flagDMCWFilter,
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Convert(dmCWResourceName(args[0], project, flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("converting conversion workspace: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Convert", args[0], flagDMCWAsync)
}

func runDMCWRollback(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Rollback(
		dmCWResourceName(args[0], project, flagDMCWRegion),
		&datamigration.RollbackConversionWorkspaceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back conversion workspace: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Rollback", args[0], flagDMCWAsync)
}

func runDMCWSeed(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &datamigration.SeedConversionWorkspaceRequest{
		AutoCommit:                   flagDMCWAutoCommit,
		SourceConnectionProfile:      flagDMCWSourceConnProfile,
		DestinationConnectionProfile: flagDMCWDestConnProfile,
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Seed(dmCWResourceName(args[0], project, flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("seeding conversion workspace: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Seed", args[0], flagDMCWAsync)
}

func runDMCWDescribeDDLs(cmd *cobra.Command, args []string) error {
	return runDMCWDescribeEntitiesFiltered(args[0], "DDL")
}

func runDMCWDescribeIssues(cmd *cobra.Command, args []string) error {
	return runDMCWDescribeEntitiesFiltered(args[0], "ISSUES")
}

func runDMCWDescribeEntities(cmd *cobra.Command, args []string) error {
	return runDMCWDescribeEntitiesFiltered(args[0], "")
}

func runDMCWDescribeEntitiesFiltered(workspace, view string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ConversionWorkspaces.DescribeDatabaseEntities(dmCWResourceName(workspace, project, flagDMCWRegion)).Context(ctx)
	if flagDMCWDescribeCommitID != "" {
		call = call.CommitId(flagDMCWDescribeCommitID)
	}
	if flagDMCWDescribeTree != "" {
		call = call.Tree(flagDMCWDescribeTree)
	}
	if flagDMCWDescribeUncomm {
		call = call.Uncommitted(true)
	}
	if flagDMCWDescribeDBName != "" {
		call = call.Filter(fmt.Sprintf(`database="%s"`, flagDMCWDescribeDBName))
	}
	if view != "" {
		call = call.View(view)
	} else if flagDMCWDescribeView != "" {
		call = call.View(flagDMCWDescribeView)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing database entities: %w", err)
	}
	return emitFormatted(resp, flagDMCWFormat)
}

func runDMCWImportRules(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if len(flagDMCWRulesFiles) == 0 {
		return fmt.Errorf("at least one --file is required")
	}
	var rules []*datamigration.RulesFile
	for _, p := range flagDMCWRulesFiles {
		body, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("reading rules file %s: %w", p, err)
		}
		rules = append(rules, &datamigration.RulesFile{
			RulesSourceFilename: p,
			RulesContent:        string(body),
		})
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &datamigration.ImportMappingRulesRequest{
		AutoCommit:  flagDMCWAutoCommit,
		RulesFormat: flagDMCWRulesFormat,
		RulesFiles:  rules,
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.MappingRules.Import(dmCWResourceName(args[0], project, flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing mapping rules: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Import rules", args[0], flagDMCWAsync)
}

func runDMCWListBackgroundJobs(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ConversionWorkspaces.SearchBackgroundJobs(dmCWResourceName(args[0], project, flagDMCWRegion)).Context(ctx)
	if flagDMCWListLimit > 0 {
		call = call.MaxSize(flagDMCWListLimit)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing background jobs: %w", err)
	}
	return emitFormatted(resp, flagDMCWFormat)
}

// --- mapping-rules subcommands ---

func runDMCWMRList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmCWResourceName(flagDMCWWorkspace, project, flagDMCWRegion)
	var all []*datamigration.MappingRule
	pageToken := ""
	for {
		call := svc.Projects.Locations.ConversionWorkspaces.MappingRules.List(parent).Context(ctx)
		if flagDMCWListPageSize > 0 {
			call = call.PageSize(flagDMCWListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing mapping rules: %w", err)
		}
		all = append(all, resp.MappingRules...)
		if flagDMCWListLimit > 0 && int64(len(all)) >= flagDMCWListLimit {
			all = all[:flagDMCWListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDMCWFormat != "" {
		return emitFormatted(all, flagDMCWFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, r := range all {
		fmt.Printf("%-40s %s\n", path.Base(r.Name), r.State)
	}
	return nil
}

func runDMCWMRDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	r, err := svc.Projects.Locations.ConversionWorkspaces.MappingRules.Get(
		dmCWMRResourceName(args[0], project, flagDMCWRegion, flagDMCWWorkspace)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing mapping rule: %w", err)
	}
	return emitFormatted(r, flagDMCWFormat)
}

func runDMCWMRDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ConversionWorkspaces.MappingRules.Delete(
		dmCWMRResourceName(args[0], project, flagDMCWRegion, flagDMCWWorkspace)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting mapping rule: %w", err)
	}
	fmt.Printf("Deleted mapping rule [%s].\n", args[0])
	return nil
}

func runDMCWMRCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	rule := &datamigration.MappingRule{}
	if err := loadYAMLOrJSONInto(flagDMCWConfigFile, rule); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmCWResourceName(flagDMCWWorkspace, project, flagDMCWRegion)
	r, err := svc.Projects.Locations.ConversionWorkspaces.MappingRules.Create(parent, rule).
		MappingRuleId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating mapping rule: %w", err)
	}
	return emitFormatted(r, "")
}
