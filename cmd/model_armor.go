package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud model-armor (#359) ---
//
// The Model Armor API (`modelarmor.googleapis.com`, v1) is not exposed by
// google.golang.org/api; both subgroups use the shared restClient from
// rest_helpers.go with a per-service endpoint.

var modelArmorCmd = &cobra.Command{Use: "model-armor", Short: "Manage Google Cloud Model Armor"}

var modelArmorRest = newRESTClient("https://modelarmor.googleapis.com/v1")

func init() {
	rootCmd.AddCommand(modelArmorCmd)
}

// --- model-armor floorsettings (#1423) ---

var modelArmorFSCmd = &cobra.Command{Use: "floorsettings", Short: "Manage Model Armor floor settings"}

var (
	flagMAFSFullURI    string
	flagMAFSFormat     string
	flagMAFSConfigFile string
	flagMAFSUpdateMask string
)

var (
	modelArmorFSDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe a FloorSetting resource",
		Args: cobra.NoArgs, RunE: runMAFSDescribe,
	}
	modelArmorFSUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update a FloorSetting resource",
		Args: cobra.NoArgs, RunE: runMAFSUpdate,
	}
)

func init() {
	for _, c := range []*cobra.Command{modelArmorFSDescribeCmd, modelArmorFSUpdateCmd} {
		c.Flags().StringVar(&flagMAFSFullURI, "full-uri", "",
			"Full resource URI of the FloorSetting (required, e.g. projects/PROJECT/locations/LOCATION/floorSetting)")
		_ = c.MarkFlagRequired("full-uri")
		c.Flags().StringVar(&flagMAFSFormat, "format", "", "Output format")
	}
	modelArmorFSUpdateCmd.Flags().StringVar(&flagMAFSConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the FloorSetting body (required)")
	_ = modelArmorFSUpdateCmd.MarkFlagRequired("config-file")
	modelArmorFSUpdateCmd.Flags().StringVar(&flagMAFSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")

	modelArmorFSCmd.AddCommand(modelArmorFSDescribeCmd, modelArmorFSUpdateCmd)
	modelArmorCmd.AddCommand(modelArmorFSCmd)
}

func runMAFSDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	var got map[string]any
	if err := modelArmorRest.do(ctx, http.MethodGet, "/"+strings.TrimPrefix(flagMAFSFullURI, "/"), nil, nil, &got); err != nil {
		return fmt.Errorf("describing floor setting: %w", err)
	}
	return emitFormatted(got, flagMAFSFormat)
}

func runMAFSUpdate(cmd *cobra.Command, args []string) error {
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMAFSConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	mask := flagMAFSUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := context.Background()
	var got map[string]any
	if err := modelArmorRest.do(ctx, http.MethodPatch, "/"+strings.TrimPrefix(flagMAFSFullURI, "/"), q, body, &got); err != nil {
		return fmt.Errorf("updating floor setting: %w", err)
	}
	fmt.Printf("Updated floor setting [%s].\n", flagMAFSFullURI)
	return emitFormatted(got, flagMAFSFormat)
}

// --- model-armor templates (#1424) ---

var modelArmorTplCmd = &cobra.Command{Use: "templates", Short: "Manage Model Armor templates"}

var (
	flagMATplLocation   string
	flagMATplFormat     string
	flagMATplConfigFile string
	flagMATplUpdateMask string
	flagMATplPageSize   int64
	flagMATplFilter     string
)

var (
	modelArmorTplCreateCmd = &cobra.Command{
		Use: "create TEMPLATE", Short: "Create a Model Armor template",
		Args: cobra.ExactArgs(1), RunE: runMATplCreate,
	}
	modelArmorTplDeleteCmd = &cobra.Command{
		Use: "delete TEMPLATE", Short: "Delete a Model Armor template",
		Args: cobra.ExactArgs(1), RunE: runMATplDelete,
	}
	modelArmorTplDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a Model Armor template",
		Args: cobra.ExactArgs(1), RunE: runMATplDescribe,
	}
	modelArmorTplListCmd = &cobra.Command{
		Use: "list", Short: "List Model Armor templates",
		Args: cobra.NoArgs, RunE: runMATplList,
	}
	modelArmorTplUpdateCmd = &cobra.Command{
		Use: "update TEMPLATE", Short: "Update a Model Armor template",
		Args: cobra.ExactArgs(1), RunE: runMATplUpdate,
	}
	modelArmorTplSanitizeModelCmd = &cobra.Command{
		Use: "sanitize-model-response TEMPLATE",
		Short: "Send a model response through the template's filters",
		Args:  cobra.ExactArgs(1), RunE: runMATplSanitizeModel,
	}
	modelArmorTplSanitizeUserCmd = &cobra.Command{
		Use: "sanitize-user-prompt TEMPLATE",
		Short: "Send a user prompt through the template's filters",
		Args:  cobra.ExactArgs(1), RunE: runMATplSanitizeUser,
	}
)

func init() {
	all := []*cobra.Command{
		modelArmorTplCreateCmd, modelArmorTplDeleteCmd, modelArmorTplDescribeCmd,
		modelArmorTplListCmd, modelArmorTplUpdateCmd,
		modelArmorTplSanitizeModelCmd, modelArmorTplSanitizeUserCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagMATplLocation, "location", "", "Model Armor location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagMATplFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		modelArmorTplCreateCmd, modelArmorTplUpdateCmd,
		modelArmorTplSanitizeModelCmd, modelArmorTplSanitizeUserCmd,
	} {
		c.Flags().StringVar(&flagMATplConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	modelArmorTplUpdateCmd.Flags().StringVar(&flagMATplUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	modelArmorTplListCmd.Flags().Int64Var(&flagMATplPageSize, "page-size", 0, "Maximum results per page")
	modelArmorTplListCmd.Flags().StringVar(&flagMATplFilter, "filter", "", "Server-side filter expression")

	modelArmorTplCmd.AddCommand(all...)
	modelArmorCmd.AddCommand(modelArmorTplCmd)
}

func maTplParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagMATplLocation), nil
}

func maTplName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := maTplParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/templates/%s", parent, id), nil
}

func runMATplCreate(cmd *cobra.Command, args []string) error {
	parent, err := maTplParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMATplConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("templateId", args[0])
	ctx := context.Background()
	var got map[string]any
	if err := modelArmorRest.do(ctx, http.MethodPost, "/"+parent+"/templates", q, body, &got); err != nil {
		return fmt.Errorf("creating template: %w", err)
	}
	fmt.Printf("Created template [%s].\n", args[0])
	return emitFormatted(got, flagMATplFormat)
}

func runMATplDelete(cmd *cobra.Command, args []string) error {
	name, err := maTplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := modelArmorRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting template: %w", err)
	}
	fmt.Printf("Deleted template [%s].\n", args[0])
	return nil
}

func runMATplDescribe(cmd *cobra.Command, args []string) error {
	name, err := maTplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := modelArmorRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing template: %w", err)
	}
	return emitFormatted(got, flagMATplFormat)
}

func runMATplList(cmd *cobra.Command, args []string) error {
	parent, err := maTplParent()
	if err != nil {
		return err
	}
	base := url.Values{}
	if flagMATplFilter != "" {
		base.Set("filter", flagMATplFilter)
	}
	ctx := context.Background()
	items, err := modelArmorRest.paginate(ctx, "/"+parent+"/templates", base, "templates", flagMATplPageSize)
	if err != nil {
		return fmt.Errorf("listing templates: %w", err)
	}
	return emitFormatted(items, flagMATplFormat)
}

func runMATplUpdate(cmd *cobra.Command, args []string) error {
	name, err := maTplName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMATplConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	mask := flagMATplUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := context.Background()
	var got map[string]any
	if err := modelArmorRest.do(ctx, http.MethodPatch, "/"+name, q, body, &got); err != nil {
		return fmt.Errorf("updating template: %w", err)
	}
	fmt.Printf("Updated template [%s].\n", args[0])
	return emitFormatted(got, flagMATplFormat)
}

func runMATplSanitizeModel(cmd *cobra.Command, args []string) error {
	name, err := maTplName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMATplConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := modelArmorRest.do(ctx, http.MethodPost, "/"+name+":sanitizeModelResponse", nil, body, &got); err != nil {
		return fmt.Errorf("sanitizing model response: %w", err)
	}
	return emitFormatted(got, flagMATplFormat)
}

func runMATplSanitizeUser(cmd *cobra.Command, args []string) error {
	name, err := maTplName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMATplConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := modelArmorRest.do(ctx, http.MethodPost, "/"+name+":sanitizeUserPrompt", nil, body, &got); err != nil {
		return fmt.Errorf("sanitizing user prompt: %w", err)
	}
	return emitFormatted(got, flagMATplFormat)
}
