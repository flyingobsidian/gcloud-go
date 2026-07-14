package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigateway "google.golang.org/api/apigateway/v1"
)

// --- gcloud api-gateway (#296) ---

var apiGatewayCmd = &cobra.Command{Use: "api-gateway", Short: "Manage API Gateway"}

func agLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func agChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func agWaitOp(ctx context.Context, svc *apigateway.Service, op *apigateway.ApigatewayOperation) (*apigateway.ApigatewayOperation, error) {
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

func agFinishOp(ctx context.Context, svc *apigateway.Service, op *apigateway.ApigatewayOperation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := agWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

var (
	flagAGLocation   string
	flagAGConfigFile string
	flagAGUpdateMask string
	flagAGFormat     string
	flagAGAsync      bool
	flagAGAPI        string
)

// --- apis ---

var agApisCmd = &cobra.Command{Use: "apis", Short: "Manage API Gateway APIs"}

var (
	agAPICreateCmd = &cobra.Command{
		Use: "create API", Short: "Create an API from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAGAPICreate,
	}
	agAPIDeleteCmd = &cobra.Command{
		Use: "delete API", Short: "Delete an API",
		Args: cobra.ExactArgs(1), RunE: runAGAPIDelete,
	}
	agAPIDescribeCmd = &cobra.Command{
		Use: "describe API", Short: "Describe an API",
		Args: cobra.ExactArgs(1), RunE: runAGAPIDescribe,
	}
	agAPIListCmd = &cobra.Command{
		Use: "list", Short: "List APIs",
		Args: cobra.NoArgs, RunE: runAGAPIList,
	}
	agAPIUpdateCmd = &cobra.Command{
		Use: "update API", Short: "Update an API from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAGAPIUpdate,
	}
)

// --- api-configs ---

var agAPIConfigsCmd = &cobra.Command{Use: "api-configs", Short: "Manage API Gateway API configs"}

var (
	agACCreateCmd = &cobra.Command{
		Use: "create CONFIG", Short: "Create an API config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAGACCreate,
	}
	agACDeleteCmd = &cobra.Command{
		Use: "delete CONFIG", Short: "Delete an API config",
		Args: cobra.ExactArgs(1), RunE: runAGACDelete,
	}
	agACDescribeCmd = &cobra.Command{
		Use: "describe CONFIG", Short: "Describe an API config",
		Args: cobra.ExactArgs(1), RunE: runAGACDescribe,
	}
	agACListCmd = &cobra.Command{
		Use: "list", Short: "List API configs for an API",
		Args: cobra.NoArgs, RunE: runAGACList,
	}
	agACUpdateCmd = &cobra.Command{
		Use: "update CONFIG", Short: "Update an API config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAGACUpdate,
	}
)

// --- gateways ---

var agGatewaysCmd = &cobra.Command{Use: "gateways", Short: "Manage API Gateway gateways"}

var (
	agGWCreateCmd = &cobra.Command{
		Use: "create GATEWAY", Short: "Create a gateway from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAGGWCreate,
	}
	agGWDeleteCmd = &cobra.Command{
		Use: "delete GATEWAY", Short: "Delete a gateway",
		Args: cobra.ExactArgs(1), RunE: runAGGWDelete,
	}
	agGWDescribeCmd = &cobra.Command{
		Use: "describe GATEWAY", Short: "Describe a gateway",
		Args: cobra.ExactArgs(1), RunE: runAGGWDescribe,
	}
	agGWListCmd = &cobra.Command{
		Use: "list", Short: "List gateways",
		Args: cobra.NoArgs, RunE: runAGGWList,
	}
	agGWUpdateCmd = &cobra.Command{
		Use: "update GATEWAY", Short: "Update a gateway from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runAGGWUpdate,
	}
)

// --- operations ---

var agOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage API Gateway operations"}

var (
	agOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: runAGOpCancel,
	}
	agOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runAGOpDescribe,
	}
	agOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: runAGOpList,
	}
	agOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for an operation to complete",
		Args: cobra.ExactArgs(1), RunE: runAGOpWait,
	}
)

func init() {
	apiCmds := []*cobra.Command{agAPICreateCmd, agAPIDeleteCmd, agAPIDescribeCmd, agAPIListCmd, agAPIUpdateCmd}
	for _, c := range apiCmds {
		c.Flags().StringVar(&flagAGLocation, "location", "global", "Location containing the API")
	}
	for _, c := range []*cobra.Command{agAPICreateCmd, agAPIUpdateCmd} {
		c.Flags().StringVar(&flagAGConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Api message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	agAPIUpdateCmd.Flags().StringVar(&flagAGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{agAPICreateCmd, agAPIDeleteCmd, agAPIUpdateCmd} {
		c.Flags().BoolVar(&flagAGAsync, "async", false, "Return the long-running operation without waiting")
	}
	agAPIDescribeCmd.Flags().StringVar(&flagAGFormat, "format", "", "Output format")
	agAPIListCmd.Flags().StringVar(&flagAGFormat, "format", "", "Output format")
	agApisCmd.AddCommand(apiCmds...)

	acCmds := []*cobra.Command{agACCreateCmd, agACDeleteCmd, agACDescribeCmd, agACListCmd, agACUpdateCmd}
	for _, c := range acCmds {
		c.Flags().StringVar(&flagAGLocation, "location", "global", "Location containing the API")
		c.Flags().StringVar(&flagAGAPI, "api", "", "API containing the config (required)")
		_ = c.MarkFlagRequired("api")
	}
	for _, c := range []*cobra.Command{agACCreateCmd, agACUpdateCmd} {
		c.Flags().StringVar(&flagAGConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the ApiConfig message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	agACUpdateCmd.Flags().StringVar(&flagAGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{agACCreateCmd, agACDeleteCmd, agACUpdateCmd} {
		c.Flags().BoolVar(&flagAGAsync, "async", false, "Return the long-running operation without waiting")
	}
	agACDescribeCmd.Flags().StringVar(&flagAGFormat, "format", "", "Output format")
	agACListCmd.Flags().StringVar(&flagAGFormat, "format", "", "Output format")
	agAPIConfigsCmd.AddCommand(acCmds...)

	gwCmds := []*cobra.Command{agGWCreateCmd, agGWDeleteCmd, agGWDescribeCmd, agGWListCmd, agGWUpdateCmd}
	for _, c := range gwCmds {
		c.Flags().StringVar(&flagAGLocation, "location", "", "Location containing the gateway (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{agGWCreateCmd, agGWUpdateCmd} {
		c.Flags().StringVar(&flagAGConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Gateway message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	agGWUpdateCmd.Flags().StringVar(&flagAGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{agGWCreateCmd, agGWDeleteCmd, agGWUpdateCmd} {
		c.Flags().BoolVar(&flagAGAsync, "async", false, "Return the long-running operation without waiting")
	}
	agGWDescribeCmd.Flags().StringVar(&flagAGFormat, "format", "", "Output format")
	agGWListCmd.Flags().StringVar(&flagAGFormat, "format", "", "Output format")
	agGatewaysCmd.AddCommand(gwCmds...)

	opCmds := []*cobra.Command{agOpCancelCmd, agOpDescribeCmd, agOpListCmd, agOpWaitCmd}
	for _, c := range opCmds {
		c.Flags().StringVar(&flagAGLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	agOpDescribeCmd.Flags().StringVar(&flagAGFormat, "format", "", "Output format")
	agOpListCmd.Flags().StringVar(&flagAGFormat, "format", "", "Output format")
	agOperationsCmd.AddCommand(opCmds...)

	apiGatewayCmd.AddCommand(agApisCmd, agAPIConfigsCmd, agGatewaysCmd, agOperationsCmd)
	rootCmd.AddCommand(apiGatewayCmd)
}

func agAPIName(id, project, location string) string {
	return agChild("apis", id, agLocationParent(project, location))
}

func agACName(id, project, location, api string) string {
	return agChild("configs", id, fmt.Sprintf("%s/apis/%s", agLocationParent(project, location), api))
}

func agGWName(id, project, location string) string {
	return agChild("gateways", id, agLocationParent(project, location))
}

func agOpName(id, project, location string) string {
	return agChild("operations", id, agLocationParent(project, location))
}

func runAGAPICreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	api := &apigateway.ApigatewayApi{}
	if err := loadYAMLOrJSONInto(flagAGConfigFile, api); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Create(agLocationParent(project, flagAGLocation), api).
		ApiId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating API: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Create API", args[0], flagAGAsync)
}

func runAGAPIDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Delete(agAPIName(args[0], project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting API: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Delete API", args[0], flagAGAsync)
}

func runAGAPIDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Apis.Get(agAPIName(args[0], project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing API: %w", err)
	}
	return emitFormatted(got, flagAGFormat)
}

func runAGAPIList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Apis.List(agLocationParent(project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing APIs: %w", err)
	}
	if flagAGFormat != "" {
		return emitFormatted(resp.Apis, flagAGFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, a := range resp.Apis {
		fmt.Printf("%-40s %s\n", path.Base(a.Name), a.State)
	}
	return nil
}

func runAGAPIUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	api := &apigateway.ApigatewayApi{}
	if err := loadYAMLOrJSONInto(flagAGConfigFile, api); err != nil {
		return err
	}
	mask := flagAGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(api))
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Patch(agAPIName(args[0], project, flagAGLocation), api).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating API: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Update API", args[0], flagAGAsync)
}

func runAGACCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ac := &apigateway.ApigatewayApiConfig{}
	if err := loadYAMLOrJSONInto(flagAGConfigFile, ac); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("%s/apis/%s", agLocationParent(project, flagAGLocation), flagAGAPI)
	op, err := svc.Projects.Locations.Apis.Configs.Create(parent, ac).
		ApiConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating API config: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Create API config", args[0], flagAGAsync)
}

func runAGACDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Configs.Delete(agACName(args[0], project, flagAGLocation, flagAGAPI)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting API config: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Delete API config", args[0], flagAGAsync)
}

func runAGACDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Apis.Configs.Get(agACName(args[0], project, flagAGLocation, flagAGAPI)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing API config: %w", err)
	}
	return emitFormatted(got, flagAGFormat)
}

func runAGACList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("%s/apis/%s", agLocationParent(project, flagAGLocation), flagAGAPI)
	resp, err := svc.Projects.Locations.Apis.Configs.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing API configs: %w", err)
	}
	if flagAGFormat != "" {
		return emitFormatted(resp.ApiConfigs, flagAGFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, c := range resp.ApiConfigs {
		fmt.Printf("%-40s %s\n", path.Base(c.Name), c.State)
	}
	return nil
}

func runAGACUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ac := &apigateway.ApigatewayApiConfig{}
	if err := loadYAMLOrJSONInto(flagAGConfigFile, ac); err != nil {
		return err
	}
	mask := flagAGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ac))
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Configs.Patch(agACName(args[0], project, flagAGLocation, flagAGAPI), ac).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating API config: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Update API config", args[0], flagAGAsync)
}

func runAGGWCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	gw := &apigateway.ApigatewayGateway{}
	if err := loadYAMLOrJSONInto(flagAGConfigFile, gw); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Gateways.Create(agLocationParent(project, flagAGLocation), gw).
		GatewayId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating gateway: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Create gateway", args[0], flagAGAsync)
}

func runAGGWDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Gateways.Delete(agGWName(args[0], project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting gateway: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Delete gateway", args[0], flagAGAsync)
}

func runAGGWDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Gateways.Get(agGWName(args[0], project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing gateway: %w", err)
	}
	return emitFormatted(got, flagAGFormat)
}

func runAGGWList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Gateways.List(agLocationParent(project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing gateways: %w", err)
	}
	if flagAGFormat != "" {
		return emitFormatted(resp.Gateways, flagAGFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, gw := range resp.Gateways {
		fmt.Printf("%-40s %s\n", path.Base(gw.Name), gw.State)
	}
	return nil
}

func runAGGWUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	gw := &apigateway.ApigatewayGateway{}
	if err := loadYAMLOrJSONInto(flagAGConfigFile, gw); err != nil {
		return err
	}
	mask := flagAGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(gw))
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Gateways.Patch(agGWName(args[0], project, flagAGLocation), gw).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating gateway: %w", err)
	}
	return agFinishOp(ctx, svc, op, "Update gateway", args[0], flagAGAsync)
}

func runAGOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(agOpName(args[0], project, flagAGLocation), &apigateway.ApigatewayCancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runAGOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(agOpName(args[0], project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagAGFormat)
}

func runAGOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(agLocationParent(project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagAGFormat != "" {
		return emitFormatted(resp.Operations, flagAGFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func runAGOpWait(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.APIGatewayService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(agOpName(args[0], project, flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching operation: %w", err)
	}
	final, err := agWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagAGFormat)
}
