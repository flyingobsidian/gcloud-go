package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

func (s *databaseMigrationCmds) initDatabaseMigrationConversionWorkspaces() {
	// Top-level workspace commands.
	topCommands := []*cobra.Command{
		s.dmCWCreateCmd, s.dmCWDeleteCmd, s.dmCWDescribeCmd, s.dmCWUpdateCmd, s.dmCWListCmd,
		s.dmCWApplyCmd, s.dmCWCommitCmd, s.dmCWConvertCmd, s.dmCWRollbackCmd, s.dmCWSeedCmd,
		s.dmCWDescribeDDLsCmd, s.dmCWDescribeEntitiesCmd, s.dmCWDescribeIssuesCmd,
		s.dmCWImportRulesCmd, s.dmCWListBackgroundJobsCmd,
	}
	for _, c := range topCommands {
		c.Flags().StringVar(&s.flagDMCWRegion, "region", "", "Region containing the conversion workspace (required)")
		_ = c.MarkFlagRequired("region")
	}

	for _, c := range []*cobra.Command{s.dmCWCreateCmd, s.dmCWUpdateCmd} {
		c.Flags().StringVar(&s.flagDMCWConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the ConversionWorkspace message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dmCWUpdateCmd.Flags().StringVar(&s.flagDMCWUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	// LRO commands get --async.
	for _, c := range []*cobra.Command{
		s.dmCWCreateCmd, s.dmCWDeleteCmd, s.dmCWUpdateCmd, s.dmCWApplyCmd, s.dmCWCommitCmd,
		s.dmCWConvertCmd, s.dmCWRollbackCmd, s.dmCWSeedCmd, s.dmCWImportRulesCmd,
	} {
		c.Flags().BoolVar(&s.flagDMCWAsync, "async", false, "Return the long-running operation without waiting")
	}

	// apply / convert / seed common options.
	for _, c := range []*cobra.Command{s.dmCWApplyCmd, s.dmCWConvertCmd, s.dmCWSeedCmd} {
		c.Flags().StringVar(&s.flagDMCWFilter, "filter", "", "AIP-160 style filter for entities to include")
	}
	// apply keeps the historical opt-in default (false); a separate variable
	// prevents the convert/seed/import-rules default flip from leaking into it.
	s.dmCWApplyCmd.Flags().BoolVar(&s.flagDMCWApplyAutoCommit, "auto-commit", false,
		"Commit the workspace automatically after the operation")
	// convert / seed / import-rules default --auto-commit to true (matching
	// gcloud-python 578.0.0). Users can pass --no-auto-commit to opt out.
	for _, c := range []*cobra.Command{s.dmCWConvertCmd, s.dmCWSeedCmd, s.dmCWImportRulesCmd} {
		c.Flags().BoolVar(&s.flagDMCWAutoCommit, "auto-commit", true,
			"Commit the workspace automatically after the operation")
		c.Flags().BoolVar(&s.flagDMCWNoAutoCommit, "no-auto-commit", false,
			"Disable automatic commit (overrides --auto-commit)")
	}
	s.dmCWApplyCmd.Flags().BoolVar(&s.flagDMCWDryRun, "dry-run", false,
		"Validate the apply without changing the destination database")
	s.dmCWApplyCmd.Flags().StringVar(&s.flagDMCWConnProfile, "connection-profile", "",
		"Fully qualified name of the destination connection profile")
	s.dmCWConvertCmd.Flags().BoolVar(&s.flagDMCWConvertFullPath, "convert-full-path", false,
		"Automatically convert the full entity path for each entity matching --filter")

	s.dmCWSeedCmd.Flags().StringVar(&s.flagDMCWSourceConnProfile, "source-connection-profile", "",
		"Fully qualified name of the source connection profile")
	s.dmCWSeedCmd.Flags().StringVar(&s.flagDMCWDestConnProfile, "destination-connection-profile", "",
		"Fully qualified name of the destination connection profile")

	s.dmCWCommitCmd.Flags().StringVar(&s.flagDMCWCommitName, "commit-name", "", "Optional name for the commit")

	// describe-entities / describe-ddls / describe-issues shared filters.
	for _, c := range []*cobra.Command{s.dmCWDescribeDDLsCmd, s.dmCWDescribeEntitiesCmd, s.dmCWDescribeIssuesCmd} {
		c.Flags().StringVar(&s.flagDMCWDescribeCommitID, "commit-id", "", "Describe entities as of this commit")
		c.Flags().StringVar(&s.flagDMCWDescribeTree, "tree", "", "Which entity tree to walk (SOURCE, DRAFT, DESTINATION)")
		c.Flags().BoolVar(&s.flagDMCWDescribeUncomm, "uncommitted", false, "Include uncommitted changes")
		c.Flags().StringVar(&s.flagDMCWDescribeDBName, "database", "", "Restrict entities to this database")
		c.Flags().StringVar(&s.flagDMCWFormat, "format", "", "Output format")
	}
	s.dmCWDescribeEntitiesCmd.Flags().StringVar(&s.flagDMCWDescribeView, "view", "",
		"Which entity view to return (BASIC or FULL)")

	s.dmCWImportRulesCmd.Flags().StringSliceVar(&s.flagDMCWRulesFiles, "file", nil,
		"Path(s) to rules files (repeat to import multiple files)")
	s.dmCWImportRulesCmd.Flags().StringVar(&s.flagDMCWRulesFormat, "rules-format", "IMPORT_RULES_FILE_FORMAT_HARBOUR_BRIDGE_SESSION_FILE",
		"Format of the rules content file")
	// --auto-commit / --no-auto-commit for import-rules are registered above
	// alongside convert/seed.
	_ = s.dmCWImportRulesCmd.MarkFlagRequired("file")

	s.dmCWDescribeCmd.Flags().StringVar(&s.flagDMCWFormat, "format", "", "Output format")
	s.dmCWListCmd.Flags().StringVar(&s.flagDMCWFormat, "format", "", "Output format")
	s.dmCWListCmd.Flags().Int64Var(&s.flagDMCWListPageSize, "page-size", 0, "Page size")
	s.dmCWListCmd.Flags().Int64Var(&s.flagDMCWListLimit, "limit", 0, "Cap total results")
	s.dmCWListCmd.Flags().StringVar(&s.flagDMCWListFilter, "filter", "", "Server-side filter expression")
	s.dmCWListCmd.Flags().BoolVar(&s.flagDMCWListURI, "uri", false, "Print resource names only")

	s.dmCWListBackgroundJobsCmd.Flags().StringVar(&s.flagDMCWFormat, "format", "", "Output format")
	s.dmCWListBackgroundJobsCmd.Flags().Int64Var(&s.flagDMCWListLimit, "limit", 0, "Cap total results (0 = no cap)")

	s.dmCWMappingRulesCmd.AddCommand(s.dmCWMRCreateCmd, s.dmCWMRDeleteCmd, s.dmCWMRDescribeCmd, s.dmCWMRListCmd)
	for _, c := range s.dmCWMappingRulesCmd.Commands() {
		c.Flags().StringVar(&s.flagDMCWRegion, "region", "", "Region containing the conversion workspace (required)")
		c.Flags().StringVar(&s.flagDMCWWorkspace, "conversion-workspace", "",
			"Name of the parent conversion workspace (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("conversion-workspace")
	}
	s.dmCWMRListCmd.Flags().Int64Var(&s.flagDMCWListPageSize, "page-size", 0, "Page size")
	s.dmCWMRListCmd.Flags().Int64Var(&s.flagDMCWListLimit, "limit", 0, "Cap total results")
	s.dmCWMRListCmd.Flags().StringVar(&s.flagDMCWFormat, "format", "", "Output format")
	s.dmCWMRDescribeCmd.Flags().StringVar(&s.flagDMCWFormat, "format", "", "Output format")
	s.dmCWMRCreateCmd.Flags().StringVar(&s.flagDMCWConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the MappingRule message body (required)")
	_ = s.dmCWMRCreateCmd.MarkFlagRequired("config-file")

	s.dmCWCmd.AddCommand(topCommands...)
	s.dmCWCmd.AddCommand(s.dmCWMappingRulesCmd)
	s.databaseMigrationCmd.AddCommand(s.dmCWCmd)
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

// dmCWAutoCommit returns the effective auto-commit setting for convert, seed,
// and import-rules. --no-auto-commit wins over --auto-commit so scripts can
// opt back into the historical false default even though gcloud-python 578.0.0
// flipped the on-by-default behaviour.
func (s *databaseMigrationCmds) dmCWAutoCommit() bool {
	if s.flagDMCWNoAutoCommit {
		return false
	}
	return s.flagDMCWAutoCommit
}

func (s *databaseMigrationCmds) runDMCWCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cw := &datamigration.ConversionWorkspace{}
	if err := loadYAMLOrJSONInto(s.flagDMCWConfigFile, cw); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Create(dmParent(project, s.flagDMCWRegion), cw).
		ConversionWorkspaceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating conversion workspace: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Create", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Delete(dmCWResourceName(args[0], project, s.flagDMCWRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting conversion workspace: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Delete", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	cw, err := svc.Projects.Locations.ConversionWorkspaces.Get(dmCWResourceName(args[0], project, s.flagDMCWRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing conversion workspace: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, cw, s.flagDMCWFormat)
}

func (s *databaseMigrationCmds) runDMCWUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cw := &datamigration.ConversionWorkspace{}
	if err := loadYAMLOrJSONInto(s.flagDMCWConfigFile, cw); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	mask := s.flagDMCWUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cw))
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Patch(dmCWResourceName(args[0], project, s.flagDMCWRegion), cw).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating conversion workspace: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Update", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmParent(project, s.flagDMCWRegion)
	var all []*datamigration.ConversionWorkspace
	pageToken := ""
	for {
		call := svc.Projects.Locations.ConversionWorkspaces.List(parent).Context(ctx)
		if s.flagDMCWListFilter != "" {
			call = call.Filter(s.flagDMCWListFilter)
		}
		if s.flagDMCWListPageSize > 0 {
			call = call.PageSize(s.flagDMCWListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing conversion workspaces: %w", err)
		}
		all = append(all, resp.ConversionWorkspaces...)
		if s.flagDMCWListLimit > 0 && int64(len(all)) >= s.flagDMCWListLimit {
			all = all[:s.flagDMCWListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagDMCWListURI {
		for _, c := range all {
			fmt.Fprintln(s.inv.stdout, c.Name)
		}
		return nil
	}
	if s.flagDMCWFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDMCWFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, c := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(c.Name), c.DisplayName)
	}
	return nil
}

func (s *databaseMigrationCmds) runDMCWApply(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	req := &datamigration.ApplyConversionWorkspaceRequest{
		AutoCommit:        s.flagDMCWApplyAutoCommit,
		ConnectionProfile: s.flagDMCWConnProfile,
		DryRun:            s.flagDMCWDryRun,
		Filter:            s.flagDMCWFilter,
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Apply(dmCWResourceName(args[0], project, s.flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("applying conversion workspace: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Apply", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWCommit(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	req := &datamigration.CommitConversionWorkspaceRequest{CommitName: s.flagDMCWCommitName}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Commit(dmCWResourceName(args[0], project, s.flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing conversion workspace: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Commit", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWConvert(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	req := &datamigration.ConvertConversionWorkspaceRequest{
		AutoCommit:      s.dmCWAutoCommit(),
		ConvertFullPath: s.flagDMCWConvertFullPath,
		Filter:          s.flagDMCWFilter,
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Convert(dmCWResourceName(args[0], project, s.flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("converting conversion workspace: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Convert", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWRollback(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Rollback(
		dmCWResourceName(args[0], project, s.flagDMCWRegion),
		&datamigration.RollbackConversionWorkspaceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back conversion workspace: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Rollback", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWSeed(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	req := &datamigration.SeedConversionWorkspaceRequest{
		AutoCommit:                   s.dmCWAutoCommit(),
		SourceConnectionProfile:      s.flagDMCWSourceConnProfile,
		DestinationConnectionProfile: s.flagDMCWDestConnProfile,
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.Seed(dmCWResourceName(args[0], project, s.flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("seeding conversion workspace: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Seed", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWDescribeDDLs(cmd *cobra.Command, args []string) error {
	return s.runDMCWDescribeEntitiesFiltered(args[0], "DDL")
}

func (s *databaseMigrationCmds) runDMCWDescribeIssues(cmd *cobra.Command, args []string) error {
	return s.runDMCWDescribeEntitiesFiltered(args[0], "ISSUES")
}

func (s *databaseMigrationCmds) runDMCWDescribeEntities(cmd *cobra.Command, args []string) error {
	return s.runDMCWDescribeEntitiesFiltered(args[0], "")
}

func (s *databaseMigrationCmds) runDMCWDescribeEntitiesFiltered(workspace, view string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ConversionWorkspaces.DescribeDatabaseEntities(dmCWResourceName(workspace, project, s.flagDMCWRegion)).Context(ctx)
	if s.flagDMCWDescribeCommitID != "" {
		call = call.CommitId(s.flagDMCWDescribeCommitID)
	}
	if s.flagDMCWDescribeTree != "" {
		call = call.Tree(s.flagDMCWDescribeTree)
	}
	if s.flagDMCWDescribeUncomm {
		call = call.Uncommitted(true)
	}
	if s.flagDMCWDescribeDBName != "" {
		call = call.Filter(fmt.Sprintf(`database="%s"`, s.flagDMCWDescribeDBName))
	}
	if view != "" {
		call = call.View(view)
	} else if s.flagDMCWDescribeView != "" {
		call = call.View(s.flagDMCWDescribeView)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing database entities: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagDMCWFormat)
}

func (s *databaseMigrationCmds) runDMCWImportRules(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if len(s.flagDMCWRulesFiles) == 0 {
		return fmt.Errorf("at least one --file is required")
	}
	var rules []*datamigration.RulesFile
	for _, p := range s.flagDMCWRulesFiles {
		body, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("reading rules file %s: %w", p, err)
		}
		rules = append(rules, &datamigration.RulesFile{
			RulesSourceFilename: p,
			RulesContent:        string(body),
		})
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	req := &datamigration.ImportMappingRulesRequest{
		AutoCommit:  s.dmCWAutoCommit(),
		RulesFormat: s.flagDMCWRulesFormat,
		RulesFiles:  rules,
	}
	op, err := svc.Projects.Locations.ConversionWorkspaces.MappingRules.Import(dmCWResourceName(args[0], project, s.flagDMCWRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing mapping rules: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Import rules", args[0], s.flagDMCWAsync)
}

func (s *databaseMigrationCmds) runDMCWListBackgroundJobs(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ConversionWorkspaces.SearchBackgroundJobs(dmCWResourceName(args[0], project, s.flagDMCWRegion)).Context(ctx)
	if s.flagDMCWListLimit > 0 {
		call = call.MaxSize(s.flagDMCWListLimit)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing background jobs: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagDMCWFormat)
}

// --- mapping-rules subcommands ---

func (s *databaseMigrationCmds) runDMCWMRList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmCWResourceName(s.flagDMCWWorkspace, project, s.flagDMCWRegion)
	var all []*datamigration.MappingRule
	pageToken := ""
	for {
		call := svc.Projects.Locations.ConversionWorkspaces.MappingRules.List(parent).Context(ctx)
		if s.flagDMCWListPageSize > 0 {
			call = call.PageSize(s.flagDMCWListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing mapping rules: %w", err)
		}
		all = append(all, resp.MappingRules...)
		if s.flagDMCWListLimit > 0 && int64(len(all)) >= s.flagDMCWListLimit {
			all = all[:s.flagDMCWListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagDMCWFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDMCWFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "STATE")
	for _, r := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(r.Name), r.State)
	}
	return nil
}

func (s *databaseMigrationCmds) runDMCWMRDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	r, err := svc.Projects.Locations.ConversionWorkspaces.MappingRules.Get(
		dmCWMRResourceName(args[0], project, s.flagDMCWRegion, s.flagDMCWWorkspace)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing mapping rule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, r, s.flagDMCWFormat)
}

func (s *databaseMigrationCmds) runDMCWMRDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ConversionWorkspaces.MappingRules.Delete(
		dmCWMRResourceName(args[0], project, s.flagDMCWRegion, s.flagDMCWWorkspace)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting mapping rule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted mapping rule [%s].\n", args[0])
	return nil
}

func (s *databaseMigrationCmds) runDMCWMRCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	rule := &datamigration.MappingRule{}
	if err := loadYAMLOrJSONInto(s.flagDMCWConfigFile, rule); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmCWResourceName(s.flagDMCWWorkspace, project, s.flagDMCWRegion)
	r, err := svc.Projects.Locations.ConversionWorkspaces.MappingRules.Create(parent, rule).
		MappingRuleId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating mapping rule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, r, "")
}
