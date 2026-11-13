package cmd

import (
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

func (s *biglakeCmds) initBiglakeIcebergCatalogs() {
	all := []*cobra.Command{
		s.blicCreateCmd, s.blicDeleteCmd, s.blicDescribeCmd, s.blicListCmd, s.blicUpdateCmd, s.blicFailoverCmd,
		s.blicGetIamCmd, s.blicSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBLICLocation, "location", "", "BigLake location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagBLICFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.blicCreateCmd, s.blicUpdateCmd, s.blicFailoverCmd} {
		c.Flags().StringVar(&s.flagBLICConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the catalog body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.blicUpdateCmd.Flags().StringVar(&s.flagBLICUpdateMask, "update-mask", "",
		"Comma-separated list of hyphen-cased fields to update (defaults to every populated field)")

	s.biglakeIcebergCatalogsCmd.AddCommand(all...)
	s.biglakeIcebergCmd.AddCommand(s.biglakeIcebergCatalogsCmd)
}

// blicCatalogsPath returns the extensions-scoped path
// `/projects/PROJ/locations/LOC/catalogs`.
func (s *biglakeCmds) blicCatalogsPath() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/projects/%s/locations/%s/catalogs", project, s.flagBLICLocation), nil
}

// blicCatalogPath returns the extensions-scoped path
// `/projects/PROJ/locations/LOC/catalogs/CATALOG`. Accepts a fully qualified
// `projects/...` id.
func (s *biglakeCmds) blicCatalogPath(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return "/" + id, nil
	}
	cats, err := s.blicCatalogsPath()
	if err != nil {
		return "", err
	}
	return cats + "/" + id, nil
}

// blicStandardCatalogName returns the biglake/v1 resource name
// `projects/PROJ/locations/LOC/catalogs/CATALOG` used by IAM RPCs.
func (s *biglakeCmds) blicStandardCatalogName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/catalogs/%s", project, s.flagBLICLocation, id), nil
}

func (s *biglakeCmds) runBLICCreate(cmd *cobra.Command, args []string) error {
	path, err := s.blicCatalogsPath()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagBLICConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("iceberg-catalog-id", args[0])
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(biglakeIcebergRest()).do(ctx, http.MethodPost, path, q, body, &got); err != nil {
		return fmt.Errorf("creating catalog: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created BigLake Iceberg catalog [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBLICFormat)
}

func (s *biglakeCmds) runBLICDelete(cmd *cobra.Command, args []string) error {
	path, err := s.blicCatalogPath(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(biglakeIcebergRest()).do(ctx, http.MethodDelete, path, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting catalog: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted BigLake Iceberg catalog [%s].\n", args[0])
	return nil
}

func (s *biglakeCmds) runBLICDescribe(cmd *cobra.Command, args []string) error {
	path, err := s.blicCatalogPath(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(biglakeIcebergRest()).do(ctx, http.MethodGet, path, nil, nil, &got); err != nil {
		return fmt.Errorf("describing catalog: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBLICFormat)
}

func (s *biglakeCmds) runBLICList(cmd *cobra.Command, args []string) error {
	path, err := s.blicCatalogsPath()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var raw map[string]any
	if err := s.inv.rest(biglakeIcebergRest()).do(ctx, http.MethodGet, path, nil, nil, &raw); err != nil {
		return fmt.Errorf("listing catalogs: %w", err)
	}
	// The Iceberg REST list may return either {"catalogs": [...]} or a raw
	// object; fall back to printing the whole response when it isn't a slice.
	if arr, ok := raw["catalogs"].([]any); ok {
		return emitFormattedTo(s.inv.stdout, arr, s.flagBLICFormat)
	}
	return emitFormattedTo(s.inv.stdout, raw, s.flagBLICFormat)
}

func (s *biglakeCmds) runBLICUpdate(cmd *cobra.Command, args []string) error {
	path, err := s.blicCatalogPath(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagBLICConfigFile, &body); err != nil {
		return err
	}
	mask := s.flagBLICUpdateMask
	if mask == "" {
		// The body uses hyphenated keys; use those top-level keys directly so
		// the resulting mask matches the field names the server expects.
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("update-mask", mask)
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(biglakeIcebergRest()).do(ctx, http.MethodPatch, path, q, body, &got); err != nil {
		return fmt.Errorf("updating catalog: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated BigLake Iceberg catalog [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBLICFormat)
}

func (s *biglakeCmds) runBLICFailover(cmd *cobra.Command, args []string) error {
	path, err := s.blicCatalogPath(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if s.flagBLICConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagBLICConfigFile, &body); err != nil {
			return err
		}
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(biglakeIcebergRest()).do(ctx, http.MethodPost, path+":failover", nil, body, &got); err != nil {
		return fmt.Errorf("failing over catalog: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Failover request issued for BigLake Iceberg catalog [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBLICFormat)
}

func (s *biglakeCmds) runBLICGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.blicStandardCatalogName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(biglakeStandardRest()).do(ctx, http.MethodGet, "/"+name+":getIamPolicy", nil, nil, &got); err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBLICFormat)
}

func (s *biglakeCmds) runBLICSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.blicStandardCatalogName(args[0])
	if err != nil {
		return err
	}
	policy := map[string]any{}
	if err := loadYAMLOrJSONInto(args[1], &policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(biglakeStandardRest()).do(ctx, http.MethodPost, "/"+name+":setIamPolicy", nil, map[string]any{"policy": policy}, &got); err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBLICFormat)
}
