package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee organizations (#1382) ---

var apigeeOrganizationsCmd = &cobra.Command{Use: "organizations", Short: "Manage Apigee organizations"}

var (
	flagApigeeOrgFormat     string
	flagApigeeOrgConfigFile string
)

var (
	apigeeOrgDescribeCmd = &cobra.Command{
		Use: "describe ORGANIZATION", Short: "Describe an Apigee organization",
		Args: cobra.ExactArgs(1), RunE: runApigeeOrgDescribe,
	}
	apigeeOrgListCmd = &cobra.Command{
		Use: "list", Short: "List Apigee organizations you have access to",
		Args: cobra.NoArgs, RunE: runApigeeOrgList,
	}
	apigeeOrgProvisionCmd = &cobra.Command{
		Use: "provision PROJECT", Short: "Provision an Apigee organization for a GCP project",
		Args: cobra.ExactArgs(1), RunE: runApigeeOrgProvision,
	}
)

func init() {
	all := []*cobra.Command{apigeeOrgDescribeCmd, apigeeOrgListCmd, apigeeOrgProvisionCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeOrgFormat, "format", "", "Output format")
	}
	apigeeOrgProvisionCmd.Flags().StringVar(&flagApigeeOrgConfigFile, "config-file", "", "Path to YAML/JSON provision request (required)")
	_ = apigeeOrgProvisionCmd.MarkFlagRequired("config-file")

	apigeeOrganizationsCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeOrganizationsCmd)
}

func runApigeeOrgDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeOrgName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing organization: %w", err)
	}
	return emitFormatted(got, flagApigeeOrgFormat)
}

func runApigeeOrgList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.List("organizations").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing organizations: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1OrganizationProjectMapping
	all = append(all, resp.Organizations...)
	return emitFormatted(all, flagApigeeOrgFormat)
}

func runApigeeOrgProvision(cmd *cobra.Command, args []string) error {
	project := args[0]
	if !strings.HasPrefix(project, "projects/") {
		project = "projects/" + project
	}
	body := &apigee.GoogleCloudApigeeV1ProvisionOrganizationRequest{}
	if err := loadYAMLOrJSONInto(flagApigeeOrgConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.ProvisionOrganization(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("provisioning organization: %w", err)
	}
	return emitFormatted(op, flagApigeeOrgFormat)
}
