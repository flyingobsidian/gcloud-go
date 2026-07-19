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

// --- gcloud edge-cache keysets (#1057) ---

var edgeCacheKeysetsCmd = &cobra.Command{Use: "keysets", Short: "Manage EdgeCacheKeyset resources"}

var (
	flagEdgeCacheKeysetsLocation    string
	flagEdgeCacheKeysetsFormat      string
	flagEdgeCacheKeysetsConfigFile  string
	flagEdgeCacheKeysetsUpdateMask  string
	flagEdgeCacheKeysetsDestination string
	flagEdgeCacheKeysetsSource      string
	flagEdgeCacheKeysetsPageSize    int64
)

var (
	edgeCacheKeysetsCreateCmd = &cobra.Command{
		Use: "create KEYSET", Short: "Create an EdgeCacheKeyset",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheKeysetsCreate,
	}
	edgeCacheKeysetsDeleteCmd = &cobra.Command{
		Use: "delete KEYSET", Short: "Delete an EdgeCacheKeyset",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheKeysetsDelete,
	}
	edgeCacheKeysetsDescribeCmd = &cobra.Command{
		Use: "describe KEYSET", Short: "Describe an EdgeCacheKeyset",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheKeysetsDescribe,
	}
	edgeCacheKeysetsExportCmd = &cobra.Command{
		Use: "export KEYSET", Short: "Export an EdgeCacheKeyset (as YAML or JSON) to a file or stdout",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheKeysetsExport,
	}
	edgeCacheKeysetsImportCmd = &cobra.Command{
		Use: "import KEYSET", Short: "Import (create or update) an EdgeCacheKeyset from a file",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheKeysetsImport,
	}
	edgeCacheKeysetsListCmd = &cobra.Command{
		Use: "list", Short: "List EdgeCacheKeyset resources",
		Args: cobra.NoArgs, RunE: runEdgeCacheKeysetsList,
	}
	edgeCacheKeysetsUpdateCmd = &cobra.Command{
		Use: "update KEYSET", Short: "Update an EdgeCacheKeyset",
		Args: cobra.ExactArgs(1), RunE: runEdgeCacheKeysetsUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		edgeCacheKeysetsCreateCmd, edgeCacheKeysetsDeleteCmd, edgeCacheKeysetsDescribeCmd,
		edgeCacheKeysetsExportCmd, edgeCacheKeysetsImportCmd, edgeCacheKeysetsListCmd,
		edgeCacheKeysetsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagEdgeCacheKeysetsLocation, "location", "", "Edge Cache location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagEdgeCacheKeysetsFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{edgeCacheKeysetsCreateCmd, edgeCacheKeysetsUpdateCmd} {
		c.Flags().StringVar(&flagEdgeCacheKeysetsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the EdgeCacheKeyset body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	edgeCacheKeysetsUpdateCmd.Flags().StringVar(&flagEdgeCacheKeysetsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated top-level key)")
	edgeCacheKeysetsExportCmd.Flags().StringVar(&flagEdgeCacheKeysetsDestination, "destination", "",
		"File to write the exported resource to (defaults to stdout)")
	edgeCacheKeysetsImportCmd.Flags().StringVar(&flagEdgeCacheKeysetsSource, "source", "",
		"Path to a YAML/JSON file with the EdgeCacheKeyset body (required)")
	_ = edgeCacheKeysetsImportCmd.MarkFlagRequired("source")
	edgeCacheKeysetsListCmd.Flags().Int64Var(&flagEdgeCacheKeysetsPageSize, "page-size", 0, "Maximum results per page")

	edgeCacheKeysetsCmd.AddCommand(all...)
	edgeCacheCmd.AddCommand(edgeCacheKeysetsCmd)
}

func edgeCacheKeysetsParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagEdgeCacheKeysetsLocation), nil
}

func edgeCacheKeysetsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := edgeCacheKeysetsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/edgeCacheKeysets/%s", parent, id), nil
}

func runEdgeCacheKeysetsCreate(cmd *cobra.Command, args []string) error {
	parent, err := edgeCacheKeysetsParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheKeysetsConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("edgeCacheKeysetId", args[0])
	ctx := context.Background()
	var op map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodPost, "/"+parent+"/edgeCacheKeysets", q, body, &op); err != nil {
		return fmt.Errorf("creating edge-cache keyset: %w", err)
	}
	fmt.Printf("Create request issued for edge-cache keyset [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheKeysetsFormat)
}

func runEdgeCacheKeysetsDelete(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting edge-cache keyset: %w", err)
	}
	fmt.Printf("Delete request issued for edge-cache keyset [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheKeysetsFormat)
}

func runEdgeCacheKeysetsDescribe(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing edge-cache keyset: %w", err)
	}
	return emitFormatted(got, flagEdgeCacheKeysetsFormat)
}

func runEdgeCacheKeysetsExport(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := edgeCacheRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("exporting edge-cache keyset: %w", err)
	}
	if flagEdgeCacheKeysetsDestination == "" {
		return emitFormatted(got, flagEdgeCacheKeysetsFormat)
	}
	f, err := os.Create(flagEdgeCacheKeysetsDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	return emitFormattedTo(f, got, flagEdgeCacheKeysetsFormat)
}

func runEdgeCacheKeysetsImport(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheKeysetsSource, &body); err != nil {
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
		fmt.Printf("Import (update) request issued for edge-cache keyset [%s].\n", args[0])
		return emitFormatted(op, flagEdgeCacheKeysetsFormat)
	}
	var httpErr *restError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusNotFound {
		return fmt.Errorf("importing edge-cache keyset: %w", err)
	}
	parent, perr := edgeCacheKeysetsParent()
	if perr != nil {
		return perr
	}
	createQ := url.Values{}
	createQ.Set("edgeCacheKeysetId", args[0])
	if err := edgeCacheRest.do(ctx, http.MethodPost, "/"+parent+"/edgeCacheKeysets", createQ, body, &op); err != nil {
		return fmt.Errorf("importing edge-cache keyset: %w", err)
	}
	fmt.Printf("Import (create) request issued for edge-cache keyset [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheKeysetsFormat)
}

func runEdgeCacheKeysetsList(cmd *cobra.Command, args []string) error {
	parent, err := edgeCacheKeysetsParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := edgeCacheRest.paginate(ctx, "/"+parent+"/edgeCacheKeysets", nil, "edgeCacheKeysets", flagEdgeCacheKeysetsPageSize)
	if err != nil {
		return fmt.Errorf("listing edge-cache keysets: %w", err)
	}
	return emitFormatted(items, flagEdgeCacheKeysetsFormat)
}

func runEdgeCacheKeysetsUpdate(cmd *cobra.Command, args []string) error {
	name, err := edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagEdgeCacheKeysetsConfigFile, &body); err != nil {
		return err
	}
	mask := flagEdgeCacheKeysetsUpdateMask
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
		return fmt.Errorf("updating edge-cache keyset: %w", err)
	}
	fmt.Printf("Update request issued for edge-cache keyset [%s].\n", args[0])
	return emitFormatted(op, flagEdgeCacheKeysetsFormat)
}
