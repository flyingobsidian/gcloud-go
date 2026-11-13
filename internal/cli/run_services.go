package cli

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	logging "google.golang.org/api/logging/v2"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run services (#1054) ---
//
// Backed by run/v2 Projects.Locations.Services.

func (s *runCmds) initRunServices() {
	all := []*cobra.Command{
		s.runServicesDeleteCmd, s.runServicesDescribeCmd, s.runServicesListCmd,
		s.runServicesReplaceCmd, s.runServicesUpdateCmd, s.runServicesUpdateTrafficCmd,
		s.runServicesLogsCmd, s.runServicesProxyCmd, s.runServicesGetIamCmd,
		s.runServicesSetIamCmd, s.runServicesAddIamCmd, s.runServicesRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagRunServicesFormat, "format", "", "Output format")
		if c == s.runServicesListCmd {
			// As in gcloud, list spans every region unless --region is given.
			c.Flags().StringVar(&s.flagRunServicesRegion, "region", "", "Cloud Run region; default all regions")
			continue
		}
		c.Flags().StringVar(&s.flagRunServicesRegion, "region", "", "Cloud Run region (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.runServicesReplaceCmd, s.runServicesUpdateCmd} {
		c.Flags().StringVar(&s.flagRunServicesConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Service body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.runServicesUpdateCmd.Flags().StringVar(&s.flagRunServicesUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")

	s.runServicesListCmd.Flags().Int64Var(&s.flagRunServicesPageSize, "page-size", 0, "Maximum results per page")
	s.runServicesListCmd.Flags().BoolVar(&s.flagRunServicesShowDel, "show-deleted", false, "Include deleted services")

	s.runServicesUpdateTrafficCmd.Flags().StringToStringVar(&s.flagRunServicesTrafficRevs, "to-revisions", nil,
		"Revision-to-percent map, e.g. REV1=60,REV2=40")
	s.runServicesUpdateTrafficCmd.Flags().BoolVar(&s.flagRunServicesTrafficLatest, "to-latest", false,
		"Route 100% of traffic to the latest ready revision")

	s.runServicesLogsCmd.Flags().Int64Var(&s.flagRunServicesLimit, "limit", 100,
		"Maximum number of log entries to return")

	s.runServicesProxyCmd.Flags().IntVar(&s.flagRunServicesProxyPort, "port", 8080,
		"Local port to listen on")
	s.runServicesProxyCmd.Flags().StringVar(&s.flagRunServicesProxyBind, "bind", "127.0.0.1",
		"Local address to bind")

	for _, c := range []*cobra.Command{s.runServicesAddIamCmd, s.runServicesRemoveIamCmd} {
		runIamFlags(c, &s.flagRunServicesIamMember, &s.flagRunServicesIamRole,
			&s.flagRunServicesIamCondExpr, &s.flagRunServicesIamCondT, &s.flagRunServicesIamCondD)
	}
	s.runServicesRemoveIamCmd.Flags().BoolVar(&s.flagRunServicesIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.runServicesCmd.AddCommand(all...)
	s.runCmd.AddCommand(s.runServicesCmd)
}

func (s *runCmds) runSvcName(project, service string) string {
	return runResourceName(project, s.flagRunServicesRegion, "services", service)
}

func (s *runCmds) runSvcDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Delete(s.runSvcName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Get(s.runSvcName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if s.flagRunServicesRegion == "" {
		v1, err := s.inv.clients().RunV1(ctx, "")
		if err != nil {
			return err
		}
		all, err := runListAllServices(ctx, s.inv.stderr, v1, project)
		if err != nil {
			return err
		}
		return emitFormattedTo(s.inv.stdout, all, s.flagRunServicesFormat)
	}
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	var all []*runv2.GoogleCloudRunV2Service
	pageToken := ""
	for {
		call := svc.Projects.Locations.Services.List(runParent(project, s.flagRunServicesRegion)).Context(ctx)
		if s.flagRunServicesPageSize > 0 {
			call = call.PageSize(s.flagRunServicesPageSize)
		}
		if s.flagRunServicesShowDel {
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
	return emitFormattedTo(s.inv.stdout, all, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcReplace(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Service{}
	if err := loadYAMLOrJSONInto(s.flagRunServicesConfigFile, body); err != nil {
		return err
	}
	body.Name = s.runSvcName(project, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.Patch(body.Name, body).
		AllowMissing(true).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing service: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Replace request issued for service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &runv2.GoogleCloudRunV2Service{}
	if err := loadYAMLOrJSONInto(s.flagRunServicesConfigFile, body); err != nil {
		return err
	}
	mask := s.flagRunServicesUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	body.Name = s.runSvcName(project, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
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
	fmt.Fprintf(s.inv.stdout, "Update request issued for service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcUpdateTraffic(cmd *cobra.Command, args []string) error {
	if len(s.flagRunServicesTrafficRevs) == 0 && !s.flagRunServicesTrafficLatest {
		return fmt.Errorf("one of --to-revisions or --to-latest is required")
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	name := s.runSvcName(project, args[0])
	current, err := svc.Projects.Locations.Services.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("loading service: %w", err)
	}
	var traffic []*runv2.GoogleCloudRunV2TrafficTarget
	if s.flagRunServicesTrafficLatest {
		traffic = append(traffic, &runv2.GoogleCloudRunV2TrafficTarget{
			Type: "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST", Percent: 100,
		})
	}
	for rev, pctStr := range s.flagRunServicesTrafficRevs {
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
	fmt.Fprintf(s.inv.stdout, "Traffic update issued for service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcLogs(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Logging(ctx)
	if err != nil {
		return err
	}
	filter := fmt.Sprintf(
		`resource.type="cloud_run_revision" AND resource.labels.service_name=%q AND resource.labels.location=%q`,
		args[0], s.flagRunServicesRegion,
	)
	limit := s.flagRunServicesLimit
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
	if s.flagRunServicesFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Entries, s.flagRunServicesFormat)
	}
	s.printLogEntries(resp.Entries)
	return nil
}

// printLogEntries writes log entries as "TIMESTAMP SEVERITY MESSAGE" lines.
func (s *runCmds) printLogEntries(entries []*logging.LogEntry) {
	for _, e := range entries {
		msg := logEntryMessage(e)
		fmt.Fprintf(s.inv.stdout, "%s %s %s\n", e.Timestamp, e.Severity, msg)
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

func (s *runCmds) runSvcProxy(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.Get(s.runSvcName(project, args[0])).Context(ctx).Do()
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
	ts, err := s.inv.clients().PlatformTokenSource(ctx)
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
	addr := fmt.Sprintf("%s:%d", s.flagRunServicesProxyBind, s.flagRunServicesProxyPort)
	fmt.Fprintf(s.inv.stdout, "Proxying %s -> %s\n", "http://"+addr, got.Uri)
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

func (s *runCmds) runSvcGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(s.runSvcName(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &runv2.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Services.SetIamPolicy(s.runSvcName(project, args[0]),
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("service [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	resource := s.runSvcName(project, args[0])
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	runIamAddBinding(policy, s.flagRunServicesIamRole, s.flagRunServicesIamMember,
		runIamBuildCondition(s.flagRunServicesIamCondExpr, s.flagRunServicesIamCondT, s.flagRunServicesIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Services.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("service [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunServicesFormat)
}

func (s *runCmds) runSvcRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().RunV2(ctx, s.flagRunServicesRegion)
	if err != nil {
		return err
	}
	resource := s.runSvcName(project, args[0])
	policy, err := svc.Projects.Locations.Services.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !runIamRemoveBinding(policy, s.flagRunServicesIamRole, s.flagRunServicesIamMember,
		runIamBuildCondition(s.flagRunServicesIamCondExpr, s.flagRunServicesIamCondT, s.flagRunServicesIamCondD),
		s.flagRunServicesIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			s.flagRunServicesIamRole, s.flagRunServicesIamMember)
	}
	updated, err := svc.Projects.Locations.Services.SetIamPolicy(resource,
		&runv2.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.runIamUpdatedIam(fmt.Sprintf("service [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagRunServicesFormat)
}
