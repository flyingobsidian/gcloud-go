package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	agentregistry "google.golang.org/api/agentregistry/v1alpha"
)

// --- gcloud agent-registry (#290, #788-#793) ---

var agentRegistryCmd = &cobra.Command{Use: "agent-registry", Short: "Manage Agent Registry"}

var (
	flagARLocation   string
	flagARFormat     string
	flagARFilter     string
	flagARConfigFile string
	flagARUpdateMask string
	flagARServiceID  string
	flagARRequestID  string
	flagARAsync      bool
)

func arLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func arChild(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func arResolveParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagARLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return arLocationParent(project, flagARLocation), nil
}

func arWaitOp(ctx context.Context, svc *agentregistry.APIService, op *agentregistry.Operation) (*agentregistry.Operation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
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

func arFinishOp(ctx context.Context, svc *agentregistry.APIService, op *agentregistry.Operation, verb, name string) error {
	if flagARAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := arWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

// --- agents ---

var agentRegistryAgentsCmd = &cobra.Command{Use: "agents", Short: "Manage Agent Registry agents"}

var (
	arAgentDescribeCmd = &cobra.Command{
		Use: "describe AGENT", Short: "Describe an agent",
		Args: cobra.ExactArgs(1), RunE: runARAgentDescribe,
	}
	arAgentListCmd = &cobra.Command{
		Use: "list", Short: "List agents in a location",
		Args: cobra.NoArgs, RunE: runARAgentList,
	}
	arAgentSearchCmd = &cobra.Command{
		Use: "search QUERY", Short: "Search agents",
		Args: cobra.ExactArgs(1), RunE: runARAgentSearch,
	}
)

// --- bindings ---

var agentRegistryBindingsCmd = &cobra.Command{Use: "bindings", Short: "Manage Agent Registry bindings"}

var (
	arBindingCreateCmd = &cobra.Command{
		Use: "create BINDING", Short: "Create a binding from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runARBindingCreate,
	}
	arBindingDeleteCmd = &cobra.Command{
		Use: "delete BINDING", Short: "Delete a binding",
		Args: cobra.ExactArgs(1), RunE: runARBindingDelete,
	}
	arBindingDescribeCmd = &cobra.Command{
		Use: "describe BINDING", Short: "Describe a binding",
		Args: cobra.ExactArgs(1), RunE: runARBindingDescribe,
	}
	arBindingFetchAvailableCmd = &cobra.Command{
		Use: "fetch-available", Short: "Fetch available bindings in a location",
		Args: cobra.NoArgs, RunE: runARBindingFetchAvailable,
	}
	arBindingListCmd = &cobra.Command{
		Use: "list", Short: "List bindings in a location",
		Args: cobra.NoArgs, RunE: runARBindingList,
	}
	arBindingUpdateCmd = &cobra.Command{
		Use: "update BINDING", Short: "Update a binding from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runARBindingUpdate,
	}
)

// --- endpoints ---

var agentRegistryEndpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Agent Registry endpoints"}

var (
	arEndpointDescribeCmd = &cobra.Command{
		Use: "describe ENDPOINT", Short: "Describe an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAREndpointDescribe,
	}
	arEndpointListCmd = &cobra.Command{
		Use: "list", Short: "List endpoints in a location",
		Args: cobra.NoArgs, RunE: runAREndpointList,
	}
)

// --- mcp-servers ---

var agentRegistryMcpServersCmd = &cobra.Command{Use: "mcp-servers", Short: "Manage Agent Registry MCP servers"}

var (
	arMcpDescribeCmd = &cobra.Command{
		Use: "describe SERVER", Short: "Describe an MCP server",
		Args: cobra.ExactArgs(1), RunE: runARMcpDescribe,
	}
	arMcpListCmd = &cobra.Command{
		Use: "list", Short: "List MCP servers in a location",
		Args: cobra.NoArgs, RunE: runARMcpList,
	}
	arMcpSearchCmd = &cobra.Command{
		Use: "search QUERY", Short: "Search MCP servers",
		Args: cobra.ExactArgs(1), RunE: runARMcpSearch,
	}
)

// --- operations ---

var agentRegistryOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Agent Registry operations"}

var (
	arOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: runAROpCancel,
	}
	arOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: runAROpDelete,
	}
	arOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runAROpDescribe,
	}
	arOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: runAROpList,
	}
	arOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for an operation to complete",
		Args: cobra.ExactArgs(1), RunE: runAROpWait,
	}
)

// --- services ---

var agentRegistryServicesCmd = &cobra.Command{Use: "services", Short: "Manage Agent Registry services"}

var (
	arSvcCreateCmd = &cobra.Command{
		Use: "create SERVICE", Short: "Create a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runARSvcCreate,
	}
	arSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a service",
		Args: cobra.ExactArgs(1), RunE: runARSvcDelete,
	}
	arSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: runARSvcDescribe,
	}
	arSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services in a location",
		Args: cobra.NoArgs, RunE: runARSvcList,
	}
	arSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runARSvcUpdate,
	}
)

func init() {
	addARFlags := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagARLocation, "location", "", "Location (required)")
		}
	}
	addFormatFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagARFormat, "format", "", "Output format")
		}
	}
	addFilterFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagARFilter, "filter", "", "Server-side list filter")
		}
	}

	// agents
	addARFlags(arAgentDescribeCmd, arAgentListCmd, arAgentSearchCmd)
	addFormatFlag(arAgentDescribeCmd, arAgentListCmd, arAgentSearchCmd)
	addFilterFlag(arAgentListCmd)
	agentRegistryAgentsCmd.AddCommand(arAgentDescribeCmd, arAgentListCmd, arAgentSearchCmd)
	agentRegistryCmd.AddCommand(agentRegistryAgentsCmd)

	// bindings
	addARFlags(arBindingCreateCmd, arBindingDeleteCmd, arBindingDescribeCmd,
		arBindingFetchAvailableCmd, arBindingListCmd, arBindingUpdateCmd)
	addFormatFlag(arBindingDescribeCmd, arBindingListCmd, arBindingFetchAvailableCmd)
	addFilterFlag(arBindingListCmd)
	arBindingCreateCmd.Flags().StringVar(&flagARConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Binding body (required)")
	_ = arBindingCreateCmd.MarkFlagRequired("config-file")
	arBindingCreateCmd.Flags().StringVar(&flagARRequestID, "request-id", "", "Optional idempotency request ID")
	arBindingCreateCmd.Flags().BoolVar(&flagARAsync, "async", false, "Do not wait for the operation to complete")
	arBindingDeleteCmd.Flags().StringVar(&flagARRequestID, "request-id", "", "Optional idempotency request ID")
	arBindingDeleteCmd.Flags().BoolVar(&flagARAsync, "async", false, "Do not wait for the operation to complete")
	arBindingUpdateCmd.Flags().StringVar(&flagARConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Binding body (required)")
	_ = arBindingUpdateCmd.MarkFlagRequired("config-file")
	arBindingUpdateCmd.Flags().StringVar(&flagARUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	arBindingUpdateCmd.Flags().StringVar(&flagARRequestID, "request-id", "", "Optional idempotency request ID")
	arBindingUpdateCmd.Flags().BoolVar(&flagARAsync, "async", false, "Do not wait for the operation to complete")
	agentRegistryBindingsCmd.AddCommand(arBindingCreateCmd, arBindingDeleteCmd, arBindingDescribeCmd,
		arBindingFetchAvailableCmd, arBindingListCmd, arBindingUpdateCmd)
	agentRegistryCmd.AddCommand(agentRegistryBindingsCmd)

	// endpoints
	addARFlags(arEndpointDescribeCmd, arEndpointListCmd)
	addFormatFlag(arEndpointDescribeCmd, arEndpointListCmd)
	addFilterFlag(arEndpointListCmd)
	agentRegistryEndpointsCmd.AddCommand(arEndpointDescribeCmd, arEndpointListCmd)
	agentRegistryCmd.AddCommand(agentRegistryEndpointsCmd)

	// mcp-servers
	addARFlags(arMcpDescribeCmd, arMcpListCmd, arMcpSearchCmd)
	addFormatFlag(arMcpDescribeCmd, arMcpListCmd, arMcpSearchCmd)
	addFilterFlag(arMcpListCmd)
	agentRegistryMcpServersCmd.AddCommand(arMcpDescribeCmd, arMcpListCmd, arMcpSearchCmd)
	agentRegistryCmd.AddCommand(agentRegistryMcpServersCmd)

	// operations
	addARFlags(arOpCancelCmd, arOpDeleteCmd, arOpDescribeCmd, arOpListCmd, arOpWaitCmd)
	addFormatFlag(arOpDescribeCmd, arOpListCmd, arOpWaitCmd)
	addFilterFlag(arOpListCmd)
	agentRegistryOperationsCmd.AddCommand(arOpCancelCmd, arOpDeleteCmd, arOpDescribeCmd, arOpListCmd, arOpWaitCmd)
	agentRegistryCmd.AddCommand(agentRegistryOperationsCmd)

	// services
	addARFlags(arSvcCreateCmd, arSvcDeleteCmd, arSvcDescribeCmd, arSvcListCmd, arSvcUpdateCmd)
	addFormatFlag(arSvcDescribeCmd, arSvcListCmd)
	addFilterFlag(arSvcListCmd)
	arSvcCreateCmd.Flags().StringVar(&flagARConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Service body (required)")
	_ = arSvcCreateCmd.MarkFlagRequired("config-file")
	arSvcCreateCmd.Flags().StringVar(&flagARServiceID, "service-id", "", "Optional server-supplied ID")
	arSvcCreateCmd.Flags().StringVar(&flagARRequestID, "request-id", "", "Optional idempotency request ID")
	arSvcCreateCmd.Flags().BoolVar(&flagARAsync, "async", false, "Do not wait for the operation to complete")
	arSvcDeleteCmd.Flags().StringVar(&flagARRequestID, "request-id", "", "Optional idempotency request ID")
	arSvcDeleteCmd.Flags().BoolVar(&flagARAsync, "async", false, "Do not wait for the operation to complete")
	arSvcUpdateCmd.Flags().StringVar(&flagARConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Service body (required)")
	_ = arSvcUpdateCmd.MarkFlagRequired("config-file")
	arSvcUpdateCmd.Flags().StringVar(&flagARUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	arSvcUpdateCmd.Flags().StringVar(&flagARRequestID, "request-id", "", "Optional idempotency request ID")
	arSvcUpdateCmd.Flags().BoolVar(&flagARAsync, "async", false, "Do not wait for the operation to complete")
	agentRegistryServicesCmd.AddCommand(arSvcCreateCmd, arSvcDeleteCmd, arSvcDescribeCmd, arSvcListCmd, arSvcUpdateCmd)
	agentRegistryCmd.AddCommand(agentRegistryServicesCmd)

	rootCmd.AddCommand(agentRegistryCmd)
}

// --- agents impl ---

func runARAgentDescribe(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "agents", args[0])
	got, err := svc.Projects.Locations.Agents.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing agent: %w", err)
	}
	return emitFormatted(got, flagARFormat)
}

func runARAgentList(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*agentregistry.Agent
	pageToken := ""
	for {
		call := svc.Projects.Locations.Agents.List(parent).Context(ctx)
		if flagARFilter != "" {
			call = call.Filter(flagARFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing agents: %w", err)
		}
		all = append(all, resp.Agents...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagARFormat != "" {
		return emitFormatted(all, flagARFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, a := range all {
		fmt.Printf("%-40s %s\n", path.Base(a.Name), a.DisplayName)
	}
	return nil
}

func runARAgentSearch(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &agentregistry.SearchAgentsRequest{SearchString: args[0]}
	resp, err := svc.Projects.Locations.Agents.Search(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("searching agents: %w", err)
	}
	return emitFormatted(resp, flagARFormat)
}

// --- bindings impl ---

func runARBindingCreate(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	body := &agentregistry.Binding{}
	if err := loadYAMLOrJSONInto(flagARConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Bindings.Create(parent, body).BindingId(args[0]).Context(ctx)
	if flagARRequestID != "" {
		call = call.RequestId(flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating binding: %w", err)
	}
	return arFinishOp(ctx, svc, op, "Create binding", args[0])
}

func runARBindingDelete(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "bindings", args[0])
	call := svc.Projects.Locations.Bindings.Delete(name).Context(ctx)
	if flagARRequestID != "" {
		call = call.RequestId(flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting binding: %w", err)
	}
	return arFinishOp(ctx, svc, op, "Delete binding", args[0])
}

func runARBindingDescribe(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "bindings", args[0])
	got, err := svc.Projects.Locations.Bindings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing binding: %w", err)
	}
	return emitFormatted(got, flagARFormat)
}

func runARBindingFetchAvailable(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var results []*agentregistry.Binding
	pageToken := ""
	for {
		call := svc.Projects.Locations.Bindings.FetchAvailable(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("fetching available bindings: %w", err)
		}
		results = append(results, resp.Bindings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(results, flagARFormat)
}

func runARBindingList(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*agentregistry.Binding
	pageToken := ""
	for {
		call := svc.Projects.Locations.Bindings.List(parent).Context(ctx)
		if flagARFilter != "" {
			call = call.Filter(flagARFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing bindings: %w", err)
		}
		all = append(all, resp.Bindings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagARFormat != "" {
		return emitFormatted(all, flagARFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, b := range all {
		fmt.Printf("%-40s %s\n", path.Base(b.Name), b.DisplayName)
	}
	return nil
}

func runARBindingUpdate(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	body := &agentregistry.Binding{}
	if err := loadYAMLOrJSONInto(flagARConfigFile, body); err != nil {
		return err
	}
	mask := flagARUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "bindings", args[0])
	call := svc.Projects.Locations.Bindings.Patch(name, body).UpdateMask(mask).Context(ctx)
	if flagARRequestID != "" {
		call = call.RequestId(flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating binding: %w", err)
	}
	return arFinishOp(ctx, svc, op, "Update binding", args[0])
}

// --- endpoints impl ---

func runAREndpointDescribe(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "endpoints", args[0])
	got, err := svc.Projects.Locations.Endpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint: %w", err)
	}
	return emitFormatted(got, flagARFormat)
}

func runAREndpointList(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*agentregistry.Endpoint
	pageToken := ""
	for {
		call := svc.Projects.Locations.Endpoints.List(parent).Context(ctx)
		if flagARFilter != "" {
			call = call.Filter(flagARFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing endpoints: %w", err)
		}
		all = append(all, resp.Endpoints...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagARFormat != "" {
		return emitFormatted(all, flagARFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, e := range all {
		fmt.Printf("%-40s\n", path.Base(e.Name))
	}
	return nil
}

// --- mcp-servers impl ---

func runARMcpDescribe(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "mcpServers", args[0])
	got, err := svc.Projects.Locations.McpServers.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing MCP server: %w", err)
	}
	return emitFormatted(got, flagARFormat)
}

func runARMcpList(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*agentregistry.McpServer
	pageToken := ""
	for {
		call := svc.Projects.Locations.McpServers.List(parent).Context(ctx)
		if flagARFilter != "" {
			call = call.Filter(flagARFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing MCP servers: %w", err)
		}
		all = append(all, resp.McpServers...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagARFormat != "" {
		return emitFormatted(all, flagARFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, m := range all {
		fmt.Printf("%-40s %s\n", path.Base(m.Name), m.DisplayName)
	}
	return nil
}

func runARMcpSearch(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &agentregistry.SearchMcpServersRequest{SearchString: args[0]}
	resp, err := svc.Projects.Locations.McpServers.Search(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("searching MCP servers: %w", err)
	}
	return emitFormatted(resp, flagARFormat)
}

// --- operations impl ---

func runAROpCancel(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "operations", args[0])
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &agentregistry.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Cancel request issued for operation %s.\n", args[0])
	return nil
}

func runAROpDelete(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "operations", args[0])
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Deleted operation %s.\n", args[0])
	return nil
}

func runAROpDescribe(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "operations", args[0])
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagARFormat)
}

func runAROpList(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*agentregistry.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagARFilter != "" {
			call = call.Filter(flagARFilter)
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
	if flagARFormat != "" {
		return emitFormatted(all, flagARFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func runAROpWait(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "operations", args[0])
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting operation: %w", err)
	}
	final, err := arWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagARFormat)
}

// --- services impl ---

func runARSvcCreate(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	body := &agentregistry.Service{}
	if err := loadYAMLOrJSONInto(flagARConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	id := flagARServiceID
	if id == "" {
		id = args[0]
	}
	call := svc.Projects.Locations.Services.Create(parent, body).ServiceId(id).Context(ctx)
	if flagARRequestID != "" {
		call = call.RequestId(flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}
	return arFinishOp(ctx, svc, op, "Create service", args[0])
}

func runARSvcDelete(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "services", args[0])
	call := svc.Projects.Locations.Services.Delete(name).Context(ctx)
	if flagARRequestID != "" {
		call = call.RequestId(flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	return arFinishOp(ctx, svc, op, "Delete service", args[0])
}

func runARSvcDescribe(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "services", args[0])
	got, err := svc.Projects.Locations.Services.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormatted(got, flagARFormat)
}

func runARSvcList(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*agentregistry.Service
	pageToken := ""
	for {
		call := svc.Projects.Locations.Services.List(parent).Context(ctx)
		if flagARFilter != "" {
			call = call.Filter(flagARFilter)
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
	if flagARFormat != "" {
		return emitFormatted(all, flagARFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, s := range all {
		fmt.Printf("%-40s %s\n", path.Base(s.Name), s.DisplayName)
	}
	return nil
}

func runARSvcUpdate(cmd *cobra.Command, args []string) error {
	parent, err := arResolveParent()
	if err != nil {
		return err
	}
	body := &agentregistry.Service{}
	if err := loadYAMLOrJSONInto(flagARConfigFile, body); err != nil {
		return err
	}
	mask := flagARUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AgentRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := arChild(parent, "services", args[0])
	call := svc.Projects.Locations.Services.Patch(name, body).UpdateMask(mask).Context(ctx)
	if flagARRequestID != "" {
		call = call.RequestId(flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return arFinishOp(ctx, svc, op, "Update service", args[0])
}
