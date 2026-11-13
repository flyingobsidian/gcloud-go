package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	policysimulator "google.golang.org/api/policysimulator/v1"
)

// --- gcloud iam simulator (#1013) ---
//
// The simulator subgroup is a thin wrapper around the policysimulator v1
// "replays" resource — a Replay is one simulation run.

func (s *iamCmds) initIamSimulator() {
	for _, c := range []*cobra.Command{s.iamSimCreateCmd, s.iamSimDescribeCmd} {
		c.Flags().StringVar(&s.flagIamSimScope, "scope", "", "Owning scope, e.g. projects/PROJECT, folders/FOLDER, or organizations/ORG (required)")
		_ = c.MarkFlagRequired("scope")
		c.Flags().StringVar(&s.flagIamSimLocation, "location", "global", "Location (defaults to global)")
		c.Flags().StringVar(&s.flagIamSimFormat, "format", "", "Output format")
	}
	s.iamSimCreateCmd.Flags().StringVar(&s.flagIamSimConfigFile, "config-file", "", "YAML/JSON file with the Replay body (required)")
	_ = s.iamSimCreateCmd.MarkFlagRequired("config-file")

	s.iamSimulatorReplaysCmd.AddCommand(s.iamSimCreateCmd, s.iamSimDescribeCmd)
	s.iamSimulatorCmd.AddCommand(s.iamSimulatorReplaysCmd)
	s.iamCmd.AddCommand(s.iamSimulatorCmd)
}

func (s *iamCmds) iamSimReplayParent() string {
	return fmt.Sprintf("%s/locations/%s", s.flagIamSimScope, s.flagIamSimLocation)
}

func (s *iamCmds) iamSimReplayName(id string) string {
	if strings.Contains(id, "/replays/") {
		return id
	}
	return fmt.Sprintf("%s/replays/%s", s.iamSimReplayParent(), id)
}

func (s *iamCmds) runIamSimCreate(cmd *cobra.Command, args []string) error {
	body := &policysimulator.GoogleCloudPolicysimulatorV1Replay{}
	if err := loadYAMLOrJSONInto(s.flagIamSimConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PolicySimulator(ctx)
	if err != nil {
		return err
	}
	parent := s.iamSimReplayParent()
	var op *policysimulator.GoogleLongrunningOperation
	switch {
	case strings.HasPrefix(s.flagIamSimScope, "projects/"):
		op, err = svc.Projects.Locations.Replays.Create(parent, body).Context(ctx).Do()
	case strings.HasPrefix(s.flagIamSimScope, "folders/"):
		op, err = svc.Folders.Locations.Replays.Create(parent, body).Context(ctx).Do()
	case strings.HasPrefix(s.flagIamSimScope, "organizations/"):
		op, err = svc.Organizations.Locations.Replays.Create(parent, body).Context(ctx).Do()
	default:
		return fmt.Errorf("--scope must start with projects/, folders/, or organizations/")
	}
	if err != nil {
		return fmt.Errorf("creating replay: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create replay initiated (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamSimFormat)
}

func (s *iamCmds) runIamSimDescribe(cmd *cobra.Command, args []string) error {
	name := s.iamSimReplayName(args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().PolicySimulator(ctx)
	if err != nil {
		return err
	}
	var got *policysimulator.GoogleCloudPolicysimulatorV1Replay
	switch {
	case strings.HasPrefix(s.flagIamSimScope, "projects/"):
		got, err = svc.Projects.Locations.Replays.Get(name).Context(ctx).Do()
	case strings.HasPrefix(s.flagIamSimScope, "folders/"):
		got, err = svc.Folders.Locations.Replays.Get(name).Context(ctx).Do()
	case strings.HasPrefix(s.flagIamSimScope, "organizations/"):
		got, err = svc.Organizations.Locations.Replays.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("--scope must start with projects/, folders/, or organizations/")
	}
	if err != nil {
		return fmt.Errorf("describing replay: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIamSimFormat)
}
