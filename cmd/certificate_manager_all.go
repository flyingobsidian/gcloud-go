package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

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

func (s *certificateManagerCmds) cmFinishOp(ctx context.Context, svc *certificatemanager.Service, op *certificatemanager.Operation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := cmWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// cmMergeTags overlays --tags entries onto an existing resource-manager tag
// map. Values from the flag win on key collision so `--tags` behaves as an
// explicit override of anything preloaded from --config-file.
func (s *certificateManagerCmds) cmMergeTags(existing map[string]string) map[string]string {
	if len(s.flagCMTags) == 0 {
		return existing
	}
	if existing == nil {
		existing = make(map[string]string, len(s.flagCMTags))
	}
	for k, v := range s.flagCMTags {
		existing[k] = v
	}
	return existing
}

// --- certificates ---

func (s *certificateManagerCmds) initCertificateManagerAll() {
	s.registerCMCrud(s.cmCertCreateCmd, s.cmCertDeleteCmd, s.cmCertDescribeCmd, s.cmCertListCmd, s.cmCertUpdateCmd, "Certificate")
	s.cmCertificatesCmd.AddCommand(s.cmCertCreateCmd, s.cmCertDeleteCmd, s.cmCertDescribeCmd, s.cmCertListCmd, s.cmCertUpdateCmd)
	s.certificateManagerCmd.AddCommand(s.cmCertificatesCmd)
}

func (s *certificateManagerCmds) registerCMCrud(create, del, describe, list, update *cobra.Command, msg string) {
	for _, c := range []*cobra.Command{create, del, describe, list, update} {
		c.Flags().StringVar(&s.flagCMLocation, "location", "global", "Location containing the resource")
	}
	for _, c := range []*cobra.Command{create, update} {
		c.Flags().StringVar(&s.flagCMConfigFile, "config-file", "",
			fmt.Sprintf("Path to a JSON/YAML file with the %s message body (required)", msg))
		_ = c.MarkFlagRequired("config-file")
	}
	update.Flags().StringVar(&s.flagCMUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{create, del, update} {
		c.Flags().BoolVar(&s.flagCMAsync, "async", false, "Return the long-running operation without waiting")
	}
	// --tags accepts comma-separated KEY=VALUE resource-manager tag bindings,
	// e.g. `--tags=123/environment=prod,123/costCenter=marketing`. Only exposed
	// on create — updating a resource's tag bindings goes through the Resource
	// Manager API, not Certificate Manager's Patch.
	create.Flags().StringToStringVar(&s.flagCMTags, "tags", nil,
		"KEY=VALUE resource-manager tags to bind to the "+msg+" on creation")
	describe.Flags().StringVar(&s.flagCMFormat, "format", "", "Output format")
	list.Flags().StringVar(&s.flagCMFormat, "format", "", "Output format")
}

func cmCertName(id, project, location string) string {
	return cmChild("certificates", id, cmLocationParent(project, location))
}

func (s *certificateManagerCmds) runCMCertCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cert := &certificatemanager.Certificate{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, cert); err != nil {
		return err
	}
	cert.Tags = s.cmMergeTags(cert.Tags)
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Certificates.Create(cmLocationParent(project, s.flagCMLocation), cert).
		CertificateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating certificate: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Create certificate", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMCertDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Certificates.Delete(cmCertName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting certificate: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Delete certificate", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMCertDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Certificates.Get(cmCertName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing certificate: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCMFormat)
}

func (s *certificateManagerCmds) runCMCertList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Certificates.List(cmLocationParent(project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing certificates: %w", err)
	}
	if s.flagCMFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Certificates, s.flagCMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Certificates), "%-40s\n", "NAME")
	for _, c := range resp.Certificates {
		fmt.Fprintln(s.inv.stdout, path.Base(c.Name))
	}
	return nil
}

func (s *certificateManagerCmds) runCMCertUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cert := &certificatemanager.Certificate{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, cert); err != nil {
		return err
	}
	mask := s.flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cert))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Certificates.Patch(cmCertName(args[0], project, s.flagCMLocation), cert).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating certificate: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Update certificate", args[0], s.flagCMAsync)
}

// --- dns-authorizations ---

func (s *certificateManagerCmds) initCertificateManagerAll2() {
	s.registerCMCrud(s.cmDACreateCmd, s.cmDADeleteCmd, s.cmDADescribeCmd, s.cmDAListCmd, s.cmDAUpdateCmd, "DnsAuthorization")
	s.cmDnsAuthCmd.AddCommand(s.cmDACreateCmd, s.cmDADeleteCmd, s.cmDADescribeCmd, s.cmDAListCmd, s.cmDAUpdateCmd)
	s.certificateManagerCmd.AddCommand(s.cmDnsAuthCmd)
}

func cmDAName(id, project, location string) string {
	return cmChild("dnsAuthorizations", id, cmLocationParent(project, location))
}

func (s *certificateManagerCmds) runCMDACreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	da := &certificatemanager.DnsAuthorization{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, da); err != nil {
		return err
	}
	da.Tags = s.cmMergeTags(da.Tags)
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsAuthorizations.Create(cmLocationParent(project, s.flagCMLocation), da).
		DnsAuthorizationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DNS authorization: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Create DNS authorization", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMDADelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsAuthorizations.Delete(cmDAName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting DNS authorization: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Delete DNS authorization", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMDADescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DnsAuthorizations.Get(cmDAName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DNS authorization: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCMFormat)
}

func (s *certificateManagerCmds) runCMDAList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.DnsAuthorizations.List(cmLocationParent(project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing DNS authorizations: %w", err)
	}
	if s.flagCMFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.DnsAuthorizations, s.flagCMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.DnsAuthorizations), "%-40s %s\n", "NAME", "DOMAIN")
	for _, d := range resp.DnsAuthorizations {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(d.Name), d.Domain)
	}
	return nil
}

func (s *certificateManagerCmds) runCMDAUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	da := &certificatemanager.DnsAuthorization{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, da); err != nil {
		return err
	}
	mask := s.flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(da))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsAuthorizations.Patch(cmDAName(args[0], project, s.flagCMLocation), da).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating DNS authorization: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Update DNS authorization", args[0], s.flagCMAsync)
}

// --- issuance-configs ---

func (s *certificateManagerCmds) initCertificateManagerAll3() {
	s.registerCMCrud(s.cmICCreateCmd, s.cmICDeleteCmd, s.cmICDescribeCmd, s.cmICListCmd, s.cmICUpdateCmd, "CertificateIssuanceConfig")
	s.cmIssuanceConfigsCmd.AddCommand(s.cmICCreateCmd, s.cmICDeleteCmd, s.cmICDescribeCmd, s.cmICListCmd, s.cmICUpdateCmd)
	s.certificateManagerCmd.AddCommand(s.cmIssuanceConfigsCmd)
}

func cmICName(id, project, location string) string {
	return cmChild("certificateIssuanceConfigs", id, cmLocationParent(project, location))
}

func (s *certificateManagerCmds) runCMICCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ic := &certificatemanager.CertificateIssuanceConfig{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, ic); err != nil {
		return err
	}
	ic.Tags = s.cmMergeTags(ic.Tags)
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateIssuanceConfigs.Create(cmLocationParent(project, s.flagCMLocation), ic).
		CertificateIssuanceConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating issuance config: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Create issuance config", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMICDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateIssuanceConfigs.Delete(cmICName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting issuance config: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Delete issuance config", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMICDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CertificateIssuanceConfigs.Get(cmICName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing issuance config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCMFormat)
}

func (s *certificateManagerCmds) runCMICList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CertificateIssuanceConfigs.List(cmLocationParent(project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing issuance configs: %w", err)
	}
	if s.flagCMFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.CertificateIssuanceConfigs, s.flagCMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.CertificateIssuanceConfigs), "%-40s\n", "NAME")
	for _, i := range resp.CertificateIssuanceConfigs {
		fmt.Fprintln(s.inv.stdout, path.Base(i.Name))
	}
	return nil
}

func (s *certificateManagerCmds) runCMICUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ic := &certificatemanager.CertificateIssuanceConfig{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, ic); err != nil {
		return err
	}
	mask := s.flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ic))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateIssuanceConfigs.Patch(cmICName(args[0], project, s.flagCMLocation), ic).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating issuance config: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Update issuance config", args[0], s.flagCMAsync)
}

// --- maps ---

func (s *certificateManagerCmds) initCertificateManagerAll4() {
	s.registerCMCrud(s.cmMapCreateCmd, s.cmMapDeleteCmd, s.cmMapDescribeCmd, s.cmMapListCmd, s.cmMapUpdateCmd, "CertificateMap")
	s.cmMapsCmd.AddCommand(s.cmMapCreateCmd, s.cmMapDeleteCmd, s.cmMapDescribeCmd, s.cmMapListCmd, s.cmMapUpdateCmd)
	s.registerCMMapEntries(s.cmMapsCmd)
	s.certificateManagerCmd.AddCommand(s.cmMapsCmd)
}

func cmMapName(id, project, location string) string {
	return cmChild("certificateMaps", id, cmLocationParent(project, location))
}

func (s *certificateManagerCmds) runCMMapCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	m := &certificatemanager.CertificateMap{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, m); err != nil {
		return err
	}
	m.Tags = s.cmMergeTags(m.Tags)
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.Create(cmLocationParent(project, s.flagCMLocation), m).
		CertificateMapId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating map: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Create map", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMMapDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.Delete(cmMapName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting map: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Delete map", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMMapDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CertificateMaps.Get(cmMapName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing map: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCMFormat)
}

func (s *certificateManagerCmds) runCMMapList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CertificateMaps.List(cmLocationParent(project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing maps: %w", err)
	}
	if s.flagCMFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.CertificateMaps, s.flagCMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.CertificateMaps), "%-40s\n", "NAME")
	for _, m := range resp.CertificateMaps {
		fmt.Fprintln(s.inv.stdout, path.Base(m.Name))
	}
	return nil
}

func (s *certificateManagerCmds) runCMMapUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	m := &certificatemanager.CertificateMap{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, m); err != nil {
		return err
	}
	mask := s.flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(m))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.Patch(cmMapName(args[0], project, s.flagCMLocation), m).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating map: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Update map", args[0], s.flagCMAsync)
}

// --- maps entries subgroup ---

func (s *certificateManagerCmds) registerCMMapEntries(parent *cobra.Command) {
	for _, c := range []*cobra.Command{s.cmMECreateCmd, s.cmMEDeleteCmd, s.cmMEDescribeCmd, s.cmMEListCmd, s.cmMEUpdateCmd} {
		c.Flags().StringVar(&s.flagCMLocation, "location", "global", "Location containing the certificate map")
		c.Flags().StringVar(&s.flagCMMEMap, "map", "", "Certificate map containing the entry (required)")
		_ = c.MarkFlagRequired("map")
	}
	for _, c := range []*cobra.Command{s.cmMECreateCmd, s.cmMEUpdateCmd} {
		c.Flags().StringVar(&s.flagCMConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the CertificateMapEntry body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.cmMEUpdateCmd.Flags().StringVar(&s.flagCMUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.cmMECreateCmd, s.cmMEDeleteCmd, s.cmMEUpdateCmd} {
		c.Flags().BoolVar(&s.flagCMAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.cmMEDescribeCmd.Flags().StringVar(&s.flagCMFormat, "format", "", "Output format")
	s.cmMEListCmd.Flags().StringVar(&s.flagCMFormat, "format", "", "Output format")

	s.cmMapEntriesCmd.AddCommand(s.cmMECreateCmd, s.cmMEDeleteCmd, s.cmMEDescribeCmd, s.cmMEListCmd, s.cmMEUpdateCmd)
	parent.AddCommand(s.cmMapEntriesCmd)
}

func cmMEParent(project, location, mapID string) string {
	return fmt.Sprintf("%s/certificateMaps/%s", cmLocationParent(project, location), mapID)
}

func cmMEName(id, project, location, mapID string) string {
	return cmChild("certificateMapEntries", id, cmMEParent(project, location, mapID))
}

func (s *certificateManagerCmds) runCMMECreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	e := &certificatemanager.CertificateMapEntry{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, e); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.Create(cmMEParent(project, s.flagCMLocation, s.flagCMMEMap), e).
		CertificateMapEntryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating map entry: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Create map entry", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMMEDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.Delete(cmMEName(args[0], project, s.flagCMLocation, s.flagCMMEMap)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting map entry: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Delete map entry", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMMEDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.Get(cmMEName(args[0], project, s.flagCMLocation, s.flagCMMEMap)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing map entry: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCMFormat)
}

func (s *certificateManagerCmds) runCMMEList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.List(cmMEParent(project, s.flagCMLocation, s.flagCMMEMap)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing map entries: %w", err)
	}
	if s.flagCMFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.CertificateMapEntries, s.flagCMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.CertificateMapEntries), "%-40s\n", "NAME")
	for _, e := range resp.CertificateMapEntries {
		fmt.Fprintln(s.inv.stdout, path.Base(e.Name))
	}
	return nil
}

func (s *certificateManagerCmds) runCMMEUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	e := &certificatemanager.CertificateMapEntry{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, e); err != nil {
		return err
	}
	mask := s.flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(e))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateMaps.CertificateMapEntries.Patch(cmMEName(args[0], project, s.flagCMLocation, s.flagCMMEMap), e).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating map entry: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Update map entry", args[0], s.flagCMAsync)
}

// --- trust-configs ---

func (s *certificateManagerCmds) initCertificateManagerAll5() {
	s.registerCMCrud(s.cmTCCreateCmd, s.cmTCDeleteCmd, s.cmTCDescribeCmd, s.cmTCListCmd, s.cmTCUpdateCmd, "TrustConfig")
	s.cmTrustConfigsCmd.AddCommand(s.cmTCCreateCmd, s.cmTCDeleteCmd, s.cmTCDescribeCmd, s.cmTCListCmd, s.cmTCUpdateCmd)
	s.certificateManagerCmd.AddCommand(s.cmTrustConfigsCmd)
}

func cmTCName(id, project, location string) string {
	return cmChild("trustConfigs", id, cmLocationParent(project, location))
}

func (s *certificateManagerCmds) runCMTCCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	tc := &certificatemanager.TrustConfig{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, tc); err != nil {
		return err
	}
	tc.Tags = s.cmMergeTags(tc.Tags)
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TrustConfigs.Create(cmLocationParent(project, s.flagCMLocation), tc).
		TrustConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating trust config: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Create trust config", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMTCDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TrustConfigs.Delete(cmTCName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting trust config: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Delete trust config", args[0], s.flagCMAsync)
}

func (s *certificateManagerCmds) runCMTCDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TrustConfigs.Get(cmTCName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing trust config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCMFormat)
}

func (s *certificateManagerCmds) runCMTCList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.TrustConfigs.List(cmLocationParent(project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing trust configs: %w", err)
	}
	if s.flagCMFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.TrustConfigs, s.flagCMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.TrustConfigs), "%-40s\n", "NAME")
	for _, t := range resp.TrustConfigs {
		fmt.Fprintln(s.inv.stdout, path.Base(t.Name))
	}
	return nil
}

func (s *certificateManagerCmds) runCMTCUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	tc := &certificatemanager.TrustConfig{}
	if err := loadYAMLOrJSONInto(s.flagCMConfigFile, tc); err != nil {
		return err
	}
	mask := s.flagCMUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(tc))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.TrustConfigs.Patch(cmTCName(args[0], project, s.flagCMLocation), tc).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating trust config: %w", err)
	}
	return s.cmFinishOp(ctx, svc, op, "Update trust config", args[0], s.flagCMAsync)
}

// --- operations ---

func (s *certificateManagerCmds) initCertificateManagerAll6() {
	for _, c := range []*cobra.Command{s.cmOpDescribeCmd, s.cmOpListCmd} {
		c.Flags().StringVar(&s.flagCMLocation, "location", "global", "Location containing the operation")
	}
	s.cmOpDescribeCmd.Flags().StringVar(&s.flagCMFormat, "format", "", "Output format")
	s.cmOpListCmd.Flags().StringVar(&s.flagCMFormat, "format", "", "Output format")

	s.cmOperationsCmd.AddCommand(s.cmOpDescribeCmd, s.cmOpListCmd)
	s.certificateManagerCmd.AddCommand(s.cmOperationsCmd)
}

func cmOpName(id, project, location string) string {
	return cmChild("operations", id, cmLocationParent(project, location))
}

func (s *certificateManagerCmds) runCMOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(cmOpName(args[0], project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagCMFormat)
}

func (s *certificateManagerCmds) runCMOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CertificateManager(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(cmLocationParent(project, s.flagCMLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagCMFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagCMFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
