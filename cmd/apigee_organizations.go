package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee organizations (#1382) ---

func (s *apigeeCmds) initApigeeOrganizations() {
	all := []*cobra.Command{s.apigeeOrgDescribeCmd, s.apigeeOrgListCmd, s.apigeeOrgProvisionCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeeOrgFormat, "format", "", "Output format")
	}
	s.apigeeOrgProvisionCmd.Flags().StringVar(&s.flagApigeeOrgConfigFile, "config-file", "", "Path to YAML/JSON provision request (required)")
	_ = s.apigeeOrgProvisionCmd.MarkFlagRequired("config-file")

	s.apigeeOrganizationsCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeOrganizationsCmd)
}

func (s *apigeeCmds) runApigeeOrgDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeOrgName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing organization: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeOrgFormat)
}

func (s *apigeeCmds) runApigeeOrgList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Apigee(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.List("organizations").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing organizations: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1OrganizationProjectMapping
	all = append(all, resp.Organizations...)
	return emitFormattedTo(s.inv.stdout, all, s.flagApigeeOrgFormat)
}

func (s *apigeeCmds) runApigeeOrgProvision(cmd *cobra.Command, args []string) error {
	project := args[0]
	if !strings.HasPrefix(project, "projects/") {
		project = "projects/" + project
	}
	body := &apigee.GoogleCloudApigeeV1ProvisionOrganizationRequest{}
	if err := loadYAMLOrJSONInto(s.flagApigeeOrgConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Apigee(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.ProvisionOrganization(project, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("provisioning organization: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagApigeeOrgFormat)
}
