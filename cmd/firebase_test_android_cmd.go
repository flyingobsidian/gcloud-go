package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	testing "google.golang.org/api/testing/v1"
)

// --- gcloud firebase test android (#1233) ---
//
// Backed by the testing v1 client: TestEnvironmentCatalog for
// models/versions/locales enumeration, and ProjectsTestMatrices for
// submitting/inspecting an Android test run.

var firebaseTestAndroidCmd = &cobra.Command{Use: "android", Short: "Android application testing"}

var (
	flagFtaFormat     string
	flagFtaConfigFile string
	flagFtaProject    string
)

var (
	firebaseTestAndroidModelsCmd = &cobra.Command{
		Use: "models", Short: "Manage supported Android device models",
	}
	firebaseTestAndroidModelsListCmd = &cobra.Command{
		Use: "list", Short: "List supported Android device models",
		Args: cobra.NoArgs, RunE: runFtaModelsList,
	}
	firebaseTestAndroidVersionsCmd = &cobra.Command{
		Use: "versions", Short: "Manage supported Android OS versions",
	}
	firebaseTestAndroidVersionsListCmd = &cobra.Command{
		Use: "list", Short: "List supported Android OS versions",
		Args: cobra.NoArgs, RunE: runFtaVersionsList,
	}
	firebaseTestAndroidLocalesCmd = &cobra.Command{
		Use: "locales", Short: "Manage supported Android locales",
	}
	firebaseTestAndroidLocalesListCmd = &cobra.Command{
		Use: "locales", Short: "List supported Android locales",
		Args: cobra.NoArgs, RunE: runFtaLocalesList,
	}
	firebaseTestAndroidRunCmd = &cobra.Command{
		Use: "run", Short: "Submit an Android test matrix (loads TestMatrix body from --config-file)",
		Args: cobra.NoArgs, RunE: runFtaRun,
	}
)

func init() {
	firebaseTestAndroidModelsCmd.AddCommand(firebaseTestAndroidModelsListCmd)
	firebaseTestAndroidVersionsCmd.AddCommand(firebaseTestAndroidVersionsListCmd)
	firebaseTestAndroidLocalesCmd.AddCommand(firebaseTestAndroidLocalesListCmd)
	firebaseTestAndroidRunCmd.Flags().StringVar(&flagFtaConfigFile, "config-file", "", "YAML/JSON file with a TestMatrix body (required)")
	_ = firebaseTestAndroidRunCmd.MarkFlagRequired("config-file")
	firebaseTestAndroidRunCmd.Flags().StringVar(&flagFtaProject, "project", "", "Owning project (falls back to the resolved project)")
	firebaseTestAndroidRunCmd.Flags().StringVar(&flagFtaFormat, "format", "", "Output format")
	firebaseTestAndroidModelsListCmd.Flags().StringVar(&flagFtaFormat, "format", "", "Output format")
	firebaseTestAndroidVersionsListCmd.Flags().StringVar(&flagFtaFormat, "format", "", "Output format")
	firebaseTestAndroidLocalesListCmd.Flags().StringVar(&flagFtaFormat, "format", "", "Output format")

	firebaseTestAndroidCmd.AddCommand(
		firebaseTestAndroidModelsCmd, firebaseTestAndroidVersionsCmd,
		firebaseTestAndroidLocalesCmd, firebaseTestAndroidRunCmd,
	)
	firebaseTestCmd.AddCommand(firebaseTestAndroidCmd)
}

func ftaCatalog(ctx context.Context) (*testing.TestEnvironmentCatalog, error) {
	project, err := resolveProject()
	if err != nil {
		return nil, err
	}
	svc, err := gcp.TestingService(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	return svc.TestEnvironmentCatalog.Get("ANDROID").ProjectId(project).Context(ctx).Do()
}

func runFtaModelsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cat, err := ftaCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching Android device catalog: %w", err)
	}
	if cat.AndroidDeviceCatalog == nil {
		return emitFormatted([]any{}, flagFtaFormat)
	}
	return emitFormatted(cat.AndroidDeviceCatalog.Models, flagFtaFormat)
}

func runFtaVersionsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cat, err := ftaCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching Android device catalog: %w", err)
	}
	if cat.AndroidDeviceCatalog == nil {
		return emitFormatted([]any{}, flagFtaFormat)
	}
	return emitFormatted(cat.AndroidDeviceCatalog.Versions, flagFtaFormat)
}

func runFtaLocalesList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cat, err := ftaCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching Android device catalog: %w", err)
	}
	if cat.AndroidDeviceCatalog == nil {
		return emitFormatted([]any{}, flagFtaFormat)
	}
	return emitFormatted(cat.AndroidDeviceCatalog.RuntimeConfiguration.Locales, flagFtaFormat)
}

func runFtaRun(cmd *cobra.Command, args []string) error {
	project := flagFtaProject
	if project == "" {
		p, err := resolveProject()
		if err != nil {
			return err
		}
		project = p
	}
	body := &testing.TestMatrix{}
	if err := loadYAMLOrJSONInto(flagFtaConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TestingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.TestMatrices.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting Android test matrix: %w", err)
	}
	fmt.Printf("Submitted Android test matrix [%s].\n", got.TestMatrixId)
	return emitFormatted(got, flagFtaFormat)
}
