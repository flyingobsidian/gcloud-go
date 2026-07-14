package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	privateca "google.golang.org/api/privateca/v1"
)

// --- privateca locations ---

var privatecaLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Explore Private CA locations",
}

var pcaLocDescribeCmd = &cobra.Command{
	Use: "describe LOCATION", Short: "Describe a Private CA location",
	Args: cobra.ExactArgs(1), RunE: runPCALocDescribe,
}
var pcaLocListCmd = &cobra.Command{
	Use: "list", Short: "List Private CA locations",
	Args: cobra.NoArgs, RunE: runPCALocList,
}

var flagPCAFormat string

func init() {
	pcaLocDescribeCmd.Flags().StringVar(&flagPCAFormat, "format", "", "Output format")
	pcaLocListCmd.Flags().StringVar(&flagPCAFormat, "format", "", "Output format")
	privatecaLocationsCmd.AddCommand(pcaLocDescribeCmd, pcaLocListCmd)
	privatecaCmd.AddCommand(privatecaLocationsCmd)
}

func runPCALocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(privatecaLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(got, flagPCAFormat)
}

func runPCALocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if flagPCAFormat != "" {
		return emitFormatted(resp.Locations, flagPCAFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

// --- privateca operations ---

var privatecaOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Private CA operations",
}

var (
	pcaOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Private CA operation",
		Args: cobra.ExactArgs(1), RunE: runPCAOpCancel,
	}
	pcaOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Private CA operation",
		Args: cobra.ExactArgs(1), RunE: runPCAOpDelete,
	}
	pcaOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Private CA operation",
		Args: cobra.ExactArgs(1), RunE: runPCAOpDescribe,
	}
	pcaOpListCmd = &cobra.Command{
		Use: "list", Short: "List Private CA operations in a location",
		Args: cobra.NoArgs, RunE: runPCAOpList,
	}
)

var flagPCAOpLocation string

func init() {
	for _, c := range []*cobra.Command{pcaOpCancelCmd, pcaOpDeleteCmd, pcaOpDescribeCmd, pcaOpListCmd} {
		c.Flags().StringVar(&flagPCAOpLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	pcaOpDescribeCmd.Flags().StringVar(&flagPCAFormat, "format", "", "Output format")
	pcaOpListCmd.Flags().StringVar(&flagPCAFormat, "format", "", "Output format")

	privatecaOperationsCmd.AddCommand(pcaOpCancelCmd, pcaOpDeleteCmd, pcaOpDescribeCmd, pcaOpListCmd)
	privatecaCmd.AddCommand(privatecaOperationsCmd)
}

func pcaOpName(id, project, location string) string {
	return pcaResourceName("operations", id, privatecaLocationParent(project, location))
}

func runPCAOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(pcaOpName(args[0], project, flagPCAOpLocation), &privateca.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runPCAOpDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(pcaOpName(args[0], project, flagPCAOpLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runPCAOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(pcaOpName(args[0], project, flagPCAOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagPCAFormat)
}

func runPCAOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(privatecaLocationParent(project, flagPCAOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagPCAFormat != "" {
		return emitFormatted(resp.Operations, flagPCAFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- privateca pools ---

var privatecaPoolsCmd = &cobra.Command{
	Use:   "pools",
	Short: "Manage Private CA pools",
}

var (
	pcaPoolCreateCmd = &cobra.Command{
		Use: "create POOL", Short: "Create a CA pool from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPCAPoolCreate,
	}
	pcaPoolDeleteCmd = &cobra.Command{
		Use: "delete POOL", Short: "Delete a CA pool",
		Args: cobra.ExactArgs(1), RunE: runPCAPoolDelete,
	}
	pcaPoolDescribeCmd = &cobra.Command{
		Use: "describe POOL", Short: "Describe a CA pool",
		Args: cobra.ExactArgs(1), RunE: runPCAPoolDescribe,
	}
	pcaPoolListCmd = &cobra.Command{
		Use: "list", Short: "List CA pools in a location",
		Args: cobra.NoArgs, RunE: runPCAPoolList,
	}
	pcaPoolUpdateCmd = &cobra.Command{
		Use: "update POOL", Short: "Update a CA pool from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPCAPoolUpdate,
	}
	pcaPoolGetCaCertsCmd = &cobra.Command{
		Use: "get-ca-certs POOL", Short: "Print PEM-encoded CA certificates for a pool",
		Args: cobra.ExactArgs(1), RunE: runPCAPoolGetCaCerts,
	}
	pcaPoolGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POOL", Short: "Print the IAM policy for a CA pool",
		Args: cobra.ExactArgs(1), RunE: runPCAPoolGetIam,
	}
	pcaPoolSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POOL POLICY_FILE", Short: "Replace the IAM policy for a CA pool",
		Args: cobra.ExactArgs(2), RunE: runPCAPoolSetIam,
	}
	pcaPoolAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding POOL", Short: "Add an IAM binding to a CA pool",
		Args: cobra.ExactArgs(1), RunE: runPCAPoolAddIam,
	}
	pcaPoolRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding POOL", Short: "Remove an IAM binding from a CA pool",
		Args: cobra.ExactArgs(1), RunE: runPCAPoolRemoveIam,
	}
)

var (
	flagPCAPoolLocation   string
	flagPCAPoolConfigFile string
	flagPCAPoolUpdateMask string
	flagPCAPoolAsync      bool
	flagPCAPoolIamMember  string
	flagPCAPoolIamRole    string
)

func init() {
	all := []*cobra.Command{pcaPoolCreateCmd, pcaPoolDeleteCmd, pcaPoolDescribeCmd, pcaPoolListCmd, pcaPoolUpdateCmd,
		pcaPoolGetCaCertsCmd, pcaPoolGetIamCmd, pcaPoolSetIamCmd, pcaPoolAddIamCmd, pcaPoolRemoveIamCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagPCAPoolLocation, "location", "", "Location containing the CA pool (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{pcaPoolCreateCmd, pcaPoolUpdateCmd} {
		c.Flags().StringVar(&flagPCAPoolConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the CaPool message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	pcaPoolUpdateCmd.Flags().StringVar(&flagPCAPoolUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{pcaPoolCreateCmd, pcaPoolDeleteCmd, pcaPoolUpdateCmd} {
		c.Flags().BoolVar(&flagPCAPoolAsync, "async", false, "Return the long-running operation without waiting")
	}
	pcaAddIAMFlags(pcaPoolAddIamCmd, &flagPCAPoolIamMember, &flagPCAPoolIamRole)
	pcaAddIAMFlags(pcaPoolRemoveIamCmd, &flagPCAPoolIamMember, &flagPCAPoolIamRole)

	privatecaPoolsCmd.AddCommand(all...)
	privatecaCmd.AddCommand(privatecaPoolsCmd)
}

func pcaPoolName(id, project, location string) string {
	return pcaResourceName("caPools", id, privatecaLocationParent(project, location))
}

func runPCAPoolCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	pool := &privateca.CaPool{}
	if err := loadYAMLOrJSONInto(flagPCAPoolConfigFile, pool); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.Create(privatecaLocationParent(project, flagPCAPoolLocation), pool).
		CaPoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating pool: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Create pool", args[0], flagPCAPoolAsync)
}

func runPCAPoolDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.Delete(pcaPoolName(args[0], project, flagPCAPoolLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting pool: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Delete pool", args[0], flagPCAPoolAsync)
}

func runPCAPoolDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	pool, err := svc.Projects.Locations.CaPools.Get(pcaPoolName(args[0], project, flagPCAPoolLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing pool: %w", err)
	}
	return emitFormatted(pool, flagPCAFormat)
}

func runPCAPoolList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CaPools.List(privatecaLocationParent(project, flagPCAPoolLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing pools: %w", err)
	}
	if flagPCAFormat != "" {
		return emitFormatted(resp.CaPools, flagPCAFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "TIER")
	for _, p := range resp.CaPools {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.Tier)
	}
	return nil
}

func runPCAPoolUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	pool := &privateca.CaPool{}
	if err := loadYAMLOrJSONInto(flagPCAPoolConfigFile, pool); err != nil {
		return err
	}
	mask := flagPCAPoolUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(pool))
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CaPools.Patch(pcaPoolName(args[0], project, flagPCAPoolLocation), pool).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating pool: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Update pool", args[0], flagPCAPoolAsync)
}

func runPCAPoolGetCaCerts(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CaPools.FetchCaCerts(pcaPoolName(args[0], project, flagPCAPoolLocation), &privateca.FetchCaCertsRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching CA certs: %w", err)
	}
	for _, chain := range resp.CaCerts {
		for _, pem := range chain.Certificates {
			fmt.Print(pem)
		}
	}
	return nil
}

func runPCAPoolGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CaPools.GetIamPolicy(pcaPoolName(args[0], project, flagPCAPoolLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagPCAFormat)
}

func runPCAPoolSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &privateca.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CaPools.SetIamPolicy(pcaPoolName(args[0], project, flagPCAPoolLocation), &privateca.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runPCAPoolAddIam(cmd *cobra.Command, args []string) error {
	return pcaPoolModifyIam(args[0], func(p *privateca.Policy) {
		pcaAddBinding(p, flagPCAPoolIamRole, flagPCAPoolIamMember)
	})
}

func runPCAPoolRemoveIam(cmd *cobra.Command, args []string) error {
	return pcaPoolModifyIam(args[0], func(p *privateca.Policy) {
		pcaRemoveBinding(p, flagPCAPoolIamRole, flagPCAPoolIamMember)
	})
}

func pcaPoolModifyIam(name string, mutate func(*privateca.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := pcaPoolName(name, project, flagPCAPoolLocation)
	policy, err := svc.Projects.Locations.CaPools.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.CaPools.SetIamPolicy(resource, &privateca.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// --- privateca certificates ---

var privatecaCertificatesCmd = &cobra.Command{
	Use:   "certificates",
	Short: "Manage Private CA certificates",
}

var (
	pcaCertCreateCmd = &cobra.Command{
		Use: "create CERTIFICATE", Short: "Create a certificate from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPCACertCreate,
	}
	pcaCertDescribeCmd = &cobra.Command{
		Use: "describe CERTIFICATE", Short: "Describe a certificate",
		Args: cobra.ExactArgs(1), RunE: runPCACertDescribe,
	}
	pcaCertListCmd = &cobra.Command{
		Use: "list", Short: "List certificates in a CA pool",
		Args: cobra.NoArgs, RunE: runPCACertList,
	}
	pcaCertUpdateCmd = &cobra.Command{
		Use: "update CERTIFICATE", Short: "Update a certificate from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPCACertUpdate,
	}
	pcaCertRevokeCmd = &cobra.Command{
		Use: "revoke CERTIFICATE", Short: "Revoke a certificate",
		Args: cobra.ExactArgs(1), RunE: runPCACertRevoke,
	}
	pcaCertExportCmd = &cobra.Command{
		Use: "export CERTIFICATE", Short: "Print the PEM-encoded certificate to stdout",
		Args: cobra.ExactArgs(1), RunE: runPCACertExport,
	}
)

var (
	flagPCACertLocation     string
	flagPCACertIssuerPool   string
	flagPCACertConfigFile   string
	flagPCACertUpdateMask   string
	flagPCACertRevokeReason string
	flagPCACertOutputFile   string
)

func init() {
	all := []*cobra.Command{pcaCertCreateCmd, pcaCertDescribeCmd, pcaCertListCmd, pcaCertUpdateCmd, pcaCertRevokeCmd, pcaCertExportCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagPCACertLocation, "location", "", "Location containing the certificate (required)")
		c.Flags().StringVar(&flagPCACertIssuerPool, "issuer-pool", "", "CA pool containing the certificate (required)")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("issuer-pool")
	}
	for _, c := range []*cobra.Command{pcaCertCreateCmd, pcaCertUpdateCmd} {
		c.Flags().StringVar(&flagPCACertConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Certificate message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	pcaCertUpdateCmd.Flags().StringVar(&flagPCACertUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	pcaCertRevokeCmd.Flags().StringVar(&flagPCACertRevokeReason, "reason", "REVOCATION_REASON_UNSPECIFIED",
		"Reason for revocation (see the ReasonForRevocation enum)")
	pcaCertExportCmd.Flags().StringVar(&flagPCACertOutputFile, "output-file", "",
		"Write PEM output to this file instead of stdout")

	privatecaCertificatesCmd.AddCommand(all...)
	privatecaCmd.AddCommand(privatecaCertificatesCmd)
}

func pcaCertName(id, project, location, pool string) string {
	parent := fmt.Sprintf("%s/caPools/%s", privatecaLocationParent(project, location), pool)
	return pcaResourceName("certificates", id, parent)
}

func runPCACertCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cert := &privateca.Certificate{}
	if err := loadYAMLOrJSONInto(flagPCACertConfigFile, cert); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("%s/caPools/%s", privatecaLocationParent(project, flagPCACertLocation), flagPCACertIssuerPool)
	got, err := svc.Projects.Locations.CaPools.Certificates.Create(parent, cert).
		CertificateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating certificate: %w", err)
	}
	return emitFormatted(got, "")
}

func runPCACertDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CaPools.Certificates.Get(pcaCertName(args[0], project, flagPCACertLocation, flagPCACertIssuerPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing certificate: %w", err)
	}
	return emitFormatted(got, flagPCAFormat)
}

func runPCACertList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("%s/caPools/%s", privatecaLocationParent(project, flagPCACertLocation), flagPCACertIssuerPool)
	resp, err := svc.Projects.Locations.CaPools.Certificates.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing certificates: %w", err)
	}
	if flagPCAFormat != "" {
		return emitFormatted(resp.Certificates, flagPCAFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "SUBJECT")
	for _, c := range resp.Certificates {
		sub := ""
		if c.CertificateDescription != nil && c.CertificateDescription.SubjectDescription != nil && c.CertificateDescription.SubjectDescription.Subject != nil {
			sub = c.CertificateDescription.SubjectDescription.Subject.CommonName
		}
		fmt.Printf("%-40s %s\n", path.Base(c.Name), sub)
	}
	return nil
}

func runPCACertUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	cert := &privateca.Certificate{}
	if err := loadYAMLOrJSONInto(flagPCACertConfigFile, cert); err != nil {
		return err
	}
	mask := flagPCACertUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cert))
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CaPools.Certificates.Patch(pcaCertName(args[0], project, flagPCACertLocation, flagPCACertIssuerPool), cert).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating certificate: %w", err)
	}
	return emitFormatted(got, "")
}

func runPCACertRevoke(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &privateca.RevokeCertificateRequest{Reason: flagPCACertRevokeReason}
	got, err := svc.Projects.Locations.CaPools.Certificates.Revoke(pcaCertName(args[0], project, flagPCACertLocation, flagPCACertIssuerPool), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("revoking certificate: %w", err)
	}
	return emitFormatted(got, "")
}

func runPCACertExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cert, err := svc.Projects.Locations.CaPools.Certificates.Get(pcaCertName(args[0], project, flagPCACertLocation, flagPCACertIssuerPool)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching certificate: %w", err)
	}
	out := cert.PemCertificate
	for _, c := range cert.PemCertificateChain {
		out += c
	}
	if flagPCACertOutputFile != "" {
		return os.WriteFile(flagPCACertOutputFile, []byte(out), 0644)
	}
	fmt.Print(out)
	return nil
}
