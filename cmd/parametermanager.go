package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	parametermanager "google.golang.org/api/parametermanager/v1"
)

// --- gcloud parametermanager (#370, #971) ---

var parameterManagerCmd = &cobra.Command{
	Use:   "parametermanager",
	Short: "Manage Parameter Manager",
}

var (
	flagPMLocation   string
	flagPMFormat     string
	flagPMFilter     string
	flagPMOrderBy    string
	flagPMPageSize   int64
	flagPMFormatVal  string
	flagPMKmsKey     string
	flagPMLabels     []string
	flagPMUpdateMask string
	flagPMPayload    string
	flagPMPayloadB64 string
	flagPMPayloadFile string
	flagPMDisabled   bool
	flagPMView       string
)

// --- parameters ---

var pmParametersCmd = &cobra.Command{Use: "parameters", Short: "Manage parameters"}

var (
	pmParamsCreateCmd = &cobra.Command{
		Use: "create PARAMETER", Short: "Create a parameter",
		Args: cobra.ExactArgs(1), RunE: runPMParamCreate,
	}
	pmParamsDeleteCmd = &cobra.Command{
		Use: "delete PARAMETER", Short: "Delete a parameter",
		Args: cobra.ExactArgs(1), RunE: runPMParamDelete,
	}
	pmParamsDescribeCmd = &cobra.Command{
		Use: "describe PARAMETER", Short: "Describe a parameter",
		Args: cobra.ExactArgs(1), RunE: runPMParamDescribe,
	}
	pmParamsListCmd = &cobra.Command{
		Use: "list", Short: "List parameters",
		Args: cobra.NoArgs, RunE: runPMParamList,
	}
	pmParamsUpdateCmd = &cobra.Command{
		Use: "update PARAMETER", Short: "Update a parameter",
		Args: cobra.ExactArgs(1), RunE: runPMParamUpdate,
	}
)

// --- versions (as subtree of parameters) ---

var pmVersionsCmd = &cobra.Command{Use: "versions", Short: "Manage parameter versions"}

var (
	pmVersionsCreateCmd = &cobra.Command{
		Use: "create VERSION --parameter=PARAMETER", Short: "Create a parameter version",
		Args: cobra.ExactArgs(1), RunE: runPMVerCreate,
	}
	pmVersionsDeleteCmd = &cobra.Command{
		Use: "delete VERSION --parameter=PARAMETER", Short: "Delete a parameter version",
		Args: cobra.ExactArgs(1), RunE: runPMVerDelete,
	}
	pmVersionsDescribeCmd = &cobra.Command{
		Use: "describe VERSION --parameter=PARAMETER", Short: "Describe a parameter version",
		Args: cobra.ExactArgs(1), RunE: runPMVerDescribe,
	}
	pmVersionsListCmd = &cobra.Command{
		Use: "list --parameter=PARAMETER", Short: "List parameter versions",
		Args: cobra.NoArgs, RunE: runPMVerList,
	}
	pmVersionsUpdateCmd = &cobra.Command{
		Use: "update VERSION --parameter=PARAMETER", Short: "Enable or disable a parameter version",
		Args: cobra.ExactArgs(1), RunE: runPMVerUpdate,
	}
	pmVersionsRenderCmd = &cobra.Command{
		Use: "render VERSION --parameter=PARAMETER", Short: "Render a parameter version (resolve secret references)",
		Args: cobra.ExactArgs(1), RunE: runPMVerRender,
	}
	pmVersionsAccessCmd = &cobra.Command{
		Use: "access VERSION --parameter=PARAMETER", Short: "Print a parameter version's payload data",
		Args: cobra.ExactArgs(1), RunE: runPMVerAccess,
	}
)

var flagPMVerParameter string

func init() {
	// Parameter-level flags
	addPMLocFlag(pmParamsCreateCmd, pmParamsDeleteCmd, pmParamsDescribeCmd, pmParamsListCmd, pmParamsUpdateCmd)
	addPMFormatFlag(pmParamsCreateCmd, pmParamsDescribeCmd, pmParamsListCmd, pmParamsUpdateCmd)
	pmParamsListCmd.Flags().StringVar(&flagPMFilter, "filter", "", "Server-side filter expression")
	pmParamsListCmd.Flags().StringVar(&flagPMOrderBy, "order-by", "", "Server-side ordering hint")
	pmParamsListCmd.Flags().Int64Var(&flagPMPageSize, "page-size", 0, "Maximum number of results per page")
	for _, c := range []*cobra.Command{pmParamsCreateCmd, pmParamsUpdateCmd} {
		c.Flags().StringVar(&flagPMFormatVal, "parameter-format", "", "Parameter format: UNFORMATTED, YAML, JSON")
		c.Flags().StringVar(&flagPMKmsKey, "kms-key", "", "Customer-managed KMS key (projects/*/locations/*/keyRings/*/cryptoKeys/*)")
		c.Flags().StringSliceVar(&flagPMLabels, "labels", nil, "Labels as key=value pairs")
	}
	pmParamsUpdateCmd.Flags().StringVar(&flagPMUpdateMask, "update-mask", "", "Comma-separated list of fields to update (defaults to every populated field)")

	pmParametersCmd.AddCommand(pmParamsCreateCmd, pmParamsDeleteCmd, pmParamsDescribeCmd, pmParamsListCmd, pmParamsUpdateCmd)

	// Version-level flags
	verAll := []*cobra.Command{pmVersionsCreateCmd, pmVersionsDeleteCmd, pmVersionsDescribeCmd, pmVersionsListCmd, pmVersionsUpdateCmd, pmVersionsRenderCmd, pmVersionsAccessCmd}
	addPMLocFlag(verAll...)
	addPMFormatFlag(pmVersionsCreateCmd, pmVersionsDescribeCmd, pmVersionsListCmd, pmVersionsUpdateCmd, pmVersionsRenderCmd)
	for _, c := range verAll {
		c.Flags().StringVar(&flagPMVerParameter, "parameter", "", "Parent parameter (required)")
		_ = c.MarkFlagRequired("parameter")
	}
	pmVersionsListCmd.Flags().StringVar(&flagPMFilter, "filter", "", "Server-side filter expression")
	pmVersionsListCmd.Flags().StringVar(&flagPMOrderBy, "order-by", "", "Server-side ordering hint")
	pmVersionsListCmd.Flags().Int64Var(&flagPMPageSize, "page-size", 0, "Maximum number of results per page")
	pmVersionsDescribeCmd.Flags().StringVar(&flagPMView, "view", "", "Response view: BASIC or FULL")
	for _, c := range []*cobra.Command{pmVersionsCreateCmd} {
		c.Flags().StringVar(&flagPMPayload, "payload-data", "", "Raw payload data (string). Mutually exclusive with --payload-file/--payload-b64")
		c.Flags().StringVar(&flagPMPayloadB64, "payload-b64", "", "Base64-encoded payload data. Mutually exclusive with --payload-data/--payload-file")
		c.Flags().StringVar(&flagPMPayloadFile, "payload-file", "", "Path to a file containing the payload. Mutually exclusive with --payload-data/--payload-b64")
		c.Flags().BoolVar(&flagPMDisabled, "disabled", false, "Create the version in a disabled state")
	}
	pmVersionsUpdateCmd.Flags().BoolVar(&flagPMDisabled, "disabled", false, "Disable this parameter version")
	pmVersionsUpdateCmd.Flags().StringVar(&flagPMUpdateMask, "update-mask", "disabled", "Comma-separated list of fields to update (defaults to \"disabled\")")

	pmVersionsCmd.AddCommand(verAll...)
	pmParametersCmd.AddCommand(pmVersionsCmd)

	parameterManagerCmd.AddCommand(pmParametersCmd)
	rootCmd.AddCommand(parameterManagerCmd)
}

func addPMLocFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagPMLocation, "location", "global", "Location for the parameter (e.g. \"global\" or a region)")
	}
}

func addPMFormatFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVar(&flagPMFormat, "format", "", "Output format")
	}
}

func pmParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	loc := flagPMLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, loc), nil
}

func pmParamName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := pmParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/parameters/%s", parent, id), nil
}

func pmVersionName(paramArg, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	paramName, err := pmParamName(paramArg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/versions/%s", paramName, id), nil
}

func pmLabelsFromFlag() map[string]string {
	if len(flagPMLabels) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, kv := range flagPMLabels {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

func pmReadPayload() (string, error) {
	set := 0
	for _, s := range []string{flagPMPayload, flagPMPayloadB64, flagPMPayloadFile} {
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
	case flagPMPayload != "":
		return base64.StdEncoding.EncodeToString([]byte(flagPMPayload)), nil
	case flagPMPayloadB64 != "":
		if _, err := base64.StdEncoding.DecodeString(flagPMPayloadB64); err != nil {
			return "", fmt.Errorf("--payload-b64: %w", err)
		}
		return flagPMPayloadB64, nil
	default:
		data, err := os.ReadFile(flagPMPayloadFile)
		if err != nil {
			return "", fmt.Errorf("reading payload file: %w", err)
		}
		return base64.StdEncoding.EncodeToString(data), nil
	}
}

// --- parameters impl ---

func runPMParamCreate(cmd *cobra.Command, args []string) error {
	parent, err := pmParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &parametermanager.Parameter{
		Format: strings.ToUpper(flagPMFormatVal),
		KmsKey: flagPMKmsKey,
		Labels: pmLabelsFromFlag(),
	}
	got, err := svc.Projects.Locations.Parameters.Create(parent, body).ParameterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating parameter: %w", err)
	}
	return emitFormatted(got, flagPMFormat)
}

func runPMParamDelete(cmd *cobra.Command, args []string) error {
	name, err := pmParamName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Parameters.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting parameter: %w", err)
	}
	fmt.Printf("Deleted parameter [%s].\n", args[0])
	return nil
}

func runPMParamDescribe(cmd *cobra.Command, args []string) error {
	name, err := pmParamName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Parameters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing parameter: %w", err)
	}
	return emitFormatted(got, flagPMFormat)
}

func runPMParamList(cmd *cobra.Command, args []string) error {
	parent, err := pmParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*parametermanager.Parameter
	pageToken := ""
	for {
		call := svc.Projects.Locations.Parameters.List(parent).Context(ctx)
		if flagPMFilter != "" {
			call = call.Filter(flagPMFilter)
		}
		if flagPMOrderBy != "" {
			call = call.OrderBy(flagPMOrderBy)
		}
		if flagPMPageSize > 0 {
			call = call.PageSize(flagPMPageSize)
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
	return emitFormatted(all, flagPMFormat)
}

func runPMParamUpdate(cmd *cobra.Command, args []string) error {
	name, err := pmParamName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &parametermanager.Parameter{
		Format: strings.ToUpper(flagPMFormatVal),
		KmsKey: flagPMKmsKey,
		Labels: pmLabelsFromFlag(),
	}
	mask := flagPMUpdateMask
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
	return emitFormatted(got, flagPMFormat)
}

// --- versions impl ---

func runPMVerCreate(cmd *cobra.Command, args []string) error {
	parent, err := pmParamName(flagPMVerParameter)
	if err != nil {
		return err
	}
	payload, err := pmReadPayload()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &parametermanager.ParameterVersion{
		Disabled: flagPMDisabled,
		Payload:  &parametermanager.ParameterVersionPayload{Data: payload},
	}
	got, err := svc.Projects.Locations.Parameters.Versions.Create(parent, body).ParameterVersionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating parameter version: %w", err)
	}
	return emitFormatted(got, flagPMFormat)
}

func runPMVerDelete(cmd *cobra.Command, args []string) error {
	name, err := pmVersionName(flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Parameters.Versions.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting parameter version: %w", err)
	}
	fmt.Printf("Deleted parameter version [%s].\n", args[0])
	return nil
}

func runPMVerDescribe(cmd *cobra.Command, args []string) error {
	name, err := pmVersionName(flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Parameters.Versions.Get(name).Context(ctx)
	if flagPMView != "" {
		call = call.View(strings.ToUpper(flagPMView))
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing parameter version: %w", err)
	}
	return emitFormatted(got, flagPMFormat)
}

func runPMVerList(cmd *cobra.Command, args []string) error {
	parent, err := pmParamName(flagPMVerParameter)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*parametermanager.ParameterVersion
	pageToken := ""
	for {
		call := svc.Projects.Locations.Parameters.Versions.List(parent).Context(ctx)
		if flagPMFilter != "" {
			call = call.Filter(flagPMFilter)
		}
		if flagPMOrderBy != "" {
			call = call.OrderBy(flagPMOrderBy)
		}
		if flagPMPageSize > 0 {
			call = call.PageSize(flagPMPageSize)
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
	return emitFormatted(all, flagPMFormat)
}

func runPMVerUpdate(cmd *cobra.Command, args []string) error {
	name, err := pmVersionName(flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &parametermanager.ParameterVersion{Disabled: flagPMDisabled}
	// Disabled is a bool and defaults to false; force-send so a false value
	// still transmits when explicitly requested via --update-mask=disabled.
	body.ForceSendFields = []string{"Disabled"}
	call := svc.Projects.Locations.Parameters.Versions.Patch(name, body).Context(ctx)
	if flagPMUpdateMask != "" {
		call = call.UpdateMask(flagPMUpdateMask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating parameter version: %w", err)
	}
	return emitFormatted(got, flagPMFormat)
}

func runPMVerRender(cmd *cobra.Command, args []string) error {
	name, err := pmVersionName(flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Parameters.Versions.Render(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rendering parameter version: %w", err)
	}
	return emitFormatted(got, flagPMFormat)
}

func runPMVerAccess(cmd *cobra.Command, args []string) error {
	name, err := pmVersionName(flagPMVerParameter, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ParameterManagerService(ctx, flagAccount)
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
	_, err = os.Stdout.Write(data)
	return err
}
