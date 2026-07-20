package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud firebase test network-profiles (#1235) ---

var firebaseTestNetworkProfilesCmd = &cobra.Command{Use: "network-profiles", Short: "List network profiles available to Firebase Test Lab"}

var (
	flagFtnpFormat string
	flagFtnpID     string
)

var (
	firebaseTestNetworkProfilesListCmd = &cobra.Command{
		Use: "list", Short: "List network profiles",
		Args: cobra.NoArgs, RunE: runFtnpList,
	}
	firebaseTestNetworkProfilesDescribeCmd = &cobra.Command{
		Use: "describe PROFILE_ID", Short: "Describe a network profile",
		Args: cobra.ExactArgs(1), RunE: runFtnpDescribe,
	}
)

func init() {
	for _, c := range []*cobra.Command{firebaseTestNetworkProfilesListCmd, firebaseTestNetworkProfilesDescribeCmd} {
		c.Flags().StringVar(&flagFtnpFormat, "format", "", "Output format")
	}

	firebaseTestNetworkProfilesCmd.AddCommand(firebaseTestNetworkProfilesListCmd, firebaseTestNetworkProfilesDescribeCmd)
	firebaseTestCmd.AddCommand(firebaseTestNetworkProfilesCmd)
}

func runFtnpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TestingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cat, err := svc.TestEnvironmentCatalog.Get("NETWORK_CONFIGURATION").ProjectId(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching network configuration catalog: %w", err)
	}
	if cat.NetworkConfigurationCatalog == nil {
		return emitFormatted([]any{}, flagFtnpFormat)
	}
	return emitFormatted(cat.NetworkConfigurationCatalog.Configurations, flagFtnpFormat)
}

func runFtnpDescribe(cmd *cobra.Command, args []string) error {
	flagFtnpID = args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TestingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cat, err := svc.TestEnvironmentCatalog.Get("NETWORK_CONFIGURATION").ProjectId(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching network configuration catalog: %w", err)
	}
	if cat.NetworkConfigurationCatalog != nil {
		for _, nc := range cat.NetworkConfigurationCatalog.Configurations {
			if nc.Id == flagFtnpID {
				return emitFormatted(nc, flagFtnpFormat)
			}
		}
	}
	return fmt.Errorf("network profile [%s] not found", flagFtnpID)
}
