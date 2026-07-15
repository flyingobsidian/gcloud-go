package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	recaptcha "google.golang.org/api/recaptchaenterprise/v1"
)

// --- gcloud recaptcha keys (#1185) ---

var recaptchaKeysCmd = &cobra.Command{Use: "keys", Short: "Manage reCAPTCHA keys"}

var (
	flagRcpKeyConfigFile     string
	flagRcpKeyDisplayName    string
	flagRcpKeyLabels         []string
	flagRcpKeyTestingScore   float64
	flagRcpKeyIntegration    string
	flagRcpKeyWeb            bool
	flagRcpKeyAndroid        bool
	flagRcpKeyIOS            bool
	flagRcpKeyExpress        bool
	flagRcpKeyAllowAllDoms   bool
	flagRcpKeyDomains        []string
	flagRcpKeyAllowAllBundle bool
	flagRcpKeyBundleIDs      []string
	flagRcpKeyAllowAllPkgs   bool
	flagRcpKeyPackageNames   []string
	flagRcpKeyIP             string
	flagRcpKeyOverride       string
	flagRcpKeySkipBilling    bool
	flagRcpKeyUpdateMask     string
	flagRcpKeyFormat         string
	flagRcpKeyPageSize       int64
)

var (
	recaptchaKeysAddIpOverrideCmd = &cobra.Command{
		Use: "add-ip-override KEY", Short: "Add an IP override to a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: runRcpKeyAddIpOverride,
	}
	recaptchaKeysCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a reCAPTCHA key",
		Args: cobra.NoArgs, RunE: runRcpKeyCreate,
	}
	recaptchaKeysDeleteCmd = &cobra.Command{
		Use: "delete KEY", Short: "Delete a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: runRcpKeyDelete,
	}
	recaptchaKeysDescribeCmd = &cobra.Command{
		Use: "describe KEY", Short: "Describe a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: runRcpKeyDescribe,
	}
	recaptchaKeysListCmd = &cobra.Command{
		Use: "list", Short: "List reCAPTCHA keys",
		Args: cobra.NoArgs, RunE: runRcpKeyList,
	}
	recaptchaKeysListIpOverridesCmd = &cobra.Command{
		Use: "list-ip-overrides KEY", Short: "List IP overrides for a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: runRcpKeyListIpOverrides,
	}
	recaptchaKeysMigrateCmd = &cobra.Command{
		Use: "migrate KEY", Short: "Migrate a legacy reCAPTCHA key to reCAPTCHA Enterprise",
		Args: cobra.ExactArgs(1), RunE: runRcpKeyMigrate,
	}
	recaptchaKeysRemoveIpOverrideCmd = &cobra.Command{
		Use: "remove-ip-override KEY", Short: "Remove an IP override from a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: runRcpKeyRemoveIpOverride,
	}
	recaptchaKeysUpdateCmd = &cobra.Command{
		Use: "update KEY", Short: "Update a reCAPTCHA key",
		Args: cobra.ExactArgs(1), RunE: runRcpKeyUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		recaptchaKeysAddIpOverrideCmd, recaptchaKeysCreateCmd, recaptchaKeysDeleteCmd,
		recaptchaKeysDescribeCmd, recaptchaKeysListCmd, recaptchaKeysListIpOverridesCmd,
		recaptchaKeysMigrateCmd, recaptchaKeysRemoveIpOverrideCmd, recaptchaKeysUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagRcpKeyFormat, "format", "", "Output format")
	}

	// create/update shared flags
	for _, c := range []*cobra.Command{recaptchaKeysCreateCmd, recaptchaKeysUpdateCmd} {
		c.Flags().StringVar(&flagRcpKeyConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the full Key body (skips individual flags)")
		c.Flags().StringVar(&flagRcpKeyDisplayName, "display-name", "",
			"Human-readable display name")
		c.Flags().StringSliceVar(&flagRcpKeyLabels, "labels", nil, "Labels as KEY=VALUE pairs")
		c.Flags().Float64Var(&flagRcpKeyTestingScore, "testing-score", 0,
			"If set, all assessments return this score (0 to 1)")
		c.Flags().StringVar(&flagRcpKeyIntegration, "integration-type", "",
			"Web integration type: score, checkbox or invisible")
		c.Flags().BoolVar(&flagRcpKeyWeb, "web", false, "Configure the key for websites")
		c.Flags().BoolVar(&flagRcpKeyAndroid, "android", false, "Configure the key for Android apps")
		c.Flags().BoolVar(&flagRcpKeyIOS, "ios", false, "Configure the key for iOS apps")
		c.Flags().BoolVar(&flagRcpKeyExpress, "express", false, "Configure the key for Express assessments")
		c.Flags().BoolVar(&flagRcpKeyAllowAllDoms, "allow-all-domains", false,
			"Skip domain enforcement (web)")
		c.Flags().StringSliceVar(&flagRcpKeyDomains, "domains", nil,
			"Domains allowed to use the key (web)")
		c.Flags().BoolVar(&flagRcpKeyAllowAllBundle, "allow-all-bundle-ids", false,
			"Skip bundle id enforcement (iOS)")
		c.Flags().StringSliceVar(&flagRcpKeyBundleIDs, "bundle-ids", nil,
			"iOS bundle ids allowed to use the key")
		c.Flags().BoolVar(&flagRcpKeyAllowAllPkgs, "allow-all-package-names", false,
			"Skip package name enforcement (Android)")
		c.Flags().StringSliceVar(&flagRcpKeyPackageNames, "package-names", nil,
			"Android package names allowed to use the key")
	}
	recaptchaKeysUpdateCmd.Flags().StringVar(&flagRcpKeyUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	recaptchaKeysListCmd.Flags().Int64Var(&flagRcpKeyPageSize, "page-size", 0,
		"Maximum number of results per page")
	recaptchaKeysListIpOverridesCmd.Flags().Int64Var(&flagRcpKeyPageSize, "page-size", 0,
		"Maximum number of results per page")

	for _, c := range []*cobra.Command{recaptchaKeysAddIpOverrideCmd, recaptchaKeysRemoveIpOverrideCmd} {
		c.Flags().StringVar(&flagRcpKeyIP, "ip", "", "IP address or CIDR range (required)")
		_ = c.MarkFlagRequired("ip")
		c.Flags().StringVar(&flagRcpKeyOverride, "override", "",
			"Override type: allow (required)")
		_ = c.MarkFlagRequired("override")
	}
	recaptchaKeysMigrateCmd.Flags().BoolVar(&flagRcpKeySkipBilling, "skip-billing-check", false,
		"Skip the billing check")

	recaptchaKeysCmd.AddCommand(all...)
	recaptchaCmd.AddCommand(recaptchaKeysCmd)
}

func rcpKeyName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := rcpProjectParent()
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

func rcpKeyLabelsFromFlag() map[string]string {
	if len(flagRcpKeyLabels) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range flagRcpKeyLabels {
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
func rcpKeyBodyFromFlags() (*recaptcha.GoogleCloudRecaptchaenterpriseV1Key, error) {
	key := &recaptcha.GoogleCloudRecaptchaenterpriseV1Key{
		DisplayName: flagRcpKeyDisplayName,
		Labels:      rcpKeyLabelsFromFlag(),
	}
	if flagRcpKeyTestingScore != 0 {
		key.TestingOptions = &recaptcha.GoogleCloudRecaptchaenterpriseV1TestingOptions{
			TestingScore: flagRcpKeyTestingScore,
		}
	}
	integration, err := rcpKeyIntegrationEnum(flagRcpKeyIntegration)
	if err != nil {
		return nil, err
	}
	platforms := 0
	if flagRcpKeyWeb {
		platforms++
		key.WebSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1WebKeySettings{
			AllowAllDomains: flagRcpKeyAllowAllDoms,
			AllowedDomains:  flagRcpKeyDomains,
			IntegrationType: integration,
		}
	}
	if flagRcpKeyAndroid {
		platforms++
		key.AndroidSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1AndroidKeySettings{
			AllowAllPackageNames: flagRcpKeyAllowAllPkgs,
			AllowedPackageNames:  flagRcpKeyPackageNames,
		}
	}
	if flagRcpKeyIOS {
		platforms++
		key.IosSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1IOSKeySettings{
			AllowAllBundleIds: flagRcpKeyAllowAllBundle,
			AllowedBundleIds:  flagRcpKeyBundleIDs,
		}
	}
	if flagRcpKeyExpress {
		platforms++
		key.ExpressSettings = &recaptcha.GoogleCloudRecaptchaenterpriseV1ExpressKeySettings{}
	}
	if platforms > 1 {
		return nil, fmt.Errorf("only one of --web, --android, --ios, --express may be set")
	}
	return key, nil
}

func rcpKeyLoadBody() (*recaptcha.GoogleCloudRecaptchaenterpriseV1Key, error) {
	if flagRcpKeyConfigFile != "" {
		key := &recaptcha.GoogleCloudRecaptchaenterpriseV1Key{}
		if err := loadYAMLOrJSONInto(flagRcpKeyConfigFile, key); err != nil {
			return nil, err
		}
		return key, nil
	}
	return rcpKeyBodyFromFlags()
}

func runRcpKeyCreate(cmd *cobra.Command, args []string) error {
	parent, err := rcpProjectParent()
	if err != nil {
		return err
	}
	body, err := rcpKeyLoadBody()
	if err != nil {
		return err
	}
	if body.DisplayName == "" {
		return fmt.Errorf("--display-name is required (or provide it via --config-file)")
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating key: %w", err)
	}
	return emitFormatted(got, flagRcpKeyFormat)
}

func runRcpKeyDelete(cmd *cobra.Command, args []string) error {
	name, err := rcpKeyName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Keys.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting key: %w", err)
	}
	fmt.Printf("Deleted key [%s].\n", args[0])
	return nil
}

func runRcpKeyDescribe(cmd *cobra.Command, args []string) error {
	name, err := rcpKeyName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing key: %w", err)
	}
	return emitFormatted(got, flagRcpKeyFormat)
}

func runRcpKeyList(cmd *cobra.Command, args []string) error {
	parent, err := rcpProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*recaptcha.GoogleCloudRecaptchaenterpriseV1Key
	pageToken := ""
	for {
		call := svc.Projects.Keys.List(parent).Context(ctx)
		if flagRcpKeyPageSize > 0 {
			call = call.PageSize(flagRcpKeyPageSize)
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
	return emitFormatted(all, flagRcpKeyFormat)
}

func runRcpKeyMigrate(cmd *cobra.Command, args []string) error {
	name, err := rcpKeyName(args[0])
	if err != nil {
		return err
	}
	body := &recaptcha.GoogleCloudRecaptchaenterpriseV1MigrateKeyRequest{
		SkipBillingCheck: flagRcpKeySkipBilling,
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.Migrate(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating key: %w", err)
	}
	fmt.Printf("Migrated key [%s].\n", args[0])
	return emitFormatted(got, flagRcpKeyFormat)
}

func runRcpKeyUpdate(cmd *cobra.Command, args []string) error {
	name, err := rcpKeyName(args[0])
	if err != nil {
		return err
	}
	body, err := rcpKeyLoadBody()
	if err != nil {
		return err
	}
	mask := flagRcpKeyUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
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
	return emitFormatted(got, flagRcpKeyFormat)
}

func runRcpKeyAddIpOverride(cmd *cobra.Command, args []string) error {
	name, err := rcpKeyName(args[0])
	if err != nil {
		return err
	}
	enum, err := rcpKeyOverrideEnum(flagRcpKeyOverride)
	if err != nil {
		return err
	}
	body := &recaptcha.GoogleCloudRecaptchaenterpriseV1AddIpOverrideRequest{
		IpOverrideData: &recaptcha.GoogleCloudRecaptchaenterpriseV1IpOverrideData{
			Ip:           flagRcpKeyIP,
			OverrideType: enum,
		},
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.AddIpOverride(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IP override: %w", err)
	}
	fmt.Printf("Added IP override [%s] to key [%s].\n", flagRcpKeyIP, args[0])
	return emitFormatted(got, flagRcpKeyFormat)
}

func runRcpKeyRemoveIpOverride(cmd *cobra.Command, args []string) error {
	name, err := rcpKeyName(args[0])
	if err != nil {
		return err
	}
	enum, err := rcpKeyOverrideEnum(flagRcpKeyOverride)
	if err != nil {
		return err
	}
	body := &recaptcha.GoogleCloudRecaptchaenterpriseV1RemoveIpOverrideRequest{
		IpOverrideData: &recaptcha.GoogleCloudRecaptchaenterpriseV1IpOverrideData{
			Ip:           flagRcpKeyIP,
			OverrideType: enum,
		},
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Keys.RemoveIpOverride(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IP override: %w", err)
	}
	fmt.Printf("Removed IP override [%s] from key [%s].\n", flagRcpKeyIP, args[0])
	return emitFormatted(got, flagRcpKeyFormat)
}

func runRcpKeyListIpOverrides(cmd *cobra.Command, args []string) error {
	name, err := rcpKeyName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*recaptcha.GoogleCloudRecaptchaenterpriseV1IpOverrideData
	pageToken := ""
	for {
		call := svc.Projects.Keys.ListIpOverrides(name).Context(ctx)
		if flagRcpKeyPageSize > 0 {
			call = call.PageSize(flagRcpKeyPageSize)
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
	return emitFormatted(all, flagRcpKeyFormat)
}
