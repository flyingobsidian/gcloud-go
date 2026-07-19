package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub runtime-project-attachments (#1165) ---

var apihubRPACmd = &cobra.Command{
	Use:   "runtime-project-attachments",
	Short: "Manage API Hub runtime project attachments",
}

var (
	flagAPRPALocation       string
	flagAPRPAFormat         string
	flagAPRPAConfigFile     string
	flagAPRPAFilter         string
	flagAPRPAPageSize       int64
	flagAPRPARuntimeProject string
)

var (
	apihubRPACreateCmd = &cobra.Command{
		Use: "create ATTACHMENT", Short: "Create a runtime project attachment",
		Args: cobra.ExactArgs(1), RunE: runAPRPACreate,
	}
	apihubRPADeleteCmd = &cobra.Command{
		Use: "delete ATTACHMENT", Short: "Delete a runtime project attachment",
		Args: cobra.ExactArgs(1), RunE: runAPRPADelete,
	}
	apihubRPADescribeCmd = &cobra.Command{
		Use: "describe ATTACHMENT", Short: "Describe a runtime project attachment",
		Args: cobra.ExactArgs(1), RunE: runAPRPADescribe,
	}
	apihubRPAListCmd = &cobra.Command{
		Use: "list", Short: "List runtime project attachments in a location",
		Args: cobra.NoArgs, RunE: runAPRPAList,
	}
	apihubRPALookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up the runtime project attachment for a runtime project",
		Args: cobra.NoArgs, RunE: runAPRPALookup,
	}
)

func init() {
	all := []*cobra.Command{
		apihubRPACreateCmd, apihubRPADeleteCmd, apihubRPADescribeCmd,
		apihubRPAListCmd, apihubRPALookupCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAPRPALocation, "location", "",
			"Location that owns the runtime project attachments (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagAPRPAFormat, "format", "", "Output format")
	}
	apihubRPACreateCmd.Flags().StringVar(&flagAPRPAConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the RuntimeProjectAttachment body (required)")
	_ = apihubRPACreateCmd.MarkFlagRequired("config-file")
	apihubRPAListCmd.Flags().StringVar(&flagAPRPAFilter, "filter", "", "Server-side filter expression")
	apihubRPAListCmd.Flags().Int64Var(&flagAPRPAPageSize, "page-size", 0, "Maximum number of results per page")
	apihubRPALookupCmd.Flags().StringVar(&flagAPRPARuntimeProject, "runtime-project", "",
		"Runtime project id or number to look up the attachment for (required)")
	_ = apihubRPALookupCmd.MarkFlagRequired("runtime-project")

	apihubRPACmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubRPACmd)
}

func apihubRPAParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagAPRPALocation), nil
}

func apihubRPAName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := apihubRPAParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/runtimeProjectAttachments/%s", parent, id), nil
}

func runAPRPACreate(cmd *cobra.Command, args []string) error {
	parent, err := apihubRPAParent()
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1RuntimeProjectAttachment{}
	if err := loadYAMLOrJSONInto(flagAPRPAConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.RuntimeProjectAttachments.Create(parent, body).
		RuntimeProjectAttachmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating runtime project attachment: %w", err)
	}
	return emitFormatted(got, flagAPRPAFormat)
}

func runAPRPADelete(cmd *cobra.Command, args []string) error {
	name, err := apihubRPAName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.RuntimeProjectAttachments.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting runtime project attachment: %w", err)
	}
	fmt.Printf("Deleted runtime project attachment [%s].\n", args[0])
	return nil
}

func runAPRPADescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubRPAName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.RuntimeProjectAttachments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing runtime project attachment: %w", err)
	}
	return emitFormatted(got, flagAPRPAFormat)
}

func runAPRPAList(cmd *cobra.Command, args []string) error {
	parent, err := apihubRPAParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1RuntimeProjectAttachment
	pageToken := ""
	for {
		call := svc.Projects.Locations.RuntimeProjectAttachments.List(parent).Context(ctx)
		if flagAPRPAFilter != "" {
			call = call.Filter(flagAPRPAFilter)
		}
		if flagAPRPAPageSize > 0 {
			call = call.PageSize(flagAPRPAPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing runtime project attachments: %w", err)
		}
		all = append(all, resp.RuntimeProjectAttachments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAPRPAFormat)
}

func runAPRPALookup(cmd *cobra.Command, args []string) error {
	// The Lookup RPC takes name = projects/{runtime-project}/locations/{location}
	// and searches across regions to find the attachment for the runtime project.
	name := fmt.Sprintf("projects/%s/locations/%s", flagAPRPARuntimeProject, flagAPRPALocation)
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.LookupRuntimeProjectAttachment(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up runtime project attachment: %w", err)
	}
	return emitFormatted(got, flagAPRPAFormat)
}
