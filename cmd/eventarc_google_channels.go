package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcGoogleChannelsCmd = &cobra.Command{
	Use:   "google-channels",
	Short: "Manage the Google-managed Eventarc channel configuration",
}

var evGChanDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe the Google Channel configuration",
	Args:  cobra.NoArgs,
	RunE:  runEvGChanDescribe,
}

var evGChanUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the Google Channel configuration from a --config-file",
	Args:  cobra.NoArgs,
	RunE:  runEvGChanUpdate,
}

var (
	flagEvGChanLocation   string
	flagEvGChanConfigFile string
	flagEvGChanUpdateMask string
	flagEvGChanFormat     string
)

func init() {
	for _, c := range []*cobra.Command{evGChanDescribeCmd, evGChanUpdateCmd} {
		eventarcAddRegionFlag(c, &flagEvGChanLocation, true)
	}
	evGChanUpdateCmd.Flags().StringVar(&flagEvGChanConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the GoogleChannelConfig message body (required)")
	_ = evGChanUpdateCmd.MarkFlagRequired("config-file")
	evGChanUpdateCmd.Flags().StringVar(&flagEvGChanUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	evGChanDescribeCmd.Flags().StringVar(&flagEvGChanFormat, "format", "", "Output format")

	eventarcGoogleChannelsCmd.AddCommand(evGChanDescribeCmd, evGChanUpdateCmd)
	eventarcCmd.AddCommand(eventarcGoogleChannelsCmd)
}

func googleChannelConfigName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/googleChannelConfig", project, location)
}

func runEvGChanDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.GetGoogleChannelConfig(googleChannelConfigName(project, flagEvGChanLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Google Channel config: %w", err)
	}
	return emitFormatted(cfg, flagEvGChanFormat)
}

func runEvGChanUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cfg := &eventarc.GoogleChannelConfig{}
	if err := loadYAMLOrJSONInto(flagEvGChanConfigFile, cfg); err != nil {
		return err
	}
	mask := flagEvGChanUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cfg))
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	out, err := svc.Projects.Locations.UpdateGoogleChannelConfig(googleChannelConfigName(project, flagEvGChanLocation), cfg).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating Google Channel config: %w", err)
	}
	return emitFormatted(out, "")
}
