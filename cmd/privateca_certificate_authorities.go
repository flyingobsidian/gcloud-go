package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
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

var privatecaRootsCmd = &cobra.Command{
	Use:   "roots",
	Short: "Manage Private CA root certificate authorities",
}

// --- subordinates ---

var privatecaSubordinatesCmd = &cobra.Command{
	Use:   "subordinates",
	Short: "Manage Private CA subordinate certificate authorities",
}

var (
	flagPCACAPool          string
	flagPCACALocation      string
	flagPCACAConfigFile    string
	flagPCACAUpdateMask    string
	flagPCACAAsync         bool
	flagPCACAIgnoreActive  bool
	flagPCACAPemCert       string
	flagPCACAOutputFile    string
	flagPCACASkipGrace     bool
	flagPCACAIgnoreDeps    bool
)

func init() {
	privatecaRootsCmd.AddCommand(pcaCASubcommands(pcaCATypeRoot)...)
	privatecaSubordinatesCmd.AddCommand(pcaCASubcommands(pcaCATypeSubordinate)...)
	privatecaCmd.AddCommand(privatecaRootsCmd, privatecaSubordinatesCmd)
}

func pcaCASubcommands(caType string) []*cobra.Command {
	create := &cobra.Command{
		Use: "create CA", Short: fmt.Sprintf("Create a %s certificate authority from a --config-file", caTypeLabel(caType)),
		Args: cobra.ExactArgs(1), RunE: pcaCARunCreate(caType),
	}
	del := &cobra.Command{
		Use: "delete CA", Short: fmt.Sprintf("Delete a %s certificate authority", caTypeLabel(caType)),
		Args: cobra.ExactArgs(1), RunE: pcaCARunDelete,
	}
	describe := &cobra.Command{
		Use: "describe CA", Short: fmt.Sprintf("Describe a %s certificate authority", caTypeLabel(caType)),
		Args: cobra.ExactArgs(1), RunE: pcaCARunDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: fmt.Sprintf("List %s certificate authorities in a CA pool", caTypeLabel(caType)),
		Args: cobra.NoArgs, RunE: pcaCARunList(caType),
	}
	update := &cobra.Command{
		Use: "update CA", Short: fmt.Sprintf("Update a %s certificate authority from a --config-file", caTypeLabel(caType)),
		Args: cobra.ExactArgs(1), RunE: pcaCARunUpdate,
	}
	disable := &cobra.Command{
		Use: "disable CA", Short: "Disable a certificate authority",
		Args: cobra.ExactArgs(1), RunE: pcaCARunDisable,
	}
	enable := &cobra.Command{
		Use: "enable CA", Short: "Enable a certificate authority",
		Args: cobra.ExactArgs(1), RunE: pcaCARunEnable,
	}
	undelete := &cobra.Command{
		Use: "undelete CA", Short: "Undelete a scheduled-for-deletion certificate authority",
		Args: cobra.ExactArgs(1), RunE: pcaCARunUndelete,
	}
	all := []*cobra.Command{create, del, describe, list, update, disable, enable, undelete}
	if caType == pcaCATypeSubordinate {
		activate := &cobra.Command{
			Use: "activate CA", Short: "Activate a subordinate CA using a signed CSR PEM",
			Args: cobra.ExactArgs(1), RunE: pcaCARunActivate,
		}
		activate.Flags().StringVar(&flagPCACAPemCert, "pem-ca-certificate", "",
			"PEM-encoded signed CA certificate (required)")
		_ = activate.MarkFlagRequired("pem-ca-certificate")
		getCsr := &cobra.Command{
			Use: "get-csr CA", Short: "Print the PEM-encoded CSR for a subordinate CA",
			Args: cobra.ExactArgs(1), RunE: pcaCARunGetCsr,
		}
		all = append(all, activate, getCsr)
	}
	for _, c := range all {
		c.Flags().StringVar(&flagPCACALocation, "location", "", "Location containing the CA pool (required)")
		c.Flags().StringVar(&flagPCACAPool, "pool", "", "CA pool containing the certificate authority (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("pool")
	}
	for _, c := range []*cobra.Command{create, update} {
		c.Flags().StringVar(&flagPCACAConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the CertificateAuthority message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	update.Flags().StringVar(&flagPCACAUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{create, del, update, disable, enable, undelete} {
		c.Flags().BoolVar(&flagPCACAAsync, "async", false, "Return the long-running operation without waiting")
	}
	del.Flags().BoolVar(&flagPCACASkipGrace, "skip-grace-period", false, "Delete without waiting for the grace period")
	del.Flags().BoolVar(&flagPCACAIgnoreActive, "ignore-active-certificates", false, "Delete even if the CA has active certificates")
	del.Flags().BoolVar(&flagPCACAIgnoreDeps, "ignore-dependent-resources", false, "Delete even if dependent resources exist")
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

func pcaCARunCreate(caType string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := resolveProject()
		if err != nil {
			return err
		}
		ca := &privateca.CertificateAuthority{}
		if err := loadYAMLOrJSONInto(flagPCACAConfigFile, ca); err != nil {
			return err
		}
		if ca.Type == "" {
			ca.Type = caType
		}
		ctx := context.Background()
		svc, err := gcp.PrivateCAService(ctx, flagAccount)
		if err != nil {
			return err
		}
		op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Create(pcaCAParent(project, flagPCACALocation, flagPCACAPool), ca).
			CertificateAuthorityId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating certificate authority: %w", err)
		}
		return pcaFinishOp(ctx, svc, op, "Create CA", args[0], flagPCACAAsync)
	}
}

func pcaCARunDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.CaPools.CertificateAuthorities.Delete(pcaCAName(args[0], project, flagPCACALocation, flagPCACAPool)).Context(ctx)
	if flagPCACAIgnoreActive {
		call = call.IgnoreActiveCertificates(true)
	}
	if flagPCACAIgnoreDeps {
		call = call.IgnoreDependentResources(true)
	}
	if flagPCACASkipGrace {
		call = call.SkipGracePeriod(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting certificate authority: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Delete CA", args[0], flagPCACAAsync)
}

func pcaCARunDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	ca, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Get(pcaCAName(args[0], project, flagPCACALocation, flagPCACAPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing certificate authority: %w", err)
	}
	return emitFormatted(ca, flagPCAFormat)
}

func pcaCARunList(caType string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		project, err := resolveProject()
		if err != nil {
			return err
		}
		ctx := context.Background()
		svc, err := gcp.PrivateCAService(ctx, flagAccount)
		if err != nil {
			return err
		}
		resp, err := svc.Projects.Locations.CaPools.CertificateAuthorities.List(pcaCAParent(project, flagPCACALocation, flagPCACAPool)).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing certificate authorities: %w", err)
		}
		var filtered []*privateca.CertificateAuthority
		for _, ca := range resp.CertificateAuthorities {
			if ca.Type == caType {
				filtered = append(filtered, ca)
			}
		}
		if flagPCAFormat != "" {
			return emitFormatted(filtered, flagPCAFormat)
		}
		fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "TIER")
		for _, ca := range filtered {
			fmt.Printf("%-40s %-15s %s\n", path.Base(ca.Name), ca.State, ca.Tier)
		}
		return nil
	}
}

func pcaCARunUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ca := &privateca.CertificateAuthority{}
	if err := loadYAMLOrJSONInto(flagPCACAConfigFile, ca); err != nil {
		return err
	}
	mask := flagPCACAUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ca))
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Patch(pcaCAName(args[0], project, flagPCACALocation, flagPCACAPool), ca).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating certificate authority: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Update CA", args[0], flagPCACAAsync)
}

func pcaCARunEnable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Enable(pcaCAName(args[0], project, flagPCACALocation, flagPCACAPool), &privateca.EnableCertificateAuthorityRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling certificate authority: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Enable CA", args[0], flagPCACAAsync)
}

func pcaCARunDisable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Disable(pcaCAName(args[0], project, flagPCACALocation, flagPCACAPool), &privateca.DisableCertificateAuthorityRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling certificate authority: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Disable CA", args[0], flagPCACAAsync)
}

func pcaCARunUndelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Undelete(pcaCAName(args[0], project, flagPCACALocation, flagPCACAPool), &privateca.UndeleteCertificateAuthorityRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting certificate authority: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Undelete CA", args[0], flagPCACAAsync)
}

func pcaCARunActivate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &privateca.ActivateCertificateAuthorityRequest{PemCaCertificate: flagPCACAPemCert}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Activate(pcaCAName(args[0], project, flagPCACALocation, flagPCACAPool), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("activating certificate authority: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Activate CA", args[0], flagPCACAAsync)
}

func pcaCARunGetCsr(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CaPools.CertificateAuthorities.Fetch(pcaCAName(args[0], project, flagPCACALocation, flagPCACAPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching CSR: %w", err)
	}
	fmt.Print(resp.PemCsr)
	return nil
}
