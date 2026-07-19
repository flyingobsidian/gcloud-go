package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud edge-cache services (#1060) ---

var edgeCacheServicesCmd = &cobra.Command{Use: "services", Short: "Manage EdgeCacheService resources"}

var (
	flagEdgeCacheServicesLocation     string
	flagEdgeCacheServicesFormat       string
	flagEdgeCacheServicesConfigFile   string
	flagEdgeCacheServicesUpdateMask   string
	flagEdgeCacheServicesDestination  string
	flagEdgeCacheServicesSource       string
	flagEdgeCacheServicesInvalidateFn string
	flagEdgeCacheServicesPageSize     int64
)

var (
	edgeCacheServicesDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete an EdgeCacheService",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheServicesDelete,
	}
	edgeCacheServicesDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe an EdgeCacheService",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheServicesDescribe,
	}
	edgeCacheServicesExportCmd = &cobra.Command{
		Use: "export SERVICE", Short: "Export an EdgeCacheService (as YAML or JSON) to a file or stdout",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheServicesExport,
	}
	edgeCacheServicesImportCmd = &cobra.Command{
		Use: "import SERVICE", Short: "Import (create or update) an EdgeCacheService from a file",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheServicesImport,
	}
	edgeCacheServicesInvalidateCmd = &cobra.Command{
		Use: "invalidate-cache SERVICE", Short: "Invalidate cache for an EdgeCacheService",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheServicesInvalidate,
	}
	edgeCacheServicesListCmd = &cobra.Command{
		Use: "list", Short: "List EdgeCacheService resources",
		Args: cobra.NoArgs, RunE: runEdgeCacheServicesList,
	}
	edgeCacheServicesUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update an EdgeCacheService",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheServicesUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		edgeCacheServicesDeleteCmd, edgeCacheServicesDescribeCmd,
		edgeCacheServicesExportCmd, edgeCacheServicesImportCmd,
		edgeCacheServicesInvalidateCmd, edgeCacheServicesListCmd,
		edgeCacheServicesUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagEdgeCacheServicesLocation, "location", "", "Edge Cache location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagEdgeCacheServicesFormat, "format", "", "Output format")
	}
	edgeCacheServicesUpdateCmd.Flags().StringVar(&flagEdgeCacheServicesConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the EdgeCacheService body (required)")
	_ = edgeCacheServicesUpdateCmd.MarkFlagRequired("config-file")
	edgeCacheServicesUpdateCmd.Flags().StringVar(&flagEdgeCacheServicesUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated top-level key)")
	edgeCacheServicesExportCmd.Flags().StringVar(&flagEdgeCacheServicesDestination, "destination", "",
		"File to write the exported resource to (defaults to stdout)")
	edgeCacheServicesImportCmd.Flags().StringVar(&flagEdgeCacheServicesSource, "source", "",
		"Path to a YAML/JSON file with the EdgeCacheService body (required)")
	_ = edgeCacheServicesImportCmd.MarkFlagRequired("source")
	edgeCacheServicesInvalidateCmd.Flags().StringVar(&flagEdgeCacheServicesInvalidateFn, "config-file", "",
		"Path to a YAML/JSON file with the InvalidateCache request body (required)")
	_ = edgeCacheServicesInvalidateCmd.MarkFlagRequired("config-file")
	edgeCacheServicesListCmd.Flags().Int64Var(&flagEdgeCacheServicesPageSize, "page-size", 0, "Maximum results per page")

	edgeCacheServicesCmd.AddCommand(all...)
	edgeCacheCmd.AddCommand(edgeCacheServicesCmd)
}

func edgeCacheServicesParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagEdgeCacheServicesLocation), nil
}

func edgeCacheServicesName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := edgeCacheServicesParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/edgeCacheServices/%s", parent, id), nil
}

func runEdgeCacheServicesDelete(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting edge-cache service: %w", err)
	}
	fmt.Printf("Delete request issued for edge-cache service [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheServicesFormat)
}

func runEdgeCacheServicesDescribe(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing edge-cache service: %w", err)
	}
	return emitFormatted(got, flagEdgeCacheServicesFormat)
}

func runEdgeCacheServicesExport(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("exporting edge-cache service: %w", err)
	}
	if flagEdgeCacheServicesDestination == "" {
		return emitFormatted(got, flagEdgeCacheServicesFormat)
	}
	f, err := os.Create(flagEdgeCacheServicesDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	return emitFormattedTo(f, got, flagEdgeCacheServicesFormat)
}

func runEdgeCacheServicesImport(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheServicesSource, &body); err != nil {
		return err
	}
	ctx := context.Background()
	// Attempt update first; if the resource does not exist, fall through to create.
	var op map[string]any
	q := url.Values{}
	if mask := joinMask(dcTopLevelKeys(body)); mask != "" {
		q.Set("updateMask", mask)
	}
	err = edgeCacheRest.do(ctx, http.MethodPatch, "/"+name, q, body, &op)
	if err == nil {
		fmt.Printf("Import (update) request issued for edge-cache service [%s].\n", args[0])
		return emitFormatted(op, flagEdgeCacheServicesFormat)
	}
	var httpErr *restError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusNotFound {
		return fmt.Errorf("importing edge-cache service: %w", err)
	}
	parent, perr := edgeCacheServicesParent()
	if perr != nil {
		return perr
	}
	createQ := url.Values{}
	createQ.Set("edgeCacheServiceId", args[0])
	if err := edgeCacheRest.do(ctx, http.MethodPost, "/"+parent+"/edgeCacheServices", createQ, body, &op); err != nil {
		return fmt.Errorf("importing edge-cache service: %w", err)
	}
	fmt.Printf("Import (create) request issued for edge-cache service [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheServicesFormat)
}

func runEdgeCacheServicesInvalidate(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheServicesInvalidateFn, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodPost, "/"+name+":invalidateCache", nil, body, &op); err != nil {
		return fmt.Errorf("invalidating edge-cache service cache: %w", err)
	}
	fmt.Printf("Invalidate-cache request issued for edge-cache service [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheServicesFormat)
}

func runEdgeCacheServicesList(cmd *cobra.Command, args []string) error {
	parent, err := edgeCacheServicesParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := edgeCacheRest.paginate(ctx, "/"+parent+"/edgeCacheServices", nil, "edgeCacheServices", flagEdgeCacheServicesPageSize)
	if err != nil {
		return fmt.Errorf("listing edge-cache services: %w", err)
	}
	return emitFormatted(items, flagEdgeCacheServicesFormat)
}

func runEdgeCacheServicesUpdate(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheServicesConfigFile, &body); err != nil {
		return err
	}
	mask := flagEdgeCacheServicesUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := context.Background()
	var op map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating edge-cache service: %w", err)
	}
	fmt.Printf("Update request issued for edge-cache service [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheServicesFormat)
}
