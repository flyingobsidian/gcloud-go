package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee environments (#1380) ---

func (s *apigeeCmds) initApigeeEnvironments() {
	all := []*cobra.Command{
		s.apigeeEnvCreateCmd, s.apigeeEnvDeleteCmd, s.apigeeEnvDescribeCmd, s.apigeeEnvUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeeEnvOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagApigeeEnvFormat, "format", "", "Output format")
	}
	s.apigeeEnvCreateCmd.Flags().StringVar(&s.flagApigeeEnvConfigFile, "config-file", "", "Path to YAML/JSON environment definition (required)")
	_ = s.apigeeEnvCreateCmd.MarkFlagRequired("config-file")
	s.apigeeEnvUpdateCmd.Flags().StringVar(&s.flagApigeeEnvConfigFile, "config-file", "", "Path to YAML/JSON environment definition (required)")
	_ = s.apigeeEnvUpdateCmd.MarkFlagRequired("config-file")

	s.apigeeEnvironmentsCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeEnvironmentsCmd)
}

func (s *apigeeCmds) apigeeEnvName(id string) (string, error) {
	return apigeeResource(s.flagApigeeEnvOrganization, "environments", id)
}

func (s *apigeeCmds) runApigeeEnvCreate(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeeEnvOrganization)
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1Environment{}
	if err := loadYAMLOrJSONInto(s.flagApigeeEnvConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = args[0]
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Environments.Create(parent, body).Name(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating environment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagApigeeEnvFormat)
}

func (s *apigeeCmds) runApigeeEnvDelete(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeEnvName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Environments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}
	if op != nil && op.Name != "" {
		fmt.Fprintf(s.inv.stdout, "Delete environment [%s] initiated (operation: %s).\n", args[0], op.Name)
	} else {
		fmt.Fprintf(s.inv.stdout, "Deleted environment [%s].\n", args[0])
	}
	return nil
}

func (s *apigeeCmds) runApigeeEnvDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeEnvName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing environment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeEnvFormat)
}

func (s *apigeeCmds) runApigeeEnvUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeEnvName(args[0])
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1Environment{}
	if err := loadYAMLOrJSONInto(s.flagApigeeEnvConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Environments.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeEnvFormat)
}
