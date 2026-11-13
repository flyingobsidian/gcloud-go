package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee products (#1383) ---

func (s *apigeeCmds) initApigeeProducts() {
	all := []*cobra.Command{
		s.apigeePrdCreateCmd, s.apigeePrdDeleteCmd, s.apigeePrdDescribeCmd,
		s.apigeePrdListCmd, s.apigeePrdUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeePrdOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagApigeePrdFormat, "format", "", "Output format")
	}
	s.apigeePrdCreateCmd.Flags().StringVar(&s.flagApigeePrdConfigFile, "config-file", "", "Path to YAML/JSON API product definition (required)")
	_ = s.apigeePrdCreateCmd.MarkFlagRequired("config-file")
	s.apigeePrdUpdateCmd.Flags().StringVar(&s.flagApigeePrdConfigFile, "config-file", "", "Path to YAML/JSON API product definition (required)")
	_ = s.apigeePrdUpdateCmd.MarkFlagRequired("config-file")

	s.apigeeProductsCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeProductsCmd)
}

func (s *apigeeCmds) apigeePrdName(id string) (string, error) {
	return apigeeResource(s.flagApigeePrdOrganization, "apiproducts", id)
}

func (s *apigeeCmds) runApigeePrdCreate(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeePrdOrganization)
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1ApiProduct{}
	if err := loadYAMLOrJSONInto(s.flagApigeePrdConfigFile, body); err != nil {
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
	got, err := svc.Organizations.Apiproducts.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating api product: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeePrdFormat)
}

func (s *apigeeCmds) runApigeePrdDelete(cmd *cobra.Command, args []string) error {
	name, err := s.apigeePrdName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Apiproducts.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting api product: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted api product [%s].\n", args[0])
	return nil
}

func (s *apigeeCmds) runApigeePrdDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.apigeePrdName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apiproducts.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api product: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeePrdFormat)
}

func (s *apigeeCmds) runApigeePrdList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeePrdOrganization)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Apiproducts.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing api products: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1ApiProduct
	all = append(all, resp.ApiProduct...)
	return emitFormattedTo(s.inv.stdout, all, s.flagApigeePrdFormat)
}

func (s *apigeeCmds) runApigeePrdUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.apigeePrdName(args[0])
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1ApiProduct{}
	if err := loadYAMLOrJSONInto(s.flagApigeePrdConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apiproducts.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating api product: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeePrdFormat)
}
