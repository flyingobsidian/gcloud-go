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

// --- gcloud edge-cache origins (#1059) ---

var edgeCacheOriginsCmd = &cobra.Command{Use: "origins", Short: "Manage EdgeCacheOrigin resources"}

var (
	flagEdgeCacheOriginsLocation    string
	flagEdgeCacheOriginsFormat      string
	flagEdgeCacheOriginsConfigFile  string
	flagEdgeCacheOriginsUpdateMask  string
	flagEdgeCacheOriginsDestination string
	flagEdgeCacheOriginsSource      string
	flagEdgeCacheOriginsPageSize    int64
)

var (
	edgeCacheOriginsCreateCmd = &cobra.Command{
		Use: "create ORIGIN", Short: "Create an EdgeCacheOrigin",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheOriginsCreate,
	}
	edgeCacheOriginsDeleteCmd = &cobra.Command{
		Use: "delete ORIGIN", Short: "Delete an EdgeCacheOrigin",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheOriginsDelete,
	}
	edgeCacheOriginsDescribeCmd = &cobra.Command{
		Use: "describe ORIGIN", Short: "Describe an EdgeCacheOrigin",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheOriginsDescribe,
	}
	edgeCacheOriginsExportCmd = &cobra.Command{
		Use: "export ORIGIN", Short: "Export an EdgeCacheOrigin (as YAML or JSON) to a file or stdout",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheOriginsExport,
	}
	edgeCacheOriginsImportCmd = &cobra.Command{
		Use: "import ORIGIN", Short: "Import (create or update) an EdgeCacheOrigin from a file",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheOriginsImport,
	}
	edgeCacheOriginsListCmd = &cobra.Command{
		Use: "list", Short: "List EdgeCacheOrigin resources",
		Args: cobra.NoArgs, RunE: runEdgeCacheOriginsList,
	}
	edgeCacheOriginsUpdateCmd = &cobra.Command{
		Use: "update ORIGIN", Short: "Update an EdgeCacheOrigin",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheOriginsUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		edgeCacheOriginsCreateCmd, edgeCacheOriginsDeleteCmd, edgeCacheOriginsDescribeCmd,
		edgeCacheOriginsExportCmd, edgeCacheOriginsImportCmd, edgeCacheOriginsListCmd,
		edgeCacheOriginsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagEdgeCacheOriginsLocation, "location", "", "Edge Cache location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagEdgeCacheOriginsFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{edgeCacheOriginsCreateCmd, edgeCacheOriginsUpdateCmd} {
		c.Flags().StringVar(&flagEdgeCacheOriginsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the EdgeCacheOrigin body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	edgeCacheOriginsUpdateCmd.Flags().StringVar(&flagEdgeCacheOriginsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated top-level key)")
	edgeCacheOriginsExportCmd.Flags().StringVar(&flagEdgeCacheOriginsDestination, "destination", "",
		"File to write the exported resource to (defaults to stdout)")
	edgeCacheOriginsImportCmd.Flags().StringVar(&flagEdgeCacheOriginsSource, "source", "",
		"Path to a YAML/JSON file with the EdgeCacheOrigin body (required)")
	_ = edgeCacheOriginsImportCmd.MarkFlagRequired("source")
	edgeCacheOriginsListCmd.Flags().Int64Var(&flagEdgeCacheOriginsPageSize, "page-size", 0, "Maximum results per page")

	edgeCacheOriginsCmd.AddCommand(all...)
	edgeCacheCmd.AddCommand(edgeCacheOriginsCmd)
}

func edgeCacheOriginsParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagEdgeCacheOriginsLocation), nil
}

func edgeCacheOriginsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := edgeCacheOriginsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/edgeCacheOrigins/%s", parent, id), nil
}

func runEdgeCacheOriginsCreate(cmd *cobra.Command, args []string) error {
	parent, err := edgeCacheOriginsParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheOriginsConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("edgeCacheOriginId", args[0])
	ctx := context.Background()
	var op map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodPost, "/"+parent+"/edgeCacheOrigins", q, body, &op); err != nil {
		return fmt.Errorf("creating edge-cache origin: %w", err)
	}
	fmt.Printf("Create request issued for edge-cache origin [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheOriginsFormat)
}

func runEdgeCacheOriginsDelete(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting edge-cache origin: %w", err)
	}
	fmt.Printf("Delete request issued for edge-cache origin [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheOriginsFormat)
}

func runEdgeCacheOriginsDescribe(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing edge-cache origin: %w", err)
	}
	return emitFormatted(got, flagEdgeCacheOriginsFormat)
}

func runEdgeCacheOriginsExport(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("exporting edge-cache origin: %w", err)
	}
	if flagEdgeCacheOriginsDestination == "" {
		return emitFormatted(got, flagEdgeCacheOriginsFormat)
	}
	f, err := os.Create(flagEdgeCacheOriginsDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	return emitFormattedTo(f, got, flagEdgeCacheOriginsFormat)
}

func runEdgeCacheOriginsImport(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheOriginsSource, &body); err != nil {
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
		fmt.Printf("Import (update) request issued for edge-cache origin [%s].\n", args[0])
		return emitFormatted(op, flagEdgeCacheOriginsFormat)
	}
	var httpErr *restError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusNotFound {
		return fmt.Errorf("importing edge-cache origin: %w", err)
	}
	parent, perr := edgeCacheOriginsParent()
	if perr != nil {
		return perr
	}
	createQ := url.Values{}
	createQ.Set("edgeCacheOriginId", args[0])
	if err := edgeCacheRest.do(ctx, http.MethodPost, "/"+parent+"/edgeCacheOrigins", createQ, body, &op); err != nil {
		return fmt.Errorf("importing edge-cache origin: %w", err)
	}
	fmt.Printf("Import (create) request issued for edge-cache origin [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheOriginsFormat)
}

func runEdgeCacheOriginsList(cmd *cobra.Command, args []string) error {
	parent, err := edgeCacheOriginsParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := edgeCacheRest.paginate(ctx, "/"+parent+"/edgeCacheOrigins", nil, "edgeCacheOrigins", flagEdgeCacheOriginsPageSize)
	if err != nil {
		return fmt.Errorf("listing edge-cache origins: %w", err)
	}
	return emitFormatted(items, flagEdgeCacheOriginsFormat)
}

func runEdgeCacheOriginsUpdate(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheOriginsConfigFile, &body); err != nil {
		return err
	}
	mask := flagEdgeCacheOriginsUpdateMask
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
		return fmt.Errorf("updating edge-cache origin: %w", err)
	}
	fmt.Printf("Update request issued for edge-cache origin [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheOriginsFormat)
}
