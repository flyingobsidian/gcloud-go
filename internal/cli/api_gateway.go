package cli

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	apigateway "google.golang.org/api/apigateway/v1"
)

// --- gcloud api-gateway (#296) ---

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

func (s *apiGatewayCmds) agFinishOp(ctx context.Context, svc *apigateway.Service, op *apigateway.ApigatewayOperation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := agWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// --- apis ---

// --- api-configs ---

// --- gateways ---

// --- operations ---

func (s *apiGatewayCmds) initApiGateway() {
	apiCmds := []*cobra.Command{s.agAPICreateCmd, s.agAPIDeleteCmd, s.agAPIDescribeCmd, s.agAPIListCmd, s.agAPIUpdateCmd}
	for _, c := range apiCmds {
		c.Flags().StringVar(&s.flagAGLocation, "location", "global", "Location containing the API")
	}
	for _, c := range []*cobra.Command{s.agAPICreateCmd, s.agAPIUpdateCmd} {
		c.Flags().StringVar(&s.flagAGConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Api message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.agAPIUpdateCmd.Flags().StringVar(&s.flagAGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.agAPICreateCmd, s.agAPIDeleteCmd, s.agAPIUpdateCmd} {
		c.Flags().BoolVar(&s.flagAGAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.agAPIDescribeCmd.Flags().StringVar(&s.flagAGFormat, "format", "", "Output format")
	s.agAPIListCmd.Flags().StringVar(&s.flagAGFormat, "format", "", "Output format")
	s.agApisCmd.AddCommand(apiCmds...)

	acCmds := []*cobra.Command{s.agACCreateCmd, s.agACDeleteCmd, s.agACDescribeCmd, s.agACListCmd, s.agACUpdateCmd}
	for _, c := range acCmds {
		c.Flags().StringVar(&s.flagAGLocation, "location", "global", "Location containing the API")
		c.Flags().StringVar(&s.flagAGAPI, "api", "", "API containing the config (required)")
		_ = c.MarkFlagRequired("api")
	}
	for _, c := range []*cobra.Command{s.agACCreateCmd, s.agACUpdateCmd} {
		c.Flags().StringVar(&s.flagAGConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the ApiConfig message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.agACUpdateCmd.Flags().StringVar(&s.flagAGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.agACCreateCmd, s.agACDeleteCmd, s.agACUpdateCmd} {
		c.Flags().BoolVar(&s.flagAGAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.agACDescribeCmd.Flags().StringVar(&s.flagAGFormat, "format", "", "Output format")
	s.agACListCmd.Flags().StringVar(&s.flagAGFormat, "format", "", "Output format")
	s.agAPIConfigsCmd.AddCommand(acCmds...)

	gwCmds := []*cobra.Command{s.agGWCreateCmd, s.agGWDeleteCmd, s.agGWDescribeCmd, s.agGWListCmd, s.agGWUpdateCmd}
	for _, c := range gwCmds {
		c.Flags().StringVar(&s.flagAGLocation, "location", "", "Location containing the gateway (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{s.agGWCreateCmd, s.agGWUpdateCmd} {
		c.Flags().StringVar(&s.flagAGConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Gateway message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.agGWCreateCmd.Flags().BoolVar(&s.flagAGGWEnableStreaming, "enable-streaming", false,
		"Creates the gateway with streaming enabled, which supports Server-Sent "+
			"Events (SSE), HTTP/2 streaming, HTTP chunked transfer, WebSockets, and "+
			"gRPC bidirectional streaming. This can only be set when the gateway is "+
			"created and cannot be changed afterwards. When omitted, API Gateway "+
			"selects the streaming mode.")
	s.agGWUpdateCmd.Flags().StringVar(&s.flagAGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.agGWCreateCmd, s.agGWDeleteCmd, s.agGWUpdateCmd} {
		c.Flags().BoolVar(&s.flagAGAsync, "async", false, "Return the long-running operation without waiting")
	}
	s.agGWDescribeCmd.Flags().StringVar(&s.flagAGFormat, "format", "", "Output format")
	s.agGWListCmd.Flags().StringVar(&s.flagAGFormat, "format", "", "Output format")
	s.agGatewaysCmd.AddCommand(gwCmds...)

	opCmds := []*cobra.Command{s.agOpCancelCmd, s.agOpDescribeCmd, s.agOpListCmd, s.agOpWaitCmd}
	for _, c := range opCmds {
		c.Flags().StringVar(&s.flagAGLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	s.agOpDescribeCmd.Flags().StringVar(&s.flagAGFormat, "format", "", "Output format")
	s.agOpListCmd.Flags().StringVar(&s.flagAGFormat, "format", "", "Output format")
	s.agOperationsCmd.AddCommand(opCmds...)

	s.apiGatewayCmd.AddCommand(s.agApisCmd, s.agAPIConfigsCmd, s.agGatewaysCmd, s.agOperationsCmd)

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

func (s *apiGatewayCmds) runAGAPICreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	api := &apigateway.ApigatewayApi{}
	if err := loadYAMLOrJSONInto(s.flagAGConfigFile, api); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Create(agLocationParent(project, s.flagAGLocation), api).
		ApiId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating API: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Create API", args[0], s.flagAGAsync)
}

func (s *apiGatewayCmds) runAGAPIDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Delete(agAPIName(args[0], project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting API: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Delete API", args[0], s.flagAGAsync)
}

func (s *apiGatewayCmds) runAGAPIDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Apis.Get(agAPIName(args[0], project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing API: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAGFormat)
}

func (s *apiGatewayCmds) runAGAPIList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Apis.List(agLocationParent(project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing APIs: %w", err)
	}
	if s.flagAGFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Apis, s.flagAGFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Apis), "%-40s %s\n", "NAME", "STATE")
	for _, a := range resp.Apis {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(a.Name), a.State)
	}
	return nil
}

func (s *apiGatewayCmds) runAGAPIUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	api := &apigateway.ApigatewayApi{}
	if err := loadYAMLOrJSONInto(s.flagAGConfigFile, api); err != nil {
		return err
	}
	mask := s.flagAGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(api))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Patch(agAPIName(args[0], project, s.flagAGLocation), api).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating API: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Update API", args[0], s.flagAGAsync)
}

func (s *apiGatewayCmds) runAGACCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ac := &apigateway.ApigatewayApiConfig{}
	if err := loadYAMLOrJSONInto(s.flagAGConfigFile, ac); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("%s/apis/%s", agLocationParent(project, s.flagAGLocation), s.flagAGAPI)
	op, err := svc.Projects.Locations.Apis.Configs.Create(parent, ac).
		ApiConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating API config: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Create API config", args[0], s.flagAGAsync)
}

func (s *apiGatewayCmds) runAGACDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Configs.Delete(agACName(args[0], project, s.flagAGLocation, s.flagAGAPI)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting API config: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Delete API config", args[0], s.flagAGAsync)
}

func (s *apiGatewayCmds) runAGACDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Apis.Configs.Get(agACName(args[0], project, s.flagAGLocation, s.flagAGAPI)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing API config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAGFormat)
}

func (s *apiGatewayCmds) runAGACList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("%s/apis/%s", agLocationParent(project, s.flagAGLocation), s.flagAGAPI)
	resp, err := svc.Projects.Locations.Apis.Configs.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing API configs: %w", err)
	}
	if s.flagAGFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.ApiConfigs, s.flagAGFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.ApiConfigs), "%-40s %s\n", "NAME", "STATE")
	for _, c := range resp.ApiConfigs {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(c.Name), c.State)
	}
	return nil
}

func (s *apiGatewayCmds) runAGACUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ac := &apigateway.ApigatewayApiConfig{}
	if err := loadYAMLOrJSONInto(s.flagAGConfigFile, ac); err != nil {
		return err
	}
	mask := s.flagAGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(ac))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Apis.Configs.Patch(agACName(args[0], project, s.flagAGLocation, s.flagAGAPI), ac).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating API config: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Update API config", args[0], s.flagAGAsync)
}

// applyAGGWEnableStreaming opts the gateway body into streaming when
// --enable-streaming is set; otherwise the --config-file value is kept.
func applyAGGWEnableStreaming(gw *apigateway.ApigatewayGateway, enable bool) {
	if enable {
		gw.StreamingMode = "STREAMING_MODE_ENABLED"
	}
}

func (s *apiGatewayCmds) runAGGWCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	gw := &apigateway.ApigatewayGateway{}
	if err := loadYAMLOrJSONInto(s.flagAGConfigFile, gw); err != nil {
		return err
	}
	applyAGGWEnableStreaming(gw, s.flagAGGWEnableStreaming)
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Gateways.Create(agLocationParent(project, s.flagAGLocation), gw).
		GatewayId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating gateway: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Create gateway", args[0], s.flagAGAsync)
}

func (s *apiGatewayCmds) runAGGWDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Gateways.Delete(agGWName(args[0], project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting gateway: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Delete gateway", args[0], s.flagAGAsync)
}

func (s *apiGatewayCmds) runAGGWDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Gateways.Get(agGWName(args[0], project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing gateway: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAGFormat)
}

func (s *apiGatewayCmds) runAGGWList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Gateways.List(agLocationParent(project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing gateways: %w", err)
	}
	if s.flagAGFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Gateways, s.flagAGFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Gateways), "%-40s %s\n", "NAME", "STATE")
	for _, gw := range resp.Gateways {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(gw.Name), gw.State)
	}
	return nil
}

func (s *apiGatewayCmds) runAGGWUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	gw := &apigateway.ApigatewayGateway{}
	if err := loadYAMLOrJSONInto(s.flagAGConfigFile, gw); err != nil {
		return err
	}
	mask := s.flagAGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(gw))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Gateways.Patch(agGWName(args[0], project, s.flagAGLocation), gw).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating gateway: %w", err)
	}
	return s.agFinishOp(ctx, svc, op, "Update gateway", args[0], s.flagAGAsync)
}

func (s *apiGatewayCmds) runAGOpCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(agOpName(args[0], project, s.flagAGLocation), &apigateway.ApigatewayCancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *apiGatewayCmds) runAGOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(agOpName(args[0], project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagAGFormat)
}

func (s *apiGatewayCmds) runAGOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(agLocationParent(project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagAGFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagAGFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func (s *apiGatewayCmds) runAGOpWait(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().APIGateway(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(agOpName(args[0], project, s.flagAGLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching operation: %w", err)
	}
	final, err := agWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagAGFormat)
}

// apiGatewayCmds holds the commands and flag values of one invocation's "apiGateway"
// command group, so invocations don't share state (#2070).
type apiGatewayCmds struct {
	inv *invocation

	apiGatewayCmd    *cobra.Command
	flagAGLocation   string
	flagAGConfigFile string
	flagAGUpdateMask string
	flagAGFormat     string
	flagAGAsync      bool
	flagAGAPI        string
	// flagAGGWEnableStreaming sets streamingMode=STREAMING_MODE_ENABLED on
	// `gateways create` (gcloud-python 585.0.0). The mode is immutable, so
	// the flag only opts in; when omitted API Gateway selects the mode.
	flagAGGWEnableStreaming bool
	agApisCmd               *cobra.Command
	agAPICreateCmd          *cobra.Command
	agAPIDeleteCmd          *cobra.Command
	agAPIDescribeCmd        *cobra.Command
	agAPIListCmd            *cobra.Command
	agAPIUpdateCmd          *cobra.Command
	agAPIConfigsCmd         *cobra.Command
	agACCreateCmd           *cobra.Command
	agACDeleteCmd           *cobra.Command
	agACDescribeCmd         *cobra.Command
	agACListCmd             *cobra.Command
	agACUpdateCmd           *cobra.Command
	agGatewaysCmd           *cobra.Command
	agGWCreateCmd           *cobra.Command
	agGWDeleteCmd           *cobra.Command
	agGWDescribeCmd         *cobra.Command
	agGWListCmd             *cobra.Command
	agGWUpdateCmd           *cobra.Command
	agOperationsCmd         *cobra.Command
	agOpCancelCmd           *cobra.Command
	agOpDescribeCmd         *cobra.Command
	agOpListCmd             *cobra.Command
	agOpWaitCmd             *cobra.Command
}

// newApiGatewayCmd returns the "apiGateway" command group for inv.
func newApiGatewayCmd(inv *invocation) *cobra.Command {
	return newApiGatewayCmds(inv).apiGatewayCmd
}

// newApiGatewayCmds builds the "apiGateway" command group for inv and returns its state.
func newApiGatewayCmds(inv *invocation) *apiGatewayCmds {
	s := &apiGatewayCmds{inv: inv}
	s.apiGatewayCmd = &cobra.Command{Use: "api-gateway", Short: "Manage API Gateway"}
	s.agApisCmd = &cobra.Command{Use: "apis", Short: "Manage API Gateway APIs"}
	s.agAPICreateCmd = &cobra.Command{
		Use: "create API", Short: "Create an API from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runAGAPICreate,
	}
	s.agAPIDeleteCmd = &cobra.Command{
		Use: "delete API", Short: "Delete an API",
		Args: cobra.ExactArgs(1), RunE: s.runAGAPIDelete,
	}
	s.agAPIDescribeCmd = &cobra.Command{
		Use: "describe API", Short: "Describe an API",
		Args: cobra.ExactArgs(1), RunE: s.runAGAPIDescribe,
	}
	s.agAPIListCmd = &cobra.Command{
		Use: "list", Short: "List APIs",
		Args: cobra.NoArgs, RunE: s.runAGAPIList,
	}
	s.agAPIUpdateCmd = &cobra.Command{
		Use: "update API", Short: "Update an API from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runAGAPIUpdate,
	}
	s.agAPIConfigsCmd = &cobra.Command{Use: "api-configs", Short: "Manage API Gateway API configs"}
	s.agACCreateCmd = &cobra.Command{
		Use: "create CONFIG", Short: "Create an API config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runAGACCreate,
	}
	s.agACDeleteCmd = &cobra.Command{
		Use: "delete CONFIG", Short: "Delete an API config",
		Args: cobra.ExactArgs(1), RunE: s.runAGACDelete,
	}
	s.agACDescribeCmd = &cobra.Command{
		Use: "describe CONFIG", Short: "Describe an API config",
		Args: cobra.ExactArgs(1), RunE: s.runAGACDescribe,
	}
	s.agACListCmd = &cobra.Command{
		Use: "list", Short: "List API configs for an API",
		Args: cobra.NoArgs, RunE: s.runAGACList,
	}
	s.agACUpdateCmd = &cobra.Command{
		Use: "update CONFIG", Short: "Update an API config from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runAGACUpdate,
	}
	s.agGatewaysCmd = &cobra.Command{Use: "gateways", Short: "Manage API Gateway gateways"}
	s.agGWCreateCmd = &cobra.Command{
		Use: "create GATEWAY", Short: "Create a gateway from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runAGGWCreate,
	}
	s.agGWDeleteCmd = &cobra.Command{
		Use: "delete GATEWAY", Short: "Delete a gateway",
		Args: cobra.ExactArgs(1), RunE: s.runAGGWDelete,
	}
	s.agGWDescribeCmd = &cobra.Command{
		Use: "describe GATEWAY", Short: "Describe a gateway",
		Args: cobra.ExactArgs(1), RunE: s.runAGGWDescribe,
	}
	s.agGWListCmd = &cobra.Command{
		Use: "list", Short: "List gateways",
		Args: cobra.NoArgs, RunE: s.runAGGWList,
	}
	s.agGWUpdateCmd = &cobra.Command{
		Use: "update GATEWAY", Short: "Update a gateway from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runAGGWUpdate,
	}
	s.agOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage API Gateway operations"}
	s.agOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an operation",
		Args: cobra.ExactArgs(1), RunE: s.runAGOpCancel,
	}
	s.agOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runAGOpDescribe,
	}
	s.agOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: s.runAGOpList,
	}
	s.agOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for an operation to complete",
		Args: cobra.ExactArgs(1), RunE: s.runAGOpWait,
	}
	s.initApiGateway()
	return s
}
