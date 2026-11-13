package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud run multi-region-services (#1051) ---
//
// The v0.279.0 generated run/v2 client does not expose the
// Projects.Locations.MultiRegionServices resource (it is a newer preview
// surface). Use the raw REST client against run.googleapis.com/v2 instead.

func (s *runCmds) initRunMultiRegionServices() {
	all := []*cobra.Command{
		s.runMRSDeleteCmd, s.runMRSDescribeCmd, s.runMRSListCmd,
		s.runMRSReplaceCmd, s.runMRSUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagRunMRSRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&s.flagRunMRSFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{s.runMRSReplaceCmd, s.runMRSUpdateCmd} {
		c.Flags().StringVar(&s.flagRunMRSConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the MultiRegionService body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.runMRSUpdateCmd.Flags().StringVar(&s.flagRunMRSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	s.runMRSListCmd.Flags().Int64Var(&s.flagRunMRSPageSize, "page-size", 0, "Maximum results per page")

	s.runMultiRegionServicesCmd.AddCommand(all...)
	s.runCmd.AddCommand(s.runMultiRegionServicesCmd)
}

// runMRSClient returns a REST client bound to the regional Cloud Run v2 endpoint.
func (s *runCmds) runMRSClient() *restClient {
	return s.inv.restClient(fmt.Sprintf("https://%s-run.googleapis.com/v2", s.flagRunMRSRegion))
}

func (s *runCmds) runMRSPath(project, mrs string) string {
	if strings.HasPrefix(mrs, "projects/") {
		return "/" + mrs
	}
	return fmt.Sprintf("/projects/%s/locations/%s/multiRegionServices/%s",
		project, s.flagRunMRSRegion, mrs)
}

func (s *runCmds) runMRSCollection(project string) string {
	return fmt.Sprintf("/projects/%s/locations/%s/multiRegionServices",
		project, s.flagRunMRSRegion)
}

func (s *runCmds) runMRSDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.runMRSClient().do(ctx, http.MethodDelete, s.runMRSPath(project, args[0]), nil, nil, &op); err != nil {
		return fmt.Errorf("deleting multi-region service: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for multi-region service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunMRSFormat)
}

func (s *runCmds) runMRSDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var mrs map[string]any
	if err := s.runMRSClient().do(ctx, http.MethodGet, s.runMRSPath(project, args[0]), nil, nil, &mrs); err != nil {
		return fmt.Errorf("describing multi-region service: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, mrs, s.flagRunMRSFormat)
}

func (s *runCmds) runMRSList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.runMRSClient().paginate(ctx, s.runMRSCollection(project), url.Values{}, "multiRegionServices", s.flagRunMRSPageSize)
	if err != nil {
		return fmt.Errorf("listing multi-region services: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagRunMRSFormat)
}

func (s *runCmds) runMRSReplace(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(s.flagRunMRSConfigFile, &body); err != nil {
		return err
	}
	body["name"] = strings.TrimPrefix(s.runMRSPath(project, args[0]), "/")
	ctx := cmd.Context()
	var op map[string]any
	if err := s.runMRSClient().do(ctx, http.MethodPut, s.runMRSPath(project, args[0]), nil, body, &op); err != nil {
		return fmt.Errorf("replacing multi-region service: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Replace request issued for multi-region service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunMRSFormat)
}

func (s *runCmds) runMRSUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(s.flagRunMRSConfigFile, &body); err != nil {
		return err
	}
	body["name"] = strings.TrimPrefix(s.runMRSPath(project, args[0]), "/")
	mask := s.flagRunMRSUpdateMask
	if mask == "" {
		fields := make([]string, 0, len(body))
		for k := range body {
			if k == "name" {
				continue
			}
			fields = append(fields, k)
		}
		mask = joinMask(fields)
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.runMRSClient().do(ctx, http.MethodPatch, s.runMRSPath(project, args[0]), q, body, &op); err != nil {
		return fmt.Errorf("updating multi-region service: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for multi-region service [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagRunMRSFormat)
}
