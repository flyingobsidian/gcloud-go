package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap oauth-clients (#1066) ---

func (s *iapCmds) initIapOauthClients() {
	all := []*cobra.Command{
		s.iapClientsCreateCmd, s.iapClientsDeleteCmd, s.iapClientsDescribeCmd,
		s.iapClientsListCmd, s.iapClientsResetSecretCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagIapClientsBrand, "brand", "",
			"Brand id (or the full projects/PROJECT/brands/BRAND path) that owns the client (required)")
		_ = c.MarkFlagRequired("brand")
		c.Flags().StringVar(&s.flagIapClientsFormat, "format", "", "Output format")
	}
	s.iapClientsCreateCmd.Flags().StringVar(&s.flagIapClientsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the IdentityAwareProxyClient body")
	s.iapClientsCreateCmd.Flags().StringVar(&s.flagIapClientsDisplayName, "display-name", "",
		"Display name for the new client (used when --config-file is not set)")
	s.iapClientsListCmd.Flags().Int64Var(&s.flagIapClientsPageSize, "page-size", 0, "Maximum results per page")

	s.iapClientsCmd.AddCommand(all...)
	s.iapCmd.AddCommand(s.iapClientsCmd)
}

func (s *iapCmds) iapClientsBrandParent() (string, error) {
	if strings.HasPrefix(s.flagIapClientsBrand, "projects/") {
		return s.flagIapClientsBrand, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/brands/%s", project, s.flagIapClientsBrand), nil
}

func (s *iapCmds) iapClientsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.iapClientsBrandParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/identityAwareProxyClients/%s", parent, id), nil
}

func (s *iapCmds) runIapClientsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.iapClientsBrandParent()
	if err != nil {
		return err
	}
	body := &iap.IdentityAwareProxyClient{}
	if s.flagIapClientsConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagIapClientsConfigFile, body); err != nil {
			return err
		}
	}
	if s.flagIapClientsDisplayName != "" {
		body.DisplayName = s.flagIapClientsDisplayName
	}
	if body.DisplayName == "" {
		return fmt.Errorf("either --config-file or --display-name is required")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.IdentityAwareProxyClients.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating IAP OAuth client: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created IAP OAuth client [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagIapClientsFormat)
}

func (s *iapCmds) runIapClientsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.iapClientsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Brands.IdentityAwareProxyClients.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting IAP OAuth client: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted IAP OAuth client [%s].\n", args[0])
	return nil
}

func (s *iapCmds) runIapClientsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.iapClientsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.IdentityAwareProxyClients.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing IAP OAuth client: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIapClientsFormat)
}

func (s *iapCmds) runIapClientsList(cmd *cobra.Command, args []string) error {
	parent, err := s.iapClientsBrandParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	var all []*iap.IdentityAwareProxyClient
	pageToken := ""
	for {
		call := svc.Projects.Brands.IdentityAwareProxyClients.List(parent).Context(ctx)
		if s.flagIapClientsPageSize > 0 {
			call = call.PageSize(s.flagIapClientsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing IAP OAuth clients: %w", err)
		}
		all = append(all, resp.IdentityAwareProxyClients...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagIapClientsFormat)
}

func (s *iapCmds) runIapClientsResetSecret(cmd *cobra.Command, args []string) error {
	name, err := s.iapClientsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.IdentityAwareProxyClients.ResetSecret(name,
		&iap.ResetIdentityAwareProxyClientSecretRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting IAP OAuth client secret: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Reset secret for IAP OAuth client [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagIapClientsFormat)
}
