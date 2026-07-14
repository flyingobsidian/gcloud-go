package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	observability "google.golang.org/api/observability/v1"
)

// --- gcloud observability (#366) ---

var observabilityCmd = &cobra.Command{Use: "observability", Short: "Manage Observability resources"}

func obLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func obChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

var (
	flagObLocation   string
	flagObConfigFile string
	flagObUpdateMask string
	flagObFormat     string
)

// --- scopes ---

var observabilityScopesCmd = &cobra.Command{Use: "scopes", Short: "Manage observability scopes"}

var (
	obScopeDescribeCmd = &cobra.Command{
		Use: "describe SCOPE", Short: "Describe an observability scope",
		Args: cobra.ExactArgs(1), RunE: runObScopeDescribe,
	}
	obScopeUpdateCmd = &cobra.Command{
		Use: "update SCOPE", Short: "Update an observability scope from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runObScopeUpdate,
	}
)

// --- trace-scopes ---

var observabilityTraceScopesCmd = &cobra.Command{Use: "trace-scopes", Short: "Manage observability trace scopes"}

var (
	obTSCreateCmd = &cobra.Command{
		Use: "create TRACE_SCOPE", Short: "Create a trace scope from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runObTSCreate,
	}
	obTSDeleteCmd = &cobra.Command{
		Use: "delete TRACE_SCOPE", Short: "Delete a trace scope",
		Args: cobra.ExactArgs(1), RunE: runObTSDelete,
	}
	obTSDescribeCmd = &cobra.Command{
		Use: "describe TRACE_SCOPE", Short: "Describe a trace scope",
		Args: cobra.ExactArgs(1), RunE: runObTSDescribe,
	}
	obTSListCmd = &cobra.Command{
		Use: "list", Short: "List trace scopes",
		Args: cobra.NoArgs, RunE: runObTSList,
	}
	obTSUpdateCmd = &cobra.Command{
		Use: "update TRACE_SCOPE", Short: "Update a trace scope from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runObTSUpdate,
	}
)

func init() {
	// scopes
	for _, c := range []*cobra.Command{obScopeDescribeCmd, obScopeUpdateCmd} {
		c.Flags().StringVar(&flagObLocation, "location", "global", "Location containing the scope")
	}
	obScopeDescribeCmd.Flags().StringVar(&flagObFormat, "format", "", "Output format")
	obScopeUpdateCmd.Flags().StringVar(&flagObConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Scope body (required)")
	_ = obScopeUpdateCmd.MarkFlagRequired("config-file")
	obScopeUpdateCmd.Flags().StringVar(&flagObUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	observabilityScopesCmd.AddCommand(obScopeDescribeCmd, obScopeUpdateCmd)
	observabilityCmd.AddCommand(observabilityScopesCmd)

	// trace-scopes
	tsAll := []*cobra.Command{obTSCreateCmd, obTSDeleteCmd, obTSDescribeCmd, obTSListCmd, obTSUpdateCmd}
	for _, c := range tsAll {
		c.Flags().StringVar(&flagObLocation, "location", "global", "Location containing the trace scope")
	}
	for _, c := range []*cobra.Command{obTSCreateCmd, obTSUpdateCmd} {
		c.Flags().StringVar(&flagObConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the TraceScope body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	obTSUpdateCmd.Flags().StringVar(&flagObUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	obTSDescribeCmd.Flags().StringVar(&flagObFormat, "format", "", "Output format")
	obTSListCmd.Flags().StringVar(&flagObFormat, "format", "", "Output format")
	observabilityTraceScopesCmd.AddCommand(tsAll...)
	observabilityCmd.AddCommand(observabilityTraceScopesCmd)

	rootCmd.AddCommand(observabilityCmd)
}

// --- scopes impl ---

func obScopeName(id, project, location string) string {
	return obChild("scopes", id, obLocationParent(project, location))
}

func runObScopeDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Scopes.Get(obScopeName(args[0], project, flagObLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing scope: %w", err)
	}
	return emitFormatted(got, flagObFormat)
}

func runObScopeUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	s := &observability.Scope{}
	if err := loadYAMLOrJSONInto(flagObConfigFile, s); err != nil {
		return err
	}
	mask := flagObUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(s))
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Scopes.Patch(obScopeName(args[0], project, flagObLocation), s).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating scope: %w", err)
	}
	return emitFormatted(got, "")
}

// --- trace-scopes impl ---

func obTSName(id, project, location string) string {
	return obChild("traceScopes", id, obLocationParent(project, location))
}

func runObTSCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ts := &observability.TraceScope{}
	if err := loadYAMLOrJSONInto(flagObConfigFile, ts); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TraceScopes.Create(obLocationParent(project, flagObLocation), ts).
		TraceScopeId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating trace scope: %w", err)
	}
	return emitFormatted(got, "")
}

func runObTSDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.TraceScopes.Delete(obTSName(args[0], project, flagObLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting trace scope: %w", err)
	}
	fmt.Printf("Deleted trace scope [%s].\n", args[0])
	return nil
}

func runObTSDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TraceScopes.Get(obTSName(args[0], project, flagObLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing trace scope: %w", err)
	}
	return emitFormatted(got, flagObFormat)
}

func runObTSList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.TraceScopes.List(obLocationParent(project, flagObLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing trace scopes: %w", err)
	}
	if flagObFormat != "" {
		return emitFormatted(resp.TraceScopes, flagObFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, ts := range resp.TraceScopes {
		fmt.Println(path.Base(ts.Name))
	}
	return nil
}

func runObTSUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ts := &observability.TraceScope{}
	if err := loadYAMLOrJSONInto(flagObConfigFile, ts); err != nil {
		return err
	}
	mask := flagObUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ts))
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TraceScopes.Patch(obTSName(args[0], project, flagObLocation), ts).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating trace scope: %w", err)
	}
	return emitFormatted(got, "")
}
