package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	policysimulator "google.golang.org/api/policysimulator/v1"
)

// --- gcloud iam simulator (#1013) ---
//
// The simulator subgroup is a thin wrapper around the policysimulator v1
// "replays" resource — a Replay is one simulation run.

var iamSimulatorCmd = &cobra.Command{Use: "simulator", Short: "Simulate IAM policy changes"}
var iamSimulatorReplaysCmd = &cobra.Command{Use: "replays", Short: "Manage IAM policy simulation Replays"}

var (
	flagIamSimScope      string
	flagIamSimLocation   string
	flagIamSimConfigFile string
	flagIamSimFormat     string
)

var (
	iamSimCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a Replay (loads the Replay body from --config-file)",
		Args: cobra.NoArgs, RunE: runIamSimCreate,
	}
	iamSimDescribeCmd = &cobra.Command{
		Use: "describe REPLAY", Short: "Describe a Replay",
		Args: cobra.ExactArgs(1), RunE: runIamSimDescribe,
	}
)

func init() {
	for _, c := range []*cobra.Command{iamSimCreateCmd, iamSimDescribeCmd} {
		c.Flags().StringVar(&flagIamSimScope, "scope", "", "Owning scope, e.g. projects/PROJECT, folders/FOLDER, or organizations/ORG (required)")
		_ = c.MarkFlagRequired("scope")
		c.Flags().StringVar(&flagIamSimLocation, "location", "global", "Location (defaults to global)")
		c.Flags().StringVar(&flagIamSimFormat, "format", "", "Output format")
	}
	iamSimCreateCmd.Flags().StringVar(&flagIamSimConfigFile, "config-file", "", "YAML/JSON file with the Replay body (required)")
	_ = iamSimCreateCmd.MarkFlagRequired("config-file")

	iamSimulatorReplaysCmd.AddCommand(iamSimCreateCmd, iamSimDescribeCmd)
	iamSimulatorCmd.AddCommand(iamSimulatorReplaysCmd)
	iamCmd.AddCommand(iamSimulatorCmd)
}

func iamSimReplayParent() string {
	return fmt.Sprintf("%s/locations/%s", flagIamSimScope, flagIamSimLocation)
}

func iamSimReplayName(id string) string {
	if strings.Contains(id, "/replays/") {
		return id
	}
	return fmt.Sprintf("%s/replays/%s", iamSimReplayParent(), id)
}

func runIamSimCreate(cmd *cobra.Command, args []string) error {
	body := &policysimulator.GoogleCloudPolicysimulatorV1Replay{}
	if err := loadYAMLOrJSONInto(flagIamSimConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PolicySimulatorService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := iamSimReplayParent()
	var op *policysimulator.GoogleLongrunningOperation
	switch {
	case strings.HasPrefix(flagIamSimScope, "projects/"):
		op, err = svc.Projects.Locations.Replays.Create(parent, body).Context(ctx).Do()
	case strings.HasPrefix(flagIamSimScope, "folders/"):
		op, err = svc.Folders.Locations.Replays.Create(parent, body).Context(ctx).Do()
	case strings.HasPrefix(flagIamSimScope, "organizations/"):
		op, err = svc.Organizations.Locations.Replays.Create(parent, body).Context(ctx).Do()
	default:
		return fmt.Errorf("--scope must start with projects/, folders/, or organizations/")
	}
	if err != nil {
		return fmt.Errorf("creating replay: %w", err)
	}
	fmt.Printf("Create replay initiated (operation: %s).\n", op.Name)
	return emitFormatted(op, flagIamSimFormat)
}

func runIamSimDescribe(cmd *cobra.Command, args []string) error {
	name := iamSimReplayName(args[0])
	ctx := context.Background()
	svc, err := gcp.PolicySimulatorService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var got *policysimulator.GoogleCloudPolicysimulatorV1Replay
	switch {
	case strings.HasPrefix(flagIamSimScope, "projects/"):
		got, err = svc.Projects.Locations.Replays.Get(name).Context(ctx).Do()
	case strings.HasPrefix(flagIamSimScope, "folders/"):
		got, err = svc.Folders.Locations.Replays.Get(name).Context(ctx).Do()
	case strings.HasPrefix(flagIamSimScope, "organizations/"):
		got, err = svc.Organizations.Locations.Replays.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("--scope must start with projects/, folders/, or organizations/")
	}
	if err != nil {
		return fmt.Errorf("describing replay: %w", err)
	}
	return emitFormatted(got, flagIamSimFormat)
}
