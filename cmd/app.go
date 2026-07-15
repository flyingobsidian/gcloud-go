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

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	appengine "google.golang.org/api/appengine/v1"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud app (#299, #865-#874) ---

var appCmd = &cobra.Command{Use: "app", Short: "Manage App Engine"}

// Shared flags for the app subcommands.
var (
	flagAppFormat               string
	flagAppFilter               string
	flagAppLimit                int64
	flagAppService              string
	flagAppVersion              string
	flagAppLevel                string
	flagAppLogs                 string
	flagAppPending              bool
	flagAppHideNoTraffic        bool
	flagAppEnvironment          string
	flagAppCertificateID        string
	flagAppNoCertificateID      bool
	flagAppCertificateMgmt      string
	flagAppSourceRange          string
	flagAppAction               string
	flagAppDescription          string
	flagAppDomainDisplayName    string
	flagAppDisplayName          string
	flagAppCertificateFile      string
	flagAppPrivateKeyFile       string
	flagAppSplits               map[string]string
	flagAppSplitBy              string
	flagAppMigrate              bool
	flagAppLaunchBrowser        bool
	flagAppAsync                bool
	flagAppMatchingAddress      string
)

// appResolveApp returns the App Engine application id (the project) and validates a project
// has been selected.
func appResolveApp() (string, error) {
	return resolveProject()
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

func appFinishOp(ctx context.Context, svc *appengine.APIService, app string, op *appengine.Operation, verb, name string) error {
	if flagAppAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := appWaitOp(ctx, svc, app, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// --- domain-mappings ---

var appDomainMappingsCmd = &cobra.Command{Use: "domain-mappings", Short: "Manage App Engine domain mappings"}

var (
	appDomMapCreateCmd = &cobra.Command{
		Use: "create DOMAIN", Short: "Create a domain mapping",
		Args: cobra.ExactArgs(1), RunE: runAppDomMapCreate,
	}
	appDomMapDeleteCmd = &cobra.Command{
		Use: "delete DOMAIN", Short: "Delete a domain mapping",
		Args: cobra.ExactArgs(1), RunE: runAppDomMapDelete,
	}
	appDomMapDescribeCmd = &cobra.Command{
		Use: "describe DOMAIN", Short: "Describe a domain mapping",
		Args: cobra.ExactArgs(1), RunE: runAppDomMapDescribe,
	}
	appDomMapListCmd = &cobra.Command{
		Use: "list", Short: "List domain mappings",
		Args: cobra.NoArgs, RunE: runAppDomMapList,
	}
	appDomMapUpdateCmd = &cobra.Command{
		Use: "update DOMAIN", Short: "Update a domain mapping",
		Args: cobra.ExactArgs(1), RunE: runAppDomMapUpdate,
	}
)

// --- firewall-rules ---

var appFirewallRulesCmd = &cobra.Command{Use: "firewall-rules", Short: "Manage App Engine firewall rules"}

var (
	appFwCreateCmd = &cobra.Command{
		Use: "create PRIORITY", Short: "Create a firewall rule",
		Args: cobra.ExactArgs(1), RunE: runAppFwCreate,
	}
	appFwDeleteCmd = &cobra.Command{
		Use: "delete PRIORITY", Short: "Delete a firewall rule",
		Args: cobra.ExactArgs(1), RunE: runAppFwDelete,
	}
	appFwDescribeCmd = &cobra.Command{
		Use: "describe PRIORITY", Short: "Describe a firewall rule",
		Args: cobra.ExactArgs(1), RunE: runAppFwDescribe,
	}
	appFwListCmd = &cobra.Command{
		Use: "list", Short: "List firewall rules",
		Args: cobra.NoArgs, RunE: runAppFwList,
	}
	appFwUpdateCmd = &cobra.Command{
		Use: "update PRIORITY", Short: "Update a firewall rule",
		Args: cobra.ExactArgs(1), RunE: runAppFwUpdate,
	}
	appFwTestIPCmd = &cobra.Command{
		Use: "test-ip IP", Short: "Display firewall rules that match a given IP",
		Args: cobra.ExactArgs(1), RunE: runAppFwTestIP,
	}
)

// --- instances ---

var appInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage App Engine instances"}

var (
	appInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete an instance",
		Args: cobra.ExactArgs(1), RunE: runAppInstDelete,
	}
	appInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe an instance",
		Args: cobra.ExactArgs(1), RunE: runAppInstDescribe,
	}
	appInstEnableDebugCmd = &cobra.Command{
		Use: "enable-debug INSTANCE", Short: "Enable debug mode on a flexible-environment instance",
		Args: cobra.ExactArgs(1), RunE: runAppInstEnableDebug,
	}
	appInstDisableDebugCmd = &cobra.Command{
		Use: "disable-debug INSTANCE", Short: "Disable debug mode on a flexible-environment instance",
		Args: cobra.ExactArgs(1), RunE: runAppInstDisableDebug,
	}
	appInstListCmd = &cobra.Command{
		Use: "list", Short: "List instances",
		Args: cobra.NoArgs, RunE: runAppInstList,
	}
)

// --- logs ---

var appLogsCmd = &cobra.Command{Use: "logs", Short: "Read App Engine app logs"}

var (
	appLogsReadCmd = &cobra.Command{
		Use: "read", Short: "Read the latest App Engine log entries",
		Args: cobra.NoArgs, RunE: runAppLogsRead,
	}
	appLogsTailCmd = &cobra.Command{
		Use: "tail", Short: "Tail App Engine log entries (best-effort)",
		Args: cobra.NoArgs, RunE: runAppLogsTail,
	}
)

// --- operations ---

var appOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage App Engine operations"}

var (
	appOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runAppOpDescribe,
	}
	appOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations",
		Args: cobra.NoArgs, RunE: runAppOpList,
	}
	appOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Poll an operation until completion",
		Args: cobra.ExactArgs(1), RunE: runAppOpWait,
	}
)

// --- regions ---

var appRegionsCmd = &cobra.Command{Use: "regions", Short: "View regional availability for App Engine"}

var (
	appRegionsListCmd = &cobra.Command{
		Use: "list", Short: "List App Engine regions",
		Args: cobra.NoArgs, RunE: runAppRegionsList,
	}
)

// --- runtimes ---

var appRuntimesCmd = &cobra.Command{Use: "runtimes", Short: "View runtimes available to App Engine"}

var (
	appRuntimesListCmd = &cobra.Command{
		Use: "list", Short: "List App Engine runtimes",
		Args: cobra.NoArgs, RunE: runAppRuntimesList,
	}
)

// --- services ---

var appServicesCmd = &cobra.Command{Use: "services", Short: "Manage App Engine services"}

var (
	appSvcBrowseCmd = &cobra.Command{
		Use: "browse SERVICE...", Short: "Print URLs that open a service in a browser",
		Args: cobra.MinimumNArgs(1), RunE: runAppSvcBrowse,
	}
	appSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE...", Short: "Delete one or more services (or one version of a service)",
		Args: cobra.MinimumNArgs(1), RunE: runAppSvcDelete,
	}
	appSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: runAppSvcDescribe,
	}
	appSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services",
		Args: cobra.NoArgs, RunE: runAppSvcList,
	}
	appSvcSetTrafficCmd = &cobra.Command{
		Use: "set-traffic [SERVICE...]", Short: "Set traffic splits across versions of one or more services",
		Args: cobra.ArbitraryArgs, RunE: runAppSvcSetTraffic,
	}
	appSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service",
		Args: cobra.ExactArgs(1), RunE: runAppSvcUpdate,
	}
)

// --- ssl-certificates ---

var appSSLCertsCmd = &cobra.Command{Use: "ssl-certificates", Short: "Manage App Engine SSL certificates"}

var (
	appSSLCreateCmd = &cobra.Command{
		Use: "create", Short: "Upload a new SSL certificate",
		Args: cobra.NoArgs, RunE: runAppSSLCreate,
	}
	appSSLDeleteCmd = &cobra.Command{
		Use: "delete CERT_ID", Short: "Delete an SSL certificate",
		Args: cobra.ExactArgs(1), RunE: runAppSSLDelete,
	}
	appSSLDescribeCmd = &cobra.Command{
		Use: "describe CERT_ID", Short: "Describe an SSL certificate",
		Args: cobra.ExactArgs(1), RunE: runAppSSLDescribe,
	}
	appSSLListCmd = &cobra.Command{
		Use: "list", Short: "List SSL certificates",
		Args: cobra.NoArgs, RunE: runAppSSLList,
	}
	appSSLUpdateCmd = &cobra.Command{
		Use: "update CERT_ID", Short: "Update an SSL certificate",
		Args: cobra.ExactArgs(1), RunE: runAppSSLUpdate,
	}
)

// --- versions ---

var appVersionsCmd = &cobra.Command{Use: "versions", Short: "Manage App Engine versions"}

var (
	appVerBrowseCmd = &cobra.Command{
		Use: "browse VERSION...", Short: "Print URLs that open specific versions in a browser",
		Args: cobra.MinimumNArgs(1), RunE: runAppVerBrowse,
	}
	appVerDeleteCmd = &cobra.Command{
		Use: "delete VERSION...", Short: "Delete versions",
		Args: cobra.MinimumNArgs(1), RunE: runAppVerDelete,
	}
	appVerDescribeCmd = &cobra.Command{
		Use: "describe VERSION", Short: "Describe a version",
		Args: cobra.ExactArgs(1), RunE: runAppVerDescribe,
	}
	appVerListCmd = &cobra.Command{
		Use: "list", Short: "List versions",
		Args: cobra.NoArgs, RunE: runAppVerList,
	}
	appVerMigrateCmd = &cobra.Command{
		Use: "migrate VERSION", Short: "Migrate 100% of traffic to a version",
		Args: cobra.ExactArgs(1), RunE: runAppVerMigrate,
	}
	appVerStartCmd = &cobra.Command{
		Use: "start VERSION...", Short: "Start serving specified versions",
		Args: cobra.MinimumNArgs(1), RunE: runAppVerStart,
	}
	appVerStopCmd = &cobra.Command{
		Use: "stop VERSION...", Short: "Stop serving specified versions",
		Args: cobra.MinimumNArgs(1), RunE: runAppVerStop,
	}
)

func init() {
	// --- domain-mappings ---
	for _, c := range []*cobra.Command{appDomMapCreateCmd, appDomMapDeleteCmd, appDomMapDescribeCmd, appDomMapListCmd, appDomMapUpdateCmd} {
		c.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
		c.Flags().BoolVar(&flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	}
	appDomMapCreateCmd.Flags().StringVar(&flagAppCertificateID, "certificate-id", "", "Manually managed certificate id to associate with this domain")
	appDomMapCreateCmd.Flags().StringVar(&flagAppCertificateMgmt, "certificate-management", "", "Certificate management type: automatic or manual")
	appDomMapUpdateCmd.Flags().StringVar(&flagAppCertificateID, "certificate-id", "", "Certificate id to associate with this domain")
	appDomMapUpdateCmd.Flags().BoolVar(&flagAppNoCertificateID, "no-certificate-id", false, "Remove certificate association from this domain")
	appDomMapUpdateCmd.Flags().StringVar(&flagAppCertificateMgmt, "certificate-management", "", "Certificate management type: automatic or manual")
	for _, c := range []*cobra.Command{appDomMapCreateCmd, appDomMapDeleteCmd, appDomMapDescribeCmd, appDomMapListCmd, appDomMapUpdateCmd} {
		appDomainMappingsCmd.AddCommand(c)
	}
	appCmd.AddCommand(appDomainMappingsCmd)

	// --- firewall-rules ---
	for _, c := range []*cobra.Command{appFwCreateCmd, appFwDeleteCmd, appFwDescribeCmd, appFwListCmd, appFwUpdateCmd, appFwTestIPCmd} {
		c.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	}
	appFwCreateCmd.Flags().StringVar(&flagAppSourceRange, "source-range", "", "IP address or CIDR range (required)")
	appFwCreateCmd.Flags().StringVar(&flagAppAction, "action", "", "ALLOW or DENY (required)")
	appFwCreateCmd.Flags().StringVar(&flagAppDescription, "description", "", "A text description of the rule")
	_ = appFwCreateCmd.MarkFlagRequired("source-range")
	_ = appFwCreateCmd.MarkFlagRequired("action")
	appFwUpdateCmd.Flags().StringVar(&flagAppSourceRange, "source-range", "", "IP address or CIDR range")
	appFwUpdateCmd.Flags().StringVar(&flagAppAction, "action", "", "ALLOW or DENY")
	appFwUpdateCmd.Flags().StringVar(&flagAppDescription, "description", "", "A text description of the rule")
	appFwListCmd.Flags().StringVar(&flagAppMatchingAddress, "matching-address", "", "Only return rules matching this IP address")
	for _, c := range []*cobra.Command{appFwCreateCmd, appFwDeleteCmd, appFwDescribeCmd, appFwListCmd, appFwUpdateCmd, appFwTestIPCmd} {
		appFirewallRulesCmd.AddCommand(c)
	}
	appCmd.AddCommand(appFirewallRulesCmd)

	// --- instances ---
	for _, c := range []*cobra.Command{appInstDeleteCmd, appInstDescribeCmd, appInstEnableDebugCmd, appInstDisableDebugCmd, appInstListCmd} {
		c.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{appInstDeleteCmd, appInstDescribeCmd, appInstEnableDebugCmd, appInstDisableDebugCmd} {
		c.Flags().StringVarP(&flagAppService, "service", "s", "", "The service ID (required)")
		c.Flags().StringVarP(&flagAppVersion, "version", "v", "", "The version ID (required)")
		_ = c.MarkFlagRequired("service")
		_ = c.MarkFlagRequired("version")
	}
	for _, c := range []*cobra.Command{appInstDeleteCmd, appInstEnableDebugCmd, appInstDisableDebugCmd} {
		c.Flags().BoolVar(&flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	}
	appInstListCmd.Flags().StringVarP(&flagAppService, "service", "s", "", "Only show instances from this service")
	appInstListCmd.Flags().StringVarP(&flagAppVersion, "version", "v", "", "Only show instances from this version")
	for _, c := range []*cobra.Command{appInstDeleteCmd, appInstDescribeCmd, appInstEnableDebugCmd, appInstDisableDebugCmd, appInstListCmd} {
		appInstancesCmd.AddCommand(c)
	}
	appCmd.AddCommand(appInstancesCmd)

	// --- logs ---
	for _, c := range []*cobra.Command{appLogsReadCmd, appLogsTailCmd} {
		c.Flags().StringVarP(&flagAppService, "service", "s", "", "Only entries from this service")
		c.Flags().StringVarP(&flagAppVersion, "version", "v", "", "Only entries from this version")
		c.Flags().StringVar(&flagAppLevel, "level", "any", "Minimum severity: any, debug, info, notice, warning, error, critical, alert, emergency")
		c.Flags().StringVar(&flagAppLogs, "logs", "", "Comma-separated log ids to include (defaults to stdout,stderr,crash.log,nginx.request,request_log)")
		c.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	}
	appLogsReadCmd.Flags().Int64Var(&flagAppLimit, "limit", 200, "Number of log entries to show")
	appLogsTailCmd.Flags().Int64Var(&flagAppLimit, "limit", 200, "Number of log entries to show")
	appLogsCmd.AddCommand(appLogsReadCmd, appLogsTailCmd)
	appCmd.AddCommand(appLogsCmd)

	// --- operations ---
	for _, c := range []*cobra.Command{appOpDescribeCmd, appOpListCmd, appOpWaitCmd} {
		c.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	}
	appOpListCmd.Flags().BoolVar(&flagAppPending, "pending", false, "Only show pending (not done) operations")
	appOpListCmd.Flags().StringVar(&flagAppFilter, "filter", "", "Filter expression applied to the API")
	appOperationsCmd.AddCommand(appOpDescribeCmd, appOpListCmd, appOpWaitCmd)
	appCmd.AddCommand(appOperationsCmd)

	// --- regions ---
	appRegionsListCmd.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	appRegionsCmd.AddCommand(appRegionsListCmd)
	appCmd.AddCommand(appRegionsCmd)

	// --- runtimes ---
	appRuntimesListCmd.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	appRuntimesListCmd.Flags().StringVar(&flagAppEnvironment, "environment", "", "Only include runtimes in this environment: STANDARD or FLEXIBLE")
	appRuntimesCmd.AddCommand(appRuntimesListCmd)
	appCmd.AddCommand(appRuntimesCmd)

	// --- services ---
	for _, c := range []*cobra.Command{appSvcBrowseCmd, appSvcDeleteCmd, appSvcDescribeCmd, appSvcListCmd, appSvcSetTrafficCmd, appSvcUpdateCmd} {
		c.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	}
	appSvcBrowseCmd.Flags().StringVarP(&flagAppVersion, "version", "v", "", "Browse the specific version")
	appSvcBrowseCmd.Flags().BoolVar(&flagAppLaunchBrowser, "launch-browser", false, "Attempt to launch a browser (best-effort; prints URL either way)")
	appSvcDeleteCmd.Flags().StringVar(&flagAppVersion, "version", "", "Delete only this version instead of the entire service")
	appSvcDeleteCmd.Flags().BoolVar(&flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	appSvcSetTrafficCmd.Flags().StringToStringVar(&flagAppSplits, "splits", nil, "Version=weight pairs describing traffic allocation (required)")
	appSvcSetTrafficCmd.Flags().StringVar(&flagAppSplitBy, "split-by", "ip", "Split traffic by: cookie, ip, random")
	appSvcSetTrafficCmd.Flags().BoolVar(&flagAppMigrate, "migrate", false, "Gracefully migrate traffic to new versions")
	appSvcSetTrafficCmd.Flags().BoolVar(&flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	_ = appSvcSetTrafficCmd.MarkFlagRequired("splits")
	appSvcUpdateCmd.Flags().StringVar(&flagAppIngress, "ingress", "", "Network ingress setting: all, internal-only, internal-and-cloud-load-balancing")
	appSvcUpdateCmd.Flags().BoolVar(&flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	for _, c := range []*cobra.Command{appSvcBrowseCmd, appSvcDeleteCmd, appSvcDescribeCmd, appSvcListCmd, appSvcSetTrafficCmd, appSvcUpdateCmd} {
		appServicesCmd.AddCommand(c)
	}
	appCmd.AddCommand(appServicesCmd)

	// --- ssl-certificates ---
	for _, c := range []*cobra.Command{appSSLCreateCmd, appSSLDeleteCmd, appSSLDescribeCmd, appSSLListCmd, appSSLUpdateCmd} {
		c.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	}
	appSSLCreateCmd.Flags().StringVar(&flagAppDisplayName, "display-name", "", "Display name for this certificate (required)")
	appSSLCreateCmd.Flags().StringVar(&flagAppCertificateFile, "certificate", "", "Path to the PEM x.509 certificate file (required)")
	appSSLCreateCmd.Flags().StringVar(&flagAppPrivateKeyFile, "private-key", "", "Path to the PEM RSA private key file (required)")
	_ = appSSLCreateCmd.MarkFlagRequired("display-name")
	_ = appSSLCreateCmd.MarkFlagRequired("certificate")
	_ = appSSLCreateCmd.MarkFlagRequired("private-key")
	appSSLUpdateCmd.Flags().StringVar(&flagAppDisplayName, "display-name", "", "New display name for the certificate")
	appSSLUpdateCmd.Flags().StringVar(&flagAppCertificateFile, "certificate", "", "Path to a replacement PEM x.509 certificate")
	appSSLUpdateCmd.Flags().StringVar(&flagAppPrivateKeyFile, "private-key", "", "Path to a replacement PEM RSA private key")
	for _, c := range []*cobra.Command{appSSLCreateCmd, appSSLDeleteCmd, appSSLDescribeCmd, appSSLListCmd, appSSLUpdateCmd} {
		appSSLCertsCmd.AddCommand(c)
	}
	appCmd.AddCommand(appSSLCertsCmd)

	// --- versions ---
	for _, c := range []*cobra.Command{appVerBrowseCmd, appVerDeleteCmd, appVerDescribeCmd, appVerListCmd, appVerMigrateCmd, appVerStartCmd, appVerStopCmd} {
		c.Flags().StringVar(&flagAppFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{appVerBrowseCmd, appVerDeleteCmd, appVerDescribeCmd, appVerMigrateCmd, appVerStartCmd, appVerStopCmd} {
		c.Flags().StringVarP(&flagAppService, "service", "s", "", "Service that owns the version(s)")
	}
	appVerBrowseCmd.Flags().BoolVar(&flagAppLaunchBrowser, "launch-browser", false, "Attempt to launch a browser (best-effort; prints URL either way)")
	appVerListCmd.Flags().StringVarP(&flagAppService, "service", "s", "", "Only list versions from this service")
	appVerListCmd.Flags().BoolVar(&flagAppHideNoTraffic, "hide-no-traffic", false, "Only show versions that are receiving traffic")
	for _, c := range []*cobra.Command{appVerDeleteCmd, appVerMigrateCmd, appVerStartCmd, appVerStopCmd} {
		c.Flags().BoolVar(&flagAppAsync, "async", false, "Return immediately instead of waiting for the operation")
	}
	for _, c := range []*cobra.Command{appVerBrowseCmd, appVerDeleteCmd, appVerDescribeCmd, appVerListCmd, appVerMigrateCmd, appVerStartCmd, appVerStopCmd} {
		appVersionsCmd.AddCommand(c)
	}
	appCmd.AddCommand(appVersionsCmd)

	// Legacy stubs for top-level app commands still to be implemented under #299.
	for _, name := range []string{"browse", "create", "deploy", "describe", "open-console", "update"} {
		registerStubCommand(appCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(appCmd)
}

var flagAppIngress string

// -----------------------------------------------------------------------------
// domain-mappings implementations
// -----------------------------------------------------------------------------

func runAppDomMapCreate(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	mgmt := strings.ToLower(flagAppCertificateMgmt)
	if mgmt == "" {
		if flagAppCertificateID == "" {
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
	if flagAppCertificateID != "" || mgmt == "manual" {
		mapping.SslSettings = &appengine.SslSettings{
			SslManagementType: strings.ToUpper(mgmt),
			CertificateId:     flagAppCertificateID,
		}
	} else {
		mapping.SslSettings = &appengine.SslSettings{SslManagementType: "AUTOMATIC"}
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Apps.DomainMappings.Create(app, mapping).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating domain mapping: %w", err)
	}
	return appFinishOp(ctx, svc, app, op, "Create domain mapping", args[0])
}

func runAppDomMapDelete(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Apps.DomainMappings.Delete(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting domain mapping: %w", err)
	}
	return appFinishOp(ctx, svc, app, op, "Delete domain mapping", args[0])
}

func runAppDomMapDescribe(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.DomainMappings.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain mapping: %w", err)
	}
	return emitFormatted(got, flagAppFormat)
}

func runAppDomMapList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
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
	if flagAppFormat != "" {
		return emitFormatted(all, flagAppFormat)
	}
	fmt.Printf("%-40s %s\n", "ID", "SSL_CERTIFICATE_ID")
	for _, dm := range all {
		cert := ""
		if dm.SslSettings != nil {
			cert = dm.SslSettings.CertificateId
		}
		fmt.Printf("%-40s %s\n", dm.Id, cert)
	}
	return nil
}

func runAppDomMapUpdate(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	if flagAppCertificateID != "" && flagAppNoCertificateID {
		return fmt.Errorf("--certificate-id and --no-certificate-id are mutually exclusive")
	}
	mgmt := strings.ToLower(flagAppCertificateMgmt)
	if mgmt == "" && (flagAppCertificateID != "" || flagAppNoCertificateID) {
		mgmt = "manual"
	}
	if mgmt != "" && mgmt != "automatic" && mgmt != "manual" {
		return fmt.Errorf("--certificate-management must be automatic or manual")
	}
	body := &appengine.DomainMapping{}
	var masks []string
	if mgmt != "" || flagAppCertificateID != "" || flagAppNoCertificateID {
		body.SslSettings = &appengine.SslSettings{}
		if mgmt != "" {
			body.SslSettings.SslManagementType = strings.ToUpper(mgmt)
		}
		if flagAppCertificateID != "" {
			body.SslSettings.CertificateId = flagAppCertificateID
		}
		if flagAppNoCertificateID {
			body.SslSettings.CertificateId = ""
			body.SslSettings.ForceSendFields = append(body.SslSettings.ForceSendFields, "CertificateId")
		}
		masks = append(masks, "ssl_settings")
	}
	if len(masks) == 0 {
		return fmt.Errorf("no update options specified")
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Apps.DomainMappings.Patch(app, args[0], body).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating domain mapping: %w", err)
	}
	return appFinishOp(ctx, svc, app, op, "Update domain mapping", args[0])
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

func runAppFwCreate(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
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
	action := strings.ToUpper(flagAppAction)
	if action != "ALLOW" && action != "DENY" {
		return fmt.Errorf("--action must be ALLOW or DENY")
	}
	body := &appengine.FirewallRule{
		Priority:    priority,
		SourceRange: flagAppSourceRange,
		Action:      action,
		Description: flagAppDescription,
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Firewall.IngressRules.Create(app, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating firewall rule: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created firewall rule [%s].\n", args[0])
	return emitFormatted(got, flagAppFormat)
}

func runAppFwDelete(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Apps.Firewall.IngressRules.Delete(app, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting firewall rule: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deleted firewall rule [%s].\n", args[0])
	return nil
}

func runAppFwDescribe(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Firewall.IngressRules.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing firewall rule: %w", err)
	}
	return emitFormatted(got, flagAppFormat)
}

func runAppFwList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*appengine.FirewallRule
	pageToken := ""
	for {
		call := svc.Apps.Firewall.IngressRules.List(app).Context(ctx)
		if flagAppMatchingAddress != "" {
			call = call.MatchingAddress(flagAppMatchingAddress)
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
	if flagAppFormat != "" {
		return emitFormatted(all, flagAppFormat)
	}
	fmt.Printf("%-12s %-8s %-32s %s\n", "PRIORITY", "ACTION", "SOURCE_RANGE", "DESCRIPTION")
	for _, r := range all {
		fmt.Printf("%-12d %-8s %-32s %s\n", r.Priority, r.Action, r.SourceRange, r.Description)
	}
	return nil
}

func runAppFwUpdate(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	body := &appengine.FirewallRule{}
	var masks []string
	if flagAppAction != "" {
		action := strings.ToUpper(flagAppAction)
		if action != "ALLOW" && action != "DENY" {
			return fmt.Errorf("--action must be ALLOW or DENY")
		}
		body.Action = action
		masks = append(masks, "action")
	}
	if flagAppSourceRange != "" {
		body.SourceRange = flagAppSourceRange
		masks = append(masks, "sourceRange")
	}
	if cmd.Flags().Changed("description") {
		body.Description = flagAppDescription
		masks = append(masks, "description")
	}
	if len(masks) == 0 {
		return fmt.Errorf("no update options specified")
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Firewall.IngressRules.Patch(app, args[0], body).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating firewall rule: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated firewall rule [%s].\n", args[0])
	return emitFormatted(got, flagAppFormat)
}

func runAppFwTestIP(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
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
		fmt.Fprintln(os.Stderr, "No rules match the IP address.")
		return nil
	}
	fmt.Fprintf(os.Stderr, "The action `%s` will apply to the IP address.\n", all[0].Action)
	if flagAppFormat != "" {
		return emitFormatted(all, flagAppFormat)
	}
	fmt.Printf("%-12s %-8s %-32s %s\n", "PRIORITY", "ACTION", "SOURCE_RANGE", "DESCRIPTION")
	for _, r := range all {
		fmt.Printf("%-12d %-8s %-32s %s\n", r.Priority, r.Action, r.SourceRange, r.Description)
	}
	return nil
}

// -----------------------------------------------------------------------------
// instances implementations
// -----------------------------------------------------------------------------

func runAppInstDelete(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Apps.Services.Versions.Instances.Delete(app, flagAppService, flagAppVersion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return appFinishOp(ctx, svc, app, op, "Delete instance", args[0])
}

func runAppInstDescribe(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Services.Versions.Instances.Get(app, flagAppService, flagAppVersion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagAppFormat)
}

func runAppInstEnableDebug(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Apps.Services.Versions.Instances.Debug(app, flagAppService, flagAppVersion, args[0], &appengine.DebugInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling debug: %w", err)
	}
	return appFinishOp(ctx, svc, app, op, "Enable debug on instance", args[0])
}

// runAppInstDisableDebug: the App Engine v1 API exposes only the Debug method (which
// enables debug on flex instances). Disabling debug requires deleting the instance,
// which is destructive; we surface an explicit error instead of silently deleting.
func runAppInstDisableDebug(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("disable-debug is not supported by the App Engine Admin v1 API; delete the instance with `gcloud app instances delete` to remove its debug session")
}

func runAppInstList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
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
		if flagAppService != "" && s.Id != flagAppService {
			continue
		}
		versions, err := appListVersions(ctx, svc, app, s.Id)
		if err != nil {
			return err
		}
		for _, v := range versions {
			if flagAppVersion != "" && v.Id != flagAppVersion {
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
	if flagAppFormat != "" {
		return emitFormatted(rows, flagAppFormat)
	}
	fmt.Printf("%-20s %-20s %-40s %s\n", "SERVICE", "VERSION", "ID", "VM_STATUS")
	for _, r := range rows {
		fmt.Printf("%-20s %-20s %-40s %s\n", r.Service, r.Version, r.Instance.Id, r.Instance.VmStatus)
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

var defaultAppLogIDs = []string{"stdout", "stderr", "crash.log", "nginx.request", "appengine.googleapis.com%2Frequest_log"}

func buildAppLogFilter(project string) string {
	parts := []string{`resource.type="gae_app"`}
	if flagAppService != "" {
		parts = append(parts, fmt.Sprintf(`resource.labels.module_id=%q`, flagAppService))
	}
	if flagAppVersion != "" {
		parts = append(parts, fmt.Sprintf(`resource.labels.version_id=%q`, flagAppVersion))
	}
	if flagAppLevel != "" && !strings.EqualFold(flagAppLevel, "any") {
		parts = append(parts, fmt.Sprintf(`severity>=%s`, strings.ToUpper(flagAppLevel)))
	}
	logs := defaultAppLogIDs
	if flagAppLogs != "" {
		logs = strings.Split(flagAppLogs, ",")
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

func runAppLogsRead(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagAppLimit < 1 || flagAppLimit > 1000 {
		return fmt.Errorf("--limit must be between 1 and 1000")
	}
	ctx := context.Background()
	svc, err := gcp.LoggingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &logging.ListLogEntriesRequest{
		Filter:        buildAppLogFilter(project),
		OrderBy:       "timestamp desc",
		PageSize:      flagAppLimit,
		ResourceNames: []string{"projects/" + project},
	}
	resp, err := svc.Entries.List(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing log entries: %w", err)
	}
	if flagAppFormat != "" {
		return emitFormatted(resp.Entries, flagAppFormat)
	}
	// Reverse so the newest entry prints last, matching gcloud.
	for i := len(resp.Entries) - 1; i >= 0; i-- {
		printAppLogEntry(resp.Entries[i])
	}
	return nil
}

func runAppLogsTail(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagAppLimit < 1 || flagAppLimit > 1000 {
		return fmt.Errorf("--limit must be between 1 and 1000")
	}
	ctx := context.Background()
	svc, err := gcp.LoggingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &logging.ListLogEntriesRequest{
		Filter:        buildAppLogFilter(project),
		OrderBy:       "timestamp desc",
		PageSize:      flagAppLimit,
		ResourceNames: []string{"projects/" + project},
	}
	resp, err := svc.Entries.List(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing log entries: %w", err)
	}
	if flagAppFormat != "" {
		return emitFormatted(resp.Entries, flagAppFormat)
	}
	for i := len(resp.Entries) - 1; i >= 0; i-- {
		printAppLogEntry(resp.Entries[i])
	}
	return nil
}

func printAppLogEntry(e *logging.LogEntry) {
	msg := e.TextPayload
	if msg == "" && e.JsonPayload != nil {
		msg = string(e.JsonPayload)
	}
	if msg == "" && e.HttpRequest != nil {
		msg = fmt.Sprintf("%s %s -> %d", e.HttpRequest.RequestMethod, e.HttpRequest.RequestUrl, e.HttpRequest.Status)
	}
	fmt.Printf("%s %s %s\n", e.Timestamp, e.Severity, strings.TrimSpace(msg))
}

// -----------------------------------------------------------------------------
// operations implementations
// -----------------------------------------------------------------------------

func runAppOpDescribe(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Operations.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagAppFormat)
}

func runAppOpList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*appengine.Operation
	pageToken := ""
	for {
		call := svc.Apps.Operations.List(app).Context(ctx)
		filter := flagAppFilter
		if flagAppPending {
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
	if flagAppFormat != "" {
		return emitFormatted(all, flagAppFormat)
	}
	fmt.Printf("%-50s %-8s\n", "ID", "DONE")
	for _, op := range all {
		fmt.Printf("%-50s %v\n", path.Base(op.Name), op.Done)
	}
	return nil
}

func runAppOpWait(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Apps.Operations.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching operation: %w", err)
	}
	if op.Done {
		fmt.Fprintf(os.Stderr, "Operation [%s] is already done.\n", args[0])
		return emitFormatted(op, flagAppFormat)
	}
	fmt.Fprintf(os.Stderr, "Waiting for operation [%s] to complete.\n", args[0])
	final, err := appWaitOp(ctx, svc, app, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagAppFormat)
}

// -----------------------------------------------------------------------------
// regions implementations
// -----------------------------------------------------------------------------

func runAppRegionsList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
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
	if flagAppFormat != "" {
		return emitFormatted(all, flagAppFormat)
	}
	fmt.Printf("%-24s %-16s %s\n", "REGION", "STANDARD", "FLEXIBLE")
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
		fmt.Printf("%-24s %-16s %s\n", l.LocationId, std, flex)
	}
	return nil
}

// -----------------------------------------------------------------------------
// runtimes implementations
// -----------------------------------------------------------------------------

func runAppRuntimesList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Apps.ListRuntimes(app).Context(ctx)
	if flagAppEnvironment != "" {
		call = call.Environment(strings.ToUpper(flagAppEnvironment))
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing runtimes: %w", err)
	}
	if flagAppFormat != "" {
		return emitFormatted(resp.Runtimes, flagAppFormat)
	}
	fmt.Printf("%-24s %-16s %-16s %s\n", "NAME", "ENVIRONMENT", "STAGE", "DISPLAY_NAME")
	for _, r := range resp.Runtimes {
		fmt.Printf("%-24s %-16s %-16s %s\n", r.Name, r.Environment, r.Stage, r.DisplayName)
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

func runAppSvcBrowse(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	host, err := appDefaultHostname(ctx, svc, app)
	if err != nil {
		return err
	}
	for _, service := range args {
		url := serviceURL(host, service, flagAppVersion)
		fmt.Println(url)
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

func runAppSvcDelete(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	for _, service := range args {
		if flagAppVersion != "" {
			op, err := svc.Apps.Services.Versions.Delete(app, service, flagAppVersion).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("deleting version %s/%s: %w", service, flagAppVersion, err)
			}
			if err := appFinishOp(ctx, svc, app, op, "Delete version", service+"/"+flagAppVersion); err != nil {
				return err
			}
			continue
		}
		op, err := svc.Apps.Services.Delete(app, service).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting service %s: %w", service, err)
		}
		if err := appFinishOp(ctx, svc, app, op, "Delete service", service); err != nil {
			return err
		}
	}
	return nil
}

func runAppSvcDescribe(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.Services.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormatted(got, flagAppFormat)
}

func runAppSvcList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	all, err := appListServices(ctx, svc, app)
	if err != nil {
		return fmt.Errorf("listing services: %w", err)
	}
	if flagAppFormat != "" {
		return emitFormatted(all, flagAppFormat)
	}
	fmt.Printf("%-24s %-16s %s\n", "SERVICE", "NUM_VERSIONS", "TRAFFIC")
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
		fmt.Printf("%-24s %-16d %s\n", s.Id, versions, split)
	}
	return nil
}

func runAppSvcSetTraffic(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	if len(flagAppSplits) == 0 {
		return fmt.Errorf("--splits is required")
	}
	shardBy := strings.ToUpper(flagAppSplitBy)
	switch shardBy {
	case "COOKIE", "IP", "RANDOM":
	default:
		return fmt.Errorf("--split-by must be cookie, ip, or random")
	}
	// Normalize weights so they sum to 1.
	sum := 0.0
	weights := make(map[string]float64, len(flagAppSplits))
	for k, v := range flagAppSplits {
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
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
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
		if flagAppMigrate {
			call = call.MigrateTraffic(true)
		}
		op, err := call.Do()
		if err != nil {
			return fmt.Errorf("setting traffic on %s: %w", s, err)
		}
		if err := appFinishOp(ctx, svc, app, op, "Set traffic", s); err != nil {
			return err
		}
	}
	return nil
}

func runAppSvcUpdate(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	if flagAppIngress == "" {
		return fmt.Errorf("no update options specified (try --ingress)")
	}
	ingress := ""
	switch strings.ToLower(flagAppIngress) {
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
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Apps.Services.Patch(app, args[0], body).UpdateMask("networkSettings").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return appFinishOp(ctx, svc, app, op, "Update service", args[0])
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

func runAppSSLCreate(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	cert, err := readFileOrError(flagAppCertificateFile)
	if err != nil {
		return err
	}
	key, err := readFileOrError(flagAppPrivateKeyFile)
	if err != nil {
		return err
	}
	body := &appengine.AuthorizedCertificate{
		DisplayName: flagAppDisplayName,
		CertificateRawData: &appengine.CertificateRawData{
			PublicCertificate: cert,
			PrivateKey:        key,
		},
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.AuthorizedCertificates.Create(app, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating SSL certificate: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created SSL certificate [%s].\n", got.Id)
	return emitFormatted(got, flagAppFormat)
}

func runAppSSLDelete(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Apps.AuthorizedCertificates.Delete(app, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting SSL certificate: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deleted SSL certificate [%s].\n", args[0])
	return nil
}

func runAppSSLDescribe(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.AuthorizedCertificates.Get(app, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing SSL certificate: %w", err)
	}
	return emitFormatted(got, flagAppFormat)
}

func runAppSSLList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
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
	if flagAppFormat != "" {
		return emitFormatted(all, flagAppFormat)
	}
	fmt.Printf("%-16s %-32s %s\n", "ID", "DISPLAY_NAME", "DOMAINS")
	for _, c := range all {
		fmt.Printf("%-16s %-32s %s\n", c.Id, c.DisplayName, strings.Join(c.DomainNames, ","))
	}
	return nil
}

func runAppSSLUpdate(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	body := &appengine.AuthorizedCertificate{}
	var masks []string
	if flagAppDisplayName != "" {
		body.DisplayName = flagAppDisplayName
		masks = append(masks, "displayName")
	}
	if flagAppCertificateFile != "" || flagAppPrivateKeyFile != "" {
		raw := &appengine.CertificateRawData{}
		if flagAppCertificateFile != "" {
			cert, err := readFileOrError(flagAppCertificateFile)
			if err != nil {
				return err
			}
			raw.PublicCertificate = cert
		}
		if flagAppPrivateKeyFile != "" {
			key, err := readFileOrError(flagAppPrivateKeyFile)
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
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Apps.AuthorizedCertificates.Patch(app, args[0], body).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating SSL certificate: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated SSL certificate [%s].\n", args[0])
	return emitFormatted(got, flagAppFormat)
}

// -----------------------------------------------------------------------------
// versions implementations
// -----------------------------------------------------------------------------

func runAppVerBrowse(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	host, err := appDefaultHostname(ctx, svc, app)
	if err != nil {
		return err
	}
	service := flagAppService
	if service == "" {
		service = "default"
	}
	for _, v := range args {
		fmt.Println(serviceURL(host, service, v))
	}
	return nil
}

func runAppVerDelete(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	service := flagAppService
	if service == "" {
		service = "default"
	}
	for _, v := range args {
		op, err := svc.Apps.Services.Versions.Delete(app, service, v).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("deleting version %s: %w", v, err)
		}
		if err := appFinishOp(ctx, svc, app, op, "Delete version", service+"/"+v); err != nil {
			return err
		}
	}
	return nil
}

func runAppVerDescribe(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	service := flagAppService
	if service == "" {
		service = "default"
	}
	got, err := svc.Apps.Services.Versions.Get(app, service, args[0]).View("FULL").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing version: %w", err)
	}
	return emitFormatted(got, flagAppFormat)
}

func runAppVerList(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
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
		if flagAppService != "" && s.Id != flagAppService {
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
			if flagAppHideNoTraffic && split == 0 {
				continue
			}
			rows = append(rows, verRow{Service: s.Id, Version: v, TrafficSplit: split})
		}
	}
	if flagAppFormat != "" {
		return emitFormatted(rows, flagAppFormat)
	}
	fmt.Printf("%-20s %-20s %-8s %-24s %s\n", "SERVICE", "VERSION.ID", "TRAFFIC", "LAST_DEPLOYED", "SERVING_STATUS")
	for _, r := range rows {
		fmt.Printf("%-20s %-20s %-8.2f %-24s %s\n", r.Service, r.Version.Id, r.TrafficSplit, r.Version.CreateTime, r.Version.ServingStatus)
	}
	return nil
}

func runAppVerMigrate(cmd *cobra.Command, args []string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	service := flagAppService
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
	return appFinishOp(ctx, svc, app, op, "Migrate traffic", service+"/"+args[0])
}

func runAppVerStart(cmd *cobra.Command, args []string) error {
	return appSetServingStatus(cmd, args, "SERVING", "Start")
}

func runAppVerStop(cmd *cobra.Command, args []string) error {
	return appSetServingStatus(cmd, args, "STOPPED", "Stop")
}

func appSetServingStatus(cmd *cobra.Command, args []string, status, verb string) error {
	app, err := appResolveApp()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AppEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	service := flagAppService
	if service == "" {
		service = "default"
	}
	body := &appengine.Version{ServingStatus: status}
	for _, v := range args {
		op, err := svc.Apps.Services.Versions.Patch(app, service, v, body).UpdateMask("servingStatus").Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("%s version %s: %w", strings.ToLower(verb), v, err)
		}
		if err := appFinishOp(ctx, svc, app, op, verb+" version", service+"/"+v); err != nil {
			return err
		}
	}
	return nil
}
