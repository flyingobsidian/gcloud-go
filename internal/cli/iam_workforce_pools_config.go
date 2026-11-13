package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud iam workforce-pools create-cred-config / create-login-config (#1780) ---
//
// Two BYOID configuration file generators shared by Workforce Identity
// Federation, matching the gcloud-python surface. Both commands accept either
// the fully-qualified `locations/LOC/workforcePools/POOL/providers/PROV`
// resource form or the short-format `POOL/PROVIDER` shipped in 581.0.0.

func (s *iamCmds) initIamWorkforcePoolsConfig() {
	// create-cred-config flags
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcOutputFile, "output-file", "", "Output file path (required)")
	_ = s.iamWpCreateCredConfigCmd.MarkFlagRequired("output-file")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcUserProject, "workforce-pool-user-project", "", "Client project number (required)")
	_ = s.iamWpCreateCredConfigCmd.MarkFlagRequired("workforce-pool-user-project")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcSubjectTokenType, "subject-token-type", "", "Subject token type (defaults to id_token)")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcCredSourceFile, "credential-source-file", "", "File containing external credential")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcCredSourceURL, "credential-source-url", "", "URL that returns the external credential")
	s.iamWpCreateCredConfigCmd.Flags().StringToStringVar(&s.flagIamWpCcCredSourceHeaders, "credential-source-headers", nil, "Headers to send with the credential-source URL")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcCredSourceType, "credential-source-type", "", "Credential source format: json or text")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcCredSourceField, "credential-source-field-name", "", "JSON field containing the token")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcExecCommand, "executable-command", "", "Absolute command to run for the credential")
	s.iamWpCreateCredConfigCmd.Flags().IntVar(&s.flagIamWpCcExecTimeoutMs, "executable-timeout-millis", 30000, "Executable timeout in ms")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcExecOutputFile, "executable-output-file", "", "Cache file for the executable output")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcServiceAccount, "service-account", "", "Service account email to impersonate")
	s.iamWpCreateCredConfigCmd.Flags().IntVar(&s.flagIamWpCcSATokenLifetime, "service-account-token-lifetime-seconds", 0, "Impersonation token lifetime seconds")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcStsLocation, "sts-location", "", "STS endpoint region (empty or 'global' for global endpoint)")
	s.iamWpCreateCredConfigCmd.Flags().StringVar(&s.flagIamWpCcUniverseDomain, "universe-domain", "", "Override universe domain (defaults to googleapis.com)")
	s.iamWpCreateCredConfigCmd.Flags().BoolVar(&s.flagIamWpCcEnableMtls, "enable-mtls", false, "Use mTLS for STS endpoints")

	// create-login-config flags
	s.iamWpCreateLoginConfigCmd.Flags().StringVar(&s.flagIamWpLcOutputFile, "output-file", "", "Output file path (required)")
	_ = s.iamWpCreateLoginConfigCmd.MarkFlagRequired("output-file")
	s.iamWpCreateLoginConfigCmd.Flags().BoolVar(&s.flagIamWpLcActivate, "activate", false, "Persist auth/login_config_file to the generated file")
	s.iamWpCreateLoginConfigCmd.Flags().StringVar(&s.flagIamWpLcUniverseDomain, "universe-domain", "", "Override universe domain (defaults to googleapis.com)")
	s.iamWpCreateLoginConfigCmd.Flags().StringVar(&s.flagIamWpLcCloudWebDomain, "universe-cloud-web-domain", "", "Override cloud web domain")
	s.iamWpCreateLoginConfigCmd.Flags().BoolVar(&s.flagIamWpLcEnableMtls, "enable-mtls", false, "Use mTLS for STS endpoints")

	s.iamWorkforcePoolsCmd.AddCommand(s.iamWpCreateCredConfigCmd, s.iamWpCreateLoginConfigCmd)
}

// normalizeWorkforceAudience accepts either a fully-qualified workforce-pool
// provider resource name or the short `POOL/PROVIDER` form shipped in 581.0.0
// and returns the resource form used in the audience field.
func normalizeWorkforceAudience(raw string) (string, error) {
	trimmed := strings.TrimPrefix(raw, "//iam.googleapis.com/")
	if strings.HasPrefix(trimmed, "locations/") {
		return trimmed, nil
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return fmt.Sprintf("locations/global/workforcePools/%s/providers/%s", parts[0], parts[1]), nil
	}
	return "", fmt.Errorf("audience must be either locations/LOC/workforcePools/POOL/providers/PROV or POOL/PROVIDER")
}

// stsBaseURL builds the STS base URL, honouring the locational/global split
// and the optional mTLS variant added in 580.0.0.
func stsBaseURL(universeDomain, stsLocation string, enableMtls bool) (string, error) {
	if universeDomain == "" {
		universeDomain = "googleapis.com"
	}
	if enableMtls && stsLocation != "" && stsLocation != "global" {
		return "", fmt.Errorf("mTLS is not supported with locational Security Token Service endpoints")
	}
	if stsLocation == "" || stsLocation == "global" {
		mtls := ""
		if enableMtls {
			mtls = "mtls."
		}
		return fmt.Sprintf("https://sts.%s%s", mtls, universeDomain), nil
	}
	return fmt.Sprintf("https://sts.%s.rep.%s", stsLocation, universeDomain), nil
}

func iamCredentialsBaseURL(universeDomain string, enableMtls bool) string {
	if universeDomain == "" {
		universeDomain = "googleapis.com"
	}
	mtls := ""
	if enableMtls {
		mtls = "mtls."
	}
	return fmt.Sprintf("https://iamcredentials.%s%s", mtls, universeDomain)
}

func (s *iamCmds) runIamWpCreateCredConfig(cmd *cobra.Command, args []string) error {
	audiencePath, err := normalizeWorkforceAudience(args[0])
	if err != nil {
		return err
	}
	universe := s.flagIamWpCcUniverseDomain
	if universe == "" {
		universe = "googleapis.com"
	}
	stsBase, err := stsBaseURL(universe, s.flagIamWpCcStsLocation, s.flagIamWpCcEnableMtls)
	if err != nil {
		return err
	}

	credSource, tokenType, err := s.buildWorkforceCredSource()
	if err != nil {
		return err
	}
	subject := s.flagIamWpCcSubjectTokenType
	if subject == "" {
		if tokenType != "" {
			subject = tokenType
		} else {
			subject = "urn:ietf:params:oauth:token-type:id_token"
		}
	}

	output := map[string]any{
		"universe_domain":             universe,
		"type":                        "external_account",
		"audience":                    "//iam.googleapis.com/" + audiencePath,
		"subject_token_type":          subject,
		"token_url":                   stsBase + "/v1/token",
		"credential_source":           credSource,
		"workforce_pool_user_project": s.flagIamWpCcUserProject,
	}
	if s.flagIamWpCcServiceAccount != "" {
		output["service_account_impersonation_url"] = fmt.Sprintf(
			"%s/v1/projects/-/serviceAccounts/%s:generateAccessToken",
			iamCredentialsBaseURL(universe, s.flagIamWpCcEnableMtls),
			s.flagIamWpCcServiceAccount,
		)
		if s.flagIamWpCcSATokenLifetime > 0 {
			output["service_account_impersonation"] = map[string]any{
				"token_lifetime_seconds": s.flagIamWpCcSATokenLifetime,
			}
		}
	} else {
		output["token_info_url"] = stsBase + "/v1/introspect"
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(s.flagIamWpCcOutputFile, data, 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created credential configuration file [%s].\n", s.flagIamWpCcOutputFile)
	return nil
}

// buildWorkforceCredSource returns the credential_source object plus an
// implied subject token type (for X.509/mTLS the caller may override).
func (s *iamCmds) buildWorkforceCredSource() (map[string]any, string, error) {
	count := 0
	if s.flagIamWpCcCredSourceFile != "" {
		count++
	}
	if s.flagIamWpCcCredSourceURL != "" {
		count++
	}
	if s.flagIamWpCcExecCommand != "" {
		count++
	}
	if count == 0 {
		return nil, "", fmt.Errorf("specify one of --credential-source-file, --credential-source-url, or --executable-command")
	}
	if count > 1 {
		return nil, "", fmt.Errorf("specify only one credential source type")
	}

	tokenFormat, err := credentialSourceFormat(s.flagIamWpCcCredSourceType, s.flagIamWpCcCredSourceField)
	if err != nil {
		return nil, "", err
	}

	switch {
	case s.flagIamWpCcCredSourceFile != "":
		src := map[string]any{"file": s.flagIamWpCcCredSourceFile}
		if tokenFormat != nil {
			src["format"] = tokenFormat
		}
		return src, "", nil
	case s.flagIamWpCcCredSourceURL != "":
		src := map[string]any{"url": s.flagIamWpCcCredSourceURL}
		if len(s.flagIamWpCcCredSourceHeaders) > 0 {
			src["headers"] = s.flagIamWpCcCredSourceHeaders
		}
		if tokenFormat != nil {
			src["format"] = tokenFormat
		}
		return src, "", nil
	case s.flagIamWpCcExecCommand != "":
		exec := map[string]any{
			"command":        s.flagIamWpCcExecCommand,
			"timeout_millis": s.flagIamWpCcExecTimeoutMs,
		}
		if s.flagIamWpCcExecOutputFile != "" {
			exec["output_file"] = s.flagIamWpCcExecOutputFile
		}
		return map[string]any{"executable": exec}, "", nil
	}
	return nil, "", fmt.Errorf("no credential source resolved")
}

func credentialSourceFormat(sourceType, fieldName string) (map[string]any, error) {
	if sourceType == "" {
		return nil, nil
	}
	sourceType = strings.ToLower(sourceType)
	if sourceType != "json" && sourceType != "text" {
		return nil, fmt.Errorf(`--credential-source-type must be either "json" or "text"`)
	}
	f := map[string]any{"type": sourceType}
	if sourceType == "json" {
		if fieldName == "" {
			return nil, fmt.Errorf("--credential-source-field-name required for JSON formatted tokens")
		}
		f["subject_token_field_name"] = fieldName
	}
	return f, nil
}

func (s *iamCmds) runIamWpCreateLoginConfig(cmd *cobra.Command, args []string) error {
	audiencePath, err := normalizeWorkforceAudience(args[0])
	if err != nil {
		return err
	}
	universe := s.flagIamWpLcUniverseDomain
	if universe == "" {
		universe = "googleapis.com"
	}
	cloudWebDomain := s.flagIamWpLcCloudWebDomain
	if cloudWebDomain == "" {
		cloudWebDomain = "cloud.google"
	}
	stsBase, err := stsBaseURL(universe, "", s.flagIamWpLcEnableMtls)
	if err != nil {
		return err
	}

	output := map[string]any{
		"universe_domain":           universe,
		"universe_cloud_web_domain": cloudWebDomain,
		"type":                      "external_account_authorized_user_login_config",
		"audience":                  "//iam.googleapis.com/" + audiencePath,
		"auth_url":                  fmt.Sprintf("https://auth.%s/authorize", cloudWebDomain),
		"token_url":                 stsBase + "/v1/oauthtoken",
		"token_info_url":            stsBase + "/v1/introspect",
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(s.flagIamWpLcOutputFile, data, 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created login configuration file [%s].\n", s.flagIamWpLcOutputFile)
	if s.flagIamWpLcActivate {
		fmt.Fprintf(s.inv.stdout, "Set auth/login_config_file requested but not persisted (unsupported in gcloud-go).\n")
	}
	return nil
}
