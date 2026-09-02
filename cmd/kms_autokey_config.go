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
	flagKmsAutokeyProject    string
	flagKmsAutokeyFormat     string
	flagKmsAutokeyConfigFile string
	flagKmsAutokeyMask       string
	flagKmsAutokeyKeyProject string
)

var kmsAutokeyDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe the AutokeyConfig for a folder or project",
	Args:  cobra.NoArgs,
	RunE:  runKmsAutokeyDescribe,
}

var kmsAutokeyShowEffectiveCmd = &cobra.Command{
	Use:   "show-effective-config",
	Short: "Show the effective AutokeyConfig for a folder or project",
	Args:  cobra.NoArgs,
	RunE:  runKmsAutokeyShowEffective,
}

var kmsAutokeyUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the AutokeyConfig for a folder or project",
	Args:  cobra.NoArgs,
	RunE:  runKmsAutokeyUpdate,
}

func init() {
	for _, c := range []*cobra.Command{
		kmsAutokeyDescribeCmd, kmsAutokeyShowEffectiveCmd, kmsAutokeyUpdateCmd,
	} {
		c.Flags().StringVar(&flagKmsAutokeyFormat, "format", "", "Output format")
		// --folder / --project are both optional; if neither is supplied
		// gcloud-python defaults to the current core/project.
		c.Flags().StringVar(&flagKmsAutokeyFolder, "folder", "", "Folder ID or resource name")
		c.Flags().StringVar(&flagKmsAutokeyProject, "project", "", "Project ID (defaults to core/project)")
	}
	kmsAutokeyUpdateCmd.Flags().StringVar(&flagKmsAutokeyConfigFile, "config-file", "", "YAML/JSON body for the AutokeyConfig")
	kmsAutokeyUpdateCmd.Flags().StringVar(&flagKmsAutokeyMask, "update-mask", "", "Fields to update; defaults to populated fields")
	// --key-project sets the KeyProject field on the AutokeyConfig; supported
	// under both folder and project scopes (570.0.0).
	kmsAutokeyUpdateCmd.Flags().StringVar(&flagKmsAutokeyKeyProject, "key-project", "", "Project that hosts Autokey-provisioned CryptoKeys (projects/PROJECT_ID)")

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

// kmsProjectAutokeyName returns "projects/PROJECT/autokeyConfig". Added in
// gcloud-python 570.0.0 to support project-level Autokey configuration.
func kmsProjectAutokeyName(project string) string {
	project = strings.TrimSpace(project)
	if strings.HasSuffix(project, "/autokeyConfig") {
		return project
	}
	if strings.HasPrefix(project, "projects/") {
		return project + "/autokeyConfig"
	}
	return fmt.Sprintf("projects/%s/autokeyConfig", project)
}

// autokeyResolveScope returns the resource name to operate on. A --folder
// wins; otherwise --project or the core/project property is used.
func autokeyResolveScope() (name string, isFolder bool, err error) {
	if flagKmsAutokeyFolder != "" {
		return kmsFolderAutokeyName(flagKmsAutokeyFolder), true, nil
	}
	project := flagKmsAutokeyProject
	if project == "" {
		p, perr := resolveProject()
		if perr != nil {
			return "", false, perr
		}
		project = p
	}
	return kmsProjectAutokeyName(project), false, nil
}

func runKmsAutokeyDescribe(cmd *cobra.Command, args []string) error {
	name, _, err := autokeyResolveScope()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// Same REST verb (GetAutokeyConfig) works for both folder/... and
	// projects/... resource names; the API routes based on the prefix.
	out, err := svc.Folders.GetAutokeyConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing autokey config: %w", err)
	}
	return emitFormatted(out, flagKmsAutokeyFormat)
}

func runKmsAutokeyShowEffective(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// --folder wins; otherwise --project (default: core/project).
	var parent string
	if flagKmsAutokeyFolder != "" {
		if strings.HasPrefix(flagKmsAutokeyFolder, "folders/") {
			parent = flagKmsAutokeyFolder
		} else {
			parent = "folders/" + flagKmsAutokeyFolder
		}
	} else {
		project := flagKmsAutokeyProject
		if project == "" {
			p, perr := resolveProject()
			if perr != nil {
				return perr
			}
			project = p
		}
		if strings.HasPrefix(project, "projects/") {
			parent = project
		} else {
			parent = "projects/" + project
		}
	}
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
	if flagKmsAutokeyConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagKmsAutokeyConfigFile, body); err != nil {
			return err
		}
	}
	if flagKmsAutokeyKeyProject != "" {
		body.KeyProject = flagKmsAutokeyKeyProject
	}
	if flagKmsAutokeyConfigFile == "" && flagKmsAutokeyKeyProject == "" {
		return fmt.Errorf("--config-file or --key-project is required")
	}
	mask := flagKmsAutokeyMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name, _, err := autokeyResolveScope()
	if err != nil {
		return err
	}
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
