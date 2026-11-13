package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	testing "google.golang.org/api/testing/v1"
)

// --- gcloud firebase test android (#1233) ---
//
// Backed by the testing v1 client: TestEnvironmentCatalog for
// models/versions/locales enumeration, and ProjectsTestMatrices for
// submitting/inspecting an Android test run.

func (s *firebaseCmds) initFirebaseTestAndroidCmd() {
	s.firebaseTestAndroidModelsCmd.AddCommand(s.firebaseTestAndroidModelsListCmd)
	s.firebaseTestAndroidVersionsCmd.AddCommand(s.firebaseTestAndroidVersionsListCmd)
	s.firebaseTestAndroidLocalesCmd.AddCommand(s.firebaseTestAndroidLocalesListCmd)
	s.firebaseTestAndroidRunCmd.Flags().StringVar(&s.flagFtaConfigFile, "config-file", "", "YAML/JSON file with a TestMatrix body (required)")
	_ = s.firebaseTestAndroidRunCmd.MarkFlagRequired("config-file")
	s.firebaseTestAndroidRunCmd.Flags().StringVar(&s.flagFtaProject, "project", "", "Owning project (falls back to the resolved project)")
	s.firebaseTestAndroidRunCmd.Flags().StringVar(&s.flagFtaFormat, "format", "", "Output format")
	s.firebaseTestAndroidModelsListCmd.Flags().StringVar(&s.flagFtaFormat, "format", "", "Output format")
	s.firebaseTestAndroidVersionsListCmd.Flags().StringVar(&s.flagFtaFormat, "format", "", "Output format")
	s.firebaseTestAndroidLocalesListCmd.Flags().StringVar(&s.flagFtaFormat, "format", "", "Output format")

	s.firebaseTestAndroidCmd.AddCommand(
		s.firebaseTestAndroidModelsCmd, s.firebaseTestAndroidVersionsCmd,
		s.firebaseTestAndroidLocalesCmd, s.firebaseTestAndroidRunCmd,
	)
	s.firebaseTestCmd.AddCommand(s.firebaseTestAndroidCmd)
}

func (s *firebaseCmds) ftaCatalog(ctx context.Context) (*testing.TestEnvironmentCatalog, error) {
	project, err := s.inv.project()
	if err != nil {
		return nil, err
	}
	svc, err := s.inv.clients().Testing(ctx)
	if err != nil {
		return nil, err
	}
	return svc.TestEnvironmentCatalog.Get("ANDROID").ProjectId(project).Context(ctx).Do()
}

func (s *firebaseCmds) runFtaModelsList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cat, err := s.ftaCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching Android device catalog: %w", err)
	}
	if cat.AndroidDeviceCatalog == nil {
		return emitFormattedTo(s.inv.stdout, []any{}, s.flagFtaFormat)
	}
	return emitFormattedTo(s.inv.stdout, cat.AndroidDeviceCatalog.Models, s.flagFtaFormat)
}

func (s *firebaseCmds) runFtaVersionsList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cat, err := s.ftaCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching Android device catalog: %w", err)
	}
	if cat.AndroidDeviceCatalog == nil {
		return emitFormattedTo(s.inv.stdout, []any{}, s.flagFtaFormat)
	}
	return emitFormattedTo(s.inv.stdout, cat.AndroidDeviceCatalog.Versions, s.flagFtaFormat)
}

func (s *firebaseCmds) runFtaLocalesList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cat, err := s.ftaCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching Android device catalog: %w", err)
	}
	if cat.AndroidDeviceCatalog == nil {
		return emitFormattedTo(s.inv.stdout, []any{}, s.flagFtaFormat)
	}
	return emitFormattedTo(s.inv.stdout, cat.AndroidDeviceCatalog.RuntimeConfiguration.Locales, s.flagFtaFormat)
}

func (s *firebaseCmds) runFtaRun(cmd *cobra.Command, args []string) error {
	project := s.flagFtaProject
	if project == "" {
		p, err := s.inv.project()
		if err != nil {
			return err
		}
		project = p
	}
	body := &testing.TestMatrix{}
	if err := loadYAMLOrJSONInto(s.flagFtaConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Testing(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.TestMatrices.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting Android test matrix: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Submitted Android test matrix [%s].\n", got.TestMatrixId)
	return emitFormattedTo(s.inv.stdout, got, s.flagFtaFormat)
}
