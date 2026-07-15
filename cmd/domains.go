package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	domainsapi "google.golang.org/api/domains/v1"
)

// --- gcloud domains (#943) ---

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage Cloud Domains",
}

var (
	flagDomLocation     string
	flagDomFile         string
	flagDomDomainName   string
	flagDomValidateOnly bool
	flagDomFilter       string
	flagDomPageSize     int64
	flagDomLimit        int64
	flagDomYearlyPrice  string
	flagDomAuthCode     string
	flagDomTag          string
	flagDomUpdateMask   string
)

// The domains v1 registrations surface always uses the "global" location. We
// keep --location as a flag for forward-compatibility but default to global.
func domLocation() string {
	if flagDomLocation != "" {
		return flagDomLocation
	}
	return "global"
}

func domParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, domLocation())
}

func domRegName(id, project string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/registrations/%s", domParent(project), id)
}

// --- Registrations subgroup ---

var domainsRegistrationsCmd = &cobra.Command{Use: "registrations", Short: "Manage domain registrations"}

var domainsRegisterCmd = &cobra.Command{
	Use:   "register DOMAIN",
	Short: "Register a domain",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomRegister,
}

var domainsRegDescribeCmd = &cobra.Command{
	Use:   "describe REGISTRATION",
	Short: "Describe a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomDescribe,
}

var domainsRegListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registrations in a location",
	Args:  cobra.NoArgs,
	RunE:  runDomList,
}

var domainsRegDeleteCmd = &cobra.Command{
	Use:   "delete REGISTRATION",
	Short: "Delete a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomDelete,
}

var domainsRegUpdateCmd = &cobra.Command{
	Use:   "update REGISTRATION",
	Short: "Update a registration (labels only via --from-file)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomUpdate,
}

var domainsRegExportCmd = &cobra.Command{
	Use:   "export REGISTRATION",
	Short: "Export a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomExport,
}

var domainsRegImportCmd = &cobra.Command{
	Use:   "import DOMAIN",
	Short: "Import an existing domain into Cloud Domains",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomImport,
}

var domainsRegInitPushCmd = &cobra.Command{
	Use:   "initiate-push-transfer REGISTRATION",
	Short: "Initiate a push transfer to another registrar",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomInitiatePushTransfer,
}

var domainsRegRenewCmd = &cobra.Command{
	Use:   "renew-domain REGISTRATION",
	Short: "Renew a registered domain",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomRenew,
}

var domainsRegRetrieveAuthCodeCmd = &cobra.Command{
	Use:   "retrieve-authorization-code REGISTRATION",
	Short: "Retrieve the authorization code for a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomRetrieveAuth,
}

var domainsRegResetAuthCodeCmd = &cobra.Command{
	Use:   "reset-authorization-code REGISTRATION",
	Short: "Reset the authorization code for a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomResetAuth,
}

var domainsRegRetrieveGDDnsCmd = &cobra.Command{
	Use:   "retrieve-google-domains-dns-records REGISTRATION",
	Short: "Retrieve Google Domains DNS records for a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomRetrieveGDDns,
}

var domainsRegRetrieveGDForwardingCfgCmd = &cobra.Command{
	Use:   "retrieve-google-domains-forwarding-config REGISTRATION",
	Short: "Retrieve Google Domains forwarding configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomRetrieveGDForwardingConfig,
}

var domainsRegRetrieveImportTransferCmd = &cobra.Command{
	Use:   "retrieve-import-transfer-parameters",
	Short: "List domains importable into Cloud Domains",
	Args:  cobra.NoArgs,
	RunE:  runDomRetrieveImportable,
}

var domainsRegRetrieveRegisterParamsCmd = &cobra.Command{
	Use:   "retrieve-register-parameters",
	Short: "Retrieve pricing and eligibility for registering a domain",
	Args:  cobra.NoArgs,
	RunE:  runDomRetrieveRegisterParams,
}

var domainsRegRetrieveTransferParamsCmd = &cobra.Command{
	Use:   "retrieve-transfer-parameters",
	Short: "Retrieve pricing and eligibility for transferring a domain",
	Args:  cobra.NoArgs,
	RunE:  runDomRetrieveTransferParams,
}

var domainsRegSearchDomainsCmd = &cobra.Command{
	Use:   "search-domains QUERY",
	Short: "Search for available domains",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomSearch,
}

var domainsRegTransferCmd = &cobra.Command{
	Use:   "transfer DOMAIN",
	Short: "Transfer a domain into Cloud Domains",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomTransfer,
}

// configure has three sub-subcommands mirroring the Python surface.
var domainsRegConfigureCmd = &cobra.Command{Use: "configure", Short: "Configure a registration's contacts, DNS, or management settings"}

var domainsRegConfigureContactsCmd = &cobra.Command{
	Use:   "contacts REGISTRATION",
	Short: "Configure contact settings for a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomConfigureContacts,
}

var domainsRegConfigureDNSCmd = &cobra.Command{
	Use:   "dns REGISTRATION",
	Short: "Configure DNS settings for a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomConfigureDNS,
}

var domainsRegConfigureMgmtCmd = &cobra.Command{
	Use:   "management REGISTRATION",
	Short: "Configure management settings for a registration",
	Args:  cobra.ExactArgs(1),
	RunE:  runDomConfigureManagement,
}

func init() {
	// Common location flag on every command that hits a project/location endpoint.
	all := []*cobra.Command{
		domainsRegisterCmd, domainsRegDescribeCmd, domainsRegListCmd,
		domainsRegDeleteCmd, domainsRegUpdateCmd, domainsRegExportCmd,
		domainsRegImportCmd, domainsRegInitPushCmd, domainsRegRenewCmd,
		domainsRegRetrieveAuthCodeCmd, domainsRegResetAuthCodeCmd,
		domainsRegRetrieveGDDnsCmd, domainsRegRetrieveGDForwardingCfgCmd,
		domainsRegRetrieveImportTransferCmd, domainsRegRetrieveRegisterParamsCmd,
		domainsRegRetrieveTransferParamsCmd, domainsRegSearchDomainsCmd,
		domainsRegTransferCmd, domainsRegConfigureContactsCmd,
		domainsRegConfigureDNSCmd, domainsRegConfigureMgmtCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDomLocation, "location", "", "Location of the registration (defaults to 'global')")
	}

	// Register / Transfer / Import / Configure take a JSON/YAML input file
	// describing the Registration (or DNS/contact settings) because the request
	// bodies are large and nested.
	domainsRegisterCmd.Flags().StringVar(&flagDomFile, "from-file", "", "JSON/YAML file with the RegisterDomainRequest body")
	domainsRegisterCmd.Flags().StringVar(&flagDomYearlyPrice, "yearly-price", "", "Yearly price acknowledgement in units.currency form (e.g. 12.00.USD)")
	domainsRegisterCmd.Flags().BoolVar(&flagDomValidateOnly, "validate-only", false, "Validate only")

	domainsRegTransferCmd.Flags().StringVar(&flagDomFile, "from-file", "", "JSON/YAML file with the TransferDomainRequest body")
	domainsRegTransferCmd.Flags().StringVar(&flagDomAuthCode, "authorization-code", "", "The transfer authorization code from the current registrar")
	domainsRegTransferCmd.Flags().StringVar(&flagDomYearlyPrice, "yearly-price", "", "Yearly price acknowledgement in units.currency form")
	domainsRegTransferCmd.Flags().BoolVar(&flagDomValidateOnly, "validate-only", false, "Validate only")

	domainsRegImportCmd.Flags().StringVar(&flagDomFile, "from-file", "", "JSON/YAML file with the ImportDomainRequest body (optional)")

	for _, c := range []*cobra.Command{domainsRegConfigureContactsCmd, domainsRegConfigureDNSCmd, domainsRegConfigureMgmtCmd} {
		c.Flags().StringVar(&flagDomFile, "from-file", "", "JSON/YAML file with the configure request body")
		c.Flags().StringVar(&flagDomUpdateMask, "update-mask", "", "Comma-separated field mask (required)")
		c.MarkFlagRequired("from-file")
	}

	domainsRegListCmd.Flags().StringVar(&flagDomFilter, "filter", "", "Server-side filter expression")
	domainsRegListCmd.Flags().Int64Var(&flagDomPageSize, "page-size", 0, "Number of results per page")
	domainsRegListCmd.Flags().Int64Var(&flagDomLimit, "limit", 0, "Maximum number of results to return")

	domainsRegRenewCmd.Flags().StringVar(&flagDomYearlyPrice, "yearly-price", "", "Yearly renewal price acknowledgement in units.currency form")
	domainsRegRenewCmd.Flags().BoolVar(&flagDomValidateOnly, "validate-only", false, "Validate only")

	domainsRegInitPushCmd.Flags().StringVar(&flagDomTag, "tag", "", "The Tag of the new registrar (required)")
	domainsRegInitPushCmd.MarkFlagRequired("tag")

	domainsRegUpdateCmd.Flags().StringVar(&flagDomFile, "from-file", "", "JSON/YAML file with the updated Registration body")
	domainsRegUpdateCmd.Flags().StringVar(&flagDomUpdateMask, "update-mask", "", "Comma-separated field mask (required)")
	domainsRegUpdateCmd.MarkFlagRequired("from-file")
	domainsRegUpdateCmd.MarkFlagRequired("update-mask")

	for _, c := range []*cobra.Command{domainsRegRetrieveRegisterParamsCmd, domainsRegRetrieveTransferParamsCmd} {
		c.Flags().StringVar(&flagDomDomainName, "domain-name", "", "The domain name to look up (required)")
		c.MarkFlagRequired("domain-name")
	}

	domainsRegistrationsCmd.AddCommand(
		domainsRegisterCmd, domainsRegDescribeCmd, domainsRegListCmd,
		domainsRegDeleteCmd, domainsRegUpdateCmd, domainsRegExportCmd,
		domainsRegImportCmd, domainsRegInitPushCmd, domainsRegRenewCmd,
		domainsRegRetrieveAuthCodeCmd, domainsRegResetAuthCodeCmd,
		domainsRegRetrieveGDDnsCmd, domainsRegRetrieveGDForwardingCfgCmd,
		domainsRegRetrieveImportTransferCmd, domainsRegRetrieveRegisterParamsCmd,
		domainsRegRetrieveTransferParamsCmd, domainsRegSearchDomainsCmd,
		domainsRegTransferCmd,
	)
	domainsRegConfigureCmd.AddCommand(
		domainsRegConfigureContactsCmd, domainsRegConfigureDNSCmd, domainsRegConfigureMgmtCmd,
	)
	domainsRegistrationsCmd.AddCommand(domainsRegConfigureCmd)
	domainsCmd.AddCommand(domainsRegistrationsCmd)
	rootCmd.AddCommand(domainsCmd)
}

// --- helpers ---

func domSvc(ctx context.Context) (*domainsapi.Service, error) {
	return gcp.DomainsService(ctx, flagAccount)
}

// --- Register ---

func runDomRegister(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &domainsapi.RegisterDomainRequest{}
	if flagDomFile != "" {
		if err := loadYAMLOrJSONInto(flagDomFile, req); err != nil {
			return err
		}
	}
	if req.Registration == nil {
		req.Registration = &domainsapi.Registration{}
	}
	if req.Registration.DomainName == "" {
		req.Registration.DomainName = args[0]
	}
	if req.YearlyPrice == nil && flagDomYearlyPrice != "" {
		money, err := parseDomainsMoney(flagDomYearlyPrice)
		if err != nil {
			return err
		}
		req.YearlyPrice = money
	}
	if flagDomValidateOnly {
		req.ValidateOnly = true
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Register(domParent(project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("registering domain: %w", err)
	}
	fmt.Printf("Register request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Describe / List / Delete / Update / Export ---

func runDomDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	reg, err := svc.Projects.Locations.Registrations.Get(domRegName(args[0], project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing registration: %w", err)
	}
	return emitFormatted(reg, "")
}

func runDomList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	var all []*domainsapi.Registration
	pageToken := ""
	for {
		call := svc.Projects.Locations.Registrations.List(domParent(project)).Context(ctx)
		if flagDomFilter != "" {
			call = call.Filter(flagDomFilter)
		}
		if flagDomPageSize > 0 {
			call = call.PageSize(flagDomPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing registrations: %w", err)
		}
		all = append(all, resp.Registrations...)
		if flagDomLimit > 0 && int64(len(all)) >= flagDomLimit {
			all = all[:flagDomLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runDomDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Delete(domRegName(args[0], project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting registration: %w", err)
	}
	fmt.Printf("Delete request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runDomUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	reg := &domainsapi.Registration{}
	if err := loadYAMLOrJSONInto(flagDomFile, reg); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Patch(domRegName(args[0], project), reg).UpdateMask(flagDomUpdateMask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating registration: %w", err)
	}
	fmt.Printf("Update request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runDomExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Export(domRegName(args[0], project), &domainsapi.ExportRegistrationRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting registration: %w", err)
	}
	fmt.Printf("Export request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Import / Initiate push / Renew ---

func runDomImport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &domainsapi.ImportDomainRequest{DomainName: args[0]}
	if flagDomFile != "" {
		if err := loadYAMLOrJSONInto(flagDomFile, req); err != nil {
			return err
		}
		if req.DomainName == "" {
			req.DomainName = args[0]
		}
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Import(domParent(project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing domain: %w", err)
	}
	fmt.Printf("Import request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runDomInitiatePushTransfer(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.InitiatePushTransfer(
		domRegName(args[0], project),
		&domainsapi.InitiatePushTransferRequest{Tag: flagDomTag},
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("initiating push transfer: %w", err)
	}
	fmt.Printf("Push transfer initiated for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runDomRenew(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &domainsapi.RenewDomainRequest{ValidateOnly: flagDomValidateOnly}
	if flagDomYearlyPrice != "" {
		money, err := parseDomainsMoney(flagDomYearlyPrice)
		if err != nil {
			return err
		}
		req.YearlyPrice = money
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.RenewDomain(domRegName(args[0], project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renewing domain: %w", err)
	}
	fmt.Printf("Renew request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Auth code / DNS / forwarding lookups ---

func runDomRetrieveAuth(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	code, err := svc.Projects.Locations.Registrations.RetrieveAuthorizationCode(domRegName(args[0], project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving authorization code: %w", err)
	}
	return emitFormatted(code, "")
}

func runDomResetAuth(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	code, err := svc.Projects.Locations.Registrations.ResetAuthorizationCode(domRegName(args[0], project), &domainsapi.ResetAuthorizationCodeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting authorization code: %w", err)
	}
	return emitFormatted(code, "")
}

func runDomRetrieveGDDns(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	var all []*domainsapi.ResourceRecordSet
	pageToken := ""
	for {
		call := svc.Projects.Locations.Registrations.RetrieveGoogleDomainsDnsRecords(domRegName(args[0], project)).Context(ctx)
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
	return emitFormatted(all, "")
}

func runDomRetrieveGDForwardingConfig(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.Registrations.RetrieveGoogleDomainsForwardingConfig(domRegName(args[0], project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving Google Domains forwarding config: %w", err)
	}
	return emitFormatted(cfg, "")
}

// --- Retrieve parameter lookups ---

func runDomRetrieveImportable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	var all []*domainsapi.Domain
	pageToken := ""
	for {
		call := svc.Projects.Locations.Registrations.RetrieveImportableDomains(domParent(project)).Context(ctx)
		if flagDomPageSize > 0 {
			call = call.PageSize(flagDomPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing importable domains: %w", err)
		}
		all = append(all, resp.Domains...)
		if flagDomLimit > 0 && int64(len(all)) >= flagDomLimit {
			all = all[:flagDomLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runDomRetrieveRegisterParams(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Registrations.RetrieveRegisterParameters(domParent(project)).DomainName(flagDomDomainName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving register parameters: %w", err)
	}
	return emitFormatted(resp, "")
}

func runDomRetrieveTransferParams(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Registrations.RetrieveTransferParameters(domParent(project)).DomainName(flagDomDomainName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving transfer parameters: %w", err)
	}
	return emitFormatted(resp, "")
}

func runDomSearch(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Registrations.SearchDomains(domParent(project)).Query(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("searching domains: %w", err)
	}
	return emitFormatted(resp, "")
}

// --- Transfer ---

func runDomTransfer(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &domainsapi.TransferDomainRequest{}
	if flagDomFile != "" {
		if err := loadYAMLOrJSONInto(flagDomFile, req); err != nil {
			return err
		}
	}
	if req.Registration == nil {
		req.Registration = &domainsapi.Registration{}
	}
	if req.Registration.DomainName == "" {
		req.Registration.DomainName = args[0]
	}
	if flagDomAuthCode != "" {
		req.AuthorizationCode = &domainsapi.AuthorizationCode{Code: flagDomAuthCode}
	}
	if req.YearlyPrice == nil && flagDomYearlyPrice != "" {
		money, err := parseDomainsMoney(flagDomYearlyPrice)
		if err != nil {
			return err
		}
		req.YearlyPrice = money
	}
	if flagDomValidateOnly {
		req.ValidateOnly = true
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.Transfer(domParent(project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("transferring domain: %w", err)
	}
	fmt.Printf("Transfer request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

// --- Configure sub-subcommands ---

func runDomConfigureContacts(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &domainsapi.ConfigureContactSettingsRequest{}
	if err := loadYAMLOrJSONInto(flagDomFile, req); err != nil {
		return err
	}
	if flagDomUpdateMask != "" {
		req.UpdateMask = flagDomUpdateMask
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.ConfigureContactSettings(domRegName(args[0], project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("configuring contacts: %w", err)
	}
	fmt.Printf("Configure request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runDomConfigureDNS(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &domainsapi.ConfigureDnsSettingsRequest{}
	if err := loadYAMLOrJSONInto(flagDomFile, req); err != nil {
		return err
	}
	if flagDomUpdateMask != "" {
		req.UpdateMask = flagDomUpdateMask
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.ConfigureDnsSettings(domRegName(args[0], project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("configuring DNS: %w", err)
	}
	fmt.Printf("Configure request issued for [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runDomConfigureManagement(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &domainsapi.ConfigureManagementSettingsRequest{}
	if err := loadYAMLOrJSONInto(flagDomFile, req); err != nil {
		return err
	}
	if flagDomUpdateMask != "" {
		req.UpdateMask = flagDomUpdateMask
	}
	ctx := context.Background()
	svc, err := domSvc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Registrations.ConfigureManagementSettings(domRegName(args[0], project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("configuring management: %w", err)
	}
	fmt.Printf("Configure request issued for [%s] (operation: %s)\n", args[0], op.Name)
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
