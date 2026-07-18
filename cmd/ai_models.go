package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai models (#1458) ---

var aiModelsCmd = &cobra.Command{Use: "models", Short: "Manage Vertex AI models"}

var (
	flagAIMRegion       string
	flagAIMFormat       string
	flagAIMConfigFile   string
	flagAIMVersion      string
	flagAIMFilter       string
	flagAIMOrderBy      string
	flagAIMPageSize     int64
	flagAIMReadMask     string
)

var (
	aiMDeleteCmd = &cobra.Command{
		Use: "delete MODEL", Short: "Delete a model",
		Args: cobra.ExactArgs(1), RunE: runAIMDelete,
	}
	aiMDescribeCmd = &cobra.Command{
		Use: "describe MODEL", Short: "Describe a model",
		Args: cobra.ExactArgs(1), RunE: runAIMDescribe,
	}
	aiMListCmd = &cobra.Command{
		Use: "list", Short: "List models",
		Args: cobra.NoArgs, RunE: runAIMList,
	}
	aiMCopyCmd = &cobra.Command{
		Use: "copy", Short: "Copy a model",
		Args: cobra.NoArgs, RunE: runAIMCopy,
	}
	aiMUploadCmd = &cobra.Command{
		Use: "upload", Short: "Upload a model",
		Args: cobra.NoArgs, RunE: runAIMUpload,
	}
	aiMListVersionsCmd = &cobra.Command{
		Use: "list-version MODEL", Short: "List versions of a model",
		Args: cobra.ExactArgs(1), RunE: runAIMListVersions,
	}
	aiMDeleteVersionCmd = &cobra.Command{
		Use: "delete-version MODEL", Short: "Delete a specific version of a model",
		Long: "MODEL may be given as \"MODEL@VERSION\" or as MODEL together with " +
			"--version=VERSION. In either form the resource name sent to the " +
			"server is MODEL@VERSION.",
		Args: cobra.ExactArgs(1), RunE: runAIMDeleteVersion,
	}
)

func init() {
	all := []*cobra.Command{
		aiMDeleteCmd, aiMDescribeCmd, aiMListCmd,
		aiMCopyCmd, aiMUploadCmd,
		aiMListVersionsCmd, aiMDeleteVersionCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIMRegion, "region", "", "Region where the model lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAIMFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{aiMCopyCmd, aiMUploadCmd} {
		c.Flags().StringVar(&flagAIMConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	aiMListCmd.Flags().StringVar(&flagAIMFilter, "filter", "", "Server-side filter expression")
	aiMListCmd.Flags().StringVar(&flagAIMOrderBy, "order-by", "", "Order-by expression")
	aiMListCmd.Flags().Int64Var(&flagAIMPageSize, "page-size", 0, "Maximum results per page")
	aiMListCmd.Flags().StringVar(&flagAIMReadMask, "read-mask", "", "Field mask for reads")
	aiMListVersionsCmd.Flags().StringVar(&flagAIMFilter, "filter", "", "Server-side filter expression")
	aiMListVersionsCmd.Flags().StringVar(&flagAIMOrderBy, "order-by", "", "Order-by expression")
	aiMListVersionsCmd.Flags().Int64Var(&flagAIMPageSize, "page-size", 0, "Maximum results per page")
	aiMListVersionsCmd.Flags().StringVar(&flagAIMReadMask, "read-mask", "", "Field mask for reads")
	aiMDeleteVersionCmd.Flags().StringVar(&flagAIMVersion, "version", "",
		"Model version ID (may also be supplied inline as MODEL@VERSION)")

	aiModelsCmd.AddCommand(all...)
	aiCmd.AddCommand(aiModelsCmd)
}

func aiMParent() (string, error) { return aiParent(flagAIMRegion) }

func aiMName(id string) (string, error) {
	parent, err := aiMParent()
	if err != nil {
		return "", err
	}
	return aiChild("models", id, parent), nil
}

// aiMVersionedName resolves MODEL[@VERSION] plus --version into a full
// "projects/.../models/<id>@<version>" resource name. Either an inline
// "@VERSION" suffix on MODEL or a --version flag is required; specifying both
// values with a conflict is rejected.
func aiMVersionedName(rawModel, versionFlag string) (string, error) {
	base := rawModel
	inline := ""
	if i := strings.Index(rawModel, "@"); i >= 0 {
		base = rawModel[:i]
		inline = rawModel[i+1:]
	}
	version := inline
	if versionFlag != "" {
		if inline != "" && inline != versionFlag {
			return "", fmt.Errorf("model version specified twice: %q inline and %q via --version", inline, versionFlag)
		}
		version = versionFlag
	}
	if version == "" {
		return "", fmt.Errorf("model version required (use MODEL@VERSION or --version)")
	}
	parent, err := aiMParent()
	if err != nil {
		return "", err
	}
	full := aiChild("models", base, parent)
	return full + "@" + version, nil
}

func runAIMDelete(cmd *cobra.Command, args []string) error {
	name, err := aiMName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Models.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting model: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete request issued for model [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIMFormat)
}

func runAIMDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiMName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Models.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing model: %w", err)
	}
	return emitFormatted(got, flagAIMFormat)
}

func runAIMList(cmd *cobra.Command, args []string) error {
	parent, err := aiMParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Model
	pageToken := ""
	for {
		call := svc.Projects.Locations.Models.List(parent).Context(ctx)
		if flagAIMFilter != "" {
			call = call.Filter(flagAIMFilter)
		}
		if flagAIMOrderBy != "" {
			call = call.OrderBy(flagAIMOrderBy)
		}
		if flagAIMPageSize > 0 {
			call = call.PageSize(flagAIMPageSize)
		}
		if flagAIMReadMask != "" {
			call = call.ReadMask(flagAIMReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing models: %w", err)
		}
		all = append(all, resp.Models...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAIMFormat)
}

func runAIMCopy(cmd *cobra.Command, args []string) error {
	parent, err := aiMParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1CopyModelRequest{}
	if err := loadYAMLOrJSONInto(flagAIMConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Models.Copy(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("copying model: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Copy request issued (operation: %s).\n", op.Name)
	return emitFormatted(op, flagAIMFormat)
}

func runAIMUpload(cmd *cobra.Command, args []string) error {
	parent, err := aiMParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1UploadModelRequest{}
	if err := loadYAMLOrJSONInto(flagAIMConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Models.Upload(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("uploading model: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Upload request issued (operation: %s).\n", op.Name)
	return emitFormatted(op, flagAIMFormat)
}

func runAIMListVersions(cmd *cobra.Command, args []string) error {
	name, err := aiMName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Model
	pageToken := ""
	for {
		call := svc.Projects.Locations.Models.ListVersions(name).Context(ctx)
		if flagAIMFilter != "" {
			call = call.Filter(flagAIMFilter)
		}
		if flagAIMOrderBy != "" {
			call = call.OrderBy(flagAIMOrderBy)
		}
		if flagAIMPageSize > 0 {
			call = call.PageSize(flagAIMPageSize)
		}
		if flagAIMReadMask != "" {
			call = call.ReadMask(flagAIMReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing model versions: %w", err)
		}
		all = append(all, resp.Models...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAIMFormat)
}

func runAIMDeleteVersion(cmd *cobra.Command, args []string) error {
	name, err := aiMVersionedName(args[0], flagAIMVersion)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIMRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Models.DeleteVersion(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting model version: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete-version request issued for [%s] (operation: %s).\n", name, op.Name)
	return emitFormatted(op, flagAIMFormat)
}
