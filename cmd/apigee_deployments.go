package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee deployments (#1378) ---

var apigeeDeploymentsCmd = &cobra.Command{Use: "deployments", Short: "Manage Apigee deployments"}

var (
	flagApigeeDeployOrganization string
	flagApigeeDeployFormat       string
)

var (
	apigeeDeployListCmd = &cobra.Command{
		Use: "list", Short: "List deployments in an organization",
		Args: cobra.NoArgs, RunE: runApigeeDeployList,
	}
)

func init() {
	all := []*cobra.Command{apigeeDeployListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeDeployOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeeDeployFormat, "format", "", "Output format")
	}

	apigeeDeploymentsCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeDeploymentsCmd)
}

func runApigeeDeployList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeeDeployOrganization)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Deployments.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing deployments: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1Deployment
	all = append(all, resp.Deployments...)
	return emitFormatted(all, flagApigeeDeployFormat)
}
