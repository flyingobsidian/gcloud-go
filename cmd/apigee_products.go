package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee products (#1383) ---

var apigeeProductsCmd = &cobra.Command{Use: "products", Short: "Manage Apigee API products"}

var (
	flagApigeePrdOrganization string
	flagApigeePrdFormat       string
	flagApigeePrdConfigFile   string
)

var (
	apigeePrdCreateCmd = &cobra.Command{
		Use: "create RESOURCE", Short: "Create an API product",
		Args: cobra.ExactArgs(1), RunE: runApigeePrdCreate,
	}
	apigeePrdDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an API product",
		Args: cobra.ExactArgs(1), RunE: runApigeePrdDelete,
	}
	apigeePrdDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an API product",
		Args: cobra.ExactArgs(1), RunE: runApigeePrdDescribe,
	}
	apigeePrdListCmd = &cobra.Command{
		Use: "list", Short: "List API products in an organization",
		Args: cobra.NoArgs, RunE: runApigeePrdList,
	}
	apigeePrdUpdateCmd = &cobra.Command{
		Use: "update RESOURCE", Short: "Update an API product",
		Args: cobra.ExactArgs(1), RunE: runApigeePrdUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		apigeePrdCreateCmd, apigeePrdDeleteCmd, apigeePrdDescribeCmd,
		apigeePrdListCmd, apigeePrdUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeePrdOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeePrdFormat, "format", "", "Output format")
	}
	apigeePrdCreateCmd.Flags().StringVar(&flagApigeePrdConfigFile, "config-file", "", "Path to YAML/JSON API product definition (required)")
	_ = apigeePrdCreateCmd.MarkFlagRequired("config-file")
	apigeePrdUpdateCmd.Flags().StringVar(&flagApigeePrdConfigFile, "config-file", "", "Path to YAML/JSON API product definition (required)")
	_ = apigeePrdUpdateCmd.MarkFlagRequired("config-file")

	apigeeProductsCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeProductsCmd)
}

func apigeePrdName(id string) (string, error) {
	return apigeeResource(flagApigeePrdOrganization, "apiproducts", id)
}

func runApigeePrdCreate(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeePrdOrganization)
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1ApiProduct{}
	if err := loadYAMLOrJSONInto(flagApigeePrdConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apiproducts.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating api product: %w", err)
	}
	return emitFormatted(got, flagApigeePrdFormat)
}

func runApigeePrdDelete(cmd *cobra.Command, args []string) error {
	name, err := apigeePrdName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Apiproducts.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting api product: %w", err)
	}
	fmt.Printf("Deleted api product [%s].\n", args[0])
	return nil
}

func runApigeePrdDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeePrdName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apiproducts.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api product: %w", err)
	}
	return emitFormatted(got, flagApigeePrdFormat)
}

func runApigeePrdList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeePrdOrganization)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Apiproducts.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing api products: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1ApiProduct
	all = append(all, resp.ApiProduct...)
	return emitFormatted(all, flagApigeePrdFormat)
}

func runApigeePrdUpdate(cmd *cobra.Command, args []string) error {
	name, err := apigeePrdName(args[0])
	if err != nil {
		return err
	}
	body := &apigee.GoogleCloudApigeeV1ApiProduct{}
	if err := loadYAMLOrJSONInto(flagApigeePrdConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apiproducts.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating api product: %w", err)
	}
	return emitFormatted(got, flagApigeePrdFormat)
}
