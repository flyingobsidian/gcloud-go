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

// --- gcloud edge-cache services (#1060) ---

func (s *edgeCacheCmds) initEdgeCacheServices() {
	all := []*cobra.Command{
		s.edgeCacheServicesDeleteCmd, s.edgeCacheServicesDescribeCmd,
		s.edgeCacheServicesExportCmd, s.edgeCacheServicesImportCmd,
		s.edgeCacheServicesInvalidateCmd, s.edgeCacheServicesListCmd,
		s.edgeCacheServicesUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagEdgeCacheServicesLocation, "location", "", "Edge Cache location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagEdgeCacheServicesFormat, "format", "", "Output format")
	}
	s.edgeCacheServicesUpdateCmd.Flags().StringVar(&s.flagEdgeCacheServicesConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the EdgeCacheService body (required)")
	_ = s.edgeCacheServicesUpdateCmd.MarkFlagRequired("config-file")
	s.edgeCacheServicesUpdateCmd.Flags().StringVar(&s.flagEdgeCacheServicesUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated top-level key)")
	s.edgeCacheServicesExportCmd.Flags().StringVar(&s.flagEdgeCacheServicesDestination, "destination", "",
		"File to write the exported resource to (defaults to stdout)")
	s.edgeCacheServicesImportCmd.Flags().StringVar(&s.flagEdgeCacheServicesSource, "source", "",
		"Path to a YAML/JSON file with the EdgeCacheService body (required)")
	_ = s.edgeCacheServicesImportCmd.MarkFlagRequired("source")
	s.edgeCacheServicesInvalidateCmd.Flags().StringVar(&s.flagEdgeCacheServicesInvalidateFn, "config-file", "",
		"Path to a YAML/JSON file with the InvalidateCache request body (required)")
	_ = s.edgeCacheServicesInvalidateCmd.MarkFlagRequired("config-file")
	s.edgeCacheServicesListCmd.Flags().Int64Var(&s.flagEdgeCacheServicesPageSize, "page-size", 0, "Maximum results per page")

	s.edgeCacheServicesCmd.AddCommand(all...)
	s.edgeCacheCmd.AddCommand(s.edgeCacheServicesCmd)
}

func (s *edgeCacheCmds) edgeCacheServicesParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagEdgeCacheServicesLocation), nil
}

func (s *edgeCacheCmds) edgeCacheServicesName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.edgeCacheServicesParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/edgeCacheServices/%s", parent, id), nil
}

func (s *edgeCacheCmds) runEdgeCacheServicesDelete(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting edge-cache service: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for edge-cache service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheServicesFormat)
}

func (s *edgeCacheCmds) runEdgeCacheServicesDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing edge-cache service: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagEdgeCacheServicesFormat)
}

func (s *edgeCacheCmds) runEdgeCacheServicesExport(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("exporting edge-cache service: %w", err)
	}
	if s.flagEdgeCacheServicesDestination == "" {
		return emitFormattedTo(s.inv.stdout, got, s.flagEdgeCacheServicesFormat)
	}
	f, err := os.Create(s.flagEdgeCacheServicesDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer f.Close()
	return emitFormattedTo(f, got, s.flagEdgeCacheServicesFormat)
}

func (s *edgeCacheCmds) runEdgeCacheServicesImport(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheServicesSource, &body); err != nil {
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
		fmt.Fprintf(s.inv.stdout, "Import (update) request issued for edge-cache service [%s].\n", args[0])
		return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheServicesFormat)
	}
	var httpErr *restError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusNotFound {
		return fmt.Errorf("importing edge-cache service: %w", err)
	}
	parent, perr := s.edgeCacheServicesParent()
	if perr != nil {
		return perr
	}
	createQ := url.Values{}
	createQ.Set("edgeCacheServiceId", args[0])
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodPost, "/"+parent+"/edgeCacheServices", createQ, body, &op); err != nil {
		return fmt.Errorf("importing edge-cache service: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Import (create) request issued for edge-cache service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheServicesFormat)
}

func (s *edgeCacheCmds) runEdgeCacheServicesInvalidate(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheServicesInvalidateFn, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(edgeCacheRest()).do(ctx, http.MethodPost, "/"+name+":invalidateCache", nil, body, &op); err != nil {
		return fmt.Errorf("invalidating edge-cache service cache: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Invalidate-cache request issued for edge-cache service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheServicesFormat)
}

func (s *edgeCacheCmds) runEdgeCacheServicesList(cmd *cobra.Command, args []string) error {
	parent, err := s.edgeCacheServicesParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(edgeCacheRest()).paginate(ctx, "/"+parent+"/edgeCacheServices", nil, "edgeCacheServices", s.flagEdgeCacheServicesPageSize)
	if err != nil {
		return fmt.Errorf("listing edge-cache services: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagEdgeCacheServicesFormat)
}

func (s *edgeCacheCmds) runEdgeCacheServicesUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.edgeCacheServicesName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagEdgeCacheServicesConfigFile, &body); err != nil {
		return err
	}
	mask := s.flagEdgeCacheServicesUpdateMask
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
		return fmt.Errorf("updating edge-cache service: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for edge-cache service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagEdgeCacheServicesFormat)
}
