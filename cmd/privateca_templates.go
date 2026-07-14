package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	privateca "google.golang.org/api/privateca/v1"
)

var privatecaTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage Private CA certificate templates",
}

var (
	pcaTplCreateCmd = &cobra.Command{
		Use: "create TEMPLATE", Short: "Create a certificate template from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPCATplCreate,
	}
	pcaTplDeleteCmd = &cobra.Command{
		Use: "delete TEMPLATE", Short: "Delete a certificate template",
		Args: cobra.ExactArgs(1), RunE: runPCATplDelete,
	}
	pcaTplDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a certificate template",
		Args: cobra.ExactArgs(1), RunE: runPCATplDescribe,
	}
	pcaTplListCmd = &cobra.Command{
		Use: "list", Short: "List certificate templates in a location",
		Args: cobra.NoArgs, RunE: runPCATplList,
	}
	pcaTplUpdateCmd = &cobra.Command{
		Use: "update TEMPLATE", Short: "Update a certificate template from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runPCATplUpdate,
	}
	pcaTplReplicateCmd = &cobra.Command{
		Use: "replicate TEMPLATE", Short: "Replicate a certificate template to another location",
		Args: cobra.ExactArgs(1), RunE: runPCATplReplicate,
	}
	pcaTplGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TEMPLATE", Short: "Print the IAM policy for a certificate template",
		Args: cobra.ExactArgs(1), RunE: runPCATplGetIam,
	}
	pcaTplSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TEMPLATE POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runPCATplSetIam,
	}
	pcaTplAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TEMPLATE", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runPCATplAddIam,
	}
	pcaTplRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TEMPLATE", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runPCATplRemoveIam,
	}
)

var (
	flagPCATplLocation      string
	flagPCATplConfigFile    string
	flagPCATplUpdateMask    string
	flagPCATplAsync         bool
	flagPCATplReplicateLoc  string
	flagPCATplIamMember     string
	flagPCATplIamRole       string
)

func init() {
	all := []*cobra.Command{pcaTplCreateCmd, pcaTplDeleteCmd, pcaTplDescribeCmd, pcaTplListCmd, pcaTplUpdateCmd,
		pcaTplReplicateCmd, pcaTplGetIamCmd, pcaTplSetIamCmd, pcaTplAddIamCmd, pcaTplRemoveIamCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagPCATplLocation, "location", "", "Location containing the template (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{pcaTplCreateCmd, pcaTplUpdateCmd} {
		c.Flags().StringVar(&flagPCATplConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the CertificateTemplate body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	pcaTplUpdateCmd.Flags().StringVar(&flagPCATplUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{pcaTplCreateCmd, pcaTplDeleteCmd, pcaTplUpdateCmd, pcaTplReplicateCmd} {
		c.Flags().BoolVar(&flagPCATplAsync, "async", false, "Return the long-running operation without waiting")
	}
	pcaTplReplicateCmd.Flags().StringVar(&flagPCATplReplicateLoc, "target-location", "",
		"Destination location for the replicated template (required)")
	_ = pcaTplReplicateCmd.MarkFlagRequired("target-location")
	pcaAddIAMFlags(pcaTplAddIamCmd, &flagPCATplIamMember, &flagPCATplIamRole)
	pcaAddIAMFlags(pcaTplRemoveIamCmd, &flagPCATplIamMember, &flagPCATplIamRole)

	privatecaTemplatesCmd.AddCommand(all...)
	privatecaCmd.AddCommand(privatecaTemplatesCmd)
}

func pcaTplName(id, project, location string) string {
	return pcaResourceName("certificateTemplates", id, privatecaLocationParent(project, location))
}

func runPCATplCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	tpl := &privateca.CertificateTemplate{}
	if err := loadYAMLOrJSONInto(flagPCATplConfigFile, tpl); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateTemplates.Create(privatecaLocationParent(project, flagPCATplLocation), tpl).
		CertificateTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating template: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Create template", args[0], flagPCATplAsync)
}

func runPCATplDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateTemplates.Delete(pcaTplName(args[0], project, flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting template: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Delete template", args[0], flagPCATplAsync)
}

func runPCATplDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CertificateTemplates.Get(pcaTplName(args[0], project, flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing template: %w", err)
	}
	return emitFormatted(got, flagPCAFormat)
}

func runPCATplList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CertificateTemplates.List(privatecaLocationParent(project, flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing templates: %w", err)
	}
	if flagPCAFormat != "" {
		return emitFormatted(resp.CertificateTemplates, flagPCAFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, t := range resp.CertificateTemplates {
		fmt.Println(path.Base(t.Name))
	}
	return nil
}

func runPCATplUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	tpl := &privateca.CertificateTemplate{}
	if err := loadYAMLOrJSONInto(flagPCATplConfigFile, tpl); err != nil {
		return err
	}
	mask := flagPCATplUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(tpl))
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateTemplates.Patch(pcaTplName(args[0], project, flagPCATplLocation), tpl).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating template: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Update template", args[0], flagPCATplAsync)
}

// runPCATplReplicate reads the source template and creates it in
// --target-location under the same ID.
func runPCATplReplicate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	src, err := svc.Projects.Locations.CertificateTemplates.Get(pcaTplName(args[0], project, flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching source template: %w", err)
	}
	// Clear output-only / immutable fields the target rejects.
	src.Name = ""
	src.CreateTime = ""
	src.UpdateTime = ""
	op, err := svc.Projects.Locations.CertificateTemplates.Create(privatecaLocationParent(project, flagPCATplReplicateLoc), src).
		CertificateTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replicating template: %w", err)
	}
	return pcaFinishOp(ctx, svc, op, "Replicate template", args[0], flagPCATplAsync)
}

func runPCATplGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CertificateTemplates.GetIamPolicy(pcaTplName(args[0], project, flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagPCAFormat)
}

func runPCATplSetIam(cmd *cobra.Command, args []string) error {
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
	got, err := svc.Projects.Locations.CertificateTemplates.SetIamPolicy(pcaTplName(args[0], project, flagPCATplLocation), &privateca.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runPCATplAddIam(cmd *cobra.Command, args []string) error {
	return pcaTplModifyIam(args[0], func(p *privateca.Policy) {
		pcaAddBinding(p, flagPCATplIamRole, flagPCATplIamMember)
	})
}

func runPCATplRemoveIam(cmd *cobra.Command, args []string) error {
	return pcaTplModifyIam(args[0], func(p *privateca.Policy) {
		pcaRemoveBinding(p, flagPCATplIamRole, flagPCATplIamMember)
	})
}

func pcaTplModifyIam(name string, mutate func(*privateca.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PrivateCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := pcaTplName(name, project, flagPCATplLocation)
	policy, err := svc.Projects.Locations.CertificateTemplates.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.CertificateTemplates.SetIamPolicy(resource, &privateca.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}
