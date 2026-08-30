package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub locations configure-and-deploy-server (#1754) ---
//
// Ports the GA promotion shipped in gcloud-python 580.0.0. Only MCP (Model
// Context Protocol) servers on Apigee X are currently supported by the
// backend; this command exposes the corresponding v1
// `apihub.projects.locations.servers.configureAndDeployServer` RPC.

var apihubLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage API Hub location-level resources",
}

var apihubLocationsConfigureAndDeployCmd = &cobra.Command{
	Use:   "configure-and-deploy-server LOCATION",
	Short: "Configure and deploy an API Hub MCP server (currently on Apigee X)",
	Args:  cobra.ExactArgs(1),
	RunE:  runApihubLocationsConfigureAndDeploy,
}

var (
	flagApihubCDFormat            string
	flagApihubCDMcpTools          []string
	flagApihubCDMcpToolsFromFile  string
	flagApihubCDApigeeXEnv        string
	flagApihubCDApigeeXProxy      string
	flagApihubCDApigeeXTargetProj string
	flagApihubCDApigeeXDisplay    string
	flagApihubCDApigeeXDescr      string
)

func init() {
	apihubLocationsConfigureAndDeployCmd.Flags().StringVar(&flagApihubCDFormat, "format", "", "Output format")
	apihubLocationsConfigureAndDeployCmd.Flags().StringArrayVar(&flagApihubCDMcpTools, "mcp-tools", nil,
		"MCP tool exposed on the server; repeatable. Each value is a comma-separated dict with keys "+
			"tool-id=..,description=..,operation=<full API Hub operation resource name>. "+
			"For tools that reference an operation by spec+path+method, use --mcp-tools-from-file")
	apihubLocationsConfigureAndDeployCmd.Flags().StringVar(&flagApihubCDMcpToolsFromFile, "mcp-tools-from-file", "",
		"Path to a YAML/JSON file containing a list of MCP tools "+
			"(supports the http_operation oneof arm as {spec,path,method})")
	apihubLocationsConfigureAndDeployCmd.Flags().StringVar(&flagApihubCDApigeeXEnv, "apigee-x-environment", "",
		"Apigee X environment to deploy into (required for Apigee X target)")
	apihubLocationsConfigureAndDeployCmd.Flags().StringVar(&flagApihubCDApigeeXProxy, "apigee-x-proxy", "",
		"Apigee X proxy resource name (required for Apigee X target)")
	apihubLocationsConfigureAndDeployCmd.Flags().StringVar(&flagApihubCDApigeeXTargetProj, "apigee-x-target-project", "",
		"Runtime project hosting the Apigee X organization (required for Apigee X target)")
	apihubLocationsConfigureAndDeployCmd.Flags().StringVar(&flagApihubCDApigeeXDisplay, "apigee-x-proxy-display-name", "",
		"Display name for the deployed Apigee X proxy revision")
	apihubLocationsConfigureAndDeployCmd.Flags().StringVar(&flagApihubCDApigeeXDescr, "apigee-x-proxy-description", "",
		"Description for the deployed Apigee X proxy revision")

	apihubLocationsCmd.AddCommand(apihubLocationsConfigureAndDeployCmd)
	apihubCmd.AddCommand(apihubLocationsCmd)
}

// parseMcpToolFlag parses a single --mcp-tools value into an
// McpToolConfig. Supports the gcloud-python alternate-delimiter syntax
// (`^|^k=v|k=v|...`) documented under `gcloud topic escaping`.
func parseMcpToolFlag(spec string) (*apihub.GoogleCloudApihubV1McpToolConfig, error) {
	delim := ","
	if strings.HasPrefix(spec, "^") {
		end := strings.Index(spec[1:], "^")
		if end <= 0 {
			return nil, fmt.Errorf("malformed alternate-delimiter syntax %q", spec)
		}
		delim = spec[1 : 1+end]
		spec = spec[2+end:]
	}
	tool := &apihub.GoogleCloudApihubV1McpToolConfig{Operation: &apihub.GoogleCloudApihubV1OperationConfig{}}
	for _, part := range strings.Split(spec, delim) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		eq := strings.Index(part, "=")
		if eq < 0 {
			return nil, fmt.Errorf("expected key=value in %q", part)
		}
		key, val := part[:eq], part[eq+1:]
		switch key {
		case "tool-id":
			tool.ToolId = val
		case "description":
			tool.Description = val
		case "operation":
			tool.Operation.Operation = val
		default:
			return nil, fmt.Errorf("unrecognised --mcp-tools key %q "+
				"(supported: tool-id, description, operation)", key)
		}
	}
	if tool.ToolId == "" || tool.Description == "" || tool.Operation.Operation == "" {
		return nil, fmt.Errorf("--mcp-tools requires tool-id, description, and operation " +
			"(use --mcp-tools-from-file for http_operation)")
	}
	return tool, nil
}

// buildApihubMcpTools returns the McpToolConfig list from --mcp-tools /
// --mcp-tools-from-file, enforcing mutual exclusion.
func buildApihubMcpTools() ([]*apihub.GoogleCloudApihubV1McpToolConfig, error) {
	if len(flagApihubCDMcpTools) > 0 && flagApihubCDMcpToolsFromFile != "" {
		return nil, fmt.Errorf("--mcp-tools and --mcp-tools-from-file are mutually exclusive")
	}
	if len(flagApihubCDMcpTools) == 0 && flagApihubCDMcpToolsFromFile == "" {
		return nil, fmt.Errorf("one of --mcp-tools or --mcp-tools-from-file is required")
	}
	if flagApihubCDMcpToolsFromFile != "" {
		var tools []*apihub.GoogleCloudApihubV1McpToolConfig
		if err := loadYAMLOrJSONInto(flagApihubCDMcpToolsFromFile, &tools); err != nil {
			return nil, err
		}
		return tools, nil
	}
	out := make([]*apihub.GoogleCloudApihubV1McpToolConfig, 0, len(flagApihubCDMcpTools))
	for _, s := range flagApihubCDMcpTools {
		t, err := parseMcpToolFlag(s)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func runApihubLocationsConfigureAndDeploy(cmd *cobra.Command, args []string) error {
	loc := args[0]
	parent := loc
	if !strings.HasPrefix(loc, "projects/") {
		var err error
		parent, err = apihubLocationParent(loc)
		if err != nil {
			return err
		}
	}

	if flagApihubCDApigeeXEnv == "" || flagApihubCDApigeeXProxy == "" || flagApihubCDApigeeXTargetProj == "" {
		return fmt.Errorf("--apigee-x-environment, --apigee-x-proxy, and --apigee-x-target-project are required " +
			"(Apigee X is currently the only supported MCP target)")
	}

	tools, err := buildApihubMcpTools()
	if err != nil {
		return err
	}

	target := &apihub.GoogleCloudApihubV1ApigeeXTargetDetails{
		Environment:   flagApihubCDApigeeXEnv,
		Proxy:         flagApihubCDApigeeXProxy,
		TargetProject: flagApihubCDApigeeXTargetProj,
	}
	if flagApihubCDApigeeXDisplay != "" || flagApihubCDApigeeXDescr != "" {
		target.Metadata = &apihub.GoogleCloudApihubV1MetaData{
			DisplayName: flagApihubCDApigeeXDisplay,
			Description: flagApihubCDApigeeXDescr,
		}
	}

	req := &apihub.GoogleCloudApihubV1ConfigureAndDeployServerRequest{
		McpServerConfig: &apihub.GoogleCloudApihubV1McpServerConfig{
			Tools:                tools,
			ApigeeXTargetDetails: target,
		},
	}

	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Servers.ConfigureAndDeployServer(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("configure-and-deploy MCP server: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Configure-and-deploy request issued (operation: %s).\n", op.Name)
	return emitFormatted(op, flagApihubCDFormat)
}
