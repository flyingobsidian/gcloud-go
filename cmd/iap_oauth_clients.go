package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap oauth-clients (#1066) ---

var iapClientsCmd = &cobra.Command{Use: "oauth-clients", Short: "Manage IAP OAuth clients"}

var (
	flagIapClientsBrand       string
	flagIapClientsFormat      string
	flagIapClientsConfigFile  string
	flagIapClientsDisplayName string
	flagIapClientsPageSize    int64
)

var (
	iapClientsCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an IAP OAuth client",
		Args: cobra.NoArgs, RunE: runIapClientsCreate,
	}
	iapClientsDeleteCmd = &cobra.Command{
		Use: "delete CLIENT", Short: "Delete an IAP OAuth client",
		Args: cobra.ExactArgs(1), RunE: runIapClientsDelete,
	}
	iapClientsDescribeCmd = &cobra.Command{
		Use: "describe CLIENT", Short: "Describe an IAP OAuth client",
		Args: cobra.ExactArgs(1), RunE: runIapClientsDescribe,
	}
	iapClientsListCmd = &cobra.Command{
		Use: "list", Short: "List IAP OAuth clients under a brand",
		Args: cobra.NoArgs, RunE: runIapClientsList,
	}
	iapClientsResetSecretCmd = &cobra.Command{
		Use: "reset-secret CLIENT", Short: "Reset the secret of an IAP OAuth client",
		Args: cobra.ExactArgs(1), RunE: runIapClientsResetSecret,
	}
)

func init() {
	all := []*cobra.Command{
		iapClientsCreateCmd, iapClientsDeleteCmd, iapClientsDescribeCmd,
		iapClientsListCmd, iapClientsResetSecretCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagIapClientsBrand, "brand", "",
			"Brand id (or the full projects/PROJECT/brands/BRAND path) that owns the client (required)")
		_ = c.MarkFlagRequired("brand")
		c.Flags().StringVar(&flagIapClientsFormat, "format", "", "Output format")
	}
	iapClientsCreateCmd.Flags().StringVar(&flagIapClientsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the IdentityAwareProxyClient body")
	iapClientsCreateCmd.Flags().StringVar(&flagIapClientsDisplayName, "display-name", "",
		"Display name for the new client (used when --config-file is not set)")
	iapClientsListCmd.Flags().Int64Var(&flagIapClientsPageSize, "page-size", 0, "Maximum results per page")

	iapClientsCmd.AddCommand(all...)
	iapCmd.AddCommand(iapClientsCmd)
}

func iapClientsBrandParent() (string, error) {
	if strings.HasPrefix(flagIapClientsBrand, "projects/") {
		return flagIapClientsBrand, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/brands/%s", project, flagIapClientsBrand), nil
}

func iapClientsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := iapClientsBrandParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/identityAwareProxyClients/%s", parent, id), nil
}

func runIapClientsCreate(cmd *cobra.Command, args []string) error {
	parent, err := iapClientsBrandParent()
	if err != nil {
		return err
	}
	body := &iap.IdentityAwareProxyClient{}
	if flagIapClientsConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagIapClientsConfigFile, body); err != nil {
			return err
		}
	}
	if flagIapClientsDisplayName != "" {
		body.DisplayName = flagIapClientsDisplayName
	}
	if body.DisplayName == "" {
		return fmt.Errorf("either --config-file or --display-name is required")
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.IdentityAwareProxyClients.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating IAP OAuth client: %w", err)
	}
	fmt.Printf("Created IAP OAuth client [%s].\n", got.Name)
	return emitFormatted(got, flagIapClientsFormat)
}

func runIapClientsDelete(cmd *cobra.Command, args []string) error {
	name, err := iapClientsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Brands.IdentityAwareProxyClients.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting IAP OAuth client: %w", err)
	}
	fmt.Printf("Deleted IAP OAuth client [%s].\n", args[0])
	return nil
}

func runIapClientsDescribe(cmd *cobra.Command, args []string) error {
	name, err := iapClientsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.IdentityAwareProxyClients.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing IAP OAuth client: %w", err)
	}
	return emitFormatted(got, flagIapClientsFormat)
}

func runIapClientsList(cmd *cobra.Command, args []string) error {
	parent, err := iapClientsBrandParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*iap.IdentityAwareProxyClient
	pageToken := ""
	for {
		call := svc.Projects.Brands.IdentityAwareProxyClients.List(parent).Context(ctx)
		if flagIapClientsPageSize > 0 {
			call = call.PageSize(flagIapClientsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing IAP OAuth clients: %w", err)
		}
		all = append(all, resp.IdentityAwareProxyClients...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagIapClientsFormat)
}

func runIapClientsResetSecret(cmd *cobra.Command, args []string) error {
	name, err := iapClientsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Brands.IdentityAwareProxyClients.ResetSecret(name,
		&iap.ResetIdentityAwareProxyClientSecretRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting IAP OAuth client secret: %w", err)
	}
	fmt.Printf("Reset secret for IAP OAuth client [%s].\n", args[0])
	return emitFormatted(got, flagIapClientsFormat)
}
