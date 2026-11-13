package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	agentregistry "google.golang.org/api/agentregistry/v1alpha"
)

// --- gcloud agent-registry (#290, #788-#793) ---

func arLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func arChild(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func (g *agentRegistryCmds) arResolveParent() (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	if g.flagARLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return arLocationParent(project, g.flagARLocation), nil
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

func (g *agentRegistryCmds) arFinishOp(ctx context.Context, svc *agentregistry.APIService, op *agentregistry.Operation, verb, name string) error {
	if g.flagARAsync {
		fmt.Fprintf(g.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(g.inv.stdout, op, "")
	}
	final, err := arWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(g.inv.stdout, final.Response, "")
	}
	return nil
}

// --- agents ---

// --- bindings ---

// --- endpoints ---

// --- mcp-servers ---

// --- operations ---

// --- services ---

func (g *agentRegistryCmds) initAgentRegistry() {
	addARFlags := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&g.flagARLocation, "location", "", "Location (required)")
		}
	}
	addFormatFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&g.flagARFormat, "format", "", "Output format")
		}
	}
	addFilterFlag := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&g.flagARFilter, "filter", "", "Server-side list filter")
		}
	}

	// agents
	addARFlags(g.arAgentDescribeCmd, g.arAgentListCmd, g.arAgentSearchCmd)
	addFormatFlag(g.arAgentDescribeCmd, g.arAgentListCmd, g.arAgentSearchCmd)
	addFilterFlag(g.arAgentListCmd)
	g.agentRegistryAgentsCmd.AddCommand(g.arAgentDescribeCmd, g.arAgentListCmd, g.arAgentSearchCmd)
	g.agentRegistryCmd.AddCommand(g.agentRegistryAgentsCmd)

	// bindings
	addARFlags(g.arBindingCreateCmd, g.arBindingDeleteCmd, g.arBindingDescribeCmd,
		g.arBindingFetchAvailableCmd, g.arBindingListCmd, g.arBindingUpdateCmd)
	addFormatFlag(g.arBindingDescribeCmd, g.arBindingListCmd, g.arBindingFetchAvailableCmd)
	addFilterFlag(g.arBindingListCmd)
	g.arBindingCreateCmd.Flags().StringVar(&g.flagARConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Binding body (required)")
	_ = g.arBindingCreateCmd.MarkFlagRequired("config-file")
	g.arBindingCreateCmd.Flags().StringVar(&g.flagARRequestID, "request-id", "", "Optional idempotency request ID")
	g.arBindingCreateCmd.Flags().BoolVar(&g.flagARAsync, "async", false, "Do not wait for the operation to complete")
	g.arBindingDeleteCmd.Flags().StringVar(&g.flagARRequestID, "request-id", "", "Optional idempotency request ID")
	g.arBindingDeleteCmd.Flags().BoolVar(&g.flagARAsync, "async", false, "Do not wait for the operation to complete")
	g.arBindingUpdateCmd.Flags().StringVar(&g.flagARConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Binding body (required)")
	_ = g.arBindingUpdateCmd.MarkFlagRequired("config-file")
	g.arBindingUpdateCmd.Flags().StringVar(&g.flagARUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	g.arBindingUpdateCmd.Flags().StringVar(&g.flagARRequestID, "request-id", "", "Optional idempotency request ID")
	g.arBindingUpdateCmd.Flags().BoolVar(&g.flagARAsync, "async", false, "Do not wait for the operation to complete")
	g.agentRegistryBindingsCmd.AddCommand(g.arBindingCreateCmd, g.arBindingDeleteCmd, g.arBindingDescribeCmd,
		g.arBindingFetchAvailableCmd, g.arBindingListCmd, g.arBindingUpdateCmd)
	g.agentRegistryCmd.AddCommand(g.agentRegistryBindingsCmd)

	// endpoints
	addARFlags(g.arEndpointDescribeCmd, g.arEndpointListCmd)
	addFormatFlag(g.arEndpointDescribeCmd, g.arEndpointListCmd)
	addFilterFlag(g.arEndpointListCmd)
	g.agentRegistryEndpointsCmd.AddCommand(g.arEndpointDescribeCmd, g.arEndpointListCmd)
	g.agentRegistryCmd.AddCommand(g.agentRegistryEndpointsCmd)

	// mcp-servers
	addARFlags(g.arMcpDescribeCmd, g.arMcpListCmd, g.arMcpSearchCmd)
	addFormatFlag(g.arMcpDescribeCmd, g.arMcpListCmd, g.arMcpSearchCmd)
	addFilterFlag(g.arMcpListCmd)
	g.agentRegistryMcpServersCmd.AddCommand(g.arMcpDescribeCmd, g.arMcpListCmd, g.arMcpSearchCmd)
	g.agentRegistryCmd.AddCommand(g.agentRegistryMcpServersCmd)

	// operations
	addARFlags(g.arOpCancelCmd, g.arOpDeleteCmd, g.arOpDescribeCmd, g.arOpListCmd, g.arOpWaitCmd)
	addFormatFlag(g.arOpDescribeCmd, g.arOpListCmd, g.arOpWaitCmd)
	addFilterFlag(g.arOpListCmd)
	g.agentRegistryOperationsCmd.AddCommand(g.arOpCancelCmd, g.arOpDeleteCmd, g.arOpDescribeCmd, g.arOpListCmd, g.arOpWaitCmd)
	g.agentRegistryCmd.AddCommand(g.agentRegistryOperationsCmd)

	// services
	addARFlags(g.arSvcCreateCmd, g.arSvcDeleteCmd, g.arSvcDescribeCmd, g.arSvcListCmd, g.arSvcUpdateCmd)
	addFormatFlag(g.arSvcDescribeCmd, g.arSvcListCmd)
	addFilterFlag(g.arSvcListCmd)
	g.arSvcCreateCmd.Flags().StringVar(&g.flagARConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Service body (required)")
	_ = g.arSvcCreateCmd.MarkFlagRequired("config-file")
	g.arSvcCreateCmd.Flags().StringVar(&g.flagARServiceID, "service-id", "", "Optional server-supplied ID")
	g.arSvcCreateCmd.Flags().StringVar(&g.flagARRequestID, "request-id", "", "Optional idempotency request ID")
	g.arSvcCreateCmd.Flags().BoolVar(&g.flagARAsync, "async", false, "Do not wait for the operation to complete")
	g.arSvcDeleteCmd.Flags().StringVar(&g.flagARRequestID, "request-id", "", "Optional idempotency request ID")
	g.arSvcDeleteCmd.Flags().BoolVar(&g.flagARAsync, "async", false, "Do not wait for the operation to complete")
	g.arSvcUpdateCmd.Flags().StringVar(&g.flagARConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Service body (required)")
	_ = g.arSvcUpdateCmd.MarkFlagRequired("config-file")
	g.arSvcUpdateCmd.Flags().StringVar(&g.flagARUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	g.arSvcUpdateCmd.Flags().StringVar(&g.flagARRequestID, "request-id", "", "Optional idempotency request ID")
	g.arSvcUpdateCmd.Flags().BoolVar(&g.flagARAsync, "async", false, "Do not wait for the operation to complete")
	g.agentRegistryServicesCmd.AddCommand(g.arSvcCreateCmd, g.arSvcDeleteCmd, g.arSvcDescribeCmd, g.arSvcListCmd, g.arSvcUpdateCmd)
	g.agentRegistryCmd.AddCommand(g.agentRegistryServicesCmd)

}

// --- agents impl ---

func (g *agentRegistryCmds) runARAgentDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "agents", args[0])
	got, err := svc.Projects.Locations.Agents.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing agent: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagARFormat)
}

func (g *agentRegistryCmds) runARAgentList(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	var all []*agentregistry.Agent
	pageToken := ""
	for {
		call := svc.Projects.Locations.Agents.List(parent).Context(ctx)
		if g.flagARFilter != "" {
			call = call.Filter(g.flagARFilter)
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
	if g.flagARFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagARFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, a := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(a.Name), a.DisplayName)
	}
	return nil
}

func (g *agentRegistryCmds) runARAgentSearch(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	req := &agentregistry.SearchAgentsRequest{SearchString: args[0]}
	resp, err := svc.Projects.Locations.Agents.Search(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("searching agents: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagARFormat)
}

// --- bindings impl ---

func (g *agentRegistryCmds) runARBindingCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	body := &agentregistry.Binding{}
	if err := loadYAMLOrJSONInto(g.flagARConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Bindings.Create(parent, body).BindingId(args[0]).Context(ctx)
	if g.flagARRequestID != "" {
		call = call.RequestId(g.flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating binding: %w", err)
	}
	return g.arFinishOp(ctx, svc, op, "Create binding", args[0])
}

func (g *agentRegistryCmds) runARBindingDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "bindings", args[0])
	call := svc.Projects.Locations.Bindings.Delete(name).Context(ctx)
	if g.flagARRequestID != "" {
		call = call.RequestId(g.flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting binding: %w", err)
	}
	return g.arFinishOp(ctx, svc, op, "Delete binding", args[0])
}

func (g *agentRegistryCmds) runARBindingDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "bindings", args[0])
	got, err := svc.Projects.Locations.Bindings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing binding: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagARFormat)
}

func (g *agentRegistryCmds) runARBindingFetchAvailable(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
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
	return emitFormattedTo(g.inv.stdout, results, g.flagARFormat)
}

func (g *agentRegistryCmds) runARBindingList(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	var all []*agentregistry.Binding
	pageToken := ""
	for {
		call := svc.Projects.Locations.Bindings.List(parent).Context(ctx)
		if g.flagARFilter != "" {
			call = call.Filter(g.flagARFilter)
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
	if g.flagARFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagARFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, b := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(b.Name), b.DisplayName)
	}
	return nil
}

func (g *agentRegistryCmds) runARBindingUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	body := &agentregistry.Binding{}
	if err := loadYAMLOrJSONInto(g.flagARConfigFile, body); err != nil {
		return err
	}
	mask := g.flagARUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "bindings", args[0])
	call := svc.Projects.Locations.Bindings.Patch(name, body).UpdateMask(mask).Context(ctx)
	if g.flagARRequestID != "" {
		call = call.RequestId(g.flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating binding: %w", err)
	}
	return g.arFinishOp(ctx, svc, op, "Update binding", args[0])
}

// --- endpoints impl ---

func (g *agentRegistryCmds) runAREndpointDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "endpoints", args[0])
	got, err := svc.Projects.Locations.Endpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagARFormat)
}

func (g *agentRegistryCmds) runAREndpointList(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	var all []*agentregistry.Endpoint
	pageToken := ""
	for {
		call := svc.Projects.Locations.Endpoints.List(parent).Context(ctx)
		if g.flagARFilter != "" {
			call = call.Filter(g.flagARFilter)
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
	if g.flagARFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagARFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s\n", "NAME")
	for _, e := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s\n", path.Base(e.Name))
	}
	return nil
}

// --- mcp-servers impl ---

func (g *agentRegistryCmds) runARMcpDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "mcpServers", args[0])
	got, err := svc.Projects.Locations.McpServers.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing MCP server: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagARFormat)
}

func (g *agentRegistryCmds) runARMcpList(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	var all []*agentregistry.McpServer
	pageToken := ""
	for {
		call := svc.Projects.Locations.McpServers.List(parent).Context(ctx)
		if g.flagARFilter != "" {
			call = call.Filter(g.flagARFilter)
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
	if g.flagARFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagARFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, m := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(m.Name), m.DisplayName)
	}
	return nil
}

func (g *agentRegistryCmds) runARMcpSearch(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	req := &agentregistry.SearchMcpServersRequest{SearchString: args[0]}
	resp, err := svc.Projects.Locations.McpServers.Search(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("searching MCP servers: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagARFormat)
}

// --- operations impl ---

func (g *agentRegistryCmds) runAROpCancel(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
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

func (g *agentRegistryCmds) runAROpDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
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

func (g *agentRegistryCmds) runAROpDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "operations", args[0])
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagARFormat)
}

func (g *agentRegistryCmds) runAROpList(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	var all []*agentregistry.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if g.flagARFilter != "" {
			call = call.Filter(g.flagARFilter)
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
	if g.flagARFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagARFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func (g *agentRegistryCmds) runAROpWait(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
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
	return emitFormattedTo(g.inv.stdout, final, g.flagARFormat)
}

// --- services impl ---

func (g *agentRegistryCmds) runARSvcCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	body := &agentregistry.Service{}
	if err := loadYAMLOrJSONInto(g.flagARConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	id := g.flagARServiceID
	if id == "" {
		id = args[0]
	}
	call := svc.Projects.Locations.Services.Create(parent, body).ServiceId(id).Context(ctx)
	if g.flagARRequestID != "" {
		call = call.RequestId(g.flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}
	return g.arFinishOp(ctx, svc, op, "Create service", args[0])
}

func (g *agentRegistryCmds) runARSvcDelete(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "services", args[0])
	call := svc.Projects.Locations.Services.Delete(name).Context(ctx)
	if g.flagARRequestID != "" {
		call = call.RequestId(g.flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	return g.arFinishOp(ctx, svc, op, "Delete service", args[0])
}

func (g *agentRegistryCmds) runARSvcDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "services", args[0])
	got, err := svc.Projects.Locations.Services.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagARFormat)
}

func (g *agentRegistryCmds) runARSvcList(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	var all []*agentregistry.Service
	pageToken := ""
	for {
		call := svc.Projects.Locations.Services.List(parent).Context(ctx)
		if g.flagARFilter != "" {
			call = call.Filter(g.flagARFilter)
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
	if g.flagARFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagARFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, s := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(s.Name), s.DisplayName)
	}
	return nil
}

func (g *agentRegistryCmds) runARSvcUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.arResolveParent()
	if err != nil {
		return err
	}
	body := &agentregistry.Service{}
	if err := loadYAMLOrJSONInto(g.flagARConfigFile, body); err != nil {
		return err
	}
	mask := g.flagARUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AgentRegistry(ctx)
	if err != nil {
		return err
	}
	name := arChild(parent, "services", args[0])
	call := svc.Projects.Locations.Services.Patch(name, body).UpdateMask(mask).Context(ctx)
	if g.flagARRequestID != "" {
		call = call.RequestId(g.flagARRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service: %w", err)
	}
	return g.arFinishOp(ctx, svc, op, "Update service", args[0])
}

// agentRegistryCmds holds the commands and flag values of one invocation's "agentRegistry"
// command group, so invocations don't share state (#2070).
type agentRegistryCmds struct {
	inv *invocation

	agentRegistryCmd           *cobra.Command
	flagARLocation             string
	flagARFormat               string
	flagARFilter               string
	flagARConfigFile           string
	flagARUpdateMask           string
	flagARServiceID            string
	flagARRequestID            string
	flagARAsync                bool
	agentRegistryAgentsCmd     *cobra.Command
	arAgentDescribeCmd         *cobra.Command
	arAgentListCmd             *cobra.Command
	arAgentSearchCmd           *cobra.Command
	agentRegistryBindingsCmd   *cobra.Command
	arBindingCreateCmd         *cobra.Command
	arBindingDeleteCmd         *cobra.Command
	arBindingDescribeCmd       *cobra.Command
	arBindingFetchAvailableCmd *cobra.Command
	arBindingListCmd           *cobra.Command
	arBindingUpdateCmd         *cobra.Command
	agentRegistryEndpointsCmd  *cobra.Command
	arEndpointDescribeCmd      *cobra.Command
	arEndpointListCmd          *cobra.Command
	agentRegistryMcpServersCmd *cobra.Command
	arMcpDescribeCmd           *cobra.Command
	arMcpListCmd               *cobra.Command
	arMcpSearchCmd             *cobra.Command
	agentRegistryOperationsCmd *cobra.Command
	arOpCancelCmd              *cobra.Command
	arOpDeleteCmd              *cobra.Command
	arOpDescribeCmd            *cobra.Command
	arOpListCmd                *cobra.Command
	arOpWaitCmd                *cobra.Command
	agentRegistryServicesCmd   *cobra.Command
	arSvcCreateCmd             *cobra.Command
	arSvcDeleteCmd             *cobra.Command
	arSvcDescribeCmd           *cobra.Command
	arSvcListCmd               *cobra.Command
	arSvcUpdateCmd             *cobra.Command
}

// newAgentRegistryCmd returns the "agentRegistry" command group for inv.
func newAgentRegistryCmd(inv *invocation) *cobra.Command {
	return newAgentRegistryCmds(inv).agentRegistryCmd
}

// newAgentRegistryCmds builds the "agentRegistry" command group for inv and returns its state.
func newAgentRegistryCmds(inv *invocation) *agentRegistryCmds {
	g := &agentRegistryCmds{inv: inv}
	g.agentRegistryCmd = &cobra.Command{Use: "agent-registry", Short: "Manage Agent Registry"}
	g.agentRegistryAgentsCmd = &cobra.Command{Use: "agents", Short: "Manage Agent Registry agents"}
	g.arAgentDescribeCmd = &cobra.Command{
		Use: "describe AGENT", Short: "Describe an agent",
		Args: cobra.ExactArgs(1), RunE: g.runARAgentDescribe,
	}
	g.arAgentListCmd = &cobra.Command{
		Use: "list", Short: "List agents in a location",
		Args: cobra.NoArgs, RunE: g.runARAgentList,
	}
	g.arAgentSearchCmd = &cobra.Command{
		Use: "search QUERY", Short: "Search agents",
		Args: cobra.ExactArgs(1), RunE: g.runARAgentSearch,
	}
	g.agentRegistryBindingsCmd = &cobra.Command{Use: "bindings", Short: "Manage Agent Registry bindings"}
	g.arBindingCreateCmd = &cobra.Command{
		Use: "create BINDING", Short: "Create a binding from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runARBindingCreate,
	}
	g.arBindingDeleteCmd = &cobra.Command{
		Use: "delete BINDING", Short: "Delete a binding",
		Args: cobra.ExactArgs(1), RunE: g.runARBindingDelete,
	}
	g.arBindingDescribeCmd = &cobra.Command{
		Use: "describe BINDING", Short: "Describe a binding",
		Args: cobra.ExactArgs(1), RunE: g.runARBindingDescribe,
	}
	g.arBindingFetchAvailableCmd = &cobra.Command{
		Use: "fetch-available", Short: "Fetch available bindings in a location",
		Args: cobra.NoArgs, RunE: g.runARBindingFetchAvailable,
	}
	g.arBindingListCmd = &cobra.Command{
		Use: "list", Short: "List bindings in a location",
		Args: cobra.NoArgs, RunE: g.runARBindingList,
	}
	g.arBindingUpdateCmd = &cobra.Command{
		Use: "update BINDING", Short: "Update a binding from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runARBindingUpdate,
	}
	g.agentRegistryEndpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Agent Registry endpoints"}
	g.arEndpointDescribeCmd = &cobra.Command{
		Use: "describe ENDPOINT", Short: "Describe an endpoint",
		Args: cobra.ExactArgs(1), RunE: g.runAREndpointDescribe,
	}
	g.arEndpointListCmd = &cobra.Command{
		Use: "list", Short: "List endpoints in a location",
		Args: cobra.NoArgs, RunE: g.runAREndpointList,
	}
	g.agentRegistryMcpServersCmd = &cobra.Command{Use: "mcp-servers", Short: "Manage Agent Registry MCP servers"}
	g.arMcpDescribeCmd = &cobra.Command{
		Use: "describe SERVER", Short: "Describe an MCP server",
		Args: cobra.ExactArgs(1), RunE: g.runARMcpDescribe,
	}
	g.arMcpListCmd = &cobra.Command{
		Use: "list", Short: "List MCP servers in a location",
		Args: cobra.NoArgs, RunE: g.runARMcpList,
	}
	g.arMcpSearchCmd = &cobra.Command{
		Use: "search QUERY", Short: "Search MCP servers",
		Args: cobra.ExactArgs(1), RunE: g.runARMcpSearch,
	}
	g.agentRegistryOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Agent Registry operations"}
	g.arOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: g.runAROpCancel,
	}
	g.arOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete an operation",
		Args: cobra.ExactArgs(1), RunE: g.runAROpDelete,
	}
	g.arOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: g.runAROpDescribe,
	}
	g.arOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: g.runAROpList,
	}
	g.arOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for an operation to complete",
		Args: cobra.ExactArgs(1), RunE: g.runAROpWait,
	}
	g.agentRegistryServicesCmd = &cobra.Command{Use: "services", Short: "Manage Agent Registry services"}
	g.arSvcCreateCmd = &cobra.Command{
		Use: "create SERVICE", Short: "Create a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runARSvcCreate,
	}
	g.arSvcDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a service",
		Args: cobra.ExactArgs(1), RunE: g.runARSvcDelete,
	}
	g.arSvcDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a service",
		Args: cobra.ExactArgs(1), RunE: g.runARSvcDescribe,
	}
	g.arSvcListCmd = &cobra.Command{
		Use: "list", Short: "List services in a location",
		Args: cobra.NoArgs, RunE: g.runARSvcList,
	}
	g.arSvcUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a service from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runARSvcUpdate,
	}
	g.initAgentRegistry()
	return g
}
