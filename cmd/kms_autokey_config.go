package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms autokey-config (#1100) ---

func (s *kmsCmds) initKmsAutokeyConfig() {
	for _, c := range []*cobra.Command{
		s.kmsAutokeyDescribeCmd, s.kmsAutokeyShowEffectiveCmd, s.kmsAutokeyUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagKmsAutokeyFormat, "format", "", "Output format")
		// --folder / --project are both optional; if neither is supplied
		// gcloud-python defaults to the current core/project.
		c.Flags().StringVar(&s.flagKmsAutokeyFolder, "folder", "", "Folder ID or resource name")
		c.Flags().StringVar(&s.flagKmsAutokeyProject, "project", "", "Project ID (defaults to core/project)")
	}
	s.kmsAutokeyUpdateCmd.Flags().StringVar(&s.flagKmsAutokeyConfigFile, "config-file", "", "YAML/JSON body for the AutokeyConfig")
	s.kmsAutokeyUpdateCmd.Flags().StringVar(&s.flagKmsAutokeyMask, "update-mask", "", "Fields to update; defaults to populated fields")
	// --key-project sets the KeyProject field on the AutokeyConfig; supported
	// under both folder and project scopes (570.0.0).
	s.kmsAutokeyUpdateCmd.Flags().StringVar(&s.flagKmsAutokeyKeyProject, "key-project", "", "Project that hosts Autokey-provisioned CryptoKeys (projects/PROJECT_ID)")

	s.kmsAutokeyConfigCmd.AddCommand(s.kmsAutokeyDescribeCmd, s.kmsAutokeyShowEffectiveCmd, s.kmsAutokeyUpdateCmd)
	s.kmsCmd.AddCommand(s.kmsAutokeyConfigCmd)
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
func (s *kmsCmds) autokeyResolveScope() (name string, isFolder bool, err error) {
	if s.flagKmsAutokeyFolder != "" {
		return kmsFolderAutokeyName(s.flagKmsAutokeyFolder), true, nil
	}
	project := s.flagKmsAutokeyProject
	if project == "" {
		p, perr := s.inv.project()
		if perr != nil {
			return "", false, perr
		}
		project = p
	}
	return kmsProjectAutokeyName(project), false, nil
}

func (s *kmsCmds) runKmsAutokeyDescribe(cmd *cobra.Command, args []string) error {
	name, _, err := s.autokeyResolveScope()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	// Same REST verb (GetAutokeyConfig) works for both folder/... and
	// projects/... resource names; the API routes based on the prefix.
	out, err := svc.Folders.GetAutokeyConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing autokey config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsAutokeyFormat)
}

func (s *kmsCmds) runKmsAutokeyShowEffective(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	// --folder wins; otherwise --project (default: core/project).
	var parent string
	if s.flagKmsAutokeyFolder != "" {
		if strings.HasPrefix(s.flagKmsAutokeyFolder, "folders/") {
			parent = s.flagKmsAutokeyFolder
		} else {
			parent = "folders/" + s.flagKmsAutokeyFolder
		}
	} else {
		project := s.flagKmsAutokeyProject
		if project == "" {
			p, perr := s.inv.project()
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
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsAutokeyFormat)
}

func (s *kmsCmds) runKmsAutokeyUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.AutokeyConfig{}
	if s.flagKmsAutokeyConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagKmsAutokeyConfigFile, body); err != nil {
			return err
		}
	}
	if s.flagKmsAutokeyKeyProject != "" {
		body.KeyProject = s.flagKmsAutokeyKeyProject
	}
	if s.flagKmsAutokeyConfigFile == "" && s.flagKmsAutokeyKeyProject == "" {
		return fmt.Errorf("--config-file or --key-project is required")
	}
	mask := s.flagKmsAutokeyMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name, _, err := s.autokeyResolveScope()
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
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsAutokeyFormat)
}
