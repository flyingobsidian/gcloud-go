package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	testing "google.golang.org/api/testing/v1"
)

// --- gcloud firebase test ios (#1234) ---

func (s *firebaseCmds) initFirebaseTestIosCmd() {
	s.firebaseTestIosModelsCmd.AddCommand(s.firebaseTestIosModelsListCmd)
	s.firebaseTestIosVersionsCmd.AddCommand(s.firebaseTestIosVersionsListCmd)
	s.firebaseTestIosLocalesCmd.AddCommand(s.firebaseTestIosLocalesListCmd)
	s.firebaseTestIosRunCmd.Flags().StringVar(&s.flagFtiConfigFile, "config-file", "", "YAML/JSON file with a TestMatrix body (required)")
	_ = s.firebaseTestIosRunCmd.MarkFlagRequired("config-file")
	s.firebaseTestIosRunCmd.Flags().StringVar(&s.flagFtiProject, "project", "", "Owning project (falls back to the resolved project)")
	s.firebaseTestIosRunCmd.Flags().StringVar(&s.flagFtiFormat, "format", "", "Output format")
	s.firebaseTestIosModelsListCmd.Flags().StringVar(&s.flagFtiFormat, "format", "", "Output format")
	s.firebaseTestIosVersionsListCmd.Flags().StringVar(&s.flagFtiFormat, "format", "", "Output format")
	s.firebaseTestIosLocalesListCmd.Flags().StringVar(&s.flagFtiFormat, "format", "", "Output format")

	s.firebaseTestIosCmd.AddCommand(
		s.firebaseTestIosModelsCmd, s.firebaseTestIosVersionsCmd,
		s.firebaseTestIosLocalesCmd, s.firebaseTestIosRunCmd,
	)
	s.firebaseTestCmd.AddCommand(s.firebaseTestIosCmd)
}

func (s *firebaseCmds) ftiCatalog(ctx context.Context) (*testing.TestEnvironmentCatalog, error) {
	project, err := s.inv.project()
	if err != nil {
		return nil, err
	}
	svc, err := s.inv.clients().Testing(ctx)
	if err != nil {
		return nil, err
	}
	return svc.TestEnvironmentCatalog.Get("IOS").ProjectId(project).Context(ctx).Do()
}

func (s *firebaseCmds) runFtiModelsList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cat, err := s.ftiCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching iOS device catalog: %w", err)
	}
	if cat.IosDeviceCatalog == nil {
		return emitFormattedTo(s.inv.stdout, []any{}, s.flagFtiFormat)
	}
	return emitFormattedTo(s.inv.stdout, cat.IosDeviceCatalog.Models, s.flagFtiFormat)
}

func (s *firebaseCmds) runFtiVersionsList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cat, err := s.ftiCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching iOS device catalog: %w", err)
	}
	if cat.IosDeviceCatalog == nil {
		return emitFormattedTo(s.inv.stdout, []any{}, s.flagFtiFormat)
	}
	return emitFormattedTo(s.inv.stdout, cat.IosDeviceCatalog.Versions, s.flagFtiFormat)
}

func (s *firebaseCmds) runFtiLocalesList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	cat, err := s.ftiCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching iOS device catalog: %w", err)
	}
	if cat.IosDeviceCatalog == nil || cat.IosDeviceCatalog.RuntimeConfiguration == nil {
		return emitFormattedTo(s.inv.stdout, []any{}, s.flagFtiFormat)
	}
	return emitFormattedTo(s.inv.stdout, cat.IosDeviceCatalog.RuntimeConfiguration.Locales, s.flagFtiFormat)
}

func (s *firebaseCmds) runFtiRun(cmd *cobra.Command, args []string) error {
	project := s.flagFtiProject
	if project == "" {
		p, err := s.inv.project()
		if err != nil {
			return err
		}
		project = p
	}
	body := &testing.TestMatrix{}
	if err := loadYAMLOrJSONInto(s.flagFtiConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Testing(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.TestMatrices.Create(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting iOS test matrix: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Submitted iOS test matrix [%s].\n", got.TestMatrixId)
	return emitFormattedTo(s.inv.stdout, got, s.flagFtiFormat)
}
