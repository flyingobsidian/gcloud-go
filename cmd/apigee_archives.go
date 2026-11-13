package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee archives (#1377) ---

func (s *apigeeCmds) initApigeeArchives() {
	all := []*cobra.Command{
		s.apigeeArcCreateCmd, s.apigeeArcDeleteCmd, s.apigeeArcDescribeCmd,
		s.apigeeArcListCmd, s.apigeeArcUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeeArcOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagApigeeArcEnvironment, "environment", "", "Apigee environment (required)")
		_ = c.MarkFlagRequired("environment")
		c.Flags().StringVar(&s.flagApigeeArcFormat, "format", "", "Output format")
	}
	s.apigeeArcCreateCmd.Flags().StringVar(&s.flagApigeeArcConfigFile, "config-file", "", "Path to YAML/JSON archive deployment definition (required)")
	_ = s.apigeeArcCreateCmd.MarkFlagRequired("config-file")
	s.apigeeArcUpdateCmd.Flags().StringVar(&s.flagApigeeArcConfigFile, "config-file", "", "Path to YAML/JSON archive deployment definition (required)")
	_ = s.apigeeArcUpdateCmd.MarkFlagRequired("config-file")
	s.apigeeArcListCmd.Flags().Int64Var(&s.flagApigeeArcPageSize, "page-size", 0, "Maximum results per page")

	s.apigeeArchivesCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeArchivesCmd)
}

func (s *apigeeCmds) apigeeArcParent() (string, error) {
	if s.flagApigeeArcEnvironment == "" {
		return "", fmt.Errorf("--environment is required")
	}
	org, err := apigeeOrgName(s.flagApigeeArcOrganization)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/environments/%s", org, s.flagApigeeArcEnvironment), nil
}

func (s *apigeeCmds) apigeeArcName(id string) (string, error) {
	parent, err := s.apigeeArcParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/archiveDeployments/%s", parent, id), nil
}

func (s *apigeeCmds) runApigeeArcCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.apigeeArcParent()
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1ArchiveDeployment{}
	if err := loadYAMLOrJSONInto(s.flagApigeeArcConfigFile, body); err != nil {
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
	op, err := svc.Organizations.Environments.ArchiveDeployments.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating archive deployment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagApigeeArcFormat)
}

func (s *apigeeCmds) runApigeeArcDelete(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeArcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Environments.ArchiveDeployments.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting archive deployment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted archive deployment [%s].\n", args[0])
	return nil
}

func (s *apigeeCmds) runApigeeArcDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeArcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Environments.ArchiveDeployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing archive deployment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeArcFormat)
}

func (s *apigeeCmds) runApigeeArcList(cmd *cobra.Command, args []string) error {
	parent, err := s.apigeeArcParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	var all []*apigee.GoogleCloudApigeeV1ArchiveDeployment
	pageToken := ""
	for {
		call := svc.Organizations.Environments.ArchiveDeployments.List(parent).Context(ctx)
		if s.flagApigeeArcPageSize > 0 {
			call = call.PageSize(s.flagApigeeArcPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing archive deployments: %w", err)
		}
		all = append(all, resp.ArchiveDeployments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagApigeeArcFormat)
}

func (s *apigeeCmds) runApigeeArcUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeArcName(args[0])
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1ArchiveDeployment{}
	if err := loadYAMLOrJSONInto(s.flagApigeeArcConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Environments.ArchiveDeployments.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating archive deployment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeArcFormat)
}
