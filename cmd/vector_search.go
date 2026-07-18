package cmd

import (
	"context"
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

var vectorSearchCmd = &cobra.Command{Use: "vector-search", Short: "Manage Google Cloud Vector Search"}

var vectorSearchRest = newRESTClient("https://vectorsearch.googleapis.com/v1")

func init() {
	rootCmd.AddCommand(vectorSearchCmd)
}

// --- vector-search collections (#1432) ---

var vsCollCmd = &cobra.Command{Use: "collections", Short: "Manage Vector Search collections"}

var (
	flagVSCollLocation   string
	flagVSCollFormat     string
	flagVSCollConfigFile string
	flagVSCollUpdateMask string
	flagVSCollPageSize   int64
	flagVSCollImportFile string
	flagVSCollExportFile string
)

var (
	vsCollCreateCmd = &cobra.Command{
		Use: "create COLLECTION", Short: "Create a Vector Search collection",
		Args: cobra.ExactArgs(1), RunE: runVSCollCreate,
	}
	vsCollDeleteCmd = &cobra.Command{
		Use: "delete COLLECTION", Short: "Delete a Vector Search collection",
		Args: cobra.ExactArgs(1), RunE: runVSCollDelete,
	}
	vsCollDescribeCmd = &cobra.Command{
		Use: "describe COLLECTION", Short: "Describe a Vector Search collection",
		Args: cobra.ExactArgs(1), RunE: runVSCollDescribe,
	}
	vsCollListCmd = &cobra.Command{
		Use: "list", Short: "List Vector Search collections",
		Args: cobra.NoArgs, RunE: runVSCollList,
	}
	vsCollUpdateCmd = &cobra.Command{
		Use: "update COLLECTION", Short: "Update a Vector Search collection",
		Args: cobra.ExactArgs(1), RunE: runVSCollUpdate,
	}
	vsCollImportCmd = &cobra.Command{
		Use: "import-data-objects COLLECTION",
		Short: "Bulk-import data objects into a collection",
		Args:  cobra.ExactArgs(1), RunE: runVSCollImportData,
	}
	vsCollExportCmd = &cobra.Command{
		Use: "export-data-objects COLLECTION",
		Short: "Bulk-export data objects from a collection",
		Args:  cobra.ExactArgs(1), RunE: runVSCollExportData,
	}
)

func init() {
	all := []*cobra.Command{
		vsCollCreateCmd, vsCollDeleteCmd, vsCollDescribeCmd, vsCollListCmd,
		vsCollUpdateCmd, vsCollImportCmd, vsCollExportCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagVSCollLocation, "location", "", "Vector Search location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVSCollFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{vsCollCreateCmd, vsCollUpdateCmd, vsCollImportCmd, vsCollExportCmd} {
		c.Flags().StringVar(&flagVSCollConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	vsCollUpdateCmd.Flags().StringVar(&flagVSCollUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	vsCollListCmd.Flags().Int64Var(&flagVSCollPageSize, "page-size", 0, "Maximum results per page")

	vsCollCmd.AddCommand(all...)
	vectorSearchCmd.AddCommand(vsCollCmd)
}

func vsCollParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagVSCollLocation), nil
}

func vsCollName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := vsCollParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/collections/%s", parent, id), nil
}

func runVSCollCreate(cmd *cobra.Command, args []string) error {
	parent, err := vsCollParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagVSCollConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("collectionId", args[0])
	ctx := context.Background()
	var op map[string]any
	if err := vectorSearchRest.do(ctx, http.MethodPost, "/"+parent+"/collections", q, body, &op); err != nil {
		return fmt.Errorf("creating collection: %w", err)
	}
	fmt.Printf("Create request issued for collection [%s].\n", args[0])
	return emitFormatted(op, flagVSCollFormat)
}

func runVSCollDelete(cmd *cobra.Command, args []string) error {
	name, err := vsCollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := vectorSearchRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting collection: %w", err)
	}
	fmt.Printf("Delete request issued for collection [%s].\n", args[0])
	return emitFormatted(op, flagVSCollFormat)
}

func runVSCollDescribe(cmd *cobra.Command, args []string) error {
	name, err := vsCollName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := vectorSearchRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing collection: %w", err)
	}
	return emitFormatted(got, flagVSCollFormat)
}

func runVSCollList(cmd *cobra.Command, args []string) error {
	parent, err := vsCollParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := vectorSearchRest.paginate(ctx, "/"+parent+"/collections", nil, "collections", flagVSCollPageSize)
	if err != nil {
		return fmt.Errorf("listing collections: %w", err)
	}
	return emitFormatted(items, flagVSCollFormat)
}

func runVSCollUpdate(cmd *cobra.Command, args []string) error {
	name, err := vsCollName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagVSCollConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	mask := flagVSCollUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := context.Background()
	var op map[string]any
	if err := vectorSearchRest.do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating collection: %w", err)
	}
	fmt.Printf("Update request issued for collection [%s].\n", args[0])
	return emitFormatted(op, flagVSCollFormat)
}

func runVSCollImportData(cmd *cobra.Command, args []string) error {
	name, err := vsCollName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagVSCollConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := vectorSearchRest.do(ctx, http.MethodPost, "/"+name+":importDataObjects", nil, body, &op); err != nil {
		return fmt.Errorf("importing data objects: %w", err)
	}
	return emitFormatted(op, flagVSCollFormat)
}

func runVSCollExportData(cmd *cobra.Command, args []string) error {
	name, err := vsCollName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagVSCollConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := vectorSearchRest.do(ctx, http.MethodPost, "/"+name+":exportDataObjects", nil, body, &op); err != nil {
		return fmt.Errorf("exporting data objects: %w", err)
	}
	return emitFormatted(op, flagVSCollFormat)
}

// --- vector-search operations (#1433) ---

var vsOpCmd = &cobra.Command{Use: "operations", Short: "Manage Vector Search operations"}

var (
	flagVSOpLocation string
	flagVSOpFormat   string
	flagVSOpFilter   string
	flagVSOpPageSize int64
	flagVSOpTimeout  time.Duration
)

var (
	vsOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Vector Search operation",
		Args: cobra.ExactArgs(1), RunE: runVSOpCancel,
	}
	vsOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Vector Search operation record",
		Args: cobra.ExactArgs(1), RunE: runVSOpDelete,
	}
	vsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Vector Search operation",
		Args: cobra.ExactArgs(1), RunE: runVSOpDescribe,
	}
	vsOpListCmd = &cobra.Command{
		Use: "list", Short: "List Vector Search operations",
		Args: cobra.NoArgs, RunE: runVSOpList,
	}
	vsOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Vector Search operation to finish",
		Args: cobra.ExactArgs(1), RunE: runVSOpWait,
	}
)

func init() {
	all := []*cobra.Command{vsOpCancelCmd, vsOpDeleteCmd, vsOpDescribeCmd, vsOpListCmd, vsOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagVSOpLocation, "location", "", "Vector Search location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVSOpFormat, "format", "", "Output format")
	}
	vsOpListCmd.Flags().StringVar(&flagVSOpFilter, "filter", "", "Server-side filter expression")
	vsOpListCmd.Flags().Int64Var(&flagVSOpPageSize, "page-size", 0, "Maximum results per page")
	vsOpWaitCmd.Flags().DurationVar(&flagVSOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to finish")

	vsOpCmd.AddCommand(all...)
	vectorSearchCmd.AddCommand(vsOpCmd)
}

func vsOpQualifiedName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, flagVSOpLocation, id), nil
}

func runVSOpCancel(cmd *cobra.Command, args []string) error {
	name, err := vsOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := vectorSearchRest.do(ctx, http.MethodPost, "/"+name+":cancel", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancel request issued for operation %s.\n", args[0])
	return nil
}

func runVSOpDelete(cmd *cobra.Command, args []string) error {
	name, err := vsOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := vectorSearchRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation %s.\n", args[0])
	return nil
}

func runVSOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := vsOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := vectorSearchRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagVSOpFormat)
}

func runVSOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	base := url.Values{}
	if flagVSOpFilter != "" {
		base.Set("filter", flagVSOpFilter)
	}
	ctx := context.Background()
	parent := fmt.Sprintf("projects/%s/locations/%s/operations", project, flagVSOpLocation)
	items, err := vectorSearchRest.paginate(ctx, "/"+parent, base, "operations", flagVSOpPageSize)
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagVSOpFormat != "" {
		return emitFormatted(items, flagVSOpFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range items {
		name, _ := o["name"].(string)
		done, _ := o["done"].(bool)
		fmt.Printf("%-40s %v\n", path.Base(name), done)
	}
	return nil
}

func runVSOpWait(cmd *cobra.Command, args []string) error {
	name, err := vsOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	op, err := vectorSearchRest.waitOperation(ctx, name, flagVSOpTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("Operation %s completed.\n", args[0])
	return emitFormatted(op, flagVSOpFormat)
}
