package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iamcredentials "google.golang.org/api/iamcredentials/v1"
)

// --- gcloud auth print-identity-token (#1022) ---
//
// Uses IAM Credentials to mint an OIDC-style ID token by impersonating a
// service account. If --impersonate-service-account isn't given (either
// via the flag or the global env), --account is required.

var (
	flagAuthPitAudiences     []string
	flagAuthPitIncludeEmail  bool
	flagAuthPitTokenFormat   string
	flagAuthPitImpersonation string
)

var authPrintIdentityTokenCmd = &cobra.Command{
	Use:   "print-identity-token",
	Short: "Print an OpenID Connect identity token by impersonating a service account",
	Args:  cobra.NoArgs,
	RunE:  runAuthPrintIdentityToken,
}

func init() {
	authPrintIdentityTokenCmd.Flags().StringVar(&flagAuthPitImpersonation, "impersonate-service-account", "", "Service account email to impersonate (falls back to --account)")
	authPrintIdentityTokenCmd.Flags().StringSliceVar(&flagAuthPitAudiences, "audiences", nil, "Audience(s) to embed in the token (required, may repeat)")
	authPrintIdentityTokenCmd.Flags().BoolVar(&flagAuthPitIncludeEmail, "include-email", true, "Include the impersonated service account email in the token")
	authPrintIdentityTokenCmd.Flags().StringVar(&flagAuthPitTokenFormat, "token-format", "standard", "Token format: standard or full")
	_ = authPrintIdentityTokenCmd.MarkFlagRequired("audiences")

	authCmd.AddCommand(authPrintIdentityTokenCmd)
}

func runAuthPrintIdentityToken(cmd *cobra.Command, args []string) error {
	sa := flagAuthPitImpersonation
	if sa == "" {
		sa = flagAccount
	}
	if sa == "" || !strings.Contains(sa, "@") {
		return fmt.Errorf("service account email required (via --impersonate-service-account or --account)")
	}
	if len(flagAuthPitAudiences) == 0 {
		return fmt.Errorf("--audiences is required")
	}
	name := "projects/-/serviceAccounts/" + sa
	audience := strings.Join(flagAuthPitAudiences, ",")
	ctx := context.Background()
	svc, err := gcp.IAMCredentialsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &iamcredentials.GenerateIdTokenRequest{
		Audience:     audience,
		IncludeEmail: flagAuthPitIncludeEmail,
	}
	resp, err := svc.Projects.ServiceAccounts.GenerateIdToken(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("generating identity token: %w", err)
	}
	fmt.Println(resp.Token)
	return nil
}
