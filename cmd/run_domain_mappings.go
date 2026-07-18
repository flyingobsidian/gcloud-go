package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	runv1 "google.golang.org/api/run/v1"
)

// --- gcloud run domain-mappings (#1049) ---
//
// Backed by the v1 Knative-style API surface
// (Namespaces.Domainmappings). Namespace = the project id. All calls are
// pinned to the regional endpoint https://REGION-run.googleapis.com/.

var runDomainMappingsCmd = &cobra.Command{Use: "domain-mappings", Short: "Manage Cloud Run domain mappings"}

var (
	flagRunDomainMappingsRegion     string
	flagRunDomainMappingsFormat     string
	flagRunDomainMappingsConfigFile string
	flagRunDomainMappingsLimit      int64
)

var (
	runDomainMappingsCreateCmd = &cobra.Command{
		Use: "create DOMAIN", Short: "Create a Cloud Run domain mapping",
		Args: cobra.ExactArgs(1), RunE: runDomainMappingsCreate,
	}
	runDomainMappingsDeleteCmd = &cobra.Command{
		Use: "delete DOMAIN", Short: "Delete a Cloud Run domain mapping",
		Args: cobra.ExactArgs(1), RunE: runDomainMappingsDelete,
	}
	runDomainMappingsDescribeCmd = &cobra.Command{
		Use: "describe DOMAIN", Short: "Describe a Cloud Run domain mapping",
		Args: cobra.ExactArgs(1), RunE: runDomainMappingsDescribe,
	}
	runDomainMappingsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run domain mappings",
		Args: cobra.NoArgs, RunE: runDomainMappingsList,
	}
)

func init() {
	all := []*cobra.Command{
		runDomainMappingsCreateCmd, runDomainMappingsDeleteCmd,
		runDomainMappingsDescribeCmd, runDomainMappingsListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagRunDomainMappingsRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&flagRunDomainMappingsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
	}
	runDomainMappingsCreateCmd.Flags().StringVar(&flagRunDomainMappingsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the DomainMapping body (required)")
	_ = runDomainMappingsCreateCmd.MarkFlagRequired("config-file")
	runDomainMappingsListCmd.Flags().Int64Var(&flagRunDomainMappingsLimit, "limit", 0,
		"Maximum number of mappings to return (0 = server default)")

	runDomainMappingsCmd.AddCommand(all...)
	runCmd.AddCommand(runDomainMappingsCmd)
}

func runDomainMappingsParent(project string) string {
	return fmt.Sprintf("namespaces/%s", project)
}

func runDomainMappingsName(project, domain string) string {
	return runNamespaceName(project, "domainmappings", domain)
}

func runDomainMappingsCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &runv1.DomainMapping{}
	if err := loadYAMLOrJSONInto(flagRunDomainMappingsConfigFile, body); err != nil {
		return err
	}
	if body.Metadata == nil {
		body.Metadata = &runv1.ObjectMeta{}
	}
	body.Metadata.Name = args[0]
	if body.Metadata.Namespace == "" {
		body.Metadata.Namespace = project
	}
	ctx := context.Background()
	svc, err := gcp.RunV1Service(ctx, flagAccount, flagRunDomainMappingsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Namespaces.Domainmappings.Create(runDomainMappingsParent(project), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating domain mapping: %w", err)
	}
	fmt.Printf("Created Cloud Run domain mapping [%s].\n", args[0])
	return emitFormatted(got, flagRunDomainMappingsFormat)
}

func runDomainMappingsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV1Service(ctx, flagAccount, flagRunDomainMappingsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Namespaces.Domainmappings.Delete(runDomainMappingsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting domain mapping: %w", err)
	}
	fmt.Printf("Deleted Cloud Run domain mapping [%s].\n", args[0])
	return emitFormatted(got, flagRunDomainMappingsFormat)
}

func runDomainMappingsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV1Service(ctx, flagAccount, flagRunDomainMappingsRegion)
	if err != nil {
		return err
	}
	got, err := svc.Namespaces.Domainmappings.Get(runDomainMappingsName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain mapping: %w", err)
	}
	return emitFormatted(got, flagRunDomainMappingsFormat)
}

func runDomainMappingsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV1Service(ctx, flagAccount, flagRunDomainMappingsRegion)
	if err != nil {
		return err
	}
	// The Knative-style list surface does not paginate; it returns a
	// single page whose size is bounded server-side (or by --limit when
	// provided).
	call := svc.Namespaces.Domainmappings.List(runDomainMappingsParent(project)).Context(ctx)
	if flagRunDomainMappingsLimit > 0 {
		call = call.Limit(flagRunDomainMappingsLimit)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing domain mappings: %w", err)
	}
	return emitFormatted(resp.Items, flagRunDomainMappingsFormat)
}
