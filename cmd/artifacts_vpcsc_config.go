package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts vpcsc-config (#1085) ---

var artifactsVpcscConfigCmd = &cobra.Command{
	Use:   "vpcsc-config",
	Short: "Manage the VPC Service Controls config for Artifact Registry",
}

var (
	flagArtVpcLocation string
	flagArtVpcFormat   string
)

var artifactsVpcscConfigDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe the VPC-SC config for a location",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsVpcscConfigDescribe,
}

var artifactsVpcscConfigAllowCmd = &cobra.Command{
	Use:   "allow",
	Short: "Allow Artifact Registry to serve requests inside a VPC-SC perimeter",
	Args:  cobra.NoArgs,
	RunE:  func(cmd *cobra.Command, args []string) error { return artSetVpcscPolicy("ALLOW") },
}

var artifactsVpcscConfigDenyCmd = &cobra.Command{
	Use:   "deny",
	Short: "Deny Artifact Registry from serving requests inside a VPC-SC perimeter",
	Args:  cobra.NoArgs,
	RunE:  func(cmd *cobra.Command, args []string) error { return artSetVpcscPolicy("DENY") },
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsVpcscConfigDescribeCmd, artifactsVpcscConfigAllowCmd, artifactsVpcscConfigDenyCmd,
	} {
		c.Flags().StringVar(&flagArtVpcLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&flagArtVpcFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	artifactsVpcscConfigCmd.AddCommand(
		artifactsVpcscConfigDescribeCmd, artifactsVpcscConfigAllowCmd, artifactsVpcscConfigDenyCmd,
	)
	artifactsCmd.AddCommand(artifactsVpcscConfigCmd)
}

func artVpcscName(project string) string {
	return artLocationParent(project, flagArtVpcLocation) + "/vpcscConfig"
}

func runArtifactsVpcscConfigDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.GetVpcscConfig(artVpcscName(project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing vpcscConfig: %w", err)
	}
	return emitFormatted(cfg, flagArtVpcFormat)
}

func artSetVpcscPolicy(policy string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.VPCSCConfig{VpcscPolicy: policy}
	out, err := svc.Projects.Locations.UpdateVpcscConfig(artVpcscName(project), body).
		UpdateMask("vpcscPolicy").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating vpcscConfig: %w", err)
	}
	return emitFormatted(out, flagArtVpcFormat)
}
