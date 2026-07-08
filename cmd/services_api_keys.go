package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apikeys "google.golang.org/api/apikeys/v2"
)

var servicesAPIKeysCmd = &cobra.Command{
	Use:   "api-keys",
	Short: "Manage API keys",
}

var apiKeyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an API key",
	Args:  cobra.NoArgs,
	RunE:  runAPIKeyCreate,
}

var apiKeyDeleteCmd = &cobra.Command{
	Use:   "delete KEY_ID_OR_NAME",
	Short: "Delete an API key",
	Args:  cobra.ExactArgs(1),
	RunE:  runAPIKeyDelete,
}

var apiKeyDescribeCmd = &cobra.Command{
	Use:   "describe KEY_ID_OR_NAME",
	Short: "Describe an API key",
	Args:  cobra.ExactArgs(1),
	RunE:  runAPIKeyDescribe,
}

var apiKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	Args:  cobra.NoArgs,
	RunE:  runAPIKeyList,
}

var apiKeyLookupCmd = &cobra.Command{
	Use:   "lookup KEY_STRING",
	Short: "Look up the resource name of an API key from its key string",
	Args:  cobra.ExactArgs(1),
	RunE:  runAPIKeyLookup,
}

var apiKeyGetKeyStringCmd = &cobra.Command{
	Use:   "get-key-string KEY_ID_OR_NAME",
	Short: "Get the string of an API key",
	Args:  cobra.ExactArgs(1),
	RunE:  runAPIKeyGetKeyString,
}

var apiKeyUpdateCmd = &cobra.Command{
	Use:   "update KEY_ID_OR_NAME",
	Short: "Update an API key",
	Args:  cobra.ExactArgs(1),
	RunE:  runAPIKeyUpdate,
}

var apiKeyUndeleteCmd = &cobra.Command{
	Use:   "undelete KEY_ID_OR_NAME",
	Short: "Undelete an API key deleted within the last 30 days",
	Args:  cobra.ExactArgs(1),
	RunE:  runAPIKeyUndelete,
}

var (
	flagAPIKeyDisplayName     string
	flagAPIKeyAnnotations     map[string]string
	flagAPIKeyAllowedIPs      []string
	flagAPIKeyAllowedReferrers []string
	flagAPIKeyAllowedBundleIDs []string
	flagAPIKeyAllowedAndroidPackages []string
	flagAPIKeyAllowedAndroidSHAs []string
	flagAPIKeyAPITargets      []string
	flagAPIKeyClearRestrictions bool
	flagAPIKeyServiceAccount  string
	flagAPIKeyListShowDeleted bool
	flagAPIKeyListFormat      string
	flagAPIKeyListPageSize    int64
	flagAPIKeyListLimit       int64
)

func init() {
	restrictionFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&flagAPIKeyDisplayName, "display-name", "", "Human-readable display name for the key")
		c.Flags().StringToStringVar(&flagAPIKeyAnnotations, "annotations", nil, "Annotations to attach to the key")
		c.Flags().StringSliceVar(&flagAPIKeyAllowedIPs, "allowed-ips", nil, "IPv4/IPv6 addresses or CIDR ranges permitted to use the key")
		c.Flags().StringSliceVar(&flagAPIKeyAllowedReferrers, "allowed-referrers", nil, "HTTP referrers permitted to use the key")
		c.Flags().StringSliceVar(&flagAPIKeyAllowedBundleIDs, "allowed-bundle-ids", nil, "iOS bundle IDs permitted to use the key")
		c.Flags().StringSliceVar(&flagAPIKeyAllowedAndroidPackages, "allowed-application", nil, "Android app package names permitted to use the key")
		c.Flags().StringSliceVar(&flagAPIKeyAllowedAndroidSHAs, "allowed-application-sha1", nil, "Android app SHA-1 fingerprints, paired with --allowed-application")
		c.Flags().StringSliceVar(&flagAPIKeyAPITargets, "api-target", nil, "API restriction of the form SERVICE_NAME[:METHOD[:METHOD...]]")
		c.Flags().StringVar(&flagAPIKeyServiceAccount, "service-account", "", "Service account email to bind the key to")
	}
	restrictionFlags(apiKeyCreateCmd)
	restrictionFlags(apiKeyUpdateCmd)
	apiKeyUpdateCmd.Flags().BoolVar(&flagAPIKeyClearRestrictions, "clear-restrictions", false, "Clear all restrictions on the key")

	apiKeyListCmd.Flags().BoolVar(&flagAPIKeyListShowDeleted, "show-deleted", false, "Include deleted keys in results")
	apiKeyListCmd.Flags().StringVar(&flagAPIKeyListFormat, "format", "", "Output format (json, yaml, or table)")
	apiKeyListCmd.Flags().Int64Var(&flagAPIKeyListPageSize, "page-size", 0, "Page size for API pagination")
	apiKeyListCmd.Flags().Int64Var(&flagAPIKeyListLimit, "limit", 0, "Maximum number of keys to list (0 = no limit)")

	servicesAPIKeysCmd.AddCommand(
		apiKeyCreateCmd, apiKeyDeleteCmd, apiKeyDescribeCmd, apiKeyListCmd,
		apiKeyLookupCmd, apiKeyGetKeyStringCmd, apiKeyUpdateCmd, apiKeyUndeleteCmd,
	)
	servicesCmd.AddCommand(servicesAPIKeysCmd)
}

// apiKeyName normalizes a user-supplied key identifier to the full resource
// name `projects/{project}/locations/global/keys/{key}`. If the input already
// looks like a full resource name it is returned unchanged.
func apiKeyName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/global/keys/%s", project, id)
}

// apiKeyParent returns `projects/{project}/locations/global`.
func apiKeyParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/global", project)
}

// parseAPITargets converts flag values like "compute.googleapis.com:get:list"
// into V2ApiTarget structs.
func parseAPITargets(entries []string) []*apikeys.V2ApiTarget {
	if len(entries) == 0 {
		return nil
	}
	out := make([]*apikeys.V2ApiTarget, 0, len(entries))
	for _, e := range entries {
		parts := strings.Split(e, ":")
		t := &apikeys.V2ApiTarget{Service: parts[0]}
		if len(parts) > 1 {
			t.Methods = parts[1:]
		}
		out = append(out, t)
	}
	return out
}

// buildAPIKeyRestrictions assembles a V2Restrictions from the current flag
// values, returning nil if no restriction fields were set.
func buildAPIKeyRestrictions() *apikeys.V2Restrictions {
	r := &apikeys.V2Restrictions{}
	set := false
	if len(flagAPIKeyAllowedIPs) > 0 {
		r.ServerKeyRestrictions = &apikeys.V2ServerKeyRestrictions{AllowedIps: flagAPIKeyAllowedIPs}
		set = true
	}
	if len(flagAPIKeyAllowedReferrers) > 0 {
		r.BrowserKeyRestrictions = &apikeys.V2BrowserKeyRestrictions{AllowedReferrers: flagAPIKeyAllowedReferrers}
		set = true
	}
	if len(flagAPIKeyAllowedBundleIDs) > 0 {
		r.IosKeyRestrictions = &apikeys.V2IosKeyRestrictions{AllowedBundleIds: flagAPIKeyAllowedBundleIDs}
		set = true
	}
	if len(flagAPIKeyAllowedAndroidPackages) > 0 {
		apps := make([]*apikeys.V2AndroidApplication, 0, len(flagAPIKeyAllowedAndroidPackages))
		for i, pkg := range flagAPIKeyAllowedAndroidPackages {
			app := &apikeys.V2AndroidApplication{PackageName: pkg}
			if i < len(flagAPIKeyAllowedAndroidSHAs) {
				app.Sha1Fingerprint = flagAPIKeyAllowedAndroidSHAs[i]
			}
			apps = append(apps, app)
		}
		r.AndroidKeyRestrictions = &apikeys.V2AndroidKeyRestrictions{AllowedApplications: apps}
		set = true
	}
	if len(flagAPIKeyAPITargets) > 0 {
		r.ApiTargets = parseAPITargets(flagAPIKeyAPITargets)
		set = true
	}
	if !set {
		return nil
	}
	return r
}

func runAPIKeyCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIKeysService(ctx, flagAccount)
	if err != nil {
		return err
	}
	key := &apikeys.V2Key{
		DisplayName:         flagAPIKeyDisplayName,
		Annotations:         flagAPIKeyAnnotations,
		Restrictions:        buildAPIKeyRestrictions(),
		ServiceAccountEmail: flagAPIKeyServiceAccount,
	}
	op, err := svc.Projects.Locations.Keys.Create(apiKeyParent(project), key).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating API key: %w", err)
	}
	fmt.Printf("Create API key in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runAPIKeyDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIKeysService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Keys.Delete(apiKeyName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting API key: %w", err)
	}
	fmt.Printf("Delete API key in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runAPIKeyDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIKeysService(ctx, flagAccount)
	if err != nil {
		return err
	}
	key, err := svc.Projects.Locations.Keys.Get(apiKeyName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing API key: %w", err)
	}
	return yamlEncode(key)
}

func runAPIKeyList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIKeysService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*apikeys.V2Key
	pageToken := ""
	for {
		call := svc.Projects.Locations.Keys.List(apiKeyParent(project)).ShowDeleted(flagAPIKeyListShowDeleted).Context(ctx)
		if flagAPIKeyListPageSize > 0 {
			call = call.PageSize(flagAPIKeyListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing API keys: %w", err)
		}
		all = append(all, resp.Keys...)
		if flagAPIKeyListLimit > 0 && int64(len(all)) >= flagAPIKeyListLimit {
			all = all[:flagAPIKeyListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return printListResults(all, flagAPIKeyListFormat, func() {
		fmt.Printf("%-40s %-30s %s\n", "NAME", "DISPLAY_NAME", "CREATE_TIME")
		for _, k := range all {
			fmt.Printf("%-40s %-30s %s\n", k.Name, k.DisplayName, k.CreateTime)
		}
	})
}

func runAPIKeyLookup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.APIKeysService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Keys.LookupKey().KeyString(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up API key: %w", err)
	}
	return yamlEncode(resp)
}

func runAPIKeyGetKeyString(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIKeysService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Keys.GetKeyString(apiKeyName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting API key string: %w", err)
	}
	fmt.Println(resp.KeyString)
	return nil
}

func runAPIKeyUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIKeysService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := apiKeyName(project, args[0])
	key := &apikeys.V2Key{Name: name}

	var masks []string
	if cmd.Flags().Changed("display-name") {
		key.DisplayName = flagAPIKeyDisplayName
		masks = append(masks, "display_name")
	}
	if cmd.Flags().Changed("annotations") {
		key.Annotations = flagAPIKeyAnnotations
		masks = append(masks, "annotations")
	}
	if cmd.Flags().Changed("service-account") {
		key.ServiceAccountEmail = flagAPIKeyServiceAccount
		masks = append(masks, "service_account_email")
	}
	if flagAPIKeyClearRestrictions {
		key.Restrictions = &apikeys.V2Restrictions{}
		masks = append(masks, "restrictions")
	} else if r := buildAPIKeyRestrictions(); r != nil {
		key.Restrictions = r
		masks = append(masks, "restrictions")
	}
	if len(masks) == 0 {
		return fmt.Errorf("no updatable fields provided")
	}

	op, err := svc.Projects.Locations.Keys.Patch(name, key).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating API key: %w", err)
	}
	fmt.Printf("Update API key in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runAPIKeyUndelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIKeysService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Keys.Undelete(apiKeyName(project, args[0]), &apikeys.V2UndeleteKeyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting API key: %w", err)
	}
	fmt.Printf("Undelete API key in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}
