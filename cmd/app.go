package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	appengine "google.golang.org/api/appengine/v1"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud app (#299, #865-#874) ---

// Shared flags for the app subcommands.

// appResolveApp returns the App Engine application id (the project) and validates a project
// has been selected.
func (g *appCmds) appResolveApp() (string, error) {
	return g.inv.project()
}

func appWaitOp(ctx context.Context, svc *appengine.APIService, app string, op *appengine.Operation) (*appengine.Operation, error) {
	backoff := time.Second
	for !op.Done {
		time.Sleep(backoff)
		if backoff < 8*time.Second {
			backoff *= 2
		}
		id := path.Base(op.Name)
		got, err := svc.Apps.Operations.Get(app, id).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (g *appCmds) appFinishOp(ctx context.Context, svc *appengine.APIService, app string, op *appengine.Operation, verb, name string) error {
	if g.flagAppAsync {
		fmt.Fprintf(g.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(g.inv.stdout, op, "")
	}
	final, err := appWaitOp(ctx, svc, app, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(g.inv.stdout, final.Response, "")
	}
	return nil
}

// --- domain-mappings ---

// --- firewall-rules ---

// --- instances ---

// --- logs ---

// --- operations ---

// --- regions ---

// --- runtimes ---

// --- services ---

// --- ssl-certificates ---

// --- versions ---

func (g *appCmds) initApp() {
	// --- domain-mappings ---
	for _, c := range []*cobra.Command{g.appDomMapCreateCmd, g.appDomMapDeleteCmd, g.appDomMapDescribeCmd, g.appDomMapListCmd, g.appDomMapUpdateCmd} {
		c.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
		c.Flags().BoolVar(&g.flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	}
	g.appDomMapCreateCmd.Flags().StringVar(&g.flagAppCertificateID, "certificate-id", "", "Manually managed certificate id to associate with this domain")
	g.appDomMapCreateCmd.Flags().StringVar(&g.flagAppCertificateMgmt, "certificate-management", "", "Certificate management type: automatic or manual")
	g.appDomMapUpdateCmd.Flags().StringVar(&g.flagAppCertificateID, "certificate-id", "", "Certificate id to associate with this domain")
	g.appDomMapUpdateCmd.Flags().BoolVar(&g.flagAppNoCertificateID, "no-certificate-id", false, "Remove certificate association from this domain")
	g.appDomMapUpdateCmd.Flags().StringVar(&g.flagAppCertificateMgmt, "certificate-management", "", "Certificate management type: automatic or manual")
	for _, c := range []*cobra.Command{g.appDomMapCreateCmd, g.appDomMapDeleteCmd, g.appDomMapDescribeCmd, g.appDomMapListCmd, g.appDomMapUpdateCmd} {
		g.appDomainMappingsCmd.AddCommand(c)
	}
	g.appCmd.AddCommand(g.appDomainMappingsCmd)

	// --- firewall-rules ---
	for _, c := range []*cobra.Command{g.appFwCreateCmd, g.appFwDeleteCmd, g.appFwDescribeCmd, g.appFwListCmd, g.appFwUpdateCmd, g.appFwTestIPCmd} {
		c.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	}
	g.appFwCreateCmd.Flags().StringVar(&g.flagAppSourceRange, "source-range", "", "IP address or CIDR range (required)")
	g.appFwCreateCmd.Flags().StringVar(&g.flagAppAction, "action", "", "ALLOW or DENY (required)")
	g.appFwCreateCmd.Flags().StringVar(&g.flagAppDescription, "description", "", "A text description of the rule")
	_ = g.appFwCreateCmd.MarkFlagRequired("source-range")
	_ = g.appFwCreateCmd.MarkFlagRequired("action")
	g.appFwUpdateCmd.Flags().StringVar(&g.flagAppSourceRange, "source-range", "", "IP address or CIDR range")
	g.appFwUpdateCmd.Flags().StringVar(&g.flagAppAction, "action", "", "ALLOW or DENY")
	g.appFwUpdateCmd.Flags().StringVar(&g.flagAppDescription, "description", "", "A text description of the rule")
	g.appFwListCmd.Flags().StringVar(&g.flagAppMatchingAddress, "matching-address", "", "Only return rules matching this IP address")
	for _, c := range []*cobra.Command{g.appFwCreateCmd, g.appFwDeleteCmd, g.appFwDescribeCmd, g.appFwListCmd, g.appFwUpdateCmd, g.appFwTestIPCmd} {
		g.appFirewallRulesCmd.AddCommand(c)
	}
	g.appCmd.AddCommand(g.appFirewallRulesCmd)

	// --- instances ---
	for _, c := range []*cobra.Command{g.appInstDeleteCmd, g.appInstDescribeCmd, g.appInstEnableDebugCmd, g.appInstDisableDebugCmd, g.appInstListCmd} {
		c.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.appInstDeleteCmd, g.appInstDescribeCmd, g.appInstEnableDebugCmd, g.appInstDisableDebugCmd} {
		c.Flags().StringVarP(&g.flagAppService, "service", "s", "", "The service ID (required)")
		c.Flags().StringVarP(&g.flagAppVersion, "version", "v", "", "The version ID (required)")
		_ = c.MarkFlagRequired("service")
		_ = c.MarkFlagRequired("version")
	}
	for _, c := range []*cobra.Command{g.appInstDeleteCmd, g.appInstEnableDebugCmd, g.appInstDisableDebugCmd} {
		c.Flags().BoolVar(&g.flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	}
	g.appInstListCmd.Flags().StringVarP(&g.flagAppService, "service", "s", "", "Only show instances from this service")
	g.appInstListCmd.Flags().StringVarP(&g.flagAppVersion, "version", "v", "", "Only show instances from this version")
	for _, c := range []*cobra.Command{g.appInstDeleteCmd, g.appInstDescribeCmd, g.appInstEnableDebugCmd, g.appInstDisableDebugCmd, g.appInstListCmd} {
		g.appInstancesCmd.AddCommand(c)
	}
	g.appCmd.AddCommand(g.appInstancesCmd)

	// --- logs ---
	for _, c := range []*cobra.Command{g.appLogsReadCmd, g.appLogsTailCmd} {
		c.Flags().StringVarP(&g.flagAppService, "service", "s", "", "Only entries from this service")
		c.Flags().StringVarP(&g.flagAppVersion, "version", "v", "", "Only entries from this version")
		c.Flags().StringVar(&g.flagAppLevel, "level", "any", "Minimum severity: any, debug, info, notice, warning, error, critical, alert, emergency")
		c.Flags().StringVar(&g.flagAppLogs, "logs", "", "Comma-separated log ids to include (defaults to stdout,stderr,crash.log,nginx.request,request_log)")
		c.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	}
	g.appLogsReadCmd.Flags().Int64Var(&g.flagAppLimit, "limit", 200, "Number of log entries to show")
	g.appLogsTailCmd.Flags().Int64Var(&g.flagAppLimit, "limit", 200, "Number of log entries to show")
	g.appLogsCmd.AddCommand(g.appLogsReadCmd, g.appLogsTailCmd)
	g.appCmd.AddCommand(g.appLogsCmd)

	// --- operations ---
	for _, c := range []*cobra.Command{g.appOpDescribeCmd, g.appOpListCmd, g.appOpWaitCmd} {
		c.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	}
	g.appOpListCmd.Flags().BoolVar(&g.flagAppPending, "pending", false, "Only show pending (not done) operations")
	g.appOpListCmd.Flags().StringVar(&g.flagAppFilter, "filter", "", "Filter expression applied to the API")
	g.appOperationsCmd.AddCommand(g.appOpDescribeCmd, g.appOpListCmd, g.appOpWaitCmd)
	g.appCmd.AddCommand(g.appOperationsCmd)

	// --- regions ---
	g.appRegionsListCmd.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	g.appRegionsCmd.AddCommand(g.appRegionsListCmd)
	g.appCmd.AddCommand(g.appRegionsCmd)

	// --- runtimes ---
	g.appRuntimesListCmd.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	g.appRuntimesListCmd.Flags().StringVar(&g.flagAppEnvironment, "environment", "", "Only include runtimes in this environment: STANDARD or FLEXIBLE")
	g.appRuntimesCmd.AddCommand(g.appRuntimesListCmd)
	g.appCmd.AddCommand(g.appRuntimesCmd)

	// --- services ---
	for _, c := range []*cobra.Command{g.appSvcBrowseCmd, g.appSvcDeleteCmd, g.appSvcDescribeCmd, g.appSvcListCmd, g.appSvcSetTrafficCmd, g.appSvcUpdateCmd} {
		c.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	}
	g.appSvcBrowseCmd.Flags().StringVarP(&g.flagAppVersion, "version", "v", "", "Browse the specific version")
	g.appSvcBrowseCmd.Flags().BoolVar(&g.flagAppLaunchBrowser, "launch-browser", false, "Attempt to launch a browser (best-effort; prints URL either way)")
	g.appSvcDeleteCmd.Flags().StringVar(&g.flagAppVersion, "version", "", "Delete only this version instead of the entire service")
	g.appSvcDeleteCmd.Flags().BoolVar(&g.flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	g.appSvcSetTrafficCmd.Flags().StringToStringVar(&g.flagAppSplits, "splits", nil, "Version=weight pairs describing traffic allocation (required)")
	g.appSvcSetTrafficCmd.Flags().StringVar(&g.flagAppSplitBy, "split-by", "ip", "Split traffic by: cookie, ip, random")
	g.appSvcSetTrafficCmd.Flags().BoolVar(&g.flagAppMigrate, "migrate", false, "Gracefully migrate traffic to new versions")
	g.appSvcSetTrafficCmd.Flags().BoolVar(&g.flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	_ = g.appSvcSetTrafficCmd.MarkFlagRequired("splits")
	g.appSvcUpdateCmd.Flags().StringVar(&g.flagAppIngress, "ingress", "", "Network ingress setting: all, internal-only, internal-and-cloud-load-balancing")
	g.appSvcUpdateCmd.Flags().BoolVar(&g.flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	for _, c := range []*cobra.Command{g.appSvcBrowseCmd, g.appSvcDeleteCmd, g.appSvcDescribeCmd, g.appSvcListCmd, g.appSvcSetTrafficCmd, g.appSvcUpdateCmd} {
		g.appServicesCmd.AddCommand(c)
	}
	g.appCmd.AddCommand(g.appServicesCmd)

	// --- ssl-certificates ---
	for _, c := range []*cobra.Command{g.appSSLCreateCmd, g.appSSLDeleteCmd, g.appSSLDescribeCmd, g.appSSLListCmd, g.appSSLUpdateCmd} {
		c.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	}
	g.appSSLCreateCmd.Flags().StringVar(&g.flagAppDisplayName, "display-name", "", "Display name for this certificate (required)")
	g.appSSLCreateCmd.Flags().StringVar(&g.flagAppCertificateFile, "certificate", "", "Path to the PEM x.509 certificate file (required)")
	g.appSSLCreateCmd.Flags().StringVar(&g.flagAppPrivateKeyFile, "private-key", "", "Path to the PEM RSA private key file (required)")
	_ = g.appSSLCreateCmd.MarkFlagRequired("display-name")
	_ = g.appSSLCreateCmd.MarkFlagRequired("certificate")
	_ = g.appSSLCreateCmd.MarkFlagRequired("private-key")
	g.appSSLUpdateCmd.Flags().StringVar(&g.flagAppDisplayName, "display-name", "", "New display name for the certificate")
	g.appSSLUpdateCmd.Flags().StringVar(&g.flagAppCertificateFile, "certificate", "", "Path to a replacement PEM x.509 certificate")
	g.appSSLUpdateCmd.Flags().StringVar(&g.flagAppPrivateKeyFile, "private-key", "", "Path to a replacement PEM RSA private key")
	for _, c := range []*cobra.Command{g.appSSLCreateCmd, g.appSSLDeleteCmd, g.appSSLDescribeCmd, g.appSSLListCmd, g.appSSLUpdateCmd} {
		g.appSSLCertsCmd.AddCommand(c)
	}
	g.appCmd.AddCommand(g.appSSLCertsCmd)

	// --- versions ---
	for _, c := range []*cobra.Command{g.appVerBrowseCmd, g.appVerDeleteCmd, g.appVerDescribeCmd, g.appVerListCmd, g.appVerMigrateCmd, g.appVerStartCmd, g.appVerStopCmd} {
		c.Flags().StringVar(&g.flagAppFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.appVerBrowseCmd, g.appVerDeleteCmd, g.appVerDescribeCmd, g.appVerMigrateCmd, g.appVerStartCmd, g.appVerStopCmd} {
		c.Flags().StringVarP(&g.flagAppService, "service", "s", "", "Service that owns the version(s)")
	}
	g.appVerBrowseCmd.Flags().BoolVar(&g.flagAppLaunchBrowser, "launch-browser", false, "Attempt to launch a browser (best-effort; prints URL either way)")
	g.appVerListCmd.Flags().StringVarP(&g.flagAppService, "service", "s", "", "Only list versions from this service")
	g.appVerListCmd.Flags().BoolVar(&g.flagAppHideNoTraffic, "hide-no-traffic", false, "Only show versions that are receiving traffic")
	for _, c := range []*cobra.Command{g.appVerDeleteCmd, g.appVerMigrateCmd, g.appVerStartCmd, g.appVerStopCmd} {
		c.Flags().BoolVar(&g.flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	}
	for _, c := range []*cobra.Command{g.appVerBrowseCmd, g.appVerDeleteCmd, g.appVerDescribeCmd, g.appVerListCmd, g.appVerMigrateCmd, g.appVerStartCmd, g.appVerStopCmd} {
		g.appVersionsCmd.AddCommand(c)
	}
	g.appCmd.AddCommand(g.appVersionsCmd)

	// Legacy stubs for top-level app commands still to be implemented under #299.
	for _, name := range []string{"browse", "create", "deploy", "describe", "open-console", "update"} {
		registerStubCommand(g.appCmd, name, "Not yet implemented")
	}

}

// -----------------------------------------------------------------------------
// domain-mappings implementations
// -----------------------------------------------------------------------------

func (g *appCmds) runAppDomMapCreate(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	mgmt := strings.ToLower(g.flagAppCertificateMgmt)
	if mgmt == "" {
		if g.flagAppCertificateID == "" {
			mgmt = "automatic"
		} else {
			mgmt = "manual"
		}
	}
	if mgmt != "automatic" && mgmt != "manual" {
		return fmt.Errorf("--certificate-management must be automatic or manual")
	}
	mapping := &appengine.DomainMapping{
		Id: args[0],
	}
	if g.flagAppCertificateID != "" || mgmt == "manual" {
		mapping.SslSettings = &appengine.SslSettings{
			SslManagementType: strings.ToUpper(mgmt),
			CertificateId:     g.flagAppCertificateID,
		}
	} else {
		mapping.SslSettings = &appengine.SslSettings{SslManagementType: "AUTOMATIC"}
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Apps.DomainMappings.Create(app, mapping).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating domain mapping: %w", err)
	}
	return g.appFinishOp(ctx, svc, app, op, "Create domain mapping", args[0])
}

func (g *appCmds) runAppDomMapDelete(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Apps.DomainMappings.Delete(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting domain mapping: %w", err)
	}
	return g.appFinishOp(ctx, svc, app, op, "Delete domain mapping", args[0])
}

func (g *appCmds) runAppDomMapDescribe(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.DomainMappings.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain mapping: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppDomMapList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	var all []*appengine.DomainMapping
	pageToken := ""
	for {
		call := svc.Apps.DomainMappings.List(app).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing domain mappings: %w", err)
		}
		all = append(all, resp.DomainMappings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "ID", "SSL_CERTIFICATE_ID")
	for _, dm := range all {
		cert := ""
		if dm.SslSettings != nil {
			cert = dm.SslSettings.CertificateId
		}
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", dm.Id, cert)
	}
	return nil
}

func (g *appCmds) runAppDomMapUpdate(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	if g.flagAppCertificateID != "" && g.flagAppNoCertificateID {
		return fmt.Errorf("--certificate-id and --no-certificate-id are mutually exclusive")
	}
	mgmt := strings.ToLower(g.flagAppCertificateMgmt)
	if mgmt == "" && (g.flagAppCertificateID != "" || g.flagAppNoCertificateID) {
		mgmt = "manual"
	}
	if mgmt != "" && mgmt != "automatic" && mgmt != "manual" {
		return fmt.Errorf("--certificate-management must be automatic or manual")
	}
	body := &appengine.DomainMapping{}
	var masks []string
	if mgmt != "" || g.flagAppCertificateID != "" || g.flagAppNoCertificateID {
		body.SslSettings = &appengine.SslSettings{}
		if mgmt != "" {
			body.SslSettings.SslManagementType = strings.ToUpper(mgmt)
		}
		if g.flagAppCertificateID != "" {
			body.SslSettings.CertificateId = g.flagAppCertificateID
		}
		if g.flagAppNoCertificateID {
			body.SslSettings.CertificateId = ""
			body.SslSettings.ForceSendFields = append(body.SslSettings.ForceSendFields, "CertificateId")
		}
		masks = append(masks, "ssl_settings")
	}
	if len(masks) == 0 {
		return fmt.Errorf("no update options specified")
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Apps.DomainMappings.Patch(app, args[0], body).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating domain mapping: %w", err)
	}
	return g.appFinishOp(ctx, svc, app, op, "Update domain mapping", args[0])
}

// -----------------------------------------------------------------------------
// firewall-rules implementations
// -----------------------------------------------------------------------------

func parsePriority(s string) (int64, error) {
	if strings.EqualFold(s, "default") {
		// The API represents the catch-all rule as Int32.MaxValue.
		return 2147483647, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

func (g *appCmds) runAppFwCreate(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	if strings.EqualFold(args[0], "default") {
		return fmt.Errorf("the `default` rule cannot be created, only updated")
	}
	priority, err := parsePriority(args[0])
	if err != nil {
		return fmt.Errorf("invalid priority %q: %w", args[0], err)
	}
	action := strings.ToUpper(g.flagAppAction)
	if action != "ALLOW" && action != "DENY" {
		return fmt.Errorf("--action must be ALLOW or DENY")
	}
	body := &appengine.FirewallRule{
		Priority:    priority,
		SourceRange: g.flagAppSourceRange,
		Action:      action,
		Description: g.flagAppDescription,
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Firewall.IngressRules.Create(app, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating firewall rule: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Created firewall rule [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppFwDelete(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Apps.Firewall.IngressRules.Delete(app, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting firewall rule: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Deleted firewall rule [%s].\n", args[0])
	return nil
}

func (g *appCmds) runAppFwDescribe(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Firewall.IngressRules.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing firewall rule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppFwList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	var all []*appengine.FirewallRule
	pageToken := ""
	for {
		call := svc.Apps.Firewall.IngressRules.List(app).Context(ctx)
		if g.flagAppMatchingAddress != "" {
			call = call.MatchingAddress(g.flagAppMatchingAddress)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing firewall rules: %w", err)
		}
		all = append(all, resp.IngressRules...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-12s %-8s %-32s %s\n", "PRIORITY", "ACTION", "SOURCE_RANGE", "DESCRIPTION")
	for _, r := range all {
		fmt.Fprintf(g.inv.stdout, "%-12d %-8s %-32s %s\n", r.Priority, r.Action, r.SourceRange, r.Description)
	}
	return nil
}

func (g *appCmds) runAppFwUpdate(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	body := &appengine.FirewallRule{}
	var masks []string
	if g.flagAppAction != "" {
		action := strings.ToUpper(g.flagAppAction)
		if action != "ALLOW" && action != "DENY" {
			return fmt.Errorf("--action must be ALLOW or DENY")
		}
		body.Action = action
		masks = append(masks, "action")
	}
	if g.flagAppSourceRange != "" {
		body.SourceRange = g.flagAppSourceRange
		masks = append(masks, "sourceRange")
	}
	if cmd.Flags().Changed("description") {
		body.Description = g.flagAppDescription
		masks = append(masks, "description")
	}
	if len(masks) == 0 {
		return fmt.Errorf("no update options specified")
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Firewall.IngressRules.Patch(app, args[0], body).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating firewall rule: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Updated firewall rule [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppFwTestIP(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	var all []*appengine.FirewallRule
	pageToken := ""
	for {
		call := svc.Apps.Firewall.IngressRules.List(app).MatchingAddress(args[0]).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("testing firewall rules for %s: %w", args[0], err)
		}
		all = append(all, resp.IngressRules...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if len(all) == 0 {
		fmt.Fprintln(g.inv.stderr, "No rules match the IP address.")
		return nil
	}
	fmt.Fprintf(g.inv.stderr, "The action `%s` will apply to the IP address.\n", all[0].Action)
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-12s %-8s %-32s %s\n", "PRIORITY", "ACTION", "SOURCE_RANGE", "DESCRIPTION")
	for _, r := range all {
		fmt.Fprintf(g.inv.stdout, "%-12d %-8s %-32s %s\n", r.Priority, r.Action, r.SourceRange, r.Description)
	}
	return nil
}

// -----------------------------------------------------------------------------
// instances implementations
// -----------------------------------------------------------------------------

func (g *appCmds) runAppInstDelete(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Apps.Services.Versions.Instances.Delete(app, g.flagAppService, g.flagAppVersion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return g.appFinishOp(ctx, svc, app, op, "Delete instance", args[0])
}

func (g *appCmds) runAppInstDescribe(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Services.Versions.Instances.Get(app, g.flagAppService, g.flagAppVersion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppInstEnableDebug(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Apps.Services.Versions.Instances.Debug(app, g.flagAppService, g.flagAppVersion, args[0], &appengine.DebugInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling debug: %w", err)
	}
	return g.appFinishOp(ctx, svc, app, op, "Enable debug on instance", args[0])
}

// runAppInstDisableDebug: the App Engine v1 API exposes only the Debug method (which
// enables debug on flex instances). Disabling debug requires deleting the instance,
// which is destructive; we surface an explicit error instead of silently deleting.
func runAppInstDisableDebug(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("disable-debug is not supported by the App Engine Admin v1 API; delete the instance with `gcloud app instances delete` to remove its debug session")
}

func (g *appCmds) runAppInstList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}

	type instanceRow struct {
		Service  string
		Version  string
		Instance *appengine.Instance
	}
	var rows []instanceRow

	// Determine services to iterate over.
	services, err := appListServices(ctx, svc, app)
	if err != nil {
		return err
	}
	for _, s := range services {
		if g.flagAppService != "" && s.Id != g.flagAppService {
			continue
		}
		versions, err := appListVersions(ctx, svc, app, s.Id)
		if err != nil {
			return err
		}
		for _, v := range versions {
			if g.flagAppVersion != "" && v.Id != g.flagAppVersion {
				continue
			}
			instances, err := appListInstances(ctx, svc, app, s.Id, v.Id)
			if err != nil {
				// Standard-environment versions have no listable instances; skip.
				continue
			}
			for _, inst := range instances {
				rows = append(rows, instanceRow{Service: s.Id, Version: v.Id, Instance: inst})
			}
		}
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, rows, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(rows), "%-20s %-20s %-40s %s\n", "SERVICE", "VERSION", "ID", "VM_STATUS")
	for _, r := range rows {
		fmt.Fprintf(g.inv.stdout, "%-20s %-20s %-40s %s\n", r.Service, r.Version, r.Instance.Id, r.Instance.VmStatus)
	}
	return nil
}

func appListInstances(ctx context.Context, svc *appengine.APIService, app, service, version string) ([]*appengine.Instance, error) {
	var all []*appengine.Instance
	pageToken := ""
	for {
		call := svc.Apps.Services.Versions.Instances.List(app, service, version).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Instances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return all, nil
}

// -----------------------------------------------------------------------------
// logs implementations
// -----------------------------------------------------------------------------

func defaultAppLogIDs() []string {
	return []string{"stdout", "stderr", "crash.log", "nginx.request", "appengine.googleapis.com%2Frequest_log"}
}

func (g *appCmds) buildAppLogFilter(project string) string {
	parts := []string{`resource.type="gae_app"`}
	if g.flagAppService != "" {
		parts = append(parts, fmt.Sprintf(`resource.labels.module_id=%q`, g.flagAppService))
	}
	if g.flagAppVersion != "" {
		parts = append(parts, fmt.Sprintf(`resource.labels.version_id=%q`, g.flagAppVersion))
	}
	if g.flagAppLevel != "" && !strings.EqualFold(g.flagAppLevel, "any") {
		parts = append(parts, fmt.Sprintf(`severity>=%s`, strings.ToUpper(g.flagAppLevel)))
	}
	logs := defaultAppLogIDs()
	if g.flagAppLogs != "" {
		logs = strings.Split(g.flagAppLogs, ",")
	}
	if len(logs) > 0 {
		var ors []string
		for _, l := range logs {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			ors = append(ors, fmt.Sprintf(`logName="projects/%s/logs/%s"`, project, l))
		}
		if len(ors) > 0 {
			parts = append(parts, "("+strings.Join(ors, " OR ")+")")
		}
	}
	return strings.Join(parts, " AND ")
}

func (g *appCmds) runAppLogsRead(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	if g.flagAppLimit < 1 || g.flagAppLimit > 1000 {
		return fmt.Errorf("--limit must be between 1 and 1000")
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Logging(ctx)
	if err != nil {
		return err
	}
	req := &logging.ListLogEntriesRequest{
		Filter:        g.buildAppLogFilter(project),
		OrderBy:       "timestamp desc",
		PageSize:      g.flagAppLimit,
		ResourceNames: []string{"projects/" + project},
	}
	resp, err := svc.Entries.List(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing log entries: %w", err)
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Entries, g.flagAppFormat)
	}
	// Reverse so the newest entry prints last, matching gcloud.
	for i := len(resp.Entries) - 1; i >= 0; i-- {
		g.printAppLogEntry(resp.Entries[i])
	}
	return nil
}

func (g *appCmds) runAppLogsTail(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	if g.flagAppLimit < 1 || g.flagAppLimit > 1000 {
		return fmt.Errorf("--limit must be between 1 and 1000")
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Logging(ctx)
	if err != nil {
		return err
	}
	req := &logging.ListLogEntriesRequest{
		Filter:        g.buildAppLogFilter(project),
		OrderBy:       "timestamp desc",
		PageSize:      g.flagAppLimit,
		ResourceNames: []string{"projects/" + project},
	}
	resp, err := svc.Entries.List(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing log entries: %w", err)
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Entries, g.flagAppFormat)
	}
	for i := len(resp.Entries) - 1; i >= 0; i-- {
		g.printAppLogEntry(resp.Entries[i])
	}
	return nil
}

func (g *appCmds) printAppLogEntry(e *logging.LogEntry) {
	msg := e.TextPayload
	if msg == "" && e.JsonPayload != nil {
		msg = string(e.JsonPayload)
	}
	if msg == "" && e.HttpRequest != nil {
		msg = fmt.Sprintf("%s %s -> %d", e.HttpRequest.RequestMethod, e.HttpRequest.RequestUrl, e.HttpRequest.Status)
	}
	fmt.Fprintf(g.inv.stdout, "%s %s %s\n", e.Timestamp, e.Severity, strings.TrimSpace(msg))
}

// -----------------------------------------------------------------------------
// operations implementations
// -----------------------------------------------------------------------------

func (g *appCmds) runAppOpDescribe(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Operations.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppOpList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	var all []*appengine.Operation
	pageToken := ""
	for {
		call := svc.Apps.Operations.List(app).Context(ctx)
		filter := g.flagAppFilter
		if g.flagAppPending {
			if filter != "" {
				filter = filter + " AND done=false"
			} else {
				filter = "done=false"
			}
		}
		if filter != "" {
			call = call.Filter(filter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-50s %-8s\n", "ID", "DONE")
	for _, op := range all {
		fmt.Fprintf(g.inv.stdout, "%-50s %v\n", path.Base(op.Name), op.Done)
	}
	return nil
}

func (g *appCmds) runAppOpWait(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Apps.Operations.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching operation: %w", err)
	}
	if op.Done {
		fmt.Fprintf(g.inv.stderr, "Operation [%s] is already done.\n", args[0])
		return emitFormattedTo(g.inv.stdout, op, g.flagAppFormat)
	}
	fmt.Fprintf(g.inv.stderr, "Waiting for operation [%s] to complete.\n", args[0])
	final, err := appWaitOp(ctx, svc, app, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(g.inv.stdout, final, g.flagAppFormat)
}

// -----------------------------------------------------------------------------
// regions implementations
// -----------------------------------------------------------------------------

func (g *appCmds) runAppRegionsList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	var all []*appengine.Location
	pageToken := ""
	for {
		call := svc.Apps.Locations.List(app).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing regions: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	sort.Slice(all, func(i, j int) bool { return all[i].LocationId < all[j].LocationId })
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-24s %-16s %s\n", "REGION", "STANDARD", "FLEXIBLE")
	for _, l := range all {
		std, flex := "-", "-"
		var meta struct {
			StandardEnvironmentAvailable bool `json:"standardEnvironmentAvailable"`
			FlexibleEnvironmentAvailable bool `json:"flexibleEnvironmentAvailable"`
		}
		if l.Metadata != nil {
			_ = json.Unmarshal(l.Metadata, &meta)
			if meta.StandardEnvironmentAvailable {
				std = "YES"
			} else {
				std = "NO"
			}
			if meta.FlexibleEnvironmentAvailable {
				flex = "YES"
			} else {
				flex = "NO"
			}
		}
		fmt.Fprintf(g.inv.stdout, "%-24s %-16s %s\n", l.LocationId, std, flex)
	}
	return nil
}

// -----------------------------------------------------------------------------
// runtimes implementations
// -----------------------------------------------------------------------------

func (g *appCmds) runAppRuntimesList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	call := svc.Apps.ListRuntimes(app).Context(ctx)
	if g.flagAppEnvironment != "" {
		call = call.Environment(strings.ToUpper(g.flagAppEnvironment))
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing runtimes: %w", err)
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Runtimes, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Runtimes), "%-24s %-16s %-16s %s\n", "NAME", "ENVIRONMENT", "STAGE", "DISPLAY_NAME")
	for _, r := range resp.Runtimes {
		fmt.Fprintf(g.inv.stdout, "%-24s %-16s %-16s %s\n", r.Name, r.Environment, r.Stage, r.DisplayName)
	}
	return nil
}

// -----------------------------------------------------------------------------
// services implementations
// -----------------------------------------------------------------------------

func appListServices(ctx context.Context, svc *appengine.APIService, app string) ([]*appengine.Service, error) {
	var all []*appengine.Service
	pageToken := ""
	for {
		call := svc.Apps.Services.List(app).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Services...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return all, nil
}

func appListVersions(ctx context.Context, svc *appengine.APIService, app, service string) ([]*appengine.Version, error) {
	var all []*appengine.Version
	pageToken := ""
	for {
		call := svc.Apps.Services.Versions.List(app, service).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Versions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return all, nil
}

func appDefaultHostname(ctx context.Context, svc *appengine.APIService, app string) (string, error) {
	got, err := svc.Apps.Get(app).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	if got.DefaultHostname != "" {
		return got.DefaultHostname, nil
	}
	return app + ".appspot.com", nil
}

func (g *appCmds) runAppSvcBrowse(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	host, err := appDefaultHostname(ctx, svc, app)
	if err != nil {
		return err
	}
	for _, service := range args {
		url := serviceURL(host, service, g.flagAppVersion)
		fmt.Fprintln(g.inv.stdout, url)
	}
	return nil
}

// serviceURL builds the URL for a service (and optional version) using the
// dot-prefixed convention (VERSION-dot-SERVICE-dot-APP).
func serviceURL(hostname, service, version string) string {
	parts := []string{}
	if version != "" {
		parts = append(parts, version)
	}
	if service != "" && service != "default" {
		parts = append(parts, service)
	}
	if len(parts) == 0 {
		return "https://" + hostname
	}
	return "https://" + strings.Join(parts, "-dot-") + "-dot-" + hostname
}

func (g *appCmds) runAppSvcDelete(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	for _, service := range args {
		if g.flagAppVersion != "" {
			op, err := svc.Apps.Services.Versions.Delete(app, service, g.flagAppVersion).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("deleting version %s/%s: %w", service, g.flagAppVersion, err)
			}
			if err := g.appFinishOp(ctx, svc, app, op, "Delete version", service+"/"+g.flagAppVersion); err != nil {
				return err
			}
			continue
		}
		op, err := svc.Apps.Services.Delete(app, service).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting service %s: %w", service, err)
		}
		if err := g.appFinishOp(ctx, svc, app, op, "Delete service", service); err != nil {
			return err
		}
	}
	return nil
}

func (g *appCmds) runAppSvcDescribe(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Services.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppSvcList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	all, err := appListServices(ctx, svc, app)
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-24s %-16s %s\n", "SERVICE", "NUM_VERSIONS", "TRAFFIC")
	for _, s := range all {
		versions := 0
		if vers, err := appListVersions(ctx, svc, app, s.Id); err == nil {
			versions = len(vers)
		}
		split := ""
		if s.Split != nil && len(s.Split.Allocations) > 0 {
			keys := make([]string, 0, len(s.Split.Allocations))
			for k := range s.Split.Allocations {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var kv []string
			for _, k := range keys {
				kv = append(kv, fmt.Sprintf("%s=%g", k, s.Split.Allocations[k]))
			}
			split = strings.Join(kv, ",")
		}
		fmt.Fprintf(g.inv.stdout, "%-24s %-16d %s\n", s.Id, versions, split)
	}
	return nil
}

func (g *appCmds) runAppSvcSetTraffic(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	if len(g.flagAppSplits) == 0 {
		return fmt.Errorf("--splits is required")
	}
	shardBy := strings.ToUpper(g.flagAppSplitBy)
	switch shardBy {
	case "COOKIE", "IP", "RANDOM":
	default:
		return fmt.Errorf("--split-by must be cookie, ip, or random")
	}
	// Normalize weights so they sum to 1.
	sum := 0.0
	weights := make(map[string]float64, len(g.flagAppSplits))
	for k, v := range g.flagAppSplits {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("invalid weight for version %s: %w", k, err)
		}
		if f < 0 {
			return fmt.Errorf("weight for version %s cannot be negative", k)
		}
		weights[k] = f
		sum += f
	}
	if sum == 0 {
		return fmt.Errorf("total split weight must be > 0")
	}
	allocations := make(map[string]float64, len(weights))
	for k, w := range weights {
		allocations[k] = w / sum
	}
	body := &appengine.Service{Split: &appengine.TrafficSplit{Allocations: allocations, ShardBy: shardBy}}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	// If no services listed, apply to every service.
	services := args
	if len(services) == 0 {
		got, err := appListServices(ctx, svc, app)
		if err != nil {
			return fmt.Errorf("listing services: %w", err)
		}
		for _, s := range got {
			services = append(services, s.Id)
		}
	}
	for _, s := range services {
		call := svc.Apps.Services.Patch(app, s, body).UpdateMask("split").Context(ctx)
		if g.flagAppMigrate {
			call = call.MigrateTraffic(true)
		}
		op, err := call.Do()
		if err != nil {
			return fmt.Errorf("setting traffic on %s: %w", s, err)
		}
		if err := g.appFinishOp(ctx, svc, app, op, "Set traffic", s); err != nil {
			return err
		}
	}
	return nil
}

func (g *appCmds) runAppSvcUpdate(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	if g.flagAppIngress == "" {
		return fmt.Errorf("no update options specified (try --ingress)")
	}
	ingress := ""
	switch strings.ToLower(g.flagAppIngress) {
	case "all":
		ingress = "INGRESS_TRAFFIC_ALLOWED_ALL"
	case "internal-only":
		ingress = "INGRESS_TRAFFIC_ALLOWED_INTERNAL_ONLY"
	case "internal-and-cloud-load-balancing":
		ingress = "INGRESS_TRAFFIC_ALLOWED_INTERNAL_AND_LB"
	default:
		return fmt.Errorf("--ingress must be all, internal-only, or internal-and-cloud-load-balancing")
	}
	body := &appengine.Service{NetworkSettings: &appengine.NetworkSettings{IngressTrafficAllowed: ingress}}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Apps.Services.Patch(app, args[0], body).UpdateMask("networkSettings").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return g.appFinishOp(ctx, svc, app, op, "Update service", args[0])
}

// -----------------------------------------------------------------------------
// ssl-certificates implementations
// -----------------------------------------------------------------------------

func readFileOrError(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	return string(data), nil
}

func (g *appCmds) runAppSSLCreate(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	cert, err := readFileOrError(g.flagAppCertificateFile)
	if err != nil {
		return err
	}
	key, err := readFileOrError(g.flagAppPrivateKeyFile)
	if err != nil {
		return err
	}
	body := &appengine.AuthorizedCertificate{
		DisplayName: g.flagAppDisplayName,
		CertificateRawData: &appengine.CertificateRawData{
			PublicCertificate: cert,
			PrivateKey:        key,
		},
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.AuthorizedCertificates.Create(app, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating SSL certificate: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Created SSL certificate [%s].\n", got.Id)
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppSSLDelete(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Apps.AuthorizedCertificates.Delete(app, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting SSL certificate: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Deleted SSL certificate [%s].\n", args[0])
	return nil
}

func (g *appCmds) runAppSSLDescribe(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.AuthorizedCertificates.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing SSL certificate: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppSSLList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	var all []*appengine.AuthorizedCertificate
	pageToken := ""
	for {
		call := svc.Apps.AuthorizedCertificates.List(app).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing SSL certificates: %w", err)
		}
		all = append(all, resp.Certificates...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-16s %-32s %s\n", "ID", "DISPLAY_NAME", "DOMAINS")
	for _, c := range all {
		fmt.Fprintf(g.inv.stdout, "%-16s %-32s %s\n", c.Id, c.DisplayName, strings.Join(c.DomainNames, ","))
	}
	return nil
}

func (g *appCmds) runAppSSLUpdate(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	body := &appengine.AuthorizedCertificate{}
	var masks []string
	if g.flagAppDisplayName != "" {
		body.DisplayName = g.flagAppDisplayName
		masks = append(masks, "displayName")
	}
	if g.flagAppCertificateFile != "" || g.flagAppPrivateKeyFile != "" {
		raw := &appengine.CertificateRawData{}
		if g.flagAppCertificateFile != "" {
			cert, err := readFileOrError(g.flagAppCertificateFile)
			if err != nil {
				return err
			}
			raw.PublicCertificate = cert
		}
		if g.flagAppPrivateKeyFile != "" {
			key, err := readFileOrError(g.flagAppPrivateKeyFile)
			if err != nil {
				return err
			}
			raw.PrivateKey = key
		}
		body.CertificateRawData = raw
		masks = append(masks, "certificateRawData")
	}
	if len(masks) == 0 {
		return fmt.Errorf("no update options specified")
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Apps.AuthorizedCertificates.Patch(app, args[0], body).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating SSL certificate: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Updated SSL certificate [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

// -----------------------------------------------------------------------------
// versions implementations
// -----------------------------------------------------------------------------

func (g *appCmds) runAppVerBrowse(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	host, err := appDefaultHostname(ctx, svc, app)
	if err != nil {
		return err
	}
	service := g.flagAppService
	if service == "" {
		service = "default"
	}
	for _, v := range args {
		fmt.Fprintln(g.inv.stdout, serviceURL(host, service, v))
	}
	return nil
}

func (g *appCmds) runAppVerDelete(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	service := g.flagAppService
	if service == "" {
		service = "default"
	}
	for _, v := range args {
		op, err := svc.Apps.Services.Versions.Delete(app, service, v).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting version %s: %w", v, err)
		}
		if err := g.appFinishOp(ctx, svc, app, op, "Delete version", service+"/"+v); err != nil {
			return err
		}
	}
	return nil
}

func (g *appCmds) runAppVerDescribe(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	service := g.flagAppService
	if service == "" {
		service = "default"
	}
	got, err := svc.Apps.Services.Versions.Get(app, service, args[0]).View("FULL").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing version: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAppFormat)
}

func (g *appCmds) runAppVerList(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	services, err := appListServices(ctx, svc, app)
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	type verRow struct {
		Service      string
		Version      *appengine.Version
		TrafficSplit float64
	}
	var rows []verRow
	for _, s := range services {
		if g.flagAppService != "" && s.Id != g.flagAppService {
			continue
		}
		versions, err := appListVersions(ctx, svc, app, s.Id)
		if err != nil {
			return fmt.Errorf("listing versions for %s: %w", s.Id, err)
		}
		for _, v := range versions {
			split := 0.0
			if s.Split != nil {
				split = s.Split.Allocations[v.Id]
			}
			if g.flagAppHideNoTraffic && split == 0 {
				continue
			}
			rows = append(rows, verRow{Service: s.Id, Version: v, TrafficSplit: split})
		}
	}
	if g.flagAppFormat != "" {
		return emitFormattedTo(g.inv.stdout, rows, g.flagAppFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(rows), "%-20s %-20s %-8s %-24s %s\n", "SERVICE", "VERSION.ID", "TRAFFIC", "LAST_DEPLOYED", "SERVING_STATUS")
	for _, r := range rows {
		fmt.Fprintf(g.inv.stdout, "%-20s %-20s %-8.2f %-24s %s\n", r.Service, r.Version.Id, r.TrafficSplit, r.Version.CreateTime, r.Version.ServingStatus)
	}
	return nil
}

func (g *appCmds) runAppVerMigrate(cmd *cobra.Command, args []string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	service := g.flagAppService
	if service == "" {
		service = "default"
	}
	body := &appengine.Service{Split: &appengine.TrafficSplit{
		Allocations: map[string]float64{args[0]: 1.0},
		ShardBy:     "IP",
	}}
	op, err := svc.Apps.Services.Patch(app, service, body).UpdateMask("split").MigrateTraffic(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating traffic to %s: %w", args[0], err)
	}
	return g.appFinishOp(ctx, svc, app, op, "Migrate traffic", service+"/"+args[0])
}

func (g *appCmds) runAppVerStart(cmd *cobra.Command, args []string) error {
	return g.appSetServingStatus(cmd, args, "SERVING", "Start")
}

func (g *appCmds) runAppVerStop(cmd *cobra.Command, args []string) error {
	return g.appSetServingStatus(cmd, args, "STOPPED", "Stop")
}

func (g *appCmds) appSetServingStatus(cmd *cobra.Command, args []string, status, verb string) error {
	app, err := g.appResolveApp()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AppEngine(ctx)
	if err != nil {
		return err
	}
	service := g.flagAppService
	if service == "" {
		service = "default"
	}
	body := &appengine.Version{ServingStatus: status}
	for _, v := range args {
		op, err := svc.Apps.Services.Versions.Patch(app, service, v, body).UpdateMask("servingStatus").Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("%s version %s: %w", strings.ToLower(verb), v, err)
		}
		if err := g.appFinishOp(ctx, svc, app, op, verb+" version", service+"/"+v); err != nil {
			return err
		}
	}
	return nil
}

// appCmds holds the commands and flag values of one invocation's "app"
// command group, so invocations don't share state (#2070).
type appCmds struct {
	inv *invocation

	appCmd                 *cobra.Command
	flagAppFormat          string
	flagAppFilter          string
	flagAppLimit           int64
	flagAppService         string
	flagAppVersion         string
	flagAppLevel           string
	flagAppLogs            string
	flagAppPending         bool
	flagAppHideNoTraffic   bool
	flagAppEnvironment     string
	flagAppCertificateID   string
	flagAppNoCertificateID bool
	flagAppCertificateMgmt string
	flagAppSourceRange     string
	flagAppAction          string
	flagAppDescription     string
	flagAppDisplayName     string
	flagAppCertificateFile string
	flagAppPrivateKeyFile  string
	flagAppSplits          map[string]string
	flagAppSplitBy         string
	flagAppMigrate         bool
	flagAppLaunchBrowser   bool
	flagAppAsync           bool
	flagAppMatchingAddress string
	appDomainMappingsCmd   *cobra.Command
	appDomMapCreateCmd     *cobra.Command
	appDomMapDeleteCmd     *cobra.Command
	appDomMapDescribeCmd   *cobra.Command
	appDomMapListCmd       *cobra.Command
	appDomMapUpdateCmd     *cobra.Command
	appFirewallRulesCmd    *cobra.Command
	appFwCreateCmd         *cobra.Command
	appFwDeleteCmd         *cobra.Command
	appFwDescribeCmd       *cobra.Command
	appFwListCmd           *cobra.Command
	appFwUpdateCmd         *cobra.Command
	appFwTestIPCmd         *cobra.Command
	appInstancesCmd        *cobra.Command
	appInstDeleteCmd       *cobra.Command
	appInstDescribeCmd     *cobra.Command
	appInstEnableDebugCmd  *cobra.Command
	appInstDisableDebugCmd *cobra.Command
	appInstListCmd         *cobra.Command
	appLogsCmd             *cobra.Command
	appLogsReadCmd         *cobra.Command
	appLogsTailCmd         *cobra.Command
	appOperationsCmd       *cobra.Command
	appOpDescribeCmd       *cobra.Command
	appOpListCmd           *cobra.Command
	appOpWaitCmd           *cobra.Command
	appRegionsCmd          *cobra.Command
	appRegionsListCmd      *cobra.Command
	appRuntimesCmd         *cobra.Command
	appRuntimesListCmd     *cobra.Command
	appServicesCmd         *cobra.Command
	appSvcBrowseCmd        *cobra.Command
	appSvcDeleteCmd        *cobra.Command
	appSvcDescribeCmd      *cobra.Command
	appSvcListCmd          *cobra.Command
	appSvcSetTrafficCmd    *cobra.Command
	appSvcUpdateCmd        *cobra.Command
	appSSLCertsCmd         *cobra.Command
	appSSLCreateCmd        *cobra.Command
	appSSLDeleteCmd        *cobra.Command
	appSSLDescribeCmd      *cobra.Command
	appSSLListCmd          *cobra.Command
	appSSLUpdateCmd        *cobra.Command
	appVersionsCmd         *cobra.Command
	appVerBrowseCmd        *cobra.Command
	appVerDeleteCmd        *cobra.Command
	appVerDescribeCmd      *cobra.Command
	appVerListCmd          *cobra.Command
	appVerMigrateCmd       *cobra.Command
	appVerStartCmd         *cobra.Command
	appVerStopCmd          *cobra.Command
	flagAppIngress         string
}

// newAppCmd returns the "app" command group for inv.
func newAppCmd(inv *invocation) *cobra.Command {
	return newAppCmds(inv).appCmd
}

// newAppCmds builds the "app" command group for inv and returns its state.
func newAppCmds(inv *invocation) *appCmds {
	g := &appCmds{inv: inv}
	g.appCmd = &cobra.Command{Use: "app", Short: "Manage App Engine"}
	g.appDomainMappingsCmd = &cobra.Command{Use: "domain-mappings", Short: "Manage App Engine domain mappings"}
	g.appDomMapCreateCmd = &cobra.Command{
		Use: "create DOMAIN", Short: "Create a domain mapping",
		Args: cobra.ExactArgs(1), RunE: g.runAppDomMapCreate,
	}
	g.appDomMapDeleteCmd = &cobra.Command{
		Use: "delete DOMAIN", Short: "Delete a domain mapping",
		Args: cobra.ExactArgs(1), RunE: g.runAppDomMapDelete,
	}
	g.appDomMapDescribeCmd = &cobra.Command{
		Use: "describe DOMAIN", Short: "Describe a domain mapping",
		Args: cobra.ExactArgs(1), RunE: g.runAppDomMapDescribe,
	}
	g.appDomMapListCmd = &cobra.Command{
		Use: "list", Short: "List domain mappings",
		Args: cobra.NoArgs, RunE: g.runAppDomMapList,
	}
	g.appDomMapUpdateCmd = &cobra.Command{
		Use: "update DOMAIN", Short: "Update a domain mapping",
		Args: cobra.ExactArgs(1), RunE: g.runAppDomMapUpdate,
	}
	g.appFirewallRulesCmd = &cobra.Command{Use: "firewall-rules", Short: "Manage App Engine firewall rules"}
	g.appFwCreateCmd = &cobra.Command{
		Use: "create PRIORITY", Short: "Create a firewall rule",
		Args: cobra.ExactArgs(1), RunE: g.runAppFwCreate,
	}
	g.appFwDeleteCmd = &cobra.Command{
		Use: "delete PRIORITY", Short: "Delete a firewall rule",
		Args: cobra.ExactArgs(1), RunE: g.runAppFwDelete,
	}
	g.appFwDescribeCmd = &cobra.Command{
		Use: "describe PRIORITY", Short: "Describe a firewall rule",
		Args: cobra.ExactArgs(1), RunE: g.runAppFwDescribe,
	}
	g.appFwListCmd = &cobra.Command{
		Use: "list", Short: "List firewall rules",
		Args: cobra.NoArgs, RunE: g.runAppFwList,
	}
	g.appFwUpdateCmd = &cobra.Command{
		Use: "update PRIORITY", Short: "Update a firewall rule",
		Args: cobra.ExactArgs(1), RunE: g.runAppFwUpdate,
	}
	g.appFwTestIPCmd = &cobra.Command{
		Use: "test-ip IP", Short: "Display firewall rules that match a given IP",
		Args: cobra.ExactArgs(1), RunE: g.runAppFwTestIP,
	}
	g.appInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage App Engine instances"}
	g.appInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an instance",
		Args: cobra.ExactArgs(1), RunE: g.runAppInstDelete,
	}
	g.appInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an instance",
		Args: cobra.ExactArgs(1), RunE: g.runAppInstDescribe,
	}
	g.appInstEnableDebugCmd = &cobra.Command{
		Use: "enable-debug INSTANCE", Short: "Enable debug mode on a flexible-environment instance",
		Args: cobra.ExactArgs(1), RunE: g.runAppInstEnableDebug,
	}
	g.appInstDisableDebugCmd = &cobra.Command{
		Use: "disable-debug INSTANCE", Short: "Disable debug mode on a flexible-environment instance",
		Args: cobra.ExactArgs(1), RunE: runAppInstDisableDebug,
	}
	g.appInstListCmd = &cobra.Command{
		Use: "list", Short: "List instances",
		Args: cobra.NoArgs, RunE: g.runAppInstList,
	}
	g.appLogsCmd = &cobra.Command{Use: "logs", Short: "Read App Engine app logs"}
	g.appLogsReadCmd = &cobra.Command{
		Use: "read", Short: "Read the latest App Engine log entries",
		Args: cobra.NoArgs, RunE: g.runAppLogsRead,
	}
	g.appLogsTailCmd = &cobra.Command{
		Use: "tail", Short: "Tail App Engine log entries (best-effort)",
		Args: cobra.NoArgs, RunE: g.runAppLogsTail,
	}
	g.appOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage App Engine operations"}
	g.appOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: g.runAppOpDescribe,
	}
	g.appOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations",
		Args: cobra.NoArgs, RunE: g.runAppOpList,
	}
	g.appOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Poll an operation until completion",
		Args: cobra.ExactArgs(1), RunE: g.runAppOpWait,
	}
	g.appRegionsCmd = &cobra.Command{Use: "regions", Short: "View regional availability for App Engine"}
	g.appRegionsListCmd = &cobra.Command{
		Use: "list", Short: "List App Engine regions",
		Args: cobra.NoArgs, RunE: g.runAppRegionsList,
	}
	g.appRuntimesCmd = &cobra.Command{Use: "runtimes", Short: "View runtimes available to App Engine"}
	g.appRuntimesListCmd = &cobra.Command{
		Use: "list", Short: "List App Engine runtimes",
		Args: cobra.NoArgs, RunE: g.runAppRuntimesList,
	}
	g.appServicesCmd = &cobra.Command{Use: "services", Short: "Manage App Engine services"}
	g.appSvcBrowseCmd = &cobra.Command{
		Use: "browse SERVICE...", Short: "Print URLs that open a service in a browser",
		Args: cobra.MinimumNArgs(1), RunE: g.runAppSvcBrowse,
	}
	g.appSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE...", Short: "Delete one or more services (or one version of a service)",
		Args: cobra.MinimumNArgs(1), RunE: g.runAppSvcDelete,
	}
	g.appSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: g.runAppSvcDescribe,
	}
	g.appSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services",
		Args: cobra.NoArgs, RunE: g.runAppSvcList,
	}
	g.appSvcSetTrafficCmd = &cobra.Command{
		Use: "set-traffic [SERVICE...]", Short: "Set traffic splits across versions of one or more services",
		Args: cobra.ArbitraryArgs, RunE: g.runAppSvcSetTraffic,
	}
	g.appSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service",
		Args: cobra.ExactArgs(1), RunE: g.runAppSvcUpdate,
	}
	g.appSSLCertsCmd = &cobra.Command{Use: "ssl-certificates", Short: "Manage App Engine SSL certificates"}
	g.appSSLCreateCmd = &cobra.Command{
		Use: "create", Short: "Upload a new SSL certificate",
		Args: cobra.NoArgs, RunE: g.runAppSSLCreate,
	}
	g.appSSLDeleteCmd = &cobra.Command{
		Use: "delete CERT_ID", Short: "Delete an SSL certificate",
		Args: cobra.ExactArgs(1), RunE: g.runAppSSLDelete,
	}
	g.appSSLDescribeCmd = &cobra.Command{
		Use: "describe CERT_ID", Short: "Describe an SSL certificate",
		Args: cobra.ExactArgs(1), RunE: g.runAppSSLDescribe,
	}
	g.appSSLListCmd = &cobra.Command{
		Use: "list", Short: "List SSL certificates",
		Args: cobra.NoArgs, RunE: g.runAppSSLList,
	}
	g.appSSLUpdateCmd = &cobra.Command{
		Use: "update CERT_ID", Short: "Update an SSL certificate",
		Args: cobra.ExactArgs(1), RunE: g.runAppSSLUpdate,
	}
	g.appVersionsCmd = &cobra.Command{Use: "versions", Short: "Manage App Engine versions"}
	g.appVerBrowseCmd = &cobra.Command{
		Use: "browse VERSION...", Short: "Print URLs that open specific versions in a browser",
		Args: cobra.MinimumNArgs(1), RunE: g.runAppVerBrowse,
	}
	g.appVerDeleteCmd = &cobra.Command{
		Use: "delete VERSION...", Short: "Delete versions",
		Args: cobra.MinimumNArgs(1), RunE: g.runAppVerDelete,
	}
	g.appVerDescribeCmd = &cobra.Command{
		Use: "describe VERSION", Short: "Describe a version",
		Args: cobra.ExactArgs(1), RunE: g.runAppVerDescribe,
	}
	g.appVerListCmd = &cobra.Command{
		Use: "list", Short: "List versions",
		Args: cobra.NoArgs, RunE: g.runAppVerList,
	}
	g.appVerMigrateCmd = &cobra.Command{
		Use: "migrate VERSION", Short: "Migrate 100% of traffic to a version",
		Args: cobra.ExactArgs(1), RunE: g.runAppVerMigrate,
	}
	g.appVerStartCmd = &cobra.Command{
		Use: "start VERSION...", Short: "Start serving specified versions",
		Args: cobra.MinimumNArgs(1), RunE: g.runAppVerStart,
	}
	g.appVerStopCmd = &cobra.Command{
		Use: "stop VERSION...", Short: "Stop serving specified versions",
		Args: cobra.MinimumNArgs(1), RunE: g.runAppVerStop,
	}
	g.initApp()
	return g
}
