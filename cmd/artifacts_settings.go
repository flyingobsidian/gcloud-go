package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// --- gcloud artifacts settings (#1082) ---

var artifactsSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage Artifact Registry per-location settings",
}

var (
	flagArtSettingsLocation string
	flagArtSettingsFormat   string
)

var artifactsSettingsDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe the settings for a location",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsSettingsDescribe,
}

var artifactsSettingsEnableRedirCmd = &cobra.Command{
	Use:   "enable-upgrade-redirection",
	Short: "Enable redirection of legacy container registry traffic to Artifact Registry",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsSettingsEnableRedir,
}

var artifactsSettingsDisableRedirCmd = &cobra.Command{
	Use:   "disable-upgrade-redirection",
	Short: "Disable redirection of legacy container registry traffic",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsSettingsDisableRedir,
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsSettingsDescribeCmd, artifactsSettingsEnableRedirCmd, artifactsSettingsDisableRedirCmd,
	} {
		c.Flags().StringVar(&flagArtSettingsLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagArtSettingsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	artifactsSettingsCmd.AddCommand(
		artifactsSettingsDescribeCmd, artifactsSettingsEnableRedirCmd, artifactsSettingsDisableRedirCmd,
	)
	artifactsCmd.AddCommand(artifactsSettingsCmd)
}

func artSettingsName(project string) string {
	return artLocationParent(project, flagArtSettingsLocation) + "/settings"
}

func runArtifactsSettingsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var out map[string]any
	if err := artifactsRest.do(context.Background(), http.MethodGet, "/"+artSettingsName(project), nil, nil, &out); err != nil {
		return fmt.Errorf("describing settings: %w", err)
	}
	return emitFormatted(out, flagArtSettingsFormat)
}

func runArtifactsSettingsEnableRedir(cmd *cobra.Command, args []string) error {
	return patchSettingsRedirection("REDIRECTION_FROM_GCR_IO_ENABLED")
}

func runArtifactsSettingsDisableRedir(cmd *cobra.Command, args []string) error {
	return patchSettingsRedirection("REDIRECTION_FROM_GCR_IO_DISABLED")
}

func patchSettingsRedirection(state string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := map[string]any{"legacyRedirectionState": state}
	q := url.Values{}
	q.Set("updateMask", "legacyRedirectionState")
	var out map[string]any
	if err := artifactsRest.do(context.Background(), http.MethodPatch, "/"+artSettingsName(project), q, body, &out); err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}
	return emitFormatted(out, flagArtSettingsFormat)
}
