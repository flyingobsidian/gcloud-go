package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcGoogleChannels() {
	for _, c := range []*cobra.Command{grp.evGChanDescribeCmd, grp.evGChanUpdateCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvGChanLocation, true)
	}
	grp.evGChanUpdateCmd.Flags().StringVar(&grp.flagEvGChanConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the GoogleChannelConfig message body (required)")
	_ = grp.evGChanUpdateCmd.MarkFlagRequired("config-file")
	grp.evGChanUpdateCmd.Flags().StringVar(&grp.flagEvGChanUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	grp.evGChanDescribeCmd.Flags().StringVar(&grp.flagEvGChanFormat, "format", "", "Output format")

	grp.eventarcGoogleChannelsCmd.AddCommand(grp.evGChanDescribeCmd, grp.evGChanUpdateCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcGoogleChannelsCmd)
}

func googleChannelConfigName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/googleChannelConfig", project, location)
}

func (grp *eventarcCmds) runEvGChanDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.GetGoogleChannelConfig(googleChannelConfigName(project, grp.flagEvGChanLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Google Channel config: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, cfg, grp.flagEvGChanFormat)
}

func (grp *eventarcCmds) runEvGChanUpdate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	cfg := &eventarc.GoogleChannelConfig{}
	if err := loadYAMLOrJSONInto(grp.flagEvGChanConfigFile, cfg); err != nil {
		return err
	}
	mask := grp.flagEvGChanUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cfg))
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	out, err := svc.Projects.Locations.UpdateGoogleChannelConfig(googleChannelConfigName(project, grp.flagEvGChanLocation), cfg).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating Google Channel config: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, out, "")
}
