package cmd

import (
	"context"
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

var runMultiRegionServicesCmd = &cobra.Command{Use: "multi-region-services", Short: "Manage Cloud Run multi-region services"}

var (
	flagRunMRSRegion     string
	flagRunMRSFormat     string
	flagRunMRSConfigFile string
	flagRunMRSUpdateMask string
	flagRunMRSPageSize   int64
)

var (
	runMRSDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a Cloud Run multi-region service",
		Args: cobra.ExactArgs(1), RunE: runMRSDelete,
	}
	runMRSDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a Cloud Run multi-region service",
		Args: cobra.ExactArgs(1), RunE: runMRSDescribe,
	}
	runMRSListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run multi-region services",
		Args: cobra.NoArgs, RunE: runMRSList,
	}
	runMRSReplaceCmd = &cobra.Command{
		Use: "replace SERVICE", Short: "Replace a Cloud Run multi-region service (PUT)",
		Args: cobra.ExactArgs(1), RunE: runMRSReplace,
	}
	runMRSUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a Cloud Run multi-region service (PATCH)",
		Args: cobra.ExactArgs(1), RunE: runMRSUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		runMRSDeleteCmd, runMRSDescribeCmd, runMRSListCmd,
		runMRSReplaceCmd, runMRSUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagRunMRSRegion, "region", "", "Cloud Run region (required)")
		c.Flags().StringVar(&flagRunMRSFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{runMRSReplaceCmd, runMRSUpdateCmd} {
		c.Flags().StringVar(&flagRunMRSConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the MultiRegionService body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	runMRSUpdateCmd.Flags().StringVar(&flagRunMRSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update")
	runMRSListCmd.Flags().Int64Var(&flagRunMRSPageSize, "page-size", 0, "Maximum results per page")

	runMultiRegionServicesCmd.AddCommand(all...)
	runCmd.AddCommand(runMultiRegionServicesCmd)
}

// runMRSClient returns a REST client bound to the regional Cloud Run v2 endpoint.
func runMRSClient() *restClient {
	return newRESTClient(fmt.Sprintf("https://%s-run.googleapis.com/v2", flagRunMRSRegion))
}

func runMRSPath(project, mrs string) string {
	if strings.HasPrefix(mrs, "projects/") {
		return "/" + mrs
	}
	return fmt.Sprintf("/projects/%s/locations/%s/multiRegionServices/%s",
		project, flagRunMRSRegion, mrs)
}

func runMRSCollection(project string) string {
	return fmt.Sprintf("/projects/%s/locations/%s/multiRegionServices",
		project, flagRunMRSRegion)
}

func runMRSDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := runMRSClient().do(ctx, http.MethodDelete, runMRSPath(project, args[0]), nil, nil, &op); err != nil {
		return fmt.Errorf("deleting multi-region service: %w", err)
	}
	fmt.Printf("Delete request issued for multi-region service [%s].\n", args[0])
	return emitFormatted(op, flagRunMRSFormat)
}

func runMRSDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var mrs map[string]any
	if err := runMRSClient().do(ctx, http.MethodGet, runMRSPath(project, args[0]), nil, nil, &mrs); err != nil {
		return fmt.Errorf("describing multi-region service: %w", err)
	}
	return emitFormatted(mrs, flagRunMRSFormat)
}

func runMRSList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := runMRSClient().paginate(ctx, runMRSCollection(project), url.Values{}, "multiRegionServices", flagRunMRSPageSize)
	if err != nil {
		return fmt.Errorf("listing multi-region services: %w", err)
	}
	return emitFormatted(items, flagRunMRSFormat)
}

func runMRSReplace(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(flagRunMRSConfigFile, &body); err != nil {
		return err
	}
	body["name"] = strings.TrimPrefix(runMRSPath(project, args[0]), "/")
	ctx := context.Background()
	var op map[string]any
	if err := runMRSClient().do(ctx, http.MethodPut, runMRSPath(project, args[0]), nil, body, &op); err != nil {
		return fmt.Errorf("replacing multi-region service: %w", err)
	}
	fmt.Printf("Replace request issued for multi-region service [%s].\n", args[0])
	return emitFormatted(op, flagRunMRSFormat)
}

func runMRSUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(flagRunMRSConfigFile, &body); err != nil {
		return err
	}
	body["name"] = strings.TrimPrefix(runMRSPath(project, args[0]), "/")
	mask := flagRunMRSUpdateMask
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
	ctx := context.Background()
	var op map[string]any
	if err := runMRSClient().do(ctx, http.MethodPatch, runMRSPath(project, args[0]), q, body, &op); err != nil {
		return fmt.Errorf("updating multi-region service: %w", err)
	}
	fmt.Printf("Update request issued for multi-region service [%s].\n", args[0])
	return emitFormatted(op, flagRunMRSFormat)
}
