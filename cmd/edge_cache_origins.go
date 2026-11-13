package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud edge-cache origins (#1059) ---

func (s *edgeCacheCmds) initEdgeCacheOrigins() {
	all := []*cobra.Command{
		s.edgeCacheOriginsCreateCmd, s.edgeCacheOriginsDeleteCmd, s.edgeCacheOriginsDescribeCmd,
		s.edgeCacheOriginsExportCmd, s.edgeCacheOriginsImportCmd, s.edgeCacheOriginsListCmd,
		s.edgeCacheOriginsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagEdgeCacheOriginsLocation, "location", "", "Edge Cache location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagEdgeCacheOriginsFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.edgeCacheOriginsCreateCmd, s.edgeCacheOriginsUpdateCmd} {
		c.Flags().StringVar(&s.flagEdgeCacheOriginsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the EdgeCacheOrigin body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.edgeCacheOriginsUpdateCmd.Flags().StringVar(&s.flagEdgeCacheOriginsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated top-level key)")
	s.edgeCacheOriginsExportCmd.Flags().StringVar(&s.flagEdgeCacheOriginsDestination, "destination", "",
		"File to write the exported resource to (defaults to stdout)")
	s.edgeCacheOriginsImportCmd.Flags().StringVar(&s.flagEdgeCacheOriginsSource, "source", "",
		"Path to a YAML/JSON file with the EdgeCacheOrigin body (required)")
	_ = s.edgeCacheOriginsImportCmd.MarkFlagRequired("source")
	s.edgeCacheOriginsListCmd.Flags().Int64Var(&s.flagEdgeCacheOriginsPageSize, "page-size", 0, "Maximum results per page")

	s.edgeCacheOriginsCmd.AddCommand(all...)
	s.edgeCacheCmd.AddCommand(s.edgeCacheOriginsCmd)
}

func (s *edgeCacheCmds) edgeCacheOriginsParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagEdgeCacheOriginsLocation), nil
}

func (s *edgeCacheCmds) edgeCacheOriginsName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.edgeCacheOriginsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/edgeCacheOrigins/%s", parent, id), nil
}

func (s *edgeCacheCmds) runEdgeCacheOriginsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.edgeCacheOriginsParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheOriginsConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("edgeCacheOriginId", args[0])
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(edgeCacheRest).do(ctx, http.MethodPost, "/"+parent+"/edgeCacheOrigins", q, body, &op); err != nil {
		return fmt.Errorf("creating edge-cache origin: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for edge-cache origin [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheOriginsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheOriginsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(edgeCacheRest).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting edge-cache origin: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for edge-cache origin [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheOriginsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheOriginsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(edgeCacheRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing edge-cache origin: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagEdgeCacheOriginsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheOriginsExport(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(edgeCacheRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("exporting edge-cache origin: %w", err)
	}
	if s.flagEdgeCacheOriginsDestination == "" {
		return emitFormattedTo(s.inv.stdout, got, s.flagEdgeCacheOriginsFormat)
	}
	f, err := os.Create(s.flagEdgeCacheOriginsDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	return emitFormattedTo(f, got, s.flagEdgeCacheOriginsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheOriginsImport(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheOriginsSource, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	// Attempt update first; if the resource does not exist, fall through to create.
	var op map[string]any
	q := url.Values{}
	if mask := joinMask(dcTopLevelKeys(body)); mask != "" {
		q.Set("updateMask", mask)
	}
	err = s.inv.rest(edgeCacheRest).do(ctx, http.MethodPatch, "/"+name, q, body, &op)
	if err == nil {
		fmt.Fprintf(s.inv.stdout, "Import (update) request issued for edge-cache origin [%s].\n", args[0])
		return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheOriginsFormat)
	}
	var httpErr *restError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusNotFound {
		return fmt.Errorf("importing edge-cache origin: %w", err)
	}
	parent, perr := s.edgeCacheOriginsParent()
	if perr != nil {
		return perr
	}
	createQ := url.Values{}
	createQ.Set("edgeCacheOriginId", args[0])
	if err := s.inv.rest(edgeCacheRest).do(ctx, http.MethodPost, "/"+parent+"/edgeCacheOrigins", createQ, body, &op); err != nil {
		return fmt.Errorf("importing edge-cache origin: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Import (create) request issued for edge-cache origin [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheOriginsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheOriginsList(cmd *cobra.Command, args []string) error {
	parent, err := s.edgeCacheOriginsParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(edgeCacheRest).paginate(ctx, "/"+parent+"/edgeCacheOrigins", nil, "edgeCacheOrigins", s.flagEdgeCacheOriginsPageSize)
	if err != nil {
		return fmt.Errorf("listing edge-cache origins: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagEdgeCacheOriginsFormat)
}

func (s *edgeCacheCmds) runEdgeCacheOriginsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheOriginsName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheOriginsConfigFile, &body); err != nil {
		return err
	}
	mask := s.flagEdgeCacheOriginsUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(edgeCacheRest).do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating edge-cache origin: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for edge-cache origin [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheOriginsFormat)
}
