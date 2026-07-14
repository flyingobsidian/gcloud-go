package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	certificatemanager "google.golang.org/api/certificatemanager/v1"
)

// --- shared helpers ---

func cmLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func cmChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func cmWaitOp(ctx context.Context, svc *certificatemanager.Service, op *certificatemanager.Operation) (*certificatemanager.Operation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func cmFinishOp(ctx context.Context, svc *certificatemanager.Service, op *certificatemanager.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := cmWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// Common shared flags.
var (
	flagCMLocation   string
	flagCMConfigFile string
	flagCMUpdateMask string
	flagCMFormat     string
	flagCMAsync      bool
)

// --- certificates ---

var cmCertificatesCmd = &cobra.Command{Use: "certificates", Short: "Manage Certificate Manager certificates"}

var (
	cmCertCreateCmd = &cobra.Command{
		Use: "create CERT", Short: "Create a certificate from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMCertCreate,
	}
	cmCertDeleteCmd = &cobra.Command{
		Use: "delete CERT", Short: "Delete a certificate",
		Args: cobra.ExactArgs(1), RunE: runCMCertDelete,
	}
	cmCertDescribeCmd = &cobra.Command{
		Use: "describe CERT", Short: "Describe a certificate",
		Args: cobra.ExactArgs(1), RunE: runCMCertDescribe,
	}
	cmCertListCmd = &cobra.Command{
		Use: "list", Short: "List certificates",
		Args: cobra.NoArgs, RunE: runCMCertList,
	}
	cmCertUpdateCmd = &cobra.Command{
		Use: "update CERT", Short: "Update a certificate from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMCertUpdate,
	}
)

func init() {
	registerCMCrud(cmCertCreateCmd, cmCertDeleteCmd, cmCertDescribeCmd, cmCertListCmd, cmCertUpdateCmd, "Certificate")
	cmCertificatesCmd.AddCommand(cmCertCreateCmd, cmCertDeleteCmd, cmCertDescribeCmd, cmCertListCmd, cmCertUpdateCmd)
	certificateManagerCmd.AddCommand(cmCertificatesCmd)
}

func registerCMCrud(create, del, describe, list, update *cobra.Command, msg string) {
	for _, c := range []*cobra.Command{create, del, describe, list, update} {
		c.Flags().StringVar(&flagCMLocation, "location", "global", "Location containing the resource")
	}
	for _, c := range []*cobra.Command{create, update} {
		c.Flags().StringVar(&flagCMConfigFile, "config-file", "",
			fmt.Sprintf("Path to a JSON/YAML file with the %s message body (required)", msg))
		_ = c.MarkFlagRequired("config-file")
	}
	update.Flags().StringVar(&flagCMUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{create, del, update} {
		c.Flags().BoolVar(&flagCMAsync, "async", false, "Return the long-running operation without waiting")
	}
	describe.Flags().StringVar(&flagCMFormat, "format", "", "Output format")
	list.Flags().StringVar(&flagCMFormat, "format", "", "Output format")
}

func cmCertName(id, project, location string) string {
	return cmChild("certificates", id, cmLocationParent(project, location))
}

func runCMCertCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cert := &certificatemanager.Certificate{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, cert); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Certificates.Create(cmLocationParent(project, flagCMLocation), cert).
		CertificateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating certificate: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Create certificate", args[0], flagCMAsync)
}

func runCMCertDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Certificates.Delete(cmCertName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting certificate: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Delete certificate", args[0], flagCMAsync)
}

func runCMCertDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Certificates.Get(cmCertName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing certificate: %w", err)
	}
	return emitFormatted(got, flagCMFormat)
}

func runCMCertList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Certificates.List(cmLocationParent(project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing certificates: %w", err)
	}
	if flagCMFormat != "" {
		return emitFormatted(resp.Certificates, flagCMFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, c := range resp.Certificates {
		fmt.Println(path.Base(c.Name))
	}
	return nil
}

func runCMCertUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cert := &certificatemanager.Certificate{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, cert); err != nil {
		return err
	}
	mask := flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cert))
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Certificates.Patch(cmCertName(args[0], project, flagCMLocation), cert).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating certificate: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Update certificate", args[0], flagCMAsync)
}

// --- dns-authorizations ---

var cmDnsAuthCmd = &cobra.Command{Use: "dns-authorizations", Short: "Manage DNS authorizations"}

var (
	cmDACreateCmd = &cobra.Command{
		Use: "create DA", Short: "Create a DNS authorization from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMDACreate,
	}
	cmDADeleteCmd = &cobra.Command{
		Use: "delete DA", Short: "Delete a DNS authorization",
		Args: cobra.ExactArgs(1), RunE: runCMDADelete,
	}
	cmDADescribeCmd = &cobra.Command{
		Use: "describe DA", Short: "Describe a DNS authorization",
		Args: cobra.ExactArgs(1), RunE: runCMDADescribe,
	}
	cmDAListCmd = &cobra.Command{
		Use: "list", Short: "List DNS authorizations",
		Args: cobra.NoArgs, RunE: runCMDAList,
	}
	cmDAUpdateCmd = &cobra.Command{
		Use: "update DA", Short: "Update a DNS authorization from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMDAUpdate,
	}
)

func init() {
	registerCMCrud(cmDACreateCmd, cmDADeleteCmd, cmDADescribeCmd, cmDAListCmd, cmDAUpdateCmd, "DnsAuthorization")
	cmDnsAuthCmd.AddCommand(cmDACreateCmd, cmDADeleteCmd, cmDADescribeCmd, cmDAListCmd, cmDAUpdateCmd)
	certificateManagerCmd.AddCommand(cmDnsAuthCmd)
}

func cmDAName(id, project, location string) string {
	return cmChild("dnsAuthorizations", id, cmLocationParent(project, location))
}

func runCMDACreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	da := &certificatemanager.DnsAuthorization{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, da); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsAuthorizations.Create(cmLocationParent(project, flagCMLocation), da).
		DnsAuthorizationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DNS authorization: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Create DNS authorization", args[0], flagCMAsync)
}

func runCMDADelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsAuthorizations.Delete(cmDAName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting DNS authorization: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Delete DNS authorization", args[0], flagCMAsync)
}

func runCMDADescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DnsAuthorizations.Get(cmDAName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS authorization: %w", err)
	}
	return emitFormatted(got, flagCMFormat)
}

func runCMDAList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DnsAuthorizations.List(cmLocationParent(project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing DNS authorizations: %w", err)
	}
	if flagCMFormat != "" {
		return emitFormatted(resp.DnsAuthorizations, flagCMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DOMAIN")
	for _, d := range resp.DnsAuthorizations {
		fmt.Printf("%-40s %s\n", path.Base(d.Name), d.Domain)
	}
	return nil
}

func runCMDAUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	da := &certificatemanager.DnsAuthorization{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, da); err != nil {
		return err
	}
	mask := flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(da))
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsAuthorizations.Patch(cmDAName(args[0], project, flagCMLocation), da).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating DNS authorization: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Update DNS authorization", args[0], flagCMAsync)
}

// --- issuance-configs ---

var cmIssuanceConfigsCmd = &cobra.Command{Use: "issuance-configs", Short: "Manage certificate issuance configs"}

var (
	cmICCreateCmd = &cobra.Command{
		Use: "create IC", Short: "Create an issuance config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMICCreate,
	}
	cmICDeleteCmd = &cobra.Command{
		Use: "delete IC", Short: "Delete an issuance config",
		Args: cobra.ExactArgs(1), RunE: runCMICDelete,
	}
	cmICDescribeCmd = &cobra.Command{
		Use: "describe IC", Short: "Describe an issuance config",
		Args: cobra.ExactArgs(1), RunE: runCMICDescribe,
	}
	cmICListCmd = &cobra.Command{
		Use: "list", Short: "List issuance configs",
		Args: cobra.NoArgs, RunE: runCMICList,
	}
	cmICUpdateCmd = &cobra.Command{
		Use: "update IC", Short: "Update an issuance config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMICUpdate,
	}
)

func init() {
	registerCMCrud(cmICCreateCmd, cmICDeleteCmd, cmICDescribeCmd, cmICListCmd, cmICUpdateCmd, "CertificateIssuanceConfig")
	cmIssuanceConfigsCmd.AddCommand(cmICCreateCmd, cmICDeleteCmd, cmICDescribeCmd, cmICListCmd, cmICUpdateCmd)
	certificateManagerCmd.AddCommand(cmIssuanceConfigsCmd)
}

func cmICName(id, project, location string) string {
	return cmChild("certificateIssuanceConfigs", id, cmLocationParent(project, location))
}

func runCMICCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ic := &certificatemanager.CertificateIssuanceConfig{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, ic); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateIssuanceConfigs.Create(cmLocationParent(project, flagCMLocation), ic).
		CertificateIssuanceConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating issuance config: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Create issuance config", args[0], flagCMAsync)
}

func runCMICDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateIssuanceConfigs.Delete(cmICName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting issuance config: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Delete issuance config", args[0], flagCMAsync)
}

func runCMICDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CertificateIssuanceConfigs.Get(cmICName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing issuance config: %w", err)
	}
	return emitFormatted(got, flagCMFormat)
}

func runCMICList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CertificateIssuanceConfigs.List(cmLocationParent(project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing issuance configs: %w", err)
	}
	if flagCMFormat != "" {
		return emitFormatted(resp.CertificateIssuanceConfigs, flagCMFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, i := range resp.CertificateIssuanceConfigs {
		fmt.Println(path.Base(i.Name))
	}
	return nil
}

func runCMICUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ic := &certificatemanager.CertificateIssuanceConfig{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, ic); err != nil {
		return err
	}
	mask := flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ic))
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateIssuanceConfigs.Patch(cmICName(args[0], project, flagCMLocation), ic).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating issuance config: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Update issuance config", args[0], flagCMAsync)
}

// --- maps ---

var cmMapsCmd = &cobra.Command{Use: "maps", Short: "Manage certificate maps"}

var (
	cmMapCreateCmd = &cobra.Command{
		Use: "create MAP", Short: "Create a certificate map from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMMapCreate,
	}
	cmMapDeleteCmd = &cobra.Command{
		Use: "delete MAP", Short: "Delete a certificate map",
		Args: cobra.ExactArgs(1), RunE: runCMMapDelete,
	}
	cmMapDescribeCmd = &cobra.Command{
		Use: "describe MAP", Short: "Describe a certificate map",
		Args: cobra.ExactArgs(1), RunE: runCMMapDescribe,
	}
	cmMapListCmd = &cobra.Command{
		Use: "list", Short: "List certificate maps",
		Args: cobra.NoArgs, RunE: runCMMapList,
	}
	cmMapUpdateCmd = &cobra.Command{
		Use: "update MAP", Short: "Update a certificate map from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMMapUpdate,
	}
)

func init() {
	registerCMCrud(cmMapCreateCmd, cmMapDeleteCmd, cmMapDescribeCmd, cmMapListCmd, cmMapUpdateCmd, "CertificateMap")
	cmMapsCmd.AddCommand(cmMapCreateCmd, cmMapDeleteCmd, cmMapDescribeCmd, cmMapListCmd, cmMapUpdateCmd)
	registerCMMapEntries(cmMapsCmd)
	certificateManagerCmd.AddCommand(cmMapsCmd)
}

func cmMapName(id, project, location string) string {
	return cmChild("certificateMaps", id, cmLocationParent(project, location))
}

func runCMMapCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	m := &certificatemanager.CertificateMap{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, m); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.Create(cmLocationParent(project, flagCMLocation), m).
		CertificateMapId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating map: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Create map", args[0], flagCMAsync)
}

func runCMMapDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.Delete(cmMapName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting map: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Delete map", args[0], flagCMAsync)
}

func runCMMapDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CertificateMaps.Get(cmMapName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing map: %w", err)
	}
	return emitFormatted(got, flagCMFormat)
}

func runCMMapList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CertificateMaps.List(cmLocationParent(project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing maps: %w", err)
	}
	if flagCMFormat != "" {
		return emitFormatted(resp.CertificateMaps, flagCMFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, m := range resp.CertificateMaps {
		fmt.Println(path.Base(m.Name))
	}
	return nil
}

func runCMMapUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	m := &certificatemanager.CertificateMap{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, m); err != nil {
		return err
	}
	mask := flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(m))
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.Patch(cmMapName(args[0], project, flagCMLocation), m).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating map: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Update map", args[0], flagCMAsync)
}

// --- maps entries subgroup ---

var cmMapEntriesCmd = &cobra.Command{Use: "entries", Short: "Manage certificate map entries"}

var (
	cmMECreateCmd = &cobra.Command{
		Use: "create ENTRY", Short: "Create a map entry from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMMECreate,
	}
	cmMEDeleteCmd = &cobra.Command{
		Use: "delete ENTRY", Short: "Delete a map entry",
		Args: cobra.ExactArgs(1), RunE: runCMMEDelete,
	}
	cmMEDescribeCmd = &cobra.Command{
		Use: "describe ENTRY", Short: "Describe a map entry",
		Args: cobra.ExactArgs(1), RunE: runCMMEDescribe,
	}
	cmMEListCmd = &cobra.Command{
		Use: "list", Short: "List map entries",
		Args: cobra.NoArgs, RunE: runCMMEList,
	}
	cmMEUpdateCmd = &cobra.Command{
		Use: "update ENTRY", Short: "Update a map entry from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMMEUpdate,
	}
)

var flagCMMEMap string

func registerCMMapEntries(parent *cobra.Command) {
	for _, c := range []*cobra.Command{cmMECreateCmd, cmMEDeleteCmd, cmMEDescribeCmd, cmMEListCmd, cmMEUpdateCmd} {
		c.Flags().StringVar(&flagCMLocation, "location", "global", "Location containing the certificate map")
		c.Flags().StringVar(&flagCMMEMap, "map", "", "Certificate map containing the entry (required)")
		_ = c.MarkFlagRequired("map")
	}
	for _, c := range []*cobra.Command{cmMECreateCmd, cmMEUpdateCmd} {
		c.Flags().StringVar(&flagCMConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the CertificateMapEntry body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	cmMEUpdateCmd.Flags().StringVar(&flagCMUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{cmMECreateCmd, cmMEDeleteCmd, cmMEUpdateCmd} {
		c.Flags().BoolVar(&flagCMAsync, "async", false, "Return the long-running operation without waiting")
	}
	cmMEDescribeCmd.Flags().StringVar(&flagCMFormat, "format", "", "Output format")
	cmMEListCmd.Flags().StringVar(&flagCMFormat, "format", "", "Output format")

	cmMapEntriesCmd.AddCommand(cmMECreateCmd, cmMEDeleteCmd, cmMEDescribeCmd, cmMEListCmd, cmMEUpdateCmd)
	parent.AddCommand(cmMapEntriesCmd)
}

func cmMEParent(project, location, mapID string) string {
	return fmt.Sprintf("%s/certificateMaps/%s", cmLocationParent(project, location), mapID)
}

func cmMEName(id, project, location, mapID string) string {
	return cmChild("certificateMapEntries", id, cmMEParent(project, location, mapID))
}

func runCMMECreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	e := &certificatemanager.CertificateMapEntry{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, e); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.Create(cmMEParent(project, flagCMLocation, flagCMMEMap), e).
		CertificateMapEntryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating map entry: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Create map entry", args[0], flagCMAsync)
}

func runCMMEDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.Delete(cmMEName(args[0], project, flagCMLocation, flagCMMEMap)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting map entry: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Delete map entry", args[0], flagCMAsync)
}

func runCMMEDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.Get(cmMEName(args[0], project, flagCMLocation, flagCMMEMap)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing map entry: %w", err)
	}
	return emitFormatted(got, flagCMFormat)
}

func runCMMEList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.List(cmMEParent(project, flagCMLocation, flagCMMEMap)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing map entries: %w", err)
	}
	if flagCMFormat != "" {
		return emitFormatted(resp.CertificateMapEntries, flagCMFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, e := range resp.CertificateMapEntries {
		fmt.Println(path.Base(e.Name))
	}
	return nil
}

func runCMMEUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	e := &certificatemanager.CertificateMapEntry{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, e); err != nil {
		return err
	}
	mask := flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(e))
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.Patch(cmMEName(args[0], project, flagCMLocation, flagCMMEMap), e).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating map entry: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Update map entry", args[0], flagCMAsync)
}

// --- trust-configs ---

var cmTrustConfigsCmd = &cobra.Command{Use: "trust-configs", Short: "Manage trust configs"}

var (
	cmTCCreateCmd = &cobra.Command{
		Use: "create TC", Short: "Create a trust config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMTCCreate,
	}
	cmTCDeleteCmd = &cobra.Command{
		Use: "delete TC", Short: "Delete a trust config",
		Args: cobra.ExactArgs(1), RunE: runCMTCDelete,
	}
	cmTCDescribeCmd = &cobra.Command{
		Use: "describe TC", Short: "Describe a trust config",
		Args: cobra.ExactArgs(1), RunE: runCMTCDescribe,
	}
	cmTCListCmd = &cobra.Command{
		Use: "list", Short: "List trust configs",
		Args: cobra.NoArgs, RunE: runCMTCList,
	}
	cmTCUpdateCmd = &cobra.Command{
		Use: "update TC", Short: "Update a trust config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runCMTCUpdate,
	}
)

func init() {
	registerCMCrud(cmTCCreateCmd, cmTCDeleteCmd, cmTCDescribeCmd, cmTCListCmd, cmTCUpdateCmd, "TrustConfig")
	cmTrustConfigsCmd.AddCommand(cmTCCreateCmd, cmTCDeleteCmd, cmTCDescribeCmd, cmTCListCmd, cmTCUpdateCmd)
	certificateManagerCmd.AddCommand(cmTrustConfigsCmd)
}

func cmTCName(id, project, location string) string {
	return cmChild("trustConfigs", id, cmLocationParent(project, location))
}

func runCMTCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	tc := &certificatemanager.TrustConfig{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, tc); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TrustConfigs.Create(cmLocationParent(project, flagCMLocation), tc).
		TrustConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating trust config: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Create trust config", args[0], flagCMAsync)
}

func runCMTCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TrustConfigs.Delete(cmTCName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting trust config: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Delete trust config", args[0], flagCMAsync)
}

func runCMTCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TrustConfigs.Get(cmTCName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing trust config: %w", err)
	}
	return emitFormatted(got, flagCMFormat)
}

func runCMTCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.TrustConfigs.List(cmLocationParent(project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing trust configs: %w", err)
	}
	if flagCMFormat != "" {
		return emitFormatted(resp.TrustConfigs, flagCMFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, t := range resp.TrustConfigs {
		fmt.Println(path.Base(t.Name))
	}
	return nil
}

func runCMTCUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	tc := &certificatemanager.TrustConfig{}
	if err := loadYAMLOrJSONInto(flagCMConfigFile, tc); err != nil {
		return err
	}
	mask := flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(tc))
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TrustConfigs.Patch(cmTCName(args[0], project, flagCMLocation), tc).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating trust config: %w", err)
	}
	return cmFinishOp(ctx, svc, op, "Update trust config", args[0], flagCMAsync)
}

// --- operations ---

var cmOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Certificate Manager operations"}

var (
	cmOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Certificate Manager operation",
		Args: cobra.ExactArgs(1), RunE: runCMOpDescribe,
	}
	cmOpListCmd = &cobra.Command{
		Use: "list", Short: "List Certificate Manager operations in a location",
		Args: cobra.NoArgs, RunE: runCMOpList,
	}
)

func init() {
	for _, c := range []*cobra.Command{cmOpDescribeCmd, cmOpListCmd} {
		c.Flags().StringVar(&flagCMLocation, "location", "global", "Location containing the operation")
	}
	cmOpDescribeCmd.Flags().StringVar(&flagCMFormat, "format", "", "Output format")
	cmOpListCmd.Flags().StringVar(&flagCMFormat, "format", "", "Output format")

	cmOperationsCmd.AddCommand(cmOpDescribeCmd, cmOpListCmd)
	certificateManagerCmd.AddCommand(cmOperationsCmd)
}

func cmOpName(id, project, location string) string {
	return cmChild("operations", id, cmLocationParent(project, location))
}

func runCMOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(cmOpName(args[0], project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagCMFormat)
}

func runCMOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CertificateManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(cmLocationParent(project, flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagCMFormat != "" {
		return emitFormatted(resp.Operations, flagCMFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
