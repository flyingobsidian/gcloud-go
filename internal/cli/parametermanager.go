package cli

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	parametermanager "google.golang.org/api/parametermanager/v1"
)

// --- gcloud parametermanager (#370, #971) ---

// --- parameters ---

// --- versions (as subtree of parameters) ---

func (g *parameterManagerCmds) initParametermanager() {
	// Parameter-level flags
	g.addPMLocFlag(g.pmParamsCreateCmd, g.pmParamsDeleteCmd, g.pmParamsDescribeCmd, g.pmParamsListCmd, g.pmParamsUpdateCmd)
	g.addPMFormatFlag(g.pmParamsCreateCmd, g.pmParamsDescribeCmd, g.pmParamsListCmd, g.pmParamsUpdateCmd)
	g.pmParamsListCmd.Flags().StringVar(&g.flagPMFilter, "filter", "", "Server-side filter expression")
	g.pmParamsListCmd.Flags().StringVar(&g.flagPMOrderBy, "order-by", "", "Server-side ordering hint")
	g.pmParamsListCmd.Flags().Int64Var(&g.flagPMPageSize, "page-size", 0, "Maximum number of results per page")
	for _, c := range []*cobra.Command{g.pmParamsCreateCmd, g.pmParamsUpdateCmd} {
		c.Flags().StringVar(&g.flagPMFormatVal, "parameter-format", "", "Parameter format: UNFORMATTED, YAML, JSON")
		c.Flags().StringVar(&g.flagPMKmsKey, "kms-key", "", "Customer-managed KMS key (projects/*/locations/*/keyRings/*/cryptoKeys/*)")
		c.Flags().StringSliceVar(&g.flagPMLabels, "labels", nil, "Labels as key=value pairs")
	}
	g.pmParamsUpdateCmd.Flags().StringVar(&g.flagPMUpdateMask, "update-mask", "", "Comma-separated list of fields to update (defaults to every populated field)")

	g.pmParametersCmd.AddCommand(g.pmParamsCreateCmd, g.pmParamsDeleteCmd, g.pmParamsDescribeCmd, g.pmParamsListCmd, g.pmParamsUpdateCmd)

	// Version-level flags
	verAll := []*cobra.Command{g.pmVersionsCreateCmd, g.pmVersionsDeleteCmd, g.pmVersionsDescribeCmd, g.pmVersionsListCmd, g.pmVersionsUpdateCmd, g.pmVersionsRenderCmd, g.pmVersionsAccessCmd}
	g.addPMLocFlag(verAll...)
	g.addPMFormatFlag(g.pmVersionsCreateCmd, g.pmVersionsDescribeCmd, g.pmVersionsListCmd, g.pmVersionsUpdateCmd, g.pmVersionsRenderCmd)
	for _, c := range verAll {
		c.Flags().StringVar(&g.flagPMVerParameter, "parameter", "", "Parent parameter (required)")
		_ = c.MarkFlagRequired("parameter")
	}
	g.pmVersionsListCmd.Flags().StringVar(&g.flagPMFilter, "filter", "", "Server-side filter expression")
	g.pmVersionsListCmd.Flags().StringVar(&g.flagPMOrderBy, "order-by", "", "Server-side ordering hint")
	g.pmVersionsListCmd.Flags().Int64Var(&g.flagPMPageSize, "page-size", 0, "Maximum number of results per page")
	g.pmVersionsDescribeCmd.Flags().StringVar(&g.flagPMView, "view", "", "Response view: BASIC or FULL")
	for _, c := range []*cobra.Command{g.pmVersionsCreateCmd} {
		c.Flags().StringVar(&g.flagPMPayload, "payload-data", "", "Raw payload data (string). Mutually exclusive with --payload-file/--payload-b64")
		c.Flags().StringVar(&g.flagPMPayloadB64, "payload-b64", "", "Base64-encoded payload data. Mutually exclusive with --payload-data/--payload-file")
		c.Flags().StringVar(&g.flagPMPayloadFile, "payload-file", "", "Path to a file containing the payload. Mutually exclusive with --payload-data/--payload-b64")
		c.Flags().BoolVar(&g.flagPMDisabled, "disabled", false, "Create the version in a disabled state")
	}
	g.pmVersionsUpdateCmd.Flags().BoolVar(&g.flagPMDisabled, "disabled", false, "Disable this parameter version")
	g.pmVersionsUpdateCmd.Flags().StringVar(&g.flagPMUpdateMask, "update-mask", "disabled", "Comma-separated list of fields to update (defaults to \"disabled\")")

	g.pmVersionsCmd.AddCommand(verAll...)
	g.pmParametersCmd.AddCommand(g.pmVersionsCmd)

	g.parameterManagerCmd.AddCommand(g.pmParametersCmd)

}

func (g *parameterManagerCmds) addPMLocFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagPMLocation, "location", "global", "Location for the parameter (e.g. \"global\" or a region)")
	}
}

func (g *parameterManagerCmds) addPMFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&g.flagPMFormat, "format", "", "Output format")
	}
}

func (g *parameterManagerCmds) pmParent() (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	loc := g.flagPMLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, loc), nil
}

func (g *parameterManagerCmds) pmParamName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := g.pmParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/parameters/%s", parent, id), nil
}

func (g *parameterManagerCmds) pmVersionName(paramArg, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	paramName, err := g.pmParamName(paramArg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/versions/%s", paramName, id), nil
}

func (g *parameterManagerCmds) pmLabelsFromFlag() map[string]string {
	if len(g.flagPMLabels) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range g.flagPMLabels {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

func (g *parameterManagerCmds) pmReadPayload() (string, error) {
	set := 0
	for _, s := range []string{g.flagPMPayload, g.flagPMPayloadB64, g.flagPMPayloadFile} {
		if s != "" {
			set++
		}
	}
	if set == 0 {
		return "", fmt.Errorf("one of --payload-data, --payload-b64, or --payload-file is required")
	}
	if set > 1 {
		return "", fmt.Errorf("only one of --payload-data, --payload-b64, or --payload-file may be set")
	}
	switch {
	case g.flagPMPayload != "":
		return base64.StdEncoding.EncodeToString([]byte(g.flagPMPayload)), nil
	case g.flagPMPayloadB64 != "":
		if _, err := base64.StdEncoding.DecodeString(g.flagPMPayloadB64); err != nil {
			return "", fmt.Errorf("--payload-b64: %w", err)
		}
		return g.flagPMPayloadB64, nil
	default:
		data, err := os.ReadFile(g.flagPMPayloadFile)
		if err != nil {
			return "", fmt.Errorf("reading payload file: %w", err)
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}
}

// --- parameters impl ---

func (g *parameterManagerCmds) runPMParamCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.pmParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	body := &parametermanager.Parameter{
		Format: strings.ToUpper(g.flagPMFormatVal),
		KmsKey: g.flagPMKmsKey,
		Labels: g.pmLabelsFromFlag(),
	}
	got, err := svc.Projects.Locations.Parameters.Create(parent, body).ParameterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating parameter: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPMFormat)
}

func (g *parameterManagerCmds) runPMParamDelete(cmd *cobra.Command, args []string) error {
	name, err := g.pmParamName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Parameters.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting parameter: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted parameter [%s].\n", args[0])
	return nil
}

func (g *parameterManagerCmds) runPMParamDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.pmParamName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Parameters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing parameter: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPMFormat)
}

func (g *parameterManagerCmds) runPMParamList(cmd *cobra.Command, args []string) error {
	parent, err := g.pmParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	var all []*parametermanager.Parameter
	pageToken := ""
	for {
		call := svc.Projects.Locations.Parameters.List(parent).Context(ctx)
		if g.flagPMFilter != "" {
			call = call.Filter(g.flagPMFilter)
		}
		if g.flagPMOrderBy != "" {
			call = call.OrderBy(g.flagPMOrderBy)
		}
		if g.flagPMPageSize > 0 {
			call = call.PageSize(g.flagPMPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing parameters: %w", err)
		}
		all = append(all, resp.Parameters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagPMFormat)
}

func (g *parameterManagerCmds) runPMParamUpdate(cmd *cobra.Command, args []string) error {
	name, err := g.pmParamName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	body := &parametermanager.Parameter{
		Format: strings.ToUpper(g.flagPMFormatVal),
		KmsKey: g.flagPMKmsKey,
		Labels: g.pmLabelsFromFlag(),
	}
	mask := g.flagPMUpdateMask
	if mask == "" {
		var fields []string
		if body.Format != "" {
			fields = append(fields, "format")
		}
		if body.KmsKey != "" {
			fields = append(fields, "kmsKey")
		}
		if body.Labels != nil {
			fields = append(fields, "labels")
		}
		mask = strings.Join(fields, ",")
	}
	call := svc.Projects.Locations.Parameters.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating parameter: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPMFormat)
}

// --- versions impl ---

func (g *parameterManagerCmds) runPMVerCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.pmParamName(g.flagPMVerParameter)
	if err != nil {
		return err
	}
	payload, err := g.pmReadPayload()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	body := &parametermanager.ParameterVersion{
		Disabled: g.flagPMDisabled,
		Payload:  &parametermanager.ParameterVersionPayload{Data: payload},
	}
	got, err := svc.Projects.Locations.Parameters.Versions.Create(parent, body).ParameterVersionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating parameter version: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPMFormat)
}

func (g *parameterManagerCmds) runPMVerDelete(cmd *cobra.Command, args []string) error {
	name, err := g.pmVersionName(g.flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Parameters.Versions.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting parameter version: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted parameter version [%s].\n", args[0])
	return nil
}

func (g *parameterManagerCmds) runPMVerDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.pmVersionName(g.flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Parameters.Versions.Get(name).Context(ctx)
	if g.flagPMView != "" {
		call = call.View(strings.ToUpper(g.flagPMView))
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing parameter version: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPMFormat)
}

func (g *parameterManagerCmds) runPMVerList(cmd *cobra.Command, args []string) error {
	parent, err := g.pmParamName(g.flagPMVerParameter)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	var all []*parametermanager.ParameterVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.Parameters.Versions.List(parent).Context(ctx)
		if g.flagPMFilter != "" {
			call = call.Filter(g.flagPMFilter)
		}
		if g.flagPMOrderBy != "" {
			call = call.OrderBy(g.flagPMOrderBy)
		}
		if g.flagPMPageSize > 0 {
			call = call.PageSize(g.flagPMPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing parameter versions: %w", err)
		}
		all = append(all, resp.ParameterVersions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagPMFormat)
}

func (g *parameterManagerCmds) runPMVerUpdate(cmd *cobra.Command, args []string) error {
	name, err := g.pmVersionName(g.flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	body := &parametermanager.ParameterVersion{Disabled: g.flagPMDisabled}
	// Disabled is a bool and defaults to false; force-send so a false value
	// still transmits when explicitly requested via --update-mask=disabled.
	body.ForceSendFields = []string{"Disabled"}
	call := svc.Projects.Locations.Parameters.Versions.Patch(name, body).Context(ctx)
	if g.flagPMUpdateMask != "" {
		call = call.UpdateMask(g.flagPMUpdateMask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating parameter version: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPMFormat)
}

func (g *parameterManagerCmds) runPMVerRender(cmd *cobra.Command, args []string) error {
	name, err := g.pmVersionName(g.flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Parameters.Versions.Render(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rendering parameter version: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPMFormat)
}

func (g *parameterManagerCmds) runPMVerAccess(cmd *cobra.Command, args []string) error {
	name, err := g.pmVersionName(g.flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ParameterManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Parameters.Versions.Get(name).View("FULL").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("accessing parameter version: %w", err)
	}
	if got.Payload == nil || got.Payload.Data == "" {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(got.Payload.Data)
	if err != nil {
		return fmt.Errorf("decoding payload: %w", err)
	}
	_, err = g.inv.stdout.Write(data)
	return err
}

// parameterManagerCmds holds the commands and flag values of one invocation's "parameterManager"
// command group, so invocations don't share state (#2070).
type parameterManagerCmds struct {
	inv *invocation

	parameterManagerCmd   *cobra.Command
	flagPMLocation        string
	flagPMFormat          string
	flagPMFilter          string
	flagPMOrderBy         string
	flagPMPageSize        int64
	flagPMFormatVal       string
	flagPMKmsKey          string
	flagPMLabels          []string
	flagPMUpdateMask      string
	flagPMPayload         string
	flagPMPayloadB64      string
	flagPMPayloadFile     string
	flagPMDisabled        bool
	flagPMView            string
	pmParametersCmd       *cobra.Command
	pmParamsCreateCmd     *cobra.Command
	pmParamsDeleteCmd     *cobra.Command
	pmParamsDescribeCmd   *cobra.Command
	pmParamsListCmd       *cobra.Command
	pmParamsUpdateCmd     *cobra.Command
	pmVersionsCmd         *cobra.Command
	pmVersionsCreateCmd   *cobra.Command
	pmVersionsDeleteCmd   *cobra.Command
	pmVersionsDescribeCmd *cobra.Command
	pmVersionsListCmd     *cobra.Command
	pmVersionsUpdateCmd   *cobra.Command
	pmVersionsRenderCmd   *cobra.Command
	pmVersionsAccessCmd   *cobra.Command
	flagPMVerParameter    string
}

// newParameterManagerCmd returns the "parameterManager" command group for inv.
func newParameterManagerCmd(inv *invocation) *cobra.Command {
	return newParameterManagerCmds(inv).parameterManagerCmd
}

// newParameterManagerCmds builds the "parameterManager" command group for inv and returns its state.
func newParameterManagerCmds(inv *invocation) *parameterManagerCmds {
	g := &parameterManagerCmds{inv: inv}
	g.parameterManagerCmd = &cobra.Command{
		Use:   "parametermanager",
		Short: "Manage Parameter Manager",
	}
	g.pmParametersCmd = &cobra.Command{Use: "parameters", Short: "Manage parameters"}
	g.pmParamsCreateCmd = &cobra.Command{
		Use: "create PARAMETER", Short: "Create a parameter",
		Args: cobra.ExactArgs(1), RunE: g.runPMParamCreate,
	}
	g.pmParamsDeleteCmd = &cobra.Command{
		Use: "delete PARAMETER", Short: "Delete a parameter",
		Args: cobra.ExactArgs(1), RunE: g.runPMParamDelete,
	}
	g.pmParamsDescribeCmd = &cobra.Command{
		Use: "describe PARAMETER", Short: "Describe a parameter",
		Args: cobra.ExactArgs(1), RunE: g.runPMParamDescribe,
	}
	g.pmParamsListCmd = &cobra.Command{
		Use: "list", Short: "List parameters",
		Args: cobra.NoArgs, RunE: g.runPMParamList,
	}
	g.pmParamsUpdateCmd = &cobra.Command{
		Use: "update PARAMETER", Short: "Update a parameter",
		Args: cobra.ExactArgs(1), RunE: g.runPMParamUpdate,
	}
	g.pmVersionsCmd = &cobra.Command{Use: "versions", Short: "Manage parameter versions"}
	g.pmVersionsCreateCmd = &cobra.Command{
		Use: "create VERSION --parameter=PARAMETER", Short: "Create a parameter version",
		Args: cobra.ExactArgs(1), RunE: g.runPMVerCreate,
	}
	g.pmVersionsDeleteCmd = &cobra.Command{
		Use: "delete VERSION --parameter=PARAMETER", Short: "Delete a parameter version",
		Args: cobra.ExactArgs(1), RunE: g.runPMVerDelete,
	}
	g.pmVersionsDescribeCmd = &cobra.Command{
		Use: "describe VERSION --parameter=PARAMETER", Short: "Describe a parameter version",
		Args: cobra.ExactArgs(1), RunE: g.runPMVerDescribe,
	}
	g.pmVersionsListCmd = &cobra.Command{
		Use: "list --parameter=PARAMETER", Short: "List parameter versions",
		Args: cobra.NoArgs, RunE: g.runPMVerList,
	}
	g.pmVersionsUpdateCmd = &cobra.Command{
		Use: "update VERSION --parameter=PARAMETER", Short: "Enable or disable a parameter version",
		Args: cobra.ExactArgs(1), RunE: g.runPMVerUpdate,
	}
	g.pmVersionsRenderCmd = &cobra.Command{
		Use: "render VERSION --parameter=PARAMETER", Short: "Render a parameter version (resolve secret references)",
		Args: cobra.ExactArgs(1), RunE: g.runPMVerRender,
	}
	g.pmVersionsAccessCmd = &cobra.Command{
		Use: "access VERSION --parameter=PARAMETER", Short: "Print a parameter version's payload data",
		Args: cobra.ExactArgs(1), RunE: g.runPMVerAccess,
	}
	g.initParametermanager()
	return g
}
