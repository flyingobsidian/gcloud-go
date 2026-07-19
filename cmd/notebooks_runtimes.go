package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks runtimes (#1064) ---

var notebooksRuntimeCmd = &cobra.Command{Use: "runtimes", Short: "Manage notebook runtimes"}

var (
	flagNotebooksRuntimeLocation   string
	flagNotebooksRuntimeFormat     string
	flagNotebooksRuntimeConfigFile string
	flagNotebooksRuntimePageSize   int64
)

var (
	notebooksRuntimeCreateCmd = &cobra.Command{
		Use: "create RUNTIME", Short: "Create a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeCreate,
	}
	notebooksRuntimeDeleteCmd = &cobra.Command{
		Use: "delete RUNTIME", Short: "Delete a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeDelete,
	}
	notebooksRuntimeDescribeCmd = &cobra.Command{
		Use: "describe RUNTIME", Short: "Describe a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeDescribe,
	}
	notebooksRuntimeListCmd = &cobra.Command{
		Use: "list", Short: "List notebook runtimes",
		Args: cobra.NoArgs, RunE: runNotebooksRuntimeList,
	}
	notebooksRuntimeDiagnoseCmd = &cobra.Command{
		Use: "diagnose RUNTIME", Short: "Run diagnostics for a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeDiagnose,
	}
	notebooksRuntimeMigrateCmd = &cobra.Command{
		Use: "migrate RUNTIME", Short: "Migrate a notebook runtime to the Workbench Instances API",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeMigrate,
	}
	notebooksRuntimeResetCmd = &cobra.Command{
		Use: "reset RUNTIME", Short: "Reset a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeReset,
	}
	notebooksRuntimeStartCmd = &cobra.Command{
		Use: "start RUNTIME", Short: "Start a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeStart,
	}
	notebooksRuntimeStopCmd = &cobra.Command{
		Use: "stop RUNTIME", Short: "Stop a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeStop,
	}
	notebooksRuntimeSwitchCmd = &cobra.Command{
		Use: "switch RUNTIME", Short: "Switch a notebook runtime to a different machine type or accelerator",
		Args: cobra.ExactArgs(1), RunE: runNotebooksRuntimeSwitch,
	}
)

func init() {
	all := []*cobra.Command{
		notebooksRuntimeCreateCmd, notebooksRuntimeDeleteCmd, notebooksRuntimeDescribeCmd,
		notebooksRuntimeListCmd,
		notebooksRuntimeDiagnoseCmd, notebooksRuntimeMigrateCmd,
		notebooksRuntimeResetCmd, notebooksRuntimeStartCmd, notebooksRuntimeStopCmd,
		notebooksRuntimeSwitchCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNotebooksRuntimeLocation, "location", "", "Notebook location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNotebooksRuntimeFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		notebooksRuntimeCreateCmd, notebooksRuntimeDiagnoseCmd,
		notebooksRuntimeMigrateCmd, notebooksRuntimeSwitchCmd,
	} {
		c.Flags().StringVar(&flagNotebooksRuntimeConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	notebooksRuntimeListCmd.Flags().Int64Var(&flagNotebooksRuntimePageSize, "page-size", 0, "Maximum results per page")

	notebooksRuntimeCmd.AddCommand(all...)
	notebooksCmd.AddCommand(notebooksRuntimeCmd)
}

func notebooksRuntimeParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagNotebooksRuntimeLocation), nil
}

func notebooksRuntimeName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := notebooksRuntimeParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/runtimes/%s", parent, id), nil
}

func runNotebooksRuntimeCreate(cmd *cobra.Command, args []string) error {
	parent, err := notebooksRuntimeParent()
	if err != nil {
		return err
	}
	body := &notebooksv1.Runtime{}
	if err := loadYAMLOrJSONInto(flagNotebooksRuntimeConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Create(parent, body).RuntimeId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notebook runtime: %w", err)
	}
	fmt.Printf("Create request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeDelete(cmd *cobra.Command, args []string) error {
	name, err := notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook runtime: %w", err)
	}
	fmt.Printf("Delete request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeDescribe(cmd *cobra.Command, args []string) error {
	name, err := notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Runtimes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notebook runtime: %w", err)
	}
	return emitFormatted(got, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeList(cmd *cobra.Command, args []string) error {
	parent, err := notebooksRuntimeParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*notebooksv1.Runtime
	pageToken := ""
	for {
		call := svc.Projects.Locations.Runtimes.List(parent).Context(ctx)
		if flagNotebooksRuntimePageSize > 0 {
			call = call.PageSize(flagNotebooksRuntimePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook runtimes: %w", err)
		}
		all = append(all, resp.Runtimes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeDiagnose(cmd *cobra.Command, args []string) error {
	name, err := notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.DiagnoseRuntimeRequest{}
	if err := loadYAMLOrJSONInto(flagNotebooksRuntimeConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Diagnose(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("diagnosing notebook runtime: %w", err)
	}
	fmt.Printf("Diagnose request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeMigrate(cmd *cobra.Command, args []string) error {
	name, err := notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.MigrateRuntimeRequest{}
	if err := loadYAMLOrJSONInto(flagNotebooksRuntimeConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Migrate(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating notebook runtime: %w", err)
	}
	fmt.Printf("Migrate request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeReset(cmd *cobra.Command, args []string) error {
	name, err := notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Reset(name, &notebooksv1.ResetRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting notebook runtime: %w", err)
	}
	fmt.Printf("Reset request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeStart(cmd *cobra.Command, args []string) error {
	name, err := notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Start(name, &notebooksv1.StartRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting notebook runtime: %w", err)
	}
	fmt.Printf("Start request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeStop(cmd *cobra.Command, args []string) error {
	name, err := notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Stop(name, &notebooksv1.StopRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping notebook runtime: %w", err)
	}
	fmt.Printf("Stop request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksRuntimeFormat)
}

func runNotebooksRuntimeSwitch(cmd *cobra.Command, args []string) error {
	name, err := notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.SwitchRuntimeRequest{}
	if err := loadYAMLOrJSONInto(flagNotebooksRuntimeConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NotebooksV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Switch(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("switching notebook runtime: %w", err)
	}
	fmt.Printf("Switch request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNotebooksRuntimeFormat)
}
