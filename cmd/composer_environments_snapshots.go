package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	composer "google.golang.org/api/composer/v1"
)

// --- gcloud composer environments snapshots (subgroup of #1502) ---

func (s *composerCmds) initComposerEnvironmentsSnapshots() {
	for _, c := range []*cobra.Command{s.composerEnvSnapSaveCmd, s.composerEnvSnapLoadCmd} {
		c.Flags().StringVar(&s.flagComposerEnvLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagComposerEnvFormat, "format", "", "Output format")
	}
	s.composerEnvSnapSaveCmd.Flags().StringVar(&s.flagComposerEnvSnapLocation, "snapshot-location", "",
		"Cloud Storage prefix where the snapshot should be written (required, e.g. gs://bucket/snapshots)")
	_ = s.composerEnvSnapSaveCmd.MarkFlagRequired("snapshot-location")

	s.composerEnvSnapLoadCmd.Flags().StringVar(&s.flagComposerEnvSnapPath, "snapshot-path", "",
		"Cloud Storage path of the snapshot to load (required, e.g. gs://bucket/snapshots/env_ts)")
	_ = s.composerEnvSnapLoadCmd.MarkFlagRequired("snapshot-path")
	s.composerEnvSnapLoadCmd.Flags().BoolVar(&s.flagComposerEnvSkipOverrides, "skip-airflow-overrides", false,
		"Skip setting Airflow overrides when loading")
	s.composerEnvSnapLoadCmd.Flags().BoolVar(&s.flagComposerEnvSkipEnv, "skip-environment-variables", false,
		"Skip setting environment variables when loading")
	s.composerEnvSnapLoadCmd.Flags().BoolVar(&s.flagComposerEnvSkipData, "skip-gcs-data-copying", false,
		"Skip copying GCS data when loading")
	s.composerEnvSnapLoadCmd.Flags().BoolVar(&s.flagComposerEnvSkipPypi, "skip-pypi-packages-installation", false,
		"Skip installing Pypi packages when loading")

	s.composerEnvSnapshotsCmd.AddCommand(s.composerEnvSnapSaveCmd, s.composerEnvSnapLoadCmd)
}

func (s *composerCmds) runComposerEnvSnapSave(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.SaveSnapshot(name, &composer.SaveSnapshotRequest{
		SnapshotLocation: s.flagComposerEnvSnapLocation,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("saving snapshot: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Save-snapshot request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerEnvSnapLoad(cmd *cobra.Command, args []string) error {
	name, err := s.composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.LoadSnapshot(name, &composer.LoadSnapshotRequest{
		SnapshotPath:                    s.flagComposerEnvSnapPath,
		SkipAirflowOverridesSetting:     s.flagComposerEnvSkipOverrides,
		SkipEnvironmentVariablesSetting: s.flagComposerEnvSkipEnv,
		SkipGcsDataCopying:              s.flagComposerEnvSkipData,
		SkipPypiPackagesInstallation:    s.flagComposerEnvSkipPypi,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Load-snapshot request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerEnvFormat)
}
