package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap settings (#1067) ---

var iapSettingsCmd = &cobra.Command{Use: "settings", Short: "Manage IAP settings"}

var (
	flagIapSettingsFormat       string
	flagIapSettingsResourceName string
	flagIapSettingsConfigFile   string
	flagIapSettingsUpdateMask   string
)

var (
	iapSettingsGetCmd = &cobra.Command{
		Use: "get", Short: "Get IAP settings for an IAP-protected resource",
		Args: cobra.NoArgs, RunE: runIapSettingsGet,
	}
	iapSettingsSetCmd = &cobra.Command{
		Use: "set", Short: "Set IAP settings for an IAP-protected resource",
		Args: cobra.NoArgs, RunE: runIapSettingsSet,
	}
)

func init() {
	for _, c := range []*cobra.Command{iapSettingsGetCmd, iapSettingsSetCmd} {
		c.Flags().StringVar(&flagIapSettingsFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagIapSettingsResourceName, "resource-name", "",
			"Full name of the IAP-protected resource (required)")
		_ = c.MarkFlagRequired("resource-name")
	}
	iapSettingsSetCmd.Flags().StringVar(&flagIapSettingsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the IapSettings body (required)")
	_ = iapSettingsSetCmd.MarkFlagRequired("config-file")
	iapSettingsSetCmd.Flags().StringVar(&flagIapSettingsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	iapSettingsCmd.AddCommand(iapSettingsGetCmd, iapSettingsSetCmd)
	iapCmd.AddCommand(iapSettingsCmd)
}

func runIapSettingsGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.V1.GetIapSettings(flagIapSettingsResourceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAP settings: %w", err)
	}
	return emitFormatted(got, flagIapSettingsFormat)
}

func runIapSettingsSet(cmd *cobra.Command, args []string) error {
	body := &iap.IapSettings{}
	if err := loadYAMLOrJSONInto(flagIapSettingsConfigFile, body); err != nil {
		return err
	}
	mask := flagIapSettingsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.V1.UpdateIapSettings(flagIapSettingsResourceName, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("setting IAP settings: %w", err)
	}
	fmt.Printf("Updated IAP settings for [%s].\n", flagIapSettingsResourceName)
	return emitFormatted(got, flagIapSettingsFormat)
}
