package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	servicemanagement "google.golang.org/api/servicemanagement/v1"
	su "google.golang.org/api/serviceusage/v1"
)

// --- Shared helpers for the Endpoints surface (#892-#894) ---

func (g *endpointsCmds) endpointsService(ctx context.Context) (*servicemanagement.APIService, error) {
	return g.inv.clients().ServiceManagement(ctx)
}

// --- Subgroup command objects ---

// --- configs ---

// --- operations ---

// --- services ---

func (g *endpointsCmds) initEndpointsAll() {
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&g.flagEPFormat, "format", "", "Output format")
		}
	}

	// configs
	g.epConfigDescribeCmd.Flags().StringVar(&g.flagEPService, "service", "", "Service name (required)")
	_ = g.epConfigDescribeCmd.MarkFlagRequired("service")
	g.epConfigListCmd.Flags().StringVar(&g.flagEPService, "service", "", "Service name (required)")
	_ = g.epConfigListCmd.MarkFlagRequired("service")
	g.epConfigListCmd.Flags().Int64Var(&g.flagEPPageSize, "page-size", 0, "Page size for API pagination")
	addFmt(g.epConfigDescribeCmd, g.epConfigListCmd)
	g.endpointsConfigsCmd.AddCommand(g.epConfigDescribeCmd, g.epConfigListCmd)
	g.endpointsCmd.AddCommand(g.endpointsConfigsCmd)

	// operations
	addFmt(g.epOpDescribeCmd, g.epOpListCmd, g.epOpWaitCmd)
	g.epOpListCmd.Flags().Int64Var(&g.flagEPPageSize, "page-size", 0, "Page size for API pagination")
	g.endpointsOperationsCmd.AddCommand(g.epOpDescribeCmd, g.epOpListCmd, g.epOpWaitCmd)
	g.endpointsCmd.AddCommand(g.endpointsOperationsCmd)

	// services
	g.epSvcDeleteCmd.Flags().BoolVar(&g.flagEPAsync, "async", false, "Do not wait for the operation to finish")
	g.epSvcUndeployCmd.Flags().BoolVar(&g.flagEPAsync, "async", false, "Do not wait for the operation to finish")
	g.epSvcDeployCmd.Flags().BoolVar(&g.flagEPAsync, "async", false, "Do not wait for the operation to finish")
	g.epSvcDeployCmd.Flags().BoolVar(&g.flagEPValidate, "validate-only", false, "Validate the configuration only; do not deploy")
	g.epSvcDeployCmd.Flags().BoolVar(&g.flagEPForce, "force", false, "Force deployment even when hazardous changes are detected")
	for _, c := range []*cobra.Command{g.epSvcAddIAMCmd, g.epSvcRemoveIAMCmd} {
		c.Flags().StringVar(&g.flagEPMember, "member", "", "IAM member, e.g. user:alice@example.com (required)")
		c.Flags().StringVar(&g.flagEPRole, "role", "", "IAM role, e.g. roles/servicemanagement.serviceConsumer (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	addFmt(g.epSvcDescribeCmd, g.epSvcListCmd, g.epSvcDeployCmd, g.epSvcGetConfigNameCmd,
		g.epSvcGetIAMCmd, g.epSvcSetIAMCmd, g.epSvcCheckIAMCmd)
	g.endpointsServicesCmd.AddCommand(
		g.epSvcDeleteCmd, g.epSvcUndeployCmd, g.epSvcDescribeCmd, g.epSvcListCmd, g.epSvcDeployCmd,
		g.epSvcGetConfigNameCmd, g.epSvcEnableCmd, g.epSvcDisableCmd,
		g.epSvcAddIAMCmd, g.epSvcRemoveIAMCmd, g.epSvcGetIAMCmd, g.epSvcSetIAMCmd, g.epSvcCheckIAMCmd,
	)
	g.endpointsCmd.AddCommand(g.endpointsServicesCmd)
}

// --- configs impl ---

func (g *endpointsCmds) runEPConfigDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Services.Configs.Get(g.flagEPService, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service config: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagEPFormat)
}

func (g *endpointsCmds) runEPConfigList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	var all []*servicemanagement.Service
	pageToken := ""
	for {
		call := svc.Services.Configs.List(g.flagEPService).Context(ctx)
		if g.flagEPPageSize > 0 {
			call = call.PageSize(g.flagEPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service configs: %w", err)
		}
		all = append(all, resp.ServiceConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagEPFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagEPFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-30s %s\n", "CONFIG_ID", "SERVICE_NAME")
	for _, c := range all {
		fmt.Fprintf(g.inv.stdout, "%-30s %s\n", c.Id, c.Name)
	}
	return nil
}

// --- operations impl ---

func epOperationsName(id string) string {
	if strings.HasPrefix(id, "operations/") {
		return id
	}
	return "operations/" + id
}

func (g *endpointsCmds) runEPOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Operations.Get(epOperationsName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagEPFormat)
}

func (g *endpointsCmds) runEPOpList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	var all []*servicemanagement.Operation
	pageToken := ""
	for {
		call := svc.Operations.List().Context(ctx)
		if g.flagEPPageSize > 0 {
			call = call.PageSize(g.flagEPPageSize)
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
	if g.flagEPFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagEPFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-60s %-6s\n", "NAME", "DONE")
	for _, op := range all {
		fmt.Fprintf(g.inv.stdout, "%-60s %-6t\n", op.Name, op.Done)
	}
	return nil
}

func (g *endpointsCmds) runEPOpWait(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	final, err := epWaitOp(ctx, svc, epOperationsName(args[0]))
	if err != nil {
		return err
	}
	return emitFormattedTo(g.inv.stdout, final, g.flagEPFormat)
}

func epWaitOp(ctx context.Context, svc *servicemanagement.APIService, name string) (*servicemanagement.Operation, error) {
	for {
		got, err := svc.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", name, err)
		}
		if got.Done {
			if got.Error != nil {
				return got, fmt.Errorf("operation %s failed: %s", name, got.Error.Message)
			}
			return got, nil
		}
	}
}

// --- services impl ---

func (g *endpointsCmds) runEPSvcDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Services.Delete(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	if g.flagEPAsync {
		fmt.Fprintf(g.inv.stderr, "Delete in progress (operation: %s).\n", op.Name)
		return emitFormattedTo(g.inv.stdout, op, g.flagEPFormat)
	}
	final, err := epWaitOp(ctx, svc, op.Name)
	if err != nil {
		return err
	}
	return emitFormattedTo(g.inv.stdout, final, g.flagEPFormat)
}

func (g *endpointsCmds) runEPSvcDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Services.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagEPFormat)
}

func (g *endpointsCmds) runEPSvcList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	var all []*servicemanagement.ManagedService
	pageToken := ""
	for {
		call := svc.Services.List().ProducerProjectId(project).Context(ctx)
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
	if g.flagEPFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagEPFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "PROJECT_ID")
	for _, s := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", s.ServiceName, s.ProducerProjectId)
	}
	return nil
}

func (g *endpointsCmds) runEPSvcDeploy(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	// Build the ConfigFile list from paths.
	source := &servicemanagement.ConfigSource{}
	var serviceName string
	for _, p := range args {
		data, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("reading %s: %w", p, err)
		}
		ftype, err := epDetectFileType(p, data)
		if err != nil {
			return err
		}
		source.Files = append(source.Files, &servicemanagement.ConfigFile{
			FilePath:     filepath.Base(p),
			FileContents: string(data),
			FileType:     ftype,
		})
		if serviceName == "" {
			serviceName = epExtractServiceName(data)
		}
	}
	if serviceName == "" {
		return fmt.Errorf("could not determine service name from config file(s)")
	}
	// Ensure the ManagedService exists.
	if _, err := svc.Services.Get(serviceName).Context(ctx).Do(); err != nil {
		if !g.flagEPValidate {
			created, err := svc.Services.Create(&servicemanagement.ManagedService{
				ServiceName:       serviceName,
				ProducerProjectId: project,
			}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("creating managed service %q: %w", serviceName, err)
			}
			if _, err := epWaitOp(ctx, svc, created.Name); err != nil {
				return err
			}
		}
	}
	submit, err := svc.Services.Configs.Submit(serviceName, &servicemanagement.SubmitConfigSourceRequest{
		ConfigSource: source,
		ValidateOnly: g.flagEPValidate,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting service config: %w", err)
	}
	final, err := epWaitOp(ctx, svc, submit.Name)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "Service config submitted for %s.\n", serviceName)
	if g.flagEPValidate {
		return emitFormattedTo(g.inv.stdout, final, g.flagEPFormat)
	}
	// Extract the newly generated config ID from the response.
	configID := epConfigIDFromOp(final)
	if configID == "" {
		return fmt.Errorf("could not determine new service config ID from operation response")
	}
	// Roll out 100% traffic to the new config.
	rollout := &servicemanagement.Rollout{
		ServiceName: serviceName,
		TrafficPercentStrategy: &servicemanagement.TrafficPercentStrategy{
			Percentages: map[string]float64{configID: 100.0},
		},
	}
	rolloutOp, err := svc.Services.Rollouts.Create(serviceName, rollout).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating rollout: %w", err)
	}
	if g.flagEPAsync {
		fmt.Fprintf(g.inv.stderr, "Rollout in progress (operation: %s).\n", rolloutOp.Name)
		return emitFormattedTo(g.inv.stdout, rolloutOp, g.flagEPFormat)
	}
	rolloutFinal, err := epWaitOp(ctx, svc, rolloutOp.Name)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "Service configuration [%s] deployed for service [%s].\n", configID, serviceName)
	return emitFormattedTo(g.inv.stdout, rolloutFinal, g.flagEPFormat)
}

func epDetectFileType(path string, data []byte) (string, error) {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".pb"), strings.HasSuffix(lower, ".descriptor"):
		return "FILE_DESCRIPTOR_SET_PROTO", nil
	case strings.HasSuffix(lower, ".proto"):
		return "PROTO_FILE", nil
	case strings.HasSuffix(lower, ".yaml"), strings.HasSuffix(lower, ".yml"), strings.HasSuffix(lower, ".json"):
		// Heuristic: OpenAPI files declare a "swagger" or "openapi" key.
		if bytesContainsAny(data, "swagger", "openapi") {
			return "OPEN_API_YAML", nil
		}
		return "SERVICE_CONFIG_YAML", nil
	}
	return "", fmt.Errorf("cannot determine service config file type for %s (supported: .yaml, .yml, .json, .pb, .descriptor, .proto)", path)
}

func bytesContainsAny(data []byte, needles ...string) bool {
	s := string(data)
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

// epExtractServiceName pulls the service name from a service-config YAML/JSON.
// OpenAPI files carry it in "host"; google.api.Service files carry it in
// "name". For proto descriptors this returns "" and the caller must obtain the
// name elsewhere.
func epExtractServiceName(data []byte) string {
	s := string(data)
	for _, key := range []string{"\nname: ", "\nhost: ", `"name": "`, `"host": "`} {
		if i := strings.Index(s, key); i >= 0 {
			rest := s[i+len(key):]
			end := strings.IndexAny(rest, "\n\r\"")
			if end > 0 {
				return strings.TrimSpace(rest[:end])
			}
		}
	}
	return ""
}

func epConfigIDFromOp(op *servicemanagement.Operation) string {
	if op == nil || op.Response == nil {
		return ""
	}
	// Response is a raw JSON object; parse it as SubmitConfigSourceResponse.
	var resp servicemanagement.SubmitConfigSourceResponse
	if err := epUnmarshalRawMessage(op.Response, &resp); err != nil {
		return ""
	}
	if resp.ServiceConfig != nil {
		return resp.ServiceConfig.Id
	}
	return ""
}

func epUnmarshalRawMessage(raw []byte, dst any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
}

func (g *endpointsCmds) runEPSvcGetConfigName(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Services.GetConfig(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting service config: %w", err)
	}
	if g.flagEPFormat != "" {
		return emitFormattedTo(g.inv.stdout, got, g.flagEPFormat)
	}
	fmt.Fprintln(g.inv.stdout, got.Id)
	return nil
}

// --- enable / disable use serviceusage (endpoints services are consumable services) ---

func (g *endpointsCmds) runEPSvcEnable(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceUsage(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Services.Enable(fmt.Sprintf("projects/%s/services/%s", project, args[0]),
		&su.EnableServiceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling service: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Enable in progress (operation: %s).\n", op.Name)
	return emitFormattedTo(g.inv.stdout, op, g.flagEPFormat)
}

func (g *endpointsCmds) runEPSvcDisable(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ServiceUsage(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Services.Disable(fmt.Sprintf("projects/%s/services/%s", project, args[0]),
		&su.DisableServiceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling service: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Disable in progress (operation: %s).\n", op.Name)
	return emitFormattedTo(g.inv.stdout, op, g.flagEPFormat)
}

// --- IAM ---

func (g *endpointsCmds) runEPSvcGetIAM(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Services.GetIamPolicy(args[0], &servicemanagement.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, policy, g.flagEPFormat)
}

func (g *endpointsCmds) runEPSvcSetIAM(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	policy := &servicemanagement.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	updated, err := svc.Services.SetIamPolicy(args[0], &servicemanagement.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, updated, g.flagEPFormat)
}

func (g *endpointsCmds) runEPSvcAddIAM(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Services.GetIamPolicy(args[0], &servicemanagement.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	epMergeBinding(policy, g.flagEPMember, g.flagEPRole)
	updated, err := svc.Services.SetIamPolicy(args[0], &servicemanagement.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, updated, g.flagEPFormat)
}

func (g *endpointsCmds) runEPSvcRemoveIAM(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Services.GetIamPolicy(args[0], &servicemanagement.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	epRemoveBinding(policy, g.flagEPMember, g.flagEPRole)
	updated, err := svc.Services.SetIamPolicy(args[0], &servicemanagement.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, updated, g.flagEPFormat)
}

func (g *endpointsCmds) runEPSvcCheckIAM(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.endpointsService(ctx)
	if err != nil {
		return err
	}
	// Ask for the full set of servicemanagement.* permissions; the API
	// returns only those actually granted to the caller.
	perms := []string{
		"servicemanagement.services.get",
		"servicemanagement.services.list",
		"servicemanagement.services.create",
		"servicemanagement.services.delete",
		"servicemanagement.services.update",
		"servicemanagement.services.bind",
		"servicemanagement.services.check",
		"servicemanagement.services.report",
		"servicemanagement.services.setIamPolicy",
		"servicemanagement.services.getIamPolicy",
	}
	resp, err := svc.Services.TestIamPermissions(args[0], &servicemanagement.TestIamPermissionsRequest{
		Permissions: perms,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("testing IAM permissions: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagEPFormat)
}

func epMergeBinding(policy *servicemanagement.Policy, member, role string) {
	for _, b := range policy.Bindings {
		if b.Role == role {
			for _, m := range b.Members {
				if m == member {
					return
				}
			}
			b.Members = append(b.Members, member)
			return
		}
	}
	policy.Bindings = append(policy.Bindings, &servicemanagement.Binding{
		Role:    role,
		Members: []string{member},
	})
}

func epRemoveBinding(policy *servicemanagement.Policy, member, role string) {
	for _, b := range policy.Bindings {
		if b.Role != role {
			continue
		}
		out := b.Members[:0]
		for _, m := range b.Members {
			if m != member {
				out = append(out, m)
			}
		}
		b.Members = out
	}
}
