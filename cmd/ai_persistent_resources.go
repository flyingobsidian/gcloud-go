package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai persistent-resources (#1460) ---

var aiPRCmd = &cobra.Command{Use: "persistent-resources", Short: "Manage Vertex AI persistent resources"}

var (
	flagAIPRRegion       string
	flagAIPRFormat       string
	flagAIPRConfigFile   string
	flagAIPRPRId         string
	flagAIPRPageSize     int64
)

var (
	aiPRCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a persistent resource",
		Args: cobra.NoArgs, RunE: runAIPRCreate,
	}
	aiPRDeleteCmd = &cobra.Command{
		Use: "delete PERSISTENT_RESOURCE", Short: "Delete a persistent resource",
		Args: cobra.ExactArgs(1), RunE: runAIPRDelete,
	}
	aiPRDescribeCmd = &cobra.Command{
		Use: "describe PERSISTENT_RESOURCE", Short: "Describe a persistent resource",
		Args: cobra.ExactArgs(1), RunE: runAIPRDescribe,
	}
	aiPRListCmd = &cobra.Command{
		Use: "list", Short: "List persistent resources",
		Args: cobra.NoArgs, RunE: runAIPRList,
	}
	aiPRRebootCmd = &cobra.Command{
		Use: "reboot PERSISTENT_RESOURCE", Short: "Reboot a persistent resource",
		Args: cobra.ExactArgs(1), RunE: runAIPRReboot,
	}
)

func init() {
	all := []*cobra.Command{
		aiPRCreateCmd, aiPRDeleteCmd, aiPRDescribeCmd, aiPRListCmd, aiPRRebootCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIPRRegion, "region", "", "Region where the persistent resource lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAIPRFormat, "format", "", "Output format")
	}
	aiPRCreateCmd.Flags().StringVar(&flagAIPRConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the PersistentResource body (required)")
	_ = aiPRCreateCmd.MarkFlagRequired("config-file")
	aiPRCreateCmd.Flags().StringVar(&flagAIPRPRId, "persistent-resource-id", "",
		"Caller-supplied persistent resource ID (required)")
	_ = aiPRCreateCmd.MarkFlagRequired("persistent-resource-id")
	aiPRListCmd.Flags().Int64Var(&flagAIPRPageSize, "page-size", 0, "Maximum results per page")

	aiPRCmd.AddCommand(all...)
	aiCmd.AddCommand(aiPRCmd)
}

func aiPRParent() (string, error) { return aiParent(flagAIPRRegion) }

func aiPRName(id string) (string, error) {
	parent, err := aiPRParent()
	if err != nil {
		return "", err
	}
	return aiChild("persistentResources", id, parent), nil
}

func runAIPRCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiPRParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1PersistentResource{}
	if err := loadYAMLOrJSONInto(flagAIPRConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIPRRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PersistentResources.Create(parent, body).
		PersistentResourceId(flagAIPRPRId).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating persistent resource: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Create request issued for persistent resource [%s] (operation: %s).\n", flagAIPRPRId, op.Name)
	return emitFormatted(op, flagAIPRFormat)
}

func runAIPRDelete(cmd *cobra.Command, args []string) error {
	name, err := aiPRName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIPRRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PersistentResources.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting persistent resource: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete request issued for persistent resource [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIPRFormat)
}

func runAIPRDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiPRName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIPRRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PersistentResources.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing persistent resource: %w", err)
	}
	return emitFormatted(got, flagAIPRFormat)
}

func runAIPRList(cmd *cobra.Command, args []string) error {
	parent, err := aiPRParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIPRRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1PersistentResource
	pageToken := ""
	for {
		call := svc.Projects.Locations.PersistentResources.List(parent).Context(ctx)
		if flagAIPRPageSize > 0 {
			call = call.PageSize(flagAIPRPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing persistent resources: %w", err)
		}
		all = append(all, resp.PersistentResources...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAIPRFormat)
}

func runAIPRReboot(cmd *cobra.Command, args []string) error {
	name, err := aiPRName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIPRRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PersistentResources.Reboot(name,
		&aiplatform.GoogleCloudAiplatformV1RebootPersistentResourceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rebooting persistent resource: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Reboot request issued for persistent resource [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIPRFormat)
}
