package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud vector-search (#394) ---
//
// The Vector Search API (`vectorsearch.googleapis.com`, v1) is not exposed by
// google.golang.org/api. Both subgroups use the shared restClient from
// rest_helpers.go with a per-service endpoint.

func vectorSearchRest() *restClient {
	return newRESTClient("https://vectorsearch.googleapis.com/v1")
}

func (s *vectorSearchCmds) initVectorSearch() {

}

// --- vector-search collections (#1432) ---

func (s *vectorSearchCmds) initVectorSearch1() {
	all := []*cobra.Command{
		s.vsCollCreateCmd, s.vsCollDeleteCmd, s.vsCollDescribeCmd, s.vsCollListCmd,
		s.vsCollUpdateCmd, s.vsCollImportCmd, s.vsCollExportCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVSCollLocation, "location", "", "Vector Search location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVSCollFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.vsCollCreateCmd, s.vsCollUpdateCmd, s.vsCollImportCmd, s.vsCollExportCmd} {
		c.Flags().StringVar(&s.flagVSCollConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.vsCollUpdateCmd.Flags().StringVar(&s.flagVSCollUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	s.vsCollListCmd.Flags().Int64Var(&s.flagVSCollPageSize, "page-size", 0, "Maximum results per page")

	s.vsCollCmd.AddCommand(all...)
	s.vectorSearchCmd.AddCommand(s.vsCollCmd)
}

func (s *vectorSearchCmds) vsCollParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagVSCollLocation), nil
}

func (s *vectorSearchCmds) vsCollName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.vsCollParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/collections/%s", parent, id), nil
}

func (s *vectorSearchCmds) runVSCollCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.vsCollParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagVSCollConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("collectionId", args[0])
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodPost, "/"+parent+"/collections", q, body, &op); err != nil {
		return fmt.Errorf("creating collection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for collection [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagVSCollFormat)
}

func (s *vectorSearchCmds) runVSCollDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vsCollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting collection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for collection [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagVSCollFormat)
}

func (s *vectorSearchCmds) runVSCollDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vsCollName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing collection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVSCollFormat)
}

func (s *vectorSearchCmds) runVSCollList(cmd *cobra.Command, args []string) error {
	parent, err := s.vsCollParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(vectorSearchRest()).paginate(ctx, "/"+parent+"/collections", nil, "collections", s.flagVSCollPageSize)
	if err != nil {
		return fmt.Errorf("listing collections: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagVSCollFormat)
}

func (s *vectorSearchCmds) runVSCollUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.vsCollName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagVSCollConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	mask := s.flagVSCollUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating collection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for collection [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagVSCollFormat)
}

func (s *vectorSearchCmds) runVSCollImportData(cmd *cobra.Command, args []string) error {
	name, err := s.vsCollName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagVSCollConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodPost, "/"+name+":importDataObjects", nil, body, &op); err != nil {
		return fmt.Errorf("importing data objects: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagVSCollFormat)
}

func (s *vectorSearchCmds) runVSCollExportData(cmd *cobra.Command, args []string) error {
	name, err := s.vsCollName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagVSCollConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodPost, "/"+name+":exportDataObjects", nil, body, &op); err != nil {
		return fmt.Errorf("exporting data objects: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagVSCollFormat)
}

// --- vector-search operations (#1433) ---

func (s *vectorSearchCmds) initVectorSearch2() {
	all := []*cobra.Command{s.vsOpCancelCmd, s.vsOpDeleteCmd, s.vsOpDescribeCmd, s.vsOpListCmd, s.vsOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVSOpLocation, "location", "", "Vector Search location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVSOpFormat, "format", "", "Output format")
	}
	s.vsOpListCmd.Flags().StringVar(&s.flagVSOpFilter, "filter", "", "Server-side filter expression")
	s.vsOpListCmd.Flags().Int64Var(&s.flagVSOpPageSize, "page-size", 0, "Maximum results per page")
	s.vsOpWaitCmd.Flags().DurationVar(&s.flagVSOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to finish")

	s.vsOpCmd.AddCommand(all...)
	s.vectorSearchCmd.AddCommand(s.vsOpCmd)
}

func (s *vectorSearchCmds) vsOpQualifiedName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, s.flagVSOpLocation, id), nil
}

func (s *vectorSearchCmds) runVSOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.vsOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodPost, "/"+name+":cancel", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for operation %s.\n", args[0])
	return nil
}

func (s *vectorSearchCmds) runVSOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vsOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation %s.\n", args[0])
	return nil
}

func (s *vectorSearchCmds) runVSOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vsOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(vectorSearchRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVSOpFormat)
}

func (s *vectorSearchCmds) runVSOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	base := url.Values{}
	if s.flagVSOpFilter != "" {
		base.Set("filter", s.flagVSOpFilter)
	}
	ctx := cmd.Context()
	parent := fmt.Sprintf("projects/%s/locations/%s/operations", project, s.flagVSOpLocation)
	items, err := s.inv.rest(vectorSearchRest()).paginate(ctx, "/"+parent, base, "operations", s.flagVSOpPageSize)
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagVSOpFormat != "" {
		return emitFormattedTo(s.inv.stdout, items, s.flagVSOpFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(items), "%-40s %s\n", "NAME", "DONE")
	for _, o := range items {
		name, _ := o["name"].(string)
		done, _ := o["done"].(bool)
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(name), done)
	}
	return nil
}

func (s *vectorSearchCmds) runVSOpWait(cmd *cobra.Command, args []string) error {
	name, err := s.vsOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	op, err := s.inv.rest(vectorSearchRest()).waitOperation(ctx, name, s.flagVSOpTimeout)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Operation %s completed.\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagVSOpFormat)
}

// vectorSearchCmds holds the commands and flag values of one invocation's "vectorSearch"
// command group, so invocations don't share state (#2070).
type vectorSearchCmds struct {
	inv *invocation

	vectorSearchCmd      *cobra.Command
	vsCollCmd            *cobra.Command
	flagVSCollLocation   string
	flagVSCollFormat     string
	flagVSCollConfigFile string
	flagVSCollUpdateMask string
	flagVSCollPageSize   int64
	vsCollCreateCmd      *cobra.Command
	vsCollDeleteCmd      *cobra.Command
	vsCollDescribeCmd    *cobra.Command
	vsCollListCmd        *cobra.Command
	vsCollUpdateCmd      *cobra.Command
	vsCollImportCmd      *cobra.Command
	vsCollExportCmd      *cobra.Command
	vsOpCmd              *cobra.Command
	flagVSOpLocation     string
	flagVSOpFormat       string
	flagVSOpFilter       string
	flagVSOpPageSize     int64
	flagVSOpTimeout      time.Duration
	vsOpCancelCmd        *cobra.Command
	vsOpDeleteCmd        *cobra.Command
	vsOpDescribeCmd      *cobra.Command
	vsOpListCmd          *cobra.Command
	vsOpWaitCmd          *cobra.Command
}

// newVectorSearchCmd returns the "vectorSearch" command group for inv.
func newVectorSearchCmd(inv *invocation) *cobra.Command {
	return newVectorSearchCmds(inv).vectorSearchCmd
}

// newVectorSearchCmds builds the "vectorSearch" command group for inv and returns its state.
func newVectorSearchCmds(inv *invocation) *vectorSearchCmds {
	s := &vectorSearchCmds{inv: inv}
	s.vectorSearchCmd = &cobra.Command{Use: "vector-search", Short: "Manage Google Cloud Vector Search"}
	s.vsCollCmd = &cobra.Command{Use: "collections", Short: "Manage Vector Search collections"}
	s.vsCollCreateCmd = &cobra.Command{
		Use: "create COLLECTION", Short: "Create a Vector Search collection",
		Args: cobra.ExactArgs(1), RunE: s.runVSCollCreate,
	}
	s.vsCollDeleteCmd = &cobra.Command{
		Use: "delete COLLECTION", Short: "Delete a Vector Search collection",
		Args: cobra.ExactArgs(1), RunE: s.runVSCollDelete,
	}
	s.vsCollDescribeCmd = &cobra.Command{
		Use: "describe COLLECTION", Short: "Describe a Vector Search collection",
		Args: cobra.ExactArgs(1), RunE: s.runVSCollDescribe,
	}
	s.vsCollListCmd = &cobra.Command{
		Use: "list", Short: "List Vector Search collections",
		Args: cobra.NoArgs, RunE: s.runVSCollList,
	}
	s.vsCollUpdateCmd = &cobra.Command{
		Use: "update COLLECTION", Short: "Update a Vector Search collection",
		Args: cobra.ExactArgs(1), RunE: s.runVSCollUpdate,
	}
	s.vsCollImportCmd = &cobra.Command{
		Use:   "import-data-objects COLLECTION",
		Short: "Bulk-import data objects into a collection",
		Args:  cobra.ExactArgs(1), RunE: s.runVSCollImportData,
	}
	s.vsCollExportCmd = &cobra.Command{
		Use:   "export-data-objects COLLECTION",
		Short: "Bulk-export data objects from a collection",
		Args:  cobra.ExactArgs(1), RunE: s.runVSCollExportData,
	}
	s.vsOpCmd = &cobra.Command{Use: "operations", Short: "Manage Vector Search operations"}
	s.vsOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Vector Search operation",
		Args: cobra.ExactArgs(1), RunE: s.runVSOpCancel,
	}
	s.vsOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Vector Search operation record",
		Args: cobra.ExactArgs(1), RunE: s.runVSOpDelete,
	}
	s.vsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Vector Search operation",
		Args: cobra.ExactArgs(1), RunE: s.runVSOpDescribe,
	}
	s.vsOpListCmd = &cobra.Command{
		Use: "list", Short: "List Vector Search operations",
		Args: cobra.NoArgs, RunE: s.runVSOpList,
	}
	s.vsOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Vector Search operation to finish",
		Args: cobra.ExactArgs(1), RunE: s.runVSOpWait,
	}
	s.initVectorSearch()
	s.initVectorSearch1()
	s.initVectorSearch2()
	return s
}
