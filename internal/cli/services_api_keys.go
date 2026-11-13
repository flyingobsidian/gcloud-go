package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apikeys "google.golang.org/api/apikeys/v2"
)

// apiKeyCheckExistingUsage mirrors get_check_existing_usage_enum: CHECK
// unless the check was explicitly turned off.
func apiKeyCheckExistingUsage(check, noCheck bool) string {
	if noCheck || !check {
		return "SKIP"
	}
	return "CHECK"
}

func (g *servicesCmds) initServicesApiKeys() {
	restrictionFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&g.flagAPIKeyDisplayName, "display-name", "", "Human-readable display name for the key")
		c.Flags().StringToStringVar(&g.flagAPIKeyAnnotations, "annotations", nil, "Annotations to attach to the key")
		c.Flags().StringSliceVar(&g.flagAPIKeyAllowedIPs, "allowed-ips", nil, "IPv4/IPv6 addresses or CIDR ranges permitted to use the key")
		c.Flags().StringSliceVar(&g.flagAPIKeyAllowedReferrers, "allowed-referrers", nil, "HTTP referrers permitted to use the key")
		c.Flags().StringSliceVar(&g.flagAPIKeyAllowedBundleIDs, "allowed-bundle-ids", nil, "iOS bundle IDs permitted to use the key")
		c.Flags().StringSliceVar(&g.flagAPIKeyAllowedAndroidPackages, "allowed-application", nil, "Android app package names permitted to use the key")
		c.Flags().StringSliceVar(&g.flagAPIKeyAllowedAndroidSHAs, "allowed-application-sha1", nil, "Android app SHA-1 fingerprints, paired with --allowed-application")
		c.Flags().StringSliceVar(&g.flagAPIKeyAPITargets, "api-target", nil, "API restriction of the form SERVICE_NAME[:METHOD[:METHOD...]]")
		c.Flags().StringVar(&g.flagAPIKeyServiceAccount, "service-account", "", "Service account email to bind the key to")
	}
	restrictionFlags(g.apiKeyCreateCmd)
	restrictionFlags(g.apiKeyUpdateCmd)
	g.apiKeyUpdateCmd.Flags().BoolVar(&g.flagAPIKeyAppendRestrictions, "append", false,
		"Merge new --allowed-application and --api-target restrictions with the "+
			"key's existing restrictions instead of replacing them")

	// --[no-]check-existing-usage added in gcloud-python 584.0.0 (#1942).
	for _, c := range []*cobra.Command{g.apiKeyUpdateCmd, g.apiKeyDeleteCmd} {
		c.Flags().BoolVar(&g.flagAPIKeyCheckUsage, "check-existing-usage", true,
			"Check existing usage before the key is changed; the request fails if the key has incompatible traffic in the last 7 days (defaults to true)")
		c.Flags().BoolVar(&g.flagAPIKeyNoCheckUsage, "no-check-existing-usage", false, "Skip the existing usage check")
		c.MarkFlagsMutuallyExclusive("check-existing-usage", "no-check-existing-usage")
	}

	g.apiKeyListCmd.Flags().BoolVar(&g.flagAPIKeyListShowDeleted, "show-deleted", false, "Include deleted keys in results")
	g.apiKeyListCmd.Flags().StringVar(&g.flagAPIKeyListFormat, "format", "", "Output format (json, yaml, or table)")
	g.apiKeyListCmd.Flags().Int64Var(&g.flagAPIKeyListPageSize, "page-size", 0, "Page size for API pagination")
	g.apiKeyListCmd.Flags().Int64Var(&g.flagAPIKeyListLimit, "limit", 0, "Maximum number of keys to list (0 = no limit)")

	g.servicesAPIKeysCmd.AddCommand(
		g.apiKeyCreateCmd, g.apiKeyDeleteCmd, g.apiKeyDescribeCmd, g.apiKeyListCmd,
		g.apiKeyLookupCmd, g.apiKeyGetKeyStringCmd, g.apiKeyUpdateCmd, g.apiKeyUndeleteCmd,
	)
	g.servicesCmd.AddCommand(g.servicesAPIKeysCmd)
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
func (g *servicesCmds) buildAPIKeyRestrictions() *apikeys.V2Restrictions {
	r := &apikeys.V2Restrictions{}
	set := false
	if len(g.flagAPIKeyAllowedIPs) > 0 {
		r.ServerKeyRestrictions = &apikeys.V2ServerKeyRestrictions{AllowedIps: g.flagAPIKeyAllowedIPs}
		set = true
	}
	if len(g.flagAPIKeyAllowedReferrers) > 0 {
		r.BrowserKeyRestrictions = &apikeys.V2BrowserKeyRestrictions{AllowedReferrers: g.flagAPIKeyAllowedReferrers}
		set = true
	}
	if len(g.flagAPIKeyAllowedBundleIDs) > 0 {
		r.IosKeyRestrictions = &apikeys.V2IosKeyRestrictions{AllowedBundleIds: g.flagAPIKeyAllowedBundleIDs}
		set = true
	}
	if len(g.flagAPIKeyAllowedAndroidPackages) > 0 {
		apps := make([]*apikeys.V2AndroidApplication, 0, len(g.flagAPIKeyAllowedAndroidPackages))
		for i, pkg := range g.flagAPIKeyAllowedAndroidPackages {
			app := &apikeys.V2AndroidApplication{PackageName: pkg}
			if i < len(g.flagAPIKeyAllowedAndroidSHAs) {
				app.Sha1Fingerprint = g.flagAPIKeyAllowedAndroidSHAs[i]
			}
			apps = append(apps, app)
		}
		r.AndroidKeyRestrictions = &apikeys.V2AndroidKeyRestrictions{AllowedApplications: apps}
		set = true
	}
	if len(g.flagAPIKeyAPITargets) > 0 {
		r.ApiTargets = parseAPITargets(g.flagAPIKeyAPITargets)
		set = true
	}
	if !set {
		return nil
	}
	return r
}

// mergeAPIKeyRestrictions merges new-application and API-target lists from `add`
// into `base`. Non-list restriction fields (IPs, referrers, iOS bundle IDs) are
// overwritten when present on `add`, matching the reference Python behavior for
// `--append` on `gcloud services api-keys update` (which only merges the
// android-application and api-target lists).
func mergeAPIKeyRestrictions(base, add *apikeys.V2Restrictions) *apikeys.V2Restrictions {
	if base == nil {
		return add
	}
	merged := *base
	if add == nil {
		return &merged
	}
	if add.ServerKeyRestrictions != nil {
		merged.ServerKeyRestrictions = add.ServerKeyRestrictions
	}
	if add.BrowserKeyRestrictions != nil {
		merged.BrowserKeyRestrictions = add.BrowserKeyRestrictions
	}
	if add.IosKeyRestrictions != nil {
		merged.IosKeyRestrictions = add.IosKeyRestrictions
	}
	if add.AndroidKeyRestrictions != nil {
		if merged.AndroidKeyRestrictions == nil {
			merged.AndroidKeyRestrictions = &apikeys.V2AndroidKeyRestrictions{}
		}
		seen := map[string]bool{}
		for _, a := range merged.AndroidKeyRestrictions.AllowedApplications {
			seen[a.PackageName+"|"+a.Sha1Fingerprint] = true
		}
		for _, a := range add.AndroidKeyRestrictions.AllowedApplications {
			key := a.PackageName + "|" + a.Sha1Fingerprint
			if seen[key] {
				continue
			}
			seen[key] = true
			merged.AndroidKeyRestrictions.AllowedApplications = append(
				merged.AndroidKeyRestrictions.AllowedApplications, a)
		}
	}
	if len(add.ApiTargets) > 0 {
		byService := map[string]*apikeys.V2ApiTarget{}
		var order []string
		for _, t := range merged.ApiTargets {
			byService[t.Service] = t
			order = append(order, t.Service)
		}
		for _, t := range add.ApiTargets {
			if existing, ok := byService[t.Service]; ok {
				seen := map[string]bool{}
				for _, m := range existing.Methods {
					seen[m] = true
				}
				for _, m := range t.Methods {
					if !seen[m] {
						existing.Methods = append(existing.Methods, m)
						seen[m] = true
					}
				}
				continue
			}
			byService[t.Service] = t
			order = append(order, t.Service)
		}
		merged.ApiTargets = merged.ApiTargets[:0]
		for _, s := range order {
			merged.ApiTargets = append(merged.ApiTargets, byService[s])
		}
	}
	return &merged
}

// errAPIKeyCreateNeedsTarget mirrors the InvalidArgumentException that
// gcloud-python 586.0.0+ raises when `services api-keys create` lacks
// --api-target.
var errAPIKeyCreateNeedsTarget = fmt.Errorf("Invalid value for [--api-target]: API keys must be created with API target restrictions. Please specify `--api-target`.")

// errAPIKeyUpdateNeedsTarget mirrors the InvalidArgumentException that
// gcloud-python 586.0.0+ raises when `services api-keys update` would leave a
// key without API target restrictions.
var errAPIKeyUpdateNeedsTarget = fmt.Errorf("Invalid value for [--api-target]: API keys must have API target restrictions. Please specify `--api-target` to restrict this key.")

// checkAPIKeyUpdateTargets reports whether an update may proceed: the key must
// already carry API targets or the caller must be supplying --api-target.
func checkAPIKeyUpdateTargets(current *apikeys.V2Key, specifyingTarget bool) error {
	if specifyingTarget {
		return nil
	}
	if current != nil && current.Restrictions != nil && len(current.Restrictions.ApiTargets) > 0 {
		return nil
	}
	return errAPIKeyUpdateNeedsTarget
}

func (g *servicesCmds) runAPIKeyCreate(cmd *cobra.Command, args []string) error {
	if !cmd.Flags().Changed("api-target") {
		return errAPIKeyCreateNeedsTarget
	}
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().APIKeys(ctx)
	if err != nil {
		return err
	}
	key := &apikeys.V2Key{
		DisplayName:         g.flagAPIKeyDisplayName,
		Annotations:         g.flagAPIKeyAnnotations,
		Restrictions:        g.buildAPIKeyRestrictions(),
		ServiceAccountEmail: g.flagAPIKeyServiceAccount,
	}
	op, err := svc.Projects.Locations.Keys.Create(apiKeyParent(project), key).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating API key: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Create API key in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) runAPIKeyDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().APIKeys(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Keys.Delete(apiKeyName(project, args[0])).
		CheckExistingUsage(apiKeyCheckExistingUsage(g.flagAPIKeyCheckUsage, g.flagAPIKeyNoCheckUsage)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting API key: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Delete API key in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) runAPIKeyDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().APIKeys(ctx)
	if err != nil {
		return err
	}
	key, err := svc.Projects.Locations.Keys.Get(apiKeyName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing API key: %w", err)
	}
	return g.inv.yaml(key)
}

func (g *servicesCmds) runAPIKeyList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().APIKeys(ctx)
	if err != nil {
		return err
	}

	var all []*apikeys.V2Key
	pageToken := ""
	for {
		call := svc.Projects.Locations.Keys.List(apiKeyParent(project)).ShowDeleted(g.flagAPIKeyListShowDeleted).Context(ctx)
		if g.flagAPIKeyListPageSize > 0 {
			call = call.PageSize(g.flagAPIKeyListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing API keys: %w", err)
		}
		all = append(all, resp.Keys...)
		if g.flagAPIKeyListLimit > 0 && int64(len(all)) >= g.flagAPIKeyListLimit {
			all = all[:g.flagAPIKeyListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return g.inv.printListResults(all, g.flagAPIKeyListFormat, func() {
		printTableHeadingTo(g.inv.stdout, len(all), "%-40s %-30s %s\n", "NAME", "DISPLAY_NAME", "CREATE_TIME")
		for _, k := range all {
			fmt.Fprintf(g.inv.stdout, "%-40s %-30s %s\n", k.Name, k.DisplayName, k.CreateTime)
		}
	})
}

func (g *servicesCmds) runAPIKeyLookup(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.inv.clients().APIKeys(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Keys.LookupKey().KeyString(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up API key: %w", err)
	}
	return g.inv.yaml(resp)
}

func (g *servicesCmds) runAPIKeyGetKeyString(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().APIKeys(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Keys.GetKeyString(apiKeyName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting API key string: %w", err)
	}
	fmt.Fprintln(g.inv.stdout, resp.KeyString)
	return nil
}

func (g *servicesCmds) runAPIKeyUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().APIKeys(ctx)
	if err != nil {
		return err
	}

	name := apiKeyName(project, args[0])
	current, err := svc.Projects.Locations.Keys.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching existing API key: %w", err)
	}
	if err := checkAPIKeyUpdateTargets(current, cmd.Flags().Changed("api-target")); err != nil {
		return err
	}
	key := &apikeys.V2Key{Name: name}

	var masks []string
	if cmd.Flags().Changed("display-name") {
		key.DisplayName = g.flagAPIKeyDisplayName
		masks = append(masks, "display_name")
	}
	if cmd.Flags().Changed("annotations") {
		key.Annotations = g.flagAPIKeyAnnotations
		masks = append(masks, "annotations")
	}
	if cmd.Flags().Changed("service-account") {
		key.ServiceAccountEmail = g.flagAPIKeyServiceAccount
		masks = append(masks, "service_account_email")
	}
	if r := g.buildAPIKeyRestrictions(); r != nil {
		if g.flagAPIKeyAppendRestrictions {
			r = mergeAPIKeyRestrictions(current.Restrictions, r)
		}
		key.Restrictions = r
		masks = append(masks, "restrictions")
	}
	if len(masks) == 0 {
		return fmt.Errorf("no updatable fields provided")
	}

	op, err := svc.Projects.Locations.Keys.Patch(name, key).UpdateMask(strings.Join(masks, ",")).
		CheckExistingUsage(apiKeyCheckExistingUsage(g.flagAPIKeyCheckUsage, g.flagAPIKeyNoCheckUsage)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating API key: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Update API key in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *servicesCmds) runAPIKeyUndelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().APIKeys(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Keys.Undelete(apiKeyName(project, args[0]), &apikeys.V2UndeleteKeyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting API key: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Undelete API key in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}
