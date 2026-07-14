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

var dmMJCmd = &cobra.Command{
	Use:   "migration-jobs",
	Short: "Manage Database Migration Service migration jobs",
}

var (
	dmMJCreateCmd = &cobra.Command{
		Use: "create JOB", Short: "Create a migration job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDMMJCreate,
	}
	dmMJDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJDelete,
	}
	dmMJDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Show details about a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJDescribe,
	}
	dmMJUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update a migration job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDMMJUpdate,
	}
	dmMJListCmd = &cobra.Command{
		Use: "list", Short: "List migration jobs in a region",
		Args: cobra.NoArgs, RunE: runDMMJList,
	}
	dmMJStartCmd = &cobra.Command{
		Use: "start JOB", Short: "Start a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJStart,
	}
	dmMJStopCmd = &cobra.Command{
		Use: "stop JOB", Short: "Stop a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJStop,
	}
	dmMJResumeCmd = &cobra.Command{
		Use: "resume JOB", Short: "Resume a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJResume,
	}
	dmMJRestartCmd = &cobra.Command{
		Use: "restart JOB", Short: "Restart a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJRestart,
	}
	dmMJPromoteCmd = &cobra.Command{
		Use: "promote JOB", Short: "Promote a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJPromote,
	}
	dmMJVerifyCmd = &cobra.Command{
		Use: "verify JOB", Short: "Verify a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJVerify,
	}
	dmMJDemoteDestCmd = &cobra.Command{
		Use: "demote-destination JOB", Short: "Demote the destination of a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJDemoteDestination,
	}
	dmMJGenerateSSHCmd = &cobra.Command{
		Use: "generate-ssh-script JOB", Short: "Generate the SSH script for a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJGenerateSSHScript,
	}
	dmMJFetchObjectsCmd = &cobra.Command{
		Use: "fetch-source-objects JOB", Short: "Fetch the source database objects of a migration job",
		Args: cobra.ExactArgs(1), RunE: runDMMJFetchSourceObjects,
	}
)

var (
	flagDMMJRegion         string
	flagDMMJFormat         string
	flagDMMJConfigFile     string
	flagDMMJUpdateMask     string
	flagDMMJListPageSize   int64
	flagDMMJListLimit      int64
	flagDMMJListFilter     string
	flagDMMJListURI        bool
	flagDMMJAsync          bool
	flagDMMJSkipValidate   bool
	flagDMMJRestartFailed  bool
	flagDMMJRestartSkipVal bool
	flagDMMJDryRun         bool
	flagDMMJSSHVM          string
	flagDMMJSSHVMPort      int64
)

func init() {
	all := []*cobra.Command{
		dmMJCreateCmd, dmMJDeleteCmd, dmMJDescribeCmd, dmMJUpdateCmd, dmMJListCmd,
		dmMJStartCmd, dmMJStopCmd, dmMJResumeCmd, dmMJRestartCmd, dmMJPromoteCmd,
		dmMJVerifyCmd, dmMJDemoteDestCmd, dmMJGenerateSSHCmd, dmMJFetchObjectsCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDMMJRegion, "region", "", "Region containing the migration job (required)")
		_ = c.MarkFlagRequired("region")
	}

	// --config-file for create/update; --update-mask for update.
	for _, c := range []*cobra.Command{dmMJCreateCmd, dmMJUpdateCmd} {
		c.Flags().StringVar(&flagDMMJConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the MigrationJob message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dmMJUpdateCmd.Flags().StringVar(&flagDMMJUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	// --async on operations that return an LRO.
	for _, c := range []*cobra.Command{
		dmMJCreateCmd, dmMJDeleteCmd, dmMJUpdateCmd, dmMJStartCmd, dmMJStopCmd,
		dmMJResumeCmd, dmMJRestartCmd, dmMJPromoteCmd, dmMJVerifyCmd, dmMJDemoteDestCmd,
	} {
		c.Flags().BoolVar(&flagDMMJAsync, "async", false, "Return the long-running operation without waiting")
	}

	for _, c := range []*cobra.Command{dmMJStartCmd, dmMJResumeCmd, dmMJRestartCmd} {
		c.Flags().BoolVar(&flagDMMJSkipValidate, "skip-validation", false,
			"Skip pre-flight validation before performing the action")
	}
	dmMJRestartCmd.Flags().BoolVar(&flagDMMJRestartFailed, "restart-failed-objects", false,
		"Only restart failed objects")

	dmMJVerifyCmd.Flags().BoolVar(&flagDMMJDryRun, "dry-run", false,
		"Verify only; do not persist changes")

	dmMJGenerateSSHCmd.Flags().StringVar(&flagDMMJSSHVM, "vm", "",
		"Name of the bastion VM to use or create (required)")
	dmMJGenerateSSHCmd.Flags().Int64Var(&flagDMMJSSHVMPort, "vm-port", 22,
		"SSH port to open on the bastion VM")
	_ = dmMJGenerateSSHCmd.MarkFlagRequired("vm")

	dmMJDescribeCmd.Flags().StringVar(&flagDMMJFormat, "format", "", "Output format")
	dmMJListCmd.Flags().StringVar(&flagDMMJFormat, "format", "", "Output format")
	dmMJListCmd.Flags().Int64Var(&flagDMMJListPageSize, "page-size", 0, "Page size for API pagination")
	dmMJListCmd.Flags().Int64Var(&flagDMMJListLimit, "limit", 0, "Cap total results (0 = no cap)")
	dmMJListCmd.Flags().StringVar(&flagDMMJListFilter, "filter", "", "Server-side filter expression")
	dmMJListCmd.Flags().BoolVar(&flagDMMJListURI, "uri", false, "Print resource names only")

	dmMJCmd.AddCommand(all...)
	databaseMigrationCmd.AddCommand(dmMJCmd)
}

func dmMJResourceName(name, project, region string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/migrationJobs/%s", project, region, name)
}

func runDMMJDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	mj, err := svc.Projects.Locations.MigrationJobs.Get(dmMJResourceName(args[0], project, flagDMMJRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing migration job: %w", err)
	}
	return emitFormatted(mj, flagDMMJFormat)
}

func runDMMJList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmParent(project, flagDMMJRegion)
	var all []*datamigration.MigrationJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.MigrationJobs.List(parent).Context(ctx)
		if flagDMMJListFilter != "" {
			call = call.Filter(flagDMMJListFilter)
		}
		if flagDMMJListPageSize > 0 {
			call = call.PageSize(flagDMMJListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing migration jobs: %w", err)
		}
		all = append(all, resp.MigrationJobs...)
		if flagDMMJListLimit > 0 && int64(len(all)) >= flagDMMJListLimit {
			all = all[:flagDMMJListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDMMJListURI {
		for _, m := range all {
			fmt.Println(m.Name)
		}
		return nil
	}
	if flagDMMJFormat != "" {
		return emitFormatted(all, flagDMMJFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "TYPE")
	for _, m := range all {
		fmt.Printf("%-40s %-15s %s\n", path.Base(m.Name), m.State, m.Type)
	}
	return nil
}

func runDMMJCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	mj := &datamigration.MigrationJob{}
	if err := loadYAMLOrJSONInto(flagDMMJConfigFile, mj); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MigrationJobs.Create(dmParent(project, flagDMMJRegion), mj).
		MigrationJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating migration job: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Create", args[0], flagDMMJAsync)
}

func runDMMJDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MigrationJobs.Delete(dmMJResourceName(args[0], project, flagDMMJRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting migration job: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Delete", args[0], flagDMMJAsync)
}

func runDMMJUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	mj := &datamigration.MigrationJob{}
	if err := loadYAMLOrJSONInto(flagDMMJConfigFile, mj); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	mask := flagDMMJUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(mj))
	}
	op, err := svc.Projects.Locations.MigrationJobs.Patch(dmMJResourceName(args[0], project, flagDMMJRegion), mj).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating migration job: %w", err)
	}
	return dmFinishOp(ctx, svc, op, "Update", args[0], flagDMMJAsync)
}

func runDMMJStart(cmd *cobra.Command, args []string) error {
	return dmMJInvokeAction(args[0], "Start", flagDMMJAsync, func(svc *datamigration.Service, name string, ctx context.Context) (*datamigration.Operation, error) {
		req := &datamigration.StartMigrationJobRequest{SkipValidation: flagDMMJSkipValidate}
		return svc.Projects.Locations.MigrationJobs.Start(name, req).Context(ctx).Do()
	})
}

func runDMMJStop(cmd *cobra.Command, args []string) error {
	return dmMJInvokeAction(args[0], "Stop", flagDMMJAsync, func(svc *datamigration.Service, name string, ctx context.Context) (*datamigration.Operation, error) {
		return svc.Projects.Locations.MigrationJobs.Stop(name, &datamigration.StopMigrationJobRequest{}).Context(ctx).Do()
	})
}

func runDMMJResume(cmd *cobra.Command, args []string) error {
	return dmMJInvokeAction(args[0], "Resume", flagDMMJAsync, func(svc *datamigration.Service, name string, ctx context.Context) (*datamigration.Operation, error) {
		req := &datamigration.ResumeMigrationJobRequest{SkipValidation: flagDMMJSkipValidate}
		return svc.Projects.Locations.MigrationJobs.Resume(name, req).Context(ctx).Do()
	})
}

func runDMMJRestart(cmd *cobra.Command, args []string) error {
	return dmMJInvokeAction(args[0], "Restart", flagDMMJAsync, func(svc *datamigration.Service, name string, ctx context.Context) (*datamigration.Operation, error) {
		req := &datamigration.RestartMigrationJobRequest{
			SkipValidation:       flagDMMJSkipValidate,
			RestartFailedObjects: flagDMMJRestartFailed,
		}
		return svc.Projects.Locations.MigrationJobs.Restart(name, req).Context(ctx).Do()
	})
}

func runDMMJPromote(cmd *cobra.Command, args []string) error {
	return dmMJInvokeAction(args[0], "Promote", flagDMMJAsync, func(svc *datamigration.Service, name string, ctx context.Context) (*datamigration.Operation, error) {
		return svc.Projects.Locations.MigrationJobs.Promote(name, &datamigration.PromoteMigrationJobRequest{}).Context(ctx).Do()
	})
}

func runDMMJVerify(cmd *cobra.Command, args []string) error {
	return dmMJInvokeAction(args[0], "Verify", flagDMMJAsync, func(svc *datamigration.Service, name string, ctx context.Context) (*datamigration.Operation, error) {
		return svc.Projects.Locations.MigrationJobs.Verify(name, &datamigration.VerifyMigrationJobRequest{}).Context(ctx).Do()
	})
}

func runDMMJDemoteDestination(cmd *cobra.Command, args []string) error {
	return dmMJInvokeAction(args[0], "Demote destination", flagDMMJAsync, func(svc *datamigration.Service, name string, ctx context.Context) (*datamigration.Operation, error) {
		return svc.Projects.Locations.MigrationJobs.DemoteDestination(name, &datamigration.DemoteDestinationRequest{}).Context(ctx).Do()
	})
}

func runDMMJGenerateSSHScript(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &datamigration.GenerateSshScriptRequest{
		Vm:     flagDMMJSSHVM,
		VmPort: flagDMMJSSHVMPort,
	}
	resp, err := svc.Projects.Locations.MigrationJobs.GenerateSshScript(dmMJResourceName(args[0], project, flagDMMJRegion), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("generating SSH script: %w", err)
	}
	fmt.Println(resp.Script)
	return nil
}

func runDMMJFetchSourceObjects(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MigrationJobs.FetchSourceObjects(dmMJResourceName(args[0], project, flagDMMJRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching source objects: %w", err)
	}
	if flagDMMJAsync {
		fmt.Fprintf(os.Stderr, "Fetch in progress (operation: %s).\n", op.Name)
		return emitFormatted(op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(op.Response, "")
}

// dmMJInvokeAction wraps the common "resolve project → build service → run
// action call that returns an Operation → optionally wait" pattern used by
// start/stop/resume/restart/promote/verify/demote actions.
func dmMJInvokeAction(name, verb string, async bool, action func(svc *datamigration.Service, resourceName string, ctx context.Context) (*datamigration.Operation, error)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := action(svc, dmMJResourceName(name, project, flagDMMJRegion), ctx)
	if err != nil {
		return fmt.Errorf("%s migration job: %w", strings.ToLower(verb), err)
	}
	return dmFinishOp(ctx, svc, op, verb, name, async)
}

// dmFinishOp either prints the operation name (when async) or blocks until it
// completes, then prints a short success message on stderr.
func dmFinishOp(ctx context.Context, svc *datamigration.Service, op *datamigration.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	op, err := waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s migration job [%s] completed.\n", verb, name)
	if op.Response != nil {
		return emitFormatted(op.Response, "")
	}
	return nil
}
