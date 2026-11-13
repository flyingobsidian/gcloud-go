package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	observability "google.golang.org/api/observability/v1"
)

// --- gcloud observability (#366) ---

func obLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func obChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// --- scopes ---

// --- trace-scopes ---

func (g *observabilityCmds) initObservability() {
	// scopes
	for _, c := range []*cobra.Command{g.obScopeDescribeCmd, g.obScopeUpdateCmd} {
		c.Flags().StringVar(&g.flagObLocation, "location", "global", "Location containing the scope")
	}
	g.obScopeDescribeCmd.Flags().StringVar(&g.flagObFormat, "format", "", "Output format")
	g.obScopeUpdateCmd.Flags().StringVar(&g.flagObConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Scope body (required)")
	_ = g.obScopeUpdateCmd.MarkFlagRequired("config-file")
	g.obScopeUpdateCmd.Flags().StringVar(&g.flagObUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	g.observabilityScopesCmd.AddCommand(g.obScopeDescribeCmd, g.obScopeUpdateCmd)
	g.observabilityCmd.AddCommand(g.observabilityScopesCmd)

	// trace-scopes
	tsAll := []*cobra.Command{g.obTSCreateCmd, g.obTSDeleteCmd, g.obTSDescribeCmd, g.obTSListCmd, g.obTSUpdateCmd}
	for _, c := range tsAll {
		c.Flags().StringVar(&g.flagObLocation, "location", "global", "Location containing the trace scope")
	}
	for _, c := range []*cobra.Command{g.obTSCreateCmd, g.obTSUpdateCmd} {
		c.Flags().StringVar(&g.flagObConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the TraceScope body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.obTSUpdateCmd.Flags().StringVar(&g.flagObUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	g.obTSDescribeCmd.Flags().StringVar(&g.flagObFormat, "format", "", "Output format")
	g.obTSListCmd.Flags().StringVar(&g.flagObFormat, "format", "", "Output format")
	g.observabilityTraceScopesCmd.AddCommand(tsAll...)
	g.observabilityCmd.AddCommand(g.observabilityTraceScopesCmd)

}

// --- scopes impl ---

func obScopeName(id, project, location string) string {
	return obChild("scopes", id, obLocationParent(project, location))
}

func (g *observabilityCmds) runObScopeDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Scopes.Get(obScopeName(args[0], project, g.flagObLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing scope: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagObFormat)
}

func (g *observabilityCmds) runObScopeUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	s := &observability.Scope{}
	if err := loadYAMLOrJSONInto(g.flagObConfigFile, s); err != nil {
		return err
	}
	mask := g.flagObUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Scopes.Patch(obScopeName(args[0], project, g.flagObLocation), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating scope: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

// --- trace-scopes impl ---

func obTSName(id, project, location string) string {
	return obChild("traceScopes", id, obLocationParent(project, location))
}

func (g *observabilityCmds) runObTSCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ts := &observability.TraceScope{}
	if err := loadYAMLOrJSONInto(g.flagObConfigFile, ts); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TraceScopes.Create(obLocationParent(project, g.flagObLocation), ts).
		TraceScopeId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating trace scope: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *observabilityCmds) runObTSDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.TraceScopes.Delete(obTSName(args[0], project, g.flagObLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting trace scope: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted trace scope [%s].\n", args[0])
	return nil
}

func (g *observabilityCmds) runObTSDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TraceScopes.Get(obTSName(args[0], project, g.flagObLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing trace scope: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagObFormat)
}

func (g *observabilityCmds) runObTSList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.TraceScopes.List(obLocationParent(project, g.flagObLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing trace scopes: %w", err)
	}
	if g.flagObFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.TraceScopes, g.flagObFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.TraceScopes), "%-40s\n", "NAME")
	for _, ts := range resp.TraceScopes {
		fmt.Fprintln(g.inv.stdout, path.Base(ts.Name))
	}
	return nil
}

func (g *observabilityCmds) runObTSUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ts := &observability.TraceScope{}
	if err := loadYAMLOrJSONInto(g.flagObConfigFile, ts); err != nil {
		return err
	}
	mask := g.flagObUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ts))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TraceScopes.Patch(obTSName(args[0], project, g.flagObLocation), ts).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating trace scope: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

// observabilityCmds holds the commands and flag values of one invocation's "observability"
// command group, so invocations don't share state (#2070).
type observabilityCmds struct {
	inv *invocation

	observabilityCmd            *cobra.Command
	flagObLocation              string
	flagObConfigFile            string
	flagObUpdateMask            string
	flagObFormat                string
	observabilityScopesCmd      *cobra.Command
	obScopeDescribeCmd          *cobra.Command
	obScopeUpdateCmd            *cobra.Command
	observabilityTraceScopesCmd *cobra.Command
	obTSCreateCmd               *cobra.Command
	obTSDeleteCmd               *cobra.Command
	obTSDescribeCmd             *cobra.Command
	obTSListCmd                 *cobra.Command
	obTSUpdateCmd               *cobra.Command
	observabilityBucketsCmd     *cobra.Command
	obBucketCreateCmd           *cobra.Command
	obBucketDescribeCmd         *cobra.Command
	obBucketListCmd             *cobra.Command
	obBucketUpdateCmd           *cobra.Command
	flagObBucketLocation        string
	flagObBucketFormat          string
	flagObBucketAsync           bool
	flagObBucketDisplayName     string
	flagObBucketDescription     string
	flagObBucketKmsKey          string
	flagObBucketClearCmek       bool
	flagObBucketShowDeleted     bool
	flagObBucketPageSize        int64
}

// newObservabilityCmd returns the "observability" command group for inv.
func newObservabilityCmd(inv *invocation) *cobra.Command {
	return newObservabilityCmds(inv).observabilityCmd
}

// newObservabilityCmds builds the "observability" command group for inv and returns its state.
func newObservabilityCmds(inv *invocation) *observabilityCmds {
	g := &observabilityCmds{inv: inv}
	g.observabilityCmd = &cobra.Command{Use: "observability", Short: "Manage Observability resources"}
	g.observabilityScopesCmd = &cobra.Command{Use: "scopes", Short: "Manage observability scopes"}
	g.obScopeDescribeCmd = &cobra.Command{
		Use: "describe SCOPE", Short: "Describe an observability scope",
		Args: cobra.ExactArgs(1), RunE: g.runObScopeDescribe,
	}
	g.obScopeUpdateCmd = &cobra.Command{
		Use: "update SCOPE", Short: "Update an observability scope from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runObScopeUpdate,
	}
	g.observabilityTraceScopesCmd = &cobra.Command{Use: "trace-scopes", Short: "Manage observability trace scopes"}
	g.obTSCreateCmd = &cobra.Command{
		Use: "create TRACE_SCOPE", Short: "Create a trace scope from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runObTSCreate,
	}
	g.obTSDeleteCmd = &cobra.Command{
		Use: "delete TRACE_SCOPE", Short: "Delete a trace scope",
		Args: cobra.ExactArgs(1), RunE: g.runObTSDelete,
	}
	g.obTSDescribeCmd = &cobra.Command{
		Use: "describe TRACE_SCOPE", Short: "Describe a trace scope",
		Args: cobra.ExactArgs(1), RunE: g.runObTSDescribe,
	}
	g.obTSListCmd = &cobra.Command{
		Use: "list", Short: "List trace scopes",
		Args: cobra.NoArgs, RunE: g.runObTSList,
	}
	g.obTSUpdateCmd = &cobra.Command{
		Use: "update TRACE_SCOPE", Short: "Update a trace scope from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runObTSUpdate,
	}
	g.observabilityBucketsCmd = &cobra.Command{Use: "buckets", Short: "Manage observability buckets"}
	g.obBucketCreateCmd = &cobra.Command{
		Use: "create BUCKET", Short: "Create an observability bucket",
		Args: cobra.ExactArgs(1), RunE: g.runObBucketCreate,
	}
	g.obBucketDescribeCmd = &cobra.Command{
		Use: "describe BUCKET", Short: "Describe an observability bucket",
		Args: cobra.ExactArgs(1), RunE: g.runObBucketDescribe,
	}
	g.obBucketListCmd = &cobra.Command{
		Use: "list", Short: "List observability buckets",
		Args: cobra.NoArgs, RunE: g.runObBucketList,
	}
	g.obBucketUpdateCmd = &cobra.Command{
		Use: "update BUCKET", Short: "Update an observability bucket",
		Args: cobra.ExactArgs(1), RunE: g.runObBucketUpdate,
	}
	g.initObservability()
	g.initObservabilityBuckets()
	return g
}
