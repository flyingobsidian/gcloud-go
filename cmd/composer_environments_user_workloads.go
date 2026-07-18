package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	composer "google.golang.org/api/composer/v1"
)

// --- gcloud composer environments user-workloads-config-maps / user-workloads-secrets ---

var composerEnvCMCmd = &cobra.Command{Use: "user-workloads-config-maps", Short: "Manage user workload ConfigMaps"}
var composerEnvSecretsCmd = &cobra.Command{Use: "user-workloads-secrets", Short: "Manage user workload Secrets"}

var (
	flagComposerUWEnv        string
	flagComposerUWConfigFile string
	flagComposerUWPageSize   int64
)

var (
	composerCMCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create a user workload ConfigMap",
		Args: cobra.ExactArgs(1), RunE: runComposerCMCreate,
	}
	composerCMDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete a user workload ConfigMap",
		Args: cobra.ExactArgs(1), RunE: runComposerCMDelete,
	}
	composerCMDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe a user workload ConfigMap",
		Args: cobra.ExactArgs(1), RunE: runComposerCMDescribe,
	}
	composerCMListCmd = &cobra.Command{
		Use: "list", Short: "List user workload ConfigMaps",
		Args: cobra.NoArgs, RunE: runComposerCMList,
	}
	composerCMUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update a user workload ConfigMap",
		Args: cobra.ExactArgs(1), RunE: runComposerCMUpdate,
	}
	composerSecCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create a user workload Secret",
		Args: cobra.ExactArgs(1), RunE: runComposerSecCreate,
	}
	composerSecDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete a user workload Secret",
		Args: cobra.ExactArgs(1), RunE: runComposerSecDelete,
	}
	composerSecDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe a user workload Secret",
		Args: cobra.ExactArgs(1), RunE: runComposerSecDescribe,
	}
	composerSecListCmd = &cobra.Command{
		Use: "list", Short: "List user workload Secrets",
		Args: cobra.NoArgs, RunE: runComposerSecList,
	}
	composerSecUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update a user workload Secret",
		Args: cobra.ExactArgs(1), RunE: runComposerSecUpdate,
	}
)

func init() {
	cms := []*cobra.Command{composerCMCreateCmd, composerCMDeleteCmd, composerCMDescribeCmd, composerCMListCmd, composerCMUpdateCmd}
	secs := []*cobra.Command{composerSecCreateCmd, composerSecDeleteCmd, composerSecDescribeCmd, composerSecListCmd, composerSecUpdateCmd}
	for _, c := range append(cms, secs...) {
		c.Flags().StringVar(&flagComposerEnvLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagComposerUWEnv, "environment", "", "Environment ID that owns the resource (required)")
		_ = c.MarkFlagRequired("environment")
		c.Flags().StringVar(&flagComposerEnvFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{composerCMCreateCmd, composerCMUpdateCmd, composerSecCreateCmd, composerSecUpdateCmd} {
		c.Flags().StringVar(&flagComposerUWConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the resource body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{composerCMListCmd, composerSecListCmd} {
		c.Flags().Int64Var(&flagComposerUWPageSize, "page-size", 0, "Maximum results per page")
	}
	composerEnvCMCmd.AddCommand(cms...)
	composerEnvSecretsCmd.AddCommand(secs...)
}

func composerUWEnvName() (string, error) {
	return composerEnvResolvedName(flagComposerUWEnv)
}

func composerUWChild(collection, id string) (string, error) {
	env, err := composerUWEnvName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", env, collection, id), nil
}

func runComposerCMCreate(cmd *cobra.Command, args []string) error {
	parent, err := composerUWEnvName()
	if err != nil {
		return err
	}
	body := &composer.UserWorkloadsConfigMap{}
	if err := loadYAMLOrJSONInto(flagComposerUWConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = fmt.Sprintf("%s/userWorkloadsConfigMaps/%s", parent, args[0])
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating configmap: %w", err)
	}
	fmt.Printf("Created ConfigMap [%s].\n", args[0])
	return emitFormatted(got, flagComposerEnvFormat)
}

func runComposerCMDelete(cmd *cobra.Command, args []string) error {
	name, err := composerUWChild("userWorkloadsConfigMaps", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting configmap: %w", err)
	}
	fmt.Printf("Deleted ConfigMap [%s].\n", args[0])
	return nil
}

func runComposerCMDescribe(cmd *cobra.Command, args []string) error {
	name, err := composerUWChild("userWorkloadsConfigMaps", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing configmap: %w", err)
	}
	return emitFormatted(got, flagComposerEnvFormat)
}

func runComposerCMList(cmd *cobra.Command, args []string) error {
	parent, err := composerUWEnvName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*composer.UserWorkloadsConfigMap
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.List(parent).Context(ctx)
		if flagComposerUWPageSize > 0 {
			call = call.PageSize(flagComposerUWPageSize)
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
	return emitFormatted(all, flagComposerEnvFormat)
}

func runComposerCMUpdate(cmd *cobra.Command, args []string) error {
	name, err := composerUWChild("userWorkloadsConfigMaps", args[0])
	if err != nil {
		return err
	}
	body := &composer.UserWorkloadsConfigMap{}
	if err := loadYAMLOrJSONInto(flagComposerUWConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = name
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsConfigMaps.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating configmap: %w", err)
	}
	fmt.Printf("Updated ConfigMap [%s].\n", args[0])
	return emitFormatted(got, flagComposerEnvFormat)
}

func runComposerSecCreate(cmd *cobra.Command, args []string) error {
	parent, err := composerUWEnvName()
	if err != nil {
		return err
	}
	body := &composer.UserWorkloadsSecret{}
	if err := loadYAMLOrJSONInto(flagComposerUWConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = fmt.Sprintf("%s/userWorkloadsSecrets/%s", parent, args[0])
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsSecrets.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating secret: %w", err)
	}
	fmt.Printf("Created Secret [%s].\n", args[0])
	return emitFormatted(got, flagComposerEnvFormat)
}

func runComposerSecDelete(cmd *cobra.Command, args []string) error {
	name, err := composerUWChild("userWorkloadsSecrets", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Environments.UserWorkloadsSecrets.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting secret: %w", err)
	}
	fmt.Printf("Deleted Secret [%s].\n", args[0])
	return nil
}

func runComposerSecDescribe(cmd *cobra.Command, args []string) error {
	name, err := composerUWChild("userWorkloadsSecrets", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsSecrets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing secret: %w", err)
	}
	return emitFormatted(got, flagComposerEnvFormat)
}

func runComposerSecList(cmd *cobra.Command, args []string) error {
	parent, err := composerUWEnvName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*composer.UserWorkloadsSecret
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.UserWorkloadsSecrets.List(parent).Context(ctx)
		if flagComposerUWPageSize > 0 {
			call = call.PageSize(flagComposerUWPageSize)
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
	return emitFormatted(all, flagComposerEnvFormat)
}

func runComposerSecUpdate(cmd *cobra.Command, args []string) error {
	name, err := composerUWChild("userWorkloadsSecrets", args[0])
	if err != nil {
		return err
	}
	body := &composer.UserWorkloadsSecret{}
	if err := loadYAMLOrJSONInto(flagComposerUWConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = name
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.UserWorkloadsSecrets.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating secret: %w", err)
	}
	fmt.Printf("Updated Secret [%s].\n", args[0])
	return emitFormatted(got, flagComposerEnvFormat)
}
