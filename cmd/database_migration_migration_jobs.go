package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

func (s *databaseMigrationCmds) initDatabaseMigrationMigrationJobs() {
	all := []*cobra.Command{
		s.dmMJCreateCmd, s.dmMJDeleteCmd, s.dmMJDescribeCmd, s.dmMJUpdateCmd, s.dmMJListCmd,
		s.dmMJStartCmd, s.dmMJStopCmd, s.dmMJResumeCmd, s.dmMJRestartCmd, s.dmMJPromoteCmd,
		s.dmMJVerifyCmd, s.dmMJDemoteDestCmd, s.dmMJGenerateSSHCmd, s.dmMJFetchObjectsCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDMMJRegion, "region", "", "Region containing the migration job (required)")
		_ = c.MarkFlagRequired("region")
	}

	// --config-file for create/update; --update-mask for update.
	for _, c := range []*cobra.Command{s.dmMJCreateCmd, s.dmMJUpdateCmd} {
		c.Flags().StringVar(&s.flagDMMJConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the MigrationJob message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dmMJUpdateCmd.Flags().StringVar(&s.flagDMMJUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	// --async on operations that return an LRO.
	for _, c := range []*cobra.Command{
		s.dmMJCreateCmd, s.dmMJDeleteCmd, s.dmMJUpdateCmd, s.dmMJStartCmd, s.dmMJStopCmd,
		s.dmMJResumeCmd, s.dmMJRestartCmd, s.dmMJPromoteCmd, s.dmMJVerifyCmd, s.dmMJDemoteDestCmd,
	} {
		c.Flags().BoolVar(&s.flagDMMJAsync, "async", false, "Return the long-running operation without waiting")
	}

	for _, c := range []*cobra.Command{s.dmMJStartCmd, s.dmMJResumeCmd, s.dmMJRestartCmd} {
		c.Flags().BoolVar(&s.flagDMMJSkipValidate, "skip-validation", false,
			"Skip pre-flight validation before performing the action")
	}
	s.dmMJRestartCmd.Flags().BoolVar(&s.flagDMMJRestartFailed, "restart-failed-objects", false,
		"Only restart failed objects")

	s.dmMJVerifyCmd.Flags().BoolVar(&s.flagDMMJDryRun, "dry-run", false,
		"Verify only; do not persist changes")

	s.dmMJGenerateSSHCmd.Flags().StringVar(&s.flagDMMJSSHVM, "vm", "",
		"Name of the bastion VM to use or create (required)")
	s.dmMJGenerateSSHCmd.Flags().Int64Var(&s.flagDMMJSSHVMPort, "vm-port", 22,
		"SSH port to open on the bastion VM")
	_ = s.dmMJGenerateSSHCmd.MarkFlagRequired("vm")

	s.dmMJDescribeCmd.Flags().StringVar(&s.flagDMMJFormat, "format", "", "Output format")
	s.dmMJListCmd.Flags().StringVar(&s.flagDMMJFormat, "format", "", "Output format")
	s.dmMJListCmd.Flags().Int64Var(&s.flagDMMJListPageSize, "page-size", 0, "Page size for API pagination")
	s.dmMJListCmd.Flags().Int64Var(&s.flagDMMJListLimit, "limit", 0, "Cap total results (0 = no cap)")
	s.dmMJListCmd.Flags().StringVar(&s.flagDMMJListFilter, "filter", "", "Server-side filter expression")
	s.dmMJListCmd.Flags().BoolVar(&s.flagDMMJListURI, "uri", false, "Print resource names only")

	s.dmMJCmd.AddCommand(all...)
	s.databaseMigrationCmd.AddCommand(s.dmMJCmd)
}

func dmMJResourceName(name, project, region string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/migrationJobs/%s", project, region, name)
}

func (s *databaseMigrationCmds) runDMMJDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	mj, err := svc.Projects.Locations.MigrationJobs.Get(dmMJResourceName(args[0], project, s.flagDMMJRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing migration job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, mj, s.flagDMMJFormat)
}

func (s *databaseMigrationCmds) runDMMJList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmParent(project, s.flagDMMJRegion)
	var all []*datamigration.MigrationJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.MigrationJobs.List(parent).Context(ctx)
		if s.flagDMMJListFilter != "" {
			call = call.Filter(s.flagDMMJListFilter)
		}
		if s.flagDMMJListPageSize > 0 {
			call = call.PageSize(s.flagDMMJListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing migration jobs: %w", err)
		}
		all = append(all, resp.MigrationJobs...)
		if s.flagDMMJListLimit > 0 && int64(len(all)) >= s.flagDMMJListLimit {
			all = all[:s.flagDMMJListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagDMMJListURI {
		for _, m := range all {
			fmt.Fprintln(s.inv.stdout, m.Name)
		}
		return nil
	}
	if s.flagDMMJFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDMMJFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-15s %s\n", "NAME", "STATE", "TYPE")
	for _, m := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-15s %s\n", path.Base(m.Name), m.State, m.Type)
	}
	return nil
}

func (s *databaseMigrationCmds) runDMMJCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	mj := &datamigration.MigrationJob{}
	if err := loadYAMLOrJSONInto(s.flagDMMJConfigFile, mj); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MigrationJobs.Create(dmParent(project, s.flagDMMJRegion), mj).
		MigrationJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating migration job: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Create", args[0], s.flagDMMJAsync)
}

func (s *databaseMigrationCmds) runDMMJDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MigrationJobs.Delete(dmMJResourceName(args[0], project, s.flagDMMJRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting migration job: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Delete", args[0], s.flagDMMJAsync)
}

func (s *databaseMigrationCmds) runDMMJUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	mj := &datamigration.MigrationJob{}
	if err := loadYAMLOrJSONInto(s.flagDMMJConfigFile, mj); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	mask := s.flagDMMJUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(mj))
	}
	op, err := svc.Projects.Locations.MigrationJobs.Patch(dmMJResourceName(args[0], project, s.flagDMMJRegion), mj).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating migration job: %w", err)
	}
	return s.dmFinishOp(ctx, svc, op, "Update", args[0], s.flagDMMJAsync)
}

func (s *databaseMigrationCmds) runDMMJStart(cmd *cobra.Command, args []string) error {
	return s.dmMJInvokeAction(args[0], "Start", s.flagDMMJAsync, func(ctx context.Context, svc *datamigration.Service, name string) (*datamigration.Operation, error) {
		req := &datamigration.StartMigrationJobRequest{SkipValidation: s.flagDMMJSkipValidate}
		return svc.Projects.Locations.MigrationJobs.Start(name, req).Context(ctx).Do()
	})
}

func (s *databaseMigrationCmds) runDMMJStop(cmd *cobra.Command, args []string) error {
	return s.dmMJInvokeAction(args[0], "Stop", s.flagDMMJAsync, func(ctx context.Context, svc *datamigration.Service, name string) (*datamigration.Operation, error) {
		return svc.Projects.Locations.MigrationJobs.Stop(name, &datamigration.StopMigrationJobRequest{}).Context(ctx).Do()
	})
}

func (s *databaseMigrationCmds) runDMMJResume(cmd *cobra.Command, args []string) error {
	return s.dmMJInvokeAction(args[0], "Resume", s.flagDMMJAsync, func(ctx context.Context, svc *datamigration.Service, name string) (*datamigration.Operation, error) {
		req := &datamigration.ResumeMigrationJobRequest{SkipValidation: s.flagDMMJSkipValidate}
		return svc.Projects.Locations.MigrationJobs.Resume(name, req).Context(ctx).Do()
	})
}

func (s *databaseMigrationCmds) runDMMJRestart(cmd *cobra.Command, args []string) error {
	return s.dmMJInvokeAction(args[0], "Restart", s.flagDMMJAsync, func(ctx context.Context, svc *datamigration.Service, name string) (*datamigration.Operation, error) {
		req := &datamigration.RestartMigrationJobRequest{
			SkipValidation:       s.flagDMMJSkipValidate,
			RestartFailedObjects: s.flagDMMJRestartFailed,
		}
		return svc.Projects.Locations.MigrationJobs.Restart(name, req).Context(ctx).Do()
	})
}

func (s *databaseMigrationCmds) runDMMJPromote(cmd *cobra.Command, args []string) error {
	return s.dmMJInvokeAction(args[0], "Promote", s.flagDMMJAsync, func(ctx context.Context, svc *datamigration.Service, name string) (*datamigration.Operation, error) {
		return svc.Projects.Locations.MigrationJobs.Promote(name, &datamigration.PromoteMigrationJobRequest{}).Context(ctx).Do()
	})
}

func (s *databaseMigrationCmds) runDMMJVerify(cmd *cobra.Command, args []string) error {
	return s.dmMJInvokeAction(args[0], "Verify", s.flagDMMJAsync, func(ctx context.Context, svc *datamigration.Service, name string) (*datamigration.Operation, error) {
		return svc.Projects.Locations.MigrationJobs.Verify(name, &datamigration.VerifyMigrationJobRequest{}).Context(ctx).Do()
	})
}

func (s *databaseMigrationCmds) runDMMJDemoteDestination(cmd *cobra.Command, args []string) error {
	return s.dmMJInvokeAction(args[0], "Demote destination", s.flagDMMJAsync, func(ctx context.Context, svc *datamigration.Service, name string) (*datamigration.Operation, error) {
		return svc.Projects.Locations.MigrationJobs.DemoteDestination(name, &datamigration.DemoteDestinationRequest{}).Context(ctx).Do()
	})
}

func (s *databaseMigrationCmds) runDMMJGenerateSSHScript(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	req := &datamigration.GenerateSshScriptRequest{
		Vm:     s.flagDMMJSSHVM,
		VmPort: s.flagDMMJSSHVMPort,
	}
	resp, err := svc.Projects.Locations.MigrationJobs.GenerateSshScript(dmMJResourceName(args[0], project, s.flagDMMJRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("generating SSH script: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, resp.Script)
	return nil
}

func (s *databaseMigrationCmds) runDMMJFetchSourceObjects(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MigrationJobs.FetchSourceObjects(dmMJResourceName(args[0], project, s.flagDMMJRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching source objects: %w", err)
	}
	if s.flagDMMJAsync {
		fmt.Fprintf(s.inv.stderr, "Fetch in progress (operation: %s).\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, op.Response, "")
}

// dmMJInvokeAction wraps the common "resolve project → build service → run
// action call that returns an Operation → optionally wait" pattern used by
// start/stop/resume/restart/promote/verify/demote actions.
func (s *databaseMigrationCmds) dmMJInvokeAction(name, verb string, async bool, action func(ctx context.Context, svc *datamigration.Service, resourceName string) (*datamigration.Operation, error)) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := action(ctx, svc, dmMJResourceName(name, project, s.flagDMMJRegion))
	if err != nil {
		return fmt.Errorf("%s migration job: %w", strings.ToLower(verb), err)
	}
	return s.dmFinishOp(ctx, svc, op, verb, name, async)
}

// dmFinishOp either prints the operation name (when async) or blocks until it
// completes, then prints a short success message on stderr.
func (s *databaseMigrationCmds) dmFinishOp(ctx context.Context, svc *datamigration.Service, op *datamigration.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	op, err := waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s migration job [%s] completed.\n", verb, name)
	if op.Response != nil {
		return emitFormattedTo(s.inv.stdout, op.Response, "")
	}
	return nil
}
