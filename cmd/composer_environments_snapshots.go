package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	composer "google.golang.org/api/composer/v1"
)

// --- gcloud composer environments snapshots (subgroup of #1502) ---

var composerEnvSnapshotsCmd = &cobra.Command{Use: "snapshots", Short: "Save and load Composer environment snapshots"}

var (
	composerEnvSnapSaveCmd = &cobra.Command{
		Use: "save ENVIRONMENT", Short: "Save a snapshot of a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvSnapSave,
	}
	composerEnvSnapLoadCmd = &cobra.Command{
		Use: "load ENVIRONMENT", Short: "Load a snapshot into a Composer environment",
		Args: cobra.ExactArgs(1), RunE: runComposerEnvSnapLoad,
	}
)

func init() {
	for _, c := range []*cobra.Command{composerEnvSnapSaveCmd, composerEnvSnapLoadCmd} {
		c.Flags().StringVar(&flagComposerEnvLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagComposerEnvFormat, "format", "", "Output format")
	}
	composerEnvSnapSaveCmd.Flags().StringVar(&flagComposerEnvSnapLocation, "snapshot-location", "",
		"Cloud Storage prefix where the snapshot should be written (required, e.g. gs://bucket/snapshots)")
	_ = composerEnvSnapSaveCmd.MarkFlagRequired("snapshot-location")

	composerEnvSnapLoadCmd.Flags().StringVar(&flagComposerEnvSnapPath, "snapshot-path", "",
		"Cloud Storage path of the snapshot to load (required, e.g. gs://bucket/snapshots/env_ts)")
	_ = composerEnvSnapLoadCmd.MarkFlagRequired("snapshot-path")
	composerEnvSnapLoadCmd.Flags().BoolVar(&flagComposerEnvSkipOverrides, "skip-airflow-overrides", false,
		"Skip setting Airflow overrides when loading")
	composerEnvSnapLoadCmd.Flags().BoolVar(&flagComposerEnvSkipEnv, "skip-environment-variables", false,
		"Skip setting environment variables when loading")
	composerEnvSnapLoadCmd.Flags().BoolVar(&flagComposerEnvSkipData, "skip-gcs-data-copying", false,
		"Skip copying GCS data when loading")
	composerEnvSnapLoadCmd.Flags().BoolVar(&flagComposerEnvSkipPypi, "skip-pypi-packages-installation", false,
		"Skip installing Pypi packages when loading")

	composerEnvSnapshotsCmd.AddCommand(composerEnvSnapSaveCmd, composerEnvSnapLoadCmd)
}

func runComposerEnvSnapSave(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.SaveSnapshot(name, &composer.SaveSnapshotRequest{
		SnapshotLocation: flagComposerEnvSnapLocation,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("saving snapshot: %w", err)
	}
	fmt.Printf("Save-snapshot request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagComposerEnvFormat)
}

func runComposerEnvSnapLoad(cmd *cobra.Command, args []string) error {
	name, err := composerEnvResolvedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.LoadSnapshot(name, &composer.LoadSnapshotRequest{
		SnapshotPath:                    flagComposerEnvSnapPath,
		SkipAirflowOverridesSetting:     flagComposerEnvSkipOverrides,
		SkipEnvironmentVariablesSetting: flagComposerEnvSkipEnv,
		SkipGcsDataCopying:              flagComposerEnvSkipData,
		SkipPypiPackagesInstallation:    flagComposerEnvSkipPypi,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}
	fmt.Printf("Load-snapshot request issued for environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagComposerEnvFormat)
}
