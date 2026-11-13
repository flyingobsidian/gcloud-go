package cli

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	privateca "google.golang.org/api/privateca/v1"
)

func (s *privatecaCmds) initPrivatecaTemplates() {
	all := []*cobra.Command{s.pcaTplCreateCmd, s.pcaTplDeleteCmd, s.pcaTplDescribeCmd, s.pcaTplListCmd, s.pcaTplUpdateCmd,
		s.pcaTplReplicateCmd, s.pcaTplGetIamCmd, s.pcaTplSetIamCmd, s.pcaTplAddIamCmd, s.pcaTplRemoveIamCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagPCATplLocation, "location", "", "Location containing the template (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{s.pcaTplCreateCmd, s.pcaTplUpdateCmd} {
		c.Flags().StringVar(&s.flagPCATplConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the CertificateTemplate body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.pcaTplUpdateCmd.Flags().StringVar(&s.flagPCATplUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.pcaTplCreateCmd, s.pcaTplDeleteCmd, s.pcaTplUpdateCmd, s.pcaTplReplicateCmd} {
		c.Flags().BoolVar(&s.flagPCATplAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.pcaTplReplicateCmd.Flags().StringVar(&s.flagPCATplReplicateLoc, "target-location", "",
		"Destination location for the replicated template (required)")
	_ = s.pcaTplReplicateCmd.MarkFlagRequired("target-location")
	pcaAddIAMFlags(s.pcaTplAddIamCmd, &s.flagPCATplIamMember, &s.flagPCATplIamRole)
	pcaAddIAMFlags(s.pcaTplRemoveIamCmd, &s.flagPCATplIamMember, &s.flagPCATplIamRole)

	s.privatecaTemplatesCmd.AddCommand(all...)
	s.privatecaCmd.AddCommand(s.privatecaTemplatesCmd)
}

func pcaTplName(id, project, location string) string {
	return pcaResourceName("certificateTemplates", id, privatecaLocationParent(project, location))
}

func (s *privatecaCmds) runPCATplCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	tpl := &privateca.CertificateTemplate{}
	if err := loadYAMLOrJSONInto(s.flagPCATplConfigFile, tpl); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateTemplates.Create(privatecaLocationParent(project, s.flagPCATplLocation), tpl).
		CertificateTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating template: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Create template", args[0], s.flagPCATplAsync)
}

func (s *privatecaCmds) runPCATplDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateTemplates.Delete(pcaTplName(args[0], project, s.flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting template: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Delete template", args[0], s.flagPCATplAsync)
}

func (s *privatecaCmds) runPCATplDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CertificateTemplates.Get(pcaTplName(args[0], project, s.flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing template: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagPCAFormat)
}

func (s *privatecaCmds) runPCATplList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.CertificateTemplates.List(privatecaLocationParent(project, s.flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing templates: %w", err)
	}
	if s.flagPCAFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.CertificateTemplates, s.flagPCAFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.CertificateTemplates), "%-40s\n", "NAME")
	for _, t := range resp.CertificateTemplates {
		fmt.Fprintln(s.inv.stdout, path.Base(t.Name))
	}
	return nil
}

func (s *privatecaCmds) runPCATplUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	tpl := &privateca.CertificateTemplate{}
	if err := loadYAMLOrJSONInto(s.flagPCATplConfigFile, tpl); err != nil {
		return err
	}
	mask := s.flagPCATplUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(tpl))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.CertificateTemplates.Patch(pcaTplName(args[0], project, s.flagPCATplLocation), tpl).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating template: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Update template", args[0], s.flagPCATplAsync)
}

// runPCATplReplicate reads the source template and creates it in
// --target-location under the same ID.
func (s *privatecaCmds) runPCATplReplicate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	src, err := svc.Projects.Locations.CertificateTemplates.Get(pcaTplName(args[0], project, s.flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching source template: %w", err)
	}
	// Clear output-only / immutable fields the target rejects.
	src.Name = ""
	src.CreateTime = ""
	src.UpdateTime = ""
	op, err := svc.Projects.Locations.CertificateTemplates.Create(privatecaLocationParent(project, s.flagPCATplReplicateLoc), src).
		CertificateTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replicating template: %w", err)
	}
	return s.pcaFinishOp(ctx, svc, op, "Replicate template", args[0], s.flagPCATplAsync)
}

func (s *privatecaCmds) runPCATplGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.CertificateTemplates.GetIamPolicy(pcaTplName(args[0], project, s.flagPCATplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagPCAFormat)
}

func (s *privatecaCmds) runPCATplSetIam(cmd *cobra.Command, args []string) error {
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
	got, err := svc.Projects.Locations.CertificateTemplates.SetIamPolicy(pcaTplName(args[0], project, s.flagPCATplLocation), &privateca.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *privatecaCmds) runPCATplAddIam(cmd *cobra.Command, args []string) error {
	return s.pcaTplModifyIam(args[0], func(p *privateca.Policy) {
		pcaAddBinding(p, s.flagPCATplIamRole, s.flagPCATplIamMember)
	})
}

func (s *privatecaCmds) runPCATplRemoveIam(cmd *cobra.Command, args []string) error {
	return s.pcaTplModifyIam(args[0], func(p *privateca.Policy) {
		pcaRemoveBinding(p, s.flagPCATplIamRole, s.flagPCATplIamMember)
	})
}

func (s *privatecaCmds) pcaTplModifyIam(name string, mutate func(*privateca.Policy)) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().PrivateCA(ctx)
	if err != nil {
		return err
	}
	resource := pcaTplName(name, project, s.flagPCATplLocation)
	policy, err := svc.Projects.Locations.CertificateTemplates.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.CertificateTemplates.SetIamPolicy(resource, &privateca.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}
