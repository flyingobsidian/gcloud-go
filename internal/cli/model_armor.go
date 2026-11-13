package cli

import (
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

func modelArmorRest() *restClient {
	return newRESTClient("https://modelarmor.googleapis.com/v1")
}

func (s *modelArmorCmds) initModelArmor() {

}

// --- model-armor floorsettings (#1423) ---

func (s *modelArmorCmds) initModelArmor1() {
	for _, c := range []*cobra.Command{s.modelArmorFSDescribeCmd, s.modelArmorFSUpdateCmd} {
		c.Flags().StringVar(&s.flagMAFSFullURI, "full-uri", "",
			"Full resource URI of the FloorSetting (required, e.g. projects/PROJECT/locations/LOCATION/floorSetting)")
		_ = c.MarkFlagRequired("full-uri")
		c.Flags().StringVar(&s.flagMAFSFormat, "format", "", "Output format")
	}
	s.modelArmorFSUpdateCmd.Flags().StringVar(&s.flagMAFSConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the FloorSetting body (required)")
	_ = s.modelArmorFSUpdateCmd.MarkFlagRequired("config-file")
	s.modelArmorFSUpdateCmd.Flags().StringVar(&s.flagMAFSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")

	s.modelArmorFSCmd.AddCommand(s.modelArmorFSDescribeCmd, s.modelArmorFSUpdateCmd)
	s.modelArmorCmd.AddCommand(s.modelArmorFSCmd)
}

func (s *modelArmorCmds) runMAFSDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(modelArmorRest()).do(ctx, http.MethodGet, "/"+strings.TrimPrefix(s.flagMAFSFullURI, "/"), nil, nil, &got); err != nil {
		return fmt.Errorf("describing floor setting: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMAFSFormat)
}

func (s *modelArmorCmds) runMAFSUpdate(cmd *cobra.Command, args []string) error {
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMAFSConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	mask := s.flagMAFSUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(modelArmorRest()).do(ctx, http.MethodPatch, "/"+strings.TrimPrefix(s.flagMAFSFullURI, "/"), q, body, &got); err != nil {
		return fmt.Errorf("updating floor setting: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated floor setting [%s].\n", s.flagMAFSFullURI)
	return emitFormattedTo(s.inv.stdout, got, s.flagMAFSFormat)
}

// --- model-armor templates (#1424) ---

func (s *modelArmorCmds) initModelArmor2() {
	all := []*cobra.Command{
		s.modelArmorTplCreateCmd, s.modelArmorTplDeleteCmd, s.modelArmorTplDescribeCmd,
		s.modelArmorTplListCmd, s.modelArmorTplUpdateCmd,
		s.modelArmorTplSanitizeModelCmd, s.modelArmorTplSanitizeUserCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMATplLocation, "location", "", "Model Armor location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagMATplFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.modelArmorTplCreateCmd, s.modelArmorTplUpdateCmd,
		s.modelArmorTplSanitizeModelCmd, s.modelArmorTplSanitizeUserCmd,
	} {
		c.Flags().StringVar(&s.flagMATplConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.modelArmorTplUpdateCmd.Flags().StringVar(&s.flagMATplUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	s.modelArmorTplListCmd.Flags().Int64Var(&s.flagMATplPageSize, "page-size", 0, "Maximum results per page")
	s.modelArmorTplListCmd.Flags().StringVar(&s.flagMATplFilter, "filter", "", "Server-side filter expression")

	s.modelArmorTplCmd.AddCommand(all...)
	s.modelArmorCmd.AddCommand(s.modelArmorTplCmd)
}

func (s *modelArmorCmds) maTplParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagMATplLocation), nil
}

func (s *modelArmorCmds) maTplName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.maTplParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/templates/%s", parent, id), nil
}

func (s *modelArmorCmds) runMATplCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.maTplParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMATplConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("templateId", args[0])
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(modelArmorRest()).do(ctx, http.MethodPost, "/"+parent+"/templates", q, body, &got); err != nil {
		return fmt.Errorf("creating template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagMATplFormat)
}

func (s *modelArmorCmds) runMATplDelete(cmd *cobra.Command, args []string) error {
	name, err := s.maTplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(modelArmorRest()).do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted template [%s].\n", args[0])
	return nil
}

func (s *modelArmorCmds) runMATplDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.maTplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(modelArmorRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing template: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMATplFormat)
}

func (s *modelArmorCmds) runMATplList(cmd *cobra.Command, args []string) error {
	parent, err := s.maTplParent()
	if err != nil {
		return err
	}
	base := url.Values{}
	if s.flagMATplFilter != "" {
		base.Set("filter", s.flagMATplFilter)
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(modelArmorRest()).paginate(ctx, "/"+parent+"/templates", base, "templates", s.flagMATplPageSize)
	if err != nil {
		return fmt.Errorf("listing templates: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagMATplFormat)
}

func (s *modelArmorCmds) runMATplUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.maTplName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMATplConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	mask := s.flagMATplUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(modelArmorRest()).do(ctx, http.MethodPatch, "/"+name, q, body, &got); err != nil {
		return fmt.Errorf("updating template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagMATplFormat)
}

func (s *modelArmorCmds) runMATplSanitizeModel(cmd *cobra.Command, args []string) error {
	name, err := s.maTplName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMATplConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(modelArmorRest()).do(ctx, http.MethodPost, "/"+name+":sanitizeModelResponse", nil, body, &got); err != nil {
		return fmt.Errorf("sanitizing model response: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMATplFormat)
}

func (s *modelArmorCmds) runMATplSanitizeUser(cmd *cobra.Command, args []string) error {
	name, err := s.maTplName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMATplConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(modelArmorRest()).do(ctx, http.MethodPost, "/"+name+":sanitizeUserPrompt", nil, body, &got); err != nil {
		return fmt.Errorf("sanitizing user prompt: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMATplFormat)
}

// modelArmorCmds holds the commands and flag values of one invocation's "modelArmor"
// command group, so invocations don't share state (#2070).
type modelArmorCmds struct {
	inv *invocation

	modelArmorCmd                 *cobra.Command
	modelArmorFSCmd               *cobra.Command
	flagMAFSFullURI               string
	flagMAFSFormat                string
	flagMAFSConfigFile            string
	flagMAFSUpdateMask            string
	modelArmorFSDescribeCmd       *cobra.Command
	modelArmorFSUpdateCmd         *cobra.Command
	modelArmorTplCmd              *cobra.Command
	flagMATplLocation             string
	flagMATplFormat               string
	flagMATplConfigFile           string
	flagMATplUpdateMask           string
	flagMATplPageSize             int64
	flagMATplFilter               string
	modelArmorTplCreateCmd        *cobra.Command
	modelArmorTplDeleteCmd        *cobra.Command
	modelArmorTplDescribeCmd      *cobra.Command
	modelArmorTplListCmd          *cobra.Command
	modelArmorTplUpdateCmd        *cobra.Command
	modelArmorTplSanitizeModelCmd *cobra.Command
	modelArmorTplSanitizeUserCmd  *cobra.Command
}

// newModelArmorCmd returns the "modelArmor" command group for inv.
func newModelArmorCmd(inv *invocation) *cobra.Command {
	return newModelArmorCmds(inv).modelArmorCmd
}

// newModelArmorCmds builds the "modelArmor" command group for inv and returns its state.
func newModelArmorCmds(inv *invocation) *modelArmorCmds {
	s := &modelArmorCmds{inv: inv}
	s.modelArmorCmd = &cobra.Command{Use: "model-armor", Short: "Manage Google Cloud Model Armor"}
	s.modelArmorFSCmd = &cobra.Command{Use: "floorsettings", Short: "Manage Model Armor floor settings"}
	s.modelArmorFSDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe a FloorSetting resource",
		Args: cobra.NoArgs, RunE: s.runMAFSDescribe,
	}
	s.modelArmorFSUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update a FloorSetting resource",
		Args: cobra.NoArgs, RunE: s.runMAFSUpdate,
	}
	s.modelArmorTplCmd = &cobra.Command{Use: "templates", Short: "Manage Model Armor templates"}
	s.modelArmorTplCreateCmd = &cobra.Command{
		Use: "create TEMPLATE", Short: "Create a Model Armor template",
		Args: cobra.ExactArgs(1), RunE: s.runMATplCreate,
	}
	s.modelArmorTplDeleteCmd = &cobra.Command{
		Use: "delete TEMPLATE", Short: "Delete a Model Armor template",
		Args: cobra.ExactArgs(1), RunE: s.runMATplDelete,
	}
	s.modelArmorTplDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a Model Armor template",
		Args: cobra.ExactArgs(1), RunE: s.runMATplDescribe,
	}
	s.modelArmorTplListCmd = &cobra.Command{
		Use: "list", Short: "List Model Armor templates",
		Args: cobra.NoArgs, RunE: s.runMATplList,
	}
	s.modelArmorTplUpdateCmd = &cobra.Command{
		Use: "update TEMPLATE", Short: "Update a Model Armor template",
		Args: cobra.ExactArgs(1), RunE: s.runMATplUpdate,
	}
	s.modelArmorTplSanitizeModelCmd = &cobra.Command{
		Use:   "sanitize-model-response TEMPLATE",
		Short: "Send a model response through the template's filters",
		Args:  cobra.ExactArgs(1), RunE: s.runMATplSanitizeModel,
	}
	s.modelArmorTplSanitizeUserCmd = &cobra.Command{
		Use:   "sanitize-user-prompt TEMPLATE",
		Short: "Send a user prompt through the template's filters",
		Args:  cobra.ExactArgs(1), RunE: s.runMATplSanitizeUser,
	}
	s.initModelArmor()
	s.initModelArmor1()
	s.initModelArmor2()
	return s
}
