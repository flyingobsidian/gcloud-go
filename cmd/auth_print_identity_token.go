package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	iamcredentials "google.golang.org/api/iamcredentials/v1"
)

// --- gcloud auth print-identity-token (#1022) ---
//
// Uses IAM Credentials to mint an OIDC-style ID token by impersonating a
// service account. If --impersonate-service-account isn't given (either
// via the flag or the global env), --account is required.

func (s *authCmds) initAuthPrintIdentityToken() {
	s.authPrintIdentityTokenCmd.Flags().StringVar(&s.flagAuthPitImpersonation, "impersonate-service-account", "", "Service account email to impersonate (falls back to --account)")
	s.authPrintIdentityTokenCmd.Flags().StringSliceVar(&s.flagAuthPitAudiences, "audiences", nil, "Audience(s) to embed in the token (required, may repeat)")
	s.authPrintIdentityTokenCmd.Flags().BoolVar(&s.flagAuthPitIncludeEmail, "include-email", true, "Include the impersonated service account email in the token")
	s.authPrintIdentityTokenCmd.Flags().StringVar(&s.flagAuthPitTokenFormat, "token-format", "standard", "Token format: standard or full")
	_ = s.authPrintIdentityTokenCmd.MarkFlagRequired("audiences")

	s.authCmd.AddCommand(s.authPrintIdentityTokenCmd)
}

func (s *authCmds) runAuthPrintIdentityToken(cmd *cobra.Command, args []string) error {
	sa := s.flagAuthPitImpersonation
	if sa == "" {
		sa = s.inv.flags.account
	}
	if sa == "" || !strings.Contains(sa, "@") {
		return fmt.Errorf("service account email required (via --impersonate-service-account or --account)")
	}
	if len(s.flagAuthPitAudiences) == 0 {
		return fmt.Errorf("--audiences is required")
	}
	name := "projects/-/serviceAccounts/" + sa
	audience := strings.Join(s.flagAuthPitAudiences, ",")
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAMCredentials(ctx)
	if err != nil {
		return err
	}
	req := &iamcredentials.GenerateIdTokenRequest{
		Audience:     audience,
		IncludeEmail: s.flagAuthPitIncludeEmail,
	}
	resp, err := svc.Projects.ServiceAccounts.GenerateIdToken(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("generating identity token: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, resp.Token)
	return nil
}
