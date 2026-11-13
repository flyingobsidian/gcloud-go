package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	privateca "google.golang.org/api/privateca/v1"
)

// --- privateca locations ---

func (s *privatecaCmds) initPrivatecaAll() {
	s.pcaLocDescribeCmd.Flags().StringVar(&s.flagPCAFormat, "format", "", "Output format")
	s.pcaLocListCmd.Flags().StringVar(&s.flagPCAFormat, "format", "", "Output format")
	s.privatecaLocationsCmd.AddCommand(s.pcaLocDescribeCmd, s.pcaLocListCmd)
	s.privatecaCmd.AddCommand(s.privatecaLocationsCmd)
}

func (s *privatecaCmds) runPCALocDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(privatecaLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagPCAFormat)
}

func (s *privatecaCmds) runPCALocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if s.flagPCAFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Locations, s.flagPCAFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Locations), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(s.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- privateca operations ---

func (s *privatecaCmds) initPrivatecaAll2() {
	for _, c := range []*cobra.Command{s.pcaOpCancelCmd, s.pcaOpDeleteCmd, s.pcaOpDescribeCmd, s.pcaOpListCmd} {
		c.Flags().StringVar(&s.flagPCAOpLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	s.pcaOpDescribeCmd.Flags().StringVar(&s.flagPCAFormat, "format", "", "Output format")
	s.pcaOpListCmd.Flags().StringVar(&s.flagPCAFormat, "format", "", "Output format")

	s.privatecaOperationsCmd.AddCommand(s.pcaOpCancelCmd, s.pcaOpDeleteCmd, s.pcaOpDescribeCmd, s.pcaOpListCmd)
	s.privatecaCmd.AddCommand(s.privatecaOperationsCmd)
}

func pcaOpName(id, project, location string) string {
	return pcaResourceName("operations", id, privatecaLocationParent(project, location))
}

func (s *privatecaCmds) runPCAOpCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(pcaOpName(args[0], project, s.flagPCAOpLocation), &privateca.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *privatecaCmds) runPCAOpDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(pcaOpName(args[0], project, s.flagPCAOpLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *privatecaCmds) runPCAOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(pcaOpName(args[0], project, s.flagPCAOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagPCAFormat)
}

func (s *privatecaCmds) runPCAOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(privatecaLocationParent(project, s.flagPCAOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagPCAFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagPCAFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- privateca pools ---

func (s *privatecaCmds) initPrivatecaAll3() {
	all := []*cobra.Command{s.pcaPoolCreateCmd, s.pcaPoolDeleteCmd, s.pcaPoolDescribeCmd, s.pcaPoolListCmd, s.pcaPoolUpdateCmd,
		s.pcaPoolGetCaCertsCmd, s.pcaPoolGetIamCmd, s.pcaPoolSetIamCmd, s.pcaPoolAddIamCmd, s.pcaPoolRemoveIamCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagPCAPoolLocation, "location", "", "Location containing the CA pool (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{s.pcaPoolCreateCmd, s.pcaPoolUpdateCmd} {
		c.Flags().StringVar(&s.flagPCAPoolConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the CaPool message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.pcaPoolUpdateCmd.Flags().StringVar(&s.flagPCAPoolUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.pcaPoolCreateCmd, s.pcaPoolDeleteCmd, s.pcaPoolUpdateCmd} {
		c.Flags().BoolVar(&s.flagPCAPoolAsync, "async", false, "Return the long-running operation without waiting")
	}
	pcaAddIAMFlags(s.pcaPoolAddIamCmd, &s.flagPCAPoolIamMember, &s.flagPCAPoolIamRole)
	pcaAddIAMFlags(s.pcaPoolRemoveIamCmd, &s.flagPCAPoolIamMember, &s.flagPCAPoolIamRole)

	s.privatecaPoolsCmd.AddCommand(all...)
	s.privatecaCmd.AddCommand(s.privatecaPoolsCmd)
}

func pcaPoolName(id, project, location string) string {
	return pcaResourceName("caPools", id, privatecaLocationParent(project, location))
}

func (s *privatecaCmds) runPCAPoolCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	pool := &privateca.CaPool{}
	if err := loadYAMLOrJSONInto(s.flagPCAPoolConfigFile, pool); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.Create(privatecaLocationParent(project, s.flagPCAPoolLocation), pool).
		CaPoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating pool: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Create pool", args[0], s.flagPCAPoolAsync)
}

func (s *privatecaCmds) runPCAPoolDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.Delete(pcaPoolName(args[0], project, s.flagPCAPoolLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting pool: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Delete pool", args[0], s.flagPCAPoolAsync)
}

func (s *privatecaCmds) runPCAPoolDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	pool, err := svc.Projects.Locations.CaPools.Get(pcaPoolName(args[0], project, s.flagPCAPoolLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing pool: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, pool, s.flagPCAFormat)
}

func (s *privatecaCmds) runPCAPoolList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CaPools.List(privatecaLocationParent(project, s.flagPCAPoolLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing pools: %w", err)
	}
	if s.flagPCAFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.CaPools, s.flagPCAFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.CaPools), "%-40s %s\n", "NAME", "TIER")
	for _, p := range resp.CaPools {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.Tier)
	}
	return nil
}

func (s *privatecaCmds) runPCAPoolUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	pool := &privateca.CaPool{}
	if err := loadYAMLOrJSONInto(s.flagPCAPoolConfigFile, pool); err != nil {
		return err
	}
	mask := s.flagPCAPoolUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(pool))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.Patch(pcaPoolName(args[0], project, s.flagPCAPoolLocation), pool).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating pool: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Update pool", args[0], s.flagPCAPoolAsync)
}

func (s *privatecaCmds) runPCAPoolGetCaCerts(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CaPools.FetchCaCerts(pcaPoolName(args[0], project, s.flagPCAPoolLocation), &privateca.FetchCaCertsRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching CA certs: %w", err)
	}
	for _, chain := range resp.CaCerts {
		for _, pem := range chain.Certificates {
			fmt.Fprint(s.inv.stdout, pem)
		}
	}
	return nil
}

func (s *privatecaCmds) runPCAPoolGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CaPools.GetIamPolicy(pcaPoolName(args[0], project, s.flagPCAPoolLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagPCAFormat)
}

func (s *privatecaCmds) runPCAPoolSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &privateca.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CaPools.SetIamPolicy(pcaPoolName(args[0], project, s.flagPCAPoolLocation), &privateca.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *privatecaCmds) runPCAPoolAddIam(cmd *cobra.Command, args []string) error {
	return s.pcaPoolModifyIam(args[0], func(p *privateca.Policy) {
		pcaAddBinding(p, s.flagPCAPoolIamRole, s.flagPCAPoolIamMember)
	})
}

func (s *privatecaCmds) runPCAPoolRemoveIam(cmd *cobra.Command, args []string) error {
	return s.pcaPoolModifyIam(args[0], func(p *privateca.Policy) {
		pcaRemoveBinding(p, s.flagPCAPoolIamRole, s.flagPCAPoolIamMember)
	})
}

func (s *privatecaCmds) pcaPoolModifyIam(name string, mutate func(*privateca.Policy)) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	resource := pcaPoolName(name, project, s.flagPCAPoolLocation)
	policy, err := svc.Projects.Locations.CaPools.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.CaPools.SetIamPolicy(resource, &privateca.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- privateca certificates ---

func (s *privatecaCmds) initPrivatecaAll4() {
	all := []*cobra.Command{s.pcaCertCreateCmd, s.pcaCertDescribeCmd, s.pcaCertListCmd, s.pcaCertUpdateCmd, s.pcaCertRevokeCmd, s.pcaCertExportCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagPCACertLocation, "location", "", "Location containing the certificate (required)")
		c.Flags().StringVar(&s.flagPCACertIssuerPool, "issuer-pool", "", "CA pool containing the certificate (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("issuer-pool")
	}
	for _, c := range []*cobra.Command{s.pcaCertCreateCmd, s.pcaCertUpdateCmd} {
		c.Flags().StringVar(&s.flagPCACertConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Certificate message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.pcaCertUpdateCmd.Flags().StringVar(&s.flagPCACertUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.pcaCertCreateCmd.Flags().StringVar(&s.flagPCACertNotBefore, "requested-not-before-time", "",
		"RFC 3339 timestamp for Certificate.requestedNotBeforeTime; requires "+
			"the issuing CA pool's issuancePolicy to set "+
			"allowRequesterSpecifiedNotBeforeTime=true")
	s.pcaCertCreateCmd.Flags().BoolVar(&s.flagPCACertGenerateKey, "generate-key", false,
		"Use this flag to have a new private key securely generated on your machine.")
	s.pcaCertCreateCmd.Flags().StringVar(&s.flagPCACertKeyOutputFile, "key-output-file", "",
		"The path where the generated private key file should be written (in PEM format).")
	s.pcaCertCreateCmd.Flags().StringVar(&s.flagPCACertKeyAlgorithm, "key-algorithm", "rsa-2048",
		"The crypto algorithm to use for generating the client-side private key. "+
			"KEY_ALGORITHM must be one of: "+strings.Join(pcaKeyAlgorithms(), ", ")+".")
	s.pcaCertRevokeCmd.Flags().StringVar(&s.flagPCACertRevokeReason, "reason", "REVOCATION_REASON_UNSPECIFIED",
		"Reason for revocation (see the ReasonForRevocation enum)")
	s.pcaCertExportCmd.Flags().StringVar(&s.flagPCACertOutputFile, "output-file", "",
		"Write PEM output to this file instead of stdout")

	s.privatecaCertificatesCmd.AddCommand(all...)
	s.privatecaCmd.AddCommand(s.privatecaCertificatesCmd)
}

func pcaCertName(id, project, location, pool string) string {
	parent := fmt.Sprintf("%s/caPools/%s", privatecaLocationParent(project, location), pool)
	return pcaResourceName("certificates", id, parent)
}

// validatePCACertKeyFlags enforces the --generate-key argument group:
// --generate-key and --key-output-file go together, and --key-algorithm
// only applies with --generate-key.
func (s *privatecaCmds) validatePCACertKeyFlags(cmd *cobra.Command) error {
	if err := validatePCAKeyAlgorithm(s.flagPCACertKeyAlgorithm); err != nil {
		return err
	}
	if s.flagPCACertGenerateKey {
		if s.flagPCACertKeyOutputFile == "" {
			return fmt.Errorf("argument --generate-key: --key-output-file must be specified.")
		}
		return nil
	}
	for _, f := range []string{"key-output-file", "key-algorithm"} {
		if cmd.Flags().Changed(f) {
			return fmt.Errorf("argument --%s: --generate-key must be specified.", f)
		}
	}
	return nil
}

func (s *privatecaCmds) runPCACertCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if err := s.validatePCACertKeyFlags(cmd); err != nil {
		return err
	}
	cert := &privateca.Certificate{}
	if err := loadYAMLOrJSONInto(s.flagPCACertConfigFile, cert); err != nil {
		return err
	}
	if s.flagPCACertGenerateKey {
		if err := s.applyPCAGeneratedKey(cert, s.flagPCACertKeyAlgorithm, s.flagPCACertKeyOutputFile); err != nil {
			return err
		}
	}
	// --requested-not-before-time wins over any value from --config-file so
	// the flag has a predictable effect regardless of the body content.
	if s.flagPCACertNotBefore != "" {
		cert.RequestedNotBeforeTime = s.flagPCACertNotBefore
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("%s/caPools/%s", privatecaLocationParent(project, s.flagPCACertLocation), s.flagPCACertIssuerPool)
	got, err := svc.Projects.Locations.CaPools.Certificates.Create(parent, cert).
		CertificateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating certificate: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *privatecaCmds) runPCACertDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CaPools.Certificates.Get(pcaCertName(args[0], project, s.flagPCACertLocation, s.flagPCACertIssuerPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing certificate: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagPCAFormat)
}

func (s *privatecaCmds) runPCACertList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("%s/caPools/%s", privatecaLocationParent(project, s.flagPCACertLocation), s.flagPCACertIssuerPool)
	resp, err := svc.Projects.Locations.CaPools.Certificates.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing certificates: %w", err)
	}
	if s.flagPCAFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Certificates, s.flagPCAFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Certificates), "%-40s %s\n", "NAME", "SUBJECT")
	for _, c := range resp.Certificates {
		sub := ""
		if c.CertificateDescription != nil && c.CertificateDescription.SubjectDescription != nil && c.CertificateDescription.SubjectDescription.Subject != nil {
			sub = c.CertificateDescription.SubjectDescription.Subject.CommonName
		}
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(c.Name), sub)
	}
	return nil
}

func (s *privatecaCmds) runPCACertUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	cert := &privateca.Certificate{}
	if err := loadYAMLOrJSONInto(s.flagPCACertConfigFile, cert); err != nil {
		return err
	}
	mask := s.flagPCACertUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cert))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CaPools.Certificates.Patch(pcaCertName(args[0], project, s.flagPCACertLocation, s.flagPCACertIssuerPool), cert).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating certificate: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *privatecaCmds) runPCACertRevoke(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	req := &privateca.RevokeCertificateRequest{Reason: s.flagPCACertRevokeReason}
	got, err := svc.Projects.Locations.CaPools.Certificates.Revoke(pcaCertName(args[0], project, s.flagPCACertLocation, s.flagPCACertIssuerPool), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("revoking certificate: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *privatecaCmds) runPCACertExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	cert, err := svc.Projects.Locations.CaPools.Certificates.Get(pcaCertName(args[0], project, s.flagPCACertLocation, s.flagPCACertIssuerPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching certificate: %w", err)
	}
	out := cert.PemCertificate
	for _, c := range cert.PemCertificateChain {
		out += c
	}
	if s.flagPCACertOutputFile != "" {
		return os.WriteFile(s.flagPCACertOutputFile, []byte(out), 0644)
	}
	fmt.Fprint(s.inv.stdout, out)
	return nil
}
