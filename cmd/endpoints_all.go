package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	servicemanagement "google.golang.org/api/servicemanagement/v1"
	su "google.golang.org/api/serviceusage/v1"
)

// --- Shared helpers for the Endpoints surface (#892-#894) ---

var (
	flagEPFormat     string
	flagEPService    string
	flagEPAsync      bool
	flagEPValidate   bool
	flagEPForce      bool
	flagEPMember     string
	flagEPRole       string
	flagEPPolicyFile string
	flagEPPageSize   int64
)

func endpointsService(ctx context.Context) (*servicemanagement.APIService, error) {
	return gcp.ServiceManagementService(ctx, flagAccount)
}

// --- Subgroup command objects ---

var (
	endpointsConfigsCmd    = &cobra.Command{Use: "configs", Short: "View service configurations"}
	endpointsOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Service Management operations"}
	endpointsServicesCmd   = &cobra.Command{Use: "services", Short: "Manage Endpoints services"}
)

// --- configs ---

var (
	epConfigDescribeCmd = &cobra.Command{
		Use: "describe CONFIG_ID", Short: "Describe a service configuration",
		Args: cobra.ExactArgs(1), RunE: runEPConfigDescribe,
	}
	epConfigListCmd = &cobra.Command{
		Use: "list", Short: "List configurations for a service",
		Args: cobra.NoArgs, RunE: runEPConfigList,
	}
)

// --- operations ---

var (
	epOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Service Management operation",
		Args: cobra.ExactArgs(1), RunE: runEPOpDescribe,
	}
	epOpListCmd = &cobra.Command{
		Use: "list", Short: "List Service Management operations",
		Args: cobra.NoArgs, RunE: runEPOpList,
	}
	epOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Service Management operation to complete",
		Args: cobra.ExactArgs(1), RunE: runEPOpWait,
	}
)

// --- services ---

var (
	epSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete an Endpoints service",
		Args: cobra.ExactArgs(1), RunE: runEPSvcDelete,
	}
	epSvcUndeployCmd = &cobra.Command{
		Use: "undeploy SERVICE", Short: "Undeploy an Endpoints service (alias for delete)",
		Args: cobra.ExactArgs(1), RunE: runEPSvcDelete,
	}
	epSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe an Endpoints service",
		Args: cobra.ExactArgs(1), RunE: runEPSvcDescribe,
	}
	epSvcListCmd = &cobra.Command{
		Use: "list", Short: "List Endpoints services owned by the current project",
		Args: cobra.NoArgs, RunE: runEPSvcList,
	}
	epSvcDeployCmd = &cobra.Command{
		Use: "deploy CONFIG_FILE [CONFIG_FILE ...]", Short: "Deploy a service configuration",
		Args: cobra.MinimumNArgs(1), RunE: runEPSvcDeploy,
	}
	epSvcGetConfigNameCmd = &cobra.Command{
		Use: "get-config-name SERVICE", Short: "Print the active service configuration ID",
		Args: cobra.ExactArgs(1), RunE: runEPSvcGetConfigName,
	}
	epSvcEnableCmd = &cobra.Command{
		Use: "enable SERVICE", Short: "Enable an Endpoints service for the project",
		Args: cobra.ExactArgs(1), RunE: runEPSvcEnable,
	}
	epSvcDisableCmd = &cobra.Command{
		Use: "disable SERVICE", Short: "Disable an Endpoints service for the project",
		Args: cobra.ExactArgs(1), RunE: runEPSvcDisable,
	}
	epSvcAddIAMCmd = &cobra.Command{
		Use: "add-iam-policy-binding SERVICE", Short: "Add an IAM policy binding to a service",
		Args: cobra.ExactArgs(1), RunE: runEPSvcAddIAM,
	}
	epSvcRemoveIAMCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SERVICE", Short: "Remove an IAM policy binding from a service",
		Args: cobra.ExactArgs(1), RunE: runEPSvcRemoveIAM,
	}
	epSvcGetIAMCmd = &cobra.Command{
		Use: "get-iam-policy SERVICE", Short: "Get the IAM policy for a service",
		Args: cobra.ExactArgs(1), RunE: runEPSvcGetIAM,
	}
	epSvcSetIAMCmd = &cobra.Command{
		Use: "set-iam-policy SERVICE POLICY_FILE", Short: "Set the IAM policy for a service",
		Args: cobra.ExactArgs(2), RunE: runEPSvcSetIAM,
	}
	epSvcCheckIAMCmd = &cobra.Command{
		Use: "check-iam-policy SERVICE", Short: "Check IAM permissions on a service",
		Args: cobra.ExactArgs(1), RunE: runEPSvcCheckIAM,
	}
)

func init() {
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagEPFormat, "format", "", "Output format")
		}
	}

	// configs
	epConfigDescribeCmd.Flags().StringVar(&flagEPService, "service", "", "Service name (required)")
	_ = epConfigDescribeCmd.MarkFlagRequired("service")
	epConfigListCmd.Flags().StringVar(&flagEPService, "service", "", "Service name (required)")
	_ = epConfigListCmd.MarkFlagRequired("service")
	epConfigListCmd.Flags().Int64Var(&flagEPPageSize, "page-size", 0, "Page size for API pagination")
	addFmt(epConfigDescribeCmd, epConfigListCmd)
	endpointsConfigsCmd.AddCommand(epConfigDescribeCmd, epConfigListCmd)
	endpointsCmd.AddCommand(endpointsConfigsCmd)

	// operations
	addFmt(epOpDescribeCmd, epOpListCmd, epOpWaitCmd)
	epOpListCmd.Flags().Int64Var(&flagEPPageSize, "page-size", 0, "Page size for API pagination")
	endpointsOperationsCmd.AddCommand(epOpDescribeCmd, epOpListCmd, epOpWaitCmd)
	endpointsCmd.AddCommand(endpointsOperationsCmd)

	// services
	epSvcDeleteCmd.Flags().BoolVar(&flagEPAsync, "async", false, "Do not wait for the operation to finish")
	epSvcUndeployCmd.Flags().BoolVar(&flagEPAsync, "async", false, "Do not wait for the operation to finish")
	epSvcDeployCmd.Flags().BoolVar(&flagEPAsync, "async", false, "Do not wait for the operation to finish")
	epSvcDeployCmd.Flags().BoolVar(&flagEPValidate, "validate-only", false, "Validate the configuration only; do not deploy")
	epSvcDeployCmd.Flags().BoolVar(&flagEPForce, "force", false, "Force deployment even when hazardous changes are detected")
	for _, c := range []*cobra.Command{epSvcAddIAMCmd, epSvcRemoveIAMCmd} {
		c.Flags().StringVar(&flagEPMember, "member", "", "IAM member, e.g. user:alice@example.com (required)")
		c.Flags().StringVar(&flagEPRole, "role", "", "IAM role, e.g. roles/servicemanagement.serviceConsumer (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	addFmt(epSvcDescribeCmd, epSvcListCmd, epSvcDeployCmd, epSvcGetConfigNameCmd,
		epSvcGetIAMCmd, epSvcSetIAMCmd, epSvcCheckIAMCmd)
	endpointsServicesCmd.AddCommand(
		epSvcDeleteCmd, epSvcUndeployCmd, epSvcDescribeCmd, epSvcListCmd, epSvcDeployCmd,
		epSvcGetConfigNameCmd, epSvcEnableCmd, epSvcDisableCmd,
		epSvcAddIAMCmd, epSvcRemoveIAMCmd, epSvcGetIAMCmd, epSvcSetIAMCmd, epSvcCheckIAMCmd,
	)
	endpointsCmd.AddCommand(endpointsServicesCmd)
}

// --- configs impl ---

func runEPConfigDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Services.Configs.Get(flagEPService, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service config: %w", err)
	}
	return emitFormatted(got, flagEPFormat)
}

func runEPConfigList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	var all []*servicemanagement.Service
	pageToken := ""
	for {
		call := svc.Services.Configs.List(flagEPService).Context(ctx)
		if flagEPPageSize > 0 {
			call = call.PageSize(flagEPPageSize)
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
	if flagEPFormat != "" {
		return emitFormatted(all, flagEPFormat)
	}
	fmt.Printf("%-30s %s\n", "CONFIG_ID", "SERVICE_NAME")
	for _, c := range all {
		fmt.Printf("%-30s %s\n", c.Id, c.Name)
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

func runEPOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Operations.Get(epOperationsName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagEPFormat)
}

func runEPOpList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	var all []*servicemanagement.Operation
	pageToken := ""
	for {
		call := svc.Operations.List().Context(ctx)
		if flagEPPageSize > 0 {
			call = call.PageSize(flagEPPageSize)
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
	if flagEPFormat != "" {
		return emitFormatted(all, flagEPFormat)
	}
	fmt.Printf("%-60s %-6s\n", "NAME", "DONE")
	for _, op := range all {
		fmt.Printf("%-60s %-6t\n", op.Name, op.Done)
	}
	return nil
}

func runEPOpWait(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	final, err := epWaitOp(ctx, svc, epOperationsName(args[0]))
	if err != nil {
		return err
	}
	return emitFormatted(final, flagEPFormat)
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

func runEPSvcDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Services.Delete(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	if flagEPAsync {
		fmt.Fprintf(os.Stderr, "Delete in progress (operation: %s).\n", op.Name)
		return emitFormatted(op, flagEPFormat)
	}
	final, err := epWaitOp(ctx, svc, op.Name)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagEPFormat)
}

func runEPSvcDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Services.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormatted(got, flagEPFormat)
}

func runEPSvcList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := endpointsService(ctx)
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
	if flagEPFormat != "" {
		return emitFormatted(all, flagEPFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "PROJECT_ID")
	for _, s := range all {
		fmt.Printf("%-40s %s\n", s.ServiceName, s.ProducerProjectId)
	}
	return nil
}

func runEPSvcDeploy(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := endpointsService(ctx)
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
		if !flagEPValidate {
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
		ValidateOnly: flagEPValidate,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("submitting service config: %w", err)
	}
	final, err := epWaitOp(ctx, svc, submit.Name)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Service config submitted for %s.\n", serviceName)
	if flagEPValidate {
		return emitFormatted(final, flagEPFormat)
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
	if flagEPAsync {
		fmt.Fprintf(os.Stderr, "Rollout in progress (operation: %s).\n", rolloutOp.Name)
		return emitFormatted(rolloutOp, flagEPFormat)
	}
	rolloutFinal, err := epWaitOp(ctx, svc, rolloutOp.Name)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Service configuration [%s] deployed for service [%s].\n", configID, serviceName)
	return emitFormatted(rolloutFinal, flagEPFormat)
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

func runEPSvcGetConfigName(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Services.GetConfig(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting service config: %w", err)
	}
	if flagEPFormat != "" {
		return emitFormatted(got, flagEPFormat)
	}
	fmt.Println(got.Id)
	return nil
}

// --- enable / disable use serviceusage (endpoints services are consumable services) ---

func runEPSvcEnable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceUsageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Services.Enable(fmt.Sprintf("projects/%s/services/%s", project, args[0]),
		&su.EnableServiceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling service: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Enable in progress (operation: %s).\n", op.Name)
	return emitFormatted(op, flagEPFormat)
}

func runEPSvcDisable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ServiceUsageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Services.Disable(fmt.Sprintf("projects/%s/services/%s", project, args[0]),
		&su.DisableServiceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling service: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Disable in progress (operation: %s).\n", op.Name)
	return emitFormatted(op, flagEPFormat)
}

// --- IAM ---

func runEPSvcGetIAM(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Services.GetIamPolicy(args[0], &servicemanagement.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagEPFormat)
}

func runEPSvcSetIAM(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
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
	return emitFormatted(updated, flagEPFormat)
}

func runEPSvcAddIAM(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Services.GetIamPolicy(args[0], &servicemanagement.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	epMergeBinding(policy, flagEPMember, flagEPRole)
	updated, err := svc.Services.SetIamPolicy(args[0], &servicemanagement.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(updated, flagEPFormat)
}

func runEPSvcRemoveIAM(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Services.GetIamPolicy(args[0], &servicemanagement.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	epRemoveBinding(policy, flagEPMember, flagEPRole)
	updated, err := svc.Services.SetIamPolicy(args[0], &servicemanagement.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(updated, flagEPFormat)
}

func runEPSvcCheckIAM(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := endpointsService(ctx)
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
	return emitFormatted(resp, flagEPFormat)
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
