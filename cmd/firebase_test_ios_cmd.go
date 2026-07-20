package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	testing "google.golang.org/api/testing/v1"
)

// --- gcloud firebase test ios (#1234) ---

var firebaseTestIosCmd = &cobra.Command{Use: "ios", Short: "iOS application testing"}

var (
	flagFtiFormat     string
	flagFtiConfigFile string
	flagFtiProject    string
)

var (
	firebaseTestIosModelsCmd = &cobra.Command{Use: "models", Short: "Manage supported iOS device models"}
	firebaseTestIosModelsListCmd = &cobra.Command{
		Use: "list", Short: "List supported iOS device models",
		Args: cobra.NoArgs, RunE: runFtiModelsList,
	}
	firebaseTestIosVersionsCmd = &cobra.Command{Use: "versions", Short: "Manage supported iOS versions"}
	firebaseTestIosVersionsListCmd = &cobra.Command{
		Use: "list", Short: "List supported iOS versions",
		Args: cobra.NoArgs, RunE: runFtiVersionsList,
	}
	firebaseTestIosLocalesCmd = &cobra.Command{Use: "locales", Short: "Manage supported iOS locales"}
	firebaseTestIosLocalesListCmd = &cobra.Command{
		Use: "list", Short: "List supported iOS locales",
		Args: cobra.NoArgs, RunE: runFtiLocalesList,
	}
	firebaseTestIosRunCmd = &cobra.Command{
		Use: "run", Short: "Submit an iOS test matrix (loads TestMatrix body from --config-file)",
		Args: cobra.NoArgs, RunE: runFtiRun,
	}
)

func init() {
	firebaseTestIosModelsCmd.AddCommand(firebaseTestIosModelsListCmd)
	firebaseTestIosVersionsCmd.AddCommand(firebaseTestIosVersionsListCmd)
	firebaseTestIosLocalesCmd.AddCommand(firebaseTestIosLocalesListCmd)
	firebaseTestIosRunCmd.Flags().StringVar(&flagFtiConfigFile, "config-file", "", "YAML/JSON file with a TestMatrix body (required)")
	_ = firebaseTestIosRunCmd.MarkFlagRequired("config-file")
	firebaseTestIosRunCmd.Flags().StringVar(&flagFtiProject, "project", "", "Owning project (falls back to the resolved project)")
	firebaseTestIosRunCmd.Flags().StringVar(&flagFtiFormat, "format", "", "Output format")
	firebaseTestIosModelsListCmd.Flags().StringVar(&flagFtiFormat, "format", "", "Output format")
	firebaseTestIosVersionsListCmd.Flags().StringVar(&flagFtiFormat, "format", "", "Output format")
	firebaseTestIosLocalesListCmd.Flags().StringVar(&flagFtiFormat, "format", "", "Output format")

	firebaseTestIosCmd.AddCommand(
		firebaseTestIosModelsCmd, firebaseTestIosVersionsCmd,
		firebaseTestIosLocalesCmd, firebaseTestIosRunCmd,
	)
	firebaseTestCmd.AddCommand(firebaseTestIosCmd)
}

func ftiCatalog(ctx context.Context) (*testing.TestEnvironmentCatalog, error) {
	project, err := resolveProject()
	if err != nil {
		return nil, err
	}
	svc, err := gcp.TestingService(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	return svc.TestEnvironmentCatalog.Get("IOS").ProjectId(project).Context(ctx).Do()
}

func runFtiModelsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cat, err := ftiCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching iOS device catalog: %w", err)
	}
	if cat.IosDeviceCatalog == nil {
		return emitFormatted([]any{}, flagFtiFormat)
	}
	return emitFormatted(cat.IosDeviceCatalog.Models, flagFtiFormat)
}

func runFtiVersionsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cat, err := ftiCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching iOS device catalog: %w", err)
	}
	if cat.IosDeviceCatalog == nil {
		return emitFormatted([]any{}, flagFtiFormat)
	}
	return emitFormatted(cat.IosDeviceCatalog.Versions, flagFtiFormat)
}

func runFtiLocalesList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cat, err := ftiCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching iOS device catalog: %w", err)
	}
	if cat.IosDeviceCatalog == nil || cat.IosDeviceCatalog.RuntimeConfiguration == nil {
		return emitFormatted([]any{}, flagFtiFormat)
	}
	return emitFormatted(cat.IosDeviceCatalog.RuntimeConfiguration.Locales, flagFtiFormat)
}

func runFtiRun(cmd *cobra.Command, args []string) error {
	project := flagFtiProject
	if project == "" {
		p, err := resolveProject()
		if err != nil {
			return err
		}
		project = p
	}
	body := &testing.TestMatrix{}
	if err := loadYAMLOrJSONInto(flagFtiConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TestingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.TestMatrices.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting iOS test matrix: %w", err)
	}
	fmt.Printf("Submitted iOS test matrix [%s].\n", got.TestMatrixId)
	return emitFormatted(got, flagFtiFormat)
}
