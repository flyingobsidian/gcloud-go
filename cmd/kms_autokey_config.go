package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms autokey-config (#1100) ---

var kmsAutokeyConfigCmd = &cobra.Command{
	Use:   "autokey-config",
	Short: "Manage Cloud KMS Autokey configurations",
}

var (
	flagKmsAutokeyFolder     string
	flagKmsAutokeyFormat     string
	flagKmsAutokeyConfigFile string
	flagKmsAutokeyMask       string
)

var kmsAutokeyDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe the AutokeyConfig for a folder",
	Args:  cobra.NoArgs,
	RunE:  runKmsAutokeyDescribe,
}

var kmsAutokeyShowEffectiveCmd = &cobra.Command{
	Use:   "show-effective-config",
	Short: "Show the effective AutokeyConfig for a project",
	Args:  cobra.NoArgs,
	RunE:  runKmsAutokeyShowEffective,
}

var kmsAutokeyUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the AutokeyConfig for a folder",
	Args:  cobra.NoArgs,
	RunE:  runKmsAutokeyUpdate,
}

func init() {
	for _, c := range []*cobra.Command{
		kmsAutokeyDescribeCmd, kmsAutokeyShowEffectiveCmd, kmsAutokeyUpdateCmd,
	} {
		c.Flags().StringVar(&flagKmsAutokeyFormat, "format", "", "Output format")
	}
	kmsAutokeyDescribeCmd.Flags().StringVar(&flagKmsAutokeyFolder, "folder", "", "Folder ID or resource name (required)")
	_ = kmsAutokeyDescribeCmd.MarkFlagRequired("folder")

	kmsAutokeyUpdateCmd.Flags().StringVar(&flagKmsAutokeyFolder, "folder", "", "Folder ID or resource name (required)")
	kmsAutokeyUpdateCmd.Flags().StringVar(&flagKmsAutokeyConfigFile, "config-file", "", "YAML/JSON body for the AutokeyConfig (required)")
	kmsAutokeyUpdateCmd.Flags().StringVar(&flagKmsAutokeyMask, "update-mask", "", "Fields to update; defaults to populated fields")
	_ = kmsAutokeyUpdateCmd.MarkFlagRequired("folder")
	_ = kmsAutokeyUpdateCmd.MarkFlagRequired("config-file")

	kmsAutokeyConfigCmd.AddCommand(kmsAutokeyDescribeCmd, kmsAutokeyShowEffectiveCmd, kmsAutokeyUpdateCmd)
	kmsCmd.AddCommand(kmsAutokeyConfigCmd)
}

// kmsFolderAutokeyName returns "folders/FOLDER/autokeyConfig", accepting either
// a bare folder id or a full "folders/FOLDER" / autokeyConfig resource name.
func kmsFolderAutokeyName(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasSuffix(raw, "/autokeyConfig") {
		return raw
	}
	if strings.HasPrefix(raw, "folders/") {
		return raw + "/autokeyConfig"
	}
	return fmt.Sprintf("folders/%s/autokeyConfig", raw)
}

func runKmsAutokeyDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := kmsFolderAutokeyName(flagKmsAutokeyFolder)
	out, err := svc.Folders.GetAutokeyConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing autokey config: %w", err)
	}
	return emitFormatted(out, flagKmsAutokeyFormat)
}

func runKmsAutokeyShowEffective(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := "projects/" + project
	out, err := svc.Projects.ShowEffectiveAutokeyConfig(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("showing effective autokey config: %w", err)
	}
	return emitFormatted(out, flagKmsAutokeyFormat)
}

func runKmsAutokeyUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &cloudkms.AutokeyConfig{}
	if err := loadYAMLOrJSONInto(flagKmsAutokeyConfigFile, body); err != nil {
		return err
	}
	mask := flagKmsAutokeyMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := kmsFolderAutokeyName(flagKmsAutokeyFolder)
	call := svc.Folders.UpdateAutokeyConfig(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating autokey config: %w", err)
	}
	return emitFormatted(out, flagKmsAutokeyFormat)
}
