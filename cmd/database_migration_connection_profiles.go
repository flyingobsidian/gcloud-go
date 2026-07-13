package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
	"gopkg.in/yaml.v3"
)

// --- gcloud database-migration connection-profiles (#782) ---

var dmConnProfilesCmd = &cobra.Command{
	Use:   "connection-profiles",
	Short: "Manage Database Migration Service connection profiles",
}

var dmCPDescribeCmd = &cobra.Command{
	Use:   "describe PROFILE",
	Short: "Show details about a connection profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMCPDescribe,
}

var dmCPListCmd = &cobra.Command{
	Use:   "list",
	Short: "List connection profiles in a region",
	Args:  cobra.NoArgs,
	RunE:  runDMCPList,
}

var dmCPDeleteCmd = &cobra.Command{
	Use:   "delete PROFILE",
	Short: "Delete a connection profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMCPDelete,
}

var dmCPCreateCmd = &cobra.Command{
	Use:   "create PROFILE",
	Short: "Create a connection profile from a --config-file",
	Long: `Create a Database Migration Service connection profile.

The --config-file argument points at a JSON or YAML file containing a
ConnectionProfile message body. gcloud-python exposes this command as a
per-engine group (create mysql, create postgresql, etc.); the generic form
here accepts any engine by loading its message body from disk. See:
https://cloud.google.com/database-migration/docs/reference/rest/v1/projects.locations.connectionProfiles#ConnectionProfile
`,
	Args: cobra.ExactArgs(1),
	RunE: runDMCPCreate,
}

var dmCPUpdateCmd = &cobra.Command{
	Use:   "update PROFILE",
	Short: "Update a connection profile from a --config-file",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMCPUpdate,
}

var dmCPTestCmd = &cobra.Command{
	Use:   "test PROFILE",
	Short: "Test a connection profile without persisting changes",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMCPTest,
}

var dmCPFetchStaticIPsCmd = &cobra.Command{
	Use:   "fetch-static-ips",
	Short: "Fetch static IPs the service uses to establish outbound connections",
	Args:  cobra.NoArgs,
	RunE:  runDMCPFetchStaticIPs,
}

var (
	flagDMCPRegion       string
	flagDMCPFormat       string
	flagDMCPListPageSize int64
	flagDMCPListLimit    int64
	flagDMCPListFilter   string
	flagDMCPListURI      bool
	flagDMCPConfigFile   string
	flagDMCPUpdateMask   string
	flagDMCPSkipValidate bool
	flagDMCPAsync        bool
)

func init() {
	// A single --region flag is shared by everything that resolves to a
	// projects/*/locations/* parent.
	for _, c := range []*cobra.Command{
		dmCPDescribeCmd, dmCPListCmd, dmCPDeleteCmd, dmCPCreateCmd,
		dmCPUpdateCmd, dmCPTestCmd, dmCPFetchStaticIPsCmd,
	} {
		c.Flags().StringVar(&flagDMCPRegion, "region", "", "Region containing the resource (required)")
		_ = c.MarkFlagRequired("region")
	}

	dmCPDescribeCmd.Flags().StringVar(&flagDMCPFormat, "format", "", "Output format (yaml, json, table, ...)")

	dmCPListCmd.Flags().StringVar(&flagDMCPFormat, "format", "", "Output format (yaml, json, table, ...)")
	dmCPListCmd.Flags().Int64Var(&flagDMCPListPageSize, "page-size", 0, "Page size for API pagination")
	dmCPListCmd.Flags().Int64Var(&flagDMCPListLimit, "limit", 0, "Cap total results (0 = no cap)")
	dmCPListCmd.Flags().StringVar(&flagDMCPListFilter, "filter", "", "Server-side filter expression")
	dmCPListCmd.Flags().BoolVar(&flagDMCPListURI, "uri", false, "Print resource names only")

	dmCPCreateCmd.Flags().StringVar(&flagDMCPConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ConnectionProfile message body (required)")
	_ = dmCPCreateCmd.MarkFlagRequired("config-file")
	dmCPCreateCmd.Flags().BoolVar(&flagDMCPSkipValidate, "skip-validation", false,
		"Do not validate the profile at creation time")
	dmCPCreateCmd.Flags().BoolVar(&flagDMCPAsync, "async", false,
		"Return the long-running operation immediately without waiting for completion")

	dmCPUpdateCmd.Flags().StringVar(&flagDMCPConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ConnectionProfile message body (required)")
	_ = dmCPUpdateCmd.MarkFlagRequired("config-file")
	dmCPUpdateCmd.Flags().StringVar(&flagDMCPUpdateMask, "update-mask", "",
		"Comma-separated set of fields to update; when omitted, defaults to every top-level field in the config file")
	dmCPUpdateCmd.Flags().BoolVar(&flagDMCPSkipValidate, "skip-validation", false,
		"Do not validate the profile at update time")
	dmCPUpdateCmd.Flags().BoolVar(&flagDMCPAsync, "async", false,
		"Return the long-running operation immediately without waiting for completion")

	dmCPDeleteCmd.Flags().BoolVar(&flagDMCPAsync, "async", false,
		"Return the long-running operation immediately without waiting for completion")

	dmCPTestCmd.Flags().BoolVar(&flagDMCPAsync, "async", false,
		"Return the long-running operation immediately without waiting for completion")

	dmConnProfilesCmd.AddCommand(
		dmCPCreateCmd,
		dmCPDeleteCmd,
		dmCPDescribeCmd,
		dmCPFetchStaticIPsCmd,
		dmCPListCmd,
		dmCPTestCmd,
		dmCPUpdateCmd,
	)
	databaseMigrationCmd.AddCommand(dmConnProfilesCmd)
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

func runDMCPDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := dmCPResourceName(args[0], project, flagDMCPRegion)
	profile, err := svc.Projects.Locations.ConnectionProfiles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connection profile: %w", err)
	}
	return emitFormatted(profile, flagDMCPFormat)
}

func runDMCPList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := dmParent(project, flagDMCPRegion)
	var all []*datamigration.ConnectionProfile
	pageToken := ""
	for {
		call := svc.Projects.Locations.ConnectionProfiles.List(parent).Context(ctx)
		if flagDMCPListFilter != "" {
			call = call.Filter(flagDMCPListFilter)
		}
		if flagDMCPListPageSize > 0 {
			call = call.PageSize(flagDMCPListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing connection profiles: %w", err)
		}
		all = append(all, resp.ConnectionProfiles...)
		if flagDMCPListLimit > 0 && int64(len(all)) >= flagDMCPListLimit {
			all = all[:flagDMCPListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagDMCPListURI {
		for _, p := range all {
			fmt.Println(p.Name)
		}
		return nil
	}
	if flagDMCPFormat != "" {
		return emitFormatted(all, flagDMCPFormat)
	}
	fmt.Printf("%-40s %-20s %s\n", "NAME", "STATE", "TYPE")
	for _, p := range all {
		fmt.Printf("%-40s %-20s %s\n", path.Base(p.Name), p.State, connectionProfileType(p))
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

func runDMCPDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := dmCPResourceName(args[0], project, flagDMCPRegion)
	op, err := svc.Projects.Locations.ConnectionProfiles.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connection profile: %w", err)
	}
	if flagDMCPAsync {
		fmt.Fprintf(os.Stderr, "Delete in progress (operation: %s).\n", op.Name)
		return emitFormatted(op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Deleted connection profile [%s].\n", args[0])
	return nil
}

func runDMCPCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	profile, err := loadConnectionProfileFile(flagDMCPConfigFile)
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmParent(project, flagDMCPRegion)
	call := svc.Projects.Locations.ConnectionProfiles.Create(parent, profile).
		ConnectionProfileId(args[0]).Context(ctx)
	if flagDMCPSkipValidate {
		call = call.SkipValidation(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating connection profile: %w", err)
	}
	if flagDMCPAsync {
		fmt.Fprintf(os.Stderr, "Create in progress (operation: %s).\n", op.Name)
		return emitFormatted(op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Created connection profile [%s].\n", args[0])
	return emitFormatted(op.Response, "")
}

func runDMCPUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	profile, err := loadConnectionProfileFile(flagDMCPConfigFile)
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := dmCPResourceName(args[0], project, flagDMCPRegion)

	mask := flagDMCPUpdateMask
	if mask == "" {
		mask = deriveConnectionProfileUpdateMask(profile)
	}

	call := svc.Projects.Locations.ConnectionProfiles.Patch(name, profile).
		UpdateMask(mask).Context(ctx)
	if flagDMCPSkipValidate {
		call = call.SkipValidation(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating connection profile: %w", err)
	}
	if flagDMCPAsync {
		fmt.Fprintf(os.Stderr, "Update in progress (operation: %s).\n", op.Name)
		return emitFormatted(op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Updated connection profile [%s].\n", args[0])
	return emitFormatted(op.Response, "")
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

func runDMCPTest(cmd *cobra.Command, args []string) error {
	// Match gcloud-python's connection_profiles.Test: fetch the existing
	// profile, then issue a Patch with validateOnly=true and a no-op
	// updateMask (displayName). The service runs validation but does not
	// persist any changes.
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := dmCPResourceName(args[0], project, flagDMCPRegion)

	current, err := svc.Projects.Locations.ConnectionProfiles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching connection profile: %w", err)
	}
	op, err := svc.Projects.Locations.ConnectionProfiles.Patch(name, current).
		UpdateMask("displayName").ValidateOnly(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("testing connection profile: %w", err)
	}
	if flagDMCPAsync {
		fmt.Fprintf(os.Stderr, "Test in progress (operation: %s).\n", op.Name)
		return emitFormatted(op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Connection profile [%s] validated successfully.\n", args[0])
	return nil
}

func runDMCPFetchStaticIPs(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmParent(project, flagDMCPRegion)
	call := svc.Projects.Locations.FetchStaticIps(parent).Context(ctx)
	if flagDMCPListPageSize > 0 {
		call = call.PageSize(flagDMCPListPageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("fetching static IPs: %w", err)
	}
	for _, ip := range resp.StaticIps {
		fmt.Println(ip)
	}
	return nil
}
