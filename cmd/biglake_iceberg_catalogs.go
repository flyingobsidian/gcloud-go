package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud biglake iceberg catalogs (#981) ---
//
// The BigLake Iceberg REST catalog API uses a bespoke URL layout and, unlike
// most Google REST APIs, its request/response bodies use hyphen-separated
// keys (e.g. `catalog-type`, `default-location`, `federated-catalog-options`)
// rather than camelCase.

var biglakeIcebergCatalogsCmd = &cobra.Command{Use: "catalogs", Short: "Manage BigLake Iceberg REST catalogs"}

var (
	flagBLICLocation   string
	flagBLICFormat     string
	flagBLICConfigFile string
	flagBLICUpdateMask string
)

var (
	blicCreateCmd = &cobra.Command{
		Use: "create CATALOG", Short: "Create a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: runBLICCreate,
	}
	blicDeleteCmd = &cobra.Command{
		Use: "delete CATALOG", Short: "Delete a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: runBLICDelete,
	}
	blicDescribeCmd = &cobra.Command{
		Use: "describe CATALOG", Short: "Describe a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: runBLICDescribe,
	}
	blicListCmd = &cobra.Command{
		Use: "list", Short: "List BigLake Iceberg REST catalogs",
		Args: cobra.NoArgs, RunE: runBLICList,
	}
	blicUpdateCmd = &cobra.Command{
		Use: "update CATALOG", Short: "Update a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: runBLICUpdate,
	}
	blicFailoverCmd = &cobra.Command{
		Use: "failover CATALOG", Short: "Fail over a BigLake Iceberg REST catalog",
		Args: cobra.ExactArgs(1), RunE: runBLICFailover,
	}
	blicGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CATALOG", Short: "Get the IAM policy for a BigLake Iceberg catalog",
		Args: cobra.ExactArgs(1), RunE: runBLICGetIam,
	}
	blicSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CATALOG POLICY_FILE", Short: "Set the IAM policy for a BigLake Iceberg catalog",
		Args: cobra.ExactArgs(2), RunE: runBLICSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		blicCreateCmd, blicDeleteCmd, blicDescribeCmd, blicListCmd, blicUpdateCmd, blicFailoverCmd,
		blicGetIamCmd, blicSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBLICLocation, "location", "", "BigLake location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagBLICFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{blicCreateCmd, blicUpdateCmd, blicFailoverCmd} {
		c.Flags().StringVar(&flagBLICConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the catalog body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	blicUpdateCmd.Flags().StringVar(&flagBLICUpdateMask, "update-mask", "",
		"Comma-separated list of hyphen-cased fields to update (defaults to every populated field)")

	biglakeIcebergCatalogsCmd.AddCommand(all...)
	biglakeIcebergCmd.AddCommand(biglakeIcebergCatalogsCmd)
}

// blicCatalogsPath returns the extensions-scoped path
// `/projects/PROJ/locations/LOC/catalogs`.
func blicCatalogsPath() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/projects/%s/locations/%s/catalogs", project, flagBLICLocation), nil
}

// blicCatalogPath returns the extensions-scoped path
// `/projects/PROJ/locations/LOC/catalogs/CATALOG`. Accepts a fully qualified
// `projects/...` id.
func blicCatalogPath(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return "/" + id, nil
	}
	cats, err := blicCatalogsPath()
	if err != nil {
		return "", err
	}
	return cats + "/" + id, nil
}

// blicStandardCatalogName returns the biglake/v1 resource name
// `projects/PROJ/locations/LOC/catalogs/CATALOG` used by IAM RPCs.
func blicStandardCatalogName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/catalogs/%s", project, flagBLICLocation, id), nil
}

func runBLICCreate(cmd *cobra.Command, args []string) error {
	path, err := blicCatalogsPath()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagBLICConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("iceberg-catalog-id", args[0])
	ctx := context.Background()
	var got map[string]any
	if err := biglakeIcebergRest.do(ctx, http.MethodPost, path, q, body, &got); err != nil {
		return fmt.Errorf("creating catalog: %w", err)
	}
	fmt.Printf("Created BigLake Iceberg catalog [%s].\n", args[0])
	return emitFormatted(got, flagBLICFormat)
}

func runBLICDelete(cmd *cobra.Command, args []string) error {
	path, err := blicCatalogPath(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := biglakeIcebergRest.do(ctx, http.MethodDelete, path, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting catalog: %w", err)
	}
	fmt.Printf("Deleted BigLake Iceberg catalog [%s].\n", args[0])
	return nil
}

func runBLICDescribe(cmd *cobra.Command, args []string) error {
	path, err := blicCatalogPath(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := biglakeIcebergRest.do(ctx, http.MethodGet, path, nil, nil, &got); err != nil {
		return fmt.Errorf("describing catalog: %w", err)
	}
	return emitFormatted(got, flagBLICFormat)
}

func runBLICList(cmd *cobra.Command, args []string) error {
	path, err := blicCatalogsPath()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var raw map[string]any
	if err := biglakeIcebergRest.do(ctx, http.MethodGet, path, nil, nil, &raw); err != nil {
		return fmt.Errorf("listing catalogs: %w", err)
	}
	// The Iceberg REST list may return either {"catalogs": [...]} or a raw
	// object; fall back to printing the whole response when it isn't a slice.
	if arr, ok := raw["catalogs"].([]any); ok {
		return emitFormatted(arr, flagBLICFormat)
	}
	return emitFormatted(raw, flagBLICFormat)
}

func runBLICUpdate(cmd *cobra.Command, args []string) error {
	path, err := blicCatalogPath(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagBLICConfigFile, &body); err != nil {
		return err
	}
	mask := flagBLICUpdateMask
	if mask == "" {
		// The body uses hyphenated keys; use those top-level keys directly so
		// the resulting mask matches the field names the server expects.
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("update-mask", mask)
	}
	ctx := context.Background()
	var got map[string]any
	if err := biglakeIcebergRest.do(ctx, http.MethodPatch, path, q, body, &got); err != nil {
		return fmt.Errorf("updating catalog: %w", err)
	}
	fmt.Printf("Updated BigLake Iceberg catalog [%s].\n", args[0])
	return emitFormatted(got, flagBLICFormat)
}

func runBLICFailover(cmd *cobra.Command, args []string) error {
	path, err := blicCatalogPath(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if flagBLICConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagBLICConfigFile, &body); err != nil {
			return err
		}
	}
	ctx := context.Background()
	var got map[string]any
	if err := biglakeIcebergRest.do(ctx, http.MethodPost, path+":failover", nil, body, &got); err != nil {
		return fmt.Errorf("failing over catalog: %w", err)
	}
	fmt.Printf("Failover request issued for BigLake Iceberg catalog [%s].\n", args[0])
	return emitFormatted(got, flagBLICFormat)
}

func runBLICGetIam(cmd *cobra.Command, args []string) error {
	name, err := blicStandardCatalogName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := biglakeStandardRest.do(ctx, http.MethodGet, "/"+name+":getIamPolicy", nil, nil, &got); err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(got, flagBLICFormat)
}

func runBLICSetIam(cmd *cobra.Command, args []string) error {
	name, err := blicStandardCatalogName(args[0])
	if err != nil {
		return err
	}
	policy := map[string]any{}
	if err := loadYAMLOrJSONInto(args[1], &policy); err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := biglakeStandardRest.do(ctx, http.MethodPost, "/"+name+":setIamPolicy", nil, map[string]any{"policy": policy}, &got); err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, flagBLICFormat)
}
