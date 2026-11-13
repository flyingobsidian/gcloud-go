package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	domainsapi "google.golang.org/api/domains/v1"
)

// --- gcloud domains (#943) ---

// The domains v1 registrations surface always uses the "global" location. We
// keep --location as a flag for forward-compatibility but default to global.
func (s *domainsCmds) domLocation() string {
	if s.flagDomLocation != "" {
		return s.flagDomLocation
	}
	return "global"
}

func (s *domainsCmds) domParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, s.domLocation())
}

func (s *domainsCmds) domRegName(id, project string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/registrations/%s", s.domParent(project), id)
}

// --- Registrations subgroup ---

func (s *domainsCmds) initDomains() {
	// Common location flag on every command that hits a project/location endpoint.
	all := []*cobra.Command{
		s.domainsRegisterCmd, s.domainsRegDescribeCmd, s.domainsRegListCmd,
		s.domainsRegDeleteCmd, s.domainsRegUpdateCmd, s.domainsRegExportCmd,
		s.domainsRegImportCmd, s.domainsRegInitPushCmd, s.domainsRegRenewCmd,
		s.domainsRegRetrieveAuthCodeCmd, s.domainsRegResetAuthCodeCmd,
		s.domainsRegRetrieveGDDnsCmd, s.domainsRegRetrieveGDForwardingCfgCmd,
		s.domainsRegRetrieveImportTransferCmd, s.domainsRegRetrieveRegisterParamsCmd,
		s.domainsRegRetrieveTransferParamsCmd, s.domainsRegSearchDomainsCmd,
		s.domainsRegTransferCmd, s.domainsRegConfigureContactsCmd,
		s.domainsRegConfigureDNSCmd, s.domainsRegConfigureMgmtCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDomLocation, "location", "", "Location of the registration (defaults to 'global')")
	}

	// Register / Transfer / Import / Configure take a JSON/YAML input file
	// describing the Registration (or DNS/contact settings) because the request
	// bodies are large and nested.
	s.domainsRegisterCmd.Flags().StringVar(&s.flagDomFile, "from-file", "", "JSON/YAML file with the RegisterDomainRequest body")
	s.domainsRegisterCmd.Flags().StringVar(&s.flagDomYearlyPrice, "yearly-price", "", "Yearly price acknowledgement in units.currency form (e.g. 12.00.USD)")
	s.domainsRegisterCmd.Flags().BoolVar(&s.flagDomValidateOnly, "validate-only", false, "Validate only")

	s.domainsRegTransferCmd.Flags().StringVar(&s.flagDomFile, "from-file", "", "JSON/YAML file with the TransferDomainRequest body")
	s.domainsRegTransferCmd.Flags().StringVar(&s.flagDomAuthCode, "authorization-code", "", "The transfer authorization code from the current registrar")
	s.domainsRegTransferCmd.Flags().StringVar(&s.flagDomYearlyPrice, "yearly-price", "", "Yearly price acknowledgement in units.currency form")
	s.domainsRegTransferCmd.Flags().BoolVar(&s.flagDomValidateOnly, "validate-only", false, "Validate only")

	s.domainsRegImportCmd.Flags().StringVar(&s.flagDomFile, "from-file", "", "JSON/YAML file with the ImportDomainRequest body (optional)")

	for _, c := range []*cobra.Command{s.domainsRegConfigureContactsCmd, s.domainsRegConfigureDNSCmd, s.domainsRegConfigureMgmtCmd} {
		c.Flags().StringVar(&s.flagDomFile, "from-file", "", "JSON/YAML file with the configure request body")
		c.Flags().StringVar(&s.flagDomUpdateMask, "update-mask", "", "Comma-separated field mask (required)")
		c.MarkFlagRequired("from-file")
	}

	s.domainsRegListCmd.Flags().StringVar(&s.flagDomFilter, "filter", "", "Server-side filter expression")
	s.domainsRegListCmd.Flags().Int64Var(&s.flagDomPageSize, "page-size", 0, "Number of results per page")
	s.domainsRegListCmd.Flags().Int64Var(&s.flagDomLimit, "limit", 0, "Maximum number of results to return")

	s.domainsRegRenewCmd.Flags().StringVar(&s.flagDomYearlyPrice, "yearly-price", "", "Yearly renewal price acknowledgement in units.currency form")
	s.domainsRegRenewCmd.Flags().BoolVar(&s.flagDomValidateOnly, "validate-only", false, "Validate only")

	s.domainsRegInitPushCmd.Flags().StringVar(&s.flagDomTag, "tag", "", "The Tag of the new registrar (required)")
	s.domainsRegInitPushCmd.MarkFlagRequired("tag")

	s.domainsRegUpdateCmd.Flags().StringVar(&s.flagDomFile, "from-file", "", "JSON/YAML file with the updated Registration body")
	s.domainsRegUpdateCmd.Flags().StringVar(&s.flagDomUpdateMask, "update-mask", "", "Comma-separated field mask (required)")
	s.domainsRegUpdateCmd.MarkFlagRequired("from-file")
	s.domainsRegUpdateCmd.MarkFlagRequired("update-mask")

	for _, c := range []*cobra.Command{s.domainsRegRetrieveRegisterParamsCmd, s.domainsRegRetrieveTransferParamsCmd} {
		c.Flags().StringVar(&s.flagDomDomainName, "domain-name", "", "The domain name to look up (required)")
		c.MarkFlagRequired("domain-name")
	}

	s.domainsRegistrationsCmd.AddCommand(
		s.domainsRegisterCmd, s.domainsRegDescribeCmd, s.domainsRegListCmd,
		s.domainsRegDeleteCmd, s.domainsRegUpdateCmd, s.domainsRegExportCmd,
		s.domainsRegImportCmd, s.domainsRegInitPushCmd, s.domainsRegRenewCmd,
		s.domainsRegRetrieveAuthCodeCmd, s.domainsRegResetAuthCodeCmd,
		s.domainsRegRetrieveGDDnsCmd, s.domainsRegRetrieveGDForwardingCfgCmd,
		s.domainsRegRetrieveImportTransferCmd, s.domainsRegRetrieveRegisterParamsCmd,
		s.domainsRegRetrieveTransferParamsCmd, s.domainsRegSearchDomainsCmd,
		s.domainsRegTransferCmd,
	)
	s.domainsRegConfigureCmd.AddCommand(
		s.domainsRegConfigureContactsCmd, s.domainsRegConfigureDNSCmd, s.domainsRegConfigureMgmtCmd,
	)
	s.domainsRegistrationsCmd.AddCommand(s.domainsRegConfigureCmd)
	s.domainsCmd.AddCommand(s.domainsRegistrationsCmd)

}

// --- helpers ---

func (s *domainsCmds) domSvc(ctx context.Context) (*domainsapi.Service, error) {
	return s.inv.clients().Domains(ctx)
}

// --- Register ---

func (s *domainsCmds) runDomRegister(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &domainsapi.RegisterDomainRequest{}
	if s.flagDomFile != "" {
		if err := loadYAMLOrJSONInto(s.flagDomFile, req); err != nil {
			return err
		}
	}
	if req.Registration == nil {
		req.Registration = &domainsapi.Registration{}
	}
	if req.Registration.DomainName == "" {
		req.Registration.DomainName = args[0]
	}
	if req.YearlyPrice == nil && s.flagDomYearlyPrice != "" {
		money, err := parseDomainsMoney(s.flagDomYearlyPrice)
		if err != nil {
			return err
		}
		req.YearlyPrice = money
	}
	if s.flagDomValidateOnly {
		req.ValidateOnly = true
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Register(s.domParent(project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("registering domain: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Register request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Describe / List / Delete / Update / Export ---

func (s *domainsCmds) runDomDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	reg, err := svc.Projects.Locations.Registrations.Get(s.domRegName(args[0], project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing registration: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, reg, "")
}

func (s *domainsCmds) runDomList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	var all []*domainsapi.Registration
	pageToken := ""
	for {
		call := svc.Projects.Locations.Registrations.List(s.domParent(project)).Context(ctx)
		if s.flagDomFilter != "" {
			call = call.Filter(s.flagDomFilter)
		}
		if s.flagDomPageSize > 0 {
			call = call.PageSize(s.flagDomPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing registrations: %w", err)
		}
		all = append(all, resp.Registrations...)
		if s.flagDomLimit > 0 && int64(len(all)) >= s.flagDomLimit {
			all = all[:s.flagDomLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *domainsCmds) runDomDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Delete(s.domRegName(args[0], project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting registration: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *domainsCmds) runDomUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	reg := &domainsapi.Registration{}
	if err := loadYAMLOrJSONInto(s.flagDomFile, reg); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Patch(s.domRegName(args[0], project), reg).UpdateMask(s.flagDomUpdateMask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating registration: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *domainsCmds) runDomExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Export(s.domRegName(args[0], project), &domainsapi.ExportRegistrationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting registration: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Export request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Import / Initiate push / Renew ---

func (s *domainsCmds) runDomImport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &domainsapi.ImportDomainRequest{DomainName: args[0]}
	if s.flagDomFile != "" {
		if err := loadYAMLOrJSONInto(s.flagDomFile, req); err != nil {
			return err
		}
		if req.DomainName == "" {
			req.DomainName = args[0]
		}
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Import(s.domParent(project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing domain: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Import request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *domainsCmds) runDomInitiatePushTransfer(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.InitiatePushTransfer(
		s.domRegName(args[0], project),
		&domainsapi.InitiatePushTransferRequest{Tag: s.flagDomTag},
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("initiating push transfer: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Push transfer initiated for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *domainsCmds) runDomRenew(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &domainsapi.RenewDomainRequest{ValidateOnly: s.flagDomValidateOnly}
	if s.flagDomYearlyPrice != "" {
		money, err := parseDomainsMoney(s.flagDomYearlyPrice)
		if err != nil {
			return err
		}
		req.YearlyPrice = money
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.RenewDomain(s.domRegName(args[0], project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renewing domain: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Renew request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Auth code / DNS / forwarding lookups ---

func (s *domainsCmds) runDomRetrieveAuth(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	code, err := svc.Projects.Locations.Registrations.RetrieveAuthorizationCode(s.domRegName(args[0], project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving authorization code: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, code, "")
}

func (s *domainsCmds) runDomResetAuth(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	code, err := svc.Projects.Locations.Registrations.ResetAuthorizationCode(s.domRegName(args[0], project), &domainsapi.ResetAuthorizationCodeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting authorization code: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, code, "")
}

func (s *domainsCmds) runDomRetrieveGDDns(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	var all []*domainsapi.ResourceRecordSet
	pageToken := ""
	for {
		call := svc.Projects.Locations.Registrations.RetrieveGoogleDomainsDnsRecords(s.domRegName(args[0], project)).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("retrieving Google Domains DNS records: %w", err)
		}
		if resp == nil {
			break
		}
		if resp.Rrset != nil {
			all = append(all, resp.Rrset...)
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *domainsCmds) runDomRetrieveGDForwardingConfig(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.Registrations.RetrieveGoogleDomainsForwardingConfig(s.domRegName(args[0], project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving Google Domains forwarding config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, cfg, "")
}

// --- Retrieve parameter lookups ---

func (s *domainsCmds) runDomRetrieveImportable(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	var all []*domainsapi.Domain
	pageToken := ""
	for {
		call := svc.Projects.Locations.Registrations.RetrieveImportableDomains(s.domParent(project)).Context(ctx)
		if s.flagDomPageSize > 0 {
			call = call.PageSize(s.flagDomPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing importable domains: %w", err)
		}
		all = append(all, resp.Domains...)
		if s.flagDomLimit > 0 && int64(len(all)) >= s.flagDomLimit {
			all = all[:s.flagDomLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *domainsCmds) runDomRetrieveRegisterParams(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Registrations.RetrieveRegisterParameters(s.domParent(project)).DomainName(s.flagDomDomainName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving register parameters: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, "")
}

func (s *domainsCmds) runDomRetrieveTransferParams(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Registrations.RetrieveTransferParameters(s.domParent(project)).DomainName(s.flagDomDomainName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving transfer parameters: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, "")
}

func (s *domainsCmds) runDomSearch(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Registrations.SearchDomains(s.domParent(project)).Query(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("searching domains: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, "")
}

// --- Transfer ---

func (s *domainsCmds) runDomTransfer(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &domainsapi.TransferDomainRequest{}
	if s.flagDomFile != "" {
		if err := loadYAMLOrJSONInto(s.flagDomFile, req); err != nil {
			return err
		}
	}
	if req.Registration == nil {
		req.Registration = &domainsapi.Registration{}
	}
	if req.Registration.DomainName == "" {
		req.Registration.DomainName = args[0]
	}
	if s.flagDomAuthCode != "" {
		req.AuthorizationCode = &domainsapi.AuthorizationCode{Code: s.flagDomAuthCode}
	}
	if req.YearlyPrice == nil && s.flagDomYearlyPrice != "" {
		money, err := parseDomainsMoney(s.flagDomYearlyPrice)
		if err != nil {
			return err
		}
		req.YearlyPrice = money
	}
	if s.flagDomValidateOnly {
		req.ValidateOnly = true
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Transfer(s.domParent(project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("transferring domain: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Transfer request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Configure sub-subcommands ---

func (s *domainsCmds) runDomConfigureContacts(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &domainsapi.ConfigureContactSettingsRequest{}
	if err := loadYAMLOrJSONInto(s.flagDomFile, req); err != nil {
		return err
	}
	if s.flagDomUpdateMask != "" {
		req.UpdateMask = s.flagDomUpdateMask
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.ConfigureContactSettings(s.domRegName(args[0], project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("configuring contacts: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Configure request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *domainsCmds) runDomConfigureDNS(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &domainsapi.ConfigureDnsSettingsRequest{}
	if err := loadYAMLOrJSONInto(s.flagDomFile, req); err != nil {
		return err
	}
	if s.flagDomUpdateMask != "" {
		req.UpdateMask = s.flagDomUpdateMask
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.ConfigureDnsSettings(s.domRegName(args[0], project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("configuring DNS: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Configure request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *domainsCmds) runDomConfigureManagement(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &domainsapi.ConfigureManagementSettingsRequest{}
	if err := loadYAMLOrJSONInto(s.flagDomFile, req); err != nil {
		return err
	}
	if s.flagDomUpdateMask != "" {
		req.UpdateMask = s.flagDomUpdateMask
	}
	ctx := cmd.Context()
	svc, err := s.domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.ConfigureManagementSettings(s.domRegName(args[0], project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("configuring management: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Configure request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// parseDomainsMoney accepts "AMOUNT.CURRENCY" e.g. "12.00.USD" and returns a Money.
func parseDomainsMoney(s string) (*domainsapi.Money, error) {
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid --yearly-price %q, expected AMOUNT.CURRENCY (e.g. 12.00.USD)", s)
	}
	currency := parts[len(parts)-1]
	amount := strings.Join(parts[:len(parts)-1], ".")
	dotIdx := strings.Index(amount, ".")
	var units, nanos int64
	if dotIdx < 0 {
		fmt.Sscanf(amount, "%d", &units)
	} else {
		fmt.Sscanf(amount[:dotIdx], "%d", &units)
		frac := amount[dotIdx+1:]
		for len(frac) < 9 {
			frac += "0"
		}
		frac = frac[:9]
		fmt.Sscanf(frac, "%d", &nanos)
	}
	return &domainsapi.Money{CurrencyCode: currency, Units: units, Nanos: nanos}, nil
}

// domainsCmds holds the commands and flag values of one invocation's "domains"
// command group, so invocations don't share state (#2070).
type domainsCmds struct {
	inv *invocation

	domainsCmd                           *cobra.Command
	flagDomLocation                      string
	flagDomFile                          string
	flagDomDomainName                    string
	flagDomValidateOnly                  bool
	flagDomFilter                        string
	flagDomPageSize                      int64
	flagDomLimit                         int64
	flagDomYearlyPrice                   string
	flagDomAuthCode                      string
	flagDomTag                           string
	flagDomUpdateMask                    string
	domainsRegistrationsCmd              *cobra.Command
	domainsRegisterCmd                   *cobra.Command
	domainsRegDescribeCmd                *cobra.Command
	domainsRegListCmd                    *cobra.Command
	domainsRegDeleteCmd                  *cobra.Command
	domainsRegUpdateCmd                  *cobra.Command
	domainsRegExportCmd                  *cobra.Command
	domainsRegImportCmd                  *cobra.Command
	domainsRegInitPushCmd                *cobra.Command
	domainsRegRenewCmd                   *cobra.Command
	domainsRegRetrieveAuthCodeCmd        *cobra.Command
	domainsRegResetAuthCodeCmd           *cobra.Command
	domainsRegRetrieveGDDnsCmd           *cobra.Command
	domainsRegRetrieveGDForwardingCfgCmd *cobra.Command
	domainsRegRetrieveImportTransferCmd  *cobra.Command
	domainsRegRetrieveRegisterParamsCmd  *cobra.Command
	domainsRegRetrieveTransferParamsCmd  *cobra.Command
	domainsRegSearchDomainsCmd           *cobra.Command
	domainsRegTransferCmd                *cobra.Command
	// configure has three sub-subcommands mirroring the Python surface.
	domainsRegConfigureCmd         *cobra.Command
	domainsRegConfigureContactsCmd *cobra.Command
	domainsRegConfigureDNSCmd      *cobra.Command
	domainsRegConfigureMgmtCmd     *cobra.Command
}

// newDomainsCmd returns the "domains" command group for inv.
func newDomainsCmd(inv *invocation) *cobra.Command {
	return newDomainsCmds(inv).domainsCmd
}

// newDomainsCmds builds the "domains" command group for inv and returns its state.
func newDomainsCmds(inv *invocation) *domainsCmds {
	s := &domainsCmds{inv: inv}
	s.domainsCmd = &cobra.Command{
		Use:   "domains",
		Short: "Manage Cloud Domains",
	}
	s.domainsRegistrationsCmd = &cobra.Command{Use: "registrations", Short: "Manage domain registrations"}
	s.domainsRegisterCmd = &cobra.Command{
		Use:   "register DOMAIN",
		Short: "Register a domain",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomRegister,
	}
	s.domainsRegDescribeCmd = &cobra.Command{
		Use:   "describe REGISTRATION",
		Short: "Describe a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomDescribe,
	}
	s.domainsRegListCmd = &cobra.Command{
		Use:   "list",
		Short: "List registrations in a location",
		Args:  cobra.NoArgs,
		RunE:  s.runDomList,
	}
	s.domainsRegDeleteCmd = &cobra.Command{
		Use:   "delete REGISTRATION",
		Short: "Delete a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomDelete,
	}
	s.domainsRegUpdateCmd = &cobra.Command{
		Use:   "update REGISTRATION",
		Short: "Update a registration (labels only via --from-file)",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomUpdate,
	}
	s.domainsRegExportCmd = &cobra.Command{
		Use:   "export REGISTRATION",
		Short: "Export a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomExport,
	}
	s.domainsRegImportCmd = &cobra.Command{
		Use:   "import DOMAIN",
		Short: "Import an existing domain into Cloud Domains",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomImport,
	}
	s.domainsRegInitPushCmd = &cobra.Command{
		Use:   "initiate-push-transfer REGISTRATION",
		Short: "Initiate a push transfer to another registrar",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomInitiatePushTransfer,
	}
	s.domainsRegRenewCmd = &cobra.Command{
		Use:   "renew-domain REGISTRATION",
		Short: "Renew a registered domain",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomRenew,
	}
	s.domainsRegRetrieveAuthCodeCmd = &cobra.Command{
		Use:   "retrieve-authorization-code REGISTRATION",
		Short: "Retrieve the authorization code for a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomRetrieveAuth,
	}
	s.domainsRegResetAuthCodeCmd = &cobra.Command{
		Use:   "reset-authorization-code REGISTRATION",
		Short: "Reset the authorization code for a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomResetAuth,
	}
	s.domainsRegRetrieveGDDnsCmd = &cobra.Command{
		Use:   "retrieve-google-domains-dns-records REGISTRATION",
		Short: "Retrieve Google Domains DNS records for a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomRetrieveGDDns,
	}
	s.domainsRegRetrieveGDForwardingCfgCmd = &cobra.Command{
		Use:   "retrieve-google-domains-forwarding-config REGISTRATION",
		Short: "Retrieve Google Domains forwarding configuration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomRetrieveGDForwardingConfig,
	}
	s.domainsRegRetrieveImportTransferCmd = &cobra.Command{
		Use:   "retrieve-import-transfer-parameters",
		Short: "List domains importable into Cloud Domains",
		Args:  cobra.NoArgs,
		RunE:  s.runDomRetrieveImportable,
	}
	s.domainsRegRetrieveRegisterParamsCmd = &cobra.Command{
		Use:   "retrieve-register-parameters",
		Short: "Retrieve pricing and eligibility for registering a domain",
		Args:  cobra.NoArgs,
		RunE:  s.runDomRetrieveRegisterParams,
	}
	s.domainsRegRetrieveTransferParamsCmd = &cobra.Command{
		Use:   "retrieve-transfer-parameters",
		Short: "Retrieve pricing and eligibility for transferring a domain",
		Args:  cobra.NoArgs,
		RunE:  s.runDomRetrieveTransferParams,
	}
	s.domainsRegSearchDomainsCmd = &cobra.Command{
		Use:   "search-domains QUERY",
		Short: "Search for available domains",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomSearch,
	}
	s.domainsRegTransferCmd = &cobra.Command{
		Use:   "transfer DOMAIN",
		Short: "Transfer a domain into Cloud Domains",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomTransfer,
	}
	s.domainsRegConfigureCmd = &cobra.Command{Use: "configure", Short: "Configure a registration's contacts, DNS, or management settings"}
	s.domainsRegConfigureContactsCmd = &cobra.Command{
		Use:   "contacts REGISTRATION",
		Short: "Configure contact settings for a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomConfigureContacts,
	}
	s.domainsRegConfigureDNSCmd = &cobra.Command{
		Use:   "dns REGISTRATION",
		Short: "Configure DNS settings for a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomConfigureDNS,
	}
	s.domainsRegConfigureMgmtCmd = &cobra.Command{
		Use:   "management REGISTRATION",
		Short: "Configure management settings for a registration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDomConfigureManagement,
	}
	s.initDomains()
	return s
}
