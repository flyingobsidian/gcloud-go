package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	privateca "google.golang.org/api/privateca/v1"
)

// Shared implementation for privateca roots + subordinates, both of which are
// filtered views over the CertificateAuthorities API. Root CAs have
// Type=SELF_SIGNED, subordinates have Type=SUBORDINATE.

const (
	pcaCATypeRoot        = "SELF_SIGNED"
	pcaCATypeSubordinate = "SUBORDINATE"
)

// --- roots ---

// --- subordinates ---

func (s *privatecaCmds) initPrivatecaCertificateAuthorities() {
	s.privatecaRootsCmd.AddCommand(s.pcaCASubcommands(pcaCATypeRoot)...)
	s.privatecaSubordinatesCmd.AddCommand(s.pcaCASubcommands(pcaCATypeSubordinate)...)
	s.privatecaCmd.AddCommand(s.privatecaRootsCmd, s.privatecaSubordinatesCmd)
}

func (s *privatecaCmds) pcaCASubcommands(caType string) []*cobra.Command {
	create := &cobra.Command{
		Use: "create CA", Short: fmt.Sprintf("Create a %s certificate authority from a --config-file", caTypeLabel(caType)),
		Args: cobra.ExactArgs(1), RunE: s.pcaCARunCreate(caType),
	}
	del := &cobra.Command{
		Use: "delete CA", Short: fmt.Sprintf("Delete a %s certificate authority", caTypeLabel(caType)),
		Args: cobra.ExactArgs(1), RunE: s.pcaCARunDelete,
	}
	describe := &cobra.Command{
		Use: "describe CA", Short: fmt.Sprintf("Describe a %s certificate authority", caTypeLabel(caType)),
		Args: cobra.ExactArgs(1), RunE: s.pcaCARunDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: fmt.Sprintf("List %s certificate authorities in a CA pool", caTypeLabel(caType)),
		Args: cobra.NoArgs, RunE: s.pcaCARunList(caType),
	}
	update := &cobra.Command{
		Use: "update CA", Short: fmt.Sprintf("Update a %s certificate authority from a --config-file", caTypeLabel(caType)),
		Args: cobra.ExactArgs(1), RunE: s.pcaCARunUpdate,
	}
	disable := &cobra.Command{
		Use: "disable CA", Short: "Disable a certificate authority",
		Args: cobra.ExactArgs(1), RunE: s.pcaCARunDisable,
	}
	enable := &cobra.Command{
		Use: "enable CA", Short: "Enable a certificate authority",
		Args: cobra.ExactArgs(1), RunE: s.pcaCARunEnable,
	}
	undelete := &cobra.Command{
		Use: "undelete CA", Short: "Undelete a scheduled-for-deletion certificate authority",
		Args: cobra.ExactArgs(1), RunE: s.pcaCARunUndelete,
	}
	all := []*cobra.Command{create, del, describe, list, update, disable, enable, undelete}
	if caType == pcaCATypeSubordinate {
		activate := &cobra.Command{
			Use: "activate CA", Short: "Activate a subordinate CA using a signed CSR PEM",
			Args: cobra.ExactArgs(1), RunE: s.pcaCARunActivate,
		}
		activate.Flags().StringVar(&s.flagPCACAPemCert, "pem-ca-certificate", "",
			"PEM-encoded signed CA certificate (required)")
		_ = activate.MarkFlagRequired("pem-ca-certificate")
		// First-party activation shortcut (gcloud-python 580.0.0): supply the
		// issuing CA by resource id instead of assembling `SubordinateConfig`
		// in a --config-file. --issuer-ca is enough on its own — location and
		// pool default to the subordinate's own --location/--pool.
		activate.Flags().StringVar(&s.flagPCACAIssuerCA, "issuer-ca", "",
			"Issuing CA id (or full resource name) for first-party activation")
		activate.Flags().StringVar(&s.flagPCACAIssuerPool, "issuer-pool", "",
			"CA pool containing --issuer-ca (defaults to --pool)")
		activate.Flags().StringVar(&s.flagPCACAIssuerLocation, "issuer-location", "",
			"Location containing --issuer-pool (defaults to --location)")
		getCsr := &cobra.Command{
			Use: "get-csr CA", Short: "Print the PEM-encoded CSR for a subordinate CA",
			Args: cobra.ExactArgs(1), RunE: s.pcaCARunGetCsr,
		}
		all = append(all, activate, getCsr)
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagPCACALocation, "location", "", "Location containing the CA pool (required)")
		c.Flags().StringVar(&s.flagPCACAPool, "pool", "", "CA pool containing the certificate authority (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("pool")
	}
	for _, c := range []*cobra.Command{create, update} {
		c.Flags().StringVar(&s.flagPCACAConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the CertificateAuthority message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	update.Flags().StringVar(&s.flagPCACAUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{create, del, update, disable, enable, undelete} {
		c.Flags().BoolVar(&s.flagPCACAAsync, "async", false, "Return the long-running operation without waiting")
	}
	del.Flags().BoolVar(&s.flagPCACASkipGrace, "skip-grace-period", false, "Delete without waiting for the grace period")
	del.Flags().BoolVar(&s.flagPCACAIgnoreActive, "ignore-active-certificates", false, "Delete even if the CA has active certificates")
	del.Flags().BoolVar(&s.flagPCACAIgnoreDeps, "ignore-dependent-resources", false, "Delete even if dependent resources exist")
	return all
}

func caTypeLabel(caType string) string {
	if caType == pcaCATypeRoot {
		return "root"
	}
	return "subordinate"
}

func pcaCAParent(project, location, pool string) string {
	return fmt.Sprintf("%s/caPools/%s", privatecaLocationParent(project, location), pool)
}

func pcaCAName(id, project, location, pool string) string {
	return pcaResourceName("certificateAuthorities", id, pcaCAParent(project, location, pool))
}

func (s *privatecaCmds) pcaCARunCreate(caType string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := s.inv.project()
		if err != nil {
			return err
		}
		ca := &privateca.CertificateAuthority{}
		if err := loadYAMLOrJSONInto(s.flagPCACAConfigFile, ca); err != nil {
			return err
		}
		if ca.Type == "" {
			ca.Type = caType
		}
		ctx := cmd.Context()
		svc, err := s.inv.clients().PrivateCA(ctx)
		if err != nil {
			return err
		}
		op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Create(pcaCAParent(project, s.flagPCACALocation, s.flagPCACAPool), ca).
			CertificateAuthorityId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating certificate authority: %w", err)
		}
		return s.pcaFinishOp(ctx, svc, op, "Create CA", args[0], s.flagPCACAAsync)
	}
}

func (s *privatecaCmds) pcaCARunDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.CaPools.CertificateAuthorities.Delete(pcaCAName(args[0], project, s.flagPCACALocation, s.flagPCACAPool)).Context(ctx)
	if s.flagPCACAIgnoreActive {
		call = call.IgnoreActiveCertificates(true)
	}
	if s.flagPCACAIgnoreDeps {
		call = call.IgnoreDependentResources(true)
	}
	if s.flagPCACASkipGrace {
		call = call.SkipGracePeriod(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting certificate authority: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Delete CA", args[0], s.flagPCACAAsync)
}

func (s *privatecaCmds) pcaCARunDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	ca, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Get(pcaCAName(args[0], project, s.flagPCACALocation, s.flagPCACAPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing certificate authority: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, ca, s.flagPCAFormat)
}

func (s *privatecaCmds) pcaCARunList(caType string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := s.inv.project()
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		svc, err := s.inv.clients().PrivateCA(ctx)
		if err != nil {
			return err
		}
		resp, err := svc.Projects.Locations.CaPools.CertificateAuthorities.List(pcaCAParent(project, s.flagPCACALocation, s.flagPCACAPool)).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing certificate authorities: %w", err)
		}
		var filtered []*privateca.CertificateAuthority
		for _, ca := range resp.CertificateAuthorities {
			if ca.Type == caType {
				filtered = append(filtered, ca)
			}
		}
		if s.flagPCAFormat != "" {
			return emitFormattedTo(s.inv.stdout, filtered, s.flagPCAFormat)
		}
		printTableHeadingTo(s.inv.stdout, len(filtered), "%-40s %-15s %s\n", "NAME", "STATE", "TIER")
		for _, ca := range filtered {
			fmt.Fprintf(s.inv.stdout, "%-40s %-15s %s\n", path.Base(ca.Name), ca.State, ca.Tier)
		}
		return nil
	}
}

func (s *privatecaCmds) pcaCARunUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ca := &privateca.CertificateAuthority{}
	if err := loadYAMLOrJSONInto(s.flagPCACAConfigFile, ca); err != nil {
		return err
	}
	mask := s.flagPCACAUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ca))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Patch(pcaCAName(args[0], project, s.flagPCACALocation, s.flagPCACAPool), ca).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating certificate authority: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Update CA", args[0], s.flagPCACAAsync)
}

func (s *privatecaCmds) pcaCARunEnable(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Enable(pcaCAName(args[0], project, s.flagPCACALocation, s.flagPCACAPool), &privateca.EnableCertificateAuthorityRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling certificate authority: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Enable CA", args[0], s.flagPCACAAsync)
}

func (s *privatecaCmds) pcaCARunDisable(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Disable(pcaCAName(args[0], project, s.flagPCACALocation, s.flagPCACAPool), &privateca.DisableCertificateAuthorityRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling certificate authority: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Disable CA", args[0], s.flagPCACAAsync)
}

func (s *privatecaCmds) pcaCARunUndelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Undelete(pcaCAName(args[0], project, s.flagPCACALocation, s.flagPCACAPool), &privateca.UndeleteCertificateAuthorityRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting certificate authority: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Undelete CA", args[0], s.flagPCACAAsync)
}

func (s *privatecaCmds) pcaCARunActivate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &privateca.ActivateCertificateAuthorityRequest{PemCaCertificate: s.flagPCACAPemCert}
	if s.flagPCACAIssuerCA != "" {
		issuer, err := s.pcaResolveIssuerCA(project)
		if err != nil {
			return err
		}
		req.SubordinateConfig = &privateca.SubordinateConfig{CertificateAuthority: issuer}
	} else if s.flagPCACAIssuerPool != "" || s.flagPCACAIssuerLocation != "" {
		return fmt.Errorf("--issuer-pool/--issuer-location require --issuer-ca")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Activate(pcaCAName(args[0], project, s.flagPCACALocation, s.flagPCACAPool), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("activating certificate authority: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Activate CA", args[0], s.flagPCACAAsync)
}

// pcaResolveIssuerCA turns --issuer-ca (plus --issuer-pool/--issuer-location
// or the current --pool/--location) into the fully qualified
// projects/…/certificateAuthorities/… resource name that
// `SubordinateConfig.CertificateAuthority` expects. A user may already pass a
// fully qualified name, in which case it is returned as-is.
func (s *privatecaCmds) pcaResolveIssuerCA(project string) (string, error) {
	if s.flagPCACAIssuerCA == "" {
		return "", nil
	}
	// Full resource name → passthrough.
	if strings.HasPrefix(s.flagPCACAIssuerCA, "projects/") {
		return s.flagPCACAIssuerCA, nil
	}
	loc := s.flagPCACAIssuerLocation
	if loc == "" {
		loc = s.flagPCACALocation
	}
	pool := s.flagPCACAIssuerPool
	if pool == "" {
		pool = s.flagPCACAPool
	}
	if loc == "" || pool == "" {
		return "", fmt.Errorf("--issuer-ca requires --issuer-location and --issuer-pool (or the surrounding --location/--pool)")
	}
	return pcaCAName(s.flagPCACAIssuerCA, project, loc, pool), nil
}

func (s *privatecaCmds) pcaCARunGetCsr(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Fetch(pcaCAName(args[0], project, s.flagPCACALocation, s.flagPCACAPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching CSR: %w", err)
	}
	fmt.Fprint(s.inv.stdout, resp.PemCsr)
	return nil
}
