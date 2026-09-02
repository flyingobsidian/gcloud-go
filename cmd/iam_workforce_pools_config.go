package cmd

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

var iamWpCreateCredConfigCmd = &cobra.Command{
	Use:   "create-cred-config AUDIENCE",
	Short: "Create a workforce-pool BYOID credential configuration file",
	Args:  cobra.ExactArgs(1),
	RunE:  runIamWpCreateCredConfig,
}

var iamWpCreateLoginConfigCmd = &cobra.Command{
	Use:   "create-login-config AUDIENCE",
	Short: "Create a workforce-pool sign-in (login) configuration file",
	Args:  cobra.ExactArgs(1),
	RunE:  runIamWpCreateLoginConfig,
}

var (
	flagIamWpCcOutputFile        string
	flagIamWpCcUserProject       string
	flagIamWpCcSubjectTokenType  string
	flagIamWpCcCredSourceFile    string
	flagIamWpCcCredSourceURL     string
	flagIamWpCcCredSourceHeaders map[string]string
	flagIamWpCcCredSourceType    string
	flagIamWpCcCredSourceField   string
	flagIamWpCcExecCommand       string
	flagIamWpCcExecTimeoutMs     int
	flagIamWpCcExecOutputFile    string
	flagIamWpCcServiceAccount    string
	flagIamWpCcSATokenLifetime   int
	flagIamWpCcStsLocation       string
	flagIamWpCcUniverseDomain    string
	flagIamWpCcEnableMtls        bool
)

var (
	flagIamWpLcOutputFile     string
	flagIamWpLcUniverseDomain string
	flagIamWpLcCloudWebDomain string
	flagIamWpLcEnableMtls     bool
	flagIamWpLcActivate       bool
)

func init() {
	// create-cred-config flags
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcOutputFile, "output-file", "", "Output file path (required)")
	_ = iamWpCreateCredConfigCmd.MarkFlagRequired("output-file")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcUserProject, "workforce-pool-user-project", "", "Client project number (required)")
	_ = iamWpCreateCredConfigCmd.MarkFlagRequired("workforce-pool-user-project")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcSubjectTokenType, "subject-token-type", "", "Subject token type (defaults to id_token)")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcCredSourceFile, "credential-source-file", "", "File containing external credential")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcCredSourceURL, "credential-source-url", "", "URL that returns the external credential")
	iamWpCreateCredConfigCmd.Flags().StringToStringVar(&flagIamWpCcCredSourceHeaders, "credential-source-headers", nil, "Headers to send with the credential-source URL")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcCredSourceType, "credential-source-type", "", "Credential source format: json or text")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcCredSourceField, "credential-source-field-name", "", "JSON field containing the token")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcExecCommand, "executable-command", "", "Absolute command to run for the credential")
	iamWpCreateCredConfigCmd.Flags().IntVar(&flagIamWpCcExecTimeoutMs, "executable-timeout-millis", 30000, "Executable timeout in ms")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcExecOutputFile, "executable-output-file", "", "Cache file for the executable output")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcServiceAccount, "service-account", "", "Service account email to impersonate")
	iamWpCreateCredConfigCmd.Flags().IntVar(&flagIamWpCcSATokenLifetime, "service-account-token-lifetime-seconds", 0, "Impersonation token lifetime seconds")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcStsLocation, "sts-location", "", "STS endpoint region (empty or 'global' for global endpoint)")
	iamWpCreateCredConfigCmd.Flags().StringVar(&flagIamWpCcUniverseDomain, "universe-domain", "", "Override universe domain (defaults to googleapis.com)")
	iamWpCreateCredConfigCmd.Flags().BoolVar(&flagIamWpCcEnableMtls, "enable-mtls", false, "Use mTLS for STS endpoints")

	// create-login-config flags
	iamWpCreateLoginConfigCmd.Flags().StringVar(&flagIamWpLcOutputFile, "output-file", "", "Output file path (required)")
	_ = iamWpCreateLoginConfigCmd.MarkFlagRequired("output-file")
	iamWpCreateLoginConfigCmd.Flags().BoolVar(&flagIamWpLcActivate, "activate", false, "Persist auth/login_config_file to the generated file")
	iamWpCreateLoginConfigCmd.Flags().StringVar(&flagIamWpLcUniverseDomain, "universe-domain", "", "Override universe domain (defaults to googleapis.com)")
	iamWpCreateLoginConfigCmd.Flags().StringVar(&flagIamWpLcCloudWebDomain, "universe-cloud-web-domain", "", "Override cloud web domain")
	iamWpCreateLoginConfigCmd.Flags().BoolVar(&flagIamWpLcEnableMtls, "enable-mtls", false, "Use mTLS for STS endpoints")

	iamWorkforcePoolsCmd.AddCommand(iamWpCreateCredConfigCmd, iamWpCreateLoginConfigCmd)
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

func runIamWpCreateCredConfig(cmd *cobra.Command, args []string) error {
	audiencePath, err := normalizeWorkforceAudience(args[0])
	if err != nil {
		return err
	}
	universe := flagIamWpCcUniverseDomain
	if universe == "" {
		universe = "googleapis.com"
	}
	stsBase, err := stsBaseURL(universe, flagIamWpCcStsLocation, flagIamWpCcEnableMtls)
	if err != nil {
		return err
	}

	credSource, tokenType, err := buildWorkforceCredSource()
	if err != nil {
		return err
	}
	subject := flagIamWpCcSubjectTokenType
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
		"workforce_pool_user_project": flagIamWpCcUserProject,
	}
	if flagIamWpCcServiceAccount != "" {
		output["service_account_impersonation_url"] = fmt.Sprintf(
			"%s/v1/projects/-/serviceAccounts/%s:generateAccessToken",
			iamCredentialsBaseURL(universe, flagIamWpCcEnableMtls),
			flagIamWpCcServiceAccount,
		)
		if flagIamWpCcSATokenLifetime > 0 {
			output["service_account_impersonation"] = map[string]any{
				"token_lifetime_seconds": flagIamWpCcSATokenLifetime,
			}
		}
	} else {
		output["token_info_url"] = stsBase + "/v1/introspect"
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(flagIamWpCcOutputFile, data, 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	fmt.Printf("Created credential configuration file [%s].\n", flagIamWpCcOutputFile)
	return nil
}

// buildWorkforceCredSource returns the credential_source object plus an
// implied subject token type (for X.509/mTLS the caller may override).
func buildWorkforceCredSource() (map[string]any, string, error) {
	count := 0
	if flagIamWpCcCredSourceFile != "" {
		count++
	}
	if flagIamWpCcCredSourceURL != "" {
		count++
	}
	if flagIamWpCcExecCommand != "" {
		count++
	}
	if count == 0 {
		return nil, "", fmt.Errorf("specify one of --credential-source-file, --credential-source-url, or --executable-command")
	}
	if count > 1 {
		return nil, "", fmt.Errorf("specify only one credential source type")
	}

	tokenFormat, err := credentialSourceFormat(flagIamWpCcCredSourceType, flagIamWpCcCredSourceField)
	if err != nil {
		return nil, "", err
	}

	switch {
	case flagIamWpCcCredSourceFile != "":
		src := map[string]any{"file": flagIamWpCcCredSourceFile}
		if tokenFormat != nil {
			src["format"] = tokenFormat
		}
		return src, "", nil
	case flagIamWpCcCredSourceURL != "":
		src := map[string]any{"url": flagIamWpCcCredSourceURL}
		if len(flagIamWpCcCredSourceHeaders) > 0 {
			src["headers"] = flagIamWpCcCredSourceHeaders
		}
		if tokenFormat != nil {
			src["format"] = tokenFormat
		}
		return src, "", nil
	case flagIamWpCcExecCommand != "":
		exec := map[string]any{
			"command":        flagIamWpCcExecCommand,
			"timeout_millis": flagIamWpCcExecTimeoutMs,
		}
		if flagIamWpCcExecOutputFile != "" {
			exec["output_file"] = flagIamWpCcExecOutputFile
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

func runIamWpCreateLoginConfig(cmd *cobra.Command, args []string) error {
	audiencePath, err := normalizeWorkforceAudience(args[0])
	if err != nil {
		return err
	}
	universe := flagIamWpLcUniverseDomain
	if universe == "" {
		universe = "googleapis.com"
	}
	cloudWebDomain := flagIamWpLcCloudWebDomain
	if cloudWebDomain == "" {
		cloudWebDomain = "cloud.google"
	}
	stsBase, err := stsBaseURL(universe, "", flagIamWpLcEnableMtls)
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
	if err := os.WriteFile(flagIamWpLcOutputFile, data, 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	fmt.Printf("Created login configuration file [%s].\n", flagIamWpLcOutputFile)
	if flagIamWpLcActivate {
		fmt.Printf("Set auth/login_config_file requested but not persisted (unsupported in gcloud-go).\n")
	}
	return nil
}
