package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts vpcsc-config (#1085) ---

func (s *artifactsCmds) initArtifactsVpcscConfig() {
	for _, c := range []*cobra.Command{
		s.artifactsVpcscConfigDescribeCmd, s.artifactsVpcscConfigAllowCmd, s.artifactsVpcscConfigDenyCmd,
	} {
		c.Flags().StringVar(&s.flagArtVpcLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagArtVpcFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	s.artifactsVpcscConfigCmd.AddCommand(
		s.artifactsVpcscConfigDescribeCmd, s.artifactsVpcscConfigAllowCmd, s.artifactsVpcscConfigDenyCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsVpcscConfigCmd)
}

func (s *artifactsCmds) artVpcscName(project string) string {
	return artLocationParent(project, s.flagArtVpcLocation) + "/vpcscConfig"
}

func (s *artifactsCmds) runArtifactsVpcscConfigDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.GetVpcscConfig(s.artVpcscName(project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing vpcscConfig: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, cfg, s.flagArtVpcFormat)
}

func (s *artifactsCmds) artSetVpcscPolicy(policy string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.VPCSCConfig{VpcscPolicy: policy}
	out, err := svc.Projects.Locations.UpdateVpcscConfig(s.artVpcscName(project), body).
		UpdateMask("vpcscPolicy").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating vpcscConfig: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtVpcFormat)
}
