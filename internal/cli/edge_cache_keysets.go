package cli

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud edge-cache keysets (#1057) ---

func (s *edgeCacheCmds) initEdgeCacheKeysets() {
	all := []*cobra.Command{
		s.edgeCacheKeysetsCreateCmd, s.edgeCacheKeysetsDeleteCmd, s.edgeCacheKeysetsDescribeCmd,
		s.edgeCacheKeysetsExportCmd, s.edgeCacheKeysetsImportCmd, s.edgeCacheKeysetsListCmd,
		s.edgeCacheKeysetsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagEdgeCacheKeysetsLocation, "location", "", "Edge Cache location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagEdgeCacheKeysetsFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.edgeCacheKeysetsCreateCmd, s.edgeCacheKeysetsUpdateCmd} {
		c.Flags().StringVar(&s.flagEdgeCacheKeysetsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the EdgeCacheKeyset body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.edgeCacheKeysetsUpdateCmd.Flags().StringVar(&s.flagEdgeCacheKeysetsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated top-level key)")
	s.edgeCacheKeysetsExportCmd.Flags().StringVar(&s.flagEdgeCacheKeysetsDestination, "destination", "",
		"File to write the exported resource to (defaults to stdout)")
	s.edgeCacheKeysetsImportCmd.Flags().StringVar(&s.flagEdgeCacheKeysetsSource, "source", "",
		"Path to a YAML/JSON file with the EdgeCacheKeyset body (required)")
	_ = s.edgeCacheKeysetsImportCmd.MarkFlagRequired("source")
	s.edgeCacheKeysetsListCmd.Flags().Int64Var(&s.flagEdgeCacheKeysetsPageSize, "page-size", 0, "Maximum results per page")

	s.edgeCacheKeysetsCmd.AddCommand(all...)
	s.edgeCacheCmd.AddCommand(s.edgeCacheKeysetsCmd)
}

func (s *edgeCacheCmds) edgeCacheKeysetsParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagEdgeCacheKeysetsLocation), nil
}

func (s *edgeCacheCmds) edgeCacheKeysetsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.edgeCacheKeysetsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/edgeCacheKeysets/%s", parent, id), nil
}

func (s *edgeCacheCmds) runEdgeCacheKeysetsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.edgeCacheKeysetsParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheKeysetsConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("edgeCacheKeysetId", args[0])
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodPost, "/"+parent+"/edgeCacheKeysets", q, body, &op); err != nil {
		return fmt.Errorf("creating edge-cache keyset: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for edge-cache keyset [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheKeysetsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheKeysetsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting edge-cache keyset: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for edge-cache keyset [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheKeysetsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheKeysetsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing edge-cache keyset: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagEdgeCacheKeysetsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheKeysetsExport(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("exporting edge-cache keyset: %w", err)
	}
	if s.flagEdgeCacheKeysetsDestination == "" {
		return emitFormattedTo(s.inv.stdout, got, s.flagEdgeCacheKeysetsFormat)
	}
	f, err := os.Create(s.flagEdgeCacheKeysetsDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	return emitFormattedTo(f, got, s.flagEdgeCacheKeysetsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheKeysetsImport(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheKeysetsSource, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	// Attempt update first; if the resource does not exist, fall through to create.
	var op map[string]any
	q := url.Values{}
	if mask := joinMask(dcTopLevelKeys(body)); mask != "" {
		q.Set("updateMask", mask)
	}
	err = s.inv.rest(edgeCacheRest()).do(ctx, http.MethodPatch, "/"+name, q, body, &op)
	if err == nil {
		fmt.Fprintf(s.inv.stdout, "Import (update) request issued for edge-cache keyset [%s].\n", args[0])
		return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheKeysetsFormat)
	}
	var httpErr *restError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusNotFound {
		return fmt.Errorf("importing edge-cache keyset: %w", err)
	}
	parent, perr := s.edgeCacheKeysetsParent()
	if perr != nil {
		return perr
	}
	createQ := url.Values{}
	createQ.Set("edgeCacheKeysetId", args[0])
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodPost, "/"+parent+"/edgeCacheKeysets", createQ, body, &op); err != nil {
		return fmt.Errorf("importing edge-cache keyset: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Import (create) request issued for edge-cache keyset [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheKeysetsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheKeysetsList(cmd *cobra.Command, args []string) error {
	parent, err := s.edgeCacheKeysetsParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(edgeCacheRest()).paginate(ctx, "/"+parent+"/edgeCacheKeysets", nil, "edgeCacheKeysets", s.flagEdgeCacheKeysetsPageSize)
	if err != nil {
		return fmt.Errorf("listing edge-cache keysets: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagEdgeCacheKeysetsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheKeysetsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheKeysetsName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheKeysetsConfigFile, &body); err != nil {
		return err
	}
	mask := s.flagEdgeCacheKeysetsUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating edge-cache keyset: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for edge-cache keyset [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheKeysetsFormat)
}
