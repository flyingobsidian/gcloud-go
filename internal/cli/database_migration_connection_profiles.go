package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
	"gopkg.in/yaml.v3"
)

// --- gcloud database-migration connection-profiles (#782) ---

func (s *databaseMigrationCmds) initDatabaseMigrationConnectionProfiles() {
	// A single --region flag is shared by everything that resolves to a
	// projects/*/locations/* parent.
	for _, c := range []*cobra.Command{
		s.dmCPDescribeCmd, s.dmCPListCmd, s.dmCPDeleteCmd, s.dmCPCreateCmd,
		s.dmCPUpdateCmd, s.dmCPTestCmd, s.dmCPFetchStaticIPsCmd,
	} {
		c.Flags().StringVar(&s.flagDMCPRegion, "region", "", "Region containing the resource (required)")
		_ = c.MarkFlagRequired("region")
	}

	s.dmCPDescribeCmd.Flags().StringVar(&s.flagDMCPFormat, "format", "", "Output format (yaml, json, table, ...)")

	s.dmCPListCmd.Flags().StringVar(&s.flagDMCPFormat, "format", "", "Output format (yaml, json, table, ...)")
	s.dmCPListCmd.Flags().Int64Var(&s.flagDMCPListPageSize, "page-size", 0, "Page size for API pagination")
	s.dmCPListCmd.Flags().Int64Var(&s.flagDMCPListLimit, "limit", 0, "Cap total results (0 = no cap)")
	s.dmCPListCmd.Flags().StringVar(&s.flagDMCPListFilter, "filter", "", "Server-side filter expression")
	s.dmCPListCmd.Flags().BoolVar(&s.flagDMCPListURI, "uri", false, "Print resource names only")

	s.dmCPCreateCmd.Flags().StringVar(&s.flagDMCPConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ConnectionProfile message body (required)")
	_ = s.dmCPCreateCmd.MarkFlagRequired("config-file")
	s.dmCPCreateCmd.Flags().BoolVar(&s.flagDMCPSkipValidate, "skip-validation", false,
		"Do not validate the profile at creation time")
	s.dmCPCreateCmd.Flags().BoolVar(&s.flagDMCPAsync, "async", false,
		"Return the long-running operation immediately without waiting for completion")

	s.dmCPUpdateCmd.Flags().StringVar(&s.flagDMCPConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ConnectionProfile message body (required)")
	_ = s.dmCPUpdateCmd.MarkFlagRequired("config-file")
	s.dmCPUpdateCmd.Flags().StringVar(&s.flagDMCPUpdateMask, "update-mask", "",
		"Comma-separated set of fields to update; when omitted, defaults to every top-level field in the config file")
	s.dmCPUpdateCmd.Flags().BoolVar(&s.flagDMCPSkipValidate, "skip-validation", false,
		"Do not validate the profile at update time")
	s.dmCPUpdateCmd.Flags().BoolVar(&s.flagDMCPAsync, "async", false,
		"Return the long-running operation immediately without waiting for completion")

	s.dmCPDeleteCmd.Flags().BoolVar(&s.flagDMCPAsync, "async", false,
		"Return the long-running operation immediately without waiting for completion")

	s.dmCPTestCmd.Flags().BoolVar(&s.flagDMCPAsync, "async", false,
		"Return the long-running operation immediately without waiting for completion")

	// gcloud-python 581.0.0 added --fetch-reserved-public-ips to
	// `connection-profiles fetch-static-ips`; maps to the FetchStaticIps
	// request's fetchReservedPublicIps query parameter.
	s.dmCPFetchStaticIPsCmd.Flags().BoolVar(&s.flagDMCPFetchReservedPublic, "fetch-reserved-public-ips", false,
		"Fetch reserved public IPs allocated for private connections in addition to the service's static outbound IPs")
	s.dmCPFetchStaticIPsCmd.Flags().Int64Var(&s.flagDMCPListPageSize, "page-size", 0, "Page size for API pagination")

	s.dmConnProfilesCmd.AddCommand(
		s.dmCPCreateCmd,
		s.dmCPDeleteCmd,
		s.dmCPDescribeCmd,
		s.dmCPFetchStaticIPsCmd,
		s.dmCPListCmd,
		s.dmCPTestCmd,
		s.dmCPUpdateCmd,
	)
	s.databaseMigrationCmd.AddCommand(s.dmConnProfilesCmd)
}

// dmCPResourceName resolves a bare profile ID (plus --region) or a fully
// qualified projects/*/locations/*/connectionProfiles/* value.
func dmCPResourceName(profileID, project, region string) string {
	if strings.HasPrefix(profileID, "projects/") {
		return profileID
	}
	return fmt.Sprintf("projects/%s/locations/%s/connectionProfiles/%s", project, region, profileID)
}

// dmParent returns the projects/*/locations/* parent for the given project
// and --region.
func dmParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

// loadConnectionProfileFile reads a JSON or YAML ConnectionProfile body from
// disk. The datamigration Go types only carry JSON tags, so YAML is decoded
// via a generic-map round-trip through JSON (which does honour those tags).
func loadConnectionProfileFile(pathname string) (*datamigration.ConnectionProfile, error) {
	data, err := os.ReadFile(pathname)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	var generic any
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	jsonBytes, err := json.Marshal(convertYAMLKeys(generic))
	if err != nil {
		return nil, fmt.Errorf("normalising config file: %w", err)
	}
	profile := &datamigration.ConnectionProfile{}
	if err := json.Unmarshal(jsonBytes, profile); err != nil {
		return nil, fmt.Errorf("decoding config file into ConnectionProfile: %w", err)
	}
	return profile, nil
}

// convertYAMLKeys converts map[interface{}]interface{} (which yaml.v3 can
// still produce for nested maps decoded via any) into map[string]interface{},
// which json.Marshal accepts. Leaves other kinds untouched.
func convertYAMLKeys(v any) any {
	switch m := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, val := range m {
			out[fmt.Sprint(k)] = convertYAMLKeys(val)
		}
		return out
	case map[string]any:
		for k, val := range m {
			m[k] = convertYAMLKeys(val)
		}
		return m
	case []any:
		for i, val := range m {
			m[i] = convertYAMLKeys(val)
		}
		return m
	default:
		return v
	}
}

// waitForDMOperation polls a long-running operation until it completes.
// Returns the terminal Operation state.
func waitForDMOperation(ctx context.Context, svc *datamigration.Service, op *datamigration.Operation) (*datamigration.Operation, error) {
	for !op.Done {
		var err error
		op, err = svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (s *databaseMigrationCmds) runDMCPDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	name := dmCPResourceName(args[0], project, s.flagDMCPRegion)
	profile, err := svc.Projects.Locations.ConnectionProfiles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connection profile: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, profile, s.flagDMCPFormat)
}

func (s *databaseMigrationCmds) runDMCPList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}

	parent := dmParent(project, s.flagDMCPRegion)
	var all []*datamigration.ConnectionProfile
	pageToken := ""
	for {
		call := svc.Projects.Locations.ConnectionProfiles.List(parent).Context(ctx)
		if s.flagDMCPListFilter != "" {
			call = call.Filter(s.flagDMCPListFilter)
		}
		if s.flagDMCPListPageSize > 0 {
			call = call.PageSize(s.flagDMCPListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing connection profiles: %w", err)
		}
		all = append(all, resp.ConnectionProfiles...)
		if s.flagDMCPListLimit > 0 && int64(len(all)) >= s.flagDMCPListLimit {
			all = all[:s.flagDMCPListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if s.flagDMCPListURI {
		for _, p := range all {
			fmt.Fprintln(s.inv.stdout, p.Name)
		}
		return nil
	}
	if s.flagDMCPFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDMCPFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-20s %s\n", "NAME", "STATE", "TYPE")
	for _, p := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-20s %s\n", path.Base(p.Name), p.State, connectionProfileType(p))
	}
	return nil
}

// connectionProfileType infers the engine from which sub-message is populated
// on the profile.
func connectionProfileType(p *datamigration.ConnectionProfile) string {
	switch {
	case p.Alloydb != nil:
		return "ALLOYDB"
	case p.Cloudsql != nil:
		return "CLOUDSQL"
	case p.Mysql != nil:
		return "MYSQL"
	case p.Oracle != nil:
		return "ORACLE"
	case p.Postgresql != nil:
		return "POSTGRESQL"
	case p.Sqlserver != nil:
		return "SQLSERVER"
	}
	return ""
}

func (s *databaseMigrationCmds) runDMCPDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	name := dmCPResourceName(args[0], project, s.flagDMCPRegion)
	op, err := svc.Projects.Locations.ConnectionProfiles.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connection profile: %w", err)
	}
	if s.flagDMCPAsync {
		fmt.Fprintf(s.inv.stderr, "Delete in progress (operation: %s).\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "Deleted connection profile [%s].\n", args[0])
	return nil
}

func (s *databaseMigrationCmds) runDMCPCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	profile, err := loadConnectionProfileFile(s.flagDMCPConfigFile)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmParent(project, s.flagDMCPRegion)
	call := svc.Projects.Locations.ConnectionProfiles.Create(parent, profile).
		ConnectionProfileId(args[0]).Context(ctx)
	if s.flagDMCPSkipValidate {
		call = call.SkipValidation(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating connection profile: %w", err)
	}
	if s.flagDMCPAsync {
		fmt.Fprintf(s.inv.stderr, "Create in progress (operation: %s).\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "Created connection profile [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op.Response, "")
}

func (s *databaseMigrationCmds) runDMCPUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	profile, err := loadConnectionProfileFile(s.flagDMCPConfigFile)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	name := dmCPResourceName(args[0], project, s.flagDMCPRegion)

	mask := s.flagDMCPUpdateMask
	if mask == "" {
		mask = deriveConnectionProfileUpdateMask(profile)
	}

	call := svc.Projects.Locations.ConnectionProfiles.Patch(name, profile).
		UpdateMask(mask).Context(ctx)
	if s.flagDMCPSkipValidate {
		call = call.SkipValidation(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating connection profile: %w", err)
	}
	if s.flagDMCPAsync {
		fmt.Fprintf(s.inv.stderr, "Update in progress (operation: %s).\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "Updated connection profile [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op.Response, "")
}

// deriveConnectionProfileUpdateMask lists every top-level field present on
// the profile message, so that a bare `update` without --update-mask sends
// all populated fields.
func deriveConnectionProfileUpdateMask(p *datamigration.ConnectionProfile) string {
	var fields []string
	if p.DisplayName != "" {
		fields = append(fields, "displayName")
	}
	if p.Labels != nil {
		fields = append(fields, "labels")
	}
	if p.Provider != "" {
		fields = append(fields, "provider")
	}
	if p.State != "" {
		fields = append(fields, "state")
	}
	if p.Alloydb != nil {
		fields = append(fields, "alloydb")
	}
	if p.Cloudsql != nil {
		fields = append(fields, "cloudsql")
	}
	if p.Mysql != nil {
		fields = append(fields, "mysql")
	}
	if p.Oracle != nil {
		fields = append(fields, "oracle")
	}
	if p.Postgresql != nil {
		fields = append(fields, "postgresql")
	}
	if p.Sqlserver != nil {
		fields = append(fields, "sqlserver")
	}
	if len(fields) == 0 {
		return "displayName"
	}
	return strings.Join(fields, ",")
}

func (s *databaseMigrationCmds) runDMCPTest(cmd *cobra.Command, args []string) error {
	// Match gcloud-python's connection_profiles.Test: fetch the existing
	// profile, then issue a Patch with validateOnly=true and a no-op
	// updateMask (displayName). The service runs validation but does not
	// persist any changes.
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	name := dmCPResourceName(args[0], project, s.flagDMCPRegion)

	current, err := svc.Projects.Locations.ConnectionProfiles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching connection profile: %w", err)
	}
	op, err := svc.Projects.Locations.ConnectionProfiles.Patch(name, current).
		UpdateMask("displayName").ValidateOnly(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("testing connection profile: %w", err)
	}
	if s.flagDMCPAsync {
		fmt.Fprintf(s.inv.stderr, "Test in progress (operation: %s).\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "Connection profile [%s] validated successfully.\n", args[0])
	return nil
}

func (s *databaseMigrationCmds) runDMCPFetchStaticIPs(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmParent(project, s.flagDMCPRegion)
	call := svc.Projects.Locations.FetchStaticIps(parent).Context(ctx)
	if s.flagDMCPListPageSize > 0 {
		call = call.PageSize(s.flagDMCPListPageSize)
	}
	if s.flagDMCPFetchReservedPublic {
		call = call.FetchReservedPublicIps(true)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("fetching static IPs: %w", err)
	}
	for _, ip := range resp.StaticIps {
		fmt.Fprintln(s.inv.stdout, ip)
	}
	return nil
}
