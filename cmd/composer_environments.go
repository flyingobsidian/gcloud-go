package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	composer "google.golang.org/api/composer/v1"
)

// --- gcloud composer environments (#1502) ---

var composerEnvCmd = &cobra.Command{Use: "environments", Short: "Manage Composer environments"}

var (
	flagComposerEnvLocation      string
	flagComposerEnvFormat        string
	flagComposerEnvConfigFile    string
	flagComposerEnvUpdateMask    string
	flagComposerEnvFilter        string
	flagComposerEnvPageSize      int64
	flagComposerEnvImageVersion  string
	flagComposerEnvSnapLocation  string
	flagComposerEnvSnapPath      string
	flagComposerEnvSkipOverrides bool
	flagComposerEnvSkipEnv       bool
	flagComposerEnvSkipData      bool
	flagComposerEnvSkipPypi      bool
	flagComposerEnvRunCommand    string
	flagComposerEnvRunSubcommand string
	flagComposerEnvRunParams     []string
	flagComposerEnvRunTimeout    time.Duration
	flagComposerEnvRunTree       bool
)

var (
	composerEnvCreateCmd = &cobra.Command{
		Use: "create ENVIRONMENT", Short: "Create a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvCreate,
	}
	composerEnvDeleteCmd = &cobra.Command{
		Use: "delete ENVIRONMENT", Short: "Delete a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvDelete,
	}
	composerEnvDescribeCmd = &cobra.Command{
		Use: "describe ENVIRONMENT", Short: "Describe a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvDescribe,
	}
	composerEnvListCmd = &cobra.Command{
		Use: "list", Short: "List Composer environments",
		Args: cobra.NoArgs, RunE: runComposerEnvList,
	}
	composerEnvUpdateCmd = &cobra.Command{
		Use: "update ENVIRONMENT", Short: "Update a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvUpdate,
	}
	composerEnvCheckUpgradeCmd = &cobra.Command{
		Use: "check-upgrade ENVIRONMENT", Short: "Check whether the environment can be upgraded to an image version",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvCheckUpgrade,
	}
	composerEnvDatabaseFailoverCmd = &cobra.Command{
		Use: "database-failover ENVIRONMENT", Short: "Trigger a Cloud SQL failover for a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvDatabaseFailover,
	}
	composerEnvFetchDatabasePropertiesCmd = &cobra.Command{
		Use: "fetch-database-properties ENVIRONMENT", Short: "Fetch Cloud SQL database properties for a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvFetchDatabaseProperties,
	}
	composerEnvRestartWebServerCmd = &cobra.Command{
		Use: "restart-web-server ENVIRONMENT", Short: "Restart the Composer web server",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvRestartWebServer,
	}
	composerEnvListWorkloadsCmd = &cobra.Command{
		Use: "list-workloads ENVIRONMENT", Short: "List Airflow workloads in a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvListWorkloads,
	}
	composerEnvListUpgradesCmd = &cobra.Command{
		Use: "list-upgrades ENVIRONMENT",
		Short: "List the image versions this environment can upgrade to",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvListUpgrades,
	}
	composerEnvListPackagesCmd = &cobra.Command{
		Use: "list-packages ENVIRONMENT",
		Short: "List Python packages installed in the Airflow worker",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvListPackages,
	}
	composerEnvRunCmd = &cobra.Command{
		Use: "run ENVIRONMENT SUBCOMMAND [-- ARG ...]",
		Short: "Run an Airflow CLI subcommand in the environment (via composer's ExecuteAirflowCommand API)",
		Args: cobra.MinimumNArgs(2), RunE: runComposerEnvRun,
	}
)

func init() {
	all := []*cobra.Command{
		composerEnvCreateCmd, composerEnvDeleteCmd, composerEnvDescribeCmd,
		composerEnvListCmd, composerEnvUpdateCmd, composerEnvCheckUpgradeCmd,
		composerEnvDatabaseFailoverCmd, composerEnvFetchDatabasePropertiesCmd,
		composerEnvRestartWebServerCmd, composerEnvListWorkloadsCmd,
		composerEnvListUpgradesCmd, composerEnvListPackagesCmd, composerEnvRunCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagComposerEnvLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagComposerEnvFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{composerEnvCreateCmd, composerEnvUpdateCmd} {
		c.Flags().StringVar(&flagComposerEnvConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Environment body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	composerEnvUpdateCmd.Flags().StringVar(&flagComposerEnvUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	composerEnvListCmd.Flags().StringVar(&flagComposerEnvFilter, "filter", "", "Server-side filter expression")
	composerEnvListCmd.Flags().Int64Var(&flagComposerEnvPageSize, "page-size", 0, "Maximum results per page")

	composerEnvCheckUpgradeCmd.Flags().StringVar(&flagComposerEnvImageVersion, "image-version", "",
		"Target Composer/Airflow image version to check (required)")
	_ = composerEnvCheckUpgradeCmd.MarkFlagRequired("image-version")

	composerEnvListWorkloadsCmd.Flags().StringVar(&flagComposerEnvFilter, "filter", "", "Server-side filter expression")
	composerEnvListWorkloadsCmd.Flags().Int64Var(&flagComposerEnvPageSize, "page-size", 0, "Maximum results per page")

	composerEnvRunCmd.Flags().StringSliceVar(&flagComposerEnvRunParams, "parameters", nil,
		"Additional Airflow CLI arguments (repeatable)")
	composerEnvRunCmd.Flags().DurationVar(&flagComposerEnvRunTimeout, "timeout", 5*time.Minute,
		"Maximum time to wait for the Airflow command to finish")

	composerEnvListPackagesCmd.Flags().BoolVar(&flagComposerEnvRunTree, "tree", false,
		"Show the pipdeptree output instead of a flat pip list")
	composerEnvListPackagesCmd.Flags().DurationVar(&flagComposerEnvRunTimeout, "timeout", 5*time.Minute,
		"Maximum time to wait for the Airflow command to finish")

	composerEnvCmd.AddCommand(all...)
	composerEnvCmd.AddCommand(composerEnvSnapshotsCmd, composerEnvStorageCmd, composerEnvCMCmd, composerEnvSecretsCmd)
	composerCmd.AddCommand(composerEnvCmd)
}

func composerEnvParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return composerLocationParent(project, flagComposerEnvLocation), nil
}

func composerEnvResolvedName(id string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return composerEnvName(project, flagComposerEnvLocation, id), nil
}

func runComposerEnvCreate(cmd *cobra.Command, args []string) error {
	parent, err := composerEnvParent()
	if err != nil {
		return err
	}
	body := &composer.Environment{}
	if err := loadYAMLOrJSONInto(flagComposerEnvConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = fmt.Sprintf("%s/environments/%s", parent, args[0])
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating environment: %w", err)
	}
	fmt.Printf("Create request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagComposerEnvFormat)
}

func runComposerEnvDelete(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}
	fmt.Printf("Delete request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagComposerEnvFormat)
}

func runComposerEnvDescribe(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing environment: %w", err)
	}
	return emitFormatted(got, flagComposerEnvFormat)
}

func runComposerEnvList(cmd *cobra.Command, args []string) error {
	parent, err := composerEnvParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*composer.Environment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.List(parent).Context(ctx)
		if flagComposerEnvPageSize > 0 {
			call = call.PageSize(flagComposerEnvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing environments: %w", err)
		}
		for _, e := range resp.Environments {
			if flagComposerEnvFilter != "" && !strings.Contains(e.Name, flagComposerEnvFilter) {
				continue
			}
			all = append(all, e)
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagComposerEnvFormat)
}

func runComposerEnvUpdate(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	body := &composer.Environment{}
	if err := loadYAMLOrJSONInto(flagComposerEnvConfigFile, body); err != nil {
		return err
	}
	mask := flagComposerEnvUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Environments.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}
	fmt.Printf("Update request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagComposerEnvFormat)
}

func runComposerEnvCheckUpgrade(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.CheckUpgrade(name, &composer.CheckUpgradeRequest{
		ImageVersion: flagComposerEnvImageVersion,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking upgrade: %w", err)
	}
	fmt.Printf("Check-upgrade request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagComposerEnvFormat)
}

func runComposerEnvDatabaseFailover(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.DatabaseFailover(name, &composer.DatabaseFailoverRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("triggering database failover: %w", err)
	}
	fmt.Printf("Failover request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagComposerEnvFormat)
}

func runComposerEnvFetchDatabaseProperties(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Environments.FetchDatabaseProperties(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching database properties: %w", err)
	}
	return emitFormatted(resp, flagComposerEnvFormat)
}

func runComposerEnvRestartWebServer(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.RestartWebServer(name, &composer.RestartWebServerRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting web server: %w", err)
	}
	fmt.Printf("Restart request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagComposerEnvFormat)
}

func runComposerEnvListWorkloads(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*composer.ComposerWorkload
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.Workloads.List(name).Context(ctx)
		if flagComposerEnvFilter != "" {
			call = call.Filter(flagComposerEnvFilter)
		}
		if flagComposerEnvPageSize > 0 {
			call = call.PageSize(flagComposerEnvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workloads: %w", err)
		}
		all = append(all, resp.Workloads...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagComposerEnvFormat)
}

func runComposerEnvListUpgrades(cmd *cobra.Command, args []string) error {
	// Use the Environment's config.softwareConfig.imageVersion as the reference
	// and issue a CheckUpgrade that returns the set of supported target versions.
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	env, err := svc.Projects.Locations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing environment: %w", err)
	}
	current := ""
	if env.Config != nil && env.Config.SoftwareConfig != nil {
		current = env.Config.SoftwareConfig.ImageVersion
	}
	op, err := svc.Projects.Locations.Environments.CheckUpgrade(name, &composer.CheckUpgradeRequest{
		ImageVersion: current,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking upgrade: %w", err)
	}
	return emitFormatted(op, flagComposerEnvFormat)
}

func runComposerEnvListPackages(cmd *cobra.Command, args []string) error {
	subcmd := "list"
	params := []string{}
	if flagComposerEnvRunTree {
		// pipdeptree is invoked via `airflow info` on Composer 2+, but the
		// canonical Python impl runs `python -m pipdeptree --warn` on the worker.
		// The Airflow-command surface doesn't expose pipdeptree directly, so we
		// fall through to `airflow info --tree` which is the API-compatible way
		// to get the dependency graph.
		params = append(params, "--tree")
	}
	out, err := runAirflowCommand(args[0], "info", subcmd, params, flagComposerEnvRunTimeout)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func runComposerEnvRun(cmd *cobra.Command, args []string) error {
	subcommand := args[1]
	trailing := args[2:]
	// Merge --parameters and trailing args, preserving order.
	params := append([]string{}, flagComposerEnvRunParams...)
	params = append(params, trailing...)
	out, err := runAirflowCommand(args[0], "airflow", subcommand, params, flagComposerEnvRunTimeout)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func runAirflowCommand(envID, command, subcommand string, parameters []string, timeout time.Duration) (string, error) {
	name, err := composerEnvResolvedName(envID)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return "", err
	}
	resp, err := svc.Projects.Locations.Environments.ExecuteAirflowCommand(name, &composer.ExecuteAirflowCommandRequest{
		Command:    command,
		Subcommand: subcommand,
		Parameters: parameters,
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("executing airflow command: %w", err)
	}
	if resp.Error != "" {
		return "", fmt.Errorf("airflow command rejected: %s", resp.Error)
	}
	var lines strings.Builder
	nextLine := int64(0)
	backoff := 2 * time.Second
	for {
		poll, err := svc.Projects.Locations.Environments.PollAirflowCommand(name, &composer.PollAirflowCommandRequest{
			ExecutionId:    resp.ExecutionId,
			Pod:            resp.Pod,
			PodNamespace:   resp.PodNamespace,
			NextLineNumber: nextLine,
		}).Context(ctx).Do()
		if err != nil {
			return "", fmt.Errorf("polling airflow command: %w", err)
		}
		for _, ln := range poll.Output {
			lines.WriteString(ln.Content)
			lines.WriteByte('\n')
			if ln.LineNumber >= nextLine {
				nextLine = ln.LineNumber + 1
			}
		}
		if poll.OutputEnd {
			if poll.ExitInfo != nil && poll.ExitInfo.ExitCode != 0 {
				return lines.String(), fmt.Errorf("airflow command exited %d: %s", poll.ExitInfo.ExitCode, poll.ExitInfo.Error)
			}
			return lines.String(), nil
		}
		select {
		case <-ctx.Done():
			_, _ = svc.Projects.Locations.Environments.StopAirflowCommand(name, &composer.StopAirflowCommandRequest{
				ExecutionId:  resp.ExecutionId,
				Pod:          resp.Pod,
				PodNamespace: resp.PodNamespace,
			}).Context(context.Background()).Do()
			return lines.String(), fmt.Errorf("timed out waiting for airflow command: %w", ctx.Err())
		case <-time.After(backoff):
		}
		if backoff < 15*time.Second {
			backoff = time.Duration(float64(backoff) * 1.5)
		}
	}
}
