package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap oauth-brands (#1065) ---

func (s *iapCmds) initIapOauthBrands() {
	all := []*cobra.Command{s.iapBrandsCreateCmd, s.iapBrandsDescribeCmd, s.iapBrandsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagIapBrandsFormat, "format", "", "Output format")
	}
	s.iapBrandsCreateCmd.Flags().StringVar(&s.flagIapBrandsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Brand body (required)")
	_ = s.iapBrandsCreateCmd.MarkFlagRequired("config-file")
	s.iapBrandsListCmd.Flags().Int64Var(&s.flagIapBrandsPageSize, "page-size", 0, "Maximum results per page")

	s.iapBrandsCmd.AddCommand(all...)
	s.iapCmd.AddCommand(s.iapBrandsCmd)
}

func (s *iapCmds) iapBrandsProjectParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s", project), nil
}

func (s *iapCmds) iapBrandsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.iapBrandsProjectParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/brands/%s", parent, id), nil
}

func (s *iapCmds) runIapBrandsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.iapBrandsProjectParent()
	if err != nil {
		return err
	}
	body := &iap.Brand{}
	if err := loadYAMLOrJSONInto(s.flagIapBrandsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating IAP OAuth brand: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created IAP OAuth brand [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagIapBrandsFormat)
}

func (s *iapCmds) runIapBrandsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.iapBrandsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing IAP OAuth brand: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIapBrandsFormat)
}

func (s *iapCmds) runIapBrandsList(cmd *cobra.Command, args []string) error {
	parent, err := s.iapBrandsProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	// The Brands.List endpoint returns the complete collection at once.
	resp, err := svc.Projects.Brands.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing IAP OAuth brands: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp.Brands, s.flagIapBrandsFormat)
}
