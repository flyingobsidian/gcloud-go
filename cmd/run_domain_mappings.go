package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	runv1 "google.golang.org/api/run/v1"
)

// --- gcloud run domain-mappings (#1049) ---
//
// Backed by the v1 Knative-style API surface
// (Namespaces.Domainmappings). Namespace = the project id. All calls are
// pinned to the regional endpoint https://REGION-run.googleapis.com/.

func (s *runCmds) initRunDomainMappings() {
	all := []*cobra.Command{
		s.runDomainMappingsCreateCmd, s.runDomainMappingsDeleteCmd,
		s.runDomainMappingsDescribeCmd, s.runDomainMappingsListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagRunDomainMappingsRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&s.flagRunDomainMappingsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
	}
	s.runDomainMappingsCreateCmd.Flags().StringVar(&s.flagRunDomainMappingsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the DomainMapping body (required)")
	_ = s.runDomainMappingsCreateCmd.MarkFlagRequired("config-file")
	s.runDomainMappingsListCmd.Flags().Int64Var(&s.flagRunDomainMappingsLimit, "limit", 0,
		"Maximum number of mappings to return (0 = server default)")

	s.runDomainMappingsCmd.AddCommand(all...)
	s.runCmd.AddCommand(s.runDomainMappingsCmd)
}

func runDomainMappingsParent(project string) string {
	return fmt.Sprintf("namespaces/%s", project)
}

func runDomainMappingsName(project, domain string) string {
	return runNamespaceName(project, "domainmappings", domain)
}

func (s *runCmds) runDomainMappingsCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &runv1.DomainMapping{}
	if err := loadYAMLOrJSONInto(s.flagRunDomainMappingsConfigFile, body); err != nil {
		return err
	}
	if body.Metadata == nil {
		body.Metadata = &runv1.ObjectMeta{}
	}
	body.Metadata.Name = args[0]
	if body.Metadata.Namespace == "" {
		body.Metadata.Namespace = project
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV1(ctx, s.flagRunDomainMappingsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Namespaces.Domainmappings.Create(runDomainMappingsParent(project), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating domain mapping: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created Cloud Run domain mapping [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagRunDomainMappingsFormat)
}

func (s *runCmds) runDomainMappingsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV1(ctx, s.flagRunDomainMappingsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Namespaces.Domainmappings.Delete(runDomainMappingsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting domain mapping: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted Cloud Run domain mapping [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagRunDomainMappingsFormat)
}

func (s *runCmds) runDomainMappingsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV1(ctx, s.flagRunDomainMappingsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Namespaces.Domainmappings.Get(runDomainMappingsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain mapping: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRunDomainMappingsFormat)
}

func (s *runCmds) runDomainMappingsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV1(ctx, s.flagRunDomainMappingsRegion)
	if err != nil {
		return err
	}
	// The Knative-style list surface does not paginate; it returns a
	// single page whose size is bounded server-side (or by --limit when
	// provided).
	call := svc.Namespaces.Domainmappings.List(runDomainMappingsParent(project)).Context(ctx)
	if s.flagRunDomainMappingsLimit > 0 {
		call = call.Limit(s.flagRunDomainMappingsLimit)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing domain mappings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp.Items, s.flagRunDomainMappingsFormat)
}
