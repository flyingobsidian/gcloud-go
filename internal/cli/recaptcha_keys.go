package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	recaptcha "google.golang.org/api/recaptchaenterprise/v1"
)

// --- gcloud recaptcha keys (#1185) ---

func (s *recaptchaCmds) initRecaptchaKeys() {
	all := []*cobra.Command{
		s.recaptchaKeysAddIpOverrideCmd, s.recaptchaKeysCreateCmd, s.recaptchaKeysDeleteCmd,
		s.recaptchaKeysDescribeCmd, s.recaptchaKeysListCmd, s.recaptchaKeysListIpOverridesCmd,
		s.recaptchaKeysMigrateCmd, s.recaptchaKeysRemoveIpOverrideCmd, s.recaptchaKeysUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagRcpKeyFormat, "format", "", "Output format")
	}

	// create/update shared flags
	for _, c := range []*cobra.Command{s.recaptchaKeysCreateCmd, s.recaptchaKeysUpdateCmd} {
		c.Flags().StringVar(&s.flagRcpKeyConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the full Key body (skips individual flags)")
		c.Flags().StringVar(&s.flagRcpKeyDisplayName, "display-name", "",
			"Human-readable display name")
		c.Flags().StringSliceVar(&s.flagRcpKeyLabels, "labels", nil, "Labels as KEY=VALUE pairs")
		c.Flags().Float64Var(&s.flagRcpKeyTestingScore, "testing-score", 0,
			"If set, all assessments return this score (0 to 1)")
		c.Flags().StringVar(&s.flagRcpKeyIntegration, "integration-type", "",
			"Web integration type: score, checkbox or invisible")
		c.Flags().BoolVar(&s.flagRcpKeyWeb, "web", false, "Configure the key for websites")
		c.Flags().BoolVar(&s.flagRcpKeyAndroid, "android", false, "Configure the key for Android apps")
		c.Flags().BoolVar(&s.flagRcpKeyIOS, "ios", false, "Configure the key for iOS apps")
		c.Flags().BoolVar(&s.flagRcpKeyExpress, "express", false, "Configure the key for Express assessments")
		c.Flags().BoolVar(&s.flagRcpKeyUniversal, "universal", false, "Configure the key for Universal integration")
		c.Flags().BoolVar(&s.flagRcpKeyAllowAllDoms, "allow-all-domains", false,
			"Skip domain enforcement (web)")
		c.Flags().StringSliceVar(&s.flagRcpKeyDomains, "domains", nil,
			"Domains allowed to use the key (web)")
		c.Flags().BoolVar(&s.flagRcpKeyAllowAllBundle, "allow-all-bundle-ids", false,
			"Skip bundle id enforcement (iOS)")
		c.Flags().StringSliceVar(&s.flagRcpKeyBundleIDs, "bundle-ids", nil,
			"iOS bundle ids allowed to use the key")
		c.Flags().BoolVar(&s.flagRcpKeyAllowAllPkgs, "allow-all-package-names", false,
			"Skip package name enforcement (Android)")
		c.Flags().StringSliceVar(&s.flagRcpKeyPackageNames, "package-names", nil,
			"Android package names allowed to use the key")
	}
	s.recaptchaKeysUpdateCmd.Flags().StringVar(&s.flagRcpKeyUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.recaptchaKeysListCmd.Flags().Int64Var(&s.flagRcpKeyPageSize, "page-size", 0,
		"Maximum number of results per page")
	s.recaptchaKeysListIpOverridesCmd.Flags().Int64Var(&s.flagRcpKeyPageSize, "page-size", 0,
		"Maximum number of results per page")

	for _, c := range []*cobra.Command{s.recaptchaKeysAddIpOverrideCmd, s.recaptchaKeysRemoveIpOverrideCmd} {
		c.Flags().StringVar(&s.flagRcpKeyIP, "ip", "", "IP address or CIDR range (required)")
		_ = c.MarkFlagRequired("ip")
		c.Flags().StringVar(&s.flagRcpKeyOverride, "override", "",
			"Override type: allow (required)")
		_ = c.MarkFlagRequired("override")
	}
	s.recaptchaKeysMigrateCmd.Flags().BoolVar(&s.flagRcpKeySkipBilling, "skip-billing-check", false,
		"Skip the billing check")

	s.recaptchaKeysCmd.AddCommand(all...)
	s.recaptchaCmd.AddCommand(s.recaptchaKeysCmd)
}

func (s *recaptchaCmds) rcpKeyName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.rcpProjectParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/keys/%s", parent, id), nil
}

func rcpKeyOverrideEnum(v string) (string, error) {
	switch strings.ToLower(v) {
	case "allow":
		return "ALLOW", nil
	default:
		return "", fmt.Errorf("--override must be allow (got %q)", v)
	}
}

func rcpKeyIntegrationEnum(v string) (string, error) {
	switch strings.ToLower(v) {
	case "":
		return "", nil
	case "score":
		return "SCORE", nil
	case "checkbox":
		return "CHECKBOX", nil
	case "invisible":
		return "INVISIBLE", nil
	default:
		return "", fmt.Errorf("--integration-type must be score, checkbox or invisible (got %q)", v)
	}
}

func (s *recaptchaCmds) rcpKeyLabelsFromFlag() map[string]string {
	if len(s.flagRcpKeyLabels) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range s.flagRcpKeyLabels {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

// rcpKeyBodyFromFlags composes a Key body from the flag values. Callers
// should apply --config-file first (which fully replaces the body) and use
// this only when --config-file is empty.
func (s *recaptchaCmds) rcpKeyBodyFromFlags() (*recaptcha.GoogleCloudRecaptchaenterpriseV1Key, error) {
	key := &recaptcha.GoogleCloudRecaptchaenterpriseV1Key{
		DisplayName: s.flagRcpKeyDisplayName,
		Labels:      s.rcpKeyLabelsFromFlag(),
	}
	if s.flagRcpKeyTestingScore != 0 {
		key.TestingOptions = &recaptcha.GoogleCloudRecaptchaenterpriseV1TestingOptions{
			TestingScore: s.flagRcpKeyTestingScore,
		}
	}
	integration, err := rcpKeyIntegrationEnum(s.flagRcpKeyIntegration)
	if err != nil {
		return nil, err
	}
	platforms := 0
	if s.flagRcpKeyWeb {
		platforms++
		key.WebSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1WebKeySettings{
			AllowAllDomains: s.flagRcpKeyAllowAllDoms,
			AllowedDomains:  s.flagRcpKeyDomains,
			IntegrationType: integration,
		}
	}
	if s.flagRcpKeyAndroid {
		platforms++
		key.AndroidSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1AndroidKeySettings{
			AllowAllPackageNames: s.flagRcpKeyAllowAllPkgs,
			AllowedPackageNames:  s.flagRcpKeyPackageNames,
		}
	}
	if s.flagRcpKeyIOS {
		platforms++
		key.IosSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1IOSKeySettings{
			AllowAllBundleIds: s.flagRcpKeyAllowAllBundle,
			AllowedBundleIds:  s.flagRcpKeyBundleIDs,
		}
	}
	if s.flagRcpKeyExpress {
		platforms++
		key.ExpressSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1ExpressKeySettings{}
	}
	if s.flagRcpKeyUniversal {
		platforms++
		key.UniversalSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1UniversalKeySettings{}
	}
	if platforms > 1 {
		return nil, fmt.Errorf("only one of --web, --android, --ios, --express, --universal may be set")
	}
	return key, nil
}

func (s *recaptchaCmds) rcpKeyLoadBody() (*recaptcha.GoogleCloudRecaptchaenterpriseV1Key, error) {
	if s.flagRcpKeyConfigFile != "" {
		key := &recaptcha.GoogleCloudRecaptchaenterpriseV1Key{}
		if err := loadYAMLOrJSONInto(s.flagRcpKeyConfigFile, key); err != nil {
			return nil, err
		}
		return key, nil
	}
	return s.rcpKeyBodyFromFlags()
}

func (s *recaptchaCmds) runRcpKeyCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.rcpProjectParent()
	if err != nil {
		return err
	}
	body, err := s.rcpKeyLoadBody()
	if err != nil {
		return err
	}
	if body.DisplayName == "" {
		return fmt.Errorf("--display-name is required (or provide it via --config-file)")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpKeyFormat)
}

func (s *recaptchaCmds) runRcpKeyDelete(cmd *cobra.Command, args []string) error {
	name, err := s.rcpKeyName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Keys.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting key: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted key [%s].\n", args[0])
	return nil
}

func (s *recaptchaCmds) runRcpKeyDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.rcpKeyName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpKeyFormat)
}

func (s *recaptchaCmds) runRcpKeyList(cmd *cobra.Command, args []string) error {
	parent, err := s.rcpProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	var all []*recaptcha.GoogleCloudRecaptchaenterpriseV1Key
	pageToken := ""
	for {
		call := svc.Projects.Keys.List(parent).Context(ctx)
		if s.flagRcpKeyPageSize > 0 {
			call = call.PageSize(s.flagRcpKeyPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing keys: %w", err)
		}
		all = append(all, resp.Keys...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagRcpKeyFormat)
}

func (s *recaptchaCmds) runRcpKeyMigrate(cmd *cobra.Command, args []string) error {
	name, err := s.rcpKeyName(args[0])
	if err != nil {
		return err
	}
	body := &recaptcha.GoogleCloudRecaptchaenterpriseV1MigrateKeyRequest{
		SkipBillingCheck: s.flagRcpKeySkipBilling,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.Migrate(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating key: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Migrated key [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpKeyFormat)
}

func (s *recaptchaCmds) runRcpKeyUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.rcpKeyName(args[0])
	if err != nil {
		return err
	}
	body, err := s.rcpKeyLoadBody()
	if err != nil {
		return err
	}
	mask := s.flagRcpKeyUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Keys.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpKeyFormat)
}

func (s *recaptchaCmds) runRcpKeyAddIpOverride(cmd *cobra.Command, args []string) error {
	name, err := s.rcpKeyName(args[0])
	if err != nil {
		return err
	}
	enum, err := rcpKeyOverrideEnum(s.flagRcpKeyOverride)
	if err != nil {
		return err
	}
	body := &recaptcha.GoogleCloudRecaptchaenterpriseV1AddIpOverrideRequest{
		IpOverrideData: &recaptcha.GoogleCloudRecaptchaenterpriseV1IpOverrideData{
			Ip:           s.flagRcpKeyIP,
			OverrideType: enum,
		},
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.AddIpOverride(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IP override: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Added IP override [%s] to key [%s].\n", s.flagRcpKeyIP, args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpKeyFormat)
}

func (s *recaptchaCmds) runRcpKeyRemoveIpOverride(cmd *cobra.Command, args []string) error {
	name, err := s.rcpKeyName(args[0])
	if err != nil {
		return err
	}
	enum, err := rcpKeyOverrideEnum(s.flagRcpKeyOverride)
	if err != nil {
		return err
	}
	body := &recaptcha.GoogleCloudRecaptchaenterpriseV1RemoveIpOverrideRequest{
		IpOverrideData: &recaptcha.GoogleCloudRecaptchaenterpriseV1IpOverrideData{
			Ip:           s.flagRcpKeyIP,
			OverrideType: enum,
		},
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.RemoveIpOverride(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IP override: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Removed IP override [%s] from key [%s].\n", s.flagRcpKeyIP, args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpKeyFormat)
}

func (s *recaptchaCmds) runRcpKeyListIpOverrides(cmd *cobra.Command, args []string) error {
	name, err := s.rcpKeyName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	var all []*recaptcha.GoogleCloudRecaptchaenterpriseV1IpOverrideData
	pageToken := ""
	for {
		call := svc.Projects.Keys.ListIpOverrides(name).Context(ctx)
		if s.flagRcpKeyPageSize > 0 {
			call = call.PageSize(s.flagRcpKeyPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing IP overrides: %w", err)
		}
		all = append(all, resp.IpOverrides...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagRcpKeyFormat)
}
