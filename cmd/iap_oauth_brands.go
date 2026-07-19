package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap oauth-brands (#1065) ---

var iapBrandsCmd = &cobra.Command{Use: "oauth-brands", Short: "Manage IAP OAuth brands"}

var (
	flagIapBrandsFormat     string
	flagIapBrandsConfigFile string
	flagIapBrandsPageSize   int64
)

var (
	iapBrandsCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an IAP OAuth brand",
		Args: cobra.NoArgs, RunE: runIapBrandsCreate,
	}
	iapBrandsDescribeCmd = &cobra.Command{
		Use: "describe BRAND", Short: "Describe an IAP OAuth brand",
		Args: cobra.ExactArgs(1), RunE: runIapBrandsDescribe,
	}
	iapBrandsListCmd = &cobra.Command{
		Use: "list", Short: "List IAP OAuth brands",
		Args: cobra.NoArgs, RunE: runIapBrandsList,
	}
)

func init() {
	all := []*cobra.Command{iapBrandsCreateCmd, iapBrandsDescribeCmd, iapBrandsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagIapBrandsFormat, "format", "", "Output format")
	}
	iapBrandsCreateCmd.Flags().StringVar(&flagIapBrandsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Brand body (required)")
	_ = iapBrandsCreateCmd.MarkFlagRequired("config-file")
	iapBrandsListCmd.Flags().Int64Var(&flagIapBrandsPageSize, "page-size", 0, "Maximum results per page")

	iapBrandsCmd.AddCommand(all...)
	iapCmd.AddCommand(iapBrandsCmd)
}

func iapBrandsProjectParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s", project), nil
}

func iapBrandsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := iapBrandsProjectParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/brands/%s", parent, id), nil
}

func runIapBrandsCreate(cmd *cobra.Command, args []string) error {
	parent, err := iapBrandsProjectParent()
	if err != nil {
		return err
	}
	body := &iap.Brand{}
	if err := loadYAMLOrJSONInto(flagIapBrandsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating IAP OAuth brand: %w", err)
	}
	fmt.Printf("Created IAP OAuth brand [%s].\n", got.Name)
	return emitFormatted(got, flagIapBrandsFormat)
}

func runIapBrandsDescribe(cmd *cobra.Command, args []string) error {
	name, err := iapBrandsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing IAP OAuth brand: %w", err)
	}
	return emitFormatted(got, flagIapBrandsFormat)
}

func runIapBrandsList(cmd *cobra.Command, args []string) error {
	parent, err := iapBrandsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// The Brands.List endpoint returns the complete collection at once.
	resp, err := svc.Projects.Brands.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing IAP OAuth brands: %w", err)
	}
	return emitFormatted(resp.Brands, flagIapBrandsFormat)
}
