package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	composer "google.golang.org/api/composer/v1"
)

// --- gcloud composer environments user-workloads-config-maps / user-workloads-secrets ---

func (s *composerCmds) initComposerEnvironmentsUserWorkloads() {
	cms := []*cobra.Command{s.composerCMCreateCmd, s.composerCMDeleteCmd, s.composerCMDescribeCmd, s.composerCMListCmd, s.composerCMUpdateCmd}
	secs := []*cobra.Command{s.composerSecCreateCmd, s.composerSecDeleteCmd, s.composerSecDescribeCmd, s.composerSecListCmd, s.composerSecUpdateCmd}
	for _, c := range append(cms, secs...) {
		c.Flags().StringVar(&s.flagComposerEnvLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagComposerUWEnv, "environment", "", "Environment ID that owns the resource (required)")
		_ = c.MarkFlagRequired("environment")
		c.Flags().StringVar(&s.flagComposerEnvFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.composerCMCreateCmd, s.composerCMUpdateCmd, s.composerSecCreateCmd, s.composerSecUpdateCmd} {
		c.Flags().StringVar(&s.flagComposerUWConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{s.composerCMListCmd, s.composerSecListCmd} {
		c.Flags().Int64Var(&s.flagComposerUWPageSize, "page-size", 0, "Maximum results per page")
	}
	s.composerEnvCMCmd.AddCommand(cms...)
	s.composerEnvSecretsCmd.AddCommand(secs...)
}

func (s *composerCmds) composerUWEnvName() (string, error) {
	return s.composerEnvResolvedName(s.flagComposerUWEnv)
}

func (s *composerCmds) composerUWChild(collection, id string) (string, error) {
	env, err := s.composerUWEnvName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", env, collection, id), nil
}

func (s *composerCmds) runComposerCMCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.composerUWEnvName()
	if err != nil {
		return err
	}
	body := &composer.UserWorkloadsConfigMap{}
	if err := loadYAMLOrJSONInto(s.flagComposerUWConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = fmt.Sprintf("%s/userWorkloadsConfigMaps/%s", parent, args[0])
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating configmap: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created ConfigMap [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerCMDelete(cmd *cobra.Command, args []string) error {
	name, err := s.composerUWChild("userWorkloadsConfigMaps", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting configmap: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted ConfigMap [%s].\n", args[0])
	return nil
}

func (s *composerCmds) runComposerCMDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.composerUWChild("userWorkloadsConfigMaps", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing configmap: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerCMList(cmd *cobra.Command, args []string) error {
	parent, err := s.composerUWEnvName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	var all []*composer.UserWorkloadsConfigMap
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.List(parent).Context(ctx)
		if s.flagComposerUWPageSize > 0 {
			call = call.PageSize(s.flagComposerUWPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing configmaps: %w", err)
		}
		all = append(all, resp.UserWorkloadsConfigMaps...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerCMUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.composerUWChild("userWorkloadsConfigMaps", args[0])
	if err != nil {
		return err
	}
	body := &composer.UserWorkloadsConfigMap{}
	if err := loadYAMLOrJSONInto(s.flagComposerUWConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = name
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating configmap: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated ConfigMap [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerSecCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.composerUWEnvName()
	if err != nil {
		return err
	}
	body := &composer.UserWorkloadsSecret{}
	if err := loadYAMLOrJSONInto(s.flagComposerUWConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = fmt.Sprintf("%s/userWorkloadsSecrets/%s", parent, args[0])
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsSecrets.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating secret: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created Secret [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerSecDelete(cmd *cobra.Command, args []string) error {
	name, err := s.composerUWChild("userWorkloadsSecrets", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Environments.UserWorkloadsSecrets.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting secret: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted Secret [%s].\n", args[0])
	return nil
}

func (s *composerCmds) runComposerSecDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.composerUWChild("userWorkloadsSecrets", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsSecrets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing secret: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerSecList(cmd *cobra.Command, args []string) error {
	parent, err := s.composerUWEnvName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	var all []*composer.UserWorkloadsSecret
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.UserWorkloadsSecrets.List(parent).Context(ctx)
		if s.flagComposerUWPageSize > 0 {
			call = call.PageSize(s.flagComposerUWPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing secrets: %w", err)
		}
		all = append(all, resp.UserWorkloadsSecrets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagComposerEnvFormat)
}

func (s *composerCmds) runComposerSecUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.composerUWChild("userWorkloadsSecrets", args[0])
	if err != nil {
		return err
	}
	body := &composer.UserWorkloadsSecret{}
	if err := loadYAMLOrJSONInto(s.flagComposerUWConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = name
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsSecrets.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating secret: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated Secret [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagComposerEnvFormat)
}
