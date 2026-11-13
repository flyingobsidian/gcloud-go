package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	composer "google.golang.org/api/composer/v1"
)

// --- gcloud composer environments (#1502) ---

func (s *composerCmds) initComposerEnvironments() {
	all := []*cobra.Command{
		s.composerEnvCreateCmd, s.composerEnvDeleteCmd, s.composerEnvDescribeCmd,
		s.composerEnvListCmd, s.composerEnvUpdateCmd, s.composerEnvCheckUpgradeCmd,
		s.composerEnvDatabaseFailoverCmd, s.composerEnvFetchDatabasePropertiesCmd,
		s.composerEnvRestartWebServerCmd, s.composerEnvListWorkloadsCmd,
		s.composerEnvListUpgradesCmd, s.composerEnvListPackagesCmd, s.composerEnvRunCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagComposerEnvFormat, "format", "", "Output format")
		if c == s.composerEnvListCmd {
			continue
		}
		c.Flags().StringVar(&s.flagComposerEnvLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
	}
	// As in gcloud, list takes --locations, falling back to composer/location.
	s.composerEnvListCmd.Flags().StringSliceVar(&s.flagComposerEnvListLocations, "locations", nil,
		"Locations in which to list environments; default the composer/location property")
	// --location is kept as a hidden alias for earlier gcloud-go releases.
	s.composerEnvListCmd.Flags().StringVar(&s.flagComposerEnvLocation, "location", "", "Composer location")
	_ = s.composerEnvListCmd.Flags().MarkHidden("location")
	s.composerEnvListCmd.Flags().Int64Var(&s.flagComposerEnvListLimit, "limit", 0, "Maximum number of environments to list (0 = no limit)")
	for _, c := range []*cobra.Command{s.composerEnvCreateCmd, s.composerEnvUpdateCmd} {
		c.Flags().StringVar(&s.flagComposerEnvConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Environment body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.composerEnvUpdateCmd.Flags().StringVar(&s.flagComposerEnvUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.composerEnvListCmd.Flags().StringVar(&s.flagComposerEnvFilter, "filter", "", "Server-side filter expression")
	s.composerEnvListCmd.Flags().Int64Var(&s.flagComposerEnvPageSize, "page-size", 0, "Maximum results per page")

	s.composerEnvCheckUpgradeCmd.Flags().StringVar(&s.flagComposerEnvImageVersion, "image-version", "",
		"Target Composer/Airflow image version to check (required)")
	_ = s.composerEnvCheckUpgradeCmd.MarkFlagRequired("image-version")

	s.composerEnvListWorkloadsCmd.Flags().StringVar(&s.flagComposerEnvFilter, "filter", "", "Server-side filter expression")
	s.composerEnvListWorkloadsCmd.Flags().Int64Var(&s.flagComposerEnvPageSize, "page-size", 0, "Maximum results per page")

	s.composerEnvRunCmd.Flags().StringSliceVar(&s.flagComposerEnvRunParams, "parameters", nil,
		"Additional Airflow CLI arguments (repeatable)")
	s.composerEnvRunCmd.Flags().DurationVar(&s.flagComposerEnvRunTimeout, "timeout", 5*time.Minute,
		"Maximum time to wait for the Airflow command to finish")

	s.composerEnvListPackagesCmd.Flags().BoolVar(&s.flagComposerEnvRunTree, "tree", false,
		"Show the pipdeptree output instead of a flat pip list")
	s.composerEnvListPackagesCmd.Flags().DurationVar(&s.flagComposerEnvRunTimeout, "timeout", 5*time.Minute,
		"Maximum time to wait for the Airflow command to finish")

	s.composerEnvCmd.AddCommand(all...)
	s.composerEnvCmd.AddCommand(s.composerEnvSnapshotsCmd, s.composerEnvStorageCmd, s.composerEnvCMCmd, s.composerEnvSecretsCmd)
	s.composerCmd.AddCommand(s.composerEnvCmd)
}

func (s *composerCmds) composerEnvParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return composerLocationParent(project, s.flagComposerEnvLocation), nil
}

func (s *composerCmds) composerEnvResolvedName(id string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return composerEnvName(project, s.flagComposerEnvLocation, id), nil
}

func (s *composerCmds) initComposerEnvironments2() {
	s.composerEnvCreateCmd.Flags().BoolVar(&s.flagComposerEnvDevelopmentMode, "enable-development-mode", false,
		"Enable development mode for the environment. Development mode constrains the "+
			"environment configuration but allows environment hibernation; use it only for test environments.")
}

// applyComposerDevelopmentMode sets mode=DEVELOPMENT when requested, leaving
// any --config-file mode otherwise.
func applyComposerDevelopmentMode(env *composer.Environment, enable bool) {
	if enable {
		env.Mode = "DEVELOPMENT"
	}
}

func (s *composerCmds) runComposerEnvCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.composerEnvParent()
	if err != nil {
		return err
	}
	body := &composer.Environment{}
	if err := loadYAMLOrJSONInto(s.flagComposerEnvConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = fmt.Sprintf("%s/environments/%s", parent, args[0])
	}
	applyComposerDevelopmentMode(body, s.flagComposerEnvDevelopmentMode)
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating environment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvDelete(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing environment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagComposerEnvFormat)
}

// errComposerNoLocations is gcloud's error when `environments list` has
// neither --locations nor the composer/location property.
var errComposerNoLocations = errors.New(
	"Missing required argument [--locations]: One or more locations in which to list environments must be provided.")

// composerEnvListLocations returns --locations (or the hidden --location),
// else the composer/location property.
func (s *composerCmds) composerEnvListLocations() ([]string, error) {
	if len(s.flagComposerEnvListLocations) > 0 {
		return s.flagComposerEnvListLocations, nil
	}
	if s.flagComposerEnvLocation != "" {
		return []string{s.flagComposerEnvLocation}, nil
	}
	location, err := s.inv.env.Property("composer", "location")
	if err != nil {
		return nil, fmt.Errorf("reading composer/location: %w", err)
	}
	if location == "" {
		return nil, errComposerNoLocations
	}
	return []string{location}, nil
}

// composerEnvPager fetches one page of environments in parent.
type composerEnvPager func(ctx context.Context, parent, pageToken string) (*composer.ListEnvironmentsResponse, error)

// listComposerEnvs lists environments in each parent in turn until they run
// out or limit have been collected (limit <= 0 means no limit), skipping
// names without the client-side --filter substring. It never returns a nil
// slice.
func listComposerEnvs(ctx context.Context, parents []string, page composerEnvPager, filter string, limit int64) ([]*composer.Environment, error) {
	all := []*composer.Environment{}
	for _, parent := range parents {
		pageToken := ""
		for {
			resp, err := page(ctx, parent, pageToken)
			if err != nil {
				return nil, fmt.Errorf("listing environments: %w", err)
			}
			for _, e := range resp.Environments {
				if filter != "" && !strings.Contains(e.Name, filter) {
					continue
				}
				all = append(all, e)
				if limit > 0 && int64(len(all)) >= limit {
					return all, nil
				}
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
	}
	return all, nil
}

func (s *composerCmds) runComposerEnvList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	locations, err := s.composerEnvListLocations()
	if err != nil {
		return err
	}
	parents := make([]string, len(locations))
	for i, loc := range locations {
		parents[i] = fmt.Sprintf("projects/%s/locations/%s", project, loc)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	page := func(ctx context.Context, parent, pageToken string) (*composer.ListEnvironmentsResponse, error) {
		call := svc.Projects.Locations.Environments.List(parent).Context(ctx)
		if s.flagComposerEnvPageSize > 0 {
			call = call.PageSize(s.flagComposerEnvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		return call.Do()
	}
	all, err := listComposerEnvs(ctx, parents, page, s.flagComposerEnvFilter, s.flagComposerEnvListLimit)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	body := &composer.Environment{}
	if err := loadYAMLOrJSONInto(s.flagComposerEnvConfigFile, body); err != nil {
		return err
	}
	mask := s.flagComposerEnvUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Update request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvCheckUpgrade(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.CheckUpgrade(name, &composer.CheckUpgradeRequest{
		ImageVersion: s.flagComposerEnvImageVersion,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking upgrade: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Check-upgrade request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvDatabaseFailover(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.DatabaseFailover(name, &composer.DatabaseFailoverRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("triggering database failover: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Failover request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvFetchDatabaseProperties(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Environments.FetchDatabaseProperties(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching database properties: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvRestartWebServer(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.RestartWebServer(name, &composer.RestartWebServerRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting web server: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restart request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvListWorkloads(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	var all []*composer.ComposerWorkload
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.Workloads.List(name).Context(ctx)
		if s.flagComposerEnvFilter != "" {
			call = call.Filter(s.flagComposerEnvFilter)
		}
		if s.flagComposerEnvPageSize > 0 {
			call = call.PageSize(s.flagComposerEnvPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvListUpgrades(cmd *cobra.Command, args []string) error {
	// Use the Environment's config.softwareConfig.imageVersion as the reference
	// and issue a CheckUpgrade that returns the set of supported target versions.
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
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
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvListPackages(cmd *cobra.Command, args []string) error {
	subcmd := "list"
	params := []string{}
	if s.flagComposerEnvRunTree {
		// pipdeptree is invoked via `airflow info` on Composer 2+, but the
		// canonical Python impl runs `python -m pipdeptree --warn` on the worker.
		// The Airflow-command surface doesn't expose pipdeptree directly, so we
		// fall through to `airflow info --tree` which is the API-compatible way
		// to get the dependency graph.
		params = append(params, "--tree")
	}
	out, err := s.runAirflowCommand(args[0], "info", subcmd, params, s.flagComposerEnvRunTimeout)
	if err != nil {
		return err
	}
	fmt.Fprintln(s.inv.stdout, out)
	return nil
}

func (s *composerCmds) runComposerEnvRun(cmd *cobra.Command, args []string) error {
	subcommand := args[1]
	trailing := args[2:]
	// Merge --parameters and trailing args, preserving order.
	params := append([]string{}, s.flagComposerEnvRunParams...)
	params = append(params, trailing...)
	out, err := s.runAirflowCommand(args[0], "airflow", subcommand, params, s.flagComposerEnvRunTimeout)
	if err != nil {
		return err
	}
	fmt.Fprintln(s.inv.stdout, out)
	return nil
}

func (s *composerCmds) runAirflowCommand(envID, command, subcommand string, parameters []string, timeout time.Duration) (string, error) {
	name, err := s.composerEnvResolvedName(envID)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(s.inv.context(), timeout)
	defer cancel()
	svc, err := s.inv.clients().Composer(ctx)
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
			}).Context(s.inv.context()).Do()
			return lines.String(), fmt.Errorf("timed out waiting for airflow command: %w", ctx.Err())
		case <-time.After(backoff):
		}
		if backoff < 15*time.Second {
			backoff = time.Duration(float64(backoff) * 1.5)
		}
	}
}
