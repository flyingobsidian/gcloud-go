package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee deployments (#1378) ---

func (s *apigeeCmds) initApigeeDeployments() {
	all := []*cobra.Command{s.apigeeDeployListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeeDeployOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagApigeeDeployFormat, "format", "", "Output format")
	}

	s.apigeeDeploymentsCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeDeploymentsCmd)
}

func (s *apigeeCmds) runApigeeDeployList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeeDeployOrganization)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Deployments.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing deployments: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1Deployment
	all = append(all, resp.Deployments...)
	return emitFormattedTo(s.inv.stdout, all, s.flagApigeeDeployFormat)
}
