package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging settings (#922) ---
//
// Settings are readable at every scope but only writable at the organization
// or folder level (the underlying API only exposes UpdateSettings for those
// two). Attempts to update a project or billing-account scope are rejected.

var loggingSettingsCmd = &cobra.Command{Use: "settings", Short: "Manage Logs Router settings"}

var (
	flagLogSettingsKmsKey             string
	flagLogSettingsStorageLocation    string
	flagLogSettingsDisableDefaultSink bool
)

var (
	loggingSettingsDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe Logs Router settings",
		Args: cobra.NoArgs, RunE: runLogSettingsDescribe,
	}
	loggingSettingsUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update Logs Router settings",
		Args: cobra.NoArgs, RunE: runLogSettingsUpdate,
	}
)

func settingsResourceName(parent string) string {
	return parent + "/settings"
}

func runLogSettingsDescribe(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := settingsResourceName(parent)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.Settings
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.GetSettings(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.GetSettings(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.GetSettings(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.GetSettings(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing settings: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogSettingsUpdate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if scope != "organizations" && scope != "folders" {
		return fmt.Errorf("updating settings is only supported at organization or folder scope")
	}
	name := settingsResourceName(parent)
	body := &logging.Settings{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	if flagLogSettingsKmsKey != "" {
		body.KmsKeyName = flagLogSettingsKmsKey
	}
	if flagLogSettingsStorageLocation != "" {
		body.StorageLocation = flagLogSettingsStorageLocation
	}
	if cmd.Flags().Changed("disable-default-sink") {
		body.DisableDefaultSink = flagLogSettingsDisableDefaultSink
		body.ForceSendFields = append(body.ForceSendFields, "DisableDefaultSink")
	}
	mask := loggingResolveMask(body)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.Settings
	switch scope {
	case "organizations":
		got, err = svc.Organizations.UpdateSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.UpdateSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func init() {
	all := []*cobra.Command{loggingSettingsDescribeCmd, loggingSettingsUpdateCmd}
	addLogScopeFlags(all...)
	addLogFormatFlag(all...)
	loggingSettingsUpdateCmd.Flags().StringVar(&flagLogSettingsKmsKey, "kms-key-name", "", "CMEK KMS key name for the Logs Router")
	loggingSettingsUpdateCmd.Flags().StringVar(&flagLogSettingsStorageLocation, "storage-location", "", "Default storage location for _Default and _Required buckets")
	loggingSettingsUpdateCmd.Flags().BoolVar(&flagLogSettingsDisableDefaultSink, "disable-default-sink", false, "Disable the _Default sink for newly-created projects and folders")
	loggingSettingsUpdateCmd.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the Settings body")
	loggingSettingsUpdateCmd.Flags().StringVar(&flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	loggingSettingsCmd.AddCommand(all...)
	loggingCmd.AddCommand(loggingSettingsCmd)
}
