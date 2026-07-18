package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	logging "google.golang.org/api/logging/v2"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run services (#1054) ---
//
// Backed by run/v2 Projects.Locations.Services.

var runServicesCmd = &cobra.Command{Use: "services", Short: "Manage Cloud Run services"}

var (
	flagRunServicesRegion     string
	flagRunServicesFormat     string
	flagRunServicesConfigFile string
	flagRunServicesUpdateMask string
	flagRunServicesPageSize   int64
	flagRunServicesShowDel    bool
	flagRunServicesLimit      int64

	flagRunServicesIamMember   string
	flagRunServicesIamRole     string
	flagRunServicesIamCondExpr string
	flagRunServicesIamCondT    string
	flagRunServicesIamCondD    string
	flagRunServicesIamAllCond  bool

	flagRunServicesTrafficRevs   map[string]string
	flagRunServicesTrafficLatest bool

	flagRunServicesProxyPort int
	flagRunServicesProxyBind string
)

var (
	runServicesDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: runSvcDelete,
	}
	runServicesDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: runSvcDescribe,
	}
	runServicesListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run services",
		Args: cobra.NoArgs, RunE: runSvcList,
	}
	runServicesReplaceCmd = &cobra.Command{
		Use: "replace SERVICE", Short: "Replace a Cloud Run service from a config file",
		Args: cobra.ExactArgs(1), RunE: runSvcReplace,
	}
	runServicesUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a Cloud Run service from a config file",
		Args: cobra.ExactArgs(1), RunE: runSvcUpdate,
	}
	runServicesUpdateTrafficCmd = &cobra.Command{
		Use: "update-traffic SERVICE", Short: "Update Cloud Run service traffic split",
		Args: cobra.ExactArgs(1), RunE: runSvcUpdateTraffic,
	}
	runServicesLogsCmd = &cobra.Command{
		Use: "logs SERVICE", Short: "Read Cloud Logging entries for a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: runSvcLogs,
	}
	runServicesProxyCmd = &cobra.Command{
		Use: "proxy SERVICE", Short: "Proxy authenticated local traffic to a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: runSvcProxy,
	}
	runServicesGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SERVICE", Short: "Get the IAM policy for a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: runSvcGetIam,
	}
	runServicesSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SERVICE POLICY_FILE", Short: "Set the IAM policy for a Cloud Run service",
		Args: cobra.ExactArgs(2), RunE: runSvcSetIam,
	}
	runServicesAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SERVICE", Short: "Add an IAM policy binding to a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: runSvcAddIam,
	}
	runServicesRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SERVICE", Short: "Remove an IAM policy binding from a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: runSvcRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		runServicesDeleteCmd, runServicesDescribeCmd, runServicesListCmd,
		runServicesReplaceCmd, runServicesUpdateCmd, runServicesUpdateTrafficCmd,
		runServicesLogsCmd, runServicesProxyCmd, runServicesGetIamCmd,
		runServicesSetIamCmd, runServicesAddIamCmd, runServicesRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagRunServicesRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&flagRunServicesFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{runServicesReplaceCmd, runServicesUpdateCmd} {
		c.Flags().StringVar(&flagRunServicesConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Service body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	runServicesUpdateCmd.Flags().StringVar(&flagRunServicesUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")

	runServicesListCmd.Flags().Int64Var(&flagRunServicesPageSize, "page-size", 0, "Maximum results per page")
	runServicesListCmd.Flags().BoolVar(&flagRunServicesShowDel, "show-deleted", false, "Include deleted services")

	runServicesUpdateTrafficCmd.Flags().StringToStringVar(&flagRunServicesTrafficRevs, "to-revisions", nil,
		"Revision-to-percent map, e.g. REV1=60,REV2=40")
	runServicesUpdateTrafficCmd.Flags().BoolVar(&flagRunServicesTrafficLatest, "to-latest", false,
		"Route 100% of traffic to the latest ready revision")

	runServicesLogsCmd.Flags().Int64Var(&flagRunServicesLimit, "limit", 100,
		"Maximum number of log entries to return")

	runServicesProxyCmd.Flags().IntVar(&flagRunServicesProxyPort, "port", 8080,
		"Local port to listen on")
	runServicesProxyCmd.Flags().StringVar(&flagRunServicesProxyBind, "bind", "127.0.0.1",
		"Local address to bind")

	for _, c := range []*cobra.Command{runServicesAddIamCmd, runServicesRemoveIamCmd} {
		runIamFlags(c, &flagRunServicesIamMember, &flagRunServicesIamRole,
			&flagRunServicesIamCondExpr, &flagRunServicesIamCondT, &flagRunServicesIamCondD)
	}
	runServicesRemoveIamCmd.Flags().BoolVar(&flagRunServicesIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	runServicesCmd.AddCommand(all...)
	runCmd.AddCommand(runServicesCmd)
}

func runSvcName(project, service string) string {
	return runResourceName(project, flagRunServicesRegion, "services", service)
}

func runSvcDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Delete(runSvcName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	fmt.Printf("Delete request issued for service [%s].\n", args[0])
	return emitFormatted(op, flagRunServicesFormat)
}

func runSvcDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Get(runSvcName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormatted(got, flagRunServicesFormat)
}

func runSvcList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	var all []*runv2.GoogleCloudRunV2Service
	pageToken := ""
	for {
		call := svc.Projects.Locations.Services.List(runParent(project, flagRunServicesRegion)).Context(ctx)
		if flagRunServicesPageSize > 0 {
			call = call.PageSize(flagRunServicesPageSize)
		}
		if flagRunServicesShowDel {
			call = call.ShowDeleted(true)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing services: %w", err)
		}
		all = append(all, resp.Services...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagRunServicesFormat)
}

func runSvcReplace(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Service{}
	if err := loadYAMLOrJSONInto(flagRunServicesConfigFile, body); err != nil {
		return err
	}
	body.Name = runSvcName(project, args[0])
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Patch(body.Name, body).
		AllowMissing(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing service: %w", err)
	}
	fmt.Printf("Replace request issued for service [%s].\n", args[0])
	return emitFormatted(op, flagRunServicesFormat)
}

func runSvcUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Service{}
	if err := loadYAMLOrJSONInto(flagRunServicesConfigFile, body); err != nil {
		return err
	}
	mask := flagRunServicesUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	body.Name = runSvcName(project, args[0])
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Services.Patch(body.Name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	fmt.Printf("Update request issued for service [%s].\n", args[0])
	return emitFormatted(op, flagRunServicesFormat)
}

func runSvcUpdateTraffic(cmd *cobra.Command, args []string) error {
	if len(flagRunServicesTrafficRevs) == 0 && !flagRunServicesTrafficLatest {
		return fmt.Errorf("one of --to-revisions or --to-latest is required")
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	name := runSvcName(project, args[0])
	current, err := svc.Projects.Locations.Services.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("loading service: %w", err)
	}
	var traffic []*runv2.GoogleCloudRunV2TrafficTarget
	if flagRunServicesTrafficLatest {
		traffic = append(traffic, &runv2.GoogleCloudRunV2TrafficTarget{
			Type: "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST", Percent: 100,
		})
	}
	for rev, pctStr := range flagRunServicesTrafficRevs {
		pct, err := parseTrafficPercent(pctStr)
		if err != nil {
			return err
		}
		traffic = append(traffic, &runv2.GoogleCloudRunV2TrafficTarget{
			Type:     "TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION",
			Revision: rev,
			Percent:  pct,
		})
	}
	patch := &runv2.GoogleCloudRunV2Service{
		Name:    current.Name,
		Traffic: traffic,
	}
	op, err := svc.Projects.Locations.Services.Patch(name, patch).
		UpdateMask("traffic").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating traffic: %w", err)
	}
	fmt.Printf("Traffic update issued for service [%s].\n", args[0])
	return emitFormatted(op, flagRunServicesFormat)
}

func runSvcLogs(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LoggingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	filter := fmt.Sprintf(
		`resource.type="cloud_run_revision" AND resource.labels.service_name=%q AND resource.labels.location=%q`,
		args[0], flagRunServicesRegion,
	)
	limit := flagRunServicesLimit
	if limit <= 0 {
		limit = 100
	}
	req := &logging.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + project},
		Filter:        filter,
		OrderBy:       "timestamp desc",
		PageSize:      limit,
	}
	resp, err := svc.Entries.List(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing log entries: %w", err)
	}
	if flagRunServicesFormat != "" {
		return emitFormatted(resp.Entries, flagRunServicesFormat)
	}
	printLogEntries(resp.Entries)
	return nil
}

// printLogEntries writes log entries as "TIMESTAMP SEVERITY MESSAGE" lines.
func printLogEntries(entries []*logging.LogEntry) {
	for _, e := range entries {
		msg := logEntryMessage(e)
		fmt.Printf("%s %s %s\n", e.Timestamp, e.Severity, msg)
	}
}

func logEntryMessage(e *logging.LogEntry) string {
	if e.TextPayload != "" {
		return e.TextPayload
	}
	if len(e.JsonPayload) > 0 {
		return string(e.JsonPayload)
	}
	if len(e.ProtoPayload) > 0 {
		return string(e.ProtoPayload)
	}
	return ""
}

// parseTrafficPercent parses a --to-revisions value (an integer 0-100).
func parseTrafficPercent(s string) (int64, error) {
	var pct int64
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &pct); err != nil {
		return 0, fmt.Errorf("invalid traffic percent %q: %w", s, err)
	}
	if pct < 0 || pct > 100 {
		return 0, fmt.Errorf("traffic percent %q out of range [0,100]", s)
	}
	return pct, nil
}

func runSvcProxy(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Get(runSvcName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	if got.Uri == "" {
		return fmt.Errorf("service [%s] has no serving URI", args[0])
	}
	target, err := url.Parse(got.Uri)
	if err != nil {
		return fmt.Errorf("parsing service URI %q: %w", got.Uri, err)
	}
	ts, err := gcp.PlatformTokenSource(ctx, flagAccount)
	if err != nil {
		return err
	}
	// The reverse proxy stamps the Host to the upstream Cloud Run
	// hostname and layers an OAuth2 access token onto each request.
	// Cloud Run accepts OAuth2 access tokens with the cloud-platform
	// scope for authenticated invocations; ID tokens would be preferred
	// but require idtoken package integration.
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &authProxyTransport{
		base:   http.DefaultTransport,
		ts:     ts,
		host:   target.Host,
		scheme: target.Scheme,
	}
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = target.Host
		req.URL.Host = target.Host
		req.URL.Scheme = target.Scheme
	}
	addr := fmt.Sprintf("%s:%d", flagRunServicesProxyBind, flagRunServicesProxyPort)
	fmt.Printf("Proxying %s -> %s\n", "http://"+addr, got.Uri)
	server := &http.Server{
		Addr:              addr,
		Handler:           proxy,
		ReadHeaderTimeout: 30 * time.Second,
	}
	return server.ListenAndServe()
}

// authProxyTransport injects a bearer token on every proxied request.
type authProxyTransport struct {
	base   http.RoundTripper
	ts     oauth2.TokenSource
	host   string
	scheme string
}

func (t *authProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.ts.Token()
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+token.AccessToken)
	clone.Host = t.host
	clone.URL.Host = t.host
	clone.URL.Scheme = t.scheme
	return t.base.RoundTrip(clone)
}

func runSvcGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(runSvcName(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagRunServicesFormat)
}

func runSvcSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &runv2.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Services.SetIamPolicy(runSvcName(project, args[0]),
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("service [%s]", args[0]))
	return emitFormatted(updated, flagRunServicesFormat)
}

func runSvcAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	resource := runSvcName(project, args[0])
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	runIamAddBinding(policy, flagRunServicesIamRole, flagRunServicesIamMember,
		runIamBuildCondition(flagRunServicesIamCondExpr, flagRunServicesIamCondT, flagRunServicesIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Services.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("service [%s]", args[0]))
	return emitFormatted(updated, flagRunServicesFormat)
}

func runSvcRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.RunV2Service(ctx, flagAccount, flagRunServicesRegion)
	if err != nil {
		return err
	}
	resource := runSvcName(project, args[0])
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !runIamRemoveBinding(policy, flagRunServicesIamRole, flagRunServicesIamMember,
		runIamBuildCondition(flagRunServicesIamCondExpr, flagRunServicesIamCondT, flagRunServicesIamCondD),
		flagRunServicesIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			flagRunServicesIamRole, flagRunServicesIamMember)
	}
	updated, err := svc.Projects.Locations.Services.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	runIamUpdatedIam(fmt.Sprintf("service [%s]", args[0]))
	return emitFormatted(updated, flagRunServicesFormat)
}
