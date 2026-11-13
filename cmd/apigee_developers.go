package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee developers (#1379) ---

func (s *apigeeCmds) initApigeeDevelopers() {
	all := []*cobra.Command{
		s.apigeeDevCreateCmd, s.apigeeDevDeleteCmd, s.apigeeDevDescribeCmd,
		s.apigeeDevListCmd, s.apigeeDevUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeeDevOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagApigeeDevFormat, "format", "", "Output format")
	}
	s.apigeeDevCreateCmd.Flags().StringVar(&s.flagApigeeDevConfigFile, "config-file", "", "Path to YAML/JSON developer definition (required)")
	_ = s.apigeeDevCreateCmd.MarkFlagRequired("config-file")
	s.apigeeDevUpdateCmd.Flags().StringVar(&s.flagApigeeDevConfigFile, "config-file", "", "Path to YAML/JSON developer definition (required)")
	_ = s.apigeeDevUpdateCmd.MarkFlagRequired("config-file")

	s.apigeeDevelopersCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeDevelopersCmd)
}

func (s *apigeeCmds) apigeeDevName(id string) (string, error) {
	return apigeeResource(s.flagApigeeDevOrganization, "developers", id)
}

func (s *apigeeCmds) runApigeeDevCreate(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeeDevOrganization)
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1Developer{}
	if err := loadYAMLOrJSONInto(s.flagApigeeDevConfigFile, body); err != nil {
		return err
	}
	if body.Email == "" {
		body.Email = args[0]
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Developers.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating developer: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeDevFormat)
}

func (s *apigeeCmds) runApigeeDevDelete(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeDevName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Developers.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting developer: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted developer [%s].\n", args[0])
	return nil
}

func (s *apigeeCmds) runApigeeDevDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeDevName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Developers.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing developer: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeDevFormat)
}

func (s *apigeeCmds) runApigeeDevList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeeDevOrganization)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Developers.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing developers: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1Developer
	all = append(all, resp.Developer...)
	return emitFormattedTo(s.inv.stdout, all, s.flagApigeeDevFormat)
}

func (s *apigeeCmds) runApigeeDevUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeDevName(args[0])
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1Developer{}
	if err := loadYAMLOrJSONInto(s.flagApigeeDevConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Developers.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating developer: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeDevFormat)
}
